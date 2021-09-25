package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var crawlerID = atomic.NewInt32(0)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	// The crawler that generated this result
	CrawlerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	Neighbors []peer.AddrInfo

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred during the crawl
	Error error

	// The above error transferred to a known error
	DialError string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r *Result) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer.
func (r *Result) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

// Crawler encapsulates a libp2p host that crawls the network.
type Crawler struct {
	*service.Service

	host         host.Host
	config       *config.Config
	pm           *pb.ProtocolMessenger
	crawledPeers int
}

// NewCrawler initializes a new crawler based on the given configuration.
func NewCrawler(h host.Host, conf *config.Config) (*Crawler, error) {
	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(conf.Protocols),
		timeout:   conf.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Crawler{
		Service: service.New(fmt.Sprintf("crawler-%02d", crawlerID.Load())),
		host:    h,
		pm:      pm,
		config:  conf,
	}
	crawlerID.Inc()

	return c, nil
}

// StartCrawling enters an endless loop and consumes crawl jobs from the crawl queue
// and publishes its result on the results queue until it is told to stop or the
// crawl queue was closed.
func (c *Crawler) StartCrawling(crawlQueue *queue.FIFO, resultsQueue *queue.FIFO) {
	c.ServiceStarted()
	defer c.ServiceStopped()
	ctx := c.ServiceContext()

	for {
		// Give the shutdown signal precedence
		select {
		case <-c.SigShutdown():
			return
		default:
		}

		select {
		case elem, ok := <-crawlQueue.Consume():
			if !ok {
				// The crawl queue was closed
				return
			}
			result := c.handleCrawlJob(ctx, elem.(peer.AddrInfo))
			resultsQueue.Push(result)
		case <-c.SigShutdown():
			return
		}
	}
}

// handleCrawlJob takes peer address information and crawls (connects and fetches neighbors).
func (c *Crawler) handleCrawlJob(ctx context.Context, pi peer.AddrInfo) Result {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.Identifier(),
		"targetID":   pi.ID.Pretty()[:16],
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	cr := Result{
		CrawlerID:      c.Identifier(),
		Peer:           filterPrivateMaddrs(pi),
		CrawlStartTime: time.Now(),
	}

	cr.ConnectStartTime = time.Now()
	cr.Error = c.connect(ctx, cr.Peer) // use filtered addr list
	cr.ConnectEndTime = time.Now()

	// If we could successfully connect to the peer we actually crawl it.
	if cr.Error == nil {

		ps := c.host.Peerstore()

		// Extract agent
		if agent, err := ps.Get(pi.ID, "AgentVersion"); err == nil {
			cr.Agent = agent.(string)
		}

		// Extract protocols
		if protocols, err := ps.GetProtocols(pi.ID); err == nil {
			cr.Protocols = protocols
		}

		// Fetch all neighbors
		cr.Neighbors, cr.Error = c.fetchNeighbors(ctx, pi)
	}

	if cr.Error != nil {
		cr.DialError = db.DialError(cr.Error)
	}

	// Free connection resources
	if err := c.host.Network().ClosePeer(pi.ID); err != nil {
		log.WithError(err).WithField("targetID", pi.ID.Pretty()[:16]).Warnln("Could not close connection to peer")
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	cr.CrawlEndTime = time.Now()

	return cr
}

// connect strips all private multi addresses in `pi` and establishes a connection to the given peer.
// It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) error {
	stats.Record(ctx, metrics.CrawlConnectsCount.M(1))

	if len(pi.Addrs) == 0 {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
	defer cancel()

	if err := c.host.Connect(ctx, pi); err != nil {
		stats.Record(ctx, metrics.CrawlConnectErrorsCount.M(1))
		return err
	}

	return nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL). The returned peers are streamed
// to the results channel.
func (c *Crawler) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) ([]peer.AddrInfo, error) {
	var allNeighbors []peer.AddrInfo
	rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(pi.ID), time.Hour, nil, time.Hour, nil)
	if err != nil {
		return allNeighbors, err
	}

	allNeighborsLk := sync.RWMutex{}
	errg := errgroup.Group{}
	for i := uint(0); i <= 15; i++ { // 15 is maximum
		count := i
		errg.Go(func() error {
			// Generate a peer with the given common prefix length
			rpi, err := rt.GenRandPeerID(count)
			if err != nil {
				return errors.Wrapf(err, "generating random peer ID with CPL %d", count)
			}

			neighbors, err := c.pm.GetClosestPeers(ctx, pi.ID, rpi)
			if err != nil {
				return errors.Wrapf(err, "getting closest peer with CPL %d", count)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()
			for _, n := range neighbors {
				allNeighbors = append(allNeighbors, *n)
			}

			return nil
		})
	}
	err = errg.Wait()
	stats.Record(c.ServiceContext(),
		metrics.FetchedNeighborsCount.M(float64(len(allNeighbors))),
	)
	return allNeighbors, err
}

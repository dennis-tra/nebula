package crawl

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

var crawlerID = atomic.NewInt32(0)

// Crawler encapsulates a libp2p host that crawls the network.
type Crawler struct {
	id           string
	host         host.Host
	config       *config.Config
	pm           *pb.ProtocolMessenger
	crawledPeers int
	done         chan struct{}
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
		id:     fmt.Sprintf("crawler-%02d", crawlerID.Inc()),
		host:   h,
		pm:     pm,
		config: conf,
		done:   make(chan struct{}),
	}

	return c, nil
}

// StartCrawling enters an endless loop and consumes crawl jobs from the crawl queue
// and publishes its result on the results queue until it is told to stop or the
// crawl queue was closed.
func (c *Crawler) StartCrawling(ctx context.Context, crawlQueue *queue.FIFO[peer.AddrInfo], resultsQueue *queue.FIFO[Result]) {
	defer close(c.done)
	for {
		// Give the shutdown signal precedence
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case pi, ok := <-crawlQueue.Consume():
			if !ok {
				// The crawl queue was closed
				return
			}
			result := c.handleCrawlJob(ctx, pi)
			resultsQueue.Push(result)
		}
	}
}

// handleCrawlJob takes peer address information and crawls (connects and fetches neighbors).
func (c *Crawler) handleCrawlJob(ctx context.Context, pi peer.AddrInfo) Result {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   utils.FmtPeerID(pi.ID),
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	cr := Result{
		CrawlerID:      c.id,
		Peer:           utils.FilterPrivateMaddrs(pi),
		CrawlStartTime: time.Now(),
		RoutingTable:   &RoutingTable{PeerID: pi.ID},
	}

	cr.ConnectStartTime = time.Now()
	cr.ConnectError = c.connect(ctx, cr.Peer) // use filtered addr list
	cr.ConnectEndTime = time.Now()

	// If we could successfully connect to the peer we actually crawl it.
	if cr.ConnectError == nil {

		// Fetch all neighbors
		cr.RoutingTable, cr.CrawlError = c.fetchNeighbors(ctx, pi)
		if cr.CrawlError != nil {
			cr.CrawlErrorStr = db.NetError(cr.CrawlError)
		}

		// Extract information from peer store
		ps := c.host.Peerstore()

		// Extract agent
		if agent, err := ps.Get(pi.ID, "AgentVersion"); err == nil {
			cr.Agent = agent.(string)
		}

		// Extract protocols
		if protocols, err := ps.GetProtocols(pi.ID); err == nil {
			cr.Protocols = protocols
		}
	}

	if cr.ConnectError != nil {
		cr.ConnectErrorStr = db.NetError(cr.ConnectError)
	}

	// Free connection resources
	if err := c.host.Network().ClosePeer(pi.ID); err != nil {
		log.WithError(err).WithField("remoteID", utils.FmtPeerID(pi.ID)).Warnln("Could not close connection to peer")
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	// Save the end time of this crawl
	cr.CrawlEndTime = time.Now()

	return cr
}

// connect establishes a connection to the given peer. It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) error {
	metrics.VisitCount.With(metrics.CrawlLabel).Inc()

	if len(pi.Addrs) == 0 {
		metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
		return fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
	defer cancel()

	if err := c.host.Connect(ctx, pi); err != nil {
		metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
		return err
	}

	return nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL).
func (c *Crawler) fetchNeighbors(ctx context.Context, pi peer.AddrInfo) (*RoutingTable, error) {
	rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(pi.ID), time.Hour, nil, time.Hour, nil)
	if err != nil {
		return nil, err
	}

	allNeighborsLk := sync.RWMutex{}
	allNeighbors := map[peer.ID]peer.AddrInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	errg := errgroup.Group{}
	for i := uint(0); i <= 15; i++ { // 15 is maximum
		count := i // Copy value
		errg.Go(func() error {
			// Generate a peer with the given common prefix length
			rpi, err := rt.GenRandPeerID(count)
			if err != nil {
				errorBits.Add(1 << count)
				return errors.Wrapf(err, "generating random peer ID with CPL %d", count)
			}

			var neighbors []*peer.AddrInfo
			for retry := 0; retry < 2; retry++ {
				neighbors, err = c.pm.GetClosestPeers(ctx, pi.ID, rpi)
				if err == nil {
					break
				}

				if utils.IsResourceLimitExceeded(err) {
					// other node has indicated that it's out of resources. Wait a bit and try again.
					time.Sleep(time.Second * time.Duration(5*(retry+1))) // may add jitter here
					continue
				}

				errorBits.Add(1 << count)
				return errors.Wrapf(err, "getting closest peer with CPL %d", count)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()
			for _, n := range neighbors {
				allNeighbors[n.ID] = *n
			}
			return nil
		})
	}
	err = errg.Wait()
	metrics.FetchedNeighborsCount.Observe(float64(len(allNeighbors)))

	routingTable := &RoutingTable{
		PeerID:    pi.ID,
		Neighbors: []peer.AddrInfo{},
		ErrorBits: uint16(errorBits.Load()),
		Error:     err,
	}

	for _, n := range allNeighbors {
		routingTable.Neighbors = append(routingTable.Neighbors, n)
	}

	return routingTable, err
}

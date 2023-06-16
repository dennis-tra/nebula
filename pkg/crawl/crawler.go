package crawl

import (
	"context"
	"fmt"
	"time"

	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/api"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

var crawlerID = atomic.NewInt32(0)

// Crawler encapsulates a libp2p host that crawls the network.
type Crawler struct {
	id           string
	host         *basichost.BasicHost
	config       *config.Crawl
	pm           *pb.ProtocolMessenger
	crawledPeers int
	client       *api.Client
	done         chan struct{}
}

// NewCrawler initializes a new crawler based on the given configuration.
func NewCrawler(h *basichost.BasicHost, conf *config.Crawl) (*Crawler, error) {
	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(conf.Protocols.Value()),
		timeout:   conf.Root.DialTimeout,
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
		client: api.NewClient(),
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
	}

	// start crawling both ways
	p2pResultCh := c.crawlP2P(ctx, cr.Peer)
	apiResultCh := c.crawlAPI(ctx, cr.Peer)

	p2pResult := <-p2pResultCh
	cr.CrawlEndTime = time.Now() // for legacy/consistency reasons we track the crawl end time here (without the API)
	apiResult := <-apiResultCh

	// merge both results
	cr.Merge(p2pResult, apiResult)

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr
}

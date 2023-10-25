package discv5

import (
	"context"
	"fmt"
	"time"

	"github.com/friendsofgo/errors"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	discover "github.com/dennis-tra/nebula-crawler/pkg/eth"
	log "github.com/sirupsen/logrus"
)

type CrawlerConfig struct {
	TrackNeighbors bool
	DialTimeout    time.Duration
	CheckExposed   bool
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	listener     *discover.UDPv5
	crawledPeers int
	done         chan struct{}
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   task.peerID.ShortString(),
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		RoutingTableFromAPI: false,
	}

	if len(task.maddrs) == 0 {
		log.WithField("peer", task.ID().String()).Warnln("no multi addresses")
	}
	cr.Agent = task.Node.Record().IdentityScheme()

	var firstResponseTimestamp *time.Time
	cr.ConnectStartTime = time.Now()
	cr.RoutingTable, firstResponseTimestamp, cr.CrawlError = c.fetchNeighbors(ctx, task)
	if firstResponseTimestamp == nil {
		cr.ConnectEndTime = time.Now()
		cr.ConnectError = cr.CrawlError
		metrics.VisitErrorsCount.With(metrics.CrawlLabel).Inc()
	} else {
		cr.ConnectEndTime = *firstResponseTimestamp
	}

	if cr.CrawlError != nil {
		cr.CrawlErrorStr = db.NetError(cr.CrawlError)
	}

	if cr.ConnectError != nil {
		cr.ConnectErrorStr = db.NetError(cr.ConnectError)
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	cr.CrawlEndTime = time.Now()
	return cr, nil
}

// fetchNeighbors sends RPC messages to the given peer and asks for its closest peers to an artificial set
// of 15 random peer IDs with increasing common prefix lengths (CPL).
func (c *Crawler) fetchNeighbors(ctx context.Context, pi PeerInfo) (*core.RoutingTable[PeerInfo], *time.Time, error) {
	// all neighbors of pi. We're using a map to not deduplicate.
	allNeighbors := map[string]PeerInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	// loop through the buckets sequentially because discv5 is also doing that
	// internally, so we won't gain much by spawning multiple parallel go
	// routines here. Stop the process as soon as we have received a timeout and
	// don't let the following calls time out as well.
	var err error
	timeouts := 0
	var firstResponseTimestamp *time.Time
	for i := 0; i <= discover.NBuckets; i++ { // 15 is maximum
		var neighbors []*enode.Node
		neighbors, err = c.listener.FindNode(pi.Node, []uint{uint(discover.HashBits - i)})
		if err != nil {
			errorBits.Add(1 << i)

			if errors.Is(err, discover.ErrTimeout) {
				timeouts += 1
				if timeouts < 2 {
					continue
				}
			}

			err = fmt.Errorf("getting closest peer with CPL %d: %w", i, err)
			break
		}

		if firstResponseTimestamp == nil {
			now := time.Now()
			firstResponseTimestamp = &now
		}

		for _, n := range neighbors {
			npi, err := NewPeerInfo(n)
			if err != nil {
				log.WithError(err).Warnln("Failed parsing ethereum node neighbor")
				continue
			}
			allNeighbors[string(npi.peerID)] = npi
		}
	}
	metrics.FetchedNeighborsCount.Observe(float64(len(allNeighbors)))

	routingTable := &core.RoutingTable[PeerInfo]{
		PeerID:    pi.ID(),
		Neighbors: []PeerInfo{},
		ErrorBits: uint16(errorBits.Load()),
		Error:     err,
	}

	for _, n := range allNeighbors {
		routingTable.Neighbors = append(routingTable.Neighbors, n)
	}

	return routingTable, firstResponseTimestamp, err
}

package discv5

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/db"

	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum/p2p/enode"

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
	allNeighborsLk := sync.RWMutex{}
	allNeighbors := map[string]PeerInfo{}
	var firstResponseTimestamp *time.Time

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	errg := errgroup.Group{}
	for i := 0; i <= discover.NBuckets; i++ { // 15 is maximum
		count := i // Copy value
		distance := uint(discover.HashBits - i)
		errg.Go(func() error {
			var (
				neighbors []*enode.Node
				err       error
			)

			for retry := 0; retry < 2; retry++ {
				neighbors, err = c.listener.FindNode(pi.Node, []uint{distance})
				if err == nil {
					break
				}
				// check error and potentially retry
				errorBits.Add(1 << count)
				return fmt.Errorf("getting closest peer with CPL %d: %w", count, err)
			}

			allNeighborsLk.Lock()
			defer allNeighborsLk.Unlock()

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
			return nil
		})
	}
	err := errg.Wait()
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

package mainline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	btdht "github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/int160"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

type CrawlerConfig struct {
	DialTimeout time.Duration
	LogErrors   bool
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	server       *btdht.Server
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

	errStr := ""

	crawlStart := time.Now()
	result, err := c.crawl(ctx, task)
	crawlEnd := time.Now()

	if err != nil {
		errStr = db.NetError(err)
	}

	connErr := err
	connErrStr := errStr

	if len(result.RoutingTable.Neighbors) != 0 {
		connErr = nil
		connErrStr = ""
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:        c.id,
		Info:             task,
		CrawlStartTime:   crawlStart,
		RoutingTable:     result.RoutingTable,
		Agent:            "",
		Protocols:        []string{},
		ConnectError:     connErr,
		ConnectErrorStr:  connErrStr,
		CrawlError:       connErr,
		CrawlErrorStr:    connErrStr,
		CrawlEndTime:     crawlEnd,
		ConnectStartTime: crawlStart,
		ConnectEndTime:   crawlEnd,
		Properties:       make(json.RawMessage, 0),
		LogErrors:        c.cfg.LogErrors,
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

type Result struct {
	RoutingTable *core.RoutingTable[PeerInfo]
}

func (c *Crawler) crawl(ctx context.Context, pi PeerInfo) (Result, error) {
	result := Result{
		RoutingTable: &core.RoutingTable[PeerInfo]{PeerID: pi.ID()},
	}
	var err error

	base := pi.NodeInfo.ID.Int160()

	allNeighborsLk := sync.RWMutex{}
	allNeighbors := map[peer.ID]PeerInfo{}

	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	errorBits := atomic.NewUint32(0)

	for i := 0; i <= 8; i++ {
		count := i // Copy value (we're still constrained to Go 1.19)

		mod := int160.FromBytes(base.Bytes())
		mod.SetBit(count, !base.GetBit(count))

		fmt.Println(count)
		qres := c.server.FindNode(addr{NodeAddr: pi.Addr}, mod, btdht.QueryRateLimiting{
			NotFirst:      true,
			NotAny:        true,
			WaitOnRetries: true,
			NoWaitFirst:   true,
		})
		fmt.Println(count, "done")

		if qres.Err != nil {
			errorBits.Add(1 << count)
			err = qres.Err
			break
		}

		tqres := qres.TraversalQueryResult(pi.Addr)
		time.Sleep(2 * time.Minute)

		allNeighborsLk.Lock()
		// for _, node := range append(tqres.Nodes, tqres.Nodes6...) {
		for _, node := range tqres.Nodes {
			neighbor, err := NewPeerInfo(node)
			if err != nil {
				log.WithError(err).Warnln("Could not build peer info")
				continue
			}

			allNeighbors[neighbor.ID()] = neighbor
		}
		allNeighborsLk.Unlock()

	}

	routingTable := &core.RoutingTable[PeerInfo]{
		PeerID:    pi.ID(),
		Neighbors: make([]PeerInfo, 0, len(allNeighbors)),
		ErrorBits: uint16(errorBits.Load()),
		Error:     err,
	}

	for _, n := range allNeighbors {
		routingTable.Neighbors = append(routingTable.Neighbors, n)
	}

	result.RoutingTable = routingTable

	return result, err
}

package core

import (
	"context"
	"fmt"
	"math"

	"github.com/dennis-tra/nebula-crawler/pkg/models"

	"github.com/dennis-tra/nebula-crawler/pkg/metrics"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/core/peer"
)

type EngineConfig struct {
	CrawlerCount   int
	WriterCount    int
	Limit          int
	TrackNeighbors bool
}

type Engine[I PeerInfo] struct {
	cfg   *EngineConfig
	stack Stack[I]

	crawlerPool *Pool[I, CrawlResult[I]]
	writerPool  *Pool[CrawlResult[I], WriteResult]

	crawlQueue map[string]I
	writeQueue map[string]CrawlResult[I]

	runData *RunData[I]

	inflight map[string]struct{}

	// A map from a peer's ID to further peer information. This map will
	// indicate if a peer has already been crawled in the past, so we don't put
	// them in the crawl queue again.
	crawled map[string]struct{}
}

// NewEngine initializes a new crawl engine.
func NewEngine[I PeerInfo](stack Stack[I], cfg *EngineConfig) (*Engine[I], error) {
	crawlers := make([]Worker[I, CrawlResult[I]], cfg.CrawlerCount)
	for i := 0; i < cfg.CrawlerCount; i++ {
		crawler, err := stack.NewCrawler()
		if err != nil {
			return nil, fmt.Errorf("new crawler worker: %w", err)
		}
		crawlers[i] = crawler
	}

	writers := make([]Worker[CrawlResult[I], WriteResult], cfg.WriterCount)
	for i := 0; i < cfg.WriterCount; i++ {
		writer, err := stack.NewWriter()
		if err != nil {
			return nil, fmt.Errorf("new writer worker: %w", err)
		}
		writers[i] = writer
	}

	return &Engine[I]{
		cfg:         cfg,
		stack:       stack,
		crawlerPool: NewPool[I, CrawlResult[I]](crawlers...),
		writerPool:  NewPool[CrawlResult[I], WriteResult](writers...),
		crawlQueue:  make(map[string]I),
		writeQueue:  make(map[string]CrawlResult[I]),
		runData: &RunData[I]{
			PeerMappings:  make(map[peer.ID]int),
			RoutingTables: make(map[peer.ID]*RoutingTable[I]),
			ConnErrs:      make(map[string]int),
		},
		inflight: make(map[string]struct{}),
		crawled:  make(map[string]struct{}),
	}, nil
}

func (e *Engine[I]) Run(ctx context.Context) (*RunData[I], error) {
	defer e.stack.OnClose()
	defer func() {
		e.runData.QueuedPeers = len(e.crawlQueue)
		e.runData.CrawledPeers = len(e.crawled)
	}()

	crawlTasks := make(chan I)
	writeTasks := make(chan CrawlResult[I])

	crawlResults := e.crawlerPool.Start(ctx, crawlTasks)
	writeResults := e.writerPool.Start(ctx, writeTasks)

	bps, err := e.stack.BootstrapPeers()
	if err != nil {
		return e.runData, fmt.Errorf("get bootstrap peers: %w", err)
	}

	for _, bp := range bps {
		e.crawlQueue[string(bp.ID())] = bp
	}

	for {
		// get a random peer to crawl and a random crawl result that we
		// should store in the database
		crawlTask, crawlOk := peekItem(e.crawlQueue)
		writeTask, writeOk := peekItem(e.writeQueue)

		// if we don't have any more peers to crawl and the crawlTasks channel
		// wasn't closed previously, we close it and set it to nil. Setting the
		// channel to nil prevents sending the crawlTask zero value to it.
		var innerCrawlTasks chan I
		if crawlOk {
			innerCrawlTasks = crawlTasks
		} else if crawlTasks != nil && len(e.inflight) == 0 {
			close(crawlTasks)
			crawlTasks = nil
		}

		var innerWriteTasks chan CrawlResult[I]
		if writeOk {
			innerWriteTasks = writeTasks
		} else if crawlTasks == nil && writeTasks != nil {
			close(writeTasks)
			writeTasks = nil
		}

		select {
		case <-ctx.Done():
			e.stack.OnClose()
			close(crawlTasks)
			close(writeTasks)
			<-crawlResults
			<-writeResults
			return e.runData, ctx.Err()
		case innerCrawlTasks <- crawlTask:
			delete(e.crawlQueue, string(crawlTask.ID()))
			e.inflight[string(crawlTask.ID())] = struct{}{}
		case innerWriteTasks <- writeTask:
			delete(e.writeQueue, string(writeTask.Info.ID()))
		case result, more := <-crawlResults:
			if !more {
				crawlResults = nil
				break
			}
			e.handleCrawlResult(result.Value, result.Error)
		case result, more := <-writeResults:
			if !more {
				writeResults = nil
				break
			}
			e.handleWriteResult(result.Value, result.Error)
		}

		if (crawlResults == nil && writeResults == nil) || e.reachedCrawlLimit() {
			break
		}
	}

	return e.runData, nil
}

func (e *Engine[I]) handleCrawlResult(cr CrawlResult[I], err error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  cr.CrawlerID,
		"remoteID":   cr.Info.ID().ShortString(),
		"isDialable": cr.ConnectError == nil && cr.CrawlError == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.CrawlerID)

	delete(e.inflight, string(cr.Info.ID()))

	// Keep track that this peer was crawled, so we don't do it again during this run
	e.crawled[string(cr.Info.ID())] = struct{}{}

	// Update prometheus metrics with the most up to date data
	metrics.DistinctVisitedPeersCount.Inc()
	metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(len(e.crawlQueue)))

	// Publish crawl result to persist queue so that the data is saved into the DB.
	e.writeQueue[string(cr.Info.ID())] = cr

	// Give the stack a chance to handle the result and keep track of network
	// specific metrics
	e.stack.OnPeerCrawled(cr, err)

	// Schedule crawls of all found neighbors unless we got the routing table from the API.
	// In this case the routing table information won't include any MultiAddresses. This means
	// we can't use these peers for further crawls.
	if !cr.RoutingTableFromAPI {
		for _, n := range cr.RoutingTable.Neighbors {
			// Don't add this peer to the queue if it's already in it
			if _, inCrawlQueue := e.crawlQueue[string(n.ID())]; inCrawlQueue {
				continue
			}

			// Don't add the peer to the queue if we have already visited it
			if _, crawled := e.crawled[string(n.ID())]; crawled {
				continue
			}

			// Schedule crawl for peer
			e.crawlQueue[string(n.ID())] = n
			e.runData.QueuedPeers += 1
		}

		// Track new peer in queue with prometheus
		metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(len(e.crawlQueue)))
	}

	if cr.ConnectError == nil {
		// Only track the neighbors if we were actually able to connect to the peer. Otherwise, we would track
		// an empty routing table of that peer. Only track the routing table in the neighbors table if at least
		// one FIND_NODE RPC succeeded.
		if e.cfg.TrackNeighbors && cr.RoutingTable.ErrorBits < math.MaxUint16 {
			e.runData.RoutingTables[cr.Info.ID()] = cr.RoutingTable
		}
	} else if cr.ConnectError != nil {
		// Log and count connection errors
		e.runData.ConnErrs[cr.ConnectErrorStr] += 1
		if cr.ConnectErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(cr.ConnectError)
		} else {
			logEntry = logEntry.WithField("dialErr", cr.ConnectErrorStr)
		}
	}

	if cr.CrawlError != nil {
		// Don't count this against the "dialability" errors - for now.
		// s.connErrs[cr.CrawlErrorStr] += 1

		// Log and count crawl errors
		if cr.CrawlErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(cr.CrawlError)
		} else {
			logEntry = logEntry.WithField("crawlErr", cr.CrawlErrorStr)
		}
	}

	logEntry.WithFields(map[string]interface{}{
		"inCrawlQueue": len(e.crawlQueue),
		"crawled":      len(e.crawled),
	}).Infoln("Handled crawl result from worker", cr.CrawlerID)
}

func (e *Engine[I]) handleWriteResult(cr WriteResult, err error) {
	if err != nil {
		return
	}

	if cr.PeerID != nil {
		e.runData.PeerMappings[cr.PID] = *cr.PeerID
	}
}

// reachedCrawlLimit returns true if the crawl limit is configured (aka != 0)
// and the crawled peers exceed this limit.
func (e *Engine[I]) reachedCrawlLimit() bool {
	return e.cfg.Limit > 0 && len(e.crawled) >= e.cfg.Limit
}

type RunData[I PeerInfo] struct {
	// A map that maps peer IDs to their database IDs. This speeds up the insertion of neighbor information as
	// the database does not need to look up every peer ID but only the ones not yet present in the database.
	// Speed up for ~11k peers: 5.5 min -> 30s
	PeerMappings map[peer.ID]int

	// A map that keeps track of all k-bucket entries of a particular peer.
	RoutingTables map[peer.ID]*RoutingTable[I]

	// A map of errors that happened during the crawl.
	ConnErrs map[string]int

	// The number of peers that are still queued to crawl after the Run method returned.
	QueuedPeers int

	// The number of peers that were crawled.
	CrawledPeers int
}

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (d *RunData[I]) TotalErrors() int {
	sum := 0
	for _, count := range d.ConnErrs {
		sum += count
	}
	return sum
}

func peekItem[K comparable, V any](queue map[K]V) (V, bool) {
	for _, item := range queue {
		return item, true
	}
	return *new(V), false
}

package core

import (
	"context"
	"fmt"
	"math"

	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// The EngineConfig object configures the core Nebula [Engine] below.
type EngineConfig struct {
	// the number of internal crawlers. This translates to how many peers to crawl in parallel.
	CrawlerCount int

	// the number of internal writers that store the crawl results to disk.
	WriterCount int

	// maximum number of peers to crawl before stopping the engine.
	Limit int

	// a flag that indicates whether we want to track and keep routing table
	// configurations of all peers in memory and write them to disk after the
	// crawl has finished.
	TrackNeighbors bool
}

// Engine is the integral data structure for orchestrating the communication
// with peers and writing the crawl results to disk. It maintains a worker pool
// of crawlers and writers that are concurrently crawling peers and writing
// the results to disk. The engine is responsible for scheduling which peer
// to crawl next, making sure to not crawl the same peer twice. At the same time
// it buffers the crawl results and schedules crawl results to be stored and
// distributes these jobs to the writers when they have capacity. The engine
// can be configured with the [EngineConfig] struct.
type Engine[I PeerInfo] struct {
	// this engine's configuration
	cfg *EngineConfig

	// the networking stack that this engine should operate with
	stack Stack[I]

	// worker pools for internal crawlers/writers.
	crawlerPool *Pool[I, CrawlResult[I]]
	writerPool  *Pool[CrawlResult[I], WriteResult]

	// queues of jobs that need to be performed. Either peers to crawl or
	// crawl results to write to disk.
	crawlQueue map[string]I
	writeQueue map[string]CrawlResult[I]

	// data structure that captures aggregate information about the engine's run
	runData *RunData[I]

	// a map that keeps track of all peers we are currently communicating with.
	inflight map[string]struct{}

	// A map from a peer's ID to further peer information. This map will
	// indicate if a peer has already been crawled in the past, so we don't put
	// them in the crawl queue again.
	crawled map[string]struct{}
}

// NewEngine initializes a new crawl engine. See the [Engine] documentation for
// more information.
func NewEngine[I PeerInfo](stack Stack[I], cfg *EngineConfig) (*Engine[I], error) {
	// initialize the configured number of crawl workers.
	crawlers := make([]Worker[I, CrawlResult[I]], cfg.CrawlerCount)
	for i := 0; i < cfg.CrawlerCount; i++ {
		crawler, err := stack.NewCrawler()
		if err != nil {
			return nil, fmt.Errorf("new crawler worker: %w", err)
		}
		crawlers[i] = crawler
	}

	// initialize the configured amount of writers that write crawl results to disk
	writers := make([]Worker[CrawlResult[I], WriteResult], cfg.WriterCount)
	for i := 0; i < cfg.WriterCount; i++ {
		writer, err := stack.NewWriter()
		if err != nil {
			return nil, fmt.Errorf("new writer worker: %w", err)
		}
		writers[i] = writer
	}

	// initialize empty maps for the different queues etc.
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

// Run is a blocking call that starts the crawler and writer worker pools to
// accept and perform tasks. It then starts by asking the network stack for a
// set of bootstrap nodes to start the crawl from. Then it sends these bootstrap
// peers to the crawl workers which then perform the act of crawling that peer
// and sending back the result which is then sent to one of the writer workers
// which in turn stores the crawl result. The engine, in the meantime, keeps
// track and exposes prometheus metrics, as well as gathers aggregate
// information about the entire crawl itself in [RunData].
func (e *Engine[I]) Run(ctx context.Context) (*RunData[I], error) {
	defer e.stack.OnClose()
	defer func() {
		e.runData.QueuedPeers = len(e.crawlQueue)
		e.runData.CrawledPeers = len(e.crawled)
	}()

	// initialize the task queues that the crawler and writer will read from
	crawlTasks := make(chan I)
	writeTasks := make(chan CrawlResult[I])

	// start the crawler and writer worker pools to read from their respective
	// task channel. The returned result channels are used to signal back any
	// task completion from any worker of the respective pool.
	crawlResults := e.crawlerPool.Start(ctx, crawlTasks)
	writeResults := e.writerPool.Start(ctx, writeTasks)

	// ask the networking stack for bootstrap peers to start the crawl from.
	bps, err := e.stack.BootstrapPeers()
	if err != nil {
		return e.runData, fmt.Errorf("get bootstrap peers: %w", err)
	}

	// put these bootstrap peers in the crawl queue, so they'll be properly scheduled.
	for _, bp := range bps {
		e.crawlQueue[string(bp.ID())] = bp
	}

	// start the core loop
	for {
		// get a random peer to crawl and a random crawl result that we
		// should store in the database
		crawlTask, crawlOk := peekItem(e.crawlQueue)
		writeTask, writeOk := peekItem(e.writeQueue)

		// if we still have crawl tasks to do, set the inner queue to the
		// original one, so that we'll try to send on that channel in the below
		// select statement. Otherwise, if the crawlTasks queue wasn't already
		// closed in the past and we don't have any requests still inflight,
		// close the channel and invalidate the variable -> we're done crawling.
		var innerCrawlTasks chan I
		if crawlOk {
			innerCrawlTasks = crawlTasks
		} else if crawlTasks != nil && len(e.inflight) == 0 {
			close(crawlTasks)
			crawlTasks = nil
		}

		// if we still have write tasks to do, set the inner queue to the
		// original one, sot that we'll try to send on that channel in the below
		// select statement. Otherwise, if the crawlTasks queue was invalidated
		// (we don't have anything to crawl) and the writeTasks queue was not
		// invalidated yet, do exactly that -> we're done writing results to
		// disk, and there isn't anything left to do.
		var innerWriteTasks chan CrawlResult[I]
		if writeOk {
			innerWriteTasks = writeTasks
		} else if crawlTasks == nil && writeTasks != nil {
			close(writeTasks)
			writeTasks = nil
		}

		select {
		case <-ctx.Done():
			// the engine was asked to stop. Clean up resources.
			e.stack.OnClose()
			close(crawlTasks)
			close(writeTasks)
			<-crawlResults
			<-writeResults
			return e.runData, ctx.Err()
		case innerCrawlTasks <- crawlTask:
			// a crawl worker was ready to accept a new task -> perform internal bookkeeping.
			delete(e.crawlQueue, string(crawlTask.ID()))
			e.inflight[string(crawlTask.ID())] = struct{}{}
		case innerWriteTasks <- writeTask:
			// a write worker was ready to accept a new task -> perform internal bookkeeping.
			delete(e.writeQueue, string(writeTask.Info.ID()))
		case result, more := <-crawlResults:
			if !more {
				// the crawlResults queue was closed. This means all crawl workers
				// have exited their go routine and scheduling new tasks would block
				// indefinitely because no one would read from the channel.
				// To avoid a hot spinning loop we set the channel to nil which
				// will keep the select statement to block. If crawlResults and
				// writeResults are nil, no work will be performed, and we can
				// exit this for loop. This is checked below.
				crawlResults = nil
				break
			}

			// a crawl worker finished a crawl task by contacting a peer. Handle it.
			e.handleCrawlResult(result.Value, result.Error)
		case result, more := <-writeResults:
			if !more {
				// the writeResults queue was closed. This means all write workers
				// have exited their go routine and scheduling new tasks would block
				// indefinitely because no one would read from the channel.
				// To avoid a hot spinning loop we set the channel to nil which
				// will keep the select statement to block. If writeResults and
				// crawlResults are nil, no work will be performed, and we can
				// exit this for loop. This is checked below.
				writeResults = nil
				break
			}

			// a write worker finished writing data to disk. Handle this event.
			e.handleWriteResult(result.Value, result.Error)
		}

		// break the for loop after 1) all workers have stopped or 2) we have
		// reached the configured maximum amount of peers we wanted to crawl.
		if (crawlResults == nil && writeResults == nil) || e.reachedCrawlLimit() {
			break
		}
	}

	// return aggregate run information. Check the above defer statement for
	// how other fields are populated.
	return e.runData, nil
}

// handleCrawlResult performs internal bookkeeping after a worker from the
// crawler pool has published a crawl result. Here, we update several internal
// bookkeeping maps and prometheus metrics as well as scheduling new peers to
// crawl.
func (e *Engine[I]) handleCrawlResult(cr CrawlResult[I], err error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  cr.CrawlerID,
		"remoteID":   cr.Info.ID().ShortString(),
		"isDialable": cr.ConnectError == nil && cr.CrawlError == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.CrawlerID)

	// This crawl operation for this peer is not inflight anymore -> delete it.
	delete(e.inflight, string(cr.Info.ID()))

	// Keep track that this peer was crawled, so we don't do it again during this run
	e.crawled[string(cr.Info.ID())] = struct{}{}

	// Update prometheus metrics with the most up to date data
	metrics.DistinctVisitedPeersCount.Inc()
	metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(len(e.crawlQueue)))

	// Publish crawl result to persist queue so that the data is saved into the DB.
	e.writeQueue[string(cr.Info.ID())] = cr

	// Give the stack a chance to handle the result and keep track of
	// network-specific metrics
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
	}).Infoln("Handled crawl result from", cr.CrawlerID)
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

// peekItem is a helper function that returns an arbitrary element from a map.
// It returns true if the map still contained at least one element or false if
// the map is empty.
func peekItem[K comparable, V any](queue map[K]V) (V, bool) {
	for _, item := range queue {
		return item, true
	}
	return *new(V), false
}

package core

import (
	"context"
	"fmt"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"

	ma "github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

// The EngineConfig object configures the core Nebula [Engine] below.
type EngineConfig struct {
	// the number of internal workers. This translates to how many peers we
	// process in parallel.
	WorkerCount int

	// the number of internal writers that store the results to disk.
	WriterCount int

	// maximum number of peers to crawl before stopping the engine. 0 means
	// to process peers until there are no more in the work queue.
	Limit int

	// if set to true, the engine won't keep track of which peers were already
	// processed to prevent processing a peer twice. The engine is solely driven
	// by what the driver will emit on its tasks channel.
	DuplicateProcessing bool

	// which type addresses should be dialed. Relevant for parking
	// peers during a crawl
	AddrDialType config.AddrType
}

// DefaultEngineConfig returns a default engine configuration that can and
// should be adjusted for different networks.
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		WorkerCount:         100,
		WriterCount:         10,
		Limit:               0,
		DuplicateProcessing: false,
		AddrDialType:        config.AddrTypeAny,
	}
}

// Validate verifies the engine configuration's invariants.
func (cfg *EngineConfig) Validate() error {
	if cfg.WorkerCount <= 0 {
		return fmt.Errorf("worker count must not be zero or negative")
	}

	if cfg.WriterCount <= 0 {
		return fmt.Errorf("writer count must not be zero or negative")
	}

	return nil
}

// Engine is the integral data structure for orchestrating the communication
// with peers and writing the processing results to disk. It maintains a pool
// of workers and writers that are concurrently processing peers and writing
// the results to disk. The engine is responsible for scheduling which peer
// to process next, making sure to not process the same peer twice. At the same
// time, it buffers the processing results and schedules results to be stored
// and distributes these tasks to the writers when they have capacity. The engine
// can be configured with the [EngineConfig] struct.
type Engine[I PeerInfo[I], R WorkResult[I]] struct {
	// this engine's configuration
	cfg *EngineConfig

	// the engine driver that provides worker implementations and peers to process
	driver Driver[I, R]

	// the channel on which the driver will emit peers to process
	tasksChan <-chan I

	// the data structure that drives the engine by handling peer processing
	// and write results. It returns new tasks to do.
	handler Handler[I, R]

	// pools for internal workers and writers. Workers process peers by crawling
	// or dialing them, and writers take the processing result and write them to
	// disk
	workerPool *Pool[I, R]
	writerPool *Pool[R, WriteResult]

	// queues of jobs that need to be performed. Either peers to process or
	// processing results to write to disk.
	peerQueue  *PriorityQueue[I]
	writeQueue *PriorityQueue[R]

	// a map that keeps track of all peers we are currently communicating with.
	inflight map[string]struct{}

	// A set of peer IDs that indicates which peers have already been processed
	// in the past, so we don't put them in the peer queue again.
	processed map[string]struct{}

	// a counter that tracks the number of handled write results
	writeCount int

	// a filter function that removes multi addresses from the given slice
	maddrFilter func([]ma.Multiaddr) []ma.Multiaddr
}

// NewEngine initializes a new engine. See the [Engine] documentation for
// more information.
func NewEngine[I PeerInfo[I], R WorkResult[I]](driver Driver[I, R], handler Handler[I, R], cfg *EngineConfig) (*Engine[I, R], error) {
	if cfg == nil {
		cfg = DefaultEngineConfig()
	} else if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// initialize the configured number of peer processing workers.
	workers := make([]Worker[I, R], cfg.WorkerCount)
	for i := 0; i < cfg.WorkerCount; i++ {
		worker, err := driver.NewWorker()
		if err != nil {
			return nil, fmt.Errorf("new peer worker: %w", err)
		}
		workers[i] = worker
	}

	// initialize the configured number of writers that write peer processing
	// results to disk
	writers := make([]Worker[R, WriteResult], cfg.WriterCount)
	for i := 0; i < cfg.WriterCount; i++ {
		writer, err := driver.NewWriter()
		if err != nil {
			return nil, fmt.Errorf("new writer worker: %w", err)
		}
		writers[i] = writer
	}

	var maddrFilter func([]ma.Multiaddr) []ma.Multiaddr = func(maddrs []ma.Multiaddr) []ma.Multiaddr { return maddrs }
	switch cfg.AddrDialType {
	case config.AddrTypePrivate:
		maddrFilter = utils.FilterPublicMaddrs
	case config.AddrTypePublic:
		maddrFilter = utils.FilterPrivateMaddrs
	}

	// initialize empty maps for the different queues etc.
	return &Engine[I, R]{
		cfg:         cfg,
		driver:      driver,
		tasksChan:   driver.Tasks(),
		handler:     handler,
		workerPool:  NewPool[I, R](workers...),
		writerPool:  NewPool[R, WriteResult](writers...),
		maddrFilter: maddrFilter,
		peerQueue:   NewPriorityQueue[I](),
		writeQueue:  NewPriorityQueue[R](),
		inflight:    make(map[string]struct{}),
		processed:   make(map[string]struct{}),
	}, nil
}

// Run is a blocking call that starts the worker and writer pools to accept and
// perform tasks. It enters an indefinite loop expecting to receive tasks from
// the driver. In the case of a crawl operation, these should be the bootstrap
// peers start the crawl from. Then it sends these peers to the workers which
// then process that peer and send back the result which is then sent to one of
// the writer workers which in turn stores the result. The engine, in the
// meantime, keeps track of and exposes prometheus metrics. The engine will
// keep running as long as the tasksChan from the driver isn't closed. If the
// channel was closed, the engine will process all remaining peers in the queue.
// Each result is passed to a handler that may return additional peers to
// process.
func (e *Engine[I, R]) Run(ctx context.Context) (map[string]I, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// initialize the task queues that the workers and writers will read from
	peerTasks := make(chan I)
	writeTasks := make(chan R)

	// start the worker and writer worker pools to read from their respective
	// task channel. The returned result channels are used to signal back any
	// task completion from any worker of the respective pool.
	peerResults := e.workerPool.Start(ctx, peerTasks)
	writerResults := e.writerPool.Start(ctx, writeTasks)

	// start the core loop
	for {
		// get a random peer to process and a random processing result that we
		// should store in the database
		peerTask, peerOk := e.peerQueue.Peek()
		writeTask, writeOk := e.writeQueue.Peek()

		// if we still have peers to process, set the inner queue to the
		// original one, so that we'll try to send on that channel in the below
		// select statement. Otherwise, if the peerTasks queue wasn't already
		// closed in the past, and we don't have any requests still inflight,
		// and we don't anticipate new tasks on the tasksChan from the driver,
		// close the channel and invalidate the variable -> we're done
		// processing peers.
		var innerPeerTasks chan I
		if peerOk {
			innerPeerTasks = peerTasks
		} else if peerTasks != nil && len(e.inflight) == 0 && e.tasksChan == nil {
			close(peerTasks)
			peerTasks = nil
		}

		// if we still have write tasks to do, set the inner queue to the
		// original one, sot that we'll try to send on that channel in the below
		// select statement. Otherwise, if the peerTasks queue was invalidated
		// (we don't have anything to process) and the writeTasks queue was not
		// invalidated yet, do exactly that -> we're done writing results to
		// disk, and there isn't anything left to do.
		var innerWriteTasks chan R
		if writeOk {
			innerWriteTasks = writeTasks
		} else if peerTasks == nil && writeTasks != nil {
			close(writeTasks)
			writeTasks = nil
		}

		select {
		case task, more := <-e.tasksChan:
			if !more {
				e.tasksChan = nil
				break
			}
			if _, found := e.inflight[string(task.ID())]; found {
				break
			}
			e.peerQueue.Push(string(task.ID()), task, 1)
		case innerPeerTasks <- peerTask:
			// a worker was ready to accept a new task -> perform internal bookkeeping.
			e.peerQueue.Drop(string(peerTask.ID()))
			e.inflight[string(peerTask.ID())] = struct{}{}
		case innerWriteTasks <- writeTask:
			// a write worker was ready to accept a new task -> perform internal bookkeeping.
			e.writeQueue.Drop(string(writeTask.PeerInfo().ID()))
		case result, more := <-peerResults:
			if !more {
				// the peerResults queue was closed. This means all workers
				// have exited their go routine and scheduling new tasks would block
				// indefinitely because no one would read from the channel.
				// To avoid a hot spinning loop we set the channel to nil which
				// will keep the select statement to block. If peerResults and
				// writerResults are nil, no work will be performed, and we can
				// exit this for loop. This is checked below.
				peerResults = nil
				break
			}

			// a worker finished a task by processing a peer. Handle it.
			e.handlePeerResult(result)
		case result, more := <-writerResults:
			if !more {
				// the writerResults queue was closed. This means all write workers
				// have exited their go routine and scheduling new tasks would block
				// indefinitely because no one would read from the channel.
				// To avoid a hot spinning loop we set the channel to nil which
				// will keep the select statement to block. If writerResults and
				// peerResults are nil, no work will be performed, and we can
				// exit this for loop. This is checked below.
				writerResults = nil
				break
			}

			// a write worker finished writing data to disk. Handle this event.
			e.handleWriteResult(result)
		case <-ctx.Done():
			// the engine was asked to stop. Clean up resources.
			log.Infoln("Closing driver...")
			e.driver.Close()
			if peerTasks != nil {
				close(peerTasks)
			}
			if writeTasks != nil {
				close(writeTasks)
			}

			// drain results channels. They'll be closed after all workers have
			// stopped working
			for range peerResults {
				// drop result
			}
			for range writerResults {
				// drop result
			}

			return e.peerQueue.All(), ctx.Err()
		}

		if peerResults == nil && writerResults == nil {
			log.Infoln("Closing driver...")
			e.driver.Close()
			return e.peerQueue.All(), nil // no work to do, natural end
		}

		// break the for loop after 1) all workers have stopped or 2) we have
		// reached the configured maximum amount of peers we wanted to process.
		if (peerResults == nil && writerResults == nil) || e.reachedProcessingLimit() {
			cancel()
		}
	}
}

// handlePeerResult performs internal bookkeeping after a worker from the pool
// has published a worker result. Here, we update several internal bookkeeping
// maps and prometheus metrics as well as scheduling new peers to process.
func (e *Engine[I, R]) handlePeerResult(result Result[R]) {
	wr := result.Value
	logEntry := wr.LogEntry()
	logEntry.Debugln("Handling worker result")

	// The operation for this peer is not inflight anymore -> delete it.
	delete(e.inflight, string(wr.PeerInfo().ID()))

	// Keep track that this peer was processed, so we don't do it again during
	// this run. Unless we explicitly allow duplicate processing.
	if !e.cfg.DuplicateProcessing {
		e.processed[string(wr.PeerInfo().ID())] = struct{}{}
		logEntry = logEntry.WithField("processed", len(e.processed))
		metrics.DistinctVisitedPeersCount.Inc()
	}

	// Publish the processing result to the writer queue so that the data is
	// saved to disk.
	e.writeQueue.Push(string(wr.PeerInfo().ID()), wr, 0)

	// let the handler work on the new peer result
	newTasks := e.handler.HandlePeerResult(result)

	// process the new tasks that came out of handling the peer result
	for _, task := range newTasks {
		mapKey := string(task.ID())

		// Don't add this peer to the queue if we're currently querying it
		if _, isInflight := e.inflight[mapKey]; isInflight {
			continue
		}

		// Don't add the peer to the queue if we have already processed it
		if _, processed := e.processed[mapKey]; processed {
			continue
		}

		// Check if we have already queued this peer. If so, merge the new
		// information with the already existing ones.
		queuedTask, isQueued := e.peerQueue.Find(mapKey)
		if isQueued {
			task = task.Merge(queuedTask)
		}

		// If we don't know any multi addresses for the peer yet, we push it
		// to the end of our priority queue by giving it a low priority.
		priority := 1
		if len(e.maddrFilter(task.Addrs())) == 0 {
			priority = 0
		}

		// If the peer was already queued we only update its priority. If the
		// peer wasn't queued, we push it to the queue.
		if isQueued {
			e.peerQueue.Update(mapKey, task, priority)
		} else {
			e.peerQueue.Push(string(task.ID()), task, priority)
		}
	}

	// Track new peer in queue with prometheus
	metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(e.peerQueue.Len()))

	logEntry.WithFields(map[string]interface{}{
		"queued":   e.peerQueue.Len(),
		"inflight": len(e.inflight),
	}).Infoln("Handled worker result")
}

func (e *Engine[I, R]) handleWriteResult(result Result[WriteResult]) {
	e.handler.HandleWriteResult(result)
	e.writeCount += 1

	if result.Value.Duration == 0 {
		return
	}

	log.WithFields(log.Fields{
		"writerID": result.Value.WriterID,
		"remoteID": result.Value.PeerID.ShortString(),
		"success":  result.Value.Error == nil,
		"written":  e.writeCount,
		"duration": result.Value.Duration,
	}).Infoln("Handled writer result")
}

// reachedProcessingLimit returns true if the processing limit is configured
// (aka != 0) and the processed peers exceed this limit.
func (e *Engine[I, R]) reachedProcessingLimit() bool {
	return e.cfg.Limit > 0 && len(e.processed) >= e.cfg.Limit
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

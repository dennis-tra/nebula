package crawl

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var ProtocolStrings = []protocol.ID{
	"/ipfs/kad/1.0.0",
	"/ipfs/kad/2.0.0",
}

// The Orchestrator handles the scheduling and managing of worker
// that crawl the network. I'm still not quite happy with that name...
type Orchestrator struct {
	// Service represents an entity that runs in a
	// separate go routine and where its lifecycle
	// needs to be handled externally.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// Connection to database
	// db *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inCrawlQueue sync.Map

	// The number of peers in the crawl queue.
	inCrawlQueueCount atomic.Uint32

	// A map from peer.ID to peer.AddrInfo to indicate if a peer has already been crawled
	// in the past, so we don't put in the crawl queue again.
	crawled sync.Map

	// The number of crawled peers.
	crawledCount atomic.Uint32

	// The queue of peer.AddrInfo's that still need to be crawled.
	crawlQueue chan peer.AddrInfo

	// The queue that the workers publish their crawl results on, so that the
	// orchestrator can handle them, e.g. update the maps above etc.
	resultsQueue chan CrawlResult

	// The list of worker node references.
	workers []*Worker

	// A map of errors that happened during the crawl.
	Errors sync.Map
}

func NewOrchestrator(ctx context.Context) (*Orchestrator, error) {
	// Initialize a single libp2p node that's shared between all workers.
	// TODO: experiment with multiple nodes.
	priv, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048) // TODO: is this really necessary? see "weak keys" handling in weizenbaum crawler.
	opts := []libp2p.Option{libp2p.Identity(priv), libp2p.NoListenAddrs}
	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	p := &Orchestrator{
		Service: service.New("orchestrator"),
		host:    h,
		// db:                dbc,
		config:            conf,
		inCrawlQueue:      sync.Map{},
		inCrawlQueueCount: atomic.Uint32{},
		crawled:           sync.Map{},
		crawledCount:      atomic.Uint32{},
		crawlQueue:        make(chan peer.AddrInfo),
		resultsQueue:      make(chan CrawlResult),
		workers:           []*Worker{},
	}
	return p, nil
}

// CrawlNetwork starts the configured amount of workers and fills
// the worker queue with bootstrap nodes to start with.
func (o *Orchestrator) CrawlNetwork(bootstrap []peer.AddrInfo) error {
	o.ServiceStarted()

	// Handle the results of crawls of a particular node in a separate go routine
	go o.handleCrawlResults()

	// Start all workers
	for i := 0; i < o.config.WorkerCount; i++ {
		w, err := NewWorker(o.host, o.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		o.workers = append(o.workers, w)
		go w.StartCrawling(o.crawlQueue, o.resultsQueue)
	}

	// Fill the queue with bootstrap nodes
	for _, b := range bootstrap {
		o.dispatchCrawl(b)
	}

	// Block until the orchestrator shuts down.
	<-o.SigShutdown()

	o.Cleanup()
	o.ServiceStopped()
	return nil
}

// Cleanup handles the release of all resources allocated by the orchestrator.
func (o *Orchestrator) Cleanup() {
	// drain crawl queue
OUTER:
	for {
		select {
		case pi := <-o.crawlQueue:
			log.WithField("targetID", pi.ID.Pretty()[:16]).Debugln("Drained peer")
		default:
			break OUTER
		}
	}
	close(o.crawlQueue)

	var wg sync.WaitGroup
	for _, w := range o.workers {
		wg.Add(1)
		go func(w *Worker) {
			w.Shutdown()
			wg.Done()
		}(w)
	}
	wg.Wait()

	// After all workers have stopped and won't send any results we can close the results channel.
	close(o.resultsQueue)
}

func (o *Orchestrator) handleCrawlResults() {
	for result := range o.resultsQueue {
		o.handleCrawlResult(result)
	}
}

func (o *Orchestrator) handleCrawlResult(cr CrawlResult) {
	defer func() {
		log.WithFields(map[string]interface{}{
			"inCrawlQueue": o.inCrawlQueueCount.Load(),
			"crawled":      o.crawledCount.Load(),
		}).Infof("Handled crawl result")

		if o.inCrawlQueueCount.Load() == 0 {
			// o.db can be null if it's a dry run
			//if o.db != nil {
			//	c := &db.Crawl{
			//		StartedAt:    o.StartTime,
			//		FinishedAt:   time.Now(),
			//		CrawledPeers: uint(o.crawledCount.Load()),
			//	}
			//
			//	o.db.Create(c)
			//}
			o.Shutdown()
		}
	}()

	//b := &backoff.Backoff{
	//	// These are the defaults
	//	Min:    30 * time.Second,
	//	Max:    5 * time.Minute,
	//	Factor: 2,
	//	Jitter: true,
	//}

	// o.db can be null if it's a dry run
	//if o.db != nil {
	//p := &db.Peer{
	//	ID:        cr.Peer.ID.Pretty(),
	//	FirstDial: time.Now(),
	//	LastDial:  time.Now(),
	//	NextDial:  time.Now().Add(b.ForAttempt(0)),
	//	Dials:     0,
	//}
	//
	//o.db.Create(p)
	//}

	ctx := o.ServiceContext()

	o.crawled.Store(cr.Peer.ID, true)
	o.crawledCount.Inc()
	o.inCrawlQueue.Delete(cr.Peer.ID)
	stats.Record(o.ServiceContext(), metrics.PeersToCrawlCount.M(float64(o.inCrawlQueueCount.Dec())))

	if cr.Agent != "" {
		ctx, _ = tag.New(o.ServiceContext(), metrics.UpsertAgentVersion(cr.Agent))
	}
	stats.Record(ctx, metrics.CrawledPeersCount.M(1))

	if cr.Error != nil {
		o.Errors.Store(cr.Error.Error(), cr.Error.Error())
		log.WithFields(log.Fields{
			"workerID": cr.WorkerID,
			"targetID": cr.Peer.ID.Pretty()[:16],
		}).WithError(cr.Error).Warnln("Error crawling peer")
		return
	}

	for _, pi := range cr.Neighbors {
		_, inCrawlQueue := o.inCrawlQueue.Load(pi.ID)
		_, crawled := o.crawled.Load(pi.ID)
		if !inCrawlQueue && !crawled {
			o.dispatchCrawl(pi)
		}
	}
}

func (o *Orchestrator) dispatchCrawl(pi peer.AddrInfo) {
	select {
	case <-o.SigShutdown():
		log.Debugln("Skipping dispatch as orchestrator shuts down")
		return
	default:
		o.inCrawlQueue.Store(pi.ID, true)
		stats.Record(o.ServiceContext(), metrics.PeersToCrawlCount.M(float64(o.inCrawlQueueCount.Inc())))
		go func() { o.crawlQueue <- pi }()
	}
}

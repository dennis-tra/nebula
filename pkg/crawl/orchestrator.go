package crawl

import (
	"context"
	"database/sql"
	"regexp"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

const agentVersionRegexPattern = `\/?go-ipfs\/(?P<core>\d+\.\d+\.\d+)-?(?P<prerelease>\w+)?\/(?P<commit>\w+)?`

var (
	agentVersionRegex = regexp.MustCompile(agentVersionRegexPattern)
	ProtocolStrings   = []protocol.ID{
		"/ipfs/kad/1.0.0",
	}
)

// The Orchestrator handles the scheduling and managing of worker
// that crawl the network. I'm still not quite happy with that name...
type Orchestrator struct {
	// Service represents an entity that runs in a
	// separate go routine and where its lifecycle
	// needs to be handled externally.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// The database handle
	dbh *sql.DB

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

	// The number of crawled peers.
	crawlErrors atomic.Uint32

	// The queue of peer.AddrInfo's that still need to be crawled.
	crawlQueue chan peer.AddrInfo

	// The queue that the workers publish their crawl results on, so that the
	// orchestrator can handle them, e.g. update the maps above etc.
	resultsQueue chan CrawlResult

	AgentVersion   map[string]int
	AgentVersionLk sync.RWMutex

	// The list of worker node references.
	workers []*Worker

	// A map of errors that happened during the crawl.
	Errors sync.Map
}

func NewOrchestrator(ctx context.Context, dbh *sql.DB) (*Orchestrator, error) {
	// Initialize a single libp2p node that's shared between all workers.
	// TODO: experiment with multiple nodes.
	// TODO: is the key pair really necessary? see "weak keys" handling in weizenbaum crawler.
	priv, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
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
		Service:           service.New("orchestrator"),
		host:              h,
		dbh:               dbh,
		config:            conf,
		inCrawlQueue:      sync.Map{},
		inCrawlQueueCount: atomic.Uint32{},
		crawled:           sync.Map{},
		crawledCount:      atomic.Uint32{},
		crawlQueue:        make(chan peer.AddrInfo),
		resultsQueue:      make(chan CrawlResult),
		AgentVersion:      map[string]int{},
		workers:           []*Worker{},
	}
	return p, nil
}

// CrawlNetwork starts the configured amount of workers and fills
// the worker queue with bootstrap nodes to start with.
func (o *Orchestrator) CrawlNetwork(bootstrap []peer.AddrInfo) error {
	o.ServiceStarted()

	// Handle the results of crawls of a particular node in a separate go routine
	// TODO: handle sync here and get rid of sync.Map/atomic.Int32?
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

func (o *Orchestrator) Shutdown() {
	defer o.Service.Shutdown()

	log.WithFields(log.Fields{
		"crawledPeers":    o.crawledCount.Load(),
		"crawlDuration":   time.Now().Sub(o.StartTime).String(),
		"dialablePeers":   o.crawledCount.Load() - o.crawlErrors.Load(),
		"undialablePeers": o.crawlErrors.Load(),
	}).Infoln("Successfully finished crawl")

	// The Database handler can be null if it's a dry run
	if o.dbh == nil {
		return
	}
	log.Infoln("Saving crawl results to database")

	crawl := &models.Crawl{
		StartedAt:       o.StartTime,
		FinishedAt:      time.Now(),
		CrawledPeers:    int(o.crawledCount.Load()),
		DialablePeers:   int(o.crawledCount.Load() - o.crawlErrors.Load()),
		UndialablePeers: int(o.crawlErrors.Load()),
	}

	err := crawl.Insert(o.ServiceContext(), o.dbh, boil.Infer())
	if err != nil {
		log.WithError(err).Warnln("Could not save crawl result")
		return
	}

	o.AgentVersionLk.Lock()
	defer o.AgentVersionLk.Unlock()

	ppfull := map[string]int{}
	ppcore := map[string]int{}
	for version, count := range o.AgentVersion {
		matches := agentVersionRegex.FindStringSubmatch(version)
		if matches == nil {
			ppfull[version] += count
		} else {
			ppcore[matches[1]] += count
		}
	}

	txn, err := o.dbh.BeginTx(o.ServiceContext(), nil)
	if err != nil {
		log.WithError(err).Warnln("Could not start txn")
		return
	}
	for version, count := range ppfull {
		pp := &models.PeerProperty{
			Property: "agent_version",
			Value:    version,
			Count:    count,
			CrawlID:  crawl.ID,
		}
		err = pp.Insert(o.ServiceContext(), txn, boil.Infer())
		if err != nil {
			log.WithError(err).Warnln("Could not insert peer property txn")
			continue
		}
	}
	for version, count := range ppcore {
		pp := &models.PeerProperty{
			Property: "agent_version_core",
			Value:    version,
			Count:    count,
			CrawlID:  crawl.ID,
		}
		err = pp.Insert(o.ServiceContext(), txn, boil.Infer())
		if err != nil {
			log.WithError(err).Warnln("Could not insert peer property txn")
			continue
		}
	}
	if err = txn.Commit(); err != nil {
		log.WithError(err).Warnln()
	}
	log.Infoln("Saved peer properties")
}

func (o *Orchestrator) handleCrawlResult(cr CrawlResult) {
	logEntry := log.WithFields(log.Fields{
		"workerID": cr.WorkerID,
		"targetID": cr.Peer.ID.Pretty()[:16],
	})
	logEntry.Debugln("Handling crawl result from worker", cr.WorkerID)

	// TODO: Another go routine?
	startUpsert := time.Now()
	err := o.upsertCrawlResult(cr.Peer.ID, cr.Peer.Addrs, cr.Error)
	if err != nil {
		log.WithError(err).Warnln("Could not update peer")
	}
	stats.Record(o.ServiceContext(), metrics.CrawledUpsertDuration.M(millisSince(startUpsert)))

	ctx := o.ServiceContext()
	o.crawled.Store(cr.Peer.ID, true)
	o.crawledCount.Inc()
	o.inCrawlQueue.Delete(cr.Peer.ID)
	stats.Record(o.ServiceContext(), metrics.PeersToCrawlCount.M(float64(o.inCrawlQueueCount.Dec())))

	o.AgentVersionLk.Lock()
	if cr.Agent != "" {
		o.AgentVersion[cr.Agent] += 1
	} else {
		o.AgentVersion["n.a."] += 1
	}
	o.AgentVersionLk.Unlock()

	stats.Record(ctx, metrics.CrawledPeersCount.M(1))

	if cr.Error != nil {
		o.Errors.Store(cr.Error.Error(), cr.Error.Error())
		o.crawlErrors.Inc()
		logEntry.WithError(cr.Error).Debugln("Error crawling peer")
	} else {
		for _, pi := range cr.Neighbors {
			_, inCrawlQueue := o.inCrawlQueue.Load(pi.ID)
			_, crawled := o.crawled.Load(pi.ID)
			if !inCrawlQueue && !crawled {
				o.dispatchCrawl(pi)
			}
		}
	}

	logEntry.WithFields(map[string]interface{}{
		"inCrawlQueue": o.inCrawlQueueCount.Load(),
		"crawled":      o.crawledCount.Load(),
	}).Infoln("Handled crawl result from worker", cr.WorkerID)

	if o.inCrawlQueueCount.Load() == 0 || (o.config.CrawlLimit > 0 && int(o.crawledCount.Load()) >= o.config.CrawlLimit) {
		o.Shutdown()
	}
}

// upsertCrawlResult inserts the given peer with its multi addresses in the database and
// upserts its currently active session
func (o *Orchestrator) upsertCrawlResult(peerID peer.ID, maddrs []ma.Multiaddr, dialErr error) error {
	// Check if we're in a dry-run
	if o.dbh == nil {
		return nil
	}

	if dialErr == nil {
		if err := db.UpsertPeer(o.ServiceContext(), o.dbh, peerID.Pretty(), maddrs); err != nil {
			return errors.Wrap(err, "upsert peer")
		}
		if err := db.UpsertSessionSuccess(o.dbh, peerID.Pretty()); err != nil {
			return errors.Wrap(err, "upsert session success")
		}
	} else if dialErr != o.ServiceContext().Err() {
		if err := db.UpsertSessionError(o.dbh, peerID.Pretty()); err != nil {
			return errors.Wrap(err, "upsert session error")
		}
	}

	return nil
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

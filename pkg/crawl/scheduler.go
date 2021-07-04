package crawl

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
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
		"/ipfs/kad/2.0.0",
	}
)

// The Scheduler handles the scheduling and managing of worker
// that crawl the network. I'm still not quite happy with that name...
type Scheduler struct {
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

	Protocols   map[string]int
	ProtocolsLk sync.RWMutex

	// The list of worker node references.
	workers []*Worker

	// A map of errors that happened during the crawl.
	Errors   map[string]int
	ErrorsLk sync.RWMutex
}

// knownErrors contains a list of known errors. Property key + string to match for
var knownErrors = map[string]string{
	"io_timeout":                 "i/o timeout",
	"connection_refused":         "connection refused",
	"protocol_not_supported":     "protocol not supported",
	"peer_id_mismatch":           "peer id mismatch",
	"no_route_to_host":           "no route to host",
	"network_unreachable":        "network is unreachable",
	"no_good_addresses":          "no good addresses",
	"context_deadline_exceeded":  "context deadline exceeded",
	"no_public_ip":               "no public IP address",
	"max_dial_attempts_exceeded": "max dial attempts exceeded",
}

func NewScheduler(ctx context.Context, dbh *sql.DB) (*Scheduler, error) {
	// Initialize a single libp2p node that's shared between all workers.
	// TODO: experiment with multiple nodes.
	// TODO: is the key pair really necessary? see "weak keys" handling in weizenbaum crawler.
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(ctx, libp2p.Identity(priv), libp2p.NoListenAddrs)
	if err != nil {
		return nil, err
	}

	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	p := &Scheduler{
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
		Protocols:         map[string]int{},
		Errors:            map[string]int{},
		workers:           []*Worker{},
	}
	return p, nil
}

// CrawlNetwork starts the configured amount of workers and fills
// the worker queue with bootstrap nodes to start with.
func (s *Scheduler) CrawlNetwork(bootstrap []peer.AddrInfo) error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	// Handle the results of crawls of a particular node in a separate go routine
	// TODO: handle sync here and get rid of sync.Map/atomic.Int32?
	go s.handleCrawlResults()

	// Start all workers
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		w, err := NewWorker(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.workers = append(s.workers, w)
		go w.StartCrawling(s.crawlQueue, s.resultsQueue)
	}

	// Fill the queue with bootstrap nodes
	for _, b := range bootstrap {
		s.dispatchCrawl(b)
	}

	// Block until the orchestrator shuts down.
	<-s.SigShutdown()

	s.Cleanup()
	return nil
}

// Cleanup handles the release of all resources allocated by the orchestrator.
func (s *Scheduler) Cleanup() {
	// drain crawl queue
OUTER:
	for {
		select {
		case pi := <-s.crawlQueue:
			log.WithField("targetID", pi.ID.Pretty()[:16]).Debugln("Drained peer")
		default:
			break OUTER
		}
	}
	close(s.crawlQueue)

	var wg sync.WaitGroup
	for _, w := range s.workers {
		wg.Add(1)
		go func(w *Worker) {
			w.Shutdown()
			wg.Done()
		}(w)
	}
	wg.Wait()

	// After all workers have stopped and won't send any results we can close the results channel.
	close(s.resultsQueue)
}

func (s *Scheduler) handleCrawlResults() {
	for result := range s.resultsQueue {
		s.handleCrawlResult(result)
	}
}

func (s *Scheduler) Shutdown() {
	defer s.Service.Shutdown()

	log.WithFields(log.Fields{
		"crawledPeers":    s.crawledCount.Load(),
		"crawlDuration":   time.Now().Sub(s.StartTime).String(),
		"dialablePeers":   s.crawledCount.Load() - s.crawlErrors.Load(),
		"undialablePeers": s.crawlErrors.Load(),
	}).Infoln("Successfully finished crawl")

	// The Database handler can be null if it's a dry run
	if s.dbh == nil {
		return
	}
	log.Infoln("Saving crawl results to database")

	crawl := &models.Crawl{
		StartedAt:       s.StartTime,
		FinishedAt:      time.Now(),
		CrawledPeers:    int(s.crawledCount.Load()),
		DialablePeers:   int(s.crawledCount.Load() - s.crawlErrors.Load()),
		UndialablePeers: int(s.crawlErrors.Load()),
	}

	err := crawl.Insert(s.ServiceContext(), s.dbh, boil.Infer())
	if err != nil {
		log.WithError(err).Warnln("Could not save crawl result")
		return
	}

	s.AgentVersionLk.Lock()
	s.ProtocolsLk.Lock()
	s.ErrorsLk.Lock()
	defer s.AgentVersionLk.Unlock()
	defer s.ProtocolsLk.Unlock()
	defer s.ErrorsLk.Unlock()

	ppfull := map[string]int{}
	ppcore := map[string]int{}
	for version, count := range s.AgentVersion {
		ppfull[version] += count

		matches := agentVersionRegex.FindStringSubmatch(version)
		if matches != nil {
			ppcore[matches[1]] += count
		}
	}

	txn, err := s.dbh.BeginTx(s.ServiceContext(), nil)
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
		err = pp.Insert(s.ServiceContext(), txn, boil.Infer())
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
		err = pp.Insert(s.ServiceContext(), txn, boil.Infer())
		if err != nil {
			log.WithError(err).Warnln("Could not insert peer property txn")
			continue
		}
	}

	for p, count := range s.Protocols {
		pp := &models.PeerProperty{
			Property: "protocol",
			Value:    p,
			Count:    count,
			CrawlID:  crawl.ID,
		}
		err = pp.Insert(s.ServiceContext(), txn, boil.Infer())
		if err != nil {
			log.WithError(err).Warnln("Could not insert peer property txn")
			continue
		}
	}

	for errKey, count := range s.Errors {
		pp := &models.PeerProperty{
			Property: "error",
			Value:    errKey,
			Count:    count,
			CrawlID:  crawl.ID,
		}
		err = pp.Insert(s.ServiceContext(), txn, boil.Infer())
		if err != nil {
			log.WithError(err).Warnln("Could not insert peer property txn")
			continue
		}
	}

	if err = txn.Commit(); err != nil {
		log.WithError(err).Warnln("Could not commit transaction")
	} else {
		log.Infoln("Saved peer properties")
	}
}

func (s *Scheduler) handleCrawlResult(cr CrawlResult) {
	logEntry := log.WithFields(log.Fields{
		"workerID": cr.WorkerID,
		"targetID": cr.Peer.ID.Pretty()[:16],
	})
	logEntry.Debugln("Handling crawl result from worker", cr.WorkerID)

	// TODO: Another go routine?
	startUpsert := time.Now()
	err := s.upsertCrawlResult(cr.Peer.ID, cr.Peer.Addrs, cr.Error)
	if err != nil {
		log.WithError(err).Warnln("Could not update peer")
	}
	stats.Record(s.ServiceContext(), metrics.CrawledUpsertDuration.M(millisSince(startUpsert)))

	s.crawled.Store(cr.Peer.ID, true)
	s.crawledCount.Inc()
	s.inCrawlQueue.Delete(cr.Peer.ID)
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(s.inCrawlQueueCount.Dec())))

	s.AgentVersionLk.Lock()
	if cr.Agent != "" {
		s.AgentVersion[cr.Agent] += 1
	} else {
		s.AgentVersion["n.a."] += 1
	}
	s.AgentVersionLk.Unlock()

	s.ProtocolsLk.Lock()
	for _, p := range cr.Protocols {
		s.Protocols[p] += 1
	}
	s.ProtocolsLk.Unlock()
	stats.Record(s.ServiceContext(), metrics.CrawledPeersCount.M(1))

	if cr.Error != nil {
		s.crawlErrors.Inc()

		s.ErrorsLk.Lock()
		known := false
		for errKey, errStr := range knownErrors {
			if strings.Contains(cr.Error.Error(), errStr) {
				s.Errors[errKey] += 1
				known = true
				break
			}
		}
		if !known {
			s.Errors["unknown"] += 1
			logEntry = logEntry.WithError(cr.Error)
		}
		s.ErrorsLk.Unlock()

		logEntry.WithError(cr.Error).Debugln("Error crawling peer")
	} else {
		for _, pi := range cr.Neighbors {
			_, inCrawlQueue := s.inCrawlQueue.Load(pi.ID)
			_, crawled := s.crawled.Load(pi.ID)
			if !inCrawlQueue && !crawled {
				s.dispatchCrawl(pi)
			}
		}
	}

	logEntry.WithFields(map[string]interface{}{
		"inCrawlQueue": s.inCrawlQueueCount.Load(),
		"crawled":      s.crawledCount.Load(),
	}).Infoln("Handled crawl result from worker", cr.WorkerID)

	if s.inCrawlQueueCount.Load() == 0 || (s.config.CrawlLimit > 0 && int(s.crawledCount.Load()) >= s.config.CrawlLimit) {
		s.Shutdown()
	}
}

// upsertCrawlResult inserts the given peer with its multi addresses in the database and
// upserts its currently active session
func (s *Scheduler) upsertCrawlResult(peerID peer.ID, maddrs []ma.Multiaddr, dialErr error) error {
	// Check if we're in a dry-run
	if s.dbh == nil {
		return nil
	}

	if dialErr == nil {
		if err := db.UpsertPeer(s.ServiceContext(), s.dbh, peerID.Pretty(), maddrs); err != nil {
			return errors.Wrap(err, "upsert peer")
		}
		if err := db.UpsertSessionSuccess(s.dbh, peerID.Pretty()); err != nil {
			return errors.Wrap(err, "upsert session success")
		}
	} else if dialErr != s.ServiceContext().Err() {
		if err := db.UpsertSessionError(s.dbh, peerID.Pretty()); err != nil {
			return errors.Wrap(err, "upsert session error")
		}
	}

	return nil
}

func (s *Scheduler) dispatchCrawl(pi peer.AddrInfo) {
	select {
	case <-s.SigShutdown():
		log.Debugln("Skipping dispatch as orchestrator shuts down")
		return
	default:
		s.inCrawlQueue.Store(pi.ID, true)
		stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(s.inCrawlQueueCount.Inc())))
		go func() { s.crawlQueue <- pi }()
	}
}

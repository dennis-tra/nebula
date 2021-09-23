package crawl

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

const agentVersionRegexPattern = `\/?go-ipfs\/(?P<core>\d+\.\d+\.\d+)-?(?P<prerelease>\w+)?\/(?P<commit>\w+)?`

var agentVersionRegex = regexp.MustCompile(agentVersionRegexPattern)

// The Scheduler handles the scheduling and managing of workers that crawl the network
// as well as handling the crawl the results by persisting them in the database.
type Scheduler struct {
	// Service represents an entity that runs in a
	// separate go routine and where its lifecycle
	// needs to be handled externally.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// The database client
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// Instance of this crawl
	crawl *models.Crawl

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inCrawlQueue map[peer.ID]peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer has already been crawled
	// in the past, so we don't put in the crawl queue again.
	crawled map[peer.ID]peer.AddrInfo

	// TODO:
	dbPeers map[peer.ID]*models.Peer

	// The queue of peer.AddrInfo's that still need to be crawled.
	crawlQueue *queue.FIFO

	// The queue that the workers publish their crawl results on, so that the
	// scheduler can handle them, e.g. update the maps above etc.
	resultsQueue *queue.FIFO

	// A queue that take crawl results and gets consumed by a worker that saves the data into the DB.
	persistQueue *queue.FIFO

	// A map of agent versions and their occurrences that happened during the crawl.
	agentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	protocols map[string]int

	// A map of errors that happened during the crawl.
	errors map[string]int

	// The list of worker node references.
	workers sync.Map
}

// knownErrors contains a list of known errors. Property key + string to match for
var knownErrors = map[string]string{
	models.DialErrorIoTimeout:               "i/o timeout",
	models.DialErrorConnectionRefused:       "connection refused",
	models.DialErrorProtocolNotSupported:    "protocol not supported",
	models.DialErrorPeerIDMismatch:          "peer id mismatch",
	models.DialErrorNoRouteToHost:           "no route to host",
	models.DialErrorNetworkUnreachable:      "network is unreachable",
	models.DialErrorNoGoodAddresses:         "no good addresses",
	models.DialErrorContextDeadlineExceeded: "context deadline exceeded",
	models.DialErrorNoPublicIP:              "no public IP address",
	models.DialErrorMaxDialAttemptsExceeded: "max dial attempts exceeded",
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(ctx context.Context, dbc *db.Client) (*Scheduler, error) {
	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, conf.DialTimeout)

	// Force direct dials will prevent swarm to run into dial backoff errors. It also prevents proxied connections.
	ctx = network.WithForceDirectDial(ctx, "prevent backoff")

	// Initialize a single libp2p node that's shared between all workers.
	// TODO: experiment with multiple nodes.
	// TODO: is the key pair really necessary? see "weak keys" handling in weizenbaum crawler.
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair")
	}

	h, err := libp2p.New(ctx, libp2p.Identity(priv), libp2p.NoListenAddrs)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	s := &Scheduler{
		Service:      service.New("scheduler"),
		host:         h,
		dbc:          dbc,
		config:       conf,
		inCrawlQueue: map[peer.ID]peer.AddrInfo{},
		crawled:      map[peer.ID]peer.AddrInfo{},
		dbPeers:      map[peer.ID]*models.Peer{},
		crawlQueue:   queue.NewFIFO(),
		resultsQueue: queue.NewFIFO(),
		persistQueue: queue.NewFIFO(),
		agentVersion: map[string]int{},
		protocols:    map[string]int{},
		errors:       map[string]int{},
		workers:      sync.Map{},
	}

	return s, nil
}

// CrawlNetwork starts the configured amount of workers and fills
// the worker queue with bootstrap nodes to start with.
func (s *Scheduler) CrawlNetwork(bootstrap []peer.AddrInfo) error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	if s.dbc != nil {
		crawl, err := s.dbc.InitCrawl(s.ServiceContext())
		if err != nil {
			return errors.Wrap(err, "creating crawl in db")
		}
		s.crawl = crawl
	}

	// Start all workers
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		w, err := NewWorker(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.workers.Store(i, w)
		go w.StartCrawling(s.crawlQueue, s.resultsQueue)
	}

	// Start all persisters
	var persisters []*Persister
	if s.dbc != nil {
		for i := 0; i < 10; i++ {
			p, err := NewPersister(s.dbc, s.config, s.crawl)
			if err != nil {
				return errors.Wrap(err, "new persister")
			}
			persisters = append(persisters, p)
			go p.StartPersisting(s.persistQueue)
		}
	}

	// Query known peers for bootstrapping
	if s.dbc != nil {
		bps, err := s.dbc.QueryBootstrapPeers(s.ServiceContext(), s.config.CrawlWorkerCount)
		if err != nil {
			log.WithError(err).Warnln("Could not query bootstrap peers")
		}
		bootstrap = append(bootstrap, bps...)
	}

	// Fill the queue with bootstrap nodes
	for _, b := range bootstrap {
		// This check is necessary as the query above could have returned a canonical bootstrap peer as well
		if _, inCrawlQueue := s.inCrawlQueue[b.ID]; !inCrawlQueue {
			s.scheduleCrawl(b)
		}
	}

	// Read from the results queue blocking - this is the main loop
	s.readResultsQueue()

	// Stop Go routine of crawl queue
	s.crawlQueue.DoneProducing()

	// Stop workers
	s.shutdownWorkers()

	// Stop Go routine of results queue
	s.resultsQueue.DoneProducing()

	// Stop Go routine of persist queue
	s.persistQueue.DoneProducing()

	// Wait for all persisters to finish
	for _, p := range persisters {
		log.Infoln("Waiting for persister to finish ", p.Identifier())
		<-p.SigDone()
	}

	// Finally, log the crawl summary
	defer s.logSummary()

	if s.dbc == nil {
		return nil
	}

	// Persist the crawl results
	if err := s.updateCrawl(context.Background()); err != nil {
		return errors.Wrap(err, "persist crawl")
	}

	if err := s.persistCrawlProperties(context.Background()); err != nil {
		return errors.Wrap(err, "persist crawl properties")
	}

	return nil
}

// readResultsQueue listens for crawl results on the resultsQueue channel and handles any
// entries in handleResult. If the scheduler is shut down it schedules a cleanup of resources
func (s *Scheduler) readResultsQueue() {
	for {
		select {
		case elem, ok := <-s.resultsQueue.Consume():
			if !ok {
				return
			}
			s.handleResult(elem.(Result))
		case <-s.SigShutdown():
			return
		}
	}
}

// handleResult takes a crawl result and persist the information in the database and schedules
// new crawls.
func (s *Scheduler) handleResult(cr Result) {
	start := time.Now()
	defer stats.Record(s.ServiceContext(), metrics.CrawlResultHandlingDuration.M(millisSince(start)))

	logEntry := log.WithFields(log.Fields{
		"workerID":   cr.WorkerID,
		"targetID":   cr.Peer.ID.Pretty()[:16],
		"isDialable": cr.Error == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.WorkerID)

	// update maps
	s.crawled[cr.Peer.ID] = cr.Peer
	stats.Record(s.ServiceContext(), metrics.CrawledPeersCount.M(1))
	delete(s.inCrawlQueue, cr.Peer.ID)
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))

	//// Check if we're in a dry-run
	//if s.dbc != nil {
	//	// persist session information
	//	if err := s.persistCrawlResult(cr); err != nil {
	//		if !errors.Is(err, context.Canceled) {
	//			log.WithError(err).Warnln("Could not persist crawl result")
	//		}
	//	}
	//}

	s.persistQueue.Push(cr)

	// track agent versions
	s.agentVersion[cr.Agent] += 1

	// track seen protocols
	for _, p := range cr.Protocols {
		s.protocols[p] += 1
	}

	// log error or schedule new crawls
	if cr.Error == nil {
		for _, pi := range cr.Neighbors {
			if _, inCrawlQueue := s.inCrawlQueue[pi.ID]; inCrawlQueue {
				continue
			}

			if _, crawled := s.crawled[pi.ID]; crawled {
				continue
			}

			s.scheduleCrawl(pi)
		}
	} else {
		// Count errors
		dialErr := determineDialError(cr.Error)
		s.errors[dialErr] += 1
		if dialErr == models.DialErrorUnknown {
			logEntry = logEntry.WithError(cr.Error)
		} else {
			logEntry = logEntry.WithField("dialErr", dialErr)
		}
	}

	logEntry.WithFields(map[string]interface{}{
		"inCrawlQueue": len(s.inCrawlQueue),
		"crawled":      len(s.crawled),
	}).Infoln("Handled crawl result from worker", cr.WorkerID)

	if len(s.inCrawlQueue) == 0 || s.config.ReachedCrawlLimit(len(s.crawled)) {
		go s.Shutdown()
	}
}

// schedule crawl takes the address information and inserts it in the crawl queue in a separate
// go routine so we don't block the results handler. Buffered channels won't work here as there could
// be thousands of peers waiting to be crawled, so we spawn a separate go routine each time.
// I'm not happy with this approach - isn't there another fan out concurrency pattern based on channels?
// This could be an approach: https://github.com/AsynkronIT/goring though it's without channels and only single consumer
func (s *Scheduler) scheduleCrawl(pi peer.AddrInfo) {
	s.inCrawlQueue[pi.ID] = pi
	s.crawlQueue.Push(pi)
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))
}

// shutdownWorkers sends shutdown signals to all workers and blocks until all have shut down.
func (s *Scheduler) shutdownWorkers() {
	var wg sync.WaitGroup
	s.workers.Range(func(_, worker interface{}) bool {
		w := worker.(*Worker)
		wg.Add(1)
		go func(w *Worker) {
			w.Shutdown()
			wg.Done()
		}(w)
		return true
	})
	wg.Wait()
}

//// shutdownPersisters
//func (s *Scheduler) shutdownPersisters() {
//	var wg sync.WaitGroup
//	s.persisters.Range(func(_, persister interface{}) bool {
//		p := persister.(*Persister)
//		wg.Add(1)
//		go func(p *Persister) {
//			p.Shutdown()
//			wg.Done()
//		}(p)
//		return true
//	})
//	wg.Wait()
//}

// logSummary logs the final results of the crawl.
func (s *Scheduler) logSummary() {
	log.Infoln("Logging crawl results...")

	log.Infoln("")
	for err, count := range s.errors {
		log.WithField("count", count).WithField("value", err).Infoln("Dial Error")
	}
	log.Infoln("")
	for agent, count := range s.agentVersion {
		log.WithField("count", count).WithField("value", agent).Infoln("Agent")
	}
	log.Infoln("")
	for protocol, count := range s.protocols {
		log.WithField("count", count).WithField("value", protocol).Infoln("Protocol")
	}
	log.Infoln("")

	log.WithFields(log.Fields{
		"crawledPeers":    len(s.crawled),
		"crawlDuration":   time.Now().Sub(s.StartTime).String(),
		"dialablePeers":   len(s.crawled) - s.TotalErrors(),
		"undialablePeers": s.TotalErrors(),
	}).Infoln("Finished crawl")
}

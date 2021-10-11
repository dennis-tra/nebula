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

// The Scheduler handles the scheduling and managing of
//   a) crawlers - They consume a queue of peer address information, visit them and publish their results
//                 on a separate results queue. This results queue is consumed by this scheduler and further
//                 processed
//   b) persisters - They consume a separate persist queue. Basically all results that are published on the
//                 crawl results queue gets passed on to the persisters. However, the scheduler investigates
//                 the crawl results and builds up aggregate information for the whole crawl. Letting the
//                 persister directly consume the results queue would not allow that.
type Scheduler struct {
	// Service represents an entity that runs in a separate go routine and where its lifecycle
	// needs to be handled externally. This is true for this scheduler, so we're embedding it here.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all crawlers.
	host host.Host

	// The database client
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// Instance of this crawl. This instance gets created right at the beginning of the crawl, so we have
	// an ID that we can link subsequent database entities with.
	crawl *models.Crawl

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put them there again.
	inCrawlQueue map[peer.ID]peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer has already been crawled
	// in the past, so we don't put them in the crawl queue again.
	crawled map[peer.ID]peer.AddrInfo // TODO: could be replaced by just a counter

	// The queue of peer.AddrInfo's that still need to be crawled.
	crawlQueue *queue.FIFO

	// The queue that the crawlers publish their results on, so that the scheduler can handle them,
	// e.g. update the maps above etc.
	resultsQueue *queue.FIFO

	// A queue that takes crawl results and gets consumed by persisters that save the data into the DB.
	persistQueue *queue.FIFO

	// A map of agent versions and their occurrences that happened during the crawl.
	agentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	protocols map[string]int

	// A map of errors that happened during the crawl.
	errors map[string]int

	// The list of worker node references.
	crawlers sync.Map
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(ctx context.Context, conf *config.Config, dbc *db.Client) (*Scheduler, error) {
	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, conf.DialTimeout)

	// Force direct dials will prevent swarm to run into dial backoff errors. It also prevents proxied connections.
	ctx = network.WithForceDirectDial(ctx, "prevent backoff")

	// Initialize a single libp2p node that's shared between all crawlers.
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
		crawlQueue:   queue.NewFIFO(),
		resultsQueue: queue.NewFIFO(),
		persistQueue: queue.NewFIFO(),
		agentVersion: map[string]int{},
		protocols:    map[string]int{},
		errors:       map[string]int{},
		crawlers:     sync.Map{},
	}

	return s, nil
}

// CrawlNetwork starts the configured amount of crawlers and fills
// the crawl queue with bootstrap nodes to start with. These bootstrap
// nodes will be enriched by nodes we have seen in the past from the
// database. It also starts the persisters
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

	// Start all crawlers
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		c, err := NewCrawler(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.crawlers.Store(i, c)
		go c.StartCrawling(s.crawlQueue, s.resultsQueue)
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

	// Indicate that we won't publish any new crawl tasks to the queue.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.crawlQueue.DoneProducing()

	// Stop crawlers - blocking
	s.shutdownCrawlers()

	// Indicate that the crawlers won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.resultsQueue.DoneProducing()

	// Indicate that we won't send any more results to the persisters. This will
	// lead the persisters to consume the queue until the end and then stop automatically,
	// so we can wait below
	// As the persister consume the queue completely this won't leak a Go routine.
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
	if err := s.updateCrawl(context.Background(), len(s.inCrawlQueue) == 0); err != nil {
		return errors.Wrap(err, "persist crawl")
	}

	if err := s.persistCrawlProperties(context.Background()); err != nil {
		return errors.Wrap(err, "persist crawl properties")
	}

	return nil
}

// readResultsQueue listens for crawl results on the resultsQueue and handles any
// entries in handleResult. If the scheduler is asked to shut down it
// breaks out of this loop and the clean-up routines above take over.
func (s *Scheduler) readResultsQueue() {
	for {
		// Give the shutdown signal precedence
		select {
		case <-s.SigShutdown():
			return
		default:
		}

		select {
		case <-s.SigShutdown():
			return
		case elem, ok := <-s.resultsQueue.Consume():
			if !ok {
				return
			}
			s.handleResult(elem.(Result))
		}
	}
}

// handleResult takes a crawl result, aggregates crawl information and publishes the result
// to the persist queue, so that the persisters can persist the information in the database.
// It also looks into the result and publishes new crawl jobs based on whether the found peers
// weren't crawled before or are not already in the queue.
func (s *Scheduler) handleResult(cr Result) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  cr.CrawlerID,
		"targetID":   cr.Peer.ID.Pretty()[:16],
		"isDialable": cr.Error == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.CrawlerID)

	// Keep track that this peer was crawled, so we don't do it again during this run
	s.crawled[cr.Peer.ID] = cr.Peer
	stats.Record(s.ServiceContext(), metrics.CrawledPeersCount.M(1))

	// Remove peer from crawl queue map as it is not in there anymore
	delete(s.inCrawlQueue, cr.Peer.ID)
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))

	// Publish crawl result to persist queue so that the data is saved into the DB.
	s.persistQueue.Push(cr)

	// Track agent versions
	s.agentVersion[cr.Agent] += 1

	// Track seen protocols
	for _, p := range cr.Protocols {
		s.protocols[p] += 1
	}

	// Log error or schedule new crawls
	if cr.Error == nil {
		for _, pi := range cr.Neighbors {
			// Don't add this peer to the queue if its already in it
			if _, inCrawlQueue := s.inCrawlQueue[pi.ID]; inCrawlQueue {
				continue
			}

			// Don't add the peer to the queue if we have already visited it
			if _, crawled := s.crawled[pi.ID]; crawled {
				continue
			}

			s.scheduleCrawl(pi)
		}
	} else {
		// Count errors
		s.errors[cr.DialError] += 1
		if cr.DialError == models.DialErrorUnknown {
			logEntry = logEntry.WithError(cr.Error)
		} else {
			logEntry = logEntry.WithField("dialErr", cr.DialError)
		}
	}

	logEntry.WithFields(map[string]interface{}{
		"inCrawlQueue": len(s.inCrawlQueue),
		"crawled":      len(s.crawled),
	}).Infoln("Handled crawl result from worker", cr.CrawlerID)

	// If the queue is empty or we have reached the configured limit we stop the crawl.
	if len(s.inCrawlQueue) == 0 || s.config.ReachedCrawlLimit(len(s.crawled)) {
		go s.Shutdown()
	}
}

// scheduleCrawl takes the address information, inserts it in the crawl queue and updates the associated map.
func (s *Scheduler) scheduleCrawl(pi peer.AddrInfo) {
	s.inCrawlQueue[pi.ID] = pi
	s.crawlQueue.Push(pi)
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))
}

// shutdownCrawlers sends shutdown signals to all crawlers and blocks until all have shut down.
func (s *Scheduler) shutdownCrawlers() {
	var wg sync.WaitGroup
	s.crawlers.Range(func(_, worker interface{}) bool {
		w := worker.(*Crawler)
		wg.Add(1)
		go func(w *Crawler) {
			w.Shutdown()
			wg.Done()
		}(w)
		return true
	})
	wg.Wait()
}

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

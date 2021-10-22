package crawl

import (
	"context"
	"regexp"
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
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
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
	// The libp2p node that's used to crawl the network. This one is also passed to all crawlers.
	host host.Host

	// The database client
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// Instance of this crawl. This instance gets created right at the beginning of the crawl, so we have
	// an ID that we can link subsequent database entities with.
	crawl *models.Crawl

	// The timestamp when the crawler was started
	crawlStart time.Time

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
	}

	return s, nil
}

// CrawlNetwork starts the configured amount of crawlers and fills
// the crawl queue with bootstrap nodes to start with. These bootstrap
// nodes will be enriched by nodes we have seen in the past from the
// database. It also starts the persisters
func (s *Scheduler) CrawlNetwork(ctx context.Context, bootstrap []peer.AddrInfo) error {
	s.crawlStart = time.Now()

	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	if err := s.initCrawl(ctx); err != nil {
		return err
	}

	// Start all crawlers
	crawlers, crawlerCancel, err := s.startCrawlers(ctx)
	if err != nil {
		return err
	}
	defer crawlerCancel()

	// Start all persisters
	persisters, persistersCancel, err := s.startPersisters(ctx)
	if err != nil {
		return err
	}
	defer persistersCancel()

	// Query known peers for bootstrapping
	bootstrap = append(bootstrap, s.cachedBootstrapPeers(ctx)...)

	// Fill the queue with bootstrap nodes
	for _, b := range bootstrap {
		s.tryScheduleCrawl(ctx, b)
	}

	// Read from the results queue blocking - this is the main loop
	s.readResultsQueue(ctx)

	// Indicate that we won't publish any new crawl tasks to the queue.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.crawlQueue.DoneProducing()

	// Stop crawlers - blocking
	crawlerCancel()
	for _, c := range crawlers {
		log.WithField("persisterID", c.id).Debugln("Waiting for crawler to stop")
		<-c.done
	}

	// Indicate that the crawlers won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.resultsQueue.DoneProducing()

	// Indicate that we won't send any more results to the persisters. This will
	// lead the persisters to consume the queue until the end and then stop automatically,
	// so we can wait below
	// As the persister consume the queue completely this won't leak a Go routine.
	s.persistQueue.DoneProducing()

	// Wait for all persisters to finish
	persistersCancel()
	for _, p := range persisters {
		log.WithField("persisterID", "").Debugln("Waiting for persister to stop")
		<-p.done
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

// cachedBootstrapPeers queries peers from the database with an active session.
// If an error occurs it is only logged and an empty slice is returned as
// bootstrap peers are usually given during crawl start. The resulting list
// of cached peers could theoretically overlap with the list of bootstrap peers
// that were given during crawl start. This double-crawl is prevented in the
// tryScheduleCrawl method.
func (s *Scheduler) cachedBootstrapPeers(ctx context.Context) []peer.AddrInfo {
	if s.dbc == nil {
		return []peer.AddrInfo{}
	}

	bps, err := s.dbc.QueryBootstrapPeers(ctx, s.config.CrawlWorkerCount)
	if err != nil {
		log.WithError(err).Warnln("Could not query bootstrap peers")
		return []peer.AddrInfo{}
	}

	return bps
}

// initCrawl inserts a row in the crawls table (if it's not a dry run) with
// the state of Started and saves it on the crawl field of the scheduler.
func (s *Scheduler) initCrawl(ctx context.Context) error {
	if s.dbc == nil {
		return nil
	}

	crawl, err := s.dbc.InitCrawl(ctx)
	if err != nil {
		return errors.Wrap(err, "creating crawl in db")
	}
	s.crawl = crawl
	s.crawlStart = crawl.StartedAt

	return nil
}

// startCrawlers initializes Crawler structs and instructs them to read the crawlQueue to _start crawling_.
// The returned cancelFunc can be used to stop the crawlers from reading from the crawlQueue and "shut down".
func (s *Scheduler) startCrawlers(ctx context.Context) ([]*Crawler, context.CancelFunc, error) {
	crawlerCtx, crawlerCancel := context.WithCancel(ctx)

	var crawlers []*Crawler
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		c, err := NewCrawler(s.host, s.config)
		if err != nil {
			crawlerCancel()
			return nil, nil, errors.Wrap(err, "new crawler")
		}
		crawlers = append(crawlers, c)
		go c.StartCrawling(crawlerCtx, s.crawlQueue, s.resultsQueue)
	}

	return crawlers, crawlerCancel, nil
}

// startPersisters initializes Persister structs and instructs them to read the persistQueue to _start persisting_.
// The returned cancelFunc can be used to stop the persisters from reading from the persistQueue and "shut down".
func (s *Scheduler) startPersisters(ctx context.Context) ([]*Persister, context.CancelFunc, error) {
	persistersCtx, persistersCancel := context.WithCancel(ctx)
	if s.dbc == nil {
		return []*Persister{}, persistersCancel, nil
	}

	var persisters []*Persister
	for i := 0; i < 10; i++ {
		p, err := NewPersister(s.dbc, s.config, s.crawl)
		if err != nil {
			persistersCancel()
			return nil, nil, errors.Wrap(err, "new persister")
		}
		persisters = append(persisters, p)
		go p.StartPersisting(persistersCtx, s.persistQueue)
	}

	return persisters, persistersCancel, nil
}

// readResultsQueue listens for crawl results on the resultsQueue and handles any
// entries in handleResult. If the scheduler is asked to shut down it
// breaks out of this loop and the clean-up routines above take over.
func (s *Scheduler) readResultsQueue(ctx context.Context) {
	var result Result
	for {
		// Give the shutdown signal precedence
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case elem, ok := <-s.resultsQueue.Consume():
			if !ok {
				return
			}
			result = elem.(Result)
		}

		s.handleResult(ctx, result)

		// If the queue is empty, or we have reached the configured limit we stop the crawl.
		if len(s.inCrawlQueue) == 0 || s.config.ReachedCrawlLimit(len(s.crawled)) {
			return
		}
	}
}

// handleResult takes a crawl result, aggregates crawl information and publishes the result
// to the persist queue, so that the persisters can persist the information in the database.
// It also looks into the result and publishes new crawl jobs based on whether the found peers
// weren't crawled before or are not already in the queue.
func (s *Scheduler) handleResult(ctx context.Context, cr Result) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  cr.CrawlerID,
		"targetID":   utils.FmtPeerID(cr.Peer.ID),
		"isDialable": cr.Error == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.CrawlerID)

	// Keep track that this peer was crawled, so we don't do it again during this run
	s.crawled[cr.Peer.ID] = cr.Peer
	stats.Record(ctx, metrics.CrawledPeersCount.M(1))

	// Remove peer from crawl queue map as it is not in there anymore
	delete(s.inCrawlQueue, cr.Peer.ID)
	stats.Record(ctx, metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))

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
			s.tryScheduleCrawl(ctx, pi)
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
}

// tryScheduleCrawl takes the address information, inserts it in the crawl queue and updates the associated map.
// The prefix "try" should indicate that there are the side-effects of checking whether the peer was already
// crawled or is already scheduled.
func (s *Scheduler) tryScheduleCrawl(ctx context.Context, pi peer.AddrInfo) {
	// Don't add this peer to the queue if it's already in it
	if _, inCrawlQueue := s.inCrawlQueue[pi.ID]; inCrawlQueue {
		return
	}

	// Don't add the peer to the queue if we have already visited it
	if _, crawled := s.crawled[pi.ID]; crawled {
		return
	}

	s.inCrawlQueue[pi.ID] = pi
	s.crawlQueue.Push(pi)
	stats.Record(ctx, metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))
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
		"crawlDuration":   time.Now().Sub(s.crawlStart).String(),
		"dialablePeers":   len(s.crawled) - s.TotalErrors(),
		"undialablePeers": s.TotalErrors(),
	}).Infoln("Finished crawl")
}

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (s *Scheduler) TotalErrors() int {
	sum := 0
	for _, count := range s.errors {
		sum += count
	}
	return sum
}

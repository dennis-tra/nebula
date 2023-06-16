package crawl

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

// The Scheduler handles the scheduling and managing of
//
//	a) crawlers - They consume a queue of peer address information, visit them and publish their results
//	              on a separate results queue. This results queue is consumed by this scheduler and further
//	              processed
//	b) persisters - They consume a separate persist queue. Basically all results that are published on the
//	              crawl results queue gets passed on to the persisters. However, the scheduler investigates
//	              the crawl results and builds up aggregate information for the whole crawl. Letting the
//	              persister directly consume the results queue would not allow that.
type Scheduler struct {
	// The libp2p node that's used to crawl the network. This one is also passed to all crawlers.
	host *basichost.BasicHost

	// The database client
	dbc db.Client

	// The configuration of timeouts etc.
	config *config.Crawl

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
	crawlQueue *queue.FIFO[peer.AddrInfo]

	// The queue that the crawlers publish their results on, so that the scheduler can handle them,
	// e.g. update the maps above etc.
	crawlResultsQueue *queue.FIFO[Result]

	// A queue that takes crawl results and gets consumed by persisters that save the data into the DB.
	persistQueue *queue.FIFO[Result]

	// The queue that the persisters publish their results on, so that the peerID -> database id mapping (peerMappings)
	// can be built.
	persistResultsQueue *queue.FIFO[*db.InsertVisitResult]

	// A map of agent versions and their occurrences that happened during the crawl.
	agentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	protocols map[string]int

	// A map of errors that happened during the crawl.
	errors map[string]int

	// A map that keeps track of all k-bucket entries of a particular peer.
	routingTables map[peer.ID]*RoutingTable

	// A map that maps peer IDs to their database IDs. This speeds up the insertion of neighbor information as
	// the database does not need to look up every peer ID but only the ones not yet present in the database.
	// Speed up for ~11k peers: 5.5 min -> 30s
	// TODO: Disable maintenance of map if dry-run is active
	peerMappings map[peer.ID]int
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Crawl, dbc db.Client) (*Scheduler, error) {

	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Initialize a single libp2p node that's shared between all crawlers.
	h, err := libp2p.New(libp2p.NoListenAddrs, libp2p.ResourceManager(rm), libp2p.UserAgent("nebula-crawler/"+conf.Root.Version()))
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	s := &Scheduler{
		host:                h.(*basichost.BasicHost),
		dbc:                 dbc,
		config:              conf,
		inCrawlQueue:        map[peer.ID]peer.AddrInfo{},
		crawled:             map[peer.ID]peer.AddrInfo{},
		crawlQueue:          queue.NewFIFO[peer.AddrInfo](),
		crawlResultsQueue:   queue.NewFIFO[Result](),
		persistQueue:        queue.NewFIFO[Result](),
		persistResultsQueue: queue.NewFIFO[*db.InsertVisitResult](),
		agentVersion:        map[string]int{},
		protocols:           map[string]int{},
		errors:              map[string]int{},
		routingTables:       map[peer.ID]*RoutingTable{},
		peerMappings:        map[peer.ID]int{},
	}

	return s, nil
}

// CrawlNetwork starts the configured amount of crawlers and fills
// the crawl queue with bootstrap nodes to start with. These bootstrap
// nodes will be enriched by nodes we have seen in the past from the
// database. It also starts the persisters
func (s *Scheduler) CrawlNetwork(ctx context.Context, bootstrap []peer.AddrInfo) error {
	log.Infoln("Starting to crawl the network")

	s.crawlStart = time.Now()

	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, s.config.Root.DialTimeout)

	// Force direct dials will prevent swarm to run into dial backoff errors. It also prevents proxied connections.
	ctx = network.WithForceDirectDial(ctx, "prevent backoff")

	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	if err := s.initCrawl(ctx); err != nil {
		return err
	}

	// Start all crawlers
	crawlers, crawlerCtx, crawlerCancel, err := s.startCrawlers(ctx)
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
	// TODO: This can still leak a Go routine. However, we're exiting here anyway...
	s.crawlQueue.DoneProducing()

	// Stop crawlers - blocking
	crawlerCancel()
	for _, c := range crawlers {
		log.WithField("crawlerID", c.id).Debugln("Waiting for crawler to stop")
		<-c.done
	}

	// Indicate that the crawlers won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However, we're exiting here anyway...
	s.crawlResultsQueue.DoneProducing()

	// Indicate that we won't send any more results to the persisters. This will
	// lead the persisters to consume the queue until the end and then stop automatically,
	// so we can wait below
	// As the persister consume the queue completely this won't leak a Go routine.
	s.persistQueue.DoneProducing()

	// Wait for all persisters to finish
	select {
	case <-ctx.Done():
		persistersCancel()
		log.Infoln("Cancelling persister context as root context stopped") // e.g. ^C
	default:
		log.Debugln("Not cancelling persister context as we stopped organically") // limit or just finished
	}
	for _, p := range persisters {
		log.WithField("persisterID", p.id).Infoln("Waiting for persister to stop")
		<-p.done
	}

	// Indicate that the persisters won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.persistResultsQueue.DoneProducing()

	// Finally, log the crawl summary
	defer s.logSummary()

	// Return early if we are in a dry-run
	if s.dbc == nil {
		return nil
	}

	// Persist the crawl results
	if err := s.updateCrawl(context.Background(), crawlerCtx, len(s.inCrawlQueue) == 0); err != nil {
		return fmt.Errorf("persist crawl: %w", err)
	}

	// Persist associated crawl properties
	if err := s.persistCrawlProperties(context.Background()); err != nil {
		return fmt.Errorf("persist crawl properties: %w", err)
	}

	// persist all neighbor information
	if s.config.PersistNeighbors {
		s.persistNeighbors(context.Background())
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

	log.Infoln("Initializing crawl...")
	crawl, err := s.dbc.InitCrawl(ctx)
	if err != nil {
		return fmt.Errorf("creating crawl in db: %w", err)
	}
	s.crawl = crawl
	s.crawlStart = crawl.StartedAt

	return nil
}

// startCrawlers initializes Crawler structs and instructs them to read the crawlQueue to _start crawling_.
// The returned cancelFunc can be used to stop the crawlers from reading from the crawlQueue and "shut down".
func (s *Scheduler) startCrawlers(ctx context.Context) ([]*Crawler, context.Context, context.CancelFunc, error) {
	crawlerCtx, crawlerCancel := context.WithCancel(ctx)

	var crawlers []*Crawler
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		c, err := NewCrawler(s.host, s.config)
		if err != nil {
			crawlerCancel()
			return nil, nil, nil, fmt.Errorf("new crawler: %w", err)
		}
		crawlers = append(crawlers, c)
		go c.StartCrawling(crawlerCtx, s.crawlQueue, s.crawlResultsQueue)
	}

	return crawlers, crawlerCtx, crawlerCancel, nil
}

// startPersisters initializes Persister structs and instructs them to read the persistQueue to _start persisting_.
// The returned cancelFunc can be used to stop the persisters from reading from the persistQueue and "shut down".
func (s *Scheduler) startPersisters(ctx context.Context) ([]*Persister, context.CancelFunc, error) {
	// Create dedicated context for the persisters
	persistersCtx, persistersCancel := context.WithCancel(ctx)
	if s.dbc == nil {
		return []*Persister{}, persistersCancel, nil
	}

	var persisters []*Persister
	for i := 0; i < 10; i++ {
		p, err := NewPersister(s.dbc, s.config, s.crawl)
		if err != nil {
			persistersCancel()
			return nil, nil, fmt.Errorf("new persister: %w", err)
		}
		persisters = append(persisters, p)
		go p.StartPersisting(persistersCtx, s.persistQueue, s.persistResultsQueue)
	}

	return persisters, persistersCancel, nil
}

// readResultsQueue listens for crawl results on the crawlResultsQueue and handles any
// entries in handleResult. If the scheduler is asked to shut down it
// breaks out of this loop and the clean-up routines above take over.
func (s *Scheduler) readResultsQueue(ctx context.Context) {
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
		case r, ok := <-s.crawlResultsQueue.Consume():
			if !ok {
				return
			}

			s.handleResult(ctx, r)

			// If the queue is empty, or we have reached the configured limit we stop the crawl.
			if len(s.inCrawlQueue) == 0 || s.config.ReachedCrawlLimit(len(s.crawled)) {
				return
			}
		case r, ok := <-s.persistResultsQueue.Consume():
			if !ok {
				return
			}

			if r != nil && r.PeerID != nil {
				s.peerMappings[r.PID] = *r.PeerID
			}
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
		"remoteID":   utils.FmtPeerID(cr.Peer.ID),
		"isDialable": cr.ConnectError == nil && cr.CrawlError == nil,
	})
	logEntry.Debugln("Handling crawl result from worker", cr.CrawlerID)

	// Keep track that this peer was crawled, so we don't do it again during this run
	s.crawled[cr.Peer.ID] = cr.Peer
	metrics.DistinctVisitedPeersCount.Inc()

	// Remove peer from crawl queue map as it is not in there anymore
	delete(s.inCrawlQueue, cr.Peer.ID)
	metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(len(s.inCrawlQueue)))

	// Publish crawl result to persist queue so that the data is saved into the DB.
	s.persistQueue.Push(cr)

	// Track agent versions
	s.agentVersion[cr.Agent] += 1

	// Track seen protocols
	for _, p := range cr.Protocols {
		s.protocols[p] += 1
	}

	// Schedule crawls of all found neighbors unless we got the routing table from the API.
	// In this case the routing table information won't include any MultiAddresses. This means
	// we can't use these peers for further crawls.
	if !cr.RoutingTableFromAPI {
		for _, rt := range cr.RoutingTable.Neighbors {
			s.tryScheduleCrawl(ctx, rt)
		}
	}

	if cr.ConnectError == nil {
		// Only track the neighbors if we were actually able to connect to the peer. Otherwise, we would track
		// an empty routing table of that peer. Only track the routing table in the neighbors table if at least
		// one FIND_NODE RPC succeeded.
		if s.config.PersistNeighbors && cr.RoutingTable.ErrorBits < math.MaxUint16 {
			s.routingTables[cr.Peer.ID] = cr.RoutingTable
		}
	} else if cr.ConnectError != nil {
		// Log and count connection errors
		s.errors[cr.ConnectErrorStr] += 1
		if cr.ConnectErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(cr.ConnectError)
		} else {
			logEntry = logEntry.WithField("dialErr", cr.ConnectErrorStr)
		}
	} else if cr.CrawlError != nil {
		// Log and count crawl errors
		s.errors[cr.CrawlErrorStr] += 1
		if cr.CrawlErrorStr == models.NetErrorUnknown {
			logEntry = logEntry.WithError(cr.CrawlError)
		} else {
			logEntry = logEntry.WithField("crawlErr", cr.CrawlErrorStr)
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

	// Schedule crawl for peer
	s.inCrawlQueue[pi.ID] = pi
	s.crawlQueue.Push(pi)

	// Track new peer in queue with prometheus
	metrics.VisitQueueLength.With(metrics.CrawlLabel).Set(float64(len(s.inCrawlQueue)))
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

// RoutingTable captures the routing table information and crawl error of a particular peer
type RoutingTable struct {
	// PeerID is the peer whose neighbors (routing table entries) are in the array below.
	PeerID peer.ID
	// The peers that are in the routing table of the above peer
	Neighbors []peer.AddrInfo
	// First error that has occurred during crawling that peer
	Error error
	// Little Endian representation of at which CPLs errors occurred during neighbors fetches.
	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	ErrorBits uint16
}

func (rt *RoutingTable) PeerIDs() []peer.ID {
	peerIDs := make([]peer.ID, len(rt.Neighbors))
	for i, neighbor := range rt.Neighbors {
		peerIDs[i] = neighbor.ID
	}
	return peerIDs
}

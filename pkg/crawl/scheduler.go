package crawl

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"time"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/network"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.opencensus.io/stats"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
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

	// The database handle
	dbh *sql.DB

	// The configuration of timeouts etc.
	config *config.Config

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inCrawlQueue map[peer.ID]peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer has already been crawled
	// in the past, so we don't put in the crawl queue again.
	crawled map[peer.ID]peer.AddrInfo

	// The queue of peer.AddrInfo's that still need to be crawled.
	crawlQueue chan peer.AddrInfo

	// The queue that the workers publish their crawl results on, so that the
	// scheduler can handle them, e.g. update the maps above etc.
	resultsQueue chan Result

	// A map of agent versions and their occurrences that happened during the crawl.
	AgentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	Protocols map[string]int

	// A map of errors that happened during the crawl.
	Errors map[string]int

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

func determineDialError(err error) string {
	for key, errStr := range knownErrors {
		if strings.Contains(err.Error(), errStr) {
			return key
		}
	}
	return models.DialErrorUnknown
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(ctx context.Context, dbh *sql.DB) (*Scheduler, error) {
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
		return nil, err
	}

	h, err := libp2p.New(ctx, libp2p.Identity(priv), libp2p.NoListenAddrs)
	if err != nil {
		return nil, err
	}

	p := &Scheduler{
		Service:      service.New("scheduler"),
		host:         h,
		dbh:          dbh,
		config:       conf,
		inCrawlQueue: map[peer.ID]peer.AddrInfo{},
		crawled:      map[peer.ID]peer.AddrInfo{},
		crawlQueue:   make(chan peer.AddrInfo),
		resultsQueue: make(chan Result),
		AgentVersion: map[string]int{},
		Protocols:    map[string]int{},
		Errors:       map[string]int{},
		workers:      sync.Map{},
	}

	return p, nil
}

// CrawlNetwork starts the configured amount of workers and fills
// the worker queue with bootstrap nodes to start with.
func (s *Scheduler) CrawlNetwork(bootstrap []peer.AddrInfo) error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	s.StartTime = time.Now()

	// Start all workers
	for i := 0; i < s.config.CrawlWorkerCount; i++ {
		w, err := NewWorker(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.workers.Store(i, w)
		go w.StartCrawling(s.crawlQueue, s.resultsQueue)
	}

	// Fill the queue with bootstrap nodes
	for _, b := range bootstrap {
		s.scheduleCrawl(b)
	}

	// Read from the results queue blocking
	s.readResultsQueue()

	// release all resources
	s.cleanup()

	defer func() {
		log.Infoln("Logging crawl results...")

		log.Infoln("")
		for err, count := range s.Errors {
			log.WithField("count", count).WithField("value", err).Infoln("Dial Error")
		}
		log.Infoln("")
		for agent, count := range s.AgentVersion {
			log.WithField("count", count).WithField("value", agent).Infoln("Agent")
		}
		log.Infoln("")
		for protocol, count := range s.Protocols {
			log.WithField("count", count).WithField("value", protocol).Infoln("Protocol")
		}
		log.Infoln("")

		log.WithFields(log.Fields{
			"crawledPeers":    len(s.crawled),
			"crawlDuration":   time.Now().Sub(s.StartTime).String(),
			"dialablePeers":   len(s.crawled) - s.TotalErrors(),
			"undialablePeers": s.TotalErrors(),
		}).Infoln("Finished crawl")
	}()

	if s.dbh != nil {
		crawl, err := s.persistCrawl(context.Background())
		if err != nil {
			return errors.Wrap(err, "persist crawl")
		}

		if err := s.persistPeerProperties(context.Background(), crawl.ID); err != nil {
			return errors.Wrap(err, "persist peer properties")
		}
	}

	return nil
}

// readResultsQueue listens for crawl results on the resultsQueue channel and handles any
// entries in handleResult. If the scheduler is shut down it schedules a cleanup of resources
func (s *Scheduler) readResultsQueue() {
	for {
		select {
		case result := <-s.resultsQueue:
			start := time.Now()
			s.handleResult(result)
			stats.Record(s.ServiceContext(), metrics.CrawlResultHandlingDuration.M(millisSince(start)))
		case <-s.SigShutdown():
			return
		}
	}
}

// handleResult takes a crawl result and persist the information in the database and schedules
// new crawls.
func (s *Scheduler) handleResult(cr Result) {
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

	// persist session information
	if err := s.upsertCrawlResult(cr); err != nil {
		log.WithError(err).Warnln("Could not update peer")
	}

	// track agent versions
	s.AgentVersion[cr.Agent] += 1

	// track seen protocols
	for _, p := range cr.Protocols {
		s.Protocols[p] += 1
	}

	// track error or schedule new crawls
	if cr.Error == nil {
		for _, pi := range cr.Neighbors {
			_, inCrawlQueue := s.inCrawlQueue[pi.ID]
			_, crawled := s.crawled[pi.ID]
			if !inCrawlQueue && !crawled {
				s.scheduleCrawl(pi)
			}
		}
	} else {
		// Count errors
		dialErr := determineDialError(cr.Error)
		s.Errors[dialErr] += 1
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

	if len(s.inCrawlQueue) == 0 || (s.config.CrawlLimit > 0 && len(s.crawled) >= s.config.CrawlLimit) {
		go s.Shutdown()
	}
}

// upsertCrawlResult inserts the given peer with its multi addresses in the database and
// upserts its currently active session
func (s *Scheduler) upsertCrawlResult(cr Result) error {
	// Check if we're in a dry-run
	if s.dbh == nil {
		return nil
	}

	startUpsert := time.Now()
	if cr.Error == nil {
		// No error, update peer record in DB
		oldMaddrStrs, err := db.UpsertPeer(s.dbh, cr.Peer.ID.Pretty(), cr.Peer.Addrs)
		if err != nil {
			return errors.Wrap(err, "upsert peer")
		}

		// Parse old multi-addresses
		oldMaddrs, err := addrsToMaddrs(oldMaddrStrs)
		if err != nil {
			return errors.Wrap(err, "addrs to maddrs")
		}

		// Check if the new multi-addresses constitute a new session
		if isNewSession(oldMaddrs, cr.Peer.Addrs) {
			// This is a new session as there is no overlap in multi-addresses - invalidate current session
			if err := db.UpsertSessionError(s.dbh, cr.Peer.ID.Pretty(), time.Now(), models.DialErrorMaddrReset); err != nil {
				return errors.Wrap(err, "upsert session error maddr reset")
			}
		}

		// Upsert peer session
		if err := db.UpsertSessionSuccess(s.dbh, cr.Peer.ID.Pretty()); err != nil {
			return errors.Wrap(err, "upsert session success")
		}
	} else if cr.Error != s.ServiceContext().Err() {
		dialErr := determineDialError(cr.Error)
		if err := db.UpsertSessionError(s.dbh, cr.Peer.ID.Pretty(), cr.ErrorTime, dialErr); err != nil {
			return errors.Wrap(err, "upsert session error")
		}
	}
	stats.Record(s.ServiceContext(), metrics.CrawledUpsertDuration.M(millisSince(startUpsert)))
	return nil
}

func isNewSession(oldMaddrs []ma.Multiaddr, newMaddrs []ma.Multiaddr) bool {
	if len(oldMaddrs) == 0 && len(newMaddrs) == 0 {
		return false
	}

	for _, oldMaddr := range oldMaddrs {
		for _, newMaddr := range newMaddrs {
			// If any multi address is equal to the previous one it is considered the same session.
			if oldMaddr.Equal(newMaddr) {
				return false
			}
		}
	}
	return true
}

func addrsToMaddrs(addrs []string) ([]ma.Multiaddr, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		maddrs[i] = maddr
	}
	return maddrs, nil
}

func maddrsToAddrs(maddrs []ma.Multiaddr) []string {
	addrs := make([]string, len(maddrs))
	for i, maddr := range maddrs {
		addrs[i] = maddr.String()
	}
	return addrs
}

// schedule crawl takes the address information and inserts it in the crawl queue in a separate
// go routine so we don't block the results handler. Buffered channels won't work here as there could
// be thousands of peers waiting to be crawled, so we spawn a separate go routine each time.
// I'm not happy with this approach - isn't there another fan out concurrency pattern based on channels?
// This could be an approach: https://github.com/AsynkronIT/goring though it's without channels and only single consumer
func (s *Scheduler) scheduleCrawl(pi peer.AddrInfo) {
	s.inCrawlQueue[pi.ID] = pi
	stats.Record(s.ServiceContext(), metrics.PeersToCrawlCount.M(float64(len(s.inCrawlQueue))))
	go func() {
		if s.IsStarted() {
			s.crawlQueue <- pi
		}
	}()
}

// cleanup handles the release of all resources allocated by the scheduler.
// Make sure to not access any maps here as this is run in a separate
// go routine from the handleResult method.
//
// remove all peers from the crawl queue. There could be pending writes on the crawl
// queue in scheduleCrawl() that would lead to `panic: send on closed channel` if we
// didn't drain the queue prior closing the channel.
func (s *Scheduler) cleanup() {
	s.drainCrawlQueue()
	close(s.crawlQueue)
	s.shutdownWorkers()
	close(s.resultsQueue)
}

// drainCrawlQueue reads all entries from crawlQueue and discards them.
func (s *Scheduler) drainCrawlQueue() {
	for {
		select {
		case pi := <-s.crawlQueue:
			log.WithField("targetID", pi.ID.Pretty()[:16]).Traceln("Drained peer")
		default:
			return
		}
	}
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

// persistCrawl writes crawl statistics to the database.
func (s *Scheduler) persistCrawl(ctx context.Context) (*models.Crawl, error) {
	log.Infoln("Persisting crawl result...")

	crawl := &models.Crawl{
		StartedAt:       s.StartTime,
		FinishedAt:      time.Now(),
		CrawledPeers:    len(s.crawled),
		DialablePeers:   len(s.crawled) - s.TotalErrors(),
		UndialablePeers: s.TotalErrors(),
	}

	return crawl, crawl.Insert(ctx, s.dbh, boil.Infer())
}

// persistPeerProperties writes peer property statistics to the database.
func (s *Scheduler) persistPeerProperties(ctx context.Context, crawlID int) error {
	log.Infoln("Persisting peer properties...")

	// Extract full and core agent versions. Core agent versions are just strings like 0.8.0 or 0.5.0
	// The full agent versions have much more information e.g., /go-ipfs/0.4.21-dev/789dab3
	avFull := map[string]int{}
	avCore := map[string]int{}
	for version, count := range s.AgentVersion {
		avFull[version] += count
		matches := agentVersionRegex.FindStringSubmatch(version)
		if matches != nil {
			avCore[matches[1]] += count
		}
	}

	txn, err := s.dbh.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("start txn")
	}

	for property, valuesMap := range map[string]map[string]int{
		"agent_version":      avFull,
		"agent_version_core": avCore,
		"protocol":           s.Protocols,
		"error":              s.Errors,
	} {
		for value, count := range valuesMap {
			pp := &models.PeerProperty{
				Property: property,
				Value:    value,
				Count:    count,
				CrawlID:  crawlID,
			}
			err := pp.Insert(ctx, txn, boil.Infer())
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"property": property,
					"value":    value,
				}).Warnln("Could not insert peer property txn")
				continue
			}
		}
	}

	return txn.Commit()
}

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (s *Scheduler) TotalErrors() int {
	sum := 0
	for _, count := range s.Errors {
		sum += count
	}
	return sum
}

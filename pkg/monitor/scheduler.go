package monitor

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

type Scheduler struct { // Service represents an entity that runs in a
	// separate go routine and where its lifecycle
	// needs to be handled externally.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// The database handle
	dbh *sql.DB

	// The configuration of timeouts etc.
	config *config.Config

	// The queue of peer.AddrInfo's that need to be dialed to.
	dialQueue chan peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inDialQueue sync.Map

	// The number of peers in the ping queue.
	inDialQueueCount atomic.Uint32

	// The queue that the workers publish their dial results on
	resultsQueue chan Result

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
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(ctx, libp2p.Identity(priv), libp2p.NoListenAddrs)
	if err != nil {
		return nil, err
	}

	m := &Scheduler{
		Service:      service.New("scheduler"),
		host:         h,
		dbh:          dbh,
		config:       conf,
		inDialQueue:  sync.Map{},
		dialQueue:    make(chan peer.AddrInfo),
		resultsQueue: make(chan Result),
		workers:      sync.Map{},
	}

	return m, nil
}

func (s *Scheduler) StartMonitoring() error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	s.StartTime = time.Now()

	// Start all workers
	for i := 0; i < s.config.MonitorWorkerCount; i++ {
		w, err := NewWorker(s.dbh, s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.workers.Store(i, w)
		go w.StartDialing(s.dialQueue, s.resultsQueue)
	}

	// Async handle the results from workers
	go s.handleResults()

	// Monitor the database and schedule dial jobs
	s.monitorDatabase()

	// release all resources
	s.cleanup()

	log.WithFields(log.Fields{
		"inDialQueue":     s.inDialQueueCount.Load(),
		"monitorDuration": time.Now().Sub(s.StartTime).String(),
	}).Infoln("Finished monitoring")

	return nil
}

func (s *Scheduler) handleResults() {
	for dr := range s.resultsQueue {
		logEntry := log.WithFields(log.Fields{
			"workerID": dr.WorkerID,
			"targetID": dr.Peer.ID.Pretty()[:16],
			"alive":    dr.Error == nil,
		})
		if dr.Error != nil {
			logEntry = logEntry.WithError(dr.Error)
		}
		logEntry.Infoln("Handling dial result from worker", dr.WorkerID)

		// Update maps
		s.inDialQueue.Delete(dr.Peer.ID)
		stats.Record(s.ServiceContext(), metrics.PeersToDialCount.M(float64(s.inDialQueueCount.Dec())))

		var err error
		if dr.Error == nil {
			err = db.UpsertSessionSuccess(s.dbh, dr.Peer.ID.Pretty())
		} else {
			dialErr := determineDialError(dr.Error)
			if ctx, err := tag.New(s.ServiceContext(), tag.Upsert(metrics.KeyError, dialErr)); err == nil {
				stats.Record(ctx, metrics.PeersToDialErrorsCount.M(1))
			}
			err = db.UpsertSessionError(s.dbh, dr.Peer.ID.Pretty(), dr.ErrorTime, dialErr)
		}

		if err != nil {
			logEntry.WithError(err).Warn("Could not update session record")
		}
	}
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (s *Scheduler) monitorDatabase() {
	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := s.fetchSessions()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		// For every session schedule that it gets pushed into the dialQueue
		for _, session := range sessions {
			if err = s.scheduleDial(session); err != nil {
				log.WithError(err).Warnln("Could not schedule dial")
			}
		}
		log.Infof("In dial queue %d peers", s.inDialQueueCount.Load())

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
		case <-s.SigShutdown():
			return
		}
	}
}

func (s *Scheduler) fetchSessions() (models.SessionSlice, error) {
	dueSessions, err := db.FetchDueSessions(s.ServiceContext(), s.dbh)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "fetch due sessions")
	}
	log.Infof("Found %d due sessions\n", len(dueSessions))

	return dueSessions, nil
}

func (s *Scheduler) scheduleDial(session *models.Session) error {
	// Parse peer ID from database
	peerID, err := peer.Decode(session.R.Peer.ID)
	if err != nil {
		return errors.Wrap(err, "decode peer ID")
	}
	logEntry := log.WithField("peerID", peerID.Pretty()[:16])

	// Parse multi addresses from database
	pi := peer.AddrInfo{ID: peerID}
	for _, maddrStr := range session.R.Peer.MultiAddresses {
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			logEntry.WithError(err).Warnln("Could not parse multi address")
			continue
		}
		pi.Addrs = append(pi.Addrs, maddr)
	}

	// Check if peer is already in dial queue
	if _, inPingQueue := s.inDialQueue.LoadOrStore(peerID, pi); inPingQueue {
		logEntry.Traceln("Peer already in dial queue")
		return nil
	}
	stats.Record(s.ServiceContext(), metrics.PeersToDialCount.M(float64(s.inDialQueueCount.Inc())))

	// Schedule dial for peer
	go func() {
		if s.IsStarted() {
			s.dialQueue <- pi
		}
	}()

	return nil
}

func (s *Scheduler) cleanup() {
	s.drainDialQueue()
	close(s.dialQueue)
	s.shutdownWorkers()
	close(s.resultsQueue)
}

// drainDialQueue reads all entries from crawlQueue and discards them.
func (s *Scheduler) drainDialQueue() {
	for {
		select {
		case pi := <-s.dialQueue:
			log.WithField("targetID", pi.ID.Pretty()[:16]).Debugln("Drained peer")
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

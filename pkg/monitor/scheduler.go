package monitor

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/models"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"
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

	// The queue of peer.AddrInfo's that need to be connected to.
	connectQueue chan peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inConnectQueue sync.Map

	// The number of peers in the ping queue.
	inConnectQueueCount atomic.Uint32

	// The queue that the workers publish their connect results on
	resultsQueue chan Result

	// The list of worker node references.
	workers sync.Map
}

func NewScheduler(ctx context.Context, dbh *sql.DB) (*Scheduler, error) {
	// Initialize a single libp2p node that's shared between all workers.
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

	m := &Scheduler{
		Service:        service.New("scheduler"),
		host:           h,
		dbh:            dbh,
		config:         conf,
		inConnectQueue: sync.Map{},
		connectQueue:   make(chan peer.AddrInfo),
		resultsQueue:   make(chan Result),
		workers:        sync.Map{},
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
		go w.StartConnecting(s.connectQueue, s.resultsQueue)
	}

	// Async handle the results from workers
	go s.handleResults()

	// Monitor the database and schedule connect jobs
	s.monitorDatabase()

	// release all resources
	s.cleanup()

	log.WithFields(log.Fields{
		"inConnectQueue":  s.inConnectQueueCount.Load(),
		"monitorDuration": time.Now().Sub(s.StartTime).String(),
	}).Infoln("Finished monitoring")

	return nil
}

func (s *Scheduler) handleResults() {
	for dr := range s.resultsQueue {
		logEntry := log.WithFields(log.Fields{
			"workerID": dr.WorkerID,
			"targetID": dr.Peer.ID.Pretty()[:16],
			"alive":    dr.Alive,
		})
		logEntry.Infoln("Handling connect result from worker", dr.WorkerID)

		// Update maps
		s.inConnectQueue.Delete(dr.Peer.ID)
		stats.Record(s.ServiceContext(), metrics.PeersToConnectCount.M(float64(s.inConnectQueueCount.Dec())))

		var err error
		if dr.Alive {
			logEntry.Traceln("Peer still reachable")
			err = db.UpsertSessionSuccess(s.dbh, dr.Peer.ID.Pretty())
		} else {
			logEntry.Traceln("Peer not reachable anymore")
			err = db.UpsertSessionError(s.dbh, dr.Peer.ID.Pretty())
		}

		if err != nil {
			logEntry.WithError(err).WithField("alive", dr.Alive).Warn("Could not update session record")
		}
	}
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (s *Scheduler) monitorDatabase() {
	for {
		sessions, err := s.fetchSessions()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		// For every session schedule that it gets pushed into the connectQueue
		for _, session := range sessions {
			if err = s.scheduleConnect(session); err != nil {
				log.WithError(err).Warnln("Could not schedule connect")
			}
		}

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

	rfSessions, err := db.FetchRecentlyFinishedSessions(s.ServiceContext(), s.dbh)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "fetch recently finished sessions")
	}
	log.Infof("Found %d recently finished sessions\n", len(rfSessions))

	return append(dueSessions, rfSessions...), nil
}

func (s *Scheduler) scheduleConnect(session *models.Session) error {
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

	// Check if peer is already in connect queue
	if _, inPingQueue := s.inConnectQueue.LoadOrStore(peerID, pi); inPingQueue {
		logEntry.Infoln("Peer already in connect queue")
		return nil
	}
	stats.Record(s.ServiceContext(), metrics.PeersToConnectCount.M(float64(s.inConnectQueueCount.Inc())))

	// Schedule connect for peer
	go func() {
		if s.IsStarted() {
			s.connectQueue <- pi
		}
	}()

	return nil
}

func (s *Scheduler) cleanup() {
	s.drainConnectQueue()
	close(s.connectQueue)
	s.shutdownWorkers()
	close(s.resultsQueue)
}

// drainConnectQueue reads all entries from crawlQueue and discards them.
func (s *Scheduler) drainConnectQueue() {
	for {
		select {
		case pi := <-s.connectQueue:
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

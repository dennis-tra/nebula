package monitor

import (
	"context"
	"database/sql"
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
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

// The Scheduler handles the scheduling and managing of
//   a) workers - They consume a queue of peer address information, visit them and publish their results
//                on a separate results queue. This results queue is consumed by this scheduler and further
//                processed
type Scheduler struct {
	// Service represents an entity that runs in a separate go routine and where its lifecycle
	// needs to be handled externally. This is true for this scheduler, so we're embedding it here.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// The database handle
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// The queue of peer.AddrInfo's that need to be dialed to.
	dialQueue *queue.FIFO

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inDialQueue sync.Map

	// The number of peers in the ping queue.
	inDialQueueCount atomic.Uint32

	// The queue that the workers publish their dial results on
	resultsQueue *queue.FIFO

	// The list of worker node references.
	workers sync.Map
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
		dbc:          dbc,
		config:       conf,
		inDialQueue:  sync.Map{},
		dialQueue:    queue.NewFIFO(),
		resultsQueue: queue.NewFIFO(),
		workers:      sync.Map{},
	}

	return m, nil
}

// StartMonitoring starts the configured amount of workers and fills
// the dial queue with peers that are due to be dialed.
func (s *Scheduler) StartMonitoring() error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	s.StartTime = time.Now()

	// Start all workers
	for i := 0; i < s.config.MonitorWorkerCount; i++ {
		w, err := NewWorker(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.workers.Store(i, w)
		go w.StartDialing(s.dialQueue, s.resultsQueue)
	}

	// Async handle the results from workers
	go s.readResultsQueue()

	// Monitor the database and schedule dial jobs
	s.monitorDatabase()

	// release all resources
	s.shutdownWorkers()

	log.WithFields(log.Fields{
		"inDialQueue":     s.inDialQueueCount.Load(),
		"monitorDuration": time.Now().Sub(s.StartTime).String(),
	}).Infoln("Finished monitoring")

	return nil
}

// readResultsQueue listens for dial results on the resultsQueue and handles any
// entries in handleResult. If the scheduler is shut down it schedules a cleanup of resources.
func (s *Scheduler) readResultsQueue() {
	for {
		// Give the shutdown signal precedence
		select {
		case <-s.SigShutdown():
			return
		default:
		}

		select {
		case elem := <-s.resultsQueue.Consume():
			s.handleResult(elem.(Result))
		case <-s.SigShutdown():
			return
		}
	}
}

func (s *Scheduler) handleResult(dr Result) {
	logEntry := log.WithFields(log.Fields{
		"workerID": dr.WorkerID,
		"targetID": dr.Peer.ID.Pretty()[:16],
		"alive":    dr.Error == nil,
	})
	if dr.Error != nil {
		logEntry = logEntry.WithError(dr.Error)
	}

	// Update maps
	s.inDialQueue.Delete(dr.Peer.ID)
	stats.Record(s.ServiceContext(), metrics.PeersToDialCount.M(float64(s.inDialQueueCount.Dec())))

	if err := s.insertRawVisit(s.ServiceContext(), dr); err != nil {
		logEntry.WithError(err).Warnln("Could not persist dial result")
	}
	logEntry.Infoln("Handled dial result from worker", dr.WorkerID)
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (s *Scheduler) monitorDatabase() {
	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := s.dbc.FetchDueSessions(s.ServiceContext())
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

func (s *Scheduler) scheduleDial(session *models.Session) error {
	// Parse peer ID from database
	peerID, err := peer.Decode(session.R.Peer.MultiHash)
	if err != nil {
		return errors.Wrap(err, "decode peer ID")
	}
	logEntry := log.WithField("peerID", peerID.Pretty()[:16])

	// Parse multi addresses from database
	pi := peer.AddrInfo{ID: peerID}
	for _, maddrStr := range session.R.Peer.R.MultiAddresses {
		maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
		if err != nil {
			logEntry.WithError(err).Warnln("Could not parse multi address")
			continue
		}
		pi.Addrs = append(pi.Addrs, maddr)
	}

	// Check if peer is already in dial queue
	if _, inPingQueue := s.inDialQueue.LoadOrStore(peerID, pi); inPingQueue {
		return nil
	}
	stats.Record(s.ServiceContext(), metrics.PeersToDialCount.M(float64(s.inDialQueueCount.Inc())))

	// Schedule dial for peer
	s.dialQueue.Push(pi)

	return nil
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

package monitor

import (
	"context"
	"database/sql"
	"sync"
	"time"

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

type Monitor struct { // Service represents an entity that runs in a
	// separate go routine and where its lifecycle
	// needs to be handled externally.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all workers.
	host host.Host

	// The database handle
	dbh *sql.DB

	// The configuration of timeouts etc.
	config *config.Config

	// The queue of peer.AddrInfo's that need to be pinged.
	pingQueue chan peer.AddrInfo

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inPingQueue sync.Map

	// The number of peers in the ping queue.
	inPingQueueCount atomic.Uint32

	// The queue that the workers publish their ping results on
	resultsQueue chan PingResult

	// A map of peers that were fetched from the database and were scheduled to be put into the ping queue
	duePeers map[peer.ID]*time.Timer

	// The list of worker node references.
	workers []*Worker
}

func NewMonitor(ctx context.Context, dbh *sql.DB) (*Monitor, error) {
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

	m := &Monitor{
		Service:          service.New("monitor"),
		host:             h,
		dbh:              dbh,
		config:           conf,
		inPingQueue:      sync.Map{},
		inPingQueueCount: atomic.Uint32{},
		pingQueue:        make(chan peer.AddrInfo),
		resultsQueue:     make(chan PingResult),
		workers:          []*Worker{},
	}
	return m, nil
}

func (m *Monitor) Shutdown() {
	defer m.Service.Shutdown()

	// drain ping queue
OUTER:
	for {
		select {
		case pi := <-m.pingQueue:
			log.WithField("targetID", pi.ID.Pretty()[:16]).Debugln("Drained peer")
		default:
			break OUTER
		}
	}
	close(m.pingQueue)

	var wg sync.WaitGroup
	for _, w := range m.workers {
		wg.Add(1)
		go func(w *Worker) {
			w.Shutdown()
			wg.Done()
		}(w)
	}
	wg.Wait()

	// After all workers have stopped and won't send any results we can close the results channel.
	close(m.resultsQueue)
}

func (m *Monitor) StartMonitoring() error {
	m.ServiceStarted()
	defer m.ServiceStopped()

	go func() {
		for pr := range m.resultsQueue {
			logEntry := log.WithFields(log.Fields{
				"targetID": pr.Peer.ID.Pretty()[:16],
				"workerID": pr.WorkerID,
				"alive":    pr.Alive,
			})
			logEntry.Infoln("Handling ping result")
			m.inPingQueue.Delete(pr.Peer.ID)
			stats.Record(m.ServiceContext(), metrics.PeersToPingCount.M(float64(m.inPingQueueCount.Dec())))

			var err error
			if pr.Alive {
				logEntry.Traceln("Pinged peer still reachable")
				err = db.UpsertSessionSuccess(m.dbh, pr.Peer.ID.Pretty())
			} else {
				logEntry.Traceln("Pinged peer not reachable anymore")
				err = db.UpsertSessionError(m.dbh, pr.Peer.ID.Pretty())
			}

			if err != nil {
				logEntry.WithError(err).WithField("alive", pr.Alive).Warn("Could not update session record")
			}
		}
	}()

	// Start all workers
	for i := 0; i < m.config.WorkerCount; i++ {
		w, err := NewWorker(m.dbh, m.host, m.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		m.workers = append(m.workers, w)
		go w.StartPinging(m.pingQueue, m.resultsQueue)
	}

	for {
		select {
		case <-m.SigShutdown():
			return nil
		default:
		}

		// Get sessions from the database and do nothing on error, just wait
		log.Infoln("Fetching due sessions from database...")
		dueSessions, err := db.FetchDueSessions(m.ServiceContext(), m.dbh)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch due sessions")
			continue
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Infoln("No sessions due to ping")
			continue
		}
		log.Infof("Found %d due sessions\n", len(dueSessions))

		// Get sessions from the database and do nothing on error, just wait
		log.Infoln("Fetching recently finished sessions from database...")
		recentlyFinishedSessions, err := db.FetchRecentlyGoneSessions(m.ServiceContext(), m.dbh)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch recently finished sessions")
			continue
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Infoln("No recently finished sessions to ping")
			continue
		}
		log.Infof("Found %d recently finished sessions\n", len(recentlyFinishedSessions))

		// For every due session schedule that it gets pushed into the pingQueue
		for _, session := range append(dueSessions, recentlyFinishedSessions...) {

			// Parse peer ID from database
			peerID, err := peer.Decode(session.R.Peer.ID)
			if err != nil {
				log.WithField("peerID", session.R.Peer.ID).WithError(err).Warnln("Could not parse peer ID")
				continue
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

			select {
			case <-m.SigShutdown():
				log.Infoln("Skipping dispatch as monitor shuts down")
				return nil
			default:
				// Check if peer is already in ping queue
				if _, inPingQueue := m.inPingQueue.LoadOrStore(peerID, pi); inPingQueue {
					logEntry.Infoln("Peer already in ping queue")
					continue
				}
				// Schedule ping for peer
				stats.Record(m.ServiceContext(), metrics.PeersToPingCount.M(float64(m.inPingQueueCount.Inc())))

				go func() { m.pingQueue <- pi }()
			}
		}

		select {
		case <-time.Tick(10 * time.Second):
		case <-m.SigShutdown():
			return nil
		}
	}
}

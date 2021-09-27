package monitor

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/volatiletech/null/v8"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

// The Scheduler handles the scheduling and managing of
//   a) dialers - They consume a queue of peer address information, visit them and publish their results
//                on a separate results queue. This results queue is consumed by this scheduler and further
//                processed
type Scheduler struct {
	// Service represents an entity that runs in a separate go routine and where its lifecycle
	// needs to be handled externally. This is true for this scheduler, so we're embedding it here.
	*service.Service

	// The libp2p node that's used to crawl the network. This one is also passed to all dialers.
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

	// The queue that the dialers publish their dial results on
	resultsQueue *queue.FIFO

	// The list of dialer node references.
	dialers sync.Map
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

	// Initialize a single libp2p node that's shared between all dialers.
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
		dialers:      sync.Map{},
	}

	return m, nil
}

// StartMonitoring starts the configured amount of dialers and fills
// the dial queue with peers that are due to be dialed.
func (s *Scheduler) StartMonitoring() error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	s.StartTime = time.Now()

	// Start all dialers
	for i := 0; i < s.config.MonitorWorkerCount; i++ {
		w, err := NewDialer(s.host, s.config)
		if err != nil {
			return errors.Wrap(err, "new dialer")
		}
		s.dialers.Store(i, w)
		go w.StartDialing(s.dialQueue, s.resultsQueue)
	}

	// Async handle the results from dialers
	go s.readResultsQueue()

	// Monitor the database and schedule dial jobs
	s.monitorDatabase()

	// release all resources
	s.shutdownDialers()

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
		"dialerID": dr.DialerID,
		"targetID": dr.Peer.ID.Pretty()[:16],
		"alive":    dr.Error == nil,
	})
	if dr.Error != nil {
		if dr.DialError == models.DialErrorUnknown {
			logEntry = logEntry.WithError(dr.Error)
		} else {
			logEntry = logEntry.WithField("error", dr.DialError)
		}
	}
	start := time.Now()
	if err := s.insertRawVisit(s.ServiceContext(), dr); err != nil {
		logEntry.WithError(err).Warnln("Could not persist dial result")
	}

	// Update maps
	s.inDialQueue.Delete(dr.Peer.ID)
	stats.Record(s.ServiceContext(), metrics.PeersToDialCount.M(float64(s.inDialQueueCount.Dec())))

	// Track dial errors for prometheus
	if dr.Error != nil {
		if ctx, err := tag.New(s.ServiceContext(), tag.Upsert(metrics.KeyError, dr.DialError)); err == nil {
			stats.Record(ctx, metrics.PeersToDialErrorsCount.M(1))
		}
	}

	logEntry.
		WithField("dialDur", dr.DialDuration()).
		WithField("persistDur", time.Since(start)).
		Infoln("Handled dial result from dialer", dr.DialerID)
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

// insertRawVisit builds up a raw_visit database entry.
func (s *Scheduler) insertRawVisit(ctx context.Context, cr Result) error {
	rv := &models.RawVisit{
		VisitStartedAt: cr.DialStartTime,
		VisitEndedAt:   cr.DialEndTime,
		DialDuration:   null.StringFrom(fmt.Sprintf("%f seconds", cr.DialDuration().Seconds())),
		Type:           models.VisitTypeDial,
		PeerMultiHash:  cr.Peer.ID.Pretty(),
		MultiAddresses: maddrsToAddrs(cr.Peer.Addrs),
	}
	if cr.Error != nil {
		rv.Error = null.StringFrom(cr.DialError)
		if len(cr.Error.Error()) > 255 {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error()[:255])
		} else {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error())
		}
	}

	return s.dbc.InsertRawVisit(ctx, rv)
}

// shutdownDialers sends shutdown signals to all dialers and blocks until all have shut down.
func (s *Scheduler) shutdownDialers() {
	var wg sync.WaitGroup
	s.dialers.Range(func(_, dialer interface{}) bool {
		d := dialer.(*Dialer)
		wg.Add(1)
		go func(w *Dialer) {
			w.Shutdown()
			wg.Done()
		}(d)
		return true
	})
	wg.Wait()
}

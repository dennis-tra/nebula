package monitor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	libp2p2 "github.com/dennis-tra/nebula-crawler/pkg/libp2p"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
)

// The Scheduler handles the scheduling and managing of
//
//	a) dialers - They consume a queue of peer address information, visit them and publish their results
//	             on a separate results queue. This results queue is consumed by this scheduler and further
//	             processed
type Scheduler struct {
	// The libp2p node that's used to crawl the network. This one is also passed to all dialers.
	host host.Host

	// The database handle
	dbc *db.DBClient

	// The configuration of timeouts etc.
	config *config.Monitor

	// The queue of peer.AddrInfo's that need to be dialed to.
	dialQueue *queue.FIFO[peer.AddrInfo]

	// A map from peer.ID to peer.AddrInfo to indicate if a peer was put in the queue, so
	// we don't put it there again.
	inDialQueue sync.Map

	// The number of peers in the dial queue.
	inDialQueueCount atomic.Uint32

	// The queue that the dialers publish their dial results on
	resultsQueue *queue.FIFO[Result]
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Monitor, dbc *db.DBClient) (*Scheduler, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Initialize a single libp2p node that's shared between all dialers.
	h, err := libp2p.New(libp2p.NoListenAddrs, libp2p.ResourceManager(rm), libp2p.UserAgent("nebula-monitor/"+conf.Root.Version()))
	if err != nil {
		return nil, err
	}

	s := &Scheduler{
		host:         h,
		dbc:          dbc,
		config:       conf,
		inDialQueue:  sync.Map{},
		dialQueue:    queue.NewFIFO[peer.AddrInfo](),
		resultsQueue: queue.NewFIFO[Result](),
	}

	return s, nil
}

// StartMonitoring starts the configured amount of dialers and fills
// the dial queue with peers that are due to be dialed.
func (s *Scheduler) StartMonitoring(ctx context.Context) error {
	start := time.Now()

	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, s.config.Root.DialTimeout)

	// Force direct dials will prevent swarm to run into dial backoff errors. It also prevents proxied connections.
	ctx = network.WithForceDirectDial(ctx, "prevent backoff")

	// Start all dialers
	var dialers []*libp2p2.Dialer
	for i := 0; i < s.config.MonitorWorkerCount; i++ {
		d, err := libp2p2.NewDialer(s.host, s.config)
		if err != nil {
			return fmt.Errorf("new dialer: %w", err)
		}

		dialers = append(dialers, d)
		go d.StartDialing(ctx, s.dialQueue, s.resultsQueue)
	}

	// Async handle the results from dialers
	go s.readResultsQueue(ctx)

	// Monitor the database and schedule dial jobs
	s.monitorDatabase(ctx)

	for _, d := range dialers {
		log.WithField("dialerId", d.id).Debugln("Waiting for dialer to stop")
		<-d.done
	}

	log.WithFields(log.Fields{
		"inDialQueue":     s.inDialQueueCount.Load(),
		"monitorDuration": time.Since(start),
	}).Infoln("Finished monitoring")

	return nil
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (s *Scheduler) monitorDatabase(ctx context.Context) {
	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := s.dbc.FetchDueOpenSessions(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		// For every session schedule that it gets pushed into the dialQueue
		for _, session := range sessions {
			if err = s.scheduleDial(ctx, session); err != nil {
				log.WithError(err).Warnln("Could not schedule dial")
			}
		}
		log.Infof("In dial queue %d peers", s.inDialQueueCount.Load())

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}

// scheduleDial takes a session entity from the database constructs a peer.AddrInfo struct and feeds
// it into the queue of peers-to-dial to be picked up by one of the dialers.
func (s *Scheduler) scheduleDial(ctx context.Context, session *models.SessionsOpen) error {
	// Parse peer ID from database
	peerID, err := peer.Decode(session.R.Peer.MultiHash)
	if err != nil {
		return fmt.Errorf("decode peer ID: %w", err)
	}
	logEntry := log.WithField("peerID", peerID.ShortString())

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
	if _, inDialQueue := s.inDialQueue.LoadOrStore(peerID, pi); inDialQueue {
		return nil
	}

	// Schedule dial for peer
	s.dialQueue.Push(pi)

	// Track new peer in queue with prometheus
	metrics.VisitQueueLength.With(metrics.DialLabel).Set(float64(s.inDialQueueCount.Inc()))

	return nil
}

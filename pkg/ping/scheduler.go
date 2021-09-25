package ping

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

// The Scheduler handles the scheduling and managing of
//   a) pingers - They consume a queue of peer address information, ping them and publish their results
//                 on a separate results queue. This results queue is consumed by this scheduler and further
//                 processed
type Scheduler struct {
	// Service represents an entity that runs in a separate go routine and where its lifecycle
	// needs to be handled externally. This is true for this scheduler, so we're embedding it here.
	*service.Service

	// The database client
	dbc *db.Client

	// The configuration of timeouts etc.
	config *config.Config

	// Instance of this crawl. This instance gets created right at the beginning of the crawl, so we have
	// an ID that we can link subsequent database entities with.
	crawl *models.Crawl

	// A map from peer.ID to peer.AddrInfo to indicate if a peer has already been pinged
	// in the past, so we don't put them in the crawl queue again.
	inPingQueue int

	// The queue of peer.AddrInfo's that still need to be pinged.
	pingQueue *queue.FIFO

	// The queue that the pingers publish their results on, so that the scheduler can handle them,
	// e.g. update the maps above etc.
	resultsQueue *queue.FIFO

	// The list of worker node references.
	pingers sync.Map
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(ctx context.Context, dbc *db.Client) (*Scheduler, error) {
	conf, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	s := &Scheduler{
		Service:      service.New("scheduler"),
		dbc:          dbc,
		config:       conf,
		pingQueue:    queue.NewFIFO(),
		resultsQueue: queue.NewFIFO(),
		pingers:      sync.Map{},
	}

	return s, nil
}

// PingNetwork starts the configured amount of pingers and fills
// the ping queue with all online peers of the most recent
// successful crawl
func (s *Scheduler) PingNetwork() error {
	s.ServiceStarted()
	defer s.ServiceStopped()

	// Start all pingers
	for i := 0; i < s.config.PingWorkerCount; i++ {
		c, err := NewPinger(s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}
		s.pingers.Store(i, c)
		go c.StartPinging(s.pingQueue, s.resultsQueue)
	}

	// Get most recent successful crawl
	crawl, err := models.Crawls(
		qm.Where(models.CrawlColumns.State+" = ?", models.CrawlStateSucceeded),
		qm.OrderBy(models.CrawlColumns.FinishedAt+" DESC"),
		qm.Limit(1),
	).One(s.ServiceContext(), s.dbc.Handle())
	if err != nil {
		return errors.Wrap(err, "fetching crawl")
	}

	// Get all peers + their multi address from most recent crawl
	peers, err := models.Peers(
		qm.InnerJoin("visits v on v.peer_id = peers.id"),
		qm.Where("v.crawl_id = ? and v.error is null", crawl.ID),
		qm.Load(models.PeerRels.MultiAddresses),
	).All(s.ServiceContext(), s.dbc.Handle())
	if err != nil {
		return err
	}

	for i, p := range peers {

		// Only push peers into the queue until the limit is reached
		if i == s.config.PingLimit && s.config.PingLimit != 0 {
			break
		}

		pi, err := db.ToAddrInfo(p)
		if err != nil {
			return err
		}
		job := Job{pi: pi, dbpeer: p}
		s.pingQueue.Push(job)
		s.inPingQueue++
	}

	log.Infof("Started pinging %d peers...", s.inPingQueue)

	s.readResultsQueue()

	// Indicate that we won't publish any new crawl tasks to the queue.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.pingQueue.DoneProducing()

	// Stop pingers - blocking
	s.shutdownPingers()

	// Indicate that the pingers won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.resultsQueue.DoneProducing()

	return nil
}

// readResultsQueue .
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
			if s.inPingQueue == 0 {
				return
			}
		}
	}
}

// handleResult takes a crawl result, aggregates crawl information and publishes the result
// to the persist queue, so that the persisters can persist the information in the database.
// It also looks into the result and publishes new crawl jobs based on whether the found peers
// weren't pinged before or are not already in the queue.
func (s *Scheduler) handleResult(cr Result) {
	start := time.Now()
	logEntry := log.WithFields(log.Fields{
		"pingerID": cr.PingerID,
		"targetID": cr.Peer.ID.Pretty()[:16],
	})
	logEntry.Debugln("Handling ping result from pinger", cr.PingerID)
	s.inPingQueue -= 1

	txn, err := s.dbc.Handle().BeginTx(s.ServiceContext(), nil)
	if err != nil {
		log.WithError(err).Warnln("Error starting txn")
		return
	}
	for _, latency := range cr.PingLatencies {
		if err := latency.Insert(s.ServiceContext(), s.dbc.Handle(), boil.Infer()); err != nil {
			log.WithError(err).Warnln("Error inserting latency")
		}
	}

	if err = txn.Commit(); err != nil {
		_ = txn.Rollback()
		log.WithError(err).Warnln("Error committing txn")
	}

	logEntry.WithFields(map[string]interface{}{
		"inPingQueue": s.inPingQueue,
		"duration":    time.Since(start),
	}).Infoln("Handled ping result from worker", cr.PingerID)
}

// shutdownPingers sends shutdown signals to all pingers and blocks until all have shut down.
func (s *Scheduler) shutdownPingers() {
	var wg sync.WaitGroup
	s.pingers.Range(func(_, pinger interface{}) bool {
		p := pinger.(*Pinger)
		wg.Add(1)
		go func(w *Pinger) {
			w.Shutdown()
			wg.Done()
		}(p)
		return true
	})
	wg.Wait()
}

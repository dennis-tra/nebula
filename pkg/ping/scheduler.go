package ping

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

// The Scheduler handles the scheduling and managing of
//
//	a) pingers - They consume a queue of peer address information, ping them and publish their results
//	              on a separate results queue. This results queue is consumed by this scheduler and further
//	              processed
type Scheduler struct {
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
}

// NewScheduler initializes a new libp2p host and scheduler instance.
func NewScheduler(conf *config.Config, dbc *db.Client) (*Scheduler, error) {
	s := &Scheduler{
		dbc:          dbc,
		config:       conf,
		pingQueue:    queue.NewFIFO(),
		resultsQueue: queue.NewFIFO(),
	}

	return s, nil
}

// PingNetwork starts the configured amount of pingers and fills
// the ping queue with all online peers of the most recent
// successful crawl
func (s *Scheduler) PingNetwork(ctx context.Context) error {
	// Start all pingers
	var pingers []*Pinger
	for i := 0; i < s.config.PingWorkerCount; i++ {
		p, err := NewPinger(s.config)
		if err != nil {
			return errors.Wrap(err, "new worker")
		}

		pingers = append(pingers, p)
		go p.StartPinging(ctx, s.pingQueue, s.resultsQueue)
	}

	// Get most recent successful crawl
	crawl, err := models.Crawls(
		qm.Where(models.CrawlColumns.State+" = ?", models.CrawlStateSucceeded),
		qm.OrderBy(models.CrawlColumns.FinishedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, s.dbc.Handle())
	if err != nil {
		return errors.Wrap(err, "fetching crawl")
	}

	// Get all peers + their multi address from most recent crawl
	peers, err := models.Peers(
		qm.InnerJoin("visits v on v.peer_id = peers.id"),
		qm.Where("v.crawl_id = ? and v.error is null", crawl.ID),
		qm.Load(models.PeerRels.MultiAddresses),
	).All(ctx, s.dbc.Handle())
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

	s.readResultsQueue(ctx)

	// Indicate that we won't publish any new ping tasks to the queue.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.pingQueue.DoneProducing()

	// Stop pingers - blocking
	for _, p := range pingers {
		log.WithField("pingerId", p.id).Debugln("Waiting for pinger to stop")
		<-p.done
	}

	// Indicate that the pingers won't send any new results as they are now stopped.
	// TODO: This can still leak a Go routine. However we're exiting here anyway...
	s.resultsQueue.DoneProducing()

	return nil
}

// readResultsQueue consumes elements from the results queue that were published by the pingers. This endless for
// loop stops if the ping queue has no entries anymore or the context was cancelled.
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
		case elem, ok := <-s.resultsQueue.Consume():
			if !ok {
				return
			}
			s.handleResult(ctx, elem.(Result))
			if s.inPingQueue == 0 {
				return
			}
		}
	}
}

// handleResult takes a ping result and saves the latencies to the database.
func (s *Scheduler) handleResult(ctx context.Context, cr Result) {
	start := time.Now()
	logEntry := log.WithFields(log.Fields{
		"pingerID": cr.PingerID,
		"remoteID": utils.FmtPeerID(cr.Peer.ID),
	})
	logEntry.Debugln("Handling ping result from pinger", cr.PingerID)
	s.inPingQueue -= 1

	txn, err := s.dbc.Handle().BeginTx(ctx, nil)
	if err != nil {
		log.WithError(err).Warnln("Error starting txn")
		return
	}
	for _, latency := range cr.PingLatencies {
		if err := latency.Insert(ctx, s.dbc.Handle(), boil.Infer()); err != nil {
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

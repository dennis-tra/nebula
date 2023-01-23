package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

var persisterID = atomic.NewInt32(0)

// Persister handles the insert/upsert/update operations for a particular crawl result.
type Persister struct {
	id             string
	config         *config.Config
	dbc            *db.Client
	crawl          *models.Crawl
	persistedPeers int
	done           chan struct{}
}

// NewPersister initializes a new persister based on the given configuration.
func NewPersister(dbc *db.Client, conf *config.Config, crawl *models.Crawl) (*Persister, error) {
	p := &Persister{
		id:             fmt.Sprintf("persister-%02d", persisterID.Inc()),
		config:         conf,
		dbc:            dbc,
		crawl:          crawl,
		persistedPeers: 0,
		done:           make(chan struct{}),
	}
	persisterID.Inc()

	return p, nil
}

// StartPersisting enters an endless loop and consumes persist jobs from the persist queue
// until it is told to stop or the persist queue was closed.
func (p *Persister) StartPersisting(ctx context.Context, persistQueue *queue.FIFO[Result], resultsQueue *queue.FIFO[*db.InsertVisitResult]) {
	defer close(p.done)
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
		case r, ok := <-persistQueue.Consume():
			if !ok {
				// The persist queue was closed
				return
			}

			ivr := p.handlePersistJob(ctx, r)
			resultsQueue.Push(ivr)
		}
	}
}

// handlePersistJob takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (p *Persister) handlePersistJob(ctx context.Context, cr Result) *db.InsertVisitResult {
	logEntry := log.WithFields(log.Fields{
		"persisterID": p.id,
		"remoteID":    utils.FmtPeerID(cr.Peer.ID),
	})
	logEntry.Debugln("Persisting peer")
	defer logEntry.Debugln("Persisted peer")

	start := time.Now()

	ivr, err := p.insertVisit(ctx, cr)
	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logEntry.WithError(err).Warnln("Error inserting raw visit")
	} else {
		p.persistedPeers++
	}
	logEntry.
		WithField("persisted", p.persistedPeers).
		WithField("success", err == nil).
		WithField("duration", time.Since(start)).
		Infoln("Persisted result from worker", cr.CrawlerID)
	return ivr
}

// insertVisit builds up a visit database entry.
func (p *Persister) insertVisit(ctx context.Context, cr Result) (*db.InsertVisitResult, error) {
	return p.dbc.PersistCrawlVisit(
		ctx,
		p.dbc.Handle(),
		p.crawl.ID,
		cr.Peer.ID,
		cr.Peer.Addrs,
		cr.Protocols,
		cr.Agent,
		cr.ConnectDuration(),
		cr.CrawlDuration(),
		cr.CrawlStartTime,
		cr.CrawlEndTime,
		cr.ConnectErrorStr,
		cr.CrawlErrorStr,
		cr.IsExposed,
	)
}

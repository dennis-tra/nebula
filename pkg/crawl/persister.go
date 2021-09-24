package crawl

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/queue"
	"github.com/dennis-tra/nebula-crawler/pkg/service"
)

var persisterID = atomic.NewInt32(0)

// Persister handles the insert/upsert/update operations for a particular crawl result.
// We're doing it asynchronously as each insert can take multiple tens of milliseconds.
// This would take too long to do synchronously during a crawl.
type Persister struct {
	*service.Service

	config         *config.Config
	dbc            *db.Client
	crawl          *models.Crawl
	persistedPeers int
}

// NewPersister initializes a new persister based on the given configuration.
func NewPersister(dbc *db.Client, conf *config.Config, crawl *models.Crawl) (*Persister, error) {
	p := &Persister{
		Service: service.New(fmt.Sprintf("persister-%02d", persisterID.Load())),
		config:  conf,
		dbc:     dbc,
		crawl:   crawl,
	}
	persisterID.Inc()
	return p, nil
}

// StartPersisting enters an endless loop and consumes persist jobs from the persist queue
// until it is told to stop or the persist queue was closed.
func (p *Persister) StartPersisting(persistQueue *queue.FIFO) {
	p.ServiceStarted()
	defer p.ServiceStopped()
	ctx := p.ServiceContext()

	for {
		// Give the shutdown signal precedence
		select {
		case <-p.SigShutdown():
			return
		default:
		}

		select {
		case elem, ok := <-persistQueue.Consume():
			if !ok {
				// The persist queue was closed
				return
			}
			p.handlePersistJob(ctx, elem.(Result))
		case <-p.SigShutdown():
			return
		}
	}
}

// handlePersistJob takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (p *Persister) handlePersistJob(ctx context.Context, cr Result) {
	logEntry := log.WithFields(log.Fields{
		"persisterID": p.Identifier(),
		"targetID":    cr.Peer.ID.Pretty()[:16],
	})
	logEntry.Debugln("Persisting peer")
	defer logEntry.Debugln("Persisted peer")

	start := time.Now()

	err := p.insertRawVisit(ctx, cr)
	if err != nil {
		logEntry.WithError(err).WithField("result", cr).Warnln("Error inserting raw visit")
	} else {
		p.persistedPeers++
	}
	logEntry.
		WithField("persisted", p.persistedPeers).
		WithField("success", err == nil).
		WithField("duration", time.Since(start)).
		Infoln("Persisted result from worker", cr.CrawlerID)
}

// insertRawVisit builds up a raw_visit database entry.
func (p *Persister) insertRawVisit(ctx context.Context, cr Result) error {
	rv := &models.RawVisit{
		CrawlID:         null.IntFrom(p.crawl.ID),
		VisitStartedAt:  cr.CrawlStartTime,
		VisitEndedAt:    cr.CrawlEndTime,
		ConnectDuration: null.StringFrom(fmt.Sprintf("%f seconds", cr.ConnectDuration().Seconds())),
		CrawlDuration:   null.StringFrom(fmt.Sprintf("%f seconds", cr.CrawlDuration().Seconds())),
		PeerMultiHash:   cr.Peer.ID.Pretty(),
		Protocols:       cr.Protocols,
		MultiAddresses:  maddrsToAddrs(cr.Peer.Addrs),
		Type:            models.VisitTypeCrawl,
	}
	if cr.Agent != "" {
		rv.AgentVersion = null.StringFrom(cr.Agent)
	}
	if cr.Error != nil {
		rv.Error = null.StringFrom(cr.DialError)
		if len(cr.Error.Error()) > 255 {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error()[:255])
		} else {
			rv.ErrorMessage = null.StringFrom(cr.Error.Error())
		}
	}

	return p.dbc.InsertRawVisit(ctx, rv)
}

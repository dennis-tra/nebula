package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
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
	persistedPeers int
}

// NewPersister initializes a new persister based on the given configuration.
func NewPersister(dbc *db.Client, conf *config.Config) (*Persister, error) {
	p := &Persister{
		Service: service.New(fmt.Sprintf("persister-%02d", persisterID.Load())),
		config:  conf,
		dbc:     dbc,
	}
	persisterID.Inc()
	return p, nil
}

// StartPersisting TODO
func (p *Persister) StartPersisting(persistQueue *queue.FIFO) {
	p.ServiceStarted()
	defer p.ServiceStopped()

	ctx := p.ServiceContext()
	logEntry := log.WithField("persisterID", p.Identifier())
	for {
		var cr Result
		select {
		case elem, ok := <-persistQueue.Consume():
			if !ok {
				// The persist queue was closed
				return
			}
			cr = elem.(Result)
		case <-p.SigShutdown():
			return
		}

		logEntry = logEntry.WithField("targetID", cr.Peer.ID.Pretty()[:16])
		logEntry.Debugln("Persisting peer")

		var err error
		if err = p.persistCrawlResult(ctx, cr); err != nil {
			logEntry.WithError(err).Warnln("Error persisting crawl result")
		} else {
			logEntry.Debugln("Persisted peer")
		}
		logEntry.WithField("persisted", p.persistedPeers).WithField("success", err == nil).Infoln("Persisted crawl result from worker", cr.WorkerID)
	}
}

// persistCrawlResult inserts the given peer with its multi addresses in the database and
// upserts its currently active session
func (p *Persister) persistCrawlResult(ctx context.Context, cr Result) error {
	var err error

	dbPeer, err := p.dbc.UpsertPeer(ctx, cr.Peer, cr.Agent)
	if err != nil {
		return errors.Wrap(err, "upsert peer")
	}

	startUpsert := time.Now()
	if cr.Error == nil {
		if err := p.dbc.UpsertSessionSuccess(dbPeer); err != nil {
			return errors.Wrap(err, "upsert session success")
		}
	} else if cr.Error != ctx.Err() {
		if err := p.dbc.UpsertSessionError(dbPeer, cr.ErrorTime, determineDialError(cr.Error)); err != nil {
			return errors.Wrap(err, "upsert session error")
		}
	}
	stats.Record(ctx, metrics.CrawledUpsertDuration.M(millisSince(startUpsert)))

	// Persist latency measurements
	if cr.Latencies != nil {
		if err := p.dbc.InsertLatencies(ctx, dbPeer, cr.Latencies); err != nil {
			return errors.Wrap(err, "insert latencies")
		}
	}

	p.persistedPeers++

	return nil
}

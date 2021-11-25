package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"go.opencensus.io/stats"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
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
	agentVersions  map[string]*models.AgentVersion
	protocols      map[string]*models.Protocol
}

// NewPersister initializes a new persister based on the given configuration.
func NewPersister(dbc *db.Client, conf *config.Config, crawl *models.Crawl, avs map[string]*models.AgentVersion, prot map[string]*models.Protocol) (*Persister, error) {
	p := &Persister{
		id:             fmt.Sprintf("persister-%02d", persisterID.Inc()),
		config:         conf,
		dbc:            dbc,
		crawl:          crawl,
		persistedPeers: 0,
		done:           make(chan struct{}),
		agentVersions:  avs,
		protocols:      prot,
	}
	persisterID.Inc()

	return p, nil
}

// StartPersisting enters an endless loop and consumes persist jobs from the persist queue
// until it is told to stop or the persist queue was closed.
func (p *Persister) StartPersisting(ctx context.Context, persistQueue *queue.FIFO) {
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
		case elem, ok := <-persistQueue.Consume():
			if !ok {
				// The persist queue was closed
				return
			}
			p.handlePersistJob(ctx, elem.(Result))
		}
	}
}

// handlePersistJob takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (p *Persister) handlePersistJob(ctx context.Context, cr Result) {
	logEntry := log.WithFields(log.Fields{
		"persisterID": p.id,
		"remoteID":    utils.FmtPeerID(cr.Peer.ID),
	})
	logEntry.Debugln("Persisting peer")
	defer logEntry.Debugln("Persisted peer")

	start := time.Now()

	err := p.insertRawVisit(ctx, cr)
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
}

func (p *Persister) agentVersionID(agent string) *int {
	av, found := p.agentVersions[agent]
	if !found {
		return nil
	}
	id := av.ID
	return &id
}

func (p *Persister) parseProtocols(ctx context.Context, protocols []string) (types.StringArray, types.Int64Array) {
	var protocolStrs []string
	var protocolIDs []int64
	for _, protocol := range protocols {
		if p, found := p.protocols[protocol]; found {
			stats.Record(ctx, metrics.ProtocolCacheHitCount.M(1))
			protocolIDs = append(protocolIDs, int64(p.ID))
		} else {
			stats.Record(ctx, metrics.ProtocolCacheMissCount.M(1))
			protocolStrs = append(protocolStrs, protocol)
		}
	}
	return protocolStrs, protocolIDs
}

// insertRawVisit builds up a raw_visit database entry.
func (p *Persister) insertRawVisit(ctx context.Context, cr Result) error {
	protocolStrs, protocolIDs := p.parseProtocols(ctx, cr.Protocols)
	rv := &models.RawVisit{
		CrawlID:         null.IntFrom(p.crawl.ID),
		VisitStartedAt:  cr.CrawlStartTime,
		VisitEndedAt:    cr.CrawlEndTime,
		ConnectDuration: null.StringFrom(fmt.Sprintf("%f seconds", cr.ConnectDuration().Seconds())),
		CrawlDuration:   null.StringFrom(fmt.Sprintf("%f seconds", cr.CrawlDuration().Seconds())),
		PeerMultiHash:   cr.Peer.ID.Pretty(),
		Protocols:       protocolStrs,
		ProtocolIds:     protocolIDs,
		MultiAddresses:  utils.MaddrsToAddrs(cr.Peer.Addrs),
		Type:            models.VisitTypeCrawl,
	}
	if cr.Agent != "" {
		avID := p.agentVersionID(cr.Agent)
		if avID != nil {
			stats.Record(ctx, metrics.AgentVersionCacheHitCount.M(1))
			rv.AgentVersionID = null.IntFromPtr(avID)
		} else {
			stats.Record(ctx, metrics.AgentVersionCacheMissCount.M(1))
			rv.AgentVersion = null.StringFrom(cr.Agent)
		}
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

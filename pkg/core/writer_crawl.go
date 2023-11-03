package core

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

type CrawlWriterConfig struct {
	AddrTrackType config.AddrType
}

// CrawlWriter handles the insert/upsert/update operations for a particular crawl result.
type CrawlWriter[I PeerInfo[I]] struct {
	id           string
	cfg          *CrawlWriterConfig
	dbc          db.Client
	dbCrawlID    int
	writtenPeers int
}

func NewCrawlWriter[I PeerInfo[I]](id string, dbc db.Client, dbCrawlID int, cfg *CrawlWriterConfig) *CrawlWriter[I] {
	return &CrawlWriter[I]{
		id:           id,
		cfg:          cfg,
		dbc:          dbc,
		dbCrawlID:    dbCrawlID,
		writtenPeers: 0,
	}
}

// Work takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (w *CrawlWriter[I]) Work(ctx context.Context, task CrawlResult[I]) (WriteResult, error) {
	if _, ok := w.dbc.(*db.NoopClient); ok {
		return WriteResult{
			InsertVisitResult: &db.InsertVisitResult{},
		}, nil
	}

	logEntry := log.WithFields(log.Fields{
		"writerID": w.id,
		"remoteID": task.Info.ID().ShortString(),
	})
	logEntry.Debugln("Storing peer")
	defer logEntry.Debugln("Stored peer")

	maddrs := task.Info.Addrs()
	switch w.cfg.AddrTrackType {
	case config.AddrTypePrivate:
		maddrs = utils.FilterPublicMaddrs(maddrs)
	case config.AddrTypePublic:
		maddrs = utils.FilterPrivateMaddrs(maddrs)
	default:
		// noop
	}

	start := time.Now()
	ivr, err := w.dbc.PersistCrawlVisit(
		ctx,
		w.dbCrawlID,
		task.Info.ID(),
		maddrs,
		task.Protocols,
		task.Agent,
		task.ConnectDuration(),
		task.CrawlDuration(),
		task.CrawlStartTime,
		task.CrawlEndTime,
		task.ConnectErrorStr,
		task.CrawlErrorStr,
		null.JSONFrom(task.Properties),
	)
	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logEntry.WithError(err).Warnln("Error inserting raw visit")
	} else {
		w.writtenPeers++
	}

	return WriteResult{
		InsertVisitResult: ivr,
		WriterID:          w.id,
		PeerID:            task.Info.ID(),
		Duration:          time.Since(start),
		Error:             err,
	}, nil
}

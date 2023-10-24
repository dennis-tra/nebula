package core

import (
	"context"
	"errors"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
	log "github.com/sirupsen/logrus"
)

type WriteResult struct {
	*db.InsertVisitResult
}

// Writer handles the insert/upsert/update operations for a particular crawl result.
type Writer[I PeerInfo] struct {
	id           string
	dbc          db.Client
	dbCrawlID    int
	writtenPeers int
}

var _ Worker[CrawlResult[PeerInfo], WriteResult] = (*Writer[PeerInfo])(nil)

func NewWriter[I PeerInfo](id string, dbc db.Client, dbCrawlID int) *Writer[I] {
	return &Writer[I]{
		id:           id,
		dbc:          dbc,
		dbCrawlID:    dbCrawlID,
		writtenPeers: 0,
	}
}

// Work takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (w *Writer[I]) Work(ctx context.Context, task CrawlResult[I]) (WriteResult, error) {
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

	start := time.Now()

	ivr, err := w.insertVisit(ctx, task)
	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logEntry.WithError(err).Warnln("Error inserting raw visit")
	} else {
		w.writtenPeers++
	}
	logEntry.
		WithField("persisted", w.writtenPeers).
		WithField("success", err == nil).
		WithField("duration", time.Since(start)).
		Infoln("Persisted result from worker", w.id)

	return WriteResult{InsertVisitResult: ivr}, nil
}

// insertVisit builds up a visit database entry.
func (w *Writer[I]) insertVisit(ctx context.Context, cr CrawlResult[I]) (*db.InsertVisitResult, error) {
	return w.dbc.PersistCrawlVisit(
		ctx,
		w.dbCrawlID,
		cr.Info.ID(),
		cr.Info.Addrs(),
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

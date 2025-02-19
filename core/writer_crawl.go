package core

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type CrawlWriterConfig struct {
	AddrTrackType config.AddrType
}

// CrawlWriter handles the insert/upsert/update operations for a particular crawl result.
type CrawlWriter[I PeerInfo[I]] struct {
	id           string
	cfg          *CrawlWriterConfig
	dbc          db.Client
	writtenPeers int
}

func NewCrawlWriter[I PeerInfo[I]](id string, dbc db.Client, cfg *CrawlWriterConfig) *CrawlWriter[I] {
	return &CrawlWriter[I]{
		id:           id,
		cfg:          cfg,
		dbc:          dbc,
		writtenPeers: 0,
	}
}

// Work takes a crawl result (persist job) and inserts a denormalized database entry of the results.
func (w *CrawlWriter[I]) Work(ctx context.Context, task CrawlResult[I]) (WriteResult, error) {
	if _, ok := w.dbc.(*db.NoopClient); ok {
		return WriteResult{}, nil
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

	var (
		errorBits uint16
		neighbors []peer.ID
	)
	if task.RoutingTable != nil {
		neighbors = make([]peer.ID, len(task.RoutingTable.Neighbors))
		for i, n := range task.RoutingTable.Neighbors {
			neighbors[i] = n.ID()
		}
		errorBits = task.RoutingTable.ErrorBits
	}

	args := &db.VisitArgs{
		PeerID:          task.Info.ID(),
		Protocols:       task.Protocols,
		AgentVersion:    task.Agent,
		DialMaddrs:      task.DialMaddrs,
		FilteredMaddrs:  task.FilteredMaddrs,
		ExtraMaddrs:     task.ExtraMaddrs,
		ConnectMaddr:    task.ConnectMaddr,
		DialErrors:      task.DialErrors,
		ConnectDuration: task.ConnectDuration(),
		CrawlDuration:   task.CrawlDuration(),
		VisitStartedAt:  task.CrawlStartTime,
		VisitEndedAt:    task.CrawlEndTime,
		ConnectErrorStr: task.ConnectErrorStr,
		CrawlErrorStr:   task.CrawlErrorStr,
		VisitType:       db.VisitTypeCrawl,
		Neighbors:       neighbors,
		ErrorBits:       errorBits,
		Properties:      task.Properties,
	}

	start := time.Now()
	err := w.dbc.InsertVisit(ctx, args)
	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		logEntry.WithError(err).Warnln("Error inserting raw visit")
	} else {
		w.writtenPeers++
	}

	return WriteResult{
		WriterID: w.id,
		PeerID:   task.Info.ID(),
		Duration: time.Since(start),
		Error:    err,
	}, nil
}

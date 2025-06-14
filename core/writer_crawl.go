package core

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/db"
)

type CrawlWriterConfig struct{}

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

	var (
		errorBits        uint16
		neighbors        []peer.ID
		neighborPrefixes []uint64
	)
	if task.RoutingTable != nil {
		neighbors = make([]peer.ID, len(task.RoutingTable.Neighbors))
		neighborPrefixes = make([]uint64, len(task.RoutingTable.Neighbors))
		for i, n := range task.RoutingTable.Neighbors {
			neighbors[i] = n.ID()
			neighborPrefixes[i] = n.DiscoveryPrefix()
		}
		errorBits = task.RoutingTable.ErrorBits
	}

	args := &db.VisitArgs{
		PeerID:           task.Info.ID(),
		DiscoveryPrefix:  task.Info.DiscoveryPrefix(),
		Protocols:        task.Protocols,
		AgentVersion:     task.Agent,
		DialMaddrs:       task.DialMaddrs,
		FilteredMaddrs:   task.FilteredMaddrs,
		ExtraMaddrs:      task.ExtraMaddrs,
		ListenMaddrs:     task.ListenMaddrs,
		ConnectMaddr:     task.ConnectMaddr,
		DialErrors:       task.DialErrors,
		ConnectDuration:  task.ConnectDuration(),
		CrawlDuration:    task.CrawlDuration(),
		VisitStartedAt:   task.CrawlStartTime,
		VisitEndedAt:     task.CrawlEndTime,
		ConnectErrorStr:  task.ConnectErrorStr,
		CrawlErrorStr:    task.CrawlErrorStr,
		VisitType:        db.VisitTypeCrawl,
		Neighbors:        neighbors,
		NeighborPrefixes: neighborPrefixes,
		ErrorBits:        errorBits,
		Properties:       task.Properties,
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

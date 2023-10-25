package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/discv5"
	"github.com/dennis-tra/nebula-crawler/pkg/libp2p"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

type Monitor struct {
	cfg *config.Monitor
	dbc db.Client
}

func New(dbc db.Client, cfg *config.Monitor) (*Monitor, error) {
	return &Monitor{
		cfg: cfg,
		dbc: dbc,
	}, nil
}

func (c *Monitor) MonitorNetwork(ctx context.Context) error {
	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	dbMonitor, err := c.dbc.InitMonitor(ctx)
	if err != nil {
		return fmt.Errorf("creating crawl in db: %w", err)
	}

	switch c.cfg.Network {
	case string(config.NetworkEthereum):
		stackCfg := &discv5.StackConfig{
			TrackNeighbors:    c.cfg.PersistNeighbors,
			BootstrapPeerStrs: c.cfg.BootstrapPeers.Value(),
		}

		stack, err := discv5.NewStack(c.dbc, dbMonitor, stackCfg)
		if err != nil {
			return fmt.Errorf("new stack: %w", err)
		}

		engineCfg := &core.EngineConfig{
			MonitorerCount: c.cfg.MonitorWorkerCount,
			WriterCount:    c.cfg.WriteWorkerCount,
			Limit:          c.cfg.MonitorLimit,
			TrackNeighbors: c.cfg.PersistNeighbors,
		}

		eng, err := core.NewEngine[discv5.PeerInfo](stack, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		runData, err := eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

		_ = runData

		return nil
	default:
		stackCfg := &libp2p.StackConfig{
			Version:           c.cfg.Root.Version(),
			Protocols:         c.cfg.Protocols.Value(),
			DialTimeout:       c.cfg.Root.DialTimeout,
			TrackNeighbors:    c.cfg.PersistNeighbors,
			CheckExposed:      c.cfg.CheckExposed,
			BootstrapPeerStrs: c.cfg.BootstrapPeers.Value(),
		}

		stack, err := libp2p.NewStack(c.dbc, dbMonitor, stackCfg)
		if err != nil {
			return fmt.Errorf("new stack: %w", err)
		}

		engineCfg := &core.EngineConfig{
			MonitorerCount: c.cfg.MonitorWorkerCount,
			WriterCount:    c.cfg.WriteWorkerCount,
			Limit:          c.cfg.MonitorLimit,
			TrackNeighbors: c.cfg.PersistNeighbors,
		}

		eng, err := core.NewEngine[libp2p.PeerInfo](stack, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		runData, err := eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

		// construct a new cleanup context
		cleanupCtx := ctx
		if ctx.Err() != nil {
			var cancel context.CancelFunc
			cleanupCtx, cancel = context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
		}

		// Persist the crawl results
		if err := updateMonitor(cleanupCtx, c.dbc, dbMonitor, ctx, runData); err != nil {
			return fmt.Errorf("persist crawl: %w", err)
		}

	}

	return nil
}

// updateMonitor writes crawl statistics to the database
func updateMonitor[I core.PeerInfo](ctx context.Context, dbc db.Client, dbMonitor *models.Monitor, crawlCtx context.Context, runData *core.RunData[I]) error {
	log.Infoln("Persisting crawl result...")

	dbMonitor.FinishedAt = null.TimeFrom(time.Now())
	dbMonitor.MonitoredPeers = null.IntFrom(runData.MonitoredPeers)
	dbMonitor.DialablePeers = null.IntFrom(runData.MonitoredPeers - runData.TotalErrors())
	dbMonitor.UndialablePeers = null.IntFrom(runData.TotalErrors())

	if runData.QueuedPeers == 0 {
		dbMonitor.State = models.MonitorStateSucceeded
	} else if errors.Is(crawlCtx.Err(), context.Canceled) {
		dbMonitor.State = models.MonitorStateCancelled
	} else {
		dbMonitor.State = models.MonitorStateFailed
	}

	return dbc.UpdateMonitor(ctx, dbMonitor)
}

// storeNeighbors fills the neighbors table with topology information
func storeNeighbors[I core.PeerInfo](ctx context.Context, dbc db.Client, dbMonitor *models.Monitor, runData *core.RunData[I]) {
	log.Infoln("Persisting neighbor information...")

	start := time.Now()
	neighborsCount := 0
	i := 0
	for p, routingTable := range runData.RoutingTables {
		if i%100 == 0 && i > 0 {
			log.Infof("Persisted %d peers and their neighbors", i)
		}
		i++
		neighborsCount += len(routingTable.Neighbors)

		var dbPeerID *int
		if id, found := runData.PeerMappings[p]; found {
			dbPeerID = &id
		}

		dbPeerIDs := []int{}
		peerIDs := []peer.ID{}
		for _, n := range routingTable.Neighbors {
			if id, found := runData.PeerMappings[n.ID()]; found {
				dbPeerIDs = append(dbPeerIDs, id)
			} else {
				peerIDs = append(peerIDs, n.ID())
			}
		}
		if err := dbc.PersistNeighbors(ctx, dbMonitor, dbPeerID, p, routingTable.ErrorBits, dbPeerIDs, peerIDs); err != nil {
			log.WithError(err).WithField("peerID", p.ShortString()).Warnln("Could not persist neighbors")
		}
	}
	log.WithFields(log.Fields{
		"duration":       time.Since(start),
		"avg":            fmt.Sprintf("%.2fms", time.Since(start).Seconds()/float64(len(runData.RoutingTables))*1000),
		"peers":          len(runData.RoutingTables),
		"totalNeighbors": neighborsCount,
	}).Infoln("Finished persisting neighbor information")
}

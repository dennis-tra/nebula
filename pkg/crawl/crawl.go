package crawl

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

type Crawl struct {
	cfg *config.Crawl
	dbc db.Client
}

func New(dbc db.Client, cfg *config.Crawl) (*Crawl, error) {
	return &Crawl{
		cfg: cfg,
		dbc: dbc,
	}, nil
}

func (c *Crawl) CrawlNetwork(ctx context.Context) error {
	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	dbCrawl, err := c.dbc.InitCrawl(ctx)
	if err != nil {
		return fmt.Errorf("creating crawl in db: %w", err)
	}

	handlerCfg := &core.CrawlHandlerConfig{
		TrackNeighbors: c.cfg.PersistNeighbors,
	}

	engineCfg := &core.EngineConfig{
		WorkerCount: c.cfg.CrawlWorkerCount,
		WriterCount: c.cfg.WriteWorkerCount,
		Limit:       c.cfg.CrawlLimit,
	}

	switch c.cfg.Network {
	case string(config.NetworkEthereum):
		driverCfg := &discv5.DriverConfig{
			TrackNeighbors:    c.cfg.PersistNeighbors,
			BootstrapPeerStrs: c.cfg.BootstrapPeers.Value(),
		}

		driver, err := discv5.NewCrawlDriver(c.dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewCrawlHandler[discv5.PeerInfo](handlerCfg)

		eng, err := core.NewEngine[discv5.PeerInfo, core.CrawlResult[discv5.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		queuedPeers, err := eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

		handler.QueuedPeers = len(queuedPeers)

		if err := persistCrawlInformation(ctx, c.dbc, dbCrawl, handler); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}

		return nil
	default:
		driverCfg := &libp2p.CrawlDriverConfig{
			Version:           c.cfg.Root.Version(),
			Protocols:         c.cfg.Protocols.Value(),
			DialTimeout:       c.cfg.Root.DialTimeout,
			TrackNeighbors:    c.cfg.PersistNeighbors,
			CheckExposed:      c.cfg.CheckExposed,
			BootstrapPeerStrs: c.cfg.BootstrapPeers.Value(),
		}

		driver, err := libp2p.NewCrawlDriver(c.dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewCrawlHandler[libp2p.PeerInfo](handlerCfg)
		eng, err := core.NewEngine[libp2p.PeerInfo, core.CrawlResult[libp2p.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		queuedPeers, err := eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

		handler.QueuedPeers = len(queuedPeers)

		if err := persistCrawlInformation(ctx, c.dbc, dbCrawl, handler); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}
	}

	return nil
}

func persistCrawlInformation[I core.PeerInfo](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) error {
	// construct a new cleanup context to store the crawl results even
	// if the user cancelled the process.
	sigs := make(chan os.Signal, 1)
	cleanupCtx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Infof("Received %s signal - Stopping...\n", sig.String())
		signal.Stop(sigs)
		cancel()
	}()

	// Persist the crawl results
	if err := updateCrawl(cleanupCtx, dbc, dbCrawl, ctx, handler); err != nil {
		return fmt.Errorf("persist crawl: %w", err)
	}

	// Persist associated crawl properties
	if err := persistCrawlProperties(cleanupCtx, dbc, dbCrawl, handler); err != nil {
		return fmt.Errorf("persist crawl properties: %w", err)
	}

	// persist all neighbor information
	if err := storeNeighbors(cleanupCtx, dbc, dbCrawl, handler); err != nil {
		return fmt.Errorf("store neighbors: %w", err)
	}

	logSummary(dbCrawl, handler)

	return nil
}

// updateCrawl writes crawl statistics to the database
func updateCrawl[I core.PeerInfo](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, crawlCtx context.Context, handler *core.CrawlHandler[I]) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	log.Infoln("Persisting crawl result...")

	dbCrawl.FinishedAt = null.TimeFrom(time.Now())
	dbCrawl.CrawledPeers = null.IntFrom(handler.CrawledPeers)
	dbCrawl.DialablePeers = null.IntFrom(handler.CrawledPeers - handler.TotalErrors())
	dbCrawl.UndialablePeers = null.IntFrom(handler.TotalErrors())

	if handler.QueuedPeers == 0 {
		dbCrawl.State = models.CrawlStateSucceeded
	} else if errors.Is(crawlCtx.Err(), context.Canceled) {
		dbCrawl.State = models.CrawlStateCancelled
	} else {
		dbCrawl.State = models.CrawlStateFailed
	}

	return dbc.UpdateCrawl(ctx, dbCrawl)
}

// persistCrawlProperties writes crawl property statistics to the database.
func persistCrawlProperties[I core.PeerInfo](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	log.Infoln("Persisting crawl properties...")

	// Extract full and core agent versionc. Core agent versions are just strings like 0.8.0 or 0.5.0
	// The full agent versions have much more information e.g., /go-ipfs/0.4.21-dev/789dab3
	avFull := map[string]int{}
	for version, count := range handler.AgentVersion {
		avFull[version] += count
	}
	pps := map[string]map[string]int{
		"agent_version": avFull,
		"protocol":      handler.Protocols,
		"error":         handler.CrawlErrs,
	}

	return dbc.PersistCrawlProperties(ctx, dbCrawl, pps)
}

// storeNeighbors fills the neighbors table with topology information
func storeNeighbors[I core.PeerInfo](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	if len(handler.RoutingTables) == 0 {
		return nil
	}

	log.Infoln("Persisting neighbor information...")

	start := time.Now()
	neighborsCount := 0
	i := 0
	for p, routingTable := range handler.RoutingTables {
		if i%100 == 0 && i > 0 {
			log.Infof("Persisted %d peers and their neighbors", i)
		}
		i++
		neighborsCount += len(routingTable.Neighbors)

		var dbPeerID *int
		if id, found := handler.PeerMappings[p]; found {
			dbPeerID = &id
		}

		dbPeerIDs := []int{}
		peerIDs := []peer.ID{}
		for _, n := range routingTable.Neighbors {
			if id, found := handler.PeerMappings[n.ID()]; found {
				dbPeerIDs = append(dbPeerIDs, id)
			} else {
				peerIDs = append(peerIDs, n.ID())
			}
		}
		if err := dbc.PersistNeighbors(ctx, dbCrawl, dbPeerID, p, routingTable.ErrorBits, dbPeerIDs, peerIDs); err != nil {
			return fmt.Errorf("persiting neighbor information: %w", err)
		}
	}
	log.WithFields(log.Fields{
		"duration":       time.Since(start),
		"avg":            fmt.Sprintf("%.2fms", time.Since(start).Seconds()/float64(len(handler.RoutingTables))*1000),
		"peers":          len(handler.RoutingTables),
		"totalNeighbors": neighborsCount,
	}).Infoln("Finished persisting neighbor information")
	return nil
}

// logSummary logs the final results of the crawl.
func logSummary[I core.PeerInfo](dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) {
	log.Infoln("Logging crawl results...")

	log.Infoln("")
	for err, count := range handler.ConnErrs {
		log.WithField("count", count).WithField("value", err).Infoln("Dial Error")
	}

	log.Infoln("")
	for err, count := range handler.CrawlErrs {
		log.WithField("count", count).WithField("value", err).Infoln("Crawl Error")
	}

	log.Infoln("")
	for agent, count := range handler.AgentVersion {
		log.WithField("count", count).WithField("value", agent).Infoln("Agent")
	}
	log.Infoln("")
	for protocol, count := range handler.Protocols {
		log.WithField("count", count).WithField("value", protocol).Infoln("Protocol")
	}
	log.Infoln("")

	log.WithFields(log.Fields{
		"crawledPeers":    handler.CrawledPeers,
		"crawlDuration":   time.Since(dbCrawl.StartedAt).String(),
		"dialablePeers":   handler.CrawledPeers - handler.TotalErrors(),
		"undialablePeers": handler.TotalErrors(),
	}).Infoln("Finished crawl")
}

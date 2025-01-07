package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/bitcoin"
	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/discv4"
	"github.com/dennis-tra/nebula-crawler/discv5"
	"github.com/dennis-tra/nebula-crawler/libp2p"
	"github.com/dennis-tra/nebula-crawler/utils"
)

var crawlConfig = &config.Crawl{
	Root:             rootConfig,
	CrawlWorkerCount: 1000,
	WriteWorkerCount: 10,
	CrawlLimit:       0,
	PersistNeighbors: false,
	FilePathUdgerDB:  "",
	Network:          string(config.NetworkIPFS),
	BootstrapPeers:   cli.NewStringSlice(),
	Protocols:        cli.NewStringSlice(string(kaddht.ProtocolDHT)),
	AddrTrackTypeStr: "public",
	AddrDialTypeStr:  "public",
	KeepENR:          false,
	CheckExposed:     false,
	UDPRespTimeout:   3 * time.Second,
}

// CrawlCommand contains the crawl sub-command configuration.
var CrawlCommand = &cli.Command{
	Name:   "crawl",
	Usage:  "Crawls the entire network starting with a set of bootstrap nodes.",
	Action: CrawlAction,
	Before: func(c *cli.Context) error {
		// based on the network setting, return the default bootstrap peers and protocols
		bootstrapPeers, protocols, err := config.ConfigureNetwork(crawlConfig.Network)
		if err != nil {
			return err
		}

		// Give CLI option precedence
		if c.IsSet("protocols") {
			crawlConfig.Protocols = cli.NewStringSlice(c.StringSlice("protocols")...)
		} else {
			crawlConfig.Protocols = protocols
		}

		if c.IsSet("bootstrap-peers") {
			crawlConfig.BootstrapPeers = cli.NewStringSlice(c.StringSlice("bootstrap-peers")...)
		} else {
			crawlConfig.BootstrapPeers = bootstrapPeers
		}

		if log.GetLevel() >= log.DebugLevel {
			log.Debugln("Using the following configuration:")
			fmt.Println(crawlConfig.String())
		}

		switch config.AddrType(strings.ToLower(crawlConfig.AddrTrackTypeStr)) {
		case config.AddrTypePrivate:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypePrivate)
		case config.AddrTypePublic:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypePublic)
		case config.AddrTypeAny:
			crawlConfig.AddrTrackTypeStr = string(config.AddrTypeAny)
		default:
			return fmt.Errorf("unknown type of addresses to track: %s (supported values are private, public, any)", crawlConfig.AddrTrackTypeStr)
		}

		switch config.AddrType(strings.ToLower(crawlConfig.AddrDialTypeStr)) {
		case config.AddrTypePrivate:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypePrivate)
		case config.AddrTypePublic:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypePublic)
		case config.AddrTypeAny:
			crawlConfig.AddrDialTypeStr = string(config.AddrTypeAny)
		default:
			return fmt.Errorf("unknown type of addresses to dial: %s (supported values are private, public, any)", crawlConfig.AddrDialTypeStr)
		}

		// Set the maximum idle connections to avoid opening and
		// closing connections to the database
		rootConfig.Database.MaxIdleConns = crawlConfig.WriteWorkerCount

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "bootstrap-peers", // TODO: rename to bootstrappers
			Usage:       "Comma separated list of multi addresses of bootstrap peers",
			EnvVars:     []string{"NEBULA_CRAWL_BOOTSTRAP_PEERS", "NEBULA_BOOTSTRAP_PEERS" /* legacy */},
			Destination: crawlConfig.BootstrapPeers,
			DefaultText: "default " + crawlConfig.Network,
		},
		&cli.StringSliceFlag{
			Name:        "protocols",
			Usage:       "Comma separated list of protocols that this crawler should look for",
			EnvVars:     []string{"NEBULA_CRAWL_PROTOCOLS", "NEBULA_PROTOCOLS" /* legacy */},
			Value:       crawlConfig.Protocols,
			Destination: crawlConfig.Protocols,
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial and crawl peers.",
			EnvVars:     []string{"NEBULA_CRAWL_WORKER_COUNT"},
			Value:       crawlConfig.CrawlWorkerCount,
			Destination: &crawlConfig.CrawlWorkerCount,
		},
		&cli.IntFlag{
			Name:        "write-workers",
			Usage:       "How many concurrent workers should write crawl results to the database.",
			EnvVars:     []string{"NEBULA_CRAWL_WRITE_WORKER_COUNT"},
			Value:       crawlConfig.WriteWorkerCount,
			Destination: &crawlConfig.WriteWorkerCount,
			Hidden:      true,
		},
		&cli.IntFlag{
			Name:        "limit",
			Usage:       "Only crawl the specified amount of peers (0 for unlimited)",
			EnvVars:     []string{"NEBULA_CRAWL_PEER_LIMIT"},
			Value:       crawlConfig.CrawlLimit,
			Destination: &crawlConfig.CrawlLimit,
		},
		&cli.BoolFlag{
			Name:        "neighbors",
			Usage:       "Whether to persist all k-bucket entries of a particular peer at the end of a crawl.",
			EnvVars:     []string{"NEBULA_CRAWL_NEIGHBORS"},
			Value:       crawlConfig.PersistNeighbors,
			Destination: &crawlConfig.PersistNeighbors,
		},
		&cli.StringFlag{
			Name:        "addr-track-type",
			Usage:       "Which type addresses should be stored to the database (private, public, any)",
			EnvVars:     []string{"NEBULA_CRAWL_ADDR_TRACK_TYPE"},
			Value:       crawlConfig.AddrTrackTypeStr,
			Destination: &crawlConfig.AddrTrackTypeStr,
		},
		&cli.StringFlag{
			Name:        "addr-dial-type",
			Usage:       "Which type of addresses should Nebula try to dial (private, public, any)",
			EnvVars:     []string{"NEBULA_CRAWL_ADDR_DIAL_TYPE"},
			Value:       crawlConfig.AddrDialTypeStr,
			Destination: &crawlConfig.AddrDialTypeStr,
		},
		&cli.StringFlag{
			Name:        "network",
			Usage:       "Which network should be crawled. Presets default bootstrap peers and protocol. Run: `nebula networks` for more information.",
			EnvVars:     []string{"NEBULA_CRAWL_NETWORK"},
			Value:       crawlConfig.Network,
			Destination: &crawlConfig.Network,
		},
		&cli.BoolFlag{
			Name:        "check-exposed",
			Usage:       "IPFS/AMINO: Whether to check if the Kubo API is exposed. Checking also includes crawling the API.",
			EnvVars:     []string{"NEBULA_CRAWL_CHECK_EXPOSED"},
			Value:       crawlConfig.CheckExposed,
			Destination: &crawlConfig.CheckExposed,
			Category:    flagCategoryNetwork,
		},
		&cli.BoolFlag{
			Name:        "keep-enr",
			Usage:       "ETHEREUM_CONSENSUS: Whether to keep the full ENR.",
			EnvVars:     []string{"NEBULA_CRAWL_KEEP_ENR"},
			Value:       crawlConfig.KeepENR,
			Destination: &crawlConfig.KeepENR,
			Category:    flagCategoryNetwork,
		},
		&cli.DurationFlag{
			Name:        "udp-response-timeout",
			Usage:       "ETHEREUM_EXECUTION: The response timeout for UDP requests in the disv4 DHT",
			EnvVars:     []string{"NEBULA_CRAWL_UDP_RESPONSE_TIMEOUT"},
			Value:       crawlConfig.UDPRespTimeout,
			Destination: &crawlConfig.UDPRespTimeout,
			Category:    flagCategoryNetwork,
		},
	},
}

// CrawlAction is the function that is called when running `nebula crawl`.
func CrawlAction(c *cli.Context) error {
	log.Infoln("Starting Nebula crawler...")
	defer log.Infoln("Stopped Nebula crawler.")

	// init convenience variables
	ctx := c.Context
	cfg := crawlConfig

	// initialize a new database client based on the given configuration.
	// Options are Postgres, JSON, and noop (dry-run).
	dbc, err := db.NewClient(ctx, rootConfig.Database)
	if err != nil {
		return fmt.Errorf("new database client: %w", err)
	}
	defer func() {
		if err := dbc.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) && !strings.Contains(err.Error(), "use of closed network connection") {
			log.WithError(err).Warnln("Failed closing database handle")
		}
	}()

	// Query some additional bootstrap peers from the database.
	// This is optional, and we only log a warning if that doesn't work.
	bpAddrInfos, err := dbc.QueryBootstrapPeers(ctx, 10)
	if err != nil {
		log.WithError(err).Warnln("Failed querying bootstrap peers")
	}
	log.WithField("limit", 10).Infof("Queried %d bootstrap peers\n", len(bpAddrInfos))

	// Inserting a crawl row into the db so that we
	// can associate results with this crawl via
	// its DB identifier
	dbCrawl, err := dbc.InitCrawl(ctx, c.App.Version)
	if err != nil {
		return fmt.Errorf("creating crawl in db: %w", err)
	}

	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, cfg.Root.DialTimeout)

	// Allow transient connections. This way we can crawl a peer even if it is relayed.
	ctx = network.WithAllowLimitedConn(ctx, "reach peers behind NATs")

	// This is a custom configuration option that only exists in our fork of go-libp2p.
	// see: https://github.com/plprobelab/go-libp2p/commit/f6d73ce3093ded293f0de032d239709069fac586
	ctx = network.WithDisableBackoff(ctx, "prevent backoff")

	handlerCfg := &core.CrawlHandlerConfig{
		TrackNeighbors: cfg.PersistNeighbors,
	}

	engineCfg := &core.EngineConfig{
		WorkerCount:         cfg.CrawlWorkerCount,
		WriterCount:         cfg.WriteWorkerCount,
		Limit:               cfg.CrawlLimit,
		AddrDialType:        cfg.AddrDialType(),
		DuplicateProcessing: false,
		TracerProvider:      cfg.Root.TracerProvider,
		MeterProvider:       cfg.Root.MeterProvider,
	}

	switch cfg.Network {
	case string(config.NetworkEthExec):

		bpEnodes, err := cfg.BootstrapEnodesV4()
		if err != nil {
			return err
		}

		for _, addrInfo := range bpAddrInfos {
			n, err := utils.ToEnode(addrInfo.ID, addrInfo.Addrs)
			if err != nil {
				// this is just a best-effort operation so only
				// log the error and continue
				log.WithError(err).WithFields(log.Fields{
					"pid":    addrInfo.ID,
					"maddrs": addrInfo.Addrs,
				}).Warnln("Failed transforming AddrInfo to *enode.Node")
				continue
			}
			bpEnodes = append(bpEnodes, n)
		}

		// configure the crawl driver
		driverCfg := &discv4.CrawlDriverConfig{
			Version:          cfg.Root.Version(),
			DialTimeout:      cfg.Root.DialTimeout,
			CrawlWorkerCount: cfg.CrawlWorkerCount,
			TrackNeighbors:   cfg.PersistNeighbors,
			BootstrapPeers:   bpEnodes,
			AddrDialType:     cfg.AddrDialType(),
			AddrTrackType:    cfg.AddrTrackType(),
			TracerProvider:   cfg.Root.TracerProvider,
			MeterProvider:    cfg.Root.MeterProvider,
			LogErrors:        cfg.Root.LogErrors,
			KeepENR:          cfg.KeepENR,
			UDPBufferSize:    cfg.Root.UDPBufferSize,
			UDPRespTimeout:   cfg.UDPRespTimeout,
		}

		// init the crawl driver
		driver, err := discv4.NewCrawlDriver(dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new discv4 driver: %w", err)
		}

		// init the result handler
		handler := core.NewCrawlHandler[discv4.PeerInfo](handlerCfg)

		// put everything together and init the engine that'll run the crawl
		eng, err := core.NewEngine[discv4.PeerInfo, core.CrawlResult[discv4.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// finally, start the crawl
		queuedPeers, runErr := eng.Run(ctx)

		// a bit ugly but, but the handler will contain crawl statistics, that
		// we'll save to the database and print to the screen
		handler.QueuedPeers = len(queuedPeers)
		if err := persistCrawlInformation(dbc, dbCrawl, handler, runErr); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}

		return nil

	case string(config.NetworkBitcoin):
		bpEnodes, err := cfg.BootstrapBitcoinEntries()
		if err != nil {
			return err
		}

		for _, addrInfo := range bpAddrInfos {
			bpEnodes = append(bpEnodes, addrInfo.Addrs...)
		}
		// configure the crawl driver
		driverCfg := &bitcoin.CrawlDriverConfig{
			Version:        cfg.Root.Version(),
			DialTimeout:    cfg.Root.DialTimeout,
			TrackNeighbors: cfg.PersistNeighbors,
			BootstrapPeers: bpEnodes,
			AddrDialType:   cfg.AddrDialType(),
			AddrTrackType:  cfg.AddrTrackType(),
			KeepENR:        crawlConfig.KeepENR,
			TracerProvider: cfg.Root.TracerProvider,
			MeterProvider:  cfg.Root.MeterProvider,
			LogErrors:      cfg.Root.LogErrors,
			UDPBufferSize:  cfg.Root.UDPBufferSize,
			UDPRespTimeout: cfg.UDPRespTimeout,
		}

		// init the crawl driver
		driver, err := bitcoin.NewCrawlDriver(dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new bitcoin driver: %w", err)
		}

		// init the result handler
		handler := core.NewCrawlHandler[bitcoin.PeerInfo](handlerCfg)

		// put everything together and init the engine that'll run the crawl
		eng, err := core.NewEngine[bitcoin.PeerInfo, core.CrawlResult[bitcoin.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// finally, start the crawl
		queuedPeers, runErr := eng.Run(ctx)

		// a bit ugly but, but the handler will contain crawl statistics, that
		// we'll save to the database and print to the screen
		handler.QueuedPeers = len(queuedPeers)
		if err := persistCrawlInformation(dbc, dbCrawl, handler, runErr); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}

		return nil

	case string(config.NetworkEthCons),
		string(config.NetworkHolesky): // use a different driver etc. for the Ethereum consensus layer + Holeksy Testnet

		bpEnodes, err := cfg.BootstrapEnodesV5()
		if err != nil {
			return err
		}

		for _, addrInfo := range bpAddrInfos {
			n, err := utils.ToEnode(addrInfo.ID, addrInfo.Addrs)
			if err != nil {
				// this is just a best-effort operation so only
				// log the error and continue
				log.WithError(err).WithFields(log.Fields{
					"pid":    addrInfo.ID,
					"maddrs": addrInfo.Addrs,
				}).Warnln("Failed transforming AddrInfo to *enode.Node")
				continue
			}
			bpEnodes = append(bpEnodes, n)
		}

		// configure the crawl driver
		driverCfg := &discv5.CrawlDriverConfig{
			Version:        cfg.Root.Version(),
			DialTimeout:    cfg.Root.DialTimeout,
			TrackNeighbors: cfg.PersistNeighbors,
			BootstrapPeers: bpEnodes,
			AddrDialType:   cfg.AddrDialType(),
			AddrTrackType:  cfg.AddrTrackType(),
			KeepENR:        crawlConfig.KeepENR,
			TracerProvider: cfg.Root.TracerProvider,
			MeterProvider:  cfg.Root.MeterProvider,
			LogErrors:      cfg.Root.LogErrors,
			UDPBufferSize:  cfg.Root.UDPBufferSize,
			UDPRespTimeout: cfg.UDPRespTimeout,
		}

		// init the crawl driver
		driver, err := discv5.NewCrawlDriver(dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new discv5 driver: %w", err)
		}

		// init the result handler
		handler := core.NewCrawlHandler[discv5.PeerInfo](handlerCfg)

		// put everything together and init the engine that'll run the crawl
		eng, err := core.NewEngine[discv5.PeerInfo, core.CrawlResult[discv5.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// finally, start the crawl
		queuedPeers, runErr := eng.Run(ctx)

		// a bit ugly but, but the handler will contain crawl statistics, that
		// we'll save to the database and print to the screen
		handler.QueuedPeers = len(queuedPeers)
		if err := persistCrawlInformation(dbc, dbCrawl, handler, runErr); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}

		return nil
	default:

		addrInfos, err := cfg.BootstrapAddrInfos()
		if err != nil {
			return err
		}

		for _, addrInfo := range addrInfos {
			bpAddrInfos = append(bpAddrInfos, addrInfo)
		}

		// configure the crawl driver
		driverCfg := &libp2p.CrawlDriverConfig{
			Version:        cfg.Root.Version(),
			Network:        config.Network(cfg.Network),
			Protocols:      cfg.Protocols.Value(),
			DialTimeout:    cfg.Root.DialTimeout,
			TrackNeighbors: cfg.PersistNeighbors,
			CheckExposed:   cfg.CheckExposed,
			BootstrapPeers: bpAddrInfos,
			AddrDialType:   cfg.AddrDialType(),
			AddrTrackType:  cfg.AddrTrackType(),
			TracerProvider: cfg.Root.TracerProvider,
			MeterProvider:  cfg.Root.MeterProvider,
			LogErrors:      cfg.Root.LogErrors,
		}

		// init the crawl driver
		driver, err := libp2p.NewCrawlDriver(dbc, dbCrawl, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		// init the result handler
		handler := core.NewCrawlHandler[libp2p.PeerInfo](handlerCfg)

		// put everything together and init the engine that'll run the crawl
		eng, err := core.NewEngine[libp2p.PeerInfo, core.CrawlResult[libp2p.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// finally, start the crawl
		queuedPeers, runErr := eng.Run(ctx)

		// a bit ugly but, but the handler will contain crawl statistics, that
		// we'll save to the database and print to the screen
		handler.QueuedPeers = len(queuedPeers)
		if err := persistCrawlInformation(dbc, dbCrawl, handler, runErr); err != nil {
			return fmt.Errorf("persist crawl information: %w", err)
		}
	}

	return nil
}

func persistCrawlInformation[I core.PeerInfo[I]](dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I], runErr error) error {
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
	if err := updateCrawl(cleanupCtx, dbc, dbCrawl, runErr, handler); err != nil {
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
func updateCrawl[I core.PeerInfo[I]](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, runErr error, handler *core.CrawlHandler[I]) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	log.Infoln("Persisting crawl result...")

	dbCrawl.FinishedAt = null.TimeFrom(time.Now())
	dbCrawl.CrawledPeers = null.IntFrom(handler.CrawledPeers)
	dbCrawl.DialablePeers = null.IntFrom(handler.CrawledPeers - handler.TotalErrors())
	dbCrawl.UndialablePeers = null.IntFrom(handler.TotalErrors())
	dbCrawl.RemainingPeers = null.IntFrom(handler.QueuedPeers)

	if runErr == nil {
		dbCrawl.State = models.CrawlStateSucceeded
	} else if errors.Is(runErr, context.Canceled) {
		dbCrawl.State = models.CrawlStateCancelled
	} else {
		dbCrawl.State = models.CrawlStateFailed
	}

	return dbc.UpdateCrawl(ctx, dbCrawl)
}

// persistCrawlProperties writes crawl property statistics to the database.
func persistCrawlProperties[I core.PeerInfo[I]](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) error {
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
		"error":         handler.ConnErrs,
	}

	return dbc.PersistCrawlProperties(ctx, dbCrawl, pps)
}

// storeNeighbors fills the neighbors table with topology information
func storeNeighbors[I core.PeerInfo[I]](ctx context.Context, dbc db.Client, dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	if len(handler.RoutingTables) == 0 {
		return nil
	}

	log.Infoln("Storing neighbor information...")

	start := time.Now()
	neighborsCount := 0
	i := 0
	for p, routingTable := range handler.RoutingTables {
		if i%100 == 0 && i > 0 {
			log.Infof("Stored %d peers and their neighbors", i)
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
		"duration":       time.Since(start).String(),
		"avg":            fmt.Sprintf("%.2fms", time.Since(start).Seconds()/float64(len(handler.RoutingTables))*1000),
		"peers":          len(handler.RoutingTables),
		"totalNeighbors": neighborsCount,
	}).Infoln("Finished storing neighbor information")
	return nil
}

// logSummary logs the final results of the crawl.
func logSummary[I core.PeerInfo[I]](dbCrawl *models.Crawl, handler *core.CrawlHandler[I]) {
	log.Infoln("Crawl summary:")

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

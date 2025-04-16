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
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/bitcoin"
	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/discv4"
	"github.com/dennis-tra/nebula-crawler/discv5"
	"github.com/dennis-tra/nebula-crawler/libp2p"
	"github.com/dennis-tra/nebula-crawler/utils"
)

var crawlConfig = &config.Crawl{
	Root:              rootConfig,
	CrawlWorkerCount:  1000,
	WriteWorkerCount:  10,
	CrawlLimit:        0,
	PersistNeighbors:  false,
	FilePathUdgerDB:   "",
	Network:           string(config.NetworkIPFS),
	BootstrapPeers:    cli.NewStringSlice(),
	Protocols:         cli.NewStringSlice(string(kaddht.ProtocolDHT)),
	AddrTrackTypeStr:  "public",
	AddrDialTypeStr:   "public",
	KeepENR:           false,
	CheckExposed:      false,
	UDPRespTimeout:    3 * time.Second,
	EnableGossipSubPX: false,
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

		// set the network ID on the database object
		rootConfig.Database.NetworkID = crawlConfig.Network

		// set the persist neighbors flag on the database object
		rootConfig.Database.PersistNeighbors = crawlConfig.PersistNeighbors

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
			Name:        "gossipsub-px",
			Usage:       "Whether to enable gossipsub peer exchange crawling",
			EnvVars:     []string{"NEBULA_CRAWL_GOSSIPSUB_PX"},
			Value:       crawlConfig.EnableGossipSubPX,
			Destination: &crawlConfig.EnableGossipSubPX,
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
		&cli.IntFlag{
			Name:        "waku-cluster-id",
			Usage:       "WAKU/WAKU_TWN: The cluster ID for the Waku network",
			EnvVars:     []string{"NEBULA_CRAWL_WAKU_CLUSTER_ID"},
			Value:       crawlConfig.WakuClusterID,
			Destination: &crawlConfig.WakuClusterID,
			Category:    flagCategoryNetwork,
			Hidden:      true,
		},
		&cli.IntSliceFlag{
			Name:        "waku-cluster-shards",
			Usage:       "WAKU_STATUS/WAKU_TWN: The cluster shards of the Waku network",
			EnvVars:     []string{"NEBULA_CRAWL_WAKU_CLUSTER_SHARDS"},
			Value:       crawlConfig.WakuClusterShards,
			Destination: crawlConfig.WakuClusterShards,
			Category:    flagCategoryNetwork,
			Hidden:      true,
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
	start := time.Now()

	// initialize a new database client based on the given configuration.
	// Options are Postgres, JSON, and noop (dry-run).
	dbc, err := rootConfig.Database.NewClient(ctx)
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
	if err := dbc.InitCrawl(ctx, c.App.Version); err != nil {
		return fmt.Errorf("creating crawl in db: %w", err)
	}

	// Set the timeout for dialing peers
	ctx = network.WithDialPeerTimeout(ctx, cfg.Root.DialTimeout)

	// Allow transient connections. This way we can crawl a peer even if it is relayed.
	ctx = network.WithAllowLimitedConn(ctx, "reach peers behind NATs")

	// This is a custom configuration option that only exists in our fork of go-libp2p.
	// see: https://github.com/plprobelab/go-libp2p/commit/f6d73ce3093ded293f0de032d239709069fac586
	ctx = network.WithDisableBackoff(ctx, "prevent backoff")

	handlerCfg := &core.CrawlHandlerConfig{}

	engineCfg := &core.EngineConfig{
		WorkerCount:         cfg.CrawlWorkerCount,
		WriterCount:         cfg.WriteWorkerCount,
		Limit:               cfg.CrawlLimit,
		AddrDialType:        cfg.AddrDialType(),
		DuplicateProcessing: false,
		TracerProvider:      cfg.Root.TracerProvider,
		MeterProvider:       cfg.Root.MeterProvider,
	}

	var (
		summary *core.Summary
		runErr  error
	)

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
		driver, err := discv4.NewCrawlDriver(dbc, driverCfg)
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
		summary, runErr = eng.Run(ctx)

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
			BootstrapPeers: bpEnodes,
			TracerProvider: cfg.Root.TracerProvider,
			MeterProvider:  cfg.Root.MeterProvider,
			LogErrors:      cfg.Root.LogErrors,
		}

		// init the crawl driver
		driver, err := bitcoin.NewCrawlDriver(dbc, driverCfg)
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
		summary, runErr = eng.Run(ctx)

	case string(config.NetworkEthCons),
		string(config.NetworkHolesky),
		string(config.NetworkPortal),
		string(config.NetworkWakuStatus),
		string(config.NetworkWakuTWN),
		string(config.NetworkGnosis):
		// use a different driver etc. for the Ethereum consensus layer + Holeksy Testnet + Waku networks

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

		protocolID, err := cfg.DiscV5ProtocolID()
		if err != nil {
			return fmt.Errorf("parse discv5 protocol ID: %w", err)
		}

		wakuClusterID, wakuClusterShards := cfg.WakuClusterConfig()

		// configure the crawl driver
		driverCfg := &discv5.CrawlDriverConfig{
			Version:           cfg.Root.Version(),
			Network:           config.Network(cfg.Network),
			DialTimeout:       cfg.Root.DialTimeout,
			BootstrapPeers:    bpEnodes,
			CrawlWorkerCount:  cfg.CrawlWorkerCount,
			AddrDialType:      cfg.AddrDialType(),
			AddrTrackType:     cfg.AddrTrackType(),
			KeepENR:           crawlConfig.KeepENR,
			TracerProvider:    cfg.Root.TracerProvider,
			MeterProvider:     cfg.Root.MeterProvider,
			LogErrors:         cfg.Root.LogErrors,
			Discv5ProtocolID:  protocolID,
			UDPBufferSize:     cfg.Root.UDPBufferSize,
			UDPRespTimeout:    cfg.UDPRespTimeout,
			WakuClusterID:     wakuClusterID,
			WakuClusterShards: wakuClusterShards,
		}

		// init the crawl driver
		driver, err := discv5.NewCrawlDriver(dbc, driverCfg)
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
		summary, runErr = eng.Run(ctx)

	default:

		addrInfos, err := cfg.BootstrapAddrInfos()
		if err != nil {
			return err
		}

		bpAddrInfos = append(bpAddrInfos, addrInfos...)

		// configure the crawl driver
		driverCfg := &libp2p.CrawlDriverConfig{
			Version:        cfg.Root.Version(),
			WorkerCount:    cfg.CrawlWorkerCount,
			Network:        config.Network(cfg.Network),
			Protocols:      cfg.Protocols.Value(),
			DialTimeout:    cfg.Root.DialTimeout,
			CheckExposed:   cfg.CheckExposed,
			BootstrapPeers: bpAddrInfos,
			AddrDialType:   cfg.AddrDialType(),
			AddrTrackType:  cfg.AddrTrackType(),
			TracerProvider: cfg.Root.TracerProvider,
			MeterProvider:  cfg.Root.MeterProvider,
			GossipSubPX:    cfg.EnableGossipSubPX,
			LogErrors:      cfg.Root.LogErrors,
		}

		// init the crawl driver
		driver, err := libp2p.NewCrawlDriver(dbc, driverCfg)
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
		summary, runErr = eng.Run(ctx)
	}

	// we're done with the crawl so seal the crawl and store aggregate information
	if err := persistCrawlInformation(dbc, summary, runErr); err != nil {
		return fmt.Errorf("persist crawl information: %w", err)
	}

	logSummary(summary, time.Since(start))

	return nil
}

func persistCrawlInformation(dbc db.Client, summary *core.Summary, runErr error) error {
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
	if err := updateCrawl(cleanupCtx, dbc, runErr, summary); err != nil {
		return fmt.Errorf("persist crawl: %w", err)
	}

	// Persist associated crawl properties
	if err := persistCrawlProperties(cleanupCtx, dbc, summary); err != nil {
		return fmt.Errorf("persist crawl properties: %w", err)
	}

	// flush any left-over information to the database.
	if err := dbc.Flush(cleanupCtx); err != nil {
		log.WithError(err).Warnln("Failed flushing information to database")
	}

	return nil
}

// updateCrawl writes crawl statistics to the database
func updateCrawl(ctx context.Context, dbc db.Client, runErr error, summary *core.Summary) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	}

	log.Infoln("Persisting crawl result...")

	args := &db.SealCrawlArgs{
		Crawled:    summary.PeersCrawled,
		Dialable:   summary.PeersDialable,
		Undialable: summary.PeersUndialable,
		Remaining:  summary.PeersRemaining,
	}

	if runErr == nil {
		args.State = db.CrawlStateSucceeded
	} else if errors.Is(runErr, context.Canceled) {
		args.State = db.CrawlStateCancelled
	} else {
		args.State = db.CrawlStateFailed
	}

	return dbc.SealCrawl(ctx, args)
}

// persistCrawlProperties writes crawl property statistics to the database.
func persistCrawlProperties(ctx context.Context, dbc db.Client, summary *core.Summary) error {
	if _, ok := dbc.(*db.NoopClient); ok {
		return nil
	} else if _, ok := dbc.(*db.ClickHouseClient); ok {
		return nil
	}

	log.Infoln("Persisting crawl properties...")
	avFull := map[string]int{}
	for version, count := range summary.AgentVersion {
		avFull[version] += count
	}
	pps := map[string]map[string]int{
		"agent_version": avFull,
		"protocol":      summary.Protocols,
		"error":         summary.ConnErrs,
	}

	return dbc.InsertCrawlProperties(ctx, pps)
}

// logSummary logs the final results of the crawl.
func logSummary(summary *core.Summary, crawlDuration time.Duration) {
	log.Infoln("")
	log.Infoln("")
	log.Infoln("Crawl summary:")

	log.Infoln("")
	for err, count := range summary.ConnErrs {
		log.WithField("count", count).WithField("value", err).Infoln("Dial Error")
	}

	log.Infoln("")
	for err, count := range summary.CrawlErrs {
		log.WithField("count", count).WithField("value", err).Infoln("Crawl Error")
	}

	log.Infoln("")
	for agent, count := range summary.AgentVersion {
		log.WithField("count", count).WithField("value", agent).Infoln("Agent")
	}
	log.Infoln("")
	for protocol, count := range summary.Protocols {
		log.WithField("count", count).WithField("value", protocol).Infoln("Protocol")
	}
	log.Infoln("")
	log.WithFields(log.Fields{
		"crawledPeers":    summary.PeersCrawled,
		"crawlDuration":   crawlDuration.String(),
		"dialablePeers":   summary.PeersDialable,
		"undialablePeers": summary.PeersUndialable,
		"remainingPeers":  summary.PeersRemaining,
	}).Infoln("Finished crawl")
}

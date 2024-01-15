package main

import (
	"context"
	"errors"
	"fmt"

	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/discv5"
	"github.com/dennis-tra/nebula-crawler/libp2p"
)

var monitorConfig = &config.Monitor{
	Root:               rootConfig,
	MonitorWorkerCount: 1000,
	WriteWorkerCount:   10,
	Network:            string(config.NetworkIPFS),
	Protocols:          cli.NewStringSlice(string(kaddht.ProtocolDHT)),
}

// MonitorCommand contains the monitor sub-command configuration.
var MonitorCommand = &cli.Command{
	Name:   "monitor",
	Usage:  "Monitors the network by periodically dialing previously crawled peers.",
	Action: MonitorAction,
	Before: func(c *cli.Context) error {
		// based on the network setting, return the default bootstrap peers and protocols
		_, protocols, err := config.ConfigureNetwork(monitorConfig.Network)
		if err != nil {
			return err
		}

		// Give CLI option precedence
		if c.IsSet("protocols") {
			monitorConfig.Protocols = cli.NewStringSlice(c.StringSlice("protocols")...)
		} else {
			monitorConfig.Protocols = protocols
		}

		log.Debugln("Using the following configuration:")
		if log.GetLevel() >= log.DebugLevel {
			fmt.Println(monitorConfig.String())
		}

		// Set the maximum idle connections to avoid opening and
		// closing connections to the database
		rootConfig.Database.MaxIdleConns = monitorConfig.WriteWorkerCount

		return nil
	},
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial peers.",
			EnvVars:     []string{"NEBULA_MONITOR_WORKER_COUNT"},
			Value:       monitorConfig.MonitorWorkerCount,
			Destination: &monitorConfig.MonitorWorkerCount,
		},
		&cli.IntFlag{
			Name:        "write-workers",
			Usage:       "How many concurrent workers should write results to the database.",
			EnvVars:     []string{"NEBULA_MONITOR_WRITE_WORKER_COUNT"},
			Value:       monitorConfig.WriteWorkerCount,
			Destination: &monitorConfig.WriteWorkerCount,
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        "network",
			Usage:       "Which network belong the database sessions to. Relevant for parsing peer IDs and muti addresses.",
			EnvVars:     []string{"NEBULA_MONITOR_NETWORK"},
			Value:       monitorConfig.Network,
			Destination: &monitorConfig.Network,
		},
	},
}

// MonitorAction is the function that is called when running `nebula monitor`.
func MonitorAction(c *cli.Context) error {
	log.Infoln("Starting Nebula monitor...")
	defer log.Infoln("Stopped Nebula monitor.")

	// Acquire database handle
	dbc, err := db.InitDBClient(c.Context, rootConfig.Database)
	if err != nil {
		return fmt.Errorf("init db client: %w", err)
	}
	defer func() {
		if err := dbc.Close(); err != nil {
			log.WithError(err).Warnln("Failed closing database handle")
		}
	}()

	handlerCfg := &core.DialHandlerConfig{}

	engineCfg := &core.EngineConfig{
		WorkerCount:         monitorConfig.MonitorWorkerCount,
		WriterCount:         monitorConfig.WriteWorkerCount,
		Limit:               0,
		DuplicateProcessing: true,
		AddrDialType:        config.AddrTypeAny,
		TracerProvider:      monitorConfig.Root.TracerProvider,
		MeterProvider:       monitorConfig.Root.MeterProvider,
	}

	switch monitorConfig.Network {
	case string(config.NetworkEthCons):
		driverCfg := &discv5.DialDriverConfig{
			Version: monitorConfig.Root.Version(),
		}

		driver, err := discv5.NewDialDriver(dbc, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewDialHandler[discv5.PeerInfo](handlerCfg)
		eng, err := core.NewEngine[discv5.PeerInfo, core.DialResult[discv5.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		_, err = eng.Run(c.Context)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}

	default:
		driverCfg := &libp2p.DialDriverConfig{
			Version:     monitorConfig.Root.Version(),
			DialTimeout: monitorConfig.Root.DialTimeout,
		}

		driver, err := libp2p.NewDialDriver(dbc, driverCfg)
		if err != nil {
			return fmt.Errorf("new driver: %w", err)
		}

		handler := core.NewDialHandler[libp2p.PeerInfo](handlerCfg)
		eng, err := core.NewEngine[libp2p.PeerInfo, core.DialResult[libp2p.PeerInfo]](driver, handler, engineCfg)
		if err != nil {
			return fmt.Errorf("new engine: %w", err)
		}

		// Set the timeout for dialing peers
		ctx := network.WithDialPeerTimeout(c.Context, monitorConfig.Root.DialTimeout)

		// Allow transient connections. This way we can crawl a peer even if it is relayed.
		ctx = network.WithUseTransient(ctx, "reach peers behind NATs")

		// This is a custom configuration option that only exists in our fork of go-libp2p.
		// see: https://github.com/plprobelab/go-libp2p/commit/f6d73ce3093ded293f0de032d239709069fac586
		ctx = network.WithDisableBackoff(ctx, "prevent backoff")

		_, err = eng.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("running crawl engine: %w", err)
		}
	}
	return nil
}

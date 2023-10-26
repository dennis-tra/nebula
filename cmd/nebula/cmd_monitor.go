package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/monitor"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
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

	// Initialize the monitoring task
	s, err := monitor.New(dbc, monitorConfig)
	if err != nil {
		return fmt.Errorf("new monitor: %w", err)
	}

	return s.MonitorNetwork(c.Context)
}

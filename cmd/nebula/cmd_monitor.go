package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/monitor"
)

var monitorConfig = &config.Monitor{
	Root:               rootConfig,
	MonitorWorkerCount: 1000,
}

// MonitorCommand contains the monitor sub-command configuration.
var MonitorCommand = &cli.Command{
	Name:   "monitor",
	Usage:  "Monitors the network by periodically dialing previously crawled peers.",
	Action: MonitorAction,
	Before: func(c *cli.Context) error {
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
	},
}

// MonitorAction is the function that is called when running `nebula monitor`.
func MonitorAction(c *cli.Context) error {
	log.Infoln("Starting Nebula monitor...")

	// Acquire database handle
	dbc, err := db.InitDBClient(c.Context, rootConfig)
	if err != nil {
		return fmt.Errorf("init db client: %w", err)
	}

	// Initialize the monitoring task
	s, err := monitor.NewScheduler(monitorConfig, dbc)
	if err != nil {
		return fmt.Errorf("creating new scheduler: %w", err)
	}

	return s.StartMonitoring(c.Context)
}

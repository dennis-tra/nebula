package main

import (
	"strconv"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/monitor"
)

// MonitorCommand contains the monitor sub-command configuration.
var MonitorCommand = &cli.Command{
	Name:   "monitor",
	Usage:  "Monitors the network by periodically dialing previously crawled peers.",
	Action: MonitorAction,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "How many concurrent workers should dial peers.",
			EnvVars:     []string{"NEBULA_MONITOR_WORKER_COUNT"},
			DefaultText: strconv.Itoa(config.DefaultConfig.MonitorWorkerCount),
			Value:       config.DefaultConfig.MonitorWorkerCount,
		},
	},
}

// MonitorAction is the function that is called when running `nebula monitor`.
func MonitorAction(c *cli.Context) error {
	log.Infoln("Starting Nebula monitor...")

	// Load configuration file
	ctx, conf, err := config.FillContext(c)
	if err != nil {
		return errors.Wrap(err, "filling context with configuration")
	}
	c.Context = ctx

	// Acquire database handle
	dbc, err := db.InitClient(c.Context)
	if err != nil {
		return err
	}

	// Start prometheus metrics endpoint
	if err = metrics.RegisterMonitorMetrics(); err != nil {
		return err
	}
	if err = metrics.ListenAndServe(conf.PrometheusHost, conf.PrometheusPort); err != nil {
		return errors.Wrap(err, "initialize metrics")
	}

	// Initialize the monitoring task
	s, err := monitor.NewScheduler(c.Context, dbc)
	if err != nil {
		return errors.Wrap(err, "creating new scheduler")
	}

	go func() {
		// Nebula was asked to stop (e.g. SIGINT) -> tell the scheduler to stop
		<-c.Context.Done()
		s.Shutdown()
	}()

	return s.StartMonitoring()
}

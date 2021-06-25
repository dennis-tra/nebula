package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dennis-tra/nebula-crawler/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// RawVersion and build tag of the
	// PCP command line tool. This is
	// replaced on build via e.g.:
	// -ldflags "-X main.RawVersion=${VERSION}"
	RawVersion  = "dev"
	ShortCommit = "5f3759df" // quake
)

func main() {
	// ShortCommit version tag
	verTag := fmt.Sprintf("v%s+%s", RawVersion, ShortCommit)

	app := &cli.App{
		Name:      "nebula",
		Usage:     "A libp2p DHT crawler that exposes timely information about the network.",
		UsageText: "nebula [global options] command [command options] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "nebula-crawler@dtrautwein.eu",
			},
		},
		Version: verTag,
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				log.SetLevel(log.DebugLevel)
			}
			if c.IsSet("log-level") {
				log.SetLevel(log.Level(c.Int("log-level")))
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Set this flag to enable debug logging.",
				EnvVars: []string{"NEBULA_DEBUG"},
			},
			&cli.IntFlag{
				Name:        "log-level",
				Usage:       "Set this flag to a value from 0 to 6. Overrides the --debug flag.",
				EnvVars:     []string{"NEBULA_LOG_LEVEL"},
				Value:       4,
				DefaultText: "4",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				EnvVars: []string{"NEBULA_CONFIG_FILE"},
			},
			&cli.IntFlag{
				Name:        "prom-port",
				Usage:       "On which port should prometheus serve the metrics endpoint",
				EnvVars:     []string{"NEBULA_PROMETHEUS_PORT"},
				DefaultText: strconv.Itoa(config.DefaultConfig.PrometheusPort),
				Value:       config.DefaultConfig.PrometheusPort,
			},
			&cli.StringFlag{
				Name:        "prom-host",
				Usage:       "On which network interface should prometheus serve the metrics endpoint",
				EnvVars:     []string{"NEBULA_PROMETHEUS_HOST"},
				DefaultText: config.DefaultConfig.PrometheusHost,
				Value:       config.DefaultConfig.PrometheusHost,
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			CrawlCommand,
			MonitorCommand,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Printf("Received %s - Stopping...\n", sig.String())
		signal.Stop(sigs)
		cancel()
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

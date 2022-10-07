package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	_ "net/http/pprof"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
		Usage:     "A libp2p DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.",
		UsageText: "nebula [global options] command [command options] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "nebula@dtrautwein.eu",
			},
		},
		Version: verTag,
		Before:  Before,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Set this flag to enable debug logging",
				EnvVars: []string{"NEBULA_DEBUG"},
			},
			&cli.IntFlag{
				Name:        "log-level",
				Usage:       "Set this flag to a value from 0 (least verbose) to 6 (most verbose). Overrides the --debug flag",
				EnvVars:     []string{"NEBULA_LOG_LEVEL"},
				Value:       4,
				DefaultText: "4",
			},
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Load configuration from `FILE`",
				EnvVars: []string{"NEBULA_CONFIG_FILE"},
			},
			&cli.DurationFlag{
				Name:        "dial-timeout",
				Usage:       "How long should be waited before a dial is considered unsuccessful",
				EnvVars:     []string{"NEBULA_DIAL_TIMEOUT"},
				DefaultText: config.DefaultConfig.DialTimeout.String(),
				Value:       config.DefaultConfig.DialTimeout,
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
				Usage:       "Where should prometheus serve the metrics endpoint",
				EnvVars:     []string{"NEBULA_PROMETHEUS_HOST"},
				DefaultText: config.DefaultConfig.PrometheusHost,
				Value:       config.DefaultConfig.PrometheusHost,
			},
			&cli.IntFlag{
				Name:        "pprof-port",
				Usage:       "Enable pprof profiling endpoint on given port",
				EnvVars:     []string{"NEBULA_PPROF_PORT"},
				DefaultText: "disabled",
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_HOST"},
				DefaultText: config.DefaultConfig.DatabaseHost,
				Value:       config.DefaultConfig.DatabaseHost,
			},
			&cli.IntFlag{
				Name:        "db-port",
				Usage:       "On which port can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_PORT"},
				DefaultText: strconv.Itoa(config.DefaultConfig.DatabasePort),
				Value:       config.DefaultConfig.DatabasePort,
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_NAME"},
				DefaultText: config.DefaultConfig.DatabaseName,
				Value:       config.DefaultConfig.DatabaseName,
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_PASSWORD"},
				DefaultText: config.DefaultConfig.DatabasePassword,
				Value:       config.DefaultConfig.DatabasePassword,
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_USER"},
				DefaultText: config.DefaultConfig.DatabaseUser,
				Value:       config.DefaultConfig.DatabaseUser,
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"NEBULA_DATABASE_SSL_MODE"},
				DefaultText: config.DefaultConfig.DatabaseSSLMode,
				Value:       config.DefaultConfig.DatabaseSSLMode,
			},
			&cli.StringSliceFlag{
				Name:        "protocols",
				Usage:       "Comma separated list of protocols that this crawler should look for",
				EnvVars:     []string{"NEBULA_PROTOCOLS"},
				DefaultText: "IPFS DHT: " + strings.Join(config.DefaultConfig.Protocols, ","),
				Value:       cli.NewStringSlice(config.DefaultConfig.Protocols...),
			},
			&cli.StringSliceFlag{
				Name:    "bootstrap-peers",
				Usage:   "Comma separated list of multi addresses of bootstrap peers",
				EnvVars: []string{"NEBULA_BOOTSTRAP_PEERS"},
			},
			&cli.IntFlag{
				Name:        "protocols-cache-size",
				Usage:       "The cache size to hold protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_CACHE_SIZE"},
				DefaultText: strconv.Itoa(config.DefaultConfig.ProtocolsCacheSize),
				Value:       config.DefaultConfig.ProtocolsCacheSize,
			},
			&cli.IntFlag{
				Name:        "agent-versions-cache-size",
				Usage:       "The cache size to hold agent versions in memory",
				EnvVars:     []string{"NEBULA_AGENT_VERSIONS_CACHE_SIZE"},
				DefaultText: strconv.Itoa(config.DefaultConfig.AgentVersionsCacheSize),
				Value:       config.DefaultConfig.AgentVersionsCacheSize,
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			CrawlCommand,
			MonitorCommand,
			ResolveCommand,
			PingCommand,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Printf("Received %s signal - Stopping...\n", sig.String())
		signal.Stop(sigs)
		cancel()
	}()

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}

// Before is executed before any subcommands are run, but after the context is ready
// If a non-nil error is returned, no subcommands are run.
func Before(c *cli.Context) error {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.IsSet("log-level") {
		ll := c.Int("log-level")
		log.SetLevel(log.Level(ll))
		if ll == int(log.TraceLevel) {
			boil.DebugMode = true
		}
	}

	if c.IsSet("pprof-port") {
		go func() {
			pprof := fmt.Sprintf("localhost:%d", c.Int("pprof-port"))
			log.Debugln("Starting profiling endpoint at", pprof)
			if err := http.ListenAndServe(pprof, nil); err != nil {
				log.WithError(err).Warnln("Error serving pprof")
			}
		}()
	}

	return nil
}

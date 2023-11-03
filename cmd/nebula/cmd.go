package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
)

const (
	flagCategoryDatabase  = "Database Configuration:"
	flagCategoryDebugging = "Debugging Configuration:"
	flagCategoryCache     = "Cache Configuration:"
	flagCategoryNetwork   = "Network Specific Configuration:"
)

// RawVersion and build tag of the Nebula command line tool.
var RawVersion = "dev"

var rootConfig = &config.Root{
	RawVersion:      RawVersion,
	Debug:           false,
	LogLevel:        4,
	LogFormat:       "text",
	LogDisableColor: false,
	DialTimeout:     5 * time.Second,
	TelemetryHost:   "0.0.0.0",
	TelemetryPort:   6666,
	Database: &config.Database{
		DryRun:                 false,
		JSONOut:                "",
		DatabaseHost:           "localhost",
		DatabasePort:           5432,
		DatabaseName:           "nebula",
		DatabasePassword:       "password",
		DatabaseUser:           "nebula",
		DatabaseSSLMode:        "disable",
		AgentVersionsCacheSize: 200,
		ProtocolsCacheSize:     100,
		ProtocolsSetCacheSize:  200,
	},
}

func main() {
	app := &cli.App{
		Name:      "nebula",
		Usage:     "A DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.",
		UsageText: "nebula [global options] command [command options] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Dennis Trautwein",
				Email: "nebula@dtrautwein.eu",
			},
		},
		Version: rootConfig.Version(),
		Before:  Before,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Set this flag to enable debug logging",
				EnvVars:     []string{"NEBULA_DEBUG"},
				Value:       rootConfig.Debug,
				Destination: &rootConfig.Debug,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "log-level",
				Usage:       "Set this flag to a value from 0 (least verbose) to 6 (most verbose). Overrides the --debug flag",
				EnvVars:     []string{"NEBULA_LOG_LEVEL"},
				Value:       rootConfig.LogLevel,
				Destination: &rootConfig.LogLevel,
				Category:    flagCategoryDebugging,
			},
			&cli.StringFlag{
				Name:        "log-format",
				Usage:       "Define the formatting of the log output (values: text, json)",
				EnvVars:     []string{"NEBULA_LOG_FORMAT"},
				Value:       rootConfig.LogFormat,
				Destination: &rootConfig.LogFormat,
				Category:    flagCategoryDebugging,
			},
			&cli.BoolFlag{
				Name:        "log-disable-color",
				Usage:       "Whether to have colorized log output (only text log format)",
				EnvVars:     []string{"NEBULA_LOG_COLOR"},
				Value:       rootConfig.LogDisableColor,
				Destination: &rootConfig.LogDisableColor,
				Category:    flagCategoryDebugging,
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"NEBULA_TELEMETRY_HOST"},
				Value:       rootConfig.TelemetryHost,
				Destination: &rootConfig.TelemetryHost,
				Category:    flagCategoryDebugging,
			},
			&cli.IntFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"NEBULA_TELEMETRY_PORT"},
				Value:       rootConfig.TelemetryPort,
				Destination: &rootConfig.TelemetryPort,
				Category:    flagCategoryDebugging,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Usage:       "Don't write anything to disk",
				EnvVars:     []string{"NEBULA_DRY_RUN", "NEBULA_CRAWL_DRY_RUN" /*<-legacy*/},
				Value:       rootConfig.Database.DryRun,
				Destination: &rootConfig.Database.DryRun,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "json-out",
				Usage:       "If set, stores results as JSON documents at `DIR` (takes precedence over database settings).",
				EnvVars:     []string{"NEBULA_JSON_OUT", "NEBULA_CRAWL_JSON_OUT" /*<-legacy*/},
				Value:       rootConfig.Database.JSONOut,
				Destination: &rootConfig.Database.JSONOut,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_HOST"},
				Value:       rootConfig.Database.DatabaseHost,
				Destination: &rootConfig.Database.DatabaseHost,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "db-port",
				Usage:       "On which port can nebula reach the database",
				EnvVars:     []string{"NEBULA_DATABASE_PORT"},
				Value:       rootConfig.Database.DatabasePort,
				Destination: &rootConfig.Database.DatabasePort,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_NAME"},
				Value:       rootConfig.Database.DatabaseName,
				Destination: &rootConfig.Database.DatabaseName,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_PASSWORD"},
				Value:       rootConfig.Database.DatabasePassword,
				Destination: &rootConfig.Database.DatabasePassword,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"NEBULA_DATABASE_USER"},
				Value:       rootConfig.Database.DatabaseUser,
				Destination: &rootConfig.Database.DatabaseUser,
				Category:    flagCategoryDatabase,
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"NEBULA_DATABASE_SSL_MODE"},
				Value:       rootConfig.Database.DatabaseSSLMode,
				Destination: &rootConfig.Database.DatabaseSSLMode,
				Category:    flagCategoryDatabase,
			},
			&cli.IntFlag{
				Name:        "agent-versions-cache-size",
				Usage:       "The cache size to hold agent versions in memory",
				EnvVars:     []string{"NEBULA_AGENT_VERSIONS_CACHE_SIZE"},
				Value:       rootConfig.Database.AgentVersionsCacheSize,
				Destination: &rootConfig.Database.AgentVersionsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-cache-size",
				Usage:       "The cache size to hold protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_CACHE_SIZE"},
				Value:       rootConfig.Database.ProtocolsCacheSize,
				Destination: &rootConfig.Database.ProtocolsCacheSize,
				Category:    flagCategoryCache,
			},
			&cli.IntFlag{
				Name:        "protocols-set-cache-size",
				Usage:       "The cache size to hold sets of protocols in memory",
				EnvVars:     []string{"NEBULA_PROTOCOLS_SET_CACHE_SIZE"},
				Value:       rootConfig.Database.ProtocolsSetCacheSize,
				Destination: &rootConfig.Database.ProtocolsSetCacheSize,
				Category:    flagCategoryCache,
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			CrawlCommand,
			MonitorCommand,
			ResolveCommand,
			NetworksCommand,
		},
	}

	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Infof("Received %s signal - Stopping...\n", sig.String())
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
	if rootConfig.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if c.IsSet("log-level") {
		ll := c.Int("log-level")
		log.SetLevel(log.Level(ll))
		if ll == int(log.TraceLevel) {
			boil.DebugMode = true
		}
	}

	switch strings.ToLower(c.String("log-format")) {
	case "text":
		log.SetFormatter(&log.TextFormatter{
			DisableColors: c.Bool("log-disable-color"),
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("unknown log format: %q", c.String("log-format"))
	}

	// Start prometheus metrics endpoint
	go metrics.ListenAndServe(rootConfig.TelemetryHost, rootConfig.TelemetryPort)

	return nil
}

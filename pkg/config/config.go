package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	// Prefix is used to determine the XDG config directory.
	Prefix = "nebula"

	// ContextKey represents the key under which the configuration
	// is put in a regular context.Context. The loaded configuration
	// is placed in the context and thus accessible everywhere where
	// a context is passed.
	ContextKey = "config"
)

// configFile contains the path suffix that's appended to
// an XDG compliant directory to find the settings file.
var configFile = filepath.Join(Prefix, "config.json")

// DefaultConfig the default configuration.
var DefaultConfig = Config{
	BootstrapPeers:     []string{}, // see init
	DialTimeout:        time.Minute,
	CrawlWorkerCount:   1000,
	CrawlLimit:         0,
	MonitorWorkerCount: 1000,
	MeasureLatencies:   false,
	MinPingInterval:    time.Second * 30,
	PingIntervalFactor: 1.2,
	PrometheusHost:     "0.0.0.0",
	PrometheusPort:     6666,
	DatabaseHost:       "0.0.0.0",
	DatabasePort:       5432,
	DatabaseName:       "nebula",
	DatabasePassword:   "password",
	DatabaseUser:       "nebula",
	Protocols:          []string{"/ipfs/kad/1.0.0", "/ipfs/kad/2.0.0"},
}

// Config contains general user configuration.
type Config struct {
	// The path where the configuration file is located.
	Path string `json:"-"`

	// Whether the configuration file existed when nebula was started
	Existed bool `json:"-"`

	// The list of multi addresses that will make up the entry points to the network.
	BootstrapPeers []string

	// The time to wait until a dial attempt is aborted.
	DialTimeout time.Duration

	// How many parallel workers should crawl the network.
	CrawlWorkerCount int

	// How many parallel workers should crawl the network.
	MonitorWorkerCount int

	// Only crawl the specified amount of peers
	CrawlLimit int

	// Whether the crawl task should measure and record latencies to peers
	MeasureLatencies bool

	// The minimum time interval between two consecutive visits of a peer
	MinPingInterval time.Duration

	// The factor with which the next ping timestamp should be calculated
	PingIntervalFactor float64

	// Determines the prometheus host bind to.
	PrometheusHost string

	// Determines the port where prometheus serves the metrics endpoint.
	PrometheusPort int

	// Determines the host address of the database.
	DatabaseHost string

	// Determines the port of the database.
	DatabasePort int

	// Determines the name of the database that should be used.
	DatabaseName string

	// Determines the password with which we access the database.
	DatabasePassword string

	// Determines the username with which we access the database.
	DatabaseUser string

	// The list of protocols that this crawler should look for.
	Protocols []string
}

func init() {
	for _, maddr := range dht.DefaultBootstrapPeers {
		DefaultConfig.BootstrapPeers = append(DefaultConfig.BootstrapPeers, maddr.String())
	}
}

// Save persists the configuration to disk using the `Path` field.
// Permissions will be 0o744
func (c *Config) Save() error {
	log.Infoln("Saving configuration file to", c.Path)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if c.Path == "" {
		c.Path, err = xdg.ConfigFile(configFile)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(c.Path, data, 0o744)
}

// Apply takes command line arguments and overwrites the respective configurations.
func (c *Config) Apply(ctx *cli.Context) {
	if ctx.IsSet("workers") {
		if ctx.Command.Name == "crawl" {
			c.CrawlWorkerCount = ctx.Int("workers")
		} else if ctx.Command.Name == "monitor" {
			c.MonitorWorkerCount = ctx.Int("workers")
		}
	}
	if ctx.IsSet("crawl-workers") {
		c.CrawlWorkerCount = ctx.Int("crawl-workers")
	}
	if ctx.IsSet("monitor-workers") {
		c.CrawlWorkerCount = ctx.Int("monitor-workers")
	}
	if ctx.IsSet("limit") {
		c.CrawlLimit = ctx.Int("limit")
	}
	if ctx.IsSet("dial-timeout") {
		c.DialTimeout = ctx.Duration("dial-timeout")
	}
	if ctx.IsSet("latencies") {
		c.MeasureLatencies = ctx.Bool("latencies")
	}
	if ctx.IsSet("min-ping-interval") {
		c.MinPingInterval = ctx.Duration("min-ping-interval")
	}
	if ctx.IsSet("ping-interval-factor") {
		c.PingIntervalFactor = ctx.Float64("ping-interval-factor")
	}
	if ctx.IsSet("prom-port") {
		c.PrometheusPort = ctx.Int("prom-port")
	}
	if ctx.IsSet("db-host") {
		c.DatabaseHost = ctx.String("db-host")
	}
	if ctx.IsSet("db-port") {
		c.DatabasePort = ctx.Int("db-port")
	}
	if ctx.IsSet("db-name") {
		c.DatabaseName = ctx.String("db-name")
	}
	if ctx.IsSet("db-password") {
		c.DatabasePassword = ctx.String("db-password")
	}
	if ctx.IsSet("db-user") {
		c.DatabaseUser = ctx.String("db-user")
	}
	if ctx.IsSet("protocols") {
		c.Protocols = ctx.StringSlice("protocols")
	}
}

// String prints the configuration as a json string
func (c *Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("%s", data)
}

// BootstrapAddrInfos parses the configured multi address strings to proper multi addresses.
func (c *Config) BootstrapAddrInfos() ([]peer.AddrInfo, error) {
	var pis []peer.AddrInfo
	for _, maddrStr := range c.BootstrapPeers {
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return nil, err
		}
		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}
		pis = append(pis, *pi)
	}
	return pis, nil
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// If no configuration file was given use xdg file.
		var err error
		path, err = xdg.ConfigFile(configFile)
		if err != nil {
			return nil, err
		}
	}

	log.Debugln("Loading configuration from:", path)
	config := DefaultConfig
	config.Path = path
	data, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
		config.Existed = true
		return &config, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	} else {
		return &config, config.Save()
	}
}

func FillContext(c *cli.Context) (context.Context, *Config, error) {
	conf, err := LoadConfig(c.String("config"))
	if err != nil {
		return c.Context, nil, err
	}

	// Apply command line argument configurations.
	conf.Apply(c)

	// Print full configuration.
	log.Debugln("Configuration (CLI params overwrite file config):\n", conf)

	// Populate the context with the configuration.
	return context.WithValue(c.Context, ContextKey, conf), conf, nil
}

func FromContext(ctx context.Context) (*Config, error) {
	obj := ctx.Value(ContextKey)
	if obj == nil {
		return nil, errors.New("config not found in context")
	}

	config, ok := obj.(*Config)
	if !ok {
		return nil, errors.New("config in context has wrong type")
	}

	return config, nil
}

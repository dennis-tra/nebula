package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	// Prefix is used to determine the XDG config directory.
	Prefix = "nebula"
)

type Network string

const (
	NetworkIPFS     Network = "IPFS"
	NetworkFilecoin Network = "FILECOIN"
	NetworkKusama   Network = "KUSAMA"
	NetworkPolkadot Network = "POLKADOT"
	NetworkRococo   Network = "ROCOCO"
	NetworkWestend  Network = "WESTEND"
)

// configFile contains the path suffix that's appended to
// an XDG compliant directory to find the settings file.
var configFile = filepath.Join(Prefix, "config.json")

// DefaultConfig the default configuration.
var DefaultConfig = Config{
	BootstrapPeers:         []string{}, // see init
	DialTimeout:            time.Minute,
	CrawlWorkerCount:       1000,
	CrawlLimit:             0,
	PersistNeighbors:       false,
	PingLimit:              0,
	PingWorkerCount:        1000,
	MonitorWorkerCount:     1000,
	MinPingInterval:        time.Second * 30,
	PingIntervalFactor:     1.2,
	TelemetryHost:          "0.0.0.0",
	TelemetryPort:          6666,
	DatabaseHost:           "0.0.0.0",
	DatabasePort:           5432,
	DatabaseName:           "nebula",
	DatabasePassword:       "password",
	DatabaseUser:           "nebula",
	DatabaseSSLMode:        "disable",
	Protocols:              []string{},
	RefreshRoutingTable:    false,
	AgentVersionsCacheSize: 200,
	ProtocolsCacheSize:     100,
	ProtocolsSetCacheSize:  200,
	FilePathUdgerDB:        "",
	Network:                NetworkIPFS,
}

// Config contains general user configuration.
type Config struct {
	// The version string of nebula
	Version string `json:"-"`

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

	// How many parallel workers should ping peers.
	PingWorkerCount int

	// How many parallel workers should crawl the network.
	MonitorWorkerCount int

	// Only crawl the specified amount of peers
	CrawlLimit int

	// Whether to persist all k-bucket entries
	PersistNeighbors bool

	// Whether to check if the Kubo API is exposed
	CheckExposed bool

	// Only ping the specified amount of peers
	PingLimit int

	// The minimum time interval between two consecutive visits of a peer
	MinPingInterval time.Duration

	// The factor with which the next ping timestamp should be calculated
	PingIntervalFactor float64

	// Determines where the prometheus and pprof hosts should bind to.
	TelemetryHost string

	// Determines the port where prometheus and pprof serve the metrics endpoint.
	TelemetryPort int

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

	// Postgres SSL mode (should be one supported in https://www.postgresql.org/docs/current/libpq-ssl.html)
	DatabaseSSLMode string

	// The list of protocols that this crawler should look for.
	Protocols []string

	// Whether the provider's routing table for should be refreshed
	RefreshRoutingTable bool

	// How many random content provide runs should be performed
	ProvideRunCount int

	// The directory where the measurement files should be saved
	ProvideOutDir string

	// The cache size to hold agent versions in memory to skip database queries.
	AgentVersionsCacheSize int

	// The cache size to hold protocols in memory to skip database queries.
	ProtocolsCacheSize int

	// The cache size to hold sets of protocols in memory to skip database queries.
	ProtocolsSetCacheSize int

	// File path to the udger datbase
	FilePathUdgerDB string

	// The network to crawl
	Network Network
}

// Init takes the command line argument and tries to read the config file from that directory.
func Init(c *cli.Context) (*Config, error) {
	conf, err := read(c.String("config"))
	if err != nil {
		return nil, errors.Wrap(err, "read config")
	}

	// Apply command line argument configurations.
	conf.apply(c)

	// Print full configuration.
	log.Debugln("Configuration (CLI params overwrite file config):\n", conf)

	// Populate the context with the configuration.
	return conf, nil
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

	return os.WriteFile(c.Path, data, 0o744)
}

// ReachedCrawlLimit returns true if the crawl limit is configured (aka != 0) and the crawled peers exceed this limit.
func (c *Config) ReachedCrawlLimit(crawled int) bool {
	return c.CrawlLimit > 0 && crawled >= c.CrawlLimit
}

// String prints the configuration as a json string
func (c *Config) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("%s", data)
}

// DatabaseSourceName returns the data source name string to be put into the sql.Open method.
func (c *Config) DatabaseSourceName() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseSSLMode,
	)
}

// BootstrapAddrInfos parses the configured multi address strings to proper multi addresses.
func (c *Config) BootstrapAddrInfos() ([]peer.AddrInfo, error) {
	peerAddrs := map[peer.ID][]ma.Multiaddr{}
	for _, maddrStr := range c.BootstrapPeers {

		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return nil, err
		}

		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}

		_, found := peerAddrs[pi.ID]
		if found {
			peerAddrs[pi.ID] = append(peerAddrs[pi.ID], pi.Addrs...)
		} else {
			peerAddrs[pi.ID] = pi.Addrs
		}
	}

	var pis []peer.AddrInfo
	for pid, addrs := range peerAddrs {
		pi := peer.AddrInfo{
			ID:    pid,
			Addrs: addrs,
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func read(path string) (*Config, error) {
	if path == "" {
		// If no configuration file was given use xdg file.
		var err error
		path, err = xdg.ConfigFile(configFile)
		if err != nil {
			return nil, err
		}
	}

	log.Infoln("Loading configuration from:", path)
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

// apply takes command line arguments and overwrites the respective configurations.
func (c *Config) apply(ctx *cli.Context) {
	c.Version = ctx.App.Version

	if ctx.IsSet("workers") {
		if ctx.Command.Name == "crawl" {
			c.CrawlWorkerCount = ctx.Int("workers")
		} else if ctx.Command.Name == "monitor" {
			c.MonitorWorkerCount = ctx.Int("workers")
		} else if ctx.Command.Name == "ping" {
			c.PingWorkerCount = ctx.Int("workers")
		}
	}
	if ctx.IsSet("limit") {
		if ctx.Command.Name == "crawl" {
			c.CrawlLimit = ctx.Int("limit")
		} else if ctx.Command.Name == "ping" {
			c.PingLimit = ctx.Int("limit")
		}
	}
	if ctx.IsSet("dial-timeout") {
		c.DialTimeout = ctx.Duration("dial-timeout")
	}
	if ctx.IsSet("neighbors") {
		c.PersistNeighbors = ctx.Bool("neighbors")
	}
	if ctx.IsSet("check-exposed") {
		c.CheckExposed = ctx.Bool("check-exposed")
	}
	if ctx.IsSet("min-ping-interval") {
		c.MinPingInterval = ctx.Duration("min-ping-interval")
	}
	if ctx.IsSet("ping-interval-factor") {
		c.PingIntervalFactor = ctx.Float64("ping-interval-factor")
	}
	if ctx.IsSet("telemetry-host") {
		c.TelemetryHost = ctx.String("telemetry-host")
	}
	if ctx.IsSet("telemetry-port") {
		c.TelemetryPort = ctx.Int("telemetry-port")
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
	if ctx.IsSet("db-sslmode") {
		c.DatabaseSSLMode = ctx.String("db-sslmode")
	}
	if ctx.IsSet("init-rt") {
		c.RefreshRoutingTable = ctx.Bool("init-rt")
	}
	if ctx.IsSet("run-count") {
		c.ProvideRunCount = ctx.Int("run-count")
	}
	if ctx.IsSet("out") {
		c.ProvideOutDir = ctx.String("out")
	}
	if ctx.IsSet("agent-versions-cache-size") {
		c.AgentVersionsCacheSize = ctx.Int("agent-versions-cache-size")
	}
	if ctx.IsSet("protocols-cache-size") {
		c.ProtocolsCacheSize = ctx.Int("protocols-cache-size")
	}
	if ctx.IsSet("protocols-set-cache-size") {
		c.ProtocolsSetCacheSize = ctx.Int("protocols-set-cache-size")
	}
	if ctx.IsSet("udger-db") {
		c.FilePathUdgerDB = ctx.String("udger-db")
	}

	if ctx.IsSet("network") {
		c.Network = Network(ctx.String("network"))
		c.configureNetwork()
	} else if len(DefaultConfig.BootstrapPeers) == 0 {
		c.configureNetwork()

		if ctx.IsSet("protocols") {
			c.Protocols = ctx.StringSlice("protocols")
		}
	}

	// Give CLI option precedence
	if ctx.IsSet("bootstrap-peers") {
		c.BootstrapPeers = ctx.StringSlice("bootstrap-peers")
	}
}

func (c *Config) configureNetwork() {
	switch c.Network {
	case NetworkFilecoin:
		c.BootstrapPeers = BootstrapPeersFilecoin
		c.Protocols = []string{"/fil/kad/testnetnet/kad/1.0.0"}
	case NetworkKusama:
		c.BootstrapPeers = BootstrapPeersKusama
		c.Protocols = []string{"/ksmcc3/kad"}
	case NetworkPolkadot:
		c.BootstrapPeers = BootstrapPeersPolkadot
		c.Protocols = []string{"/dot/kad"}
	case NetworkRococo:
		c.BootstrapPeers = BootstrapPeersRococo
		c.Protocols = []string{"/rococo/kad"}
	case NetworkWestend:
		c.BootstrapPeers = BootstrapPeersWestend
		c.Protocols = []string{"/wnd2/kad"}
	case NetworkIPFS:
		fallthrough
	default:
		c.BootstrapPeers = []string{}
		for _, maddr := range dht.DefaultBootstrapPeers {
			c.BootstrapPeers = append(c.BootstrapPeers, maddr.String())
		}
		c.Protocols = []string{"/ipfs/kad/1.0.0"}
	}
}

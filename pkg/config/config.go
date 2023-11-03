package config

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/urfave/cli/v2"
)

type Network string

const (
	NetworkIPFS     Network = "IPFS"
	NetworkAmino    Network = "AMINO"
	NetworkFilecoin Network = "FILECOIN"
	NetworkKusama   Network = "KUSAMA"
	NetworkPolkadot Network = "POLKADOT"
	NetworkRococo   Network = "ROCOCO"
	NetworkWestend  Network = "WESTEND"
	NetworkArabica  Network = "ARABICA"
	NetworkMocha    Network = "MOCHA"
	NetworkBlockRa  Network = "BLOCKSPACE_RACE"
	NetworkEthCons  Network = "ETHEREUM_CONSENSUS"
)

func Networks() []Network {
	return []Network{
		NetworkIPFS,
		NetworkAmino,
		NetworkFilecoin,
		NetworkKusama,
		NetworkPolkadot,
		NetworkRococo,
		NetworkWestend,
		NetworkArabica,
		NetworkMocha,
		NetworkBlockRa,
		NetworkEthCons,
	}
}

type AddrType string

const (
	AddrTypePrivate AddrType = "private"
	AddrTypePublic  AddrType = "public"
	AddrTypeAny     AddrType = "any"
)

// Root contains general user configuration.
type Root struct {
	// The version string of nebula
	RawVersion string

	// Enables debug logging (equivalent to log level 5)
	Debug bool

	// Specific log level from 0 (least verbose) to 6 (most verbose)
	LogLevel int

	// Specify the log format (text or json)
	LogFormat string

	// Whether to have colorized log output
	LogDisableColor bool

	// The time to wait until a dial attempt is aborted.
	DialTimeout time.Duration

	// Determines where the prometheus and pprof hosts should bind to.
	TelemetryHost string

	// Determines the port where prometheus and pprof serve the metrics endpoint.
	TelemetryPort int

	// Contains all configuration parameters for interacting with the database
	Database *Database
}

// Version returns the actual version string which includes VCS information
func (r *Root) Version() string {
	shortCommit := ""
	dirty := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				shortCommit = setting.Value[:7]
			case "vcs.modified":
				dirty, _ = strconv.ParseBool(setting.Value)
			}
		}
	}

	versionStr := "v" + r.RawVersion

	if shortCommit != "" {
		versionStr += "+" + shortCommit
	}

	if dirty {
		versionStr += "+dirty"
	}

	return versionStr
}

// String prints the configuration as a json string
func (r *Root) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return fmt.Sprintf("%s", data)
}

type Database struct {
	// Whether to skip database interactions
	DryRun bool

	// File path to the JSON output directory
	JSONOut string

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

	// The cache size to hold agent versions in memory to skip database queries.
	AgentVersionsCacheSize int

	// The cache size to hold protocols in memory to skip database queries.
	ProtocolsCacheSize int

	// The cache size to hold sets of protocols in memory to skip database queries.
	ProtocolsSetCacheSize int
}

// DatabaseSourceName returns the data source name string to be put into the sql.Open method.
func (c *Database) DatabaseSourceName() string {
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

// Crawl contains general user configuration.
type Crawl struct {
	Root *Root

	// The list of multi addresses that will make up the entry points to the network.
	BootstrapPeers *cli.StringSlice

	// The list of protocols that this crawler should look for.
	Protocols *cli.StringSlice

	// How many parallel workers should crawl the network.
	CrawlWorkerCount int

	// How many parallel workers should write crawl results to the database
	WriteWorkerCount int

	// Only crawl the specified amount of peers
	CrawlLimit int

	// Whether to persist all k-bucket entries
	PersistNeighbors bool

	// Whether to check if the Kubo API is exposed
	CheckExposed bool

	// File path to the udger datbase
	FilePathUdgerDB string

	// The network to crawl
	Network string

	// Which type addresses should be stored to the database (private, public, both)
	AddrTrackTypeStr string

	// Which type of addresses should Nebula try to dial (private, public, both)
	AddrDialTypeStr string
}

func (c *Crawl) AddrTrackType() AddrType {
	return AddrType(c.AddrTrackTypeStr)
}

func (c *Crawl) AddrDialType() AddrType {
	return AddrType(c.AddrDialTypeStr)
}

// String prints the configuration as a json string
func (c *Crawl) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func ConfigureNetwork(network string) (*cli.StringSlice, *cli.StringSlice, error) {
	var (
		bootstrapPeers *cli.StringSlice
		protocols      *cli.StringSlice
	)

	switch Network(network) {
	case NetworkFilecoin:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersFilecoin...)
		protocols = cli.NewStringSlice("/fil/kad/testnetnet/kad/1.0.0")
	case NetworkKusama:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersKusama...)
		protocols = cli.NewStringSlice("/ksmcc3/kad")
	case NetworkPolkadot:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersPolkadot...)
		protocols = cli.NewStringSlice("/dot/kad")
	case NetworkRococo:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersRococo...)
		protocols = cli.NewStringSlice("/rococo/kad")
	case NetworkWestend:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersWestend...)
		protocols = cli.NewStringSlice("/wnd2/kad")
	case NetworkArabica:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersArabica...)
		protocols = cli.NewStringSlice("/celestia/arabica-6/kad/1.0.0")
	case NetworkMocha:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersMocha...)
		protocols = cli.NewStringSlice("/celestia/mocha/kad/1.0.0")
	case NetworkBlockRa:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersBlockspaceRace...)
		protocols = cli.NewStringSlice("/celestia/blockspacerace-0/kad/1.0.0")
	case NetworkEthCons:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersEthereumConsensus...)
		protocols = cli.NewStringSlice("discv5") // TODO
	case NetworkIPFS, NetworkAmino:
		bps := []string{}
		for _, maddr := range kaddht.DefaultBootstrapPeers {
			bps = append(bps, maddr.String())
		}
		bootstrapPeers = cli.NewStringSlice(bps...)
		protocols = cli.NewStringSlice(string(kaddht.ProtocolDHT))
	default:
		return nil, nil, fmt.Errorf("unknown network identifier: %s", network)
	}

	return bootstrapPeers, protocols, nil
}

type Monitor struct {
	Root *Root

	// How many parallel workers should crawl the network.
	MonitorWorkerCount int

	// How many parallel workers should write crawl results to the database
	WriteWorkerCount int

	// The network to crawl
	Network string

	// The list of protocols that this crawler should look for.
	Protocols *cli.StringSlice
}

// String prints the configuration as a json string
func (m *Monitor) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

type Resolve struct {
	Root *Root

	FilePathUdgerDB string
	BatchSize       int
}

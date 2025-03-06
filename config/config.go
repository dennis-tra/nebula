package config

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/p2p/discover/v5wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	ltcwire "github.com/ltcsuite/ltcd/wire"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/db"
)

type Network string

const (
	NetworkIPFS           Network = "IPFS"
	NetworkAmino          Network = "AMINO"
	NetworkFilecoin       Network = "FILECOIN"
	NetworkKusama         Network = "KUSAMA"
	NetworkPolkadot       Network = "POLKADOT"
	NetworkRococo         Network = "ROCOCO"
	NetworkWestend        Network = "WESTEND"
	NetworkCelestia       Network = "CELESTIA"
	NetworkArabica        Network = "ARABICA"
	NetworkMocha          Network = "MOCHA"
	NetworkBlockRa        Network = "BLOCKSPACE_RACE"
	NetworkEthCons        Network = "ETHEREUM_CONSENSUS"
	NetworkEthExec        Network = "ETHEREUM_EXECUTION"
	NetworkHolesky        Network = "HOLESKY"
	NetworkPortal         Network = "PORTAL"
	NetworkAvailMainnetFN Network = "AVAIL_MAINNET_FN"
	NetworkAvailMainnetLC Network = "AVAIL_MAINNET_LC"
	NetworkAvailTuringLC  Network = "AVAIL_TURING_LC"
	NetworkAvailTuringFN  Network = "AVAIL_TURING_FN"
	NetworkPactus         Network = "PACTUS"
	NetworkBitcoin        Network = "BITCOIN"
	NetworkLitecoin       Network = "LITECOIN"
	NetworkDoge           Network = "DOGE"
	NetworkDria           Network = "DRIA"
	NetworkWakuStatus     Network = "WAKU_STATUS"
	NetworkWakuTWN        Network = "WAKU_TWN"
	NetworkGnosis         Network = "GNOSIS"
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
		NetworkCelestia,
		NetworkArabica,
		NetworkMocha,
		NetworkBlockRa,
		NetworkEthCons,
		NetworkEthExec,
		NetworkHolesky,
		NetworkPortal,
		NetworkAvailMainnetFN,
		NetworkAvailMainnetLC,
		NetworkAvailTuringLC,
		NetworkAvailTuringFN,
		NetworkPactus,
		NetworkBitcoin,
		NetworkLitecoin,
		NetworkDoge,
		NetworkDria,
		NetworkWakuStatus,
		NetworkWakuTWN,
		NetworkGnosis,
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
	// Enables debug logging (equivalent to log level 5)
	Debug bool

	// Specific log level from 0 (least verbose) to 6 (most verbose)
	LogLevel int

	// Specify the log format (text or json)
	LogFormat string

	// Whether to log the full error string
	LogErrors bool

	// Whether to have colorized log output
	LogDisableColor bool

	// The time to wait until a dial attempt is aborted.
	DialTimeout time.Duration

	// Determines where the prometheus and pprof hosts should bind to.
	MetricsHost string

	// Determines the port where prometheus and pprof serve the metrics endpoint.
	MetricsPort int

	// Host of the trace collector like Jaeger
	TracesHost string

	// Port of the trace collector
	TracesPort int

	// Contains all configuration parameters for interacting with the database
	Database *Database

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider

	// The buffer size of the UDP sockets (applicable to ETHEREUM_{CONSENSUS,EXECUTION)
	UDPBufferSize int

	// The raw version of Nebula in the for X.Y.Z. Raw, because it's missing, e.g., commit information (set by GoReleaser or in Makefile)
	RawVersion string

	// The commit hash used to build the Nebula binary (set by GoReleaser or in Makefile)
	BuildCommit string

	// The date when Nebula was built (set by GoReleaser or in Makefile)
	BuildDate string

	// Who built Nebula (set by GoReleaser or in Makefile)
	BuiltBy string
}

// Version returns the actual version string which includes VCS information
func (r *Root) Version() string {
	shortCommit := ""
	dirty := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				shortCommit = setting.Value
				if len(shortCommit) > 8 {
					shortCommit = shortCommit[:8]
				}
			case "vcs.modified":
				dirty, _ = strconv.ParseBool(setting.Value)
			}
		}
	}

	versionStr := r.RawVersion

	if r.BuildCommit != "" {
		if len(r.BuildCommit) > 8 {
			r.BuildCommit = r.BuildCommit[:8]
		}
		shortCommit = r.BuildCommit
	}

	if !strings.HasSuffix(versionStr, shortCommit) {
		versionStr += "-" + shortCommit
	}

	if dirty {
		versionStr += "+dirty"
	}

	return versionStr
}

func (r *Root) BuildAuthor() string {
	if r.BuildDate != "" && r.BuiltBy != "" {
		return fmt.Sprintf("built at %s by %s", r.BuildDate, r.BuiltBy)
	} else if r.BuildDate != "" {
		return fmt.Sprintf("built at %s", r.BuildDate)
	} else if r.BuiltBy != "" {
		return fmt.Sprintf("built by %s", r.BuiltBy)
	}
	return ""
}

// String prints the configuration as a json string
func (r *Root) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return fmt.Sprintf("%s", data)
}

type Database struct {
	// Whether to skip database interactions
	DryRun bool

	// The network identifier that we are collecting data for
	NetworkID string

	// File path to the JSON output directory
	JSONOut string

	// Determines the database engine to which the data should be written
	DatabaseEngine string

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

	// The database SSL configuration. For Postgres SSL mode should be
	// one of the supported values here: https://www.postgresql.org/docs/current/libpq-ssl.html)
	// For clickhouse only a yes or no value is supported
	DatabaseSSL string

	// Whether to apply the database migrations on startup
	ApplyMigrations bool

	// Name of the cluster for creating the migrations table cluster wide
	ClickHouseClusterName string

	// Engine to use for the migrations table, defaults to TinyLog
	ClickHouseMigrationsTableEngine string

	// The maximum number of records to hold in memory before flushing the data to clickhouse
	ClickHouseBatchSize int

	// The maximum time to hold records in memory before flushing the data to clickhouse
	ClickHouseBatchInterval time.Duration

	// Whether to use the replicated merge tree engine variants for migrations
	ClickHouseReplicatedTableEngines bool

	// The cache size to hold agent versions in memory to skip database queries.
	AgentVersionsCacheSize int

	// The cache size to hold protocols in memory to skip database queries.
	ProtocolsCacheSize int

	// The cache size to hold sets of protocols in memory to skip database queries.
	ProtocolsSetCacheSize int

	// Set the maximum idle connections for the database handler.
	MaxIdleConns int

	// Whether to store the routing table of the entire network
	PersistNeighbors bool

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
}

func (cfg *Database) PostgresClientConfig() *db.PostgresClientConfig {
	if cfg.DatabasePort == 0 {
		cfg.DatabasePort = 5432
	}

	if cfg.DatabaseSSL == "" {
		cfg.DatabaseSSL = "disable"
	}

	return &db.PostgresClientConfig{
		DatabaseHost:           cfg.DatabaseHost,
		DatabasePort:           cfg.DatabasePort,
		DatabaseName:           cfg.DatabaseName,
		DatabasePassword:       cfg.DatabasePassword,
		DatabaseUser:           cfg.DatabaseUser,
		DatabaseSSL:            cfg.DatabaseSSL,
		ApplyMigrations:        cfg.ApplyMigrations,
		AgentVersionsCacheSize: cfg.AgentVersionsCacheSize,
		ProtocolsCacheSize:     cfg.ProtocolsCacheSize,
		ProtocolsSetCacheSize:  cfg.ProtocolsSetCacheSize,
		MaxIdleConns:           cfg.MaxIdleConns,
		PersistNeighbors:       cfg.PersistNeighbors,
		MeterProvider:          cfg.MeterProvider,
		TracerProvider:         cfg.TracerProvider,
	}
}

func (cfg *Database) ClickHouseClientConfig() *db.ClickHouseClientConfig {
	if cfg.DatabasePort == 0 {
		cfg.DatabasePort = 9000
	}

	databaseSSL := false
	switch strings.ToLower(cfg.DatabaseSSL) {
	case "yes", "true", "1":
		databaseSSL = true
	}

	return &db.ClickHouseClientConfig{
		DatabaseHost:           cfg.DatabaseHost,
		DatabasePort:           cfg.DatabasePort,
		DatabaseName:           cfg.DatabaseName,
		DatabaseUser:           cfg.DatabaseUser,
		DatabasePassword:       cfg.DatabasePassword,
		DatabaseSSL:            databaseSSL,
		ClusterName:            cfg.ClickHouseClusterName,
		MigrationsTableEngine:  cfg.ClickHouseMigrationsTableEngine,
		ReplicatedTableEngines: cfg.ClickHouseReplicatedTableEngines,
		ApplyMigrations:        cfg.ApplyMigrations,
		BatchSize:              cfg.ClickHouseBatchSize,
		BatchTimeout:           cfg.ClickHouseBatchInterval,
		NetworkID:              cfg.NetworkID,
		PersistNeighbors:       cfg.PersistNeighbors,
		MeterProvider:          cfg.MeterProvider,
		TracerProvider:         cfg.TracerProvider,
	}
}

// NewClient will initialize the right database client based on the given
// configuration. This can either be a Postgres, ClickHouse, JSON, or noop
// client. The noop client is a dummy implementation of the [Client] interface
// that does nothing when the methods are called. That's the one used if the
// user specifies `--dry-run` on the command line. The JSON client is used when
// the user specifies a JSON output directory. Then JSON files with crawl
// information are written to that directory. In any other case, the Postgres
// or ClickHouse client is used based on the configured database engine.
func (cfg *Database) NewClient(ctx context.Context) (db.Client, error) {
	var (
		dbc db.Client
		err error
	)

	// dry run has precedence. Then, if a JSON output directory is given, use
	// the JSON client. In any other case, use the one configured via the engine
	// command line flag client.
	if cfg.DryRun {
		dbc = db.NewNoopClient()
	} else if cfg.JSONOut != "" {
		dbc, err = db.NewJSONClient(cfg.JSONOut)
	} else {
		switch strings.ToLower(cfg.DatabaseEngine) {
		case "postgres", "pg":
			dbc, err = db.NewPostgresClient(ctx, cfg.PostgresClientConfig())
		case "clickhouse", "ch":
			dbc, err = db.NewClickHouseClient(ctx, cfg.ClickHouseClientConfig())
		default:
			return nil, fmt.Errorf("unknown database engine: %s", cfg.DatabaseEngine)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("init db client: %w", err)
	}

	return dbc, nil
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

	// File path to the udger datbase
	FilePathUdgerDB string

	// The network to crawl
	Network string

	// Which type addresses should be stored to the database (private, public, both)
	AddrTrackTypeStr string

	// Which type of addresses should Nebula try to dial (private, public, both)
	AddrDialTypeStr string

	// Whether to check if the Kubo API is exposed
	CheckExposed bool

	// Whether to keep the full enr record alongside all parsed kv-pairs
	KeepENR bool

	// The UDP response timeout when crawling the discv4 and discv5 DHTs
	UDPRespTimeout time.Duration

	// WakuClusterID represents the identifier for the Waku cluster.
	WakuClusterID int

	// WakuClusterShards defines shard indices for Waku cluster operations.
	WakuClusterShards *cli.IntSlice
}

func (c *Crawl) AddrTrackType() AddrType {
	return AddrType(c.AddrTrackTypeStr)
}

func (c *Crawl) AddrDialType() AddrType {
	return AddrType(c.AddrDialTypeStr)
}

func (c *Crawl) BootstrapAddrInfos() ([]peer.AddrInfo, error) {
	addrInfoMap := map[peer.ID][]ma.Multiaddr{}
	for _, maddrStr := range c.BootstrapPeers.Value() {

		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return nil, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		}

		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, fmt.Errorf("parse addr info from maddr %s: %w", maddr, err)
		}

		_, found := addrInfoMap[pi.ID]
		if found {
			addrInfoMap[pi.ID] = append(addrInfoMap[pi.ID], pi.Addrs...)
		} else {
			addrInfoMap[pi.ID] = pi.Addrs
		}
	}

	addrInfos := make([]peer.AddrInfo, 0, len(addrInfoMap))
	for pid, maddrs := range addrInfoMap {
		addrInfos = append(addrInfos, peer.AddrInfo{
			ID:    pid,
			Addrs: maddrs,
		})
	}

	return addrInfos, nil
}

func (c *Crawl) BootstrapEnodesV4() ([]*enode.Node, error) {
	nodesMap := map[enode.ID]*enode.Node{}
	for _, url := range c.BootstrapPeers.Value() {
		n, err := enode.ParseV4(url)
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap enode URL %s: %w", url, err)
		}
		nodesMap[n.ID()] = n
	}

	enodes := make([]*enode.Node, 0, len(nodesMap))
	for _, node := range nodesMap {
		enodes = append(enodes, node)
	}

	return enodes, nil
}

func (c *Crawl) BootstrapEnodesV5() ([]*enode.Node, error) {
	nodesMap := map[enode.ID]*enode.Node{}
	for _, enr := range c.BootstrapPeers.Value() {
		n, err := enode.Parse(enode.ValidSchemes, enr)
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap enr: %w", err)
		}
		nodesMap[n.ID()] = n
	}

	enodes := make([]*enode.Node, 0, len(nodesMap))
	for _, node := range nodesMap {
		enodes = append(enodes, node)
	}

	return enodes, nil
}

func (c *Crawl) DiscV5ProtocolID() ([6]byte, error) {
	protocols := c.Protocols.Value()
	if len(protocols) != 1 {
		return [6]byte{}, fmt.Errorf("invalid number of protocol IDs configured: %d", len(protocols))
	}

	protocolStr := protocols[0]
	if len(protocolStr) != 6 {
		return [6]byte{}, fmt.Errorf("invalid length of protocol ID %q: %d", protocolStr, len(protocolStr))
	}

	var protocolID [6]byte
	copy(protocolID[:], protocolStr)
	return protocolID, nil
}

// WakuClusterConfig doesn't return an error if the network isn't a waku network
func (c *Crawl) WakuClusterConfig() (uint32, []uint32) {
	if c.WakuClusterID != 0 && len(c.WakuClusterShards.Value()) != 0 {
		shards := make([]uint32, len(c.WakuClusterShards.Value()))
		for i, shard := range c.WakuClusterShards.Value() {
			shards[i] = uint32(shard)
		}
		return uint32(c.WakuClusterID), shards
	}

	switch Network(c.Network) {
	case NetworkWakuStatus:
		return 16, []uint32{32, 64, 128}
	case NetworkWakuTWN:
		return 1, []uint32{0, 1, 2, 3, 4, 5, 6, 7}
	default:
		return 0, []uint32{}
	}
}

// String prints the configuration as a json string
func (c *Crawl) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func BitcoinNetwork(network string) wire.BitcoinNet {
	switch Network(network) {
	case NetworkBitcoin:
		return wire.MainNet
	case NetworkLitecoin:
		// https://github.com/ltcsuite/ltcd/blob/ec0bfad6e7104120d872c89fb1fa7aa7033a4ae2/wire/protocol.go#L151
		return wire.BitcoinNet(ltcwire.MainNet)
	case NetworkDoge:
		// https://github.com/dogecoin/libdohj/blob/3dd6d7548843fe1891366e0325a05587eb21d1aa/core/src/main/java/org/libdohj/params/DogecoinMainNetParams.java#L40
		return wire.BitcoinNet(0xc0c0c0c0)
	default:
		panic(fmt.Sprintf("unknown network: %s", network))
	}
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
	case NetworkCelestia:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersCelestia...)
		protocols = cli.NewStringSlice("/celestia/celestia/kad/1.0.0")
	case NetworkArabica:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersArabica...)
		protocols = cli.NewStringSlice("/celestia/arabica-10/kad/1.0.0") // the `-10` suffix seems to be variable
	case NetworkMocha:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersMocha...)
		protocols = cli.NewStringSlice("/celestia/mocha-4/kad/1.0.0") // the `-4` suffix seems to be variable
	case NetworkBlockRa:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersBlockspaceRace...)
		protocols = cli.NewStringSlice("/celestia/blockspacerace-0/kad/1.0.0")
	case NetworkEthCons:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersEthereumConsensus...)
		protocols = cli.NewStringSlice(string(v5wire.DefaultProtocolID[:]))
	case NetworkGnosis:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersGnosis...)
		protocols = cli.NewStringSlice(string(v5wire.DefaultProtocolID[:]))
	case NetworkEthExec:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersEthereumExecution...)
		protocols = cli.NewStringSlice("discv4") // TODO
	case NetworkHolesky:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersHolesky...)
		protocols = cli.NewStringSlice(string(v5wire.DefaultProtocolID[:]))
	case NetworkPortal:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersPortalMainnet...)
		protocols = cli.NewStringSlice(string(v5wire.DefaultProtocolID[:]))
	case NetworkAvailMainnetFN:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersAvailMainnetFullNode...)
		protocols = cli.NewStringSlice("/Avail/kad")
	case NetworkAvailMainnetLC:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersAvailMainnetLightClient...)
		protocols = cli.NewStringSlice("/avail_kad/id/1.0.0-b91746")
	case NetworkAvailTuringLC:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersAvailTuringLightClient...)
		protocols = cli.NewStringSlice("/avail_kad/id/1.0.0-6f0996")
	case NetworkAvailTuringFN:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersAvailTuringFullNode...)
		protocols = cli.NewStringSlice("/Avail/kad")
	case NetworkDria:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersDria...)
		protocols = cli.NewStringSlice("/dria/kad/0.2")
	case NetworkIPFS, NetworkAmino:
		bps := []string{}
		for _, maddr := range kaddht.DefaultBootstrapPeers {
			bps = append(bps, maddr.String())
		}
		bootstrapPeers = cli.NewStringSlice(bps...)
		protocols = cli.NewStringSlice(string(kaddht.ProtocolDHT))
	case NetworkPactus:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersPactusFullNode...)
		protocols = cli.NewStringSlice("/pactus/gossip/v1/kad/1.0.0")
	case NetworkBitcoin:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersBitcoinMainnet...)
		protocols = cli.NewStringSlice("bitcoin")
	case NetworkLitecoin:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersLitecoinMainnet...)
		protocols = cli.NewStringSlice("litecoin")
	case NetworkDoge:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersDogeMainnet...)
		protocols = cli.NewStringSlice("doge")
	case NetworkWakuStatus:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersWakuStatus...)
		protocols = cli.NewStringSlice("d5waku")
	case NetworkWakuTWN:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersWakuTWN...)
		protocols = cli.NewStringSlice("d5waku")
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

	BatchSize              int
	FilePathUdgerDB        string
	FilePathMaxmindCountry string
	FilePathMaxmindASN     string
}

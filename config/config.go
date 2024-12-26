package config

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/p2p/enode"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
	NetworkAvailMainnetFN Network = "AVAIL_MAINNET_FN"
	NetworkAvailMainnetLC Network = "AVAIL_MAINNET_LC"
	NetworkAvailTuringLC  Network = "AVAIL_TURING_LC"
	NetworkAvailTuringFN  Network = "AVAIL_TURING_FN"
	NetworkPactus         Network = "PACTUS"
	NetworkBitcoin        Network = "BITCOIN"
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
		NetworkAvailMainnetFN,
		NetworkAvailMainnetLC,
		NetworkAvailTuringLC,
		NetworkAvailTuringFN,
		NetworkPactus,
		NetworkBitcoin,
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

	// Set the maximum idle connections for the database handler.
	MaxIdleConns int

	// MeterProvider is the meter provider to use when initialising metric instruments.
	MeterProvider metric.MeterProvider

	// TracerProvider is the tracer provider to use when initialising tracing
	TracerProvider trace.TracerProvider
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

func GetSeedsFromDNS(dnsSeeds []string) []string {
	wait := sync.WaitGroup{}
	results := make(chan []net.IP)

	for _, seed := range dnsSeeds {
		wait.Add(1)
		go func(address string) {
			defer wait.Done()
			ips, err := net.LookupIP(address)
			if err != nil {
				return
			}
			results <- ips
		}(seed)
	}

	go func() {
		wait.Wait()
		close(results)
	}()

	seeds := []string{}
	for ips := range results {
		for _, ip := range ips {
			seeds = append(seeds, net.JoinHostPort(ip.String(), chaincfg.MainNetParams.DefaultPort))
		}
	}

	// Note that this will likely include duplicate seeds. The crawler deduplicates them.
	return seeds
}

func (c *Crawl) BootstrapBitcoinEntries() ([]ma.Multiaddr, error) {
	addresses := GetSeedsFromDNS(c.BootstrapPeers.Value())
	addrInfos := make([]ma.Multiaddr, 0, len(addresses))

	for _, addr := range addresses {
		maddr := toMultiAddr(addr)
		// if err != nil {
		// 	return nil, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		// }

		// Directly append to the slice
		if maddr != nil {
			addrInfos = append(addrInfos, maddr)
		}
	}

	return addrInfos, nil
}

func toMultiAddr(addr string) ma.Multiaddr {

	ip_type := "4"
	if addr[0] == '[' {
		ip_type = "6"
	}

	parts := strings.Split(addr, ":")
	ip_addr := strings.Join(parts[:len(parts)-1], ":")
	port := parts[len(parts)-1]

	address := fmt.Sprintf("/ip%s/%s/tcp/%s", ip_type, ip_addr, port)
	multiAddr, _ := ma.NewMultiaddr(address)
	return multiAddr
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
		protocols = cli.NewStringSlice("discv5") // TODO
	case NetworkEthExec:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersEthereumExecution...)
		protocols = cli.NewStringSlice("discv4") // TODO
	case NetworkHolesky:
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersHolesky...)
		protocols = cli.NewStringSlice("discv5") // TODO
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
		bootstrapPeers = cli.NewStringSlice(BootstrapPeersBitcoinDNSSeeds...)
		protocols = cli.NewStringSlice("bitcoin")
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

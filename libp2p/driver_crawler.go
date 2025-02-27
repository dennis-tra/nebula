package libp2p

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"time"

	kbucket "github.com/libp2p/go-libp2p-kbucket"

	"github.com/libp2p/go-libp2p/p2p/net/swarm"

	"github.com/libp2p/go-libp2p"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	ma "github.com/multiformats/go-multiaddr"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/kubo"
	"github.com/dennis-tra/nebula-crawler/utils"
)

// Host is the interface that's required for crawling libp2p peers. Actually
// the *basichost.Host is required but to allow testing we define this interface
// here. That allows us to inject a mock host.
type Host interface {
	host.Host
	IDService() identify.IDService
}

type PeerInfo struct {
	peer.AddrInfo
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func (p PeerInfo) ID() peer.ID {
	return p.AddrInfo.ID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.AddrInfo.Addrs
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	if p.AddrInfo.ID != other.AddrInfo.ID {
		panic("merge peer ID mismatch")
	}

	return PeerInfo{
		AddrInfo: peer.AddrInfo{
			ID:    p.AddrInfo.ID,
			Addrs: utils.MergeMaddrs(p.AddrInfo.Addrs, other.AddrInfo.Addrs),
		},
	}
}

func (p PeerInfo) DiscoveryPrefix() uint64 {
	kadID := kbucket.ConvertPeerID(p.AddrInfo.ID)
	return binary.BigEndian.Uint64(kadID[:8])
}

func (p PeerInfo) DeduplicationKey() string {
	return p.AddrInfo.ID.String()
}

type CrawlDriverConfig struct {
	Version        string
	Network        config.Network
	Protocols      []string
	DialTimeout    time.Duration
	CheckExposed   bool
	BootstrapPeers []peer.AddrInfo
	AddrTrackType  config.AddrType
	AddrDialType   config.AddrType
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
	LogErrors      bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	crawlerCfg := DefaultCrawlerConfig()
	crawlerCfg.DialTimeout = cfg.DialTimeout
	crawlerCfg.CheckExposed = cfg.CheckExposed
	crawlerCfg.AddrDialType = cfg.AddrDialType
	crawlerCfg.LogErrors = cfg.LogErrors
	return crawlerCfg
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{
		AddrTrackType: cfg.AddrTrackType,
	}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	hosts        []Host
	dbc          db.Client
	tasksChan    chan PeerInfo
	crawlerCount int
	writerCount  int
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	// The Avail light clients verify the agent version:
	// https://github.com/availproject/avail-light/blob/0ddc5d50d6f3d7217c448d6d008846c6b8c4fec3/src/network/p2p/event_loop.rs#L296
	// Spoof it
	userAgent := "nebula/" + cfg.Version
	if cfg.Network == config.NetworkAvailTuringLC || cfg.Network == config.NetworkAvailMainnetLC {
		userAgent = "avail-light-client/rust-client"
	}

	hosts := make([]Host, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		h, err := newLibp2pHost(userAgent)
		if err != nil {
			return nil, fmt.Errorf("new libp2p host: %w", err)
		}
		hosts = append(hosts, h)
	}

	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, addrInfo := range cfg.BootstrapPeers {
		addrInfo := addrInfo
		tasksChan <- PeerInfo{AddrInfo: addrInfo}
	}
	close(tasksChan)

	return &CrawlDriver{
		cfg:          cfg,
		hosts:        hosts,
		dbc:          dbc,
		tasksChan:    tasksChan,
		crawlerCount: 0,
		writerCount:  0,
	}, nil
}

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	h := d.hosts[d.crawlerCount%len(d.hosts)]

	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(d.cfg.Protocols),
		timeout:   d.cfg.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Crawler{
		id:     fmt.Sprintf("crawler-%02d", d.crawlerCount),
		host:   h,
		pm:     pm,
		cfg:    d.cfg.CrawlerConfig(),
		client: kubo.NewClient(),
	}

	d.crawlerCount += 1

	return c, nil
}

func (d *CrawlDriver) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", d.writerCount), d.dbc, d.cfg.WriterConfig())
	d.writerCount += 1
	return w, nil
}

func (d *CrawlDriver) Tasks() <-chan PeerInfo {
	return d.tasksChan
}

func (d *CrawlDriver) Close() {}

func newLibp2pHost(userAgent string) (Host, error) {
	// Configure the resource manager to not limit anything
	// Don't use a connection manager that could potentially
	// prune any connections. We clean up after ourselves.
	cm := connmgr.NullConnMgr{}
	rm := network.NullResourceManager{}

	// Initialize a single libp2p node that's shared between all crawlers.
	h, err := libp2p.New(
		libp2p.UserAgent(userAgent),
		libp2p.ResourceManager(&rm),
		libp2p.ConnectionManager(cm),
		libp2p.DisableMetrics(),
		libp2p.SwarmOpts(swarm.WithReadOnlyBlackHoleDetector()),
		libp2p.UDPBlackHoleSuccessCounter(nil),
		libp2p.IPv6BlackHoleSuccessCounter(nil),
	)
	return h.(Host), err
}

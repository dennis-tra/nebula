package mainline

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"golang.org/x/time/rate"

	v2 "github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
)

type PeerInfo struct {
	krpc.NodeInfo
	peerID peer.ID
	maddrs []ma.Multiaddr // can only contain a single maddr
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func NewPeerInfo(info krpc.NodeInfo) (PeerInfo, error) {
	mh, err := multihash.Encode(info.ID[:], uint64(multicodec.Identity))
	if err != nil {
		return PeerInfo{}, fmt.Errorf("decode node id in multihash: %w", err)
	}

	peerID, err := peer.IDFromBytes(mh)
	if err != nil {
		return PeerInfo{}, fmt.Errorf("multihash to peer ID: %w", err)
	}

	maddr, err := manet.FromNetAddr(info.Addr.UDP())
	if err != nil {
		return PeerInfo{}, fmt.Errorf("multiaddress from net addr %s: %w", info.Addr.UDP(), err)
	}

	return PeerInfo{
		NodeInfo: info,
		peerID:   peerID,
		maddrs:   []ma.Multiaddr{maddr},
	}, nil
}

func (p PeerInfo) ID() peer.ID {
	return p.peerID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.maddrs
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	return p
}

type CrawlDriverConfig struct {
	Version        string
	DialTimeout    time.Duration
	TrackNeighbors bool
	BootstrapPeers []krpc.NodeInfo
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	LogErrors      bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout: cfg.DialTimeout,
		LogErrors:   cfg.LogErrors,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	dbc          db.Client
	servers      []*v2.Server
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	serverCfg := v2.NewDefaultServerConfig()
	serverCfg.Passive = true
	serverCfg.SendLimiter = rate.NewLimiter(rate.Inf, 0)
	serverCfg.StartingNodes = func() ([]v2.Addr, error) {
		return []v2.Addr{}, nil
	}

	// create a libp2p host per CPU core to distribute load
	servers := make([]*v2.Server, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		s, err := v2.NewServer(serverCfg)
		if err != nil {
			return nil, fmt.Errorf("new DHT server: %w", err)
		}

		servers = append(servers, s)
	}

	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, ni := range cfg.BootstrapPeers {
		pi, err := NewPeerInfo(ni)
		if err != nil {
			return nil, fmt.Errorf("new peer info from enr: %w", err)
		}

		tasksChan <- pi
	}
	close(tasksChan)

	return &CrawlDriver{
		cfg:       cfg,
		dbc:       dbc,
		servers:   servers,
		dbCrawl:   crawl,
		tasksChan: tasksChan,
		crawler:   make([]*Crawler, 0),
	}, nil
}

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	// evenly assign a libp2p hosts to crawler workers
	server := d.servers[d.crawlerCount%len(d.servers)]

	c := &Crawler{
		id:     fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:    d.cfg.CrawlerConfig(),
		server: server,
		done:   make(chan struct{}),
	}

	d.crawlerCount += 1

	d.crawler = append(d.crawler, c)

	log.Debugln("Started crawler worker", c.id)

	return c, nil
}

func (d *CrawlDriver) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", d.writerCount), d.dbc, d.dbCrawl.ID, d.cfg.WriterConfig())
	d.writerCount += 1
	return w, nil
}

func (d *CrawlDriver) Tasks() <-chan PeerInfo {
	return d.tasksChan
}

func (d *CrawlDriver) Close() {
	for _, server := range d.servers {
		server.Close()
	}
}

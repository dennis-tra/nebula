package bitcoin

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type PeerInfo struct {
	id       string
	maddrs   []ma.Multiaddr
	services wire.ServiceFlag
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func (p PeerInfo) ID() peer.ID {
	return peer.ID(p.id)
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.maddrs
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	if p.ID() != other.ID() {
		panic("merge peer ID mismatch")
	}

	return PeerInfo{
		id:       p.id,
		maddrs:   utils.MergeMaddrs(p.maddrs, other.maddrs),
		services: p.services & other.services,
	}
}

func (p PeerInfo) DeduplicationKey() string {
	return p.id
}

func (p PeerInfo) DiscoveryPrefix() uint64 {
	if len(p.id) < 8 {
		buf := make([]byte, 8)
		copy(buf[8-len(p.id):], p.id)
		return binary.BigEndian.Uint64(buf)
	}

	return binary.BigEndian.Uint64([]byte(p.id)[:8])
}

var serviceFlags = []wire.ServiceFlag{
	wire.SFNodeNetwork,
	wire.SFNodeGetUTXO,
	wire.SFNodeBloom,
	wire.SFNodeWitness,
	wire.SFNodeXthin,
	wire.SFNodeBit5,
	wire.SFNodeCF,
	wire.SFNode2X,
	wire.SFNodeNetworkLimited,
}

type CrawlDriverConfig struct {
	Version        string
	BitcoinNetwork wire.BitcoinNet
	DialTimeout    time.Duration
	BootstrapPeers []ma.Multiaddr
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
	LogErrors      bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:    cfg.DialTimeout,
		BitcoinNetwork: cfg.BitcoinNetwork,
		LogErrors:      cfg.LogErrors,
		Version:        cfg.Version,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	dbc          db.Client
	tasksChan    chan PeerInfo
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, maddr := range cfg.BootstrapPeers {
		_, addr, err := manet.DialArgs(maddr)
		if err != nil {
			log.WithError(err).Warnln("Invalid bootstrap peer", maddr)
			continue
		}

		tasksChan <- PeerInfo{
			id:       addr,
			maddrs:   []ma.Multiaddr{maddr},
			services: 0,
		}
	}
	close(tasksChan)

	return &CrawlDriver{
		cfg:       cfg,
		dbc:       dbc,
		tasksChan: tasksChan,
		crawler:   make([]*Crawler, 0),
	}, nil
}

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	c := &Crawler{
		id:   fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:  d.cfg.CrawlerConfig(),
		done: make(chan struct{}),
	}

	d.crawlerCount += 1

	d.crawler = append(d.crawler, c)

	log.Debugln("Started crawler worker", c.id)

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

func (d *CrawlDriver) Close() {
}

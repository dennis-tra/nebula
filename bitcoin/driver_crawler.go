package bitcoin

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type AddrInfo struct {
	id   string
	Addr []ma.Multiaddr
}
type PeerInfo struct {
	AddrInfo
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func (p PeerInfo) ID() peer.ID {
	return peer.ID(p.AddrInfo.id)

}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.AddrInfo.Addr
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	if p.AddrInfo.id != other.AddrInfo.id {
		panic("merge peer ID mismatch")
	}

	return PeerInfo{
		AddrInfo: AddrInfo{
			id:   p.AddrInfo.id,
			Addr: utils.MergeMaddrs(p.AddrInfo.Addr, other.AddrInfo.Addr),
		},
	}
}

func (p PeerInfo) DeduplicationKey() string {
	return p.AddrInfo.id
}

type CrawlDriverConfig struct {
	Version        string
	TrackNeighbors bool
	DialTimeout    time.Duration
	BootstrapPeers []ma.Multiaddr
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
	LogErrors      bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout: cfg.DialTimeout,
		LogErrors:   cfg.LogErrors,
		Version:     cfg.Version,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	dbc          db.Client
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {

	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, addrInfo := range cfg.BootstrapPeers {
		tasksChan <- PeerInfo{
			AddrInfo: AddrInfo{
				id:   addrInfo.String(),
				Addr: []ma.Multiaddr{addrInfo},
			},
		}
	}
	close(tasksChan)

	return &CrawlDriver{
		cfg:       cfg,
		dbc:       dbc,
		dbCrawl:   crawl,
		tasksChan: tasksChan,
		crawler:   make([]*Crawler, 0),
	}, nil
}

// NewWorker is called multiple times but only log the configured buffer sizes once
var logOnce sync.Once

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
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", d.writerCount), d.dbc, d.dbCrawl.ID, d.cfg.WriterConfig())
	d.writerCount += 1
	return w, nil
}

func (d *CrawlDriver) Tasks() <-chan PeerInfo {
	return d.tasksChan
}

func (d *CrawlDriver) Close() {
}

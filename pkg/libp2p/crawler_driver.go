package libp2p

import (
	"fmt"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/config"

	"github.com/libp2p/go-libp2p"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/dennis-tra/nebula-crawler/pkg/api"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

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

type CrawlDriverConfig struct {
	Version           string
	Protocols         []string
	DialTimeout       time.Duration
	TrackNeighbors    bool
	CheckExposed      bool
	BootstrapPeerStrs []string
	AddrTrackType     config.AddrType
	AddrDialType      config.AddrType
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		TrackNeighbors: cfg.TrackNeighbors,
		DialTimeout:    cfg.DialTimeout,
		CheckExposed:   cfg.CheckExposed,
		AddrDialType:   cfg.AddrDialType,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{
		AddrTrackType: cfg.AddrTrackType,
	}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	host         *basichost.BasicHost
	dbc          db.Client
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	crawlerCount int
	writerCount  int
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, dbCrawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Initialize a single libp2p node that's shared between all crawlers.
	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.ResourceManager(rm),
		libp2p.UserAgent("nebula/"+cfg.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	peerAddrs := map[peer.ID][]ma.Multiaddr{}
	for _, maddrStr := range cfg.BootstrapPeerStrs {

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

	tasksChan := make(chan PeerInfo, len(peerAddrs))
	for pid, addrs := range peerAddrs {
		pid := pid
		addrs := addrs
		tasksChan <- PeerInfo{
			AddrInfo: peer.AddrInfo{
				ID:    pid,
				Addrs: addrs,
			},
		}
	}
	close(tasksChan)

	return &CrawlDriver{
		cfg:          cfg,
		host:         h.(*basichost.BasicHost),
		dbc:          dbc,
		dbCrawl:      dbCrawl,
		tasksChan:    tasksChan,
		crawlerCount: 0,
		writerCount:  0,
	}, nil
}

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	ms := &msgSender{
		h:         d.host,
		protocols: protocol.ConvertFromStrings(d.cfg.Protocols),
		timeout:   d.cfg.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Crawler{
		id:     fmt.Sprintf("crawler-%02d", d.crawlerCount),
		host:   d.host,
		pm:     pm,
		cfg:    d.cfg.CrawlerConfig(),
		client: api.NewClient(),
	}

	d.crawlerCount += 1

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

func (d *CrawlDriver) Close() {}

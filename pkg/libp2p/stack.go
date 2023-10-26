package libp2p

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/models"

	"github.com/dennis-tra/nebula-crawler/pkg/api"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/libp2p/go-libp2p"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
)

type PeerInfo struct {
	peer.AddrInfo
}

var _ core.PeerInfo = (*PeerInfo)(nil)

func (p PeerInfo) ID() peer.ID {
	return p.AddrInfo.ID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.AddrInfo.Addrs
}

type CrawlStackConfig struct {
	Version           string
	Protocols         []string
	DialTimeout       time.Duration
	TrackNeighbors    bool
	CheckExposed      bool
	BootstrapPeerStrs []string
}

func (cfg *CrawlStackConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		TrackNeighbors: cfg.TrackNeighbors,
		DialTimeout:    cfg.DialTimeout,
		CheckExposed:   cfg.CheckExposed,
	}
}

type CrawlStack struct {
	cfg          *CrawlStackConfig
	host         *basichost.BasicHost
	dbc          db.Client
	crawl        *models.Crawl
	crawlerCount int
	writerCount  int
}

var _ core.Stack[PeerInfo] = (*CrawlStack)(nil)

func NewCrawlStack(dbc db.Client, crawl *models.Crawl, cfg *CrawlStackConfig) (*CrawlStack, error) {
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

	return &CrawlStack{
		cfg:          cfg,
		host:         h.(*basichost.BasicHost),
		dbc:          dbc,
		crawl:        crawl,
		crawlerCount: 0,
		writerCount:  0,
	}, nil
}

func (s *CrawlStack) NewCrawler() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	ms := &msgSender{
		h:         s.host,
		protocols: protocol.ConvertFromStrings(s.cfg.Protocols),
		timeout:   s.cfg.DialTimeout,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	c := &Crawler{
		id:     fmt.Sprintf("crawler-%02d", s.crawlerCount),
		host:   s.host,
		pm:     pm,
		cfg:    s.cfg.CrawlerConfig(),
		client: api.NewClient(),
	}

	s.crawlerCount += 1

	return c, nil
}

func (s *CrawlStack) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	id := fmt.Sprintf("writer-%02d", s.writerCount)
	w := core.NewCrawlWriter[PeerInfo](id, s.dbc, s.crawl.ID)
	s.writerCount += 1
	return w, nil
}

func (s *CrawlStack) BootstrapPeers() ([]PeerInfo, error) {
	peerAddrs := map[peer.ID][]ma.Multiaddr{}
	for _, maddrStr := range s.cfg.BootstrapPeerStrs {

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

	pis := make([]PeerInfo, 0, len(s.cfg.BootstrapPeerStrs))
	for pid, addrs := range peerAddrs {
		pid := pid
		addrs := addrs
		pi := PeerInfo{
			AddrInfo: peer.AddrInfo{
				ID:    pid,
				Addrs: addrs,
			},
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func (s *CrawlStack) OnClose() {}

type DialStackConfig struct {
	Version string
	// Protocols         []string
	// DialTimeout       time.Duration
	// TrackNeighbors    bool
	// CheckExposed      bool
	// BootstrapPeerStrs []string
}

type DialStack struct {
	cfg  *DialStackConfig
	host host.Host
	dbc  db.Client
}

func NewDialStack(dbc db.Client, cfg *DialStackConfig) (*DialStack, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Initialize a single libp2p node that's shared between all dialers.
	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.ResourceManager(rm),
		libp2p.UserAgent("nebula/"+cfg.Version),
	)
	if err != nil {
		return nil, err
	}

	return &DialStack{
		cfg:  cfg,
		host: h,
		dbc:  dbc,
	}, nil
}

package libp2p

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/api"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
)

type CrawlerConfig struct {
	TrackNeighbors bool
	DialTimeout    time.Duration
	CheckExposed   bool
	AddrDialType   config.AddrType
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	host         *basichost.BasicHost
	pm           *pb.ProtocolMessenger
	crawledPeers int
	client       *api.Client
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   task.ID().ShortString(),
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:      c.id,
		Info:           task,
		CrawlStartTime: time.Now(),
	}

	// adhere to the addr-dial-type command line flag and only work with
	// private/public addresses if the user asked for it.
	crawlInfo := task
	switch c.cfg.AddrDialType {
	case config.AddrTypePrivate:
		crawlInfo = PeerInfo{AddrInfo: utils.AddrInfoFilterPublicMaddrs(task.AddrInfo)}
	case config.AddrTypePublic:
		crawlInfo = PeerInfo{AddrInfo: utils.AddrInfoFilterPrivateMaddrs(task.AddrInfo)}
	default:
		// use any address
	}

	// start crawling both ways
	p2pResultCh := c.crawlP2P(ctx, crawlInfo)
	apiResultCh := c.crawlAPI(ctx, crawlInfo)

	p2pResult := <-p2pResultCh
	cr.CrawlEndTime = time.Now() // for legacy/consistency reasons we track the crawl end time here (without the API)
	apiResult := <-apiResultCh

	// merge both results
	mergeResults(&cr, p2pResult, apiResult)

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

func mergeResults(r *core.CrawlResult[PeerInfo], p2pRes P2PResult, apiRes APIResult) {
	if p2pRes.RoutingTable == nil {
		r.RoutingTable = &core.RoutingTable[PeerInfo]{PeerID: r.Info.ID()}
	} else {
		r.RoutingTable = p2pRes.RoutingTable
	}

	r.Agent = p2pRes.Agent
	r.Protocols = p2pRes.Protocols
	r.ConnectStartTime = p2pRes.ConnectStartTime
	r.ConnectEndTime = p2pRes.ConnectEndTime
	r.ConnectError = p2pRes.ConnectError
	r.ConnectErrorStr = p2pRes.ConnectErrorStr
	r.CrawlError = p2pRes.CrawlError
	r.CrawlErrorStr = p2pRes.CrawlErrorStr

	properties := map[string]any{}
	// If we attempted to crawl the API (only if we had at least one IP address for the peer)
	// and we received either the ID or routing table information
	if apiRes.Attempted {
		properties["is_exposed"] = apiRes.ID != nil || apiRes.RoutingTable != nil
	}

	var err error
	r.Properties, err = json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	if apiRes.ID != nil && (r.Agent == "" || len(r.Protocols) == 0) {
		r.Agent = apiRes.ID.AgentVersion
		r.Protocols = apiRes.ID.Protocols
	}

	var apiResMaddrs []ma.Multiaddr
	if apiRes.ID != nil {
		maddrs, err := utils.AddrsToMaddrs(apiRes.ID.Addresses)
		if err == nil {
			apiResMaddrs = maddrs
		}
	}

	if len(apiResMaddrs) > 0 {
		r.Info.AddrInfo.Addrs = apiResMaddrs
	} else if len(p2pRes.ListenAddrs) > 0 {
		r.Info.AddrInfo.Addrs = p2pRes.ListenAddrs
	}

	if len(r.RoutingTable.Neighbors) == 0 && apiRes.RoutingTable != nil {
		// construct routing table struct from API response
		rt := &core.RoutingTable[PeerInfo]{
			PeerID:    r.Info.ID(),
			Neighbors: []PeerInfo{},
		}

		for _, bucket := range apiRes.RoutingTable.Buckets {
			for _, p := range bucket.Peers {
				pid, err := peer.Decode(p.ID)
				if err != nil {
					continue
				}

				rt.Neighbors = append(rt.Neighbors, PeerInfo{
					AddrInfo: peer.AddrInfo{
						ID:    pid,
						Addrs: []ma.Multiaddr{},
					},
				})
			}
		}

		r.RoutingTable = rt
		r.RoutingTableFromAPI = true
	}
}

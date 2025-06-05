package libp2p

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
	"github.com/dennis-tra/nebula-crawler/kubo"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type CrawlerConfig struct {
	DialTimeout  time.Duration
	CheckExposed bool
	AddrDialType config.AddrType
	LogErrors    bool
	GossipSubPX  bool
	Clock        clock.Clock
}

func DefaultCrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:  15 * time.Second,
		CheckExposed: false,
		AddrDialType: config.AddrTypePublic,
		LogErrors:    false,
		GossipSubPX:  false,
		Clock:        clock.New(),
	}
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	host         Host
	pm           *pb.ProtocolMessenger
	psTopics     map[string]struct{}
	crawledPeers int
	client       *kubo.Client
	stateChan    chan string
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

	if c.cfg.GossipSubPX {
		c.stateChan <- "busy"
		defer func() { c.stateChan <- "idle" }()
	}

	// adhere to the addr-dial-type command line flag and only work with
	// private/public addresses if the user asked for it.
	var filterFn func(info peer.AddrInfo) (peer.AddrInfo, peer.AddrInfo)
	switch c.cfg.AddrDialType {
	case config.AddrTypePrivate:
		filterFn = utils.AddrInfoFilterPublicMaddrs
	case config.AddrTypePublic:
		filterFn = utils.AddrInfoFilterPrivateMaddrs
	default:
		// use all maddrs
		filterFn = func(info peer.AddrInfo) (peer.AddrInfo, peer.AddrInfo) {
			return info, peer.AddrInfo{ID: info.ID}
		}
	}
	dialAddrInfo, filteredAddrInfo := filterFn(task.AddrInfo)

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:      c.id,
		Info:           task,
		CrawlStartTime: time.Now(),
		LogErrors:      c.cfg.LogErrors,
		DialMaddrs:     dialAddrInfo.Addrs,
		FilteredMaddrs: filteredAddrInfo.Addrs,
	}

	// start crawling both ways
	crawlInfo := PeerInfo{AddrInfo: dialAddrInfo}
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
	r.DialErrors = db.MaddrErrors(r.DialMaddrs, p2pRes.ConnectError)

	// only track a crawl error if there was one AND we didn't get any neighbors.
	// as soon as we have a single neighbor we consider this as a successful crawl.
	// the details about which bucket requests have failed can be analysed with
	// the error bits.
	if p2pRes.CrawlError != nil && (p2pRes.RoutingTable == nil || len(p2pRes.RoutingTable.Neighbors) == 0) {
		r.CrawlError = p2pRes.CrawlError
		r.CrawlErrorStr = p2pRes.CrawlErrorStr
	}

	// keep track of the multi address that we used to connect to the peer
	r.ConnectMaddr = p2pRes.ConnectMaddr
	if r.ConnectMaddr == nil {
		r.ConnectMaddr = apiRes.ConnectMaddr
	}

	properties := map[string]any{}
	// If we attempted to crawl the API (only if we had at least one IP address for the peer)
	// and we received either the ID or routing table information
	if apiRes.Attempted {
		properties["is_exposed"] = apiRes.ID != nil || apiRes.RoutingTable != nil
	}

	// treat ErrConnectionClosedImmediately as no error because we were able
	// to connect
	if p2pRes.CrawlError != nil && strings.Contains(p2pRes.CrawlError.Error(), "connection failed") {
		properties["direct_close"] = true
	}

	// keep track of all unknown connection errors
	if p2pRes.ConnectErrorStr == pgmodels.NetErrorUnknown && p2pRes.ConnectError != nil {
		properties["connect_error"] = p2pRes.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if p2pRes.CrawlErrorStr == pgmodels.NetErrorUnknown && p2pRes.CrawlError != nil {
		properties["crawl_error"] = p2pRes.CrawlError.Error()
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

	// determine addresses we didn't know before we crawled the peer.
	// here, we build the ExtraMaddrs list which will contain multi addresses
	// that we didn't know when connecting to the peer but only discovered after
	// the peer told us about it in the identify exchange.
	knownMaddrs := map[string]struct{}{}
	for _, maddr := range r.Info.AddrInfo.Addrs {
		knownMaddrs[string(maddr.Bytes())] = struct{}{}
	}

	r.ExtraMaddrs = []ma.Multiaddr{}
	for _, maddr := range slices.Concat(apiRes.ListenMaddrs(), p2pRes.ListenMaddrs) {
		if _, ok := knownMaddrs[string(maddr.Bytes())]; ok {
			continue
		}
		r.ExtraMaddrs = append(r.ExtraMaddrs, maddr)
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

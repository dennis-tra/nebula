package crawl

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"
)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	// The crawler that generated this result
	CrawlerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	RoutingTable *RoutingTable

	// Indicates whether the above routing table information was queried through the API.
	// The API routing table does not include MultiAddresses, so we won't use them for further crawls.
	RoutingTableFromAPI bool

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred when connecting to the peer
	ConnectError error

	// The above error transferred to a known error
	ConnectErrorStr string

	// Any error that has occurred during fetching neighbor information
	CrawlError error

	// The above error transferred to a known error
	CrawlErrorStr string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time

	// Whether kubos RPC API is exposed
	IsExposed null.Bool
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r *Result) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer. This includes dialing and the identity protocol.
func (r *Result) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

func (r *Result) Merge(p2pRes P2PResult, apiRes APIResult) {
	if p2pRes.RoutingTable == nil {
		r.RoutingTable = &RoutingTable{PeerID: r.Peer.ID}
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

	// If we attempted to crawl the API (only if we had at least one IP address for the peer)
	// and we received either the ID or routing table information
	if apiRes.Attempted {
		r.IsExposed = null.BoolFrom(apiRes.ID != nil || apiRes.RoutingTable != nil)
	}

	if apiRes.ID != nil && (r.Agent == "" || len(r.Protocols) == 0) {
		r.Agent = apiRes.ID.AgentVersion
		r.Protocols = apiRes.ID.Protocols
	}

	if len(r.RoutingTable.Neighbors) == 0 && apiRes.RoutingTable != nil {
		// construct routing table struct from API response
		rt := &RoutingTable{
			PeerID:    r.Peer.ID,
			Neighbors: []peer.AddrInfo{},
		}

		for _, bucket := range apiRes.RoutingTable.Buckets {
			for _, p := range bucket.Peers {
				pid, err := peer.Decode(p.ID)
				if err != nil {
					continue
				}

				rt.Neighbors = append(rt.Neighbors, peer.AddrInfo{
					ID:    pid,
					Addrs: []ma.Multiaddr{},
				})
			}
		}

		r.RoutingTable = rt
		r.RoutingTableFromAPI = true
	}

}

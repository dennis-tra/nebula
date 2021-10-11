package crawl

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

// Result captures data that is gathered from crawling a single peer.
type Result struct {
	// The crawler that generated this result
	CrawlerID string

	// The crawled peer
	Peer peer.AddrInfo

	// The neighbors of the crawled peer
	Neighbors []peer.AddrInfo

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Any error that has occurred during the crawl
	Error error

	// The above error transferred to a known error
	DialError string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r *Result) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer.
func (r *Result) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

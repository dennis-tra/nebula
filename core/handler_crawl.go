package core

import (
	"context"
	"encoding/json"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
)

// CrawlResult captures data that is gathered from crawling a single peer.
type CrawlResult[I PeerInfo[I]] struct {
	// The crawler that generated this result
	CrawlerID string

	// Information about the crawled peer
	Info I

	// The extracted routing table of the crawled peer
	RoutingTable *RoutingTable[I]

	// The agent version of the crawled peer
	Agent string

	// The protocols the peer supports
	Protocols []string

	// Indicates whether the above routing table information was queried through the API.
	// The API routing table does not include MultiAddresses, so we won't use them for further crawls.
	RoutingTableFromAPI bool

	// The multi addresses we tried to dial. This can fewer addresses than we
	// found in the network because, e.g., the crawler won't try to connect to IP
	// addresses in the private CIDRs by default. It could also be that the peer
	// advertised multi addresses with protocols  that the crawler does not yet
	// support (unlikely though).
	DialMaddrs []ma.Multiaddr

	// A list of multi addresses that we found in the network for the
	// given peer but didn't try to dial. The union of filtered_maddrs and
	// dial_maddrs are all addresses we've found for the given peer in the
	// network. Nebula doesn't try to dial addresses from private IP addresses
	// by default (configurable though).
	FilteredMaddrs []ma.Multiaddr

	// A List of multi addresses that the peer additionally listens on.
	// After the crawler has connected to the peer, that peer will push all
	// addresses it listens on to the crawler. This list can contain additional
	// addresses that were not found in the network through the regular
	// discovery protocol.
	ExtraMaddrs []ma.Multiaddr

	// All addresses that the remote peer claims to listen on.
	ListenMaddrs []ma.Multiaddr

	// The multi address of the connection that we have established to the peer
	ConnectMaddr ma.Multiaddr

	// A list of errors that belong to each of the addresses stored in
	// DialMaddrs. The list is guaranteed to have the same length as the
	// DialMaddrs if a connection could not be established.
	DialErrors []string

	// An error that summaries the DialErrors into a single one (deprecated and
	// used in postgres). This error aggregation should be done later in the
	// analysis process.
	ConnectError error

	// The above error transformed to a known error
	ConnectErrorStr string

	// Any error that has occurred during fetching neighbor information
	CrawlError error

	// The above error transformed to a known error
	CrawlErrorStr string

	// When was the crawl started
	CrawlStartTime time.Time

	// When did this crawl end
	CrawlEndTime time.Time

	// When was the connection attempt made
	ConnectStartTime time.Time

	// As it can take some time to handle the result we track the timestamp explicitly
	ConnectEndTime time.Time

	// Additional properties of that specific peer we have crawled
	Properties json.RawMessage

	// Debug flag that indicates whether to log the full error string
	LogErrors bool
}

func (r CrawlResult[I]) PeerInfo() I {
	return r.Info
}

func (r CrawlResult[I]) LogEntry() *log.Entry {
	rtSize := -1
	if r.RoutingTable != nil {
		rtSize = len(r.RoutingTable.Neighbors)
	}
	logEntry := log.WithFields(log.Fields{
		"remoteID":   r.Info.ID().ShortString(),
		"isDialable": r.ConnectError == nil && r.CrawlError == nil,
		"duration":   r.CrawlDuration(),
		"rtSize":     rtSize,
	})

	if r.ConnectError != nil {
		if r.LogErrors || r.ConnectErrorStr == pgmodels.NetErrorUnknown {
			logEntry = logEntry.WithField("connErr", r.ConnectError)
		} else {
			logEntry = logEntry.WithField("connErr", r.ConnectErrorStr)
		}
	}

	if r.CrawlError != nil {
		if r.LogErrors || r.CrawlErrorStr == pgmodels.NetErrorUnknown {
			logEntry = logEntry.WithField("crawlErr", r.CrawlError)
		} else {
			logEntry = logEntry.WithField("crawlErr", r.CrawlErrorStr)
		}
	}

	return logEntry
}

func (r CrawlResult[I]) IsSuccess() bool {
	return r.ConnectError == nil && r.CrawlError == nil
}

// CrawlDuration returns the time it took to crawl to the peer (connecting + fetching neighbors)
func (r CrawlResult[I]) CrawlDuration() time.Duration {
	return r.CrawlEndTime.Sub(r.CrawlStartTime)
}

// ConnectDuration returns the time it took to connect to the peer. This includes dialing and the identity protocol.
func (r CrawlResult[I]) ConnectDuration() time.Duration {
	return r.ConnectEndTime.Sub(r.ConnectStartTime)
}

type CrawlHandlerConfig struct{}

// CrawlHandler is the default implementation for a [Handler] that can be used
// as the basis for crawl operations.
type CrawlHandler[I PeerInfo[I]] struct {
	cfg *CrawlHandlerConfig

	// A map of agent versions and their occurrences that happened during the crawl.
	AgentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	Protocols map[string]int

	// A map of errors that happened when trying to dial a peer.
	ConnErrs map[string]int

	// A map of errors that happened during the crawl.
	CrawlErrs map[string]int

	// The number of peers that were crawled.
	CrawledPeers int
}

func NewCrawlHandler[I PeerInfo[I]](cfg *CrawlHandlerConfig) *CrawlHandler[I] {
	return &CrawlHandler[I]{
		cfg:          cfg,
		AgentVersion: make(map[string]int),
		Protocols:    make(map[string]int),
		ConnErrs:     make(map[string]int),
		CrawlErrs:    make(map[string]int),
		CrawledPeers: 0,
	}
}

func (h *CrawlHandler[I]) HandlePeerResult(ctx context.Context, result Result[CrawlResult[I]]) []I {
	cr := result.Value

	// count the number of peers that we have crawled
	h.CrawledPeers += 1

	// Track agent versions
	h.AgentVersion[cr.Agent] += 1

	// Track seen protocols
	for _, p := range cr.Protocols {
		h.Protocols[p] += 1
	}

	if cr.ConnectError != nil {
		// Count connection errors
		h.ConnErrs[cr.ConnectErrorStr] += 1
	}

	if cr.CrawlError != nil {
		h.CrawlErrs[cr.CrawlErrorStr] += 1
	}

	// Schedule crawls of all found neighbors unless we got the routing table from the API.
	// In this case, the routing table information won't include any MultiAddresses. This means
	// we can't use these peers for further crawls.
	if !cr.RoutingTableFromAPI && cr.RoutingTable != nil {
		return cr.RoutingTable.Neighbors
	}

	return nil
}

func (h *CrawlHandler[I]) HandleWriteResult(ctx context.Context, result Result[WriteResult]) {
}

func (h *CrawlHandler[I]) Summary(state *EngineState) *Summary {
	return &Summary{
		PeersCrawled:    h.CrawledPeers,
		PeersDialable:   h.CrawledPeers - h.TotalErrors(),
		PeersUndialable: h.TotalErrors(),
		PeersRemaining:  state.PeersQueued,
		AgentVersion:    h.AgentVersion,
		Protocols:       h.Protocols,
		ConnErrs:        h.ConnErrs,
		CrawlErrs:       h.CrawlErrs,
	}
}

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (h *CrawlHandler[I]) TotalErrors() int {
	sum := 0
	for _, count := range h.ConnErrs {
		sum += count
	}
	return sum
}

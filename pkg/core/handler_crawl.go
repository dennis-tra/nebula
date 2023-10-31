package core

import (
	"math"

	"github.com/libp2p/go-libp2p/core/peer"
)

type CrawlHandlerConfig struct {
	// a flag that indicates whether we want to track and keep routing table
	// configurations of all peers in memory and write them to disk after the
	// crawl has finished.
	TrackNeighbors bool
}

// CrawlHandler is the default implementation for a [Handler] that can be used
// as the basis for crawl operations.
type CrawlHandler[I PeerInfo[I]] struct {
	cfg *CrawlHandlerConfig

	// A map that maps peer IDs to their database IDs. This speeds up the insertion of neighbor information as
	// the database does not need to look up every peer ID but only the ones not yet present in the database.
	// Speed up for ~11k peers: 5.5 min -> 30s
	PeerMappings map[peer.ID]int

	// A map that keeps track of all k-bucket entries of a particular peer.
	RoutingTables map[peer.ID]*RoutingTable[I]

	// A map of agent versions and their occurrences that happened during the crawl.
	AgentVersion map[string]int

	// A map of protocols and their occurrences that happened during the crawl.
	Protocols map[string]int

	// A map of errors that happened when trying to dial a peer.
	ConnErrs map[string]int

	// A map of errors that happened during the crawl.
	CrawlErrs map[string]int

	// The number of peers we would still need to crawl after the Run method has returned.
	QueuedPeers int

	// The number of peers that were crawled.
	CrawledPeers int
}

func NewCrawlHandler[I PeerInfo[I]](cfg *CrawlHandlerConfig) *CrawlHandler[I] {
	return &CrawlHandler[I]{
		cfg:           cfg,
		PeerMappings:  make(map[peer.ID]int),
		RoutingTables: make(map[peer.ID]*RoutingTable[I]),
		AgentVersion:  make(map[string]int),
		Protocols:     make(map[string]int),
		ConnErrs:      make(map[string]int),
		CrawlErrs:     make(map[string]int),
		QueuedPeers:   0,
		CrawledPeers:  0,
	}
}

func (h *CrawlHandler[I]) HandlePeerResult(result Result[CrawlResult[I]]) []I {
	cr := result.Value

	// count the number of peers that we have crawled
	h.CrawledPeers += 1

	// Track agent versions
	h.AgentVersion[cr.Agent] += 1

	// Track seen protocols
	for _, p := range cr.Protocols {
		h.Protocols[p] += 1
	}

	if cr.ConnectError == nil {
		// Only track the neighbors if we were actually able to connect to the peer. Otherwise, we would track
		// an empty routing table of that peer. Only track the routing table in the neighbors table if at least
		// one FIND_NODE RPC succeeded.
		if h.cfg.TrackNeighbors && cr.RoutingTable.ErrorBits < math.MaxUint16 {
			h.RoutingTables[cr.Info.ID()] = cr.RoutingTable
		}
	} else {
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

func (h *CrawlHandler[I]) HandleWriteResult(result Result[WriteResult]) {
	wr := result.Value

	if wr.InsertVisitResult != nil && wr.InsertVisitResult.PeerID != nil {
		h.PeerMappings[wr.PID] = *wr.InsertVisitResult.PeerID
	}
}

// TotalErrors counts the total amount of errors - equivalent to undialable peers during this crawl.
func (h *CrawlHandler[I]) TotalErrors() int {
	sum := 0
	for _, count := range h.CrawlErrs {
		sum += count
	}
	return sum
}

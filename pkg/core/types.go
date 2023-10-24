package core

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type PeerInfo interface {
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
}

type Stack[I PeerInfo] interface {
	NewCrawler() (Worker[I, CrawlResult[I]], error)
	NewWriter() (Worker[CrawlResult[I], WriteResult], error)
	BootstrapPeers() ([]I, error)
	OnPeerCrawled(result CrawlResult[I], err error)
	OnClose()
}

// RoutingTable captures the routing table information and crawl error of a particular peer
type RoutingTable[I PeerInfo] struct {
	// PeerID is the peer whose neighbors (routing table entries) are in the array below.
	PeerID peer.ID
	// The peers that are in the routing table of the above peer
	Neighbors []I
	// First error that has occurred during crawling that peer
	Error error
	// Little Endian representation of at which CPLs errors occurred during neighbors fetches.
	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	ErrorBits uint16
}

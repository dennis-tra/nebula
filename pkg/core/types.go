package core

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// PeerInfo is the interface that any peer information struct must conform to.
type PeerInfo interface {
	// ID should return the peer's/node's identifier mapped into a libp2p peer.ID.
	ID() peer.ID

	// Addrs should return all addresses that this peer is reachable at in multi address format.
	Addrs() []multiaddr.Multiaddr
}

// A Stack is a networking stack that must implements hook for events that the
// engine publishes (OnPeerCrawled, OnClose) as well as constructors for
// the internal crawlers and writers that the engine will operate on.
type Stack[I PeerInfo] interface {
	// NewCrawler returns a new [Worker] that takes a [PeerInfo], performs its
	// crawl duties by contacting that peer, and returns the resulting CrawlResult.
	NewCrawler() (Worker[I, CrawlResult[I]], error)

	// NewWriter returns a new [Worker] that takes a CrawlResult, performs its
	// duties by storing that result somewhere, and returns information about
	// how that all went.
	NewWriter() (Worker[CrawlResult[I], WriteResult], error)

	// BootstrapPeers returns a list of [PeerInfo]s from which the engine will
	// start the crawl. The source of the BootstrapPeers can be static or
	// dynamic (loaded from a database). Currently, now dynamic loading is
	// implemented. If the need arises, we might want to pass a context here.
	BootstrapPeers() ([]I, error)

	// OnPeerCrawled gives the stack a chance to perform some bookkeeping of
	// the crawl result. This method is called everytime after we have contacted
	// a peer.
	OnPeerCrawled(result CrawlResult[I], err error)

	// OnClose is called when the engine is about to shut down. This gives the
	// stack a chance to clean up internal resources.
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

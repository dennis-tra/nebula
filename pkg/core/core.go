package core

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

// PeerInfo is the interface that any peer information struct must conform to.
type PeerInfo[T any] interface {
	// ID should return the peer's/node's identifier mapped into a libp2p peer.ID.
	ID() peer.ID

	// Addrs should return all addresses that this peer is reachable at in multi address format.
	Addrs() []multiaddr.Multiaddr

	// Merge takes another peer info and merges it information into the callee
	// peer info struct. The implementation of Merge may panic if the peer IDs
	// don't match.
	Merge(other T) T
}

// A Driver is a data structure that provides the necessary implementations and
// tasks for the engine to operate.
type Driver[I PeerInfo[I], R WorkResult[I]] interface {
	// NewWorker returns a new [Worker] that takes a [PeerInfo], performs its
	// duties by contacting that peer, and returns the resulting WorkResult.
	NewWorker() (Worker[I, R], error)

	// NewWriter returns a new [Worker] that takes a [WorkResult], performs its
	// duties by storing that result somewhere, and returns information about
	// how that all went.
	NewWriter() (Worker[R, WriteResult], error)

	// Tasks returns a channel on which the driver should emit peer processing
	// tasks. This method will only be called once by the engine. The engine
	// will keep running until the returned channel was closed. Closing the
	// channel signals the engine that we don't anticipate to schedule any more
	// tasks. However, this doesn't mean that the engine will stop right away.
	// It will first process all remaining tasks it has in its queue. If you
	// want to preemptively stop the engine, cancel the context you passed into
	// [Engine.Run].
	Tasks() <-chan I

	// Close is called when the engine is about to shut down. This gives the
	// stack a chance to clean up internal resources. Implementation must be
	// idempotent as it may be called multiple times.
	Close()
}

// Handler defines the interface that the engine will call every time
// it has received a result from any of its workers.
type Handler[I PeerInfo[I], R WorkResult[I]] interface {
	// HandlePeerResult is called when the worker that has processed a peer
	// has emitted a new processing result. This can be a [CrawlResult] or
	// [DialResult] at the moment.
	HandlePeerResult(Result[R]) []I

	// HandleWriteResult is called when the writer has written a [CrawlResult]
	// or [DialResult] to disk.
	HandleWriteResult(Result[WriteResult])
}

// RoutingTable captures the routing table information and crawl error of a particular peer
type RoutingTable[I PeerInfo[I]] struct {
	// PeerID is the peer whose neighbors (routing table entries) are in the array below.
	PeerID peer.ID
	// The peers that are in the routing table of the above peer
	Neighbors []I
	// First error that has occurred during crawling that peer
	Error error
	// Little Endian representation of at which CPLs errors occurred during
	// neighbors fetches.
	// errorBits tracks at which CPL errors have occurred.
	// 0000 0000 0000 0000 - No error
	// 0000 0000 0000 0001 - An error has occurred at CPL 0
	// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
	ErrorBits uint16
}

// WorkResult must be implemented by the result that a Worker which processes
// peers returns.
type WorkResult[I PeerInfo[I]] interface {
	// PeerInfo returns information of the peer that was processed
	PeerInfo() I

	// LogEntry returns logging information that can be used by the engine
	LogEntry() *log.Entry
}

// WriteResult must be returned by write workers.
type WriteResult struct {
	*db.InsertVisitResult
	WriterID string
	PeerID   peer.ID
	Duration time.Duration
	Error    error
}

package monitor

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// Result captures data that is gathered from pinging a single peer.
type Result struct {
	DialerID string

	// The dialed peer
	Peer peer.AddrInfo

	// If error is set the peer was not dialable
	Error error

	// The above error transferred to a known error
	DialError string

	// When was the dial started
	DialStartTime time.Time

	// When did this crawl end
	DialEndTime time.Time
}

// DialDuration returns the time it took to dial the peer
func (r *Result) DialDuration() time.Duration {
	return r.DialEndTime.Sub(r.DialStartTime)
}

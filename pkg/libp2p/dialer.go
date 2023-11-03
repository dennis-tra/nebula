package libp2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/metrics"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

// Dialer encapsulates a libp2p host that dials peers.
type Dialer struct {
	id          string
	host        host.Host
	dialedPeers uint64
}

var _ core.Worker[PeerInfo, core.DialResult[PeerInfo]] = (*Dialer)(nil)

// Work TODO
func (d *Dialer) Work(ctx context.Context, task PeerInfo) (core.DialResult[PeerInfo], error) {
	// Creating log entry
	logEntry := log.WithFields(log.Fields{
		"dialerID":  d.id,
		"remoteID":  task.ID().ShortString(),
		"dialCount": d.dialedPeers,
	})
	logEntry.Debugln("Dialing peer")
	defer logEntry.Debugln("Dialed peer")

	// Initialize dial result
	dr := core.DialResult[PeerInfo]{
		DialerID:      d.id,
		Info:          task,
		DialStartTime: time.Now(),
	}

	pi := task.AddrInfo

	// Try to dial the peer 3 times
retryLoop:
	for retry := 0; retry < 3; retry++ {

		// Update log entry
		logEntry = logEntry.WithField("retry", retry)

		// Add peer information to peer store so that DialPeer can pick it up from there
		// Do this in every retry due to the TTL of one minute
		d.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, time.Minute)

		// Actually dial the peer
		if err := d.dial(ctx, pi.ID); err != nil {
			dr.Error = err
			dr.DialError = db.NetError(dr.Error)

			if errors.Is(err, context.Canceled) {
				break retryLoop
			}

			sleepDur := time.Duration(float64(retry+1) * float64(10*time.Second))
			errMsg := fmt.Sprintf("Dial failed, sleeping %s", sleepDur)

			switch dr.DialError {
			case models.NetErrorPeerIDMismatch:
				logEntry.WithError(err).Debugln("Dial failed due to peer ID mismatch - stopping retry")
				break retryLoop
			case models.NetErrorNoPublicIP, models.NetErrorNoGoodAddresses:
				logEntry.WithError(err).Debugln("Dial failed due to no public ip - stopping retry")
				break retryLoop
			case models.NetErrorMaxDialAttemptsExceeded:
				sleepDur = 70 * time.Second
				errMsg = fmt.Sprintf("Max dial attempts exceeded, sleeping longer %s", sleepDur)
			case models.NetErrorConnectionRefused:
				// The monitoring task receives a lot of "connection refused" messages. I guess there is
				// a limit somewhere of how often a peer can connect. I could imagine that this rate limiting
				// is set to one minute ¯\_(ツ)_/¯
				if retry >= 1 {
					logEntry.WithError(err).Debugf("Received 'connection refused' the second time - stopping retry")
					break retryLoop
				}
				sleepDur = 70 * time.Second
				errMsg = fmt.Sprintf("Connection refused, sleeping longer %s", sleepDur)
			default:
			}
			logEntry.WithError(err).Debugf(errMsg)
			select {
			case <-time.After(sleepDur):
			case <-ctx.Done():
				break retryLoop
			}
			continue retryLoop
		}

		// Dial was successful - reset error
		dr.Error = nil
		dr.DialError = ""

		break retryLoop
	}

	// Close established connection to prevent running out of FDs?
	if err := d.host.Network().ClosePeer(pi.ID); err != nil {
		logEntry.WithError(err).Warnln("Could not close connection to peer")
	}

	dr.DialEndTime = time.Now()

	d.dialedPeers += 1

	return dr, nil
}

func (d *Dialer) dial(ctx context.Context, peerID peer.ID) error {
	metrics.VisitCount.With(metrics.DialLabel).Inc()

	if _, err := d.host.Network().DialPeer(ctx, peerID); err != nil {
		metrics.VisitErrorsCount.With(metrics.DialLabel).Inc()
		return err
	}

	return nil
}

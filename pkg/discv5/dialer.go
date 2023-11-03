package discv5

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/eth"
)

// Dialer encapsulates a libp2p host that dials peers.
type Dialer struct {
	id          string
	dialedPeers uint64
	listener    *eth.UDPv5
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

	newEnr, err := d.listener.RequestENR(task.Node)
	dr.DialEndTime = time.Now()

	if err != nil {
		dr.Error = err
		dr.DialError = db.NetError(dr.Error)
	} else {
		task.Node = newEnr
		dr.Info = task
	}

	d.dialedPeers += 1

	return dr, nil
}

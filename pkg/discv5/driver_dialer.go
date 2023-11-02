package discv5

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
)

type DialDriverConfig struct {
	Version string
}

type DialDriver struct {
	cfg         *DialDriverConfig
	dbc         *db.DBClient
	taskQueue   chan PeerInfo
	start       chan struct{}
	shutdown    chan struct{}
	done        chan struct{}
	dialerCount int
	writerCount int
}

var _ core.Driver[PeerInfo, core.DialResult[PeerInfo]] = (*DialDriver)(nil)

func NewDialDriver(dbc *db.DBClient, cfg *DialDriverConfig) (*DialDriver, error) {
	d := &DialDriver{
		cfg:       cfg,
		dbc:       dbc,
		taskQueue: make(chan PeerInfo),
		start:     make(chan struct{}),
		shutdown:  make(chan struct{}),
		done:      make(chan struct{}),
	}

	go d.monitorDatabase()

	return d, nil
}

func (d *DialDriver) NewWorker() (core.Worker[PeerInfo, core.DialResult[PeerInfo]], error) {
	dialer := &Dialer{
		id: fmt.Sprintf("dialer-%02d", d.dialerCount),
	}

	d.dialerCount += 1

	return dialer, nil
}

func (d *DialDriver) NewWriter() (core.Worker[core.DialResult[PeerInfo], core.WriteResult], error) {
	id := fmt.Sprintf("writer-%02d", d.writerCount)
	w := core.NewDialWriter[PeerInfo](id, d.dbc)
	d.writerCount += 1
	return w, nil
}

func (d *DialDriver) Tasks() <-chan PeerInfo {
	close(d.start)
	return d.taskQueue
}

func (d *DialDriver) Close() {
	close(d.shutdown)
	<-d.done
	close(d.taskQueue)
}

// monitorDatabase checks every 10 seconds if there are peer sessions that are due to be renewed.
func (d *DialDriver) monitorDatabase() {
	defer close(d.done)

	select {
	case <-d.start:
	case <-d.shutdown:
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-d.shutdown
		cancel()
	}()

	for {
		log.Infof("Looking for sessions to check...")
		sessions, err := d.dbc.FetchDueOpenSessions(ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).Warnln("Could not fetch sessions")
			goto TICK
		}

		for _, session := range sessions {
			peerID, err := peer.Decode(session.R.Peer.MultiHash)
			if err != nil {
				log.WithField("mhash", session.R.Peer.MultiHash).
					WithError(err).
					Warnln("Could not parse multi address")
				continue
			}
			logEntry := log.WithField("peerID", peerID.ShortString())

			// Parse multi addresses from database
			addrInfo := peer.AddrInfo{ID: peerID}
			for _, maddrStr := range session.R.Peer.R.MultiAddresses {
				maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
				if err != nil {
					logEntry.WithError(err).Warnln("Could not parse multi address")
					continue
				}
				addrInfo.Addrs = append(addrInfo.Addrs, maddr)
			}

			select {
			case d.taskQueue <- PeerInfo{}: // TODO
				continue
			case <-ctx.Done():
				// fallthrough
			}
			break
		}
		// log.Infof("In dial queue %d peers", s.inDialQueueCount.Load())

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}

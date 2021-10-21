package provide

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
)

type Requester struct {
	h            host.Host
	dht          *kaddht.IpfsDHT
	pm           *pb.ProtocolMessenger
	ec           chan<- Event
	monitorCount atomic.Int32
}

func NewRequester(ctx context.Context, conf *config.Config, ec chan<- Event) (*Requester, error) {
	// Generate a new identity
	key, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair")
	}

	// Initialize a new libp2p host with the above identity
	var dht *kaddht.IpfsDHT
	h, err := libp2p.New(ctx,
		libp2p.Identity(key),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dht, err = kaddht.New(ctx, h)
			return dht, err
		}))
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	ms := &messageSenderImpl{
		host:      h,
		protocols: protocol.ConvertFromStrings(conf.Protocols),
		strmap:    make(map[peer.ID]*peerMessageSender),
		local:     h.ID(),
		ec:        ec,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	return &Requester{
		h:   h,
		dht: dht,
		pm:  pm,
		ec:  ec,
	}, nil
}

func (r *Requester) Bootstrap(ctx context.Context) error {
	logEntry := log.WithField("type", "requester")
	for _, bp := range kaddht.GetDefaultBootstrapPeerAddrInfos() {
		logEntry.WithField("remoteID", bp.ID.Pretty()[:16]).Infoln("Connecting to bootstrap peer")
		if err := r.h.Connect(ctx, bp); err != nil {
			return errors.Wrap(err, "connecting to bootstrap peer")
		}
	}
	return nil
}

func (r *Requester) RefreshRoutingTable() {
	<-r.dht.RefreshRoutingTable()
}

func (r *Requester) logEntry() *log.Entry {
	return log.WithField("type", "requester")
}

func (r *Requester) MonitorProviders(ctx context.Context, c *Content) ([]peer.ID, error) {
	r.logEntry().Infoln("Getting closest peers to monitor")
	peers, err := r.dht.GetClosestPeers(ctx, string(c.cid.Hash()))
	if err != nil {
		return nil, errors.Wrap(err, "get closest peers")
	}
	r.logEntry().Infof("Found %d peers", len(peers))

	go r.monitorPeers(ctx, c, peers)

	return peers, nil
}

func (r *Requester) monitorPeers(ctx context.Context, c *Content, peers []peer.ID) {
	r.logEntry().Infoln("Start monitoring closest peers")
	defer r.logEntry().Infoln("All monitoring routines stopped")

	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		r.monitorCount.Inc()
		go r.monitorPeer(ctx, c, p, &wg)
	}
	wg.Wait()
}

func (r *Requester) monitorPeer(ctx context.Context, c *Content, p peer.ID, wg *sync.WaitGroup) {
	defer wg.Done()
	defer r.monitorCount.Dec()

	logEntry := r.logEntry().WithField("remoteID", p.Pretty()[:16])
	logEntry.WithField("monitorCount", r.monitorCount.Load()).Infoln("Start monitoring peer")

	for {
		select {
		case <-time.Tick(time.Millisecond * 500):
		case <-ctx.Done():
			logEntry.WithField("monitorCount", r.monitorCount.Load()-1).Infoln("Stop monitoring peer")
			return
		}

		provs, _, err := r.pm.GetProviders(ctx, p, c.mhash)
		if err != nil {
			logEntry.WithField("monitorCount", r.monitorCount.Load()-1).WithError(err).Warnln("Failed to get providers")
			return
		} else if len(provs) > 0 {
			logEntry.WithField("monitorCount", r.monitorCount.Load()-1).Infoln("Found provider record")
			return
		}
	}
}

package provide

import (
	"context"
	"sync"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/config"

	"github.com/dennis-tra/nebula-crawler/pkg/service"
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
)

type Requester struct {
	*service.Service
	h   host.Host
	dht *kaddht.IpfsDHT
	pm  *pb.ProtocolMessenger
	ec  chan<- Event
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
		Service: service.New("requester"),
		h:       h,
		dht:     dht,
		pm:      pm,
		ec:      ec,
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

func (r *Requester) MonitorProviders(c *Content) error {
	r.logEntry().Infoln("Getting closest peers to monitor for provider records")
	peers, err := r.dht.GetClosestPeers(r.Ctx(), string(c.cid.Hash()))
	if err != nil {
		return errors.Wrap(err, "get closest peers")
	}
	r.logEntry().Infof("Found %d peers", len(peers))

	go r.monitorPeers(c, peers)

	return nil
}

func (r *Requester) monitorPeers(c *Content, peers []peer.ID) {
	r.Service.ServiceStarted()
	defer r.Service.ServiceStopped()

	r.logEntry().Infoln("Start monitoring closest peers")
	defer r.logEntry().Infoln("All monitoring routines stopped")

	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go r.monitorPeer(c, p, &wg)
	}
	wg.Wait()
}

func (r *Requester) monitorPeer(c *Content, p peer.ID, wg *sync.WaitGroup) {
	defer wg.Done()

	logEntry := r.logEntry().WithField("remoteID", p.Pretty()[:16])
	logEntry.Infoln("Start monitoring peer")

	for {
		select {
		case <-time.Tick(time.Millisecond * 500):
		case <-r.SigShutdown():
			logEntry.Infoln("Stop monitoring peer")
			return
		}

		provs, _, err := r.pm.GetProviders(r.Ctx(), p, c.mhash)
		if err != nil {
			logEntry.WithError(err).Warnln("Failed to get providers")
			return
		} else if len(provs) > 0 {
			logEntry.Infoln("Found provider record")
			return
		}
	}
}

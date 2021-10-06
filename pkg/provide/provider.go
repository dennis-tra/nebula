package provide

import (
	"context"

	"bou.ke/monkey"

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

	"github.com/dennis-tra/nebula-crawler/pkg/config"
)

type Provider struct {
	h   host.Host
	dht *kaddht.IpfsDHT
	ec  chan<- Event
}

func NewProvider(ctx context.Context, conf *config.Config, ec chan<- Event) (*Provider, error) {
	key, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair")
	}

	localID, err := peer.IDFromPublicKey(key.GetPublic())
	if err != nil {
		return nil, errors.Wrap(err, "id from public key")
	}

	ms := &messageSenderImpl{
		protocols: protocol.ConvertFromStrings(conf.Protocols),
		strmap:    make(map[peer.ID]*peerMessageSender),
		local:     localID,
		ec:        ec,
	}
	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	// When kaddht tries to instantiate a new protocol messenger hand it our implementation. There is no option to
	// exchange the protocol messenger or message sender implementation.
	monkey.Patch(pb.NewProtocolMessenger, func(msgSender pb.MessageSender, opts ...pb.ProtocolMessengerOption) (*pb.ProtocolMessenger, error) {
		return pm, nil
	})

	var dht *kaddht.IpfsDHT
	h, err := libp2p.New(ctx,
		libp2p.Identity(key),
		libp2p.DefaultListenAddrs,
		libp2p.Transport(NewTCPTransport(localID, ec)),
		libp2p.Transport(NewWSTransport(localID, ec)),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dht, err = kaddht.New(ctx, h)
			return dht, err
		}))
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	// Set the remaining information
	// TODO: race condition if the above init already tries to access these fields?
	ms.host = h

	// Remove monkey patched function.
	monkey.UnpatchAll()

	return &Provider{
		h:   h,
		dht: dht,
		ec:  ec,
	}, nil
}

func (p *Provider) logEntry() *log.Entry {
	return log.WithField("type", "provider")
}

func (p *Provider) Bootstrap(ctx context.Context) error {
	for _, bp := range kaddht.GetDefaultBootstrapPeerAddrInfos() {
		p.logEntry().WithField("remoteID", bp.ID.Pretty()[:16]).Infoln("Connecting to bootstrap peer")
		if err := p.h.Connect(ctx, bp); err != nil {
			return errors.Wrap(err, "connecting to bootstrap peer")
		}
	}
	return nil
}

func (p *Provider) RefreshRoutingTable() {
	p.logEntry().Infoln("Start refreshing routing table")
	defer p.logEntry().Infoln("Done refreshing routing table")
	<-p.dht.RefreshRoutingTable()
}

func (p *Provider) Provide(ctx context.Context, content *Content) error {
	p.logEntry().Infoln("Start providing content")
	defer p.logEntry().Infoln("Done providing content")
	return p.dht.Provide(ctx, content.cid, true)
}

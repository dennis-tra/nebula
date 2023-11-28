package discv5

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/eth"
)

type DialDriverConfig struct {
	Version string
}

type DialDriver struct {
	cfg         *DialDriverConfig
	dbc         *db.DBClient
	peerstore   *enode.DB
	taskQueue   chan PeerInfo
	start       chan struct{}
	shutdown    chan struct{}
	done        chan struct{}
	dialerCount int
	writerCount int
}

var _ core.Driver[PeerInfo, core.DialResult[PeerInfo]] = (*DialDriver)(nil)

func NewDialDriver(dbc *db.DBClient, cfg *DialDriverConfig) (*DialDriver, error) {
	peerstore, err := enode.OpenDB("") // in memory db
	if err != nil {
		return nil, fmt.Errorf("open in-memory peerstore: %w", err)
	}

	d := &DialDriver{
		cfg:       cfg,
		dbc:       dbc,
		peerstore: peerstore,
		taskQueue: make(chan PeerInfo),
		start:     make(chan struct{}),
		shutdown:  make(chan struct{}),
		done:      make(chan struct{}),
	}

	go d.monitorDatabase()

	return d, nil
}

func (d *DialDriver) NewWorker() (core.Worker[PeerInfo, core.DialResult[PeerInfo]], error) {
	// If I'm not using the below elliptic curve, some Ethereum clients will reject communication
	priv, err := ecdsa.GenerateKey(ethcrypto.S256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("new ethereum ecdsa key: %w", err)
	}

	ethNode := enode.NewLocalNode(d.peerstore, priv)

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, fmt.Errorf("listen on udp port: %w", err)
	}

	discv5Cfg := eth.Config{
		PrivateKey:   priv,
		ValidSchemes: enode.ValidSchemes,
	}

	listener, err := eth.ListenV5(conn, ethNode, discv5Cfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv5: %w", err)
	}

	dialer := &Dialer{
		id:       fmt.Sprintf("dialer-%02d", d.dialerCount),
		listener: listener,
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
			// take multi hash and decode into PeerID
			peerID, err := peer.Decode(session.R.Peer.MultiHash)
			if err != nil {
				log.WithField("mhash", session.R.Peer.MultiHash).
					WithError(err).
					Warnln("Could not parse multi address")
				continue
			}
			logEntry := log.WithField("peerID", peerID.ShortString())

			// init ENR and fill with secp256k1 public key from peerID
			var r enr.Record
			pubKey, err := peerID.ExtractPublicKey()
			if err != nil {
				logEntry.WithError(err).Warnln("Could not extract public key from peer ID")
				continue
			}

			raw, err := pubKey.Raw()
			if err != nil {
				logEntry.WithError(err).Warnln("Could not extract raw bytes from public key")
				continue
			}

			x, y := secp256k1.DecompressPubkey(raw)
			r.Set(enode.Secp256k1(ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}))

			// Parse multi addresses from database
			addrInfo := peer.AddrInfo{ID: peerID}
			for _, maddrStr := range session.R.Peer.R.MultiAddresses {
				maddr, err := ma.NewMultiaddr(maddrStr.Maddr)
				if err != nil {
					logEntry.WithError(err).Warnln("Could not parse multi address")
					continue
				}

				ip4 := false
				if comp, err := maddr.ValueForProtocol(ma.P_IP4); err == nil {
					ip4 = true
					if ip := net.ParseIP(comp); ip != nil {
						r.Set(enr.IPv4(ip))
					}
				} else if comp, err := maddr.ValueForProtocol(ma.P_IP6); err == nil {
					if ip := net.ParseIP(comp); ip != nil {
						r.Set(enr.IPv6(ip))
					}
				}

				if comp, err := maddr.ValueForProtocol(ma.P_UDP); err == nil {
					if udp, err := strconv.Atoi(comp); err == nil && udp <= math.MaxUint16 {
						if ip4 {
							r.Set(enr.UDP(uint16(udp)))
						} else {
							r.Set(enr.UDP6(uint16(udp)))
						}
					}
				} else if comp, err := maddr.ValueForProtocol(ma.P_TCP); err == nil {
					if tcp, err := strconv.Atoi(comp); err == nil && tcp <= math.MaxUint16 {
						if ip4 {
							r.Set(enr.TCP(uint16(tcp)))
						} else {
							r.Set(enr.TCP6(uint16(tcp)))
						}
					}
				}

				addrInfo.Addrs = append(addrInfo.Addrs, maddr)
			}

			// use custom identity scheme to not check the signature.
			node, err := enode.New(nebulaIdentityScheme{}, &r)
			if err != nil {
				logEntry.WithError(err).Warnln("Could not construct new enode.Node struct")
				continue
			}

			select {
			case d.taskQueue <- PeerInfo{
				Node:   node,
				peerID: addrInfo.ID,
				maddrs: addrInfo.Addrs,
			}:
				continue
			case <-ctx.Done():
				// fallthrough
			}
			break
		}

	TICK:
		select {
		case <-time.Tick(10 * time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}

// nebulaIdentityScheme is an always valid ID scheme. When a new [enode.Node] is
// constructed, the Verify method won't check the signature, and we just assume
// the record is valid. However, the NodeAddr method returns the correct node
// identifier.
type nebulaIdentityScheme struct{}

// Verify doesn't check the signature or anything. It assumes all records to be
// valid.
func (nebulaIdentityScheme) Verify(r *enr.Record, sig []byte) error {
	return nil
}

// NodeAddr returns the node's ID. The logic is copied from the [enode.V4ID]
// implementation.
func (nebulaIdentityScheme) NodeAddr(r *enr.Record) []byte {
	var pubkey enode.Secp256k1
	err := r.Load(&pubkey)
	if err != nil {
		return nil
	}
	buf := make([]byte, 64)
	math.ReadBits(pubkey.X, buf[:32])
	math.ReadBits(pubkey.Y, buf[32:])
	return crypto.Keccak256(buf)
}

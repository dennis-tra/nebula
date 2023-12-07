package devp2p

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"golang.org/x/net/context"
)

type Config struct {
	DialTimeout         time.Duration
	Caps                []p2p.Cap
	HighestProtoVersion uint
}

func DefaultConfig() *Config {
	return &Config{
		DialTimeout: time.Minute,
		Caps: []p2p.Cap{
			{Name: "eth", Version: 66},
			{Name: "eth", Version: 67},
			{Name: "eth", Version: 68},
		},
		HighestProtoVersion: 68,
	}
}

type Client struct {
	cfg     *Config
	dialer  net.Dialer
	privKey *ecdsa.PrivateKey

	connsMu sync.RWMutex
	conns   map[peer.ID]*Conn
}

func NewClient(privKey *ecdsa.PrivateKey, cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		cfg:     cfg,
		privKey: privKey,
		dialer: net.Dialer{
			Timeout: cfg.DialTimeout,
		},
		conns: map[peer.ID]*Conn{},
	}
}

func (c *Client) Connect(ctx context.Context, pi peer.AddrInfo) error {
	pubKey, err := pi.ID.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("extract public key: %w", err)
	}

	raw, err := pubKey.Raw()
	if err != nil {
		return fmt.Errorf("raw bytes from public key: %w", err)
	}

	x, y := secp256k1.DecompressPubkey(raw)
	ecdsaPubKey := ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}

	var fd net.Conn
	for _, maddr := range pi.Addrs {
		ipAddr, err := maddr.ValueForProtocol(ma.P_IP4)
		if err != nil {
			ipAddr, err = maddr.ValueForProtocol(ma.P_IP6)
			if err != nil {
				continue
			}
		}

		port, err := maddr.ValueForProtocol(ma.P_TCP)
		if err != nil {
			continue
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
		fd, err = c.dialer.DialContext(timeoutCtx, "tcp", fmt.Sprintf("%s:%s", ipAddr, port))
		if err != nil {
			cancel()
			return fmt.Errorf("failed dialing node: %w", err)
		}
		cancel()

		break
	}

	if fd == nil {
		return err
	}

	ethConn := &Conn{
		Conn:                       rlpx.NewConn(fd, &ecdsaPubKey),
		ourKey:                     c.privKey,
		negotiatedProtoVersion:     0,
		negotiatedSnapProtoVersion: 0,
		ourHighestProtoVersion:     c.cfg.HighestProtoVersion,
		ourHighestSnapProtoVersion: 0,
		caps:                       c.cfg.Caps,
	}

	// initiate authed session
	if err := fd.SetDeadline(time.Now().Add(10 * time.Second)); err != nil { // TODO: parameterize
		log.WithError(err).Warnln("Failed to set connection deadline")
	}

	_, err = ethConn.Handshake(c.privKey) // returns remote pubKey -> unused
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	c.connsMu.Lock()
	c.conns[pi.ID] = ethConn
	c.connsMu.Unlock()

	return nil
}

func (c *Client) Identify(pid peer.ID) (*Hello, error) {
	c.connsMu.RLock()
	conn, found := c.conns[pid]
	c.connsMu.RUnlock()

	if !found {
		return nil, fmt.Errorf("no connection to %s", pid)
	}

	pub0 := crypto.FromECDSAPub(&c.privKey.PublicKey)[1:]
	req := &Hello{
		Version: 5,
		Caps:    c.cfg.Caps,
		ID:      pub0,
	}

	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil { // TODO: parameterize
		log.WithError(err).Warnln("Failed to set connection deadline")
	}

	if err := conn.Write(req); err != nil {
		return nil, fmt.Errorf("write to conn: %w", err)
	}

	resp := conn.Read()

	switch respMsg := resp.(type) {
	case *Hello:
		if respMsg.Version >= 5 {
			conn.SetSnappy(true)
		}
		return respMsg, nil
	case *Error:
		return nil, fmt.Errorf("reading handshake response failed: %w", respMsg)
	case *Disconnect:
		return nil, fmt.Errorf("reading handshake response failed: %s", respMsg.Reason)
	default:
		return nil, fmt.Errorf("unexpected handshake response message type: %T", resp)
	}
}

func (c *Client) Close() {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	for pid, conn := range c.conns {
		delete(c.conns, pid)
		if err := conn.Close(); err != nil {
			log.WithError(err).WithField("remoteID", pid.ShortString()).Warnln("Failed closing devp2p connection")
		}
	}
}

func (c *Client) CloseConn(pid peer.ID) error {
	c.connsMu.Lock()
	defer c.connsMu.Unlock()

	conn, found := c.conns[pid]
	if !found {
		return nil
	}

	delete(c.conns, pid)

	return conn.Close()
}

package devp2p

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"time"

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

func (c *Client) Connect(ctx context.Context, pid peer.ID, maddr ma.Multiaddr) error {
	pubKey, err := pid.ExtractPublicKey()
	if err != nil {
		return fmt.Errorf("extract public key: %w", err)
	}

	raw, err := pubKey.Raw()
	if err != nil {
		return fmt.Errorf("raw bytes from public key: %w", err)
	}

	x, y := secp256k1.DecompressPubkey(raw)
	ecdsaPubKey := ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}

	ipAddr, err := maddr.ValueForProtocol(ma.P_IP4)
	if err != nil {
		ipAddr, err = maddr.ValueForProtocol(ma.P_IP6)
		if err != nil {
			return fmt.Errorf("no ip in multi address: %w", err)
		}
	}

	port, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return fmt.Errorf("no tcp address in multi address: %w", err)
	}

	fd, err := c.dialer.Dial("tcp", fmt.Sprintf("%s:%s", ipAddr, port))
	if err != nil {
		return fmt.Errorf("failed dialing node: %w", err)
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
	_, err = ethConn.Handshake(c.privKey) // returns remote pubKey -> unused
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	c.connsMu.Lock()
	c.conns[pid] = ethConn
	c.connsMu.Unlock()

	return nil
}

func (c *Client) Identify(pid peer.ID) error {
	c.connsMu.RLock()
	conn, found := c.conns[pid]
	c.connsMu.RUnlock()

	if !found {
		return fmt.Errorf("no connection to %s", pid)
	}

	pub0 := crypto.FromECDSAPub(&c.privKey.PublicKey)[1:]
	req := &Hello{
		Version: 5,
		Caps:    c.cfg.Caps,
		ID:      pub0,
	}

	if err := conn.Write(req); err != nil {
		return fmt.Errorf("write to conn: %w", err)
	}

	resp := conn.Read()

	helloResp, ok := resp.(*Hello)
	if !ok {
		return fmt.Errorf("unexpected handshake response message type: %T", resp)
	}

	if helloResp.Version >= 5 {
		conn.SetSnappy(true)
	}

	return nil
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

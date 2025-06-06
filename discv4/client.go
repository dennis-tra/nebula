package discv4

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/devp2p/ethtest"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var errUseOfClosedNetworkConnectionStr = "use of closed network connection"

type ClientConfig struct {
	DialTimeout         time.Duration
	Caps                []p2p.Cap
	HighestProtoVersion uint
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DialTimeout: 5 * time.Second,
		Caps: []p2p.Cap{
			// pretend to speak everything ¯\_(ツ)_/¯
			{Name: "eth", Version: 62},
			{Name: "eth", Version: 63},
			{Name: "eth", Version: 64},
			{Name: "eth", Version: 65},
			{Name: "eth", Version: 66},
			{Name: "eth", Version: 67},
			{Name: "eth", Version: 68},
			{Name: "eth", Version: 69},
			{Name: "eth", Version: 70},
			{Name: "eth", Version: 100},
			{Name: "snap", Version: 1},
		},
		HighestProtoVersion: 100,
	}
}

type Client struct {
	cfg     *ClientConfig
	dialer  net.Dialer
	privKey *ecdsa.PrivateKey
}

func NewClient(privKey *ecdsa.PrivateKey, cfg *ClientConfig) *Client {
	if cfg == nil {
		cfg = DefaultClientConfig()
	}

	return &Client{
		cfg:     cfg,
		privKey: privKey,
		dialer: net.Dialer{
			Timeout: cfg.DialTimeout,
		},
	}
}

func (c *Client) Connect(ctx context.Context, pi PeerInfo) (*ethtest.Conn, error) {
	logEntry := log.WithField("remoteID", pi.ID().ShortString())

	var conn net.Conn
	addrPort, ok := pi.Node.TCPEndpoint()
	if !ok {
		return nil, fmt.Errorf("no good ip address: %s:%d", pi.Node.IP(), pi.Node.TCP())
	}

	tctx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
	defer cancel()

	conn, err := c.dialer.DialContext(tctx, "tcp", addrPort.String())
	if err != nil {
		return nil, fmt.Errorf("failed dialing node: %w", err)
	}

	ethConn := &ethtest.Conn{
		Conn:                   rlpx.NewConn(conn, pi.Pubkey()),
		OurKey:                 c.privKey,
		Caps:                   c.cfg.Caps,
		OurHighestProtoVersion: c.cfg.HighestProtoVersion,
	}

	// cancel handshake if outer context is canceled and
	// cancel this go routine when this function exits
	exit := make(chan struct{})
	defer close(exit)
	go func() {
		select {
		case <-exit:
		case <-ctx.Done():
			if err := ethConn.Close(); err != nil && !strings.Contains(err.Error(), errUseOfClosedNetworkConnectionStr) {
				logEntry.WithError(err).Warnln("Failed closing devp2p connection")
			}
		}
	}()

	// set a deadline for the handshake
	if err := conn.SetDeadline(time.Now().Add(c.dialer.Timeout)); err != nil {
		logEntry.WithError(err).Warnln("Failed to set connection deadline")
	}

	_, err = ethConn.Conn.Handshake(c.privKey) // also returns the public key of the remote -> unused
	if err != nil {
		if ierr := ethConn.Close(); ierr != nil && !strings.Contains(ierr.Error(), errUseOfClosedNetworkConnectionStr) { // inner error
			logEntry.WithError(ierr).Warnln("Failed closing devp2p connection")
		}

		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	return ethConn, ctx.Err()
}

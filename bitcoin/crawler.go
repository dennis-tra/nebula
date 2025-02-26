package bitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
)

type CrawlerConfig struct {
	DialTimeout    time.Duration
	BitcoinNetwork wire.BitcoinNet
	LogErrors      bool
	Version        string
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	crawledPeers int
	done         chan struct{}
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   task.ID().ShortString(),
		"crawlCount": c.crawledPeers,
	})
	defer logEntry.Debugln("Crawled peer")

	crawlStart := time.Now()

	// start crawling
	result := c.crawlBitcoin(ctx, task)
	properties := c.PeerProperties(result)

	var (
		connectErr      error
		connectErrorStr string
		dialErrors      []string
	)
	if len(result.ConnectErrors) > 0 {
		// TODO: join all errors together
		connectErr = result.ConnectErrors[0]
		connectErrorStr = db.NetError(result.ConnectErrors[0])

		for _, err := range result.ConnectErrors {
			dialErrors = append(dialErrors, db.NetError(err))
		}

		// keep track of all unknown connection errors
		if connectErrorStr == pgmodels.NetErrorUnknown {
			properties["connect_error"] = result.ConnectErrors[0].Error()
		}
	}

	// keep track of all unknown crawl errors
	var crawlErrorStr string
	if result.CrawlError != nil {
		crawlErrorStr = db.NetError(result.CrawlError)
		properties["crawl_error"] = result.CrawlError.Error()
	}

	data, err := json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        result.RoutingTable,
		Agent:               result.Agent,
		Protocols:           result.Protocols,
		DialMaddrs:          task.Addrs(),
		ConnectMaddr:        result.ConnectMaddr,
		ExtraMaddrs:         []ma.Multiaddr{},
		FilteredMaddrs:      []ma.Multiaddr{},
		DialErrors:          dialErrors,
		ConnectError:        connectErr,
		ConnectErrorStr:     connectErrorStr,
		CrawlError:          result.CrawlError,
		CrawlErrorStr:       crawlErrorStr,
		CrawlEndTime:        time.Now(),
		ConnectStartTime:    result.ConnectStartTime,
		ConnectEndTime:      result.ConnectEndTime,
		Properties:          data,
		LogErrors:           c.cfg.LogErrors,
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

func (c *Crawler) PeerProperties(result BitcoinResult) map[string]any {
	properties := map[string]any{}

	properties["last_block"] = result.LastBlock

	return properties
}

type BitcoinResult struct {
	ConnectStartTime time.Time
	ConnectEndTime   time.Time
	ConnectErrors    []error
	CrawlError       error
	ConnectMaddr     ma.Multiaddr

	Agent     string
	Protocols []string
	LastBlock int32

	RoutingTable *core.RoutingTable[PeerInfo]
}

func (c *Crawler) crawlBitcoin(ctx context.Context, pi PeerInfo) BitcoinResult {
	result := BitcoinResult{
		RoutingTable: &core.RoutingTable[PeerInfo]{
			PeerID: pi.ID(),
		},
	}

	// connect to the peer
	var conn net.Conn
	result.ConnectStartTime = time.Now()
	conn, result.ConnectErrors = c.connect(ctx, pi.Addrs())
	result.ConnectEndTime = time.Now()

	if conn == nil {
		return result
	}
	// sendVerMsg.ProtocolVersion = int32(wire.ProtocolVersion)
	// sendVerMsg.Services = wire.SFNodeNetwork
	// sendVerMsg.Timestamp = time.Now()
	// sendVerMsg.UserAgent = "nebula/" + c.cfg.Version
	// sendVerMsg.DisableRelayTx = true

	cfg := &peer.Config{
		UserAgentName:    "nebula",
		UserAgentVersion: c.cfg.Version,
		ChainParams:      &chaincfg.MainNetParams,
		DisableRelayTx:   true,
		Listeners: peer.MessageListeners{
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				fmt.Println("version")
				return nil
			},
			OnVerAck: func(p *peer.Peer, msg *wire.MsgVerAck) {
				fmt.Println("verack")
			},
			OnAddr: func(peer *peer.Peer, msg *wire.MsgAddr) {
				fmt.Println("addr", len(msg.AddrList))
			},
			OnAddrV2: func(p *peer.Peer, msg *wire.MsgAddrV2) {
				fmt.Println("addrv2", len(msg.AddrList))
			},
		},
	}
	p, err := peer.NewOutboundPeer(cfg, conn.RemoteAddr().String())
	if err != nil {
		result.ConnectErrors = append(result.ConnectErrors, err)
		return result
	}
	p.AssociateConnection(conn)

	done := make(chan struct{})
	p.QueueMessage(wire.NewMsgGetAddr(), done)

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	go func() {
		select {
		case <-timeoutCtx.Done():
			p.Disconnect()
		}
	}()

	p.WaitForDisconnect()
	cancel()
	return result

	// wait for the close go routine below to exit
	closeDone := make(chan struct{})
	defer func() {
		select {
		case <-closeDone:
		case <-ctx.Done():
		}
	}()

	// close the connection at the end of crawling this bitcoin node by
	// canceling the closeCtx on exit of this function which then triggers the
	// call to Close() below.
	closeCtx, closeCancel := context.WithCancel(ctx)
	defer closeCancel()

	// start to asynchronously monitor the closeCtx and timeout channel. If
	// any of the two triggers, we close the connection. We use a timeout
	// channel here instead of a timeout context because we can reset it after
	// a successful handshake.
	// The closeCtx is closed when this function exits. Then at the end, after
	// we successfully closed the connection, we close the closeDone channel
	// which blocks the exit of the method in the above `defer` statement.
	timeout := time.NewTimer(c.cfg.DialTimeout)
	go func() {
		defer close(closeDone)
		select {
		case <-timeout.C:
		case <-closeCtx.Done():
		}
		if err := conn.Close(); err != nil {
			log.WithError(err).WithField("remoteID", pi.ID().ShortString()).Warnln("Could not close connection to peer")
		}
	}()

	// We successfully connected to the peer, keep track of the connect-multiaddr
	maddr, err := manet.FromNetAddr(conn.RemoteAddr())
	if err != nil {
		log.WithError(err).WithField("addr", conn.RemoteAddr()).Warnln("Could not construct connect maddr")
	} else {
		result.ConnectMaddr = maddr
	}

	// perform the handshake before we can drain the peers from the node
	versionInfo, err := c.handshake(conn)
	if err != nil {
		// TODO: loop through dial multiaddresses, find the one that matches conn and
		//       assign the error to it.
		result.ConnectErrors = []error{err}
	} else {

		for _, flag := range serviceFlags {
			and := versionInfo.Services & flag
			if and != 0 {
				result.Protocols = append(result.Protocols, fmt.Sprintf("/%d/%s", versionInfo.ProtocolVersion, and.String()))
			}
		}

		result.Agent = versionInfo.UserAgent
		result.LastBlock = versionInfo.LastBlock

		// the handshake was successful, give 10s to drain the peers
		timeout.Reset(10 * time.Second)

		neighbors, err := c.drainPeer(conn, uint32(versionInfo.ProtocolVersion))
		if err != nil && len(neighbors) == 0 {
			result.CrawlError = err
		}

		result.RoutingTable = &core.RoutingTable[PeerInfo]{
			PeerID:    pi.ID(),
			Neighbors: make([]PeerInfo, 0, len(neighbors)),
			ErrorBits: uint16(0),
			Error:     err,
		}

		for _, n := range neighbors {
			result.RoutingTable.Neighbors = append(result.RoutingTable.Neighbors, n)
		}
	}

	return result
}

// connect establishes a connection to the given peer
func (c *Crawler) connect(ctx context.Context, maddrs []ma.Multiaddr) (net.Conn, []error) {
	var errs []error
	for _, maddr := range maddrs {
		network, address, err := manet.DialArgs(maddr)
		if err != nil {
			log.WithError(err).WithField("maddr", maddr).Warnln("Could not parse maddr")
			continue
		}

		dialer := net.Dialer{Timeout: c.cfg.DialTimeout}
		conn, err := dialer.DialContext(ctx, network, address)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return conn, nil
	}

	return nil, errs
}

func (c *Crawler) handshake(conn net.Conn) (*wire.MsgVersion, error) {
	nonce, err := wire.RandomUint64()
	if err != nil {
		return nil, fmt.Errorf("could not generate nonce: %v", err)
	}

	localTCPAddr := conn.LocalAddr().(*net.TCPAddr)
	remoteTCPAddr := conn.RemoteAddr().(*net.TCPAddr)

	localAddr := wire.NewNetAddressIPPort(
		localTCPAddr.IP,
		uint16(localTCPAddr.Port),
		wire.SFNodeNetwork,
	)
	remoteAddr := wire.NewNetAddressIPPort(
		remoteTCPAddr.IP,
		uint16(remoteTCPAddr.Port),
		wire.SFNodeNetwork,
	)

	sendVerMsg := wire.NewMsgVersion(localAddr, remoteAddr, nonce, 0)
	sendVerMsg.ProtocolVersion = int32(wire.ProtocolVersion)
	sendVerMsg.Services = wire.SFNodeNetwork
	sendVerMsg.Timestamp = time.Now()
	sendVerMsg.UserAgent = "nebula/" + c.cfg.Version
	sendVerMsg.DisableRelayTx = true

	if err := c.WriteMessage(conn, wire.ProtocolVersion, sendVerMsg); err != nil {
		return nil, fmt.Errorf("could not write version message: %w", err)
	}

	// Read the response version.
	msg, _, err := c.ReadMessage(conn, wire.ProtocolVersion)
	if err != nil {
		return nil, fmt.Errorf("could not read version message: %w", err)
	}

	recVerMsg, ok := msg.(*wire.MsgVersion)
	if !ok {
		return nil, fmt.Errorf("did not receive version message: %T", recVerMsg)
	}

	// Send version acknowledgement
	if err := c.WriteMessage(conn, uint32(recVerMsg.ProtocolVersion), wire.NewMsgVerAck()); err != nil {
		return recVerMsg, nil // don't consider this as an error because we got what we want
	}

	return recVerMsg, nil
}

func (c *Crawler) drainPeer(conn net.Conn, protocolVersion uint32) (map[string]PeerInfo, error) {
	if err := c.WriteMessage(conn, protocolVersion, wire.NewMsgGetAddr()); err != nil {
		return nil, fmt.Errorf("send getaddr: %w", err)
	}

	neighbors := map[string]PeerInfo{}
	for {
		msg, _, err := c.ReadMessage(conn, protocolVersion)
		if err != nil {
			return neighbors, fmt.Errorf("read message: %w", err)
		}

		switch tmsg := msg.(type) {
		case *wire.MsgAddr:
			foundNewAddr := false
			for _, addr := range tmsg.AddrList {
				fmt.Println(conn.RemoteAddr(), addr.IP, addr.Port)

				netAddr, ok := netip.AddrFromSlice(addr.IP)
				if !ok {
					continue
				}

				addrPort := netip.AddrPortFrom(netAddr, addr.Port)

				id := addrPort.String()
				if _, ok := neighbors[id]; ok {
					continue
				}

				foundNewAddr = true

				ipMaddr, err := manet.FromIP(addr.IP)
				if err != nil {
					log.WithError(err).WithField("ip", addr.IP).Warnln("Could not construct maddr")
					continue
				}

				trptComp, err := ma.NewComponent("tcp", strconv.Itoa(int(addr.Port)))
				if err != nil {
					log.WithError(err).WithField("port", addr.Port).Warnln("Could not construct transport component")
					continue
				}

				neighbors[id] = PeerInfo{
					id:        id,
					maddrs:    []ma.Multiaddr{ipMaddr.Encapsulate(trptComp)},
					timestamp: addr.Timestamp,
					services:  addr.Services,
				}
			}

			if !foundNewAddr {
				return neighbors, nil
			}

			if err = c.WriteMessage(conn, protocolVersion, wire.NewMsgGetAddr()); err != nil {
				return neighbors, fmt.Errorf("send getaddr: %w", err)
			}
		case *wire.MsgAddrV2:

			foundNewAddr := false
			for _, addr := range tmsg.AddrList {
				id := addr.Addr.String()
				if _, ok := neighbors[id]; ok {
					continue
				}

				maddr, err := manet.FromNetAddr(addr.Addr)
				if err != nil {
					log.WithError(err).WithField("addr", addr.Addr).Warnln("Could not construct maddr")
					continue
				}

				trptComp, err := ma.NewComponent(addr.Addr.Network(), strconv.Itoa(int(addr.Port)))
				if err != nil {
					log.WithError(err).WithField("net", addr.Addr.Network()).WithField("port", addr.Port).Warnln("Could not construct transport component")
					continue
				}

				foundNewAddr = true
				neighbors[id] = PeerInfo{
					id:        id,
					maddrs:    []ma.Multiaddr{maddr.Encapsulate(trptComp)},
					timestamp: addr.Timestamp,
					services:  addr.Services,
				}

			}

			if !foundNewAddr {
				return neighbors, nil
			}

			if err = c.WriteMessage(conn, protocolVersion, wire.NewMsgGetAddr()); err != nil {
				return neighbors, fmt.Errorf("send getaddr: %w", err)
			}
		case *wire.MsgPing:
			err = c.WriteMessage(conn, protocolVersion, wire.NewMsgPong(tmsg.Nonce))
			if err != nil {
				return neighbors, fmt.Errorf("send pong: %w", err)
			}
		case *wire.MsgVerAck:
			// nothing
		case *wire.MsgInv:
			// nothing
		default:
			log.WithField("type", fmt.Sprintf("%T", msg)).Warnln("Unknown message type")
			// nothing to do
		}
	}
}

func (c *Crawler) WriteMessage(conn net.Conn, protocolVersion uint32, msg wire.Message) error {
	return wire.WriteMessage(conn, msg, protocolVersion, c.cfg.BitcoinNetwork)
}

func (c *Crawler) ReadMessage(conn net.Conn, protocolVersion uint32) (wire.Message, []byte, error) {
	return wire.ReadMessage(conn, protocolVersion, c.cfg.BitcoinNetwork)
}

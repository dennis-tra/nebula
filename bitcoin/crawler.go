package bitcoin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

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
	closeCtx, closeCancel := context.WithTimeout(ctx, 30*time.Second)
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

	versionMsg, pver, err := c.handshake(conn)
	if err != nil {
		result.ConnectErrors = append(result.ConnectErrors, err)
		return result
	}

	for _, flag := range serviceFlags {
		and := versionMsg.Services & flag
		if and != 0 {
			result.Protocols = append(result.Protocols, fmt.Sprintf("/%d/%s", versionMsg.ProtocolVersion, and.String()))
		}
	}

	result.Agent = versionMsg.UserAgent
	result.LastBlock = versionMsg.LastBlock

	// supportsSendAddrV2 := false
	verAckReceived := false
	foundNew := false
	neighbors := map[string]PeerInfo{}
loop:
	for {
		msg, _, err := c.ReadMessage(conn, pver)
		if errors.Is(err, wire.ErrUnknownMessage) {
			continue
		} else if err != nil {
			result.CrawlError = err
			break loop
		}

		timeout.Reset(c.cfg.DialTimeout)

		switch tmsg := msg.(type) {
		case *wire.MsgVersion:
			// already received at this point. Usually we would reject this
			// message as a duplicate, but we are generous here.
			continue

		case *wire.MsgVerAck:
			if verAckReceived {
				// already received at this point. Usually we would reject this
				// message as a duplicate, but we are generous here.
				continue
			}

			// handshake complete
			verAckReceived = true

			if _, err := wire.WriteMessageWithEncodingN(conn, wire.NewMsgGetAddr(), pver, c.cfg.BitcoinNetwork, wire.BaseEncoding); err != nil {
				result.CrawlError = err
				break loop
			}

		case *wire.MsgSendAddrV2:
			if pver < wire.AddrV2Version {
				continue
			}
			// supportsSendAddrV2 = true

		case *wire.MsgGetAddr:
		case *wire.MsgAddr:
			neighbors, foundNew = c.handleMsgAddr(tmsg.AddrList, neighbors)
			if !foundNew {
				break loop
			}

			if _, err := wire.WriteMessageWithEncodingN(conn, wire.NewMsgGetAddr(), pver, c.cfg.BitcoinNetwork, wire.BaseEncoding); err != nil {
				result.CrawlError = err
				break loop
			}

		case *wire.MsgAddrV2:
			neighbors, foundNew = c.handleMsgAddrV2(tmsg.AddrList, neighbors)
			if !foundNew {
				break loop
			}

			if _, err := wire.WriteMessageWithEncodingN(conn, wire.NewMsgGetAddr(), pver, c.cfg.BitcoinNetwork, wire.BaseEncoding); err != nil {
				result.CrawlError = err
				break loop
			}

		case *wire.MsgPing:
			if pver <= wire.BIP0031Version {
				continue
			}

			if _, err := wire.WriteMessageWithEncodingN(conn, wire.NewMsgPong(tmsg.Nonce), pver, c.cfg.BitcoinNetwork, wire.BaseEncoding); err != nil {
				result.CrawlError = err
				break loop
			}
		case *wire.MsgPong:
		case *wire.MsgAlert:
		case *wire.MsgMemPool:
		case *wire.MsgTx:
		case *wire.MsgBlock:
		case *wire.MsgInv:
		case *wire.MsgHeaders:
		case *wire.MsgNotFound:
		case *wire.MsgGetData:
		case *wire.MsgGetBlocks:
		case *wire.MsgGetHeaders:
		case *wire.MsgGetCFilters:
		case *wire.MsgGetCFHeaders:
		case *wire.MsgGetCFCheckpt:
		case *wire.MsgCFilter:
		case *wire.MsgCFHeaders:
		case *wire.MsgFeeFilter:
		case *wire.MsgFilterAdd:
		case *wire.MsgFilterClear:
		case *wire.MsgFilterLoad:
		case *wire.MsgMerkleBlock:
		case *wire.MsgReject:
		case *wire.MsgSendHeaders:
		default:
			log.WithField("type", fmt.Sprintf("%T", tmsg)).Warnln("Unknown message type")
		}
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

	if len(neighbors) > 0 {
		result.CrawlError = nil
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
		return conn, errs
	}

	return nil, errs
}

func (c *Crawler) localVersionMsg(conn net.Conn) *wire.MsgVersion {
	remoteTCPAddr := conn.RemoteAddr().(*net.TCPAddr)

	// Create a wire.NetAddress with only the services set to use as the
	// "addrme" in the version message.
	//
	// Older nodes previously added the IP and port information to the
	// address manager which proved to be unreliable as an inbound
	// connection from a peer didn't necessarily mean the peer itself
	// accepted inbound connections.
	//
	// Also, the timestamp is unused in the version message.
	localAddr := &wire.NetAddress{
		Services: wire.SFNodeNetwork, // TODO: which service to set?
	}

	remoteAddr := &wire.NetAddress{
		Timestamp: time.Now(),
		Services:  0,
		IP:        remoteTCPAddr.IP,
		Port:      uint16(remoteTCPAddr.Port),
	}

	msg := wire.NewMsgVersion(localAddr, remoteAddr, uint64(rand.Int63()), 0)
	msg.DisableRelayTx = true
	msg.Services = localAddr.Services
	msg.Timestamp = time.Now()
	if err := msg.AddUserAgent("nebula", c.cfg.Version, "network crawler"); err != nil {
		panic(err)
	}

	return msg
}

func (c *Crawler) handshake(conn net.Conn) (*wire.MsgVersion, uint32, error) {
	sendVerMsg := c.localVersionMsg(conn)

	if err := c.WriteMessage(conn, wire.ProtocolVersion, sendVerMsg); err != nil {
		return nil, 0, fmt.Errorf("could not write version message: %w", err)
	}

	// Read the response version.
	msg, _, err := c.ReadMessage(conn, wire.ProtocolVersion)
	if err != nil {
		return nil, 0, fmt.Errorf("could not read version message: %w", err)
	}

	recVerMsg, ok := msg.(*wire.MsgVersion)
	if !ok {
		return nil, 0, fmt.Errorf("did not receive version message: %T", recVerMsg)
	}

	pver := wire.ProtocolVersion
	if uint32(recVerMsg.ProtocolVersion) < wire.ProtocolVersion {
		pver = uint32(recVerMsg.ProtocolVersion)
	}

	if pver >= wire.AddrV2Version {
		if err := c.WriteMessage(conn, pver, wire.NewMsgSendAddrV2()); err != nil {
			return nil, 0, fmt.Errorf("could not send sendaddrv2 message: %w", err)
		}
	}

	// Send version acknowledgement
	if err := c.WriteMessage(conn, pver, wire.NewMsgVerAck()); err != nil {
		return nil, 0, fmt.Errorf("could not send verack message: %w", err)
	}

	return recVerMsg, pver, nil
}

func (c *Crawler) WriteMessage(conn net.Conn, protocolVersion uint32, msg wire.Message) error {
	_, err := wire.WriteMessageWithEncodingN(conn, msg, protocolVersion, c.cfg.BitcoinNetwork, wire.LatestEncoding)
	return err
}

func (c *Crawler) ReadMessage(conn net.Conn, protocolVersion uint32) (wire.Message, []byte, error) {
	_, msg, bytes, err := wire.ReadMessageWithEncodingN(conn, protocolVersion, c.cfg.BitcoinNetwork, wire.LatestEncoding)
	return msg, bytes, err
}

func (c *Crawler) handleMsgAddr(netAddrs []*wire.NetAddress, neighbors map[string]PeerInfo) (map[string]PeerInfo, bool) {
	foundNewAddr := false
	for _, addr := range netAddrs {
		id := net.JoinHostPort(addr.IP.String(), strconv.Itoa(int(addr.Port)))
		if _, ok := neighbors[id]; ok {
			continue
		}

		ipComp, err := manet.FromIP(addr.IP)
		if err != nil {
			log.WithError(err).WithField("ip", addr.IP).Warnln("Could not construct maddr")
			continue
		}

		trptComp, err := ma.NewComponent("tcp", strconv.Itoa(int(addr.Port)))
		if err != nil {
			log.WithError(err).WithField("port", addr.Port).Warnln("Could not construct transport component")
			continue
		}

		foundNewAddr = true
		neighbors[id] = PeerInfo{
			id:       id,
			maddrs:   []ma.Multiaddr{ipComp.Encapsulate(trptComp)},
			services: addr.Services,
		}
	}

	return neighbors, foundNewAddr
}

func (c *Crawler) handleMsgAddrV2(netAddrs []*wire.NetAddressV2, neighbors map[string]PeerInfo) (map[string]PeerInfo, bool) {
	foundNewAddr := false
	for _, addr := range netAddrs {
		host := addr.Addr.String()
		portStr := strconv.Itoa(int(addr.Port))
		id := net.JoinHostPort(host, portStr)
		if _, ok := neighbors[id]; ok {
			continue
		}

		var maddr ma.Multiaddr

		if len(host) == wire.TorV2EncodedSize && host[wire.TorV2EncodedSize-6:] == ".onion" {
			onion := strings.TrimSuffix(host, ".onion")
			ipComp, err := ma.NewComponent("onion", net.JoinHostPort(onion, portStr))
			if err != nil {
				log.WithField("addr", host).WithError(err).Warnln("Could not construct onion component")
				continue
			}
			maddr = ipComp
		} else if len(host) == wire.TorV3EncodedSize && host[wire.TorV3EncodedSize-6:] == ".onion" {
			onion := strings.TrimSuffix(host, ".onion")
			ipComp, err := ma.NewComponent("onion3", net.JoinHostPort(onion, portStr))
			if err != nil {
				log.WithField("addr", host).WithError(err).Warnln("Could not construct onion3 component")
				continue
			}
			maddr = ipComp
		} else if ip := net.ParseIP(host); ip != nil {
			ipComp, err := manet.FromIP(ip)
			if err != nil {
				log.WithError(err).WithField("ip", ip).Warnln("Could not construct maddr")
				continue
			}

			trptComp, err := ma.NewComponent("tcp", portStr)
			if err != nil {
				log.WithError(err).WithField("port", addr.Port).Warnln("Could not construct transport component")
				continue
			}
			maddr = ipComp.Encapsulate(trptComp)
		} else {
			log.WithField("addr", host).Warnln("Could not construct ip component")
			continue
		}

		foundNewAddr = true
		neighbors[host] = PeerInfo{
			id:       host,
			maddrs:   []ma.Multiaddr{maddr},
			services: addr.Services,
		}
	}

	return neighbors, foundNewAddr
}

package bitcoin

import (
	"context"
	"encoding/json"
	"net"

	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/cenkalti/backoff/v4"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
)

const MaxCrawlRetriesAfterTimeout = 2 // magic

type CrawlerConfig struct {
	DialTimeout  time.Duration
	AddrDialType config.AddrType
	KeepENR      bool
	LogErrors    bool
	Version      string
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	conn         net.Conn
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
	bitcoinResult := <-c.crawlBitcoin(ctx, task)

	properties := c.PeerProperties(&task.AddrInfo)

	if bitcoinResult.Transport != "" {
		properties["transport"] = bitcoinResult.Transport
	}

	if bitcoinResult.ConnClosedImmediately {
		properties["direct_close"] = true
	}

	if bitcoinResult.GenTCPAddr {
		properties["gen_tcp_addr"] = true
	}

	// keep track of all unknown connection errors
	if bitcoinResult.ConnectErrorStr == models.NetErrorUnknown && bitcoinResult.ConnectError != nil {
		properties["connect_error"] = bitcoinResult.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if bitcoinResult.ErrorStr == models.NetErrorUnknown && bitcoinResult.Error != nil {
		properties["crawl_error"] = bitcoinResult.Error.Error()
	}

	data, err := json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	if len(bitcoinResult.ListenAddrs) > 0 {
		task.AddrInfo.Addr = bitcoinResult.ListenAddrs
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        bitcoinResult.RoutingTable,
		Agent:               bitcoinResult.Agent,
		Protocols:           bitcoinResult.Protocols,
		ConnectError:        bitcoinResult.ConnectError,
		ConnectErrorStr:     bitcoinResult.ConnectErrorStr,
		CrawlError:          bitcoinResult.Error,
		CrawlErrorStr:       bitcoinResult.ErrorStr,
		CrawlEndTime:        time.Now(),
		ConnectStartTime:    bitcoinResult.ConnectStartTime,
		ConnectEndTime:      bitcoinResult.ConnectEndTime,
		Properties:          data,
		LogErrors:           c.cfg.LogErrors,
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

func (c *Crawler) PeerProperties(node *AddrInfo) map[string]any {
	// TODO: to be implemented later
	properties := map[string]any{}

	properties["NA"] = true

	return properties
}

type BitcoinResult struct {
	ConnectStartTime      time.Time
	ConnectEndTime        time.Time
	ConnectError          error
	ConnectErrorStr       string
	Agent                 string
	ProtocolVersion       int32
	Protocols             []string
	ListenAddrs           []ma.Multiaddr
	Transport             string // the transport of a successful connection
	ConnClosedImmediately bool   // whether conn was no error but still unconnected
	GenTCPAddr            bool   // whether a TCP address was generated
	Error                 error
	ErrorStr              string
	RoutingTable          *core.RoutingTable[PeerInfo]
}

func (c *Crawler) crawlBitcoin(ctx context.Context, pi PeerInfo) chan BitcoinResult {
	resultCh := make(chan BitcoinResult)

	go func() {
		result := BitcoinResult{}

		// sanitize the given addresses like removing UDP-only addresses and
		// adding corresponding TCP addresses.
		sanitizedAddrs, generated := sanitizeAddrs(pi.Addrs())

		// keep track if we generated a TCP address to dial
		result.GenTCPAddr = generated

		addrInfo := peer.AddrInfo{
			ID:    pi.ID(),
			Addrs: sanitizedAddrs,
		}

		var conn net.Conn
		result.ConnectStartTime = time.Now()
		conn, result.ConnectError = c.connect(ctx, addrInfo) // use filtered addr list
		c.conn = conn
		result.ConnectEndTime = time.Now()

		neighbours := make([]PeerInfo, 0, 50)

		// If we could successfully connect to the peer we actually crawl it.
		if result.ConnectError == nil {

			nodeRes, err := c.Handshake()
			result.Agent = nodeRes.UserAgent
			result.ProtocolVersion = nodeRes.ProtocolVersion
			if err != nil {
				log.Debugf("[%s] Handshake failed: %v", sanitizedAddrs, err)
			}

			err = c.WriteMessage(wire.NewMsgGetAddr())
			if err != nil {
				log.Warningf("[%s] GetAddr failed: %v", sanitizedAddrs, err)
			}

			// keep track of the transport of the open connection
			result.Transport = "tcp"

			firstReceived := -1
			tolerateMessages := 5
			// The nodes send a lot of inv messages
			tolerateInvMessages := 50
			tolerateVerAckMessages := 10
			toleratePingMessages := 10

			otherMessages := []string{}
			for {
				// We can't really tell when we're done receiving peers, so we stop either
				// when we get a smaller-than-normal set size or when we've received too
				// many unrelated messages.

				if len(otherMessages) > tolerateMessages {
					log.WithField("address", pi.Addr).WithField("num_peers", len(neighbours)).WithField("otherMessages", otherMessages).Debugf("Giving up with results after tolerating messages")
					break
				}

				msg, _, err := c.ReadMessage()
				if err != nil {
					otherMessages = append(otherMessages, err.Error())
					// log.WithField("address", addrInfo).WithField("num_peers", len(neighbours)).WithField("err", err).WithField("otherMessages", otherMessages).Warningf("Giving up with results after tolerating messages: .")
					log.Warningf("[%s] Failed to read message: %v", pi.Addr, err)
					continue
				}

				switch tmsg := msg.(type) {
				case *wire.MsgAddr:
					neighbours = append(neighbours, func() []PeerInfo {
						mapped := make([]PeerInfo, len(tmsg.AddrList))
						for i, addr := range tmsg.AddrList {
							maStr := fmt.Sprintf("/ip4/%s/tcp/%d", addr.IP.String(), addr.Port)
							maddr, err := ma.NewMultiaddr(maStr)
							if err != nil {
								continue // Skip invalid addresses
							}

							mapped[i] = PeerInfo{
								AddrInfo: AddrInfo{
									id:   maddr.String(),
									Addr: []ma.Multiaddr{maddr},
								},
							}
						}
						return mapped
					}()...)

					if firstReceived == -1 {
						firstReceived = len(tmsg.AddrList)
					} else if firstReceived > len(tmsg.AddrList) || firstReceived == 0 {
						// Probably done.
						break
					}
				case *wire.MsgAddrV2:
					neighbours = append(neighbours, func() []PeerInfo {
						mapped := make([]PeerInfo, len(tmsg.AddrList))
						for i, addr1 := range tmsg.AddrList {
							addr := addr1.ToLegacy()
							maStr := fmt.Sprintf("/ip4/%s/tcp/%d", addr.IP.String(), addr.Port)
							maddr, err := ma.NewMultiaddr(maStr)
							if err != nil {
								continue // Skip invalid addresses
							}

							mapped[i] = PeerInfo{
								AddrInfo: AddrInfo{
									id:   maddr.String(),
									Addr: []ma.Multiaddr{maddr},
								},
							}
						}
						return mapped
					}()...)

					if firstReceived == -1 {
						firstReceived = len(tmsg.AddrList)
					} else if firstReceived > len(tmsg.AddrList) || firstReceived == 0 {
						// Probably done.
						break
					}
				case *wire.MsgPing:
					log.Infoln("Sending Pong message...")
					toleratePingMessages--
					err = c.WriteMessage(wire.NewMsgPong(tmsg.Nonce))
					if err != nil {
						log.Infof("Pong msg err: %s", err)
						break
					}
					if toleratePingMessages < 0 {
						log.Infoln("Ran out of limit to tolerate Ping messages")
						break
					}
				case *wire.MsgVerAck:
					tolerateVerAckMessages--
					if tolerateVerAckMessages < 0 {
						log.Infoln("Ran out of limit to tolerate Ver Ack messages")
						break
					}
				case *wire.MsgInv:
					tolerateInvMessages--
					if tolerateInvMessages < 0 {
						log.Debugln("Ran out of limit to tolerate Inv messages")
						break
					}
				default:
					otherMessages = append(otherMessages, tmsg.Command())
					log.WithField("msg_type", tmsg.Command()).Infof("Found other message from %s", pi.Addr)
				}
				err = c.WriteMessage(wire.NewMsgGetAddr())
				if err != nil {
					log.Infoln("Error when requesting peers")
				}

			}

			if len(neighbours) > 0 {
				log.WithField("num_peers", len(neighbours)).WithField("addr", pi.Addr).Infoln("Found peers")
			} else {
				log.WithField("addr", pi.Addr).Infoln("Found no peers")
			}

		} else {
			result.Error = result.ConnectError
		}

		result.RoutingTable = &core.RoutingTable[PeerInfo]{
			PeerID:    pi.ID(),
			Neighbors: neighbours,
			ErrorBits: uint16(0), // FIXME
			Error:     result.Error,
		}

		// if there was a connection error, parse it to a known one
		if result.ConnectError != nil {
			result.ConnectErrorStr = db.NetError(result.ConnectError)
		} else {
			// Free connection resources
			if err := c.conn.Close(); err != nil {
				log.WithError(err).WithField("remoteID", pi.ID().ShortString()).Warnln("Could not close connection to peer")
			}
		}

		if result.Error != nil {
			result.ErrorStr = db.NetError(result.Error)
		}

		// send the result back and close channel
		select {
		case resultCh <- result:
		case <-ctx.Done():
		}

		close(resultCh)
	}()

	return resultCh
}

type BitcoinNodeResult struct {
	ProtocolVersion int32
	UserAgent       string
	pver            int32
}

func (c *Crawler) Handshake() (BitcoinNodeResult, error) {
	result := BitcoinNodeResult{}
	if c.conn == nil {
		return result, fmt.Errorf("peer is not connected, can't handshake")
	}

	log.WithField("Address", c.conn.RemoteAddr()).Debug("Starting handshake.")

	nonce, err := wire.RandomUint64()
	if err != nil {
		return result, err
	}

	localAddr := wire.NewNetAddressIPPort(
		c.conn.LocalAddr().(*net.TCPAddr).IP,
		uint16(c.conn.LocalAddr().(*net.TCPAddr).Port),
		wire.SFNodeNetwork,
	)
	remoteAddr := wire.NewNetAddressIPPort(
		c.conn.RemoteAddr().(*net.TCPAddr).IP,
		uint16(c.conn.RemoteAddr().(*net.TCPAddr).Port),
		wire.SFNodeNetwork,
	)

	msgVersion := wire.NewMsgVersion(localAddr, remoteAddr, nonce, 0)

	msgVersion.ProtocolVersion = int32(wire.ProtocolVersion)
	msgVersion.Services = wire.SFNodeNetwork
	msgVersion.Timestamp = time.Now()
	msgVersion.UserAgent = "nebula/" + c.cfg.Version

	if err := c.WriteMessage(msgVersion); err != nil {
		return result, err
	}

	// Read the response version.
	msg, _, err := c.ReadMessage()
	if err != nil {
		return result, err
	}
	vmsg, ok := msg.(*wire.MsgVersion)
	if !ok {
		return result, fmt.Errorf("did not receive version message: %T", vmsg)
	}

	result.ProtocolVersion = vmsg.ProtocolVersion
	result.UserAgent = vmsg.UserAgent

	// Negotiate protocol version.
	if uint32(vmsg.ProtocolVersion) < wire.ProtocolVersion {
		result.pver = vmsg.ProtocolVersion
	}
	log.Debugf("[%s] -> Version: %s", c.conn.RemoteAddr(), vmsg.UserAgent)

	// Normally we'd check if vmsg.Nonce == p.nonce but the crawler does not
	// accept external connections so we skip it.

	// Send verack.
	if err := c.WriteMessage(wire.NewMsgVerAck()); err != nil {
		return result, err
	}

	return result, nil
}

// connect establishes a connection to the given peer. It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) (net.Conn, error) {
	if len(pi.Addrs) == 0 {
		return nil, fmt.Errorf("skipping node as it has no public IP address")
	}

	// init an exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.MaxInterval = 4 * time.Second
	bo.MaxElapsedTime = time.Minute

	var retry int = 0

	for {
		logEntry := log.WithFields(log.Fields{
			"timeout":  c.cfg.DialTimeout.String(),
			"remoteID": pi.ID.String(),
			"retry":    retry,
			"maddrs":   pi.Addrs,
		})
		logEntry.Debugln("Connecting to peer", pi.ID.ShortString())

		netAddr, _ := manet.ToNetAddr(pi.Addrs[0])

		conn, err := net.DialTimeout(netAddr.Network(), netAddr.String(), c.cfg.DialTimeout) // TODO: change this dialout to 5 secs

		if err == nil {
			return conn, nil
		}

		// TODO: support actual bitcoin errors
		switch {
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionRefused]):
			// Might be transient because the remote doesn't want us to connect. Try again!
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorConnectionGated]):
			// Hints at a configuration issue and should not happen, but if it
			// does it could be transient. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Connection gated!")
		case strings.Contains(err.Error(), db.ErrorStr[models.NetErrorCantAssignRequestedAddress]):
			// Transient error due to local UDP issues. Try again!
		case strings.Contains(err.Error(), "dial backoff"):
			// should not happen because we disabled backoff checks with our
			// go-libp2p fork. Try again anyway, but at least log a warning.
			logEntry.WithError(err).Warnln("Dial backoff!")
		case strings.Contains(err.Error(), "RESOURCE_LIMIT_EXCEEDED (201)"): // thrown by a circuit relay
			// We already have too many open connections over a relay. Try again!
		default:
			logEntry.WithError(err).Debugln("Failed connecting to peer", pi.ID.ShortString())
			return nil, err
		}

		sleepDur := bo.NextBackOff()
		if sleepDur == backoff.Stop {
			logEntry.WithError(err).Debugln("Exceeded retries connecting to peer", pi.ID.ShortString())
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleepDur):
			retry += 1
			continue
		}

	}
}

func (c *Crawler) WriteMessage(msg wire.Message) error {
	return wire.WriteMessage(c.conn, msg, wire.ProtocolVersion, wire.MainNet)
}

func (c *Crawler) ReadMessage() (wire.Message, []byte, error) {
	return wire.ReadMessage(c.conn, wire.ProtocolVersion, wire.MainNet)
}

// sanitizeAddrs takes the list of multi addresses and removes any UDP-only
// multi address because we cannot dial UDP only addresses anyway. However, if
// there is no other reliable transport address like TCP or QUIC we use the UDP
// IP address + port and craft a TCP address out of it. The UDP address will
// still be removed and replaced with TCP.
func sanitizeAddrs(maddrs []ma.Multiaddr) ([]ma.Multiaddr, bool) {
	newMaddrs := make([]ma.Multiaddr, 0, len(maddrs))
	for _, maddr := range maddrs {
		if _, err := maddr.ValueForProtocol(ma.P_TCP); err == nil {
			newMaddrs = append(newMaddrs, maddr)
		} else if _, err := maddr.ValueForProtocol(ma.P_UDP); err == nil {
			_, quicErr := maddr.ValueForProtocol(ma.P_QUIC)
			_, quicV1Err := maddr.ValueForProtocol(ma.P_QUIC_V1)
			if quicErr == nil || quicV1Err == nil {
				newMaddrs = append(newMaddrs, maddr)
			}
		}
	}

	if len(newMaddrs) > 0 {
		return newMaddrs, false
	}

	for i, maddr := range maddrs {
		udp, err := maddr.ValueForProtocol(ma.P_UDP)
		if err != nil {
			continue
		}

		ip := ""
		ip4, err := maddr.ValueForProtocol(ma.P_IP4)
		if err != nil {
			ip6, err := maddr.ValueForProtocol(ma.P_IP6)
			if err != nil {
				continue
			}
			ip = "/ip6/" + ip6
		} else {
			ip = "/ip4/" + ip4
		}

		tcpMaddr, err := ma.NewMultiaddr(ip + "/tcp/" + udp)
		if err != nil {
			continue
		}

		newMaddrs = append(newMaddrs, maddrs[i+1:]...)
		newMaddrs = append(newMaddrs, tcpMaddr)

		return newMaddrs, true
	}

	return maddrs, false
}

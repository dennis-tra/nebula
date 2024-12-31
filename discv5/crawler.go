package discv5

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-msgio/pbio"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/waku-org/go-waku/waku/v2/protocol/metadata/pb"
	"go.uber.org/atomic"

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
	MaxJitter    time.Duration
}

type Crawler struct {
	id           string
	cfg          *CrawlerConfig
	host         *basichost.BasicHost
	listener     *discover.UDPv5
	crawledPeers int
	done         chan struct{}
}

var _ core.Worker[PeerInfo, core.CrawlResult[PeerInfo]] = (*Crawler)(nil)

func (c *Crawler) Work(ctx context.Context, task PeerInfo) (core.CrawlResult[PeerInfo], error) {
	// add a startup jitter delay to prevent all workers to crawl at exactly the
	// same time and potentially overwhelm the machine that Nebula is running on
	if c.crawledPeers == 0 {
		jitter := time.Duration(0)
		if c.cfg.MaxJitter > 0 { // could be <= 0 if the worker count is 1
			jitter = time.Duration(rand.Int63n(int64(c.cfg.MaxJitter)))
		}
		select {
		case <-time.After(jitter):
		case <-ctx.Done():
		}
	}

	logEntry := log.WithFields(log.Fields{
		"crawlerID":  c.id,
		"remoteID":   task.peerID.ShortString(),
		"crawlCount": c.crawledPeers,
	})
	logEntry.Debugln("Crawling peer")
	defer logEntry.Debugln("Crawled peer")

	crawlStart := time.Now()

	// start crawling both ways
	libp2pResultCh := c.crawlLibp2p(ctx, task)
	discV5ResultCh := c.crawlDiscV5(ctx, task)

	libp2pResult := <-libp2pResultCh
	discV5Result := <-discV5ResultCh

	properties := c.PeerProperties(task.Node)

	if libp2pResult.Transport != "" {
		properties["transport"] = libp2pResult.Transport
	}

	if libp2pResult.ConnClosedImmediately {
		properties["direct_close"] = true
	}

	if libp2pResult.GenTCPAddr {
		properties["gen_tcp_addr"] = true
	}

	// keep track of all unknown connection errors
	if libp2pResult.ConnectErrorStr == models.NetErrorUnknown && libp2pResult.ConnectError != nil {
		properties["connect_error"] = libp2pResult.ConnectError.Error()
	}

	// keep track of all unknown crawl errors
	if discV5Result.ErrorStr == models.NetErrorUnknown && discV5Result.Error != nil {
		properties["crawl_error"] = discV5Result.Error.Error()
	}

	data, err := json.Marshal(properties)
	if err != nil {
		log.WithError(err).WithField("properties", properties).Warnln("Could not marshal peer properties")
	}

	if len(libp2pResult.ListenAddrs) > 0 {
		task.maddrs = libp2pResult.ListenAddrs
	}

	cr := core.CrawlResult[PeerInfo]{
		CrawlerID:           c.id,
		Info:                task,
		CrawlStartTime:      crawlStart,
		RoutingTableFromAPI: false,
		RoutingTable:        discV5Result.RoutingTable,
		Agent:               libp2pResult.Agent,
		Protocols:           libp2pResult.Protocols,
		ConnectError:        libp2pResult.ConnectError,
		ConnectErrorStr:     libp2pResult.ConnectErrorStr,
		CrawlError:          discV5Result.Error,
		CrawlErrorStr:       discV5Result.ErrorStr,
		CrawlEndTime:        time.Now(),
		ConnectStartTime:    libp2pResult.ConnectStartTime,
		ConnectEndTime:      libp2pResult.ConnectEndTime,
		Properties:          data,
		LogErrors:           c.cfg.LogErrors,
		WakuCluster:         libp2pResult.WakuCluster,
	}

	// We've now crawled this peer, so increment
	c.crawledPeers++

	return cr, nil
}

func (c *Crawler) PeerProperties(node *enode.Node) map[string]any {
	properties := map[string]any{}

	if ip := node.IP(); ip != nil {
		properties["ip"] = ip.String()
	}

	properties["seq"] = node.Record().Seq()
	properties["signature"] = node.Record().Signature()

	if node.UDP() != 0 {
		properties["udp"] = node.UDP()
	}

	if node.TCP() != 0 {
		properties["tcp"] = node.TCP()
	}

	var enrEntryEth2 ENREntryEth2
	if err := node.Load(&enrEntryEth2); err == nil {
		properties["fork_digest"] = enrEntryEth2.ForkDigest.String()
		properties["next_fork_version"] = enrEntryEth2.NextForkVersion.String()
		properties["next_fork_epoch"] = enrEntryEth2.NextForkEpoch.String()
	}

	var enrEntryAttnets ENREntryAttnets
	if err := node.Load(&enrEntryAttnets); err == nil {
		properties["attnets_num"] = enrEntryAttnets.AttnetsNum
		properties["attnets"] = enrEntryAttnets.Attnets
	}

	var enrEntrySyncCommsSubnet ENREntrySyncCommsSubnet
	if err := node.Load(&enrEntrySyncCommsSubnet); err == nil {
		properties["syncnets"] = enrEntrySyncCommsSubnet.SyncNets
	}

	var enrEntryOpStack ENREntryOpStack
	if err := node.Load(&enrEntryOpStack); err == nil {
		properties["opstack_chain_id"] = enrEntryOpStack.ChainID
		properties["opstack_version"] = enrEntryOpStack.Version
	}

	if c.cfg.KeepENR {
		properties["enr"] = node.String()
	}

	return properties
}

type Libp2pResult struct {
	ConnectStartTime      time.Time
	ConnectEndTime        time.Time
	ConnectError          error
	ConnectErrorStr       string
	Agent                 string
	Protocols             []string
	ListenAddrs           []ma.Multiaddr
	Transport             string // the transport of a successful connection
	ConnClosedImmediately bool   // whether conn was no error but still unconnected
	GenTCPAddr            bool   // whether a TCP address was generated
	WakuCluster           uint32
}

func (c *Crawler) crawlLibp2p(ctx context.Context, pi PeerInfo) chan Libp2pResult {
	resultCh := make(chan Libp2pResult)

	go func() {
		result := Libp2pResult{}

		// sanitize the given addresses like removing UDP-only addresses and
		// adding corresponding TCP addresses.
		sanitizedAddrs, generated := sanitizeAddrs(pi.Addrs())

		// keep track if we generated a TCP address to dial
		result.GenTCPAddr = generated

		addrInfo := peer.AddrInfo{
			ID:    pi.ID(),
			Addrs: sanitizedAddrs,
		}

		var conn network.Conn
		result.ConnectStartTime = time.Now()
		conn, result.ConnectError = c.connect(ctx, addrInfo) // use filtered addr list
		result.ConnectEndTime = time.Now()
		c.host.SetStreamHandler("/vac/waku/metadata/1.0.0", func(stream network.Stream) {
			request := &pb.WakuMetadataRequest{}
			writer := pbio.NewDelimitedWriter(stream)
			reader := pbio.NewDelimitedReader(stream, math.MaxInt32)
			err := reader.ReadMsg(request)
			if err != nil {
				// fmt.Println("reading request from peer", stream.Conn().RemotePeer(), err)
				if err := stream.Reset(); err != nil {
					fmt.Println("resetting connection", err)
				}
				return
			}
			// fmt.Println("received metadata request from peer ", stream.Conn().RemotePeer(), request)
			result.WakuCluster = *request.ClusterId
			response := new(pb.WakuMetadataResponse)
			// TODO: fetch from config
			clusterId := uint32(16)
			response.ClusterId = &clusterId
			response.Shards = []uint32{32}
			err = writer.WriteMsg(response)
			if err != nil {
				// fmt.Println("writing response to peer", stream.Conn().RemotePeer(), err)
				if err := stream.Reset(); err != nil {
					fmt.Println("resetting connection", err)
				}
				return
			}
			// fmt.Println("sent metadata response to peer", stream.Conn().RemotePeer(), response)

		})
		// If we could successfully connect to the peer we actually crawl it.
		if result.ConnectError == nil {

			// keep track of the transport of the open connection
			result.Transport = conn.ConnState().Transport

			// wait for the Identify exchange to complete (no-op if already done)
			// the internal timeout is set to 30 s. When crawling we only allow 5s.
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				// identification timed out.
			case <-c.host.IDService().IdentifyWait(conn):
				// identification may have succeeded.
			}

			// Extract information from peer store
			ps := c.host.Peerstore()

			// Extract agent
			if agent, err := ps.Get(pi.ID(), "AgentVersion"); err == nil {
				result.Agent = agent.(string)
			}

			// Extract protocols
			if protocols, err := ps.GetProtocols(pi.ID()); err == nil {
				result.Protocols = make([]string, len(protocols))
				for i := range protocols {
					result.Protocols[i] = string(protocols[i])
				}
			}

			// Extract listen addresses
			result.ListenAddrs = ps.Addrs(pi.ID())
		}

		// if there was a connection error, parse it to a known one
		if result.ConnectError != nil {
			result.ConnectErrorStr = db.NetError(result.ConnectError)
		}

		// Free connection resources
		if err := c.host.Network().ClosePeer(pi.ID()); err != nil {
			log.WithError(err).WithField("remoteID", pi.ID().ShortString()).Warnln("Could not close connection to peer")
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

// connect establishes a connection to the given peer. It also handles metric capturing.
func (c *Crawler) connect(ctx context.Context, pi peer.AddrInfo) (network.Conn, error) {
	if len(pi.Addrs) == 0 {
		return nil, fmt.Errorf("skipping node as it has no public IP address") // change knownErrs map if changing this msg
	}

	// init an exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.MaxInterval = 10 * time.Second
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

		// save addresses into the peer store temporarily
		c.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)

		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
		conn, err := c.host.Network().DialPeer(timeoutCtx, pi.ID)
		cancel()

		if err == nil {
			return conn, nil
		}

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

		for _, remaining := range maddrs[i+1:] {
			newMaddrs = append(newMaddrs, remaining)
		}

		newMaddrs = append(newMaddrs, tcpMaddr)

		return newMaddrs, true
	}

	return maddrs, false
}

type DiscV5Result struct {
	// The time we received the first successful response
	RespondedAt *time.Time

	// The updated ethereum node record
	ENR *enode.Node

	// The neighbors of the crawled peer
	RoutingTable *core.RoutingTable[PeerInfo]

	// The time the draining of bucket entries was finished
	DoneAt time.Time

	// The combined error of crawling the peer's buckets
	Error error

	// The above error mapped to a known string
	ErrorStr string
}

func (c *Crawler) crawlDiscV5(ctx context.Context, pi PeerInfo) chan DiscV5Result {
	resultCh := make(chan DiscV5Result)

	go func() {
		result := DiscV5Result{}

		// all neighbors of pi. We're using a map to deduplicate.
		allNeighbors := map[string]PeerInfo{}

		// errorBits tracks at which CPL errors have occurred.
		// 0000 0000 0000 0000 - No error
		// 0000 0000 0000 0001 - An error has occurred at CPL 0
		// 1000 0000 0000 0001 - An error has occurred at CPL 0 and 15
		errorBits := atomic.NewUint32(0)

		timeouts := 0
		enr, err := c.listener.RequestENR(pi.Node)
		if err != nil {
			timeouts += 1
			result.ENR = pi.Node
		} else {
			result.ENR = enr
			now := time.Now()
			result.RespondedAt = &now
		}

		// loop through the buckets sequentially because discv5 is also doing that
		// internally, so we won't gain much by spawning multiple parallel go
		// routines here. Stop the process as soon as we have received a timeout and
		// don't let the following calls time out as well.
		for i := 0; i <= discover.NBuckets; i++ { // 17 is maximum
			var neighbors []*enode.Node
			neighbors, err = c.listener.FindNode(pi.Node, []uint{uint(discover.HashBits - i)})
			if err != nil {

				if errors.Is(err, discover.ErrTimeout) {
					timeouts += 1
					if timeouts < MaxCrawlRetriesAfterTimeout {
						continue
					}
				}

				errorBits.Add(1 << i)
				err = fmt.Errorf("getting closest peer with CPL %d: %w", i, err)
				break
			}
			timeouts = 0

			if result.RespondedAt == nil {
				now := time.Now()
				result.RespondedAt = &now
			}

			for _, n := range neighbors {
				npi, err := NewPeerInfo(n)
				if err != nil {
					log.WithError(err).Warnln("Failed parsing ethereum node neighbor")
					continue
				}
				allNeighbors[string(npi.peerID)] = npi
			}
		}

		result.DoneAt = time.Now()
		// if we have at least a successful result, don't record error
		if noSuccessfulRequest(err, errorBits.Load()) {
			result.Error = err
		}

		result.RoutingTable = &core.RoutingTable[PeerInfo]{
			PeerID:    pi.ID(),
			Neighbors: []PeerInfo{},
			ErrorBits: uint16(errorBits.Load()),
			Error:     result.Error,
		}

		for _, n := range allNeighbors {
			result.RoutingTable.Neighbors = append(result.RoutingTable.Neighbors, n)
		}

		// if there was a connection error, parse it to a known one
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

// noSuccessfulRequest returns true if the given error is non nil, and all bits
// of the given errorBits are set. This means that no successful request has
// been made. This is equivalent to verifying that all righmost bits are equal
// to 1, or that the errorBits is a power of 2 minus 1.
//
// Examples:
// 0b00000011 -> true
// 0b00000111 -> true
// 0b00001101 -> false
func noSuccessfulRequest(err error, errorBits uint32) bool {
	return err != nil && errorBits&(errorBits+1) == 0
}

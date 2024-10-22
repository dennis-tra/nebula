package discv4

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"math"
	"net"
	"net/netip"
	"sync"
	"syscall"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/tele"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type PeerInfo struct {
	*enode.Node
	peerID  peer.ID
	maddrs  []ma.Multiaddr
	udpAddr netip.AddrPort
}

var _ core.PeerInfo[PeerInfo] = (*PeerInfo)(nil)

func NewPeerInfo(node *enode.Node) (PeerInfo, error) {
	pubKey := node.Pubkey()
	if pubKey == nil {
		return PeerInfo{}, fmt.Errorf("no public key")
	}

	pubBytes := elliptic.Marshal(secp256k1.S256(), pubKey.X, pubKey.Y)
	secpKey, err := crypto.UnmarshalSecp256k1PublicKey(pubBytes)
	if err != nil {
		return PeerInfo{}, fmt.Errorf("unmarshal secp256k1 public key: %w", err)
	}

	peerID, err := peer.IDFromPublicKey(secpKey)
	if err != nil {
		return PeerInfo{}, fmt.Errorf("peer ID from public key: %w", err)
	}

	var ipScheme string
	if p4 := node.IP().To4(); len(p4) == net.IPv4len {
		ipScheme = "ip4"
	} else {
		ipScheme = "ip6"
	}

	maddrs := []ma.Multiaddr{}
	if node.UDP() != 0 {
		maddrStr := fmt.Sprintf("/%s/%s/udp/%d", ipScheme, node.IP(), node.UDP())
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return PeerInfo{}, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		}
		maddrs = append(maddrs, maddr)
	}

	if node.TCP() != 0 {
		maddrStr := fmt.Sprintf("/%s/%s/tcp/%d", ipScheme, node.IP(), node.TCP())
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return PeerInfo{}, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		}
		maddrs = append(maddrs, maddr)
	}

	ipAddr, ok := netip.AddrFromSlice(node.IP())
	if !ok {
		return PeerInfo{}, fmt.Errorf("failed to convert ip to netip.Addr: %s", node.IP())
	}
	udpAddr := netip.AddrPortFrom(ipAddr, uint16(node.UDP()))

	pi := PeerInfo{
		Node:    node,
		peerID:  peerID,
		maddrs:  maddrs,
		udpAddr: udpAddr,
	}

	return pi, nil
}

func (p PeerInfo) ID() peer.ID {
	return p.peerID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.maddrs
}

func (p PeerInfo) Merge(other PeerInfo) PeerInfo {
	p.maddrs = utils.MergeMaddrs(p.maddrs, other.maddrs)
	return p
}

func (p PeerInfo) DeduplicationKey() string {
	return p.Node.String()
}

type CrawlDriverConfig struct {
	Version          string
	TrackNeighbors   bool
	CrawlWorkerCount int
	DialTimeout      time.Duration
	BootstrapPeers   []*enode.Node
	AddrDialType     config.AddrType
	AddrTrackType    config.AddrType
	MeterProvider    metric.MeterProvider
	TracerProvider   trace.TracerProvider
	LogErrors        bool
	KeepENR          bool
	UDPBufferSize    int
	UDPRespTimeout   time.Duration
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:  cfg.DialTimeout,
		AddrDialType: cfg.AddrDialType,
		LogErrors:    cfg.LogErrors,
		MaxJitter:    time.Duration(cfg.CrawlWorkerCount/50) * time.Second, // e.g., 3000 workers evenly distributed over 60s
		KeepENR:      false,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{
		AddrTrackType: cfg.AddrTrackType,
	}
}

type CrawlDriver struct {
	cfg            *CrawlDriverConfig
	dbc            db.Client
	client         *Client
	dbCrawl        *models.Crawl
	tasksChan      chan PeerInfo
	peerstore      *enode.DB
	crawlerCount   int
	writerCount    int
	crawler        []*Crawler
	unhandledChan  chan discover.ReadPacket
	taskDoneAtChan chan time.Time

	// Telemetry
	unhandledPacketsCounter metric.Int64Counter
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("new ethereum ecdsa key: %w", err)
	}

	clientCfg := DefaultClientConfig()
	clientCfg.DialTimeout = cfg.DialTimeout
	client := NewClient(priv, clientCfg)

	peerstore, err := enode.OpenDB("") // in memory db
	if err != nil {
		return nil, fmt.Errorf("open in-memory peerstore: %w", err)
	}

	// Init channels:
	//   unhandledChan:  this is a channel that will receive all unhandled
	//                   packets from all discv4 UDP listeners.
	//   tasksChan:      this is the channel that the engine will consume. It
	//                   receives all peers that should be crawled.
	//   taskDoneAtChan: every time a crawl worker has completed one crawl, it
	//                   will emit a timestamp on this channel. We use this in
	//                   the monitoring of unhandled packets. If the last crawl
	//                   is longer ago than 10s, and we haven't received an
	//                   unhandled Neighbors packet, we close the channel
	unhandledChan := make(chan discover.ReadPacket, discover.BucketSize*cfg.CrawlWorkerCount)
	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	taskDoneAtChan := make(chan time.Time, cfg.CrawlWorkerCount)

	for _, node := range cfg.BootstrapPeers {
		pi, err := NewPeerInfo(node)
		if err != nil {
			return nil, fmt.Errorf("new peer info from enr: %w", err)
		}
		tasksChan <- pi
	}

	meter := cfg.MeterProvider.Meter(tele.MeterName)
	unhandledPacketsCounter, err := meter.Int64Counter("unhandled_packets")
	if err != nil {
		return nil, fmt.Errorf("create unhandled packets counter: %w", err)
	}

	// set the discovery response timeout
	discover.RespTimeout = cfg.UDPRespTimeout

	d := &CrawlDriver{
		cfg:            cfg,
		dbc:            dbc,
		client:         client,
		dbCrawl:        crawl,
		peerstore:      peerstore,
		tasksChan:      tasksChan,
		taskDoneAtChan: taskDoneAtChan,
		unhandledChan:  unhandledChan,
		crawler:        make([]*Crawler, 0, cfg.CrawlWorkerCount),

		// Telemetry
		unhandledPacketsCounter: unhandledPacketsCounter,
	}

	// hand responsibility of tasksChan to this function. It will close the
	// channel if the workers have been idle for more than 10s. This will signal
	// the engine that we also don't expect any more late unhandled packets.
	d.monitorUnhandledPackets()

	return d, nil
}

// NewWorker is called multiple times but only log the configured buffer sizes once
var logOnce sync.Once

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
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

	if err = conn.SetReadBuffer(d.cfg.UDPBufferSize); err != nil {
		log.Warnln("Failed to set read buffer size on UDP listener", err)
	}

	rcvbuf, sndbuf, err := getUDPBufferSize(conn)
	logOnce.Do(func() {
		logEntry := log.WithFields(log.Fields{
			"rcvbuf": rcvbuf,
			"sndbuf": sndbuf,
			"rcvtgt": d.cfg.UDPBufferSize, // receive target
		})
		if rcvbuf < d.cfg.UDPBufferSize {
			logEntry.Warnln("Failed to increase UDP buffer sizes, using default")
		} else {
			logEntry.Infoln("Configured UDP buffer sizes")
		}
	})

	log.Debugln("Listening on UDP port ", conn.LocalAddr().String(), " for Ethereum discovery")

	discvxCfg := discover.Config{
		PrivateKey: priv,
		Unhandled:  d.unhandledChan,
	}
	listener, err := discover.ListenV4(conn, ethNode, discvxCfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv4: %w", err)
	}

	crawlerCfg := d.cfg.CrawlerConfig()
	crawlerCfg.KeepENR = d.cfg.KeepENR

	c := &Crawler{
		id:           fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:          crawlerCfg,
		client:       d.client,
		listener:     listener,
		taskDoneChan: d.taskDoneAtChan,
		done:         make(chan struct{}),
	}

	d.crawlerCount += 1

	d.crawler = append(d.crawler, c)

	log.WithFields(log.Fields{
		"addr": conn.LocalAddr().String(),
	}).Debugln("Started crawler worker", c.id)

	return c, nil
}

func (d *CrawlDriver) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", d.writerCount), d.dbc, d.dbCrawl.ID, d.cfg.WriterConfig())
	d.writerCount += 1
	return w, nil
}

func (d *CrawlDriver) Tasks() <-chan PeerInfo {
	return d.tasksChan
}

func (d *CrawlDriver) Close() {
	for _, c := range d.crawler {
		c.listener.Close()
	}
	close(d.unhandledChan)

	// wait for the go routine that reads the unhandled packets to close
	select {
	case <-d.tasksChan:
	case <-time.After(time.Second):
		log.Warnln("Timed out waiting for packetsDone channel to close")
	}
}

func (d *CrawlDriver) monitorUnhandledPackets() {
	go func() {
		defer close(d.tasksChan)

		timeout := 10 * time.Second
		latestTaskDone := time.Now()
		timer := time.NewTimer(math.MaxInt64)

	LOOP:
		for {
			select {
			case <-timer.C:
				log.Infof("No Neighbors packet received from any crawler worker for %s. Stop monitoring unhandled packets.", timeout)
				break LOOP
			case taskDoneAt := <-d.taskDoneAtChan:
				if taskDoneAt.After(latestTaskDone) {
					latestTaskDone = taskDoneAt
					timer.Reset(timeout)
				}
			case packet, more := <-d.unhandledChan:
				if !more {
					break LOOP
				}

				rawpacket, _, _, err := v4wire.Decode(packet.Data)
				if err != nil {
					continue
				}

				neighborsPacket, ok := rawpacket.(*v4wire.Neighbors)
				if !ok {
					continue
				}

				d.unhandledPacketsCounter.Add(context.TODO(), 1)
				for _, n := range neighborsPacket.Nodes {
					node, err := discover.NodeFromRPC(packet.Addr, n, nil)
					if err != nil {
						continue
					}

					pi, err := NewPeerInfo(node)
					if err != nil {
						continue
					}

					d.tasksChan <- pi
				}
			}
		}
	}()
}

// getUDPBufferSize reads the receive and send buffer sizes from the system
func getUDPBufferSize(conn *net.UDPConn) (rcvbuf int, sndbuf int, err error) {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return 0, 0, err
	}

	var (
		rcverr error
		snderr error
	)
	err = rawConn.Control(func(fd uintptr) {
		rcvbuf, rcverr = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		sndbuf, snderr = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
	})
	if rcverr != nil {
		err = rcverr
	} else if snderr != nil {
		err = snderr
	}

	return
}

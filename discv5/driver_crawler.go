package discv5

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"io"
	"math"
	"net"
	"net/netip"
	"runtime"
	"sync"
	"time"

	secp256k1v4 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p"
	mplex "github.com/libp2p/go-libp2p-mplex"
	libp2pconfig "github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/dennis-tra/nebula-crawler/config"
	"github.com/dennis-tra/nebula-crawler/core"
	"github.com/dennis-tra/nebula-crawler/db"
	"github.com/dennis-tra/nebula-crawler/db/models"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type PeerInfo struct {
	*enode.Node
	peerID peer.ID
	maddrs []ma.Multiaddr
	enr    string
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

	if quicAddr, ok := node.QUICEndpoint(); ok {
		addr := quicAddr.Addr()
		if addr.Is4() {
			ipScheme = "ip4"
		} else {
			ipScheme = "ip6"
		}

		maddrStr := fmt.Sprintf("/%s/%s/udp/%d/quic-v1", ipScheme, addr.String(), quicAddr.Port())
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return PeerInfo{}, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		}

		maddrs = append(maddrs, maddr)
	}

	pi := PeerInfo{
		Node:   node,
		peerID: peerID,
		maddrs: maddrs,
		enr:    node.String(),
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
	// TODO: this should probably be p.enr but a change here needs to
	// be coordinated with changes to our analysis scripts, so I'm keeping it as
	// it is.
	return string(p.peerID)
}

type CrawlDriverConfig struct {
	Version        string
	TrackNeighbors bool
	DialTimeout    time.Duration
	BootstrapPeers []*enode.Node
	AddrDialType   config.AddrType
	AddrTrackType  config.AddrType
	KeepENR        bool
	MeterProvider  metric.MeterProvider
	TracerProvider trace.TracerProvider
	LogErrors      bool
	UDPBufferSize  int
	UDPRespTimeout time.Duration
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:  cfg.DialTimeout,
		AddrDialType: cfg.AddrDialType,
		KeepENR:      cfg.KeepENR,
		LogErrors:    cfg.LogErrors,
	}
}

func (cfg *CrawlDriverConfig) WriterConfig() *core.CrawlWriterConfig {
	return &core.CrawlWriterConfig{
		AddrTrackType: cfg.AddrTrackType,
	}
}

type CrawlDriver struct {
	cfg          *CrawlDriverConfig
	dbc          db.Client
	hosts        []host.Host
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	peerstore    *enode.DB
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	// create a libp2p host per CPU core to distribute load
	hosts := make([]host.Host, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		h, err := newLibp2pHost(cfg.Version)
		if err != nil {
			return nil, fmt.Errorf("new libp2p host: %w", err)
		}
		hosts = append(hosts, h)
	}

	tasksChan := make(chan PeerInfo, len(cfg.BootstrapPeers))
	for _, node := range cfg.BootstrapPeers {
		pi, err := NewPeerInfo(node)
		if err != nil {
			return nil, fmt.Errorf("new peer info from enr: %w", err)
		}
		tasksChan <- pi
	}
	close(tasksChan)

	peerstore, err := enode.OpenDB("") // in memory db
	if err != nil {
		return nil, fmt.Errorf("open in-memory peerstore: %w", err)
	}

	// set the discovery response timeout
	discover.RespTimeoutV5 = cfg.UDPRespTimeout

	return &CrawlDriver{
		cfg:       cfg,
		dbc:       dbc,
		hosts:     hosts,
		dbCrawl:   crawl,
		tasksChan: tasksChan,
		peerstore: peerstore,
		crawler:   make([]*Crawler, 0),
	}, nil
}

// NewWorker is called multiple times but only log the configured buffer sizes once
var logOnce sync.Once

func (d *CrawlDriver) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	// If I'm not using the below elliptic curve, some Ethereum clients will reject communication
	priv, err := ecdsa.GenerateKey(ethcrypto.S256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("new ethereum ecdsa key: %w", err)
	}

	laddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 0,
	}

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return nil, fmt.Errorf("listen on udp4 port: %w", err)
	}

	if err = conn.SetReadBuffer(d.cfg.UDPBufferSize); err != nil {
		log.Warnln("Failed to set read buffer size on UDP listener", err)
	}

	rcvbuf, sndbuf, err := utils.GetUDPBufferSize(conn)
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

	ethNode := enode.NewLocalNode(d.peerstore, priv)

	// I'm not really sure if the below is strictly necessary.
	udpAddr := conn.LocalAddr().(*net.UDPAddr)
	if udpAddr.IP.IsUnspecified() {
		ethNode.SetFallbackIP(net.ParseIP("127.0.0.1"))
	} else {
		ethNode.SetFallbackIP(udpAddr.IP)
	}
	ethNode.SetFallbackUDP(udpAddr.Port)

	cfg := discover.Config{
		PrivateKey:              priv,
		ValidSchemes:            enode.ValidSchemes,
		NoFindnodeLivenessCheck: true,
		RefreshInterval:         100 * time.Hour, // turn off
	}
	listener, err := discover.ListenV5(conn, ethNode, cfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv5: %w", err)
	}

	// evenly assign a libp2p hosts to crawler workers
	h := d.hosts[d.crawlerCount%len(d.hosts)]

	c := &Crawler{
		id:       fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:      d.cfg.CrawlerConfig(),
		host:     h.(*libp2pconfig.ClosableBasicHost).BasicHost,
		listener: listener,
		done:     make(chan struct{}),
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

	for _, h := range d.hosts {
		if err := h.Close(); err != nil {
			log.WithError(err).WithField("localID", h.ID().String()).Warnln("Failed closing libp2p host")
		}
	}
}

func newLibp2pHost(version string) (host.Host, error) {
	// Configure the resource manager to not limit anything
	var noSubnetLimit []rcmgr.ConnLimitPerSubnet
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)

	v4PrefixLimits := []rcmgr.NetworkPrefixLimit{
		{
			Network:   netip.MustParsePrefix("0.0.0.0/0"),
			ConnCount: math.MaxInt, // Unlimited
		},
	}

	v6PrefixLimits := []rcmgr.NetworkPrefixLimit{
		{
			Network:   netip.MustParsePrefix("::1/0"),
			ConnCount: math.MaxInt, // Unlimited
		},
	}

	rm, err := rcmgr.NewResourceManager(limiter, rcmgr.WithLimitPerSubnet(noSubnetLimit, noSubnetLimit), rcmgr.WithNetworkPrefixLimit(v4PrefixLimits, v6PrefixLimits))
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

	// Don't use a connection manager that could potentially
	// prune any connections. We _theoretically_ clean up after
	//	// ourselves.
	cm := connmgr.NullConnMgr{}

	ecdsaKey, err := ecdsa.GenerateKey(ethcrypto.S256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate secp256k1 key: %w", err)
	}

	privBytes := elliptic.Marshal(ethcrypto.S256(), ecdsaKey.X, ecdsaKey.Y)
	secpKey := (*crypto.Secp256k1PrivateKey)(secp256k1v4.PrivKeyFromBytes(privBytes))

	// Initialize a single libp2p node that's shared between all crawlers.
	// Context: https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#network-fundamentals
	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.ResourceManager(rm),
		libp2p.Identity(secpKey),
		libp2p.Security(noise.ID, noise.New),
		libp2p.UserAgent("nebula/"+version),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Muxer(mplex.ID, mplex.DefaultTransport),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		libp2p.DisableMetrics(),
		libp2p.ConnectionManager(cm),
		libp2p.EnableRelay(), // enable the relay transport
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	// According to Diva, these are required protocols. Some of them are just
	// assumed to be required. We just read from the stream indefinitely to
	// gain time for the identify exchange to finish. We just pretend to support
	// these protocols and keep the stream busy until we have gathered all the
	// information we were interested in. This includes the agend version and
	// all supported protocols.
	h.SetStreamHandler("/eth2/beacon_chain/req/ping/1/ssz_snappy", func(s network.Stream) { io.ReadAll(s) })
	h.SetStreamHandler("/eth2/beacon_chain/req/status/1/ssz_snappy", func(s network.Stream) { io.ReadAll(s) })
	h.SetStreamHandler("/eth2/beacon_chain/req/metadata/1/ssz_snappy", func(s network.Stream) { io.ReadAll(s) })
	h.SetStreamHandler("/eth2/beacon_chain/req/metadata/2/ssz_snappy", func(s network.Stream) { io.ReadAll(s) })
	h.SetStreamHandler("/eth2/beacon_chain/req/goodbye/1/ssz_snappy", func(s network.Stream) { io.ReadAll(s) })
	h.SetStreamHandler("/meshsub/1.1.0", func(s network.Stream) { io.ReadAll(s) }) // https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/p2p-interface.md#the-gossip-domain-gossipsub

	log.WithField("peerID", h.ID().String()).Infoln("Started libp2p host")

	return h, nil
}

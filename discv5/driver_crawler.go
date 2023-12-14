package discv5

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	secp256k1v4 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
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
	"github.com/dennis-tra/nebula-crawler/discvx"
	"github.com/dennis-tra/nebula-crawler/utils"
)

type PeerInfo struct {
	*enode.Node
	peerID peer.ID
	maddrs []ma.Multiaddr
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

	pi := PeerInfo{
		Node:   node,
		peerID: peerID,
		maddrs: maddrs,
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

	discv5Cfg := discvx.Config{
		PrivateKey:   priv,
		ValidSchemes: enode.ValidSchemes,
	}

	listener, err := discvx.ListenV5(conn, ethNode, discv5Cfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv5: %w", err)
	}

	// evenly assign a libp2p hosts to crawler workers
	h := d.hosts[d.crawlerCount%len(d.hosts)]

	c := &Crawler{
		id:       fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:      d.cfg.CrawlerConfig(),
		host:     h.(*basichost.BasicHost),
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
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
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

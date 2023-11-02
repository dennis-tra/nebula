package discv5

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"net"
	"time"

	secp256k1v4 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/eth"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/dennis-tra/nebula-crawler/pkg/utils"
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
	return p // TODO: merge
}

type CrawlDriverConfig struct {
	Version           string
	TrackNeighbors    bool
	DialTimeout       time.Duration
	BootstrapPeerStrs []string
	AddrDialType      config.AddrType
	AddrTrackType     config.AddrType
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:  cfg.DialTimeout,
		AddrDialType: cfg.AddrDialType,
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
	host         host.Host
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	peerstore    *enode.DB
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlDriver)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *CrawlDriverConfig) (*CrawlDriver, error) {
	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, fmt.Errorf("new resource manager: %w", err)
	}

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
		libp2p.UserAgent("nebula/"+cfg.Version),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Muxer(mplex.ID, mplex.DefaultTransport),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
	)
	if err != nil {
		return nil, fmt.Errorf("new libp2p host: %w", err)
	}

	log.WithField("peerID", h.ID().String()).Infoln("Started libp2p host")

	nodesMap := map[enode.ID]*enode.Node{}
	for _, enrs := range cfg.BootstrapPeerStrs {
		n, err := enode.Parse(enode.ValidSchemes, enrs)
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap enr: %w", err)
		}
		nodesMap[n.ID()] = n
	}

	tasksChan := make(chan PeerInfo, len(nodesMap))
	for _, node := range nodesMap {
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
		host:      h,
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

	discv5Cfg := eth.Config{
		PrivateKey:   priv,
		ValidSchemes: enode.ValidSchemes,
	}

	listener, err := eth.ListenV5(conn, ethNode, discv5Cfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv5: %w", err)
	}

	c := &Crawler{
		id:       fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:      d.cfg.CrawlerConfig(),
		host:     d.host.(*basichost.BasicHost),
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

	if err := d.host.Close(); err != nil {
		log.WithError(err).Warnln("Failed closing libp2p host")
	}
}

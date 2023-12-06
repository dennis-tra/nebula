package discv4

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/dennis-tra/nebula-crawler/devp2p"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
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
	Version           string
	TrackNeighbors    bool
	DialTimeout       time.Duration
	BootstrapPeerStrs []string
	AddrDialType      config.AddrType
	AddrTrackType     config.AddrType
	MeterProvider     metric.MeterProvider
	TracerProvider    trace.TracerProvider
	LogErrors         bool
}

func (cfg *CrawlDriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		DialTimeout:  cfg.DialTimeout,
		AddrDialType: cfg.AddrDialType,
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
	clients      []*devp2p.Client
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
	clients := make([]*devp2p.Client, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		// If I'm not using the below elliptic curve, some Ethereum clients will reject communication
		priv, err := ecdsa.GenerateKey(ethcrypto.S256(), crand.Reader)
		if err != nil {
			return nil, fmt.Errorf("new ethereum ecdsa key: %w", err)
		}

		c := devp2p.NewClient(priv, nil)
		if err != nil {
			return nil, fmt.Errorf("new devp2p host: %w", err)
		}
		clients = append(clients, c)
	}

	nodesMap := map[enode.ID]*enode.Node{}
	for _, url := range cfg.BootstrapPeerStrs {
		n, err := enode.ParseV4(url)
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap enode URL %s: %w", url, err)
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
		clients:   clients,
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

	discvxCfg := discvx.Config{
		PrivateKey: priv,
	}

	listener, err := discvx.ListenV4(conn, ethNode, discvxCfg)
	if err != nil {
		return nil, fmt.Errorf("listen discv4: %w", err)
	}

	// evenly assign a libp2p hosts to crawler workers
	client := d.clients[d.crawlerCount%len(d.clients)]

	c := &Crawler{
		id:       fmt.Sprintf("crawler-%02d", d.crawlerCount),
		cfg:      d.cfg.CrawlerConfig(),
		client:   client,
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
}

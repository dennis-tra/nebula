package discv5

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/dennis-tra/nebula-crawler/pkg/core"
	"github.com/dennis-tra/nebula-crawler/pkg/db"
	"github.com/dennis-tra/nebula-crawler/pkg/eth"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

type PeerInfo struct {
	*enode.Node
	peerID peer.ID
	maddrs []ma.Multiaddr
}

var _ core.PeerInfo = (*PeerInfo)(nil)

func NewPeerInfo(node *enode.Node) (PeerInfo, error) {
	pubkey := node.Pubkey()
	if pubkey == nil {
		return PeerInfo{}, fmt.Errorf("no public key")
	}

	ppk, err := crypto.UnmarshalSecp256k1PublicKey(elliptic.Marshal(pubkey.Curve, pubkey.X, pubkey.Y))
	if err != nil {
		return PeerInfo{}, err
	}

	peerID, err := peer.IDFromPublicKey(ppk)
	if err != nil {
		return PeerInfo{}, fmt.Errorf("peer ID from public key: %w", err)
	}

	ipScheme := "ip4"
	if node.IP().To16() != nil {
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

	return PeerInfo{
		Node:   node,
		peerID: peerID,
		maddrs: maddrs,
	}, nil
}

func (p PeerInfo) ID() peer.ID {
	return peer.ID(p.Node.ID().String())
	return p.peerID
}

func (p PeerInfo) Addrs() []ma.Multiaddr {
	return p.maddrs
}

type DriverConfig struct {
	TrackNeighbors    bool
	BootstrapPeerStrs []string
}

func (cfg *DriverConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{}
}

type CrawlStack struct {
	cfg          *DriverConfig
	dbc          db.Client
	dbCrawl      *models.Crawl
	tasksChan    chan PeerInfo
	peerstore    *enode.DB
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Driver[PeerInfo, core.CrawlResult[PeerInfo]] = (*CrawlStack)(nil)

func NewCrawlDriver(dbc db.Client, crawl *models.Crawl, cfg *DriverConfig) (*CrawlStack, error) {
	peerstore, err := enode.OpenDB("") // in memory db
	if err != nil {
		return nil, fmt.Errorf("open in-memory peerstore: %w", err)
	}

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

	return &CrawlStack{
		cfg:       cfg,
		dbc:       dbc,
		dbCrawl:   crawl,
		peerstore: peerstore,
		crawler:   make([]*Crawler, 0),
	}, nil
}

func (s *CrawlStack) NewWorker() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
	// If I'm not using the below elliptic curve, some Ethereum clients will reject communication
	priv, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("new ethereum ecdsa key: %w", err)
	}

	ethNode := enode.NewLocalNode(s.peerstore, priv)

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
		id:       fmt.Sprintf("crawler-%02d", s.crawlerCount),
		cfg:      s.cfg.CrawlerConfig(),
		listener: listener,
		done:     make(chan struct{}),
	}

	s.crawlerCount += 1

	s.crawler = append(s.crawler, c)

	log.WithFields(log.Fields{
		"addr": conn.LocalAddr().String(),
	}).Debugln("Started crawler worker", c.id)

	return c, nil
}

func (s *CrawlStack) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewCrawlWriter[PeerInfo](fmt.Sprintf("writer-%02d", s.writerCount), s.dbc, s.dbCrawl.ID)
	s.writerCount += 1
	return w, nil
}

func (s *CrawlStack) Tasks() <-chan PeerInfo {
	return s.tasksChan
}

func (s *CrawlStack) Close() {
	for _, c := range s.crawler {
		c.listener.Close()
	}
}

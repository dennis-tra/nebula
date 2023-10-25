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

	maddrs := []ma.Multiaddr{}
	if node.UDP() != 0 {
		maddrStr := fmt.Sprintf("/ip4/%s/udp/%d", node.IP(), node.UDP())
		maddr, err := ma.NewMultiaddr(maddrStr)
		if err != nil {
			return PeerInfo{}, fmt.Errorf("parse multiaddress %s: %w", maddrStr, err)
		}
		maddrs = append(maddrs, maddr)
	}

	if node.TCP() != 0 {
		maddrStr := fmt.Sprintf("/ip4/%s/tcp/%d", node.IP(), node.TCP())
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

type StackConfig struct {
	TrackNeighbors    bool
	BootstrapPeerStrs []string
}

func (cfg *StackConfig) CrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{}
}

type Stack struct {
	cfg          *StackConfig
	dbc          db.Client
	dbCrawl      *models.Crawl
	peerstore    *enode.DB
	crawlerCount int
	writerCount  int
	crawler      []*Crawler
}

var _ core.Stack[PeerInfo] = (*Stack)(nil)

func NewStack(dbc db.Client, crawl *models.Crawl, cfg *StackConfig) (*Stack, error) {
	peerstore, err := enode.OpenDB("") // in memory db
	if err != nil {
		return nil, fmt.Errorf("open in-memory peerstore: %w", err)
	}

	return &Stack{
		cfg:       cfg,
		dbc:       dbc,
		dbCrawl:   crawl,
		peerstore: peerstore,
		crawler:   make([]*Crawler, 0),
	}, nil
}

func (s *Stack) NewCrawler() (core.Worker[PeerInfo, core.CrawlResult[PeerInfo]], error) {
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

func (s *Stack) NewWriter() (core.Worker[core.CrawlResult[PeerInfo], core.WriteResult], error) {
	w := core.NewWriter[PeerInfo](fmt.Sprintf("writer-%02d", s.writerCount), s.dbc, s.dbCrawl.ID)
	s.writerCount += 1
	return w, nil
}

func (s *Stack) BootstrapPeers() ([]PeerInfo, error) {
	nodesMap := map[enode.ID]*enode.Node{}
	for _, enrs := range s.cfg.BootstrapPeerStrs {
		n, err := enode.Parse(enode.ValidSchemes, enrs)
		if err != nil {
			return nil, fmt.Errorf("parse bootstrap enr: %w", err)
		}
		nodesMap[n.ID()] = n
	}

	pis := make([]PeerInfo, 0, len(s.cfg.BootstrapPeerStrs))
	for _, node := range nodesMap {
		pi, err := NewPeerInfo(node)
		if err != nil {
			return nil, fmt.Errorf("new peer info from enr: %w", err)
		}
		pis = append(pis, pi)
	}

	return pis, nil
}

func (s *Stack) OnPeerCrawled(cr core.CrawlResult[PeerInfo], err error) {
}

func (s *Stack) OnClose() {
	for _, c := range s.crawler {
		c.listener.Close()
	}
}

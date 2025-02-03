package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	pgmodels "github.com/dennis-tra/nebula-crawler/db/models/pg"
)

type JSONClient struct {
	out string

	prefix string

	visitsFile       *os.File
	neighborsFile    *os.File
	visitEncoder     *json.Encoder
	neighborsEncoder *json.Encoder

	peerMapLk   sync.RWMutex
	peerMap     map[peer.ID]int
	peerCounter int

	crawlMu sync.Mutex
	crawl   *pgmodels.Crawl
}

func NewJSONClient(out string) (*JSONClient, error) {
	log.WithField("out", out).Infoln("Initializing JSON client")

	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, fmt.Errorf("make json out directory: %w", err)
	}
	prefix := path.Join(out, time.Now().Format("2006-01-02T15:04"))

	vf, err := os.Create(prefix + "_visits.ndjson")
	if err != nil {
		return nil, fmt.Errorf("create visits file: %w", err)
	}

	nf, err := os.Create(prefix + "_neighbors.ndjson")
	if err != nil {
		return nil, fmt.Errorf("create neighbors file: %w", err)
	}

	client := &JSONClient{
		out:              out,
		prefix:           prefix,
		visitsFile:       vf,
		neighborsFile:    nf,
		visitEncoder:     json.NewEncoder(vf),
		neighborsEncoder: json.NewEncoder(nf),
		peerMap:          map[peer.ID]int{},
	}

	return client, nil
}

func (c *JSONClient) InitCrawl(ctx context.Context, version string) (err error) {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	defer func() {
		if err != nil {
			c.crawl = nil
		}
	}()

	if c.crawl != nil {
		return fmt.Errorf("crawl already initialized")
	}

	now := time.Now()

	c.crawl = &pgmodels.Crawl{
		State:     pgmodels.CrawlStateStarted,
		StartedAt: now,
		Version:   version,
		UpdatedAt: now,
		CreatedAt: now,
	}

	data, err := json.Marshal(c.crawl)
	if err != nil {
		return fmt.Errorf("marshal crawl json: %w", err)
	}

	if err = os.WriteFile(c.prefix+"_crawl.json", data, 0o644); err != nil {
		return fmt.Errorf("write crawl json: %w", err)
	}

	return nil
}

func (c *JSONClient) SealCrawl(ctx context.Context, args *SealCrawlArgs) (err error) {
	c.crawlMu.Lock()
	defer c.crawlMu.Unlock()

	original := *c.crawl
	defer func() {
		// roll back in case of an error
		if err != nil {
			c.crawl = &original
		}
	}()

	now := time.Now()
	c.crawl.UpdatedAt = now
	c.crawl.CrawledPeers = null.IntFrom(args.Crawled)
	c.crawl.DialablePeers = null.IntFrom(args.Dialable)
	c.crawl.UndialablePeers = null.IntFrom(args.Undialable)
	c.crawl.RemainingPeers = null.IntFrom(args.Remaining)
	c.crawl.State = string(args.State)
	c.crawl.FinishedAt = null.TimeFrom(now)

	data, err := json.Marshal(c.crawl)
	if err != nil {
		return fmt.Errorf("marshal crawl json: %w", err)
	}

	if err = os.WriteFile(c.prefix+"_crawl.json", data, 0o644); err != nil {
		return fmt.Errorf("write crawl json: %w", err)
	}

	return nil
}

func (c *JSONClient) QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error) {
	return []peer.AddrInfo{}, nil
}

func (c *JSONClient) InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error {
	data, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("marshal properties json: %w", err)
	}

	if err = os.WriteFile(c.prefix+"_crawl_properties.json", data, 0o644); err != nil {
		return fmt.Errorf("write properties json: %w", err)
	}

	return nil
}

type JSONVisit struct {
	PeerID          peer.ID
	Maddrs          []ma.Multiaddr
	Protocols       []string
	AgentVersion    string
	ConnectDuration string
	CrawlDuration   string
	VisitStartedAt  time.Time
	VisitEndedAt    time.Time
	ConnectErrorStr string
	CrawlErrorStr   string
	Properties      null.JSON
}

func (c *JSONClient) InsertVisit(ctx context.Context, args *VisitArgs) error {
	data := JSONVisit{
		PeerID:          args.PeerID,
		Maddrs:          args.Maddrs,
		Protocols:       args.Protocols,
		AgentVersion:    args.AgentVersion,
		ConnectDuration: args.ConnectDuration.String(),
		CrawlDuration:   args.CrawlDuration.String(),
		VisitStartedAt:  args.VisitStartedAt,
		VisitEndedAt:    args.VisitEndedAt,
		ConnectErrorStr: args.ConnectErrorStr,
		CrawlErrorStr:   args.CrawlErrorStr,
		Properties:      args.Properties,
	}

	if err := c.visitEncoder.Encode(data); err != nil {
		return fmt.Errorf("encoding visit: %w", err)
	}

	return nil
}

type JSONNeighbors struct {
	PeerID      peer.ID
	NeighborIDs []peer.ID
	ErrorBits   string
}

func (c *JSONClient) InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error {
	data := JSONNeighbors{
		PeerID:      peerID,
		NeighborIDs: neighbors,
		ErrorBits:   fmt.Sprintf("%016b", errorBits),
	}

	if err := c.neighborsEncoder.Encode(data); err != nil {
		return fmt.Errorf("encoding visit: %w", err)
	}

	return nil
}

func (c *JSONClient) SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error) {
	return []peer.AddrInfo{}, nil
}

func (c *JSONClient) Close() error {
	err1 := c.visitsFile.Close()
	err2 := c.neighborsFile.Close()
	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed closing JSON files: %w", fmt.Errorf("%s: %w (neighbors)", err1, err2))
	} else if err1 != nil {
		return fmt.Errorf("failed closing visits file: %w", err1)
	} else if err2 != nil {
		return fmt.Errorf("failed closing neighbors files: %w", err2)
	}

	return nil
}

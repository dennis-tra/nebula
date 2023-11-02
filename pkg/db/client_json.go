package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/friendsofgo/errors"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
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
}

var _ Client = (*JSONClient)(nil)

// InitJSONClient .
func InitJSONClient(out string) (Client, error) {
	log.WithField("out", out).Infoln("Initializing JSON client")

	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, fmt.Errorf("make json out directory: %w", err)
	}
	prefix := path.Join(out, time.Now().Format("2006-01-02T15:04"))

	vf, err := os.Create(prefix + "_visits.json")
	if err != nil {
		return nil, fmt.Errorf("create visits file: %w", err)
	}

	nf, err := os.Create(prefix + "_neighbors.json")
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

func (c *JSONClient) InitCrawl(ctx context.Context) (*models.Crawl, error) {
	crawl := &models.Crawl{
		State:     models.CrawlStateStarted,
		StartedAt: time.Now(),
	}

	data, err := json.Marshal(crawl)
	if err != nil {
		return nil, fmt.Errorf("marshal crawl json: %w", err)
	}

	if err = os.WriteFile(c.prefix+"_crawl.json", data, 0o644); err != nil {
		return nil, fmt.Errorf("write crawl json: %w", err)
	}

	return crawl, nil
}

func (c *JSONClient) UpdateCrawl(ctx context.Context, crawl *models.Crawl) error {
	data, err := json.Marshal(crawl)
	if err != nil {
		return fmt.Errorf("marshal crawl json: %w", err)
	}

	if err = os.WriteFile(c.prefix+"_crawl.json", data, 0o644); err != nil {
		return fmt.Errorf("write crawl json: %w", err)
	}

	return nil
}

func (c *JSONClient) PersistCrawlProperties(ctx context.Context, crawl *models.Crawl, properties map[string]map[string]int) error {
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

func (c *JSONClient) PersistCrawlVisit(ctx context.Context, crawlID int, peerID peer.ID, maddrs []ma.Multiaddr, protocols []string, agentVersion string, connectDuration time.Duration, crawlDuration time.Duration, visitStartedAt time.Time, visitEndedAt time.Time, connectErrorStr string, crawlErrorStr string, properties null.JSON) (*InsertVisitResult, error) {
	data := JSONVisit{
		PeerID:          peerID,
		Maddrs:          maddrs,
		Protocols:       protocols,
		AgentVersion:    agentVersion,
		ConnectDuration: connectDuration.String(),
		CrawlDuration:   crawlDuration.String(),
		VisitStartedAt:  visitStartedAt,
		VisitEndedAt:    visitEndedAt,
		ConnectErrorStr: connectErrorStr,
		CrawlErrorStr:   crawlErrorStr,
		Properties:      properties,
	}

	if err := c.visitEncoder.Encode(data); err != nil {
		return nil, fmt.Errorf("encoding visit: %w", err)
	}

	return &InsertVisitResult{PID: peerID}, nil
}

type JSONNeighbors struct {
	PeerID      peer.ID
	NeighborIDs []peer.ID
	ErrorBits   string
}

func (c *JSONClient) PersistNeighbors(ctx context.Context, crawl *models.Crawl, dbPeerID *int, peerID peer.ID, errorBits uint16, dbNeighborsIDs []int, neighbors []peer.ID) error {
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

func (n *JSONClient) Close() error {
	err1 := n.visitsFile.Close()
	err2 := n.neighborsFile.Close()
	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed closing JSON files: %w", errors.Wrap(err1, err2.Error()+" (neighbors)"))
	} else if err1 != nil {
		return fmt.Errorf("failed closing visits file: %w", err1)
	} else if err2 != nil {
		return fmt.Errorf("failed closing neighbors files: %w", err2)
	}

	return nil
}

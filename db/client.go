package db

import (
	"context"
	"io"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/volatiletech/null/v8"

	"github.com/libp2p/go-libp2p/core/peer"
)

type CrawlState string

const (
	CrawlStateStarted   CrawlState = "started"
	CrawlStateSucceeded CrawlState = "succeeded"
	CrawlStateCancelled CrawlState = "cancelled"
	CrawlStateFailed    CrawlState = "failed"
)

type VisitType string

const (
	VisitTypeDial  VisitType = "dial"
	VisitTypeCrawl VisitType = "crawl"
)

type SealCrawlArgs struct {
	Crawled    int
	Dialable   int
	Undialable int
	Remaining  int
	State      CrawlState
}

type VisitArgs struct {
	PeerID          peer.ID
	Maddrs          []ma.Multiaddr
	Protocols       []string
	AgentVersion    string
	DialDuration    *time.Duration
	ConnectDuration *time.Duration
	CrawlDuration   *time.Duration
	VisitStartedAt  time.Time
	VisitEndedAt    time.Time
	ConnectErrorStr string
	CrawlErrorStr   string
	VisitType       VisitType
	Neighbors       []peer.ID
	ErrorBits       uint16
	Properties      null.JSON
}

type Client interface {
	io.Closer
	// InitCrawl initializes a new crawl instance in the database.
	// The clients are responsible for tracking the crawl's ID and associate
	// later database queries with it. This is necessary because different
	// database engines have different types of IDs. ClickHouse uses string
	// IDs and Postgres uses integers. Making the [Client] interface generic
	// on that ID would complicate the code a lot, so we require Clients to
	// keep state. This is added complexy traded for code clarity. It's a trade-
	// off and IMO this is less bad.
	InitCrawl(ctx context.Context, version string) error

	// SealCrawl marks the crawl (that the Client tracks internally) as done.
	SealCrawl(ctx context.Context, args *SealCrawlArgs) error

	// QueryBootstrapPeers fetches peers from the database that can be used
	// for bootstrapping into the network. The result will contain from zero up
	// to limit entries.
	QueryBootstrapPeers(ctx context.Context, limit int) ([]peer.AddrInfo, error)

	// InsertVisit .
	InsertVisit(ctx context.Context, args *VisitArgs) error

	// InsertCrawlProperties .
	InsertCrawlProperties(ctx context.Context, properties map[string]map[string]int) error

	// InsertNeighbors .
	InsertNeighbors(ctx context.Context, peerID peer.ID, neighbors []peer.ID, errorBits uint16) error

	// SelectPeersToProbe .
	SelectPeersToProbe(ctx context.Context) ([]peer.AddrInfo, error)
}

var (
	_ Client = (*PostgresClient)(nil)
	_ Client = (*NoopClient)(nil)
	_ Client = (*JSONClient)(nil)
	_ Client = (*ClickHouseClient)(nil)
)

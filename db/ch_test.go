package db

import (
	"fmt"
	"testing"
	"time"

	pt "github.com/libp2p/go-libp2p/core/test"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/test"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/suite"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/net/context"

	"github.com/dennis-tra/nebula-crawler/utils"
)

type ClickHouseTestSuite struct {
	suite.Suite

	client *ClickHouseClient
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ClickHouseTestSuite) SetupSuite() {
	cfg := ClickHouseClientConfig{
		DatabaseHost:     "localhost",
		DatabasePort:     9001,
		DatabaseName:     "nebula_test",
		DatabasePassword: "password_test",
		DatabaseUser:     "nebula_test",
		DatabaseSSL:      false,
		ApplyMigrations:  true,
		BatchSize:        10_000,
		BatchTimeout:     time.Second,
		NetworkID:        "test_network",
		PersistNeighbors: true,
		MeterProvider:    mnoop.NewMeterProvider(),
		TracerProvider:   tnoop.NewTracerProvider(),
	}

	ctx := suite.timeoutCtx()
	client, err := NewClickHouseClient(ctx, &cfg)
	suite.Require().NoError(err)

	suite.client = client

	suite.clearDatabase(ctx)
}

func (suite *ClickHouseTestSuite) TearDownSuite() {
	suite.Assert().NoError(suite.client.Close())
}

func (suite *ClickHouseTestSuite) SetupTest() {
	// nothing
}

func (suite *ClickHouseTestSuite) TearDownTest() {
	ctx := suite.timeoutCtx()
	suite.clearDatabase(ctx)
	suite.resetClient()
}

func (suite *ClickHouseTestSuite) clearDatabase(ctx context.Context) {
	for _, table := range []string{"crawls", "visits", "neighbors"} {
		err := suite.client.conn.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE TRUE", table))
		suite.Assert().NoError(err)
	}
}

func (suite *ClickHouseTestSuite) resetClient() {
	suite.client.crawlMu.Lock()
	defer suite.client.crawlMu.Unlock()

	suite.client.crawl = nil
	suite.client.cfg.NetworkID = "test_network"
}

func (suite *ClickHouseTestSuite) timeoutCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	suite.T().Cleanup(cancel)
	return ctx
}

func (suite *ClickHouseTestSuite) TestInitCrawl() {
	ctx := suite.timeoutCtx()

	err := suite.client.InitCrawl(ctx, "v1")
	suite.Require().NoError(err)

	storedCrawl, err := suite.client.selectCrawl(ctx, suite.client.crawl.ID)
	suite.Require().NoError(err)

	suite.Assert().Equal(suite.client.crawl.ID, storedCrawl.ID)
	suite.Assert().Equal(suite.client.crawl.State, storedCrawl.State)
	suite.Assert().Equal(suite.client.crawl.FinishedAt, storedCrawl.FinishedAt)
	suite.Assert().Equal(suite.client.crawl.UpdatedAt.Truncate(time.Millisecond).UTC(), storedCrawl.UpdatedAt.UTC())
	suite.Assert().Equal(suite.client.crawl.CreatedAt.Truncate(time.Millisecond).UTC(), storedCrawl.CreatedAt.UTC())
	suite.Assert().Equal(suite.client.crawl.CrawledPeers, storedCrawl.CrawledPeers)
	suite.Assert().Equal(suite.client.crawl.DialablePeers, storedCrawl.DialablePeers)
	suite.Assert().Equal(suite.client.crawl.UndialablePeers, storedCrawl.UndialablePeers)
	suite.Assert().Equal(suite.client.crawl.RemainingPeers, storedCrawl.RemainingPeers)
	suite.Assert().Equal(suite.client.crawl.Version, storedCrawl.Version)
	suite.Assert().Equal(suite.client.crawl.NetworkID, storedCrawl.NetworkID)

	// crawl already exists
	err = suite.client.InitCrawl(ctx, "v1")
	suite.Assert().Error(err)
}

func (suite *ClickHouseTestSuite) TestInitCrawl_differentNetworkFail() {
	ctx := suite.timeoutCtx()

	err := suite.client.InitCrawl(ctx, "v1")
	suite.Require().NoError(err)

	suite.resetClient()

	suite.client.cfg.NetworkID = "different_network"

	err = suite.client.InitCrawl(ctx, "v1")
	suite.Assert().Error(err)
}

func (suite *ClickHouseTestSuite) TestSealCrawl_notInitialized() {
	ctx := suite.timeoutCtx()

	args := &SealCrawlArgs{
		Crawled:    1,
		Dialable:   2,
		Undialable: 3,
		Remaining:  4,
		State:      CrawlStateSucceeded,
	}

	err := suite.client.SealCrawl(ctx, args)
	suite.Assert().Error(err)
}

func (suite *ClickHouseTestSuite) TestSealCrawl_happyPath() {
	ctx := suite.timeoutCtx()

	err := suite.client.InitCrawl(ctx, "v1")
	suite.Require().NoError(err)

	args := &SealCrawlArgs{
		Crawled:    1,
		Dialable:   2,
		Undialable: 3,
		Remaining:  4,
		State:      CrawlStateSucceeded,
	}

	err = suite.client.SealCrawl(ctx, args)
	suite.Assert().NoError(err)

	storedCrawl, err := suite.client.selectCrawl(ctx, suite.client.crawl.ID)
	suite.Require().NoError(err)

	// tests that the internal state was updated
	suite.Assert().Equal(suite.client.crawl.ID, storedCrawl.ID)
	suite.Assert().Equal(suite.client.crawl.State, storedCrawl.State)
	suite.Assert().Equal(suite.client.crawl.FinishedAt.Truncate(time.Millisecond).UTC(), storedCrawl.FinishedAt.UTC())
	suite.Assert().Equal(suite.client.crawl.UpdatedAt.Truncate(time.Millisecond).UTC(), storedCrawl.UpdatedAt.UTC())
	suite.Assert().Equal(suite.client.crawl.CreatedAt.Truncate(time.Millisecond).UTC(), storedCrawl.CreatedAt.UTC())
	suite.Assert().Equal(suite.client.crawl.CrawledPeers, storedCrawl.CrawledPeers)
	suite.Assert().Equal(suite.client.crawl.DialablePeers, storedCrawl.DialablePeers)
	suite.Assert().Equal(suite.client.crawl.UndialablePeers, storedCrawl.UndialablePeers)
	suite.Assert().Equal(suite.client.crawl.RemainingPeers, storedCrawl.RemainingPeers)
	suite.Assert().Equal(suite.client.crawl.Version, storedCrawl.Version)
	suite.Assert().Equal(suite.client.crawl.NetworkID, storedCrawl.NetworkID)

	suite.Assert().EqualValues(args.State, storedCrawl.State)
	suite.Assert().EqualValues(args.Crawled, *storedCrawl.CrawledPeers)
	suite.Assert().EqualValues(args.Dialable, *storedCrawl.DialablePeers)
	suite.Assert().EqualValues(args.Undialable, *storedCrawl.UndialablePeers)
	suite.Assert().EqualValues(args.Remaining, *storedCrawl.RemainingPeers)
}

func (suite *ClickHouseTestSuite) TestSealCrawl_insertVisit() {
	ctx := suite.timeoutCtx()

	pid, err := pt.RandPeerID()
	suite.Require().NoError(err)

	neighbors := test.GeneratePeerIDs(100)

	start := time.Now().UTC()
	end := start.Add(time.Second)

	connDur := 100 * time.Millisecond
	crawlDur := 100 * time.Millisecond

	args := &VisitArgs{
		PeerID: pid,
		DialMaddrs: []multiaddr.Multiaddr{
			utils.MustMultiaddr(suite.T(), "/ip4/127.0.0.1/tcp/1234"),
		},
		Protocols:        []string{"/ipfs/1.0.0"},
		AgentVersion:     "my-agent",
		ConnectDuration:  connDur,
		CrawlDuration:    crawlDur,
		VisitStartedAt:   start,
		VisitEndedAt:     end,
		ConnectErrorStr:  "conn_err",
		CrawlErrorStr:    "crawl_err",
		VisitType:        "dial",
		Neighbors:        neighbors,
		NeighborPrefixes: make([]uint64, len(neighbors)),
		ErrorBits:        20,
	}

	err = suite.client.InitCrawl(ctx, "v1")
	suite.Require().NoError(err)

	err = suite.client.InsertVisit(ctx, args)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.client.Flush(ctx))
	time.Sleep(100 * time.Millisecond)

	storedVisit, err := suite.client.selectLatestVisit(ctx)
	suite.Require().NoError(err)

	suite.Assert().Equal(suite.client.crawl.ID, *storedVisit.CrawlID)
	suite.Assert().Equal(args.PeerID.String(), storedVisit.PeerID)
	suite.Assert().Equal(args.AgentVersion, *storedVisit.AgentVersion)
	suite.Assert().Equal(args.VisitStartedAt.Truncate(time.Millisecond), storedVisit.VisitStartedAt)
	suite.Assert().Equal(args.VisitEndedAt.Truncate(time.Millisecond), storedVisit.VisitEndedAt)
	suite.Assert().Equal(args.CrawlErrorStr, *storedVisit.CrawlError)

	storedNeighbors, err := suite.client.selectNeighbors(ctx, suite.client.crawl.ID)
	suite.Require().NoError(err)

	suite.Assert().Len(storedNeighbors, len(neighbors))
	for i := range neighbors {
		suite.Assert().Equal(args.ErrorBits, storedNeighbors[i].ErrorBits)
	}
}

func (suite *ClickHouseTestSuite) TestSealCrawl_queryBootstrapPeers() {
	ctx := suite.timeoutCtx()

	count := 100
	neighbors := test.GeneratePeerIDs(count)
	maddr := utils.MustMultiaddr(suite.T(), "/ip4/127.0.0.1/tcp/1234")

	for i, neighbor := range neighbors {
		var connectMaddr multiaddr.Multiaddr
		if i%2 == 0 {
			connectMaddr = maddr
		}
		args := &VisitArgs{
			PeerID:         neighbor,
			DialMaddrs:     []multiaddr.Multiaddr{maddr},
			ConnectMaddr:   connectMaddr,
			VisitStartedAt: time.Now().Add(-time.Minute).UTC(),
			VisitEndedAt:   time.Now().Add(-time.Minute).UTC(),
		}

		err := suite.client.InsertVisit(ctx, args)
		suite.Require().NoError(err)
	}

	suite.Require().NoError(suite.client.Flush(ctx))

	peers, err := suite.client.QueryBootstrapPeers(ctx, count)
	suite.Require().NoError(err)
	suite.Assert().Len(peers, count/2)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClickHouseTestSuite(t *testing.T) {
	suite.Run(t, new(ClickHouseTestSuite))
}

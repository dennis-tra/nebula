package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func clearDatabase(ctx context.Context, db *sql.DB) error {
	if _, err := models.Sessions().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.Peers().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.Crawls().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.AgentVersions().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.Protocols().DeleteAll(ctx, db); err != nil {
		return err
	}
	return nil
}

func setup(t *testing.T) (context.Context, *Client, func(t *testing.T)) {
	ctx := context.Background()

	c := config.DefaultConfig
	c.DatabaseName = "nebula_test"
	c.DatabaseUser = "nebula_test"
	c.DatabasePassword = "password_test"
	c.DatabasePort = 2345

	client, err := InitClient(ctx, &c)
	require.NoError(t, err)

	return ctx, client, func(t *testing.T) {
		err = clearDatabase(context.Background(), client.dbh)
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)
	}
}

func TestClient_InitCrawl(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	assert.NotZero(t, crawl.ID)
	assert.Equal(t, models.CrawlStateStarted, crawl.State)
	assert.NotZero(t, crawl.StartedAt)
	assert.False(t, crawl.CrawledPeers.Valid)
	assert.False(t, crawl.DialablePeers.Valid)
	assert.False(t, crawl.UndialablePeers.Valid)
	assert.False(t, crawl.FinishedAt.Valid)

	dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawl.ID)).One(ctx, client.dbh)
	require.NoError(t, err)

	assert.Equal(t, crawl.ID, dbCrawl.ID)
	assert.Equal(t, models.CrawlStateStarted, dbCrawl.State)
	assert.NotZero(t, dbCrawl.StartedAt)
	assert.False(t, dbCrawl.CrawledPeers.Valid)
	assert.False(t, dbCrawl.DialablePeers.Valid)
	assert.False(t, dbCrawl.UndialablePeers.Valid)
	assert.False(t, dbCrawl.FinishedAt.Valid)
}

func TestClient_GetOrCreateAgentVersion(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	id, err := client.GetOrCreateAgentVersionID(ctx, client.Handle(), "")
	assert.Error(t, err)
	assert.Equal(t, 0, id)

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	prevID := id

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	assert.Equal(t, prevID, id)

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-2")
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	assert.NotEqual(t, prevID, id)
}

func TestClient_PersistCrawlProperties(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	props := map[string]map[string]int{}
	props["agent_version"] = map[string]int{
		"agent-1": 1,
		"agent-2": 2,
	}
	props["protocol"] = map[string]int{
		"protocols-1": 1,
		"protocols-2": 2,
	}
	props["error"] = map[string]int{
		"unknown":    1,
		"io_timeout": 2,
	}

	err = client.PersistCrawlProperties(ctx, crawl, props)
	require.NoError(t, err)

	cps, err := models.CrawlProperties(models.CrawlPropertyWhere.CrawlID.EQ(crawl.ID)).All(ctx, client.dbh)
	require.NoError(t, err)

	for _, cp := range cps {
		if !cp.ProtocolID.IsZero() {
			protocol, err := models.Protocols(models.ProtocolWhere.ID.EQ(cp.ProtocolID.Int)).One(ctx, client.dbh)
			require.NoError(t, err)
			assert.Equal(t, cp.Count, props["protocol"][protocol.Protocol])
		} else if !cp.AgentVersionID.IsZero() {
			agentVersion, err := models.AgentVersions(models.AgentVersionWhere.ID.EQ(cp.AgentVersionID.Int)).One(ctx, client.dbh)
			require.NoError(t, err)
			assert.Equal(t, cp.Count, props["agent_version"][agentVersion.AgentVersion])
		} else if !cp.Error.IsZero() {
			assert.Equal(t, cp.Count, props["error"][cp.Error.String])
		}
	}
	assert.Equal(t, 6, len(cps))
}

func TestClient_PersistCrawlVisit(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)
	m, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/2345")

	err = client.PersistCrawlVisit(
		ctx,
		client.dbh,
		crawl.ID,
		"my-long-peer-id",
		[]ma.Multiaddr{m},
		[]string{"protocol-1", "protocol-2"},
		"agent-1",
		time.Second,
		time.Second,
		time.Now().Add(-time.Second),
		time.Now(),
		"",
		"",
	)
	require.NoError(t, err)
}

func TestClient_QueryBootstrapPeers(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	peers, err := client.QueryBootstrapPeers(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}

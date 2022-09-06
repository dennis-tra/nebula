package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func setupTest(t *testing.T) (context.Context, *Client, func()) {
	ctx := context.Background()

	dbh, err := sql.Open("postgres", "dbname=nebula_test user=nebula_test password=password sslmode=disable")
	require.NoError(t, err)

	client := &Client{
		dbh: dbh,
	}

	err = clearDatabase(ctx, dbh)
	require.NoError(t, err)

	return ctx, client, func() {
		err := clearDatabase(ctx, dbh)
		require.NoError(t, err)
	}
}

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

func TestClient_InitCrawl(t *testing.T) {
	ctx, client, teardown := setupTest(t)
	defer teardown()

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	assert.NotZero(t, crawl.ID)
	assert.Equal(t, models.CrawlStateStarted, crawl.State)
	assert.NotZero(t, crawl.StartedAt)
	assert.False(t, crawl.CrawledPeers.Valid)
	assert.False(t, crawl.DialablePeers.Valid)
	assert.False(t, crawl.UndialablePeers.Valid)
	assert.False(t, crawl.FinishedAt.Valid)

	dbCrawl, err := models.Crawls(qm.Where(models.CrawlColumns.ID+"=?", crawl.ID)).One(ctx, client.dbh)
	require.NoError(t, err)

	assert.Equal(t, crawl.ID, dbCrawl.ID)
	assert.Equal(t, models.CrawlStateStarted, dbCrawl.State)
	assert.NotZero(t, dbCrawl.StartedAt)
	assert.False(t, dbCrawl.CrawledPeers.Valid)
	assert.False(t, dbCrawl.DialablePeers.Valid)
	assert.False(t, dbCrawl.UndialablePeers.Valid)
	assert.False(t, dbCrawl.FinishedAt.Valid)
}

func TestClient_QueryPeers(t *testing.T) {
	ctx, client, teardown := setupTest(t)
	defer teardown()

	var pis []peer.AddrInfo
	multiHashes := []string{
		"Qmbjb4FuGtmG59fSQuZj5g561QrvpT5FYnkbJxGXcgGLgx",
		"QmXhDcdvZtUjPMpHu9XkYg8G7Sh6rJvGzaWHkn7FQbS37W",
	}

	for _, mh := range multiHashes {
		p := models.Peer{
			MultiHash: mh,
		}
		err := p.Insert(ctx, client.dbh, boil.Infer())
		require.NoError(t, err)

		id, err := peer.Decode(mh)
		require.NoError(t, err)

		pi := peer.AddrInfo{ID: id}
		pis = append(pis, pi)
	}

	// All known
	peers, err := client.QueryPeers(ctx, pis)
	require.NoError(t, err)
	assert.Len(t, peers, len(multiHashes))

	for i, p := range peers {
		pi := pis[i]
		assert.Equal(t, pi.ID.Pretty(), p.MultiHash)
	}

	// Subset
	peers, err = client.QueryPeers(ctx, pis[1:])
	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, pis[1].ID.Pretty(), peers[0].MultiHash)

	// middle unknown
	id, err := peer.Decode("QmeWW3hHSZwztzH78QmAFPw2r2os1B4QLuSWrVCraSnjWc")
	require.NoError(t, err)
	unknownPi := peer.AddrInfo{ID: id}
	peers, err = client.QueryPeers(ctx, []peer.AddrInfo{pis[0], unknownPi, pis[1]})
	require.NoError(t, err)
	assert.Len(t, peers, 2)

	// all unknown
	peers, err = client.QueryPeers(ctx, []peer.AddrInfo{unknownPi})
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}

func TestClient_QueryBootstrapPeers(t *testing.T) {
	ctx, client, teardown := setupTest(t)
	defer teardown()
	peers, err := client.QueryBootstrapPeers(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}

func TestClient_GetAllAgentVersions(t *testing.T) {
	ctx, client, teardown := setupTest(t)
	defer teardown()

	avs, err := client.GetAllAgentVersions(ctx)
	require.NoError(t, err)
	assert.Len(t, avs, 0)

	testAvStr1 := "test-agent-version-1"
	testAvStr2 := "test-agent-version-2"
	testAvStr3 := "test-agent-version-3"

	av1 := models.AgentVersion{AgentVersion: testAvStr1}
	require.NoError(t, av1.Insert(ctx, client.dbh, boil.Infer()))

	avs, err = client.GetAllAgentVersions(ctx)
	require.NoError(t, err)
	assert.Len(t, avs, 1)
	assert.Equal(t, avs[testAvStr1].AgentVersion, testAvStr1)
	assert.NotZero(t, avs[testAvStr1].ID)

	av2 := models.AgentVersion{AgentVersion: testAvStr2}
	require.NoError(t, av2.Insert(ctx, client.dbh, boil.Infer()))
	av3 := models.AgentVersion{AgentVersion: testAvStr3}
	require.NoError(t, av3.Insert(ctx, client.dbh, boil.Infer()))

	avs, err = client.GetAllAgentVersions(ctx)
	require.NoError(t, err)
	assert.Len(t, avs, 3)
	assert.Equal(t, avs[testAvStr1].AgentVersion, testAvStr1)
	assert.Equal(t, avs[testAvStr2].AgentVersion, testAvStr2)
	assert.Equal(t, avs[testAvStr3].AgentVersion, testAvStr3)
	assert.NotZero(t, avs[testAvStr1].ID)
	assert.NotZero(t, avs[testAvStr2].ID)
	assert.NotZero(t, avs[testAvStr3].ID)
}

func TestClient_GetProtocols(t *testing.T) {
	ctx, client, teardown := setupTest(t)
	defer teardown()

	protocols, err := client.GetAllProtocols(ctx)
	require.NoError(t, err)
	assert.Len(t, protocols, 0)

	testProtStr1 := "test-protocol-1"
	testProtStr2 := "test-protocol-2"
	testProtStr3 := "test-protocol-3"

	prot1 := models.Protocol{Protocol: testProtStr1}
	require.NoError(t, prot1.Insert(ctx, client.dbh, boil.Infer()))

	protocols, err = client.GetAllProtocols(ctx)
	require.NoError(t, err)
	assert.Len(t, protocols, 1)
	assert.Equal(t, protocols[testProtStr1].Protocol, testProtStr1)
	assert.NotZero(t, protocols[testProtStr1].ID)

	prot2 := models.Protocol{Protocol: testProtStr2}
	require.NoError(t, prot2.Insert(ctx, client.dbh, boil.Infer()))
	prot3 := models.Protocol{Protocol: testProtStr3}
	require.NoError(t, prot3.Insert(ctx, client.dbh, boil.Infer()))

	protocols, err = client.GetAllProtocols(ctx)
	require.NoError(t, err)
	assert.Len(t, protocols, 3)
	assert.Equal(t, protocols[testProtStr1].Protocol, testProtStr1)
	assert.Equal(t, protocols[testProtStr2].Protocol, testProtStr2)
	assert.Equal(t, protocols[testProtStr3].Protocol, testProtStr3)
	assert.NotZero(t, protocols[testProtStr1].ID)
	assert.NotZero(t, protocols[testProtStr2].ID)
	assert.NotZero(t, protocols[testProtStr3].ID)
}

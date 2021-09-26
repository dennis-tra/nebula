package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func setupTest(t *testing.T) (context.Context, *Client, func()) {
	ctx := context.Background()

	dbh, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
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

package db

import (
	"context"
	"database/sql"
	"github.com/dennis-tra/nebula-crawler/pkg/config"
	ma "github.com/multiformats/go-multiaddr"
	"testing"
	"time"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	id, err := client.GetOrCreateAgentVersion(ctx, client.Handle(), "")
	assert.Error(t, err)
	assert.Equal(t, 0, id)

	id, err = client.GetOrCreateAgentVersion(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	prevID := id

	id, err = client.GetOrCreateAgentVersion(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	assert.Equal(t, prevID, id)

	id, err = client.GetOrCreateAgentVersion(ctx, client.Handle(), "agent-2")
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
	)
	require.NoError(t, err)
}

//	func TestClient_QueryPeers(t *testing.T) {
//		ctx, client, teardown := setupTest(t)
//		defer teardown()
//
//		var pis []peer.AddrInfo
//		multiHashes := []string{
//			"Qmbjb4FuGtmG59fSQuZj5g561QrvpT5FYnkbJxGXcgGLgx",
//			"QmXhDcdvZtUjPMpHu9XkYg8G7Sh6rJvGzaWHkn7FQbS37W",
//		}
//
//		for _, mh := range multiHashes {
//			p := models.Peer{
//				MultiHash: mh,
//			}
//			err := p.Insert(ctx, client.dbh, boil.Infer())
//			require.NoError(t, err)
//
//			id, err := peer.Decode(mh)
//			require.NoError(t, err)
//
//			pi := peer.AddrInfo{ID: id}
//			pis = append(pis, pi)
//		}
//
//		// All known
//		peers, err := client.QueryPeers(ctx, pis)
//		require.NoError(t, err)
//		assert.Len(t, peers, len(multiHashes))
//
//		for i, p := range peers {
//			pi := pis[i]
//			assert.Equal(t, pi.ID.Pretty(), p.MultiHash)
//		}
//
//		// Subset
//		peers, err = client.QueryPeers(ctx, pis[1:])
//		require.NoError(t, err)
//		assert.Len(t, peers, 1)
//		assert.Equal(t, pis[1].ID.Pretty(), peers[0].MultiHash)
//
//		// middle unknown
//		id, err := peer.Decode("QmeWW3hHSZwztzH78QmAFPw2r2os1B4QLuSWrVCraSnjWc")
//		require.NoError(t, err)
//		unknownPi := peer.AddrInfo{ID: id}
//		peers, err = client.QueryPeers(ctx, []peer.AddrInfo{pis[0], unknownPi, pis[1]})
//		require.NoError(t, err)
//		assert.Len(t, peers, 2)
//
//		// all unknown
//		peers, err = client.QueryPeers(ctx, []peer.AddrInfo{unknownPi})
//		require.NoError(t, err)
//		assert.Len(t, peers, 0)
//	}
func TestClient_QueryBootstrapPeers(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	peers, err := client.QueryBootstrapPeers(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}

//func TestClient_GetAllAgentVersions(t *testing.T) {
//	ctx, client, teardown := setupTest(t)
//	defer teardown()
//
//	avs, err := client.GetAllAgentVersions(ctx)
//	require.NoError(t, err)
//	assert.Len(t, avs, 0)
//
//	testAvStr1 := "test-agent-version-1"
//	testAvStr2 := "test-agent-version-2"
//	testAvStr3 := "test-agent-version-3"
//
//	av1 := models.AgentVersion{AgentVersion: testAvStr1}
//	require.NoError(t, av1.Insert(ctx, client.dbh, boil.Infer()))
//
//	avs, err = client.GetAllAgentVersions(ctx)
//	require.NoError(t, err)
//	assert.Len(t, avs, 1)
//	assert.Equal(t, avs[testAvStr1].AgentVersion, testAvStr1)
//	assert.NotZero(t, avs[testAvStr1].ID)
//
//	av2 := models.AgentVersion{AgentVersion: testAvStr2}
//	require.NoError(t, av2.Insert(ctx, client.dbh, boil.Infer()))
//	av3 := models.AgentVersion{AgentVersion: testAvStr3}
//	require.NoError(t, av3.Insert(ctx, client.dbh, boil.Infer()))
//
//	avs, err = client.GetAllAgentVersions(ctx)
//	require.NoError(t, err)
//	assert.Len(t, avs, 3)
//	assert.Equal(t, avs[testAvStr1].AgentVersion, testAvStr1)
//	assert.Equal(t, avs[testAvStr2].AgentVersion, testAvStr2)
//	assert.Equal(t, avs[testAvStr3].AgentVersion, testAvStr3)
//	assert.NotZero(t, avs[testAvStr1].ID)
//	assert.NotZero(t, avs[testAvStr2].ID)
//	assert.NotZero(t, avs[testAvStr3].ID)
//}
//
//func TestClient_GetProtocols(t *testing.T) {
//	ctx, client, teardown := setupTest(t)
//	defer teardown()
//
//	protocols, err := client.GetAllProtocols(ctx)
//	require.NoError(t, err)
//	assert.Len(t, protocols, 0)
//
//	testProtStr1 := "test-protocol-1"
//	testProtStr2 := "test-protocol-2"
//	testProtStr3 := "test-protocol-3"
//
//	prot1 := models.Protocol{Protocol: testProtStr1}
//	require.NoError(t, prot1.Insert(ctx, client.dbh, boil.Infer()))
//
//	protocols, err = client.GetAllProtocols(ctx)
//	require.NoError(t, err)
//	assert.Len(t, protocols, 1)
//	assert.Equal(t, protocols[testProtStr1].Protocol, testProtStr1)
//	assert.NotZero(t, protocols[testProtStr1].ID)
//
//	prot2 := models.Protocol{Protocol: testProtStr2}
//	require.NoError(t, prot2.Insert(ctx, client.dbh, boil.Infer()))
//	prot3 := models.Protocol{Protocol: testProtStr3}
//	require.NoError(t, prot3.Insert(ctx, client.dbh, boil.Infer()))
//
//	protocols, err = client.GetAllProtocols(ctx)
//	require.NoError(t, err)
//	assert.Len(t, protocols, 3)
//	assert.Equal(t, protocols[testProtStr1].Protocol, testProtStr1)
//	assert.Equal(t, protocols[testProtStr2].Protocol, testProtStr2)
//	assert.Equal(t, protocols[testProtStr3].Protocol, testProtStr3)
//	assert.NotZero(t, protocols[testProtStr1].ID)
//	assert.NotZero(t, protocols[testProtStr2].ID)
//	assert.NotZero(t, protocols[testProtStr3].ID)
//}

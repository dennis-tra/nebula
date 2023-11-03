package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	lp2ptest "github.com/libp2p/go-libp2p/core/test"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/config"
	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func clearDatabase(ctx context.Context, db *sql.DB) error {
	if _, err := models.Sessions().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.PeerLogs().DeleteAll(ctx, db); err != nil {
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
	if _, err := models.ProtocolsSets().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.Visits().DeleteAll(ctx, db); err != nil {
		return err
	}
	if _, err := models.CrawlProperties().DeleteAll(ctx, db); err != nil {
		return err
	}
	return nil
}

func setup(t *testing.T) (context.Context, *DBClient, func(t *testing.T)) {
	ctx := context.Background()

	c := config.Database{
		DatabaseHost:           "localhost",
		DatabasePort:           2345,
		DatabaseName:           "nebula_test",
		DatabasePassword:       "password",
		DatabaseUser:           "nebula_test",
		DatabaseSSLMode:        "disable",
		AgentVersionsCacheSize: 100,
		ProtocolsCacheSize:     100,
		ProtocolsSetCacheSize:  100,
	}

	client, err := InitDBClient(ctx, &c)
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
	assert.Nil(t, id)
	client.agentVersions.Purge()

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	prevID := id
	client.agentVersions.Purge()

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-1")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	assert.Equal(t, *prevID, *id)
	client.agentVersions.Purge()

	id, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-2")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	assert.NotEqual(t, *prevID, *id)
}

func TestClient_GetOrCreateProtocol(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	id, err := client.GetOrCreateProtocol(ctx, client.Handle(), "")
	assert.Error(t, err)
	assert.Nil(t, id)
	client.protocols.Purge()

	id, err = client.GetOrCreateProtocol(ctx, client.Handle(), "protocol-1")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	prevID := id
	client.protocols.Purge()

	id, err = client.GetOrCreateProtocol(ctx, client.Handle(), "protocol-1")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	assert.Equal(t, *prevID, *id)
	client.protocols.Purge()

	id, err = client.GetOrCreateProtocol(ctx, client.Handle(), "protocol-2")
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	assert.NotEqual(t, *prevID, *id)
}

func TestClient_GetOrCreateProtocolsSetID(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	id, err := client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{})
	assert.Error(t, err)
	assert.Nil(t, id)
	client.protocolsSets.Purge()

	id, err = client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{"protocol-1", "protocol-2"})
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	prevID := id
	client.protocolsSets.Purge()

	id, err = client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{"protocol-2", "protocol-1"})
	assert.NoError(t, err)
	assert.Equal(t, *prevID, *id)
	client.protocolsSets.Purge()

	id, err = client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{"protocol-2", "protocol-1", "protocol-3"})
	assert.NoError(t, err)
	assert.Greater(t, *id, 0)
	assert.NotEqual(t, *prevID, *id)
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

func TestClient_QueryBootstrapPeers(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	peers, err := client.QueryBootstrapPeers(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}

func TestClient_PersistCrawlVisit(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	peerID, err := lp2ptest.RandPeerID()
	require.NoError(t, err)

	ma1, err := multiaddr.NewMultiaddr("/ip4/100.0.0.1/tcp/2000")
	require.NoError(t, err)

	ma2, err := multiaddr.NewMultiaddr("/ip4/100.0.0.2/udp/3000")
	require.NoError(t, err)

	protocols := []string{"protocol-1", "protocol-2"}
	agentVersion := "agent-1"

	visitStart := time.Now().Add(-time.Second)
	visitEnd := time.Now()
	ivr, err := client.PersistCrawlVisit(
		ctx,
		crawl.ID,
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		protocols,
		agentVersion,
		time.Second,
		time.Second,
		visitStart,
		visitEnd,
		models.NetErrorIoTimeout,
		"",
		null.JSONFrom(marshalProperties(t, "is_exposed", true)),
	)
	require.NoError(t, err)

	assert.Nil(t, ivr.SessionID)
	assert.NotNil(t, ivr.PeerID)
	assert.NotNil(t, ivr.VisitID)
}

func TestClient_SessionScenario_1(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	peerID, err := lp2ptest.RandPeerID()
	require.NoError(t, err)

	ma1, err := multiaddr.NewMultiaddr("/ip4/100.0.0.1/tcp/2000")
	require.NoError(t, err)

	ma2, err := multiaddr.NewMultiaddr("/ip4/100.0.0.2/udp/3000")
	require.NoError(t, err)

	protocols := []string{"protocol-1", "protocol-2"}
	agentVersion := "agent-1"

	visitStart := time.Now().Add(-time.Second)
	visitEnd := time.Now()

	ivr, err := client.PersistCrawlVisit(
		ctx,
		crawl.ID,
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		protocols,
		agentVersion,
		time.Second,
		time.Second,
		visitStart,
		visitEnd,
		"",
		"",
		null.JSONFrom(marshalProperties(t, "is_exposed", true)),
	)
	require.NoError(t, err)

	dbPeer := fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	assert.Equal(t, dbPeer.R.AgentVersion.AgentVersion, agentVersion)
	assert.Equal(t, dbPeer.MultiHash, peerID.String())
	assert.Len(t, dbPeer.R.MultiAddresses, 2)
	assert.True(t, unmarshalProperties(t, dbPeer.Properties.JSON)["is_exposed"].(bool))

	for _, ma := range dbPeer.R.MultiAddresses {
		assert.True(t, ma.Maddr == ma1.String() || ma.Maddr == ma2.String())
	}
	session := dbPeer.R.SessionsOpen
	sessionID1 := session.ID
	assert.Equal(t, session.PeerID, dbPeer.ID)
	assert.Equal(t, session.SuccessfulVisitsCount, 1)
	assert.Equal(t, session.FailedVisitsCount, int16(0))
	assert.Equal(t, session.State, models.SessionStateOpen)
	assert.InDelta(t, session.FirstSuccessfulVisit.UnixNano(), visitStart.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, session.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.True(t, session.FirstFailedVisit.IsZero())
	assert.True(t, session.FinishReason.IsZero())
	assert.True(t, session.LastFailedVisit.IsZero())

	visitStart = time.Now().Add(-time.Second)
	visitEnd = time.Now()
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		"",
	)
	require.NoError(t, err)
	dbPeer = fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	assert.True(t, dbPeer.Properties.Valid)

	session = dbPeer.R.SessionsOpen
	assert.Equal(t, session.PeerID, dbPeer.ID)
	assert.Equal(t, session.SuccessfulVisitsCount, 2)
	assert.Equal(t, session.FailedVisitsCount, int16(0))
	assert.Equal(t, session.State, models.SessionStateOpen)
	assert.NotEqual(t, session.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.InDelta(t, session.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.True(t, session.FirstFailedVisit.IsZero())
	assert.True(t, session.FinishReason.IsZero())
	assert.True(t, session.LastFailedVisit.IsZero())

	visitStart = time.Now().Add(-time.Second)
	visitEnd = time.Now()
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		models.NetErrorConnectionRefused,
	)
	require.NoError(t, err)
	dbPeer = fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	assert.Nil(t, dbPeer.R.SessionsOpen)

	s, err := models.Sessions(models.SessionWhere.ID.EQ(*ivr.SessionID)).One(ctx, client.Handle())
	require.NoError(t, err)

	assert.Equal(t, s.PeerID, dbPeer.ID)
	assert.Equal(t, s.SuccessfulVisitsCount, 2)
	assert.Equal(t, s.FailedVisitsCount, int16(1))
	assert.Equal(t, s.State, models.SessionStateClosed)
	assert.NotEqual(t, s.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.InDelta(t, s.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, s.FirstFailedVisit.Time.UnixNano(), visitStart.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, s.LastFailedVisit.Time.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.Equal(t, s.FinishReason.String, models.NetErrorConnectionRefused)

	crawl, err = client.InitCrawl(ctx)
	require.NoError(t, err)
	visitStart = time.Now().Add(-time.Second)
	visitEnd = time.Now()

	ivr, err = client.PersistCrawlVisit(
		ctx,
		crawl.ID,
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		[]string{},
		"",
		time.Second,
		time.Second,
		visitStart,
		visitEnd,
		"",
		"",
		null.JSONFrom(marshalProperties(t, "is_exposed", true)),
	)
	require.NoError(t, err)

	visitStart = time.Now().Add(-time.Second)
	visitEnd = time.Now()
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		models.NetErrorNegotiateSecurityProtocol,
	)
	dbPeer = fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	assert.Equal(t, dbPeer.R.AgentVersion.AgentVersion, agentVersion)
	assert.Equal(t, dbPeer.MultiHash, peerID.String())
	assert.Len(t, dbPeer.R.MultiAddresses, 2)

	newSession, err := models.Sessions(models.SessionWhere.ID.EQ(*ivr.SessionID)).One(ctx, client.Handle())
	require.NoError(t, err)

	sessionID2 := newSession.ID
	require.NotEqual(t, sessionID1, sessionID2)

	assert.Equal(t, newSession.PeerID, dbPeer.ID)
	assert.Equal(t, newSession.SuccessfulVisitsCount, 1)
	assert.Equal(t, newSession.FailedVisitsCount, int16(1))
	assert.Equal(t, newSession.State, models.SessionStateClosed)
	assert.NotEqual(t, newSession.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.InDelta(t, newSession.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, newSession.FirstFailedVisit.Time.UnixNano(), visitStart.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, newSession.LastFailedVisit.Time.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.Equal(t, newSession.FinishReason.String, models.NetErrorNegotiateSecurityProtocol)

	err = s.Reload(ctx, client.Handle())
	require.NoError(t, err)

	// untouched:
	assert.Equal(t, s.PeerID, dbPeer.ID)
	assert.Equal(t, s.SuccessfulVisitsCount, 2)
	assert.Equal(t, s.FailedVisitsCount, int16(1))
	assert.Equal(t, s.State, models.SessionStateClosed)
	assert.NotEqual(t, s.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.NotEqual(t, s.LastVisitedAt.UnixMicro(), visitStart.UnixMicro())
	assert.NotEqual(t, s.FirstFailedVisit.Time.UnixMicro(), visitStart.UnixMicro())
	assert.NotEqual(t, s.LastFailedVisit.Time.UnixMicro(), visitEnd.UnixMicro())
	assert.Equal(t, s.FinishReason.String, models.NetErrorConnectionRefused)
}

func TestClient_SessionScenario_2(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	crawl, err := client.InitCrawl(ctx)
	require.NoError(t, err)

	peerID, err := lp2ptest.RandPeerID()
	require.NoError(t, err)

	ma1, err := multiaddr.NewMultiaddr("/ip4/100.0.0.3/tcp/2000")
	require.NoError(t, err)

	ma2, err := multiaddr.NewMultiaddr("/ip4/100.0.0.4/udp/3000")
	require.NoError(t, err)

	protocols := []string{"protocol-2", "protocol-3"}
	agentVersion := "agent-1"

	visitStart := time.Now()
	visitEnd := time.Now().Add(time.Second)
	ivr, err := client.PersistCrawlVisit(
		ctx,
		crawl.ID,
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		protocols,
		agentVersion,
		time.Second,
		time.Second,
		visitStart,
		visitEnd,
		"",
		"",
		null.JSONFrom(marshalProperties(t, "is_exposed", false)),
	)
	require.NoError(t, err)
	visitStart = time.Now().Add(100 * time.Hour)
	visitEnd = time.Now().Add(100 * time.Hour).Add(time.Second)
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		"",
	)

	visitStart = time.Now().Add(101 * time.Hour).Add(time.Second)
	visitEnd = time.Now().Add(101 * time.Hour)
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		models.NetErrorIoTimeout,
	)
	require.NoError(t, err)
	dbPeer := fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	session := dbPeer.R.SessionsOpen
	assert.Equal(t, session.PeerID, dbPeer.ID)
	assert.Equal(t, session.SuccessfulVisitsCount, 2)
	assert.Equal(t, session.FailedVisitsCount, int16(1))
	assert.Equal(t, session.RecoveredCount, 0)
	assert.Equal(t, session.State, models.SessionStatePending)
	assert.NotEqual(t, session.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.InDelta(t, session.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, session.FirstFailedVisit.Time.UnixNano(), visitStart.UnixNano(), float64(time.Microsecond))
	assert.Equal(t, session.FinishReason.String, models.NetErrorIoTimeout)
	assert.InDelta(t, session.LastFailedVisit.Time.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))

	visitStart = time.Now().Add(-time.Second)
	visitEnd = time.Now()
	ivr, err = client.PersistDialVisit(
		peerID,
		[]multiaddr.Multiaddr{ma1, ma2},
		time.Second,
		visitStart,
		visitEnd,
		"",
	)
	require.NoError(t, err)
	dbPeer = fetchPeer(t, ctx, client.Handle(), *ivr.PeerID)

	session = dbPeer.R.SessionsOpen
	assert.Equal(t, session.PeerID, dbPeer.ID)
	assert.Equal(t, session.SuccessfulVisitsCount, 3)
	assert.Equal(t, session.FailedVisitsCount, int16(0))
	assert.Equal(t, session.State, models.SessionStateOpen)
	assert.Equal(t, session.RecoveredCount, 1)
	assert.NotEqual(t, session.FirstSuccessfulVisit.UnixMicro(), visitStart.UnixMicro())
	assert.InDelta(t, session.LastSuccessfulVisit.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.InDelta(t, session.LastVisitedAt.UnixNano(), visitEnd.UnixNano(), float64(time.Microsecond))
	assert.True(t, session.FirstFailedVisit.IsZero())
	assert.True(t, session.FinishReason.IsZero())
	assert.True(t, session.LastFailedVisit.IsZero())

	count, err := models.Sessions().Count(ctx, client.Handle())
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestClient_UpsertPeer(t *testing.T) {
	ctx, client, teardown := setup(t)
	defer teardown(t)

	dbAgentVersionID, err := client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-1")
	require.NoError(t, err)

	dbProtocolsSetID, err := client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{"protocol-1", "protocol-2"})
	require.NoError(t, err)

	peerID, err := lp2ptest.RandPeerID()
	require.NoError(t, err)

	dbPeerID, err := client.UpsertPeer(peerID.String(), null.IntFromPtr(nil), null.IntFromPtr(nil), null.JSONFromPtr(nil))
	require.NoError(t, err)
	assert.NotZero(t, dbPeerID)

	dbPeer := fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.True(t, dbPeer.AgentVersionID.IsZero())
	assert.True(t, dbPeer.ProtocolsSetID.IsZero())
	assert.True(t, dbPeer.Properties.IsZero())

	dbPeerID, err = client.UpsertPeer(peerID.String(), null.IntFromPtr(dbAgentVersionID), null.IntFromPtr(nil), null.JSONFromPtr(nil))
	require.NoError(t, err)

	dbPeer = fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.Equal(t, dbPeer.AgentVersionID.Int, *dbAgentVersionID)
	assert.True(t, dbPeer.ProtocolsSetID.IsZero())
	assert.True(t, dbPeer.Properties.IsZero())

	dbPeerID, err = client.UpsertPeer(peerID.String(), null.IntFromPtr(nil), null.IntFromPtr(dbProtocolsSetID), null.JSONFromPtr(nil))
	require.NoError(t, err)

	dbPeer = fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.Equal(t, dbPeer.AgentVersionID.Int, *dbAgentVersionID)
	assert.Equal(t, dbPeer.ProtocolsSetID.Int, *dbProtocolsSetID)
	assert.True(t, dbPeer.Properties.IsZero())

	dbPeerID, err = client.UpsertPeer(peerID.String(), null.IntFromPtr(nil), null.IntFromPtr(dbProtocolsSetID), null.JSONFrom(marshalProperties(t, "is_exposed", false)))
	require.NoError(t, err)

	dbPeer = fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.Equal(t, dbPeer.AgentVersionID.Int, *dbAgentVersionID)
	assert.Equal(t, dbPeer.ProtocolsSetID.Int, *dbProtocolsSetID)
	assert.False(t, unmarshalProperties(t, dbPeer.Properties.JSON)["is_exposed"].(bool))

	dbPeerID, err = client.UpsertPeer(peerID.String(), null.IntFromPtr(nil), null.IntFromPtr(nil), null.JSONFromPtr(nil))
	require.NoError(t, err)

	dbPeer = fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.Equal(t, dbPeer.AgentVersionID.Int, *dbAgentVersionID)
	assert.Equal(t, dbPeer.ProtocolsSetID.Int, *dbProtocolsSetID)
	assert.False(t, unmarshalProperties(t, dbPeer.Properties.JSON)["is_exposed"].(bool))

	dbAgentVersionID, err = client.GetOrCreateAgentVersionID(ctx, client.Handle(), "agent-2")
	require.NoError(t, err)

	dbProtocolsSetID, err = client.GetOrCreateProtocolsSetID(ctx, client.Handle(), []string{"protocol-3", "protocol-2"})
	require.NoError(t, err)

	dbPeerID, err = client.UpsertPeer(peerID.String(), null.IntFromPtr(dbAgentVersionID), null.IntFromPtr(dbProtocolsSetID), null.JSONFrom(marshalProperties(t, "is_exposed", true)))
	require.NoError(t, err)

	dbPeer = fetchPeer(t, ctx, client.Handle(), dbPeerID)
	assert.Equal(t, dbPeer.AgentVersionID.Int, *dbAgentVersionID)
	assert.Equal(t, dbPeer.ProtocolsSetID.Int, *dbProtocolsSetID)
	assert.True(t, unmarshalProperties(t, dbPeer.Properties.JSON)["is_exposed"].(bool))
}

func fetchPeer(t *testing.T, ctx context.Context, exec boil.ContextExecutor, dbPeerID int) *models.Peer {
	dbPeer, err := models.Peers(
		models.PeerWhere.ID.EQ(dbPeerID),
		qm.Load(models.PeerRels.AgentVersion),
		qm.Load(models.PeerRels.MultiAddresses),
		qm.Load(models.PeerRels.ProtocolsSet),
		qm.Load(models.PeerRels.SessionsOpen),
	).One(ctx, exec)
	require.NoError(t, err)
	return dbPeer
}

func marshalProperties(t testing.TB, args ...any) []byte {
	if len(args)%2 != 0 {
		t.Fatal("args must be multiple of 2")
	}

	properties := map[string]any{}
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		require.True(t, ok)
		value := args[i+1]

		properties[key] = value
	}

	data, err := json.Marshal(properties)
	require.NoError(t, err)
	return data
}

func unmarshalProperties(t testing.TB, data []byte) map[string]any {
	properties := map[string]any{}
	err := json.Unmarshal(data, &properties)
	require.NoError(t, err)
	return properties
}

package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	_ "github.com/lib/pq"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func setup(t *testing.T, ctx context.Context) (*sql.DB, string, func(*testing.T)) {
	db, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
	require.NoError(t, err)

	peerID := "some-id"

	_, err = models.Sessions(qm.Where("peer_id = ?", peerID)).DeleteAll(ctx, db)
	require.NoError(t, err)

	err = UpsertPeer(ctx, db, peerID, []ma.Multiaddr{})
	require.NoError(t, err)

	return db, peerID, func(t *testing.T) {
		_, err = models.Sessions(qm.Where("peer_id = ?", peerID)).DeleteAll(ctx, db)
		require.NoError(t, err)
	}
}

func TestUpsertSession(t *testing.T) {
	ctx := context.Background()

	db, peerID, teardown := setup(t, ctx)
	defer teardown(t)

	err := UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	session, err := models.Sessions(qm.Where("peer_id = ?", peerID)).One(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, session.PeerID, peerID)
	assert.Equal(t, session.FirstSuccessfulDial, session.LastSuccessfulDial)
	assert.Equal(t, session.NextDialAttempt.Time, session.LastSuccessfulDial.Add(30*time.Second))
	assert.Equal(t, session.FirstFailedDial.Local(), time.Unix(0, 0))
	assert.Equal(t, session.Finished, false)
	assert.Equal(t, session.SuccessfulDials, 1)
	assert.Equal(t, session.CreatedAt, session.UpdatedAt)
	assert.Equal(t, session.CreatedAt, session.FirstSuccessfulDial)

	err = UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	session, err = models.Sessions(qm.Where("peer_id = ?", peerID)).One(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, session.PeerID, peerID)
	assert.Greaterf(t, session.LastSuccessfulDial.UnixNano(), session.FirstSuccessfulDial.UnixNano(), "")
	assert.Equal(t, session.NextDialAttempt.Time, session.LastSuccessfulDial.Add(30*time.Second))
	assert.Equal(t, session.FirstFailedDial.Local(), time.Unix(0, 0))
	assert.Equal(t, session.Finished, false)
	assert.Equal(t, session.SuccessfulDials, 2)
	assert.Greaterf(t, session.UpdatedAt.UnixNano(), session.CreatedAt.UnixNano(), "")
	assert.Equal(t, session.CreatedAt, session.FirstSuccessfulDial)

	session.LastSuccessfulDial = session.LastSuccessfulDial.Add(time.Minute)
	_, err = session.Update(ctx, db, boil.Infer())
	require.NoError(t, err)

	session, err = models.Sessions(qm.Where("peer_id = ?", peerID)).One(ctx, db)
	require.NoError(t, err)

	// TODO:
	// assert.Equal(t, session.PeerID, peerID)
	// assert.Equal(t, session.NextDialAttempt.Time.String(), session.LastSuccessfulDial.Add(time.Duration(float64(session.LastSuccessfulDial.Sub(session.FirstSuccessfulDial))*1.5)).String())
	// assert.Equal(t, session.FirstFailedDial.Local(), time.Unix(0, 0))
	// assert.Equal(t, session.Finished, false)
	// assert.Equal(t, session.SuccessfulDials, 2)
	// assert.Greaterf(t, session.UpdatedAt.UnixNano(), session.CreatedAt.UnixNano(), "")
	// assert.Equal(t, session.CreatedAt, session.FirstSuccessfulDial)

	err = UpsertSessionError(db, peerID)
	require.NoError(t, err)

	err = session.Reload(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, true, session.Finished)
	assert.Equal(t, 2, session.SuccessfulDials)
	assert.False(t, session.NextDialAttempt.Valid)
}

func TestNextDialSession(t *testing.T) {
	ctx := context.Background()
	db, peerID, teardown := setup(t, ctx)
	defer teardown(t)

	err := UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	err = UpsertSessionErrorTS(db, peerID, time.Now())
	require.NoError(t, err)

	err = UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	err = UpsertSessionError(db, peerID)
	require.NoError(t, err)

	sessions, err := models.Sessions(qm.Where("peer_id = ?", peerID)).All(ctx, db)
	require.NoError(t, err)
	assert.Lenf(t, sessions, 2, "")
}

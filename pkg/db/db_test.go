package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/nebula-crawler/pkg/models"
)

func TestUpsertSession(t *testing.T) {
	db, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
	require.NoError(t, err)

	peerID := "some-id"
	ctx := context.Background()

	_, err = models.Sessions(qm.Where("peer_id = ?", peerID)).DeleteAll(ctx, db)
	require.NoError(t, err)

	err = UpsertPeer(ctx, db, peerID, []ma.Multiaddr{})
	require.NoError(t, err)

	err = UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	session, err := models.Sessions(qm.Where("peer_id = ?", peerID)).One(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, session.Finished, false)
	assert.Equal(t, session.SuccessfulDials, 1)

	err = UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	err = session.Reload(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, session.Finished, false)
	assert.Equal(t, session.SuccessfulDials, 2)

	err = UpsertSessionError(db, peerID)
	require.NoError(t, err)

	err = session.Reload(ctx, db)
	require.NoError(t, err)

	assert.Equal(t, session.Finished, true)
	assert.Equal(t, session.SuccessfulDials, 2)
}

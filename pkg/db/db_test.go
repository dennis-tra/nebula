package db

import (
	"context"
	"database/sql"
	"fmt"
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

func setup(t *testing.T) (context.Context, *sql.DB, string, func(*testing.T)) {
	ctx := context.Background()

	db, err := sql.Open("postgres", "dbname=nebula user=nebula password=password sslmode=disable")
	require.NoError(t, err)

	peerID := fmt.Sprintf("some-id-%d", time.Now().Nanosecond())

	_, err = models.Sessions().DeleteAll(ctx, db)
	require.NoError(t, err)

	err = UpsertPeer(ctx, db, peerID, []ma.Multiaddr{})
	require.NoError(t, err)

	return ctx, db, peerID, func(t *testing.T) {
		_, err = models.Sessions().DeleteAll(ctx, db)
		require.NoError(t, err)
	}
}

func TestUpsertSessionSuccess_insertsRowIfNotExists(t *testing.T) {
	ctx, db, peerID, teardown := setup(t)
	defer teardown(t)

	err := UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	s, err := FetchSession(ctx, db, peerID)
	require.NoError(t, err)

	assert.Equal(t, s.PeerID, peerID)
	assert.Equal(t, s.FirstSuccessfulDial, s.LastSuccessfulDial)
	assert.Equal(t, s.FirstFailedDial.Local(), time.Unix(0, 0))
	assert.Equal(t, s.NextDialAttempt.Time, s.LastSuccessfulDial.Add(MinInterval))
	assert.Equal(t, s.SuccessfulDials, 1)
	assert.Equal(t, s.Finished, false)
	assert.Equal(t, s.CreatedAt, s.UpdatedAt)
	assert.Equal(t, s.CreatedAt, s.FirstSuccessfulDial)
	assert.False(t, s.FailureReason.Valid)
	assert.False(t, s.MinDuration.Valid)
	assert.False(t, s.MaxDuration.Valid)
}

func TestUpsertSessionSuccess_upsertsRowIfExists(t *testing.T) {
	ctx, db, peerID, teardown := setup(t)
	defer teardown(t)

	sleepDur := 100 * time.Millisecond
	tolerance := 5 * time.Millisecond

	// Create a session object
	err := UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	// Wait a second and measure time until the session was updated (for the assertion below)
	start := time.Now()
	time.Sleep(sleepDur)

	// Upsert the same session object
	err = UpsertSessionSuccess(db, peerID)
	dur := time.Since(start)
	require.NoError(t, err)

	// Fetch object from database
	s, err := FetchSession(ctx, db, peerID)
	require.NoError(t, err)

	// Assert things
	assert.Equal(t, s.PeerID, peerID)
	assert.NotEqual(t, s.FirstSuccessfulDial, s.LastSuccessfulDial)
	assert.InDelta(t, s.FirstSuccessfulDial.Add(dur).UnixNano(), s.LastSuccessfulDial.UnixNano(), float64(tolerance), "The last successful dial should roughly be sleepDur larger")
	assert.Equal(t, s.FirstFailedDial.Local(), time.Unix(0, 0))
	assert.Equal(t, s.SuccessfulDials, 2)
	assert.Equal(t, s.Finished, false)
	assert.InDelta(t, s.CreatedAt.Add(dur).UnixNano(), s.UpdatedAt.UnixNano(), float64(tolerance), "The last successful dial should roughly be sleepDur larger")
	assert.False(t, s.FailureReason.Valid)
	assert.False(t, s.MinDuration.Valid)
	assert.False(t, s.MaxDuration.Valid)
	assert.Equal(t, s.CreatedAt, s.FirstSuccessfulDial)
	assert.Equal(t, s.NextDialAttempt.Time, s.LastSuccessfulDial.Add(MinInterval))
}

func createSession(t *testing.T, ctx context.Context, db *sql.DB, peerID string) *models.Session {
	// Create a session object
	err := UpsertSessionSuccess(db, peerID)
	require.NoError(t, err)

	// Fetch object from database
	s, err := FetchSession(ctx, db, peerID)
	require.NoError(t, err)
	return s
}

func TestUpsertSessionSuccess_nextDialCalculation(t *testing.T) {
	tests := []struct {
		name     string
		uptime   time.Duration
		nextDial time.Duration
	}{
		{
			name:     "lower limit",
			uptime:   10 * time.Second,
			nextDial: MinInterval,
		},
		{
			name:     "intermediate",
			uptime:   5 * time.Minute,
			nextDial: time.Duration(float64(5*time.Minute) * 1.1),
		},
		{
			name:     "upper limit",
			uptime:   time.Hour,
			nextDial: MaxInterval,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, db, peerID, teardown := setup(t)
			defer teardown(t)

			uptime := tt.uptime

			// Create a session object
			s := createSession(t, ctx, db, peerID)

			// Move its creation time 5 minutes in the past
			now := time.Now()
			newTime := now.Add(-uptime)
			// s.CreatedAt = newTime // has no effect due to sqlboiler
			s.UpdatedAt = newTime
			s.FirstSuccessfulDial = newTime
			s.LastSuccessfulDial = newTime

			_, err := s.Update(boil.SkipTimestamps(ctx), db, boil.Infer())
			require.NoError(t, err)

			// Upsert the same session object
			upsert := time.Now()
			err = UpsertSessionSuccess(db, peerID)
			require.NoError(t, err)

			// Fetch session from database again
			s, err = FetchSession(ctx, db, peerID)
			require.NoError(t, err)

			// Assert things
			assert.Equal(t, s.PeerID, peerID)
			assert.Equal(t, s.FirstSuccessfulDial.UnixNano(), newTime.UnixNano())
			assert.Greater(t, s.LastSuccessfulDial.UnixNano(), newTime.UnixNano())
			assert.Equal(t, upsert.Unix(), s.LastSuccessfulDial.Unix())

			assert.Equal(t, time.Now().Add(tt.nextDial).Unix(), s.NextDialAttempt.Time.Unix())
		})
	}
}

func TestUpsertSession(t *testing.T) {
	ctx, db, peerID, teardown := setup(t)
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
	ctx, db, peerID, teardown := setup(t)
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

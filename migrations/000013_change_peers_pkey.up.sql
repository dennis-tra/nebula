-- The former id column should hold the peer_id string and
-- a new serial database ID should be used. Doing it this
-- way will prevent the database from storing the peer_id
-- of the form QmbLHAnMo... multiple times. Thus saving
-- disk space.

-- Begin the transaction
BEGIN;

-- Create new peer ID column holding the actual QmbLHAnMo... multi hash string
ALTER TABLE peers
    RENAME COLUMN id TO multi_hash;

-- Create new id column holding the internal serial ID of that peer
ALTER TABLE peers
    ADD COLUMN id SERIAL NOT NULL;

-- Put all relevant tables aside
ALTER TABLE sessions
    RENAME TO sessions_old;

-- drop all relevant constraints
ALTER TABLE sessions_old
    DROP CONSTRAINT fk_session_peer;

ALTER TABLE sessions_old
    DROP CONSTRAINT uq_peer_id_first_failed_dial;

-- Make the new serial id column the new primary key on the peers table.
ALTER TABLE peers
    DROP CONSTRAINT peers_pkey;
ALTER TABLE peers
    ADD PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_peers_multi_hash ON peers (multi_hash);

-- Create new sessions table with improved column alignment that saves ~30% of storage
-- The `sessions` table keeps track of online sessions of peers.
CREATE TABLE sessions
(
    -- A unique id that identifies a particular session
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               SERIAL      NOT NULL,
    -- When was the peer successfully dialed the first time
    first_successful_dial TIMESTAMPTZ NOT NULL,
    -- When was the most recent successful dial to the peer above
    last_successful_dial  TIMESTAMPTZ NOT NULL,
    -- When should we try to dial the peer again
    next_dial_attempt     TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    -- This cannot be null because otherwise the unique constraint
    -- uq_peer_id_first_failed_dial would not work (nulls are distinct).
    -- An unset value corresponds to the timestamp 1970-01-01
    first_failed_dial     TIMESTAMPTZ NOT NULL,
    -- The duration that this peer was online due to multiple subsequent successful dials
    min_duration          INTERVAL,
    -- The duration that from the first successful dial to the point were it was unreachable
    max_duration          INTERVAL,
    -- How many subsequent successful dials could we track
    successful_dials      INTEGER     NOT NULL,
    -- Why was this sessions marked as finished.
    finish_reason         dial_error,
    -- When was this session instance updated the last time
    updated_at            TIMESTAMPTZ NOT NULL,
    -- When was this session instance created
    created_at            TIMESTAMPTZ NOT NULL,
    -- indicates whether this session is finished or not. Equivalent to check for
    -- 1970-01-01 in the first_failed_dial field.
    finished              BOOLEAN     NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- There shouldn't be two active sessions for the same peer.
    -- An active session is a session where first_failed_dial is unset (aka. 1970-01-01)
    CONSTRAINT uq_peer_id_first_failed_dial UNIQUE (peer_id, first_failed_dial),
    -- Add a constraint that if a session is finished the reason can't be null.
    -- If the session is not finished the reason must be null.
    CONSTRAINT con_finish_reason_not_null_for_finished CHECK (
            (finished = TRUE AND finish_reason IS NOT NULL)
            OR
            (finished = FALSE AND finish_reason IS NULL)
        ),
    PRIMARY KEY (id)
);

INSERT INTO sessions (peer_id,
                      first_successful_dial,
                      last_successful_dial,
                      next_dial_attempt,
                      first_failed_dial,
                      min_duration,
                      max_duration,
                      successful_dials,
                      finish_reason,
                      updated_at,
                      created_at,
                      finished)
SELECT p.id,
       first_successful_dial,
       last_successful_dial,
       next_dial_attempt,
       first_failed_dial,
       min_duration,
       max_duration,
       successful_dials,
       finish_reason,
       s.updated_at,
       s.created_at,
       finished
FROM sessions_old s
         INNER JOIN peers p ON s.peer_id = p.multi_hash;

DROP TABLE sessions_old;

-- End the transaction
COMMIT;

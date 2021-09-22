-- Begin the transaction
BEGIN;

ALTER TABLE peers
    RENAME COLUMN id TO old_serial_id;

ALTER TABLE peers
    RENAME COLUMN multi_hash TO id;

-- Put all relevant tables aside
ALTER TABLE sessions
    RENAME TO sessions_old;

-- drop all relevant constraints
ALTER TABLE sessions_old
    DROP CONSTRAINT fk_sessions_peer;
ALTER TABLE sessions_old
    DROP CONSTRAINT uq_peer_id_first_failed_dial;
DROP INDEX idx_peers_multi_hash;

-- Make the new serial id column the new primary key on the peers table.
ALTER TABLE peers
    DROP CONSTRAINT peers_pkey;
ALTER TABLE peers
    ADD PRIMARY KEY (id);

-- The `sessions` table keeps track of online sessions of peers.
CREATE TABLE IF NOT EXISTS sessions
(
    -- A unique id that identifies a particular session
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- When was the peer successfully dialed the first time
    first_successful_dial TIMESTAMPTZ  NOT NULL,
    -- When was the most recent successful dial to the peer above
    last_successful_dial  TIMESTAMPTZ  NOT NULL,
    -- When should we try to dial the peer again
    next_dial_attempt     TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    -- This cannot be null because otherwise the unique constraint
    -- uq_peer_id_first_failed_dial would not work (nulls are distinct).
    -- An unset value corresponds to the timestamp 1970-01-01
    first_failed_dial     TIMESTAMPTZ  NOT NULL,
    -- The duration that this peer was online due to multiple subsequent successful dials
    min_duration          INTERVAL,
    -- The duration that from the first successful dial to the point were it was unreachable
    max_duration          INTERVAL,
    -- indicates whether this session is finished or not. Equivalent to check for
    -- 1970-01-01 in the first_failed_dial field.
    finished              BOOLEAN      NOT NULL,
    -- How many subsequent successful dials could we track
    successful_dials      INTEGER      NOT NULL,
    -- When was this session instance updated the last time
    updated_at            TIMESTAMPTZ  NOT NULL,
    -- When was this session instance created
    created_at            TIMESTAMPTZ  NOT NULL,
    -- Why was this sessions marked as finished.
    finish_reason         dial_error,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_session_peer FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
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
         INNER JOIN peers p ON s.peer_id = p.old_serial_id;

DROP TABLE sessions_old;

ALTER TABLE peers
    DROP COLUMN old_serial_id;

-- End the transaction
COMMIT;

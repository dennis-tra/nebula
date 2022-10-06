BEGIN;

-- The `sessions` table keeps track of online sessions of peers.
CREATE TABLE sessions
(
    -- A unique id that identifies a particular session
    id                    INT GENERATED ALWAYS AS IDENTITY,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               INT         NOT NULL,
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
    -- What's the reason why we closed this session.
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
    CONSTRAINT uq_sessions_peer_id_first_failed_dial UNIQUE (peer_id, first_failed_dial),
    -- Add a constraint that if a session is finished the reason can't be null.
    -- If the session is not finished the reason must be null.
    CONSTRAINT con_finish_reason_not_null_for_finished CHECK (
            (finished = TRUE AND finish_reason IS NOT NULL)
            OR
            (finished = FALSE AND finish_reason IS NULL)
        ),

    PRIMARY KEY (id)
);

COMMIT;
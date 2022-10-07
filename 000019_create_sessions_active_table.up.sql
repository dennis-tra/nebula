BEGIN;

-- The `sessions` table keeps track of all active sessions of peers.
CREATE TABLE sessions
(
    -- A unique id that identifies this particular active session
    id                      INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the remote peer ID.
    peer_id                 INT         NOT NULL,
    -- Timestamp of the first time we were able to visit that peer.
    first_successful_visit  TIMESTAMPTZ NOT NULL,
    -- Timestamp of the last time we were able to visit that peer.
    last_successful_visit   TIMESTAMPTZ NOT NULL,
    -- Number of successful visits in this session.
    successful_visits_count INTEGER     NOT NULL,
    -- Number of failed visits before closing this session.
    failed_visits_count     SMALLINT    NOT NULL,
    -- When was this session instance updated the last time
    updated_at              TIMESTAMPTZ NOT NULL,
    -- When was this session instance created
    created_at              TIMESTAMPTZ NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE TABLE sessions_active
(
    -- Timestamp when we should start visiting this peer again.
    next_visit_attempt TIMESTAMPTZ NOT NULL,
    -- What's the first error before we close this session.
    finish_reason      dial_error,
    -- When did we first notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ,

    -- There shouldn't be two active sessions for the same peer.
    CONSTRAINT uq_sessions_active_peer_id UNIQUE (peer_id)

) INHERITS (sessions);

CREATE TABLE sessions_closed
(
    -- When did we first notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ NOT NULL,
    -- When did we first notice that this peer is not reachable.
    last_failed_visit TIMESTAMPTZ NOT NULL,
    -- What's the reason why we closed this session.
    finish_reason     dial_error  NOT NULL,
    -- The duration that this peer was online due to multiple subsequent successful dials
    min_duration      INTERVAL    NOT NULL,
    -- The duration that from the first successful dial until the first failed dial
    max_duration      INTERVAL    NOT NULL
) INHERITS (sessions) PARTITION BY RANGE (created_at);

-- The `sessions_active` table keeps track of all active sessions of peers.
CREATE TABLE sessions_active
(
    -- A unique id that identifies this particular active session
    id                      INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the remote peer ID.
    peer_id                 INT         NOT NULL,
    -- Timestamp of the first time we were able to visit that peer.
    first_successful_visit  TIMESTAMPTZ NOT NULL,
    -- Timestamp of the last time we were able to visit that peer.
    last_successful_visit   TIMESTAMPTZ NOT NULL,
    -- Timestamp when we should start visiting this peer again.
    next_visit_attempt      TIMESTAMPTZ NOT NULL,
    -- When did we notice that this peer is not reachable.
    -- This cannot be null because otherwise the unique constraint
    -- uq_peer_id_first_failed_visit would not work (nulls are distinct).
    -- An unset value corresponds to the timestamp 1970-01-01
    first_failed_visit      TIMESTAMPTZ,
    -- Number of successful visits in this session.
    successful_visits_count INTEGER     NOT NULL,
    -- Number of failed visits before closing this session.
    failed_visits_count     SMALLINT    NOT NULL,
    -- What's the first error before we close this session.
    finish_reason           dial_error,
    -- When was this session instance updated the last time
    updated_at              TIMESTAMPTZ NOT NULL,
    -- When was this session instance created
    created_at              TIMESTAMPTZ NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- There shouldn't be two active sessions for the same peer.
    CONSTRAINT uq_sessions_active_peer_id UNIQUE (peer_id),

    PRIMARY KEY (id)
);

CREATE INDEX idx_sessions_active_next_visit_attempt ON sessions_active (next_visit_attempt);

COMMIT;
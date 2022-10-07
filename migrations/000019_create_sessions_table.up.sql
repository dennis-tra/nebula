BEGIN;

CREATE TYPE session_state AS ENUM (
    'open',
    'pending',
    'closed'
    );

CREATE TABLE sessions
(
    -- A unique id that identifies this particular session
    id                      INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the remote peer ID.
    peer_id                 INT           NOT NULL,
    -- Timestamp of the first time we were able to visit that peer.
    first_successful_visit  TIMESTAMPTZ   NOT NULL,
    -- Timestamp of the last time we were able to visit that peer.
    last_successful_visit   TIMESTAMPTZ   NOT NULL,
    -- Timestamp when we should start visiting this peer again.
    next_visit_attempt_at   TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ,
    -- When did we first notice that this peer is not reachable.
    last_failed_visit       TIMESTAMPTZ,
    -- When was this session instance updated the last time
    updated_at              TIMESTAMPTZ   NOT NULL,
    -- When was this session instance created
    created_at              TIMESTAMPTZ   NOT NULL,
    -- The duration that this peer was online due to multiple subsequent successful dials
    min_duration            INTERVAL,
    -- The duration that from the first successful dial until the first failed dial
    max_duration            INTERVAL,
    -- Number of successful visits in this session.
    successful_visits_count INTEGER       NOT NULL,
    -- The state this session is in.
    state                   session_state NOT NULL,
    -- Number of failed visits before closing this session.
    failed_visits_count     SMALLINT      NOT NULL,
    -- The number of times this session went from pending to open again.
    recovered_count         INTEGER       NOT NULL,
    -- What's the first error before we close this session.
    finish_reason           dial_error,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id, state, created_at)

) PARTITION BY LIST (state);

-------------------------------------------------
-- Create partition for open sessions
-------------------------------------------------

CREATE TABLE sessions_open PARTITION OF sessions(
    id DEFAULT nextval('sessions_id_seq'),
    next_visit_attempt_at NOT NULL
    ) FOR VALUES IN ('open', 'pending') WITH (fillfactor = 90);
-- we're doing lots of updates on this table so decrease fill factor.

-- There shouldn't be two active sessions for the same peer.
CREATE UNIQUE INDEX uq_sessions_open_peer_id ON sessions_open (peer_id);

-- Create index on visit attempt to efficiently query due sessions.
CREATE INDEX idx_sessions_open_next_visit_attempt_at ON sessions_open (next_visit_attempt_at);

-------------------------------------------------
-- Create partition for closed sessions
-------------------------------------------------
CREATE TABLE sessions_closed PARTITION OF sessions (
    first_failed_visit NOT NULL,
    last_failed_visit NOT NULL,
    min_duration NOT NULL,
    max_duration NOT NULL,
    finish_reason NOT NULL
    ) FOR VALUES IN ('closed') PARTITION BY RANGE (created_at);

CREATE INDEX idx_sessions_closed_created_at ON sessions_closed (created_at);

CREATE TABLE sessions_closed_2022_10 PARTITION OF sessions_closed
    FOR VALUES FROM ('2022-10-01') TO ('2022-11-01');

COMMIT;
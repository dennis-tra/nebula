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
    next_visit_due_at       TIMESTAMPTZ,
    -- When did we notice that this peer is not reachable.
    first_failed_visit      TIMESTAMPTZ,
    -- When did we first notice that this peer is not reachable.
    last_failed_visit       TIMESTAMPTZ,
    -- When did we last visit this peer. For indexing purposes.
    last_visited_at         TIMESTAMPTZ   NOT NULL,
    -- When was this session instance updated the last time
    updated_at              TIMESTAMPTZ   NOT NULL,
    -- When was this session instance created
    created_at              TIMESTAMPTZ   NOT NULL,
    -- Number of successful visits in this session.
    successful_visits_count INTEGER       NOT NULL,
    -- The number of times this session went from pending to open again.
    recovered_count         INTEGER       NOT NULL,
    -- The state this session is in.
    state                   session_state NOT NULL,
    -- Number of failed visits before closing this session.
    failed_visits_count     SMALLINT      NOT NULL,
    -- What's the first error before we close this session.
    finish_reason           net_error,
    -- The uptime time range for this session measured from first- to last_successful_visit to
    uptime                  TSTZRANGE     NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_sessions_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id, state, last_visited_at)

) PARTITION BY LIST (state);

-------------------------------------------------
-- Create partition for open sessions
-------------------------------------------------

CREATE TABLE sessions_open PARTITION OF sessions(
    id DEFAULT nextval('sessions_id_seq'),
    next_visit_due_at NOT NULL
) FOR VALUES IN ('open', 'pending') WITH (fillfactor = 80);

-- we're doing lots of updates on this table so decrease fill factor.

-- There shouldn't be two active sessions for the same peer.
CREATE UNIQUE INDEX uq_sessions_open_peer_id ON sessions_open (peer_id);

-- Create index on visit attempt to efficiently query due sessions.
-- CREATE INDEX idx_sessions_open_next_visit_due_at ON sessions_open (next_visit_due_at); commented out as it may not be needed

-------------------------------------------------
-- Create partition for closed sessions
-------------------------------------------------
CREATE TABLE sessions_closed PARTITION OF sessions (
    first_failed_visit NOT NULL,
    last_failed_visit NOT NULL,
    uptime NOT NULL,
    finish_reason NOT NULL
) FOR VALUES IN ('closed') PARTITION BY RANGE (last_visited_at);

CREATE INDEX idx_sessions_closed_last_visited_at ON sessions_closed (last_visited_at);
CREATE INDEX idx_sessions_closed_uptime ON sessions_closed USING GIST (uptime);

COMMIT;
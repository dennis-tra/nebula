-- Begin the transaction
BEGIN;

CREATE TYPE visit_type AS ENUM (
    'crawl',
    'dial'
    );

-- Rows in the `visits` table represent the event of visiting a peer.
-- It contains foreign key constraints and the likes...
-- This was the most accurate name for me as it also captures the idea
-- of not reaching a particular peer and thus err out.
CREATE TABLE visits
(
    -- A unique id that identifies this visit
    id                INT GENERATED ALWAYS AS IDENTITY,
    -- Which peer did we meet
    peer_id           INT         NOT NULL,
    -- In which crawl did we meet this peer (can be null if recorded during monitoring)
    crawl_id          INT,
    -- The session that this visit belongs to (no foreign key constraint).
    session_id        INT,
    -- The agent version that this peer reported during this visit.
    agent_version_id  INT,
    -- The set of supported protocols that this peer reported.
    protocols_set_id  INT,
    -- The type of this visit (done here for column alignment)
    type              visit_type  NOT NULL,
    -- The error that happened for this visit.
    error             dial_error,
    -- The time it took to connect with the peer
    visit_started_at  TIMESTAMPTZ NOT NULL,
    -- The time it took to connect with the peer
    visit_ended_at    TIMESTAMPTZ NOT NULL,
    -- When did this visit happen
    created_at        TIMESTAMPTZ NOT NULL,
    -- All multi addresses of this peer.
    -- The time it took to dial the peer or until an error occurred
    dial_duration     INTERVAL,
    -- The time it took to connect with the peer or until an error occurred
    connect_duration  INTERVAL,
    -- The time it took to crawl the peer also if an error occurred
    crawl_duration    INTERVAL,
    -- An array of all multi address IDs of the remote peer.
    multi_address_ids INT[],

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_visits_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_visits_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE SET NULL,
    -- The session ID should always point to an existing session instance in the DB
    CONSTRAINT fk_visits_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,
    -- The protocol set ID should always point to an existing protocol set in the DB
    CONSTRAINT fk_visits_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL,

    PRIMARY KEY (id, visit_started_at)
) PARTITION BY RANGE (visit_started_at);

CREATE INDEX idx_visits_visit_started_at ON visits (visit_started_at);
CREATE INDEX idx_visits_visit_started_at_peer_id ON visits (visit_started_at, peer_id);

COMMIT;

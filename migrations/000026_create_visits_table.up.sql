-- Begin the transaction
BEGIN;

-- Rows in the `visits` table represent the event of visiting a peer.
-- This is the highly normalized representation of the raw_visits table.
-- It contains foreign key constraints and the likes...
-- This was the most accurate name for me as it also captures the idea
-- of not reaching a particular peer and thus err out.
-- Other names I've considered:
--   - encounters - this implies actually being able to connect to the peer
--   - meetings - this also implies being able to talk to each other
--   - connections - closest competitor imo - however the pegasys team already used that name
CREATE TABLE visits
(
    -- A unique id that identifies this visit
    id               SERIAL,
    -- Which peer did we meet
    peer_id          SERIAL      NOT NULL,
    -- In which crawl did we meet this peer (can be null if recorded during monitoring)
    crawl_id         INT,
    -- Which peer session did we update with this visit
    session_id       INT,
    -- The time it took to dial the peer or until an error occurred
    dial_duration    INTERVAL,
    -- The time it took to connect with the peer or until an error occurred
    connect_duration INTERVAL,
    -- The time it took to crawl the peer also if an error occurred
    crawl_duration   INTERVAL,
    -- When was this visit updated the last time
    updated_at       TIMESTAMPTZ NOT NULL,
    -- When did this visit happen
    created_at       TIMESTAMPTZ NOT NULL,
    -- The type of this visit (done here for column alignment)
    type             visit_type  NOT NULL,
    -- The type of this visit (done here for column alignment)
    error            dial_error,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_visits_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id),
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_visits_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id),
    -- The session ID should always point to an existing session instance in the DB
    CONSTRAINT fk_visits_session_id FOREIGN KEY (session_id) REFERENCES sessions (id),

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;

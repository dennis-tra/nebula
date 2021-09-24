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
    -- A unique id that identifies this encounter
    id         SERIAL,
    -- Which peer did we meet
    peer_id    SERIAL      NOT NULL,
    -- In which crawl did we meet this peer (can be null if recorded during monitoring)
    crawl_id   INT,
    -- Which peer session did we update with this encounter
    session_id INT,

    -- When was this encounter updated the last time
    updated_at TIMESTAMPTZ NOT NULL,
    -- When did this encounter happen
    created_at TIMESTAMPTZ NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_encounter_peer FOREIGN KEY (peer_id) REFERENCES peers (id),
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_encounter_crawl FOREIGN KEY (crawl_id) REFERENCES crawls (id),
    -- The session ID should always point to an existing session instance in the DB
    CONSTRAINT fk_encounter_session FOREIGN KEY (session_id) REFERENCES sessions (id),

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;

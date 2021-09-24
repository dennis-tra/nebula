-- Begin the transaction
BEGIN;

-- The `raw_visits` table captures information about
-- a 'visit' of the crawler with another peer. The data
-- in this table is purposely very denormalized. There
-- are also no foreign key constraints or the likes on this table.
-- Doing it this way allows quick inserts into the database.
CREATE TABLE raw_visits
(
    -- This field identifies this encounter.
    id                 SERIAL,
    -- During which crawl did we visit this peer
    crawl_id           SERIAL       NOT NULL,
    -- The time it took to connect with the peer
    connect_latency    INTERVAL,
    -- The time it took to connect with the peer
    connect_started_at TIMESTAMPTZ  NOT NULL,
    -- The peer ID multi hash of which we want to track the multi address
    peer_multi_hash    VARCHAR(150) NOT NULL,
    -- Which agent version did this peer have at this visit
    agent_version      VARCHAR(255),
    -- Which protocols does this peer support
    protocols          VARCHAR(255) ARRAY,
    -- All multi addresses for this peer
    multi_addresses    VARCHAR(255) ARRAY,
    -- The error if one occurred
    error              TEXT,

    -- When was this peer visited
    created_at         TIMESTAMPTZ  NOT NULL,

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;

-- Begin the transaction
BEGIN;

CREATE TYPE visit_type AS ENUM (
    'crawl',
    'dial'
    );

-- The `raw_visits` table captures information about
-- a 'visit' of the crawler with another peer. The data
-- in this table is purposely very denormalized. There
-- are also no foreign key constraints or the likes on this table.
-- Doing it this way allows quick inserts into the database.
CREATE TABLE raw_visits
(
    -- This field identifies this encounter.
    id               SERIAL,
    -- During which crawl did we visit this peer
    crawl_id         INT,
    -- The time it took to connect with the peer
    visit_started_at TIMESTAMPTZ  NOT NULL,
    -- The time it took to connect with the peer
    visit_ended_at   TIMESTAMPTZ  NOT NULL,
    -- The time it took to dial the peer or until an error occurred
    dial_duration    INTERVAL,
    -- The time it took to connect with the peer or until an error occurred
    connect_duration INTERVAL,
    -- The time it took to crawl the peer also if an error occurred
    crawl_duration   INTERVAL,
    -- The type of this visit
    type             visit_type NOT NULL,
    -- Which agent version did this peer have at this visit
    agent_version    VARCHAR(255),
    -- The peer ID multi hash of which we want to track the multi address
    peer_multi_hash  VARCHAR(150) NOT NULL,
    -- Which protocols does this peer support
    protocols        VARCHAR(255) ARRAY,
    -- All multi addresses for this peer
    multi_addresses  VARCHAR(255) ARRAY,
    -- Associated dial error
    dial_error       dial_error,
    -- The error if one occurred
    error            TEXT,

    -- When was this peer visited
    created_at       TIMESTAMPTZ  NOT NULL,

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;

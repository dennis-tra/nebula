-- Begin the transaction
BEGIN;

-- Put all current crawls tables aside
ALTER TABLE crawls
    RENAME TO crawls_old;

-- The different state in which a crawl can be
CREATE TYPE crawl_state AS ENUM (
    'started',
    'cancelled',
    'failed',
    'succeeded'
    );

-- The `crawls` table captures the result of one particular crawl.
-- We are re-adding the table here so that we can properly align the columns
-- and make the current schema more obvious.
CREATE TABLE crawls
(
    -- A unique id that identifies a crawl
    id               SERIAL,
    -- The state of this crawl
    state            crawl_state NOT NULL,
    -- When did the crawl process start
    started_at       TIMESTAMPTZ NOT NULL,
    -- When did the crawl process finish
    finished_at      TIMESTAMPTZ,
    -- When was this crawl updated the last time
    updated_at       TIMESTAMPTZ NOT NULL,
    -- When was this crawl instance created (different from started_at)
    created_at       TIMESTAMPTZ NOT NULL,
    -- How many peers were visited in this crawl
    crawled_peers    INTEGER,
    -- How many peers were successfully dialed
    dialable_peers   INTEGER,
    -- How many peers were found but couldn't be reached
    undialable_peers INTEGER,

    PRIMARY KEY (id)
);

INSERT INTO crawls (id,
                    state,
                    started_at,
                    finished_at,
                    updated_at,
                    created_at,
                    crawled_peers,
                    dialable_peers,
                    undialable_peers)
SELECT id,
       'succeeded',
       started_at,
       finished_at,
       updated_at,
       created_at,
       crawled_peers,
       dialable_peers,
       undialable_peers
FROM crawls_old;

ALTER TABLE peer_properties
    DROP CONSTRAINT fk_peer_property_crawl;

ALTER TABLE peer_properties
    ADD CONSTRAINT fk_peer_property_crawl
        FOREIGN KEY (crawl_id)
            REFERENCES crawls (id)
            ON DELETE CASCADE;

DROP TABLE crawls_old;

-- End the transaction
COMMIT;

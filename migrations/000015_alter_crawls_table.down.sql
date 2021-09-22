-- Begin the transaction
BEGIN;

-- Put all current crawls tables aside
ALTER TABLE crawls
    RENAME TO crawls_old;


-- The `crawls` table captures the result of one particular crawl.
CREATE TABLE IF NOT EXISTS crawls
(
    -- A unique id that identifies a crawl
    id               SERIAL,
    -- When did the crawl process start
    started_at       TIMESTAMPTZ NOT NULL,
    -- When did the crawl process finish
    finished_at      TIMESTAMPTZ NOT NULL,
    -- How many peers were visited in this crawl
    crawled_peers    INTEGER     NOT NULL,
    -- How many peers were successfully dialed
    dialable_peers   INTEGER     NOT NULL,
    -- How many peers were found but couldn't be reached
    undialable_peers INTEGER     NOT NULL,
    -- When was this crawl updated the last time
    updated_at       TIMESTAMPTZ NOT NULL,
    -- When was this crawl instance created (different from started_at)
    created_at       TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (id)
);


INSERT INTO crawls (id,
                    started_at,
                    finished_at,
                    updated_at,
                    created_at,
                    crawled_peers,
                    dialable_peers,
                    undialable_peers)
SELECT id,
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
DROP TYPE crawl_state;

-- End the transaction
COMMIT;

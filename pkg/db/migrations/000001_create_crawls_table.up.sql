BEGIN;

-- The different states in which a crawl can be:
--   1) started: The crawl was started. In this state finished_at, crawled_peers, dialable_peers, undialable_peers is NULL.
--   2) cancelled: The crawl was stopped by the user.
--   3) failed: The crawl failed for whatever reason.
--   4) succeeded: The crawl finished successfully.
CREATE TYPE crawl_state AS ENUM (
    'started',
    'cancelled',
    'failed',
    'succeeded'
    );

COMMENT ON TYPE crawl_state IS ''
    'The different states in which a crawl can be:'
    ' 1) started: The crawl was started. In this state finished_at, crawled_peers, dialable_peers, undialable_peers is NULL.'
    ' 2) cancelled: The crawl was stopped by the user.'
    ' 3) failed: The crawl failed for whatever reason.'
    ' 4) succeeded: The crawl finished successfully.';

-- Captures the state and aggregated results of one particular crawl.
CREATE TABLE crawls
(
    -- An internal unique id that identifies a crawl.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The state of this crawl.
    state            crawl_state NOT NULL,
    -- Timestamp of when this crawl process started.
    started_at       TIMESTAMPTZ NOT NULL,
    -- Timestamp of when this crawl process finished.
    finished_at      TIMESTAMPTZ CHECK ( finished_at > started_at ),
    -- Timestamp of when this crawl row was updated the last time.
    updated_at       TIMESTAMPTZ NOT NULL CHECK ( updated_at >= created_at ),
    -- Timestamp of when this crawl instance was created which can slightly differ from the started_at timestamp.
    created_at       TIMESTAMPTZ NOT NULL,
    -- Number of _visited_ peers during this crawl.
    crawled_peers    INTEGER CHECK ( crawled_peers = dialable_peers + undialable_peers ),
    -- Number of successfully dialed peers during this crawl.
    dialable_peers   INTEGER CHECK ( dialable_peers >= 0 ),
    -- Number of peers that could not be reached during this crawl.
    undialable_peers INTEGER CHECK ( undialable_peers >= 0 ),

    PRIMARY KEY (id)
);

COMMENT ON TABLE crawls IS 'Captures the state and aggregated results of one particular crawl.';
COMMENT ON COLUMN crawls.id IS 'An internal unique id that identifies a crawl.';
COMMENT ON COLUMN crawls.state IS 'The state of this crawl.';
COMMENT ON COLUMN crawls.started_at IS 'Timestamp of when this crawl process started.';
COMMENT ON COLUMN crawls.finished_at IS 'Timestamp of when this crawl process finished.';
COMMENT ON COLUMN crawls.updated_at IS 'Timestamp of when this crawl row was updated the last time.';
COMMENT ON COLUMN crawls.created_at IS 'Timestamp of when this crawl instance was created which can slightly differ from the started_at timestamp.';
COMMENT ON COLUMN crawls.crawled_peers IS 'Number of _visited_ peers during this crawl.';
COMMENT ON COLUMN crawls.dialable_peers IS 'Number of successfully dialed peers during this crawl.';
COMMENT ON COLUMN crawls.undialable_peers IS 'Number of peers that could not be reached during this crawl.';

CREATE INDEX idx_crawls_started_at ON crawls (started_at);

COMMIT;
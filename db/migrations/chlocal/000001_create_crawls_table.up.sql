-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations

-- Captures the state and aggregated results of one particular crawl.
CREATE TABLE crawls
(
    -- An unique id that identifies a crawl. Use UUIDv7 for temporal sorting and uniqueness.
    id               UUID,
    -- The state of this crawl.
    state            Enum('started', 'cancelled', 'failed', 'succeeded'),
    -- Timestamp of when this crawl process finished.
    finished_at      Nullable(DATETIME64(3)),
    -- Timestamp of when this crawl row was updated the last time.
    updated_at       DATETIME64(3),
    -- Timestamp of when this crawl instance was created.
    created_at       DATETIME64(3),
    -- Number of _visited_ peers during this crawl.
    crawled_peers    Nullable(Int),
    -- Number of successfully dialed peers during this crawl.
    dialable_peers   Nullable(Int),
    -- Number of peers that could not be reached during this crawl.
    undialable_peers Nullable(Int),
    -- Number of remaining peers in the crawl queue if the process was canceled.
    remaining_peers  Nullable(Int),
    -- The version of Nebula that produced the crawl.
    version          String,
    -- A unique identifier of the network that was crawled.
    network_id       String
) ENGINE ReplacingMergeTree(updated_at)
    PRIMARY KEY (id, created_at)
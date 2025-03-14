CREATE TABLE neighbors
(
    -- identifies the crawl that produced this neighbor mapping
    crawl_id                     UUID,

    -- the date when the crawl was created/started. This is used to partition
    -- the data. Could have used ClickHouse's `UUIDv7ToDateTime` and the
    -- crawl_id column but then querying by date to remove partitions would
    -- be a pain. Hence, we add another column which isn't a big deal as it
    -- should compress extremely well.
    crawl_created_at             DATETIME64(3),

    -- to save space on disk we use the first 64 bits of the identifier that's
    -- used in the respective discovery protocol of the given network. In libp2p
    -- based networks these are usually the first 64 bit of the sha256 hash
    -- of the multihash component of the peer ID. For discv4 and discv5 it's
    -- different.
    peer_discovery_id_prefix     UInt64,

    -- same as the peer discovery ID
    neighbor_discovery_id_prefix UInt64,

    -- a bit string representing which request for which bucket failed. If the
    -- first bit is set then the request for the 0-th bucket has failed.
    error_bits                   UInt16
) ENGINE ReplicatedMergeTree()
    PRIMARY KEY (
        crawl_id,
        peer_discovery_id_prefix,
        neighbor_discovery_id_prefix
    )
    -- add weekly partitioning. Mode "3" is in accordance with ISO 8601:1988,
    -- considers Monday the first day of the week, and is also used by
    -- ClickHouse's `toISOWeek()` compatibility function.
    PARTITION BY toStartOfWeek(crawl_created_at, 3)
    TTL toDateTime(crawl_created_at) + INTERVAL 1 YEAR;

CREATE TABLE neighbors
(
    crawl_id                     UUID,
    peer_discovery_id_prefix     UInt64,
    neighbor_discovery_id_prefix UInt64,
    error_bits                   UInt16
) ENGINE ReplicatedMergeTree()
      PRIMARY KEY (crawl_id, peer_discovery_id_prefix, neighbor_discovery_id_prefix)

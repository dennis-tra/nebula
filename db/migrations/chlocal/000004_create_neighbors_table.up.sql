-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations

CREATE TABLE neighbors
(
    crawl_id                     UUID,
    peer_discovery_id_prefix     UInt64,
    neighbor_discovery_id_prefix UInt64,
    error_bits                   UInt16
) ENGINE MergeTree()
      PRIMARY KEY (crawl_id, peer_discovery_id_prefix, neighbor_discovery_id_prefix)

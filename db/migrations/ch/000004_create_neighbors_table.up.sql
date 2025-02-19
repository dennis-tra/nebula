CREATE TABLE neighbors
(
    crawl_id               UUID,
    peer_kad_id_prefix     UInt64,
    neighbor_kad_id_prefix UInt64,
    error_bits             UInt16
) ENGINE MergeTree()
      PRIMARY KEY (crawl_id, peer_kad_id_prefix, neighbor_kad_id_prefix)

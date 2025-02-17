CREATE TABLE neighbors
(
    crawl_id    UUID,
    peer_id     String,
    neighbor_id String,
    error_bits  UInt16
) ENGINE MergeTree()
      PRIMARY KEY (crawl_id, peer_id, neighbor_id)


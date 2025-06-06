CREATE TABLE discovery_id_prefixes_x_peer_ids
(
    discovery_id_prefix UInt64,
    peer_id             String
) ENGINE ReplicatedReplacingMergeTree()
      PRIMARY KEY (discovery_id_prefix)
-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations

CREATE TABLE discovery_id_prefixes_x_peer_ids
(
    discovery_id_prefix UInt64,
    peer_id             String
) ENGINE ReplacingMergeTree()
      PRIMARY KEY (discovery_id_prefix)
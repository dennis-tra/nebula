BEGIN;

CREATE OR REPLACE FUNCTION insert_neighbors(
    crawl INT,
    peer_db_id INT,
    peer_multi_hash TEXT,
    neighbors_db_ids INT[],
    neighbors_multi_hashes TEXT[],
    crawl_error_bits SMALLINT
) RETURNS VOID AS
$insert_neighbors$
WITH neighbor_peers AS (
    SELECT ndbid peer_id FROM unnest(neighbors_db_ids) ndbid
    UNION
    SELECT upserted_peer peer_id
    FROM unnest(neighbors_multi_hashes) neighbors_multi_hash
             CROSS JOIN upsert_peer(neighbors_multi_hash) upserted_peer
    ORDER BY peer_id
)
INSERT INTO neighbors (crawl_id, peer_id, neighbor_ids, error_bits)
SELECT crawl, COALESCE(peer_db_id, upsert_peer(peer_multi_hash)), array_agg(peer_id), crawl_error_bits
FROM neighbor_peers
GROUP BY 1, 2, 4;
$insert_neighbors$ LANGUAGE sql;


COMMIT;

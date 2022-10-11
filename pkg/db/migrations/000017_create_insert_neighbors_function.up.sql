BEGIN;

CREATE FUNCTION insert_neighbors(
    crawl INT,
    peer_multi_hash TEXT,
    neighbors_multi_hashes TEXT[],
    crawl_error_bits SMALLINT
) RETURNS VOID AS
$insert_neighbors$
    WITH neighbor_peers AS (
        SELECT upserted_peer
        FROM unnest(neighbors_multi_hashes) neighbors_multi_hash
        CROSS JOIN upsert_peer(neighbors_multi_hash) upserted_peer
    )
    INSERT INTO neighbors (crawl_id, peer_id, neighbor_ids, error_bits)
    SELECT crawl, upsert_peer(peer_multi_hash), array_agg(upserted_peer), crawl_error_bits
    FROM neighbor_peers
    GROUP BY 1, 2, 4;
$insert_neighbors$ LANGUAGE sql;

COMMIT;

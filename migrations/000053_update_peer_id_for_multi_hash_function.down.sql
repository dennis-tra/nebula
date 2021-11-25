BEGIN;

DROP FUNCTION peer_id_for_multi_hash;
DROP FUNCTION insert_neighbors;


-- upserts a peer database entry to receive an internal DB id for a given multi hash.
CREATE OR REPLACE FUNCTION peer_id_for_multi_hash(mhash varchar)
    RETURNS INT
    LANGUAGE plpgsql
AS
$$
DECLARE
    peer_id INT;
BEGIN
    INSERT INTO peers (multi_hash, updated_at, created_at)
    VALUES (mhash, NOW(), NOW())
    ON CONFLICT (multi_hash) DO UPDATE SET updated_at=excluded.updated_at
    RETURNING id INTO peer_id;

    RETURN peer_id;
END;
$$;

CREATE FUNCTION insert_neighbors(crawl INT, peer_multi_hash VARCHAR, neighbors_multi_hashes VARCHAR[])
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    INSERT INTO neighbors (crawl_id, peer_id, neighbor_ids)
    VALUES (crawl, peer_id_for_multi_hash(peer_multi_hash),
            (SELECT array_agg(peer_id_for_multi_hash)
             FROM unnest(neighbors_multi_hashes)
                      CROSS JOIN peer_id_for_multi_hash(unnest)));
END;
$$;

COMMIT;

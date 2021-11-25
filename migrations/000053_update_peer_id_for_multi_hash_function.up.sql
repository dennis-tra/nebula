BEGIN;

DROP FUNCTION peer_id_for_multi_hash;
DROP FUNCTION insert_neighbors;

-- upserts a peer database entry to receive an internal DB id for a given multi hash.
CREATE OR REPLACE FUNCTION peer_id_for_multi_hash(
    mhash TEXT
) RETURNS INT AS
$peer_id_for_multi_hash$
DECLARE
    peer_id INT;
    peer    peers%rowtype;
BEGIN
    SELECT *
    FROM peers p
    WHERE p.multi_hash = mhash
    INTO peer;

    IF peer IS NULL THEN
        INSERT INTO peers (multi_hash, updated_at, created_at)
        VALUES (mhash, NOW(), NOW())
        RETURNING id INTO peer_id;

        RETURN peer_id;
    END IF;

    RETURN peer.id;
END;
$peer_id_for_multi_hash$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION insert_neighbors(
    crawl INT,
    peer_multi_hash TEXT,
    neighbors_multi_hashes TEXT[]
) RETURNS VOID AS
$insert_neighbors$
BEGIN
    INSERT INTO neighbors (crawl_id, peer_id, neighbor_ids)
    VALUES (crawl, peer_id_for_multi_hash(peer_multi_hash),
            (SELECT array_agg(peer_id_for_multi_hash)
             FROM unnest(neighbors_multi_hashes)
                      CROSS JOIN peer_id_for_multi_hash(unnest)));
END;
$insert_neighbors$ LANGUAGE plpgsql;

COMMIT;

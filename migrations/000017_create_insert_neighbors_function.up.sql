BEGIN;

CREATE FUNCTION insert_neighbors(crawl INT, peer_multi_hash VARCHAR, neighbors_multi_hashes VARCHAR[])
    RETURNS VOID
    LANGUAGE plpgsql
AS
$$
BEGIN
    INSERT INTO neighbors (--
        crawl_id,
        peer_id,
        neighbor_ids)
    VALUES (--
               crawl,
               upsert_peer(peer_multi_hash, NULL, NULL),
               (
                   SELECT array_agg(upsert_peer)
                   FROM unnest(neighbors_multi_hashes)
                            CROSS JOIN upsert_peer(unnest, NULL, NULL)));
END;
$$;

COMMIT;

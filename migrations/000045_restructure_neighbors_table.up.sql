-- Begin the transaction
BEGIN;

DROP TABLE neighbors;

-- The `neighbors` save topology information for a particular crawl
CREATE TABLE neighbors
(
    -- During which crawl was this neighbor identified
    crawl_id     INT,
    -- The peer and their neighbor (entry in its k-buckets)
    peer_id      INT,
    -- The peers to which the peer above is connected
    neighbor_ids INT[],

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_neighbors_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_neighbors_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE CASCADE,

    PRIMARY KEY (crawl_id, peer_id)
);

CREATE INDEX idx_neighbors_peer_id ON neighbors (peer_id);


CREATE FUNCTION peer_id_for_multi_hash(mhash varchar)
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

-- End the transaction
COMMIT;

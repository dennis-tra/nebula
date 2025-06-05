BEGIN;

-- The `neighbors` save topology information for a particular crawl
CREATE TABLE neighbors
(
    -- During which crawl was this neighbor identified
    crawl_id     INT,
    -- The peer and their neighbor (entry in its k-buckets)
    peer_id      INT,
    -- The peers to which the peer above is connected
    neighbor_ids INT[],

    error_bits   SMALLINT DEFAULT 0 NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_neighbors_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_neighbors_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE CASCADE,

    PRIMARY KEY (crawl_id, peer_id)
) PARTITION BY RANGE (crawl_id);

-- End the transaction
COMMIT;

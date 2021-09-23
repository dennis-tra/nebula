-- Begin the transaction
BEGIN;

-- The `neighbors` save topology information for a particular crawl
CREATE TABLE neighbors
(
    id          SERIAL,
    -- During which crawl was this neighbor identified
    crawl_id     SERIAL,
    -- The peer and their neighbor (entry in its k-buckets)
    peer_id     SERIAL,
    -- The peer to which the peer above is connected
    neighbor_id SERIAL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_neighbors_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The neighbor ID should always point to an existing peer in the DB
    CONSTRAINT fk_neighbors_neighbor_id FOREIGN KEY (neighbor_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The crawl ID should always point to an existing crawl in the DB
    CONSTRAINT fk_neighbors_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE INDEX idx_neighbors_peer_id ON neighbors (peer_id);
CREATE INDEX idx_neighbors_neighbor_id ON neighbors (neighbor_id);

-- End the transaction
COMMIT;

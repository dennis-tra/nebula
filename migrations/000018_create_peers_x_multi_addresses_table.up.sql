-- Begin the transaction
BEGIN;

-- The `peers_x_multi_addresses` table keeps track of
-- the association of multi addresses to peers.
CREATE TABLE peers_x_multi_addresses
(
    -- The peer ID of which we want to track the multi address
    peer_id          INT NOT NULL,
    -- The ID of the multi address that has been seen for the above peer
    multi_address_id INT NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_peers_x_multi_addresses_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The maddr ID should always point to an existing multi address in the DB
    CONSTRAINT fk_peers_x_multi_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (peer_id)
);

-- End the transaction
COMMIT;

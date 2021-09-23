-- Begin the transaction
BEGIN;

-- The `peers_multi_addresses` table keeps track of
-- the association of multi addresses to peers.
CREATE TABLE peers_multi_addresses
(
    -- The peer ID of which we want to track the multi address
    peer_id  SERIAL,
    -- The ID of the multi address that has been seen for the above peer
    multi_address_id SERIAL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_peers_multi_addresses_peer FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The maddr ID should always point to an existing multi address in the DB
    CONSTRAINT fk_peers_multi_addresses_maddr FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (peer_id, multi_address_id)
);

CREATE INDEX idx_multi_addresses_pkey_reversed ON peers_multi_addresses (multi_address_id, peer_id);

-- End the transaction
COMMIT;

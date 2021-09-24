-- Begin the transaction
BEGIN;

-- The `peers_x_properties` table keeps track of
-- the most up to date association of a peer to its properties
CREATE TABLE peers_x_properties
(
    -- The peer ID of which we want to track the multi address
    peer_id  SERIAL,
    -- The ID of the property that with this peer is currently associated
    property_id SERIAL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_peers_x_properties_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    -- The property ID should always point to an existing property in the DB
    CONSTRAINT fk_peers_x_properties_property_id FOREIGN KEY (property_id) REFERENCES properties (id) ON DELETE CASCADE,

    PRIMARY KEY (peer_id, property_id)
);

CREATE INDEX idx_peers_x_properties_pkey_reversed ON peers_x_properties (property_id, peer_id);

-- End the transaction
COMMIT;

-- Begin the transaction
BEGIN;

-- The `visits_x_multi_addresses` table keeps track of
-- the association of multi addresses to visits.
CREATE TABLE visits_x_multi_addresses
(
    -- The visit ID of which we want to track the multi addresses
    visit_id         SERIAL,
    -- The ID of the multi address that has been seen for the above peer
    multi_address_id SERIAL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_visits_x_multi_addresses_peer FOREIGN KEY (visit_id) REFERENCES visits (id) ON DELETE CASCADE,
    -- The maddr ID should always point to an existing multi address in the DB
    CONSTRAINT fk_visits_x_multi_addresses_maddr FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (visit_id, multi_address_id)
);

CREATE INDEX idx_visits_x_multi_addresses_pkey_reversed ON visits_x_multi_addresses (multi_address_id, visit_id);

-- End the transaction
COMMIT;

-- Begin the transaction
BEGIN;

ALTER TABLE multi_addresses
    DROP COLUMN addr,
    DROP COLUMN country,
    DROP COLUMN cloud_provider;

-- The `multi_addresses_x_ip_addresses` table keeps track of
-- the association of multi addresses to their ip addresses.
-- A multi address can contain multiple IP addresses (relays, dnsaddr protocol, etc.).
-- IP addresses can be associated to multiple multi addresses.
CREATE TABLE multi_addresses_x_ip_addresses
(
    -- The ID of the multi address
    multi_address_id SERIAL,
    -- The ID of the IP address entry
    ip_address_id    SERIAL,
    -- When was this association created
    resolved_at      TIMESTAMPTZ NOT NULL,

    -- The multi address ID should always point to an existing multi address in the DB
    CONSTRAINT fk_multi_addresses_x_ip_addresses_multi_address_ip FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    -- The IP address ID should always point to an existing IP address in the DB
    CONSTRAINT fk_multi_addresses_x_ip_addresses_ip_address_id FOREIGN KEY (ip_address_id) REFERENCES ip_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, ip_address_id, resolved_at)
);

CREATE INDEX idx_multi_addresses_x_ip_addresses_1 ON multi_addresses_x_ip_addresses (ip_address_id, multi_address_id);
CREATE INDEX idx_multi_addresses_x_ip_addresses_2 ON multi_addresses_x_ip_addresses (multi_address_id, ip_address_id);

-- End the transaction
COMMIT;

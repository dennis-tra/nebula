-- Begin the transaction
BEGIN;

DROP TABLE multi_addresses_x_ip_addresses;

ALTER TABLE multi_addresses
    ADD COLUMN addr           INET,
    ADD COLUMN country        VARCHAR(3),
    ADD COLUMN cloud_provider VARCHAR(16);

-- End the transaction
COMMIT;

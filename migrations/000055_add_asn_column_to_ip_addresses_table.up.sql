BEGIN;

ALTER TABLE ip_addresses
    ADD COLUMN asn INT;

END;

-- Begin the transaction
BEGIN;

ALTER TABLE visits DROP CONSTRAINT fk_visits_multi_addresses_set_id;
ALTER TABLE visits DROP COLUMN multi_addresses_set_id;

DROP TABLE multi_addresses_sets;

-- End the transaction
COMMIT;

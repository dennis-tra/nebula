-- Begin the transaction
BEGIN;

ALTER TABLE peers DROP CONSTRAINT fk_peers_protocols_set_id;
ALTER TABLE peers DROP COLUMN protocols_set_id;

ALTER TABLE visits DROP CONSTRAINT fk_visits_protocols_set_id;
ALTER TABLE visits DROP COLUMN protocols_set_id;

DROP TABLE protocols_sets;
DROP TABLE protocols;

-- End the transaction
COMMIT;

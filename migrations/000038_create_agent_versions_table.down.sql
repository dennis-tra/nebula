-- Begin the transaction
BEGIN;

ALTER TABLE peers DROP CONSTRAINT fk_peers_agent_version_id;
ALTER TABLE peers DROP COLUMN agent_version_id;

ALTER TABLE visits DROP CONSTRAINT fk_visits_agent_version_id;
ALTER TABLE visits DROP COLUMN agent_version_id;

DROP TABLE agent_versions;

-- End the transaction
COMMIT;

BEGIN;

ALTER TABLE raw_visits
    DROP COLUMN agent_version_id;
ALTER TABLE raw_visits
    DROP COLUMN protocol_ids;

COMMIT;

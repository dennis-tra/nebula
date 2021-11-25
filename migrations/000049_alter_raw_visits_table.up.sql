BEGIN;

-- acts as a substitute for the agent_version column. If the crawler knows the ID of the found agent version
-- it can already provide it here. This means less work for the database.
ALTER TABLE raw_visits
    ADD COLUMN agent_version_id INT;

-- acts as a substitute for the protocols column. If the crawler knows the IDs of the found protocols
-- it can already provide it here. It's also possible to partially fill both columns.
ALTER TABLE raw_visits
    ADD COLUMN protocol_ids INT[];

COMMIT;

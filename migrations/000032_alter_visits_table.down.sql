-- Begin the transaction
BEGIN;

-- Put all current crawls tables aside
ALTER TABLE visits DROP COLUMN visit_started_at,  DROP COLUMN visit_ended_at;

-- End the transaction
COMMIT;

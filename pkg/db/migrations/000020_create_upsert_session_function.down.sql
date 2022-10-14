BEGIN;

DROP FUNCTION IF EXISTS upsert_session;
DROP FUNCTION IF EXISTS calc_max_failed_visits;
DROP FUNCTION IF EXISTS calc_next_visit;

COMMIT;

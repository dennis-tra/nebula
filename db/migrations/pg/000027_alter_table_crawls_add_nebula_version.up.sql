BEGIN;

ALTER TABLE crawls ADD COLUMN version TEXT;

UPDATE crawls SET version = 'undefined';

ALTER TABLE crawls ALTER COLUMN version SET NOT NULL;

COMMIT;
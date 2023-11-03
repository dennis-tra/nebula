BEGIN;

ALTER TABLE crawls ADD COLUMN remaining_peers INT;

COMMENT ON COLUMN crawls.remaining_peers IS 'The number of remaining peers in the crawl queue if the process was cancelled.';

UPDATE crawls
SET remaining_peers = 0
WHERE state = 'succeeded';

COMMIT;
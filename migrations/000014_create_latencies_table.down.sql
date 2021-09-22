-- Begin the transaction
BEGIN;

DROP TABLE IF EXISTS latencies;
DROP INDEX idx_latencies_peer_id;

-- End the transaction
COMMIT;

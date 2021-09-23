-- Begin the transaction
BEGIN;

-- Migrate array of multi addresses from peer to multi_addresses table.
INSERT INTO multi_addresses (maddr, updated_at, created_at)
SELECT DISTINCT unnest(multi_addresses), NOW(), NOW()
FROM peers;

-- From the migration above the association of peer to multi address was lost
-- The following query fills the peers_multi_addresses association table.
INSERT INTO peers_multi_addresses (peer_id, multi_address_id)
WITH peer_maddr_table(id, maddr) AS (SELECT id, unnest(multi_addresses) FROM peers)
SELECT p.id AS peer_id, m.id AS multi_address_id
FROM peer_maddr_table AS p
         INNER JOIN multi_addresses AS m ON p.maddr = m.maddr;

-- Drop multi addresses from peers table.
ALTER TABLE peers
    DROP COLUMN multi_addresses;
ALTER TABLE peers
    DROP COLUMN old_multi_addresses;

-- End the transaction
COMMIT;

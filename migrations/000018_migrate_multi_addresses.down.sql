-- Begin the transaction
BEGIN;

-- Add dropped columns
ALTER TABLE peers
    ADD COLUMN multi_addresses VARCHAR(255) ARRAY;
ALTER TABLE peers
    ADD COLUMN old_multi_addresses VARCHAR(255) ARRAY;

-- Migrate multi addresses back to peer
UPDATE peers
SET multi_addresses=subquery.maddrs
FROM (SELECT pma.peer_id AS "peer_id", array_agg(maddr) AS "maddrs"
      FROM multi_addresses ma
               INNER JOIN peers_x_multi_addresses pma ON pma.multi_address_id = ma.id
               INNER JOIN peers p ON p.id = pma.peer_id
      GROUP BY pma.peer_id) AS subquery
WHERE peers.id = subquery.peer_id;

-- Delete all rows from association table
DELETE
FROM peers_x_multi_addresses;

-- Delete all rows from multi_addresses table
DELETE
FROM multi_addresses;

-- End the transaction
COMMIT;

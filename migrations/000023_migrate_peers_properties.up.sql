-- Begin the transaction
BEGIN;

-- Migrate existing agent version of peer
INSERT INTO properties (property, value, updated_at, created_at)
SELECT 'agent_version', agent_version, NOW(), NOW()
FROM peers
ON CONFLICT DO NOTHING;

-- Migrate existing protocols of peer
INSERT INTO properties (property, value, updated_at, created_at)
SELECT DISTINCT 'protocol', unnest(protocol), NOW(), NOW()
FROM peers
ON CONFLICT DO NOTHING;

-- Create entries in association table
INSERT INTO peers_properties (peer_id, property_id)
SELECT peers.id, properties.id
FROM peers
         INNER JOIN properties ON peers.agent_version = properties.value AND 'agent_version' = properties.property;

-- Create entries in association table
INSERT INTO peers_properties (peer_id, property_id)
SELECT subquery.id, properties.id
FROM (select id, unnest(protocol) as protocol
      from peers) as subquery INNER JOIN properties on subquery.protocol = properties.value and 'protocol' = properties.property;

-- Drop peers columns
ALTER TABLE peers
    DROP COLUMN agent_version;
ALTER TABLE peers
    DROP COLUMN protocol;

-- End the transaction
COMMIT;

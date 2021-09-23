-- Begin the transaction
BEGIN;

ALTER TABLE peers
    ADD COLUMN agent_version VARCHAR(255),
    ADD COLUMN protocol      VARCHAR(255) ARRAY;

UPDATE peers
SET protocol=subquery.array_agg
FROM (SELECT peers_properties.peer_id, array_agg(properties.value)
      FROM properties
               INNER JOIN peers_properties ON peers_properties.property_id = properties.id
               INNER JOIN peers ON peers.id = peers_properties.peer_id
      WHERE properties.property = 'protocol'
      GROUP BY peers_properties.peer_id) AS subquery
WHERE peers.id = subquery.peer_id;


UPDATE peers
SET agent_version=subquery.value
FROM (SELECT peers_properties.peer_id, properties.value
      FROM properties
               INNER JOIN peers_properties ON peers_properties.property_id = properties.id
               INNER JOIN peers ON peers.id = peers_properties.peer_id
      WHERE properties.property = 'agent_version') AS subquery
WHERE peers.id = subquery.peer_id;

DELETE
FROM peers_properties;

-- End the transaction
COMMIT;

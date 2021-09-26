-- Begin the transaction
BEGIN;

-- agent_versions
CREATE TABLE agent_versions
(
    -- The ID for this agent version
    id            SERIAL,
    -- When was this agent version updated the last time, used to retrieve the ID after an upsert operation
    updated_at    TIMESTAMPTZ   NOT NULL,
    -- When was this agent version created
    created_at    TIMESTAMPTZ   NOT NULL,

    -- The property name
    agent_version VARCHAR(1000) NOT NULL,

    -- There should only be one protocol
    CONSTRAINT uq_agent_versions_agent_version UNIQUE (agent_version),

    PRIMARY KEY (id)
);

-- migrate agent_versions
INSERT INTO agent_versions (agent_version, updated_at, created_at)
SELECT value, p.updated_at, p.created_at
FROM properties p
WHERE p.property = 'agent_version';

-- add agent version column
ALTER TABLE visits
    ADD COLUMN agent_version_id INT;

-- migrate agent versions
UPDATE visits
SET agent_version_id = subquery.agent_version_id
FROM (SELECT visit_id, av.id as agent_version_id
      FROM visits_x_properties vxp
               INNER JOIN properties p ON p.id = vxp.property_id
               INNER JOIN agent_versions av ON p.value = av.agent_version
      WHERE p.property = 'agent_version') AS subquery
WHERE visits.id = subquery.visit_id;


ALTER TABLE visits
    ADD CONSTRAINT fk_visits_agent_version_id FOREIGN KEY (agent_version_id)
        REFERENCES agent_versions (id)
        ON DELETE NO ACTION;


-- Create a column on the visits table to associate a visit with a set of protocols
ALTER TABLE peers
    ADD COLUMN agent_version_id INT;

-- migrate agent versions
UPDATE peers
SET agent_version_id = subquery.agent_version_id
FROM (SELECT peer_id, av.id as agent_version_id
      FROM peers_x_properties vxp
               INNER JOIN properties p ON p.id = vxp.property_id
               INNER JOIN agent_versions av ON p.value = av.agent_version
      WHERE p.property = 'agent_version') AS subquery
WHERE peers.id = subquery.peer_id;

ALTER TABLE peers
    ADD CONSTRAINT fk_peers_agent_version_id FOREIGN KEY (agent_version_id)
        REFERENCES agent_versions (id)
        ON DELETE NO ACTION;

-- End the transaction
COMMIT;

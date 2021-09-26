-- Begin the transaction
BEGIN;

-- Activate intarray extension for efficient array operations
CREATE EXTENSION IF NOT EXISTS intarray;

-- The protocols table holds all the different protocols that the crawler came across
CREATE TABLE protocols
(
    -- The ID for this protocol
    id         SERIAL,
    -- When was this protocol updated the last time, used to retrieve the ID after an upsert operation
    updated_at TIMESTAMPTZ   NOT NULL,
    -- When was this protocol created
    created_at TIMESTAMPTZ   NOT NULL,

    -- The property name
    protocol   VARCHAR(1000) NOT NULL,

    -- There should only be one protocol
    CONSTRAINT uq_protocols_protocol UNIQUE (protocol),

    PRIMARY KEY (id)
);

-- migrate protocols from the properties table
INSERT INTO protocols (protocol, updated_at, created_at)
SELECT value, p.updated_at, p.created_at
FROM properties p
WHERE p.property = 'protocol';

-- Since the set of protocols for a particular peer doesn't change very often in between crawls. The
-- visits_x_properties table is blowing up quite quickly. This table holds particular sets of protocols
-- that peers support. Each visit is then linked to just one of these sets.
CREATE TABLE protocols_sets
(
    -- The ID for this set of properties
    id           SERIAL,
    -- The properties in this set
    protocol_ids INT ARRAY NOT NULL,

    -- Don't allow identical sets in the database
    EXCLUDE USING GIST(protocol_ids WITH =),

    PRIMARY KEY (id)
);

-- Allow efficient lookups of particular protocol sets.
CREATE INDEX idx_protocols_sets_protocol_ids on protocols_sets USING GIN (protocol_ids);

-- A temporary table for this transaction. This table holds all sets of protocols for each visit.
CREATE TEMP TABLE visits_agg_protocols ON COMMIT DROP AS (
    SELECT visit_id, uniq(sort(array_agg(prot.id))) as protocols_set
    FROM visits_x_properties vxp
             INNER JOIN properties p on p.id = vxp.property_id
             INNER JOIN protocols prot ON p.value = prot.protocol
    WHERE p.property = 'protocol'
    GROUP BY 1);

-- This temporary table holds all distinct protocol sets
CREATE TEMP TABLE distinct_visits_agg_protocols ON COMMIT DROP AS (
    SELECT DISTINCT protocols_set
    FROM visits_agg_protocols);

-- Save all the distinct sets in to the protocols_sets table
INSERT
INTO protocols_sets (protocol_ids)
SELECT distinct_visits_agg_protocols.protocols_set
FROM distinct_visits_agg_protocols;

-- Create a column on the visits table to associate a visit with a set of protocols
ALTER TABLE visits
    ADD COLUMN protocols_set_id INT;

-- For each visit in the visits_agg_protocols table find the associated protocol set
-- then set the protocols_set_id column to that set.
WITH visits_x_protocols_sets AS (
    SELECT ag.visit_id AS visit_id, ps.id AS protocols_set_id
    FROM protocols_sets ps
             INNER JOIN visits_agg_protocols ag ON ag.protocols_set = ps.protocol_ids
)
UPDATE visits
SET protocols_set_id = vxps.protocols_set_id
FROM visits_x_protocols_sets vxps
WHERE vxps.visit_id = visits.id;

ALTER TABLE visits ADD CONSTRAINT fk_visits_protocols_set_id FOREIGN KEY (protocols_set_id)
    REFERENCES protocols_sets (id)
    ON DELETE NO ACTION;


CREATE TEMP TABLE peers_agg_protocols ON COMMIT DROP AS (
    SELECT peer_id, uniq(sort(array_agg(prot.id))) as protocols_set
    FROM peers_x_properties vxp
             INNER JOIN properties p on p.id = vxp.property_id
             INNER JOIN protocols prot ON p.value = prot.protocol
    WHERE p.property = 'protocol'
    GROUP BY 1);

-- Create a column on the visits table to associate a visit with a set of protocols
ALTER TABLE peers
    ADD COLUMN protocols_set_id INT;

WITH peers_x_protocols_sets AS (
    SELECT ag.peer_id AS visit_id, ps.id AS protocols_set_id
    FROM protocols_sets ps
             INNER JOIN peers_agg_protocols ag ON ag.protocols_set = ps.protocol_ids
)
UPDATE peers
SET protocols_set_id = pxps.protocols_set_id
FROM peers_x_protocols_sets pxps
WHERE pxps.visit_id = peers.id;

ALTER TABLE peers ADD CONSTRAINT fk_peers_protocols_set_id FOREIGN KEY (protocols_set_id)
    REFERENCES protocols_sets (id)
    ON DELETE NO ACTION;

-- End the transaction
COMMIT;

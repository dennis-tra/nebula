-- Begin the transaction
BEGIN;

CREATE TABLE multi_addresses_sets
(
    -- The ID for this set of multi addresses
    id                SERIAL,

    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,

    -- The multi addresses in this set
    multi_address_ids INT ARRAY NOT NULL,

    -- Don't allow identical sets in the database
    EXCLUDE USING GIST(multi_address_ids gist__intbig_ops WITH =),

    PRIMARY KEY (id)
);

CREATE TEMP TABLE visits_agg_multi_addresses ON COMMIT DROP AS (
    SELECT visit_id, uniq(sort(array_agg(multi_address_id))) as multi_address_ids
    FROM visits_x_multi_addresses
    GROUP BY 1);

-- This temporary table holds all distinct protocol sets
CREATE TEMP TABLE distinct_visits_agg_protocols ON COMMIT DROP AS (
    SELECT DISTINCT multi_address_ids
    FROM visits_agg_multi_addresses);


-- Create distinct sets of multi_addresses from all visits
INSERT
INTO multi_addresses_sets (multi_address_ids, updated_at, created_at)
SELECT distinct_visits_agg_protocols.multi_address_ids, NOW(), NOW()
FROM distinct_visits_agg_protocols;


-- Create a column to associate a visit with a set of protocols
ALTER TABLE visits
    ADD COLUMN multi_addresses_set_id INT;

-- associate a visit with the protocol
WITH visit_x_multi_addresses_set AS (
    SELECT ag.visit_id AS visit_id, mas.id AS multi_addresses_set
    FROM multi_addresses_sets mas
             INNER JOIN visits_agg_multi_addresses ag ON ag.multi_address_ids = mas.multi_address_ids
)
UPDATE visits
SET multi_addresses_set_id = vxma.multi_addresses_set
FROM visit_x_multi_addresses_set vxma
WHERE vxma.visit_id = visits.id;

ALTER TABLE visits
    ADD CONSTRAINT fk_visits_multi_addresses_set_id FOREIGN KEY (multi_addresses_set_id)
        REFERENCES multi_addresses_sets (id)
        ON DELETE NO ACTION;

CREATE INDEX idx_multi_addresses_sets_multi_address_ids ON multi_addresses_sets USING GIN (multi_address_ids);

-- End the transaction
COMMIT;

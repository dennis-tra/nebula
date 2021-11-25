BEGIN;

-- Table: agent_versions
ALTER TABLE agent_versions
    ALTER COLUMN agent_version TYPE VARCHAR(1000);

-- Table: latencies
ALTER TABLE latencies
    ALTER COLUMN address TYPE VARCHAR(100);

-- Table: multi_addresses
ALTER TABLE multi_addresses
    ALTER COLUMN maddr TYPE VARCHAR(200);

-- Table: peers
ALTER TABLE peers
    ALTER COLUMN multi_hash TYPE VARCHAR(100);

-- Table: protocols
ALTER TABLE protocols
    ALTER COLUMN protocol TYPE VARCHAR(1000);

-- Table: raw_visits
ALTER TABLE raw_visits
    ALTER COLUMN agent_version TYPE VARCHAR(255);
ALTER TABLE raw_visits
    ALTER COLUMN peer_multi_hash TYPE VARCHAR(150);
ALTER TABLE raw_visits
    ALTER COLUMN protocols TYPE VARCHAR(255)[];
ALTER TABLE raw_visits
    ALTER COLUMN multi_addresses TYPE VARCHAR(255)[];

COMMIT;

BEGIN;

-- Table: agent_versions
ALTER TABLE agent_versions
    ALTER COLUMN agent_version TYPE TEXT;

-- Table: latencies
ALTER TABLE latencies
    ALTER COLUMN address TYPE TEXT;

-- Table: multi_addresses
ALTER TABLE multi_addresses
    ALTER COLUMN maddr TYPE TEXT;

-- Table: peers
ALTER TABLE peers
    ALTER COLUMN multi_hash TYPE TEXT;

-- Table: protocols
ALTER TABLE protocols
    ALTER COLUMN protocol TYPE TEXT;

-- Table: raw_visits
ALTER TABLE raw_visits
    ALTER COLUMN agent_version TYPE TEXT;
ALTER TABLE raw_visits
    ALTER COLUMN peer_multi_hash TYPE TEXT;
ALTER TABLE raw_visits
    ALTER COLUMN protocols TYPE TEXT[];
ALTER TABLE raw_visits
    ALTER COLUMN multi_addresses TYPE TEXT[];


COMMIT;

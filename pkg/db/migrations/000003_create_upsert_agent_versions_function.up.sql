BEGIN;

CREATE OR REPLACE FUNCTION upsert_agent_version(
    new_agent_version TEXT,
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_agent_version$
    WITH sel AS (
        SELECT id, agent_version
        FROM agent_versions
        WHERE agent_version = new_agent_version
    ), ups AS (
        INSERT INTO agent_versions (agent_version, created_at)
        SELECT new_agent_version, new_created_at
        WHERE NOT EXISTS (SELECT NULL FROM sel) AND new_agent_version IS NOT NULL
        ON CONFLICT ON CONSTRAINT uq_agent_versions_agent_version DO UPDATE
            SET agent_version = new_agent_version
        RETURNING id, agent_version
    )
    SELECT id FROM sel
    UNION ALL
    SELECT id FROM ups;
$upsert_agent_version$ LANGUAGE sql;

COMMENT ON FUNCTION upsert_agent_version IS 'Takes an agent version string and inserts it into the database if it does not exist. If it does exist it updates the last_seen_at field and also returns the ID.';

COMMIT;

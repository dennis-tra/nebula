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

COMMENT ON FUNCTION upsert_agent_version IS
    'Takes an agent version string and inserts it into the database if it does not exist.'
    'Returns its ID. The function tries to minimize the insert operations and only does'
    'them if the agent version does not already exist in the database. If there was an'
    'insert from another transaction in between it resorts to an upsert by overwriting'
    'the existing agent version with the same value. This is done so that the ID is returned.';

COMMIT;

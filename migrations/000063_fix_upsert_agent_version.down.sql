BEGIN;

CREATE OR REPLACE FUNCTION upsert_agent_version(
    new_agent_version TEXT,
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_agent_version$
DECLARE
    agent_version_id INT;
BEGIN
    IF new_agent_version IS NULL THEN
        RETURN NULL;
    END IF;

    WITH insert_agent_versions AS (
        SELECT new_agent_version_table new_av
        FROM unnest(('{' || new_agent_version || '}')::TEXT[]) new_agent_version_table
                 LEFT JOIN agent_versions av ON av.agent_version = new_agent_version_table
        WHERE av.id IS NULL
    )
    INSERT
    INTO agent_versions (agent_version, created_at, updated_at)
    SELECT insert_agent_versions.new_av, new_created_at, new_created_at
    FROM insert_agent_versions
    ON CONFLICT DO NOTHING;

    SELECT id FROM agent_versions av WHERE av.agent_version = new_agent_version INTO agent_version_id;

    RETURN agent_version_id;
END;
$upsert_agent_version$ LANGUAGE plpgsql;

COMMIT;

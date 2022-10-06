BEGIN;

CREATE OR REPLACE FUNCTION upsert_agent_version(
    new_agent_version TEXT
) RETURNS INT AS
$upsert_agent_version$
DECLARE
    agent_version_id  INT;
    agent_version_ret agent_versions%rowtype;
BEGIN
    new_agent_version = TRIM(new_agent_version);

    IF new_agent_version IS NULL OR new_agent_version = '' THEN
        RETURN NULL;
    END IF;

    SELECT *
    FROM agent_versions av
    WHERE av.agent_version = new_agent_version
    INTO agent_version_ret;

    IF agent_version_ret IS NULL THEN
        INSERT INTO agent_versions (agent_version, created_at)
        VALUES (new_agent_version, NOW())
        RETURNING id INTO agent_version_id;

        RETURN agent_version_id;
    END IF;

    RETURN agent_version_ret.id;
END
$upsert_agent_version$ LANGUAGE plpgsql;

COMMENT ON FUNCTION upsert_agent_version IS 'Takes an agent version string and inserts it into the database if it does not exist. If it does exist it updates the last_seen_at field and also returns the ID.';

COMMIT;

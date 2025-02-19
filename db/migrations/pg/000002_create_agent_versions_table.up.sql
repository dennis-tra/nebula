BEGIN;

-- Holds all discovered agent_versions
CREATE TABLE agent_versions
(
    -- A unique id that identifies a agent version.
    id            INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this agent version was seen the last time.
    created_at    TIMESTAMPTZ NOT NULL,
    -- Agent version string as reported from the remote peer.
    agent_version TEXT        NOT NULL CHECK ( TRIM(agent_version) != '' ),

    -- There should only be unique agent version strings in this table.
    CONSTRAINT uq_agent_versions_agent_version UNIQUE (agent_version),

    PRIMARY KEY (id)
);

COMMENT ON TABLE agent_versions IS 'Holds all discovered agent_versions';
COMMENT ON COLUMN agent_versions.id IS 'A unique id that identifies a agent version.';
COMMENT ON COLUMN agent_versions.created_at IS 'Timestamp of when this agent version was seen the last time.';
COMMENT ON COLUMN agent_versions.agent_version IS 'Agent version string as reported from the remote peer.';

COMMIT;

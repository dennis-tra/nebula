BEGIN;

-- The `peers` table keeps track of all peers ever found in the DHT
CREATE TABLE peers
(
    -- The peer ID in the form of Qm... or 12D3...
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The current agent version of the peer (updated if changed).
    agent_version_id INT,
    -- The set of protocols that this peer currently supports (updated if changed).
    protocols_set_id INT,
    -- The peer ID in the form of Qm... or 12D3...
    multi_hash       TEXT        NOT NULL,

    -- When were the multi addresses updated the last time.
    updated_at       TIMESTAMPTZ NOT NULL,
    -- When was this peer instance created.
    -- This gives a pretty accurate idea of
    -- when this peer was seen the first time.
    created_at       TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_peers_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,
    CONSTRAINT fk_peers_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL,

    -- There should only ever be distinct peer multi hash here
    CONSTRAINT uq_peers_multi_hash UNIQUE (multi_hash),

    PRIMARY KEY (id)
);

COMMIT;
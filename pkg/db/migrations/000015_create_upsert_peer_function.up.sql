BEGIN;

CREATE OR REPLACE FUNCTION upsert_peer(
    new_multi_hash TEXT,
    new_agent_version_id INT,
    new_protocols_set_id INT,
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_peer$
    WITH sel AS (
        SELECT id, multi_hash, agent_version_id, protocols_set_id
        FROM peers
        WHERE multi_hash = new_multi_hash
    ), ups AS (
        INSERT INTO peers AS p (multi_hash, agent_version_id, protocols_set_id, updated_at, created_at)
        SELECT new_multi_hash, new_agent_version_id, new_protocols_set_id, new_created_at, new_created_at
        WHERE NOT EXISTS (SELECT NULL FROM sel)
        ON CONFLICT ON CONSTRAINT uq_peers_multi_hash DO UPDATE
            SET multi_hash       = EXCLUDED.multi_hash,
                agent_version_id = coalesce(EXCLUDED.agent_version_id, p.agent_version_id),
                protocols_set_id = coalesce(EXCLUDED.protocols_set_id, p.protocols_set_id)
        RETURNING id, multi_hash
    ), upd AS (
        UPDATE peers
        SET agent_version_id = coalesce(new_agent_version_id, agent_version_id),
            protocols_set_id = coalesce(new_protocols_set_id, protocols_set_id),
            updated_at       = new_created_at
        WHERE id = (SELECT id FROM sel) AND (
            coalesce(agent_version_id, -1) != coalesce(new_agent_version_id, -1) OR
            coalesce(protocols_set_id, -1) != coalesce(new_protocols_set_id, -1)
        )
    )
    SELECT id FROM sel
    UNION ALL
    SELECT id FROM ups;
$upsert_peer$ LANGUAGE sql;

COMMIT;

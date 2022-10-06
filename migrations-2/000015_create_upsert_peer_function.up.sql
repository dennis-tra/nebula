BEGIN;

CREATE OR REPLACE FUNCTION upsert_peer(
    new_multi_hash TEXT,
    new_agent_version_id INT,
    new_protocols_set_id INT,
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_peer$
DECLARE
    peer_id INT;
    peer    peers%rowtype;
BEGIN
    SELECT *
    FROM peers p
    WHERE p.multi_hash = new_multi_hash
    INTO peer;

    IF peer IS NULL THEN
        INSERT INTO peers (multi_hash, agent_version_id, protocols_set_id, updated_at, created_at)
        VALUES (new_multi_hash, new_agent_version_id, new_protocols_set_id, new_created_at, new_created_at)
        RETURNING id INTO peer_id;

        RETURN peer_id;
    END IF;

    IF peer.agent_version_id != coalesce(new_agent_version_id, peer.agent_version_id) OR
       peer.protocols_set_id != coalesce(new_protocols_set_id, peer.protocols_set_id) THEN
        UPDATE peers
        SET agent_version_id = coalesce(new_agent_version_id, peer.agent_version_id),
            protocols_set_id = coalesce(new_protocols_set_id, peer.protocols_set_id),
            updated_at       = new_created_at
        WHERE id = peer.id;
    END IF;

    RETURN peer.id;
END;
$upsert_peer$ LANGUAGE plpgsql;

COMMIT;

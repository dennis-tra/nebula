BEGIN;

ALTER TABLE peers
    ADD COLUMN is_exposed BOOLEAN;

UPDATE peers
SET is_exposed = (properties ->> 'is_exposed')::BOOLEAN
WHERE properties ->> 'is_exposed' IS NOT NULL;

ALTER TABLE peers
    DROP COLUMN properties;

ALTER TABLE visits
    ADD COLUMN is_exposed BOOLEAN;

UPDATE visits
SET is_exposed = (peer_properties ->> 'is_exposed')::BOOLEAN
WHERE peer_properties ->> 'is_exposed' IS NOT NULL;

ALTER TABLE visits
    DROP COLUMN peer_properties;

DROP FUNCTION IF EXISTS upsert_peer;
CREATE OR REPLACE FUNCTION upsert_peer(
    new_multi_hash       TEXT,
    new_agent_version_id INT DEFAULT NULL,
    new_protocols_set_id INT DEFAULT NULL,
    new_is_exposed       BOOL DEFAULT NULL,
    new_created_at       TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_peer$
WITH sel AS (
    SELECT id, multi_hash, agent_version_id, protocols_set_id
    FROM peers
    WHERE multi_hash = new_multi_hash
), ups AS (
    INSERT INTO peers AS p (multi_hash, agent_version_id, protocols_set_id, is_exposed, updated_at, created_at)
        SELECT new_multi_hash, new_agent_version_id, new_protocols_set_id, new_is_exposed, new_created_at, new_created_at
        WHERE NOT EXISTS (SELECT NULL FROM sel)
        ON CONFLICT ON CONSTRAINT uq_peers_multi_hash DO UPDATE
            SET multi_hash       = EXCLUDED.multi_hash,
                agent_version_id = coalesce(EXCLUDED.agent_version_id, p.agent_version_id),
                protocols_set_id = coalesce(EXCLUDED.protocols_set_id, p.protocols_set_id),
                is_exposed       = coalesce(EXCLUDED.is_exposed, p.is_exposed)
        RETURNING id, multi_hash
), upd AS (
    UPDATE peers
        SET agent_version_id = coalesce(new_agent_version_id, agent_version_id),
            protocols_set_id = coalesce(new_protocols_set_id, protocols_set_id),
            is_exposed = coalesce(new_is_exposed, is_exposed),
            updated_at       = new_created_at
        WHERE id = (SELECT id FROM sel) AND (
                    coalesce(agent_version_id, -1) != coalesce(new_agent_version_id, -1) OR
                    coalesce(protocols_set_id, -1) != coalesce(new_protocols_set_id, -1) OR
                    (new_is_exposed IS NOT NULL AND is_exposed != new_is_exposed)
            )
        RETURNING peers.id
)
SELECT id FROM sel
UNION
SELECT id FROM ups
UNION
SELECT id FROM upd;
$upsert_peer$ LANGUAGE sql;

DROP TRIGGER on_peer_update ON peers;
DROP FUNCTION IF EXISTS insert_peer_log;
CREATE OR REPLACE FUNCTION insert_peer_log()
    RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.agent_version_id != NEW.agent_version_id THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'agent_version_id', OLD.agent_version_id, NEW.agent_version_id, NOW());
    END IF;

    IF OLD.protocols_set_id != NEW.protocols_set_id THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'protocols_set_id', OLD.protocols_set_id, NEW.protocols_set_id, NOW());
    END IF;

    IF OLD.is_exposed != NEW.is_exposed THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'is_exposed', OLD.is_exposed, NEW.is_exposed, NOW());
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER on_peer_update
    BEFORE UPDATE
    ON peers
    FOR EACH ROW
EXECUTE PROCEDURE insert_peer_log();


DROP FUNCTION IF EXISTS insert_visit;
CREATE OR REPLACE FUNCTION insert_visit(
    new_crawl_id INT,
    new_peer_multi_hash TEXT,
    new_multi_addresses TEXT[],
    new_agent_version_id INT,
    new_protocols_set_id INT,
    new_dial_duration INTERVAL,
    new_connect_duration INTERVAL,
    new_crawl_duration INTERVAL,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_type visit_type,
    new_connect_error net_error,
    new_crawl_error net_error,
    new_is_exposed BOOL
) RETURNS RECORD AS
$insert_visit$
DECLARE
    new_peer_id             INT;
    new_multi_addresses_ids INT[];
    new_session_id          INT;
    new_visit_id            INT;
BEGIN

    SELECT upsert_peer(new_peer_multi_hash, new_agent_version_id, new_protocols_set_id, new_is_exposed, new_visit_ended_at)
    INTO new_peer_id;

    SELECT array_agg(id) FROM upsert_multi_addresses(new_multi_addresses) INTO new_multi_addresses_ids;

    DELETE
    FROM peers_x_multi_addresses pxma
    WHERE peer_id = new_peer_id;

    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT new_peer_id, new_multi_address_id
    FROM unnest(new_multi_addresses_ids) new_multi_address_id
    ON CONFLICT DO NOTHING;

    SELECT upsert_session(new_peer_id, new_visit_started_at, new_visit_ended_at, new_connect_error) INTO new_session_id;

    -- Now we're able to create the normalized visit instance
    INSERT INTO visits (peer_id, crawl_id, session_id, dial_duration, connect_duration, crawl_duration,
                        visit_started_at, visit_ended_at, created_at, type, connect_error, crawl_error,
                        agent_version_id, protocols_set_id, multi_address_ids, is_exposed)
    SELECT new_peer_id,
           new_crawl_id,
           new_session_id,
           new_dial_duration,
           new_connect_duration,
           new_crawl_duration,
           new_visit_started_at,
           new_visit_ended_at,
           NOW(),
           new_type,
           new_connect_error,
           new_crawl_error,
           new_agent_version_id,
           new_protocols_set_id,
           new_multi_addresses_ids,
           new_is_exposed
    RETURNING id INTO new_visit_id;

    RETURN ROW(new_peer_id, new_visit_id, new_session_id);
END;
$insert_visit$ LANGUAGE plpgsql;

COMMIT;
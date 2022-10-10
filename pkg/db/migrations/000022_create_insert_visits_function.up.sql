BEGIN;

CREATE OR REPLACE FUNCTION insert_visit(
    new_crawl_id INT,
    new_peer_multi_hash TEXT,
    new_multi_addresses TEXT[],
    new_protocols TEXT[],
    new_protocol_ids INT[],
    new_agent_version TEXT,
    new_agent_version_id INT,
    new_dial_duration INTERVAL,
    new_connect_duration INTERVAL,
    new_crawl_duration INTERVAL,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_type visit_type,
    new_connect_error net_error,
    new_crawl_error net_error
) RETURNS INT AS
$insert_visit$
DECLARE
    new_visit_id int;
BEGIN

    WITH all_protocol_ids AS (
        SELECT id
        FROM upsert_protocols(new_protocols, new_visit_ended_at)
        UNION ALL
        SELECT id
        FROM unnest(new_protocol_ids) id
    ), upserted_protocol_ids AS (
        SELECT array_agg(DISTINCT id) ids
        FROM all_protocol_ids
        ORDER BY 1
    ), upserted_protocols_set_id AS (
        SELECT upsert_protocol_set_id(upi.ids,sha256(upi.ids::TEXT::BYTEA)) id
        FROM upserted_protocol_ids upi WHERE upi IS NOT NULL
    ), upserted_agent_version_id AS (
        SELECT coalesce(upsert_agent_version(new_agent_version, new_visit_ended_at), new_agent_version_id) id
    ), upserted_peer_id AS (
        SELECT upsert_peer(new_peer_multi_hash, (SELECT id FROM upserted_agent_version_id), (SELECT id FROM upserted_protocols_set_id), new_visit_ended_at) id
    ), upserted_multi_addresses AS (
        SELECT upsert_multi_addresses(new_multi_addresses) multi_address_id
    ), multi_address_diff_table AS (
        SELECT pxma.multi_address_id existing_id, uma.multi_address_id new_id
        FROM peers_x_multi_addresses pxma
            FULL OUTER JOIN upserted_multi_addresses uma ON uma.multi_address_id = pxma.multi_address_id AND peer_id = (SELECT id FROM upserted_peer_id)
    ), delete_multi_addresses AS (
        DELETE FROM peers_x_multi_addresses pxma
        WHERE peer_id = (SELECT id FROM upserted_peer_id)
            AND EXISTS (
                SELECT FROM multi_address_diff_table madt
                WHERE pxma.multi_address_id = madt.existing_id
                  AND madt.new_id IS NULL
            )
    ), insert_multi_addresses AS (
        INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
        SELECT (SELECT id FROM upserted_peer_id), madf.new_id
        FROM multi_address_diff_table madf
        WHERE madf.existing_id IS NULL
        ON CONFLICT DO NOTHING
    ), upsert_session AS (
        SELECT upsert_session((SELECT id FROM upserted_peer_id), new_visit_started_at, new_visit_ended_at, new_connect_error) id
    )

    -- Now we're able to create the normalized visit instance
    INSERT INTO visits (
        peer_id, crawl_id, session_id, dial_duration, connect_duration, crawl_duration, visit_started_at, visit_ended_at,
        created_at, type, connect_error, crawl_error, agent_version_id, protocols_set_id, multi_address_ids
    ) SELECT
        (SELECT id FROM upserted_peer_id), new_crawl_id, (SELECT id FROM upsert_session), new_dial_duration, new_connect_duration, new_crawl_duration,
        new_visit_started_at, new_visit_ended_at, NOW(), new_type, new_connect_error, new_crawl_error, (SELECT id FROM upserted_agent_version_id),
        (SELECT id FROM upserted_protocols_set_id), (SELECT array_agg(multi_address_id) FROM upserted_multi_addresses)
    RETURNING id INTO new_visit_id;

    RETURN new_visit_id;
END;
$insert_visit$ LANGUAGE plpgsql;

COMMIT;

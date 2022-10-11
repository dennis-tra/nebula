BEGIN;

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
    new_crawl_error net_error
) RETURNS INT AS
$insert_visit$
DECLARE
    new_peer_id             INT;
    new_multi_addresses_ids INT[];
    new_session_id          INT;
    new_visit_id            INT;
BEGIN

    SELECT upsert_peer(new_peer_multi_hash, new_agent_version_id, new_protocols_set_id, new_visit_ended_at)
    INTO new_peer_id;

    SELECT array_agg(id) FROM upsert_multi_addresses(new_multi_addresses) INTO new_multi_addresses_ids;

    DELETE
    FROM peers_x_multi_addresses pxma
    WHERE peer_id = new_peer_id;

    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT new_peer_id, new_multi_address_id
    FROM unnest(new_multi_addresses_ids) new_multi_address_id;

    SELECT upsert_session(new_peer_id, new_visit_started_at, new_visit_ended_at, new_connect_error) INTO new_session_id;

    -- Now we're able to create the normalized visit instance
    INSERT INTO visits (peer_id, crawl_id, session_id, dial_duration, connect_duration, crawl_duration,
                        visit_started_at, visit_ended_at,
                        created_at, type, connect_error, crawl_error, agent_version_id, protocols_set_id,
                        multi_address_ids)
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
           new_multi_addresses_ids
    RETURNING id INTO new_visit_id;

    RETURN new_visit_id;
END;
$insert_visit$ LANGUAGE plpgsql;

COMMIT;

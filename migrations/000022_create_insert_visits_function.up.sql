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
    new_error dial_error
) RETURNS INT AS
$insert_visit$
DECLARE
    upserted_protocol_ids      INT[];
    upserted_protocols_set_id  INT;
    upserted_peer_id           INT;
    upserted_multi_address_ids INT[];
    upserted_agent_version_id  INT;
    new_visit_id               int;
BEGIN

    SELECT upsert_protocols(new_protocols, new_visit_ended_at) INTO upserted_protocol_ids;
    SELECT upsert_protocol_set_id(new_protocol_ids || upserted_protocol_ids) INTO upserted_protocols_set_id;
    SELECT upsert_agent_version(new_agent_version, new_visit_ended_at) INTO upserted_agent_version_id;
    SELECT upsert_peer(new_peer_multi_hash, coalesce(upserted_agent_version_id, new_agent_version_id),
                       upserted_protocols_set_id, new_visit_ended_at)
    INTO upserted_peer_id;

    SELECT upsert_multi_addresses(new_multi_addresses) INTO upserted_multi_address_ids;

    -- Update current list of advertised multi addresses of remote peer.
    DELETE FROM peers_x_multi_addresses WHERE peer_id = upserted_peer_id;
    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT upserted_peer_id, ma.id
    FROM (SELECT unnest(upserted_multi_address_ids) id) ma
    ON CONFLICT DO NOTHING;

    SELECT upsert_session(upserted_peer_id, new_visit_ended_at, new_visit_started_at, new_error);

    -- Now we're able to create the normalized visit instance
    INSERT
    INTO visits (peer_id,
                 crawl_id,
                 dial_duration,
                 connect_duration,
                 crawl_duration,
                 visit_started_at,
                 visit_ended_at,
                 created_at,
                 type,
                 error,
                 agent_version_id,
                 protocols_set_id,
                 multi_address_ids)
    VALUES (upserted_peer_id,
            new_crawl_id,
            new_dial_duration,
            new_connect_duration,
            new_crawl_duration,
            new_visit_started_at,
            new_visit_ended_at,
            NOW(),
            new_type,
            new_error,
            coalesce(upserted_agent_version_id, new_agent_version_id),
            upserted_protocols_set_id,
            upserted_multi_address_ids)
    RETURNING id INTO new_visit_id;

    RETURN new_visit_id;
END;
$insert_visit$ LANGUAGE plpgsql;

COMMIT;

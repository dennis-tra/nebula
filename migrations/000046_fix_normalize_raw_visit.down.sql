-- Begin the transaction
BEGIN;

DROP TRIGGER insert_raw_visit ON raw_visits;
DROP FUNCTION normalize_raw_visit();

CREATE FUNCTION normalize_raw_visit() RETURNS TRIGGER AS
$normalize_raw_visit$
DECLARE
    upserted_peer_id                int;
    upserted_session_id             int;
    new_multi_addresses_ids         int[];
    upserted_multi_addresses_set_id int;
    new_protocols_ids               int[];
    upserted_protocols_set_id       int;
    upserted_agent_version_id       int;
BEGIN

    -- UPSERT PROTOCOLS
    WITH protocols_ids_table AS (
        -- upsert protocols
        INSERT INTO protocols (protocol, updated_at, created_at)
            SELECT unnest(NEW.protocols), NEW.created_at, NEW.created_at ORDER BY 1
            ON CONFLICT (protocol) DO UPDATE SET updated_at = excluded.updated_at
            RETURNING id)
    SELECT sort(array_agg(pi.id)) protocols_ids -- create sorted array of protocol IDs
    FROM protocols_ids_table pi
    INTO new_protocols_ids;

    -- Check if set of protocol IDs already exists
    SELECT id
    FROM protocols_sets ps
    WHERE ps.protocol_ids = new_protocols_ids
    INTO upserted_protocols_set_id;

    -- If the set of protocol IDs does not exist update it
    IF upserted_protocols_set_id IS NULL THEN
        INSERT
        INTO protocols_sets (protocol_ids) (SELECT new_protocols_ids WHERE new_protocols_ids IS NOT NULL)
        RETURNING id INTO upserted_protocols_set_id;
    END IF;

    -- UPSERT AGENT VERSION
    IF NEW.agent_version IS NOT NULL THEN
        INSERT INTO agent_versions (agent_version, created_at, updated_at)
        VALUES (NEW.agent_version, NOW(), NOW())
        ON CONFLICT (agent_version) DO UPDATE SET updated_at=excluded.updated_at
        RETURNING id INTO upserted_agent_version_id;
    END IF;

    -- Create a row in the peers table to track this peer ID and receive the internal DB id
    INSERT INTO peers (multi_hash, agent_version_id, protocols_set_id, updated_at, created_at)
    VALUES (NEW.peer_multi_hash, upserted_agent_version_id, upserted_protocols_set_id, NEW.created_at, NEW.created_at)
    ON CONFLICT (multi_hash) DO UPDATE SET updated_at=excluded.updated_at
    RETURNING id INTO upserted_peer_id;

    -- update the currently active session
    IF NEW.error IS NULL THEN
        INSERT INTO sessions (peer_id,
                              first_successful_dial,
                              last_successful_dial,
                              first_failed_dial,
                              next_dial_attempt,
                              successful_dials,
                              finished,
                              created_at,
                              updated_at)
        VALUES (upserted_peer_id, NOW(), NOW(), '1970-01-01', NOW() + '30s'::interval, 1, false, NOW(), NOW())
        ON CONFLICT ON CONSTRAINT uq_peer_id_first_failed_dial DO UPDATE
            SET last_successful_dial = EXCLUDED.last_successful_dial,
                successful_dials     = sessions.successful_dials + 1,
                updated_at           = EXCLUDED.updated_at,
                next_dial_attempt    =
                    CASE
                        WHEN 0.5 *
                             (EXCLUDED.last_successful_dial - sessions.first_successful_dial) <
                             '30s'::interval THEN
                                EXCLUDED.last_successful_dial + '30s'::interval
                        WHEN 0.5 *
                             (EXCLUDED.last_successful_dial - sessions.first_successful_dial) >
                             '15m'::interval THEN
                                EXCLUDED.last_successful_dial + '15m'::interval
                        ELSE
                                EXCLUDED.last_successful_dial +
                                0.5 *
                                (EXCLUDED.last_successful_dial - sessions.first_successful_dial)
                        END
        RETURNING id INTO upserted_session_id;
    ELSE
        UPDATE sessions
        SET first_failed_dial = NEW.visit_started_at,
            min_duration      = last_successful_dial - first_successful_dial,
            max_duration      = NOW() - first_successful_dial,
            finished          = true,
            updated_at        = NOW(),
            next_dial_attempt = null,
            finish_reason     = NEW.error
        WHERE peer_id = upserted_peer_id
          AND finished = false
        RETURNING id INTO upserted_session_id;
    END IF;

    -- UPSERT MULTI ADDRESSES

    WITH multi_addresses_ids_table AS (
        INSERT INTO multi_addresses (maddr, updated_at, created_at)
            SELECT DISTINCT unnest(NEW.multi_addresses), NEW.created_at, NEW.created_at -- the DISTINCT is the FIX
            ORDER BY 1
            ON CONFLICT (maddr) DO UPDATE SET updated_at = excluded.updated_at
            RETURNING id)
    SELECT sort(array_agg(mai.id)) multi_addresses_ids
    FROM multi_addresses_ids_table mai
    INTO new_multi_addresses_ids;

    IF new_multi_addresses_ids IS NOT NULL THEN
        SELECT id
        FROM multi_addresses_sets mas
        WHERE mas.multi_address_ids = new_multi_addresses_ids
        INTO upserted_multi_addresses_set_id;

        IF upserted_multi_addresses_set_id IS NULL THEN
            INSERT
            INTO multi_addresses_sets (multi_address_ids, updated_at, created_at) (SELECT new_multi_addresses_ids, NOW(), NOW())
            RETURNING id INTO upserted_multi_addresses_set_id;
        END IF;
    END IF;

    -- remove current association multi addresses and properties to peers
    DELETE FROM peers_x_multi_addresses WHERE peer_id = upserted_peer_id;

    -- Add multi address association
    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT upserted_peer_id, ma.id
    FROM multi_addresses ma
    WHERE maddr = ANY (NEW.multi_addresses);

    -- Now we're able to create the normalized visit instance
    INSERT
    INTO visits (peer_id,
                 crawl_id,
                 session_id,
                 dial_duration,
                 connect_duration,
                 crawl_duration,
                 visit_started_at,
                 visit_ended_at,
                 updated_at,
                 created_at,
                 type,
                 error,
                 agent_version_id,
                 protocols_set_id,
                 multi_addresses_set_id)
    VALUES (upserted_peer_id,
            NEW.crawl_id,
            upserted_session_id,
            NEW.dial_duration,
            NEW.connect_duration,
            NEW.crawl_duration,
            NEW.visit_started_at,
            NEW.visit_ended_at,
            NEW.created_at,
            NEW.created_at,
            NEW.type,
            NEW.error,
            upserted_agent_version_id,
            upserted_protocols_set_id,
            upserted_multi_addresses_set_id);

    DELETE FROM raw_visits WHERE id = NEW.id;

    RETURN NEW;
END;
$normalize_raw_visit$ LANGUAGE plpgsql;

CREATE TRIGGER "insert_raw_visit"
    AFTER INSERT
    ON "raw_visits"
    FOR EACH ROW
EXECUTE PROCEDURE normalize_raw_visit();


-- End the transaction
COMMIT;

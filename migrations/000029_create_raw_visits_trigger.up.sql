CREATE FUNCTION normalize_raw_visit() RETURNS TRIGGER AS
$normalize_raw_visit$
DECLARE
    inserted_peer_id    int;
    inserted_visit_id   int;
    upserted_session_id int;
BEGIN

    -- Create a row in the peers table to track this peer ID and receive the internal DB id
    INSERT INTO peers (multi_hash, updated_at, created_at)
    VALUES (NEW.peer_multi_hash, NEW.created_at, NEW.created_at)
    ON CONFLICT (multi_hash) DO UPDATE SET updated_at=excluded.updated_at
    RETURNING id INTO inserted_peer_id;

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
        VALUES (inserted_peer_id, NOW(), NOW(), '1970-01-01', NOW() + '30s'::interval, 1, false, NOW(), NOW())
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
        SET first_failed_dial = NEW.created_at,
            min_duration      = last_successful_dial - first_successful_dial,
            max_duration      = NOW() - first_successful_dial,
            finished          = true,
            updated_at        = NOW(),
            next_dial_attempt = null,
            finish_reason     = 'unknown'
        WHERE peer_id = inserted_peer_id
          AND finished = false
        RETURNING id INTO upserted_session_id;
    END IF;


    -- Now we're able to create the normalized visit instance
    INSERT
    INTO visits (peer_id, crawl_id, session_id, updated_at, created_at)
    VALUES (inserted_peer_id, NEW.crawl_id, upserted_session_id, NEW.created_at, NEW.created_at)
    RETURNING id INTO inserted_visit_id;

    -- First attempt to insert all properties into the properties table and retrieve the IDs
    -- of freshly created and already existing entries. Use these IDs to fill the visits_x_properties
    -- table and associate the properties with this visit.
    WITH all_visit_properties as (
        WITH visit_properties as (
            SELECT 'agent_version'   as protocol,
                   NEW.agent_version as val,
                   NEW.created_at    as updated_at,
                   NEW.created_at    as created_at
            UNION
            SELECT 'error'                 as protocol,
                   NEW.dial_error::varchar as val,
                   NEW.created_at          as updated_at,
                   NEW.created_at          as created_at
            WHERE NEW.dial_error IS NOT NULL
            UNION
            SELECT 'protocol'            as protocol,
                   unnest(NEW.protocols) as val,
                   NEW.created_at        as updated_at,
                   NEW.created_at        as created_at
            )
            INSERT INTO properties (property, value, updated_at, created_at)
                SELECT vp.protocol, vp.val, vp.updated_at, vp.created_at FROM visit_properties vp ORDER BY protocol, val
                ON CONFLICT (property, value) DO UPDATE SET updated_at = excluded.updated_at
                RETURNING id)
    INSERT
    INTO visits_x_properties (visit_id, property_id)
    SELECT inserted_visit_id, p2.id
    FROM all_visit_properties AS p2;

    -- Take the multi addresses of the peer and insert them into the association table
    WITH multi_addresses_id_table AS (INSERT INTO multi_addresses (maddr, updated_at, created_at)
        VALUES (unnest(NEW.multi_addresses), NEW.created_at, NEW.created_at)
        ON CONFLICT (maddr) DO UPDATE SET updated_at = excluded.updated_at
        RETURNING id, maddr)
    INSERT
    INTO visits_x_multi_addresses (visit_id, multi_address_id)
    SELECT inserted_visit_id, p.id
    FROM multi_addresses_id_table AS p
    ORDER BY maddr; -- Order to prevent dead lock

    RETURN NEW;
END;
$normalize_raw_visit$ LANGUAGE plpgsql;

CREATE TRIGGER "insert_raw_visit"
    AFTER INSERT
    ON "raw_visits"
    FOR EACH ROW
EXECUTE PROCEDURE normalize_raw_visit();

-- Begin the transaction
BEGIN;

DROP TRIGGER insert_raw_visit ON raw_visits;
DROP FUNCTION normalize_raw_visit();


CREATE FUNCTION normalize_raw_visit() RETURNS TRIGGER AS
$normalize_raw_visit$
DECLARE
    upserted_peer_id    int;
    inserted_visit_id   int;
    upserted_session_id int;
BEGIN

    -- Create a row in the peers table to track this peer ID and receive the internal DB id
    INSERT INTO peers (multi_hash, updated_at, created_at)
    VALUES (NEW.peer_multi_hash, NEW.created_at, NEW.created_at)
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

    -- Now we're able to create the normalized visit instance
    INSERT
    INTO visits (peer_id, crawl_id, session_id, dial_duration, connect_duration, crawl_duration,
                 visit_started_at,
                 visit_ended_at,
                 updated_at, created_at,
                 type, error)
    VALUES (upserted_peer_id, NEW.crawl_id, upserted_session_id, NEW.dial_duration, NEW.connect_duration,
            NEW.crawl_duration, NEW.visit_started_at, NEW.visit_ended_at, NEW.created_at, NEW.created_at, NEW.type,
            NEW.error)
    RETURNING id INTO inserted_visit_id;

    -- Take the multi addresses of the peer and insert them into the association table
    WITH multi_addresses_id_table AS (INSERT INTO multi_addresses (maddr, updated_at, created_at)
        VALUES (unnest(NEW.multi_addresses), NEW.created_at, NEW.created_at)
        ON CONFLICT (maddr) DO UPDATE SET updated_at = excluded.updated_at
        RETURNING id, maddr)
    INSERT
    INTO visits_x_multi_addresses (visit_id, multi_address_id)
    SELECT inserted_visit_id, p.id
    FROM multi_addresses_id_table AS p
    ORDER BY maddr;
    -- Order to prevent dead lock

    -- If this visit instance is a dial then we didn't gather any new
    -- multi address, protocol, agent information, so exit early
    IF NEW.type = 'dial' THEN
        RETURN NEW;
    END IF;

    -- First attempt to insert all properties into the properties table and retrieve the IDs
    -- of freshly created and already existing entries. Use these IDs to fill the visits_x_properties
    -- table and associate the properties with this visit.
    WITH visit_properties as (
        WITH peer_properties as (
            SELECT 'agent_version'   as property,
                   NEW.agent_version as val,
                   NEW.created_at    as updated_at,
                   NEW.created_at    as created_at
            WHERE NEW.agent_version IS NOT NULL
            UNION
            SELECT 'error'            as property,
                   NEW.error::varchar as val,
                   NEW.created_at     as updated_at,
                   NEW.created_at     as created_at
            WHERE NEW.error IS NOT NULL
            UNION
            SELECT 'protocol'            as property,
                   unnest(NEW.protocols) as val,
                   NEW.created_at        as updated_at,
                   NEW.created_at        as created_at
            )
            INSERT INTO properties (property, value, updated_at, created_at)
                SELECT vp.property, vp.val, vp.updated_at, vp.created_at FROM peer_properties vp ORDER BY property, val
                ON CONFLICT (property, value) DO UPDATE SET updated_at = excluded.updated_at
                RETURNING id)
    INSERT
    INTO visits_x_properties (visit_id, property_id)
    SELECT inserted_visit_id, p2.id
    FROM visit_properties AS p2;

    -- remove current association multi addresses and properties to peers
    DELETE FROM peers_x_multi_addresses WHERE peer_id = upserted_peer_id;
    DELETE FROM peers_x_properties WHERE peer_id = upserted_peer_id;

    -- Add multi address association
    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT upserted_peer_id, ma.id
    FROM multi_addresses ma
    WHERE maddr = ANY (NEW.multi_addresses);

    -- Add properties association
    WITH peer_properties as (
        SELECT 'agent_version'   as property,
               NEW.agent_version as val
        UNION
        SELECT 'error'            as property,
               NEW.error::varchar as val
        WHERE NEW.error IS NOT NULL
        UNION
        SELECT 'protocol'            as property,
               unnest(NEW.protocols) as val
    )
    INSERT
    INTO peers_x_properties (peer_id, property_id)
    SELECT upserted_peer_id, p.id
    FROM peer_properties pp
             INNER JOIN properties p ON p.property = pp.property AND p.value = pp.val
    ORDER BY p.property, p.value;


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

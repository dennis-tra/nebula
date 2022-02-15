BEGIN;

CREATE TABLE raw_visits
(
    id               SERIAL,
    crawl_id         INT,
    visit_started_at TIMESTAMPTZ  NOT NULL,
    visit_ended_at   TIMESTAMPTZ  NOT NULL,
    dial_duration    INTERVAL,
    connect_duration INTERVAL,
    crawl_duration   INTERVAL,
    type             visit_type   NOT NULL,
    agent_version    VARCHAR(255),
    peer_multi_hash  VARCHAR(150) NOT NULL,
    protocols        VARCHAR(255) ARRAY,
    multi_addresses  VARCHAR(255) ARRAY,
    error            dial_error,
    error_message    TEXT,
    created_at       TIMESTAMPTZ  NOT NULL,
    agent_version_id INT,
    protocol_ids     INT[],

    PRIMARY KEY (id)
);

CREATE OR REPLACE FUNCTION normalize_raw_visit() RETURNS TRIGGER AS
$normalize_raw_visit$
DECLARE
    upserted_protocol_ids           INT[];
    upserted_peer_id                INT;
    upserted_session_id             INT;
    new_multi_addresses_ids         INT[];
    upserted_multi_addresses_set_id INT;
    upserted_protocols_set_id       INT;
    upserted_agent_version_id       INT;
BEGIN

    SELECT upsert_protocols(NEW.protocols, NEW.created_at) INTO upserted_protocol_ids;
    SELECT upsert_protocol_set_id(NEW.protocol_ids || upserted_protocol_ids) INTO upserted_protocols_set_id;
    SELECT upsert_agent_version(NEW.agent_version, NEW.created_at) INTO upserted_agent_version_id;
    SELECT upsert_peer(NEW.peer_multi_hash, coalesce(upserted_agent_version_id, NEW.agent_version_id),
                       upserted_protocols_set_id, NEW.created_at)
    INTO upserted_peer_id;

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
        VALUES (upserted_peer_id, NEW.visit_ended_at, NEW.visit_ended_at, '1970-01-01',
                NEW.visit_ended_at + '30s'::interval, 1, false, NOW(), NOW())
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
            max_duration      = NEW.visit_started_at - first_successful_dial,
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
            coalesce(upserted_agent_version_id, NEW.agent_version_id),
            upserted_protocols_set_id,
            upserted_multi_addresses_set_id);

    DELETE FROM raw_visits WHERE id = NEW.id;


    RETURN NEW;
END;
$normalize_raw_visit$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS insert_raw_visit ON raw_visits;

CREATE TRIGGER "insert_raw_visit"
    AFTER INSERT
    ON "raw_visits"
    FOR EACH ROW
EXECUTE PROCEDURE normalize_raw_visit();

DROP FUNCTION IF EXISTS insert_visit;

COMMIT;

CREATE OR REPLACE FUNCTION upsert_session(
    visit_peer_id INT,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_error dial_error
)
    RETURNS INT AS
$$
DECLARE
    upserted_session_id INT;
BEGIN
    IF new_error IS NULL THEN
        INSERT INTO sessions (peer_id,
                              first_successful_dial,
                              last_successful_dial,
                              first_failed_dial,
                              next_dial_attempt,
                              successful_dials,
                              finished,
                              created_at,
                              updated_at)
        VALUES (visit_peer_id, new_visit_ended_at, new_visit_ended_at, '1970-01-01',
                new_visit_ended_at + '30s'::interval, 1, false, NOW(), NOW())
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
        SET first_failed_dial = new_visit_started_at,
            min_duration      = last_successful_dial - first_successful_dial,
            max_duration      = new_visit_started_at - first_successful_dial,
            finished          = true,
            updated_at        = NOW(),
            next_dial_attempt = null,
            finish_reason     = new_error
        WHERE peer_id = visit_peer_id
          AND finished = false
        RETURNING id INTO upserted_session_id;
    END IF;

    RETURN upserted_session_id;
END;
$$ LANGUAGE 'plpgsql';

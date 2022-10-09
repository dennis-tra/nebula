BEGIN;

CREATE OR REPLACE FUNCTION calc_next_visit(
    new_visited_at TIMESTAMPTZ,
    last_visited_at TIMESTAMPTZ DEFAULT NULL,
    factor FLOAT DEFAULT 1.2,
    min_interval INTERVAL DEFAULT '1m'::INTERVAL,
    max_interval INTERVAL DEFAULT '15m'::INTERVAL
)
    RETURNS TIMESTAMPTZ AS
$$
BEGIN
    RETURN new_visited_at + LEAST(max_interval, GREATEST(min_interval, factor * (new_visited_at - last_visited_at)));
END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;


CREATE OR REPLACE FUNCTION calc_max_failed_visits(
    first_successful_visit TIMESTAMPTZ,
    last_successful_visit TIMESTAMPTZ
)
    RETURNS INT AS
$$
DECLARE
    uptime INTERVAL;
BEGIN
    SELECT last_successful_visit - first_successful_visit INTO uptime;
    IF uptime < '1h'::INTERVAL THEN
        RETURN 0;
    ELSIF uptime < '6h'::INTERVAL THEN
        RETURN 1;
    ELSIF uptime < '24h'::INTERVAL THEN
        RETURN 2;
    ELSE
        RETURN 3;
    END IF;
END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;

CREATE OR REPLACE FUNCTION upsert_session(
    visit_peer_id INT,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_error dial_error
)
    RETURNS INT AS
$$
DECLARE
    max_failed_visits   INT;
    upserted_session_id INT;
    upserted_session    sessions%rowtype;
BEGIN
    SELECT *
    FROM sessions_open
    WHERE peer_id = visit_peer_id
    INTO upserted_session;

    IF upserted_session IS NULL THEN
        IF new_error IS NULL THEN
            -- If there was no session object in the database but this
            -- visit was successful we create a new open sessions.
            INSERT INTO sessions_open (--
                peer_id, first_successful_visit, last_successful_visit, next_visit_due_at, updated_at,
                created_at,
                successful_visits_count, state, recovered_count,
                failed_visits_count)
            VALUES (visit_peer_id, new_visit_started_at, new_visit_ended_at,
                    (SELECT calc_next_visit(new_visit_ended_at)),
                    NOW(),
                    NOW(), 1, 'open', 0, 0)
            RETURNING id INTO upserted_session_id;
            RETURN upserted_session_id;
        ELSE
            -- If there is no open session object in the database and
            -- this visit yielded an error there is no point in
            -- opening a session.
            RETURN NULL;
        END IF;
    END IF;

    IF new_error IS NULL THEN
        -- So we found an open session in the database and could
        -- again successfully connect to the peer. Update it:
        UPDATE sessions_open
        SET state                    = 'open',                                    -- if the state was `pending` previously, we need to set it back to open
            last_successful_visit    = new_visit_ended_at,
            successful_visits_count  = upserted_session.successful_visits_count + 1,
            updated_at               = NOW(),
            first_failed_visit       = NULL,
            failed_visits_count      = 0,
            finish_reason            = NULL,
            recovered_count          = upserted_session.recovered_count +
                                       (upserted_session.state = 'pending')::INT, -- if the state was `pending` this will yield `1` and thus increment the recovered_count
            next_visit_due_at = (SELECT calc_next_visit(new_visit_ended_at,
                                                               upserted_session.last_successful_visit))
        WHERE peer_id = visit_peer_id;
    ELSE
        SELECT calc_max_failed_visits(upserted_session.first_successful_visit, upserted_session.last_successful_visit)
        INTO max_failed_visits;

        -- So we found an open session in the database but could
        -- not connect to the remote peer. Update it:
        IF upserted_session.state = 'open' AND max_failed_visits > 0 THEN
            UPDATE sessions
            SET state                    = 'pending',
                first_failed_visit       = new_visit_ended_at,
                last_failed_visit        = new_visit_ended_at,
                failed_visits_count      = upserted_session.failed_visits_count + 1,
                updated_at               = NOW(),
                finish_reason            = new_error,
                next_visit_due_at = new_visit_ended_at + max_failed_visits * '1m'::INTERVAL
            WHERE peer_id = visit_peer_id
              AND state = 'open';
        ELSIF upserted_session.state = 'pending' AND upserted_session.failed_visits_count < max_failed_visits THEN
            UPDATE sessions
            SET last_failed_visit        = new_visit_ended_at,
                failed_visits_count      = upserted_session.failed_visits_count + 1,
                updated_at               = NOW(),
                next_visit_due_at = (SELECT calc_next_visit(new_visit_ended_at,
                                                                   upserted_session.last_successful_visit))
            WHERE peer_id = visit_peer_id
              AND state = 'pending';
        ELSE
            UPDATE sessions
            SET state                    = 'closed',
                first_failed_visit       = COALESCE(upserted_session.first_failed_visit, new_visit_ended_at),
                last_failed_visit        = new_visit_ended_at,
                failed_visits_count      = upserted_session.failed_visits_count + 1,
                min_duration             = last_successful_visit - first_successful_visit,
                max_duration             = COALESCE(upserted_session.first_failed_visit, new_visit_ended_at) -
                                           first_successful_visit,
                updated_at               = NOW(),
                finish_reason            = COALESCE(upserted_session.finish_reason, new_error),
                next_visit_due_at = NULL
            WHERE peer_id = visit_peer_id
              AND state != 'closed';
        END IF;
    END IF;

    RETURN upserted_session.id;
END;
$$ LANGUAGE 'plpgsql';


COMMIT;

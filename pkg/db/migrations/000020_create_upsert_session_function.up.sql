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
$$ LANGUAGE 'plpgsql' ;


CREATE OR REPLACE FUNCTION calc_max_failed_visits(
    first_successful_visit TIMESTAMPTZ,
    last_successful_visit TIMESTAMPTZ,
    error net_error
)
    RETURNS INT AS
$$
DECLARE
    uptime INTERVAL;
BEGIN
    SELECT last_successful_visit - first_successful_visit INTO uptime;
    IF error = 'no_good_addresses' OR error = 'no_public_ip' OR error = 'no_route_to_host' OR error = 'peer_id_mismatch' THEN
        RETURN 0;
    ELSIF uptime < '1h'::INTERVAL THEN
        RETURN 0;
    ELSIF uptime < '6h'::INTERVAL THEN
        RETURN 1;
    ELSIF uptime < '24h'::INTERVAL THEN
        RETURN 2;
    ELSE
        RETURN 3;
    END IF;
END;
$$ LANGUAGE 'plpgsql' ;


CREATE OR REPLACE FUNCTION upsert_session(
    visit_peer_id INT,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_error net_error
) RETURNS INT AS
$upsert_session$
    WITH existing_session AS (
        SELECT *
        FROM sessions_open
        WHERE peer_id = visit_peer_id
    ), max_failed_visits AS (
        SELECT es.id, calc_max_failed_visits(es.first_successful_visit, es.last_successful_visit, new_error) max_visits
        FROM existing_session AS es
    ), new_session AS (
        INSERT INTO sessions_open (
            peer_id, first_successful_visit, last_successful_visit, last_visited_at, next_visit_due_at, updated_at,
            created_at, successful_visits_count, state, recovered_count, failed_visits_count, uptime)
        SELECT visit_peer_id, new_visit_started_at, new_visit_ended_at, new_visit_ended_at, (SELECT calc_next_visit(new_visit_ended_at)),
                NOW(), NOW(), 1, 'open', 0, 0, TSTZRANGE(new_visit_started_at, NULL)
        WHERE NOT EXISTS (SELECT NULL FROM existing_session) AND new_error IS NULL
        RETURNING id
    ), update_session_no_error AS (
        UPDATE sessions_open AS so
        SET state                    = 'open', -- if the state was `pending` previously, we need to set it back to open
            last_successful_visit    = new_visit_ended_at,
            last_visited_at          = new_visit_ended_at,
            successful_visits_count  = es.successful_visits_count + 1,
            updated_at               = NOW(),
            first_failed_visit       = NULL,
            last_failed_visit        = NULL,
            failed_visits_count      = 0,
            finish_reason            = NULL,
            uptime                   = TSTZRANGE(es.first_successful_visit, new_visit_ended_at),
            recovered_count          = es.recovered_count + (es.state = 'pending')::INT, -- if the state was `pending` this will yield `1` and thus increment the recovered_count
            next_visit_due_at        = (SELECT calc_next_visit(new_visit_ended_at, es.last_successful_visit))
        FROM existing_session AS es
        WHERE so.id = es.id AND new_error IS NULL
        RETURNING so.id
    ), update_open_session_error AS (
        UPDATE sessions_open AS so
        SET state                    = 'pending',
            first_failed_visit       = new_visit_ended_at,
            last_failed_visit        = new_visit_ended_at,
            last_visited_at          = new_visit_ended_at,
            failed_visits_count      = es.failed_visits_count + 1,
            updated_at               = NOW(),
            finish_reason            = new_error,
            next_visit_due_at        = new_visit_ended_at + mfv.max_visits * '1m'::INTERVAL
        FROM existing_session AS es INNER JOIN max_failed_visits AS mfv USING (id)
        WHERE so.id = es.id AND es.state = 'open' AND new_error IS NOT NULL AND mfv.max_visits > 0
        RETURNING so.id
    ), update_pending_session_error AS (
        UPDATE sessions_open AS so
        SET last_failed_visit        = new_visit_ended_at,
            last_visited_at          = new_visit_ended_at,
            failed_visits_count      = es.failed_visits_count + 1,
            updated_at               = NOW(),
            next_visit_due_at        = (SELECT calc_next_visit(new_visit_ended_at, es.last_successful_visit))
        FROM existing_session AS es INNER JOIN max_failed_visits AS mfv USING (id)
        WHERE so.id = es.id AND es.state = 'pending' AND new_error IS NOT NULL AND es.failed_visits_count < mfv.max_visits
        RETURNING so.id
    ), close_session_error AS (
        UPDATE sessions AS s
        SET state                    = 'closed',
            first_failed_visit       = COALESCE(es.first_failed_visit, new_visit_ended_at),
            last_failed_visit        = new_visit_ended_at,
            last_visited_at          = new_visit_ended_at,
            failed_visits_count      = es.failed_visits_count + 1,
            uptime                   = TSTZRANGE(lower(es.uptime), es.last_successful_visit),
            updated_at               = NOW(),
            finish_reason            = COALESCE(es.finish_reason, new_error),
            next_visit_due_at        = NULL
        FROM existing_session AS es INNER JOIN max_failed_visits AS mfv USING (id)
        WHERE s.id = es.id AND es.state != 'closed' AND new_error IS NOT NULL
            AND NOT EXISTS (SELECT NULL FROM update_session_no_error)
            AND NOT EXISTS (SELECT NULL FROM update_open_session_error)
            AND NOT EXISTS (SELECT NULL FROM update_pending_session_error)
        RETURNING s.id
    )
    SELECT id FROM existing_session
    UNION
    SELECT id FROM new_session
    UNION
    SELECT id FROM close_session_error;
$upsert_session$ LANGUAGE 'sql';

COMMIT;

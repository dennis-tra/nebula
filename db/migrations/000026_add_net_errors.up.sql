BEGIN;

ALTER TYPE net_error ADD VALUE 'connection_reset_by_peer';
ALTER TYPE net_error ADD VALUE 'cant_assign_requested_address';
ALTER TYPE net_error ADD VALUE 'connection_gated';
ALTER TYPE net_error ADD VALUE 'cant_connect_over_relay';
ALTER TYPE net_error RENAME VALUE 'no_public_ip' TO 'no_ip_address';

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
    IF error = 'no_good_addresses' OR error = 'no_ip_address' OR error = 'no_route_to_host' OR error = 'peer_id_mismatch' THEN
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

COMMIT
-- Define error types
CREATE TYPE dial_error AS ENUM (
    'io_timeout',
    'connection_refused',
    'protocol_not_supported',
    'peer_id_mismatch',
    'no_route_to_host',
    'network_unreachable',
    'no_good_addresses',
    'context_deadline_exceeded',
    'no_public_ip',
    'max_dial_attempts_exceeded',
    'unknown'
    );

-- Add finish reason column to session table
ALTER TABLE sessions
    ADD finish_reason dial_error;

-- Set all finish reason to 'unknown' for all finished sessions
UPDATE sessions
SET finish_reason = 'unknown'
WHERE finished = true;

-- Add a constraint that if a session is finished the reason can't be null.
-- If the session is not finished the reason must be null.
ALTER TABLE sessions
    ADD CONSTRAINT con_finish_reason_not_null_for_finished CHECK (
            (finished = TRUE AND finish_reason IS NOT NULL)
            OR
            (finished = FALSE AND finish_reason IS NULL)
        );

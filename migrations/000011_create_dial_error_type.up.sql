BEGIN;

-- The different types of errors that can happen when trying to dial a remote peer.
CREATE TYPE dial_error AS ENUM (
    'unknown',
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
    'maddr_reset',
    'stream_reset',
    'host_is_down',
    'negotiate_security_protocol_no_trailing_new_line'
    );

COMMENT ON TYPE dial_error IS 'The different types of errors that can happen when trying to dial a remote peer.';

COMMIT;

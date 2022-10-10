BEGIN;

-- The different types of errors that can happen when trying to dial or crawl a remote peer.
CREATE TYPE net_error AS ENUM (
    'unknown',
    'io_timeout',
    'no_recent_network_activity',
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
    'negotiate_security_protocol',
    'negotiate_stream_multiplexer'
    );

COMMENT ON TYPE net_error IS 'The different types of errors that can happen when trying to dial or crawl a remote peer.';

COMMIT;



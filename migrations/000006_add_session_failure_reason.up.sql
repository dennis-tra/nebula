CREATE TYPE failure_type AS ENUM (
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

ALTER TABLE sessions ADD failure_reason failure_type;

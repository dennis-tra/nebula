BEGIN;

ALTER TYPE net_error ADD VALUE 'connection_reset_by_peer';
ALTER TYPE net_error ADD VALUE 'cant_assign_requested_address';
ALTER TYPE net_error ADD VALUE 'connection_gated';
ALTER TYPE net_error RENAME VALUE 'no_public_ip' TO 'no_ip_address';

COMMIT
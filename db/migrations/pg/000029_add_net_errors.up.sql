BEGIN;

ALTER TYPE net_error ADD VALUE 'black_hole_refused';
ALTER TYPE net_error ADD VALUE 'no_transport_for_protocol';

COMMIT
BEGIN;

-- Activate intarray extension for efficient array operations
CREATE EXTENSION IF NOT EXISTS intarray;

-- Since the set of protocols for a particular peer does not change very often in between crawls,
-- this table holds particular sets of protocols which other tables can reference and save space.
CREATE TABLE protocols_sets
(
    -- An internal unique id that identifies a unique set of protocols.
    -- We could also just use the hash below but since protocol sets are
    -- referenced many times having just a 4 byte instead of 32 byte ID
    -- can make huge storage difference.
    id           INT GENERATED ALWAYS AS IDENTITY,
    -- The protocol IDs of this protocol set. The IDs reference the protocols table (no foreign key checks).
    -- Note: there's an invariant regarding the INT type. Don't increase it to BIGINT without changing protocolsSetHash.
    protocol_ids INT[] NOT NULL CHECK ( array_length(protocol_ids, 1) IS NOT NULL ),
    -- The hash digest of the sorted protocol ids to allow a unique constraint
    hash         BYTEA NOT NULL,

    CONSTRAINT uq_protocols_sets_hash UNIQUE (hash),

    PRIMARY KEY (id)
);

CREATE INDEX idx_protocols_sets_protocol_ids on protocols_sets USING GIST (protocol_ids);

COMMENT ON TABLE protocols_sets IS ''
    'Since the set of protocols for a particular peer does not change very often in between crawls,'
    'this table holds particular sets of protocols which other tables can reference and save space.';
COMMENT ON COLUMN protocols_sets.id IS 'An internal unique id that identifies a unique set of protocols.';
COMMENT ON COLUMN protocols_sets.protocol_ids IS 'The protocol IDs of this protocol set. The IDs reference the protocols table (no foreign key checks).';
COMMENT ON COLUMN protocols_sets.hash IS 'The hash digest of the sorted protocol ids to allow a unique constraint.';

COMMIT;

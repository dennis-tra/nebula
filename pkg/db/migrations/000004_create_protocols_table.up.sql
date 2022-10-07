BEGIN;

-- Holds all the different protocols that the crawler came across
CREATE TABLE protocols
(
    -- A unique id that identifies a agent version.
    id         INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this protocol was seen the last time.
    created_at TIMESTAMPTZ NOT NULL,
    -- The full protocol string.
    protocol   TEXT        NOT NULL,

    -- There should only be unique protocol strings in this table
    CONSTRAINT uq_protocols_protocol UNIQUE (protocol),

    PRIMARY KEY (id)
);

COMMENT ON TABLE protocols IS 'Holds all the different protocols that the crawler came across.';
COMMENT ON COLUMN protocols.id IS 'A unique id that identifies a agent version.';
COMMENT ON COLUMN protocols.created_at IS 'Timestamp of when this protocol was seen the last time.';
COMMENT ON COLUMN protocols.protocol IS 'The full protocol string.';

COMMIT;

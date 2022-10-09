BEGIN;

CREATE OR REPLACE FUNCTION upsert_protocols(
    new_protocols TEXT[],
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS TABLE (id INT) AS
$upsert_protocols$
    WITH input AS (
        SELECT DISTINCT unnest AS protocol
        FROM UNNEST(new_protocols) unnest
    ), sel AS (-- select all existing protocols
        SELECT protocols.id, protocols.protocol
        FROM input
        INNER JOIN protocols USING (protocol)
    ), ups AS (-- upsert all protocols that don't exist yet
        INSERT INTO protocols (protocol, created_at)
        SELECT input.protocol, new_created_at
        FROM input
            LEFT JOIN sel USING (protocol)
        WHERE sel.protocol IS NULL
        ORDER BY input.protocol
        ON CONFLICT ON CONSTRAINT uq_protocols_protocol DO UPDATE
            SET protocol = protocols.protocol
        RETURNING id, protocol
    )
    SELECT id FROM sel
    UNION
    SELECT id FROM ups
    ORDER BY id;
$upsert_protocols$ LANGUAGE sql;


CREATE OR REPLACE FUNCTION upsert_protocol(
    new_protocol TEXT,
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_protocol$
    WITH sel AS (
        SELECT id, protocol
        FROM protocols
        WHERE protocol = new_protocol
    ), ups AS (
        INSERT INTO protocols (protocol, created_at)
        SELECT new_protocol, new_created_at
        WHERE NOT EXISTS (SELECT NULL FROM sel)
        ON CONFLICT ON CONSTRAINT uq_protocols_protocol DO UPDATE
            SET protocol = new_protocol
        RETURNING id, protocol
    )
    SELECT id FROM sel
    UNION
    SELECT id FROM ups;
$upsert_protocol$ LANGUAGE sql;


COMMIT;

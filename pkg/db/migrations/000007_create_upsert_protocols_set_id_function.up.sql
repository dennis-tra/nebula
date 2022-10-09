BEGIN;

CREATE OR REPLACE FUNCTION upsert_protocol_set_id(
    new_protocol_ids INT[],
    hash BYTEA
) RETURNS INT AS
$upsert_protocol_set_id$
    WITH sel AS (
        SELECT id, protocol_ids
        FROM protocols_sets
        WHERE protocol_ids = new_protocol_ids
    ), ups AS (
        INSERT INTO protocols_sets (protocol_ids, hash)
        SELECT new_protocol_ids, hash
        WHERE NOT EXISTS (SELECT NULL FROM sel)
        ON CONFLICT (hash) DO UPDATE
            SET protocol_ids = new_protocol_ids
        RETURNING id, protocol_ids
    )
    SELECT id FROM sel
    UNION
    SELECT id FROM ups;

$upsert_protocol_set_id$ LANGUAGE sql;

COMMIT;

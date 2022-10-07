BEGIN;

CREATE OR REPLACE FUNCTION upsert_protocols(
    new_protocols TEXT[],
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT[] AS
$upsert_protocols$
DECLARE
    upserted_protocol_ids INT[];
BEGIN
    -- Check if the array of protocols is null
    IF new_protocols IS NULL THEN
        RETURN NULL;
    END IF;

    -- Left join the existing protocols on the unnested version of the protocols array.
    -- Every row where the joined table is NULL is corresponds to a protocol that
    -- needs to be inserted.
    WITH insert_protocols AS (--
        SELECT DISTINCT new_protocols_table new_protocol
        FROM unnest(new_protocols) new_protocols_table
                 LEFT JOIN protocols p ON p.protocol = new_protocols_table
        WHERE p.id IS NULL)
    INSERT
    INTO protocols (protocol, created_at)
    SELECT TRIM(new_protocol), new_created_at
    FROM insert_protocols
    WHERE TRIM(new_protocol) != '' -- filter protocols that are just the empty string
    ORDER BY new_protocol
    ON CONFLICT DO NOTHING;

    -- Gather list of all protocol IDs to be returned to the caller.
    SELECT sort(array_agg(DISTINCT id))
    FROM protocols
    WHERE protocol = ANY (new_protocols)
    INTO upserted_protocol_ids;

    RETURN upserted_protocol_ids;
END;
$upsert_protocols$ LANGUAGE plpgsql;

COMMIT;

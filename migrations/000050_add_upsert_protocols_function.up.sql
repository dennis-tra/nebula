-- Begin the transaction
BEGIN;

CREATE OR REPLACE FUNCTION upsert_protocols(
    new_protocols TEXT[],
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT[] AS
$upsert_protocols$
DECLARE
    upserted_protocol_ids INT[];
BEGIN
    IF new_protocols IS NULL OR array_length(new_protocols, 1) = 0 THEN
        RETURN NULL;
    END IF;

    WITH insert_protocols AS (
        SELECT new_protocols_table new_protocol
        FROM unnest(new_protocols) new_protocols_table
                 LEFT JOIN protocols p ON p.protocol = new_protocols_table
        WHERE p.id IS NULL)
    INSERT
    INTO protocols (protocol, updated_at, created_at)
    SELECT new_protocol,
           new_created_at,
           new_created_at
    FROM insert_protocols;

    SELECT sort(array_agg(id))
    FROM protocols
    WHERE protocol = ANY (new_protocols)
    GROUP BY id
    INTO upserted_protocol_ids;

    RETURN upserted_protocol_ids;
END;
$upsert_protocols$ LANGUAGE plpgsql;

COMMIT;


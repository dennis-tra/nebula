BEGIN;

CREATE OR REPLACE FUNCTION upsert_protocol_set_id(
    new_protocol_ids INT[]
) RETURNS INT AS
$upsert_protocol_set_id$
DECLARE
    insert_protocol_set_ids   INT[];
    upserted_protocols_set_id INT;
BEGIN
    -- filter duplicates and nulls from array and sort it ID ascending
    SELECT array(SELECT DISTINCT unnest(new_protocol_ids) WHERE 1 IS NOT NULL ORDER BY 1) INTO insert_protocol_set_ids;

    IF insert_protocol_set_ids IS NULL OR array_length(insert_protocol_set_ids, 1) IS NULL THEN
        RETURN NULL;
    END IF;

    -- Check if set of protocol IDs already exists
    SELECT id
    FROM protocols_sets ps
    WHERE ps.protocol_ids = insert_protocol_set_ids
    INTO upserted_protocols_set_id;

    -- If the set of protocol IDs does not exist update it
    IF upserted_protocols_set_id IS NULL THEN
        INSERT
        INTO protocols_sets (protocol_ids) (SELECT insert_protocol_set_ids WHERE insert_protocol_set_ids IS NOT NULL)
        ON CONFLICT DO NOTHING
        RETURNING id INTO upserted_protocols_set_id;

        IF upserted_protocols_set_id IS NULL THEN
            SELECT id
            FROM protocols_sets ps
            WHERE ps.protocol_ids = insert_protocol_set_ids
            INTO upserted_protocols_set_id;
        END IF;
    END IF;

    RETURN upserted_protocols_set_id;
END;
$upsert_protocol_set_id$ LANGUAGE plpgsql;

COMMIT;

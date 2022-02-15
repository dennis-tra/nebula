BEGIN;

CREATE OR REPLACE FUNCTION upsert_multi_addresses_set_id(
    new_multi_address_ids INT[]
) RETURNS INT AS
$upsert_multi_addresses_set_id$
DECLARE
    insert_multi_address_set_ids    INT[];
    upserted_multi_addresses_set_id INT;
BEGIN
    -- filter duplicates and nulls from array and sort it ID ascending
    SELECT array(SELECT DISTINCT unnest(new_multi_address_ids) WHERE 1 IS NOT NULL ORDER BY 1)
    INTO insert_multi_address_set_ids;

    IF insert_multi_address_set_ids IS NULL OR array_length(insert_multi_address_set_ids, 1) IS NULL THEN
        RETURN NULL;
    END IF;

    -- Check if set of multi_address IDs already exists
    SELECT id
    FROM multi_addresses_sets ps
    WHERE ps.multi_address_ids = insert_multi_address_set_ids
    INTO upserted_multi_addresses_set_id;

    -- If the set of multi_address IDs does not exist update it
    IF upserted_multi_addresses_set_id IS NULL THEN
        INSERT
        INTO multi_addresses_sets (multi_address_ids, updated_at, created_at) (SELECT insert_multi_address_set_ids, NOW(), NOW()
                                                                               WHERE insert_multi_address_set_ids IS NOT NULL)
        ON CONFLICT DO NOTHING
        RETURNING id INTO upserted_multi_addresses_set_id;

        IF upserted_multi_addresses_set_id IS NULL THEN
            SELECT id
            FROM multi_addresses_sets ps
            WHERE ps.multi_address_ids = insert_multi_address_set_ids
            INTO upserted_multi_addresses_set_id;
        END IF;
    END IF;

    RETURN upserted_multi_addresses_set_id;
END;
$upsert_multi_addresses_set_id$ LANGUAGE plpgsql;

COMMIT;

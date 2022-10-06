BEGIN;

CREATE OR REPLACE FUNCTION upsert_multi_addresses(
    new_multi_addresses TEXT[],
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT[] AS
$upsert_multi_addresses$
DECLARE
    upserted_multi_address_ids INT[];
BEGIN
    -- Check if the array of multi addresses is null
    IF new_multi_addresses IS NULL THEN
        RETURN NULL;
    END IF;

    -- Left join the existing multi addresses on the unnested version of the multi addresses array.
    -- Every row where the joined table is NULL is corresponds to a multi address that
    -- needs to be inserted.
    WITH insert_multi_addresses AS (--
        SELECT DISTINCT new_multi_addresses_table new_multi_address
        FROM unnest(new_multi_addresses) new_multi_addresses_table
                 LEFT JOIN multi_addresses ma ON ma.maddr = new_multi_addresses_table
        WHERE ma.id IS NULL)
    INSERT
    INTO multi_addresses (maddr, updated_at, created_at)
    SELECT new_multi_address,
           new_created_at,
           new_created_at
    FROM insert_multi_addresses
    WHERE TRIM(new_multi_address) != ''
    ORDER BY new_multi_address
    ON CONFLICT DO NOTHING;

    SELECT sort(array_agg(DISTINCT id))
    FROM multi_addresses
    WHERE maddr = ANY (new_multi_addresses)
    INTO upserted_multi_address_ids;

    RETURN upserted_multi_address_ids;
END
$upsert_multi_addresses$ LANGUAGE plpgsql;

COMMIT;

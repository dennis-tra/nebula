CREATE OR REPLACE FUNCTION upsert_multi_addresses(
    new_peer_id INT,
    new_multi_addresses TEXT[]
) RETURNS INT AS
$upsert_multi_addresses$
DECLARE
    upserted_multi_addresses_set_id INT;
    new_multi_addresses_ids         INT[];
BEGIN

    WITH multi_addresses_ids_table AS (
        INSERT INTO multi_addresses (maddr, updated_at, created_at)
            SELECT DISTINCT unnest(new_multi_addresses), NOW(), NOW() -- the DISTINCT is the FIX
            ORDER BY 1
            ON CONFLICT (maddr) DO UPDATE SET updated_at = excluded.updated_at
            RETURNING id)
    SELECT sort(array_agg(mai.id)) multi_addresses_ids
    FROM multi_addresses_ids_table mai
    INTO new_multi_addresses_ids;

    IF new_multi_addresses_ids IS NOT NULL THEN
        SELECT id
        FROM multi_addresses_sets mas
        WHERE mas.multi_address_ids = new_multi_addresses_ids
        INTO upserted_multi_addresses_set_id;

        IF upserted_multi_addresses_set_id IS NULL THEN
            INSERT
            INTO multi_addresses_sets (multi_address_ids, updated_at, created_at) (SELECT new_multi_addresses_ids, NOW(), NOW())
            RETURNING id INTO upserted_multi_addresses_set_id;
        END IF;
    END IF;

    -- remove current association multi addresses and properties to peers
    DELETE FROM peers_x_multi_addresses WHERE peer_id = new_peer_id;

    -- Add multi address association
    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT new_peer_id, ma.id
    FROM multi_addresses ma
    WHERE maddr = ANY (new_multi_addresses);

    RETURN upserted_multi_addresses_set_id;
END;
$upsert_multi_addresses$ LANGUAGE plpgsql;


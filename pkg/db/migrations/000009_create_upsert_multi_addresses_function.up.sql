BEGIN;

CREATE OR REPLACE FUNCTION upsert_multi_addresses(
    new_multi_addresses TEXT[],
    new_created_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS TABLE (id INT) AS
$upsert_multi_addresses$
    WITH input AS (
        SELECT DISTINCT unnest AS maddr
        FROM UNNEST(new_multi_addresses) unnest
    ), sel AS (-- select all existing multi_addresses
        SELECT multi_addresses.id, multi_addresses.maddr
        FROM input
        INNER JOIN multi_addresses USING (maddr)
    ), ups AS (-- upsert all multi_addresses that don't exist yet
        INSERT INTO multi_addresses (maddr, updated_at, created_at)
        SELECT input.maddr, new_created_at, new_created_at
        FROM input
            LEFT JOIN sel USING (maddr)
        WHERE sel.maddr IS NULL
        ORDER BY input.maddr
        ON CONFLICT ON CONSTRAINT uq_multi_addresses_address DO UPDATE
            SET maddr = multi_addresses.maddr
        RETURNING id, maddr
    )
    SELECT id FROM sel
    UNION
    SELECT id FROM ups
    ORDER BY id;
$upsert_multi_addresses$ LANGUAGE sql;

COMMIT;

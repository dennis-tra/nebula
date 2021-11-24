BEGIN;

DO
$$
    BEGIN

        -- drop the default value for the id column so that we can drop the sequence
        ALTER TABLE peers
            ALTER id DROP DEFAULT;

        -- drop the sequence for the id column
        DROP SEQUENCE peers_id_seq;

        -- transform id column into a bigint
        ALTER TABLE peers
            ALTER COLUMN id TYPE BIGINT;

        -- mark the id column as an identity type
        ALTER TABLE peers
            ALTER id ADD GENERATED ALWAYS AS IDENTITY;


        -- tell the identity column to start at the maximum id of the current table
        IF (SELECT max(id) FROM peers) IS NOT NULL THEN
            PERFORM setval(pg_get_serial_sequence('peers', 'id'), coalesce((SELECT max(id) FROM peers), 0));
        END IF;


        -- change foreign key table column types
        ALTER TABLE latencies
            ALTER COLUMN peer_id TYPE BIGINT;
        ALTER TABLE neighbors
            ALTER COLUMN peer_id TYPE BIGINT;
        ALTER TABLE neighbors
            ALTER COLUMN neighbor_ids TYPE BIGINT[];
        ALTER TABLE peers_x_multi_addresses
            ALTER COLUMN peer_id TYPE BIGINT;
        ALTER TABLE sessions
            ALTER COLUMN peer_id TYPE BIGINT;
        ALTER TABLE visits
            ALTER COLUMN peer_id TYPE BIGINT;

    END
$$;


COMMIT;

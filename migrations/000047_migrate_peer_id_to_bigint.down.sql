BEGIN;

DO
$$
    BEGIN

        -- drop identity from id column
        ALTER TABLE peers
            ALTER COLUMN id DROP IDENTITY;

        -- transform id column into a int and create sequence
        CREATE SEQUENCE peers_id_seq AS INT;
        ALTER TABLE peers
            ALTER COLUMN id TYPE INT;
        ALTER SEQUENCE peers_id_seq OWNED BY peers.id;
        ALTER TABLE peers
            ALTER COLUMN id SET DEFAULT nextval('peers_id_seq');
        ALTER TABLE peers
            ALTER COLUMN id SET NOT NULL;

        -- tell the identity column to start at the maximum id of the current table
        IF (SELECT max(id) FROM peers) IS NOT NULL THEN
            PERFORM setval(pg_get_serial_sequence('peers', 'id'), coalesce((SELECT max(id) FROM peers), 0));
        END IF;

        -- change foreign key table column types
        ALTER TABLE latencies
            ALTER COLUMN peer_id TYPE INT;
        ALTER TABLE neighbors
            ALTER COLUMN peer_id TYPE INT;
        ALTER TABLE neighbors
            ALTER COLUMN neighbor_ids TYPE INT[];
        ALTER TABLE peers_x_multi_addresses
            ALTER COLUMN peer_id TYPE INT;
        ALTER TABLE sessions
            ALTER COLUMN peer_id TYPE INT;
        ALTER TABLE visits
            ALTER COLUMN peer_id TYPE INT;

    END
$$;
COMMIT;

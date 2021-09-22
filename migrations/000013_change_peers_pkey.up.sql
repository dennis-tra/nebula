-- The former id column should hold the peer_id string and
-- a new serial database ID should be used. Doing it this
-- way will prevent the database from storing the peer_id
-- of the form QmbLHAnMo... multiple times. Thus saving
-- disk space.

-- Create new peer_id column holding the actual QmbLHAnMo... string
ALTER TABLE peers
    RENAME COLUMN id TO peer_id;

-- Create new id column holding the internal serial ID of that peer
ALTER TABLE peers
    ADD COLUMN id SERIAL NOT NULL;

------------------------------
---- SESSIONS ----
------------------------------

-- Move peer id string to temporary column
ALTER TABLE sessions
    RENAME COLUMN peer_id TO old_peer_id;
-- Drop foreign key constraints
ALTER TABLE sessions
    DROP CONSTRAINT fk_session_peer;
ALTER TABLE sessions
    DROP CONSTRAINT uq_peer_id_first_failed_dial;
-- Add new peer_id foreign key column pointing to the
-- new id column on the peers table.
ALTER TABLE sessions
    ADD COLUMN peer_id SERIAL NOT NULL;

-- Fill the peer_id column on the sessions table with the new serial ID.
UPDATE sessions
SET peer_id=subquery.id
FROM (SELECT id, peer_id FROM peers) AS subquery
WHERE sessions.old_peer_id = subquery.peer_id;

------------------------------
---- LATENCIES ----
------------------------------

-- Move peer id string to temporary column
ALTER TABLE latencies
    RENAME COLUMN peer_id TO old_peer_id;
-- Drop foreign key constraint
ALTER TABLE latencies
    DROP CONSTRAINT fk_latencies_peer;
-- Add new peer_id foreign key column pointing to the
-- new id column on the peers table.
ALTER TABLE latencies
    ADD COLUMN peer_id SERIAL NOT NULL;

-- Fill the peer_id column on the latencies table with the new serial ID.
UPDATE latencies
SET peer_id=subquery.id
FROM (SELECT id, peer_id FROM peers) AS subquery
WHERE latencies.old_peer_id = subquery.peer_id;

------------------------------
---- CONNECTIONS ----
------------------------------

-- Move peer id string to temporary column
ALTER TABLE connections
    RENAME COLUMN peer_id TO old_peer_id;
-- Add new peer_id foreign key column pointing to the
-- new id column on the peers table.
ALTER TABLE connections
    ADD COLUMN peer_id SERIAL NOT NULL;

-- Fill the peer_id column on the connections table with the new serial ID.
UPDATE connections
SET peer_id=subquery.id
FROM (SELECT id, peer_id FROM peers) AS subquery
WHERE connections.old_peer_id = subquery.peer_id;

------------------------------
---- NEIGHBOURS ----
------------------------------

-- Move peer id string to temporary column
ALTER TABLE neighbours
    RENAME COLUMN peer_id TO old_peer_id;
ALTER TABLE neighbours
    RENAME COLUMN neighbour_peer_id TO old_neighbour_peer_id;

-- Add new peer_id foreign key columns pointing to the
-- new id column on the peers table.
ALTER TABLE neighbours
    ADD COLUMN peer_id      SERIAL NOT NULL,
    ADD COLUMN neighbour_id SERIAL NOT NULL;

-- Fill the peer_id column on the neighbours table with the new serial ID.
UPDATE neighbours
SET peer_id=subquery.id
FROM (SELECT id, peer_id FROM peers) AS subquery
WHERE neighbours.old_peer_id = subquery.peer_id;

-- Fill the neighbour_id column on the neighbours table with the new serial ID.
UPDATE neighbours
SET neighbour_id=subquery.id
FROM (SELECT id, peer_id FROM peers) AS subquery
WHERE neighbours.old_neighbour_peer_id = subquery.peer_id;

------------------------------------------------------------------------------------------------------------------------
-- Make the new serial id column the new primary key on the peers table.
ALTER TABLE peers
    DROP CONSTRAINT peers_pkey;
ALTER TABLE peers
    ADD PRIMARY KEY (id);
------------------------------------------------------------------------------------------------------------------------

------------------------------
---- SESSIONS ----
------------------------------

-- Add a foreign key constraint on the sessions table to only allow
-- valid IDs for the peer_id.
ALTER TABLE sessions
    ADD CONSTRAINT fk_sessions_peer
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;
ALTER TABLE sessions
    ADD CONSTRAINT uq_peer_id_first_failed_dial
        UNIQUE (peer_id, first_failed_dial);

-- Remove old peer_id data from sessions table.
ALTER TABLE sessions
    DROP COLUMN old_peer_id;

------------------------------
---- LATENCIES ----
------------------------------

-- Add a foreign key constraint on the latencies table to only allow
-- valid IDs for the peer_id.
ALTER TABLE latencies
    ADD CONSTRAINT fk_latencies_peer
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;

-- Remove old peer_id data from sessions table.
ALTER TABLE latencies
    DROP COLUMN old_peer_id;

------------------------------
---- CONNECTIONS ----
------------------------------

-- Add a foreign key constraint on the connections table to only allow
-- valid IDs for the peer_id.
ALTER TABLE connections
    ADD CONSTRAINT fk_connections_peer
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;

-- Remove old peer_id data from sessions table.
ALTER TABLE connections
    DROP COLUMN old_peer_id;

------------------------------
---- NEIGHBOURS ----
------------------------------
-- Add a foreign key constraint on the neighbours table to only allow
-- valid IDs for the peer_id.
ALTER TABLE neighbours
    ADD CONSTRAINT fk_neighbours_peer
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;
ALTER TABLE neighbours
    ADD CONSTRAINT fk_neighbours_neighbour
        FOREIGN KEY (neighbour_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;

-- Remove old peer_id data from neighbours table.
ALTER TABLE neighbours
    DROP COLUMN old_peer_id,
    DROP COLUMN old_neighbour_peer_id;

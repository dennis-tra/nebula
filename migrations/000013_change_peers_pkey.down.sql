ALTER TABLE peers
    RENAME COLUMN id TO old_serial_id;
ALTER TABLE peers
    RENAME COLUMN peer_id TO id;

------------------------------
---- SESSIONS ----
------------------------------

ALTER TABLE sessions
    RENAME COLUMN peer_id TO old_serial_peer_id;
ALTER TABLE sessions
    DROP CONSTRAINT fk_sessions_peer;
ALTER TABLE sessions
    DROP CONSTRAINT uq_peer_id_first_failed_dial;
ALTER TABLE sessions
    ADD COLUMN peer_id VARCHAR(100) NOT NULL DEFAULT '';

UPDATE sessions
SET peer_id=subquery.id
FROM (SELECT old_serial_id, id FROM peers) AS subquery
WHERE sessions.old_serial_peer_id = subquery.old_serial_id;

------------------------------
---- LATENCIES ----
------------------------------
ALTER TABLE latencies
    RENAME COLUMN peer_id TO old_serial_peer_id;
ALTER TABLE latencies
    DROP CONSTRAINT fk_latencies_peer;
ALTER TABLE latencies
    ADD COLUMN peer_id VARCHAR(100) NOT NULL DEFAULT '';

UPDATE latencies
SET peer_id=subquery.id
FROM (SELECT old_serial_id, id FROM peers) AS subquery
WHERE latencies.old_serial_peer_id = subquery.old_serial_id;

------------------------------
---- CONNECTIONS ----
------------------------------
ALTER TABLE connections
    RENAME COLUMN peer_id TO old_serial_peer_id;
ALTER TABLE connections
    DROP CONSTRAINT fk_connections_peer;
ALTER TABLE connections
    ADD COLUMN peer_id VARCHAR(100) NOT NULL DEFAULT '';

UPDATE connections
SET peer_id=subquery.id
FROM (SELECT old_serial_id, id FROM peers) AS subquery
WHERE connections.old_serial_peer_id = subquery.old_serial_id;

------------------------------
---- NEIGHBOURS ----
------------------------------
ALTER TABLE neighbours
    RENAME COLUMN peer_id TO old_serial_peer_id;
ALTER TABLE neighbours
    RENAME COLUMN neighbour_id TO old_serial_neighbour_id;
ALTER TABLE neighbours
    DROP CONSTRAINT fk_neighbours_peer;
ALTER TABLE neighbours
    DROP CONSTRAINT fk_neighbours_neighbour;
ALTER TABLE neighbours
    ADD COLUMN peer_id VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE neighbours
    ADD COLUMN neighbour_peer_id VARCHAR(100) NOT NULL DEFAULT '';

UPDATE neighbours
SET peer_id=subquery.id
FROM (SELECT old_serial_id, id FROM peers) AS subquery
WHERE neighbours.old_serial_peer_id = subquery.old_serial_id;

UPDATE neighbours
SET neighbour_peer_id=subquery.id
FROM (SELECT old_serial_id, id FROM peers) AS subquery
WHERE neighbours.old_serial_neighbour_id = subquery.old_serial_id;

------------------------------------------------------------------------------------------------------------------------
ALTER TABLE peers
    DROP CONSTRAINT peers_pkey;
ALTER TABLE peers
    ADD PRIMARY KEY (id);
DROP INDEX idx_peers_peer_id;
------------------------------------------------------------------------------------------------------------------------

------------------------------
---- SESSIONS ----
------------------------------
ALTER TABLE sessions
    ADD CONSTRAINT fk_session_peer -- no plural "sessions"
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;
ALTER TABLE sessions
    ADD CONSTRAINT uq_peer_id_first_failed_dial
        UNIQUE (peer_id, first_failed_dial);

ALTER TABLE sessions
    DROP COLUMN old_serial_peer_id;

------------------------------
---- LATENCIES ----
------------------------------
ALTER TABLE latencies
    ADD CONSTRAINT fk_latencies_peer
        FOREIGN KEY (peer_id)
            REFERENCES peers (id)
            ON DELETE CASCADE;
ALTER TABLE latencies
    DROP COLUMN old_serial_peer_id;

------------------------------
---- CONNECTIONS ----
------------------------------

-- no constraints were present before, so don't adding any
ALTER TABLE connections
    DROP COLUMN old_serial_peer_id;

------------------------------
---- NEIGHBOURS ----
------------------------------
-- no constraints were present before, so don't adding any
ALTER TABLE neighbours
    DROP COLUMN old_serial_peer_id;
ALTER TABLE neighbours
    DROP COLUMN old_serial_neighbour_id;

------------------------------------------------------------------------------------------------------------------------
ALTER TABLE peers
    DROP COLUMN old_serial_id;

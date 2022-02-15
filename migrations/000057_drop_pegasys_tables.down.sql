BEGIN;

CREATE TABLE IF NOT EXISTS pegasys_neighbours
(
    id                SERIAL,
    peer_id           VARCHAR(100) NOT NULL,
    neighbour_peer_id VARCHAR(100) NOT NULL,
    created_at        TIMESTAMPTZ,
    crawl_start_at    TIMESTAMPTZ,

    PRIMARY KEY (id)
);


CREATE TABLE IF NOT EXISTS pegasys_connections
(
    id           SERIAL,
    peer_id      VARCHAR(100) NOT NULL,
    dial_attempt TIMESTAMPTZ,
    latency      INTERVAL,
    is_succeed   BOOLEAN,
    error        VARCHAR(100),

    PRIMARY KEY (id)
);

COMMIT;

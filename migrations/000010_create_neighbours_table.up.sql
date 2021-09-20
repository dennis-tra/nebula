-- The `neighbours` table keeps all neighbour peer id
CREATE TABLE IF NOT EXISTS neighbours
(
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- The neighbour peer ID in the form of Qm... or 12D3...
    neighbour_peer_id     VARCHAR(100) NOT NULL,
    -- Time of add this neighbour
    created_at          TIMESTAMPTZ,
    -- Time of starting the crawling
    crawl_start_at      TIMESTAMPTZ,

    PRIMARY KEY (id)
);

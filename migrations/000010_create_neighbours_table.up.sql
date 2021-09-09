-- The `neightbours` table keeps all neightbour peer id
CREATE TABLE IF NOT EXISTS neightbours
(
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- The neightbour peer ID in the form of Qm... or 12D3...
    neightbour_peer_id     VARCHAR(100) NOT NULL,
    -- Time of add this neightbour
    created_at          TIMESTAMPTZ,

    PRIMARY KEY (peer_id, neightbour_peer_id)
);

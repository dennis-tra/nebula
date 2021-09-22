-- The `connections` table keeps track of latency of each libp2p Connect call
CREATE TABLE IF NOT EXISTS connections
(
    -- A unique id that identifies a particular session
    id           SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id      VARCHAR(100) NOT NULL,
    -- Time of dial
    dial_attempt TIMESTAMPTZ,
    -- Latency
    latency      INTERVAL,
    -- Fail or success
    is_succeed   BOOLEAN,
    -- error message
    error        VARCHAR(100),

    PRIMARY KEY (id)
);

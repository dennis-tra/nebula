-- The `monitoring_connections` table keeps track of latency for each dial
CREATE TABLE IF NOT EXISTS monitoring_connections
(
    -- A unique id that identifies a particular session
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- ipv4 or ipv6 address of the peer
    ip_address            VARCHAR(100),

    -- Time of dial
    dial_attempt          TIMESTAMPTZ,
    -- Latency 
    latency               INTERVAL,
    -- 
    is_succeed            BOOLEAN,

    PRIMARY KEY (id)
);

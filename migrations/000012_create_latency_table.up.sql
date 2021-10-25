-- The `latency` table keeps track of avg latency
CREATE TABLE IF NOT EXISTS latency
(
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- Time of dial
    dial_attempts         INT,
    -- Latency 
    avg_latency           INTERVAL,

    PRIMARY KEY (peer_id)
);

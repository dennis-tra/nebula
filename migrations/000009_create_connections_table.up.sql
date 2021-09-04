-- The `connections` table keeps track of latency for each dial
CREATE TABLE IF NOT EXISTS connections
(
    -- A unique id that identifies a particular session
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- multi address of the peer
    multi_address         VARCHAR(255) ARRAY,
    -- The version string of agent
    agent_version         VARCHAR(255),
    -- Time of dial
    dial_attempt          TIMESTAMPTZ,
    -- Latency 
    latency               INTERVAL,
    -- Fail or success
    is_succeed            BOOLEAN,

    PRIMARY KEY (id)
);

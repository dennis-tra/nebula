-- The `peers` table keeps track of all peers ever found in the DHT
CREATE TABLE IF NOT EXISTS peers
(
    -- The peer ID in the form of Qm... or 12D3...
    id              VARCHAR(100),
    -- The peer ID in the form of Qm... or 12D3...
    multi_addresses VARCHAR(255) ARRAY,

    -- When were the multi addresses updated the last time.
    updated_at      TIMESTAMPTZ NOT NULL,
    -- When was this peer instance created.
    -- This gives a pretty accurate idea of
    -- when this peer was seen the first time.
    created_at      TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (id)
);

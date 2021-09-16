-- The `latencies` table holds latency measurement results.
CREATE TABLE IF NOT EXISTS latencies
(
    -- A unique id that identifies this latency measurement
    id                 SERIAL,
    -- The peer that we measured the latency to.
    peer_id            VARCHAR(100) NOT NULL,
    -- The peer that we measured the latency to.
    address            VARCHAR(100) NOT NULL,
    -- The average round trip time (RTT) latency in seconds
    ping_latency_s_avg FLOAT        NOT NULL,
    -- The standard deviation of the RTT in seconds
    ping_latency_s_std FLOAT        NOT NULL,
    -- The minimum observed ping RTT in seconds
    ping_latency_s_min FLOAT        NOT NULL,
    -- The minimum observed ping RTT in seconds
    ping_latency_s_max FLOAT        NOT NULL,
    -- The number of sent ping packets
    ping_packets_sent  INT          NOT NULL,
    -- The number of received ping packets
    ping_packets_recv  INT          NOT NULL,
    -- The number of duplicate ping packets received for one sent ping packet
    ping_packets_dupl  INT          NOT NULL,
    -- The percentage of packets lost
    ping_packet_loss   FLOAT        NOT NULL,

    -- When was this crawl updated the last time
    updated_at         TIMESTAMPTZ  NOT NULL,
    -- When was this crawl instance created (different from started_at)
    created_at         TIMESTAMPTZ  NOT NULL,

    -- The peer ID should always point to an existing peer in the DB
    CONSTRAINT fk_session_peer FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE INDEX idx_latencies_peer_id ON latencies (peer_id);

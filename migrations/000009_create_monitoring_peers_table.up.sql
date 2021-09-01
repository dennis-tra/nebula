-- The `monitoring_peers` table keeps track of monitoring peers.
CREATE TABLE IF NOT EXISTS monitoring_peers
(
    -- A unique id that identifies a particular session
    id                    SERIAL,
    -- The peer ID in the form of Qm... or 12D3...
    peer_id               VARCHAR(100) NOT NULL,
    -- ipv4 or ipv6 address of the peer
    ip_address            VARCHAR(100),
    -- geo location according the ip address
    geo_location          VARCHAR(100),

    -- When should we try to dial the peer again
    next_dial_attempt     TIMESTAMPTZ,
    -- The duration that this peer was online due to multiple subsequent successful dials
    min_duration          INTERVAL,
    -- The duration that from the first successful dial to the point were it was unreachable
    max_duration          INTERVAL,
    -- How many subsequent successful dials could we track
    successful_dials      INTEGER,
    -- timestamp for the last successfull dial
    last_successful_at    TIMESTAMPTZ,
    -- timestamp for the last failed dial
    last_failed_at        TIMESTAMPTZ,

    PRIMARY KEY (id)
);

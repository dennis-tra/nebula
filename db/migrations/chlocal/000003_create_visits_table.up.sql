-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations

-- Captures the information when crawling a peer
CREATE TABLE visits
(
    -- the associated crawl that is the reason for this visit. If this is null
    -- the visit was done by Nebula's monitoring mode which operates independent of
    -- individual crawls.
    crawl_id         Nullable(UUID),

    -- the peer ID of the peer that we crawled in base58 format
    peer_id          String,

    -- the agent version string that the peer reported to have
    agent_version    String,

    -- a sorted list of protocols/capabilities that the peer claims to support
    protocols        Array(LowCardinality(String)),

    -- a list of unsorted multi addresses that the crawler tried to dial. This can be
    -- fewer addresses than we found in the network because, e.g., the crawler
    -- won't try to connect to IP addresses in private CIDRs by default. It
    -- could also be that the peer advertised multi addresses with protocols
    -- that the crawler does not yet support (unlikely though). If libp2p
    -- detects that an IP address is unreachable but we have multiple multi
    -- addresses with that IP address in the dial_maddrs list, we might not
    -- try to dial these additional maddrs. The dial_errors will contain a
    -- not_dialed entry for the respective multi address. This can also
    -- happen if there is a webtransport address and another plain quic address
    -- if the quic connection fails we won't even try the webtransport address.
    dial_maddrs      Array(String),

    -- a list of unsorted multi addresses that we found in the network for the
    -- given peer but didn't try to dial. The union of filtered_maddrs and
    -- dial_maddrs are all addresses we've found for the given peer in the
    -- network. Nebula doesn't try to dial addresses from private IP addresses
    -- by default (configurable though).
    filtered_maddrs Array(String),

    -- a list of unsorted multi addresses that the peer additionally listens on.
    -- After the crawler has connected to the peer, that peer will push all
    -- addresses it listens on to the crawler. This list can contain additional
    -- addresses that were not found in the network through the regular
    -- discovery protocol.
    extra_maddrs     Array(String),

    -- in case we could not connect to the peer, this field will contain a list
    -- of errors that occurred for each of the multi addresses in dial_maddrs.
    dial_errors      Array(LowCardinality(String)),

    -- in case we could successfully connect, this field contains the multi
    -- address that worked.
    connect_maddr    Nullable(String),

    -- in case we could connect to the peer but then crawling failed, this field
    -- will contain the associated error. TODO: nil if a single request succeeded
    crawl_error      LowCardinality(Nullable(String)),

    -- the timestamp when we started processing (dialing) the peer
    visit_started_at DateTime64(3),

    -- the timestamp when the processing was completed by the client
    visit_ended_at   DateTime64(3),

    -- an object of arbitrary key value pairs with network-specific information.
    peer_properties  JSON()
) ENGINE MergeTree() PRIMARY KEY (visit_started_at)
-- add weekly partitioning. Mode "3" is in accordance with ISO 8601:1988,
-- considers Monday the first day of the week, and is also used by
-- ClickHouse's `toISOWeek()` compatibility function.
PARTITION BY toStartOfMonth(visit_started_at)
TTL toDateTime(visit_started_at) + INTERVAL 1 YEAR;


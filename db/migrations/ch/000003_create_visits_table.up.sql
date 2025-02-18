-- Captures the information when crawling a peer
CREATE TABLE visits
(
    -- the associated crawl that is the reason for this visit. If this is null
    -- the visit was done by the monitoring task which operates independent of
    -- individual crawls.
    crawl_id         Nullable(UUID),
    -- the peer ID of the peer that we crawled in base58 format
    peer_id          String,
    -- the agent version string that the peer reported to have
    agent_version    String,
    -- the set of protocols/capabilities that the peer claims to support
    protocols        Array(LowCardinality(String)),
    -- a list of unsorted multi addresses that the peer advertised to be reachable
    -- at. This is not necessarily the set of addresses we tried to dial because
    -- it could contain only private addresses which we don't even try to dial.
    listen_maddrs    Array(String),
    -- in case we could not connect to the peer, this field will contain a list
    -- of errors that occurred for each of the multi addresses that we dialed.
    dial_errors      Array(LowCardinality(String)),
    -- in case we could successfully connect, this field contains the multi
    -- address that worked.
    connect_maddr    Nullable(String),
    -- in case we could connect to the peer but then crawling failed, this field
    -- will contain the associated error.
    crawl_error      LowCardinality(Nullable(String)),
    -- the timestamp when we started processing (dialing) the peer
    visit_started_at DateTime64(3),
    -- the timestamp when the processing was completed by the client
    visit_ended_at   DateTime64(3),
    -- an object of arbitrary key value pairs with network-specific information.
    peer_properties  JSON()
) ENGINE MergeTree()
      PRIMARY KEY (visit_started_at)
      TTL toDateTime(visit_started_at) + INTERVAL 180 DAY;


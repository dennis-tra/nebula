CREATE TABLE visit
(
    crawl_id              Nullable(UUID),
    peer_id               String,
    agent_version         String,
    protocols             Array(LowCardinality(String)),
    type                  Enum('crawl', 'dial'),
    multi_addresses       Array(String),
    connect_multi_address Nullable(String),
    connect_errors        Array(LowCardinality(String)),
    crawl_error           LowCardinality(Nullable(String)),
    visit_started_at      DateTime64(3),
    visit_ended_at        DateTime64(3)
--     peer_properties       JSON()
) ENGINE MergeTree()
      PRIMARY KEY (visit_started_at)
TTL toDateTime(visit_started_at) + INTERVAL 180 DAY;


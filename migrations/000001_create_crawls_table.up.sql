CREATE TABLE IF NOT EXISTS crawls(
     id SERIAL PRIMARY KEY,
     started_at TIMESTAMPTZ NOT NULL,
     finished_at TIMESTAMPTZ NOT NULL,
     crawled_peers INTEGER NOT NULL,
     dialable_peers INTEGER NOT NULL,
     undialable_peers INTEGER NOT NULL,
--      neighbor_count_distribution JSONB NOT NULL,
--      connect_duration_distribution JSONB NOT NULL,
     created_at TIMESTAMPTZ NOT NULL,
     updated_at TIMESTAMPTZ NOT NULL,
     deleted_at TIMESTAMPTZ
);

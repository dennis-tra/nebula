CREATE TABLE IF NOT EXISTS peers(
     id VARCHAR(100) PRIMARY KEY,
     first_dial TIMESTAMPTZ NOT NULL,
     last_dial TIMESTAMPTZ NOT NULL,
     next_dial TIMESTAMPTZ NOT NULL,
     failed_dial TIMESTAMPTZ,
     dials INTEGER NOT NULL,
     created_at TIMESTAMPTZ NOT NULL,
     updated_at TIMESTAMPTZ NOT NULL,
     deleted_at TIMESTAMPTZ
);

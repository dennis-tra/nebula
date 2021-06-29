CREATE TABLE IF NOT EXISTS multi_addresses (
    peer_id VARCHAR(100) PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

ALTER TABLE multi_addresses
    ADD CONSTRAINT fk_multi_address_peer
    FOREIGN KEY (peer_id) REFERENCES peers (id)
    ON DELETE CASCADE;

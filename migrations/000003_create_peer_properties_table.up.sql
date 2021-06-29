CREATE TABLE IF NOT EXISTS peer_properties (
    id SERIAL PRIMARY KEY,
    property VARCHAR(50) NOT NULL,
    value VARCHAR(50) NOT NULL,
    count INT NOT NULL,
    crawl_id SERIAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE peer_properties ADD CONSTRAINT fk_peer_property_crawl
    FOREIGN KEY (crawl_id)
    REFERENCES crawls (id)
    ON DELETE CASCADE;

-- Begin the transaction
BEGIN;

ALTER TABLE crawl_properties
    RENAME TO crawl_properties_old;

CREATE TABLE crawl_properties
(
    id               SERIAL PRIMARY KEY,
    crawl_id         SERIAL      NOT NULL,
    protocol_id      INT,
    agent_version_id INT,
    error            dial_error,
    count            INT         NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);

ALTER TABLE crawl_properties
    ADD CONSTRAINT fk_crawl_properties_crawl_id
        FOREIGN KEY (crawl_id)
            REFERENCES crawls (id)
            ON DELETE CASCADE;

ALTER TABLE crawl_properties
    ADD CONSTRAINT fk_crawl_properties_protocol_id
        FOREIGN KEY (protocol_id)
            REFERENCES protocols (id)
            ON DELETE NO ACTION;

ALTER TABLE crawl_properties
    ADD CONSTRAINT fk_crawl_properties_agent_version_id
        FOREIGN KEY (agent_version_id)
            REFERENCES agent_versions (id)
            ON DELETE NO ACTION;

INSERT INTO crawl_properties (crawl_id, protocol_id, count, created_at, updated_at)
SELECT cp.crawl_id, prot.id protocol_id, count, cp.created_at, cp.updated_at
FROM crawl_properties_old cp
         INNER JOIN properties p on cp.property_id = p.id
         INNER JOIN protocols prot on prot.protocol = p.value
WHERE p.property = 'protocol';

INSERT INTO crawl_properties (crawl_id, agent_version_id, count, created_at, updated_at)
SELECT cp.crawl_id, av.id agent_version_id, count, cp.created_at, cp.updated_at
FROM crawl_properties_old cp
         INNER JOIN properties p on cp.property_id = p.id
         INNER JOIN agent_versions av on av.agent_version = p.value
WHERE p.property = 'agent_version';

INSERT INTO crawl_properties (crawl_id, error, count, created_at, updated_at)
SELECT cp.crawl_id, p.value::dial_error, count, cp.created_at, cp.updated_at
FROM crawl_properties_old cp
         INNER JOIN properties p on cp.property_id = p.id
WHERE p.property = 'error';

DROP TABLE crawl_properties_old;

-- End the transaction
COMMIT;

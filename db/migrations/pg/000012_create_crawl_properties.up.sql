BEGIN;

-- Rows in the `crawl_properties` table contain aggregated information of one particular crawl.
CREATE TABLE crawl_properties
(
    -- An internal unique id that identifies this ip address.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- A reference to the crawl that this aggregated crawl property belongs to.
    crawl_id         INT NOT NULL,
    -- If not NULL the count value corresponds to the number of peers we found with this protocol during the referenced crawl.
    protocol_id      INT,
    -- If not NULL the count value corresponds to the number of peers we found with this agent version during the referenced crawl.
    agent_version_id INT,
    -- If not NULL the count value corresponds to the number of peers we failed to connect to with this error type during the referenced crawl.
    error            net_error,
    -- The aggregated count of one of the properties above (protocol, agent version, dial error).
    count            INT NOT NULL CHECK ( count >= 0 ),

    CONSTRAINT fk_crawl_properties_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE CASCADE,
    CONSTRAINT fk_crawl_properties_protocol_id FOREIGN KEY (protocol_id) REFERENCES protocols (id) ON DELETE NO ACTION,
    CONSTRAINT fk_crawl_properties_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE NO ACTION,

    PRIMARY KEY (id)
);


COMMENT ON TABLE crawl_properties IS 'Rows in the `crawl_properties` table contain aggregated information of one particular crawl.';
COMMENT ON COLUMN crawl_properties.id IS 'An internal unique id that identifies this ip address.';
COMMENT ON COLUMN crawl_properties.crawl_id IS 'A reference to the crawl that this aggregated crawl property belongs to.';
COMMENT ON COLUMN crawl_properties.protocol_id IS 'If not NULL the count value corresponds to the number of peers we found with this protocol during the referenced crawl.';
COMMENT ON COLUMN crawl_properties.agent_version_id IS 'If not NULL the count value corresponds to the number of peers we found with this agent version during the referenced crawl.';
COMMENT ON COLUMN crawl_properties.error IS 'If not NULL the count value corresponds to the number of peers we failed to connect to with this error type during the referenced crawl.';
COMMENT ON COLUMN crawl_properties.count IS 'The aggregated count of one of the properties above (protocol, agent version, dial error).';

COMMIT;

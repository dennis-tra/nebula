BEGIN;

ALTER SEQUENCE visits_id_seq RENAME TO visits_old_id_seq;

ALTER TABLE visits
    RENAME TO visits_old;
ALTER TABLE visits_2022_12
    RENAME TO visits_old_2022_12;
ALTER TABLE visits_2023_01
    RENAME TO visits_old_2023_01;
ALTER TABLE visits_2023_02
    RENAME TO visits_old_2023_02;
ALTER TABLE visits_2023_03
    RENAME TO visits_old_2023_03;
ALTER TABLE visits_2023_04
    RENAME TO visits_old_2023_04;
ALTER TABLE visits_2023_05
    RENAME TO visits_old_2023_05;
ALTER TABLE visits_2023_06
    RENAME TO visits_old_2023_06;
ALTER TABLE visits_2023_07
    RENAME TO visits_old_2023_07;
ALTER TABLE visits_2023_08
    RENAME TO visits_old_2023_08;
ALTER TABLE visits_2023_09
    RENAME TO visits_old_2023_09;
ALTER TABLE visits_2023_10
    RENAME TO visits_old_2023_10;
ALTER TABLE visits_2023_11
    RENAME TO visits_old_2023_11;

ALTER INDEX visits_pkey RENAME TO visits_old_pkey;
ALTER INDEX idx_visits_peer_id RENAME TO idx_visits_old_peer_id;
ALTER INDEX idx_visits_visit_started_at RENAME TO idx_visits_old_visit_started_at;

ALTER TABLE visits_old
    RENAME CONSTRAINT fk_visits_agent_version_id TO fk_visits_old_agent_version_id;
ALTER TABLE visits_old
    RENAME CONSTRAINT fk_visits_crawl_id TO fk_visits_old_crawl_id;
ALTER TABLE visits_old
    RENAME CONSTRAINT fk_visits_peer_id TO fk_visits_old_peer_id;
ALTER TABLE visits_old
    RENAME CONSTRAINT fk_visits_protocols_set_id TO fk_visits_old_protocols_set_id;

CREATE TABLE visits
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY,
    peer_id           INT         NOT NULL,
    crawl_id          INT,
    session_id        INT,
    agent_version_id  INT,
    protocols_set_id  INT,
    type              visit_type  NOT NULL,
    connect_error     net_error,
    crawl_error       net_error,
    visit_started_at  TIMESTAMPTZ NOT NULL,
    visit_ended_at    TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL,
    dial_duration     INTERVAL,
    connect_duration  INTERVAL,
    crawl_duration    INTERVAL,
    multi_address_ids INT[],
    is_exposed        BOOL,

    CONSTRAINT fk_visits_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_visits_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls (id) ON DELETE SET NULL,
    CONSTRAINT fk_visits_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,
    CONSTRAINT fk_visits_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL,

    PRIMARY KEY (id, visit_started_at)
) PARTITION BY RANGE (visit_started_at);

CREATE INDEX idx_visits_visit_started_at ON visits (visit_started_at);
CREATE INDEX idx_visits_peer_id ON visits (peer_id);

ALTER SEQUENCE visits_id_seq START 2147483648;
ALTER SEQUENCE visits_id_seq RESTART 2147483648;
ALTER SEQUENCE visits_id_seq MINVALUE 2147483648;

ALTER FUNCTION insert_visit(new_crawl_id integer, new_peer_multi_hash text, new_multi_addresses text[], new_agent_version_id integer, new_protocols_set_id integer, new_dial_duration interval, new_connect_duration interval, new_crawl_duration interval, new_visit_started_at timestamp with time zone, new_visit_ended_at timestamp with time zone, new_type visit_type, new_connect_error net_error, new_crawl_error net_error, new_is_exposed boolean)
    RENAME TO insert_visit_old;

CREATE OR REPLACE FUNCTION insert_visit(
    new_crawl_id INT,
    new_peer_multi_hash TEXT,
    new_multi_addresses TEXT[],
    new_agent_version_id INT,
    new_protocols_set_id INT,
    new_dial_duration INTERVAL,
    new_connect_duration INTERVAL,
    new_crawl_duration INTERVAL,
    new_visit_started_at TIMESTAMPTZ,
    new_visit_ended_at TIMESTAMPTZ,
    new_type visit_type,
    new_connect_error net_error,
    new_crawl_error net_error,
    new_is_exposed BOOL
) RETURNS RECORD AS
$insert_visit$
DECLARE
    new_peer_id             INT;
    new_multi_addresses_ids INT[];
    new_session_id          INT;
    new_visit_id            BIGINT;
BEGIN

    SELECT upsert_peer(new_peer_multi_hash, new_agent_version_id,
                       new_protocols_set_id, new_is_exposed, new_visit_ended_at)
    INTO new_peer_id;

    SELECT array_agg(id)
    FROM upsert_multi_addresses(new_multi_addresses)
    INTO new_multi_addresses_ids;

    DELETE
    FROM peers_x_multi_addresses pxma
    WHERE peer_id = new_peer_id;

    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT new_peer_id, new_multi_address_id
    FROM unnest(new_multi_addresses_ids) new_multi_address_id
    ON CONFLICT DO NOTHING;

    SELECT upsert_session(new_peer_id, new_visit_started_at, new_visit_ended_at,
                          new_connect_error)
    INTO new_session_id;

    -- Now we're able to create the normalized visit instance
    INSERT INTO visits (peer_id, crawl_id, session_id, dial_duration,
                        connect_duration, crawl_duration,
                        visit_started_at, visit_ended_at, created_at, type,
                        connect_error, crawl_error,
                        agent_version_id, protocols_set_id, multi_address_ids,
                        is_exposed)
    SELECT new_peer_id,
           new_crawl_id,
           new_session_id,
           new_dial_duration,
           new_connect_duration,
           new_crawl_duration,
           new_visit_started_at,
           new_visit_ended_at,
           NOW(),
           new_type,
           new_connect_error,
           new_crawl_error,
           new_agent_version_id,
           new_protocols_set_id,
           new_multi_addresses_ids,
           new_is_exposed
    RETURNING id INTO new_visit_id;

    RETURN ROW (new_peer_id, new_visit_id, new_session_id);
END;
$insert_visit$ LANGUAGE plpgsql;

-- ALTER TABLE visits OWNER TO nebula_ipfs; -- optional?

COMMIT;
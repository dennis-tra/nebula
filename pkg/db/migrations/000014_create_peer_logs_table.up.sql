BEGIN;

CREATE TABLE peer_logs
(
    id         INT GENERATED ALWAYS AS IDENTITY,
    peer_id    INT         NOT NULL,
    field      TEXT        NOT NULL,
    old        TEXT        NOT NULL,
    new        TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_peer_logs_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_peer_logs_peer_id_created_at ON peer_logs (peer_id, created_at);

CREATE OR REPLACE FUNCTION insert_peer_log()
    RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.agent_version_id != NEW.agent_version_id THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'agent_version_id', OLD.agent_version_id, NEW.agent_version_id, NOW());
    END IF;

    IF OLD.protocols_set_id != NEW.protocols_set_id THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'protocols_set_id', OLD.protocols_set_id, NEW.protocols_set_id, NOW());
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER on_peer_update
    BEFORE UPDATE
    ON peers
    FOR EACH ROW
EXECUTE PROCEDURE insert_peer_log();

COMMIT;

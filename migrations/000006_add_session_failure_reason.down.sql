ALTER TABLE sessions
    DROP CONSTRAINT con_finish_reason_not_null_for_finished;
ALTER TABLE sessions
    DROP COLUMN finish_reason;

DROP TYPE dial_error;


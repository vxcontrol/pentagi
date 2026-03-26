-- +goose Up
-- +goose StatementBegin
CREATE TYPE MSGLOG_RESULT_FORMAT AS ENUM ('plain', 'markdown', 'terminal');

ALTER TABLE msglogs ADD COLUMN result_format MSGLOG_RESULT_FORMAT NULL DEFAULT 'plain';

UPDATE msglogs SET result_format = 'plain';

ALTER TABLE msglogs ALTER COLUMN result_format SET NOT NULL;

CREATE INDEX msglogs_result_format_idx ON msglogs(result_format);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE msglogs DROP COLUMN result_format;

DROP TYPE MSGLOG_RESULT_FORMAT;
-- +goose StatementEnd
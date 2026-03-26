-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

DROP INDEX IF EXISTS assistantlogs_message_idx;
CREATE INDEX assistantlogs_message_idx ON assistantlogs USING GIN (message gin_trgm_ops);
CREATE INDEX assistantlogs_result_idx ON assistantlogs USING GIN (result gin_trgm_ops);
CREATE INDEX assistantlogs_thinking_idx ON assistantlogs USING GIN (thinking gin_trgm_ops);

DROP INDEX IF EXISTS msglogs_message_idx;
CREATE INDEX msglogs_message_idx ON msglogs USING GIN (message gin_trgm_ops);
CREATE INDEX msglogs_result_idx ON msglogs USING GIN (result gin_trgm_ops);
CREATE INDEX msglogs_thinking_idx ON msglogs USING GIN (thinking gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS assistantlogs_message_idx;
DROP INDEX IF EXISTS assistantlogs_result_idx;
DROP INDEX IF EXISTS assistantlogs_thinking_idx;
CREATE INDEX assistantlogs_message_idx ON assistantlogs(message);

DROP INDEX IF EXISTS msglogs_message_idx;
DROP INDEX IF EXISTS msglogs_result_idx;
DROP INDEX IF EXISTS msglogs_thinking_idx;
CREATE INDEX msglogs_message_idx ON msglogs(message);
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
-- A headless flow provisions a sandbox container but never runs the agent
-- loop, so it needs no LLM provider; model_provider_name/model_provider_type
-- still carry a placeholder ('headless'/'custom') to satisfy their NOT NULL
-- constraints without inventing a new PROVIDER_TYPE value.
ALTER TABLE flows ADD COLUMN headless BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX flows_headless_idx ON flows(headless) WHERE headless;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS flows_headless_idx;
ALTER TABLE flows DROP COLUMN headless;
-- +goose StatementEnd

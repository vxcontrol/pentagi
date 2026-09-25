-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS toolcalls_call_id_idx;
CREATE INDEX toolcalls_call_id_idx ON toolcalls USING HASH (call_id);
-- +goose StatementEnd

-- +goose Down

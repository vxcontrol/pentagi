-- +goose Up
-- +goose StatementBegin
ALTER TABLE flows ADD COLUMN trace_id TEXT NULL;

CREATE INDEX flows_trace_id_idx ON flows(trace_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE flows DROP COLUMN trace_id;
-- +goose StatementEnd

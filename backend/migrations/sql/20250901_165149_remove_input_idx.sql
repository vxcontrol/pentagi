-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS tasks_input_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX tasks_input_idx ON tasks(input);
-- +goose StatementEnd

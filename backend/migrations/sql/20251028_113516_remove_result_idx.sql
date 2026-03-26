-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS tasks_result_idx;
DROP INDEX IF EXISTS subtasks_result_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX tasks_result_idx ON tasks(result);
CREATE INDEX subtasks_result_idx ON subtasks(result);
-- +goose StatementEnd

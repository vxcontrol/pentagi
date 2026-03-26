-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS subtasks_description_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX subtasks_description_idx ON subtasks(description);
-- +goose StatementEnd

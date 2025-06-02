-- +goose Up
-- +goose StatementBegin
ALTER TABLE subtasks ADD COLUMN context TEXT NULL DEFAULT '';

UPDATE subtasks SET context = '';

ALTER TABLE subtasks ALTER COLUMN context SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE subtasks DROP COLUMN context;
-- +goose StatementEnd
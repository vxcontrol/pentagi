-- +goose Up
-- +goose StatementBegin
ALTER TABLE msglogs ADD COLUMN thinking TEXT NULL;

ALTER TABLE assistantlogs ADD COLUMN thinking TEXT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE msglogs DROP COLUMN thinking;

ALTER TABLE assistantlogs DROP COLUMN thinking;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
ALTER TYPE PROVIDER_TYPE ADD VALUE IF NOT EXISTS 'minimax';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL does not support removing enum values directly.
-- The 'minimax' value will remain in the enum type.
-- +goose StatementEnd

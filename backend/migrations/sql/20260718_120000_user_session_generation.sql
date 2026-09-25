-- +goose Up
-- +goose StatementBegin

-- Sessions live entirely in a signed cookie, so there is no registry to delete
-- from. This counter is the server side of a logout: a cookie carries the
-- generation in force at issue, and the middleware refuses one that lags.
ALTER TABLE users ADD COLUMN session_generation BIGINT NOT NULL DEFAULT 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users DROP COLUMN session_generation;

-- +goose StatementEnd

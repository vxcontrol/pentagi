-- +goose Up
-- +goose StatementBegin
ALTER TYPE PROVIDER_TYPE ADD VALUE IF NOT EXISTS 'deepseek';
ALTER TYPE PROVIDER_TYPE ADD VALUE IF NOT EXISTS 'glm';
ALTER TYPE PROVIDER_TYPE ADD VALUE IF NOT EXISTS 'kimi';
ALTER TYPE PROVIDER_TYPE ADD VALUE IF NOT EXISTS 'qwen';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL 不支持从 ENUM 删除值，降级时无需操作
-- +goose StatementEnd

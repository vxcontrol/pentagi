-- +goose Up
-- +goose StatementBegin
CREATE TYPE TOKEN_STATUS AS ENUM ('active', 'revoked');

CREATE TABLE api_tokens (
  id          BIGINT         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  token_id    TEXT           NOT NULL,
  user_id     BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id     BIGINT         NOT NULL REFERENCES roles(id),
  name        TEXT           NULL,
  ttl         BIGINT         NOT NULL,
  status      TOKEN_STATUS   NOT NULL DEFAULT 'active',
  created_at  TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP,
  deleted_at  TIMESTAMPTZ    NULL,
  
  CONSTRAINT api_tokens_token_id_unique UNIQUE (token_id)
);

-- Partial unique index for name per user (only when name is not null and not deleted)
CREATE UNIQUE INDEX api_tokens_name_user_unique_idx ON api_tokens(name, user_id) 
  WHERE name IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX api_tokens_token_id_idx ON api_tokens(token_id);
CREATE INDEX api_tokens_user_id_idx ON api_tokens(user_id);
CREATE INDEX api_tokens_status_idx ON api_tokens(status);
CREATE INDEX api_tokens_deleted_at_idx ON api_tokens(deleted_at);

CREATE TRIGGER update_api_tokens_modified
  BEFORE UPDATE ON api_tokens
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Add privileges for Admin role (role_id = 1)
INSERT INTO privileges (role_id, name) VALUES
    (1, 'settings.tokens.admin'),
    (1, 'settings.tokens.create'),
    (1, 'settings.tokens.view'),
    (1, 'settings.tokens.edit'),
    (1, 'settings.tokens.delete'),
    (1, 'settings.tokens.subscribe')
    ON CONFLICT DO NOTHING;

-- Add privileges for User role (role_id = 2)
INSERT INTO privileges (role_id, name) VALUES
    (2, 'settings.tokens.create'),
    (2, 'settings.tokens.view'),
    (2, 'settings.tokens.edit'),
    (2, 'settings.tokens.delete'),
    (2, 'settings.tokens.subscribe')
    ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM privileges WHERE name IN (
  'settings.tokens.create',
  'settings.tokens.view',
  'settings.tokens.edit',
  'settings.tokens.delete',
  'settings.tokens.admin',
  'settings.tokens.subscribe'
);

DROP INDEX IF EXISTS api_tokens_name_user_unique_idx;
DROP TABLE IF EXISTS api_tokens;
DROP TYPE IF EXISTS TOKEN_STATUS;
-- +goose StatementEnd

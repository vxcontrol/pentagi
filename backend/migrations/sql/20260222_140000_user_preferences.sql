-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'settings.user.admin'),
  (1, 'settings.user.view'),
  (1, 'settings.user.edit'),
  (1, 'settings.user.subscribe'),
  (2, 'settings.user.view'),
  (2, 'settings.user.edit'),
  (2, 'settings.user.subscribe');

CREATE TABLE user_preferences (
  id             BIGINT       PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id        BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  preferences    JSONB        NOT NULL DEFAULT '{"favoriteFlows": []}'::JSONB,
  created_at     TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT user_preferences_user_id_unique UNIQUE (user_id)
);

CREATE INDEX user_preferences_user_id_idx ON user_preferences(user_id);
CREATE INDEX user_preferences_preferences_idx ON user_preferences USING GIN (preferences);

INSERT INTO user_preferences (user_id, preferences)
SELECT id, '{"favoriteFlows": []}'::JSONB FROM users
ON CONFLICT DO NOTHING;

CREATE OR REPLACE TRIGGER update_user_preferences_modified
  BEFORE UPDATE ON user_preferences
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_preferences;

DELETE FROM privileges WHERE name IN (
  'settings.user.admin',
  'settings.user.view',
  'settings.user.edit',
  'settings.user.subscribe'
);
-- +goose StatementEnd

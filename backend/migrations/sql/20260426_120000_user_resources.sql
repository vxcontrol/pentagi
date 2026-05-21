-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'flow_files.admin'),
  (1, 'flow_files.view'),
  (1, 'flow_files.upload'),
  (1, 'flow_files.edit'),
  (1, 'flow_files.delete'),
  (1, 'flow_files.download'),
  (1, 'flow_files.subscribe'),
  (2, 'flow_files.view'),
  (2, 'flow_files.upload'),
  (2, 'flow_files.edit'),
  (2, 'flow_files.delete'),
  (2, 'flow_files.download'),
  (2, 'flow_files.subscribe')
  ON CONFLICT DO NOTHING;

INSERT INTO privileges (role_id, name) VALUES
  (1, 'resources.admin'),
  (1, 'resources.view'),
  (1, 'resources.upload'),
  (1, 'resources.edit'),
  (1, 'resources.delete'),
  (1, 'resources.download'),
  (1, 'resources.subscribe'),
  (2, 'resources.view'),
  (2, 'resources.upload'),
  (2, 'resources.edit'),
  (2, 'resources.delete'),
  (2, 'resources.download'),
  (2, 'resources.subscribe')
  ON CONFLICT DO NOTHING;

CREATE TABLE user_resources (
  id         BIGINT      PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  hash       TEXT        NOT NULL DEFAULT '',
  name       TEXT        NOT NULL,
  path       TEXT        NOT NULL,
  size       BIGINT      NOT NULL DEFAULT 0,
  is_dir     BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT user_resources_user_path_unique UNIQUE (user_id, path),
  CONSTRAINT user_resources_name_not_empty   CHECK (length(trim(name)) > 0),
  CONSTRAINT user_resources_path_not_empty   CHECK (length(trim(path)) > 0)
);

CREATE INDEX user_resources_user_id_idx    ON user_resources(user_id);
CREATE INDEX user_resources_hash_idx       ON user_resources(hash) WHERE hash != '';
CREATE INDEX user_resources_path_idx       ON user_resources USING btree (path text_pattern_ops);
CREATE INDEX user_resources_user_path_idx  ON user_resources(user_id, path text_pattern_ops);
CREATE INDEX user_resources_updated_at_idx ON user_resources(updated_at DESC);

CREATE OR REPLACE TRIGGER update_user_resources_modified
  BEFORE UPDATE ON user_resources
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_resources;

DELETE FROM privileges WHERE name IN (
  'flow_files.admin',
  'flow_files.view',
  'flow_files.upload',
  'flow_files.edit',
  'flow_files.delete',
  'flow_files.download',
  'flow_files.subscribe'
);

DELETE FROM privileges WHERE name IN (
  'resources.admin',
  'resources.view',
  'resources.upload',
  'resources.edit',
  'resources.delete',
  'resources.download',
  'resources.subscribe'
);
-- +goose StatementEnd

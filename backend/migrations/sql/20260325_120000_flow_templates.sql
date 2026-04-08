-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'templates.admin'),
  (1, 'templates.create'),
  (1, 'templates.view'),
  (1, 'templates.edit'),
  (1, 'templates.delete'),
  (1, 'templates.subscribe'),
  (2, 'templates.create'),
  (2, 'templates.view'),
  (2, 'templates.edit'),
  (2, 'templates.delete'),
  (2, 'templates.subscribe')
  ON CONFLICT DO NOTHING;

CREATE TABLE flow_templates (
  id             BIGINT       PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id        BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title          TEXT         NOT NULL,
  text           TEXT         NOT NULL,
  created_at     TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT flow_templates_title_not_empty CHECK (length(trim(title)) > 0),
  CONSTRAINT flow_templates_text_not_empty CHECK (length(trim(text)) > 0)
);

CREATE INDEX flow_templates_user_id_idx ON flow_templates(user_id);
CREATE INDEX flow_templates_created_at_idx ON flow_templates(created_at DESC);

CREATE OR REPLACE TRIGGER update_flow_templates_modified
  BEFORE UPDATE ON flow_templates
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS flow_templates;

DELETE FROM privileges WHERE name IN (
  'templates.admin',
  'templates.create',
  'templates.view',
  'templates.edit',
  'templates.delete',
  'templates.subscribe'
);
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'settings.admin'),
  (1, 'settings.view'),
  (2, 'settings.view');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM privileges WHERE name IN (
  'settings.admin',
  'settings.view'
);
-- +goose StatementEnd
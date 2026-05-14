-- +goose Up
-- +goose StatementBegin

INSERT INTO privileges (role_id, name) VALUES
  (1, 'toolcalls.admin'),
  (1, 'toolcalls.view'),
  (2, 'toolcalls.view')
  ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM privileges WHERE name IN ('toolcalls.admin', 'toolcalls.view');

-- +goose StatementEnd

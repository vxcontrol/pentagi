-- +goose Up
-- +goose StatementBegin

INSERT INTO privileges (role_id, name) VALUES
  (1, 'anonymize.call'),
  (2, 'anonymize.call')
  ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM privileges WHERE name = 'anonymize.call';

-- +goose StatementEnd

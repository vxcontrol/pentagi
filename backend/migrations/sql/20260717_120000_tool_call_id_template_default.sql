-- +goose Up
-- +goose StatementBegin

UPDATE flows      SET tool_call_id_template = 'call_{r:24:x}' WHERE coalesce(tool_call_id_template, '') = '';
UPDATE assistants SET tool_call_id_template = 'call_{r:24:x}' WHERE coalesce(tool_call_id_template, '') = '';

ALTER TABLE flows      ALTER COLUMN tool_call_id_template SET DEFAULT 'call_{r:24:x}';
ALTER TABLE assistants ALTER COLUMN tool_call_id_template SET DEFAULT 'call_{r:24:x}';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE flows      ALTER COLUMN tool_call_id_template DROP DEFAULT;
ALTER TABLE assistants ALTER COLUMN tool_call_id_template DROP DEFAULT;

-- +goose StatementEnd

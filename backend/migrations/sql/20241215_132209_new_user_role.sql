-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
    (2, 'roles.view'),
    (2, 'providers.view'),
    (2, 'prompts.view'),
    (2, 'screenshots.view'),
    (2, 'screenshots.download'),
    (2, 'screenshots.subscribe'),
    (2, 'msglogs.view'),
    (2, 'msglogs.subscribe'),
    (2, 'termlogs.view'),
    (2, 'termlogs.subscribe'),
    (2, 'flows.create'),
    (2, 'flows.delete'),
    (2, 'flows.edit'),
    (2, 'flows.view'),
    (2, 'flows.subscribe'),
    (2, 'tasks.view'),
    (2, 'tasks.subscribe'),
    (2, 'subtasks.view'),
    (2, 'containers.view'),
    (2, 'agentlogs.view'),
    (2, 'agentlogs.subscribe'),
    (2, 'vecstorelogs.view'),
    (2, 'vecstorelogs.subscribe'),
    (2, 'searchlogs.view'),
    (2, 'searchlogs.subscribe')
    ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM privileges WHERE role_id = 2;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
    (1, 'agentlogs.admin'),
    (1, 'agentlogs.view'),
    (1, 'agentlogs.subscribe'),
    (1, 'vecstorelogs.admin'),
    (1, 'vecstorelogs.view'),
    (1, 'vecstorelogs.subscribe'),
    (1, 'searchlogs.admin'),
    (1, 'searchlogs.view'),
    (1, 'searchlogs.subscribe')
    ON CONFLICT DO NOTHING;

CREATE TABLE agentlogs (
  id           BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  initiator    MSGCHAIN_TYPE   NOT NULL DEFAULT 'primary_agent',
  executor     MSGCHAIN_TYPE   NOT NULL DEFAULT 'primary_agent',
  task         TEXT            NOT NULL,
  result       TEXT            NOT NULL DEFAULT '',
  flow_id      BIGINT          NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id      BIGINT          NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id   BIGINT          NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX agentlogs_initiator_idx ON agentlogs(initiator);
CREATE INDEX agentlogs_executor_idx ON agentlogs(executor);
CREATE INDEX agentlogs_task_idx ON agentlogs(task);
CREATE INDEX agentlogs_flow_id_idx ON agentlogs(flow_id);
CREATE INDEX agentlogs_task_id_idx ON agentlogs(task_id);
CREATE INDEX agentlogs_subtask_id_idx ON agentlogs(subtask_id);

CREATE TYPE VECSTORE_ACTION_TYPE AS ENUM ('retrieve', 'store');

CREATE TABLE vecstorelogs (
  id           BIGINT                 PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  initiator    MSGCHAIN_TYPE          NOT NULL DEFAULT 'primary_agent',
  executor     MSGCHAIN_TYPE          NOT NULL DEFAULT 'primary_agent',
  filter       JSON                   NOT NULL DEFAULT '{}',
  query        TEXT                   NOT NULL,
  action       VECSTORE_ACTION_TYPE   NOT NULL,
  result       TEXT                   NOT NULL,
  flow_id      BIGINT                 NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id      BIGINT                 NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id   BIGINT                 NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX vecstorelogs_initiator_idx ON vecstorelogs(initiator);
CREATE INDEX vecstorelogs_executor_idx ON vecstorelogs(executor);
CREATE INDEX vecstorelogs_query_idx ON vecstorelogs(query);
CREATE INDEX vecstorelogs_action_idx ON vecstorelogs(action);
CREATE INDEX vecstorelogs_flow_id_idx ON vecstorelogs(flow_id);
CREATE INDEX vecstorelogs_task_id_idx ON vecstorelogs(task_id);
CREATE INDEX vecstorelogs_subtask_id_idx ON vecstorelogs(subtask_id);

CREATE TYPE SEARCHENGINE_TYPE AS ENUM ('google', 'tavily', 'traversaal', 'browser');

CREATE TABLE searchlogs (
  id           BIGINT              PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  initiator    MSGCHAIN_TYPE       NOT NULL DEFAULT 'primary_agent',
  executor     MSGCHAIN_TYPE       NOT NULL DEFAULT 'primary_agent',
  engine       SEARCHENGINE_TYPE   NOT NULL,
  query        TEXT                NOT NULL,
  result       TEXT                NOT NULL DEFAULT '',
  flow_id      BIGINT              NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id      BIGINT              NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id   BIGINT              NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ         DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX searchlogs_initiator_idx ON searchlogs(initiator);
CREATE INDEX searchlogs_executor_idx ON searchlogs(executor);
CREATE INDEX searchlogs_engine_idx ON searchlogs(engine);
CREATE INDEX searchlogs_query_idx ON searchlogs(query);
CREATE INDEX searchlogs_flow_id_idx ON searchlogs(flow_id);
CREATE INDEX searchlogs_task_id_idx ON searchlogs(task_id);
CREATE INDEX searchlogs_subtask_id_idx ON searchlogs(subtask_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE agentlogs;
DROP TABLE vecstorelogs;
DROP TABLE searchlogs;
DROP TYPE VECSTORE_ACTION_TYPE;
DROP TYPE SEARCHENGINE_TYPE;

DELETE FROM privileges WHERE name IN (
  'agentlogs.admin',
  'agentlogs.view',
  'agentlogs.subscribe',
  'vecstorelogs.admin',
  'vecstorelogs.view',
  'vecstorelogs.subscribe',
  'searchlogs.admin',
  'searchlogs.view',
  'searchlogs.subscribe'
);
-- +goose StatementEnd

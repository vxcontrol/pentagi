-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
  id     BIGINT    PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  name   TEXT      NOT NULL,

  CONSTRAINT roles_name_unique UNIQUE (name)
);

CREATE INDEX roles_name_idx ON roles(name);

INSERT INTO roles (name) VALUES
    ('Admin'),
    ('User')
    ON CONFLICT DO NOTHING;

CREATE TABLE privileges (
  id        BIGINT   PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  role_id   BIGINT   NOT NULL REFERENCES roles(id),
  name      TEXT     NOT NULL,

  CONSTRAINT privileges_role_name_unique UNIQUE (role_id, name)
);

CREATE INDEX privileges_role_id_idx ON privileges(role_id);
CREATE INDEX privileges_name_idx ON privileges(name);

INSERT INTO privileges (role_id, name) VALUES
    (1, 'users.create'),
    (1, 'users.delete'),
    (1, 'users.edit'),
    (1, 'users.view'),
    (1, 'roles.view'),
    (1, 'providers.view'),
    (1, 'prompts.view'),
    (1, 'prompts.edit'),
    (1, 'screenshots.admin'),
    (1, 'screenshots.view'),
    (1, 'screenshots.download'),
    (1, 'screenshots.subscribe'),
    (1, 'msglogs.admin'),
    (1, 'msglogs.view'),
    (1, 'msglogs.subscribe'),
    (1, 'termlogs.admin'),
    (1, 'termlogs.view'),
    (1, 'termlogs.subscribe'),
    (1, 'flows.admin'),
    (1, 'flows.create'),
    (1, 'flows.delete'),
    (1, 'flows.edit'),
    (1, 'flows.view'),
    (1, 'flows.subscribe'),
    (1, 'tasks.admin'),
    (1, 'tasks.view'),
    (1, 'tasks.subscribe'),
    (1, 'subtasks.admin'),
    (1, 'subtasks.view'),
    (1, 'containers.admin'),
    (1, 'containers.view')
    ON CONFLICT DO NOTHING;

CREATE TYPE USER_TYPE AS ENUM ('local','oauth');
CREATE TYPE USER_STATUS AS ENUM ('created','active','blocked');

CREATE TABLE users (
  id                         BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  hash                       TEXT          NOT NULL DEFAULT MD5(RANDOM()::text),
  type                       USER_TYPE     NOT NULL DEFAULT 'local',
  mail                       TEXT          NOT NULL,
  name                       TEXT          NOT NULL DEFAULT '',
  password                   TEXT          DEFAULT NULL,
  status                     USER_STATUS   NOT NULL DEFAULT 'created',
  role_id                    BIGINT        NOT NULL DEFAULT '2' REFERENCES roles(id),
  password_change_required   BOOLEAN       NOT NULL DEFAULT false,
  provider                   TEXT          NULL,
  created_at                 TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT users_mail_unique UNIQUE (mail),
  CONSTRAINT users_hash_unique UNIQUE (hash)
);

CREATE INDEX users_role_id_idx ON users(role_id);
CREATE INDEX users_hash_idx ON users(hash);

INSERT INTO users (mail, name, password, status, role_id, password_change_required) VALUES
    (
      'admin@pentagi.com',
      'admin',
      '$2a$10$deVOk0o1nYRHpaVXjIcyCuRmaHvtoMN/2RUT7w5XbZTeiWKEbXx9q',
      'active',
      1,
      true
    )
    ON CONFLICT DO NOTHING;

CREATE TABLE prompts (
  id                         BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type                       TEXT          NOT NULL,
  user_id                    BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  prompt                     TEXT          NOT NULL,

  CONSTRAINT prompts_type_user_id_unique UNIQUE (type, user_id)
);

CREATE INDEX prompts_type_idx ON prompts(type);
CREATE INDEX prompts_user_id_idx ON prompts(user_id);
CREATE INDEX prompts_prompt_idx ON prompts(prompt);

CREATE TYPE FLOW_STATUS AS ENUM ('created','running','waiting','finished','failed');

CREATE TABLE flows (
  id               BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  status           FLOW_STATUS   NOT NULL DEFAULT 'created',
  title            TEXT          NOT NULL DEFAULT 'untitled',
  model            TEXT          NOT NULL,
  model_provider   TEXT          NOT NULL,
  language         TEXT          NOT NULL,
  functions        JSON          NOT NULL DEFAULT '{}',
  prompts          JSON          NOT NULL,
  user_id          BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at       TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,
  deleted_at       TIMESTAMPTZ   NULL
);

CREATE INDEX flows_status_idx ON flows(status);
CREATE INDEX flows_title_idx ON flows(title);
CREATE INDEX flows_language_idx ON flows(language);
CREATE INDEX flows_user_id_idx ON flows(user_id);

CREATE TYPE CONTAINER_TYPE AS ENUM ('primary','secondary');
CREATE TYPE CONTAINER_STATUS AS ENUM ('starting','running','stopped','deleted','failed');

CREATE TABLE containers (
  id           BIGINT             PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type         CONTAINER_TYPE     NOT NULL DEFAULT 'primary',
  name         TEXT               NOT NULL DEFAULT MD5(RANDOM()::text),
  image        TEXT               NOT NULL,
  status       CONTAINER_STATUS   NOT NULL DEFAULT 'starting',
  local_id     TEXT,
  local_dir    TEXT,
  flow_id      BIGINT             NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ        DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMPTZ        DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT containers_local_id_unique UNIQUE (local_id)
);

CREATE INDEX containers_type_idx ON containers(type);
CREATE INDEX containers_name_idx ON containers(name);
CREATE INDEX containers_status_idx ON containers(status);
CREATE INDEX containers_flow_id_idx ON containers(flow_id);

CREATE TYPE TASK_STATUS AS ENUM ('created','running','waiting','finished','failed');

CREATE TABLE tasks (
  id           BIGINT         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  status       TASK_STATUS    NOT NULL DEFAULT 'created',
  title        TEXT           NOT NULL DEFAULT 'untitled',
  input        TEXT           NOT NULL,
  result       TEXT           NOT NULL DEFAULT '',
  flow_id      BIGINT         NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP,
  updated_at   TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX tasks_status_idx ON tasks(status);
CREATE INDEX tasks_title_idx ON tasks(title);
CREATE INDEX tasks_input_idx ON tasks(input);
CREATE INDEX tasks_result_idx ON tasks(result);
CREATE INDEX tasks_flow_id_idx ON tasks(flow_id);

CREATE TYPE SUBTASK_STATUS AS ENUM ('created','running','waiting','finished','failed');

CREATE TABLE subtasks (
  id            BIGINT           PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  status        SUBTASK_STATUS   NOT NULL DEFAULT 'created',
  title         TEXT             NOT NULL,
  description   TEXT             NOT NULL,
  result        TEXT             NOT NULL DEFAULT '',
  task_id       BIGINT           NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX subtasks_status_idx ON subtasks(status);
CREATE INDEX subtasks_title_idx ON subtasks(title);
CREATE INDEX subtasks_description_idx ON subtasks(description);
CREATE INDEX subtasks_result_idx ON subtasks(result);
CREATE INDEX subtasks_task_id_idx ON subtasks(task_id);

CREATE TYPE TOOLCALL_STATUS AS ENUM ('received','running','finished','failed');

CREATE TABLE toolcalls (
  id            BIGINT            PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  call_id       TEXT              NOT NULL,
  status        TOOLCALL_STATUS   NOT NULL DEFAULT 'received',
  name          TEXT              NOT NULL,
  args          JSON              NOT NULL,
  result        TEXT              NOT NULL DEFAULT '',
  flow_id       BIGINT            NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id       BIGINT            NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id    BIGINT            NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ       DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ       DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX toolcalls_call_id_idx ON toolcalls(call_id);
CREATE INDEX toolcalls_status_idx ON toolcalls(status);
CREATE INDEX toolcalls_name_idx ON toolcalls(name);
CREATE INDEX toolcalls_flow_id_idx ON toolcalls(flow_id);
CREATE INDEX toolcalls_task_id_idx ON toolcalls(task_id);
CREATE INDEX toolcalls_subtask_id_idx ON toolcalls(subtask_id);

CREATE TYPE MSGCHAIN_TYPE AS ENUM (
  'primary_agent',
  'reporter',
  'generator',
  'refiner',
  'reflector',
  'enricher',
  'adviser',
  'coder',
  'memorist',
  'searcher',
  'installer',
  'pentester',
  'summarizer'
);

CREATE TABLE msgchains (
  id               BIGINT          PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type             MSGCHAIN_TYPE   NOT NULL DEFAULT 'primary_agent',
  model            TEXT            NOT NULL,
  model_provider   TEXT            NOT NULL,
  usage_in         BIGINT          NOT NULL DEFAULT 0,
  usage_out        BIGINT          NOT NULL DEFAULT 0,
  chain            JSON            NOT NULL,
  flow_id          BIGINT          NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id          BIGINT          NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id       BIGINT          NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at       TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX msgchains_type_idx ON msgchains(type);
CREATE INDEX msgchains_flow_id_idx ON msgchains(flow_id);
CREATE INDEX msgchains_task_id_idx ON msgchains(task_id);
CREATE INDEX msgchains_subtask_id_idx ON msgchains(subtask_id);

CREATE TYPE TERMLOG_TYPE AS ENUM ('stdin', 'stdout','stderr');

CREATE TABLE termlogs (
  id             BIGINT         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type           TERMLOG_TYPE   NOT NULL,
  text           TEXT           NOT NULL,
  container_id   BIGINT         NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
  created_at     TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX termlogs_type_idx ON termlogs(type);
-- CREATE INDEX termlogs_text_idx ON termlogs(text);
CREATE INDEX termlogs_container_id_idx ON termlogs(container_id);

CREATE TYPE MSGLOG_TYPE AS ENUM ('thoughts', 'browser', 'terminal', 'file', 'search', 'advice', 'ask', 'input', 'done');

CREATE TABLE msglogs (
  id           BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type         MSGLOG_TYPE   NOT NULL,
  message      TEXT          NOT NULL,
  result       TEXT          NOT NULL DEFAULT '',
  flow_id      BIGINT        NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id      BIGINT        NULL REFERENCES tasks(id) ON DELETE CASCADE,
  subtask_id   BIGINT        NULL REFERENCES subtasks(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX msglogs_type_idx ON msglogs(type);
CREATE INDEX msglogs_message_idx ON msglogs(message);
-- CREATE INDEX msglogs_result_idx ON msglogs(result);
CREATE INDEX msglogs_flow_id_idx ON msglogs(flow_id);
CREATE INDEX msglogs_task_id_idx ON msglogs(task_id);
CREATE INDEX msglogs_subtask_id_idx ON msglogs(subtask_id);

CREATE TABLE screenshots (
  id           BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  name         TEXT          NOT NULL,
  url          TEXT          NOT NULL,
  flow_id      BIGINT        NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX screenshots_flow_id_idx ON screenshots(flow_id);
CREATE INDEX screenshots_name_idx ON screenshots(name);
CREATE INDEX screenshots_url_idx ON screenshots(url);

CREATE OR REPLACE FUNCTION update_modified_column()   
RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_flows_modified
  BEFORE UPDATE ON flows
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_tasks_modified
  BEFORE UPDATE ON tasks
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_subtasks_modified
  BEFORE UPDATE ON subtasks
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_containers_modified
  BEFORE UPDATE ON containers
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_toolcalls_modified
  BEFORE UPDATE ON toolcalls
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_msgchains_modified
  BEFORE UPDATE ON msgchains
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE screenshots;
DROP TABLE msglogs;
DROP TABLE termlogs;
DROP TABLE msgchains;
DROP TABLE toolcalls;
DROP TABLE subtasks;
DROP TABLE tasks;
DROP TABLE containers;
DROP TABLE flows;
DROP TABLE users;
DROP TABLE roles;
DROP TABLE privileges;
DROP TYPE MSGLOG_TYPE;
DROP TYPE TERMLOG_TYPE;
DROP TYPE MSGCHAIN_TYPE;
DROP TYPE TOOLCALL_STATUS;
DROP TYPE SUBTASK_STATUS;
DROP TYPE TASK_STATUS;
DROP TYPE CONTAINER_STATUS;
DROP TYPE CONTAINER_TYPE;
DROP TYPE FLOW_STATUS;
DROP TYPE USER_STATUS;
DROP TYPE USER_TYPE;
DROP FUNCTION update_modified_column;
-- +goose StatementEnd

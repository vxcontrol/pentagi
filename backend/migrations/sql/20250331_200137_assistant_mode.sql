-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'assistants.admin'),
  (1, 'assistants.create'),
  (1, 'assistants.delete'),
  (1, 'assistants.edit'),
  (1, 'assistants.view'),
  (1, 'assistants.subscribe'),
  (1, 'assistantlogs.admin'),
  (1, 'assistantlogs.view'),
  (1, 'assistantlogs.subscribe'),
  (2, 'assistants.create'),
  (2, 'assistants.delete'),
  (2, 'assistants.edit'),
  (2, 'assistants.view'),
  (2, 'assistants.subscribe'),
  (2, 'assistantlogs.view'),
  (2, 'assistantlogs.subscribe');

ALTER TABLE msgchains ALTER COLUMN type DROP DEFAULT;
ALTER TABLE agentlogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE agentlogs ALTER COLUMN executor DROP DEFAULT;
ALTER TABLE vecstorelogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE vecstorelogs ALTER COLUMN executor DROP DEFAULT;
ALTER TABLE searchlogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE searchlogs ALTER COLUMN executor DROP DEFAULT;

CREATE TYPE MSGCHAIN_TYPE_NEW AS ENUM (
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
  'summarizer',
  'tool_call_fixer',
  'assistant'
);

ALTER TABLE msgchains 
    ALTER COLUMN type TYPE MSGCHAIN_TYPE_NEW USING type::text::MSGCHAIN_TYPE_NEW;

ALTER TABLE agentlogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING initiator::text::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING executor::text::MSGCHAIN_TYPE_NEW;

ALTER TABLE vecstorelogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING initiator::text::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING executor::text::MSGCHAIN_TYPE_NEW;

ALTER TABLE searchlogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING initiator::text::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING executor::text::MSGCHAIN_TYPE_NEW;

DROP TYPE MSGCHAIN_TYPE;

ALTER TYPE MSGCHAIN_TYPE_NEW RENAME TO MSGCHAIN_TYPE;

ALTER TABLE msgchains 
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN type SET DEFAULT 'primary_agent';

ALTER TABLE agentlogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

ALTER TABLE vecstorelogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

ALTER TABLE searchlogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

CREATE TYPE MSGLOG_TYPE_NEW AS ENUM (
  'answer',
  'report',
  'thoughts',
  'browser',
  'terminal',
  'file',
  'search',
  'advice',
  'ask',
  'input',
  'done'
);

ALTER TABLE msglogs 
    ALTER COLUMN type TYPE MSGLOG_TYPE_NEW USING type::text::MSGLOG_TYPE_NEW;

DROP TYPE MSGLOG_TYPE;

ALTER TYPE MSGLOG_TYPE_NEW RENAME TO MSGLOG_TYPE;

ALTER TABLE msglogs 
    ALTER COLUMN type SET NOT NULL;

CREATE TYPE ASSISTANT_STATUS AS ENUM ('created','running','waiting','finished','failed');

CREATE TABLE assistants (
  id               BIGINT             PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  status           ASSISTANT_STATUS   NOT NULL DEFAULT 'created',
  title            TEXT               NOT NULL DEFAULT 'untitled',
  model            TEXT               NOT NULL,
  model_provider   TEXT               NOT NULL,
  language         TEXT               NOT NULL,
  functions        JSON               NOT NULL DEFAULT '{}',
  prompts          JSON               NOT NULL,
  trace_id         TEXT               NULL,
  flow_id          BIGINT             NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  use_agents       BOOLEAN            NOT NULL DEFAULT FALSE,
  msgchain_id      BIGINT             NULL REFERENCES msgchains(id) ON DELETE CASCADE,
  created_at       TIMESTAMPTZ        DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ        DEFAULT CURRENT_TIMESTAMP,
  deleted_at       TIMESTAMPTZ        NULL
);

CREATE INDEX assistants_status_idx ON assistants(status);
CREATE INDEX assistants_title_idx ON assistants(title);
CREATE INDEX assistants_model_provider_idx ON assistants(model_provider);
CREATE INDEX assistants_trace_id_idx ON assistants(trace_id);
CREATE INDEX assistants_flow_id_idx ON assistants(flow_id);
CREATE INDEX assistants_msgchain_id_idx ON assistants(msgchain_id);

CREATE TABLE assistantlogs (
  id              BIGINT                 PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  type            MSGLOG_TYPE            NOT NULL,
  message         TEXT                   NOT NULL,
  result          TEXT                   NOT NULL DEFAULT '',
  result_format   MSGLOG_RESULT_FORMAT   NOT NULL DEFAULT 'plain',
  flow_id         BIGINT                 NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  assistant_id    BIGINT                 NOT NULL REFERENCES assistants(id) ON DELETE CASCADE,
  created_at      TIMESTAMPTZ            DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX assistantlogs_type_idx ON assistantlogs(type);
CREATE INDEX assistantlogs_message_idx ON assistantlogs(message);
CREATE INDEX assistantlogs_result_format_idx ON assistantlogs(result_format);
CREATE INDEX assistantlogs_flow_id_idx ON assistantlogs(flow_id);
CREATE INDEX assistantlogs_assistant_id_idx ON assistantlogs(assistant_id);

CREATE OR REPLACE TRIGGER update_assistants_modified
  BEFORE UPDATE ON assistants
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE assistants;
DROP TABLE assistantlogs;
DROP TYPE ASSISTANT_STATUS;

DELETE FROM privileges WHERE name IN (
  'assistants.admin',
  'assistants.create',
  'assistants.delete',
  'assistants.edit',
  'assistants.view',
  'assistants.subscribe',
  'assistantlogs.admin',
  'assistantlogs.view',
  'assistantlogs.subscribe'
);

DELETE FROM msgchains WHERE type = 'assistant';

ALTER TABLE msgchains ALTER COLUMN type DROP DEFAULT;
ALTER TABLE agentlogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE agentlogs ALTER COLUMN executor DROP DEFAULT;
ALTER TABLE vecstorelogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE vecstorelogs ALTER COLUMN executor DROP DEFAULT;
ALTER TABLE searchlogs ALTER COLUMN initiator DROP DEFAULT;
ALTER TABLE searchlogs ALTER COLUMN executor DROP DEFAULT;

CREATE TYPE MSGCHAIN_TYPE_NEW AS ENUM (
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
  'summarizer',
  'tool_call_fixer'
);

ALTER TABLE msgchains 
    ALTER COLUMN type TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN type::text = 'assistant' THEN 'primary_agent'::text
      ELSE type::text
    END::MSGCHAIN_TYPE_NEW;

ALTER TABLE agentlogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN initiator::text = 'assistant' THEN 'primary_agent'::text
      ELSE initiator::text
    END::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN executor::text = 'assistant' THEN 'primary_agent'::text
      ELSE executor::text
    END::MSGCHAIN_TYPE_NEW;

ALTER TABLE vecstorelogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN initiator::text = 'assistant' THEN 'primary_agent'::text
      ELSE initiator::text
    END::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN executor::text = 'assistant' THEN 'primary_agent'::text
      ELSE executor::text
    END::MSGCHAIN_TYPE_NEW;

ALTER TABLE searchlogs 
    ALTER COLUMN initiator TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN initiator::text = 'assistant' THEN 'primary_agent'::text
      ELSE initiator::text
    END::MSGCHAIN_TYPE_NEW,
    ALTER COLUMN executor TYPE MSGCHAIN_TYPE_NEW USING 
    CASE 
      WHEN executor::text = 'assistant' THEN 'primary_agent'::text
      ELSE executor::text
    END::MSGCHAIN_TYPE_NEW;

DROP TYPE MSGCHAIN_TYPE;

ALTER TYPE MSGCHAIN_TYPE_NEW RENAME TO MSGCHAIN_TYPE;

ALTER TABLE msgchains 
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN type SET DEFAULT 'primary_agent';

ALTER TABLE agentlogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

ALTER TABLE vecstorelogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

ALTER TABLE searchlogs 
    ALTER COLUMN initiator SET NOT NULL,
    ALTER COLUMN initiator SET DEFAULT 'primary_agent',
    ALTER COLUMN executor SET NOT NULL,
    ALTER COLUMN executor SET DEFAULT 'primary_agent';

DELETE FROM msglogs WHERE type = 'answer' OR type = 'report';

CREATE TYPE MSGLOG_TYPE_NEW AS ENUM (
  'thoughts',
  'browser',
  'terminal',
  'file',
  'search',
  'advice',
  'ask',
  'input',
  'done'
);

ALTER TABLE msglogs 
    ALTER COLUMN type TYPE MSGLOG_TYPE_NEW USING type::text::MSGLOG_TYPE_NEW;

DROP TYPE MSGLOG_TYPE;

ALTER TYPE MSGLOG_TYPE_NEW RENAME TO MSGLOG_TYPE;

ALTER TABLE msglogs 
    ALTER COLUMN type SET NOT NULL;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
INSERT INTO privileges (role_id, name) VALUES
  (1, 'settings.providers.admin'),
  (1, 'settings.providers.view'),
  (1, 'settings.providers.edit'),
  (1, 'settings.providers.subscribe'),
  (1, 'settings.prompts.admin'),
  (1, 'settings.prompts.view'),
  (1, 'settings.prompts.edit'),
  (2, 'settings.providers.view'),
  (2, 'settings.providers.edit'),
  (2, 'settings.providers.subscribe'),
  (2, 'settings.prompts.view'),
  (2, 'settings.prompts.edit');

-- Replace old prompt permissions with new settings-namespaced ones
DELETE FROM privileges WHERE name IN (
  'prompts.view',
  'prompts.edit'
);

-- Move prompts from flow/assistant to separate table and load them each time from the database
ALTER TABLE flows DROP COLUMN prompts;
ALTER TABLE assistants DROP COLUMN prompts;

CREATE TYPE PROVIDER_TYPE AS ENUM (
  'openai',
  'anthropic',
  'gemini',
  'bedrock',
  'ollama',
  'custom'
);

CREATE TABLE providers (
  id               BIGINT        PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id          BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type             PROVIDER_TYPE NOT NULL,
  name             TEXT          NOT NULL,
  config           JSON          NOT NULL,
  created_at       TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,
  deleted_at       TIMESTAMPTZ   NULL
);

CREATE INDEX providers_user_id_idx ON providers(user_id);
CREATE INDEX providers_type_idx ON providers(type);
CREATE INDEX providers_name_user_id_idx ON providers(name, user_id);
CREATE UNIQUE INDEX providers_name_user_id_unique ON providers(name, user_id) WHERE deleted_at IS NULL;

-- Add model providers type column and separate name from type
ALTER TABLE flows ADD COLUMN model_provider_type PROVIDER_TYPE NULL;
UPDATE flows SET model_provider_type = model_provider::PROVIDER_TYPE;
ALTER TABLE flows ALTER COLUMN model_provider_type SET NOT NULL;
CREATE INDEX flows_model_provider_type_idx ON flows(model_provider_type);
DROP INDEX IF EXISTS flows_model_provider_idx;
ALTER TABLE flows RENAME COLUMN model_provider TO model_provider_name;
CREATE INDEX flows_model_provider_name_idx ON flows(model_provider_name);

ALTER TABLE assistants ADD COLUMN model_provider_type PROVIDER_TYPE NULL;
UPDATE assistants SET model_provider_type = model_provider::PROVIDER_TYPE;
ALTER TABLE assistants ALTER COLUMN model_provider_type SET NOT NULL;
CREATE INDEX assistants_model_provider_type_idx ON assistants(model_provider_type);
DROP INDEX IF EXISTS assistants_model_provider_idx;
ALTER TABLE assistants RENAME COLUMN model_provider TO model_provider_name;
CREATE INDEX assistants_model_provider_name_idx ON assistants(model_provider_name);

-- ENUM values correspond to template files in backend/pkg/templates/prompts/
CREATE TYPE PROMPT_TYPE AS ENUM (
  'primary_agent',
  'assistant',
  'pentester',
  'question_pentester',
  'coder',
  'question_coder',
  'installer',
  'question_installer',
  'searcher',
  'question_searcher',
  'memorist',
  'question_memorist',
  'adviser',
  'question_adviser',
  'generator',
  'subtasks_generator',
  'refiner',
  'subtasks_refiner',
  'reporter',
  'task_reporter',
  'reflector',
  'question_reflector',
  'enricher',
  'question_enricher',
  'toolcall_fixer',
  'input_toolcall_fixer',
  'summarizer',
  'image_chooser',
  'language_chooser',
  'flow_descriptor',
  'task_descriptor',
  'execution_logs',
  'full_execution_context',
  'short_execution_context'
);

-- Validate existing prompt types are compatible with new ENUM before migration
DO $$
DECLARE
  invalid_types TEXT[];
BEGIN
  SELECT ARRAY_AGG(DISTINCT type) INTO invalid_types
  FROM prompts
  WHERE type::TEXT NOT IN (
    'execution_logs', 'full_execution_context', 'short_execution_context',
    'question_enricher', 'question_adviser', 'question_coder', 'question_installer',
    'question_memorist', 'question_pentester', 'question_searcher', 'question_reflector',
    'input_toolcall_fixer', 'assistant', 'primary_agent', 'flow_descriptor',
    'task_descriptor', 'image_chooser', 'language_chooser', 'task_reporter',
    'toolcall_fixer', 'reporter', 'subtasks_generator', 'generator',
    'subtasks_refiner', 'refiner', 'enricher', 'reflector', 'adviser',
    'coder', 'installer', 'pentester', 'memorist', 'searcher', 'summarizer'
  );

  IF array_length(invalid_types, 1) > 0 THEN
    RAISE EXCEPTION 'Found invalid prompt types that cannot be converted to ENUM: %', 
      array_to_string(invalid_types, ', ');
  END IF;
END$$;

DROP INDEX IF EXISTS prompts_type_idx;
DROP INDEX IF EXISTS prompts_prompt_idx;

ALTER TABLE prompts 
    ALTER COLUMN type TYPE PROMPT_TYPE USING type::text::PROMPT_TYPE;

CREATE INDEX prompts_type_idx ON prompts(type);

ALTER TABLE prompts 
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

CREATE OR REPLACE TRIGGER update_providers_modified
  BEFORE UPDATE ON providers
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE OR REPLACE TRIGGER update_prompts_modified
  BEFORE UPDATE ON prompts
  FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE flows DROP COLUMN model_provider_type;
ALTER TABLE assistants DROP COLUMN model_provider_type;
ALTER TABLE flows RENAME COLUMN model_provider_name TO model_provider;
ALTER TABLE assistants RENAME COLUMN model_provider_name TO model_provider;

-- Delete unsupported model providers
DROP INDEX IF EXISTS flows_model_provider_name_idx;
DROP INDEX IF EXISTS assistants_model_provider_name_idx;
DELETE FROM flows WHERE model_provider NOT IN ('openai', 'anthropic', 'custom');
DELETE FROM assistants WHERE model_provider NOT IN ('openai', 'anthropic', 'custom');
CREATE INDEX flows_model_provider_idx ON flows(model_provider);
CREATE INDEX assistants_model_provider_idx ON assistants(model_provider);

DROP TABLE providers;
DROP TYPE PROVIDER_TYPE;

DELETE FROM privileges WHERE name IN (
  'settings.providers.admin',
  'settings.providers.view',
  'settings.providers.edit',
  'settings.providers.subscribe',
  'settings.prompts.admin',
  'settings.prompts.view',
  'settings.prompts.edit'
);

INSERT INTO privileges (role_id, name) VALUES
  (1, 'prompts.view'),
  (1, 'prompts.edit'),
  (2, 'prompts.view');

-- Convert prompts.type back to TEXT while preserving user data
DROP INDEX IF EXISTS prompts_type_idx;

ALTER TABLE prompts 
    ALTER COLUMN type TYPE TEXT USING type::text;

CREATE INDEX prompts_type_idx ON prompts(type);
CREATE INDEX prompts_prompt_idx ON prompts(prompt);

DROP TRIGGER IF EXISTS update_prompts_modified ON prompts;
ALTER TABLE prompts DROP COLUMN created_at;
ALTER TABLE prompts DROP COLUMN updated_at;

DROP TYPE PROMPT_TYPE;

-- Restore prompts to flows/assistants
ALTER TABLE flows ADD COLUMN prompts JSON NULL;
ALTER TABLE assistants ADD COLUMN prompts JSON NULL;

UPDATE flows SET prompts = '{}';
UPDATE assistants SET prompts = '{}';

ALTER TABLE flows ALTER COLUMN prompts SET NOT NULL;
ALTER TABLE assistants ALTER COLUMN prompts SET NOT NULL;
-- +goose StatementEnd

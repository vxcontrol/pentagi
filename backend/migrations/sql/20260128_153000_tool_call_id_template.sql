-- +goose Up
-- +goose StatementBegin
ALTER TABLE flows ADD COLUMN tool_call_id_template TEXT NULL;
ALTER TABLE assistants ADD COLUMN tool_call_id_template TEXT NULL;

UPDATE flows SET tool_call_id_template = 'call_{r:24:x}';
UPDATE assistants SET tool_call_id_template = 'call_{r:24:x}';

ALTER TABLE flows ALTER COLUMN tool_call_id_template SET NOT NULL;
ALTER TABLE assistants ALTER COLUMN tool_call_id_template SET NOT NULL;

CREATE INDEX flows_tool_call_id_template_idx ON flows(tool_call_id_template) WHERE tool_call_id_template IS NOT NULL;
CREATE INDEX assistants_tool_call_id_template_idx ON assistants(tool_call_id_template) WHERE tool_call_id_template IS NOT NULL;

-- Add new prompt types for tool call ID detection
CREATE TYPE PROMPT_TYPE_NEW AS ENUM (
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
  'short_execution_context',
  'tool_call_id_collector',
  'tool_call_id_detector'
);

-- Update the searchlogs table to use the new enum type
ALTER TABLE prompts
    ALTER COLUMN type TYPE PROMPT_TYPE_NEW USING type::text::PROMPT_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE PROMPT_TYPE;
ALTER TYPE PROMPT_TYPE_NEW RENAME TO PROMPT_TYPE;

-- Set the column as NOT NULL
ALTER TABLE prompts
    ALTER COLUMN type SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS flows_tool_call_id_template_idx;
DROP INDEX IF EXISTS assistants_tool_call_id_template_idx;

ALTER TABLE flows DROP COLUMN IF EXISTS tool_call_id_template;
ALTER TABLE assistants DROP COLUMN IF EXISTS tool_call_id_template;

-- Revert the changes by removing tool call ID collector and detector from the enum
CREATE TYPE PROMPT_TYPE_NEW AS ENUM (
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

-- Update the prompts table to use the new enum type
ALTER TABLE prompts
    ALTER COLUMN type TYPE PROMPT_TYPE_NEW USING type::text::PROMPT_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE PROMPT_TYPE;
ALTER TYPE PROMPT_TYPE_NEW RENAME TO PROMPT_TYPE;

-- Set the column as NOT NULL
ALTER TABLE prompts
    ALTER COLUMN type SET NOT NULL;
-- +goose StatementEnd

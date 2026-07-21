-- +goose Up
-- +goose StatementBegin
-- Add the MiniMax provider to the provider_type enum
CREATE TYPE PROVIDER_TYPE_NEW AS ENUM (
  'openai',
  'anthropic',
  'gemini',
  'bedrock',
  'ollama',
  'custom',
  'deepseek',
  'glm',
  'kimi',
  'qwen',
  'minimax'
);

-- Update columns to use the new enum type
ALTER TABLE providers
    ALTER COLUMN type TYPE PROVIDER_TYPE_NEW USING type::text::PROVIDER_TYPE_NEW;

ALTER TABLE flows
    ALTER COLUMN model_provider_type TYPE PROVIDER_TYPE_NEW USING model_provider_type::text::PROVIDER_TYPE_NEW;

ALTER TABLE assistants
    ALTER COLUMN model_provider_type TYPE PROVIDER_TYPE_NEW USING model_provider_type::text::PROVIDER_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE PROVIDER_TYPE;
ALTER TYPE PROVIDER_TYPE_NEW RENAME TO PROVIDER_TYPE;

-- Ensure NOT NULL constraints are preserved
ALTER TABLE providers
    ALTER COLUMN type SET NOT NULL;

ALTER TABLE flows
    ALTER COLUMN model_provider_type SET NOT NULL;

ALTER TABLE assistants
    ALTER COLUMN model_provider_type SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Delete providers using the MiniMax type before reverting the enum
DELETE FROM providers WHERE type IN ('minimax');
DELETE FROM flows WHERE model_provider_type IN ('minimax');
DELETE FROM assistants WHERE model_provider_type IN ('minimax');

-- Create new enum type without the MiniMax provider
CREATE TYPE PROVIDER_TYPE_NEW AS ENUM (
  'openai',
  'anthropic',
  'gemini',
  'bedrock',
  'ollama',
  'custom',
  'deepseek',
  'glm',
  'kimi',
  'qwen'
);

-- Update columns to use the new enum type
ALTER TABLE providers
    ALTER COLUMN type TYPE PROVIDER_TYPE_NEW USING type::text::PROVIDER_TYPE_NEW;

ALTER TABLE flows
    ALTER COLUMN model_provider_type TYPE PROVIDER_TYPE_NEW USING model_provider_type::text::PROVIDER_TYPE_NEW;

ALTER TABLE assistants
    ALTER COLUMN model_provider_type TYPE PROVIDER_TYPE_NEW USING model_provider_type::text::PROVIDER_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE PROVIDER_TYPE;
ALTER TYPE PROVIDER_TYPE_NEW RENAME TO PROVIDER_TYPE;

-- Ensure NOT NULL constraints are preserved
ALTER TABLE providers
    ALTER COLUMN type SET NOT NULL;

ALTER TABLE flows
    ALTER COLUMN model_provider_type SET NOT NULL;

ALTER TABLE assistants
    ALTER COLUMN model_provider_type SET NOT NULL;
-- +goose StatementEnd

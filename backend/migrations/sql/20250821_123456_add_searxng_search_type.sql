-- +goose Up
-- +goose StatementBegin
-- Add searxng to the searchengine_type enum
CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'traversaal',
  'browser',
  'duckduckgo',
  'perplexity',
  'searxng'
);

-- Update the searchlogs table to use the new enum type
ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING engine::text::SEARCHENGINE_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE SEARCHENGINE_TYPE;
ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

-- Set the column as NOT NULL
ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert the changes by removing searxng from the enum
CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'traversaal',
  'browser',
  'duckduckgo',
  'perplexity'
);

-- Update the searchlogs table to use the new enum type
ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING engine::text::SEARCHENGINE_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE SEARCHENGINE_TYPE;
ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

-- Set the column as NOT NULL
ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd
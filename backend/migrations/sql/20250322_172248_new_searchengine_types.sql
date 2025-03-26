-- +goose Up
-- +goose StatementBegin
ALTER TABLE searchlogs ALTER COLUMN engine DROP DEFAULT;

CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'traversaal',
  'browser',
  'duckduckgo',
  'perplexity'
);

ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING engine::text::SEARCHENGINE_TYPE_NEW;

DROP TYPE SEARCHENGINE_TYPE;

ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE searchlogs ALTER COLUMN engine DROP DEFAULT;

CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'traversaal',
  'browser'
);

ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING
    CASE
      WHEN engine::text = 'duckduckgo' THEN 'google'::text
      WHEN engine::text = 'perplexity' THEN 'browser'::text
      ELSE engine::text
    END::SEARCHENGINE_TYPE_NEW;

DROP TYPE SEARCHENGINE_TYPE;

ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd
-- +goose Up
ALTER TYPE SEARCHENGINE_TYPE ADD VALUE IF NOT EXISTS 'parallel';

-- +goose Down
-- Do not rewrite historical Parallel searches as another provider. Rollback is
-- safe only before this engine has been used; otherwise retain the migration.
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM searchlogs WHERE engine::text = 'parallel') THEN
    RAISE EXCEPTION 'Cannot remove parallel search attribution while searchlogs contain Parallel searches';
  END IF;
END $$;
-- +goose StatementEnd

CREATE TYPE SEARCHENGINE_TYPE_OLD AS ENUM (
  'google', 'tavily', 'firecrawl', 'traversaal', 'browser', 'duckduckgo',
  'perplexity', 'searxng', 'sploitus'
);
ALTER TABLE searchlogs
  ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_OLD USING engine::text::SEARCHENGINE_TYPE_OLD;
DROP TYPE SEARCHENGINE_TYPE;
ALTER TYPE SEARCHENGINE_TYPE_OLD RENAME TO SEARCHENGINE_TYPE;

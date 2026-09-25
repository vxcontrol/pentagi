-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS langchain_pg_embedding_meta_doc_type_flow_id
  ON langchain_pg_embedding ((cmetadata ->> 'doc_type'), (cmetadata ->> 'flow_id'));

DROP INDEX CONCURRENTLY IF EXISTS langchain_pg_embedding_meta_doc_type_partial;

ANALYZE langchain_pg_embedding;

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS langchain_pg_embedding_meta_doc_type_flow_id;

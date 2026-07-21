-- name: GetKnowledgeDocument :one
-- Fetch a single knowledge document by its UUID (admin view — no user_id check).
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND e.uuid::text = sqlc.arg(uuid);

-- name: GetUserKnowledgeDocument :one
-- Fetch a single knowledge document by UUID, scoped to a specific user.
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND e.uuid::text = sqlc.arg(uuid)
  AND (e.cmetadata ->> 'user_id') = sqlc.arg(user_id);

-- name: ListAllKnowledgeDocuments :many
-- List all knowledge documents excluding the noisy memory type (admin view).
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND COALESCE(e.cmetadata ->> 'doc_type', '') NOT IN ('memory')
ORDER BY e.uuid;

-- name: ListFlowKnowledgeDocuments :many
-- List non-memory knowledge documents belonging to a specific flow (admin scoped).
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND COALESCE(e.cmetadata ->> 'doc_type', '') NOT IN ('memory')
  AND (e.cmetadata ->> 'flow_id') = sqlc.arg(flow_id)
ORDER BY e.uuid;

-- name: ListUserKnowledgeDocuments :many
-- List all non-memory knowledge documents owned by a specific user (user-scoped view).
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND COALESCE(e.cmetadata ->> 'doc_type', '') NOT IN ('memory')
  AND (e.cmetadata ->> 'user_id') = sqlc.arg(user_id)
ORDER BY e.uuid;

-- name: UpdateKnowledgeDocument :one
-- Update an existing document's embedding, text and metadata atomically.
-- embedding must be formatted as a PostgreSQL vector literal: '[f1,f2,...]'
-- cmetadata must be valid JSON text.
UPDATE langchain_pg_embedding
SET
  embedding = sqlc.arg(embedding)::vector,
  document  = sqlc.arg(document),
  cmetadata = sqlc.arg(cmetadata)::json
WHERE uuid::text = sqlc.arg(uuid)
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain')
RETURNING
  uuid::text                              AS id,
  COALESCE(document, '')                  AS document,
  COALESCE(cmetadata::text, '{}')         AS cmetadata;

-- name: UpdateKnowledgeDocumentMetadata :one
-- Update only the document's metadata, leaving its text and embedding intact.
-- Used when an update changes no content (e.g. rename) so re-embedding — and
-- therefore a configured embedder — is unnecessary.
-- cmetadata must be valid JSON text.
UPDATE langchain_pg_embedding
SET cmetadata = sqlc.arg(cmetadata)::json
WHERE uuid::text = sqlc.arg(uuid)
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain')
RETURNING
  uuid::text                              AS id,
  COALESCE(document, '')                  AS document,
  COALESCE(cmetadata::text, '{}')         AS cmetadata;

-- name: DeleteKnowledgeDocument :exec
-- Delete a knowledge document by UUID (admin — no user_id check).
DELETE FROM langchain_pg_embedding
WHERE uuid::text = sqlc.arg(uuid)
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain');

-- name: DeleteUserKnowledgeDocument :exec
-- Delete a knowledge document by UUID, only if it belongs to the given user.
DELETE FROM langchain_pg_embedding
WHERE uuid::text = sqlc.arg(uuid)
  AND (cmetadata ->> 'user_id') = sqlc.arg(user_id)
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain');

-- name: SearchKnowledgeDocuments :many
-- Vector similarity search over all knowledge documents (admin view, no user filter).
-- Returns rows ordered by cosine similarity descending (highest score first).
-- embedding    query vector as a PostgreSQL vector literal, e.g. '[0.1,0.2,...]'
-- max_distance cosine-distance upper bound (exclusive); equals (1 - score_threshold),
--              e.g. pass 0.8 to get documents with similarity score > 0.2
-- lim          maximum number of rows to return
SELECT
  e.uuid::text                                                         AS id,
  COALESCE(e.document, '')                                             AS document,
  COALESCE(e.cmetadata::text, '{}')                                   AS cmetadata,
  (1.0 - (e.embedding <=> sqlc.arg(embedding)::vector))::float8       AS score
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND COALESCE(e.cmetadata ->> 'doc_type', '') NOT IN ('memory')
  AND (e.embedding <=> sqlc.arg(embedding)::vector)::float8 < sqlc.arg(max_distance)::float8
ORDER BY e.embedding <=> sqlc.arg(embedding)::vector
LIMIT sqlc.arg(lim)::int;

-- name: SearchUserKnowledgeDocuments :many
-- Vector similarity search scoped to a specific user (by cmetadata user_id).
-- Returns rows ordered by cosine similarity descending (highest score first).
-- embedding    query vector as a PostgreSQL vector literal, e.g. '[0.1,0.2,...]'
-- max_distance cosine-distance upper bound (exclusive); equals (1 - score_threshold)
-- lim          maximum number of rows to return
-- user_id      owner filter as a decimal text string (e.g. "42")
SELECT
  e.uuid::text                                                         AS id,
  COALESCE(e.document, '')                                             AS document,
  COALESCE(e.cmetadata::text, '{}')                                   AS cmetadata,
  (1.0 - (e.embedding <=> sqlc.arg(embedding)::vector))::float8       AS score
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND COALESCE(e.cmetadata ->> 'doc_type', '') NOT IN ('memory')
  AND (e.cmetadata ->> 'user_id') = sqlc.arg(user_id)
  AND (e.embedding <=> sqlc.arg(embedding)::vector)::float8 < sqlc.arg(max_distance)::float8
ORDER BY e.embedding <=> sqlc.arg(embedding)::vector
LIMIT sqlc.arg(lim)::int;

-- name: DeleteFlowMemoryDocuments :exec
-- Delete all memory-type documents for a specific flow.
-- Called on flow deletion to free long-term memory that will never be re-used.
-- flow_id is the decimal text representation of the flow ID (e.g. "55"), matching the
-- text result of (cmetadata ->> 'flow_id') which uses JSON ->> extraction.
DELETE FROM langchain_pg_embedding
WHERE collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain')
  AND COALESCE(cmetadata ->> 'doc_type', '') = 'memory'
  AND (cmetadata ->> 'flow_id') = sqlc.arg(flow_id);

-- name: InsertKnowledgeDocument :one
-- Insert a document with a pre-computed embedding vector and return its UUID.
-- embedding must be formatted as a PostgreSQL vector literal: '[f1,f2,...]'
-- cmetadata must be valid JSON text.
INSERT INTO langchain_pg_embedding (uuid, document, embedding, cmetadata, collection_id)
SELECT
  sqlc.arg(uuid)::uuid,
  sqlc.arg(document),
  sqlc.arg(embedding)::vector,
  sqlc.arg(cmetadata)::json,
  c.uuid
FROM langchain_pg_collection c
WHERE c.name = 'langchain'
RETURNING uuid::text AS id;

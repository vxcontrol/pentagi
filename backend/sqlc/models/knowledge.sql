-- name: GetKnowledgeDocument :one
-- Fetch a single knowledge document by its UUID (admin view — no user_id check).
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND e.uuid::text = $1;

-- name: GetUserKnowledgeDocument :one
-- Fetch a single knowledge document by UUID, scoped to a specific user.
SELECT
  e.uuid::text                              AS id,
  COALESCE(e.document, '')                  AS document,
  COALESCE(e.cmetadata::text, '{}')         AS cmetadata
FROM langchain_pg_embedding e
INNER JOIN langchain_pg_collection c ON e.collection_id = c.uuid
WHERE c.name = 'langchain'
  AND e.uuid::text = $1
  AND (e.cmetadata ->> 'user_id') = $2;

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
  AND (e.cmetadata ->> 'flow_id') = $1
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
  AND (e.cmetadata ->> 'user_id') = $1
ORDER BY e.uuid;

-- name: UpdateKnowledgeDocument :one
-- Update an existing document's embedding, text and metadata atomically.
-- $2 must be formatted as a PostgreSQL vector literal: '[f1,f2,...]'
-- $4 must be valid JSON text.
UPDATE langchain_pg_embedding
SET
  embedding = $2::vector,
  document  = $3,
  cmetadata = $4::json
WHERE uuid::text = $1
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain')
RETURNING
  uuid::text                              AS id,
  COALESCE(document, '')                  AS document,
  COALESCE(cmetadata::text, '{}')         AS cmetadata;

-- name: DeleteKnowledgeDocument :exec
-- Delete a knowledge document by UUID (admin — no user_id check).
DELETE FROM langchain_pg_embedding
WHERE uuid::text = $1
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain');

-- name: DeleteUserKnowledgeDocument :exec
-- Delete a knowledge document by UUID, only if it belongs to the given user.
DELETE FROM langchain_pg_embedding
WHERE uuid::text = $1
  AND (cmetadata ->> 'user_id') = $2
  AND collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain');

-- name: DeleteFlowMemoryDocuments :exec
-- Delete all memory-type documents for a specific flow.
-- Called on flow deletion to free long-term memory that will never be re-used.
-- $1 is the decimal text representation of the flow ID (e.g. "55"), matching the
-- text result of (cmetadata ->> 'flow_id') which uses JSON ->> extraction.
DELETE FROM langchain_pg_embedding
WHERE collection_id = (SELECT uuid FROM langchain_pg_collection WHERE name = 'langchain')
  AND COALESCE(cmetadata ->> 'doc_type', '') = 'memory'
  AND (cmetadata ->> 'flow_id') = $1;

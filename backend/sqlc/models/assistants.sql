-- name: GetFlowAssistants :many
SELECT
  a.*
FROM assistants a
INNER JOIN flows f ON a.flow_id = f.id
WHERE a.flow_id = $1 AND f.deleted_at IS NULL AND a.deleted_at IS NULL
ORDER BY a.created_at DESC;

-- name: GetUserFlowAssistants :many
SELECT
  a.*
FROM assistants a
INNER JOIN flows f ON a.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE a.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL AND a.deleted_at IS NULL
ORDER BY a.created_at DESC;

-- name: GetFlowAssistant :one
SELECT
  a.*
FROM assistants a
INNER JOIN flows f ON a.flow_id = f.id
WHERE a.id = $1 AND a.flow_id = $2 AND f.deleted_at IS NULL AND a.deleted_at IS NULL;

-- name: GetUserFlowAssistant :one
SELECT
  a.*
FROM assistants a
INNER JOIN flows f ON a.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE a.id = $1 AND a.flow_id = $2 AND f.user_id = $3 AND f.deleted_at IS NULL AND a.deleted_at IS NULL;

-- name: GetAssistant :one
SELECT
  a.*
FROM assistants a
WHERE a.id = $1 AND a.deleted_at IS NULL;

-- name: GetAssistantUseAgents :one
SELECT use_agents
FROM assistants
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAssistant :one
INSERT INTO assistants (
  title, status, model, model_provider_name, model_provider_type, language, tool_call_id_template, functions, flow_id, use_agents
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateAssistant :one
UPDATE assistants
SET title = $1, model = $2, language = $3, tool_call_id_template = $4, functions = $5, trace_id = $6, msgchain_id = $7
WHERE id = $8
RETURNING *;

-- name: UpdateAssistantUseAgents :one
UPDATE assistants
SET use_agents = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantStatus :one
UPDATE assistants
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantTitle :one
UPDATE assistants
SET title = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantModel :one
UPDATE assistants
SET model = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantLanguage :one
UPDATE assistants
SET language = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantToolCallIDTemplate :one
UPDATE assistants
SET tool_call_id_template = $1
WHERE id = $2
RETURNING *;

-- name: UpdateAssistantsProviderNameByOldName :many
-- The assistants counterpart of UpdateFlowsProviderNameByOldName. Assistants
-- carry their own provider reference, independent of their flow's, and the
-- table has no user_id — ownership is derived by joining through flows, so the
-- statement can never cross a tenant boundary. Idempotent for the same reason:
-- a rewritten row no longer matches old_name.
UPDATE assistants a
SET model_provider_name = sqlc.arg(new_name)
FROM flows f
WHERE a.flow_id = f.id
  AND f.user_id = sqlc.arg(user_id)
  AND a.model_provider_name = sqlc.arg(old_name)
  AND a.deleted_at IS NULL
  AND f.deleted_at IS NULL
RETURNING a.id, a.status, a.title, a.model, a.model_provider_name, a.language, a.functions, a.trace_id, a.flow_id, a.use_agents, a.msgchain_id, a.created_at, a.updated_at, a.deleted_at, a.model_provider_type, a.tool_call_id_template;

-- name: DeleteAssistant :one
UPDATE assistants
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

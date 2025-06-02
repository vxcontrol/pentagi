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
  title, status, model, model_provider, language, functions, prompts, flow_id, use_agents
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateAssistant :one
UPDATE assistants
SET title = $1, model = $2, language = $3, functions = $4, prompts = $5, trace_id = $6, msgchain_id = $7
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

-- name: DeleteAssistant :one
UPDATE assistants
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

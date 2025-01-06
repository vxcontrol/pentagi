-- name: GetFlows :many
SELECT
  f.*
FROM flows f
WHERE f.deleted_at IS NULL
ORDER BY f.created_at DESC;

-- name: GetUserFlows :many
SELECT
  f.*
FROM flows f
INNER JOIN users u ON f.user_id = u.id
WHERE f.user_id = $1 AND f.deleted_at IS NULL
ORDER BY f.created_at DESC;

-- name: GetFlow :one
SELECT
  f.*
FROM flows f
WHERE f.id = $1 AND f.deleted_at IS NULL;

-- name: GetUserFlow :one
SELECT
  f.*
FROM flows f
INNER JOIN users u ON f.user_id = u.id
WHERE f.id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL;

-- name: CreateFlow :one
INSERT INTO flows (
  title, status, model, model_provider, language, functions, prompts, user_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateFlow :one
UPDATE flows
SET title = $1, model = $2, language = $3, functions = $4, prompts = $5, trace_id = $6
WHERE id = $7
RETURNING *;

-- name: UpdateFlowStatus :one
UPDATE flows
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowTitle :one
UPDATE flows
SET title = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowLanguage :one
UPDATE flows
SET language = $1
WHERE id = $2
RETURNING *;

-- name: DeleteFlow :one
UPDATE flows
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

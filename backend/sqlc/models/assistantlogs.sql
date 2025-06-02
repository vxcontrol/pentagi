-- name: GetFlowAssistantLogs :many
SELECT
  al.*
FROM assistantlogs al
INNER JOIN assistants a ON al.assistant_id = a.id
INNER JOIN flows f ON al.flow_id = f.id
WHERE al.flow_id = $1 AND al.assistant_id = $2 AND f.deleted_at IS NULL AND a.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetUserFlowAssistantLogs :many
SELECT
  al.*
FROM assistantlogs al
INNER JOIN assistants a ON al.assistant_id = a.id
INNER JOIN flows f ON al.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE al.flow_id = $1 AND al.assistant_id = $2 AND f.user_id = $3 AND f.deleted_at IS NULL AND a.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetFlowAssistantLog :one
SELECT
  al.*
FROM assistantlogs al
INNER JOIN assistants a ON al.assistant_id = a.id
INNER JOIN flows f ON al.flow_id = f.id
WHERE al.id = $1 AND f.deleted_at IS NULL AND a.deleted_at IS NULL;

-- name: CreateAssistantLog :one
INSERT INTO assistantlogs (
  type,
  message,
  thinking,
  flow_id,
  assistant_id
)
VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CreateResultAssistantLog :one
INSERT INTO assistantlogs (
  type,
  message,
  thinking,
  result,
  result_format,
  flow_id,
  assistant_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateAssistantLog :one
UPDATE assistantlogs
SET type = $1, message = $2, thinking = $3, result = $4, result_format = $5
WHERE id = $6
RETURNING *;

-- name: UpdateAssistantLogContent :one
UPDATE assistantlogs
SET type = $1, message = $2, thinking = $3
WHERE id = $4
RETURNING *;

-- name: UpdateAssistantLogResult :one
UPDATE assistantlogs
SET result = $1, result_format = $2
WHERE id = $3
RETURNING *;

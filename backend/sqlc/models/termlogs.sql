-- name: GetFlowTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN flows f ON tl.flow_id = f.id
WHERE tl.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY tl.created_at ASC;

-- name: GetUserFlowTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN flows f ON tl.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE tl.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY tl.created_at ASC;

-- name: GetTaskTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN flows f ON tl.flow_id = f.id
WHERE tl.task_id = $1 AND f.deleted_at IS NULL
ORDER BY tl.created_at ASC;

-- name: GetSubtaskTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN flows f ON tl.flow_id = f.id
WHERE tl.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY tl.created_at ASC;

-- name: GetContainerTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN flows f ON tl.flow_id = f.id
WHERE tl.container_id = $1 AND f.deleted_at IS NULL
ORDER BY tl.created_at ASC;

-- name: GetTermLog :one
SELECT
  tl.*
FROM termlogs tl
WHERE tl.id = $1;

-- name: CreateTermLog :one
INSERT INTO termlogs (
  type,
  text,
  container_id,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

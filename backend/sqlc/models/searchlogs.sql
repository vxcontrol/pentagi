-- name: GetFlowSearchLogs :many
SELECT
  sl.*
FROM searchlogs sl
INNER JOIN flows f ON sl.flow_id = f.id
WHERE sl.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY sl.created_at ASC;

-- name: GetUserFlowSearchLogs :many
SELECT
  sl.*
FROM searchlogs sl
INNER JOIN flows f ON sl.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE sl.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY sl.created_at ASC;

-- name: GetTaskSearchLogs :many
SELECT
  sl.*
FROM searchlogs sl
INNER JOIN flows f ON sl.flow_id = f.id
INNER JOIN tasks t ON sl.task_id = t.id
WHERE sl.task_id = $1 AND f.deleted_at IS NULL
ORDER BY sl.created_at ASC;

-- name: GetSubtaskSearchLogs :many
SELECT
  sl.*
FROM searchlogs sl
INNER JOIN flows f ON sl.flow_id = f.id
INNER JOIN subtasks s ON sl.subtask_id = s.id
WHERE sl.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY sl.created_at ASC;

-- name: GetFlowSearchLog :one
SELECT
  sl.*
FROM searchlogs sl
INNER JOIN flows f ON sl.flow_id = f.id
WHERE sl.id = $1 AND sl.flow_id = $2 AND f.deleted_at IS NULL;

-- name: CreateSearchLog :one
INSERT INTO searchlogs (
  initiator,
  executor,
  engine,
  query,
  result,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

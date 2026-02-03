-- name: GetFlowScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
WHERE s.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetUserFlowScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE s.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetTaskScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
INNER JOIN tasks t ON s.task_id = t.id
WHERE s.task_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetSubtaskScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
INNER JOIN subtasks st ON s.subtask_id = st.id
WHERE s.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetScreenshot :one
SELECT
  s.*
FROM screenshots s
WHERE s.id = $1;

-- name: CreateScreenshot :one
INSERT INTO screenshots (
  name,
  url,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

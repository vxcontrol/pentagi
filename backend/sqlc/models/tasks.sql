-- name: GetFlowTasks :many
SELECT
  t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id
WHERE t.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY t.created_at ASC;

-- name: GetUserFlowTasks :many
SELECT
  t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE t.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY t.created_at ASC;

-- name: GetFlowTask :one
SELECT
  t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id
WHERE t.id = $1 AND t.flow_id = $2 AND f.deleted_at IS NULL;

-- name: GetUserFlowTask :one
SELECT
  t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE t.id = $1 AND t.flow_id = $2 AND f.user_id = $3 AND f.deleted_at IS NULL;

-- name: GetTask :one
SELECT
  t.*
FROM tasks t
WHERE t.id = $1;

-- name: CreateTask :one
INSERT INTO tasks (
  status,
  title,
  input,
  flow_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateTaskResult :one
UPDATE tasks
SET result = $1
WHERE id = $2
RETURNING *;

-- name: UpdateTaskFinishedResult :one
UPDATE tasks
SET status = 'finished', result = $1
WHERE id = $2
RETURNING *;

-- name: UpdateTaskFailedResult :one
UPDATE tasks
SET status = 'failed', result = $1
WHERE id = $2
RETURNING *;

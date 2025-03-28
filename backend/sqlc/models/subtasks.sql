-- name: GetFlowSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE t.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: GetFlowTaskSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE s.task_id = $1 AND t.flow_id = $2 AND f.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: GetUserFlowSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE t.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: GetUserFlowTaskSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE s.task_id = $1 AND t.flow_id = $2 AND f.user_id = $3 AND f.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: GetTaskSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE s.task_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetTaskPlannedSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE s.task_id = $1 AND (s.status = 'created' OR s.status = 'waiting') AND f.deleted_at IS NULL
ORDER BY s.id ASC;

-- name: GetTaskCompletedSubtasks :many
SELECT
  s.*
FROM subtasks s
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE s.task_id = $1 AND (s.status != 'created' AND s.status != 'waiting') AND f.deleted_at IS NULL
ORDER BY s.id ASC;

-- name: GetSubtask :one
SELECT
  s.*
FROM subtasks s
WHERE s.id = $1;

-- name: CreateSubtask :one
INSERT INTO subtasks (
  status,
  title,
  description,
  task_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateSubtaskStatus :one
UPDATE subtasks
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateSubtaskResult :one
UPDATE subtasks
SET result = $1
WHERE id = $2
RETURNING *;

-- name: UpdateSubtaskFinishedResult :one
UPDATE subtasks
SET status = 'finished', result = $1
WHERE id = $2
RETURNING *;

-- name: UpdateSubtaskFailedResult :one
UPDATE subtasks
SET status = 'failed', result = $1
WHERE id = $2
RETURNING *;

-- name: DeleteSubtask :exec
DELETE FROM subtasks
WHERE id = $1;

-- name: DeleteSubtasks :exec
DELETE FROM subtasks
WHERE id = ANY(@ids::BIGINT[]);

-- name: GetFlowMsgLogs :many
SELECT
  ml.*
FROM msglogs ml
INNER JOIN flows f ON ml.flow_id = f.id
WHERE ml.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY ml.created_at ASC;

-- name: GetUserFlowMsgLogs :many
SELECT
  ml.*
FROM msglogs ml
INNER JOIN flows f ON ml.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE ml.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY ml.created_at ASC;

-- name: GetTaskMsgLogs :many
SELECT
  ml.*
FROM msglogs ml
INNER JOIN tasks t ON ml.task_id = t.id
WHERE ml.task_id = $1
ORDER BY ml.created_at ASC;

-- name: GetSubtaskMsgLogs :many
SELECT
  ml.*
FROM msglogs ml
INNER JOIN subtasks s ON ml.subtask_id = s.id
WHERE ml.subtask_id = $1
ORDER BY ml.created_at ASC;

-- name: CreateMsgLog :one
INSERT INTO msglogs (
  type,
  message,
  thinking,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: CreateResultMsgLog :one
INSERT INTO msglogs (
  type,
  message,
  thinking,
  result,
  result_format,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateMsgLogResult :one
UPDATE msglogs
SET result = $1, result_format = $2
WHERE id = $3
RETURNING *;

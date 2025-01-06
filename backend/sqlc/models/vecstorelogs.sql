-- name: GetFlowVectorStoreLogs :many
SELECT
  vl.*
FROM vecstorelogs vl
INNER JOIN flows f ON vl.flow_id = f.id
WHERE vl.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY vl.created_at ASC;

-- name: GetUserFlowVectorStoreLogs :many
SELECT
  vl.*
FROM vecstorelogs vl
INNER JOIN flows f ON vl.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE vl.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY vl.created_at ASC;

-- name: GetTaskVectorStoreLogs :many
SELECT
  vl.*
FROM vecstorelogs vl
INNER JOIN flows f ON vl.flow_id = f.id
INNER JOIN tasks t ON vl.task_id = t.id
WHERE vl.task_id = $1 AND f.deleted_at IS NULL
ORDER BY vl.created_at ASC;

-- name: GetSubtaskVectorStoreLogs :many
SELECT
  vl.*
FROM vecstorelogs vl
INNER JOIN flows f ON vl.flow_id = f.id
INNER JOIN subtasks s ON vl.subtask_id = s.id
WHERE vl.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY vl.created_at ASC;

-- name: GetFlowVectorStoreLog :one
SELECT
  vl.*
FROM vecstorelogs vl
INNER JOIN flows f ON vl.flow_id = f.id
WHERE vl.id = $1 AND vl.flow_id = $2 AND f.deleted_at IS NULL;

-- name: CreateVectorStoreLog :one
INSERT INTO vecstorelogs (
  initiator,
  executor,
  filter,
  query,
  action,
  result,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

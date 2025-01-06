-- name: GetFlowAgentLogs :many
SELECT
  al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
WHERE al.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetUserFlowAgentLogs :many
SELECT
  al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE al.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetTaskAgentLogs :many
SELECT
  al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
INNER JOIN tasks t ON al.task_id = t.id
WHERE al.task_id = $1 AND f.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetSubtaskAgentLogs :many
SELECT
  al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
INNER JOIN subtasks s ON al.subtask_id = s.id
WHERE al.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY al.created_at ASC;

-- name: GetFlowAgentLog :one
SELECT
  al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
WHERE al.id = $1 AND al.flow_id = $2 AND f.deleted_at IS NULL;

-- name: CreateAgentLog :one
INSERT INTO agentlogs (
  initiator,
  executor,
  task,
  result,
  flow_id,
  task_id,
  subtask_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

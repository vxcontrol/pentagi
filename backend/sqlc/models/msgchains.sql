-- name: GetSubtaskMsgChains :many
SELECT
  mc.*
FROM msgchains mc
WHERE mc.subtask_id = $1
ORDER BY mc.created_at DESC;

-- name: GetSubtaskPrimaryMsgChains :many
SELECT
  mc.*
FROM msgchains mc
WHERE mc.subtask_id = $1 AND mc.type = 'primary_agent'
ORDER BY mc.created_at DESC;

-- name: GetSubtaskTypeMsgChains :many
SELECT
  mc.*
FROM msgchains mc
WHERE mc.subtask_id = $1 AND mc.type = $2
ORDER BY mc.created_at DESC;

-- name: GetTaskMsgChains :many
SELECT
  mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
WHERE mc.task_id = $1 OR s.task_id = $1
ORDER BY mc.created_at DESC;

-- name: GetTaskPrimaryMsgChains :many
SELECT
  mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
WHERE (mc.task_id = $1 OR s.task_id = $1) AND mc.type = 'primary_agent'
ORDER BY mc.created_at DESC;

-- name: GetTaskPrimaryMsgChainIDs :many
SELECT DISTINCT
  mc.id,
  mc.subtask_id
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
WHERE (mc.task_id = $1 OR s.task_id = $1) AND mc.type = 'primary_agent';

-- name: GetTaskTypeMsgChains :many
SELECT
  mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
WHERE (mc.task_id = $1 OR s.task_id = $1) AND mc.type = $2
ORDER BY mc.created_at DESC;

-- name: GetFlowMsgChains :many
SELECT
  mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id
WHERE mc.flow_id = $1 OR t.flow_id = $1
ORDER BY mc.created_at DESC;

-- name: GetFlowTypeMsgChains :many
SELECT
  mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id
WHERE (mc.flow_id = $1 OR t.flow_id = $1) AND mc.type = $2
ORDER BY mc.created_at DESC;

-- name: GetFlowTaskTypeLastMsgChain :one
SELECT
  mc.*
FROM msgchains mc
WHERE mc.flow_id = $1 AND (mc.task_id = $2 OR $2 IS NULL) AND mc.type = $3
ORDER BY mc.created_at DESC
LIMIT 1;

-- name: GetMsgChain :one
SELECT
  mc.*
FROM msgchains mc
WHERE mc.id = $1;

-- name: CreateMsgChain :one
INSERT INTO msgchains (
  type,
  model,
  model_provider,
  usage_in,
  usage_out,
  chain,
  flow_id,
  task_id,
  subtask_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateMsgChain :one
UPDATE msgchains
SET chain = $1
WHERE id = $2
RETURNING *;

-- name: UpdateMsgChainUsage :one
UPDATE msgchains
SET usage_in = usage_in + $1, usage_out = usage_out + $2
WHERE id = $3
RETURNING *;

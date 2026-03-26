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
  usage_cache_in,
  usage_cache_out,
  usage_cost_in,
  usage_cost_out,
  duration_seconds,
  chain,
  flow_id,
  task_id,
  subtask_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: UpdateMsgChain :one
UPDATE msgchains
SET chain = $1, duration_seconds = duration_seconds + $2
WHERE id = $3
RETURNING *;

-- name: UpdateMsgChainUsage :one
UPDATE msgchains
SET 
  usage_in = usage_in + $1, 
  usage_out = usage_out + $2,
  usage_cache_in = usage_cache_in + $3,
  usage_cache_out = usage_cache_out + $4,
  usage_cost_in = usage_cost_in + $5,
  usage_cost_out = usage_cost_out + $6,
  duration_seconds = duration_seconds + $7
WHERE id = $8
RETURNING *;

-- name: GetFlowUsageStats :one
SELECT
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE (mc.flow_id = $1 OR t.flow_id = $1) AND f.deleted_at IS NULL;

-- name: GetTaskUsageStats :one
SELECT
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON mc.task_id = t.id OR s.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE (mc.task_id = $1 OR s.task_id = $1) AND f.deleted_at IS NULL;

-- name: GetSubtaskUsageStats :one
SELECT
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE mc.subtask_id = $1 AND f.deleted_at IS NULL;

-- name: GetAllFlowsUsageStats :many
SELECT
  COALESCE(mc.flow_id, t.flow_id) AS flow_id,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL
GROUP BY COALESCE(mc.flow_id, t.flow_id)
ORDER BY COALESCE(mc.flow_id, t.flow_id);

-- name: GetUsageStatsByProvider :many
SELECT
  mc.model_provider,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1
GROUP BY mc.model_provider
ORDER BY mc.model_provider;

-- name: GetUsageStatsByModel :many
SELECT
  mc.model,
  mc.model_provider,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1
GROUP BY mc.model, mc.model_provider
ORDER BY mc.model, mc.model_provider;

-- name: GetUsageStatsByType :many
SELECT
  mc.type,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1
GROUP BY mc.type
ORDER BY mc.type;

-- name: GetUsageStatsByTypeForFlow :many
SELECT
  mc.type,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE (mc.flow_id = $1 OR t.flow_id = $1) AND f.deleted_at IS NULL
GROUP BY mc.type
ORDER BY mc.type;

-- name: GetUsageStatsByDayLastWeek :many
SELECT
  DATE(mc.created_at) AS date,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE mc.created_at >= NOW() - INTERVAL '7 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(mc.created_at)
ORDER BY date DESC;

-- name: GetUsageStatsByDayLastMonth :many
SELECT
  DATE(mc.created_at) AS date,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE mc.created_at >= NOW() - INTERVAL '30 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(mc.created_at)
ORDER BY date DESC;

-- name: GetUsageStatsByDayLast3Months :many
SELECT
  DATE(mc.created_at) AS date,
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE mc.created_at >= NOW() - INTERVAL '90 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(mc.created_at)
ORDER BY date DESC;

-- name: GetUserTotalUsageStats :one
SELECT
  COALESCE(SUM(mc.usage_in), 0)::bigint AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0)::bigint AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0)::bigint AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0)::bigint AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0)::double precision AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0)::double precision AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1;

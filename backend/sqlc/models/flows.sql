-- name: GetFlows :many
SELECT
  f.*
FROM flows f
WHERE f.deleted_at IS NULL
ORDER BY f.created_at DESC;

-- name: GetUserFlows :many
SELECT
  f.*
FROM flows f
INNER JOIN users u ON f.user_id = u.id
WHERE f.user_id = $1 AND f.deleted_at IS NULL
ORDER BY f.created_at DESC;

-- name: GetFlow :one
SELECT
  f.*
FROM flows f
WHERE f.id = $1 AND f.deleted_at IS NULL;

-- name: GetUserFlow :one
SELECT
  f.*
FROM flows f
INNER JOIN users u ON f.user_id = u.id
WHERE f.id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL;

-- name: CreateFlow :one
INSERT INTO flows (
  title, status, model, model_provider_name, model_provider_type, language, tool_call_id_template, functions, user_id
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateFlow :one
UPDATE flows
SET title = $1, model = $2, language = $3, tool_call_id_template = $4, functions = $5, trace_id = $6
WHERE id = $7
RETURNING *;

-- name: UpdateFlowStatus :one
UPDATE flows
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowTitle :one
UPDATE flows
SET title = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowLanguage :one
UPDATE flows
SET language = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowToolCallIDTemplate :one
UPDATE flows
SET tool_call_id_template = $1
WHERE id = $2
RETURNING *;

-- name: UpdateFlowProvider :one
UPDATE flows
SET model_provider_name = $1, model_provider_type = $2, tool_call_id_template = $3, model = $4
WHERE id = $5
RETURNING *;

-- name: UpdateFlowsProviderNameByOldName :many
-- Bulk-renames every flow row of a user still pointing at a provider's old name
-- (the user renamed a custom LLM provider, or deleted one and the reference is
-- reset to the built-in name of its type). Matching on old_name makes the
-- statement idempotent: once rewritten, a row no longer matches, so it is safe
-- to call unconditionally and to retry.
UPDATE flows
SET model_provider_name = sqlc.arg(new_name)
WHERE user_id = sqlc.arg(user_id) AND model_provider_name = sqlc.arg(old_name) AND deleted_at IS NULL
RETURNING *;

-- name: DeleteFlow :one
UPDATE flows
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- ==================== Flows Analytics Queries ====================

-- name: GetFlowStats :one
-- Get total count of tasks, subtasks, and assistants for a specific flow
SELECT
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.id = $1 AND f.deleted_at IS NULL;

-- name: GetUserTotalFlowsStats :one
-- Get total count of flows, tasks, subtasks, and assistants for a user
SELECT
  COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.user_id = $1 AND f.deleted_at IS NULL;

-- name: GetFlowsStatsByDayLastWeek :many
-- Get flows stats by day for the last week
SELECT
  DATE(f.created_at) AS date,
  COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.created_at >= NOW() - INTERVAL '7 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(f.created_at)
ORDER BY date DESC;

-- name: GetFlowsStatsByDayLastMonth :many
-- Get flows stats by day for the last month
SELECT
  DATE(f.created_at) AS date,
  COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.created_at >= NOW() - INTERVAL '30 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(f.created_at)
ORDER BY date DESC;

-- name: GetFlowsStatsByDayLast3Months :many
-- Get flows stats by day for the last 3 months
SELECT
  DATE(f.created_at) AS date,
  COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.created_at >= NOW() - INTERVAL '90 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(f.created_at)
ORDER BY date DESC;

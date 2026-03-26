-- name: GetSubtaskToolcalls :many
SELECT
  tc.*
FROM toolcalls tc
INNER JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE tc.subtask_id = $1 AND f.deleted_at IS NULL
ORDER BY tc.created_at DESC;

-- name: GetCallToolcall :one
SELECT
  tc.*
FROM toolcalls tc
WHERE tc.call_id = $1;

-- name: CreateToolcall :one
INSERT INTO toolcalls (
  call_id,
  status,
  name,
  args,
  flow_id,
  task_id,
  subtask_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateToolcallStatus :one
UPDATE toolcalls
SET 
  status = $1,
  duration_seconds = duration_seconds + $2
WHERE id = $3
RETURNING *;

-- name: UpdateToolcallFinishedResult :one
UPDATE toolcalls
SET 
  status = 'finished', 
  result = $1,
  duration_seconds = duration_seconds + $2
WHERE id = $3
RETURNING *;

-- name: UpdateToolcallFailedResult :one
UPDATE toolcalls
SET 
  status = 'failed', 
  result = $1,
  duration_seconds = duration_seconds + $2
WHERE id = $3
RETURNING *;

-- ==================== Toolcalls Analytics Queries ====================

-- name: GetFlowToolcallsStats :one
-- Get total execution time and count of toolcalls for a specific flow
SELECT
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN tasks t ON tc.task_id = t.id
LEFT JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN flows f ON tc.flow_id = f.id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL 
  AND (tc.task_id IS NULL OR t.id IS NOT NULL)
  AND (tc.subtask_id IS NULL OR s.id IS NOT NULL);

-- name: GetTaskToolcallsStats :one
-- Get total execution time and count of toolcalls for a specific task
SELECT
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN tasks t ON tc.task_id = t.id OR s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE (tc.task_id = $1 OR s.task_id = $1) AND f.deleted_at IS NULL
  AND (tc.subtask_id IS NULL OR s.id IS NOT NULL);

-- name: GetSubtaskToolcallsStats :one
-- Get total execution time and count of toolcalls for a specific subtask
SELECT
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
INNER JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE tc.subtask_id = $1 AND f.deleted_at IS NULL AND s.id IS NOT NULL AND t.id IS NOT NULL;

-- name: GetAllFlowsToolcallsStats :many
-- Get toolcalls stats for all flows
SELECT
  COALESCE(tc.flow_id, t.flow_id) AS flow_id,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL
GROUP BY COALESCE(tc.flow_id, t.flow_id)
ORDER BY COALESCE(tc.flow_id, t.flow_id);

-- name: GetToolcallsStatsByFunction :many
-- Get toolcalls stats grouped by function name for a user
SELECT
  tc.name AS function_name,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds,
  COALESCE(AVG(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE NULL END), 0.0)::double precision AS avg_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1
GROUP BY tc.name
ORDER BY total_duration_seconds DESC;

-- name: GetToolcallsStatsByFunctionForFlow :many
-- Get toolcalls stats grouped by function name for a specific flow
SELECT
  tc.name AS function_name,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds,
  COALESCE(AVG(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE NULL END), 0.0)::double precision AS avg_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE (tc.flow_id = $1 OR t.flow_id = $1) AND f.deleted_at IS NULL
GROUP BY tc.name
ORDER BY total_duration_seconds DESC;

-- name: GetToolcallsStatsByDayLastWeek :many
-- Get toolcalls stats by day for the last week
SELECT
  DATE(tc.created_at) AS date,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE tc.created_at >= NOW() - INTERVAL '7 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(tc.created_at)
ORDER BY date DESC;

-- name: GetToolcallsStatsByDayLastMonth :many
-- Get toolcalls stats by day for the last month
SELECT
  DATE(tc.created_at) AS date,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE tc.created_at >= NOW() - INTERVAL '30 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(tc.created_at)
ORDER BY date DESC;

-- name: GetToolcallsStatsByDayLast3Months :many
-- Get toolcalls stats by day for the last 3 months
SELECT
  DATE(tc.created_at) AS date,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE tc.created_at >= NOW() - INTERVAL '90 days' AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(tc.created_at)
ORDER BY date DESC;

-- name: GetUserTotalToolcallsStats :one
-- Get total toolcalls stats for a user
SELECT
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
LEFT JOIN subtasks s ON tc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR tc.task_id = t.id
INNER JOIN flows f ON (tc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL AND f.user_id = $1
  AND (tc.task_id IS NULL OR t.id IS NOT NULL)
  AND (tc.subtask_id IS NULL OR s.id IS NOT NULL);

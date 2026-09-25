-- name: GetFlowToolcalls :many
SELECT
  tc.*
FROM toolcalls tc
INNER JOIN flows f ON tc.flow_id = f.id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY tc.created_at ASC;

-- name: GetFlowToolcall :one
SELECT
  tc.*
FROM toolcalls tc
INNER JOIN flows f ON tc.flow_id = f.id
WHERE tc.id = $1 AND tc.flow_id = $2 AND f.deleted_at IS NULL;

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
INNER JOIN flows f ON f.id = tc.flow_id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL;

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
  tc.flow_id AS flow_id,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
INNER JOIN flows f ON f.id = tc.flow_id
WHERE f.deleted_at IS NULL
GROUP BY tc.flow_id
ORDER BY tc.flow_id;

-- name: GetToolcallsStatsByFunction :many
-- Get toolcalls stats grouped by function name for a user
SELECT
  tc.name AS function_name,
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds,
  COALESCE(AVG(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE NULL END), 0.0)::double precision AS avg_duration_seconds
FROM toolcalls tc
INNER JOIN flows f ON f.id = tc.flow_id
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
INNER JOIN flows f ON f.id = tc.flow_id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL
GROUP BY tc.name
ORDER BY total_duration_seconds DESC;

-- name: GetToolcallsStatsByDay :many
-- One dense row per calendar day in the caller's timezone, zeros included.
WITH bounds AS (
  SELECT
    ((NOW() AT TIME ZONE sqlc.arg(tz)::text)::date - sqlc.arg(days)::int) AS first_day,
    (NOW() AT TIME ZONE sqlc.arg(tz)::text)::date AS last_day
),
days AS (
  SELECT generate_series(b.first_day, b.last_day, INTERVAL '1 day')::date AS day
  FROM bounds b
),
agg AS (
  SELECT
    (tc.created_at AT TIME ZONE sqlc.arg(tz)::text)::date AS day,
    COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END)::bigint AS total_count,
    SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END)::double precision AS total_duration_seconds
  FROM toolcalls tc
  INNER JOIN flows f ON f.id = tc.flow_id
  WHERE f.deleted_at IS NULL
    AND f.user_id = sqlc.arg(user_id)
    -- The coarse bounds keep the index; the exact test repeats the grouping
    -- expression, because an ambiguous local midnight resolves to the later
    -- instant and would drop the earlier hour of that day.
    AND tc.created_at >= (((SELECT first_day FROM bounds) - 1)::timestamp AT TIME ZONE sqlc.arg(tz)::text)
    AND tc.created_at < (((SELECT last_day FROM bounds) + 2)::timestamp AT TIME ZONE sqlc.arg(tz)::text)
    AND (tc.created_at AT TIME ZONE sqlc.arg(tz)::text)::date
        BETWEEN (SELECT first_day FROM bounds) AND (SELECT last_day FROM bounds)
  GROUP BY 1
)
SELECT
  (d.day::timestamp AT TIME ZONE sqlc.arg(tz)::text) AS date,
  COALESCE(a.total_count, 0)::bigint AS total_count,
  COALESCE(a.total_duration_seconds, 0.0)::double precision AS total_duration_seconds
FROM days d
LEFT JOIN agg a ON a.day = d.day
ORDER BY date DESC;


-- name: GetUserTotalToolcallsStats :one
-- Get total toolcalls stats for a user
SELECT
  COALESCE(COUNT(CASE WHEN tc.status IN ('finished', 'failed') THEN 1 END), 0)::bigint AS total_count,
  COALESCE(SUM(CASE WHEN tc.status IN ('finished', 'failed') THEN tc.duration_seconds ELSE 0 END), 0.0)::double precision AS total_duration_seconds
FROM toolcalls tc
INNER JOIN flows f ON f.id = tc.flow_id
WHERE f.deleted_at IS NULL AND f.user_id = $1;

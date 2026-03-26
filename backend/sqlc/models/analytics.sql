-- name: GetFlowsForPeriodLastWeek :many
-- Get flow IDs created in the last week for analytics
SELECT id, title
FROM flows
WHERE created_at >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL AND user_id = $1
ORDER BY created_at DESC;

-- name: GetFlowsForPeriodLastMonth :many
-- Get flow IDs created in the last month for analytics
SELECT id, title
FROM flows
WHERE created_at >= NOW() - INTERVAL '30 days' AND deleted_at IS NULL AND user_id = $1
ORDER BY created_at DESC;

-- name: GetFlowsForPeriodLast3Months :many
-- Get flow IDs created in the last 3 months for analytics
SELECT id, title
FROM flows
WHERE created_at >= NOW() - INTERVAL '90 days' AND deleted_at IS NULL AND user_id = $1
ORDER BY created_at DESC;

-- name: GetTasksForFlow :many
-- Get all tasks for a flow
SELECT id, title, created_at, updated_at
FROM tasks
WHERE flow_id = $1
ORDER BY id ASC;

-- name: GetSubtasksForTasks :many
-- Get all subtasks for multiple tasks
SELECT id, task_id, title, status, created_at, updated_at
FROM subtasks
WHERE task_id = ANY(@task_ids::BIGINT[])
ORDER BY id ASC;

-- name: GetMsgchainsForFlow :many
-- Get all msgchains for a flow (including task and subtask level)
SELECT id, type, flow_id, task_id, subtask_id, duration_seconds, created_at, updated_at
FROM msgchains
WHERE flow_id = $1
ORDER BY created_at ASC;

-- name: GetToolcallsForFlow :many
-- Get all toolcalls for a flow
SELECT tc.id, tc.status, tc.flow_id, tc.task_id, tc.subtask_id, tc.duration_seconds, tc.created_at, tc.updated_at
FROM toolcalls tc
LEFT JOIN tasks t ON tc.task_id = t.id
LEFT JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN flows f ON tc.flow_id = f.id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL
  AND (tc.task_id IS NULL OR t.id IS NOT NULL)
  AND (tc.subtask_id IS NULL OR s.id IS NOT NULL)
ORDER BY tc.created_at ASC;

-- name: GetAssistantsCountForFlow :one
-- Get total count of assistants for a specific flow
SELECT COALESCE(COUNT(id), 0)::bigint AS total_assistants_count
FROM assistants
WHERE flow_id = $1 AND deleted_at IS NULL;

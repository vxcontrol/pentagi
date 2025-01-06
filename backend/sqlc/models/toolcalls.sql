-- name: GetSubtaskToolcalls :many
SELECT
  tc.*
FROM toolcalls tc
INNER JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN tasks t ON s.task_id = t.id
INNER JOIN flows f ON t.flow_id = f.id
WHERE tc.subtask_id = $1
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
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateToolcallFinishedResult :one
UPDATE toolcalls
SET status = 'finished', result = $1
WHERE id = $2
RETURNING *;

-- name: UpdateToolcallFailedResult :one
UPDATE toolcalls
SET status = 'failed', result = $1
WHERE id = $2
RETURNING *;

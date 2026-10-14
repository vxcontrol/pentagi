-- name: CreateFlowInputUpdate :one
INSERT INTO flow_input_updates (flow_id, input)
VALUES ($1, $2)
RETURNING id, flow_id, task_id, input, status, created_at, updated_at;

-- name: ClaimNextFlowInputUpdate :one
UPDATE flow_input_updates AS current_update
SET status = 'processing', updated_at = CURRENT_TIMESTAMP
WHERE current_update.id = (
  SELECT queued_update.id FROM flow_input_updates AS queued_update
  WHERE queued_update.flow_id = $1 AND queued_update.status = 'queued'
  ORDER BY id
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING id, flow_id, task_id, input, status, created_at, updated_at;

-- name: UpdateFlowInputUpdateStatus :exec
UPDATE flow_input_updates
SET status = $3, updated_at = CURRENT_TIMESTAMP
WHERE flow_id = $1 AND id = $2;

-- name: CreateTaskForFlowInputUpdate :one
WITH eligible_update AS (
  SELECT eligible_update_row.id, eligible_update_row.flow_id
  FROM flow_input_updates AS eligible_update_row
  WHERE eligible_update_row.id = $5 AND eligible_update_row.flow_id = $4
    AND eligible_update_row.status = 'processing' AND eligible_update_row.task_id IS NULL
  FOR UPDATE
), created_task AS (
  INSERT INTO tasks (status, title, input, flow_id)
  SELECT $1, $2, $3, eligible_update.flow_id
  FROM eligible_update
  RETURNING id, status, title, input, result, flow_id, created_at, updated_at
), linked_update AS (
  UPDATE flow_input_updates
  SET task_id = created_task.id, status = 'completed', updated_at = CURRENT_TIMESTAMP
  FROM created_task
  WHERE flow_input_updates.id = $5
  RETURNING flow_input_updates.id
)
SELECT created_task.* FROM created_task INNER JOIN linked_update ON TRUE;

-- name: RecoverFlowInputUpdates :exec
UPDATE flow_input_updates
SET status = 'queued', updated_at = CURRENT_TIMESTAMP
WHERE flow_id = $1 AND status = 'processing';

-- name: FailFlowInputUpdates :exec
UPDATE flow_input_updates
SET status = 'failed', updated_at = CURRENT_TIMESTAMP
WHERE flow_id = $1 AND status IN ('queued', 'processing');

-- name: GetFlowInputUpdateCounts :one
SELECT
  count(*) FILTER (WHERE status = 'queued')::BIGINT AS queued,
  count(*) FILTER (WHERE status = 'processing')::BIGINT AS processing,
  count(*) FILTER (WHERE status = 'failed')::BIGINT AS failed
FROM flow_input_updates
WHERE flow_id = $1;

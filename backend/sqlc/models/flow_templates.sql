-- name: GetFlowTemplate :one
SELECT * FROM flow_templates
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: GetFlowTemplatesByUserID :many
SELECT * FROM flow_templates
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateFlowTemplate :one
INSERT INTO flow_templates (
  user_id,
  title,
  text
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: UpdateFlowTemplate :one
UPDATE flow_templates
SET 
  title = $3,
  text = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteFlowTemplate :exec
DELETE FROM flow_templates
WHERE id = $1 AND user_id = $2;

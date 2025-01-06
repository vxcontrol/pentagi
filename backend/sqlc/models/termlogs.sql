-- name: GetFlowTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN containers c ON tl.container_id = c.id
INNER JOIN flows f ON c.flow_id = f.id
WHERE c.flow_id = $1
ORDER BY tl.created_at ASC;

-- name: GetUserFlowTermLogs :many
SELECT
  tl.*
FROM termlogs tl
INNER JOIN containers c ON tl.container_id = c.id
INNER JOIN flows f ON c.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE c.flow_id = $1 AND f.user_id = $2
ORDER BY tl.created_at ASC;

-- name: GetTermLog :one
SELECT
  tl.*
FROM termlogs tl
WHERE tl.id = $1;

-- name: CreateTermLog :one
INSERT INTO termlogs (
  type,
  text,
  container_id
)
VALUES (
  $1, $2, $3
)
RETURNING *;

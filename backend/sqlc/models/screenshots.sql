-- name: GetFlowScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
WHERE s.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetUserFlowScreenshots :many
SELECT
  s.*
FROM screenshots s
INNER JOIN flows f ON s.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE s.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: GetScreenshot :one
SELECT
  s.*
FROM screenshots s
WHERE s.id = $1;

-- name: CreateScreenshot :one
INSERT INTO screenshots (
  name,
  url,
  flow_id
)
VALUES (
  $1, $2, $3
)
RETURNING *;

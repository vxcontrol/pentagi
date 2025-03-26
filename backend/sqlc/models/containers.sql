-- name: GetContainers :many
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
WHERE f.deleted_at IS NULL
ORDER BY c.created_at DESC;

-- name: GetUserContainers :many
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE f.user_id = $1 AND f.deleted_at IS NULL
ORDER BY c.created_at DESC;

-- name: GetRunningContainers :many
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
WHERE c.status = 'running' AND f.deleted_at IS NULL
ORDER BY c.created_at DESC;

-- name: GetFlowContainers :many
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
WHERE c.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY c.created_at DESC;

-- name: GetFlowPrimaryContainer :one
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
WHERE c.flow_id = $1 AND c.type = 'primary' AND f.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT 1;

-- name: GetUserFlowContainers :many
SELECT
  c.*
FROM containers c
INNER JOIN flows f ON c.flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE c.flow_id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL
ORDER BY c.created_at DESC;

-- name: CreateContainer :one
INSERT INTO containers (
  type, name, image, status, flow_id, local_id, local_dir
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT ON CONSTRAINT containers_local_id_unique
DO UPDATE SET
  type = EXCLUDED.type,
  name = EXCLUDED.name,
  image = EXCLUDED.image,
  status = EXCLUDED.status,
  flow_id = EXCLUDED.flow_id,
  local_dir = EXCLUDED.local_dir
RETURNING *;

-- name: UpdateContainerStatusLocalID :one
UPDATE containers
SET status = $1, local_id = $2
WHERE id = $3
RETURNING *;

-- name: UpdateContainerStatus :one
UPDATE containers
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateContainerLocalID :one
UPDATE containers
SET local_id = $1
WHERE id = $2
RETURNING *;

-- name: UpdateContainerLocalDir :one
UPDATE containers
SET local_dir = $1
WHERE id = $2
RETURNING *;

-- name: UpdateContainerImage :one
UPDATE containers
SET image = $1
WHERE id = $2
RETURNING *;

-- name: GetProviders :many
SELECT
  p.*
FROM providers p
WHERE p.deleted_at IS NULL
ORDER BY p.created_at ASC;

-- name: GetProvidersByType :many
SELECT
  p.*
FROM providers p
WHERE p.type = $1 AND p.deleted_at IS NULL
ORDER BY p.created_at ASC;

-- name: GetProvider :one
SELECT
  p.*
FROM providers p
WHERE p.id = $1 AND p.deleted_at IS NULL;

-- name: GetUserProvider :one
SELECT
  p.*
FROM providers p
INNER JOIN users u ON p.user_id = u.id
WHERE p.id = $1 AND p.user_id = $2 AND p.deleted_at IS NULL;

-- name: GetUserProviders :many
SELECT
  p.*
FROM providers p
INNER JOIN users u ON p.user_id = u.id
WHERE p.user_id = $1 AND p.deleted_at IS NULL
ORDER BY p.created_at ASC;

-- name: GetUserProvidersByType :many
SELECT
  p.*
FROM providers p
INNER JOIN users u ON p.user_id = u.id
WHERE p.user_id = $1 AND p.type = $2 AND p.deleted_at IS NULL
ORDER BY p.created_at ASC;

-- name: GetUserProviderByName :one
SELECT
  p.*
FROM providers p
INNER JOIN users u ON p.user_id = u.id
WHERE p.name = $1 AND p.user_id = $2 AND p.deleted_at IS NULL;

-- name: CreateProvider :one
INSERT INTO providers (
  user_id,
  type,
  name,
  config
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateProvider :one
UPDATE providers
SET config = $2, name = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserProvider :one
UPDATE providers
SET config = $3, name = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProvider :one
UPDATE providers
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUserProvider :one
UPDATE providers
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetAPITokens :many
SELECT
  t.*
FROM api_tokens t
WHERE t.deleted_at IS NULL
ORDER BY t.created_at DESC;

-- name: GetAPIToken :one
SELECT
  t.*
FROM api_tokens t
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: GetAPITokenByTokenID :one
SELECT
  t.*
FROM api_tokens t
WHERE t.token_id = $1 AND t.deleted_at IS NULL;

-- name: GetUserAPITokens :many
SELECT
  t.*
FROM api_tokens t
INNER JOIN users u ON t.user_id = u.id
WHERE t.user_id = $1 AND t.deleted_at IS NULL
ORDER BY t.created_at DESC;

-- name: GetUserAPIToken :one
SELECT
  t.*
FROM api_tokens t
INNER JOIN users u ON t.user_id = u.id
WHERE t.id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL;

-- name: GetUserAPITokenByTokenID :one
SELECT
  t.*
FROM api_tokens t
INNER JOIN users u ON t.user_id = u.id
WHERE t.token_id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL;

-- name: CreateAPIToken :one
INSERT INTO api_tokens (
  token_id,
  user_id,
  role_id,
  name,
  ttl,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateAPIToken :one
UPDATE api_tokens
SET name = $2, status = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserAPIToken :one
UPDATE api_tokens
SET name = $3, status = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteAPIToken :one
UPDATE api_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUserAPIToken :one
UPDATE api_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteUserAPITokenByTokenID :one
UPDATE api_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE token_id = $1 AND user_id = $2
RETURNING *;

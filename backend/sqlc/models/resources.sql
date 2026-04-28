-- name: GetUserResourcesRoot :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE user_id = sqlc.arg(user_id) AND path NOT LIKE '%/%'
ORDER BY updated_at DESC, name ASC;

-- name: GetUserResourcesInDir :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE user_id = sqlc.arg(user_id)
  AND ((path = sqlc.arg(dir_path) AND is_dir = true)
    OR (path LIKE sqlc.arg(child_prefix) AND path NOT LIKE sqlc.arg(deep_prefix)))
ORDER BY updated_at DESC, name ASC;

-- name: GetUserResourcesRecursive :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE user_id = sqlc.arg(user_id)
  AND (path = sqlc.arg(dir_path) OR path LIKE sqlc.arg(child_prefix))
ORDER BY updated_at DESC, name ASC;

-- name: GetAllResourcesRoot :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE path NOT LIKE '%/%'
ORDER BY updated_at DESC, name ASC;

-- name: GetAllResourcesInDir :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE (path = sqlc.arg(dir_path) AND is_dir = true)
   OR (path LIKE sqlc.arg(child_prefix) AND path NOT LIKE sqlc.arg(deep_prefix))
ORDER BY updated_at DESC, name ASC;

-- name: GetAllResourcesRecursive :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE path = sqlc.arg(dir_path) OR path LIKE sqlc.arg(child_prefix)
ORDER BY updated_at DESC, name ASC;

-- name: GetUserResourceByID :one
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE id = sqlc.arg(id);

-- name: GetUserResourcesByIDs :many
SELECT id, user_id, hash, name, path, size, is_dir, created_at, updated_at
FROM user_resources
WHERE id = ANY(sqlc.arg(ids)::bigint[])
ORDER BY id ASC;

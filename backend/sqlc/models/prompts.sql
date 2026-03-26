-- name: GetPrompts :many
SELECT
  p.*
FROM prompts p
ORDER BY p.user_id ASC, p.type ASC;

-- name: GetUserPrompts :many
SELECT
  p.*
FROM prompts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.user_id = $1
ORDER BY p.type ASC;

-- name: GetUserPrompt :one
SELECT
  p.*
FROM prompts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.id = $1 AND p.user_id = $2;

-- name: GetUserPromptByType :one
SELECT
  p.*
FROM prompts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.type = $1 AND p.user_id = $2
LIMIT 1;

-- name: CreateUserPrompt :one
INSERT INTO prompts (
  type,
  user_id,
  prompt
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdatePrompt :one
UPDATE prompts
SET prompt = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUserPrompt :one
UPDATE prompts
SET prompt = $1
WHERE id = $2 AND user_id = $3
RETURNING *;

-- name: UpdateUserPromptByType :one
UPDATE prompts
SET prompt = $1
WHERE type = $2 AND user_id = $3
RETURNING *;

-- name: DeletePrompt :exec
DELETE FROM prompts
WHERE id = $1;

-- name: DeleteUserPrompt :exec
DELETE FROM prompts
WHERE id = $1 AND user_id = $2;

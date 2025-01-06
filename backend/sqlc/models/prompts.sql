-- name: GetUserPrompts :many
SELECT
  p.*
FROM prompts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.user_id = $1
ORDER BY p.type ASC;

-- name: GetUserTypePrompt :one
SELECT
  p.*
FROM prompts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.type = $1 AND p.user_id = $2;

-- name: CreateUserPrompt :one
INSERT INTO prompts (
  type,
  user_id,
  prompt
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateUserTypePrompt :one
UPDATE prompts
SET prompt = $1
WHERE type = $2 AND user_id = $3
RETURNING *;

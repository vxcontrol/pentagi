-- name: GetUsers :many
SELECT
  u.*,
  r.name AS role_name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM users u
INNER JOIN roles r ON u.role_id = r.id
ORDER BY u.created_at DESC;

-- name: GetUser :one
SELECT
  u.*,
  r.name AS role_name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM users u
INNER JOIN roles r ON u.role_id = r.id
WHERE u.id = $1;

-- name: GetUserByHash :one
SELECT
  u.*,
  r.name AS role_name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM users u
INNER JOIN roles r ON u.role_id = r.id
WHERE u.hash = $1;

-- name: CreateUser :one
INSERT INTO users (
  type,
  mail,
  name,
  password,
  status,
  role_id,
  password_change_required
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET status = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUserName :one
UPDATE users
SET name = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUserPasswordChangeRequired :one
UPDATE users
SET password_change_required = $1
WHERE id = $2
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users
SET role_id = $1
WHERE id = $2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

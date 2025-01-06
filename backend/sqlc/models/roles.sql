-- name: GetRoles :many
SELECT
  r.id,
  r.name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM roles r
ORDER BY r.id ASC;

-- name: GetRole :one
SELECT
  r.id,
  r.name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM roles r
WHERE r.id = $1;

-- name: GetRoleByName :one
SELECT
  r.id,
  r.name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM roles r
WHERE r.name = $1;

-- name: GetUserPreferencesByUserID :one
SELECT * FROM user_preferences
WHERE user_id = $1 LIMIT 1;

-- name: CreateUserPreferences :one
INSERT INTO user_preferences (
  user_id,
  preferences
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: UpdateUserPreferences :one
UPDATE user_preferences
SET preferences = $2
WHERE user_id = $1
RETURNING *;

-- name: DeleteUserPreferences :exec
DELETE FROM user_preferences
WHERE user_id = $1;

-- name: UpsertUserPreferences :one
INSERT INTO user_preferences (
  user_id,
  preferences
) VALUES (
  $1,
  $2
)
ON CONFLICT (user_id) DO UPDATE
SET preferences = EXCLUDED.preferences
RETURNING *;

-- name: AddFavoriteFlow :one
INSERT INTO user_preferences (user_id, preferences)
VALUES (
  sqlc.arg(user_id)::bigint,
  jsonb_build_object('favoriteFlows', jsonb_build_array(sqlc.arg(flow_id)::bigint))
)
ON CONFLICT (user_id) DO UPDATE
SET preferences = jsonb_set(
  user_preferences.preferences,
  '{favoriteFlows}',
  CASE
    WHEN user_preferences.preferences->'favoriteFlows' @> to_jsonb(sqlc.arg(flow_id)::bigint) THEN
      user_preferences.preferences->'favoriteFlows'
    ELSE
      user_preferences.preferences->'favoriteFlows' || to_jsonb(sqlc.arg(flow_id)::bigint)
  END
)
RETURNING *;

-- name: DeleteFavoriteFlow :one
UPDATE user_preferences
SET preferences = jsonb_set(
  preferences,
  '{favoriteFlows}',
  (
    SELECT COALESCE(jsonb_agg(elem), '[]'::jsonb)
    FROM jsonb_array_elements(preferences->'favoriteFlows') elem
    WHERE elem::text::bigint != sqlc.arg(flow_id)::bigint
  )
)
WHERE user_id = sqlc.arg(user_id)::bigint
RETURNING *;

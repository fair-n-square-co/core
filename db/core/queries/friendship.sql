-- name: CreateFriendship :one
INSERT INTO friendship (user_a, user_b, status, status_actor_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFriendshipByID :one
SELECT * FROM friendship
WHERE id = $1;

-- name: GetFriendshipByPair :one
SELECT * FROM friendship
WHERE user_a = $1 AND user_b = $2;

-- name: UpdateFriendshipStatus :one
UPDATE friendship
SET status = $2, status_actor_id = $3
WHERE id = $1
RETURNING *;

-- name: ListFriendshipsForUser :many
SELECT * FROM friendship
WHERE (user_a = $1 OR user_b = $1)
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
ORDER BY updated_at DESC;

-- name: CreateFriendEvent :one
INSERT INTO friend_event (friendship_id, actor_id, type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListEventsForFriendship :many
SELECT * FROM friend_event
WHERE friendship_id = $1
ORDER BY created_at ASC;

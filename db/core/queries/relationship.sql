-- name: CreateRelationship :one
INSERT INTO relationship (user_a, user_b, status, status_actor_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRelationshipByID :one
SELECT * FROM relationship
WHERE id = $1;

-- name: GetRelationshipByPair :one
SELECT * FROM relationship
WHERE user_a = $1 AND user_b = $2;

-- name: UpdateRelationshipStatus :one
UPDATE relationship
SET status = $2, status_actor_id = $3
WHERE id = $1
RETURNING *;

-- name: ListRelationshipsForUser :many
SELECT * FROM relationship
WHERE (user_a = $1 OR user_b = $1)
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
ORDER BY updated_at DESC;

-- name: CreateFriendEvent :one
INSERT INTO friend_event (relationship_id, actor_id, type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListEventsForRelationship :many
SELECT * FROM friend_event
WHERE relationship_id = $1
ORDER BY created_at ASC;

-- name: CreateUser :one
INSERT INTO "user" (auth_subject)
VALUES ($1)
RETURNING *;

-- name: GetUserByAuthSubject :one
SELECT * FROM "user"
WHERE auth_subject = $1;

-- name: GetUserByID :one
SELECT * FROM "user"
WHERE id = $1;

-- name: CreatePersona :one
INSERT INTO persona (id)
VALUES ($1)
RETURNING id;

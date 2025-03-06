-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES ( gen_random_uuid(), now(),now(),$1,$2)
RETURNING *;


-- name: DeleteUsers :exec
DELETE FROM users;


-- name: GetUsers :many
SELECT * FROM users;


-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

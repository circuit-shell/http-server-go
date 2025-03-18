-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES ( gen_random_uuid(), now(),now(),$1,$2)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;


-- name: UpdateUser :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

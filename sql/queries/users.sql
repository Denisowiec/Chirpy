-- name: CreateUser :one
INSERT INTO users (created_at, updated_at, email, hashed_password) VALUES (NOW(), NOW(), $1, $2) RETURNING *;

-- name: Reset :exec
TRUNCATE users CASCADE;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
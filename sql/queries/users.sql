-- name: CreateUser :one
INSERT INTO users (created_at, updated_at, email) VALUES (NOW(), NOW(), $1) RETURNING *;

-- name: Reset :exec
TRUNCATE users CASCADE;
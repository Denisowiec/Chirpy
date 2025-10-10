-- name: CreateChirp :one
INSERT INTO chirps (created_at, updated_at, body, user_id) VALUES (NOW(), NOW(), $1, $2) RETURNING *;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetChirpsForUser :many
SELECT * FROM chirps WHERE user_id = $1 ORDER BY created_at ASC;

-- name: GetChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: DeleteChirp :one
DELETE FROM chirps WHERE id = $1 AND user_id = $2 RETURNING *;
-- name: SetRefToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at,        revoked_at) VALUES 
                              ($1,      NOW(),      NOW(),      $2, $3, NULL      ) RETURNING *;

-- name: GetRefToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: GetUserFromRefToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1;

-- name: RevokeToken :one
UPDATE refresh_tokens SET revoked_at = Now(), updated_at = Now() WHERE token = $1 RETURNING *;
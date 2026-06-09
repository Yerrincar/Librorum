-- name: CreateSession :one
INSERT INTO sessions (user_id, hash) VALUES ($1,$2) RETURNING *;

-- name: FindSessionByTokenHash :one
SELECT * FROM sessions WHERE hash = $1;

-- name: DeleteSessionByTokenHash :one
DELETE FROM sessions WHERE hash = $1 RETURNING *;

-- name: DeleteExpiredSessions :one
DELETE FROM sessions WHERE expires_at < NOW() RETURNING *;

-- name: DeleteSessionByUserID :many
DELETE FROM sessions WHERE user_id = $1 RETURNING *;

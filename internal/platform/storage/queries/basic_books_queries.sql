-- name: SelectBooks :many
SELECT * FROM library_items ORDER BY created_at LIMIT $1 OFFSET $2; 

-- name: InsertUser :one
INSERT INTO users (username, email, password_hash, display_name) VALUES ($1,$2,$3,$4) RETURNING *;

-- name: SelectUserByUsername :one 
SELECT * FROM users WHERE username = $1;

-- name: SelectUserByID :one 
SELECT * FROM users WHERE id = $1;

-- name: SelectUserByEmail :one 
SELECT * FROM users WHERE email = $1;

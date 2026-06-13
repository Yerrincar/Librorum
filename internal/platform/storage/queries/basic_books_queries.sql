-- name: SelectBooks :many
SELECT * FROM library_items WHERE user_id = $1 AND kind = $2 ORDER BY created_at LIMIT $3 OFFSET $4; 

-- name: SelectUserByUsername :one 
SELECT * FROM users WHERE username = $1;

-- name: SelectUserByID :one 
SELECT * FROM users WHERE id = $1;

-- name: SelectUserByEmail :one 
SELECT * FROM users WHERE email = $1;

-- name: InsertUser :one
INSERT INTO users (username, email, password_hash, display_name) VALUES ($1,$2,$3,$4) RETURNING *;

-- name: InsertBook :one
INSERT INTO library_items (user_id, kind, title, author, description, language, publication_year, genres, rating, 
    ownership_status, reading_status, publication_status, current_chapter, total_chapters,
    read_at, cover_path, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
    RETURNING *;

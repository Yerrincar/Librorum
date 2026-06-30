-- name: SelectBooksByUser :many
SELECT * FROM library_items WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: SelectBooksByUserAndKind :many
SELECT * FROM library_items WHERE user_id = $1 AND kind = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4;

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

-- name: UpdateLibraryItems :one
UPDATE library_items SET title = $3, author = $4, rating = $5, cover_path = COALESCE(NULLIF($6, ''), cover_path), read_at = $7, description = $8,
language = $9, genres = $10, ownership_status = $11, reading_status=$12, current_chapter = $13, total_chapters = $14, notes = $15,  updated_at = now() WHERE id = $1 AND user_id = $2 RETURNING *;

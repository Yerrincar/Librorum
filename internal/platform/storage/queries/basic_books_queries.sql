-- name: SelectBooksByUser :many
SELECT * FROM library_items 
WHERE 
    user_id = $1 AND (sqlc.narg('reading_year')::int IS NULL OR EXTRACT(YEAR FROM read_at) = sqlc.narg('reading_year')::int)
ORDER BY 
CASE WHEN sqlc.arg('sort_by')::text = 'read_at_asc' THEN read_at END ASC,
CASE WHEN sqlc.arg('sort_by')::text = 'read_at_desc' THEN read_at END DESC,
CASE WHEN sqlc.arg('sort_by')::text = 'title_asc' THEN title END ASC,
CASE WHEN sqlc.arg('sort_by')::text = 'title_desc' THEN title END DESC,
CASE WHEN sqlc.arg('sort_title')::text = 'asc' THEN title END ASC,
CASE WHEN sqlc.arg('sort_title')::text = 'desc' THEN title END DESC,
id ASC LIMIT $2 OFFSET $3;

-- name: SelectBooksByUserAndKind :many
SELECT * FROM library_items 
WHERE 
    user_id = $1 AND kind = $2 AND (sqlc.narg('reading_year')::int IS NULL OR EXTRACT(YEAR FROM read_at) = sqlc.narg('reading_year')::int)
ORDER BY 
CASE WHEN sqlc.arg('sort_by')::text = 'read_at_asc' THEN read_at END ASC,
CASE WHEN sqlc.arg('sort_by')::text = 'read_at_desc' THEN read_at END DESC,
CASE WHEN sqlc.arg('sort_by')::text = 'title_asc' THEN title END ASC,
CASE WHEN sqlc.arg('sort_by')::text = 'title_desc' THEN title END DESC,
CASE WHEN sqlc.arg('sort_title')::text = 'asc' THEN title END ASC,
CASE WHEN sqlc.arg('sort_title')::text = 'desc' THEN title END DESC,
id ASC LIMIT $3 OFFSET $4;

-- name: SelectReadYearByUsername :many
SELECT DISTINCT EXTRACT(YEAR FROM read_at)::int AS reading_year
FROM library_items
WHERE user_id = $1
  AND read_at IS NOT NULL
ORDER BY reading_year DESC;

-- name: SelectReadYearByUsernameAndKind :many
SELECT DISTINCT EXTRACT(YEAR FROM read_at)::int AS reading_year
FROM library_items
WHERE user_id = $1 
  AND kind = $2
  AND read_at IS NOT NULL
ORDER BY reading_year DESC;

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
UPDATE library_items SET 
    title = sqlc.arg('title'), 
    author = sqlc.arg('author'), 
    rating = sqlc.arg('rating'), 
    cover_path = COALESCE(NULLIF(sqlc.arg('cover_path')::text, ''), cover_path), 
    read_at = sqlc.arg('read_at'), 
    description = sqlc.arg('description'),
    language = sqlc.arg('language'), 
    genres = sqlc.arg('genres'), 
    ownership_status = sqlc.arg('ownership_status'), 
    reading_status = sqlc.arg('reading_status'), 
    current_chapter = sqlc.arg('current_chapter'), 
    total_chapters = sqlc.arg('total_chapters'), 
    notes = sqlc.arg('notes'), 
    updated_at = now() 
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') 
RETURNING *;

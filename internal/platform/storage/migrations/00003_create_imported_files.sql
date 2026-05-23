-- +goose Up
CREATE TABLE imported_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    library_item_id BIGINT REFERENCES library_items(id) ON DELETE SET NULL,
    original_file_name TEXT NOT NULL,
    stored_file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    media_type TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    sha256 TEXT NOT NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, sha256),
    UNIQUE (user_id, file_path)
);

CREATE INDEX imported_files_user_id_idx ON imported_files(user_id);
CREATE INDEX imported_files_library_item_id_idx ON imported_files(library_item_id);

-- +goose Down
DROP TABLE IF EXISTS imported_files;

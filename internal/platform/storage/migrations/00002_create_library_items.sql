-- +goose Up
CREATE TABLE library_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('book', 'manga', 'manhwa')),
    title TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT '',
    publication_year INTEGER CHECK (publication_year IS NULL OR publication_year BETWEEN 0 AND 3000),
    genres TEXT[] NOT NULL DEFAULT '{}',
    rating NUMERIC(3,1) CHECK (rating IS NULL OR rating BETWEEN 0 AND 5),
    ownership_status TEXT NOT NULL DEFAULT 'none' CHECK (
        ownership_status IN ('none', 'owned_physical', 'owned_digital', 'owned_physical_and_digital', 'wishlist')
    ),
    reading_status TEXT NOT NULL DEFAULT 'unread' CHECK (
        reading_status IN ('unread', 'to_read', 'reading', 'read', 'dropped')
    ),
    publication_status TEXT NOT NULL DEFAULT 'unknown' CHECK (
        publication_status IN ('unknown', 'finished', 'ongoing', 'hiatus')
    ),
    current_chapter NUMERIC(8,2) CHECK (current_chapter IS NULL OR current_chapter >= 0),
    total_chapters NUMERIC(8,2) CHECK (total_chapters IS NULL OR total_chapters >= 0),
    read_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    cover_path TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    search_vector TSVECTOR NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (total_chapters IS NULL OR current_chapter IS NULL OR current_chapter <= total_chapters)
);

CREATE INDEX library_items_user_id_idx ON library_items(user_id);
CREATE INDEX library_items_user_kind_idx ON library_items(user_id, kind);
CREATE INDEX library_items_user_reading_status_idx ON library_items(user_id, reading_status);
CREATE INDEX library_items_user_ownership_status_idx ON library_items(user_id, ownership_status);
CREATE INDEX library_items_user_publication_status_idx ON library_items(user_id, publication_status);
CREATE INDEX library_items_user_read_at_idx ON library_items(user_id, read_at DESC NULLS LAST);
CREATE INDEX library_items_genres_idx ON library_items USING GIN (genres);
CREATE INDEX library_items_search_idx ON library_items USING GIN (search_vector);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_library_items_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector =
        setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(NEW.author, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(array_to_string(NEW.genres, ' '), '')), 'C') ||
        setweight(to_tsvector('simple', coalesce(NEW.description, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER library_items_set_search_vector
BEFORE INSERT OR UPDATE OF title, author, genres, description ON library_items
FOR EACH ROW
EXECUTE FUNCTION set_library_items_search_vector();

CREATE TRIGGER library_items_set_updated_at
BEFORE UPDATE ON library_items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS library_items_set_updated_at ON library_items;
DROP TRIGGER IF EXISTS library_items_set_search_vector ON library_items;
DROP TABLE IF EXISTS library_items;
DROP FUNCTION IF EXISTS set_library_items_search_vector();

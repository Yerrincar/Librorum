-- +goose Up
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hash TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '30 Days'
);

-- +goose Down
DROP TABLE IF EXISTS sessions;

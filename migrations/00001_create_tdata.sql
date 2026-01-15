-- +goose Up
CREATE TABLE IF NOT EXISTS tdata (
    id SERIAL PRIMARY KEY,
    shortURL TEXT NOT NULL,
    originalURL TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_short_url ON tdata (shortURL);

-- +goose Down
DROP TABLE IF EXISTS tdata;
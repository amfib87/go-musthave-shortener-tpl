-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE UNIQUE INDEX IF NOT EXISTS idx_tdata_originalurl_unique ON tdata (originalURL);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP INDEX IF EXISTS idx_tdata_originalurl_unique;
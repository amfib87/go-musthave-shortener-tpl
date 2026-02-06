-- +goose Up
ALTER TABLE tdata
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE tdata
DROP COLUMN IF EXISTS is_deleted;
-- +goose Up
-- Добавляем столбец с временным значением по умолчанию
ALTER TABLE tdata
ADD COLUMN IF NOT EXISTS userid VARCHAR(128) NOT NULL DEFAULT 'unknown';

-- Удаляем значение по умолчанию
ALTER TABLE tdata
ALTER COLUMN userid DROP DEFAULT;
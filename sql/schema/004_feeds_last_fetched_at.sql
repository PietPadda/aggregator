-- 004_feeds_last_fetched_at.sql

-- +goose Up
ALTER TABLE feeds
ADD COLUMN last_fetched_at TIMESTAMP;

-- +goose Down
ALTER TABLE feeds
DROP COLUMN last_fetched_at;
-- 005_posts.sql

-- +goose Up
CREATE TABLE posts (
    -- define table columns
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE, -- critical to make URL unique
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL,
    -- and link to feeds
    FOREIGN KEY (feed_id) 
        REFERENCES feeds(id) 
        ON DELETE CASCADE -- delete record if feed deleted
);

-- +goose Down
DROP TABLE posts;
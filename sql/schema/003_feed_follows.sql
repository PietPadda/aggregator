-- 003_feed_follows.sql

-- +goose Up
CREATE TABLE feed_follows (
    -- define table columns
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    feed_id UUID NOT NULL,
    -- link to users
    FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE, -- delete record if user deleted
    -- and link to feeds (many-to-one)
    FOREIGN KEY (feed_id) 
        REFERENCES feeds(id) 
        ON DELETE CASCADE, -- delete record if feed deleted
    -- unique pair of user/feed only
    UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
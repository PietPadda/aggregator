-- feeds.sql

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: ListFeedsWithCreator :many
SELECT f.name AS feedName, f.url AS feedURL, u.name AS userName
FROM feeds f INNER JOIN users u
ON u.id = f.user_id;

-- name: GetFeedByURL :one
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds
WHERE url = $1 -- url to match the inputy
LIMIT 1; -- ensure only one record is returned
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

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
  updated_at = NOW(),
  last_fetched_at = NOW()
WHERE id = $1; -- use feed_id (unique as it's a pk)

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds          -- we return ALL cols for ScrapeFeeds
ORDER BY last_fetched_at ASC -- from oldest to newest
NULLS FIRST                   -- any null fetched at record first, these are EVEN older!
LIMIT 1;                     -- we should only get 1, as there MIGHT be more than one
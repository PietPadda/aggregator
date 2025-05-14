-- feed_follows.sql

-- name: CreateFeedFollows :one
-- create new record using WITH CTE pattern
WITH inserted_feed_follow AS (
    -- insert into table
    INSERT INTO feed_follows (
        id,
        created_at,
        updated_at,
        user_id,
        feed_id
    )
    -- value placeholders
    VALUES (
        $1, $2, $3, $4, $5
    ) RETURNING * -- return after insert (populates the CTE record!)
)
-- select the record data and user and feed name from inserted_feed_follow
SELECT
    iff.id,
    iff.created_at,
    iff.updated_at,
    iff.user_id,
    iff.feed_id,
    u.name AS userName,
    f.name AS feedName
FROM inserted_feed_follow iff
-- inner join users (omit other users)
INNER JOIN users u ON u.id = iff.user_id
-- inner join feeds (omit other feeds)
INNER JOIN feeds f ON f.id = iff.feed_id;

-- name: GetFeedFollowsForUser :many
-- get all feed follows for a user
SELECT
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    u.name AS userName,
    f.name AS feedName
FROM feed_follows ff
-- inner join users (omit other users)
INNER JOIN users u ON u.id = ff.user_id
-- inner join feeds (omit other feeds)
INNER JOIN feeds f ON f.id = ff.feed_id
-- where clause to filter by user_id
WHERE ff.user_id = $1
-- order by created_at descending (otherwise random with where clause)
ORDER BY ff.created_at DESC;

-- name: DeleteFeedFollowByUserAndFeed :one
-- delete feed follow record by url for a user
DELETE FROM feed_follows ff
-- using feeds table (PostgreSQL doesn't support inner join on delete)
USING feeds f
-- where clause to filter record
WHERE f.url = $1         -- matches url
  AND ff.user_id = $2    -- matches user_id
  AND ff.feed_id = f.id  -- feed follow id matches feed id
RETURNING ff.*;          -- get the deleted record from feed follows table!

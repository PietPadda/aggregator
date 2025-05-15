-- posts.sql

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT 
    p.id,
    p.created_at,
    p.updated_at,
    p.title,
    p.url,
    p.description,
    p.published_at,
    p.feed_id
FROM posts p
-- inner join feed_follows (omit other feeds and users)
INNER JOIN feed_follows ff ON p.feed_id = ff.feed_id
-- match with current user
WHERE ff.user_id = $1
-- order by published_at descending, NULLS LAST (as they're older)
-- THEN order by updated_ desc, to prevent random NULL selection
ORDER BY p.published_at DESC NULLS LAST,
         p.created_at DESC
LIMIT $2;
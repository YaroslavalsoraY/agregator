-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, 
                   published_at, title, url, 
                   description, feed_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);

-- name: GetPostsForUser :many
SELECT posts.title, posts.description FROM posts
INNER JOIN feeds
ON feeds.id = posts.feed_id
WHERE feeds.user_id = $1
ORDER BY posts.updated_at DESC
LIMIT $2;
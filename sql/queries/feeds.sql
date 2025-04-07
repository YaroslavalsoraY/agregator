-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT f.name, f.url, u.name
FROM feeds f
INNER JOIN users u
ON f.user_id = u.id;

-- name: GetFeedByUrl :one
SELECT id, name FROM feeds
WHERE url = $1;

-- name: GetFeedByID :one
SELECT name FROM feeds
WHERE id = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds SET updated_at = NOW(), last_fetched_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT url, id FROM feeds
ORDER BY last_fetched_at NULLS FIRST;
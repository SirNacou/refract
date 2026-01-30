-- name: ListURLs :many
SELECT * FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetURLByShortCode :one
SELECT * FROM urls
WHERE short_code = $1;

-- name: CreateURL :one
INSERT INTO urls (id, short_code, original_url, user_id, domain, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
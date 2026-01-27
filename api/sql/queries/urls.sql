-- name: ListURLs :many
SELECT * FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;
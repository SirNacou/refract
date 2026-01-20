-- name: GetURLByShortCode :one
SELECT * FROM urls
WHERE short_code = $1
LIMIT 1;

-- name: CreateURL :one
INSERT InTO urls (
    snowflake_id,
    short_code,
    destination_url,
    title,
    notes,
    status,
    created_at,
    updated_at,
    expires_at,
    creator_user_id,
    total_clicks,
    last_clicked_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;
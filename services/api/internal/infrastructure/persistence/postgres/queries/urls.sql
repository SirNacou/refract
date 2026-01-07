-- name: CreateURL :exec
INSERT INTO urls (
    id,
    short_code,
    original_url,
    domain,
    expires_at,
    has_fixed_expiration,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetURLByShortCode :one
SELECT 
    id,
    short_code,
    original_url,
    domain,
    created_at,
    updated_at,
    expires_at,
    has_fixed_expiration,
    click_count,
    is_active,
    metadata
FROM urls
WHERE short_code = $1
LIMIT 1;

-- name: ExistsByShortCode :one
SELECT EXISTS(
    SELECT 1 FROM urls WHERE short_code = $1
);

-- name: IncrementClickCount :exec
UPDATE urls
SET click_count = click_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE short_code = $1;

-- name: UpdateExpiration :exec
UPDATE urls
SET expires_at = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE short_code = $2;

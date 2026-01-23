-- name: GetClickSummaryByURL :many
SELECT * FROM click_summary_hourly 
WHERE url_id = $1 AND hour >= $2 AND hour <= $3
ORDER BY hour DESC;
-- name: GetGeographicBreakdown :many
SELECT country_code, country_name, COUNT(*) as clicks
FROM click_events
WHERE url_id = $1 AND time >= $2 AND time <= $3
GROUP BY country_code, country_name
ORDER BY clicks DESC
LIMIT 50;
-- name: GetBrowserBreakdown :many
SELECT browser, COUNT(*) as clicks
FROM click_events
WHERE url_id = $1 AND time >= $2 AND time <= $3
GROUP BY browser
ORDER BY clicks DESC;
-- name: GetReferrerBreakdown :many
SELECT referrer, COUNT(*) as clicks
FROM click_events
WHERE url_id = $1 AND time >= $2 AND time <= $3
GROUP BY referrer
ORDER BY clicks DESC
LIMIT 50;
-- name: GetClickEventsByURL :many
SELECT * FROM click_events
WHERE url_id = $1 AND time >= $2 AND time <= $3
ORDER BY time DESC
LIMIT 100;
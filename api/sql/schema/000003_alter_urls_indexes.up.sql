
DROP INDEX idx_urls_short_code;

CREATE UNIQUE INDEX idx_urls_active_short_code ON urls (short_code)
WHERE status = 'active';
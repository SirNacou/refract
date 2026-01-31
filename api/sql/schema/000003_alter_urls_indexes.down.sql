CREATE UNIQUE INDEX idx_urls_short_code ON urls (short_code);

DROP INDEX idx_urls_active_short_code;
-- Rollback: 00002_create_urls
-- Warning: This will delete all URL data

DROP TRIGGER IF EXISTS update_urls_updated_at ON urls;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS urls;
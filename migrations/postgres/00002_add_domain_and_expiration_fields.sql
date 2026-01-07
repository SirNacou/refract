-- +goose Up
-- Add domain and expiration behavior fields to URLs table

-- Add domain column (required field, validated against whitelist)
ALTER TABLE urls 
ADD COLUMN domain VARCHAR(255) NOT NULL DEFAULT 'short.link';

-- Add has_fixed_expiration column (distinguishes two expiration behaviors)
-- true = fixed expiration (never renews)
-- false = activity-based (renews on every click to +12 months)
ALTER TABLE urls 
ADD COLUMN has_fixed_expiration BOOLEAN NOT NULL DEFAULT false;

-- Remove default from domain after adding the column
ALTER TABLE urls 
ALTER COLUMN domain DROP DEFAULT;

-- Index for querying by domain
CREATE INDEX idx_urls_domain ON urls(domain);

-- +goose Down
DROP INDEX IF EXISTS idx_urls_domain;
ALTER TABLE urls DROP COLUMN IF EXISTS has_fixed_expiration;
ALTER TABLE urls DROP COLUMN IF EXISTS domain;

-- +goose Up
-- Create URLs table for URL shortener
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + INTERVAL '12 months') NOT NULL,
    
    -- Metrics
    click_count BIGINT DEFAULT 0 NOT NULL,
    
    -- Status
    is_active BOOLEAN DEFAULT true NOT NULL,
    
    -- Extensibility - for custom metadata like UTM params, QR settings, etc.
    metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    
    -- Constraints
    CONSTRAINT short_code_min_length CHECK (char_length(short_code) >= 4),
    CONSTRAINT short_code_max_length CHECK (char_length(short_code) <= 20),
    CONSTRAINT original_url_not_empty CHECK (char_length(original_url) > 0)
);

-- Performance indexes
-- Partial index for active URLs only (most queries)
CREATE INDEX idx_urls_short_code_active ON urls(short_code) WHERE is_active = true;

-- Index for cleanup jobs (find expired URLs)
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE is_active = true;

-- Index for listing recent URLs
CREATE INDEX idx_urls_created_at ON urls(created_at DESC);

-- GIN index for JSONB queries on metadata
CREATE INDEX idx_urls_metadata ON urls USING GIN(metadata);

-- Function to auto-update updated_at timestamp
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

-- Trigger to auto-update updated_at on every UPDATE
CREATE TRIGGER update_urls_updated_at 
    BEFORE UPDATE ON urls
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_urls_updated_at ON urls;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS urls;

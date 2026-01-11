-- Migration: 00002_create_urls
-- Description: Create urls table for short URL storage with indexes and triggers
-- Created: 2026-01-11

CREATE TABLE urls (
    snowflake_id BIGINT PRIMARY KEY,
    custom_alias TEXT,
    destination_url TEXT NOT NULL,
    title TEXT,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'expired', 'disabled', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    creator_user_id TEXT NOT NULL
        REFERENCES users(zitadel_sub) ON DELETE CASCADE,
    total_clicks BIGINT NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_urls_creator ON urls(creator_user_id, created_at DESC);
CREATE INDEX idx_urls_status ON urls(status) WHERE status = 'active';
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE UNIQUE INDEX idx_urls_custom_alias ON urls(custom_alias) WHERE custom_alias IS NOT NULL;

-- Trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger
CREATE TRIGGER update_urls_updated_at
    BEFORE UPDATE ON urls
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
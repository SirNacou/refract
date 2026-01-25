-- Migration: 00003_create_api_keys
-- Description: Create api_keys table for programmatic API access
-- Created: 2026-01-11

CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL
        REFERENCES users(zitadel_sub) ON DELETE CASCADE,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'revoked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    usage_count BIGINT NOT NULL DEFAULT 0
);

-- Indexes
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user ON api_keys(user_id, created_at DESC) WHERE status = 'active';
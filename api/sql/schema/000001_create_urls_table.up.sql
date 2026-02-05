CREATE TABLE urls (
    -- Snowflake IDs fit into a 64-bit integer (BIGINT in Postgres).
    -- We do NOT use SERIAL/AUTO_INCREMENT because the Go app generates this ID.
    id BIGINT PRIMARY KEY,

    -- The unique slug (e.g., 'u8K2a').
    -- VARCHAR(20) provides breathing room, though 7 chars is standard.
    -- Collation "C" improves lookup performance for strict ASCII slugs.
    short_code VARCHAR(20) COLLATE "C" NOT NULL,

    -- Use TEXT for URLs. In Postgres, there is no performance penalty 
    original_url TEXT NOT NULL,

    user_id VARCHAR(255) NOT NULL, 

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ, -- Nullable: NULL means "never expires"
    
    -- Feature flags
    status TEXT CHECK (status IN ('active', 'disabled', 'expired')) NOT NULL DEFAULT 'active'
);

-- 1. CRITICAL: The Lookup Index
-- This makes redirect lookups O(1) or O(log N). 
-- "UNIQUE" enforces no duplicate slugs.
CREATE UNIQUE INDEX idx_urls_short_code ON urls (short_code);

-- 2. The Dashboard Index
-- Speeds up "Show me all my links" queries.
CREATE INDEX idx_urls_user_id_created_at ON urls (user_id, created_at DESC);
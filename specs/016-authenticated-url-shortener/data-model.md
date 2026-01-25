# Data Model: Authenticated URL Shortener Platform

**Feature**: 016-authenticated-url-shortener  
**Date**: 2026-01-09  
**Status**: Phase 1 - Design  
**Related**: [spec.md](./spec.md) | [plan.md](./plan.md) | [research.md](./research.md)

---

## Overview

This document defines the complete data model for the distributed URL shortener platform, including entity definitions, database schemas, relationships, indexes, and constraints. The system uses PostgreSQL 16 for relational data (users, URLs, API keys) and TimescaleDB for time-series analytics data (click events, aggregations).

**Design Principles**:
- Domain-Driven Design: Entities represent domain concepts (URL, User, APIKey)
- CQRS-optimized: Separate tables for write models (users, urls) and read models (analytics summaries)
- Time-series optimized: Hypertables for click events with automatic partitioning
- Privacy-first: IP anonymization, no personally identifiable information in analytics

---

## 1. Entity Definitions

### 1.1 User

Represents an authenticated account holder. Authentication delegated to Zitadel identity provider (no password storage in system).

**Domain Attributes**:
- **Zitadel Subject ID** (string, primary key): Unique identifier from Zitadel JWT `sub` claim
- **Email** (string): User email from Zitadel (for display, communication)
- **Created At** (timestamp): Account creation time in this system
- **Last Sync At** (timestamp): Last time Zitadel metadata was synced
- **Status** (enum): Account status (active, suspended, deleted)

**Relationships**:
- Owns multiple **ShortURL** entities (one-to-many)
- Owns multiple **APIKey** entities (one-to-many)

**Business Rules**:
- User accounts cannot be created directly (only via Zitadel authentication flow)
- Users can only access/modify their own URLs and API keys (authorization boundary)
- Deleting user account cascades to deactivate all URLs, anonymize analytics, delete API keys (FR-118)

---

### 1.2 ShortURL

Represents a shortened URL with metadata and lifecycle management.

**Domain Attributes**:
- **Snowflake ID** (int64, primary key): Distributed unique identifier (64-bit)
- **Short Code** (string, derived): Base62-encoded Snowflake ID (8-11 characters, URL-safe)
- **Custom Alias** (string, optional): User-provided vanity alias (3-50 characters)
- **Destination URL** (string): Original long URL to redirect to
- **Title** (string, optional): User-provided descriptive title
- **Notes** (text, optional): User notes for organization
- **Status** (enum): active, expired, disabled, deleted
- **Created At** (timestamp): URL creation time
- **Updated At** (timestamp): Last modification time
- **Expires At** (timestamp, optional): Automatic expiration date (FR-013)
- **Creator User ID** (string, foreign key): Zitadel subject ID of owner
- **Total Clicks** (int64): Cached click count (denormalized for performance)
- **Last Clicked At** (timestamp, optional): Most recent click time

**Relationships**:
- Belongs to one **User** (many-to-one)
- Has many **ClickEvent** records (one-to-many, time-series)
- Has many **AnalyticsSummary** records (one-to-many, aggregated metrics)

**Business Rules**:
- Snowflake ID generated once at creation, immutable (FR-008)
- Short code derived from Snowflake ID (not stored separately, computed on read)
- Custom alias must be globally unique (cannot collide with auto-generated codes or other aliases)
- Custom alias validates: 3-50 chars, alphanumeric + hyphens, no reserved words (admin, api, health) (FR-010)
- Destination URL validated against Safe Browsing API before creation (FR-012, FR-042)
- Status transitions: active → disabled, active → expired (if expires_at passed), any → deleted
- Expired/disabled/deleted URLs return error pages instead of redirecting (FR-018, FR-019)

---

### 1.3 ClickEvent

Represents a single redirect event captured for analytics. Stored in TimescaleDB hypertable (time-series optimized).

**Domain Attributes**:
- **Time** (timestamptz, part of composite key): Event timestamp (partitioning key)
- **Event ID** (UUID): Unique event identifier
- **URL ID** (int64, foreign key): Snowflake ID of short URL
- **Referrer** (string, optional): HTTP Referer header (source of click)
- **User Agent** (string): Full user agent string
- **IP Address** (inet): Anonymized IP (last octet zeroed for IPv4, last 80 bits zeroed for IPv6) (FR-020)
- **Country Code** (char(2)): ISO 3166-1 alpha-2 country code
- **Country Name** (string): Full country name (e.g., "United States")
- **City** (string, optional): City name from GeoIP lookup
- **Latitude** (float, optional): Geographic latitude
- **Longitude** (float, optional): Geographic longitude
- **Device Type** (enum): desktop, mobile, tablet, bot (FR-022)
- **Browser** (string): Browser name and version (e.g., "Chrome 120.0")
- **Operating System** (string): OS name (e.g., "Windows 10", "iOS 17.2")

**Relationships**:
- Belongs to one **ShortURL** (many-to-one)

**Business Rules**:
- Time must be present and used as partitioning key (TimescaleDB requirement)
- IP address anonymized before storage (privacy requirement)
- Geographic data derived from anonymized IP (MaxMind GeoLite2 lookup)
- Device/browser/OS parsed from user agent (user-agent parsing library)
- Events immutable after insertion (append-only, no updates)
- Retention policy: Auto-delete events older than 5 years (FR-051)
- Compression policy: Compress chunks older than 7 days (90% storage reduction)

---

### 1.4 APIKey

Represents programmatic access credentials for API-based URL creation.

**Domain Attributes**:
- **ID** (int64, primary key): Auto-increment key ID
- **Key Hash** (string): BLAKE2b-256 hash of full API key (64 hex characters)
- **Key Prefix** (string): First 16 characters of API key (e.g., "refract_abc12345") for identification
- **Name** (string): User-provided descriptive name (e.g., "Blog automation")
- **Status** (enum): active, revoked
- **Created At** (timestamp): Key generation time
- **Last Used At** (timestamp, optional): Most recent API request timestamp
- **Usage Count** (int64): Total number of API requests made with this key
- **Owner User ID** (string, foreign key): Zitadel subject ID of creator

**Relationships**:
- Belongs to one **User** (many-to-one)

**Business Rules**:
- Full API key (e.g., "refract_abc12345xyz...") shown only once at generation (never retrievable)
- Key hash computed with BLAKE2b-256 before storage (FR-034)
- Validation: Hash provided key, lookup by hash in database (<1ms validation)
- Key prefix stored for display in UI/logs without exposing full key
- Revoked keys reject all API requests with 401 Unauthorized
- Rate limiting: 1000 requests/hour per API key (FR-041)
- Users can create unlimited API keys (no hard limit, subject to rate limits)

---

### 1.5 AnalyticsSummary

Represents pre-aggregated metrics for fast dashboard queries. Computed via TimescaleDB continuous aggregates.

**Domain Attributes**:
- **Time Bucket** (timestamptz): Start of aggregation period (hour/day/week)
- **URL ID** (int64, foreign key): Snowflake ID of short URL
- **Total Clicks** (int64): Count of all click events in period
- **Unique Visitors** (int64): Count of distinct IP addresses in 24-hour window (FR-027)
- **Mobile Clicks** (int64): Count of clicks from mobile devices
- **Desktop Clicks** (int64): Count of clicks from desktop devices
- **Top Referrers** (JSONB): Array of {referrer: string, count: int} objects (top 10)
- **Top Countries** (JSONB): Array of {country_code: string, count: int} objects (top 10)
- **Top Cities** (JSONB): Array of {city: string, count: int} objects (top 10)
- **Top Browsers** (JSONB): Array of {browser: string, count: int} objects (top 5)

**Relationships**:
- Belongs to one **ShortURL** (many-to-one)

**Business Rules**:
- Continuous aggregates refresh every 5 minutes (FR-024: near real-time)
- Hourly buckets for last 30 days (detailed view)
- Daily buckets for 30 days - 1 year (summary view)
- Monthly buckets for 1+ years (long-term trends)
- JSONB columns store top-N arrays for dashboard queries without joins
- Read-only table (materialized view), do not insert directly

---

## 2. Database Schemas

### 2.1 PostgreSQL Schema (Relational Data)

#### Table: `users`

```sql
CREATE TABLE users (
    zitadel_sub TEXT PRIMARY KEY,                      -- Zitadel subject ID (e.g., "123456789012345678")
    email TEXT NOT NULL,                               -- Email from Zitadel token
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),     -- First login time
    last_sync_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),   -- Last Zitadel metadata sync
    status TEXT NOT NULL DEFAULT 'active'              -- active, suspended, deleted
        CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

**Storage Estimates**:
- 10,000 users × ~200 bytes/row = ~2 MB
- Indexes: ~1 MB
- **Total: ~3 MB**

---

#### Table: `urls`

```sql
CREATE TABLE urls (
    snowflake_id BIGINT PRIMARY KEY,                   -- 64-bit Snowflake ID
    custom_alias TEXT UNIQUE,                          -- User-provided vanity alias (if set)
    destination_url TEXT NOT NULL,                     -- Original long URL
    title TEXT,                                        -- User-provided title
    notes TEXT,                                        -- User notes
    status TEXT NOT NULL DEFAULT 'active'              -- active, expired, disabled, deleted
        CHECK (status IN ('active', 'expired', 'disabled', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),     
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                            -- Optional expiration
    creator_user_id TEXT NOT NULL                      -- Foreign key to users.zitadel_sub
        REFERENCES users(zitadel_sub) ON DELETE CASCADE,
    total_clicks BIGINT NOT NULL DEFAULT 0,            -- Cached count (denormalized)
    last_clicked_at TIMESTAMPTZ                        -- Most recent click
);

-- Indexes for performance
CREATE INDEX idx_urls_creator ON urls(creator_user_id, created_at DESC);
CREATE INDEX idx_urls_status ON urls(status) WHERE status = 'active';
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE UNIQUE INDEX idx_urls_custom_alias ON urls(custom_alias) WHERE custom_alias IS NOT NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_urls_updated_at
    BEFORE UPDATE ON urls
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Storage Estimates**:
- 500,000 URLs × ~500 bytes/row (with destination URL, metadata) = ~250 MB
- Indexes: ~100 MB
- **Total: ~350 MB**

**Index Rationale**:
- `idx_urls_creator`: User dashboard queries (list my URLs sorted by date)
- `idx_urls_status`: Filter active URLs for redirector lookups
- `idx_urls_expires_at`: Background job to mark expired URLs
- `idx_urls_custom_alias`: Enforce uniqueness, fast custom alias lookups

---

#### Table: `api_keys`

```sql
CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,                          -- Auto-increment key ID
    user_id TEXT NOT NULL                              -- Foreign key to users.zitadel_sub
        REFERENCES users(zitadel_sub) ON DELETE CASCADE,
    key_hash TEXT NOT NULL UNIQUE,                     -- BLAKE2b-256 hash (64 hex chars)
    key_prefix TEXT NOT NULL,                          -- First 16 chars for display
    name TEXT NOT NULL,                                -- User-provided description
    status TEXT NOT NULL DEFAULT 'active'              -- active, revoked
        CHECK (status IN ('active', 'revoked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,                          -- Updated on each API request
    usage_count BIGINT NOT NULL DEFAULT 0              -- Total requests made
);

CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user ON api_keys(user_id, created_at DESC) WHERE status = 'active';
```

**Storage Estimates**:
- 1,000 API keys (10% of users) × ~200 bytes/row = ~200 KB
- Indexes: ~100 KB
- **Total: ~300 KB**

**Index Rationale**:
- `idx_api_keys_hash`: Fast validation lookups (O(1) hash lookup)
- `idx_api_keys_user`: User settings page (list my active API keys)

---

### 2.2 TimescaleDB Schema (Time-Series Analytics)

#### Hypertable: `click_events`

```sql
CREATE TABLE click_events (
    time TIMESTAMPTZ NOT NULL,                         -- Event timestamp (partitioning key)
    event_id UUID NOT NULL DEFAULT gen_random_uuid(),  -- Unique event identifier
    url_id BIGINT NOT NULL,                            -- Foreign key to urls.snowflake_id
    referrer TEXT,                                     -- HTTP Referer header
    user_agent TEXT NOT NULL,                          -- Full user agent string
    ip_address INET NOT NULL,                          -- Anonymized IP
    country_code CHAR(2),                              -- ISO 3166-1 alpha-2
    country_name TEXT,                                 -- Full country name
    city TEXT,                                         -- City from GeoIP
    latitude DOUBLE PRECISION,                         -- Geographic coordinates
    longitude DOUBLE PRECISION,
    device_type TEXT NOT NULL                          -- desktop, mobile, tablet, bot
        CHECK (device_type IN ('desktop', 'mobile', 'tablet', 'bot')),
    browser TEXT,                                      -- Browser name and version
    operating_system TEXT                              -- OS name and version
);

-- Convert to hypertable (partitioned by time)
SELECT create_hypertable('click_events', 'time', chunk_time_interval => INTERVAL '1 day');

-- Compression policy: Compress chunks older than 7 days
ALTER TABLE click_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'url_id',         -- Segment by URL for query efficiency
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('click_events', INTERVAL '7 days');

-- Retention policy: Drop chunks older than 5 years
SELECT add_retention_policy('click_events', INTERVAL '5 years');

-- Indexes for common queries
CREATE INDEX idx_click_events_url_time ON click_events (url_id, time DESC);
CREATE INDEX idx_click_events_country ON click_events (country_code, time DESC);
```

**Storage Estimates** (uncompressed):
- 10M events/month × 12 months × 5 years = 600M events
- 600M events × ~300 bytes/row = ~180 GB uncompressed
- **With 90% compression: ~18 GB compressed**

**Performance Notes**:
- Chunk size: 1 day (optimal for retention policy and query patterns)
- Compression after 7 days reduces hot data size
- Queries with `WHERE url_id = ? AND time >= ?` use chunk exclusion (10x faster)

---

#### Continuous Aggregate: `click_summary_hourly`

```sql
CREATE MATERIALIZED VIEW click_summary_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS hour,
    url_id,
    COUNT(*) AS total_clicks,
    COUNT(DISTINCT ip_address) AS unique_visitors,
    COUNT(*) FILTER (WHERE device_type = 'mobile') AS mobile_clicks,
    COUNT(*) FILTER (WHERE device_type = 'desktop') AS desktop_clicks,
    COUNT(*) FILTER (WHERE device_type = 'tablet') AS tablet_clicks,
    COUNT(*) FILTER (WHERE device_type = 'bot') AS bot_clicks,
    jsonb_agg(
        jsonb_build_object('referrer', referrer, 'count', ref_count) 
        ORDER BY ref_count DESC
    ) FILTER (WHERE referrer IS NOT NULL) AS top_referrers
FROM (
    SELECT 
        time_bucket('1 hour', time) AS time,
        url_id,
        ip_address,
        device_type,
        referrer,
        COUNT(*) AS ref_count
    FROM click_events
    GROUP BY 1, 2, 3, 4, 5
) subq
GROUP BY hour, url_id
WITH NO DATA;

-- Refresh policy: Update every 5 minutes
SELECT add_continuous_aggregate_policy('click_summary_hourly',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '5 minutes');
```

**Storage Estimates**:
- Hourly buckets: 500K URLs × 24 hours × 30 days = ~360M rows
- With continuous aggregate compression: ~5 GB

---

#### Continuous Aggregate: `click_summary_daily`

```sql
CREATE MATERIALIZED VIEW click_summary_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', hour) AS day,
    url_id,
    SUM(total_clicks) AS total_clicks,
    SUM(unique_visitors) AS unique_visitors,
    SUM(mobile_clicks) AS mobile_clicks,
    SUM(desktop_clicks) AS desktop_clicks
FROM click_summary_hourly
GROUP BY day, url_id
WITH NO DATA;

-- Refresh policy: Update daily
SELECT add_continuous_aggregate_policy('click_summary_daily',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');
```

**Storage Estimates**:
- Daily buckets: 500K URLs × 365 days × 5 years = ~912M rows
- With continuous aggregate: ~10 GB

---

## 3. Relationships Diagram

```
┌─────────────┐
│   users     │
│─────────────│
│ zitadel_sub │◄────┐
│ email       │     │
│ created_at  │     │ 1:N (owns)
└─────────────┘     │
                    │
        ┌───────────┼───────────┐
        │                       │
┌───────▼─────────┐   ┌────────▼────────┐
│     urls        │   │   api_keys      │
│─────────────────│   │─────────────────│
│ snowflake_id    │   │ id              │
│ custom_alias    │   │ user_id (FK)    │
│ destination_url │   │ key_hash        │
│ creator_user_id │   │ key_prefix      │
│ status          │   │ status          │
└────────┬────────┘   └─────────────────┘
         │
         │ 1:N (has clicks)
         │
┌────────▼─────────────────────┐
│      click_events            │
│──────────────────────────────│
│ time                         │
│ url_id (FK)                  │
│ ip_address (anonymized)      │
│ country_code                 │
│ device_type                  │
└──────────────────────────────┘
         │
         │ aggregates to
         │
┌────────▼──────────────────────┐
│  click_summary_hourly (MV)    │
│───────────────────────────────│
│ hour (time_bucket)            │
│ url_id                        │
│ total_clicks                  │
│ unique_visitors               │
└───────────────────────────────┘
         │
         │ rolls up to
         │
┌────────▼──────────────────────┐
│  click_summary_daily (MV)     │
│───────────────────────────────│
│ day (time_bucket)             │
│ url_id                        │
│ total_clicks                  │
└───────────────────────────────┘
```

**Key Relationships**:
1. **User → URLs**: One user owns many URLs (CASCADE delete)
2. **User → API Keys**: One user owns many API keys (CASCADE delete)
3. **URL → Click Events**: One URL has many click events (time-series)
4. **Click Events → Hourly Summary**: Continuous aggregate (auto-refresh)
5. **Hourly Summary → Daily Summary**: Continuous aggregate (auto-refresh)

---

## 4. Constraints

### 4.1 Primary Key Constraints

- **users.zitadel_sub**: TEXT primary key (Zitadel subject ID format: numeric string)
- **urls.snowflake_id**: BIGINT primary key (64-bit Snowflake ID, globally unique)
- **api_keys.id**: BIGSERIAL primary key (auto-increment)
- **click_events**: Composite key (time, event_id) for TimescaleDB hypertable

### 4.2 Foreign Key Constraints

- **urls.creator_user_id** → **users.zitadel_sub** (ON DELETE CASCADE)
- **api_keys.user_id** → **users.zitadel_sub** (ON DELETE CASCADE)
- **click_events.url_id** → **urls.snowflake_id** (ON DELETE SET NULL - retain historical data)

### 4.3 Unique Constraints

- **users.zitadel_sub**: Unique (enforced by primary key)
- **urls.snowflake_id**: Unique (enforced by primary key)
- **urls.custom_alias**: Unique (partial index WHERE custom_alias IS NOT NULL)
- **api_keys.key_hash**: Unique (prevent hash collisions)

### 4.4 Check Constraints

- **users.status**: IN ('active', 'suspended', 'deleted')
- **urls.status**: IN ('active', 'expired', 'disabled', 'deleted')
- **urls.custom_alias**: LENGTH 3-50, regex `^[a-zA-Z0-9-]+$` (alphanumeric + hyphens)
- **api_keys.status**: IN ('active', 'revoked')
- **click_events.device_type**: IN ('desktop', 'mobile', 'tablet', 'bot')

### 4.5 Not Null Constraints

Critical fields that cannot be NULL:
- **users**: zitadel_sub, email, created_at, last_sync_at, status
- **urls**: snowflake_id, destination_url, status, created_at, updated_at, creator_user_id
- **api_keys**: id, user_id, key_hash, key_prefix, name, status, created_at
- **click_events**: time, url_id, user_agent, ip_address, device_type

---

## 5. Indexes

### 5.1 Performance Indexes

| Table | Index Name | Columns | Purpose |
|-------|-----------|---------|---------|
| users | idx_users_email | email | Search users by email |
| users | idx_users_created_at | created_at DESC | List recent users |
| urls | idx_urls_creator | creator_user_id, created_at DESC | User dashboard (list my URLs) |
| urls | idx_urls_status | status (WHERE status='active') | Redirector queries (active only) |
| urls | idx_urls_expires_at | expires_at (WHERE NOT NULL) | Expiration job |
| urls | idx_urls_custom_alias | custom_alias (UNIQUE, WHERE NOT NULL) | Custom alias lookups |
| api_keys | idx_api_keys_hash | key_hash (UNIQUE) | API key validation |
| api_keys | idx_api_keys_user | user_id, created_at DESC (WHERE status='active') | User settings page |
| click_events | idx_click_events_url_time | url_id, time DESC | Analytics queries by URL |
| click_events | idx_click_events_country | country_code, time DESC | Geographic analytics |

### 5.2 Index Strategy

**Partial Indexes**:
- `idx_urls_status` only indexes active URLs (reduces index size by 80% as most URLs are deleted/disabled)
- `idx_urls_custom_alias` only indexes rows with custom aliases (50% of URLs)
- `idx_api_keys_user` only indexes active keys (revoked keys excluded)

**Composite Indexes**:
- `(creator_user_id, created_at DESC)`: Supports `ORDER BY created_at DESC` without separate sort
- `(url_id, time DESC)`: TimescaleDB chunk exclusion for time-range queries

**Index Maintenance**:
- Rebuild indexes monthly: `REINDEX TABLE urls;`
- Monitor bloat: `SELECT * FROM pgstattuple('urls');`
- Autovacuum tuned for high-write tables (click_events)

---

## 6. Data Integrity Rules

### 6.1 Application-Level Validation

**Before INSERT/UPDATE**:
1. **URLs**:
   - Destination URL: Valid HTTP/HTTPS, max 2048 chars, not blocked by Safe Browsing API
   - Custom alias: Regex `^[a-zA-Z0-9-]{3,50}$`, not reserved word (admin, api, health, dashboard)
   - Snowflake ID: Generated once, never modified
   - Expires at: If set, must be future date

2. **API Keys**:
   - Key hash: Computed with BLAKE2b-256, exactly 64 hex characters
   - Key prefix: First 16 characters of full key (includes "refract_" prefix)
   - Name: 1-100 characters, non-empty

3. **Click Events**:
   - IP address: Anonymized before storage (last octet zeroed for IPv4)
   - Time: Must not be in future (clock skew tolerance: 1 minute)
   - Device type: Parsed from user agent, defaults to "desktop" if unknown

### 6.2 Database-Level Enforcement

**Triggers**:
- `update_urls_updated_at`: Auto-update `updated_at` column on every UPDATE
- (Future) `urls_status_cascade`: When URL deleted, publish event to analytics processor

**Functions**:
- `update_updated_at_column()`: Reusable trigger function for timestamp maintenance

**Domain Validation**:
- Check constraints enforce enum values (status, device_type)
- Foreign key constraints prevent orphaned records

---

## 7. Data Lifecycle

### 7.1 URL Lifecycle States

```
   ┌──────────┐
   │  CREATE  │
   └────┬─────┘
        │
        ▼
   ┌──────────┐      expires_at reached      ┌──────────┐
   │  ACTIVE  │───────────────────────────►  │ EXPIRED  │
   └────┬─────┘                              └────┬─────┘
        │                                          │
        │ user disables                            │
        ▼                                          │
   ┌──────────┐                                   │
   │ DISABLED │                                   │
   └────┬─────┘                                   │
        │                                          │
        │ user deletes                             │
        └──────────────────────────────────────────┤
                                                   │
                                                   ▼
                                              ┌──────────┐
                                              │ DELETED  │
                                              └──────────┘
```

**State Transitions**:
- `active → expired`: Automatic (background job checks expires_at)
- `active → disabled`: User action (deactivate URL)
- `disabled → active`: User action (reactivate URL)
- `any → deleted`: User action (soft delete, URLs retained for analytics)

### 7.2 Data Retention

**Click Events**:
- **Hot data** (0-7 days): Uncompressed, fast queries
- **Warm data** (7 days - 5 years): Compressed (90% reduction), slower queries acceptable
- **Cold data** (5+ years): Auto-deleted by retention policy

**Continuous Aggregates**:
- **Hourly summaries**: 30 days retention
- **Daily summaries**: 1 year retention
- **Monthly summaries**: 5 years retention

**User Account Deletion**:
1. Set `users.status = 'deleted'`
2. Anonymize user email: `email = 'deleted_' || zitadel_sub`
3. Cascade: Set all user URLs to `status = 'deleted'`
4. Cascade: Delete all API keys
5. Retain click_events (anonymized IP already)

---

## 8. Example Queries

### 8.1 Redirect Query (Hot Path)

```sql
-- Redirector service: Lookup destination URL by Snowflake ID
SELECT destination_url, status
FROM urls
WHERE snowflake_id = $1 AND status = 'active';
```

**Performance**: Index-only scan on primary key, <1ms

---

### 8.2 User Dashboard Query

```sql
-- API service: List user's URLs with click counts
SELECT 
    snowflake_id,
    custom_alias,
    destination_url,
    title,
    status,
    created_at,
    total_clicks,
    last_clicked_at
FROM urls
WHERE creator_user_id = $1
ORDER BY created_at DESC
LIMIT 50 OFFSET $2;
```

**Performance**: Uses idx_urls_creator, ~10ms for 50 rows

---

### 8.3 Analytics Dashboard Query

```sql
-- API service: Hourly click trend (last 24 hours)
SELECT 
    hour,
    total_clicks,
    unique_visitors,
    mobile_clicks,
    desktop_clicks
FROM click_summary_hourly
WHERE url_id = $1 
    AND hour >= NOW() - INTERVAL '24 hours'
ORDER BY hour ASC;
```

**Performance**: Continuous aggregate, <5ms for 24 rows

---

### 8.4 Top Referrers Query

```sql
-- API service: Top referrers for URL (last 7 days)
SELECT 
    referrer,
    COUNT(*) as clicks
FROM click_events
WHERE url_id = $1
    AND time >= NOW() - INTERVAL '7 days'
    AND referrer IS NOT NULL
GROUP BY referrer
ORDER BY clicks DESC
LIMIT 10;
```

**Performance**: Uses idx_click_events_url_time + chunk exclusion, ~50ms

---

### 8.5 Geographic Distribution Query

```sql
-- API service: Click distribution by country (last 30 days)
SELECT 
    country_code,
    country_name,
    COUNT(*) as clicks
FROM click_events
WHERE url_id = $1
    AND time >= NOW() - INTERVAL '30 days'
GROUP BY country_code, country_name
ORDER BY clicks DESC;
```

**Performance**: Hypertable with chunk exclusion, ~100ms

---

## 9. Migration Strategy

### 9.1 Schema Evolution

**Version Control**:
- Migrations stored in `migrations/postgres/`
- Numbered sequentially: `00001_create_users.sql`, `00002_create_urls.sql`
- Use `migrate` tool or `pgmigrate` for execution

**Backwards Compatibility**:
- Additive changes only (add columns, add indexes)
- Never drop columns (mark as deprecated, remove after 2 versions)
- Use `ALTER TABLE ADD COLUMN ... DEFAULT NULL` (no table rewrite)

**Zero-Downtime Migrations**:
1. Add new column with NULL/default value
2. Deploy application code that writes to both old and new columns
3. Backfill existing rows (in batches, off-peak hours)
4. Deploy code that reads from new column only
5. Drop old column (after grace period)

### 9.2 Data Seeding

**Development Environment**:
```sql
-- Seed test user
INSERT INTO users (zitadel_sub, email) 
VALUES ('test_user_123', 'test@example.com');

-- Seed sample URLs
INSERT INTO urls (snowflake_id, destination_url, creator_user_id, title)
VALUES 
    (1234567890123456, 'https://example.com/article1', 'test_user_123', 'Test Article 1'),
    (1234567890123457, 'https://example.com/article2', 'test_user_123', 'Test Article 2');

-- Seed sample click events
INSERT INTO click_events (time, url_id, ip_address, device_type, user_agent)
SELECT 
    NOW() - (random() * INTERVAL '30 days'),
    1234567890123456,
    '192.168.1.0',
    'desktop',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0'
FROM generate_series(1, 1000);
```

---

## 10. Security Considerations

### 10.1 SQL Injection Prevention

- **Use parameterized queries exclusively** (SQLc, SQLx generate safe queries)
- Never concatenate user input into SQL strings
- Validate inputs before queries (custom alias regex, URL format)

### 10.2 Data Encryption

- **At Rest**: PostgreSQL transparent data encryption (TDE) or disk-level encryption
- **In Transit**: TLS 1.3 for all database connections (enforce `sslmode=require`)
- **API Keys**: BLAKE2b hashing prevents plaintext storage

### 10.3 Access Control

- **Database Roles**:
  - `api_service_user`: SELECT, INSERT, UPDATE on urls, users, api_keys
  - `redirector_user`: SELECT on urls (read-only)
  - `analytics_processor_user`: INSERT on click_events, SELECT on continuous aggregates
- **Row-Level Security (RLS)**:
  - Users can only access their own URLs: `WHERE creator_user_id = current_user_id()`
  - Enforced at application layer (not PostgreSQL RLS for performance)

### 10.4 Audit Logging

- Enable PostgreSQL audit extension (pgAudit) for compliance
- Log DDL changes (CREATE, ALTER, DROP)
- Log DML on sensitive tables (users, api_keys)
- Rotate logs daily, retain 90 days

---

## 11. Performance Optimization

### 11.1 Connection Pooling

```go
// Go service connection pool configuration
dsn := "postgres://user:pass@localhost/refract?sslmode=require"
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)          // Max connections per service instance
db.SetMaxIdleConns(10)          // Idle connections kept warm
db.SetConnMaxLifetime(1 * time.Hour)  // Recycle connections hourly
```

### 11.2 Query Optimization

**EXPLAIN ANALYZE**:
- Profile slow queries (>100ms) with `EXPLAIN (ANALYZE, BUFFERS)`
- Target: Bitmap Index Scan or Index Scan (avoid Seq Scan on large tables)

**Common Optimizations**:
- Add covering indexes for frequently accessed columns
- Use `WHERE status = 'active'` filters (leverage partial indexes)
- Batch inserts for click_events (100 events per transaction)

### 11.3 Caching Strategy

**Application-Level Cache**:
- Cache Snowflake ID → destination URL mapping (Redis L2, in-memory L1)
- Cache user ID → email mapping (reduce Zitadel token introspection)
- Cache GeoIP lookups (IP → country/city, 24-hour TTL)

**Query Result Cache** (PostgreSQL):
- Enable shared_buffers = 25% of RAM
- Increase effective_cache_size = 75% of RAM
- Use prepared statements (query plan caching)

---

## Appendix A: Reserved Short Codes

The following short codes and custom aliases are reserved and cannot be used by users:

- `admin`, `api`, `app`, `auth`
- `dashboard`, `docs`, `help`, `health`
- `login`, `logout`, `register`, `signup`
- `settings`, `status`, `support`
- `www`, `web`, `assets`, `static`

**Enforcement**: Check constraint on `urls.custom_alias`:

```sql
ALTER TABLE urls ADD CONSTRAINT chk_custom_alias_not_reserved
    CHECK (custom_alias NOT IN ('admin', 'api', 'app', ...));
```

---

## Appendix B: Snowflake ID Format

**64-bit Structure**:
```
| 1 bit (unused) | 41 bits (timestamp) | 10 bits (worker ID) | 12 bits (sequence) |
|----------------|---------------------|---------------------|---------------------|
|       0        |   1704844800000     |        0023         |       4095          |
```

**Timestamp Epoch**: 2024-01-01 00:00:00 UTC (reduces bit usage vs Unix epoch)

**Base62 Encoding**:
- Alphabet: `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
- Example: `123456789012345678` → `dBvJIX9uyO` (10 characters)

**Decoding for Lookup**:
```
short_code "dBvJIX9uyO" → decode Base62 → snowflake_id 123456789012345678
                                             ↓
                                     SELECT * FROM urls WHERE snowflake_id = 123456789012345678
```

---

## Appendix C: Schema Versioning

**Current Schema Version**: `v1.0.0` (2026-01-09)

**Change Log**:
- `v1.0.0` (2026-01-09): Initial schema (users, urls, api_keys, click_events, continuous aggregates)

**Future Versions**:
- `v1.1.0`: Add `url_tags` table for categorization
- `v1.2.0`: Add `utm_parameters` to click_events for campaign tracking
- `v2.0.0`: Multi-tenancy support (organizations table)

---

**Document Status**: ✅ Complete - Ready for contract generation (Phase 1)

**Next Steps**:
1. Generate OpenAPI contracts for API service
2. Generate OpenAPI contracts for Redirector service
3. Define event schemas (Redis Stream click events JSON format)
4. Create `quickstart.md` with local development setup

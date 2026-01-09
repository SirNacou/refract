# Research: Authenticated URL Shortener Platform

**Feature**: 016-authenticated-url-shortener  
**Date**: 2026-01-09  
**Purpose**: Research technology choices, best practices, and architectural patterns for distributed URL shortener

---

## Overview

This document consolidates research findings to resolve technical decisions for building a distributed URL shortener with authentication, analytics, and API access. Research focused on performance optimization (< 50ms redirects), scalability (10K concurrent users), and operational simplicity aligned with Refract Constitution.

---

## 1. Authentication Architecture (Zitadel Integration)

### Decision
**Use Zitadel as Identity Provider (IdP) via OpenID Connect (OIDC)**

### Rationale
- **Zero Password Management**: Zitadel handles registration, login, password resets, email verification
- **Existing Infrastructure**: Zitadel already deployed and operational
- **Standards-Based**: OIDC/OAuth2 industry standard with robust library support (Go: `coreos/go-oidc`, TS: `@auth/core`)
- **Multi-Provider Support**: Zitadel natively supports Google, GitHub OAuth providers
- **Reduced Attack Surface**: No password hashing, session management, or CSRF complexity in application services

### Implementation Approach
1. **Frontend (TanStack Start)**: 
   - Use `@auth/core` with Zitadel provider
   - Authorization Code Flow with PKCE for SPAs
   - Store access token in memory (httpOnly cookies for refresh token)
   
2. **API Service (Go)**:
   - Validate JWT tokens on every request (`go-oidc` library)
   - Extract user subject (`sub` claim) as user ID
   - Cache Zitadel JWKS (JSON Web Key Set) with 1-hour TTL
   
3. **Token Validation Middleware**:
   ```go
   // Pseudocode
   func AuthMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           token := extractBearerToken(r)
           claims, err := verifier.Verify(r.Context(), token)
           if err != nil {
               http.Error(w, "Unauthorized", 401)
               return
           }
           ctx := context.WithValue(r.Context(), "userID", claims.Subject)
           next.ServeHTTP(w, r.WithContext(ctx))
       })
   }
   ```

### Alternatives Considered
- **Custom Auth (email/password)**: Rejected - reinvents wheel, increases security responsibility, violates DRY
- **Firebase Auth**: Rejected - vendor lock-in, monthly costs, Zitadel already deployed
- **Auth0**: Rejected - expensive at scale, Zitadel open-source and self-hosted

### Best Practices
- Never store full JWT in localStorage (XSS vulnerability)
- Use short-lived access tokens (15 min), long-lived refresh tokens (7 days)
- Implement token refresh logic before expiration
- Validate token signature + expiration + issuer on every request
- Use Zitadel's token introspection endpoint for API key validation

**References**:
- Zitadel OIDC Documentation: https://zitadel.com/docs/guides/integrate/login/oidc
- go-oidc Library: https://github.com/coreos/go-oidc

---

## 2. Short Code Generation (Snowflake ID + Base62)

### Decision
**Generate 64-bit Snowflake IDs, encode to Base62 for short codes**

### Rationale
- **Distributed Generation**: No central coordination required (each service has worker ID)
- **Collision-Free**: Guaranteed uniqueness across services
- **Time-Ordered**: IDs sortable by creation time (benefits analytics queries)
- **Reversible**: Decode Base62 → Snowflake ID for O(1) lookups (no short_code index needed)
- **Compact**: ~10-11 characters in Base62 (within spec's 7-10 char goal with optimization)

### Implementation Approach
1. **Snowflake ID Structure** (64 bits):
   ```
   1 bit: unused (always 0)
   41 bits: timestamp (milliseconds since epoch)
   10 bits: worker ID (1024 unique workers)
   12 bits: sequence (4096 IDs per millisecond per worker)
   ```

2. **Base62 Encoding**:
   ```
   Alphabet: 0-9, a-z, A-Z (62 characters, URL-safe)
   Snowflake ID 281474976710655 → Base62: "AzL8n0Y58m7" (11 chars)
   ```

3. **Optimization for Shorter Codes** (8-9 characters):
   - Use custom epoch (e.g., 2024-01-01) to reduce timestamp bits
   - Reduce worker ID bits to 6 (64 workers max)
   - Achieves ~8-9 characters for first 10 years

4. **Worker ID Assignment**:
   - Environment variable `WORKER_ID` (0-63 for API service instances)
   - Redirector service uses separate range (64-127)
   - Prevents collisions across services

### Code Example (Go)
```go
type SnowflakeGenerator struct {
    epoch      int64
    workerID   int64
    sequence   int64
    lastTimeMS int64
    mutex      sync.Mutex
}

func (g *SnowflakeGenerator) NextID() int64 {
    g.mutex.Lock()
    defer g.mutex.Unlock()
    
    nowMS := time.Now().UnixMilli() - g.epoch
    if nowMS == g.lastTimeMS {
        g.sequence = (g.sequence + 1) & 4095  // 12 bits
        if g.sequence == 0 {
            // Wait for next millisecond
            for nowMS <= g.lastTimeMS {
                nowMS = time.Now().UnixMilli() - g.epoch
            }
        }
    } else {
        g.sequence = 0
    }
    g.lastTimeMS = nowMS
    
    return (nowMS << 22) | (g.workerID << 12) | g.sequence
}

func EncodeBase62(id int64) string {
    const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    if id == 0 {
        return "0"
    }
    result := ""
    for id > 0 {
        result = string(alphabet[id%62]) + result
        id /= 62
    }
    return result
}
```

### Alternatives Considered
- **Random + Collision Check**: Rejected - requires DB query on every generation, 99th percentile latency
- **UUID v7**: Rejected - 22 chars in Base62 (too long), not integer (loses sortability benefits)
- **Sqids/Hashids with auto-increment**: Rejected - requires centralized sequence or DB sequence pooling in distributed setup
- **NanoID**: Rejected - probabilistic collisions (1% at 1B IDs), not reversible

### Best Practices
- Validate custom aliases separately (user-provided, 3-50 chars)
- Reserve short codes for edge cases: `admin`, `api`, `health` (prevent user claiming these)
- Index `snowflake_id` (primary key), `custom_alias` (unique), skip indexing `short_code` (derived field)
- Cache Snowflake ID → destination URL mapping (not short code → URL)

**References**:
- Twitter Snowflake: https://github.com/twitter-archive/snowflake
- Base62 Encoding: https://en.wikipedia.org/wiki/Base62

---

## 3. Multi-Tier Caching Strategy

### Decision
**L1: In-memory LRU cache (10K entries) | L2: Redis cache (1M entries, 1hr TTL)**

### Rationale
- **Performance**: L1 serves hot URLs in <1ms (no network I/O), L2 serves warm URLs in 2-5ms (Redis), cold misses hit DB in 10-20ms
- **Hit Rate**: 90%+ combined hit rate (Pareto principle: 20% of URLs generate 80% of traffic)
- **Scalability**: L2 shared across redirector instances (cache hits benefit all nodes)
- **Operational Simplicity**: Redis already required for rate limiting and event streams

### Implementation Approach
1. **L1 Cache (Redirector Service - Rust)**:
   ```rust
   use lru::LruCache;
   
   struct L1Cache {
       cache: Arc<Mutex<LruCache<i64, String>>>,
       max_entries: usize,
   }
   
   impl L1Cache {
       async fn get(&self, snowflake_id: i64) -> Option<String> {
           self.cache.lock().await.get(&snowflake_id).cloned()
       }
       
       async fn put(&self, snowflake_id: i64, destination: String) {
           self.cache.lock().await.put(snowflake_id, destination);
       }
   }
   ```

2. **L2 Cache (Redis)**:
   ```go
   // Key: "url:{snowflake_id}", Value: destination URL
   func GetURL(snowflakeID int64) (string, error) {
       key := fmt.Sprintf("url:%d", snowflakeID)
       dest, err := redisClient.Get(ctx, key).Result()
       if err == redis.Nil {
           return "", ErrCacheMiss
       }
       return dest, err
   }
   
   func SetURL(snowflakeID int64, destination string) error {
       key := fmt.Sprintf("url:%d", snowflakeID)
       return redisClient.Set(ctx, key, destination, 1*time.Hour).Err()
   }
   ```

3. **Cache Invalidation**:
   - **Update URL**: Invalidate L2 (publish Redis event), L1 expires naturally (LRU eviction)
   - **Delete URL**: Invalidate both L1+L2, set tombstone marker `url:{id} = DELETED` (1 day TTL)
   - **Deactivate URL**: Same as delete (prevent serving deactivated URLs from stale cache)

4. **Cache Warming**:
   - On redirector startup, pre-warm L1 with top 1000 most-clicked URLs (query analytics DB)
   - Background job refreshes L2 for URLs with clicks in last 24 hours

### Cache Flow Diagram
```
Request: GET /abc123
  ↓
Decode Base62 → Snowflake ID
  ↓
L1 Cache Lookup (in-memory)
  ├─ HIT → Return destination (301 redirect)
  └─ MISS ↓
       L2 Cache Lookup (Redis)
         ├─ HIT → Return destination + warm L1
         └─ MISS ↓
              Database Lookup (PostgreSQL)
                ├─ FOUND → Return destination + warm L2 + warm L1
                └─ NOT_FOUND → 404 page
```

### Alternatives Considered
- **Single-tier Redis only**: Rejected - every redirect requires network call (5ms overhead), can't meet <50ms p95 at 10K concurrent
- **CDN edge caching (Cloudflare)**: Rejected - can't invalidate instantly (updates take 5-30min), loses control over cache logic
- **No caching (DB only)**: Rejected - PostgreSQL can't handle 10K qps on single table, latency exceeds 100ms

### Best Practices
- Use separate Redis DB index for cache (e.g., DB 0) vs rate limiting (DB 1)
- Monitor cache hit rates (target: L1 >70%, L2 >20%, combined >90%)
- Set reasonable TTLs: L1 (no expiration, LRU eviction), L2 (1 hour)
- Implement circuit breaker: If Redis down, skip L2 (degrade to L1 + DB)

**References**:
- Rust LRU Crate: https://docs.rs/lru/latest/lru/
- Redis Caching Patterns: https://redis.io/docs/manual/patterns/caching/

---

## 4. Time-Series Analytics (TimescaleDB)

### Decision
**Use TimescaleDB (PostgreSQL extension) for click events and analytics aggregations**

### Rationale
- **Time-Series Optimized**: Hypertables automatically partition by time, 20x faster than standard PostgreSQL for time-range queries
- **Columnar Compression**: Reduces storage by 90% for historical data (600M rows → 60GB compressed)
- **Continuous Aggregates**: Materialized views auto-refresh hourly/daily metrics (enables <5sec dashboard queries)
- **Single Database Cluster**: Avoid operational complexity of separate ClickHouse/InfluxDB cluster
- **PostgreSQL Compatibility**: Same SQL, same tools (pg_dump, pgAdmin), same connection pooling

### Implementation Approach
1. **Hypertable Schema**:
   ```sql
   CREATE TABLE click_events (
       time TIMESTAMPTZ NOT NULL,
       url_id BIGINT NOT NULL,
       user_agent TEXT,
       referrer TEXT,
       ip_address INET,          -- Anonymized (last octet zeroed)
       country_code CHAR(2),
       city TEXT,
       device_type VARCHAR(20),
       browser VARCHAR(50)
   );
   
   SELECT create_hypertable('click_events', 'time');
   
   -- Compression policy: Compress chunks older than 7 days
   ALTER TABLE click_events SET (
       timescaledb.compress,
       timescaledb.compress_segmentby = 'url_id'
   );
   SELECT add_compression_policy('click_events', INTERVAL '7 days');
   
   -- Retention policy: Drop chunks older than 5 years
   SELECT add_retention_policy('click_events', INTERVAL '5 years');
   ```

2. **Continuous Aggregates**:
   ```sql
   -- Hourly aggregations for dashboard
   CREATE MATERIALIZED VIEW click_summary_hourly
   WITH (timescaledb.continuous) AS
   SELECT
       time_bucket('1 hour', time) AS hour,
       url_id,
       COUNT(*) AS total_clicks,
       COUNT(DISTINCT ip_address) AS unique_visitors,
       COUNT(*) FILTER (WHERE device_type = 'mobile') AS mobile_clicks
   FROM click_events
   GROUP BY hour, url_id;
   
   -- Refresh policy: Update every 5 minutes
   SELECT add_continuous_aggregate_policy('click_summary_hourly',
       start_offset => INTERVAL '1 day',
       end_offset => INTERVAL '5 minutes',
       schedule_interval => INTERVAL '5 minutes');
   ```

3. **Analytics Queries**:
   ```sql
   -- Top referrers for URL (last 7 days)
   SELECT referrer, COUNT(*) as clicks
   FROM click_events
   WHERE url_id = $1
       AND time >= NOW() - INTERVAL '7 days'
   GROUP BY referrer
   ORDER BY clicks DESC
   LIMIT 10;
   
   -- Click trend (hourly, last 24 hours)
   SELECT hour, total_clicks, unique_visitors
   FROM click_summary_hourly
   WHERE url_id = $1
       AND hour >= NOW() - INTERVAL '24 hours'
   ORDER BY hour;
   ```

4. **Event Ingestion Pipeline**:
   ```
   Redirector → Redis Stream (click events JSON)
       ↓
   Analytics Processor (Go consumer)
       ↓
   Batch insert to TimescaleDB (100 events per transaction)
   ```

### Alternatives Considered
- **PostgreSQL (standard tables)**: Rejected - full table scans on 600M rows, queries timeout, no auto-partitioning
- **ClickHouse**: Rejected - requires separate cluster, different SQL dialect, higher operational overhead
- **Elasticsearch**: Rejected - overkill for structured time-series (designed for full-text search), expensive at scale
- **InfluxDB**: Rejected - requires learning InfluxQL, separate backup/restore process, less mature than PostgreSQL

### Best Practices
- Partition hypertables by day (default), not hour (too many chunks)
- Always filter by time range in queries (enable chunk exclusion)
- Use `time_bucket()` for aggregations (TimescaleDB-optimized)
- Index frequently queried dimensions: `CREATE INDEX ON click_events (url_id, time DESC)`
- Monitor chunk count: `SELECT * FROM timescaledb_information.chunks` (keep under 10K)

**References**:
- TimescaleDB Documentation: https://docs.timescale.com/
- Compression Guide: https://docs.timescale.com/use-timescale/latest/compression/
- Continuous Aggregates: https://docs.timescale.com/use-timescale/latest/continuous-aggregates/

---

## 5. API Key Management (BLAKE2 Hashing)

### Decision
**Hash API keys with BLAKE2b, store hash + 8-char prefix, validate by re-hashing**

### Rationale
- **Security**: Database breach doesn't expose plaintext keys (unlike storing raw keys)
- **Performance**: BLAKE2b faster than bcrypt/Argon2 (optimized for short inputs), <1ms validation
- **Usability**: Store key prefix (`refract_abc12345...`) for user identification in logs/UI
- **Industry Standard**: GitHub, Stripe use similar pattern (hash + prefix)

### Implementation Approach
1. **Key Generation**:
   ```go
   import "crypto/rand"
   import "encoding/base64"
   
   func GenerateAPIKey() (key string, hash string, prefix string, err error) {
       // Generate 32 random bytes
       randomBytes := make([]byte, 32)
       if _, err := rand.Read(randomBytes); err != nil {
           return "", "", "", err
       }
       
       // Encode to base64 (URL-safe)
       key = "refract_" + base64.URLEncoding.EncodeToString(randomBytes)
       
       // Extract prefix (first 16 chars including "refract_")
       prefix = key[:16]
       
       // Hash full key with BLAKE2b-256
       hash = hashAPIKey(key)
       
       return key, hash, prefix, nil
   }
   
   func hashAPIKey(key string) string {
       h := blake2b.Sum256([]byte(key))
       return hex.EncodeToString(h[:])
   }
   ```

2. **Key Validation**:
   ```go
   func ValidateAPIKey(providedKey string) (userID string, err error) {
       // Hash provided key
       hash := hashAPIKey(providedKey)
       
       // Lookup in database by hash
       apiKey, err := repo.GetAPIKeyByHash(ctx, hash)
       if err != nil {
           return "", ErrInvalidAPIKey
       }
       
       // Check if revoked
       if apiKey.Status == "revoked" {
           return "", ErrAPIKeyRevoked
       }
       
       // Update last_used timestamp (async)
       go repo.UpdateLastUsed(ctx, apiKey.ID)
       
       return apiKey.UserID, nil
   }
   ```

3. **Database Schema**:
   ```sql
   CREATE TABLE api_keys (
       id BIGSERIAL PRIMARY KEY,
       user_id TEXT NOT NULL,                    -- Zitadel subject ID
       key_hash TEXT NOT NULL UNIQUE,            -- BLAKE2b hash
       key_prefix VARCHAR(16) NOT NULL,          -- For display (refract_abc12345)
       name VARCHAR(100),                        -- User-provided description
       status VARCHAR(20) DEFAULT 'active',      -- active, revoked
       created_at TIMESTAMPTZ DEFAULT NOW(),
       last_used_at TIMESTAMPTZ,
       usage_count BIGINT DEFAULT 0,
       FOREIGN KEY (user_id) REFERENCES users(zitadel_sub)
   );
   
   CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
   CREATE INDEX idx_api_keys_user ON api_keys(user_id) WHERE status = 'active';
   ```

4. **Rate Limiting**:
   ```go
   // Redis key: "ratelimit:apikey:{key_hash}:hour"
   func CheckRateLimit(keyHash string, limit int64) error {
       key := fmt.Sprintf("ratelimit:apikey:%s:hour", keyHash)
       count, err := redisClient.Incr(ctx, key).Result()
       if err != nil {
           return err
       }
       
       if count == 1 {
           // Set 1-hour expiration on first increment
           redisClient.Expire(ctx, key, 1*time.Hour)
       }
       
       if count > limit {
           return ErrRateLimitExceeded
       }
       return nil
   }
   ```

### Alternatives Considered
- **Plaintext storage**: Rejected - catastrophic if database breached (all keys compromised)
- **JWT tokens as API keys**: Rejected - stateless means can't revoke without blacklist, payload size larger
- **bcrypt/Argon2 hashing**: Rejected - too slow for API key validation (50-100ms), designed for passwords
- **HMAC-SHA256**: Rejected - requires secret key management, BLAKE2b simpler (no key needed)

### Best Practices
- Show full key only once at generation (never retrievable again)
- Display prefix in UI for identification: `refract_abc12345... (Created Jan 9, 2026)`
- Rate limit by key hash (1000 requests/hour per spec)
- Log API usage with key prefix (not full key) for debugging
- Implement key rotation: Allow users to generate new key, deprecate old after grace period

**References**:
- BLAKE2 Specification: https://www.blake2.net/
- GitHub API Key Security: https://github.blog/2021-04-05-behind-githubs-new-authentication-token-formats/

---

## 6. Safe Browsing Integration (Malicious URL Detection)

### Decision
**Integrate Google Safe Browsing API v4 (Update API + Lookup API)**

### Rationale
- **Comprehensive Database**: 1B+ malicious URLs (phishing, malware, unwanted software)
- **Real-Time Updates**: Database updated every 30 minutes
- **Cost-Effective**: 10,000 free queries/day, $0.50 per 1,000 additional (spec: 100 URLs/hour/user = ~50K queries/month = free tier)
- **Industry Standard**: Used by Chrome, Firefox, Safari

### Implementation Approach
1. **API Client (Go)**:
   ```go
   import "github.com/google/safebrowsing"
   
   sb, err := safebrowsing.NewSafeBrowser(safebrowsing.Config{
       APIKey: os.Getenv("SAFE_BROWSING_API_KEY"),
       DBPath: "/tmp/safebrowsing.db",
       UpdatePeriod: 30 * time.Minute,
   })
   
   func CheckURL(url string) (bool, error) {
       threats, err := sb.LookupURLs([]string{url})
       if err != nil {
           return false, err
       }
       
       // Any threat type = block creation
       return len(threats[0]) == 0, nil  // true = safe, false = malicious
   }
   ```

2. **Validation Flow**:
   ```
   User submits long URL
       ↓
   API validates format (valid HTTP/HTTPS)
       ↓
   Check Safe Browsing API (local cache + remote lookup)
       ├─ SAFE → Generate short code, create URL
       └─ MALICIOUS → Return 400 error: "URL flagged as malicious"
   ```

3. **Caching Strategy**:
   - Cache safe URLs for 24 hours (reduce API calls)
   - Cache malicious URLs for 7 days (rare changes)
   - Redis key: `safebrowsing:{hash(url)}` → `safe` or `malicious`

4. **Graceful Degradation**:
   ```go
   func CreateShortURL(destination string) error {
       safe, err := CheckURL(destination)
       if err != nil {
           // API down - log error, allow creation with flag
           log.Warn("Safe Browsing API unavailable", "error", err)
           // Option 1: Allow creation (track for retroactive check)
           // Option 2: Reject creation (fail-safe)
           return ErrSafeBrowsingUnavailable
       }
       
       if !safe {
           return ErrMaliciousURL
       }
       
       // Proceed with creation
       return nil
   }
   ```

### Alternatives Considered
- **URLhaus (abuse.ch)**: Rejected - smaller database (2M URLs), less frequent updates
- **VirusTotal**: Rejected - expensive ($100/month for commercial use), overkill for URL checking
- **Custom blacklist**: Rejected - requires constant maintenance, misses new threats
- **No validation**: Rejected - violates FR-042, enables phishing via short URLs

### Best Practices
- Enable all threat types: `MALWARE`, `SOCIAL_ENGINEERING`, `UNWANTED_SOFTWARE`
- Update local database every 30 minutes (Safe Browsing recommends)
- Implement retry with exponential backoff (API rate limits: 500 req/min)
- Log blocked URLs for manual review (false positives rare but possible)
- Provide appeal mechanism: User can request review via support

**References**:
- Google Safe Browsing API: https://developers.google.com/safe-browsing/v4
- Go Library: https://github.com/google/safebrowsing

---

## 7. Frontend Architecture (TanStack Start)

### Decision
**Use TanStack Start (React-based meta-framework) with TanStack Query for data fetching**

### Rationale
- **User Specification**: "Only frontend is tanstack start" (explicit requirement)
- **Modern Stack**: Built on Vite (fast HMR), supports SSR/SSG, file-based routing
- **Data Management**: TanStack Query handles caching, revalidation, optimistic updates (perfect for analytics dashboard)
- **Type Safety**: Full TypeScript support with shared types between frontend/backend

### Implementation Approach
1. **Project Structure**:
   ```
   frontend/
   ├── app/
   │   ├── routes/
   │   │   ├── index.tsx               # Landing page (public)
   │   │   ├── dashboard.tsx           # URL list (authenticated)
   │   │   ├── create.tsx              # URL creation form
   │   │   ├── analytics.$id.tsx       # Analytics detail
   │   │   └── settings/
   │   │       └── api-keys.tsx        # API key management
   │   ├── components/
   │   │   ├── AuthGuard.tsx           # Protected route wrapper
   │   │   ├── URLList.tsx
   │   │   ├── AnalyticsCharts.tsx     # Recharts visualizations
   │   │   └── URLForm.tsx
   │   └── utils/
   │       ├── api-client.ts           # Fetch wrapper with auth headers
   │       └── auth.ts                 # Zitadel OIDC integration
   ```

2. **Authentication Integration**:
   ```tsx
   // app/utils/auth.ts
   import { createServerFn } from '@tanstack/start'
   import { getWebRequest } from 'vinxi/http'
   
   export const getUser = createServerFn('GET', async () => {
       const request = getWebRequest()
       const token = request.headers.get('Authorization')?.split(' ')[1]
       
       if (!token) return null
       
       // Validate JWT with Zitadel
       const user = await validateToken(token)
       return user
   })
   
   // app/components/AuthGuard.tsx
   export function AuthGuard({ children }: { children: React.ReactNode }) {
       const { data: user, isLoading } = useQuery({
           queryKey: ['user'],
           queryFn: getUser
       })
       
       if (isLoading) return <div>Loading...</div>
       if (!user) {
           // Redirect to Zitadel login
           window.location.href = '/api/auth/login'
           return null
       }
       
       return <>{children}</>
   }
   ```

3. **Data Fetching (TanStack Query)**:
   ```tsx
   // app/routes/dashboard.tsx
   import { useQuery } from '@tanstack/react-query'
   
   export default function Dashboard() {
       const { data: urls, isLoading } = useQuery({
           queryKey: ['urls'],
           queryFn: async () => {
               const res = await fetch('/api/urls', {
                   headers: { Authorization: `Bearer ${getToken()}` }
               })
               return res.json()
           },
           refetchInterval: 30000  // Refresh every 30s
       })
       
       if (isLoading) return <div>Loading URLs...</div>
       
       return (
           <div>
               <h1>My Short URLs</h1>
               <URLList urls={urls} />
           </div>
       )
   }
   ```

4. **Real-Time Analytics**:
   ```tsx
   // app/routes/analytics.$id.tsx
   import { useQuery } from '@tanstack/react-query'
   import { LineChart, Line, XAxis, YAxis } from 'recharts'
   
   export default function Analytics({ params }: { params: { id: string } }) {
       const { data: analytics } = useQuery({
           queryKey: ['analytics', params.id],
           queryFn: async () => {
               const res = await fetch(`/api/analytics/${params.id}`)
               return res.json()
           },
           refetchInterval: 5000  // FR-024: 5-second real-time updates
       })
       
       return (
           <div>
               <h1>Analytics for {analytics.shortCode}</h1>
               <LineChart data={analytics.hourlyClicks}>
                   <XAxis dataKey="hour" />
                   <YAxis />
                   <Line type="monotone" dataKey="clicks" stroke="#8884d8" />
               </LineChart>
           </div>
       )
   }
   ```

### Best Practices
- Use Server Functions for authenticated API calls (SSR-safe)
- Implement optimistic updates for URL creation (immediate UI feedback)
- Enable Query DevTools in development
- Prefetch analytics data on dashboard hover
- Use Suspense boundaries for loading states

**References**:
- TanStack Start: https://tanstack.com/start
- TanStack Query: https://tanstack.com/query
- Recharts: https://recharts.org/

---

## 8. IP Geolocation (MaxMind GeoLite2)

### Decision
**Use MaxMind GeoLite2 City database (self-hosted, free)**

### Rationale
- **Free Tier**: GeoLite2 City database free for commercial use (updated weekly)
- **Offline Lookup**: No external API calls (low latency, no rate limits)
- **City-Level Accuracy**: 55-80% accuracy for country/city (sufficient for analytics)
- **Small Database**: 70MB compressed (loads in memory for <1ms lookups)

### Implementation Approach
1. **Database Setup**:
   ```bash
   # Download GeoLite2 City database (requires MaxMind account - free)
   curl -o GeoLite2-City.tar.gz \
       "https://download.maxmind.com/app/geoip_download?license_key=YOUR_KEY&suffix=tar.gz"
   
   # Extract .mmdb file
   tar -xzf GeoLite2-City.tar.gz
   mv GeoLite2-City_*/GeoLite2-City.mmdb /var/lib/geoip/
   
   # Weekly cron job to update database
   0 2 * * 0 /usr/local/bin/update-geoip.sh
   ```

2. **Go Integration**:
   ```go
   import "github.com/oschwald/geoip2-golang"
   
   db, err := geoip2.Open("/var/lib/geoip/GeoLite2-City.mmdb")
   defer db.Close()
   
   func GetGeoData(ipStr string) (*GeoData, error) {
       ip := net.ParseIP(ipStr)
       record, err := db.City(ip)
       if err != nil {
           return nil, err
       }
       
       return &GeoData{
           CountryCode: record.Country.IsoCode,
           CountryName: record.Country.Names["en"],
           City: record.City.Names["en"],
           Latitude: record.Location.Latitude,
           Longitude: record.Location.Longitude,
       }, nil
   }
   ```

3. **Privacy (IP Anonymization)**:
   ```go
   func AnonymizeIP(ip net.IP) net.IP {
       if ip.To4() != nil {
           // IPv4: Zero last octet (192.168.1.100 → 192.168.1.0)
           ip[3] = 0
       } else {
           // IPv6: Zero last 80 bits
           for i := 6; i < 16; i++ {
               ip[i] = 0
           }
       }
       return ip
   }
   ```

4. **Caching Strategy**:
   - Cache IP → GeoData for 24 hours (IPs rarely change location)
   - Redis key: `geoip:{anonymized_ip}` → JSON geo data
   - Reduces mmdb lookups by 95% (same IPs repeat across clicks)

### Alternatives Considered
- **MaxMind GeoIP2 Precision** (paid): Rejected - $50/month, overkill for analytics
- **IPinfo.io API**: Rejected - 50K free requests/month (insufficient), $99/month for 250K
- **ipapi.co**: Rejected - 1K free requests/day (insufficient for 10M events/month)
- **Cloudflare Workers** (CF-IPCountry header): Rejected - requires Cloudflare proxy, not self-hosted

### Best Practices
- Update GeoLite2 database weekly (accuracy improves with updates)
- Handle missing data gracefully (some IPs have no city, only country)
- Use connection pooling for mmdb file (reader is thread-safe)
- Monitor lookup latency (target <1ms)

**References**:
- MaxMind GeoLite2: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
- Go Library: https://github.com/oschwald/geoip2-golang

---

## Summary of Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Authentication** | Zitadel (OIDC) | Existing deployment, zero password management |
| **API Service** | Go 1.22 + Chi router | CQRS handlers, domain logic, proven in project history |
| **Redirector** | Rust 1.75 + Axum | <50ms latency requirement, zero-cost async |
| **Analytics Processor** | Go 1.22 | Event stream consumer, batch inserts |
| **Frontend** | TanStack Start + React | User-specified, SSR/SSG, TanStack Query for data |
| **Primary Database** | PostgreSQL 16 | Relational data (users, URLs, API keys) |
| **Analytics Database** | TimescaleDB (PG extension) | Time-series optimization, 20x faster queries |
| **Cache** | Redis/Valkey 7.2+ | L2 distributed cache, rate limiting, event streams |
| **Short Code** | Snowflake ID + Base62 | Distributed generation, collision-free, 8-11 chars |
| **API Keys** | BLAKE2b hashing | Secure storage, fast validation (<1ms) |
| **Safe Browsing** | Google Safe Browsing API v4 | 1B+ malicious URLs, free tier sufficient |
| **Geolocation** | MaxMind GeoLite2 City | Offline lookups, free, 55-80% city accuracy |

---

## Next Steps

1. **Phase 1 - Design**:
   - Create data-model.md (entity schemas, relationships)
   - Generate API contracts (OpenAPI specs for API + Redirector services)
   - Define event schemas (click events JSON format)
   - Write quickstart.md (local development setup)

2. **Phase 2 - Implementation** (via `/speckit.tasks`):
   - Setup microservices structure
   - Implement domain layer (URL entity, ShortCode value object)
   - Build CQRS handlers (commands + queries)
   - Integrate Zitadel authentication
   - Implement multi-tier caching
   - Setup TimescaleDB hypertables
   - Build analytics processor
   - Create frontend with TanStack Start

3. **Phase 3 - Testing & Deployment**:
   - Write contract tests for API endpoints
   - Integration tests for redirector + cache
   - Load testing (validate <50ms p95 at 10K concurrent)
   - Docker Compose for local dev
   - Production deployment (Kubernetes or cloud VMs)


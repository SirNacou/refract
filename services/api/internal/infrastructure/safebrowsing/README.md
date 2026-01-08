# Safe Browsing Package

This package provides Google Safe Browsing (GSB) compliant URL canonicalization for the Refract URL shortener service.

## Purpose

URL canonicalization is the process of normalizing URLs to a standard form according to Google Safe Browsing specifications. This ensures consistent URL representation when checking URLs against the GSB API for malicious content.

**Important**: This package currently implements **canonicalization only**. It does NOT yet integrate with the Google Safe Browsing API. That integration is planned for future phases.

## Why Canonicalization?

When checking URLs against the GSB API, different representations of the same URL (e.g., `HTTP://EXAMPLE.COM:80/` vs `http://example.com/`) must be normalized to ensure consistent threat detection. The GSB API requires URLs to be canonicalized before lookup.

### Use Case in Refract

1. **Storage**: Store the user's original URL exactly as provided (no canonicalization)
2. **GSB Check**: When checking against GSB API (future), canonicalize the URL transiently for lookup
3. **Privacy**: Each URL shortening request creates a unique short link (no deduplication)

## Canonicalization Rules

This implementation follows the [Google Safe Browsing v4 URL Canonicalization Specification](https://developers.google.com/safe-browsing/v4/urls-hashing#canonicalization) strictly:

1. **Remove control characters**: Tab (`\t`), carriage return (`\r`), line feed (`\n`)
2. **Lowercase scheme**: `HTTP` → `http`
3. **Lowercase host**: `EXAMPLE.COM` → `example.com`
4. **Convert IDN to Punycode**: `münchen.de` → `xn--mnchen-3ya.de`
5. **Remove default ports**: `:80` for `http`, `:443` for `https`
6. **Normalize percent encoding**:
   - Decode unreserved characters: `%7E` → `~`, `%41` → `A`
   - Uppercase hex digits for reserved: `%2f` → `%2F`
   - Keep reserved characters encoded: `%2F`, `%3F`, etc.
7. **Normalize path**: Remove `.`, `..`, and `//` segments
8. **Remove fragments**: `#section` is stripped
9. **Preserve query parameters**: Query string kept as-is (no sorting, no removal)

### Unreserved Characters (Always Decoded)

Per RFC 3986: `A-Z a-z 0-9 - . _ ~`

### Reserved Characters (Keep Encoded)

Per RFC 3986: `:/?#[]@!$&'()*+,;=`

## Usage

### Basic Canonicalization

```go
package main

import (
    "fmt"
    "github.com/yourusername/refract/services/api/internal/infrastructure/safebrowsing"
)

func main() {
    canonicalizer := safebrowsing.NewCanonicalizer()
    
    canonical, err := canonicalizer.CanonicalizeForGSB("HTTP://EXAMPLE.COM:80/path?query=value#fragment")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(canonical)
    // Output: http://example.com/path?query=value
}
```

### Batch Processing

```go
urls := []string{
    "HTTP://EXAMPLE.COM:80/",
    "https://münchen.de/path",
    "http://example.com/%7Euser/",
}

results, errs := canonicalizer.CanonicalizeBatch(urls)

for i, result := range results {
    if errs[i] != nil {
        fmt.Printf("Error: %v\n", errs[i])
        continue
    }
    fmt.Printf("%s -> %s\n", urls[i], result)
}
```

### Integration Example (Future)

```go
// In your URL creation handler
func (h *CreateURLHandler) Handle(cmd CreateURLCommand) error {
    // 1. Validate and store original URL (no canonicalization)
    originalURL := cmd.OriginalURL
    
    // 2. Canonicalize for GSB check only
    canonicalizer := safebrowsing.NewCanonicalizer()
    canonicalURL, err := canonicalizer.CanonicalizeForGSB(originalURL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    
    // 3. Check against GSB API (future implementation)
    isMalicious, err := gsbClient.CheckURL(canonicalURL)
    if err != nil {
        // Log error, decide whether to fail or proceed
    }
    if isMalicious {
        return errors.New("URL flagged as malicious by Safe Browsing")
    }
    
    // 4. Create short link with ORIGINAL URL
    shortCode := generateShortCode()
    url := domain.NewURL(originalURL, shortCode)
    return h.repo.Save(url)
}
```

## Examples

### Basic Normalizations

```go
// Scheme and host normalization
"HTTP://EXAMPLE.COM/" → "http://example.com/"

// Default port removal
"http://example.com:80/" → "http://example.com/"
"https://example.com:443/" → "https://example.com/"

// Fragment removal
"http://example.com/#fragment" → "http://example.com/"

// Path normalization
"http://example.com/./path/../file" → "http://example.com/file"
"http://example.com//path//file" → "http://example.com/path/file"
```

### Percent Encoding

```go
// Decode unreserved characters
"http://example.com/%7Euser" → "http://example.com/~user"
"http://example.com/file%2Etxt" → "http://example.com/file.txt"

// Uppercase hex for reserved characters
"http://example.com/path%2fto%2ffile" → "http://example.com/path%2Fto%2Ffile"

// Mixed example
"http://example.com/%7Epath%2f%41file" → "http://example.com/~path%2FAfile"
```

### Internationalized Domain Names (IDN)

```go
// Convert IDN to Punycode
"http://münchen.de/" → "http://xn--mnchen-3ya.de/"
"http://日本.jp/path" → "http://xn--wgv71a.jp/path"
"http://كوم.example/test" → "http://xn--mgbh0fb.example/test"
```

### Complex Real-World URLs

```go
// Multiple normalizations
"HTTP://WWW.EXAMPLE.COM:80/./path/%7Euser/../file?b=2&a=1#section"
→ "http://www.example.com/path/~user/file?b=2&a=1"

// Control characters
"http://example.com/pa\r\nth?q=\tval" → "http://example.com/path?q=val"

// IPv6 addresses (preserved as-is)
"http://[2001:db8::1]:80/" → "http://[2001:db8::1]/"
```

## Testing

### Run Tests

```bash
cd services/api
go test ./internal/infrastructure/safebrowsing/... -v
```

### Check Coverage

```bash
go test ./internal/infrastructure/safebrowsing/... -cover
```

Target: >90% code coverage

### Run Benchmarks

```bash
go test -bench=. -benchmem ./internal/infrastructure/safebrowsing/...
```

Expected performance: <10μs per URL

### Official GSB Test Vectors

The test suite includes official test vectors from Google's Safe Browsing specification to ensure compliance:

```go
// These are real test cases from Google's GSB spec
"http://host/%25%32%35" → "http://host/%25"
"http://host/%25%32%35%25%32%35" → "http://host/%25%25"
"http://host/%2525252525252525" → "http://host/%25"
"http://host/asdf%25%32%35asd" → "http://host/asdf%25asd"
// ... and many more
```

## Integration Guide: Future Phases

### Phase 1: GSB API Client (4-6 hours)

**Goal**: Implement Google Safe Browsing API v4 client for threat detection.

**Files to Create**:
```
services/api/internal/infrastructure/safebrowsing/
├── client.go           (GSB API client)
├── client_test.go      (Client tests)
├── cache.go            (In-memory threat cache)
└── types.go            (API request/response types)
```

**Steps**:
1. Get GSB API key from [Google Cloud Console](https://console.cloud.google.com/)
2. Implement `Client` struct with methods:
   - `CheckURL(canonicalURL string) (ThreatInfo, error)`
   - `CheckURLs([]string) ([]ThreatInfo, error)` - batch checking
3. Implement local caching to reduce API calls:
   - Cache negative results (safe URLs) for 30 minutes
   - Cache positive results (threats) for 5 minutes
   - Use LRU eviction with max 10,000 entries
4. Add retry logic with exponential backoff
5. Handle API rate limiting (10,000 requests/day for free tier)

**Configuration**:
```go
// Add to internal/config/config.go
type Config struct {
    // ... existing fields ...
    GSBAPIKey     string `env:"GSB_API_KEY"`
    GSBCacheSize  int    `env:"GSB_CACHE_SIZE" envDefault:"10000"`
    GSBCacheTTL   int    `env:"GSB_CACHE_TTL_MINUTES" envDefault:"30"`
}
```

**Example Client Usage**:
```go
client := safebrowsing.NewClient(apiKey)

canonical, _ := canonicalizer.CanonicalizeForGSB(userURL)
threat, err := client.CheckURL(canonical)

if err != nil {
    // Log and decide: fail safe (reject) or fail open (allow)
}

if threat.IsMalicious {
    return errors.New("URL flagged as malicious")
}
```

### Phase 2: Application Integration (3-4 hours)

**Goal**: Wire GSB checking into URL creation flow.

**Files to Modify**:
```
services/api/internal/
├── app/dependencies.go           (Initialize GSB client)
├── application/commands/
│   └── create_url.go             (Add GSB check before save)
└── domain/url/errors.go          (Add ErrMaliciousURL)
```

**Steps**:
1. Initialize `safebrowsing.Client` in `app/dependencies.go`
2. Inject GSB client into `CreateURLCommandHandler`
3. Add GSB check before creating URL:
   ```go
   // In create_url.go Handle method
   canonical, err := h.canonicalizer.CanonicalizeForGSB(cmd.OriginalURL)
   if err != nil {
       return domain.ErrInvalidURL
   }
   
   threat, err := h.gsbClient.CheckURL(canonical)
   if err != nil {
       h.logger.Error("GSB check failed", "error", err)
       // Decision: fail safe (reject) or fail open (allow)?
       // Recommend: fail open for better UX, log for monitoring
   }
   
   if threat.IsMalicious {
       return domain.ErrMaliciousURL
   }
   ```
4. Add observability: metrics for GSB check latency, cache hit rate, failures
5. Add feature flag: `ENABLE_GSB_CHECK=true` to allow disabling

**Error Handling Strategy**:
- **GSB API down**: Log error, allow URL creation (fail open)
- **Malicious URL**: Reject with clear error message
- **Invalid URL**: Reject before GSB check (no API call needed)

### Phase 3: Database Caching (4-6 hours)

**Goal**: Persist GSB check results to avoid redundant API calls.

**Migration**:
```sql
-- Migration: 00004_add_gsb_fields.sql
ALTER TABLE urls ADD COLUMN gsb_status VARCHAR(20);
ALTER TABLE urls ADD COLUMN gsb_checked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE urls ADD COLUMN gsb_threats JSONB;

CREATE INDEX idx_urls_gsb_checked_at ON urls(gsb_checked_at) 
WHERE gsb_checked_at IS NOT NULL;

-- Enum values: 'safe', 'malicious', 'pending', 'error'
```

**Files to Modify**:
```
services/api/internal/
├── domain/url/entity.go                    (Add GSB fields)
├── infrastructure/persistence/postgres/
│   ├── queries/urls.sql                    (Add GSB columns)
│   └── url_repository.go                   (Update Save/FindBy methods)
└── application/commands/create_url.go      (Check DB cache first)
```

**Logic Flow**:
```go
// 1. Check database cache first
existingURL, _ := h.repo.FindByOriginalURL(cmd.OriginalURL)
if existingURL != nil && existingURL.GSBCheckedAt.After(time.Now().Add(-24*time.Hour)) {
    // Use cached result if checked within last 24 hours
    if existingURL.GSBStatus == "malicious" {
        return domain.ErrMaliciousURL
    }
    // Safe, proceed with new short link creation
}

// 2. If not cached or stale, check API
canonical, _ := h.canonicalizer.CanonicalizeForGSB(cmd.OriginalURL)
threat, err := h.gsbClient.CheckURL(canonical)

// 3. Save result to URL entity
url.GSBStatus = threat.Status
url.GSBCheckedAt = time.Now()
url.GSBThreats = threat.ToJSON()

h.repo.Save(url)
```

**Benefits**:
- Reduce GSB API calls (stay within free tier)
- Faster URL creation (cache hit = no API call)
- Historical record of threat status changes

### Phase 4: Background Re-checking (4-6 hours)

**Goal**: Periodically re-check URLs as they can become malicious over time.

**Files to Create**:
```
services/api/internal/
├── jobs/
│   ├── gsb_recheck.go          (Background job)
│   └── gsb_recheck_test.go
└── app/jobs.go                  (Job scheduler setup)
```

**Steps**:
1. Implement cron job to re-check URLs:
   - Check URLs last checked >7 days ago
   - Batch process 1000 URLs per run
   - Run every 6 hours
2. Handle URLs that become malicious:
   - Set `is_active = false`
   - Update `gsb_status = 'malicious'`
   - Log for admin review
3. Add admin notification system:
   - Email/Slack alert for newly malicious URLs
   - Dashboard to review flagged URLs
4. Consider: Add `deactivation_reason` column

**Cron Setup** (using `github.com/robfig/cron/v3`):
```go
// In app/jobs.go
func (a *Application) StartJobs() {
    c := cron.New()
    
    // Re-check stale URLs every 6 hours
    c.AddFunc("0 */6 * * *", func() {
        job := jobs.NewGSBRecheckJob(a.urlRepo, a.gsbClient)
        if err := job.Run(); err != nil {
            a.logger.Error("GSB recheck job failed", "error", err)
        }
    })
    
    c.Start()
}
```

**Database Query**:
```sql
-- Find URLs needing re-check
SELECT id, original_url, gsb_checked_at
FROM urls
WHERE is_active = true
  AND gsb_checked_at < NOW() - INTERVAL '7 days'
  AND (gsb_status = 'safe' OR gsb_status IS NULL)
ORDER BY gsb_checked_at ASC NULLS FIRST
LIMIT 1000;
```

## Architecture Decisions

### Why Store Original URLs (Not Canonical)?

1. **User Intent**: Preserve exactly what the user provided
2. **Query Params Matter**: For tracking, analytics, referrals (e.g., `?utm_source=twitter`)
3. **Fragment Use Cases**: Some apps use fragments for client-side routing
4. **Transparency**: Users might want to see their exact input when managing links
5. **GSB Requirement**: Canonicalize transiently only for API lookup, not storage

### Why No Deduplication?

1. **Privacy-first**: Each request creates a unique link (no tracking across users)
2. **Query Params**: `?ref=alice` vs `?ref=bob` should be different links
3. **Expiration**: Users might want different expiration times for same URL
4. **Simpler**: No complex canonical_url matching logic needed

### Why No `is_public` Field?

1. **Out of Scope**: Current requirements don't include public link pools
2. **YAGNI**: Add only when needed (avoid premature optimization)
3. **Auth First**: Would need authentication before public/private makes sense

## Performance Considerations

### Canonicalization Performance

- **Target**: <10μs per URL on modern hardware
- **Batch Processing**: Use `CanonicalizeBatch()` for multiple URLs
- **Memory**: Minimal allocations, uses `strings.Builder` internally

### GSB API Performance (Future)

- **API Latency**: ~100-300ms per request (external API call)
- **Caching Critical**: In-memory cache reduces API calls by 80-90%
- **Batch Checking**: GSB API supports checking up to 500 URLs in one request
- **Rate Limits**: Free tier = 10,000 requests/day (monitor usage)

### Database Performance (Future)

- **Index**: `gsb_checked_at` for efficient re-check queries
- **Partial Index**: Only index non-null values to save space
- **JSONB**: `gsb_threats` uses JSONB for flexible threat data storage

## Security Considerations

### API Key Management

```bash
# Never commit API keys to git
# Store in environment variables
export GSB_API_KEY="your-api-key-here"

# Or use secret management service
# AWS Secrets Manager, Google Secret Manager, HashiCorp Vault
```

### Fail-Safe vs Fail-Open

**Recommendation: Fail-Open** (allow URL creation if GSB API fails)

**Reasoning**:
- Better user experience (service stays available)
- GSB API outages shouldn't break core functionality
- Still log failures for monitoring
- Can switch to fail-safe via feature flag if abuse detected

### Privacy

- **No Logging Original URLs**: Log canonical URLs only in GSB checks
- **GDPR Compliance**: Document GSB data sharing in privacy policy
- **User Notification**: Inform users that URLs are checked for safety

## Dependencies

### Current Dependencies

```go
// go.mod
require (
    golang.org/x/net v0.34.0  // For idna.Punycode (IDN support)
)
```

### Future Dependencies (Phase 1)

```go
// Additional dependencies for GSB API integration
require (
    github.com/hashicorp/golang-lru/v2 v2.0.7  // For LRU cache
    github.com/cenkalti/backoff/v4 v4.2.1       // For retry with exponential backoff
)
```

## Monitoring & Observability (Future)

### Metrics to Track

```go
// GSB Check Metrics
gsb_check_total{status="safe|malicious|error"}
gsb_check_duration_seconds{cache_hit="true|false"}
gsb_cache_hit_ratio
gsb_api_errors_total{error_type="timeout|rate_limit|auth|unknown"}

// URL Creation Metrics
url_creation_total{rejected_reason="malicious|invalid|error"}
url_creation_duration_seconds
```

### Logging

```go
// Structured logging example
logger.Info("GSB check completed",
    "canonical_url", canonical,
    "status", threat.Status,
    "cache_hit", cacheHit,
    "duration_ms", duration.Milliseconds(),
)
```

## References

- [Google Safe Browsing API v4](https://developers.google.com/safe-browsing/v4)
- [GSB URL Canonicalization Spec](https://developers.google.com/safe-browsing/v4/urls-hashing#canonicalization)
- [RFC 3986 - URI Generic Syntax](https://tools.ietf.org/html/rfc3986)
- [RFC 5891 - IDNA: Punycode](https://tools.ietf.org/html/rfc5891)

## Contributing

### Running Tests Before Commit

```bash
# Format code
go fmt ./internal/infrastructure/safebrowsing/...

# Vet code
go vet ./internal/infrastructure/safebrowsing/...

# Run tests
go test ./internal/infrastructure/safebrowsing/... -v -cover

# Run benchmarks
go test -bench=. -benchmem ./internal/infrastructure/safebrowsing/...
```

### Adding New Test Cases

When adding tests, prioritize:
1. **Official GSB vectors**: Always test against Google's spec
2. **Edge cases**: IPv6, IDN, malformed URLs, empty strings
3. **Real-world URLs**: Complex URLs found in production
4. **Performance**: Add benchmarks for performance-critical paths

## License

This package is part of the Refract URL shortener project. See repository root for license information.

---

**Status**: Canonicalization implementation complete. GSB API integration pending (Phases 1-4).

**Last Updated**: January 2026

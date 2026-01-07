# Architecture Decisions

**TL;DR**: We use Domain-Driven Design (DDD) for isolated business logic + CQRS pattern (commands write, queries read) + clean infrastructure adapters. This makes the code testable, maintainable, and independent of frameworks.

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│         HTTP Layer (Infrastructure)                  │
│  Router → Middleware → Handlers → DTOs              │
│  (Maps HTTP requests to domain commands/queries)    │
└──────────────────┬──────────────────────────────────┘
                   │ depends on
┌──────────────────▼──────────────────────────────────┐
│      Application Layer (CQRS - Use Cases)            │
│  CreateURLCommand → CreateURLHandler                │
│  GetURLQuery → GetURLHandler                        │
│  (Orchestrates domain logic, validates inputs)      │
└──────────────────┬──────────────────────────────────┘
                   │ uses
┌──────────────────▼──────────────────────────────────┐
│     Domain Layer (Pure Business Logic)               │
│  URL Entity • Value Objects • Domain Errors         │
│  Interfaces: Repository • Generator • Validator     │
│  (No frameworks, no HTTP, no database knowledge)    │
└──────────────────┬──────────────────────────────────┘
                   │ implemented by
┌──────────────────▼──────────────────────────────────┐
│        Infrastructure Layer (Adapters)               │
│  PostgreSQL Repository • Sqids Generator            │
│  Whitelist Domain Validator                         │
│  (Concrete implementations of domain interfaces)    │
└─────────────────────────────────────────────────────┘
```

---

## Core Design Decisions

### 1. Why Domain-Driven Design (DDD)?

**The Problem**: If business logic lives everywhere (handlers, database layer, utils), it becomes:
- Hard to test (need DB mocks, HTTP mocks)
- Coupled to frameworks (change HTTP library = change logic)
- Scattered (where do you look for a business rule?)

**Our Solution**: Isolate all business logic in `internal/domain/url/`

**Benefits**:
```
Domain Layer (internal/domain/url/)
├── NO imports from HTTP, database, or external libraries
├── Pure functions and value objects
├── Easy to test (just call functions, no mocking)
├── Framework-independent (could use this logic in CLI, gRPC, etc.)
└── Single source of truth for business rules
```

**Example**: URL expiration logic lives in domain, not in handlers:
```go
// In domain/url/entity.go (no framework dependencies)
func (u *URL) IsExpired() bool {
    return time.Now().After(u.expiresAt)
}

// In HTTP handler (can use domain logic)
func (h *URLHandler) GetURLMetadata(w http.ResponseWriter, r *http.Request) {
    result, err := h.getHandler.Handle(r.Context(), query)
    if result.URL.IsExpired() {
        // Domain logic is testable: no HTTP needed
    }
}
```

---

### 2. Why CQRS (Command Query Responsibility Segregation)?

**The Problem**: In typical CRUD, create and read logic live together, but they have different concerns:
- **Commands** (writes) need validation, business rules, side effects
- **Queries** (reads) just need to fetch and return data

**Our Solution**: Separate into two handlers

**Commands** (for writes):
```go
// CreateURLCommand - change state
type CreateURLCommand struct {
    OriginalURL string
    Domain      string
    ExpiresAt   *time.Time
}

// Responsibilities:
// 1. Validate inputs (value objects)
// 2. Check business rules (domain whitelist)
// 3. Generate short code
// 4. Persist to database
// 5. Return result
```

**Queries** (for reads):
```go
// GetURLQuery - read state, no changes
type GetURLQuery struct {
    ShortCode string
}

// Responsibilities:
// 1. Fetch from database
// 2. Check if valid (expired, inactive)
// 3. Return result
```

**Benefits**:
- Clear intent (is this changing data or reading it?)
- Easy to scale differently (cache queries, but not commands)
- Easier to test (command logic separate from query logic)
- Ready for CQRS databases later (separate write/read stores)

---

### 3. Why Sqids for Short Code Generation?

**Options Considered**:
1. **Random alphanumeric** (UUID, etc.) → Risk of collisions, not reversible
2. **Increment counter** (simple IDs) → Predictable, sequential, not hard to guess
3. **Sqids (bit-shuffling)** → Deterministic, zero collisions, reversible ✅

**Our Choice**: Sqids

**Why Sqids?**
```
┌─────────────────────────────────────────┐
│ Process: ID → Encode → Short Code       │
├─────────────────────────────────────────┤
│ ID 1      → "UkLWZg"                    │
│ ID 2      → "gbHJdm"                    │
│ ID 3      → "VqXmZF"                    │
└─────────────────────────────────────────┘

Benefits:
✅ Deterministic: Same ID always = same code
✅ No collisions: Each ID has unique code
✅ Reversible: Can decode "UkLWZg" → 1
✅ Compact: 6 chars instead of 36 (UUID)
✅ No reserved words: Can configure alphabet
```

**Usage**:
```go
// Create: Reserve ID from sequence → Encode with Sqids
id := repo.NextID(ctx)              // Returns 1
shortCode := generator.Generate(id) // Returns "UkLWZg"
url := NewURL(id, shortCode, ...)   // Save with both

// Redirect: Decode short code → Lookup by ID (index is fast)
id := generator.Decode("UkLWZg")    // Returns 1
url := repo.FindByID(ctx, id)       // O(1) lookup
```

---

### 4. Two Expiration Types (Fixed vs Activity-Based)

**The Problem**: Different URLs need different expiration strategies:
- Marketing campaigns: Expire at specific date (never renew)
- Personal links: Expire from inactivity (renew on use)

**Our Solution**: Two types, chosen at creation time

#### Type A: Fixed Expiration
```go
// User specifies when URL expires
urlEntity, err := url.NewURLWithFixedExpiration(
    id, shortCode, originalURL, domain,
    time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC),
)

// URL expires at that time, NEVER renews
hasFixedExpiration: true
```

**Use Cases**:
- Time-limited promotions ("Black Friday sale expires Jan 10")
- Event-specific links ("Conference code valid until May 15")
- Temporary access ("Download link valid for 7 days")

#### Type B: Activity-Based Expiration
```go
// User does NOT provide expiration
urlEntity := url.NewURLWithActivityExpiration(
    id, shortCode, originalURL, domain,
)

// URL expires 12 months from now, renews on every click
hasFixedExpiration: false
expiresAt: now + 12 months
```

**Use Cases**:
- Personal sharing ("Share your resume link long-term")
- Documentation links ("Keep this FAQ link fresh")
- Long-lived references

**In Database**:
```sql
-- Fixed expiration
INSERT INTO urls (..., expires_at, has_fixed_expiration)
VALUES (..., '2027-12-31', true);

-- Activity-based (with future click-handler)
INSERT INTO urls (..., expires_at, has_fixed_expiration)
VALUES (..., '2027-01-07', false);
-- When clicked: UPDATE expires_at = NOW() + 12 months
```

---

### 5. Error Handling Strategy

**The Problem**: Errors need to travel through layers:
- Domain logic generates `DomainError`
- Application layer propagates it
- HTTP handler converts to JSON response

**Our Solution**: `DomainError` struct with type mapping

```go
// Domain layer defines error types
type ErrorType int
const (
    ErrorTypeNotFound     // → 404
    ErrorTypeValidation   // → 400
    ErrorTypeConflict     // → 409
    ErrorTypeInternal     // → 500
)

// Domain layer returns typed errors
func NewShortCode(s string) (ShortCode, error) {
    if !isValid(s) {
        return ShortCode{}, 
            NewValidationError("INVALID_SHORT_CODE", "Must be 4-20 alphanumeric")
    }
}

// HTTP layer converts to status code
func HandleDomainError(w http.ResponseWriter, err error) {
    var domainErr *url.DomainError
    if errors.As(err, &domainErr) {
        w.WriteHeader(domainErr.HTTPStatus()) // Auto-maps type → status
        json.NewEncoder(w).Encode(ErrorResponse{
            Error: ErrorDetail{
                Code:    domainErr.Code,    // "INVALID_SHORT_CODE"
                Message: domainErr.Message, // User-friendly message
            },
        })
    }
}
```

**Error Mapping Table**:

| Error Type | HTTP Status | When | Example |
|-----------|------------|------|---------|
| NOT_FOUND | 404 | URL doesn't exist | `curl /api/urls/invalid` |
| VALIDATION | 400 | Invalid input | URL without scheme, domain not whitelisted |
| CONFLICT | 409 | Duplicate data | Same short code twice (shouldn't happen) |
| INTERNAL | 500 | System error | Database crash, panic recovery |

**User-Friendly Response**:
```json
{
  "error": {
    "code": "INVALID_DOMAIN",
    "message": "Domain 'foo.com' is not allowed. Allowed domains: [short.link]"
  }
}
```

---

### 6. Value Objects for Validation

**The Problem**: Strings are everywhere. How do you know if a string is a valid short code?

**Our Solution**: Value Objects - immutable, validated-at-construction types

```go
// Instead of: func CreateURL(originalUrl string, domain string) { ... }
// We use:     func CreateURL(originalUrl OriginalURL, domain Domain) { ... }

// ✅ Creating a value object forces validation EARLY
shortCode, err := url.NewShortCode("abc")     // ❌ Too short, returns error
shortCode, err := url.NewShortCode("UkLWZg")  // ✅ Valid, returns ShortCode

// ✅ Type system ensures correctness
func CreateURL(original OriginalURL, domain Domain) {
    // Compiler guarantees these are valid - no need to re-validate
}

// ✅ Cannot be null or empty
type ShortCode struct {
    value string // Private: cannot be modified
}
```

**Example**:
```go
// Domain layer - value objects catch errors early
originalURL, err := url.NewOriginalURL("not-a-url")
// ❌ Error: "URL must include a scheme (http:// or https://)"

originalURL, err := url.NewOriginalURL("https://example.com")
// ✅ Valid, returns OriginalURL struct

// Application layer can trust the data
cmd := CreateURLCommand{
    OriginalURL: "https://example.com", // String
    Domain: "short.link",               // String
}

// Convert to value objects once at boundary
originalURL, _ := url.NewOriginalURL(cmd.OriginalURL)
domain, _ := url.NewDomain(cmd.Domain)

// From here on, all logic uses value objects (guaranteed valid)
urlEntity := url.NewURLWithActivityExpiration(id, shortCode, originalURL, domain)
```

---

### 7. Repository Pattern for Persistence

**The Problem**: Business logic shouldn't know about databases. But URLs need to be saved somewhere.

**Our Solution**: Repository interface in domain, implementation in infrastructure

```go
// Domain layer (NO database knowledge)
type Repository interface {
    NextID(ctx context.Context) (int64, error)
    Save(ctx context.Context, url *URL) error
    FindByShortCode(ctx context.Context, code ShortCode) (*URL, error)
}

// Infrastructure layer (implements domain interface)
type PostgresURLRepository struct {
    pool    *pgxpool.Pool
    queries *generated.Queries
}

func (r *PostgresURLRepository) Save(ctx context.Context, url *URL) error {
    // Actual PostgreSQL implementation
    return r.queries.CreateURL(ctx, params)
}

// Dependency injection
repo := postgres.NewPostgresURLRepository(dbPool)
handler := commands.NewCreateURLHandler(repo, generator, validator)
// Handler doesn't know it's PostgreSQL - just knows the interface
```

**Benefits**:
- Business logic never knows it's using PostgreSQL
- Can swap to MySQL, MongoDB later without changing domain code
- Easy to test with mock repository
- Framework-independent

---

### 8. Middleware Stack Order

**Why order matters**: Middleware is a nested chain. Outer middleware runs first!

```
┌─────────────────────────────────────────┐
│ 1. Recovery (outermost - catch panics)  │
├─────────────────────────────────────────┤
│ 2. Logger (log all requests)            │
├─────────────────────────────────────────┤
│ 3. CORS (add headers)                   │
├─────────────────────────────────────────┤
│ 4. Handler (innermost - actual logic)   │
└─────────────────────────────────────────┘

Execution flow:
Request → Recovery → Logger → CORS → Handler → Handler code
Response ← Recovery ← Logger ← CORS ← Handler ← Handler code
```

**Why this order**:
1. **Recovery first** (outermost) - catches panics from ALL inner middleware
2. **Logger second** - logs every request (even panics)
3. **CORS third** - modifies headers (nice to log)
4. **Handler last** - does actual work

---

## Key Files to Know

| File | Purpose | Importance |
|------|---------|-----------|
| `internal/domain/url/entity.go` | URL aggregate root, business methods | 🔥 Core logic |
| `internal/domain/url/errors.go` | Domain error types | 🔥 Error handling |
| `internal/domain/url/value_objects.go` | Validated value objects | 🔥 Data validation |
| `internal/application/commands/create_url.go` | Create URL command handler | 🔥 Use case |
| `internal/application/queries/get_url.go` | Get URL query handler | 🔥 Use case |
| `internal/infrastructure/persistence/postgres/url_repository.go` | Database implementation | 📊 Persistence |
| `internal/infrastructure/shortcode/sqids_generator.go` | Short code generation | 🔑 Key adapter |
| `internal/infrastructure/http/router/router.go` | Route setup | 🌐 HTTP setup |
| `cmd/api/main.go` | Dependency wiring | 🔌 Startup |

---

## Design Principles Used

1. **Single Responsibility**: Each layer has one job (domain = logic, infra = frameworks)
2. **Dependency Inversion**: Domain defines interfaces, infrastructure implements them
3. **Don't Repeat Yourself**: Value object validation in one place
4. **Fail Fast**: Validate at boundaries (value objects, not deep in logic)
5. **Framework Independence**: Business logic has zero framework dependencies

---

## How to Extend Without Breaking This Design

### Add a new endpoint:
1. Create command/query in `internal/application/`
2. Domain logic stays in `internal/domain/`
3. Infrastructure adapters in `internal/infrastructure/`
4. Handler in `internal/infrastructure/http/handlers/`
5. Route in `internal/infrastructure/http/router/`

### Add a new database query:
1. Write SQL in `internal/infrastructure/persistence/postgres/queries/urls.sql`
2. Run `just sqlc-generate`
3. Implement repository method in `url_repository.go`

### Add validation:
1. Create value object in `internal/domain/url/value_objects.go`
2. Use in domain entity or command handler

---

## Key Takeaways

- 🏛️ **DDD isolates business logic** from frameworks and databases
- 🔄 **CQRS separates commands (writes) from queries (reads)**
- 🔐 **Value objects validate data at boundaries** (fail fast)
- 🔌 **Interfaces decouple business from infrastructure**
- 📚 **Middleware stack order matters** (recovery → logger → cors → handler)
- 🎯 **Always go Domain → Application → Infrastructure** (dependency direction)
- ✅ **Everything is testable** without mocks or databases

---

*Last updated: Phase 18 complete (Jan 7, 2026)*

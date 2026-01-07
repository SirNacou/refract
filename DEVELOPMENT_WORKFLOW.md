# Development Workflow

**TL;DR**: Run `just up` to start, code in `internal/`, rebuild Docker on changes, use `just` commands for common tasks.

## Quick Start

```bash
# Start all services (PostgreSQL + migrations + API)
just up

# Verify it's running
curl http://localhost:8080/health

# View logs
just api-logs

# Stop services
just down
```

---

## Project Structure

```
services/api/
│
├── cmd/api/
│   └── main.go                     # 🚀 Entry point - wires dependencies
│
├── internal/
│   ├── config/
│   │   └── config.go               # Load env vars, validate config
│   │
│   ├── domain/url/                 # 🔐 BUSINESS LOGIC (don't touch from HTTP)
│   │   ├── entity.go               # URL aggregate root
│   │   ├── errors.go               # Domain error types
│   │   ├── value_objects.go        # ShortCode, OriginalURL, Domain
│   │   ├── repository.go           # Interface only (no impl)
│   │   ├── shortcode_generator.go  # Interface only (no impl)
│   │   └── domain_validator.go     # Interface only (no impl)
│   │
│   ├── application/                # 🔄 USE CASES (commands & queries)
│   │   ├── commands/
│   │   │   └── create_url.go       # Command for creating URLs
│   │   └── queries/
│   │       └── get_url.go          # Query for getting URL metadata
│   │
│   ├── infrastructure/             # 🔌 ADAPTERS (frameworks & libraries)
│   │   ├── persistence/postgres/   # Database implementation
│   │   │   ├── url_repository.go   # Implements Repository interface
│   │   │   ├── queries/
│   │   │   │   └── urls.sql        # SQLc queries
│   │   │   └── generated/          # SQLc generated code (don't edit)
│   │   ├── shortcode/
│   │   │   └── sqids_generator.go  # Implements ShortCodeGenerator
│   │   ├── validation/
│   │   │   └── domain_validator.go # Implements DomainValidator
│   │   └── http/                   # HTTP layer
│   │       ├── dto/                # Request/response data classes
│   │       ├── middleware/         # HTTP middleware
│   │       ├── handlers/           # HTTP request handlers
│   │       └── router/             # Route setup
│   │
│   ├── database/
│   │   └── postgres.go             # Database connection (legacy, keep for now)
│   │
│   └── models/
│       └── url.go                  # Legacy model (not used, can remove)
│
├── go.mod / go.sum                 # Dependencies
├── sqlc.yaml                       # SQLc configuration
├── Dockerfile                      # Container build
└── .dockerignore
```

---

## All Available Commands

```bash
# ⭐ Services
just up                    # Start all services
just down                  # Stop all services
just rebuild               # Rebuild images and restart
just clean                 # DELETE everything (⚠️ DESTRUCTIVE)
just status                # Show running containers

# 📋 Logs
just logs                  # View all logs (all services)
just api-logs              # View only API logs
just migration-logs        # View migration logs
just postgres-logs         # View PostgreSQL logs
just tail-api              # Real-time API logs (last 100 lines)

# 🗄️ Database
just db-shell              # Connect to PostgreSQL
just migrate-status        # Show migration history
just migrate-up            # Run migrations manually (usually automatic)
just migrate-create NAME   # Create new migration file

# 🛠️ Development
just sqlc-generate         # Regenerate SQLc code from SQL
just fmt                   # Format all Go code
just test                  # Run tests (when available)
just lint                  # Run linter (requires golangci-lint)

# 🚀 Utilities
just install-deps          # Install dev tools (goose, etc.)
just restart-api           # Restart API service only
```

---

## Common Tasks

### 1. Create a New Endpoint

**Example**: Add `DELETE /api/urls/{shortCode}` to deactivate URLs

#### Step 1: Create a command in `internal/application/commands/`

Create `internal/application/commands/deactivate_url.go`:
```go
package commands

import (
    "context"
    "github.com/SirNacou/refract/services/api/internal/domain/url"
)

type DeactivateURLCommand struct {
    ShortCode string
}

type DeactivateURLHandler struct {
    repo url.Repository
}

func NewDeactivateURLHandler(repo url.Repository) *DeactivateURLHandler {
    return &DeactivateURLHandler{repo: repo}
}

func (h *DeactivateURLHandler) Handle(ctx context.Context, cmd DeactivateURLCommand) error {
    // Validate
    shortCode, err := url.NewShortCode(cmd.ShortCode)
    if err != nil {
        return err
    }
    
    // Find URL
    urlEntity, err := h.repo.FindByShortCode(ctx, shortCode)
    if err != nil {
        return err
    }
    
    // Business logic: deactivate
    urlEntity.Deactivate()
    
    // Persist
    return h.repo.Save(ctx, urlEntity)
}
```

#### Step 2: Add repository method in `internal/infrastructure/persistence/postgres/url_repository.go`

(The `Save` method already handles updates, so just use that)

#### Step 3: Create handler in `internal/infrastructure/http/handlers/url_handler.go`

```go
func (h *URLHandler) DeactivateURL(w http.ResponseWriter, r *http.Request) {
    shortCode := r.PathValue("shortCode")
    if shortCode == "" {
        writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{...})
        return
    }
    
    cmd := commands.DeactivateURLCommand{ShortCode: shortCode}
    err := h.deactivateHandler.Handle(r.Context(), cmd)
    if err != nil {
        middleware.HandleDomainError(w, err, h.logger)
        return
    }
    
    writeJSON(w, http.StatusNoContent, nil)
}
```

#### Step 4: Wire in `cmd/api/main.go`

```go
// Add to handler creation section:
deactivateHandler := commands.NewDeactivateURLHandler(repo)

// Update URLHandler:
urlHandler := handlers.NewURLHandler(
    createURLHandler, 
    getURLHandler,
    deactivateHandler,  // NEW
    logger,
)
```

#### Step 5: Add route in `internal/infrastructure/http/router/router.go`

```go
mux.HandleFunc("DELETE /api/urls/{shortCode}", cfg.URLHandler.DeactivateURL)
```

#### Step 6: Rebuild and test

```bash
docker compose down
docker compose up -d --build

# Test
curl -X DELETE http://localhost:8080/api/urls/UkLWZg
```

---

### 2. Add a New Database Query

**Example**: Add ability to list all URLs created by domain

#### Step 1: Add SQL query to `internal/infrastructure/persistence/postgres/queries/urls.sql`

```sql
-- name: ListURLsByDomain :many
SELECT 
    id,
    short_code,
    original_url,
    domain,
    created_at,
    updated_at,
    expires_at,
    has_fixed_expiration,
    click_count,
    is_active,
    metadata
FROM urls
WHERE domain = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
```

#### Step 2: Regenerate SQLc code

```bash
just sqlc-generate
```

This creates functions in `internal/infrastructure/persistence/postgres/generated/urls.sql.go`

#### Step 3: Add repository method in `url_repository.go`

```go
func (r *PostgresURLRepository) ListByDomain(
    ctx context.Context,
    domain url.Domain,
    limit int,
    offset int,
) ([]*url.URL, error) {
    rows, err := r.queries.ListURLsByDomain(ctx, generated.ListURLsByDomainParams{
        Domain: domain.String(),
        Limit:  int32(limit),
        Offset: int32(offset),
    })
    if err != nil {
        return nil, url.NewInternalError("DB_ERROR", "Failed to list URLs", err)
    }
    
    urls := make([]*url.URL, 0, len(rows))
    for _, row := range rows {
        urlEntity, err := rowToEntity(row)
        if err != nil {
            return nil, err
        }
        urls = append(urls, urlEntity)
    }
    return urls, nil
}
```

#### Step 4: Use in query handler or command

Create `internal/application/queries/list_urls_by_domain.go` or add to existing handler.

---

### 3. Add Validation for a New Field

**Example**: Validate that short codes can't use profanity

#### Step 1: Create a value object in `internal/domain/url/value_objects.go`

Update `NewShortCode()`:
```go
func NewShortCode(s string) (ShortCode, error) {
    s = strings.TrimSpace(s)
    
    if s == "" {
        return ShortCode{}, NewValidationError("INVALID_SHORT_CODE", "Short code cannot be empty")
    }
    
    if !shortCodeRegex.MatchString(s) {
        return ShortCode{}, NewValidationError("INVALID_SHORT_CODE", "Short code must be 4-20 alphanumeric")
    }
    
    // NEW: Check profanity list
    if isProfanity(s) {
        return ShortCode{}, NewValidationError("INVALID_SHORT_CODE", "Short code contains profanity")
    }
    
    return ShortCode{value: s}, nil
}

func isProfanity(code string) bool {
    profanityList := []string{"badword1", "badword2"}
    for _, word := range profanityList {
        if strings.EqualFold(code, word) {
            return true
        }
    }
    return false
}
```

#### Step 2: That's it! The validation now happens everywhere short codes are created

The value object constructor is called in:
- `CreateURLHandler.Handle()` when generating codes
- `GetURLHandler.Handle()` when parsing request
- `PostgresURLRepository.FindByShortCode()` when querying

---

### 4. View Logs

```bash
# Real-time all logs
just logs

# Just API
just api-logs

# Last 30 lines
just api-logs | tail -30

# Search for errors
just api-logs | grep ERROR

# JSON parsing (if using jq)
just api-logs | jq 'select(.level == "ERROR")'
```

---

### 5. Connect to Database

```bash
just db-shell

# Once in psql:
\d urls                    # Describe table
SELECT * FROM urls;        # View all rows
SELECT * FROM goose_db_version;  # View migrations
\q                        # Exit
```

---

### 6. Run Migrations Manually

```bash
# Status
just migrate-status

# Run any pending migrations
just migrate-up

# Create new migration
just migrate-create add_user_id_to_urls
# Edit: migrations/postgres/00003_add_user_id_to_urls.sql
# Restart services for auto-run
just rebuild
```

---

### 7. Code Changes Workflow

```bash
# 1. Make code changes in internal/
# 2. Test locally
curl http://localhost:8080/api/urls -d '...'

# 3. If go.mod changed or Dockerfile changed:
just rebuild

# 4. If only .go files changed:
docker compose restart api  # Rebuilds if needed

# 5. View logs
just api-logs
```

---

### 8. Add a New Error Type

In `internal/domain/url/errors.go`:

```go
// Add to ErrorType enum
const (
    ErrorTypeNotFound ErrorType = iota
    ErrorTypeValidation
    ErrorTypeConflict
    ErrorTypeInternal
    ErrorTypeRateLimited  // NEW
)

// Add to HTTPStatus() method
func (e *DomainError) HTTPStatus() int {
    switch e.Type {
    case ErrorTypeNotFound:
        return 404
    case ErrorTypeValidation:
        return 400
    case ErrorTypeConflict:
        return 409
    case ErrorTypeInternal:
        return 500
    case ErrorTypeRateLimited:  // NEW
        return 429
    default:
        return 500
    }
}

// Add constructor
func NewRateLimitError(code, message string) *DomainError {
    return &DomainError{
        Type:    ErrorTypeRateLimited,
        Code:    code,
        Message: message,
    }
}
```

Now HTTP handlers automatically get correct status codes:
```go
if tooManyRequests {
    return url.NewRateLimitError("TOO_MANY_REQUESTS", "Rate limit exceeded")
}
// Automatically responds with 429 status
```

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Port 8080 already in use | `lsof -i :8080` to find process, `just down` to stop |
| Database won't connect | `just db-shell` to test, check Docker logs `just postgres-logs` |
| SQLc code out of date | Run `just sqlc-generate` after changing SQL |
| Import errors in IDE | Restart IDE / `go mod tidy` |
| Container won't build | `docker compose down -v` to remove volumes, `just rebuild` |
| Migrations didn't run | Check `just migrate-status`, verify SQL syntax, restart with `just rebuild` |

---

## Build/Run Cycle

```
┌──────────────────────────────────┐
│  Edit Go files (.go)             │
├──────────────────────────────────┤
│  ✅ Change: just restart-api     │
│  ❌ No restart needed             │
│     (depends on running container)│
└──────────────────────────────────┘

┌──────────────────────────────────┐
│  Edit SQL queries                 │
├──────────────────────────────────┤
│  1. Run: just sqlc-generate      │
│  2. Edit repository method       │
│  3. Run: just rebuild            │
└──────────────────────────────────┘

┌──────────────────────────────────┐
│  Edit Dockerfile or go.mod       │
├──────────────────────────────────┤
│  Run: just rebuild               │
│  (rebuilds image and restarts)   │
└──────────────────────────────────┘
```

---

## Key Takeaways

- ✅ **Always think DDD**: Domain logic → Application logic → HTTP handlers
- ✅ **Use value objects** for validation at boundaries
- ✅ **Repository pattern** for database access (interface-based)
- ✅ **SQLc for type safety**: Edit SQL, regenerate, use generated functions
- ✅ **Middleware stack** is important (recovery → logger → cors → handler)
- ✅ **Error types map to HTTP status** automatically
- ✅ **Tests don't need Docker** (use domain layer directly)

---

*Last updated: Phase 18 complete (Jan 7, 2026)*

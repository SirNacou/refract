# Snowflake ID Generator - Infrastructure Layer

Infrastructure wrapper for the Snowflake ID generator, providing environment-based configuration and dependency injection support.

## Overview

This package wraps the domain-layer Snowflake generator (`domain/url`) with infrastructure concerns:
- Environment variable configuration (`WORKER_ID`)
- Dependency injection interface (`IDGenerator`)
- Error handling and validation
- Production-ready defaults

## Usage

### Basic Usage

```go
import "github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"

// Create generator with explicit worker ID (testing)
gen, err := idgen.NewSnowflakeGeneratorWithWorkerID(1)
if err != nil {
    log.Fatal(err)
}

id, err := gen.NextID()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Generated ID:", id)
```

### Production Usage (Environment-based)

```go
// Set WORKER_ID environment variable:
// export WORKER_ID=42

gen, err := idgen.NewSnowflakeGenerator()
if err != nil {
    log.Fatal(err)
}

id, err := gen.NextID()
```

### Dependency Injection

```go
type URLService struct {
    idGen idgen.IDGenerator
}

func NewURLService(idGen idgen.IDGenerator) *URLService {
    return &URLService{idGen: idGen}
}

func (s *URLService) CreateURL() error {
    id, err := s.idGen.NextID()
    // ...
}
```

## Worker ID Assignment

Following the architecture specification (FR-008):

| Service Type | Worker ID Range | Example |
|--------------|-----------------|---------|
| API Service  | 0-63            | API instance 1: `WORKER_ID=0`, API instance 2: `WORKER_ID=1` |
| Redirector   | 64-127          | Redirector instance 1: `WORKER_ID=64` |
| Reserved     | 128-1023        | Future services |

### Environment Configuration

**Development (local):**
```bash
# .env file
WORKER_ID=0
```

**Production (Docker Compose):**
```yaml
services:
  api-1:
    environment:
      - WORKER_ID=0
  api-2:
    environment:
      - WORKER_ID=1
```

**Production (Kubernetes):**
```yaml
env:
  - name: WORKER_ID
    value: "0"
```

## Configuration

| Environment Variable | Default | Valid Range | Description |
|---------------------|---------|-------------|-------------|
| `WORKER_ID`         | `0`     | 0-1023      | Unique worker identifier for distributed ID generation |

### Validation Rules

- Worker ID must be numeric (int64)
- Must be between 0 and 1023 (inclusive)
- Invalid values return error on generator creation
- Missing `WORKER_ID` defaults to 0 (safe for local development)

## Performance

Based on benchmarks:

- **Throughput**: ~3.3M IDs/second per worker
- **Latency**: ~300ns per ID generation
- **Memory**: 0 allocations per ID
- **Concurrency**: Thread-safe with mutex (300ns overhead)

## Architecture

```
Application Layer
    ↓
IDGenerator Interface (this package)
    ↓
SnowflakeGenerator Wrapper (this package)
    ↓
domain/url.SnowflakeGenerator (domain logic)
```

### Why This Wrapper Exists

1. **Separation of Concerns**: Domain layer doesn't know about environment variables
2. **Configuration**: Centralizes infrastructure configuration
3. **Testability**: Easy to mock `IDGenerator` interface
4. **Dependency Injection**: Compatible with DI frameworks

## Testing

```bash
# Run all tests
go test ./internal/infrastructure/idgen

# Run with coverage
go test -cover ./internal/infrastructure/idgen

# Run benchmarks
go test -bench=. ./internal/infrastructure/idgen
```

## Error Handling

```go
gen, err := idgen.NewSnowflakeGenerator()
if err != nil {
    // Possible errors:
    // - Invalid WORKER_ID environment variable
    // - Worker ID out of range (0-1023)
    log.Fatalf("Failed to create ID generator: %v", err)
}

id, err := gen.NextID()
if err != nil {
    // Possible errors:
    // - Clock moved backwards (rare)
    // - Sequence exhausted (handled internally with sleep)
    log.Fatalf("Failed to generate ID: %v", err)
}
```

## Integration Example

```go
package main

import (
    "github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
)

func main() {
    // Initialize ID generator on application startup
    idGenerator, err := idgen.NewSnowflakeGenerator()
    if err != nil {
        log.Fatal(err)
    }

    // Inject into services
    urlService := NewURLService(idGenerator)
    
    // Use throughout application
    urlService.CreateShortURL(...)
}
```

## References

- Domain implementation: `internal/domain/url/snowflake.go`
- Specification: `specs/016-authenticated-url-shortener/research.md` (lines 73-170)
- Functional requirement: FR-008 (Snowflake ID generation)

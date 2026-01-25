# Backend OIDC Migration Tasks

**Status:** Complete  
**Priority:** Medium  
**Estimated Effort:** 2-3 days  

**Goal:** Convert `services/api` from Zitadel-specific auth to generic OIDC with strict audience validation. This enables switching identity providers (Keycloak, Auth0, Okta) without code changes.

## Implementation

This migration uses the **[coreos/go-oidc](https://github.com/coreos/go-oidc)** library, which is:
- Well-maintained and widely used (Kubernetes, Dex, etc.)
- Handles OIDC discovery, JWKS fetching, and caching automatically
- Battle-tested and security-audited

## Current State

The codebase now uses generic OIDC:
- `services/api/internal/config/config.go` - `OIDCConfig` struct with `OIDC_*` env vars (backward compat with `ZITADEL_*`)
- `services/api/internal/infrastructure/auth/oidc_verifier.go` - Thin wrapper around coreos/go-oidc
- `services/api/internal/infrastructure/server/middleware/auth.go` - middleware using OIDCVerifier
- `services/api/internal/infrastructure/server/dto/health.go` - health check with `OIDC` dependency name

## Target Configuration

Required environment variables:
- `OIDC_ISSUER`: OIDC provider URL (e.g., `https://auth.example.com`)
- `OIDC_AUDIENCE`: Expected `aud` claim in JWTs (e.g., `refract-api`)

---

## Phase 1: Configuration & Environment

**Effort:** ~2 hours | **Risk:** Low | **Status:** COMPLETE

### 1.1 Update Config Schema
**File:** `services/api/internal/config/config.go`
- [x] Add `OIDCConfig` struct with:
  - `Issuer string` (required, URL)
  - `Audience string` (required)
  - `JWKSCacheTTL time.Duration` (optional, default 10m)
  - `ClockSkewSeconds int` (optional, default 60)
- [x] Add env var mapping:
  - `OIDC_ISSUER` → `Issuer`
  - `OIDC_AUDIENCE` → `Audience`
  - `OIDC_JWKS_CACHE_TTL` → `JWKSCacheTTL`
  - `OIDC_CLOCK_SKEW_SECONDS` → `ClockSkewSeconds`
- [x] Add backward compatibility mapping:
  - If `ZITADEL_ISSUER` exists and `OIDC_ISSUER` missing, map with deprecation warning
  - If `ZITADEL_URL` exists and `OIDC_ISSUER` missing, map with deprecation warning

### 1.2 Update Config Tests
**File:** `services/api/internal/config/config_test.go`
- [x] Add tests for new `OIDC_*` env vars
- [x] Add tests for backward compatibility mapping
- [x] Update existing `ZITADEL_*` tests to use new names

### 1.3 Update Environment Documentation
- [x] Update `.env.example` with new OIDC vars
- [x] Document migration from ZITADEL_* to OIDC_*

---

## Phase 2: OIDC Verifier Implementation

**Effort:** ~4-6 hours | **Risk:** Medium (core security component) | **Status:** COMPLETE

### 2.1 Create Generic OIDC Verifier
**File:** `services/api/internal/infrastructure/auth/oidc_verifier.go`
- [x] Implement `OIDCVerifier` struct with:
  - Issuer URL
  - Expected audience
  - JWKS cache with TTL
  - HTTP client for discovery/JWKS fetching
- [x] Implement discovery client:
  - `GET {issuer}/.well-known/openid-configuration`
  - Parse and cache `jwks_uri`
- [x] Implement JWKS fetching:
  - Fetch keys from `jwks_uri`
  - Cache in memory with TTL
  - Support key rotation (refresh on unknown `kid`)
- [x] Implement JWT verification:
  - Validate signature using JWKS
  - Validate `iss` == configured issuer
  - Validate `aud` contains expected audience (strict)
  - Validate `exp` not expired (allow clock skew)
  - Restrict allowed algorithms (RS256, RS384, RS512)
- [x] Extract claims:
  - `sub` (required)
  - `email` (optional, if present)

### 2.2 Add OIDC Verifier Tests
**File:** `services/api/internal/infrastructure/auth/oidc_verifier_test.go`
- [x] Test discovery document parsing
- [x] Test JWKS fetching and caching
- [x] Test JWT signature validation
- [x] Test issuer mismatch rejection
- [x] Test audience mismatch rejection (strict)
- [x] Test expired token rejection
- [x] Test unknown `kid` triggers JWKS refresh

---

## Phase 3: Middleware Migration

**Effort:** ~2 hours | **Risk:** Medium (affects all authenticated routes) | **Status:** COMPLETE

### 3.1 Update Auth Middleware
**File:** `services/api/internal/infrastructure/server/middleware/auth.go`
- [x] Remove Zitadel SDK imports
- [x] Replace `NewAuthMiddleware(zitadel *auth.ZitadelProvider)` with `NewAuthMiddleware(verifier *auth.OIDCVerifier)`
- [x] Update middleware logic:
  - Extract `Authorization: Bearer <token>` header
  - Verify token using OIDC verifier
  - Extract `sub` and optional `email` claims
  - Set context values:
    - `TokenTypeKey` → "jwt" (keep existing)
    - `UserIDKey` → `sub` claim
    - `UserEmailKey` → `email` claim (if present)
- [x] Update error handling to use consistent error format

### 3.2 Add Auth Middleware Tests
**File:** `services/api/internal/infrastructure/server/middleware/auth_test.go`
- [x] Test valid token acceptance
- [x] Test missing Authorization header rejection
- [x] Test invalid token rejection
- [x] Test expired token rejection
- [x] Test audience mismatch rejection
- [x] Test issuer mismatch rejection

---

## Phase 4: Application Bootstrap

**Effort:** ~1 hour | **Risk:** Low | **Status:** COMPLETE

### 4.1 Update Main Application
**File:** `services/api/cmd/api/main.go`
- [x] Replace `auth.NewZitadelProvider()` with `auth.NewOIDCVerifier()`
- [x] Update middleware construction:
  - `authMiddleware := middleware.NewAuthMiddleware(verifier)`
- [x] Update health check dependency name

### 4.2 Update Health Check DTO
**File:** `services/api/internal/infrastructure/server/dto/health.go`
- [x] Rename `Zitadel DependencyStatus` to `OIDC DependencyStatus`
- [x] Update any references in health check logic

---

## Phase 5: Cleanup & Documentation

**Effort:** ~2 hours | **Risk:** Low | **Status:** COMPLETE

### 5.1 Remove Zitadel Dependencies
- [x] Remove `github.com/zitadel/zitadel-go/v3` from `go.mod`
- [x] Delete `services/api/internal/infrastructure/auth/zitadel_provider.go`
- [x] Remove any remaining Zitadel imports

### 5.2 Update Documentation
- [ ] Update API documentation to reference generic OIDC
- [ ] Add setup instructions for different providers:
  - Zitadel (current default)
  - Keycloak
  - Auth0
  - Okta
- [ ] Document required audience configuration for each provider

### 5.3 Update Docker/Deployment Config
- [ ] Update any docker-compose or deployment configs to use new env var names
- [ ] Document migration steps for existing deployments

---

## Phase 6: Validation & Testing

**Effort:** ~3-4 hours | **Risk:** High (must catch security issues) | **Status:** PENDING

### 6.1 Integration Tests
- [ ] Test end-to-end with Zitadel provider
- [ ] Test end-to-end with Keycloak (if possible)
- [ ] Test token refresh scenarios
- [ ] Test concurrent requests with JWKS cache

### 6.2 Security Validation
- [ ] Verify strict audience checking works
- [ ] Test token replay protection
- [ ] Validate algorithm restrictions
- [ ] Test clock skew handling

---

## Migration Notes

### Audience Configuration
**Recommended:** Set `OIDC_AUDIENCE=refract-api`

**Zitadel Setup:**
- Configure Zitadel to issue access tokens with `aud: "refract-api"`
- This may require setting up a resource/API in Zitadel console

**Alternative Providers:**
- Keycloak: Configure client with `audience` or use client ID as audience
- Auth0: Configure API identifier
- Okta: Configure resource server

### Backward Compatibility
- Temporary mapping from `ZITADEL_*` to `OIDC_*` with deprecation warnings
- Remove mapping in future version once migration complete

### Security Considerations
- Always validate `iss` and `aud` claims
- Use HTTPS for all OIDC endpoints
- Implement proper JWKS caching with reasonable TTL
- Restrict allowed JWT algorithms
- Consider rate limiting for discovery/JWKS endpoints

---

## Rollback Plan

If migration fails:
1. Revert to previous Zitadel-specific implementation
2. Restore `ZITADEL_*` environment variables
3. Re-add Zitadel SDK dependency
4. Test with existing Zitadel configuration

Document rollback steps in deployment documentation.

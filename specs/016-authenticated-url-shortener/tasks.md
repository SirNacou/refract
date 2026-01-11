# Tasks: Authenticated URL Shortener Platform

**Feature Branch**: `016-authenticated-url-shortener`
**Input**: Design documents from `/specs/016-authenticated-url-shortener/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/, research.md, quickstart.md

**Organization**: Tasks are grouped by user story (P1, P2, P3, P4) to enable independent implementation and testing of each story.

**Tests**: Tests included in Phase 8 (T176-T180). Per constitution Principle IV, contract/integration tests MUST pass before production deployment. Tests may be deferred during rapid MVP iteration but are REQUIRED for production readiness.

---

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1-US5 from spec.md) this task belongs to
- **File paths**: Exact locations per plan.md structure

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create distributed system structure with 3 services + frontend + infrastructure

- [X] T001 Create repository root structure: services/, frontend/, migrations/, docker-compose.yml, justfile
- [X] T002 [P] Initialize API service (Go): services/api/ with go.mod, cmd/api/main.go, internal/ structure
- [X] T003 [P] Initialize Redirector service (Rust): services/redirector/ with Cargo.toml, src/main.rs
- [X] T004 [P] Initialize Analytics Processor (Go): services/analytics-processor/ with go.mod, cmd/processor/main.go
- [X] T005 [P] Initialize Frontend (TanStack Start): frontend/ with package.json, app/ structure, tsconfig.json
- [X] T006 [P] Create migrations directory: migrations/postgres/ with .gitkeep
- [X] T007 [P] Create Docker Compose configuration in docker-compose.yml with PostgreSQL, Redis, Zitadel services
- [X] T008 [P] Create Justfile with commands: `just up`, `just down`, `just migrate`, `just test`
- [X] T009 [P] Create .env.example with all required environment variables per quickstart.md
- [X] T010 [P] Create .gitignore files for each service (Go, Rust, TypeScript, Docker volumes)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

**Constitution Alignment**: Implements Principles I-III (Domain-Driven Architecture, CQRS, Infrastructure Adapters)

### API Service Foundation

- [X] T011 Create domain layer structure: services/api/internal/domain/url/, domain/apikey/, domain/user/
- [X] T012 Create application layer structure: services/api/internal/application/commands/, application/queries/
- [X] T013 Create infrastructure layer structure: services/api/internal/infrastructure/persistence/postgres/, infrastructure/cache/, infrastructure/auth/, infrastructure/idgen/, infrastructure/server/
- [X] T014 Create Snowflake ID generator in services/api/internal/infrastructure/idgen/snowflake_generator.go (FR-008: 64-bit distributed IDs with worker coordination, timestamp+sequence generation)
- [X] T015 Add Base62 encoder/decoder functions to services/api/internal/infrastructure/idgen/snowflake_generator.go (FR-008: EncodeBase62, DecodeBase62 for URL-safe short codes)
- [X] T016 [P] Create config management in services/api/internal/config/config.go (database, Redis, Zitadel URLs, worker ID)
- [X] T017 [P] Create database connection pooling setup in services/api/internal/infrastructure/persistence/postgres/connection.go
- [X] T018 [P] Create Redis client setup in services/api/internal/infrastructure/cache/redis_cache.go (L2 cache interface)
- [X] T019 [P] Create Zitadel OIDC provider in services/api/internal/infrastructure/auth/zitadel_provider.go (JWT validation, FR-002)
- [X] T020 [P] Create HTTP router setup with Chi in services/api/internal/infrastructure/server/router.go
- [X] T021 [P] Create authentication middleware in services/api/internal/infrastructure/server/middleware/auth.go (JWT + API key validation)
- [X] T022 [P] Create rate limiting middleware in services/api/internal/infrastructure/server/middleware/rate_limit.go (FR-040, FR-041)
- [X] T023 [P] Configure CORS middleware using go-chi/cors in services/api/internal/infrastructure/server/middleware/cors.go (use SecurityConfig CORS settings)
- [X] T024 [P] Create logging middleware in services/api/internal/infrastructure/server/middleware/logging.go (structured JSON logs, Principle V)
- [X] T025 [P] Create error handling utilities in services/api/internal/infrastructure/server/errors.go
- [X] T026 [P] Create DTO structures in services/api/internal/infrastructure/server/dto/ (request/response models per OpenAPI spec)

### Database Migrations

- [ ] T027 Create migration 00001_create_users.sql in migrations/postgres/ (users table per data-model.md)
- [ ] T028 Create migration 00002_create_urls.sql in migrations/postgres/ (urls table with indexes per data-model.md)
- [ ] T029 Create migration 00003_create_api_keys.sql in migrations/postgres/ (api_keys table per data-model.md)
- [ ] T030 Create migration 00004_create_timescale_hypertables.sql in migrations/postgres/ (click_events hypertable, continuous aggregates per data-model.md)
- [ ] T031 Add migration runner to services/api/cmd/api/main.go (check/run migrations on startup)

### Redirector Service Foundation

- [ ] T032 Create Axum server setup in services/redirector/src/main.rs with graceful shutdown
- [ ] T033 [P] Create config management in services/redirector/src/config.rs (database, Redis, worker ID, L1 cache size)
- [ ] T034 [P] Create PostgreSQL connection pool in services/redirector/src/repository/postgres.rs using SQLx
- [ ] T035 [P] Create L1 in-memory LRU cache in services/redirector/src/cache/memory.rs (10K capacity, FR-017)
- [ ] T036 [P] Create L2 Redis cache client in services/redirector/src/cache/redis.rs (FR-017)
- [ ] T037 [P] Create Redis Stream publisher in services/redirector/src/events/publisher.rs (click events to analytics)
- [ ] T038 [P] Create GeoIP lookup utilities in services/redirector/src/geo/ (MaxMind GeoLite2 integration, FR-021)
- [ ] T039 [P] Create user agent parser in services/redirector/src/parser/user_agent.rs (device type, browser, OS, FR-022)

### Analytics Processor Foundation

- [ ] T040 Create Redis Stream consumer in services/analytics-processor/internal/consumer/click_events.go (consumer group, batch processing)
- [ ] T041 [P] Create TimescaleDB repository in services/analytics-processor/internal/repository/timescale_repository.go (batch insert click events)
- [ ] T042 [P] Create GeoIP lookup in services/analytics-processor/internal/geo/maxmind.go (same as redirector for consistency)
- [ ] T043 [P] Create config management in services/analytics-processor/internal/config/config.go

### Frontend Foundation

- [ ] T044 Setup TanStack Start project structure in frontend/app/ with routes/, components/, utils/
- [ ] T045 [P] Create API client wrapper in frontend/app/utils/api-client.ts (fetch with auth headers, error handling)
- [ ] T046 [P] Create Zitadel OIDC integration in frontend/app/utils/auth.ts (login, logout, token refresh)
- [ ] T047 [P] Create AuthGuard component in frontend/app/components/AuthGuard.tsx (protected route wrapper)
- [ ] T048 [P] Install and configure TanStack Query in frontend/app/root.tsx (QueryClientProvider)
- [ ] T049 [P] Install and configure Recharts for analytics visualizations

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 5 - User Authentication (Priority: P1) 🎯 MVP Prerequisite

**Goal**: Implement Zitadel authentication so users can sign up, log in, and access protected pages

**Independent Test**: Visitor can sign up via Zitadel, log in with JWT token, access dashboard, log out (per spec.md scenarios 1-5)

**Why First**: Authentication is a prerequisite for US1 (URL creation requires login). Must be complete before any authenticated features.

### Domain Layer (US5)

- [ ] T050 [P] [US5] Create User entity in services/api/internal/domain/user/entity.go (Zitadel subject ID, email, status)
- [ ] T051 [P] [US5] Create UserRepository interface in services/api/internal/domain/user/repository.go (CreateOrUpdate, GetByZitadelSub)
- [ ] T052 [P] [US5] Create domain errors in services/api/internal/domain/user/errors.go (UserNotFound, UserSuspended)

### Application Layer (US5)

- [ ] T053 [P] [US5] Create SyncUser command in services/api/internal/application/commands/sync_user.go (upsert user from JWT claims, FR-003)
- [ ] T054 [P] [US5] Create GetUserProfile query in services/api/internal/application/queries/get_user_profile.go

### Infrastructure Layer (US5)

- [ ] T055 [US5] Implement PostgresUserRepository in services/api/internal/infrastructure/persistence/postgres/user_repository.go (implements domain interface)

### HTTP Layer (US5)

- [ ] T056 [P] [US5] Create /api/v1/auth/callback handler in services/api/internal/infrastructure/server/handlers/auth.go (exchange code for tokens, sync user)
- [ ] T057 [P] [US5] Create /api/v1/auth/logout handler in services/api/internal/infrastructure/server/handlers/auth.go (invalidate Zitadel session, FR-005)
- [ ] T058 [P] [US5] Create /api/v1/users/me handler (get current user profile)

### Frontend (US5)

- [ ] T059 [P] [US5] Create homepage in frontend/app/routes/index.tsx (public landing page with Sign Up / Log In buttons)
- [ ] T060 [P] [US5] Create auth callback route in frontend/app/routes/auth/callback.tsx (handle Zitadel redirect, store tokens)
- [ ] T061 [P] [US5] Create logout route in frontend/app/routes/auth/logout.tsx (clear tokens, redirect to homepage)
- [ ] T062 [US5] Implement login flow in frontend/app/utils/auth.ts (redirect to Zitadel, handle PKCE, FR-004)
- [ ] T063 [US5] Create dashboard route in frontend/app/routes/dashboard.tsx (protected, shows "Welcome {email}" placeholder)

**Checkpoint**: Authentication complete - users can sign up, log in, log out, access protected dashboard

---

## Phase 4: User Story 1 - Create and Share Short URLs (Priority: P1) 🎯 MVP Core

**Goal**: Logged-in users can create short URLs (auto-generated or custom alias) and anyone can visit them to redirect

**Independent Test**: User logs in, creates short URL, receives working short link, anyone visiting short link redirects to destination (per spec.md scenarios 1-5)

**Why This Priority**: This is the core value proposition - without URL shortening, there is no product.

### Domain Layer (US1)

- [ ] T064 [P] [US1] Create URL aggregate root in services/api/internal/domain/url/entity.go (Snowflake ID, custom alias, destination URL, status, expiration, metadata)
- [ ] T065 [P] [US1] Create ShortCode value object in services/api/internal/domain/url/short_code.go (Base62 encode/decode, validation)
- [ ] T066 [P] [US1] Create URLRepository interface in services/api/internal/domain/url/repository.go (Create, GetBySnowflakeID, GetByCustomAlias, GetByCreatorID)
- [ ] T067 [P] [US1] Create domain errors in services/api/internal/domain/url/errors.go (URLNotFound, AliasAlreadyTaken, InvalidURL, MaliciousURL)
- [ ] T068 [US1] Add business rules to URL entity: ValidateDestinationURL, ValidateCustomAlias (3-50 chars, alphanumeric+hyphens, no reserved words per data-model.md Appendix A)

### Application Layer (US1)

- [ ] T069 [P] [US1] Create CreateShortURL command in services/api/internal/application/commands/create_url.go (generate Snowflake ID, validate URL, check Safe Browsing API, FR-012)
- [ ] T070 [P] [US1] Create GetURLByShortCode query in services/api/internal/application/queries/get_url.go (decode Base62 or lookup custom alias)

### Infrastructure Layer (US1)

- [ ] T071 [US1] Implement PostgresURLRepository in services/api/internal/infrastructure/persistence/postgres/url_repository.go (implements domain interface, uses prepared statements)
- [ ] T072 [P] [US1] Create Safe Browsing API client in services/api/internal/infrastructure/safebrowsing/client.go (validate URLs against malicious domains, FR-042)
- [ ] T073 [P] [US1] Add cache warming logic to PostgresURLRepository (insert into Redis L2 on create)

### HTTP Layer (US1)

- [ ] T074 [US1] Create POST /api/v1/urls handler in services/api/internal/infrastructure/server/handlers/urls.go (create short URL, requires JWT auth)
- [ ] T075 [P] [US1] Add request validation to POST /api/v1/urls (validate DTO against OpenAPI schema)
- [ ] T076 [P] [US1] Add response mapping (domain URL → DTO URLResponse per api-service.openapi.yaml)

### Redirector Service (US1)

- [ ] T077 [US1] Implement GET /:shortCode handler in services/redirector/src/handlers/redirect.rs (lookup URL, return 301 redirect, FR-015, FR-016)
- [ ] T078 [US1] Add multi-tier cache lookup to redirect handler: L1 → L2 → DB fallback (FR-017)
- [ ] T079 [US1] Add HTTP status logic: 301 (active), 404 (not found), 410 (expired/disabled), FR-018, FR-019
- [ ] T080 [US1] Generate HTML error pages for 404/410 in services/redirector/src/handlers/error_pages.rs
- [ ] T081 [US1] Publish click event to Redis Stream asynchronously (non-blocking, FR-049)

### Analytics Processor (US1)

- [ ] T082 [US1] Implement click event consumption in services/analytics-processor/internal/consumer/click_events.go (read from Redis Stream, batch 100 events)
- [ ] T083 [US1] Add IP anonymization logic (zero last octet IPv4, last 80 bits IPv6, FR-020)
- [ ] T084 [US1] Add GeoIP enrichment (country, city, lat/lon, FR-021)
- [ ] T085 [US1] Add user agent parsing (device type, browser, OS, FR-022)
- [ ] T086 [US1] Batch insert enriched events into TimescaleDB click_events hypertable

### Frontend (US1)

- [ ] T087 [P] [US1] Create URL creation form component in frontend/app/components/URLForm.tsx (destination URL, optional custom alias, title, notes, expiration)
- [ ] T088 [P] [US1] Create URL list component in frontend/app/components/URLList.tsx (display short URLs with click counts)
- [ ] T089 [US1] Implement POST /api/v1/urls API call in frontend/app/utils/api-client.ts (with auth header)
- [ ] T090 [US1] Update dashboard route in frontend/app/routes/dashboard.tsx (show URLForm + URLList, empty state if no URLs)
- [ ] T091 [US1] Create URL creation page in frontend/app/routes/create.tsx (dedicated page with URLForm)
- [ ] T092 [US1] Add TanStack Query mutation for URL creation (optimistic updates, error handling)

### Health Checks (US1)

- [ ] T093 [P] [US1] Create /health endpoint in services/api/internal/infrastructure/server/handlers/health.go (check DB, Redis, Zitadel connectivity)
- [ ] T094 [P] [US1] Create /health endpoint in services/redirector/src/handlers/health.rs (check DB, L1/L2 cache, event stream)
- [ ] T095 [P] [US1] Create /metrics endpoint in services/redirector/src/handlers/metrics.rs (Prometheus format: redirect latency, cache hit rates)

**Checkpoint**: MVP Core Complete - Users can create short URLs, anyone can visit them and redirect, click events captured

---

## Phase 5: User Story 2 - View Click Analytics (Priority: P2)

**Goal**: Users can view detailed analytics for their short URLs (clicks, geography, devices, referrers, time trends)

**Independent Test**: User creates URL, generates clicks from different sources/locations, views analytics dashboard with accurate stats and filters (per spec.md scenarios 1-5)

**Why This Priority**: Analytics provide the "why" for using this service over competitors.

### Application Layer (US2)

- [ ] T096 [P] [US2] Create GetAnalyticsSummary query in services/api/internal/application/queries/get_analytics.go (query TimescaleDB continuous aggregates, FR-023)
- [ ] T097 [P] [US2] Create GetClickTimeSeries query (hourly/daily/weekly buckets based on date range, FR-023)
- [ ] T098 [P] [US2] Create GetGeographicDistribution query (clicks by country/city from click_events, FR-021)
- [ ] T099 [P] [US2] Create GetTopReferrers query (top 10 referrer domains, FR-026)
- [ ] T100 [P] [US2] Create GetDeviceBreakdown query (desktop/mobile/tablet percentages, FR-022)

### Infrastructure Layer (US2)

- [ ] T101 [US2] Implement TimescaleAnalyticsRepository in services/api/internal/infrastructure/persistence/postgres/analytics_repository.go (query click_summary_hourly, click_summary_daily, click_events)
- [ ] T102 [US2] Add query optimization: use time_bucket, chunk exclusion filters (WHERE time >= ? AND url_id = ?)
- [ ] T103 [US2] Add caching for analytics queries (Redis, 5-minute TTL for dashboard views)

### HTTP Layer (US2)

- [ ] T104 [US2] Create GET /api/v1/analytics/:id handler in services/api/internal/infrastructure/server/handlers/analytics.go (return comprehensive analytics, FR-024)
- [ ] T105 [US2] Add query parameters: start_date, end_date, granularity (hour/day/week/month) per api-service.openapi.yaml
- [ ] T106 [P] [US2] Add authorization check (user can only view analytics for their own URLs)
- [ ] T107 [P] [US2] Create GET /api/v1/analytics/:id/export handler (CSV export of click data)

### Frontend (US2)

- [ ] T108 [P] [US2] Create AnalyticsCharts component in frontend/app/components/AnalyticsCharts.tsx (Recharts visualizations: line chart for time series, bar chart for countries, pie chart for devices)
- [ ] T109 [P] [US2] Create AnalyticsSummary component in frontend/app/components/AnalyticsSummary.tsx (total clicks, unique visitors, average clicks/day)
- [ ] T110 [P] [US2] Create TopReferrers component in frontend/app/components/TopReferrers.tsx (table of top referrer domains)
- [ ] T111 [US2] Create analytics detail route in frontend/app/routes/analytics.$id.tsx (full analytics page with charts, date range filter)
- [ ] T112 [US2] Add TanStack Query for analytics fetching (auto-refresh every 5 seconds, FR-024)
- [ ] T113 [US2] Update URLList component to show total clicks + link to analytics page per URL
- [ ] T114 [US2] Add date range picker to analytics page (Last 7 days, Last 30 days, custom range)

**Checkpoint**: Analytics Complete - Users can view detailed analytics with filters and visualizations

---

## Phase 6: User Story 3 - Manage Short URLs (Priority: P3)

**Goal**: Users can search/filter, edit, disable, and delete their short URLs

**Independent Test**: User can search URLs by title, edit URL metadata, disable URL (visitors see "Link Disabled"), delete URL (shows 404) (per spec.md scenarios 1-5)

**Why This Priority**: Enables long-term use as users create more links.

### Application Layer (US3)

- [ ] T115 [P] [US3] Create ListUserURLs query in services/api/internal/application/queries/list_urls.go (pagination, search, sorting, FR-028, FR-029)
- [ ] T116 [P] [US3] Create UpdateURLMetadata command in services/api/internal/application/commands/update_url.go (update title, notes, destination URL, FR-030)
- [ ] T117 [P] [US3] Create DisableURL command in services/api/internal/application/commands/disable_url.go (set status to 'disabled', FR-031)
- [ ] T118 [P] [US3] Create EnableURL command in services/api/internal/application/commands/enable_url.go (set status to 'active', FR-031)
- [ ] T119 [P] [US3] Create DeleteURL command in services/api/internal/application/commands/delete_url.go (soft delete: set status to 'deleted', FR-032)

### Infrastructure Layer (US3)

- [ ] T120 [US3] Add search/filter methods to PostgresURLRepository: ListByCreatorWithFilters (search by title/destination, filter by status, sort by clicks/date)
- [ ] T121 [US3] Add cache invalidation logic: on update/disable/delete, invalidate Redis L2 cache, set tombstone marker
- [ ] T122 [US3] Add update methods to PostgresURLRepository (UpdateMetadata, UpdateStatus with updated_at trigger)

### HTTP Layer (US3)

- [ ] T123 [US3] Create GET /api/v1/urls handler in services/api/internal/infrastructure/server/handlers/urls.go (list with pagination, search, sort per api-service.openapi.yaml)
- [ ] T124 [P] [US3] Create GET /api/v1/urls/:id handler (get single URL details)
- [ ] T125 [P] [US3] Create PATCH /api/v1/urls/:id handler (update URL metadata)
- [ ] T126 [P] [US3] Create POST /api/v1/urls/:id/disable handler (disable URL)
- [ ] T127 [P] [US3] Create POST /api/v1/urls/:id/enable handler (enable URL)
- [ ] T128 [P] [US3] Create DELETE /api/v1/urls/:id handler (soft delete URL)
- [ ] T129 [US3] Add authorization checks to all handlers (user can only manage their own URLs)

### Redirector Service (US3)

- [ ] T130 [US3] Update redirect handler to check URL status: return 410 Gone for disabled/deleted URLs (HTML error page)

### Frontend (US3)

- [ ] T131 [P] [US3] Add search bar to URLList component (search by title or destination URL)
- [ ] T132 [P] [US3] Add filter dropdown to URLList component (filter by status: all/active/disabled)
- [ ] T133 [P] [US3] Add sort dropdown to URLList component (sort by creation date, click count, last modified)
- [ ] T134 [P] [US3] Create EditURLModal component in frontend/app/components/EditURLModal.tsx (form to edit title, notes, destination URL)
- [ ] T135 [US3] Add action buttons to URLList items: Edit, Disable/Enable, Delete (with confirmation dialog)
- [ ] T136 [US3] Implement TanStack Query mutations for update/disable/enable/delete (optimistic updates, invalidate cache)
- [ ] T137 [US3] Add URL status badges to URLList component (active: green, disabled: yellow, expired: orange, deleted: red)

**Checkpoint**: URL Management Complete - Users can search, filter, edit, disable, and delete URLs

---

## Phase 7: User Story 4 - Programmatic URL Creation via API Key (Priority: P4)

**Goal**: Users can generate API keys and create short URLs programmatically via HTTP requests

**Independent Test**: User generates API key, uses it to create short URLs via curl/HTTP client, sees URLs in dashboard (per spec.md scenarios 1-5)

**Why This Priority**: Enables advanced use cases (bulk shortening, automation) but core web UI delivers primary value.

### Domain Layer (US4)

- [ ] T138 [P] [US4] Create APIKey entity in services/api/internal/domain/apikey/entity.go (ID, key hash, key prefix, name, status, usage count, owner)
- [ ] T139 [P] [US4] Create BLAKE2b hashing logic in services/api/internal/domain/apikey/hash.go (hash key, validate hash, FR-034)
- [ ] T140 [P] [US4] Create APIKeyRepository interface in services/api/internal/domain/apikey/repository.go (Create, GetByHash, ListByUserID, Revoke, IncrementUsage)
- [ ] T141 [P] [US4] Create domain errors in services/api/internal/domain/apikey/errors.go (APIKeyNotFound, APIKeyRevoked, InvalidAPIKey)

### Application Layer (US4)

- [ ] T142 [P] [US4] Create GenerateAPIKey command in services/api/internal/application/commands/generate_api_key.go (generate 32-byte random key, hash with BLAKE2b, extract prefix, FR-034)
- [ ] T143 [P] [US4] Create RevokeAPIKey command in services/api/internal/application/commands/revoke_api_key.go (set status to 'revoked', FR-036)
- [ ] T144 [P] [US4] Create ListUserAPIKeys query in services/api/internal/application/queries/list_api_keys.go (list with include_revoked filter)
- [ ] T145 [P] [US4] Create ValidateAPIKey query in services/api/internal/application/queries/validate_api_key.go (hash provided key, lookup by hash, check status, FR-035)

### Infrastructure Layer (US4)

- [ ] T146 [US4] Implement PostgresAPIKeyRepository in services/api/internal/infrastructure/persistence/postgres/apikey_repository.go (implements domain interface)
- [ ] T147 [US4] Add API key rate limiting: use Redis incr with 1-hour expiration (1000 requests/hour per key, FR-041)

### HTTP Layer (US4)

- [ ] T148 [US4] Update authentication middleware to support X-API-Key header (validate API key if no JWT, extract user ID from API key)
- [ ] T149 [P] [US4] Create POST /api/v1/api-keys handler in services/api/internal/infrastructure/server/handlers/api_keys.go (generate API key, return full key once)
- [ ] T150 [P] [US4] Create GET /api/v1/api-keys handler (list user's API keys, exclude revoked by default)
- [ ] T151 [P] [US4] Create DELETE /api/v1/api-keys/:id handler (revoke API key)
- [ ] T152 [US4] Update POST /api/v1/urls handler to accept API key authentication (use same CreateShortURL command, FR-035)

### Frontend (US4)

- [ ] T153 [P] [US4] Create API key settings page in frontend/app/routes/settings/api-keys.tsx (list API keys, generate new, revoke)
- [ ] T154 [P] [US4] Create APIKeyList component in frontend/app/components/APIKeyList.tsx (table with key prefix, name, created date, last used, usage count, revoke button)
- [ ] T155 [P] [US4] Create GenerateAPIKeyModal component in frontend/app/components/GenerateAPIKeyModal.tsx (form with key name input, show full key once with copy button)
- [ ] T156 [US4] Implement TanStack Query mutations for generate/revoke API keys
- [ ] T157 [US4] Add navigation link to API Keys settings in dashboard header

**Checkpoint**: API Key Access Complete - Users can generate API keys, create URLs via API, manage keys

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that span multiple user stories and production readiness

### Performance Optimization

- [ ] T158 [P] Add index monitoring and optimization queries to migrations (EXPLAIN ANALYZE for hot queries)
- [ ] T159 [P] Implement connection pooling tuning for all services (max connections, idle timeout per plan.md)
- [ ] T160 [P] Add query result caching for frequently accessed data (user profile, URL metadata)
- [ ] T161 [P] Optimize TimescaleDB continuous aggregate refresh intervals (balance freshness vs load)

### Security Hardening

- [ ] T162 [P] Add rate limiting for all API endpoints (per-user limits on top of per-API-key limits)
- [ ] T163 [P] Add input sanitization for all user inputs (URL validation, alias validation, title/notes XSS prevention, FR-046)
- [ ] T164 [P] Add CSRF protection to all state-changing endpoints (FR-045)
- [ ] T165 [P] Add HTTPS enforcement in production (redirect HTTP to HTTPS, FR-043)
- [ ] T166 [P] Add SQL injection prevention audit (ensure all queries use parameterized statements)

### Observability

- [ ] T167 [P] Add structured logging to all services (request ID, user ID, trace context per Principle V)
- [ ] T168 [P] Add Prometheus metrics to API service (request latency, error rates, active connections)
- [ ] T169 [P] Add Prometheus metrics to Redirector service (redirect latency histogram, cache hit rates by tier)
- [ ] T170 [P] Add Prometheus metrics to Analytics Processor (event processing lag, batch size, insert rate)
- [ ] T171 [P] Add health check aggregation dashboard (monitor all services, dependencies)

### Documentation

- [ ] T172 [P] Create API documentation from OpenAPI specs (hosted at /docs endpoint)
- [ ] T173 [P] Create deployment guide in specs/016-authenticated-url-shortener/deployment.md (Kubernetes manifests, environment variables)
- [ ] T174 [P] Update quickstart.md with troubleshooting section (common errors, solutions)
- [ ] T175 [P] Create runbook for operations in specs/016-authenticated-url-shortener/runbook.md (scaling, backups, incident response)

### Testing & Validation

- [ ] T176 [P] Run quickstart.md validation end-to-end (verify all setup steps work, all scenarios pass)
- [ ] T177 [P] Add load testing for redirector service (validate <50ms p95 latency at 10K concurrent requests, FR-016)
- [ ] T178 [P] Add integration test for analytics pipeline (verify click events flow from redirector → analytics processor → TimescaleDB with <5s latency, FR-024)
- [ ] T179 [P] Add contract tests for API service (validate all endpoints match OpenAPI spec)
- [ ] T180 [P] Add contract tests for Redirector service (validate redirect behavior matches OpenAPI spec)

### Infrastructure

- [ ] T181 [P] Create Dockerfile for API service in services/api/Dockerfile (multi-stage build, minimal image)
- [ ] T182 [P] Create Dockerfile for Redirector service in services/redirector/Dockerfile (multi-stage build, minimal image)
- [ ] T183 [P] Create Dockerfile for Analytics Processor in services/analytics-processor/Dockerfile
- [ ] T184 [P] Create Dockerfile for Frontend in frontend/Dockerfile (build static assets, serve with Node)
- [ ] T185 [P] Add Docker Compose production profile in docker-compose.prod.yml (includes all services, Traefik reverse proxy)
- [ ] T186 [P] Create Kubernetes manifests in k8s/ (Deployments, Services, HPA, ConfigMaps, Secrets)

### Cleanup

- [ ] T187 [P] Code review and refactoring across all services (remove dead code, improve naming)
- [ ] T188 [P] Security audit (dependency scanning, vulnerability checks)
- [ ] T189 [P] Performance profiling (identify bottlenecks, optimize hot paths)
- [ ] T190 [P] Final constitution compliance check (verify all 5 principles followed)
- [ ] T191 [P] Create user account deletion flow in services/api/internal/application/commands/delete_user.go (cascade deactivate URLs, anonymize email per data-model.md lines 641-645, delete API keys per Edge Case 7)
- [ ] T192 [P] Add TimescaleDB retention/compression validation test in tests/integration/timescale_retention_test.go (verify 5-year retention policy, 90% compression, query performance at 600M+ rows per technical constraints)
- [ ] T193 [P] [OPTIONAL] Implement destination unavailable warning banner in services/redirector/src/handlers/redirect.rs (optional HEAD request to check destination health, show warning if 404/500 per Edge Case 10)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 5 (Phase 3)**: Depends on Foundational - PREREQUISITE for all other stories (authentication required)
- **User Story 1 (Phase 4)**: Depends on US5 completion (users must be authenticated to create URLs)
- **User Story 2 (Phase 5)**: Depends on US1 completion (need URLs to have analytics)
- **User Story 3 (Phase 6)**: Depends on US1 completion (need URLs to manage them)
- **User Story 4 (Phase 7)**: Depends on US1 completion (API keys create URLs using same logic)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### Critical Path (MVP)

```
Setup → Foundational → US5 (Auth) → US1 (URL Creation + Redirect) → MVP COMPLETE
```

**MVP Scope**: User Story 5 + User Story 1

- Users can sign up, log in, create short URLs, visit them to redirect
- Click events captured but analytics not yet viewable (US2 adds analytics dashboard)
- Total tasks for MVP: T001-T095 (95 tasks)

### Extended Features (Post-MVP)

```
MVP → US2 (Analytics) → US3 (URL Management) → US4 (API Keys) → FULL FEATURE SET
```

### User Story Dependencies

- **US5 (Auth)**: Can start after Foundational - No dependencies on other stories
- **US1 (Create URLs)**: Depends on US5 - Independently testable once US5 complete
- **US2 (Analytics)**: Depends on US1 - Can integrate with US1 but independently testable
- **US3 (Manage URLs)**: Depends on US1 - Can integrate with US1 but independently testable
- **US4 (API Keys)**: Depends on US1 - Can integrate with US1 but independently testable

### Within Each User Story

- Domain layer before application layer (entities before commands/queries)
- Application layer before infrastructure layer (interfaces before implementations)
- Infrastructure layer before HTTP layer (repositories before handlers)
- Backend before frontend (API endpoints before UI components)

### Parallel Opportunities

**Within Setup (Phase 1)**: All tasks T002-T010 can run in parallel

**Within Foundational (Phase 2)**:

- API Service Foundation (T011-T026): Most tasks marked [P] can run in parallel
- Database Migrations (T027-T031): Must run sequentially (dependencies between migrations)
- Redirector Foundation (T033-T039): Most tasks marked [P] can run in parallel
- Analytics Foundation (T040-T043): All tasks marked [P] can run in parallel
- Frontend Foundation (T045-T049): All tasks marked [P] can run in parallel

**Within User Stories**:

- Domain layer entities marked [P] can run in parallel (different files)
- Application layer commands/queries marked [P] can run in parallel
- HTTP handlers marked [P] can run in parallel
- Frontend components marked [P] can run in parallel

**Across User Stories** (after US5 complete):

- US1, US2, US3, US4 can be developed in parallel by different team members IF:
  - Each story has its own domain/application/infrastructure/HTTP layers
  - US1 completed first (or work on US1 core + others' prep in parallel)
  - Integration happens at the end (US2/US3/US4 integrate with US1 components)

---

## Parallel Example: User Story 1 Core Implementation

```bash
# After Foundational phase complete, launch US1 domain layer in parallel:
Task T064: "Create URL aggregate root in services/api/internal/domain/url/entity.go"
Task T065: "Create ShortCode value object in services/api/internal/domain/url/short_code.go"
Task T066: "Create URLRepository interface in services/api/internal/domain/url/repository.go"
Task T067: "Create domain errors in services/api/internal/domain/url/errors.go"

# Then launch US1 application layer in parallel (after domain layer complete):
Task T069: "Create CreateShortURL command in services/api/internal/application/commands/create_url.go"
Task T070: "Create GetURLByShortCode query in services/api/internal/application/queries/get_url.go"

# Then launch US1 infrastructure + HTTP in parallel (after application layer complete):
Task T071: "Implement PostgresURLRepository" (infrastructure)
Task T072: "Create Safe Browsing API client" (infrastructure)
Task T073: "Add cache warming logic" (infrastructure)
Task T074: "Create POST /api/v1/urls handler" (HTTP - depends on T071)
Task T075: "Add request validation" (HTTP)
Task T076: "Add response mapping" (HTTP)

# Redirector and Analytics can start in parallel with HTTP layer:
Task T077: "Implement GET /:shortCode handler" (redirector)
Task T082: "Implement click event consumption" (analytics processor)

# Frontend components can start once HTTP layer has endpoint specs:
Task T087: "Create URL creation form component"
Task T088: "Create URL list component"
```

---

## Implementation Strategy

### MVP First (US5 + US1 Only)

1. Complete Phase 1: Setup → Project structure ready
2. Complete Phase 2: Foundational → Infrastructure ready (CRITICAL BLOCKER)
3. Complete Phase 3: US5 (Auth) → Users can sign up, log in
4. Complete Phase 4: US1 (URL Creation + Redirect) → Users can create short URLs, redirects work
5. **STOP and VALIDATE**: Run quickstart.md scenarios for US5 + US1
6. Deploy/demo MVP (authentication + URL shortening + redirect)

**MVP Deliverable**: 95 tasks (T001-T095)
**MVP Value**: Users can create short URLs and share them - core product value delivered

### Incremental Delivery (Post-MVP)

1. **MVP (US5 + US1)** → Test independently → Deploy
2. **+ US2 (Analytics)** → T096-T114 (19 tasks) → Test independently → Deploy
3. **+ US3 (URL Management)** → T115-T137 (23 tasks) → Test independently → Deploy
4. **+ US4 (API Keys)** → T138-T157 (20 tasks) → Test independently → Deploy
5. **+ Polish** → T158-T190 (33 tasks) → Final production-ready system

**Each increment adds value without breaking previous functionality**

### Parallel Team Strategy

With multiple developers (recommended team size: 4-6):

**Phase 1-2 (Setup + Foundational)**: Everyone works together

- Developer 1: API service foundation
- Developer 2: Redirector service foundation
- Developer 3: Analytics processor foundation
- Developer 4: Frontend foundation + migrations

**Phase 3 (US5 Auth)**: Everyone works together (authentication is prerequisite)

**Phase 4-7 (User Stories)**: Parallel development

- Developer 1: US1 (API service + domain logic)
- Developer 2: US1 (Redirector service)
- Developer 3: US1 (Analytics processor)
- Developer 4: US1 (Frontend)
- Once US1 complete, developers can split to different stories:
  - Developer 1 → US2 (Analytics backend)
  - Developer 2 → US3 (URL Management backend)
  - Developer 3 → US4 (API Keys backend)
  - Developer 4 → US2/US3/US4 (Frontend for all stories)

**Phase 8 (Polish)**: Everyone works together on cross-cutting concerns

---

## Notes

- **[P] marker**: Tasks marked [P] can run in parallel (different files, no blocking dependencies)
- **[Story] label**: Maps task to user story (US1-US5 from spec.md) for traceability
- **File paths**: Follow exact structure from plan.md (3 services + frontend + migrations)
- **Constitution alignment**: Foundational phase implements all 5 principles before user story work begins
- **Independent testing**: Each user story can be tested independently once its phase completes
- **MVP focus**: US5 + US1 deliver core value (authentication + URL shortening + redirect)
- **Incremental delivery**: Add US2 (analytics), US3 (management), US4 (API keys) post-MVP
- **Commit strategy**: Commit after each task or logical group (e.g., all domain entities for a story)
- **Validation**: Run quickstart.md scenarios at each checkpoint to verify story completeness
- **Performance targets**: Redirector must achieve <50ms p95 latency (FR-016), analytics <5s visibility (FR-024)

**Total Tasks**: 193 tasks

- Setup: 10 tasks
- Foundational: 39 tasks (BLOCKS all stories)
- US5 (Auth): 14 tasks (MVP prerequisite)
- US1 (URL Creation): 32 tasks (MVP core)
- US2 (Analytics): 19 tasks (Post-MVP)
- US3 (URL Management): 23 tasks (Post-MVP)
- US4 (API Keys): 20 tasks (Post-MVP)
- Polish: 36 tasks (Production ready - includes T191-T193)

**MVP Task Count**: 95 tasks (Setup + Foundational + US5 + US1)
**Full Feature Set**: 193 tasks (all phases)

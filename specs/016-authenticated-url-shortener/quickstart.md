# Quickstart Guide: Authenticated URL Shortener Platform

**Feature**: 016-authenticated-url-shortener  
**Date**: 2026-01-09  
**Related**: [spec.md](./spec.md) | [plan.md](./plan.md) | [data-model.md](./data-model.md)

---

## Overview

This guide helps you set up a complete local development environment for the distributed URL shortener platform in under 30 minutes. You'll run all services (API, Redirector, Analytics Processor, Frontend) plus dependencies (PostgreSQL, TimescaleDB, Redis) using Docker Compose. **Note**: This setup assumes you have Zitadel already deployed in external Docker containers.

**Prerequisites**:
- Docker 24+ and Docker Compose 2.20+
- Go 1.22+ (for API and Analytics services)
- Rust 1.75+ (for Redirector service)
- Node.js 20+ (for Frontend)
- 8GB RAM minimum (16GB recommended)
- 10GB free disk space

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Local Development Stack                   │
├─────────────────────────────────────────────────────────────┤
│  Frontend (TanStack Start)         :5173                    │
│  API Service (Go)                  :8080                     │
│  Redirector Service (Rust)         :3000                    │
│  Analytics Processor (Go)          (background worker)       │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL 16 + TimescaleDB       :5432                    │
│  Redis/Valkey 7.2                  :6379                    │
│  MaxMind GeoLite2 Database         (file: /data/geoip/)     │
├─────────────────────────────────────────────────────────────┤
│  External Dependencies (Already Deployed):                   │
│  Zitadel (Identity Provider)       (configured via .env)    │
└─────────────────────────────────────────────────────────────┘
```

---

## Step 1: Clone Repository and Setup

```bash
# Clone repository
git clone https://github.com/refract/url-shortener.git
cd url-shortener

# Checkout feature branch
git checkout 016-authenticated-url-shortener

# Create necessary directories
mkdir -p data/postgres data/redis data/geoip logs

# Copy environment template
cp .env.example .env
```

---

## Step 2: Configure Environment Variables

Edit `.env` file with your local configuration:

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=refract
POSTGRES_PASSWORD=dev_password_change_in_prod
POSTGRES_DB=url_shortener
DATABASE_URL=postgres://refract:dev_password_change_in_prod@localhost:5432/url_shortener?sslmode=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Zitadel Configuration (External Instance)
ZITADEL_URL=https://your-zitadel-instance.com  # URL of your existing Zitadel deployment
ZITADEL_CLIENT_ID=your_client_id_here          # From your Zitadel application configuration
ZITADEL_CLIENT_SECRET=your_client_secret_here  # From your Zitadel application configuration
ZITADEL_ISSUER=https://your-zitadel-instance.com  # Must match the issuer in JWT tokens

# API Service Configuration
API_PORT=8080
API_BASE_URL=http://localhost:8080
JWT_ISSUER=https://your-zitadel-instance.com  # Must match ZITADEL_ISSUER
WORKER_ID=1  # Snowflake ID worker ID (0-63)

# Redirector Service Configuration
REDIRECTOR_PORT=3000
REDIRECTOR_BASE_URL=http://localhost:3000
REDIRECTOR_WORKER_ID=64  # Separate range from API service

# Frontend Configuration
VITE_API_URL=http://localhost:8080
VITE_SHORT_URL_BASE=http://localhost:3000
VITE_ZITADEL_AUTHORITY=https://your-zitadel-instance.com  # Must match ZITADEL_ISSUER
VITE_ZITADEL_CLIENT_ID=your_client_id_here

# Analytics Processor Configuration
ANALYTICS_BATCH_SIZE=100
ANALYTICS_FLUSH_INTERVAL=1s

# MaxMind GeoIP Configuration
GEOIP_DB_PATH=./data/geoip/GeoLite2-City.mmdb
GEOIP_LICENSE_KEY=your_maxmind_license_key_here

# Safe Browsing API Configuration (Google)
SAFE_BROWSING_API_KEY=your_google_api_key_here

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

**Note**: Replace placeholder values with your actual external Zitadel configuration:
- `your-zitadel-instance.com`: Your existing Zitadel deployment URL
- `your_client_id_here`: Client ID from your Zitadel application (see Step 6)
- `your_client_secret_here`: Client Secret from your Zitadel application
- Ensure your Zitadel instance is accessible from your development environment

---

## Step 3: Start Infrastructure Services (Docker Compose)

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Wait for services to be healthy (30-60 seconds)
docker-compose ps

# Check logs if services fail to start
docker-compose logs postgres
docker-compose logs redis
```

**Docker Compose Configuration** (`docker-compose.yml`):

```yaml
version: '3.9'

services:
  postgres:
    image: timescale/timescaledb:latest-pg16
    container_name: url-shortener-postgres
    environment:
      POSTGRES_USER: refract
      POSTGRES_PASSWORD: dev_password_change_in_prod
      POSTGRES_DB: url_shortener
    ports:
      - "5432:5432"
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
      - ./migrations/postgres:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U refract"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: valkey/valkey:7.2-alpine
    container_name: url-shortener-redis
    command: valkey-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - ./data/redis:/data
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
```

**Note**: Zitadel is NOT included in this docker-compose.yml as it's already deployed externally. Services will connect to your existing Zitadel instance via the environment variables configured in Step 2.

---

## Step 4: Run Database Migrations

```bash
# Install migration tool (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations/postgres -database "$DATABASE_URL" up

# Verify migrations
migrate -path migrations/postgres -database "$DATABASE_URL" version

# Check tables created
psql $DATABASE_URL -c "\dt"
# Expected tables: users, urls, api_keys, click_events

# Verify TimescaleDB extension
psql $DATABASE_URL -c "SELECT * FROM timescaledb_information.hypertables;"
# Expected: click_events hypertable
```

**Migration Files** (`migrations/postgres/`):

- `00001_create_users.up.sql`: Create users table
- `00002_create_urls.up.sql`: Create urls table with indexes
- `00003_create_api_keys.up.sql`: Create api_keys table
- `00004_create_timescale_hypertables.up.sql`: Create click_events hypertable and continuous aggregates

---

## Step 5: Download MaxMind GeoLite2 Database

```bash
# Register for free MaxMind account: https://www.maxmind.com/en/geolite2/signup
# Get license key from: https://www.maxmind.com/en/accounts/current/license-key

# Download GeoLite2 City database
curl -o data/geoip/GeoLite2-City.tar.gz \
  "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_LICENSE_KEY&suffix=tar.gz"

# Extract .mmdb file
tar -xzf data/geoip/GeoLite2-City.tar.gz -C data/geoip --strip-components=1
rm data/geoip/GeoLite2-City.tar.gz

# Verify database exists
ls -lh data/geoip/GeoLite2-City.mmdb
# Expected: ~70MB file
```

---

## Step 6: Configure Zitadel Application (One-Time Setup)

**Prerequisites**: Your external Zitadel instance must be accessible and you must have admin access.

### 6.1 Access Your Zitadel Instance

```bash
# Navigate to your Zitadel admin console
# Example: https://your-zitadel-instance.com
# Login with your Zitadel admin credentials
```

### 6.2 Create Application for URL Shortener

1. Navigate to your organization/project
2. Click **"Applications"** → **"New"**
3. Configure application:
   - **Name**: "URL Shortener"
   - **Type**: Web
4. Configure OIDC settings:
   - **Redirect URIs**: `http://localhost:5173/auth/callback`
   - **Post Logout URIs**: `http://localhost:5173`
   - **Grant Types**: Authorization Code, Refresh Token
   - **Response Types**: Code
5. Click **"Create"**
6. **Copy Client ID and Client Secret** → Update your `.env` file:
   ```bash
   ZITADEL_CLIENT_ID=<paste-client-id-here>
   ZITADEL_CLIENT_SECRET=<paste-client-secret-here>
   ```

### 6.3 Configure OAuth Providers (Optional)

If you want users to sign in with Google or GitHub:

1. Navigate to **"Identity Providers"** in your Zitadel console
2. Add **Google** provider:
   - Enter Google OAuth Client ID and Secret
   - Enable for your application
3. Add **GitHub** provider:
   - Enter GitHub OAuth App Client ID and Secret
   - Enable for your application

### 6.4 Verify Zitadel Endpoints

Ensure your Zitadel instance has the following endpoints accessible:

**Test connectivity**:
```bash
# Replace with your actual Zitadel URL
ZITADEL_URL=https://your-zitadel-instance.com

# Test OIDC discovery endpoint
curl $ZITADEL_URL/.well-known/openid-configuration

# Expected: JSON response with authorization_endpoint, token_endpoint, etc.
```

**Zitadel Configuration Values** (for reference):
- **Issuer**: `https://your-zitadel-instance.com`
- **Authorization Endpoint**: `https://your-zitadel-instance.com/oauth/v2/authorize`
- **Token Endpoint**: `https://your-zitadel-instance.com/oauth/v2/token`
- **UserInfo Endpoint**: `https://your-zitadel-instance.com/oidc/v1/userinfo`
- **JWKS URI**: `https://your-zitadel-instance.com/oauth/v2/keys`

---

## Step 7: Start Backend Services

### 7.1 Start API Service (Go)

```bash
cd services/api

# Install dependencies
go mod download

# Run database migrations (if not done in Step 4)
go run cmd/migrate/main.go up

# Start API service
go run cmd/api/main.go

# Expected output:
# 2026-01-09T12:00:00Z INFO  Starting API service version=1.0.0 port=8080
# 2026-01-09T12:00:01Z INFO  Connected to PostgreSQL host=localhost:5432
# 2026-01-09T12:00:01Z INFO  Connected to Redis host=localhost:6379
# 2026-01-09T12:00:01Z INFO  Zitadel OIDC initialized issuer=http://localhost:8081
# 2026-01-09T12:00:01Z INFO  API service listening addr=:8080
```

**Health Check**:
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","version":"1.0.0","dependencies":{"database":{"status":"up"},"cache":{"status":"up"},"zitadel":{"status":"up"}}}
```

---

### 7.2 Start Redirector Service (Rust)

```bash
cd services/redirector

# Install dependencies
cargo build

# Start redirector service
cargo run --release

# Expected output:
# [2026-01-09T12:00:00Z INFO  redirector] Starting redirector service version=1.0.0 port=3000
# [2026-01-09T12:00:01Z INFO  redirector] Connected to PostgreSQL host=localhost:5432
# [2026-01-09T12:00:01Z INFO  redirector] Connected to Redis host=localhost:6379
# [2026-01-09T12:00:02Z INFO  redirector] L1 cache initialized capacity=10000
# [2026-01-09T12:00:02Z INFO  redirector] Redirector service listening addr=0.0.0.0:3000
```

**Health Check**:
```bash
curl http://localhost:3000/health
# Expected: {"status":"healthy","version":"1.0.0","uptime_seconds":5,"dependencies":{"database":{"status":"up"},"cache_l1":{"status":"up","size":0},"cache_l2":{"status":"up"}}}
```

---

### 7.3 Start Analytics Processor (Go)

```bash
cd services/analytics-processor

# Install dependencies
go mod download

# Start analytics processor
go run cmd/processor/main.go

# Expected output:
# 2026-01-09T12:00:00Z INFO  Starting analytics processor version=1.0.0
# 2026-01-09T12:00:01Z INFO  Connected to TimescaleDB host=localhost:5432
# 2026-01-09T12:00:01Z INFO  Connected to Redis Stream stream=click_events group=analytics_processor
# 2026-01-09T12:00:01Z INFO  Consumer ready batch_size=100 flush_interval=1s
# 2026-01-09T12:00:02Z INFO  Processing events count=0 lag=0ms
```

---

## Step 8: Start Frontend (TanStack Start)

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Expected output:
# VITE v5.0.0  ready in 1234 ms
# 
# ➜  Local:   http://localhost:5173/
# ➜  Network: use --host to expose
# ➜  press h to show help
```

**Verify Frontend**:
1. Open browser: http://localhost:5173
2. Click "Sign Up" → Redirected to Zitadel
3. Register with test account: `test@example.com` / `Password1!`
4. After authentication, redirected to dashboard

---

## Step 9: Verify End-to-End Flow

### 9.1 Create Short URL (Web UI)

1. Navigate to http://localhost:5173/dashboard
2. Click "Create Short URL"
3. Enter destination URL: `https://example.com/article`
4. Click "Shorten"
5. Copy short URL: `http://localhost:3000/dBvJIX9uyO`

### 9.2 Test Redirect

```bash
# Visit short URL (follow redirect)
curl -L http://localhost:3000/dBvJIX9uyO
# Expected: HTML content from https://example.com/article

# Check redirect headers
curl -I http://localhost:3000/dBvJIX9uyO
# Expected:
# HTTP/1.1 301 Moved Permanently
# Location: https://example.com/article
# Cache-Control: public, max-age=3600
# X-Short-Code: dBvJIX9uyO
```

### 9.3 Verify Analytics

```bash
# Wait 5 seconds for analytics processing (FR-024)
sleep 5

# Query analytics API
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/analytics/1234567890123456

# Expected:
# {
#   "url_id": 1234567890123456,
#   "summary": {
#     "total_clicks": 1,
#     "unique_visitors": 1
#   },
#   "time_series": [...],
#   "geographic_distribution": [...]
# }
```

### 9.4 Test API Key Creation

```bash
# Generate API key via UI: Settings → API Keys → Generate New Key
# Copy API key: refract_abc123xyz456...

# Create URL via API
curl -X POST http://localhost:8080/api/v1/urls \
  -H "X-API-Key: refract_abc123xyz456..." \
  -H "Content-Type: application/json" \
  -d '{
    "destination_url": "https://example.com/api-test",
    "title": "API Test URL"
  }'

# Expected:
# {
#   "id": 9876543210987654,
#   "short_code": "xYz9Abc",
#   "short_url": "http://localhost:3000/xYz9Abc",
#   "destination_url": "https://example.com/api-test",
#   "status": "active",
#   ...
# }
```

---

## Step 10: Seed Test Data (Optional)

```bash
# Run seed script to generate sample URLs and click events
cd scripts
go run seed.go --users=10 --urls-per-user=50 --clicks-per-url=100

# Expected output:
# Seeding database...
# Created 10 users
# Created 500 URLs
# Generated 50,000 click events
# Seed completed in 12.34s
```

**Verify Seeded Data**:
```bash
# Check URL count
psql $DATABASE_URL -c "SELECT COUNT(*) FROM urls;"
# Expected: 500

# Check click event count
psql $DATABASE_URL -c "SELECT COUNT(*) FROM click_events;"
# Expected: 50,000

# Check continuous aggregate refresh
psql $DATABASE_URL -c "SELECT COUNT(*) FROM click_summary_hourly;"
# Expected: ~500-1000 rows (depending on time spread)
```

---

## Common Issues and Troubleshooting

### Issue 1: PostgreSQL Connection Refused

**Symptoms**: `dial tcp [::1]:5432: connect: connection refused`

**Solution**:
```bash
# Check if PostgreSQL container is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres

# Wait for health check to pass
docker-compose ps postgres | grep healthy
```

---

### Issue 2: Redirector L2 Cache Misses

**Symptoms**: `401 Unauthorized` on API requests, logs show "failed to verify token"

**Solution**:
```bash
# Verify Zitadel JWKS endpoint accessible from your development environment
curl https://your-zitadel-instance.com/oauth/v2/keys

# If connection fails, check:
# 1. Zitadel instance is running and accessible
# 2. Network connectivity (firewall, VPN, etc.)
# 3. .env ZITADEL_ISSUER matches your actual Zitadel URL

# Check .env ZITADEL_ISSUER matches JWT issuer claim
# Token issuer must exactly match ZITADEL_ISSUER value

# Debug JWT token (decode at jwt.io)
# Verify:
# - iss claim matches ZITADEL_ISSUER
# - aud claim matches ZITADEL_CLIENT_ID
# - exp claim is in future (not expired)
```

---

### Issue 3: API Service Can't Validate JWT

**Symptoms**: All redirects show `cache_tier: db` in logs (high latency)

**Solution**:
```bash
# Check Redis connectivity
docker-compose logs redis
redis-cli ping  # Expected: PONG

# Check cache keys in Redis
redis-cli --scan --pattern "url:*" | head -10

# Verify API service populating cache on URL creation
# Check logs: "cache warmed" message after URL creation

# Manually warm cache for testing
redis-cli SET "url:1234567890123456" "https://example.com/article" EX 3600
```

---

### Issue 4: Analytics Not Updating

**Symptoms**: Dashboard shows 0 clicks after redirect

**Solution**:
```bash
# Check analytics processor logs
cd services/analytics-processor
go run cmd/processor/main.go

# Expected: "Processing events count=X"

# Check Redis Stream has events
redis-cli XLEN click_events
# Expected: > 0

# Manually consume event for debugging
redis-cli XREAD COUNT 1 STREAMS click_events 0

# Check TimescaleDB events inserted
psql $DATABASE_URL -c "SELECT COUNT(*) FROM click_events WHERE time > NOW() - INTERVAL '1 minute';"
# Expected: > 0

# Force continuous aggregate refresh
psql $DATABASE_URL -c "CALL refresh_continuous_aggregate('click_summary_hourly', NULL, NULL);"
```

---

## Development Workflow

### Running Tests

```bash
# API Service Tests
cd services/api
go test ./... -v -cover

# Redirector Service Tests
cd services/redirector
cargo test

# Frontend Tests
cd frontend
npm test

# Integration Tests (all services must be running)
cd tests/integration
go test ./... -v
```

---

### Hot Reload

**API Service** (Go):
```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

cd services/api
air  # Watches for .go file changes, auto-restarts
```

**Redirector Service** (Rust):
```bash
# Install cargo-watch
cargo install cargo-watch

cd services/redirector
cargo watch -x run  # Recompiles on .rs file changes
```

**Frontend** (TanStack Start):
```bash
cd frontend
npm run dev  # Built-in HMR (Hot Module Replacement)
```

---

### Database Management

**Reset Database**:
```bash
# Drop all tables and re-run migrations
migrate -path migrations/postgres -database "$DATABASE_URL" drop -f
migrate -path migrations/postgres -database "$DATABASE_URL" up
```

**Backup Database**:
```bash
# Dump database to file
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql

# Restore from backup
psql $DATABASE_URL < backup_20260109.sql
```

**View Logs**:
```bash
# Tail PostgreSQL logs
docker-compose logs -f postgres

# Tail Redis logs
docker-compose logs -f redis

# Tail all service logs
docker-compose logs -f
```

---

### Monitoring and Metrics

**Prometheus Metrics**:
```bash
# API service metrics
curl http://localhost:8080/metrics

# Redirector service metrics
curl http://localhost:3000/metrics
```

**Cache Statistics**:
```bash
# L1 cache (in redirector service logs)
curl http://localhost:3000/health | jq '.dependencies.cache_l1'

# L2 cache (Redis)
redis-cli INFO stats
```

**Database Performance**:
```bash
# Active queries
psql $DATABASE_URL -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"

# Table sizes
psql $DATABASE_URL -c "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# TimescaleDB chunk statistics
psql $DATABASE_URL -c "SELECT * FROM timescaledb_information.chunks WHERE hypertable_name = 'click_events';"
```

---

## Next Steps

After completing this quickstart:

1. **Read Architecture Documentation**: See `plan.md` for detailed architecture decisions
2. **Explore Data Model**: See `data-model.md` for entity relationships and schemas
3. **Review API Contracts**: See `contracts/api-service.openapi.yaml` for full API reference
4. **Implement Features**: Follow feature branch tasks from `/speckit.tasks` output
5. **Write Tests**: Add contract tests, integration tests, load tests

---

## Support

**Documentation**:
- Spec: `specs/016-authenticated-url-shortener/spec.md`
- Plan: `specs/016-authenticated-url-shortener/plan.md`
- Data Model: `specs/016-authenticated-url-shortener/data-model.md`

**Issues**:
- Open GitHub issue: https://github.com/refract/url-shortener/issues
- Check existing issues: Search for error messages

**Community**:
- Discord: #url-shortener channel
- Weekly sync: Fridays 2pm UTC

---

**Quickstart Status**: ✅ Complete - Ready for development

**Estimated Setup Time**: 20-30 minutes (including downloads and migrations)

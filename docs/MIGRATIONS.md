# Database Migration Architecture

This document describes the new containerized migration system for the Refract application.

## Overview

The migration system is now containerized in a separate `migrations` service that runs before any application services start. This ensures all database schemas are properly initialized and prevents conflicts between different migration tools.

### Key Benefits

✅ **Automatic Initialization** - Migrations run automatically on deployment  
✅ **Conflict Resolution** - Separate version tracking for each migration system  
✅ **Guaranteed Order** - Migrations run in the correct order (PostgreSQL → ClickHouse → Drizzle)  
✅ **Health Checks** - Service dependencies ensure databases are ready before migrations  
✅ **Easy Rollback** - Each system can be rolled back independently if needed  

---

## Architecture

### Service Dependency Order

```
Docker Compose Up
    ↓
[PostgreSQL, ClickHouse, Valkey] start
    ↓
[Migrations] service waits for healthy databases
    ↓
[Migrations] runs all migrations sequentially:
    1. PostgreSQL migrations (golang-migrate)
    2. ClickHouse migrations (golang-migrate)  
    3. Drizzle migrations (BetterAuth)
    ↓
[API, Frontend, Worker, Redirector] start after migrations complete
```

### Migration Systems & Version Tracking

| System | Tool | Database | Version Table | Location |
|--------|------|----------|--------------|----------|
| **API Schema** | golang-migrate | PostgreSQL | `schema_migrations` | `api/sql/schema/` |
| **ClickHouse** | golang-migrate | ClickHouse | `schema_migrations` | `api/clickhouse/schema/` |
| **BetterAuth** | Drizzle ORM | PostgreSQL | `__drizzle_migrations` | `frontend/drizzle/` |

**Important:** Each system maintains its own version tracking independently to prevent conflicts.

---

## Database Schema Details

### PostgreSQL (refract)

#### API Tables (golang-migrate)
- **urls** - Main URL shortening table
  - Columns: id (snowflake), short_code, original_url, user_id, status, created_at, updated_at, expires_at, title
  - Indexes: `idx_urls_active_short_code` (conditional), `idx_urls_user_id_created_at`
  - Migrations: 5 total (versions 000001-000005)

#### BetterAuth Tables (Drizzle)
- **user** - User accounts
- **session** - User sessions with FK to user
- **account** - OAuth/auth accounts with FK to user
- **verification** - Email verification codes
- **jwks** - JWT key storage
- **todos** - Application data
- Migrations: 3 total (versions 0000-0002)

#### Version Tracking Tables
- `schema_migrations` - Created by golang-migrate, tracks API & ClickHouse versions
- `__drizzle_migrations` - Created by Drizzle, tracks BetterAuth schema versions

### ClickHouse (refract)

- **clicks** - Click events with 30-day TTL
  - Columns: short_code, clicked_at, ip_address, user_agent, referer, date, hour
  - Engine: MergeTree, partitioned by date

- **url_daily_stats** - Pre-aggregated daily statistics
  - Columns: date, short_code, total_clicks, unique_ips (approximate via HyperLogLog)
  - Engine: AggregatingMergeTree

- **url_daily_stats_mv** - Materialized view
  - Automatically aggregates click data to daily stats

---

## File Structure

```
refract/
├── Dockerfile.migrations          # Migration container image
├── scripts/
│   ├── migrate.sh                 # Main orchestration script
│   └── validate-schema.sh         # Schema validation utility
├── docker-compose.yml             # Production with migrations service
├── docker-compose.dev.yml         # Development with migrations service
├── api/
│   ├── sql/schema/                # PostgreSQL migrations (golang-migrate)
│   │   ├── 000001_create_urls_table.up.sql
│   │   ├── 000002_add_domain_field.up.sql
│   │   ├── 000003_alter_urls_indexes.up.sql
│   │   ├── 000004_remove_domain.up.sql
│   │   └── 000005_add_title.up.sql
│   └── clickhouse/schema/         # ClickHouse migrations (golang-migrate)
│       ├── 000001_clicks_table.up.sql
│       ├── 000002_create_urls_table.up.sql
│       └── 000003_add_title.up.sql
└── frontend/
    ├── drizzle/                   # Drizzle migrations (Drizzle ORM)
    │   ├── 0000_fair_talisman.sql
    │   ├── 0001_stiff_toxin.sql
    │   └── 0002_graceful_bromley.sql
    ├── auth-schema.ts             # BetterAuth schema definition
    ├── drizzle.config.ts          # Drizzle configuration
    └── src/db/schema.ts           # App data schemas
```

---

## Running Migrations

### Automatic (Production & Development)

Migrations run automatically when you start the containers:

```bash
# Production
docker compose up

# Development
docker compose -f docker-compose.dev.yml up
```

The migrations service will:
1. Wait for PostgreSQL and ClickHouse to be healthy
2. Run all pending migrations in order
3. Exit with status 0 on success or non-zero on failure
4. Block other services from starting until complete

### Manual (Local Development)

You can still run migrations manually from your host machine:

```bash
# Install golang-migrate if needed
brew install golang-migrate  # or your OS equivalent

# Migrate PostgreSQL
migrate -path ./api/sql/schema \
  -database "postgresql://postgres:postgres@localhost:5432/refract?sslmode=disable" \
  up

# Migrate ClickHouse
migrate -path ./api/clickhouse/schema \
  -database "clickhouse://default@localhost:9000/refract" \
  up

# Migrate Drizzle (BetterAuth)
cd frontend && DATABASE_URL="postgresql://postgres:postgres@localhost:5432/refract" npm run db:push
```

### Validating Schema After Migrations

```bash
# In a running container
docker exec refract-migrations /usr/local/bin/validate-schema.sh

# Or run locally (requires psql and curl)
./scripts/validate-schema.sh
```

---

## Adding New Migrations

### PostgreSQL (API)

```bash
# Create a new migration
migrate create -ext sql -dir ./api/sql/schema -seq your_migration_name

# This creates:
# - api/sql/schema/000006_your_migration_name.up.sql
# - api/sql/schema/000006_your_migration_name.down.sql
```

### ClickHouse

```bash
# Create a new migration
migrate create -ext sql -dir ./api/clickhouse/schema -seq your_migration_name
```

### Drizzle (BetterAuth)

```bash
cd frontend

# Generate migration from schema changes
npm run db:generate -- --name your_migration_name

# Push to database immediately
npm run db:push
```

---

## Conflict Resolution Strategy

### Problem: Schema Split Ownership

Previously, both golang-migrate and Drizzle were managing the same PostgreSQL database, creating conflicts:
- golang-migrate tracked versions in `schema_migrations`
- Drizzle tracked versions in `drizzle_meta`
- No single source of truth for schema state

### Solution: Separate Version Tracking

Each migration system now maintains its own version tracking table:

1. **golang-migrate** (`schema_migrations`)
   - Tracks API migrations (000001-000005)
   - Tracks ClickHouse migrations (versions separate from API)
   - Managed independently

2. **Drizzle** (`__drizzle_migrations`)
   - Tracks BetterAuth schema changes (0000-0002)
   - Uses Drizzle's internal tracking
   - Managed independently

3. **Schema Ownership Map**
   - API Schema (urls table) → golang-migrate manages
   - ClickHouse Tables → golang-migrate manages
   - Auth Schema (user, session, account, etc.) → Drizzle manages
   - Version tracking fully separated by tool

### Benefits of This Approach

✅ No conflicts between migration tools  
✅ Can rollback each system independently  
✅ Clear separation of concerns  
✅ Each team can manage their schema independently  
✅ Easy to debug issues in specific system  

---

## Troubleshooting

### Migrations Fail to Start

**Check the logs:**
```bash
docker logs refract-migrations
```

**Verify database connectivity:**
```bash
# PostgreSQL
docker exec refract-postgres pg_isready -U postgres

# ClickHouse
curl http://localhost:8123/ping
```

### Service Refuses to Start

Make sure migrations completed successfully:
```bash
docker compose logs migrations
```

Services won't start if migrations service exits with non-zero status.

### Need to Rollback

**Rollback PostgreSQL migrations:**
```bash
migrate -path ./api/sql/schema \
  -database "postgresql://postgres:postgres@localhost:5432/refract" \
  down 1
```

**Rollback ClickHouse migrations:**
```bash
migrate -path ./api/clickhouse/schema \
  -database "clickhouse://default@localhost:9000/refract" \
  down 1
```

**Rollback Drizzle:**
```bash
cd frontend && npm run db:drop  # Careful! Drops all tables
```

### Schema Validation Fails

Run the validation script to identify issues:
```bash
docker exec refract-migrations /usr/local/bin/validate-schema.sh
```

Check that all required tables exist and have correct indexes.

---

## Environment Variables

Configure migrations using these environment variables in your `.env` file:

```bash
# PostgreSQL Configuration
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=refract
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/refract

# ClickHouse Configuration
CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_HTTP_PORT=8123
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
```

---

## Migration Execution Details

### Sequence of Operations

When `docker compose up` runs:

```
1. PostgreSQL container starts
   └─ Healthcheck: pg_isready
   
2. ClickHouse container starts
   └─ Healthcheck: HTTP /ping
   
3. Migrations service starts (waits for both databases healthy)
   ├─ Waits up to 60s for PostgreSQL
   ├─ Waits up to 60s for ClickHouse
   ├─ Runs: migrate -path ./api/sql/schema up
   ├─ Runs: migrate -path ./api/clickhouse/schema up
   ├─ Runs: npm run db:push (Drizzle)
   └─ Exits with status 0 on success
   
4. API service starts (waits for migrations + postgres healthy)
5. Frontend service starts (waits for migrations + api started)
6. Worker service starts (waits for migrations + valkey + clickhouse healthy)
7. Redirector service starts (waits for migrations + postgres healthy)
```

### Error Handling

If any migration fails:
- The migrations service exits with non-zero status
- Other services will not start (depends_on condition fails)
- Check `docker logs refract-migrations` for details
- Fix the issue and restart: `docker compose restart`

---

## Integration with Justfile

Update your Justfile tasks to use the new migrations:

```bash
# Old way (manual before deploy)
just migrate-all
  → just migrate-up
  → just migrate-ch-up
  → just migrate-frontend

# New way (automatic)
docker compose up  # Migrations run automatically
```

You can keep the Justfile tasks for local development, but production deployments will use the containerized migrations.

---

## FAQ

**Q: Can I still run migrations manually?**  
A: Yes, both methods work. Manual migrations are useful for local development, containers are for production consistency.

**Q: What if I add a new migration but don't deploy it yet?**  
A: It will run on next deployment. The migrations service checks for all pending migrations.

**Q: Can I rollback individual systems?**  
A: Yes, each migration system is independent. You can rollback PostgreSQL without affecting ClickHouse.

**Q: What happens if a migration is already applied?**  
A: golang-migrate and Drizzle track versions separately and won't re-apply migrations.

**Q: How do I verify all migrations ran?**  
A: Run `./scripts/validate-schema.sh` or check version tables:

```bash
# Check PostgreSQL migration version
psql -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;"

# Check Drizzle migration version  
psql -c "SELECT json_extract(value, '$.tag') FROM __drizzle_migrations LIMIT 1;"
```

---

## Support & Documentation

For more details:
- [golang-migrate docs](https://github.com/golang-migrate/migrate)
- [Drizzle ORM docs](https://orm.drizzle.team/)
- [ClickHouse docs](https://clickhouse.com/docs)

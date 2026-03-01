# Quick Start: Running Migrations

## Starting the Application

### Production
```bash
docker compose up
```
The migrations service will automatically run all migrations before app services start.

### Development
```bash
docker compose -f docker-compose.dev.yml up
```

## Key Changes

### What's New
- ✅ Migrations run automatically in a separate container
- ✅ All databases initialize before app services start
- ✅ No manual migration steps required on deployment
- ✅ Conflicts between migration systems resolved

### File Structure
```
Dockerfile.migrations           # New: Migration container image
scripts/
  ├── migrate.sh               # New: Orchestration script
  └── validate-schema.sh       # New: Validation utility
docker-compose.yml             # Updated: Added migrations service
docker-compose.dev.yml         # Updated: Added migrations service
docs/MIGRATIONS.md             # New: Full documentation
```

## Manual Commands (Local Development)

### View Migration Logs
```bash
docker logs refract-migrations
```

### Validate Schema After Migrations
```bash
docker exec refract-migrations /usr/local/bin/validate-schema.sh
```

### Run Migrations Manually (if needed)
```bash
# PostgreSQL
migrate -path ./api/sql/schema \
  -database "postgresql://postgres:postgres@localhost:5432/refract" up

# ClickHouse
migrate -path ./api/clickhouse/schema \
  -database "clickhouse://default@localhost:9000/refract" up

# Drizzle (BetterAuth)
cd frontend && DATABASE_URL="postgresql://postgres:postgres@localhost:5432/refract" npm run db:push
```

### Create New Migrations

**PostgreSQL:**
```bash
migrate create -ext sql -dir ./api/sql/schema -seq migration_name
```

**ClickHouse:**
```bash
migrate create -ext sql -dir ./api/clickhouse/schema -seq migration_name
```

**Drizzle:**
```bash
cd frontend && npm run db:generate -- --name migration_name
```

## Environment Variables

Add to `.env` if using non-default database values:

```bash
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=refract
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/refract

CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_HTTP_PORT=8123
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
```

## Service Startup Order

```
1. PostgreSQL, ClickHouse, Valkey start
2. Migrations service waits for databases to be healthy
3. Migrations service runs all pending migrations
4. API, Frontend, Worker, Redirector start (after migrations complete)
```

## Troubleshooting

### Migrations Failed
```bash
docker logs refract-migrations
```

### Database Connection Issues
```bash
# Check PostgreSQL
docker exec refract-postgres pg_isready -U postgres

# Check ClickHouse
curl http://localhost:8123/ping
```

### Need Fresh Database
```bash
docker compose down -v  # Remove all volumes
docker compose up       # Start fresh
```

## Additional Resources

- Full documentation: `docs/MIGRATIONS.md`
- Validation script: `scripts/validate-schema.sh`
- Orchestration script: `scripts/migrate.sh`

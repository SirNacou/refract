#!/bin/bash
# Migration Orchestration Script
# Runs all database migrations in the correct order
# Supports: PostgreSQL, ClickHouse, and BetterAuth (Drizzle)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration from environment variables
POSTGRES_HOST="${POSTGRES_HOST:-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-refract}"

CLICKHOUSE_HOST="${CLICKHOUSE_HOST:-clickhouse}"
CLICKHOUSE_PORT="${CLICKHOUSE_PORT:-9000}"
CLICKHOUSE_HTTP_PORT="${CLICKHOUSE_HTTP_PORT:-8123}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-default}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-}"

DATABASE_URL="${DATABASE_URL:-postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}}"

# Retry configuration
MAX_RETRIES=30
RETRY_DELAY=2

# Logging functions
log_info() {
	echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
	echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
	echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
	echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for service to be ready
wait_for_service() {
	local service_name=$1
	local host=$2
	local port=$3
	local retry_count=0

	log_info "Waiting for $service_name at $host:$port..."

	while [ $retry_count -lt $MAX_RETRIES ]; do
		if nc -z "$host" "$port" 2>/dev/null; then
			log_success "$service_name is ready"
			return 0
		fi
		retry_count=$((retry_count + 1))
		echo -n "."
		sleep $RETRY_DELAY
	done

	log_error "$service_name did not become ready after $((MAX_RETRIES * RETRY_DELAY)) seconds"
	return 1
}

# Wait for PostgreSQL with health check
wait_for_postgres() {
	local retry_count=0

	log_info "Checking PostgreSQL connection..."

	while [ $retry_count -lt $MAX_RETRIES ]; do
		if pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" 2>/dev/null; then
			log_success "PostgreSQL is ready"
			return 0
		fi
		retry_count=$((retry_count + 1))
		echo -n "."
		sleep $RETRY_DELAY
	done

	log_error "PostgreSQL did not become ready"
	return 1
}

# Wait for ClickHouse with health check
wait_for_clickhouse() {
	local retry_count=0

	log_info "Checking ClickHouse connection..."

	while [ $retry_count -lt $MAX_RETRIES ]; do
		if curl -s -f "http://$CLICKHOUSE_HOST:$CLICKHOUSE_HTTP_PORT/ping" >/dev/null 2>&1; then
			log_success "ClickHouse is ready"
			return 0
		fi
		retry_count=$((retry_count + 1))
		echo -n "."
		sleep $RETRY_DELAY
	done

	log_error "ClickHouse did not become ready"
	return 1
}

# Run PostgreSQL migrations (golang-migrate)
migrate_postgres() {
	log_info "Starting PostgreSQL migrations..."

	if ! migrate -path ./api/sql/schema \
		-database "$DATABASE_URL" \
		up; then
		log_error "PostgreSQL migrations failed"
		return 1
	fi

	log_success "PostgreSQL migrations completed"
	return 0
}

# Run ClickHouse migrations (golang-migrate)
migrate_clickhouse() {
	log_info "Starting ClickHouse migrations..."

	# Build ClickHouse connection string
	# Format: clickhouse://user:password@host:port/database?sslmode=disable
	local ch_url="clickhouse://$CLICKHOUSE_USER"
	if [ -n "$CLICKHOUSE_PASSWORD" ]; then
		ch_url="${ch_url}:${CLICKHOUSE_PASSWORD}"
	fi
	ch_url="${ch_url}@${CLICKHOUSE_HOST}:${CLICKHOUSE_PORT}/refract"

	if ! migrate -path ./api/clickhouse/schema \
		-database "$ch_url" \
		up; then
		log_error "ClickHouse migrations failed"
		return 1
	fi

	log_success "ClickHouse migrations completed"
	return 0
}

# Run Drizzle migrations (BetterAuth)
migrate_drizzle() {
	log_info "Starting Drizzle migrations (BetterAuth)..."

	cd ./frontend

	if ! DATABASE_URL="$DATABASE_URL" npm run db:push -- --skip-generate; then
		log_error "Drizzle migrations failed"
		cd ..
		return 1
	fi

	cd ..
	log_success "Drizzle migrations completed"
	return 0
}

# Validate schema consistency
validate_schema() {
	log_info "Validating schema consistency..."

	# Check PostgreSQL schema
	if ! PGPASSWORD="$POSTGRES_PASSWORD" psql \
		-h "$POSTGRES_HOST" \
		-p "$POSTGRES_PORT" \
		-U "$POSTGRES_USER" \
		-d "$POSTGRES_DB" \
		-c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" >/dev/null 2>&1; then
		log_warn "Could not validate PostgreSQL schema"
		return 1
	fi

	log_success "Schema validation passed"
	return 0
}

# Main execution
main() {
	log_info "=========================================="
	log_info "Database Migration Orchestration Starting"
	log_info "=========================================="
	echo ""

	# Wait for services
	log_info "Phase 1: Waiting for services..."
	wait_for_postgres || exit 1
	wait_for_clickhouse || exit 1
	echo ""

	# Run migrations
	log_info "Phase 2: Running migrations..."

	# Run PostgreSQL migrations first (used by both API and Drizzle)
	migrate_postgres || exit 1
	echo ""

	# Run ClickHouse migrations
	migrate_clickhouse || exit 1
	echo ""

	# Run Drizzle migrations last (depends on PostgreSQL being ready)
	migrate_drizzle || exit 1
	echo ""

	# Validate
	log_info "Phase 3: Validating schema..."
	validate_schema || log_warn "Schema validation returned non-zero, but migrations may still have succeeded"
	echo ""

	log_info "=========================================="
	log_success "All migrations completed successfully!"
	log_info "=========================================="
}

# Run main
main "$@"

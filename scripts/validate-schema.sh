#!/bin/bash
# Schema Validation Script
# Verifies that all database schemas are properly initialized

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Logging functions
log_info() {
	echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
	echo -e "${GREEN}[✓]${NC} $1"
}

log_error() {
	echo -e "${RED}[✗]${NC} $1"
}

log_warn() {
	echo -e "${YELLOW}[!]${NC} $1"
}

# Check PostgreSQL schema
check_postgres_schema() {
	log_info "Checking PostgreSQL schema..."

	# Check if urls table exists
	if PGPASSWORD="$POSTGRES_PASSWORD" psql \
		-h "$POSTGRES_HOST" \
		-p "$POSTGRES_PORT" \
		-U "$POSTGRES_USER" \
		-d "$POSTGRES_DB" \
		-c "SELECT 1 FROM information_schema.tables WHERE table_name = 'urls' AND table_schema = 'public';" 2>/dev/null | grep -q 1; then
		log_success "PostgreSQL urls table exists"
	else
		log_error "PostgreSQL urls table NOT FOUND"
		return 1
	fi

	# Check if BetterAuth tables exist
	for table in "user" "session" "account" "verification"; do
		if PGPASSWORD="$POSTGRES_PASSWORD" psql \
			-h "$POSTGRES_HOST" \
			-p "$POSTGRES_PORT" \
			-U "$POSTGRES_USER" \
			-d "$POSTGRES_DB" \
			-c "SELECT 1 FROM information_schema.tables WHERE table_name = '$table' AND table_schema = 'public';" 2>/dev/null | grep -q 1; then
			log_success "PostgreSQL $table table exists"
		else
			log_error "PostgreSQL $table table NOT FOUND"
			return 1
		fi
	done

	# Check schema_migrations table (golang-migrate)
	if PGPASSWORD="$POSTGRES_PASSWORD" psql \
		-h "$POSTGRES_HOST" \
		-p "$POSTGRES_PORT" \
		-U "$POSTGRES_USER" \
		-d "$POSTGRES_DB" \
		-c "SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations' AND table_schema = 'public';" 2>/dev/null | grep -q 1; then
		log_success "PostgreSQL schema_migrations table exists (golang-migrate)"
	else
		log_warn "PostgreSQL schema_migrations table NOT FOUND (golang-migrate may not have run)"
	fi

	# Check drizzle_meta table
	if PGPASSWORD="$POSTGRES_PASSWORD" psql \
		-h "$POSTGRES_HOST" \
		-p "$POSTGRES_PORT" \
		-U "$POSTGRES_USER" \
		-d "$POSTGRES_DB" \
		-c "SELECT 1 FROM information_schema.tables WHERE table_name = '__drizzle_migrations' AND table_schema = 'public';" 2>/dev/null | grep -q 1; then
		log_success "PostgreSQL __drizzle_migrations table exists (Drizzle)"
	else
		log_warn "PostgreSQL __drizzle_migrations table NOT FOUND (Drizzle may not have run)"
	fi

	return 0
}

# Check ClickHouse schema
check_clickhouse_schema() {
	log_info "Checking ClickHouse schema..."

	# Check if database exists
	if curl -s "http://$CLICKHOUSE_HOST:$CLICKHOUSE_HTTP_PORT/?query=SHOW%20DATABASES%20LIKE%20%27refract%27" | grep -q "refract"; then
		log_success "ClickHouse refract database exists"
	else
		log_error "ClickHouse refract database NOT FOUND"
		return 1
	fi

	# Check if clicks table exists
	if curl -s "http://$CLICKHOUSE_HOST:$CLICKHOUSE_HTTP_PORT/?query=EXISTS%20refract.clicks" | grep -q "1"; then
		log_success "ClickHouse clicks table exists"
	else
		log_error "ClickHouse clicks table NOT FOUND"
		return 1
	fi

	# Check if url_daily_stats table exists
	if curl -s "http://$CLICKHOUSE_HOST:$CLICKHOUSE_HTTP_PORT/?query=EXISTS%20refract.url_daily_stats" | grep -q "1"; then
		log_success "ClickHouse url_daily_stats table exists"
	else
		log_error "ClickHouse url_daily_stats table NOT FOUND"
		return 1
	fi

	return 0
}

# Check migration version tracking
check_migration_versions() {
	log_info "Checking migration versions..."

	# Check PostgreSQL migrations
	if PGPASSWORD="$POSTGRES_PASSWORD" psql \
		-h "$POSTGRES_HOST" \
		-p "$POSTGRES_PORT" \
		-U "$POSTGRES_USER" \
		-d "$POSTGRES_DB" \
		-c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null | tail -1 | grep -E '[0-9]+' >/dev/null; then
		version=$(PGPASSWORD="$POSTGRES_PASSWORD" psql \
			-h "$POSTGRES_HOST" \
			-p "$POSTGRES_PORT" \
			-U "$POSTGRES_USER" \
			-d "$POSTGRES_DB" \
			-t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null)
		log_success "PostgreSQL migrations: version $version"
	else
		log_warn "Could not read PostgreSQL migration versions"
	fi
}

# Main execution
main() {
	log_info "=========================================="
	log_info "Database Schema Validation"
	log_info "=========================================="
	echo ""

	local errors=0

	check_postgres_schema || ((errors++))
	echo ""

	check_clickhouse_schema || ((errors++))
	echo ""

	check_migration_versions
	echo ""

	if [ $errors -eq 0 ]; then
		log_info "=========================================="
		log_success "All schema validations passed!"
		log_info "=========================================="
		return 0
	else
		log_info "=========================================="
		log_error "Schema validation failed with $errors error(s)"
		log_info "=========================================="
		return 1
	fi
}

main "$@"

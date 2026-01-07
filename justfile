# List available commands
default:
    @just --list

# Start all services
up:
    docker-compose up -d

# Stop all services
down:
    docker-compose down

# View logs from all services
logs:
    docker-compose logs -f

# View API logs only
api-logs:
    docker-compose logs -f api

# View migration logs
migration-logs:
    docker-compose logs migrations

# Rebuild and restart services
rebuild:
    docker-compose up -d --build

# Stop services and remove volumes (DESTRUCTIVE)
clean:
    docker-compose down -v

# Run database migrations manually
migrate-up:
    docker-compose run --rm migrations

# Create a new migration file
migrate-create NAME:
    goose -dir migrations/postgres create {{NAME}} sql

# Check migration status
migrate-status:
    docker-compose exec postgres psql -U postgres -d refract -c "SELECT * FROM goose_db_version;"

# Connect to PostgreSQL
db-shell:
    docker-compose exec postgres psql -U postgres -d refract

# Connect to Valkey CLI
valkey-shell:
    docker-compose exec valkey valkey-cli -a "${VALKEY_PASSWORD:-valkey}"

# View Valkey logs
valkey-logs:
    docker-compose logs -f valkey

# Show Valkey info and stats
valkey-info:
    docker-compose exec valkey valkey-cli -a "${VALKEY_PASSWORD:-valkey}" INFO

# Run API tests (when you have them)
test:
    cd services/api && go test ./...

# Format Go code
fmt:
    cd services/api && go fmt ./...

# Run Go linter (requires golangci-lint)
lint:
    cd services/api && golangci-lint run

# Install development dependencies
install-deps:
    go install github.com/pressly/goose/v3/cmd/goose@latest

# Generate SQLc code from SQL queries
sqlc-generate:
    docker run --rm -v "$(pwd):/src" -w /src/services/api sqlc/sqlc generate

# Show service status
status:
    docker-compose ps

# Restart API service only
restart-api:
    docker-compose restart api

# View API service logs in real-time
tail-api:
    docker-compose logs -f --tail=100 api

# View Postgres logs
postgres-logs:
    docker-compose logs -f postgres

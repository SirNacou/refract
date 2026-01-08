# List available commands
default:
    @just --list

# ========================================
# Docker Services
# ========================================

# Start all services (rebuilds to pick up code changes)
up:
    docker-compose up -d --build

# Stop all services
down:
    docker-compose down

# Restart all services
restart:
    docker-compose restart

# View all service logs
logs:
    docker-compose logs -f

# Show service status
ps:
    docker-compose ps

# Stop and remove all data (DESTRUCTIVE)
clean:
    docker-compose down -v

# ========================================
# Development Workflow
# ========================================

# Quick rebuild and restart API with logs
dev:
    docker-compose up -d --build --no-deps api
    @echo "✅ API restarted. Showing logs (Ctrl+C to exit):"
    docker-compose logs -f --tail=50 api

# Reset everything and start fresh (DESTRUCTIVE)
reset:
    docker-compose down -v
    docker-compose up -d --build
    @echo "✅ Fresh start complete."

# ========================================
# Database
# ========================================

# Run migrations
migrate:
    docker-compose run --rm migrations

# Create new migration file
migrate-new NAME:
    goose -dir migrations/postgres create {{NAME}} sql

# Connect to database shell
db:
    docker-compose exec postgres psql -U postgres -d refract

# View database logs
db-logs:
    docker-compose logs -f postgres

# ========================================
# Valkey (Redis)
# ========================================

# Connect to Valkey CLI
valkey:
    docker-compose exec valkey valkey-cli -a "${VALKEY_PASSWORD:-valkey}"

# Show Valkey info
valkey-info:
    docker-compose exec valkey valkey-cli -a "${VALKEY_PASSWORD:-valkey}" INFO

# View Valkey logs
valkey-logs:
    docker-compose logs -f valkey

# ========================================
# Go Development
# ========================================

# Run tests
test:
    cd services/api && go test ./...

# Format code
fmt:
    cd services/api && go fmt ./...

# Run linter
lint:
    cd services/api && golangci-lint run

# Generate SQLc code
sqlc:
    docker run --rm -v "$(pwd):/src" -w /src/services/api sqlc/sqlc generate

# Install dev dependencies
deps:
    go install github.com/pressly/goose/v3/cmd/goose@latest

# ========================================
# Redirector Service
# ========================================

# View redirector logs
redirector-logs:
    docker-compose logs -f redirector

# Restart redirector service
restart-redirector:
    docker-compose restart redirector

# Rebuild and restart redirector (fast dev cycle)
dev-redirector:
    docker-compose up -d --build --no-deps redirector
    @echo "✅ Redirector restarted. Showing logs (Ctrl+C to exit):"
    docker-compose logs -f --tail=50 redirector

# Check redirector health
redirector-health:
    curl -s http://localhost:8081/health | jq

# Development helper tasks for this project
# Usage:
#   just up        -> bring up docker-compose services (detached, rebuild)
#   just down      -> stop and remove compose services
#   just migrate   -> run DB migrations (uses `migrate` CLI; will attempt to install it if missing)
#   just test      -> run Go, frontend (npm) and workspace tests when available

set shell := ["bash", "-euxo", "pipefail", "-c"]

# Default DATABASE_URL if not set by env. Override by setting DATABASE_URL in your environment.
DATABASE_URL := ${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/refract?sslmode=disable}

# Start services
up:
	@echo "Starting services with docker compose..."
	@docker compose up -d --build

# Stop services
down:
	@echo "Stopping services..."
	@docker compose down

# Run database migrations
migrate:
	@echo "Running migrations (migrations/postgres)..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations/postgres -database "$DATABASE_URL" up; \
	else \
		echo "migrate CLI not found; installing via 'go install'..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		migrate -path migrations/postgres -database "$DATABASE_URL" up; \
	fi

# Run tests (Go + frontend + Rust when available)
test:
	@echo "Running Go tests..."
	@go test ./...
	@if [ -d frontend ]; then \
		echo "Running frontend tests (frontend)..."; \
		cd frontend; \
		npm test; \
	fi
	@if command -v cargo >/dev/null 2>&1; then \
		echo "Running Rust workspace tests (if any)..."; \
		cargo test --workspace; \
	fi

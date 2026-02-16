set shell := ["bash", "-c"]
set windows-shell := ["powershell", "-NoLogo", "-Command"]

database_url := "postgres://postgres:postgres@localhost:5432/refract?sslmode=disable"
ch_host := "localhost"
ch_port := "9000"
ch_user := "default"
ch_password := "default"
ch_database := "refract"
ch_database_url := "clickhouse://" + ch_host + ":" + ch_port + "?username=" + ch_user + "'&'password=" + ch_password + "'&'database=" + ch_database + "'&'x-multi-statement=true"

default:
    @echo "Available commands:"
    @echo "  dev-up                Start development environment"
    @echo "  dev-down              Stop development environment"
    @echo "  create-migration name Create a new database migration"
    @echo "  migrate name         Apply or revert database migrations"
    @echo "  migrate-all          Run all migrations (api, clickhouse, frontend)"
    @echo "  generate             Generate code from SQL queries"

dev-up name:
    @docker compose -f docker-compose.dev.yml up --build -d {{ name }}

dev-watch:
    @docker compose -f docker-compose.dev.yml up --build --watch

dev-down:
    @docker compose -f docker-compose.dev.yml down

create-migration name:
    @echo "Creating new migration: {{ name }}"
    @migrate create --ext sql --dir ./api/sql/schema -seq {{ name }}

create-ch-migration name:
    @echo "Creating new ClickHouse migration: {{ name }}"
    @migrate create --ext sql --dir ./api/clickhouse/schema -seq {{ name }}

migrate name:
    @migrate -path ./api/sql/schema -database {{ database_url }} {{ name }}

migrate-ch name:
    @migrate -path ./api/clickhouse/schema -database {{ ch_database_url }} {{ name }}

migrate-all:
    @echo "Running all migrations..."
    @just migrate-up
    @just migrate-ch-up
    @just migrate-frontend

migrate-up:
    @echo "Running API PostgreSQL migrations..."
    @migrate -path ./api/sql/schema -database {{ database_url }} up

migrate-ch-up:
    @echo "Running ClickHouse migrations..."
    @migrate -path ./api/clickhouse/schema -database {{ ch_database_url }} up

migrate-frontend:
    @echo "Running Frontend Drizzle migrations..."
    @cd frontend && DATABASE_URL={{ database_url }} bunx drizzle-kit push

generate:
    @sqlc generate -f ./api/sqlc.yaml

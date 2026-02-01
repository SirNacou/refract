set shell := ["bash", "-c"]
set windows-shell := ["powershell", "-NoLogo", "-Command"]

database_url := "postgres://postgres:postgres@localhost:5432/refract?sslmode=disable"
ch_database_url := 'clickhouse://localhost:9000?username=default"&"password=default"&"database=refract'

default:
  @echo "Available commands:"
  @echo "  dev-up                Start development environment"
  @echo "  dev-down              Stop development environment"
  @echo "  create-migration name Create a new database migration"
  @echo "  migrate name         Apply or revert database migrations"
  @echo "  generate             Generate code from SQL queries" 

dev-up:
  @docker compose -f docker-compose.dev.yml up --build --watch

dev-down:
  @docker compose -f docker-compose.dev.yml down

create-migration name:
  @echo "Creating new migration: {{name}}"
  @migrate create --ext sql --dir ./api/sql/schema -seq {{name}}

create-ch-migration name:
  @echo "Creating new migration: {{name}}"
  @migrate create --ext sql --dir ./api/clickhouse/schema -seq {{name}}

migrate name:
  @migrate -path ./api/sql/schema -database {{database_url}} {{name}}

migrate-ch name:
  @migrate -path ./api/clickhouse/schema -database {{ch_database_url}} {{name}}

generate:
  @sqlc generate -f ./api/sqlc.yaml
set shell := ["bash", "-c"]

database_url := "postgres://postgres:postgres@localhost:5432/refract?sslmode=disable"

dev-up:
  @docker compose -f docker-compose.dev.yml up --build -d

dev-down:
  @docker compose -f docker-compose.dev.yml down

create-migration name:
  @echo "Creating new migration: {{name}}"
  @migrate create --ext sql --dir ./api/sql/schema -seq {{name}}

migrate name:
  @migrate -path ./api/sql/schema -database {{database_url}} {{name}}

generate:
  @sqlc generate -f ./api/sqlc.yaml
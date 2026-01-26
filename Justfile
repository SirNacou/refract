set shell := ["bash", "-c"]

dev-up:
  @docker compose -f docker-compose.dev.yml up --build -d

dev-down:
  @docker compose -f docker-compose.dev.yml down
FROM rust:1-alpine AS builder

WORKDIR /app

COPY Cargo.toml ./services/redirector/
COPY --from=root Cargo.toml Cargo.lock ./

WORKDIR /app/services/redirector

RUN mkdir src && touch src/lib.rs
RUN cargo build --release

# Copy actual source code and build
COPY . .
RUN cargo build --release

# Stage 2: Runtime
FROM debian:bookworm-slim
# Or alpine:latest, depending on glibc/musl requirements
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/redirector /usr/local/bin/redirector

# Use a non-root user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

CMD ["redirector"]
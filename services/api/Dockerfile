# Build stage
FROM golang:1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files from services/api
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

# Copy source code
COPY services/api/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /api ./api
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8080

CMD ["./api"]

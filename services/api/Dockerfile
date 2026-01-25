# Stage 1: Build the Go application
FROM golang:1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myapp ./cmd/api/*

# Stage 2: Create a minimal runtime image
# Use 'scratch' for the smallest possible image
# For apps requiring SSL certs, consider 'distroless/static' or a minimal Linux base
FROM scratch AS final

# Copy the compiled binary from the builder stage
COPY --from=builder /app/myapp .

# Command to run the executable when the container starts
CMD ["./myapp"]

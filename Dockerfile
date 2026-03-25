# Build stage
FROM golang:1.24-bookworm AS builder

# Install build dependencies for sqlite3 (CGO)
RUN apt-get update && apt-get install -y gcc libc6-dev

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build self-contained binary
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o xtmator ./cmd/xtmator/main.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/xtmator .

# Expose the app port
EXPOSE 3333

# Run the app (database will be created in /app/xtmator.db)
CMD ["./xtmator"]

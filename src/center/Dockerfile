# OFA Center Dockerfile
# Multi-stage build for optimized image

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o center ./cmd/center

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/center .

# Copy config files
COPY configs/ ./configs/

# Create data directory for SQLite
RUN mkdir -p /app/data

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["./center"]
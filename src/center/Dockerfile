# OFA Center Dockerfile
# Multi-stage build for production deployment
# Version: v8.4.0

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o center ./cmd/center

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Copy binary from builder
COPY --from=builder /build/center .
COPY --from=builder /build/configs ./configs

# Create data directory
RUN mkdir -p /data

# Expose ports
EXPOSE 8080 9090 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/status || exit 1

# Run
ENTRYPOINT ["./center"]
CMD ["-c", "configs/center.yaml"]

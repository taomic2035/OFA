# OFA Go Agent Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o agent ./cmd/agent

# Final image
FROM alpine:3.18

WORKDIR /app

# Copy binary
COPY --from=builder /app/agent .

# Run
CMD ["./agent"]
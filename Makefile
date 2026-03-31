# OFA Makefile

.PHONY: all build test clean proto docker

# Default target
all: build

# Build Center
build-center:
	cd src/center && go build -o ../build/center ./cmd/center

# Build Agent (Go)
build-agent:
	cd src/agent/go && go build -o ../../build/agent ./cmd/agent

# Build all
build: build-center build-agent

# Run tests
test-center:
	cd src/center && go test ./...

test-agent:
	cd src/agent/go && go test ./...

test: test-center test-agent

# Generate protobuf files
proto:
	bash scripts/gen_proto.sh

# Clean build artifacts
clean:
	rm -rf build/*

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development
dev-center:
	cd src/center && go run ./cmd/center

dev-agent:
	cd src/agent/go && go run ./cmd/agent

# Initialize database
init-db:
	docker exec -it ofa-postgres psql -U ofa -d ofa -c "CREATE TABLE IF NOT EXISTS agents (id VARCHAR(64) PRIMARY KEY, name VARCHAR(255), type INTEGER, status INTEGER, last_seen TIMESTAMP);"

# Format code
fmt:
	cd src/center && go fmt ./...
	cd src/agent/go && go fmt ./...

# Lint
lint:
	cd src/center && golangci-lint run
	cd src/agent/go && golangci-lint run
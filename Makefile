# OFA Makefile
# 构建和部署命令集合

.PHONY: all build test clean proto docker docker-build docker-push deploy-compose deploy-k8s stop-compose stop-k8s status-k8s logs-k8s bench fmt lint help

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# 默认目标
all: build test

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

# Docker push
docker-push:
	docker push ofa/center:$(VERSION)
	docker push ofa/center:latest

# Docker Compose 部署
deploy-compose:
	cd deployments && docker-compose up -d
	@echo "Waiting for services to start..."
	@sleep 10
	cd deployments && docker-compose ps

stop-compose:
	cd deployments && docker-compose down

# Kubernetes 部署
deploy-k8s:
	kubectl apply -f deployments/kubernetes.yaml
	kubectl rollout status deployment/center -n ofa

stop-k8s:
	kubectl delete -f deployments/kubernetes.yaml

status-k8s:
	kubectl get all -n ofa
	kubectl get pods -n ofa

logs-k8s:
	kubectl logs -f deployment/center -n ofa

# 性能基准测试
bench:
	cd src/center && go test ./... -bench=. -benchmem

# 帮助信息
help:
	@echo "OFA Makefile Commands:"
	@echo ""
	@echo "Build:"
	@echo "  build           Build Center and Agent binaries"
	@echo "  build-center    Build Center only"
	@echo "  build-agent     Build Agent only"
	@echo ""
	@echo "Test:"
	@echo "  test            Run all tests"
	@echo "  test-center     Run Center tests"
	@echo "  test-agent      Run Agent tests"
	@echo "  bench           Run performance benchmarks"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build    Build Docker images"
	@echo "  docker-push     Push Docker images"
	@echo "  docker-up       Start Docker Compose"
	@echo "  docker-down     Stop Docker Compose"
	@echo "  docker-logs     Show Docker logs"
	@echo ""
	@echo "Deploy:"
	@echo "  deploy-compose  Deploy with Docker Compose"
	@echo "  stop-compose    Stop Docker Compose deployment"
	@echo "  deploy-k8s      Deploy to Kubernetes"
	@echo "  stop-k8s        Stop Kubernetes deployment"
	@echo "  status-k8s      Show Kubernetes status"
	@echo "  logs-k8s        Show Center logs in Kubernetes"
	@echo ""
	@echo "Development:"
	@echo "  dev-center      Run Center in dev mode"
	@echo "  dev-agent       Run Agent in dev mode"
	@echo "  fmt             Format code"
	@echo "  lint            Run linters"
	@echo "  proto           Generate protobuf files"
	@echo "  init-db         Initialize database tables"
#!/bin/bash

# OFA 快速启动脚本

set -e

echo "=========================================="
echo "  OFA - Omni Federated Agents"
echo "=========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印帮助
print_help() {
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start       启动所有服务"
    echo "  stop        停止所有服务"
    echo "  restart     重启所有服务"
    echo "  status      查看服务状态"
    echo "  logs        查看日志"
    echo "  build       构建服务"
    echo "  test        运行测试"
    echo "  clean       清理构建产物"
    echo "  help        显示帮助信息"
}

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}检查依赖...${NC}"

    # 检查 Docker
    if command -v docker &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} Docker 已安装"
    else
        echo -e "  ${RED}✗${NC} Docker 未安装"
        exit 1
    fi

    # 检查 Docker Compose
    if command -v docker-compose &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} Docker Compose 已安装"
    else
        echo -e "  ${RED}✗${NC} Docker Compose 未安装"
        exit 1
    fi

    echo ""
}

# 启动服务
start_services() {
    echo -e "${YELLOW}启动服务...${NC}"
    docker-compose -f deployments/docker-compose.yaml up -d
    echo -e "${GREEN}服务已启动${NC}"
    echo ""
    echo "服务地址:"
    echo "  - REST API: http://localhost:8080"
    echo "  - gRPC API: localhost:9090"
    echo ""
    echo "使用 '$0 logs' 查看日志"
}

# 停止服务
stop_services() {
    echo -e "${YELLOW}停止服务...${NC}"
    docker-compose -f deployments/docker-compose.yaml down
    echo -e "${GREEN}服务已停止${NC}"
}

# 重启服务
restart_services() {
    stop_services
    start_services
}

# 查看状态
check_status() {
    echo -e "${YELLOW}服务状态:${NC}"
    docker-compose -f deployments/docker-compose.yaml ps
}

# 查看日志
view_logs() {
    docker-compose -f deployments/docker-compose.yaml logs -f
}

# 构建
build_services() {
    echo -e "${YELLOW}构建服务...${NC}"

    # 检查 Go
    if command -v go &> /dev/null; then
        echo "构建 Center..."
        cd src/center && go build -o ../../build/center ./cmd/center && cd ../..

        echo "构建 Agent..."
        cd src/agent/go && go build -o ../../../build/agent ./cmd/agent && cd ../../..

        echo -e "${GREEN}构建完成${NC}"
    else
        echo -e "${RED}Go 未安装，跳过本地构建${NC}"
        echo "使用 Docker 构建..."
        docker-compose -f deployments/docker-compose.yaml build
    fi
}

# 运行测试
run_tests() {
    echo -e "${YELLOW}运行测试...${NC}"

    if command -v go &> /dev/null; then
        echo "Center 测试..."
        cd src/center && go test -v ./... && cd ../..

        echo "Agent 测试..."
        cd src/agent/go && go test -v ./... && cd ../../..

        echo -e "${GREEN}测试完成${NC}"
    else
        echo -e "${RED}Go 未安装${NC}"
        exit 1
    fi
}

# 清理
clean_up() {
    echo -e "${YELLOW}清理构建产物...${NC}"
    rm -rf build/*
    echo -e "${GREEN}清理完成${NC}"
}

# 主逻辑
case "${1:-help}" in
    start)
        check_dependencies
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    status)
        check_status
        ;;
    logs)
        view_logs
        ;;
    build)
        build_services
        ;;
    test)
        run_tests
        ;;
    clean)
        clean_up
        ;;
    help|*)
        print_help
        ;;
esac
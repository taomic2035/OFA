#!/bin/bash
# OFA Center 部署脚本
# 支持 Docker Compose 和 Kubernetes 部署

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
DEPLOY_MODE="${1:-compose}"  # compose or k8s
VERSION="${2:-latest}"
NAMESPACE="ofa"

echo -e "${GREEN}OFA Center Deployment Script${NC}"
echo "Deploy mode: ${DEPLOY_MODE}"
echo "Version: ${VERSION}"
echo ""

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"

    if [ "${DEPLOY_MODE}" = "compose" ]; then
        if ! command -v docker &> /dev/null; then
            echo -e "${RED}Error: Docker is required for compose deployment${NC}"
            exit 1
        fi
        if ! command -v docker-compose &> /dev/null; then
            echo -e "${RED}Error: docker-compose is required${NC}"
            exit 1
        fi
    elif [ "${DEPLOY_MODE}" = "k8s" ]; then
        if ! command -v kubectl &> /dev/null; then
            echo -e "${RED}Error: kubectl is required for Kubernetes deployment${NC}"
            exit 1
        fi
    fi

    echo -e "${GREEN}Dependencies OK${NC}"
}

# 构建 Docker 镜像
build_images() {
    echo -e "${YELLOW}Building Docker images...${NC}"

    cd src/center
    docker build -t ofa/center:${VERSION} -t ofa/center:latest .
    cd ../..

    echo -e "${GREEN}Images built successfully${NC}"
}

# Docker Compose 部署
deploy_compose() {
    echo -e "${YELLOW}Deploying with Docker Compose...${NC}"

    cd deployments

    # 停止现有部署
    docker-compose down --remove-orphans 2>/dev/null || true

    # 启动新部署
    docker-compose up -d

    # 等待服务就绪
    echo "Waiting for services to be ready..."
    sleep 15

    # 检查服务状态
    docker-compose ps

    cd ..

    echo -e "${GREEN}Docker Compose deployment completed${NC}"
    echo -e "${GREEN}Access Center at http://localhost:8080${NC}"
}

# Kubernetes 部署
deploy_k8s() {
    echo -e "${YELLOW}Deploying to Kubernetes...${NC}"

    # 创建命名空间（如果不存在）
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

    # 应用配置
    kubectl apply -f deployments/kubernetes.yaml

    # 等待部署就绪
    echo "Waiting for deployments to be ready..."
    kubectl rollout status deployment/center -n ${NAMESPACE} --timeout=300s
    kubectl rollout status deployment/postgres -n ${NAMESPACE} --timeout=120s
    kubectl rollout status deployment/redis -n ${NAMESPACE} --timeout=60s

    # 显示状态
    echo -e "${GREEN}Kubernetes deployment completed${NC}"
    kubectl get pods -n ${NAMESPACE}
    kubectl get services -n ${NAMESPACE}
}

# 显示帮助信息
show_help() {
    echo "Usage: ./deploy.sh [mode] [version]"
    echo ""
    echo "Modes:"
    echo "  compose  - Deploy with Docker Compose (default)"
    echo "  k8s      - Deploy to Kubernetes"
    echo ""
    echo "Examples:"
    echo "  ./deploy.sh compose"
    echo "  ./deploy.sh k8s v1.0.0"
    echo "  ./deploy.sh compose latest"
}

# 主流程
case "${DEPLOY_MODE}" in
    compose)
        check_dependencies
        build_images
        deploy_compose
        ;;
    k8s)
        check_dependencies
        build_images
        deploy_k8s
        ;;
    help|--help|-h)
        show_help
        exit 0
        ;;
    *)
        echo -e "${RED}Unknown deploy mode: ${DEPLOY_MODE}${NC}"
        show_help
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Deployment completed!${NC}"
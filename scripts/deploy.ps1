# OFA Center 部署脚本 (Windows PowerShell)
# 支持 Docker Compose 和 Kubernetes 部署

param(
    [Parameter(Position=0)]
    [string]$Mode = "compose",  # compose or k8s

    [Parameter(Position=1)]
    [string]$Version = "latest",

    [string]$Namespace = "ofa"
)

# 颜色输出函数
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Success "OFA Center Deployment Script"
Write-Host "Deploy mode: $Mode"
Write-Host "Version: $Version"
Write-Host ""

# 检查依赖
function Check-Dependencies {
    Write-Warning "Checking dependencies..."

    if ($Mode -eq "compose") {
        if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
            Write-Error "Error: Docker is required for compose deployment"
            exit 1
        }
        if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
            # 尝试使用 docker compose (新版)
            if (-not (docker compose version 2>$null)) {
                Write-Error "Error: docker-compose is required"
                exit 1
            }
        }
    } elseif ($Mode -eq "k8s") {
        if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
            Write-Error "Error: kubectl is required for Kubernetes deployment"
            exit 1
        }
    }

    Write-Success "Dependencies OK"
}

# 构建 Docker 镜像
function Build-Images {
    Write-Warning "Building Docker images..."

    Push-Location src/center
    docker build -t "ofa/center:$Version" -t "ofa/center:latest" .
    Pop-Location

    Write-Success "Images built successfully"
}

# Docker Compose 部署
function Deploy-Compose {
    Write-Warning "Deploying with Docker Compose..."

    Push-Location deployments

    # 停止现有部署
    docker-compose down --remove-orphans 2>$null
    if (-not $?) {
        # 尝试新版命令
        docker compose down --remove-orphans 2>$null
    }

    # 启动新部署
    docker-compose up -d
    if (-not $?) {
        docker compose up -d
    }

    # 等待服务就绪
    Write-Host "Waiting for services to be ready..."
    Start-Sleep -Seconds 15

    # 检查服务状态
    docker-compose ps
    if (-not $?) {
        docker compose ps
    }

    Pop-Location

    Write-Success "Docker Compose deployment completed"
    Write-Success "Access Center at http://localhost:8080"
}

# Kubernetes 部署
function Deploy-K8s {
    Write-Warning "Deploying to Kubernetes..."

    # 创建命名空间（如果不存在）
    $namespaceYaml = @"
apiVersion: v1
kind: Namespace
metadata:
  name: $Namespace
"@
    $namespaceYaml | kubectl apply -f -

    # 应用配置
    kubectl apply -f deployments/kubernetes.yaml

    # 等待部署就绪
    Write-Host "Waiting for deployments to be ready..."
    kubectl rollout status deployment/center -n $Namespace --timeout=300s
    kubectl rollout status deployment/postgres -n $Namespace --timeout=120s
    kubectl rollout status deployment/redis -n $Namespace --timeout=60s

    # 显示状态
    Write-Success "Kubernetes deployment completed"
    kubectl get pods -n $Namespace
    kubectl get services -n $Namespace
}

# 显示帮助信息
function Show-Help {
    Write-Host "Usage: .\deploy.ps1 [mode] [version]"
    Write-Host ""
    Write-Host "Modes:"
    Write-Host "  compose  - Deploy with Docker Compose (default)"
    Write-Host "  k8s      - Deploy to Kubernetes"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\deploy.ps1 compose"
    Write-Host "  .\deploy.ps1 k8s v1.0.0"
    Write-Host "  .\deploy.ps1 compose latest"
}

# 主流程
switch ($Mode) {
    "compose" {
        Check-Dependencies
        Build-Images
        Deploy-Compose
    }
    "k8s" {
        Check-Dependencies
        Build-Images
        Deploy-K8s
    }
    { $_ -in "help", "--help", "-h" } {
        Show-Help
        exit 0
    }
    default {
        Write-Error "Unknown deploy mode: $Mode"
        Show-Help
        exit 1
    }
}

Write-Host ""
Write-Success "Deployment completed!"
# OFA 部署指南

## 快速开始

### 环境要求

| 组件 | 版本要求 |
|------|----------|
| Go | 1.22+ |
| Docker | 20.10+ |
| Kubernetes | 1.24+ (可选) |
| PostgreSQL | 14+ (生产环境) |
| Redis | 7+ (生产环境) |
| etcd | 3.5+ (生产环境) |
| NATS | 2.9+ (生产环境) |

---

## 一、开发环境部署

### 1.1 本地编译

```powershell
# 设置Go代理
$env:GOPROXY = "https://goproxy.cn,direct"

# 编译Center
cd D:\vibecoding\OFA\src\center
go build -o ..\..\build\center.exe ./cmd/center

# 编译Agent
cd D:\vibecoding\OFA\src\agent\go
go build -o ..\..\..\build\agent.exe ./cmd/agent
```

### 1.2 运行服务

```powershell
# 启动Center
.\build\center.exe

# 启动Agent（另一个终端）
.\build\agent.exe
```

### 1.3 验证服务

```powershell
# 健康检查
curl http://localhost:8080/health

# 查看指标
curl http://localhost:8080/metrics

# API测试
curl http://localhost:8080/api/v1/agents
curl http://localhost:8080/api/v1/skills
```

---

## 二、Docker部署

### 2.1 构建镜像

```powershell
cd D:\vibecoding\OFA

# 构建Center镜像
docker build -t ofa/center:latest -f Dockerfile.center .

# 构建Agent镜像
docker build -t ofa/agent:latest -f Dockerfile.agent .
```

### 2.2 Docker Compose部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  center:
    image: ofa/center:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DATABASE_TYPE=sqlite
      - DATABASE_PATH=/data/ofa.db
    volumes:
      - ofa-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  agent:
    image: ofa/agent:latest
    depends_on:
      - center
    environment:
      - CENTER_URL=center:9090
    restart: always

volumes:
  ofa-data:
```

```powershell
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

---

## 三、生产环境部署

### 3.1 完整架构

```
                    ┌─────────────────────────────────────┐
                    │           负载均衡器                 │
                    │         (Nginx/HAProxy)            │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    │                             │
            ┌───────▼───────┐             ┌───────▼───────┐
            │   Center 1    │             │   Center 2    │
            │   (gRPC+REST)  │             │   (gRPC+REST)  │
            └───────┬───────┘             └───────┬───────┘
                    │                             │
                    └──────────────┬──────────────┘
                                   │
    ┌──────────────┬───────────────┼───────────────┬──────────────┐
    │              │               │               │              │
┌───▼───┐    ┌─────▼─────┐   ┌─────▼─────┐   ┌─────▼─────┐   ┌───▼───┐
│Redis  │    │PostgreSQL │   │   etcd    │   │   NATS    │   │Prometheus│
└───────┘    └───────────┘   └───────────┘   └───────────┘   └───────┘
```

### 3.2 Kubernetes部署

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ofa
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ofa-config
  namespace: ofa
data:
  center.yaml: |
    server:
      grpc_port: 9090
      rest_port: 8080
    database:
      type: postgres
      host: postgres-service
      port: 5432
      name: ofa
    cache:
      type: redis
      host: redis-service
      port: 6379
    config_center:
      type: etcd
      endpoints:
        - etcd-service:2379
---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ofa-center
  namespace: ofa
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ofa-center
  template:
    metadata:
      labels:
        app: ofa-center
    spec:
      containers:
      - name: center
        image: ofa/center:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: ofa-config
---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: ofa-center
  namespace: ofa
spec:
  selector:
    app: ofa-center
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP
---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ofa-ingress
  namespace: ofa
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.ofa.example.com
    secretName: ofa-tls
  rules:
  - host: api.ofa.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ofa-center
            port:
              number: 8080
```

### 3.3 部署命令

```bash
# 创建命名空间
kubectl apply -f k8s/namespace.yaml

# 部署配置
kubectl apply -f k8s/configmap.yaml

# 部署服务
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 部署Ingress
kubectl apply -f k8s/ingress.yaml

# 查看状态
kubectl get pods -n ofa
kubectl get services -n ofa
kubectl logs -f deployment/ofa-center -n ofa
```

---

## 四、监控部署

### 4.1 Prometheus配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ofa-center'
    static_configs:
      - targets: ['ofa-center:8080']
    metrics_path: /metrics
```

### 4.2 Grafana Dashboard

导入预置Dashboard，监控关键指标：

- Agent数量和状态
- 任务吞吐量和延迟
- 消息投递率
- 资源使用情况

---

## 五、安全配置

### 5.1 TLS配置

```yaml
# TLS证书配置
server:
  tls:
    enabled: true
    cert_file: /etc/ofa/certs/server.crt
    key_file: /etc/ofa/certs/server.key
    ca_file: /etc/ofa/certs/ca.crt
```

### 5.2 JWT认证

```yaml
# JWT配置
auth:
  jwt:
    enabled: true
    signing_method: EdDSA
    public_key: /etc/ofa/keys/public.pem
    private_key: /etc/ofa/keys/private.pem
    access_token_ttl: 15m
    refresh_token_ttl: 24h
```

### 5.3 RBAC配置

```yaml
# RBAC配置
rbac:
  enabled: true
  roles:
    admin:
      permissions: ["*"]
    operator:
      permissions: ["agents:*", "tasks:*", "skills:read"]
    viewer:
      permissions: ["*:read"]
```

---

## 六、备份与恢复

### 6.1 数据库备份

```bash
# PostgreSQL备份
pg_dump -h localhost -U ofa -d ofa > backup_$(date +%Y%m%d).sql

# 恢复
psql -h localhost -U ofa -d ofa < backup_20260330.sql
```

### 6.2 配置备份

```bash
# etcd备份
etcdctl snapshot save backup.db

# 恢复
etcdctl snapshot restore backup.db
```

---

## 七、故障排查

### 7.1 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| Agent无法连接 | 网络不通/地址错误 | 检查网络和配置 |
| 任务执行失败 | Agent离线/技能不存在 | 检查Agent状态和技能 |
| 性能下降 | 资源不足/连接池满 | 扩容或调整配置 |
| 内存泄漏 | goroutine泄漏 | 检查pprof |

### 7.2 日志查看

```bash
# 查看Center日志
kubectl logs -f deployment/ofa-center -n ofa

# 查看Agent日志
kubectl logs -f deployment/ofa-agent -n ofa

# 导出日志
kubectl logs deployment/ofa-center -n ofa > center.log
```

### 7.3 健康检查

```bash
# API健康检查
curl http://localhost:8080/health

# 深度健康检查
curl http://localhost:8080/health?deep=true

# 就绪检查
curl http://localhost:8080/ready
```

---

## 八、性能调优

### 8.1 关键参数

```yaml
# 性能配置
performance:
  # gRPC配置
  grpc:
    max_connections: 10000
    keepalive_time: 30s
    keepalive_timeout: 10s

  # HTTP配置
  http:
    read_timeout: 30s
    write_timeout: 30s
    max_header_bytes: 1MB

  # 连接池
  pool:
    max_idle: 100
    max_active: 1000
    idle_timeout: 5m

  # 缓存
  cache:
    l1_size: 10000
    l2_ttl: 1h
```

### 8.2 资源限制

```yaml
# Kubernetes资源限制
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

---

## 九、版本升级

### 9.1 滚动升级

```bash
# 更新镜像
kubectl set image deployment/ofa-center center=ofa/center:0.9.1 -n ofa

# 查看升级状态
kubectl rollout status deployment/ofa-center -n ofa

# 回滚
kubectl rollout undo deployment/ofa-center -n ofa
```

### 9.2 数据迁移

```bash
# 运行迁移脚本
./scripts/migrate.sh {old_version} {new_version}
```

---

*文档更新时间: 2026-04-10*

---

## 十、快速部署脚本

### 10.1 使用部署脚本

项目提供了便捷的部署脚本：

**Linux/Mac:**
```bash
# Docker Compose 部署
./scripts/deploy.sh compose

# Kubernetes 部署
./scripts/deploy.sh k8s

# 帮助信息
./scripts/deploy.sh --help
```

**Windows PowerShell:**
```powershell
# Docker Compose 部署
.\scripts\deploy.ps1 compose

# Kubernetes 部署
.\scripts\deploy.ps1 k8s

# 帮助信息
.\scripts\deploy.ps1 --help
```

### 10.2 使用 Makefile

```bash
# 查看所有命令
make help

# 构建和测试
make build test

# Docker 构建
make docker-build

# Docker Compose 部署
make deploy-compose

# Kubernetes 部署
make deploy-k8s

# 查看 Kubernetes 状态
make status-k8s
```

---

## 十一、Helm Chart 部署

### 11.1 使用 Helm 安装

```bash
# 安装
helm install ofa-center deployments/helm -n ofa --create-namespace

# 自定义配置
helm install ofa-center deployments/helm -n ofa \
  --set replicaCount=3 \
  --set image.tag=v5.7.0 \
  --set env.JWT_SECRET="your-secret"

# 更新
helm upgrade ofa-center deployments/helm -n ofa

# 卸载
helm uninstall ofa-center -n ofa
```

### 11.2 Helm Chart 配置

主要配置项在 `deployments/helm/values.yaml`：

```yaml
replicaCount: 1

image:
  repository: ofa/center
  tag: "latest"

ingress:
  enabled: true
  hosts:
    - host: ofa.example.com
      paths:
        - path: /api
          servicePort: rest

autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 5
```

---

## 十二、生产环境增强配置

### 12.1 应用生产配置

```bash
# 应用包含 HPA、Ingress、NetworkPolicy 的完整配置
kubectl apply -f deployments/kubernetes.yaml
kubectl apply -f deployments/kubernetes-production.yaml
```

生产配置包含：
- **Ingress**: 外部访问入口，支持 TLS
- **HorizontalPodAutoscaler**: 自动扩缩容
- **PersistentVolumeClaim**: 数据持久化
- **ServiceMonitor**: Prometheus 监控集成
- **NetworkPolicy**: 网络安全策略
- **PodDisruptionBudget**: 最小可用副本保障

### 12.2 生产配置文件

参考 `configs/center-production.yaml` 完整生产配置示例。
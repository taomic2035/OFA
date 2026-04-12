# OFA Center 部署指南 (v8.4.0)

## 一、部署概述

OFA Center 支持多种部署方式：
- Docker Compose (单机部署)
- Kubernetes (集群部署)
- Helm Chart (云原生部署)

## 二、Docker Compose 部署

### 快速启动

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f center
```

### 服务列表

| 服务 | 端口 | 说明 |
|------|------|------|
| Center | 8080, 9090, 8081 | 主服务 (REST/gRPC/WS) |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Prometheus | 9091 | 监控指标 |
| Grafana | 3000 | 监控面板 |
| Adminer | 8082 | 数据库管理 |
| Redis Commander | 8083 | Redis 管理 |

### 环境变量

```bash
# 创建 .env 文件
JWT_SECRET=your-production-secret
CLAUDE_API_KEY=your-claude-key
OPENAI_API_KEY=your-openai-key
```

## 三、Kubernetes 部署

### 部署步骤

```bash
# 创建命名空间
kubectl apply -f deploy/k8s/namespace.yaml

# 创建配置
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml

# 部署数据库
kubectl apply -f deploy/k8s/postgres.yaml
kubectl apply -f deploy/k8s/redis.yaml

# 部署 Center
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

# 查看状态
kubectl get pods -n ofa-system
kubectl get services -n ofa-system
```

### 资源配置

| 组件 | CPU | 内存 | 存储 |
|------|-----|------|------|
| Center | 250-500m | 256-512Mi | - |
| PostgreSQL | 100-200m | 256-512Mi | 10Gi |
| Redis | 50-100m | 64-128Mi | 1Gi |

## 四、Helm Chart 部署

### 安装

```bash
# 安装 Chart
helm install ofa-center deploy/helm/ofa-center \
  --namespace ofa-system \
  --create-namespace

# 自定义配置
helm install ofa-center deploy/helm/ofa-center \
  -f my-values.yaml \
  --namespace ofa-system
```

### 配置项

```yaml
# my-values.yaml
replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi

postgresql:
  auth:
    password: custom-password

monitoring:
  grafana:
    adminPassword: secure-password
```

## 五、监控配置

### Prometheus 指标

Center 服务暴露以下指标端点：
- `/metrics` - Prometheus 格式指标

关键指标：
- `ofa_connections_active` - 活跃 WebSocket 连接
- `ofa_messages_received_total` - 收到的消息数
- `ofa_messages_sent_total` - 发送的消息数
- `ofa_identity_operations_total` - 身份操作数
- `ofa_chat_requests_total` - 对话请求数

### Grafana 面板

访问 Grafana: http://localhost:3000
- 用户名: admin
- 密码: admin123

预置面板：
- Center 性能监控
- WebSocket 连接状态
- 数据库性能
- Redis 缓存命中率

## 六、生产环境建议

### 安全配置

1. **修改默认密码**
   ```yaml
   postgresql:
     auth:
       password: strong-password
   ```

2. **配置 JWT 密钥**
   ```bash
   JWT_SECRET=$(openssl rand -base64 32)
   ```

3. **启用 TLS**
   ```yaml
   ingress:
     enabled: true
     annotations:
       cert-manager.io/cluster-issuer: letsencrypt
   ```

### 高可用配置

1. **增加副本数**
   ```yaml
   replicaCount: 3
   autoscaling:
     enabled: true
     maxReplicas: 10
   ```

2. **配置数据库副本**
   ```yaml
   postgresql:
     replica:
       enabled: true
       replicas: 2
   ```

### 性能优化

1. **调整资源限制**
   ```yaml
   resources:
     limits:
       cpu: 1000m
       memory: 1Gi
   ```

2. **Redis 配置**
   ```yaml
   redis:
     master:
       config:
         maxmemory: 512mb
         maxmemory-policy: allkeys-lru
   ```

## 七、故障排查

### 常见问题

1. **连接数据库失败**
   ```bash
   # 检查数据库状态
   kubectl logs -n ofa-system ofa-postgres-0
   
   # 检查网络
   kubectl exec -n ofa-system ofa-center-xxx -- ping ofa-postgres
   ```

2. **WebSocket 连接超时**
   ```bash
   # 检查 Center 状态
   curl http://localhost:8080/api/v1/status
   
   # 检查日志
   kubectl logs -n ofa-system -l app=ofa-center
   ```

3. **Redis 缓存问题**
   ```bash
   # 连接 Redis
   redis-cli -h localhost -p 6379
   
   # 检查状态
   redis-cli INFO
   ```

---

*文档版本: v8.4.0*
*更新时间: 2026-04-12*

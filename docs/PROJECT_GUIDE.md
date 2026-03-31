# OFA 项目指南

<p align="center">
  <img src="images/logo.svg" alt="OFA Logo" width="150">
</p>

## 目录

1. [项目概述](#项目概述)
2. [快速开始](#快速开始)
3. [架构设计](#架构设计)
4. [开发指南](#开发指南)
5. [部署指南](#部署指南)
6. [API参考](#api参考)
7. [常见问题](#常见问题)

---

## 项目概述

### 什么是 OFA?

OFA (Omni Federated Agents) 是一个开源的多设备分布式智能体系统。它允许各种设备（手机、电脑、手表、IoT设备）作为一个统一的智能网络协同工作。

### 核心概念

| 概念 | 描述 |
|------|------|
| **Center** | 算力中心，负责任务调度、Agent管理、消息路由 |
| **Agent** | 设备节点，运行在手机、电脑、IoT设备上 |
| **Skill** | 技能模块，Agent可执行的具体功能 |
| **Task** | 任务，由Center分配给Agent执行 |

### 应用场景

1. **智能家居联动**: 手机控制灯光、空调、门锁等设备
2. **跨设备协同**: 手机、手表、电脑之间的数据同步和任务协作
3. **分布式AI推理**: 多设备协同完成AI模型推理
4. **工业物联网**: 设备监控、数据采集、异常预警
5. **健康监测**: 可穿戴设备采集健康数据并分析

---

## 快速开始

### 环境要求

- Go 1.22+
- Docker (可选)
- Kubernetes (可选)

### 5分钟快速启动

```bash
# 1. 克隆项目
git clone https://github.com/taomic2035/OFA.git
cd OFA

# 2. 构建
make build

# 3. 启动 Center
./build/center

# 4. 新终端启动 Agent
./build/agent

# 5. 测试 API
curl http://localhost:8080/health
```

### 配置文件

```yaml
# configs/center.yaml
server:
  grpc_port: 9090
  http_port: 8080

database:
  type: sqlite
  path: ./data/ofa.db

llm:
  default_provider: openai
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4
```

---

## 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        Center 集群                           │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │ Center1 │  │ Center2 │  │ Center3 │  │ CenterN │       │
│  │(Leader) │  │(Follower)│ │(Follower)│ │(Follower)│      │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘       │
│       │            │            │            │              │
│       └────────────┴────────────┴────────────┘              │
│                          │                                   │
│              ┌───────────┴───────────┐                      │
│              │    服务发现 & 负载均衡    │                     │
│              └───────────────────────┘                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│    Agent     │   │    Agent     │   │    Agent     │
│   (Mobile)   │   │  (Desktop)   │   │    (IoT)     │
│  Android/iOS │   │   Win/Mac    │   │ Smart Home   │
└──────────────┘   └──────────────┘   └──────────────┘
```

### 核心模块

| 模块 | 路径 | 功能 |
|------|------|------|
| 调度器 | `pkg/scheduler` | 任务分配策略 |
| 执行器 | `pkg/executor` | 任务执行引擎 |
| LLM管理器 | `pkg/llm` | 大模型集成 |
| 协作管理器 | `pkg/collab` | Agent协作 |
| 去中心化 | `pkg/decentralized` | P2P网络 |

---

## 开发指南

### 项目结构

```
OFA/
├── src/
│   ├── center/          # Center 服务源码
│   │   ├── cmd/         # 入口程序
│   │   ├── internal/    # 内部包
│   │   └── pkg/         # 公共包
│   ├── agent/go/        # Go Agent 源码
│   └── sdk/             # 各平台 SDK
│       ├── android/     # Android SDK
│       ├── ios/         # iOS SDK
│       ├── python/      # Python SDK
│       └── ...
├── docs/                # 文档
├── deployments/         # 部署配置
├── configs/             # 配置文件
└── skills/              # 技能库
```

### 开发命令

```bash
# 构建所有组件
make build

# 运行测试
make test

# 代码检查
make lint

# 生成 Proto
make proto

# 运行开发服务器
make dev
```

### 添加自定义技能

```go
// 1. 实现技能接口
type MySkill struct{}

func (s *MySkill) ID() string {
    return "my.custom.skill"
}

func (s *MySkill) Actions() []string {
    return []string{"action1", "action2"}
}

func (s *MySkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
    // 实现逻辑
    return result, nil
}

// 2. 注册技能
agent.RegisterSkill(&MySkill{})
```

---

## 部署指南

### Docker 部署

```bash
# 构建镜像
docker build -t ofa-center:latest .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/configs:/app/configs \
  ofa-center:latest

# 使用 Docker Compose
docker-compose up -d
```

### Kubernetes 部署

```bash
# 创建命名空间
kubectl create namespace ofa

# 部署服务
kubectl apply -f deployments/kubernetes/

# 检查状态
kubectl get pods -n ofa
kubectl get services -n ofa
```

### 生产环境配置

```yaml
# 生产环境推荐配置
resources:
  requests:
    cpu: "2"
    memory: "4Gi"
  limits:
    cpu: "4"
    memory: "8Gi"

replicas: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

---

## API参考

### REST API

```bash
# 健康检查
GET /health

# Agent 管理
GET    /api/v1/agents
POST   /api/v1/agents
DELETE /api/v1/agents/{id}

# 任务管理
GET    /api/v1/tasks
POST   /api/v1/tasks
GET    /api/v1/tasks/{id}

# LLM 对话
POST /api/v1/llm/chat
POST /api/v1/llm/stream

# 协作管理
POST /api/v1/collab
POST /api/v1/collab/{id}/start
GET  /api/v1/collab/{id}
```

### gRPC API

```protobuf
service AgentService {
    rpc Connect(stream AgentMessage) returns (stream CenterMessage);
    rpc ExecuteTask(TaskRequest) returns (TaskResponse);
    rpc GetAgentInfo(AgentInfoRequest) returns (AgentInfoResponse);
}
```

---

## 常见问题

### Q: 如何处理网络不稳定?

Agent SDK 内置自动重连机制，断开后会自动尝试重新连接。

### Q: 如何扩展 Agent 能力?

通过注册自定义 Skill 实现。参考 [开发指南](#开发指南)。

### Q: 如何保证数据安全?

- 所有通信支持 TLS 加密
- 支持端到端加密
- JWT 认证机制
- RBAC 权限控制

### Q: 支持哪些 LLM 模型?

- OpenAI (GPT-4, GPT-3.5)
- Claude (claude-sonnet-4-6, claude-haiku)
- 本地模型 (Ollama 兼容)

---

## 联系我们

- **Issues**: [GitHub Issues](https://github.com/taomic2035/OFA/issues)
- **Discussions**: [GitHub Discussions](https://github.com/taomic2035/OFA/discussions)

---

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

---

*文档更新时间: 2026-03-31*
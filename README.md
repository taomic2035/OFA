# OFA - Omni Federated Agents

多设备分布式智能体系统

## 项目简介

OFA是一个跨设备的多Agent分布式系统，支持手机、平板、电脑、手表等智能设备的协同工作。系统由Center（算力中心）和多个Agent（设备节点）组成，通过gRPC实现高效通信。

**v9.0 AI深度集成** - 新增大语言模型集成、自动代码生成、智能Agent协作、去中心化增强等特性。

## 版本状态

当前版本: **v9.0.0**

| 版本 | 状态 | 主要特性 |
|------|------|----------|
| v8.0.0 | ✅ 完成 | 正式发布 - 安全审计、OpenAPI文档、性能基准测试 |
| v8.1.0 | ✅ 完成 | 增强版本 - WASM技能、插件系统、10平台SDK |
| v9.0.0 | ✅ 完成 | AI深度集成 - LLM、代码生成、协作、去中心化 |

## 核心特性

### 基础能力
- 🌐 **跨平台支持**: Go、Android、iOS、Desktop、Web、Python、Rust、Node.js、C++ 等 10 平台 SDK
- 🔄 **智能调度**: 多种调度策略（能力优先、负载均衡、延迟优先、功耗感知、混合策略）
- 💬 **灵活通信**: P2P、广播、组播、NATS消息队列
- 🔌 **能力扩展**: 可插拔的Skill/Tool架构、WASM技能、插件系统
- 🔒 **安全可靠**: JWT认证、RBAC权限、mTLS、端到端加密

### 企业级特性
- 🏢 **多Center集群**: 服务发现、负载均衡、故障转移、数据同步
- 📦 **能力市场**: 技能仓库、版本管理、依赖解析、安全验证
- 🔄 **工作流引擎**: 步骤编排、定时调度、事件触发
- 👥 **多租户支持**: 租户隔离、资源配额、计费系统
- 📊 **可观测性**: 分布式追踪、日志聚合、告警管理

### AI能力
- 🖥️ **边缘计算**: 边缘Center、本地处理、边云协同、数据预处理
- 🤖 **AI推理**: 模型推理、GPU调度、分布式推理、模型量化
- 🎓 **联邦学习**: 分布式训练、隐私保护、梯度聚合、安全聚合

### v9.0 新增特性

#### 🔮 LLM深度集成
- **多提供商支持**: OpenAI、Claude、本地模型(Ollama兼容)
- **统一接口**: Chat、Stream、Embed 统一API
- **Prompt管理**: 模板注册、变量验证、上下文管理
- **LLM Agent**: 工具调用、记忆管理、多轮对话
- **RAG检索增强**: 向量存储、知识库管理、相似度搜索

#### ⚡ 自动代码生成
- **API生成**: 模型、处理器、路由、测试代码自动生成
- **SDK生成**: Go、TypeScript、Python SDK 一键生成
- **文档生成**: Markdown、HTML、OpenAPI 文档自动生成
- **Proto生成**: gRPC 协议文件自动生成

#### 🤝 智能Agent协作
- **协作模式**: 顺序、并行、管道、MapReduce、共识、拍卖、协商
- **任务编排**: 依赖管理、优先级调度、状态追踪
- **任务分配**: 多策略分配、Agent评分、动态负载均衡
- **结果聚合**: 多策略聚合、成本计算、统计报告
- **协商机制**: 提议投票、协议生成、冲突解决

#### 🌐 去中心化增强
- **网络类型**: 全P2P、混合、联邦、网状网络
- **共识机制**: PBFT、Raft、PoA、PoS、加权投票
- **数据复制**: 多复制策略、哈希校验、自动修复
- **Peer发现**: DHT、Gossip、广播、注册中心
- **信任管理**: 多信任算法、评分衰减、信任等级

## 快速开始

### 环境要求

- Go 1.22+
- Docker (可选)
- Kubernetes (可选)

### 安装

```bash
# 克隆仓库
git clone https://github.com/ofa/ofa.git
cd ofa

# 构建
go build -o build/center ./src/center/cmd/center
go build -o build/agent ./src/agent/cmd/agent
```

### 运行

```bash
# 启动 Center
./build/center

# 启动 Agent
./build/agent --center localhost:9090 --id agent-001
```

### 配置

```yaml
# config.yaml
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

decentralized:
  enabled: true
  network_type: hybrid
```

## 使用示例

### LLM 对话

```go
import "ofa/pkg/llm"

// 创建 LLM 管理器
manager := llm.NewManager()
manager.RegisterAdapter("openai", llm.NewOpenAIAdapter("api-key", "gpt-4"))

// 发送请求
resp, _ := manager.Chat(context.Background(), "openai", &llm.ChatRequest{
    Messages: []llm.Message{{Role: "user", Content: "你好"}},
})
fmt.Println(resp.Content)
```

### Agent 协作

```go
import "ofa/pkg/collab"

// 创建协作
manager := collab.NewCollaborationManager()
collab, _ := manager.CreateCollaboration(ctx, &collab.CreateCollabRequest{
    Name: "数据处理",
    Type: collab.CollabTypeParallel,
    Tasks: []*collab.CollabTask{
        {ID: "t1", Name: "任务1", SkillID: "data.process"},
        {ID: "t2", Name: "任务2", SkillID: "data.process"},
    },
})

// 启动协作
manager.StartCollaboration(ctx, collab.ID)
```

### 代码生成

```go
import "ofa/pkg/codegen"

// 定义 API
apiSpec := codegen.APISpec{
    Name: "UserService",
    Models: []codegen.ModelSpec{
        {Name: "User", Fields: []codegen.FieldSpec{
            {Name: "ID", Type: "int64"},
            {Name: "Name", Type: "string"},
        }},
    },
}

// 生成代码
apiGen := codegen.NewAPIGenerator(generator)
apiGen.GenerateModels(apiSpec, "./models")
apiGen.GenerateHandlers(apiSpec, "./handlers")
```

更多示例请参考 [用户指南](docs/USER_GUIDE.md)。

## API 端点

### REST API

```bash
# 健康检查
GET /health

# Prometheus 指标
GET /metrics

# Agent 管理
GET/POST /api/v1/agents

# 任务管理
GET/POST /api/v1/tasks

# LLM 对话
POST /api/v1/llm/chat
POST /api/v1/llm/stream

# 协作管理
GET/POST /api/v1/collab
POST /api/v1/collab/{id}/start

# 去中心化网络
GET /api/v1/network/stats
GET /api/v1/network/nodes
```

## SDK 支持

| 平台 | 语言 | 位置 | 状态 |
|------|------|------|------|
| Go | Go | `src/agent/go/` | ✅ |
| Android | Java | `src/sdk/android/` | ✅ |
| iOS | Swift | `src/sdk/ios/` | ✅ |
| Desktop | Go | `src/sdk/desktop/` | ✅ |
| Web | TypeScript | `src/sdk/web/` | ✅ |
| Lite (手表) | Go | `src/sdk/lite/` | ✅ |
| IoT (智能家居) | Go | `src/sdk/iot/` | ✅ |
| Python | Python | `src/sdk/python/` | ✅ |
| Rust | Rust | `src/sdk/rust/` | ✅ |
| Node.js | TypeScript | `src/sdk/nodejs/` | ✅ |
| C++ | C++17 | `src/sdk/cpp/` | ✅ |

## 部署

### Docker

```bash
docker build -t ofa-center:latest .
docker run -d -p 8080:8080 -p 9090:9090 ofa-center:latest
```

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### 开发环境

```bash
docker-compose -f deployments/docker-compose.dev.yml up -d
```

## 项目结构

```
OFA/
├── build/                    # 编译产物
├── src/
│   ├── center/              # Center 源码
│   │   ├── cmd/             # 入口
│   │   ├── internal/        # 内部包
│   │   └── pkg/             # 公共包
│   │       ├── llm/         # LLM 集成
│   │       ├── codegen/     # 代码生成
│   │       ├── collab/      # Agent 协作
│   │       ├── decentralized/ # 去中心化
│   │       ├── wasm/        # WASM 支持
│   │       ├── plugin/      # 插件系统
│   │       └── ...
│   ├── agent/go/            # Agent 源码
│   └── sdk/                 # SDK 源码
├── docs/                    # 文档
├── deployments/             # 部署配置
└── configs/                 # 配置文件
```

## 文档

- [用户指南](docs/USER_GUIDE.md) - 详细使用说明
- [设备接入指南](docs/DEVICE_GUIDE.md) - 移动设备、IoT设备接入
- [更新日志](CHANGELOG.md) - 版本更新记录
- [项目状态](PROJECT_STATUS.md) - 当前开发状态
- [架构设计](docs/03-ARCHITECTURE_DESIGN.md) - 系统架构
- [API 文档](docs/API.md) - API 详细说明
- [部署指南](docs/DEPLOYMENT.md) - 部署说明

## 模块统计

| 类别 | 模块数 | 说明 |
|------|--------|------|
| 核心框架 | 11 | 入口、配置、模型、存储、调度、服务、版本 |
| LLM | 6 | 管理器、适配器、Prompt、Agent、向量、服务 |
| 代码生成 | 4 | 生成器、API、SDK、文档 |
| Agent协作 | 5 | 管理器、编排器、分配器、聚合器、协商器 |
| 去中心化 | 7 | 管理器、Peer、共识、复制、发现、同步、信任 |
| 通信层 | 8 | gRPC、REST、P2P、路由器、广播、队列、存储 |
| 安全 | 4 | JWT、mTLS、E2E加密、安全审计 |
| AI能力 | 6 | 管理器、推理、分布式、量化、调参、版本 |
| 其他 | 30+ | 集群、工作流、RBAC、边缘计算、联邦学习... |

**总计: 119 个源文件**

## 许可证

MIT License
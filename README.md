# OFA - Omni Federated Agents

<p align="center">
  <img src="docs/images/logo.svg" alt="OFA Logo" width="200">
</p>

<p align="center">
  <strong>多设备分布式智能体系统 | Multi-Device Distributed Agent System</strong>
</p>

<p align="center">
  <em>Omni - 全能 | Federated - 分布式 | Agents - 智能体</em>
</p>

<p align="center">
  <a href="#简介">中文</a> | <a href="#introduction">English</a>
</p>

<p align="center">
  <a href="https://github.com/taomic2035/OFA/releases"><img src="https://img.shields.io/github/v/release/taomic2035/OFA?include_prereleases" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/taomic2035/OFA"><img src="https://goreportcard.com/badge/github.com/taomic2035/OFA" alt="Go Report Card"></a>
  <a href="https://github.com/taomic2035/OFA/blob/main/LICENSE"><img src="https://img.shields.io/github/license/taomic2035/OFA" alt="License"></a>
  <a href="https://github.com/taomic2035/OFA/stargazers"><img src="https://img.shields.io/github/stars/taomic2035/OFA?style=social" alt="Stars"></a>
</p>

---

## 简介

OFA (Omni Federated Agents) 是一个跨设备的多Agent分布式系统，支持手机、平板、电脑、手表、IoT设备等智能设备的协同工作。系统由 Center（算力中心）和多个 Agent（设备节点）组成，通过 gRPC 和 MQTT 实现高效通信。

### 核心特性

#### 🚀 v1.0 新特性 - MCP协议支持

| 特性 | 描述 |
|------|------|
| **MCP协议集成** | Android SDK 支持 Model Context Protocol，AI Agent 可直接调用设备工具 |
| **40+ 内置工具** | 系统工具、设备工具、数据工具、AI工具、UI自动化工具，全面覆盖设备能力 |
| **离线执行** | L1-L4 四级离线支持，工具在离线状态仍可执行 |
| **OpenAI兼容** | ToolCallingAdapter 提供 OpenAI Function Calling 格式支持 |
| **UI自动化** | 基于 AccessibilityService 的跨应用UI操作，支持点击、滑动、输入等 |

#### 📱 多平台支持

支持 10+ 平台的 SDK：

| 平台 | 语言 | 状态 |
|------|------|------|
| Android | Java/Kotlin | ✅ |
| iOS | Swift | ✅ |
| Desktop | Go | ✅ |
| Web | TypeScript | ✅ |
| 可穿戴设备 (手表/手环) | Go | ✅ |
| IoT (智能家居) | Go | ✅ |
| Python | Python | ✅ |
| Rust | Rust | ✅ |
| Node.js | TypeScript | ✅ |
| C++ | C++17 | ✅ |

#### 🏢 企业级特性

- **多租户支持**: 租户隔离、资源配额、计费系统
- **高可用集群**: 服务发现、负载均衡、故障转移
- **安全认证**: JWT、mTLS、端到端加密、RBAC权限
- **可观测性**: Prometheus指标、分布式追踪、日志聚合

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/taomic2035/OFA.git
cd OFA

# 构建 Center 服务
cd src/center
go build -o ../../build/center ./cmd/center

# 构建 Agent 客户端
cd ../agent/go
go build -o ../../../build/agent ./cmd/agent

# 启动服务
./build/center
./build/agent --center localhost:9090
```

### 文档

- [项目指南](docs/PROJECT_GUIDE.md) - 完整的项目指南
- [用户指南](docs/USER_GUIDE.md) - 详细使用说明
- [设备接入指南](docs/DEVICE_GUIDE.md) - Android/iOS/IoT/可穿戴设备接入
- [架构设计](docs/03-ARCHITECTURE_DESIGN.md) - 系统架构说明
- [API文档](docs/API.md) - REST/gRPC API参考
- [部署指南](docs/DEPLOYMENT.md) - Docker/Kubernetes部署
- [开发指南](docs/DEVELOPMENT.md) - 开发环境配置
- [更新日志](CHANGELOG.md) - 版本更新记录

### Web Dashboard

OFA 提供 Web 管理控制台，支持可视化的 Agent/Task 管理:

```bash
# 启动 Dashboard
cd src/dashboard
npm install
npm run dev
# 访问 http://localhost:3000
```

功能模块:
- 📊 **控制台** - 系统概览、统计卡片、实时活动流
- 🤖 **智能体管理** - Agent 列表、搜索、详情、删除
- 📋 **任务管理** - 任务列表、新建表单、状态筛选
- 📈 **系统监控** - 实时指标、WebSocket 更新
- 💬 **消息中心** - 消息发送、广播、历史记录
- ⚙️ **系统设置** - 连接配置、显示设置

### 许可证

[MIT License](LICENSE)

---

## Introduction

OFA (Omni Federated Agents) is a cross-device distributed agent system that supports smartphones, tablets, computers, wearables, and IoT devices. The system consists of a Center (compute hub) and multiple Agents (device nodes), communicating via gRPC and MQTT.

### Key Features

#### 🚀 v1.0 New Features - MCP Protocol Support

| Feature | Description |
|---------|-------------|
| **MCP Integration** | Android SDK supports Model Context Protocol, AI agents can call device tools directly |
| **40+ Built-in Tools** | System, device, data, AI, and UI automation tools covering full device capabilities |
| **Offline Execution** | L1-L4 offline levels, tools work offline |
| **OpenAI Compatible** | ToolCallingAdapter provides OpenAI Function Calling format support |
| **UI Automation** | Cross-app UI operations via AccessibilityService - click, swipe, input, and more |

#### 📱 Multi-Platform Support

SDKs for 10+ platforms: Android, iOS, Desktop, Web, Wearables, IoT, Python, Rust, Node.js, C++

#### 🏢 Enterprise Features

- **Multi-tenancy**: Tenant isolation, resource quotas, billing
- **High Availability**: Service discovery, load balancing, failover
- **Security**: JWT, mTLS, E2E encryption, RBAC
- **Observability**: Prometheus metrics, distributed tracing, log aggregation

### Quick Start

```bash
# Clone repository
git clone https://github.com/taomic2035/OFA.git
cd OFA

# Build Center service
cd src/center
go build -o ../../build/center ./cmd/center

# Build Agent client
cd ../agent/go
go build -o ../../../build/agent ./cmd/agent

# Run services
./build/center
./build/agent --center localhost:9090
```

### Documentation

- [User Guide](docs/USER_GUIDE.md) - Detailed usage instructions
- [Device Integration Guide](docs/DEVICE_GUIDE.md) - Android/iOS/IoT/Wearable integration
- [Dashboard](src/dashboard/README.md) - Web management console
- [Architecture Design](docs/03-ARCHITECTURE_DESIGN.md) - System architecture
- [API Reference](docs/API.md) - REST/gRPC API reference
- [Deployment Guide](docs/DEPLOYMENT.md) - Docker/Kubernetes deployment
- [Changelog](CHANGELOG.md) - Version history

### License

[MIT License](LICENSE)

---

## Star History

如果这个项目对您有帮助，请给我们一个 ⭐️！

If this project helps you, please give us a ⭐️!
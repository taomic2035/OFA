# OFA - Omni Federated Agents

<p align="center">
  <img src="docs/images/logo.svg" alt="OFA Logo" width="200">
</p>

<p align="center">
  <strong>多设备分布式智能体系统 | Multi-Device Distributed Agent System</strong>
</p>

<p align="center">
  <em>万物皆为我所用，万物皆是我。</em>
</p>

<p align="center">
  <a href="https://github.com/taomic2035/OFA/releases"><img src="https://img.shields.io/github/v/release/taomic2035/OFA?include_prereleases" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/taomic2035/OFA"><img src="https://goreportcard.com/badge/github.com/taomic2035/OFA" alt="Go Report Card"></a>
  <a href="https://github.com/taomic2035/OFA/blob/main/LICENSE"><img src="https://img.shields.io/github/license/taomic2035/OFA" alt="License"></a>
  <a href="https://github.com/taomic2035/OFA/stargazers"><img src="https://img.shields.io/github/stars/taomic2035/OFA?style=social" alt="Stars"></a>
</p>

---

## 核心理念

OFA 采用 **Center-Agent 分布式架构**，实现跨设备的人格统一与智能协同。

| 角色 | 职责 | 特性 |
|------|------|------|
| **Center** | 永远在线的灵魂载体 | 最终基准、冲突仲裁、数据纠偏 |
| **Agent** | 设备端载体 | 可离线、可更换、定期同步 |

**设计原则**：
- Center 保持最终数据基准，负责冲突决策与纠偏
- 设备可随时离线或更换，Center 始终在线维护人格一致性
- 所有设备共享同一人格，通过 Center 实现跨设备同步

---

## 版本历程

### v9.x 性能保障系列 (Current)

构建完整的性能保障体系，实现生产级稳定运行。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v9.3.0 | 健康检查完善 | HealthChecker、自愈机制、降级策略、告警集成 |
| v9.2.0 | 缓存优化 | TieredCache、L1/L2多级缓存、热点数据、LFU驱逐 |
| v9.1.0 | API Gateway | RateLimiter、CircuitBreaker、ResponseCache、中间件链 |
| v9.0.0 | 性能压测框架 | BenchmarkRunner、WebSocket压测、身份压测、Chat压测 |

### v8.x LLM与多平台系列

实现 LLM 智能对话和多平台 SDK 全覆盖。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v8.5.0 | 测试覆盖完善 | LLM Provider测试、Chat API测试、Audio Stream测试 |
| v8.4.0 | 部署方案完善 | Docker Compose、Kubernetes、Helm Chart |
| v8.3.0 | Web SDK TypeScript | OFAWebAgent、WebSocket、AudioPlayer、ChatClient |
| v8.2.0 | E2E测试框架 | 测试计划文档、Center测试脚本、Go E2E测试 |
| v8.1.0 | iOS SDK基础实现 | OFAiOSAgent、SwiftUI支持、iPhone/iPad/Mac/watch |
| v8.0.1 | Android音频播放 | AudioPlayer PCM流、AudioStreamReceiver |
| v8.0.0 | LLM智能对话 | Claude/OpenAI Provider、Chat REST API、流式对话 |

### v7.x 通信与安全系列

建立 Center-Agent WebSocket 通信桥接和安全认证体系。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v7.5.0 | CI/CD集成 | GitHub Actions、多平台构建、Docker推送、Release发布 |
| v7.4.0 | SDK架构文档 | SDK_ARCHITECTURE.md、SCENE_DESIGN.md |
| v7.3.0 | 场景联动实现 | SceneEngine、跑步/会议/健康异常场景 |
| v7.2.0 | Agent Token + RBAC | JWT认证、设备Token、角色权限系统 |
| v7.1.0 | PostgreSQL + Redis | 持久化存储、双层缓存、迁移脚本 |
| v7.0.0 | WebSocket通信桥接 | ConnectionManager、14种消息类型、状态推送 |

### v5.x 外在呈现系列

构建完整的数字人外在呈现能力，形成 **内在灵魂 → 外在呈现** 闭环。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v5.6.4 | TTS REST API | 语音合成API、声音克隆API、身份-声音映射API |
| v5.6.3 | Android TTS客户端 | TTSClient、流式合成、音频播放、声音克隆 |
| v5.6.2 | TTS服务集成 | TTSService、会话管理、身份声音映射 |
| v5.6.1 | TTS单元测试 | 协议测试、引擎测试、Mock提供者 |
| v5.6.0 | TTS引擎核心 | TTSProvider、Volcengine、Doubao、WebSocket协议、30+音色 |
| v5.5.0 | 多端展示系统 | MultiDisplayProfile、DisplayEngine、设备适配、状态同步 |
| v5.4.0 | 形象个性化系统 | AvatarPersonalization、PersonalizationEngine、场景适应、风格管理 |
| v5.3.0 | 表情动作系统 | ExpressionGestureProfile、ExpressionGestureEngine、情绪映射 |
| v5.2.0 | 表达内容系统 | SpeechContentProfile、SpeechContentEngine、文化表达、三观影响 |
| v5.1.0 | 语音合成系统 | VoiceProfile、VoiceEngine、情绪语音联动、TTS集成 |
| v5.0.0 | 外在形象系统 | Avatar、AvatarEngine、面型体型、风格偏好、3D模型引用 |

### v4.x 灵魂特征系列

构建数字人的内在灵魂特征，包括情绪、三观、身份、文化等核心维度。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v4.6.0 | 人际关系系统 | Relationship、SocialNetwork、AttachmentStyle、关系画像 |
| v4.5.0 | 情绪行为联动 | EmotionBehavior、决策影响、表达影响、应对策略 |
| v4.4.0 | 人生阶段系统 | LifeStage、LifeEvent、LifeLesson、发展指标 |
| v4.3.0 | 地域文化影响 | RegionalCulture、Hofstede文化维度、沟通风格、社交风格 |
| v4.2.0 | 社会身份画像 | EducationBackground、CareerProfile、SocialClass、IdentityProfile |
| v4.1.0 | 三观系统完善 | Worldview、LifeView、EnhancedValueSystem、道德判断框架 |
| v4.0.0 | 情绪系统核心 | 七情六欲模型、马斯洛需求层次、EmotionEngine |

### v3.x 多设备协同系列

实现多设备间的高效协同与智能调度。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v3.7.0 | 安全增强 | SecurityService、AES-GCM/CBC、安全会话、密钥管理 |
| v3.6.0 | 数据同步优化 | DataSyncService、增量同步、冲突检测与解决 |
| v3.5.0 | 设备群组管理 | DeviceGroup、群组创建、成员管理、群组广播 |
| v3.4.0 | 跨设备通知 | NotificationHub、智能分发、优先级管理、勿扰模式 |
| v3.3.0 | 任务协同执行 | TaskCoordinator、任务拆分、结果合并、失败重试 |
| v3.2.0 | 场景感知路由 | SceneRouter、智能路由、自定义规则 |
| v3.1.0 | 设备状态同步 | StateSyncManager、实时同步、状态广播 |
| v3.0.0 | 设备消息总线 | MessageBus、离线消息、优先级队列 |

### v2.x 去中心化架构系列

建立 Center-Agent 分布式架构基础。

| 版本 | 特性 | 核心组件 |
|------|------|----------|
| v2.9.0 | 性格进化引擎 | EvolutionEngine、稳定性检测、MBTI收敛、趋势分析 |
| v2.8.0 | 设备生命周期与信任链 | TrustManager、设备优先级、信任级别、设备更换 |
| v2.7.0 | 数据持久化增强 | PostgreSQL + Redis 混合存储 |
| v2.6.0 | Center权威与冲突仲裁 | ConflictArbiter、统一决策、纠偏机制 |
| v2.5.0 | 身份同步完善 | JSON解析、行为上报HTTP、性格推断引擎 |
| v2.4.0 | 行为上报与性格推断 | BehaviorCollector、实时性格推断 |
| v2.3.0 | 运行模式简化 | 默认SYNC模式 |
| v2.2.0 | Memory跨设备同步 | MemorySyncService、冲突自动解决 |
| v2.1.0 | Center角色转变 | 从控制中心转为数据中心 |
| v2.0.0 | 身份同步基础层 | IdentityManager、所有设备共享人格 |

### v1.x 基础能力系列

| 版本 | 特性 |
|------|------|
| v1.4.0 | PostgreSQL持久化、Redis缓存、用户画像系统 |
| v1.3.0 | WebView自动化引擎 |
| v1.2.1 | 视觉智能增强 (ML Kit OCR、模板匹配) |
| v1.2.0 | 扩展App适配器 (抖音、小红书、滴滴) |

---

## AutomationEngine

OFA Android SDK 提供完整的UI自动化引擎，支持7个自动化阶段：

| 阶段 | 能力 |
|------|------|
| Phase 1 | 无障碍服务基础 (AccessibilityEngine, NodeFinder, GesturePerformer) |
| Phase 2 | UI自动化增强 (ScrollHelper, PageMonitor, ScreenCapture) |
| Phase 3 | 应用适配层 (美团/饿了么/淘宝/京东/抖音/小红书/滴滴) |
| Phase 4 | ROM系统层 (SystemAutomationEngine, KeepAlive) |
| Phase 5 | 集成优化 (AutomationOrchestrator, SkillBridge, ErrorRecovery) |
| Phase 6 | 视觉智能 (MlKitOcrEngine, TemplateMatcher, VisionEngine) |
| Phase 7 | WebView自动化 (JsExecutor, WebFormFiller, WebEventListener) |

---

## MCP协议支持

| 特性 | 说明 |
|------|------|
| MCP协议集成 | Android SDK 支持 Model Context Protocol |
| 50+ 内置工具 | 系统工具、设备工具、数据工具、AI工具、UI自动化工具 |
| 离线执行 | L1-L4 四级离线支持 |
| OpenAI兼容 | ToolCallingAdapter 提供 Function Calling 格式支持 |
| 双LLM支持 | 云端 (OpenAI/Anthropic) + 本地 (TensorFlow Lite) |

---

## 多平台支持

| 平台 | 语言 | 状态 |
|------|------|------|
| Android | Java/Kotlin | ✅ |
| iOS | Swift | ✅ |
| Desktop | Go | ✅ |
| Web | TypeScript | ✅ |
| Wearable | Go | ✅ |
| IoT | Go | ✅ |
| Python | Python | ✅ |
| Rust | Rust | ✅ |
| Node.js | TypeScript | ✅ |
| C++ | C++17 | ✅ |

---

## 项目规模

| 组件 | 文件数 |
|------|--------|
| Center (Go) | 145+ |
| Android SDK (Java) | 82+ |
| iOS SDK (Swift) | 10 |
| Web SDK (TypeScript) | 8 |
| Unit Tests | 55+ |
| Benchmark | 8 |
| Deployment Configs | 15 |

---

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/taomic2035/OFA.git
cd OFA

# 构建 Center 服务
cd src/center
go build -o ../../build/center ./cmd/center

# 启动服务
./build/center
```

### Android SDK 集成

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:1.4.0'
}
```

```java
// 初始化
OFAAndroidAgent agent = OFAAndroidAgent.getInstance(context);
agent.initialize("center.example.com:9090");

// 自动化引擎
AutomationOrchestrator automation = agent.getAutomationOrchestrator();
automation.execute("search", Map.of("query", "奶茶"));

// 视觉能力
VisionAutomationEngine vision = automation.getVisionEngine();
vision.clickText("确定");
```

---

## 文档

| 文档 | 说明 |
|------|------|
| [API文档](docs/API.md) | REST/gRPC API 参考 (80+ 端点) |
| [阶段总结](docs/PHASE_SUMMARY.md) | v2.x - v9.x 版本历程 |
| [SDK架构](docs/SDK_ARCHITECTURE.md) | Android SDK 设计说明 |
| [场景设计](docs/SCENE_DESIGN.md) | 场景引擎架构设计 |
| [部署指南](docs/DEPLOYMENT.md) | Docker/Kubernetes 部署 |
| [E2E测试计划](docs/E2E_TEST_PLAN.md) | 端到端联调验证 |
| [压测文档](docs/BENCHMARK.md) | 性能压测框架 |

---

## 企业级特性

- **多租户支持**: 租户隔离、资源配额、计费系统
- **高可用集群**: 服务发现、负载均衡、故障转移
- **安全认证**: JWT、mTLS、端到端加密、RBAC权限
- **可观测性**: Prometheus指标、分布式追踪、日志聚合

---

## 许可证

[MIT License](LICENSE)

---

如果这个项目对您有帮助，请给我们一个 ⭐️！
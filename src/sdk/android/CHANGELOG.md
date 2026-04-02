# OFA Android SDK v1.0.1 - Dual LLM Support

## 发布日期: 2026-04-02

## 更新概述

本次更新为 OFA Android SDK 添加了双 LLM 架构支持，实现云端 LLM 和本地 LLM 的无缝切换，支持离线 AI 推理。

## 新增功能

### 双 LLM 架构

#### LLM 提供者接口
- ✅ `LLMProvider` - 统一 LLM 接口
- ✅ `LLMConfig` - 配置构建器
- ✅ `LLMResponse` - 响应结构（含 ToolCall 支持）
- ✅ `Message` - 聊天消息（支持多角色）
- ✅ `StreamCallback` - 流式回调接口
- ✅ `LLMStats` - 统计追踪

#### 云端 LLM (CloudLLMProvider)
- ✅ OpenAI 兼容 HTTP API
- ✅ 支持自定义端点
- ✅ API Key 认证
- ✅ 流式响应
- ✅ Tool Calling 支持

#### 本地 LLM (LocalLLMProvider)
- ✅ TensorFlow Lite 推理引擎
- ✅ GPU 加速支持（可选）
- ✅ 温度采样
- ✅ 流式生成
- ✅ 嵌入向量生成

#### LLM 编排器 (LLMOrchestrator)
- ✅ 多提供者管理
- ✅ 自动故障转移
- ✅ 负载均衡策略
- ✅ 健康检查

### OFAAgent Builder 新增方法

```java
// LLM 配置
Builder llmProvider(LLMProvider provider)
Builder fallbackLLMProvider(LLMProvider provider)
Builder cloudLLM(String endpoint, String apiKey, String model)
Builder localLLM(String modelPath)
Builder autoLLMFailover(boolean enable)
```

### OFAAgent 新增方法

```java
// LLM 访问
boolean hasLLM()
LLMProvider getLLMProvider()
LLMOrchestrator getLLMOrchestrator()
```

## 项目结构更新

```
sdk/src/main/java/com/ofa/agent/
├── llm/
│   ├── LLMProvider.java         # LLM 接口
│   ├── LLMConfig.java           # 配置
│   ├── LLMResponse.java         # 响应
│   ├── Message.java             # 消息
│   ├── StreamCallback.java      # 流式回调
│   ├── LLMStats.java            # 统计
│   ├── cloud/
│   │   └── CloudLLMProvider.java
│   ├── local/
│   │   ├── LocalLLMProvider.java
│   │   ├── TFLiteEngine.java
│   │   └── Tokenizer.java
│   ├── orchestrator/
│   │   └── LLMOrchestrator.java
│   └── tool/
│       └── LLMChatTool.java     # MCP 工具包装
├── grpc/
│   ├── AgentGrpc.java           # gRPC Stub
│   └── AgentOuterClass.java     # Protobuf 消息
└── constraint/
    ├── ConstraintType.java      # 约束类型
    ├── ConstraintResult.java    # 检查结果
    ├── ConstraintRule.java      # 约束规则
    └── ConstraintChecker.java   # 检查器
```

## 编译系统更新

### Gradle 配置
- ✅ 多模块结构 (`sdk/` 子模块)
- ✅ Gradle 8.2 + AGP 8.2.0
- ✅ Java 17 兼容性
- ✅ Maven Publishing 配置

### 编译命令
```bash
cd src/sdk/android
export JAVA_HOME="D:/Java/jdk-17"
export ANDROID_HOME="D:/Android/Sdk"
./gradlew.bat assembleRelease --no-daemon
```

### 输出位置
- AAR: `sdk/build/outputs/aar/sdk-release.aar`

## 依赖更新

```groovy
// TensorFlow Lite (本地 LLM)
implementation 'org.tensorflow:tensorflow-lite:2.14.0'
implementation 'org.tensorflow:tensorflow-lite-support:0.4.4'

// OkHttp (云端 LLM HTTP 客户端)
implementation 'com.squareup.okhttp3:okhttp:4.12.0'
```

## Bug 修复

- 修复 `ConstraintChecker` 多类定义问题（拆分为独立文件）
- 修复 `LLMChatTool` 缺少接口方法实现
- 修复 `LLMOrchestrator.embed()` 返回类型不匹配
- 修复 `ToolDefinition` 构造函数兼容性
- 修复多个工具类 JSONException 未捕获问题
- 修复 `CameraTool` 字节数组类型错误
- 修复 `SensorTool` 重复 case 标签
- 修复 `NFCTool` API 兼容性问题

## 兼容性

- Android SDK: 24+ (Android 7.0)
- Target SDK: 34 (Android 14)
- Java: 17
- Gradle: 8.2
- Android Gradle Plugin: 8.2.0

## 下一步计划

- [ ] TensorFlow Lite 模型加载优化
- [ ] 更多本地模型支持 (Gemma, Phi, etc.)
- [ ] 量化模型支持
- [ ] 单元测试完善
- [ ] 性能基准测试

---

**版本**: v1.0.1
**作者**: OFA Team
**日期**: 2026-04-02
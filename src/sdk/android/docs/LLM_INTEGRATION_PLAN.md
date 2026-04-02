# Android SDK LLM 能力规划

## 状态: ✅ 已实现 (v1.0.1)

本文档描述的架构已在 v1.0.1 版本中实现。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        OFA Android SDK                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    LLM Provider Layer                    │   │
│  │  ┌─────────────────┐    ┌──────────────────────────┐   │   │
│  │  │  CloudLLMProvider │    │   LocalLLMProvider        │   │   │
│  │  │  (OpenAI/Claude)  │    │   (TensorFlow Lite)       │   │   │
│  │  │  - API Key Auth   │    │   - Local Model Inference │   │   │
│  │  │  - REST/WS Call   │    │   - Offline Capable       │   │   │
│  │  │  - Streaming      │    │   - Low Latency           │   │   │
│  │  └────────┬─────────┘    └────────────┬─────────────┘   │   │
│  │           │                           │                  │   │
│  │           └───────────┬───────────────┘                  │   │
│  │                       ▼                                  │   │
│  │           ┌───────────────────────┐                     │   │
│  │           │    LLMOrchestrator    │                     │   │
│  │           │  - Auto Failover      │                     │   │
│  │           │  - Load Balancing     │                     │   │
│  │           │  - Offline Detection  │                     │   │
│  │           └───────────┬───────────┘                     │   │
│  └───────────────────────┼─────────────────────────────────┘   │
│                          │                                      │
│  ┌───────────────────────┼─────────────────────────────────┐   │
│  │                 Agent Communication                      │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │   Center    │  │  P2P Agent  │  │   MCP Server    │  │   │
│  │  │  Connection │  │  Discovery  │  │   Tool Host     │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 模块设计

### 1. LLM Provider 接口

```java
public interface LLMProvider {
    // 基础能力
    String getId();
    String getName();
    boolean isAvailable();
    boolean supportsOffline();

    // 聊天能力
    CompletableFuture<LLMResponse> chat(List<Message> messages);
    CompletableFuture<LLMResponse> chatWithTools(List<Message> messages, List<ToolDefinition> tools);

    // 流式响应
    void streamChat(List<Message> messages, StreamCallback callback);

    // Embedding
    CompletableFuture<float[]> embed(String text);

    // 配置
    void configure(LLMConfig config);
    LLMStats getStats();
}
```

### 2. CloudLLMProvider (云端 LLM)

**支持平台**：
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Azure OpenAI
- 自定义 OpenAI 兼容端点

**特性**：
- API Key 认证
- 流式响应
- Function Calling
- Token 计数
- 超时重试

### 3. LocalLLMProvider (本地 LLM)

**技术栈**：
- TensorFlow Lite
- 支持 GGUF/GGML 格式
- 量化模型 (4-bit, 8-bit)

**支持的模型**：
- Phi-2 (2.7B) - 轻量级
- Gemma (2B/7B)
- Llama 3 (8B) - 中等规模
- Qwen (1.8B/7B) - 中文优化

**特性**：
- 完全离线运行
- 低延迟推理
- 模型热切换
- 内存优化

### 4. LLMOrchestrator (编排器)

**职责**：
- 自动选择最佳 Provider
- 离线检测与切换
- 负载均衡
- 降级策略

**切换策略**：
```
Online → CloudLLMProvider (优先)
      → LocalLLMProvider (降级)

Offline → LocalLLMProvider (唯一选项)
```

### 5. 与 Agent 系统集成

**LLM 作为 MCP Tool**：
```java
// LLMTool - 暴露 LLM 能力给 MCP
public class LLMTool implements ToolExecutor {
    // llm.chat - 聊天
    // llm.embed - 文本嵌入
    // llm.complete - 文本补全
}
```

**Agent 使用 LLM**：
```java
// AI Agent 通过 LLM 进行推理
OFAAgent agent = new OFAAgent.Builder(context)
    .llmProvider(new CloudLLMProvider("https://api.openai.com/v1", "sk-xxx"))
    .fallbackLLMProvider(new LocalLLMProvider("gemma-2b.tflite"))
    .build();

// 自动工具调用
agent.getAIAgentInterface().chat("今天天气怎么样？帮我拍照记录");
// LLM 自动决定调用 weather.query 和 camera.capture
```

## 文件结构

```
src/sdk/android/src/main/java/com/ofa/agent/
├── llm/
│   ├── LLMProvider.java           # 接口定义
│   ├── LLMConfig.java             # 配置
│   ├── LLMResponse.java           # 响应
│   ├── LLMStats.java              # 统计
│   ├── Message.java               # 消息结构
│   ├── StreamCallback.java        # 流式回调
│   │
│   ├── cloud/
│   │   ├── CloudLLMProvider.java  # 云端实现
│   │   ├── OpenAIClient.java      # OpenAI 客户端
│   │   ├── ClaudeClient.java      # Claude 客户端
│   │   └── OpenAICompatibleClient.java  # 兼容客户端
│   │
│   ├── local/
│   │   ├── LocalLLMProvider.java  # 本地实现
│   │   ├── TFLiteEngine.java      # TF Lite 引擎
│   │   ├── ModelManager.java      # 模型管理
│   │   └── Tokenizer.java         # 分词器
│   │
│   ├── orchestrator/
│   │   ├── LLMOrchestrator.java   # 编排器
│   │   ├── FailoverStrategy.java  # 故障转移
│   │   └── LoadBalancer.java      # 负载均衡
│   │
│   └── tool/
│       ├── LLMChatTool.java       # LLM 聊天工具
│       ├── LLMEmbedTool.java      # Embedding 工具
│       └── LLMCompleteTool.java   # 补全工具
│
├── agent/
│   └── OFAAgent.java              # 添加 LLM 支持
```

## 配置示例

### 云端 LLM 配置

```java
LLMConfig cloudConfig = LLMConfig.builder()
    .providerType(LLMProvider.Type.CLOUD)
    .endpoint("https://api.openai.com/v1")
    .apiKey("sk-xxxxxxxx")
    .model("gpt-4-turbo-preview")
    .maxTokens(4096)
    .temperature(0.7f)
    .timeout(30000)
    .build();
```

### 本地 LLM 配置

```java
LLMConfig localConfig = LLMConfig.builder()
    .providerType(LLMProvider.Type.LOCAL)
    .modelPath("/data/local/tmp/gemma-2b.tflite")
    .tokenizerPath("/data/local/tmp/tokenizer.json")
    .maxTokens(2048)
    .temperature(0.7f)
    .threads(4)
    .gpuAccel(true)
    .build();
```

### Agent 配置

```java
OFAAgent agent = new OFAAgent.Builder(context)
    // 基础配置
    .agentId("android-001")
    .centerAddress("192.168.1.100")
    .centerPort(9090)

    // 云端 LLM (主)
    .llmProvider(new CloudLLMProvider(cloudConfig))

    // 本地 LLM (备用/离线)
    .fallbackLLMProvider(new LocalLLMProvider(localConfig))

    // 自动切换
    .autoLLMFailover(true)

    // 离线等级
    .offlineLevel(OfflineLevel.L3)
    .build();
```

## 离线场景

### L1 - 完全离线
- 使用 LocalLLMProvider
- 所有工具本地执行
- 无网络依赖

### L2 - 局域网
- 使用 LocalLLMProvider
- P2P Agent 协作
- 可访问局域网资源

### L3 - 弱网
- 优先 LocalLLMProvider
- 缓存云端请求
- 网络恢复后同步

### L4 - 在线
- 优先 CloudLLMProvider
- LocalLLMProvider 降级
- 完整功能可用

## 依赖

```gradle
dependencies {
    // TensorFlow Lite
    implementation 'org.tensorflow:tensorflow-lite:2.14.0'
    implementation 'org.tensorflow:tensorflow-lite-gpu:2.14.0'
    implementation 'org.tensorflow:tensorflow-lite-support:0.4.4'

    // OkHttp (云端 LLM)
    implementation 'com.squareup.okhttp3:okhttp:4.12.0'
    implementation 'com.squareup.okhttp3:okhttp-sse:4.12.0'

    // JSON
    implementation 'com.google.code.gson:gson:2.10.1'
}
```

## 实现优先级

| 优先级 | 模块 | 工作量 |
|--------|------|--------|
| P0 | LLMProvider 接口 | 2h |
| P0 | CloudLLMProvider (OpenAI) | 4h |
| P1 | LocalLLMProvider (TFLite) | 8h |
| P1 | LLMOrchestrator | 4h |
| P2 | LLM Tools (MCP) | 2h |
| P2 | Claude Client | 2h |
| P3 | GPU 加速 | 4h |
| P3 | 模型下载管理 | 4h |

**总计**: ~30h

---

*创建时间: 2026-04-01*
# OFA Android SDK 架构设计文档

## 1. 概述

OFA Android SDK 是一个完整的智能 Agent 开发框架，支持：

- **独立运行**: 完全本地执行，无需网络
- **云端协作**: 连接 OFA Center，接收远程任务
- **Agent 通信**: 与其他 Agent 直接通信协作
- **智能自动化**: UI 自动化、意图理解、技能编排
- **社交通知**: 智能消息分发到多个社交渠道

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Application Layer                               │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      OFAAndroidAgent                             │   │
│  │                    (Unified Entry Point)                          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────┤
│                          Mode Layer                                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     AgentModeManager                              │   │
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────────┐      │   │
│  │  │  STANDALONE   │ │   CONNECTED   │ │      HYBRID       │      │   │
│  │  │  独立运行     │ │   连接Center  │ │   混合模式        │      │   │
│  │  └───────────────┘ └───────────────┘ └───────────────────┘      │   │
│  └─────────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────┤
│                        Communication Layer                               │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────────────┐ │
│  │    Center     │ │     Peer      │ │          Local                │ │
│  │  Connection   │ │    Network    │ │     Execution Engine          │ │
│  │   (gRPC)      │ │  (NSD + P2P)  │ │                               │ │
│  └───────────────┘ └───────────────┘ └───────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────┤
│                        Capability Layer                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │
│  │   Intent    │ │   Skill     │ │ Automation  │ │    Social       │   │
│  │   Engine    │ │  Executor   │ │ Orchestrator│ │  Orchestrator   │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────┘   │
├─────────────────────────────────────────────────────────────────────────┤
│                          AI Layer                                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │
│  │  LocalAI    │ │    MAB      │ │   Smart     │ │  Operation      │   │
│  │   Engine    │ │  Decision   │ │  Decision   │ │  Recommender    │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────┘   │
├─────────────────────────────────────────────────────────────────────────┤
│                        Foundation Layer                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐   │
│  │   Memory    │ │     MCP     │ │    Tool     │ │     LLM         │   │
│  │   System    │ │   Server    │ │  Registry   │ │   Provider      │   │
│  │  (L1/L2/L3) │ │             │ │             │ │ (Cloud + Local) │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 3. 核心模块详解

### 3.1 OFAAndroidAgent (统一入口)

**职责**: 提供 Agent 的创建、初始化、任务执行、状态管理的统一接口

**设计原则**:
- Builder 模式创建，支持灵活配置
- 单例模式，全局唯一实例
- 支持运行时模式切换

```java
public class OFAAndroidAgent {
    // 核心组件
    private final AgentProfile profile;           // Agent 身份
    private final AgentModeManager modeManager;   // 模式管理
    private final LocalExecutionEngine localEngine; // 本地执行

    // 子系统
    private UserMemoryManager memoryManager;
    private AutomationOrchestrator automationOrchestrator;
    private SocialOrchestrator socialOrchestrator;

    // 创建方式
    OFAAndroidAgent agent = new OFAAndroidAgent.Builder(context)
        .runMode(AgentProfile.RunMode.HYBRID)
        .center("center.ofa.com", 9090)
        .enableAutomation(true)
        .enableSocial(true)
        .enablePeerNetwork(true)
        .build();
}
```

### 3.2 AgentModeManager (模式管理)

**职责**: 管理运行模式，智能任务路由

**三种模式对比**:

| 特性 | STANDALONE | CONNECTED | HYBRID |
|------|------------|-----------|--------|
| Center 连接 | ❌ | ✅ 必需 | ✅ 可选 |
| Peer 网络 | ❌ | ✅ | ✅ |
| 本地执行 | ✅ | ✅ | ✅ |
| 远程任务 | ❌ | ✅ | ✅ |
| 云端 LLM | ❌ | ✅ | 降级 |
| 网络依赖 | 无 | 必需 | 可选 |

**任务路由逻辑 (HYBRID 模式)**:

```
TaskRequest
     ↓
AgentModeManager.executeTask()
     ↓
┌────────────────────────────────────┐
│         Decision Logic             │
│  1. Can execute offline?           │
│  2. Needs cloud capability?        │
│  3. Is Center available?           │
│  4. Is network available?          │
└────────────────────────────────────┘
     ↓
┌────┴────┐
↓         ↓
Local    Center
Engine   Connection
```

### 3.3 LocalExecutionEngine (本地执行引擎)

**职责**: 执行所有本地任务

**支持的任务类型**:

| 类型 | 处理器 | 说明 |
|------|--------|------|
| `intent` | IntentEngine | 意图识别 |
| `skill` | CompositeSkillExecutor | 技能执行 |
| `automation` | AutomationOrchestrator | UI 自动化 |
| `social` | SocialOrchestrator | 社交通知 |
| `memory` | UserMemoryManager | 记忆操作 |
| `nl` | NaturalLanguageProcessor | 自然语言处理 |

### 3.4 CenterConnection (Center 连接)

**职责**: 与 OFA Center 的通信

**协议**: gRPC 双向流

**消息类型**:
- `RegisterRequest`: Agent 注册
- `HeartbeatRequest`: 心跳保活
- `TaskAssignment`: 任务分配
- `TaskResult`: 任务结果
- `ConfigUpdate`: 配置更新
- `BroadcastMessage`: 广播消息

**连接流程**:

```
Agent                           Center
  │                               │
  │──── RegisterRequest ─────────>│
  │<─── RegisterResponse ─────────│
  │                               │
  │<═══ Bidirectional Stream ════>│
  │                               │
  │──── HeartbeatRequest ────────>│ (每30秒)
  │<─── HeartbeatResponse ────────│
  │                               │
  │<─── TaskAssignment ───────────│
  │──── TaskResult ──────────────>│
  │                               │
```

### 3.5 PeerNetwork (Agent 间通信)

**职责**: 发现和通信本地网络中的其他 Agent

**技术栈**:
- **发现**: NSD (Network Service Discovery) / mDNS
- **通信**: TCP Socket P2P

**服务发现流程**:

```
Agent A                          Agent B
   │                               │
   │<── NSD Registration ─────────>│
   │                               │
   │<── NSD Discovery ────────────>│
   │                               │
   │<── Service Found ────────────>│
   │                               │
   │<── Service Resolved ─────────>│
   │   (Host:Port, AgentInfo)      │
   │                               │
   │<════ TCP Connection ═════════>│
   │                               │
   │<── P2P Message ──────────────>│
   │                               │
```

### 3.6 AgentProfile (Agent 身份)

**职责**: 定义 Agent 的身份、能力、配置

**核心属性**:

```java
public class AgentProfile {
    // 身份
    String agentId;           // 唯一标识
    String name;              // 显示名称
    AgentType type;           // FULL/MOBILE/LITE/IOT/EDGE

    // 能力列表
    List<Capability> capabilities;

    // 运行配置
    RunMode preferredRunMode;  // STANDALONE/CONNECTED/HYBRID
    boolean allowRemoteControl;
    boolean allowPeerCommunication;

    // 状态
    AgentStatus status;        // OFFLINE/ONLINE/BUSY/IDLE
}
```

**能力类型**:

| 能力 ID | 名称 | 说明 |
|---------|------|------|
| `ui_automation` | UI 自动化 | AccessibilityService |
| `social_notification` | 社交通知 | 多渠道消息分发 |
| `local_llm` | 本地 LLM | TFLite 推理 |
| `cloud_llm` | 云端 LLM | API 调用 |
| `intent_understanding` | 意图理解 | NLP 处理 |
| `memory_system` | 记忆系统 | L1/L2/L3 存储 |
| `skill_orchestration` | 技能编排 | 多步骤任务 |
| `contact_access` | 联系人访问 | 通讯录读取 |

## 4. 数据流设计

### 4.1 任务执行流程

```
用户输入 (自然语言/技能调用/自动化指令)
     ↓
TaskRequest 创建
     ↓
OFAAndroidAgent.execute()
     ↓
AgentModeManager.executeTask()
     ↓
┌────────────────────────────────────┐
│        模式决策                     │
│  STANDALONE → LocalExecutionEngine │
│  CONNECTED → CenterConnection      │
│  HYBRID → 智能路由                 │
└────────────────────────────────────┘
     ↓
具体执行器 (Intent/Skill/Automation/Social)
     ↓
TaskResult 返回
     ↓
结果处理 (UI展示/回调通知/状态更新)
```

### 4.2 社交通知流程

```
用户消息: "约张三明天吃饭"
     ↓
MessageClassifier.classify()
     ↓
ClassificationResult {
    type: "invitation",
    urgency: LOW,
    recommendedChannel: "wechat"
}
     ↓
ChannelSelector.selectBestChannel()
     ↓
检查渠道可用性 (微信已安装? 联系人有微信ID?)
     ↓
MessageSender.send(channel="wechat", recipient="张三")
     ↓
通过 AccessibilityService 自动化发送
     ↓
DeliveryRecord {
    success: true,
    channel: "wechat",
    duration: 2500ms
}
```

### 4.3 记忆系统流程

```
用户操作 (选择"喜茶"作为首选奶茶店)
     ↓
UserMemoryManager.set("preferred_tea_shop", "喜茶")
     ↓
L1 Cache: 内存缓存 (毫秒级访问)
     ↓
L2 Database: Room 持久化
     ↓
L3 Archive: 文件归档 (可选)
     ↓
下次操作时:
UserMemoryManager.getSuggestions("preferred_tea")
     ↓
返回: [{key: "preferred_tea_shop", value: "喜茶", score: 0.95}]
```

## 5. 线程模型

```
┌─────────────────────────────────────────────────────────────────┐
│                        Main Thread (UI)                          │
│  - 用户交互                                                      │
│  - 回调通知                                                      │
│  - 状态更新                                                      │
└─────────────────────────────────────────────────────────────────┘
                              ↑
                              │ Handler
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Background Threads                           │
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  gRPC Stream    │  │  Task Executor  │  │  Peer Network   │ │
│  │  (Center连接)   │  │  (任务执行)     │  │  (P2P通信)      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  Automation     │  │  AI Inference   │  │  Memory Sync    │ │
│  │  (UI操作)       │  │  (本地推理)     │  │  (数据同步)     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 6. 权限模型

### 6.1 必需权限

```xml
<!-- 网络通信 -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

<!-- UI 自动化 -->
<uses-permission android:name="android.permission.SYSTEM_ALERT_WINDOW" />

<!-- 前台服务 -->
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
```

### 6.2 可选权限

```xml
<!-- 社交通知 -->
<uses-permission android:name="android.permission.READ_CONTACTS" />
<uses-permission android:name="android.permission.CALL_PHONE" />
<uses-permission android:name="android.permission.SEND_SMS" />

<!-- 设备能力 -->
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.BLUETOOTH" />
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.READ_CALENDAR" />

<!-- 文件访问 -->
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
```

### 6.3 特殊权限

```xml
<!-- 无障碍服务 (UI 自动化必需) -->
<uses-permission android:name="android.permission.BIND_ACCESSIBILITY_SERVICE" />

<!-- 系统级权限 (ROM 内置) -->
<uses-permission android:name="android.permission.INSTALL_PACKAGES" />
<uses-permission android:name="android.permission.WRITE_SECURE_SETTINGS" />
```

## 7. 扩展机制

### 7.1 自定义工具

```java
// 1. 实现 ToolExecutor 接口
public class MyCustomTool implements ToolExecutor {
    @Override
    public ToolResult execute(Map<String, String> params) {
        // 实现工具逻辑
        return ToolResult.success(result);
    }

    @Override
    public long getEstimatedTimeMs() {
        return 1000;
    }
}

// 2. 注册到 ToolRegistry
ToolRegistry registry = agent.getToolRegistry();
registry.register(
    ToolDefinition.create(
        "my.custom.tool",
        "自定义工具描述",
        "param1", "string", true, "参数1"
    ),
    new MyCustomTool()
);
```

### 7.2 自定义技能

```java
// 1. 创建技能定义
SkillDefinition skill = new SkillDefinition.Builder("my_skill")
    .name("我的技能")
    .description("技能描述")
    .addStep(SkillStep.builder("TOOL")
        .toolId("my.custom.tool")
        .param("param1", "$input")
        .build())
    .addStep(SkillStep.builder("DELAY")
        .delayMs(1000)
        .build())
    .build();

// 2. 注册到 SkillRegistry
SkillRegistry registry = SkillRegistry.getInstance(context);
registry.register(skill);
```

### 7.3 自定义 App 适配器

```java
// 1. 继承 BaseAppAdapter
public class MyAppAdapter extends BaseAppAdapter {
    public MyAppAdapter(Context context) {
        super(context, "com.example.myapp");
    }

    @Override
    public boolean search(AutomationEngine engine, String query) {
        // 实现搜索逻辑
    }

    @Override
    public boolean selectProduct(AutomationEngine engine, String productName) {
        // 实现选择商品逻辑
    }
}

// 2. 注册到 AppAdapterManager
AppAdapterManager manager = orchestrator.getAdapterManager();
manager.registerAdapter(new MyAppAdapter(context));
```

## 8. 性能优化

### 8.1 内存优化

- L1 Cache 使用 LRU 策略，限制大小
- 图片使用缩略图，避免大图加载
- 及时释放不用的资源

### 8.2 电池优化

- 使用 WorkManager 进行后台任务
- 批量处理网络请求
- 合理使用 WakeLock

### 8.3 网络优化

- 请求去重和合并
- 离线缓存策略
- 指数退避重试

## 9. 安全考虑

### 9.1 数据安全

- 敏感数据加密存储
- 使用 Android Keystore
- HTTPS 通信

### 9.2 权限最小化

- 按需申请权限
- 运行时权限检查
- 权限被拒绝时的降级策略

### 9.3 隐私保护

- 用户数据本地优先
- 明确的数据使用说明
- 用户可删除所有数据
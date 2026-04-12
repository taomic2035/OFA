# OFA Android SDK 架构文档

## 一、概述

OFA Android SDK 是去中心化分布式 Agent 系统的设备端实现，遵循核心愿景 **"万物皆为我所用，万物皆是我"**。

### 设计理念

| 角色 | 职责 | 特性 |
|------|------|------|
| **Center** | 永远在线的灵魂载体 | 最终基准、冲突仲裁、数据纠偏 |
| **Agent** | 设备端载体 (本 SDK) | 可离线、可更换、定期同步 |

### SDK 定位
- 作为 Agent 设备端载体
- 管理设备状态和连接
- 实现场景感知和联动
- 维护记忆系统和技能执行

---

## 二、核心架构

### 模块组成

```
SDK (Java)
├── core/                       # 核心组件
│   ├── OFAAndroidAgent          # Agent 主类 - 统一入口
│   ├── AgentProfile             # 设备画像 - 能力和状态
│   ├── AgentModeManager         # 模式管理 - STANDALONE/SYNC
│   ├── PeerNetwork              # P2P 网络 - 设备发现
│   └── CenterConnection         # Center 连接 - WebSocket/gRPC
│
├── distributed/                # 分布式系统
│   ├── DistributedOrchestrator # 统一协调入口
│   ├── SceneDetector           # 场景感知检测
│   ├── CrossDeviceRouter       # 跨设备消息路由
│   ├── EventBus                # 事件订阅/发布
│   └── HealthDataBridge        # 健康数据桥接
│
├── social/                     # 社通系统
│   ├── MessageClassifier       # 消息分类
│   ├── ChannelSelector         # 渠道选择
│   └── SocialOrchestrator      # 社交协调
│
├── automation/                 # UI 自动化
│   ├── AutomationEngine        # 自动化引擎
│   └── AccessibilityService    # 无障碍服务
│
├── memory/                     # 记忆系统
│   ├── UserMemoryManager       # 记忆管理
│   ├── L1Cache                 # 短期记忆 (1小时)
│   ├── L2Cache                 # 中期记忆 (1天)
│   └── L3Cache                 # 长期记忆 (永久)
│
├── intent/                     # 意图理解
│   ├── IntentParser            # 意图解析
│   └── ContextAnalyzer         # 上下文分析
│
├── skill/                      # 技能编排
│   ├── SkillRegistry           # 技能注册
│   ├── SkillExecutor           # 技能执行
│   └── SkillOrchestrator       # 技能编排
│
├── mcp/                        # MCP 协议
│   ├── MCPClient               # MCP 客户端
│   └── MCPSession              # MCP 会话
│
├── tool/                       # 工具系统
│   ├── ToolRegistry            # 工具注册
│   ├── ToolExecutor            # 工具执行
│   └── ToolResult              # 工具结果
│
├── ai/                         # AI Agent 接口
│   ├── AIClient                # AI 客户端
│   └── AIResponse              # AI 响应
│
├── llm/                        # LLM 提供者
│   ├── LLMProvider             # LLM 提供者接口
│   ├── CloudLLM                # 云端 LLM (Claude/GPT)
│   └── LocalLLM                # 本地 LLM (TFLite)
│
└── grpc/                       # gRPC 通信
    ├── GRPCClient              # gRPC 客户端
    └── GRPCStream              # gRPC 流式通信
```

---

## 三、核心组件详解

### 3.1 OFAAndroidAgent (主入口)

**职责**: Agent 系统统一入口，协调所有组件

```java
public class OFAAndroidAgent {
    // 核心组件
    private AgentProfile profile;
    private AgentModeManager modeManager;
    private CenterConnection centerConnection;
    private PeerNetwork peerNetwork;
    
    // 业务组件
    private DistributedOrchestrator orchestrator;
    private UserMemoryManager memoryManager;
    private SkillOrchestrator skillOrchestrator;
    
    // 初始化
    public void initialize(Context context, AgentConfig config);
    
    // 模式切换
    public void setMode(AgentMode mode);
    
    // 连接 Center
    public void connectCenter(String centerAddress);
    
    // 断开连接
    public void disconnect();
    
    // 获取状态
    public AgentStatus getStatus();
}
```

### 3.2 AgentProfile (设备画像)

**职责**: 设备能力、状态、身份管理

```java
public class AgentProfile {
    // 设备信息
    private String agentId;
    private String deviceType;      // phone, watch, glasses, tablet
    private String deviceName;
    private String osVersion;
    
    // 能力
    private List<String> capabilities;  // voice, display, camera, health
    private Map<String, Object> resources;  // battery, network, cpu
    
    // 身份绑定
    private String identityId;
    
    // 状态
    private AgentStatus status;     // online, offline, busy, error
    private long lastHeartbeat;
}
```

### 3.3 AgentModeManager (模式管理)

**职责**: 运行模式切换和管理

```java
public enum AgentMode {
    STANDALONE,  // 完全独立，无 Center 连接
    SYNC         // 定期与 Center 同步
}

public class AgentModeManager {
    // 设置模式
    public void setMode(AgentMode mode);
    
    // 获取当前模式
    public AgentMode getCurrentMode();
    
    // 同步数据
    public void syncWithCenter();
    
    // 处理冲突
    public void handleConflict(ConflictData data);
}
```

### 3.4 CenterConnection (Center 连接)

**职责**: WebSocket/gRPC 连接管理

```java
public class CenterConnection {
    // WebSocket 连接
    private WebSocketClient wsClient;
    
    // 连接
    public void connect(String centerAddress);
    
    // 注册 Agent
    public void register(AgentProfile profile);
    
    // 发送心跳
    public void sendHeartbeat();
    
    // 发送消息
    public void sendMessage(WebSocketMessage message);
    
    // 接收消息
    public void onMessageReceived(WebSocketMessage message);
    
    // 断线重连
    public void reconnect();
}
```

---

## 四、分布式系统详解

### 4.1 DistributedOrchestrator (统一协调入口)

**职责**: 协调场景感知、跨设备路由、事件系统

```java
public class DistributedOrchestrator {
    // 组件
    private SceneDetector sceneDetector;
    private CrossDeviceRouter router;
    private EventBus eventBus;
    private HealthDataBridge healthBridge;
    
    // 场景检测
    public void detectScene(SceneContext context);
    
    // 路由消息
    public void routeMessage(Message message);
    
    // 发布事件
    public void publishEvent(String eventType, Object data);
    
    // 订阅事件
    public void subscribe(String eventType, EventHandler handler);
    
    // 健康数据
    public void reportHealthData(HealthData data);
}
```

### 4.2 SceneDetector (场景感知)

**职责**: 检测用户当前场景

```java
public enum SceneType {
    RUNNING,     // 跑步
    WALKING,     // 步行
    DRIVING,     // 驾驶
    MEETING,     // 会议
    SLEEPING,    // 睡眠
    EXERCISE,    // 运动
    WORK,        // 工作
    HOME,        // 家庭
    TRAVEL       // 旅行
}

public class SceneDetector {
    // 检测场景
    public SceneType detect(Map<String, Object> context);
    
    // 检测置信度
    public float getConfidence(SceneType type);
    
    // 场景变化回调
    public void onSceneChanged(SceneChangeListener listener);
}
```

### 4.3 CrossDeviceRouter (跨设备路由)

**职责**: 根据场景路由消息到合适设备

```java
public class CrossDeviceRouter {
    // 路由规则
    private List<RoutingRule> rules;
    
    // 路由消息
    public String route(Message message, SceneType scene);
    
    // 添加规则
    public void addRule(RoutingRule rule);
    
    // 默认规则
    // - RUNNING: 路由到手机显示
    // - MEETING: 眼镜静默提醒
    // - DRIVING: 延迟非紧急消息
    // - HEALTH_ALERT: 广播到所有设备
}
```

### 4.4 EventBus (事件系统)

**职责**: 组件间事件订阅/发布

```java
public class EventBus {
    // 发布事件
    public void publish(String eventType, Object data);
    
    // 订阅事件
    public void subscribe(String eventType, EventHandler handler);
    
    // 取消订阅
    public void unsubscribe(String eventType, EventHandler handler);
    
    // 标准事件类型
    // - SCENE_CHANGED: 场景变化
    // - HEALTH_ALERT: 健康告警
    // - MESSAGE_RECEIVED: 消息接收
    // - DEVICE_CONNECTED: 设备连接
    // - DEVICE_DISCONNECTED: 设备断开
}
```

---

## 五、记忆系统详解

### 5.1 UserMemoryManager (记忆管理)

**职责**: 三级记忆存储管理

```java
public class UserMemoryManager {
    // 三级存储
    private L1Cache l1Cache;  // 短期记忆 (1小时)
    private L2Cache l2Cache;  // 中期记忆 (1天)
    private L3Cache l3Cache;  // 长期记忆 (永久)
    
    // 存储记忆
    public void store(Memory memory);
    
    // 查询记忆
    public List<Memory> query(MemoryQuery query);
    
    // 归档记忆
    public void archive(String memoryId);
    
    // 同步到 Center
    public void syncToCenter();
    
    // 从 Center 恢复
    public void restoreFromCenter();
}
```

### 5.2 三级存储架构

| 级别 | 存储 | 持续时间 | 内容 |
|------|------|---------|------|
| L1 | 内存缓存 | 1小时 | 当前会话上下文 |
| L2 | 本地数据库 | 1天 | 今日交互记录 |
| L3 | 持久存储 | 永久 | 重要记忆归档 |

---

## 六、技能系统详解

### 6.1 SkillOrchestrator (技能编排)

**职责**: 技能注册和执行编排

```java
public class SkillOrchestrator {
    // 技能注册表
    private SkillRegistry registry;
    
    // 执行技能
    public SkillResult execute(String skillId, Map<String, Object> input);
    
    // 编排多个技能
    public List<SkillResult> orchestrate(List<String> skillIds, Map<String, Object> input);
    
    // 注册技能
    public void register(Skill skill);
}
```

### 6.2 技能类型

| 类型 | 说明 | 示例 |
|------|------|------|
| text.process | 文本处理 | 分析、摘要、翻译 |
| voice.recognize | 语音识别 | 转文字、意图提取 |
| health.monitor | 健康监测 | 心率、血压、睡眠 |
| notification.send | 通知发送 | 提醒、告警、消息 |
| automation.execute | 自动化执行 | UI 操作、任务执行 |

---

## 七、WebSocket 通信协议

### 7.1 消息类型

| 类型 | 方向 | 说明 |
|------|------|------|
| Register | Agent → Center | Agent 注册 |
| RegisterAck | Center → Agent | 注册确认 |
| Heartbeat | Agent ↔ Center | 心跳维护 |
| StateUpdate | Center → Agent | 状态推送 |
| TaskAssign | Center → Agent | 任务分配 |
| TaskResult | Agent → Center | 任务结果 |
| SyncRequest | Agent → Center | 同步请求 |
| SyncResponse | Center → Agent | 同步响应 |
| BehaviorReport | Agent → Center | 行为上报 |
| EmotionUpdate | Center → Agent | 情绪更新 |
| Error | Center → Agent | 错误响应 |

### 7.2 注册流程

```
Agent                           Center
  │                               │
  │  1. Register (agentId, type)  │
  │ ──────────────────────────────>│
  │                               │
  │  2. RegisterAck (sessionId)   │
  │ <─────────────────────────────│
  │                               │
  │  3. Heartbeat (每30秒)        │
  │ ──────────────────────────────>│
  │                               │
```

---

## 八、使用示例

### 8.1 基础使用

```java
// 创建 Agent
OFAAndroidAgent agent = new OFAAndroidAgent();

// 配置
AgentConfig config = new AgentConfig.Builder()
    .setDeviceType("phone")
    .setCenterAddress("ws://center.example.com:8080/ws")
    .setMode(AgentMode.SYNC)
    .build();

// 初始化
agent.initialize(context, config);

// 连接 Center
agent.connectCenter(config.getCenterAddress());

// 获取状态
AgentStatus status = agent.getStatus();
```

### 8.2 场景检测

```java
// 获取分布式协调器
DistributedOrchestrator orchestrator = agent.getOrchestrator();

// 订阅场景变化
orchestrator.subscribe("SCENE_CHANGED", (event) -> {
    SceneType scene = (SceneType) event.getData();
    Log.d("Scene", "Current scene: " + scene);
});

// 手动检测场景
Map<String, Object> context = new HashMap<>();
context.put("activity_type", "running");
context.put("heart_rate", 140);

SceneType scene = orchestrator.detectScene(context);
```

### 8.3 健康数据上报

```java
// 获取健康数据桥接
HealthDataBridge healthBridge = orchestrator.getHealthBridge();

// 上报心率
HealthData heartRate = new HealthData.Builder()
    .setType("heart_rate")
    .setValue(120)
    .setTimestamp(System.currentTimeMillis())
    .build();

healthBridge.report(heartRate);

// 上报血压
HealthData bp = new HealthData.Builder()
    .setType("blood_pressure")
    .setValue(Map.of("systolic", 125, "diastolic", 85))
    .build();

healthBridge.report(bp);
```

### 8.4 记忆存储

```java
// 获取记忆管理器
UserMemoryManager memoryManager = agent.getMemoryManager();

// 存储记忆
Memory memory = new Memory.Builder()
    .setType("interaction")
    .setContent("用户询问天气情况")
    .setTimestamp(System.currentTimeMillis())
    .build();

memoryManager.store(memory);

// 查询记忆
MemoryQuery query = new MemoryQuery.Builder()
    .setType("interaction")
    .setTimeRange(startTime, endTime)
    .build();

List<Memory> memories = memoryManager.query(query);
```

---

## 九、配置说明

### 9.1 AgentConfig

```java
public class AgentConfig {
    // 必填
    private String deviceType;       // phone, watch, glasses
    private String centerAddress;    // Center WebSocket 地址
    
    // 可选
    private AgentMode mode;          // 默认 SYNC
    private int heartbeatInterval;   // 默认 30秒
    private int syncInterval;        // 默认 5分钟
    private boolean enableCache;     // 默认 true
    private String identityId;       // 绑定身份
}
```

### 9.2 场景检测配置

```java
public class SceneDetectorConfig {
    private float minConfidence;     // 最小置信度 0.7
    private long detectionInterval;  // 检测间隔 5秒
    private boolean enableAutoDetect; // 自动检测 true
}
```

---

## 十、常见问题

### Q1: Agent 如何处理断线？
- 自动尝试重连 (最多 3 次)
- 重连成功后恢复心跳
- 离线期间数据缓存本地
- 重连后自动同步

### Q2: 场景检测准确率如何提升？
- 结合多数据源 (位置、活动、心率)
- 使用置信度阈值过滤
- 历史数据学习优化
- 用户反馈校准

### Q3: 记忆同步如何避免冲突？
- 版本号追踪
- Center 统一仲裁
- 最后修改优先
- 重要数据累加合并

### Q4: 多设备如何协同？
- EventBus 统一事件系统
- CrossDeviceRouter 智能路由
- Center 作为消息枢纽
- 场景感知动态切换

---

## 十一、版本历史

| 版本 | 功能 |
|------|------|
| v1.0.0 | 核心 Agent 框架 |
| v1.1.0 | 分布式协同系统 |
| v1.2.0 | Center 连接管理 |
| v2.0.0 | 身份同步系统 |
| v3.0.0 | 场景感知联动 |
| v7.0.0 | WebSocket 实时通信 |

---

*文档版本: v7.4.0*
*更新时间: 2026-04-11*
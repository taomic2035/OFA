# OFA API 文档

## 概述

OFA提供两种API接口：
- **REST API**: HTTP接口，端口8080
- **gRPC API**: 高性能RPC接口，端口9090

---

## REST API

### 基础信息

- 基础URL: `http://localhost:8080`
- 内容类型: `application/json`
- 认证: 无 (开发版本)

### 端点列表

#### 健康检查

```
GET /health
```

**响应:**

```json
{
    "status": "healthy",
    "version": "v{version}"
}
```

---

#### 获取Agent列表

```
GET /api/v1/agents
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| type | int | Agent类型过滤 |
| status | int | 状态过滤 |
| page | int | 页码，默认1 |
| page_size | int | 每页数量，默认20 |

**响应:**

```json
{
    "agents": [
        {
            "agent_id": "abc123",
            "status": 1,
            "capabilities": [
                {
                    "id": "text.process",
                    "name": "Text Process",
                    "version": "1.0.0"
                }
            ],
            "last_seen": 1711622400
        }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
}
```

---

#### 获取单个Agent

```
GET /api/v1/agents/{id}
```

**响应:**

```json
{
    "success": true,
    "agent": {
        "agent_id": "abc123",
        "status": 1,
        "capabilities": []
    }
}
```

---

#### 删除Agent

```
DELETE /api/v1/agents/{id}
```

**响应:**

```json
{
    "success": true
}
```

---

#### 提交任务

```
POST /api/v1/tasks
```

**请求体:**

```json
{
    "skill_id": "text.process",
    "input": "eyJ0ZXh0IjoiaGVsbG8iLCJvcGVyYXRpb24iOiJ1cHBlcmNhc2UifQ==",
    "target_agent": "",
    "priority": 0,
    "timeout_ms": 30000
}
```

> 注: input字段为Base64编码的JSON字符串

**输入JSON解码后:**

```json
{
    "text": "hello",
    "operation": "uppercase"
}
```

**响应:**

```json
{
    "success": true,
    "task_id": "task-abc123"
}
```

---

#### 获取任务列表

```
GET /api/v1/tasks
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| status | int | 状态过滤 |
| agent_id | string | Agent过滤 |
| page | int | 页码 |
| page_size | int | 每页数量 |

**响应:**

```json
{
    "tasks": [
        {
            "task_id": "task-abc123",
            "skill_id": "text.process",
            "status": 3,
            "output": "eyJyZXN1bHQiOiJIRUxMTyJ9",
            "created_at": 1711622400,
            "completed_at": 1711622401,
            "duration_ms": 100
        }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
}
```

---

#### 获取任务状态

```
GET /api/v1/tasks/{id}
```

**响应:**

```json
{
    "success": true,
    "task": {
        "task_id": "task-abc123",
        "skill_id": "text.process",
        "status": 3,
        "output": "eyJyZXN1bHQiOiJIRUxMTyJ9",
        "error": ""
    }
}
```

---

#### 取消任务

```
POST /api/v1/tasks/{id}/cancel
```

**请求体:**

```json
{
    "reason": "User requested"
}
```

**响应:**

```json
{
    "success": true
}
```

---

#### 发送消息

```
POST /api/v1/messages
```

**请求体:**

```json
{
    "from_agent": "agent-1",
    "to_agent": "agent-2",
    "action": "ping",
    "payload": "e30=",
    "ttl": 3600
}
```

**响应:**

```json
{
    "success": true,
    "msg_id": "msg-abc123"
}
```

---

#### 广播消息

```
POST /api/v1/messages/broadcast
```

**请求体:**

```json
{
    "from_agent": "agent-1",
    "action": "announcement",
    "payload": "e30=",
    "ttl": 3600
}
```

**响应:**

```json
{
    "success": true,
    "delivered_count": 5
}
```

---

#### 组播消息

```
POST /api/v1/messages/multicast
```

**请求体:**

```json
{
    "from_agent": "agent-1",
    "to_agents": ["agent-2", "agent-3"],
    "action": "notify",
    "payload": "e30="
}
```

**响应:**

```json
{
    "success": true,
    "delivered_count": 2
}
```

---

#### 获取技能列表

```
GET /api/v1/skills
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| category | string | 分类过滤 |

**响应:**

```json
{
    "skills": [
        {
            "id": "text.process",
            "name": "Text Process",
            "version": "1.0.0",
            "category": "text"
        },
        {
            "id": "json.process",
            "name": "JSON Process",
            "version": "1.0.0",
            "category": "data"
        }
    ]
}
```

---

#### 获取系统信息

```
GET /api/v1/system/info
```

**响应:**

```json
{
    "version": "v{version}",
    "uptime_seconds": 3600,
    "agent_count": 5,
    "task_count": 10
}
```

---

#### 获取系统指标

```
GET /api/v1/system/metrics
```

**响应:**

```json
{
    "metrics": {
        "agents_online": 5,
        "tasks_pending": 2,
        "tasks_completed": 100,
        "tasks_failed": 3
    }
}
```

---

## gRPC API

### 服务列表

| 服务 | 说明 |
|------|------|
| AgentService | Agent管理与任务执行 |
| MessageService | 消息通信 |
| ManagementService | 系统管理 |

### AgentService

```protobuf
service AgentService {
    rpc Connect(stream AgentMessage) returns (stream CenterMessage);
    rpc SubmitTask(SubmitTaskRequest) returns (SubmitTaskResponse);
    rpc GetTaskStatus(GetTaskStatusRequest) returns (GetTaskStatusResponse);
    rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
    rpc SubscribeTask(SubscribeTaskRequest) returns (stream TaskEvent);
    rpc RegisterCapabilities(RegisterCapabilitiesRequest) returns (RegisterCapabilitiesResponse);
    rpc GetCapabilities(GetCapabilitiesRequest) returns (GetCapabilitiesResponse);
}
```

### MessageService

```protobuf
service MessageService {
    rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
    rpc Broadcast(BroadcastRequest) returns (BroadcastResponse);
    rpc Multicast(MulticastRequest) returns (MulticastResponse);
    rpc Subscribe(SubscribeMessageRequest) returns (stream Message);
}
```

### ManagementService

```protobuf
service ManagementService {
    rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
    rpc GetAgent(GetAgentRequest) returns (GetAgentResponse);
    rpc DeleteAgent(DeleteAgentRequest) returns (DeleteAgentResponse);
    rpc ListTasks(ListTasksRequest) returns (ListTasksResponse);
    rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
    rpc ListSkills(ListSkillsRequest) returns (ListSkillsResponse);
    rpc InstallSkill(InstallSkillRequest) returns (InstallSkillResponse);
    rpc GetSystemInfo(GetSystemInfoRequest) returns (GetSystemInfoResponse);
    rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
}
```

---

## 枚举类型

### AgentType

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | FULL (完整版) |
| 2 | MOBILE (移动版) |
| 3 | LITE (轻量版) |
| 4 | IOT (物联网) |
| 5 | EDGE (边缘计算) |

### AgentStatus

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | ONLINE |
| 2 | BUSY |
| 3 | IDLE |
| 4 | OFFLINE |

### TaskStatus

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | PENDING |
| 2 | RUNNING |
| 3 | COMPLETED |
| 4 | FAILED |
| 5 | CANCELLED |
| 6 | TIMEOUT |

---

## 错误处理

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
    "error": "错误描述信息"
}
```

---

## 示例

### PowerShell 示例

```powershell
# 健康检查
Invoke-RestMethod -Uri "http://localhost:8080/health"

# 提交任务
$body = @{
    skill_id = "text.process"
    input = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('{"text":"hello","operation":"uppercase"}'))
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### curl 示例

```bash
# 健康检查
curl http://localhost:8080/health

# 提交任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"skill_id":"text.process","input":"eyJ0ZXh0IjoiaGVsbG8iLCJvcGVyYXRpb24iOiJ1cHBlcmNhc2UifQ=="}'
```

---

## Android SDK API

### 核心模块

| 模块 | 包路径 | 说明 |
|------|--------|------|
| Intent | `com.ofa.agent.intent` | 意图理解系统 |
| Skill | `com.ofa.agent.skill` | 技能编排系统 |
| Memory | `com.ofa.agent.memory` | 用户记忆系统 |
| Tool | `com.ofa.agent.tool` | 工具系统 |
| LLM | `com.ofa.agent.llm` | LLM集成 |
| MCP | `com.ofa.agent.mcp` | MCP协议 |

### Intent API

```java
// 获取意图注册表
IntentRegistry registry = IntentRegistry.getInstance();

// 创建意图引擎
IntentEngine engine = new IntentEngine(registry);

// 识别意图
UserIntent intent = engine.recognizeOne("帮我点一杯珍珠奶茶");
// → intentId="order_food", confidence=0.85, slots={item="珍珠奶茶", count="1"}

// 执行意图
TaskExecutor executor = new TaskExecutor(toolRegistry);
executor.execute(intent, context);
```

### Skill API

```java
// 创建技能定义
SkillDefinition skill = new SkillDefinition.Builder()
    .id("order_bubble_tea")
    .name("点奶茶")
    .step("launch_app", StepType.TOOL, "app.launch", Map.of("app", "美团"))
    .step("search", StepType.TOOL, "app.search", Map.of("query", "${drink_name}"))
    .step("select_sweetness", StepType.INPUT, "选择甜度", null)
    .step("confirm", StepType.CONFIRM, "确认下单", null)
    .build();

// 执行技能
CompositeSkillExecutor executor = new CompositeSkillExecutor(context, toolRegistry);
SkillResult result = executor.execute(skill, inputs).get();
```

### Memory API

```java
// 获取记忆管理器
UserMemoryManager memory = UserMemoryManager.getInstance(context);

// 记录偏好
memory.rememberPreference("bubble_tea.drink_name", "珍珠奶茶", "food",
    Map.of("sweetness", "五分糖", "ice", "少冰"));

// 获取推荐值
String recommended = memory.getRecommendedValue("bubble_tea.drink_name");
// → "珍珠奶茶" (使用最多)

// 获取智能默认值
SmartDefault defaults = memory.getSmartDefault("bubble_tea.drink_name");
// → recommended, lastUsed, mostUsed, confidence

// 导出记忆
memory.exportMemories(callback);
```

### Tool API

```java
// 获取工具注册表
ToolRegistry registry = ToolRegistry.getInstance();

// 执行工具
ToolResult result = registry.execute("app.launch", Map.of("app", "美团"));

// 注册自定义工具
registry.register(new ToolDefinition.Builder()
    .id("custom.action")
    .name("自定义动作")
    .handler((params, ctx) -> ToolResult.success("done"))
    .build());
```

### LLM API

```java
// 配置LLM
LLMConfig config = new LLMConfig.Builder()
    .provider(LLMProviderType.CLOUD)
    .apiKey("your-api-key")
    .model("claude-3-sonnet")
    .build();

// 获取LLM Orchestrator
LLMOrchestrator orchestrator = LLMOrchestrator.getInstance(config);

// 发送请求
LLMResponse response = orchestrator.chat("帮我点一杯奶茶");
```

### 三层记忆架构

```
┌─────────────────────────────────────┐
│         L1: MemoryCache             │
│   内存缓存 (LRU策略, <1ms访问)       │
└─────────────────────────────────────┘
                 ↓ 未命中
┌─────────────────────────────────────┐
│         L2: Room Database           │
│   持久化存储 (可靠存储)              │
└─────────────────────────────────────┘
                 ↓ 归档
┌─────────────────────────────────────┐
│         L3: MemoryArchive           │
│   文件归档 (冷数据备份)              │
└─────────────────────────────────────┘
```

### 12种步骤类型

| 类型 | 说明 | 示例 |
|------|------|------|
| TOOL | 执行工具 | 启动APP、发送消息 |
| INTENT | 触发意图识别 | 解析用户指令 |
| DELAY | 延时等待 | 等待3秒 |
| WAIT_FOR | 等待条件 | 等外卖送达 |
| CONDITION | 条件分支 | 判断是否需要支付 |
| ASSIGN | 变量赋值 | 设置默认值 |
| INPUT | 获取用户输入 | 选择甜度 |
| CONFIRM | 请求确认 | 确认下单 |
| NOTIFY | 发送通知 | 提醒外卖快到了 |
| PARALLEL | 并行执行 | 同时搜索多个APP |
| LOOP | 循环执行 | 每分钟检查状态 |
| SUB_SKILL | 调用子技能 | 执行支付流程 |

### 22个内置意图

| 类别 | 意图 |
|------|------|
| 查询 | weather_query, stock_query, news_query, traffic_query, price_query, search_query, info_query, location_query |
| 操作 | app_launch, app_close, call_contact, send_message, send_email, play_media, take_photo, set_timer |
| 设置 | setting_change, alarm_set, reminder_set, schedule_add |
| 其他 | order_food, control_device |

---

## v3.x 多设备协同 API (新增)

### 消息总线 API (v3.0.0)

#### 发送设备消息

```
POST /api/v3/messages/device
```

**请求体:**

```json
{
    "id": "msg-001",
    "from_agent": "agent-phone",
    "to_agent": "agent-watch",
    "identity_id": "identity-001",
    "type": "data",
    "priority": 2,
    "payload": {
        "key": "value"
    },
    "ttl": 3600
}
```

**响应:**

```json
{
    "success": true,
    "queued": true
}
```

#### 获取离线消息

```
GET /api/v3/messages/offline/{agent_id}
```

**响应:**

```json
{
    "messages": [
        {
            "id": "msg-001",
            "from_agent": "center",
            "payload": {}
        }
    ],
    "count": 1
}
```

---

### 设备状态同步 API (v3.1.0)

#### 上报设备状态

```
POST /api/v3/devices/{agent_id}/state
```

**请求体:**

```json
{
    "battery_level": 85,
    "battery_charging": false,
    "network_type": "wifi",
    "network_strength": 90,
    "scene": "meeting",
    "location": {
        "latitude": 39.9,
        "longitude": 116.4,
        "accuracy": 10
    }
}
```

**响应:**

```json
{
    "success": true,
    "synced": true
}
```

#### 获取设备状态

```
GET /api/v3/devices/{agent_id}/state
```

**响应:**

```json
{
    "state": {
        "battery_level": 85,
        "network_type": "wifi",
        "scene": "meeting",
        "online": true,
        "last_update": 1711622400
    }
}
```

#### 获取身份下所有设备状态

```
GET /api/v3/identities/{identity_id}/devices/states
```

**响应:**

```json
{
    "devices": [
        {
            "agent_id": "agent-phone",
            "state": {}
        },
        {
            "agent_id": "agent-watch",
            "state": {}
        }
    ]
}
```

#### 订阅状态变更

```
WebSocket /api/v3/devices/states/subscribe
```

**消息格式:**

```json
{
    "agent_id": "agent-phone",
    "changes": ["battery", "scene"],
    "old_state": {},
    "new_state": {}
}
```

---

### 场景感知路由 API (v3.2.0)

#### 路由消息

```
POST /api/v3/route
```

**请求体:**

```json
{
    "identity_id": "identity-001",
    "message_type": "social",
    "priority": 2,
    "scene": "running",
    "payload": {}
}
```

**响应:**

```json
{
    "target_agents": ["agent-watch"],
    "action": "deliver",
    "reason": "Scene running -> route to watch"
}
```

#### 配置路由规则

```
POST /api/v3/route/rules
```

**请求体:**

```json
{
    "identity_id": "identity-001",
    "scenes": ["running", "walking"],
    "message_types": ["social", "message"],
    "target_device_types": ["watch"],
    "action": "deliver",
    "priority": 100
}
```

**响应:**

```json
{
    "success": true,
    "rule_id": "rule-001"
}
```

#### 获取路由规则

```
GET /api/v3/route/rules/{identity_id}
```

---

### 任务协同 API (v3.3.0)

#### 创建协同任务

```
POST /api/v3/tasks/collaborative
```

**请求体:**

```json
{
    "identity_id": "identity-001",
    "skill_id": "data.collect",
    "split_strategy": "parallel",
    "merge_strategy": "aggregate",
    "target_agents": ["agent-phone", "agent-watch", "agent-tablet"],
    "input": {},
    "constraints": {
        "max_sub_tasks": 3,
        "timeout_per_task": 30000,
        "min_success_count": 2
    }
}
```

**响应:**

```json
{
    "success": true,
    "task_id": "collab-001",
    "sub_tasks": [
        {
            "sub_task_id": "sub-001",
            "agent_id": "agent-phone"
        },
        {
            "sub_task_id": "sub-002",
            "agent_id": "agent-watch"
        }
    ]
}
```

#### 获取子任务状态

```
GET /api/v3/tasks/collaborative/{task_id}/subtasks
```

**响应:**

```json
{
    "sub_tasks": [
        {
            "sub_task_id": "sub-001",
            "status": "completed",
            "result": {}
        },
        {
            "sub_task_id": "sub-002",
            "status": "running"
        }
    ],
    "progress": 0.5
}
```

#### 上报子任务结果

```
POST /api/v3/tasks/collaborative/{task_id}/subtasks/{sub_task_id}/result
```

**请求体:**

```json
{
    "agent_id": "agent-phone",
    "success": true,
    "result": {
        "data": "collected"
    },
    "duration_ms": 1000
}
```

#### 获取合并结果

```
GET /api/v3/tasks/collaborative/{task_id}/result
```

**响应:**

```json
{
    "success": true,
    "result": {
        "merged": "aggregated result"
    },
    "sub_task_count": 3,
    "success_count": 2
}
```

---

### 跨设备通知 API (v3.4.0)

#### 创建通知

```
POST /api/v3/notifications
```

**请求体:**

```json
{
    "identity_id": "identity-001",
    "type": "message",
    "priority": "normal",
    "title": "New Message",
    "body": "You have a new message",
    "source_app": "com.example.app",
    "category": "social",
    "group_id": "group-001",
    "actions": [
        {
            "action_id": "open",
            "title": "Open",
            "type": "open"
        },
        {
            "action_id": "dismiss",
            "title": "Dismiss",
            "type": "dismiss"
        }
    ],
    "target_devices": [],
    "ttl": 3600
}
```

**响应:**

```json
{
    "success": true,
    "notification_id": "notif-001",
    "target_devices": ["agent-phone", "agent-watch"]
}
```

#### 获取通知

```
GET /api/v3/notifications/{notification_id}
```

**响应:**

```json
{
    "notification": {
        "notification_id": "notif-001",
        "status": "delivered",
        "delivered_to": ["agent-phone"],
        "read_by": [],
        "created_at": 1711622400
    }
}
```

#### 获取身份通知列表

```
GET /api/v3/identities/{identity_id}/notifications
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| status | string | 状态过滤 (pending/delivered/read/dismissed) |
| type | string | 类型过滤 |
| unread_only | bool | 只返回未读 |
| limit | int | 数量限制 |

#### 标记通知已读

```
POST /api/v3/notifications/{notification_id}/read
```

**请求体:**

```json
{
    "agent_id": "agent-phone"
}
```

#### 标记通知忽略

```
POST /api/v3/notifications/{notification_id}/dismiss
```

**请求体:**

```json
{
    "agent_id": "agent-phone"
}
```

#### 全部标记已读

```
POST /api/v3/identities/{identity_id}/notifications/read-all
```

**请求体:**

```json
{
    "agent_id": "agent-phone"
}
```

**响应:**

```json
{
    "success": true,
    "count": 5
}
```

#### 获取通知统计

```
GET /api/v3/identities/{identity_id}/notifications/stats
```

**响应:**

```json
{
    "stats": {
        "total": 10,
        "pending": 2,
        "delivered": 3,
        "read": 4,
        "dismissed": 1,
        "unread": 5
    }
}
```

---

### Android SDK v3.x API

#### 消息总线客户端

```java
// 获取消息总线
MessageBus messageBus = agent.getMessageBus();

// 发送消息
Message msg = new Message();
msg.fromAgent = agentId;
msg.toAgent = "agent-watch";
msg.type = Message.TYPE_DATA;
msg.priority = Message.PRIORITY_NORMAL;
msg.payload = Map.of("key", "value");

messageBus.send(msg);

// 添加消息监听器
messageBus.addListener(message -> {
    if (message.type == Message.TYPE_DATA) {
        // 处理数据消息
    }
});

// 获取离线消息
List<Message> offline = messageBus.getOfflineMessages();
```

#### 设备状态同步

```java
// 获取状态同步服务
StateSyncService stateSync = agent.getStateSyncService();

// 上报状态
stateSync.reportBattery(85, false);
stateSync.reportNetwork("wifi", 90);
stateSync.reportScene("meeting");

// 添加状态变更监听器
stateSync.addStateChangeListener(change -> {
    if (change.field == "battery") {
        // 电池状态变更
    }
});

// 获取当前状态
DeviceState state = stateSync.getCurrentState();
```

#### 场景感知路由

```java
// 获取场景路由器
SceneAwareRouter router = agent.getSceneAwareRouter();

// 本地路由决策
RoutingResult result = router.routeLocally(
    "social",
    Message.PRIORITY_NORMAL
);

if (result.targetAgents.contains("watch")) {
    // 路由到手表
}

// 添加路由规则
router.addRule(new RoutingRule.Builder()
    .scenes(Set.of("running"))
    .messageTypes(Set.of("social"))
    .targetDeviceTypes(Set.of("watch"))
    .action(RoutingAction.DELIVER)
    .build());
```

#### 任务协同

```java
// 获取任务协同器
TaskCollaborator collaborator = agent.getTaskCollaborator();

// 注册为执行者
collaborator.registerExecutor("data.collect", subTask -> {
    // 执行子任务
    Object result = collectData();
    // 上报结果
    collaborator.reportResult(subTask.subTaskId, true, result);
});

// 创建协同任务
CollaborativeTask task = new CollaborativeTask.Builder()
    .skillId("data.collect")
    .splitStrategy(SplitStrategy.PARALLEL)
    .mergeStrategy(MergeStrategy.AGGREGATE)
    .targetAgents(Set.of("phone", "watch", "tablet"))
    .build();

// 发起协同
collaborator.initiateCollaboration(task);
```

#### 通知客户端

```java
// 获取通知客户端
NotificationClient notificationClient = agent.getNotificationClient();

// 设置本地通知处理器
notificationClient.setLocalHandler(new LocalNotificationHandler() {
    @Override
    public void showNotification(CrossDeviceNotification notification) {
        // 显示本地通知
        showSystemNotification(notification);
    }

    @Override
    public void cancelNotification(String notificationId) {
        // 取消本地通知
        cancelSystemNotification(notificationId);
    }
});

// 添加通知监听器
notificationClient.addListener(new NotificationListener() {
    @Override
    public void onNotificationReceived(CrossDeviceNotification notification) {
        // 收到新通知
    }

    @Override
    public void onNotificationUpdated(CrossDeviceNotification notification) {
        // 通知状态更新
    }
});

// 标记已读
notificationClient.markAsRead("notif-001");

// 获取未读数
int unreadCount = notificationClient.getUnreadCount();

// 执行通知动作
notificationClient.executeAction("notif-001", "open");
```

---

### v3.x 新增枚举类型

#### MessageType

| 值 | 说明 |
|---|------|
| data | 数据消息 |
| command | 命令消息 |
| event | 事件消息 |
| notification | 通知消息 |

#### MessagePriority

| 值 | 说明 |
|---|------|
| 0 | LOW |
| 1 | NORMAL |
| 2 | HIGH |
| 3 | URGENT |

#### NotificationType

| 值 | 说明 |
|---|------|
| message | 普通消息 |
| alert | 告警 |
| reminder | 提醒 |
| system | 系统通知 |
| social | 社交消息 |
| health | 健康提醒 |
| calendar | 日历提醒 |
| call | 通话通知 |

#### NotificationPriority

| 值 | 说明 |
|---|------|
| min | 最低优先级 |
| low | 低优先级 |
| normal | 正常优先级 |
| high | 高优先级 |
| max | 最高优先级 (勿扰时段仍显示) |

#### NotificationStatus

| 值 | 说明 |
|---|------|
| pending | 待发送 |
| delivered | 已送达 |
| read | 已读 |
| dismissed | 已忽略 |
| expired | 已过期 |

#### SplitStrategy

| 值 | 说明 |
|---|------|
| none | 不拆分 |
| parallel | 并行拆分 |
| sequence | 顺序拆分 |
| map_reduce | MapReduce |
| by_device | 按设备拆分 |

#### MergeStrategy

| 值 | 说明 |
|---|------|
| none | 不合并 |
| all | 收集所有结果 |
| first | 取首个完成结果 |
| consensus | 共识合并 |
| aggregate | 聚合统计 |
| best | 取最佳结果 |

#### DeviceScene

| 值 | 说明 |
|---|------|
| unknown | 未知 |
| idle | 空闲 |
| running | 跑步 |
| walking | 步行 |
| driving | 驾驶 |
| meeting | 会议 |
| sleeping | 睡眠 |
| working | 工作 |
| relaxing | 休息 |

#### RoutingAction

| 值 | 说明 |
|---|------|
| deliver | 直接送达 |
| delay | 延迟送达 |
| filter | 过滤不送 |
| broadcast | 广播所有 |
| quiet | 勿扰模式 |

---

## v4.x 灵魂特征系统 API (新增)

### 概述

v4.x 系统实现"灵魂特征"，让数字灵魂更加真实、立体。每个子系统遵循统一架构：

- **Center端**: 模型文件 + 引擎文件（深度管理）
- **Agent端**: 状态文件 + 客户端文件（轻量接收）

---

### 情绪系统 API (v4.0.0)

#### 获取情绪状态

```
GET /api/v4/emotion/{identity_id}
```

**响应:**

```json
{
    "emotions": {
        "joy": 0.6,
        "anger": 0.1,
        "sadness": 0.2,
        "fear": 0.1,
        "love": 0.5,
        "disgust": 0.0,
        "desire": 0.4
    },
    "dominant_emotion": "joy",
    "emotion_stability": 0.75,
    "last_trigger": "positive_interaction"
}
```

#### 触发情绪

```
POST /api/v4/emotion/{identity_id}/trigger
```

**请求体:**

```json
{
    "trigger_type": "event",
    "event_data": {
        "type": "positive_interaction",
        "intensity": 0.7,
        "source": "social"
    }
}
```

**响应:**

```json
{
    "success": true,
    "emotion_change": {
        "joy": "+0.3",
        "love": "+0.2"
    }
}
```

#### 获取欲望状态

```
GET /api/v4/desire/{identity_id}
```

**响应:**

```json
{
    "desires": {
        "physiological": {"level": 0.8, "satisfied": 0.9},
        "safety": {"level": 0.7, "satisfied": 0.85},
        "love_belonging": {"level": 0.6, "satisfied": 0.5},
        "esteem": {"level": 0.5, "satisfied": 0.6},
        "self_actualization": {"level": 0.4, "satisfied": 0.3}
    },
    "dominant_desire": "love_belonging",
    "priority_queue": ["love_belonging", "esteem", "physiological"]
}
```

#### Android SDK 情绪 API

```java
// 获取情绪客户端
EmotionClient emotionClient = EmotionClient.getInstance();

// 获取当前情绪
EmotionState state = emotionClient.getCurrentState();
double joy = state.getEmotion("joy");  // 获取喜悦值
String dominant = state.getDominantEmotion();  // 获取主导情绪

// 判断情绪状态
if (emotionClient.isHappy()) {
    // 喜悦主导
}
if (emotionClient.isStressed()) {
    // 压力状态
}

// 获取欲望状态
DesireState desireState = emotionClient.getDesireState();
String unsatisfied = desireState.getMostUnsatisfiedDesire();  // 最未满足的欲望

// 添加状态监听器
emotionClient.addListener(state -> {
    // 情绪状态变化
});
```

---

### 三观系统 API (v4.1.0)

#### 获取世界观

```
GET /api/v4/philosophy/{identity_id}/worldview
```

**响应:**

```json
{
    "world_essence": "complex_adaptive",
    "social_cognition": "cooperative_competitive",
    "future_view": "optimistic_progressive",
    "human_relationship_view": "interdependent",
    "certainty_score": 0.7
}
```

#### 获取人生观

```
GET /api/v4/philosophy/{identity_id}/lifeview
```

**响应:**

```json
{
    "life_meaning": "growth_contribution",
    "time_orientation": "balanced_present_future",
    "life_attitude": "active_exploring",
    "success_definition": "self_realization_balance",
    "happiness_source": "relationships_growth"
}
```

#### 获取价值观

```
GET /api/v4/philosophy/{identity_id}/values
```

**响应:**

```json
{
    "core_values": ["honesty", "growth", "family", "freedom"],
    "value_ranking": {
        "honesty": 1,
        "growth": 2,
        "family": 3,
        "freedom": 4
    },
    "moral_framework": "utilitarian_care",
    "decision_principles": ["minimize_harm", "promote_growth"]
}
```

#### 获取三观决策上下文

```
GET /api/v4/philosophy/{identity_id}/decision-context
```

**响应:**

```json
{
    "value_alignment": 0.85,
    "moral_judgment": "acceptable",
    "life_goal_relevance": 0.7,
    "worldview_compatibility": 0.8,
    "recommendation": "align_with_values",
    "reasoning": "此行动符合诚实和成长的核心价值观"
}
```

#### Android SDK 三观 API

```java
// 获取三观客户端
PhilosophyClient philosophyClient = PhilosophyClient.getInstance();

// 获取世界观
PhilosophyState.Worldview worldview = philosophyClient.getWorldview();
String essence = worldview.worldEssence;  // 世界本质

// 获取价值观
PhilosophyState.ValueSystem values = philosophyClient.getValueSystem();
List<String> coreValues = values.coreValues;  // 核心价值观

// 判断价值观对齐
if (philosophyClient.alignsWithValue("honesty")) {
    // 与诚实价值观对齐
}

// 获取决策建议
String recommendation = philosophyClient.getDecisionRecommendation(context);
```

---

### 社会身份 API (v4.2.0)

#### 获取教育背景

```
GET /api/v4/social/{identity_id}/education
```

**响应:**

```json
{
    "highest_level": "master",
    "major_field": "computer_science",
    "school_tier": "top",
    "academic_performance": "excellent",
    "continuous_learning": true,
    "learning_fields": ["ai", "psychology"]
}
```

#### 获取职业画像

```
GET /api/v4/social/{identity_id}/career
```

**响应:**

```json
{
    "occupation": "software_engineer",
    "industry": "technology",
    "career_stage": "mid_level",
    "job_satisfaction": 0.7,
    "work_mode": "hybrid",
    "income_level": "middle_high"
}
```

#### 获取社会阶层

```
GET /api/v4/social/{identity_id}/social-class
```

**响应:**

```json
{
    "economic_capital": {"income": "middle_high", "wealth": "middle"},
    "cultural_capital": {"education": "high", "taste": "intellectual"},
    "social_capital": {"connections": "moderate", "network_quality": "good"},
    "class_position": "middle_class",
    "mobility_trajectory": "upward"
}
```

#### 获取身份认同

```
GET /api/v4/social/{identity_id}/identity-profile
```

**响应:**

```json
{
    "primary_roles": ["professional", "family_member"],
    "role_priorities": {"professional": 1, "family_member": 2},
    "identity_confidence": 0.8,
    "role_conflicts": []
}
```

#### Android SDK 社会身份 API

```java
// 获取社会身份客户端
SocialIdentityClient socialClient = SocialIdentityClient.getInstance();

// 获取教育背景
SocialIdentityState.EducationBackground edu = socialClient.getEducationBackground();
String level = edu.highestLevel;  // 最高学历

// 获取职业画像
SocialIdentityState.CareerProfile career = socialClient.getCareerProfile();
String occupation = career.occupation;  // 职业

// 获取社会阶层
SocialIdentityState.SocialClassProfile socialClass = socialClient.getSocialClassProfile();
String position = socialClass.classPosition;  // 阶层位置

// 判断身份
if (socialClient.isProfessional()) {
    // 职业身份主导
}
```

---

### 地域文化 API (v4.3.0)

#### 获取地域文化

```
GET /api/v4/culture/{identity_id}
```

**响应:**

```json
{
    "birthplace": {"province": "广东", "city": "深圳"},
    "current_location": {"province": "北京", "city": "北京"},
    "city_level": "first_tier",
    "region": "north",
    "hofstede_dimensions": {
        "collectivism": 0.6,
        "power_distance": 0.5,
        "uncertainty_avoidance": 0.4,
        "long_term_orientation": 0.7
    },
    "communication_style": "context_dependent",
    "social_style": "reserved",
    "migration_history": [
        {"from": "广东", "to": "北京", "year": 2015}
    ]
}
```

#### 获取文化决策上下文

```
GET /api/v4/culture/{identity_id}/decision-context
```

**响应:**

```json
{
    "communication_adjustment": "formal",
    "social_expectation": "group_harmony",
    "conflict_handling": "indirect_negotiation",
    "decision_influence": {
        "collectivism_weight": 0.6,
        "authority_weight": 0.5
    }
}
```

#### Android SDK 地域文化 API

```java
// 获取地域文化客户端
RegionalCultureClient cultureClient = RegionalCultureClient.getInstance();

// 获取当前地域
RegionalCultureState.CultureLocation location = cultureClient.getCurrentLocation();
String province = location.province;

// 获取文化维度
RegionalCultureState.HofstedeDimensions dims = cultureClient.getHofstedeDimensions();
double collectivism = dims.collectivism;

// 判断沟通风格
if (cultureClient.isDirectCommunicator()) {
    // 直接沟通风格
}

// 获取文化建议
String suggestion = cultureClient.getCommunicationSuggestion("business");
```

---

### 人生阶段 API (v4.4.0)

#### 获取人生阶段

```
GET /api/v4/lifestage/{identity_id}
```

**响应:**

```json
{
    "current_stage": "early_adult",
    "stage_progress": 0.3,
    "stage_characteristics": {
        "exploration": "high",
        "stability_seeking": "medium",
        "identity_building": "high"
    },
    "development_metrics": {
        "physical_mental": {"score": 0.85, "trend": "stable"},
        "social": {"score": 0.7, "trend": "improving"},
        "career": {"score": 0.6, "trend": "improving"},
        "self_realization": {"score": 0.5, "trend": "exploring"}
    }
}
```

#### 获取人生事件

```
GET /api/v4/lifestage/{identity_id}/events
```

**响应:**

```json
{
    "events": [
        {
            "event_id": "evt-001",
            "type": "education",
            "title": "大学毕业",
            "impact": 0.8,
            "timestamp": 1711622400
        },
        {
            "event_id": "evt-002",
            "type": "career",
            "title": "首次就业",
            "impact": 0.7,
            "timestamp": 1711622500
        }
    ],
    "total": 2
}
```

#### 添加人生事件

```
POST /api/v4/lifestage/{identity_id}/events
```

**请求体:**

```json
{
    "type": "career",
    "title": "晋升",
    "description": "晋升为高级工程师",
    "impact": 0.6,
    "lessons": ["持续学习的重要性", "团队合作的价值"]
}
```

#### Android SDK 人生阶段 API

```java
// 获取人生阶段客户端
LifeStageClient lifeStageClient = LifeStageClient.getInstance();

// 获取当前阶段
LifeStageState stage = lifeStageClient.getCurrentState();
String currentStage = stage.currentStage;  // 当前阶段

// 获取发展指标
LifeStageState.DevelopmentMetrics metrics = lifeStageClient.getDevelopmentMetrics();
double socialScore = metrics.social.score;  // 社会发展得分

// 判断阶段特征
if (lifeStageClient.isInExplorationPhase()) {
    // 处于探索阶段
}

// 获取阶段建议
String advice = lifeStageClient.getStageAdvice();
```

---

### 情绪行为联动 API (v4.5.0)

#### 获取情绪决策影响

```
GET /api/v4/behavior/{identity_id}/decision-influence
```

**响应:**

```json
{
    "risk_preference": "moderate",
    "impulse_control": 0.6,
    "social_tendency": "approach",
    "decision_style": "analytical_emotional",
    "risk_tolerance": 0.5,
    "delay_discounting": 0.3
}
```

#### 获取情绪表达影响

```
GET /api/v4/behavior/{identity_id}/expression-influence
```

**响应:**

```json
{
    "tone_style": "warm_friendly",
    "wording_preference": "positive_constructive",
    "emoji_usage": "moderate",
    "response_style": "thoughtful_responsive",
    "emotional_expressiveness": 0.7
}
```

#### 获取行为指导

```
GET /api/v4/behavior/{identity_id}/guidance
```

**响应:**

```json
{
    "decision_suggestion": "考虑情绪因素，选择保守方案",
    "expression_suggestion": "温和表达，避免冲动言辞",
    "risk_warning": "当前情绪可能导致冲动决策",
    "regulation_suggestion": "先冷静5分钟再做决定",
    "recommended_coping": "problem_focused"
}
```

#### Android SDK 情绪行为 API

```java
// 获取情绪行为客户端
EmotionBehaviorClient behaviorClient = EmotionBehaviorClient.getInstance();

// 获取决策影响
EmotionBehaviorState.DecisionInfluence decision = behaviorClient.getDecisionInfluence();
String riskPref = decision.riskPreference;  // 风险偏好

// 获取表达影响
EmotionBehaviorState.ExpressionInfluence expression = behaviorClient.getExpressionInfluence();
String toneStyle = expression.toneStyle;  // 语调风格

// 获取行为建议
EmotionBehaviorState.BehaviorGuidance guidance = behaviorClient.getBehaviorGuidance();
String suggestion = guidance.decisionSuggestion;

// 判断情绪风险
if (behaviorClient.hasImpulseRisk()) {
    // 存在冲动风险
}
```

---

### 人际关系 API (v4.6.0)

#### 获取社交网络

```
GET /api/v4/relationship/{identity_id}/network
```

**响应:**

```json
{
    "total_contacts": 50,
    "close_contacts": 8,
    "support_contacts": 5,
    "weak_ties": 35,
    "network_density": 0.3,
    "network_diversity": 0.6,
    "social_capital": 0.7,
    "network_health": 0.8,
    "has_support": true
}
```

#### 获取依恋风格

```
GET /api/v4/relationship/{identity_id}/attachment
```

**响应:**

```json
{
    "primary_style": "secure",
    "anxiety_level": 0.2,
    "avoidance_level": 0.3,
    "style_description": "安全型依恋，信任他人，适度依赖"
}
```

#### 获取社交风格

```
GET /api/v4/relationship/{identity_id}/social-style
```

**响应:**

```json
{
    "directness": 0.6,
    "expressiveness": 0.5,
    "listening_style": "active",
    "social_energy": 0.7,
    "small_talk_comfort": 0.5,
    "deep_talk_preference": 0.7,
    "group_vs_one_on_one": "one_on_one"
}
```

#### 获取社交建议

```
GET /api/v4/relationship/{identity_id}/guidance
```

**响应:**

```json
{
    "should_socialize": true,
    "social_energy_level": 0.7,
    "preferred_group_size": "one_on_one",
    "preferred_depth": "deep",
    "conflict_risk": 0.2,
    "needs_self_care": false,
    "maintenance_needed": ["friend_a", "friend_b"],
    "tension_relationships": [],
    "boundary_reminders": ["注意边界"],
    "growth_opportunities": ["尝试新社交圈"]
}
```

#### Android SDK 人际关系 API

```java
// 获取人际关系客户端
RelationshipClient relationshipClient = RelationshipClient.getInstance();

// 获取网络摘要
RelationshipState.NetworkSummary network = relationshipClient.getNetworkSummary();
int closeContacts = network.closeContacts;  // 亲密联系人数

// 获取依恋风格
RelationshipState.AttachmentStyle attachment = relationshipClient.getAttachmentStyle();
String style = attachment.primaryStyle;  // 主要依恋风格

// 判断依恋类型
if (relationshipClient.isSecurelyAttached()) {
    // 安全型依恋
}

// 获取社交建议
RelationshipState.SocialGuidance guidance = relationshipClient.getSocialGuidance();
boolean shouldSocialize = guidance.shouldSocialize;

// 获取社交推荐
String approach = relationshipClient.getRecommendedSocialApproach();
String advice = relationshipClient.getRelationshipAdvice();
```

---

### v4.x 新增枚举类型

#### EmotionType

| 值 | 说明 |
|---|------|
| joy | 喜悦 |
| anger | 愤怒 |
| sadness | 悲哀 |
| fear | 恐惧 |
| love | 喜爱 |
| disgust | 厌恶 |
| desire | 欲望 |

#### DesireType (马斯洛需求)

| 值 | 说明 |
|---|------|
| physiological | 生理需求 |
| safety | 安全需求 |
| love_belonging | 归属与爱 |
| esteem | 尊重需求 |
| self_actualization | 自我实现 |

#### LifeStageType

| 值 | 说明 |
|---|------|
| childhood | 童年 |
| adolescence | 青春期 |
| youth | 青年 |
| early_adult | 成年早期 |
| mid_adult | 中年 |
| mature | 成熟期 |
| elderly | 老年 |

#### AttachmentStyleType

| 值 | 说明 |
|---|------|
| secure | 安全型 |
| anxious | 焦虑型 |
| avoidant | 回避型 |
| disorganized | 混乱型 |

#### CopingStrategyType

| 值 | 说明 |
|---|------|
| problem_focused | 问题聚焦 |
| emotion_focused | 情绪聚焦 |
| avoidance | 回避 |
| support_seeking | 寻求支持 |
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
    "version": "v0.1.0"
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
    "version": "v0.1.0",
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
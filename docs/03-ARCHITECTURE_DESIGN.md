# OFA 架构设计文档

---

# 一、系统架构概览

## 1.1 整体架构

```
┌────────────────────────────────────────────────────────────────────────────┐
│                              用户层 (User Layer)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │  手机App  │  │  平板App  │  │  桌面App  │  │  Web端   │  │  手表App  │     │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘     │
└───────┼─────────────┼─────────────┼─────────────┼─────────────┼───────────┘
        │             │             │             │             │
        └─────────────┴─────────────┴──────┬──────┴─────────────┘
                                           │
┌──────────────────────────────────────────┼──────────────────────────────────┐
│                              Agent层 (Agent Layer)                         │
│  ┌───────────────────────────────────────┴───────────────────────────────┐ │
│  │                           Agent Runtime                                │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │ │
│  │  │  Core模块   │  │  通信模块   │  │  执行模块   │  │  能力模块   │   │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────┬──────────────────────────────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                           Center层 (Center Layer)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │  API网关    │  │  调度引擎   │  │  消息路由   │  │  存储服务   │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
│         │                │                │                │             │
│  ┌──────┴────────────────┴────────────────┴────────────────┴──────┐      │
│  │                        服务总线 (Service Bus)                    │      │
│  └─────────────────────────────────────────────────────────────────┘      │
└───────────────────────────────────────────────────────────────────────────┘
                                           │
┌──────────────────────────────────────────┼──────────────────────────────────┐
│                           基础设施层 (Infrastructure)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │
│  │ PostgreSQL  │  │    Redis    │  │    etcd     │  │    NATS     │       │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘
```

## 1.2 架构分层

| 层级 | 职责 | 组件 |
|------|------|------|
| **用户层** | 用户交互入口 | 各平台客户端App |
| **Agent层** | 设备端智能体 | Agent Runtime、Skills、Tools |
| **Center层** | 管理调度中心 | API网关、调度引擎、消息路由、存储服务 |
| **基础设施层** | 数据与消息支撑 | 数据库、缓存、配置中心、消息队列 |

---

# 二、核心组件设计

## 2.1 Center组件

### 2.1.1 API网关 (API Gateway)

**职责：**
- 统一入口，路由请求
- 认证授权
- 限流熔断
- 协议转换

**接口分类：**
```
/api/v1/
├── agent/           # Agent管理
│   ├── register     # 注册
│   ├── heartbeat    # 心跳
│   ├── unregister   # 注销
│   └── status       # 状态查询
├── task/            # 任务管理
│   ├── submit       # 提交任务
│   ├── status       # 任务状态
│   ├── cancel       # 取消任务
│   └── result       # 获取结果
├── message/         # 消息服务
│   ├── send         # 发送消息
│   ├── broadcast    # 广播消息
│   └── subscribe    # 订阅消息
├── skill/           # 技能管理
│   ├── list         # 技能列表
│   ├── install      # 安装技能
│   └── uninstall    # 卸载技能
└── system/          # 系统管理
    ├── config       # 配置管理
    ├── monitor      # 监控信息
    └── logs         # 日志查询
```

### 2.1.2 调度引擎 (Scheduler)

**职责：**
- 任务分解
- Agent选择
- 任务分发
- 结果聚合

**调度流程：**
```
任务提交 → 任务解析 → 能力匹配 → Agent选择 → 任务分发 → 执行监控 → 结果收集
    │                              │
    │    ┌─────────────────────────┘
    │    │
    │    ▼
    │  调度策略:
    │  - 能力优先
    │  - 负载优先
    │  - 就近优先
    │  - 功耗优先
    │
    ▼
调度决策 → Agent队列
```

**调度器结构：**
```go
type Scheduler struct {
    taskQueue    *PriorityQueue
    agentPool    *AgentPool
    policy       SchedulePolicy
    dispatcher   *Dispatcher
    monitor      *TaskMonitor
}

type SchedulePolicy interface {
    Select(task *Task, agents []*Agent) *Agent
}

// 策略实现
type CapabilityFirstPolicy struct{}  // 能力优先
type LoadBalancePolicy struct{}       // 负载均衡
type LatencyFirstPolicy struct{}      // 延迟优先
type PowerAwarePolicy struct{}        // 功耗感知
type HybridPolicy struct{}            // 混合策略
```

### 2.1.3 消息路由 (Message Router)

**职责：**
- 消息路由转发
- Agent间通信
- 消息持久化
- 离线消息

**路由表结构：**
```go
type RouteTable struct {
    // AgentID -> Connection
    agentConnections sync.Map

    // GroupID -> []AgentID
    groupMembers sync.Map

    // 订阅关系
    subscriptions *SubscriptionManager
}
```

**消息流转：**
```
┌─────────┐    ┌──────────────┐    ┌─────────┐
│ Agent A │───▶│ MessageRouter│───▶│ Agent B │
└─────────┘    └──────────────┘    └─────────┘
                     │
                     │ 离线?
                     ▼
              ┌──────────────┐
              │ OfflineQueue │
              └──────────────┘
```

### 2.1.4 存储服务 (Storage)

**数据模型：**

```sql
-- Agent注册表
CREATE TABLE agents (
    id              VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128),
    type            VARCHAR(32),      -- full/mobile/lite/iot/edge
    status          VARCHAR(16),      -- online/offline/busy
    capabilities    JSONB,            -- 能力列表
    metadata        JSONB,            -- 元数据
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP
);

-- 任务记录
CREATE TABLE tasks (
    id              VARCHAR(64) PRIMARY KEY,
    parent_id       VARCHAR(64),
    source_agent    VARCHAR(64),
    target_agent    VARCHAR(64),
    task_type       VARCHAR(32),
    priority        INTEGER,
    status          VARCHAR(16),
    input           JSONB,
    output          JSONB,
    error           TEXT,
    created_at      TIMESTAMP,
    started_at      TIMESTAMP,
    completed_at    TIMESTAMP
);

-- 消息记录
CREATE TABLE messages (
    id              VARCHAR(64) PRIMARY KEY,
    msg_type        VARCHAR(16),
    from_agent      VARCHAR(64),
    to_agent        VARCHAR(64),
    action          VARCHAR(64),
    payload         JSONB,
    status          VARCHAR(16),
    created_at      TIMESTAMP,
    delivered_at    TIMESTAMP
);

-- Skill注册表
CREATE TABLE skills (
    id              VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128),
    version         VARCHAR(16),
    category        VARCHAR(32),
    description     TEXT,
    schema          JSONB,
    created_at      TIMESTAMP
);

-- Agent-Skill关联
CREATE TABLE agent_skills (
    agent_id        VARCHAR(64),
    skill_id        VARCHAR(64),
    installed_at    TIMESTAMP,
    PRIMARY KEY (agent_id, skill_id)
);
```

## 2.2 Agent组件

### 2.2.1 Core模块

**职责：**
- Agent生命周期管理
- 配置管理
- 日志管理
- 健康检查

```go
type AgentCore struct {
    config      *AgentConfig
    identity    *AgentIdentity
    state       *AgentState
    connector   *Connector
    executor    *Executor
    skillMgr    *SkillManager
    toolMgr     *ToolManager
}

type AgentIdentity struct {
    ID           string
    Name         string
    Type         AgentType
    DeviceInfo   DeviceInfo
}

type AgentState struct {
    Status       AgentStatus
    CPUUsage     float64
    MemoryUsage  float64
    BatteryLevel int
    NetworkType  string
    NetworkLatency time.Duration
}
```

### 2.2.2 通信模块

**职责：**
- 与Center建立连接
- 消息收发
- 心跳维持
- 断线重连

```go
type Connector struct {
    centerAddr    string
    conn          *grpc.ClientConn
    stream        proto.AgentService_ConnectClient

    // 消息处理
    sendChan      chan *proto.Message
    recvChan      chan *proto.Message

    // 心跳
    heartbeatTicker *time.Ticker
    lastHeartbeat   time.Time
}

func (c *Connector) Connect() error
func (c *Connector) Send(msg *proto.Message) error
func (c *Connector) Receive() (*proto.Message, error)
func (c *Connector) Heartbeat() error
func (c *Connector) Reconnect() error
```

### 2.2.3 执行模块

**职责：**
- 任务接收
- Skill/Tool调度
- 执行监控
- 结果返回

```go
type Executor struct {
    taskQueue    chan *Task
    workers      []*Worker
    skillMgr     *SkillManager
    resultChan   chan *TaskResult
}

type Worker struct {
    id          int
    taskChan    chan *Task
    resultChan  chan *TaskResult
}

func (w *Worker) Execute(task *Task) *TaskResult {
    // 1. 解析任务
    // 2. 匹配Skill
    // 3. 调用Tool
    // 4. 返回结果
}
```

### 2.2.4 能力模块

**Skill管理器：**
```go
type SkillManager struct {
    skills      map[string]*Skill
    loader      *SkillLoader
    validator   *SkillValidator
}

type Skill struct {
    ID          string
    Name        string
    Version     string
    Category    string
    Inputs      []ParamSpec
    Outputs     []ParamSpec
    Requires    []string
    Handler     SkillHandler
}

type SkillHandler interface {
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}
```

**Tool管理器：**
```go
type ToolManager struct {
    tools       map[string]*Tool
    loader      *ToolLoader
    sandbox     *Sandbox
}

type Tool struct {
    ID          string
    Name        string
    Type        ToolType    // native/script/api/model
    Executable  string
    Config      map[string]interface{}
}

type ToolType int
const (
    ToolNative ToolType = iota
    ToolScript
    ToolAPI
    ToolModel
)
```

---

# 三、通信协议设计

## 3.1 gRPC服务定义

### 3.1.1 Agent服务

```protobuf
syntax = "proto3";

package ofa;

option go_package = "github.com/ofa/proto";

// Agent服务 - Center与Agent通信
service AgentService {
    // 双向流式通信
    rpc Connect(stream AgentMessage) returns (stream CenterMessage);

    // 任务相关
    rpc SubmitTask(TaskRequest) returns (TaskResponse);
    rpc GetTaskStatus(TaskStatusRequest) returns (TaskStatusResponse);
    rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);

    // 能力相关
    rpc RegisterCapabilities(CapabilitiesRequest) returns (CapabilitiesResponse);
    rpc GetCapabilities(GetCapabilitiesRequest) returns (GetCapabilitiesResponse);
}

// Agent消息
message AgentMessage {
    string msg_id = 1;
    oneof payload {
        RegisterRequest register = 10;
        HeartbeatRequest heartbeat = 11;
        TaskResult task_result = 12;
        AgentEvent event = 13;
    }
}

// Center消息
message CenterMessage {
    string msg_id = 1;
    oneof payload {
        RegisterResponse register = 10;
        TaskAssignment task = 11;
        ConfigUpdate config = 12;
        BroadcastMessage broadcast = 13;
    }
}

// 注册请求
message RegisterRequest {
    string agent_id = 1;
    string name = 2;
    AgentType type = 3;
    DeviceInfo device_info = 4;
    repeated Capability capabilities = 5;
}

enum AgentType {
    AGENT_TYPE_UNKNOWN = 0;
    AGENT_TYPE_FULL = 1;
    AGENT_TYPE_MOBILE = 2;
    AGENT_TYPE_LITE = 3;
    AGENT_TYPE_IOT = 4;
    AGENT_TYPE_EDGE = 5;
}

message DeviceInfo {
    string os = 1;
    string os_version = 2;
    string model = 3;
    string manufacturer = 4;
    int64 total_memory = 5;
    int32 cpu_cores = 6;
    string arch = 7;
}

message Capability {
    string id = 1;
    string name = 2;
    string version = 3;
    string category = 4;
    map<string, string> metadata = 5;
}

// 心跳
message HeartbeatRequest {
    string agent_id = 1;
    AgentStatus status = 2;
    ResourceUsage resources = 3;
}

enum AgentStatus {
    AGENT_STATUS_UNKNOWN = 0;
    AGENT_STATUS_ONLINE = 1;
    AGENT_STATUS_BUSY = 2;
    AGENT_STATUS_IDLE = 3;
    AGENT_STATUS_OFFLINE = 4;
}

message ResourceUsage {
    double cpu_usage = 1;
    double memory_usage = 2;
    int32 battery_level = 3;
    string network_type = 4;
    int32 network_latency_ms = 5;
}

// 任务
message TaskAssignment {
    string task_id = 1;
    string parent_task_id = 2;
    string skill_id = 3;
    map<string, bytes> input = 4;
    int32 priority = 5;
    int64 timeout_ms = 6;
    map<string, string> metadata = 7;
}

message TaskResult {
    string task_id = 1;
    TaskStatus status = 2;
    map<string, bytes> output = 3;
    string error = 4;
    int64 duration_ms = 5;
}

enum TaskStatus {
    TASK_STATUS_UNKNOWN = 0;
    TASK_STATUS_PENDING = 1;
    TASK_STATUS_RUNNING = 2;
    TASK_STATUS_COMPLETED = 3;
    TASK_STATUS_FAILED = 4;
    TASK_STATUS_CANCELLED = 5;
    TASK_STATUS_TIMEOUT = 6;
}
```

### 3.1.2 消息服务

```protobuf
// 消息服务 - Agent间通信
service MessageService {
    // 点对点消息
    rpc SendMessage(MessageRequest) returns (MessageResponse);

    // 广播消息
    rpc Broadcast(BroadcastRequest) returns (BroadcastResponse);

    // 组播消息
    rpc Multicast(MulticastRequest) returns (MulticastResponse);

    // 订阅消息
    rpc Subscribe(SubscribeRequest) returns (stream Message);
}

message Message {
    string msg_id = 1;
    string from_agent = 2;
    string to_agent = 3;
    string action = 4;
    bytes payload = 5;
    int64 timestamp = 6;
    int32 ttl = 7;
}

message MessageRequest {
    Message message = 1;
    bool require_ack = 2;
    int64 timeout_ms = 3;
}

message MessageResponse {
    string msg_id = 1;
    bool success = 2;
    string error = 3;
}
```

## 3.2 通信流程

### 3.2.1 Agent注册流程

```
Agent                    Center
  │                        │
  │──── RegisterRequest ──▶│
  │                        │── 验证身份
  │                        │── 分配ID
  │                        │── 存储信息
  │◀── RegisterResponse ───│
  │                        │
  │───── Heartbeat ───────▶│ (周期性)
  │◀──── Heartbeat ACK ────│
  │                        │
```

### 3.2.2 任务执行流程

```
User/Agent              Center                  TargetAgent
    │                     │                        │
    │──── SubmitTask ────▶│                        │
    │                     │── 解析任务              │
    │                     │── 匹配能力              │
    │                     │── 选择Agent             │
    │                     │──── TaskAssignment ───▶│
    │                     │                        │── 执行任务
    │                     │◀─── TaskResult ────────│
    │◀─── TaskResult ─────│                        │
    │                     │                        │
```

### 3.2.3 Agent间通信流程

```
AgentA                  Center                  AgentB
  │                       │                       │
  │─── SendMessage ──────▶│                       │
  │                       │── 查找AgentB          │
  │                       │─── Forward Message ──▶│
  │                       │◀── ACK ───────────────│
  │◀── ACK ───────────────│                       │
  │                       │                       │
```

---

# 四、部署架构

## 4.1 单机部署

```
┌─────────────────────────────────────────────────────────┐
│                      单机部署                            │
│  ┌─────────────────────────────────────────────────────┐│
│  │                    Center服务                        ││
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   ││
│  │  │ API网关 │ │ 调度器  │ │ 路由器  │ │ 存储服务│   ││
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘   ││
│  │       └────────────┴────────────┴────────────┘      ││
│  └─────────────────────────────────────────────────────┘│
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       ││
│  │ PostgreSQL  │ │    Redis    │ │    etcd     │       ││
│  └─────────────┘ └─────────────┘ └─────────────┘       ││
└─────────────────────────────────────────────────────────┘
```

## 4.2 集群部署

```
┌─────────────────────────────────────────────────────────────────────┐
│                           负载均衡器                                 │
│                         (Nginx/HAProxy)                             │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   Center 1    │   │   Center 2    │   │   Center 3    │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  PostgreSQL   │   │    Redis      │   │     etcd      │
│   (主从)      │   │   (集群)      │   │   (集群)      │
└───────────────┘   └───────────────┘   └───────────────┘
```

## 4.3 Kubernetes部署

```yaml
# center-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ofa-center
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ofa-center
  template:
    metadata:
      labels:
        app: ofa-center
    spec:
      containers:
      - name: center
        image: ofa/center:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: ofa-secrets
              key: db-host
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
---
apiVersion: v1
kind: Service
metadata:
  name: ofa-center
spec:
  selector:
    app: ofa-center
  ports:
  - port: 8080
    name: http
  - port: 9090
    name: grpc
  type: LoadBalancer
```

---

# 五、安全架构

## 5.1 认证流程

```
Agent                    Center
  │                        │
  │─── 1. Register ───────▶│
  │    (Device ID, Secret) │
  │                        │── 验证设备
  │                        │── 生成Token
  │◀── 2. JWT Token ───────│
  │                        │
  │─── 3. API Request ────▶│
  │    (Authorization: Bearer Token)
  │                        │── 验证Token
  │                        │── 检查权限
  │◀── 4. Response ────────│
```

## 5.2 Token结构

```json
{
  "header": {
    "alg": "EdDSA",
    "typ": "JWT"
  },
  "payload": {
    "iss": "ofa-center",
    "sub": "agent-uuid",
    "aud": "ofa-agent",
    "exp": 1234567890,
    "iat": 1234560000,
    "type": "mobile",
    "capabilities": ["skill-1", "skill-2"],
    "permissions": ["task:execute", "message:send"]
  },
  "signature": "..."
}
```

## 5.3 通信加密

```
┌─────────────────────────────────────────────────────────┐
│                    TLS 1.3 加密通道                      │
│  ┌─────────┐                    ┌─────────┐            │
│  │  Agent  │◀──────────────────▶│ Center  │            │
│  └─────────┘                    └─────────┘            │
│       │                              │                 │
│       │    gRPC over TLS             │                 │
│       │    (双向认证)                │                 │
└───────┴──────────────────────────────┴─────────────────┘
```

---

# 六、监控架构

## 6.1 监控指标

| 类别 | 指标 | 描述 |
|------|------|------|
| **系统指标** | CPU/Memory/Disk/Network | 资源使用情况 |
| **Agent指标** | 在线数量/任务执行数/错误率 | Agent状态 |
| **任务指标** | 提交数/完成数/延迟/失败率 | 任务统计 |
| **通信指标** | 消息数/延迟/丢包率 | 通信质量 |

## 6.2 监控架构

```
┌───────────┐   ┌───────────┐   ┌───────────┐
│  Center   │   │   Agent   │   │   Task    │
│  Metrics  │   │  Metrics  │   │  Metrics  │
└─────┬─────┘   └─────┬─────┘   └─────┬─────┘
      │               │               │
      └───────────────┼───────────────┘
                      │
                      ▼
              ┌───────────────┐
              │   Prometheus  │
              └───────┬───────┘
                      │
                      ▼
              ┌───────────────┐
              │    Grafana    │
              │  (Dashboard)  │
              └───────────────┘
```

---

# 七、容量规划

## 7.1 性能目标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 支持Agent数 | 100,000+ | 单Center集群 |
| 消息吞吐量 | 100,000条/秒 | 并发消息处理 |
| 任务调度延迟 | < 100ms | 从提交到分配 |
| 消息延迟 | < 50ms | 局域网内 |
| API响应时间 | < 100ms | 99%请求 |

## 7.2 资源估算

**Center节点（单节点）：**
- CPU: 4核
- 内存: 8GB
- 存储: 100GB SSD
- 网络: 1Gbps

**数据库：**
- PostgreSQL: 4核/16GB/500GB SSD
- Redis: 4核/32GB
- etcd: 2核/8GB/50GB SSD

---

# 八、扩展机制

## 8.1 插件架构

```go
// 插件接口
type Plugin interface {
    Name() string
    Version() string
    Init(config map[string]interface{}) error
    Start() error
    Stop() error
}

// 插件类型
type SkillPlugin interface {
    Plugin
    Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

type ToolPlugin interface {
    Plugin
    Run(args []string) ([]byte, error)
}

type SchedulerPlugin interface {
    Plugin
    Schedule(task *Task, agents []*Agent) *Agent
}
```

## 8.2 扩展点

| 扩展点 | 接口 | 用途 |
|--------|------|------|
| 调度策略 | SchedulerPlugin | 自定义调度算法 |
| 认证方式 | AuthPlugin | 自定义认证 |
| 存储后端 | StoragePlugin | 自定义存储 |
| 消息协议 | ProtocolPlugin | 自定义协议 |
| 监控指标 | MetricsPlugin | 自定义指标 |
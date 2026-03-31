# OFA 方案设计文档

---

# 一、Center服务端设计

## 1.1 项目结构

```
src/center/
├── cmd/
│   └── center/
│       └── main.go              # 入口文件
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP处理器
│   │   ├── middleware/          # 中间件
│   │   └── router.go            # 路由配置
│   ├── grpc/
│   │   ├── agent_service.go     # Agent服务
│   │   ├── message_service.go   # 消息服务
│   │   └── task_service.go      # 任务服务
│   ├── scheduler/
│   │   ├── scheduler.go         # 调度器
│   │   ├── policy/              # 调度策略
│   │   └── queue.go             # 任务队列
│   ├── router/
│   │   ├── router.go            # 消息路由
│   │   └── route_table.go       # 路由表
│   ├── storage/
│   │   ├── postgres/            # PostgreSQL
│   │   ├── redis/               # Redis
│   │   └── etcd/                # etcd
│   ├── agent/
│   │   ├── manager.go           # Agent管理
│   │   ├── registry.go          # 注册表
│   │   └── monitor.go           # 状态监控
│   └── config/
│       ├── config.go            # 配置管理
│       └── loader.go            # 配置加载
├── pkg/
│   ├── proto/                   # Protobuf生成
│   ├── auth/                    # 认证模块
│   ├── crypto/                  # 加密模块
│   └── log/                     # 日志模块
├── configs/
│   ├── config.yaml              # 主配置
│   └── skills/                  # 技能配置
└── go.mod
```

## 1.2 核心模块实现

### 1.2.1 Agent管理器

```go
package agent

import (
    "context"
    "sync"
    "time"

    "github.com/ofa/center/internal/storage"
)

type Manager struct {
    registry   *Registry
    storage    storage.AgentStorage
    monitor    *Monitor
    eventChan  chan AgentEvent
    mu         sync.RWMutex
}

type Agent struct {
    ID           string
    Name         string
    Type         AgentType
    Status       AgentStatus
    Capabilities []Capability
    LastSeen     time.Time
    Connection   Connection
    Metadata     map[string]string
}

func NewManager(storage storage.AgentStorage) *Manager {
    return &Manager{
        registry:  NewRegistry(),
        storage:   storage,
        monitor:   NewMonitor(),
        eventChan: make(chan AgentEvent, 1000),
    }
}

// Register 注册Agent
func (m *Manager) Register(ctx context.Context, req *RegisterRequest) (*Agent, error) {
    agent := &Agent{
        ID:           generateID(),
        Name:         req.Name,
        Type:         req.Type,
        Status:       StatusOnline,
        Capabilities: req.Capabilities,
        LastSeen:     time.Now(),
        Metadata:     req.Metadata,
    }

    // 持久化
    if err := m.storage.Save(ctx, agent); err != nil {
        return nil, err
    }

    // 加入注册表
    m.registry.Add(agent)

    // 发送事件
    m.eventChan <- AgentEvent{
        Type:  EventRegister,
        Agent: agent,
    }

    return agent, nil
}

// Heartbeat 处理心跳
func (m *Manager) Heartbeat(ctx context.Context, agentID string, status *AgentStatus) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    agent, ok := m.registry.Get(agentID)
    if !ok {
        return ErrAgentNotFound
    }

    agent.LastSeen = time.Now()
    agent.Status = status.Status
    agent.Resources = status.Resources

    // 更新存储
    return m.storage.UpdateStatus(ctx, agentID, status)
}

// CheckHealth 检查Agent健康状态
func (m *Manager) CheckHealth(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.checkAgentsHealth(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (m *Manager) checkAgentsHealth(ctx context.Context) {
    now := time.Now()
    timeout := 60 * time.Second

    m.registry.ForEach(func(agent *Agent) {
        if now.Sub(agent.LastSeen) > timeout {
            agent.Status = StatusOffline
            m.storage.UpdateStatus(ctx, agent.ID, &AgentStatus{
                Status: StatusOffline,
            })
            m.eventChan <- AgentEvent{
                Type:  EventOffline,
                Agent: agent,
            }
        }
    })
}
```

### 1.2.2 任务调度器

```go
package scheduler

import (
    "context"
    "time"

    "github.com/ofa/center/internal/agent"
)

type Scheduler struct {
    queue      *PriorityQueue
    policy     SchedulePolicy
    agentMgr   *agent.Manager
    dispatcher *Dispatcher
    resultChan chan *TaskResult
}

type Task struct {
    ID          string
    ParentID    string
    Type        TaskType
    Priority    int
    SkillID     string
    Input       map[string]interface{}
    Status      TaskStatus
    TargetAgent string
    CreatedAt   time.Time
    StartedAt   time.Time
    CompletedAt time.Time
}

func NewScheduler(agentMgr *agent.Manager, policy SchedulePolicy) *Scheduler {
    return &Scheduler{
        queue:      NewPriorityQueue(),
        policy:     policy,
        agentMgr:   agentMgr,
        dispatcher: NewDispatcher(),
        resultChan: make(chan *TaskResult, 1000),
    }
}

// Submit 提交任务
func (s *Scheduler) Submit(ctx context.Context, task *Task) error {
    task.ID = generateID()
    task.Status = TaskStatusPending
    task.CreatedAt = time.Now()

    return s.queue.Push(task)
}

// Schedule 执行调度
func (s *Scheduler) Schedule(ctx context.Context) {
    for {
        select {
        case task := <-s.queue.Pop():
            s.scheduleTask(ctx, task)
        case <-ctx.Done():
            return
        }
    }
}

func (s *Scheduler) scheduleTask(ctx context.Context, task *Task) {
    // 获取可用Agent
    agents := s.agentMgr.GetAvailableAgents(task.SkillID)
    if len(agents) == 0 {
        task.Status = TaskStatusFailed
        task.Error = "no available agent"
        return
    }

    // 使用策略选择Agent
    selected := s.policy.Select(task, agents)
    task.TargetAgent = selected.ID
    task.Status = TaskStatusRunning
    task.StartedAt = time.Now()

    // 分发任务
    s.dispatcher.Dispatch(ctx, task, selected)
}

// 调度策略接口
type SchedulePolicy interface {
    Select(task *Task, agents []*agent.Agent) *agent.Agent
}

// 混合策略
type HybridPolicy struct {
    weights PolicyWeights
}

type PolicyWeights struct {
    Capability float64 // 能力权重
    Load       float64 // 负载权重
    Latency    float64 // 延迟权重
    Power      float64 // 功耗权重
}

func (p *HybridPolicy) Select(task *Task, agents []*agent.Agent) *agent.Agent {
    var best *agent.Agent
    var bestScore float64

    for _, a := range agents {
        score := p.calculateScore(task, a)
        if score > bestScore {
            bestScore = score
            best = a
        }
    }

    return best
}

func (p *HybridPolicy) calculateScore(task *Task, agent *agent.Agent) float64 {
    score := 0.0

    // 能力匹配分
    if agent.HasCapability(task.SkillID) {
        score += p.weights.Capability * 100
    }

    // 负载分 (负载越低分越高)
    score += p.weights.Load * (100 - agent.Resources.CPUUsage)

    // 延迟分 (延迟越低分越高)
    latency := float64(agent.Resources.NetworkLatencyMs)
    score += p.weights.Latency * (100 - latency)

    // 电量分 (电量越高分越高)
    score += p.weights.Power * float64(agent.Resources.BatteryLevel)

    return score
}
```

### 1.2.3 消息路由器

```go
package router

import (
    "context"
    "sync"

    "github.com/ofa/proto"
)

type Router struct {
    connections  sync.Map // agentID -> Connection
    subscriptions *SubscriptionManager
    offlineQueue *OfflineQueue
}

type Connection interface {
    Send(ctx context.Context, msg *proto.Message) error
    Close() error
}

func NewRouter() *Router {
    return &Router{
        subscriptions: NewSubscriptionManager(),
        offlineQueue:  NewOfflineQueue(),
    }
}

// Route 路由消息
func (r *Router) Route(ctx context.Context, msg *proto.Message) error {
    switch msg.To {
    case "broadcast":
        return r.broadcast(ctx, msg)
    default:
        return r.routeToAgent(ctx, msg)
    }
}

// routeToAgent 路由到指定Agent
func (r *Router) routeToAgent(ctx context.Context, msg *proto.Message) error {
    conn, ok := r.connections.Load(msg.To)
    if !ok {
        // Agent离线，存入离线队列
        return r.offlineQueue.Push(msg)
    }

    return conn.(Connection).Send(ctx, msg)
}

// broadcast 广播消息
func (r *Router) broadcast(ctx context.Context, msg *proto.Message) error {
    var errors []error

    r.connections.Range(func(key, value interface{}) bool {
        if err := value.(Connection).Send(ctx, msg); err != nil {
            errors = append(errors, err)
        }
        return true
    })

    if len(errors) > 0 {
        return errors[0]
    }
    return nil
}

// multicast 组播消息
func (r *Router) multicast(ctx context.Context, msg *proto.Message, agentIDs []string) error {
    for _, agentID := range agentIDs {
        if err := r.routeToAgent(ctx, &proto.Message{
            MsgId:    msg.MsgId,
            From:     msg.From,
            To:       agentID,
            Action:   msg.Action,
            Payload:  msg.Payload,
        }); err != nil {
            return err
        }
    }
    return nil
}

// AddConnection 添加连接
func (r *Router) AddConnection(agentID string, conn Connection) {
    r.connections.Store(agentID, conn)

    // 处理离线消息
    go r.processOfflineMessages(agentID, conn)
}

// RemoveConnection 移除连接
func (r *Router) RemoveConnection(agentID string) {
    r.connections.Delete(agentID)
}

func (r *Router) processOfflineMessages(agentID string, conn Connection) {
    messages := r.offlineQueue.PopAll(agentID)
    for _, msg := range messages {
        conn.Send(context.Background(), msg)
    }
}
```

---

# 二、Agent客户端设计

## 2.1 Android Agent

### 2.1.1 项目结构

```
platforms/android/
├── app/
│   ├── src/main/
│   │   ├── java/com/ofa/agent/
│   │   │   ├── OFAAgent.kt           # 主入口
│   │   │   ├── core/
│   │   │   │   ├── AgentCore.kt      # 核心模块
│   │   │   │   ├── AgentConfig.kt    # 配置
│   │   │   │   └── AgentState.kt     # 状态
│   │   │   ├── communication/
│   │   │   │   ├── Connector.kt      # 连接器
│   │   │   │   ├── GrpcClient.kt     # gRPC客户端
│   │   │   │   └── MessageHandler.kt # 消息处理
│   │   │   ├── executor/
│   │   │   │   ├── TaskExecutor.kt   # 任务执行
│   │   │   │   └── Worker.kt         # 工作线程
│   │   │   ├── skills/
│   │   │   │   ├── SkillManager.kt   # 技能管理
│   │   │   │   └── builtin/          # 内置技能
│   │   │   ├── tools/
│   │   │   │   ├── ToolManager.kt    # 工具管理
│   │   │   │   └── builtin/          # 内置工具
│   │   │   └── ui/                   # 用户界面
│   │   ├── jni/                      # Native代码
│   │   └── assets/                   # 资源文件
│   └── build.gradle.kts
└── build.gradle.kts
```

### 2.1.2 核心实现

```kotlin
// OFAAgent.kt
package com.ofa.agent

class OFAAgent private constructor(
    private val config: AgentConfig
) {
    private val core: AgentCore
    private val connector: Connector
    private val executor: TaskExecutor
    private val skillManager: SkillManager
    private val toolManager: ToolManager

    init {
        core = AgentCore(config)
        connector = Connector(config.centerAddress)
        executor = TaskExecutor()
        skillManager = SkillManager()
        toolManager = ToolManager()

        setupComponents()
    }

    private fun setupComponents() {
        // 注册消息处理器
        connector.setMessageHandler { message ->
            when (message.action) {
                "task.assign" -> handleTaskAssign(message)
                "config.update" -> handleConfigUpdate(message)
                "skill.install" -> handleSkillInstall(message)
                else -> handleUnknownMessage(message)
            }
        }
    }

    suspend fun start() {
        // 1. 初始化核心
        core.initialize()

        // 2. 加载技能
        skillManager.loadBuiltinSkills()
        toolManager.loadBuiltinTools()

        // 3. 连接Center
        connector.connect()

        // 4. 注册Agent
        val capabilities = skillManager.getCapabilities()
        connector.register(
            AgentInfo(
                id = core.agentId,
                name = config.agentName,
                type = AgentType.MOBILE,
                capabilities = capabilities
            )
        )

        // 5. 启动心跳
        connector.startHeartbeat {
            AgentStatus(
                status = core.status,
                resources = core.getResourceUsage()
            )
        }

        // 6. 启动执行器
        executor.start()
    }

    private suspend fun handleTaskAssign(message: Message) {
        val task = parseTask(message.payload)
        val result = executor.execute(task)
        connector.sendTaskResult(task.id, result)
    }

    companion object {
        @Volatile
        private var instance: OFAAgent? = null

        fun getInstance(config: AgentConfig): OFAAgent {
            return instance ?: synchronized(this) {
                instance ?: OFAAgent(config).also { instance = it }
            }
        }
    }
}

// Connector.kt
package com.ofa.agent.communication

import io.grpc.ManagedChannel
import io.grpc.ManagedChannelBuilder
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*

class Connector(
    private val centerAddress: String
) {
    private var channel: ManagedChannel? = null
    private var stub: AgentServiceGrpcKt.AgentServiceCoroutineStub? = null
    private var sendChannel: SendChannel<Message>? = null
    private var messageHandler: (suspend (Message) -> Unit)? = null

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    suspend fun connect() {
        channel = ManagedChannelBuilder.forTarget(centerAddress)
            .usePlaintext()
            .build()

        stub = AgentServiceGrpcKt.AgentServiceCoroutineStub(channel!!)

        // 建立双向流
        val requestFlow = MutableSharedFlow<AgentMessage>()
        val responseFlow = stub!!.connect(requestFlow)

        // 接收消息
        scope.launch {
            responseFlow.collect { message ->
                messageHandler?.invoke(parseMessage(message))
            }
        }

        sendChannel = requestFlow.asSendChannel()
    }

    suspend fun register(info: AgentInfo) {
        val request = RegisterRequest {
            agentId = info.id
            name = info.name
            type = info.type.toProto()
            capabilities.addAll(info.capabilities.map { it.toProto() })
        }

        sendChannel?.send(AgentMessage {
            register = request
        })
    }

    fun startHeartbeat(statusProvider: () -> AgentStatus) {
        scope.launch {
            while (isActive) {
                delay(30000) // 30秒心跳

                val status = statusProvider()
                sendChannel?.send(AgentMessage {
                    heartbeat = HeartbeatRequest {
                        agentId = core.agentId
                        status = status.status.toProto()
                        resources = status.resources.toProto()
                    }
                })
            }
        }
    }

    fun setMessageHandler(handler: suspend (Message) -> Unit) {
        messageHandler = handler
    }

    suspend fun sendTaskResult(taskId: String, result: TaskResult) {
        sendChannel?.send(AgentMessage {
            taskResult = result.toProto()
        })
    }
}

// TaskExecutor.kt
package com.ofa.agent.executor

import kotlinx.coroutines.*
import java.util.concurrent.Executors

class TaskExecutor(
    private val workerCount: Int = Runtime.getRuntime().availableProcessors()
) {
    private val executor = Executors.newFixedThreadPool(workerCount)
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val skillManager: SkillManager = SkillManager.getInstance()
    private val toolManager: ToolManager = ToolManager.getInstance()

    suspend fun execute(task: Task): TaskResult = withContext(Dispatchers.IO) {
        val startTime = System.currentTimeMillis()

        try {
            // 1. 获取技能
            val skill = skillManager.getSkill(task.skillId)
                ?: return@withContext TaskResult(
                    taskId = task.id,
                    status = TaskStatus.FAILED,
                    error = "Skill not found: ${task.skillId}"
                )

            // 2. 验证输入
            val validationError = skill.validateInput(task.input)
            if (validationError != null) {
                return@withContext TaskResult(
                    taskId = task.id,
                    status = TaskStatus.FAILED,
                    error = validationError
                )
            }

            // 3. 执行技能
            val output = skill.execute(task.input)

            // 4. 返回结果
            TaskResult(
                taskId = task.id,
                status = TaskStatus.COMPLETED,
                output = output,
                durationMs = System.currentTimeMillis() - startTime
            )
        } catch (e: Exception) {
            TaskResult(
                taskId = task.id,
                status = TaskStatus.FAILED,
                error = e.message ?: "Unknown error",
                durationMs = System.currentTimeMillis() - startTime
            )
        }
    }
}
```

## 2.2 技能系统设计

```kotlin
// Skill.kt
package com.ofa.agent.skills

interface Skill {
    val id: String
    val name: String
    val version: String
    val category: String
    val description: String

    val inputs: List<ParamSpec>
    val outputs: List<ParamSpec>
    val requires: List<String>

    suspend fun execute(input: Map<String, Any?>): Map<String, Any?>

    fun validateInput(input: Map<String, Any?>): String? {
        for (spec in inputs) {
            if (spec.required && !input.containsKey(spec.name)) {
                return "Missing required parameter: ${spec.name}"
            }
            // 类型验证...
        }
        return null
    }
}

data class ParamSpec(
    val name: String,
    val type: ParamType,
    val required: Boolean = true,
    val description: String = "",
    val defaultValue: Any? = null
)

enum class ParamType {
    STRING, INTEGER, FLOAT, BOOLEAN, OBJECT, ARRAY, BINARY
}

// 内置技能示例：文本处理
class TextProcessingSkill : Skill {
    override val id = "text.process"
    override val name = "Text Processing"
    override val version = "1.0.0"
    override val category = "text"

    override val inputs = listOf(
        ParamSpec("text", ParamType.STRING, true, "Text to process"),
        ParamSpec("operation", ParamType.STRING, true, "Operation: uppercase, lowercase, reverse, wordcount")
    )

    override val outputs = listOf(
        ParamSpec("result", ParamType.STRING, true, "Processing result")
    )

    override val requires = emptyList<String>()

    override suspend fun execute(input: Map<String, Any?>): Map<String, Any?> {
        val text = input["text"] as String
        val operation = input["operation"] as String

        val result = when (operation) {
            "uppercase" -> text.uppercase()
            "lowercase" -> text.lowercase()
            "reverse" -> text.reversed()
            "wordcount" -> text.split("\\s+".toRegex()).size.toString()
            else -> throw IllegalArgumentException("Unknown operation: $operation")
        }

        return mapOf("result" to result)
    }
}

// SkillManager.kt
class SkillManager private constructor() {
    private val skills = mutableMapOf<String, Skill>()
    private val loaders = mutableListOf<SkillLoader>()

    companion object {
        @Volatile private var instance: SkillManager? = null

        fun getInstance(): SkillManager {
            return instance ?: synchronized(this) {
                instance ?: SkillManager().also { instance = it }
            }
        }
    }

    fun registerLoader(loader: SkillLoader) {
        loaders.add(loader)
    }

    fun loadBuiltinSkills() {
        registerSkill(TextProcessingSkill())
        registerSkill(JsonProcessingSkill())
        registerSkill(FileOperationSkill())
        registerSkill(NetworkRequestSkill())
        // ... 更多内置技能
    }

    fun registerSkill(skill: Skill) {
        skills[skill.id] = skill
    }

    fun getSkill(id: String): Skill? = skills[id]

    fun getCapabilities(): List<Capability> {
        return skills.values.map { skill ->
            Capability(
                id = skill.id,
                name = skill.name,
                version = skill.version,
                category = skill.category
            )
        }
    }
}
```

---

# 三、SDK设计

## 3.1 Java/Kotlin SDK

```kotlin
// OFAClient.kt
package com.ofa.sdk

class OFAClient private constructor(
    private val config: ClientConfig
) {
    private val grpcClient: GrpcClient

    init {
        grpcClient = GrpcClient(config.centerAddress)
    }

    /**
     * 提交任务
     */
    suspend fun submitTask(
        skillId: String,
        input: Map<String, Any?>,
        targetAgent: String? = null,
        priority: Int = 0
    ): TaskHandle {
        val taskId = grpcClient.submitTask(skillId, input, targetAgent, priority)
        return TaskHandle(taskId, grpcClient)
    }

    /**
     * 发送消息
     */
    suspend fun sendMessage(
        toAgent: String,
        action: String,
        payload: Map<String, Any?>
    ): MessageHandle {
        return grpcClient.sendMessage(toAgent, action, payload)
    }

    /**
     * 广播消息
     */
    suspend fun broadcast(
        action: String,
        payload: Map<String, Any?>
    ): List<String> {
        return grpcClient.broadcast(action, payload)
    }

    /**
     * 获取Agent列表
     */
    suspend fun listAgents(
        type: AgentType? = null,
        status: AgentStatus? = null
    ): List<AgentInfo> {
        return grpcClient.listAgents(type, status)
    }

    /**
     * 订阅Agent事件
     */
    fun subscribeEvents(
        filter: EventFilter = EventFilter()
    ): Flow<AgentEvent> {
        return grpcClient.subscribeEvents(filter)
    }

    companion object {
        fun create(config: ClientConfig): OFAClient {
            return OFAClient(config)
        }

        fun create(centerAddress: String): OFAClient {
            return OFAClient(ClientConfig(centerAddress))
        }
    }
}

// TaskHandle.kt
class TaskHandle(
    val taskId: String,
    private val client: GrpcClient
) {
    /**
     * 等待任务完成
     */
    suspend fun await(timeout: Duration = Duration.ofMinutes(5)): TaskResult {
        return client.waitForTask(taskId, timeout)
    }

    /**
     * 获取任务状态
     */
    suspend fun getStatus(): TaskStatus {
        return client.getTaskStatus(taskId)
    }

    /**
     * 取消任务
     */
    suspend fun cancel(): Boolean {
        return client.cancelTask(taskId)
    }

    /**
     * 监听任务进度
     */
    fun watchProgress(): Flow<TaskProgress> {
        return client.watchTaskProgress(taskId)
    }
}

// 使用示例
suspend fun example() {
    val client = OFAClient.create("center.example.com:9090")

    // 提交任务
    val handle = client.submitTask(
        skillId = "text.process",
        input = mapOf(
            "text" to "Hello World",
            "operation" to "uppercase"
        )
    )

    // 等待结果
    val result = handle.await()
    println("Result: ${result.output}")

    // 发送消息
    client.sendMessage(
        toAgent = "agent-123",
        action = "notify",
        payload = mapOf("message" to "Task completed")
    )
}
```

## 3.2 Python SDK

```python
# ofa_sdk/client.py
from typing import Dict, List, Any, Optional
from dataclasses import dataclass
from enum import Enum
import grpc
import asyncio
from concurrent.futures import ThreadPoolExecutor

class TaskStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

@dataclass
class TaskResult:
    task_id: str
    status: TaskStatus
    output: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    duration_ms: int = 0

@dataclass
class AgentInfo:
    id: str
    name: str
    type: str
    status: str
    capabilities: List[str]

class OFAClient:
    """OFA Python SDK Client"""

    def __init__(self, center_address: str):
        self.center_address = center_address
        self._channel = None
        self._stub = None

    def connect(self):
        """连接到Center"""
        self._channel = grpc.insecure_channel(self.center_address)
        import ofa_pb2_grpc
        self._stub = ofa_pb2_grpc.AgentServiceStub(self._channel)

    def close(self):
        """关闭连接"""
        if self._channel:
            self._channel.close()

    def submit_task(
        self,
        skill_id: str,
        input: Dict[str, Any],
        target_agent: Optional[str] = None,
        priority: int = 0,
        timeout: float = 300.0
    ) -> TaskResult:
        """
        提交任务并等待结果

        Args:
            skill_id: 技能ID
            input: 输入参数
            target_agent: 目标Agent (可选)
            priority: 优先级
            timeout: 超时时间(秒)

        Returns:
            TaskResult: 任务结果
        """
        import ofa_pb2
        import json

        request = ofa_pb2.TaskRequest(
            skill_id=skill_id,
            input=json.dumps(input).encode(),
            target_agent=target_agent or "",
            priority=priority
        )

        response = self._stub.SubmitTask(request, timeout=timeout)

        return TaskResult(
            task_id=response.task_id,
            status=TaskStatus(response.status),
            output=json.loads(response.output) if response.output else None,
            error=response.error or None,
            duration_ms=response.duration_ms
        )

    async def submit_task_async(
        self,
        skill_id: str,
        input: Dict[str, Any],
        **kwargs
    ) -> TaskResult:
        """异步提交任务"""
        loop = asyncio.get_event_loop()
        with ThreadPoolExecutor() as executor:
            return await loop.run_in_executor(
                executor,
                lambda: self.submit_task(skill_id, input, **kwargs)
            )

    def send_message(
        self,
        to_agent: str,
        action: str,
        payload: Dict[str, Any]
    ) -> bool:
        """发送消息"""
        import ofa_pb2
        import json

        request = ofa_pb2.MessageRequest(
            message=ofa_pb2.Message(
                to_agent=to_agent,
                action=action,
                payload=json.dumps(payload).encode()
            )
        )

        response = self._stub.SendMessage(request)
        return response.success

    def list_agents(
        self,
        agent_type: Optional[str] = None,
        status: Optional[str] = None
    ) -> List[AgentInfo]:
        """获取Agent列表"""
        import ofa_pb2

        request = ofa_pb2.ListAgentsRequest(
            type=agent_type or "",
            status=status or ""
        )

        response = self._stub.ListAgents(request)

        return [
            AgentInfo(
                id=agent.id,
                name=agent.name,
                type=agent.type,
                status=agent.status,
                capabilities=list(agent.capabilities)
            )
            for agent in response.agents
        ]

    def __enter__(self):
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

# 使用示例
if __name__ == "__main__":
    with OFAClient("localhost:9090") as client:
        # 提交任务
        result = client.submit_task(
            skill_id="text.process",
            input={"text": "Hello", "operation": "uppercase"}
        )
        print(f"Result: {result.output}")

        # 获取Agent列表
        agents = client.list_agents()
        for agent in agents:
            print(f"Agent: {agent.name} - {agent.status}")
```

---

# 四、部署方案

## 4.1 Docker部署

```dockerfile
# Dockerfile.center
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o center ./cmd/center

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/center .
COPY --from=builder /app/configs ./configs

EXPOSE 8080 9090
CMD ["./center"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  center:
    build:
      context: .
      dockerfile: Dockerfile.center
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - ETCD_ENDPOINTS=etcd:2379
    depends_on:
      - postgres
      - redis
      - etcd
    networks:
      - ofa-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=ofa
      - POSTGRES_USER=ofa
      - POSTGRES_PASSWORD=ofa123
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - ofa-network

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    networks:
      - ofa-network

  etcd:
    image: quay.io/coreos/etcd:v3.5
    command: etcd -advertise-client-urls=http://etcd:2379 -listen-client-urls=http://0.0.0.0:2379
    volumes:
      - etcd-data:/etcd-data
    networks:
      - ofa-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./deployments/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - ofa-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - ofa-network

networks:
  ofa-network:
    driver: bridge

volumes:
  postgres-data:
  redis-data:
  etcd-data:
  grafana-data:
```

## 4.2 Kubernetes部署

```yaml
# k8s/center-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ofa-center
  labels:
    app: ofa-center
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
        - name: REDIS_HOST
          value: "redis-service"
        - name: ETCD_ENDPOINTS
          value: "etcd-service:2379"
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2000m"
            memory: "2Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
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
    targetPort: 8080
    name: http
  - port: 9090
    targetPort: 9090
    name: grpc
  type: LoadBalancer
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ofa-center-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ofa-center
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```
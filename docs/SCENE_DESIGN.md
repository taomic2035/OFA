# OFA 场景联动设计文档

## 一、概述

场景联动系统是 OFA 分布式 Agent 架构的核心能力，实现跨设备智能协同。基于核心愿景 **"万物皆为我所用，万物皆是我"**，让不同设备在不同场景下智能配合，形成整体智能。

### 设计目标

| 目标 | 说明 |
|------|------|
| 场景感知 | 自动检测用户当前场景 (跑步、会议、健康异常等) |
| 智能路由 | 根据场景将消息路由到合适设备 |
| 协同执行 | 多设备协同完成任务 |
| 状态同步 | 跨设备状态实时同步 |

---

## 二、架构设计

### 2.1 场景引擎架构

```
SceneEngine (场景引擎)
├── SceneDetector (场景检测器接口)
│   ├── RunningDetector (跑步检测)
│   ├── MeetingDetector (会议检测)
│   └── HealthAlertDetector (健康异常检测)
│
├── SceneHandler (场景处理器接口)
│   ├── NotificationHandler (通知处理)
│   ├── RoutingHandler (路由处理)
│   └── AlertHandler (告警处理)
│
├── TriggerRule (触发规则)
│   ├── TriggerCondition (触发条件)
│   └── SceneAction (场景动作)
│
└── SceneListener (场景监听器)
    └── RunningSceneOrchestrator (跑步协调器)
    ├── MeetingSceneOrchestrator (会议协调器)
    └── HealthAlertSceneOrchestrator (健康告警协调器)
```

### 2.2 核心接口

```go
// SceneDetector - 场景检测器接口
type SceneDetector interface {
    Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error)
    GetType() SceneType
}

// SceneHandler - 场景处理器接口
type SceneHandler interface {
    Handle(ctx context.Context, scene *SceneState) error
    CanHandle(sceneType SceneType) bool
}

// SceneListener - 场景监听器接口
type SceneListener interface {
    OnSceneStart(scene *SceneState)
    OnSceneEnd(scene *SceneState)
    OnSceneAction(scene *SceneState, action *SceneAction)
}
```

---

## 三、场景类型定义

### 3.1 场景枚举

```go
type SceneType string

const (
    SceneRunning     SceneType = "running"      // 跑步
    SceneWalking     SceneType = "walking"      // 步行
    SceneDriving     SceneType = "driving"      // 驾驶
    SceneMeeting     SceneType = "meeting"      // 会议
    SceneSleeping    SceneType = "sleeping"     // 睡眠
    SceneExercise    SceneType = "exercise"     // 运动
    SceneWork        SceneType = "work"         // 工作
    SceneHome        SceneType = "home"         // 家庭
    SceneTravel      SceneType = "travel"       // 旅行
    SceneHealthAlert SceneType = "health_alert" // 健康异常
)
```

### 3.2 场景状态

```go
type SceneState struct {
    ID           string                 `json:"id"`
    Type         SceneType              `json:"type"`
    IdentityID   string                 `json:"identity_id"`
    AgentID      string                 `json:"agent_id"`
    StartTime    time.Time              `json:"start_time"`
    EndTime      *time.Time             `json:"end_time,omitempty"`
    Duration     time.Duration          `json:"duration"`
    Confidence   float64                `json:"confidence"`    // 检测置信度
    Context      map[string]interface{} `json:"context"`       // 场景上下文
    Actions      []SceneAction          `json:"actions"`       // 待执行动作
    Active       bool                   `json:"active"`
}
```

---

## 四、典型场景设计

### 4.1 跑步场景

**检测源**: 手表运动传感器

**联动逻辑**:

```
手表检测跑步
    │
    ├── SceneEngine.DetectScene() → SceneRunning
    │
    ├── RunningSceneOrchestrator.OnSceneStart()
    │   ├── 创建 RunningSession
    │   ├── 路由规则生效: RouteToPhone=true
    │   └── 手表端过滤: FilterOnWatch=true (仅紧急消息)
    │
    ├── 持续监测
    │   ├── 心率监测: >160 触发告警
    │   ├── 定期路由状态到手机显示
    │   └── 更新运动指标 (距离、步数、卡路里)
    │
    └── 跑步结束
        ├── HandleRunningEnd()
        ├── 生成运动总结
        └── 解除路由规则
```

**配置参数**:

| 参数 | 默认值 | 说明 |
|------|--------|------|
| MinDuration | 60s | 最小持续时间才触发路由 |
| HeartRateThreshold | 100 | 心率阈值 |
| RouteToPhone | true | 路由到手机显示 |
| FilterOnWatch | true | 手表端过滤非紧急消息 |

**路由规则**:

```yaml
rules:
  - name: running_route
    conditions:
      - type: scene
        value: running
        operator: equals
      - type: duration
        value: 60s
        operator: greater_than
    actions:
      - type: route
        target: phone
        payload:
          message_type: running_status
```

---

### 4.2 会议场景

**检测源**: 手机日历事件

**联动逻辑**:

```
手机日历检测会议
    │
    ├── SceneEngine.DetectScene() → SceneMeeting
    │
    ├── MeetingSceneOrchestrator.OnSceneStart()
    │   ├── 创建 MeetingSession
    │   ├── 启用 DND 模式
    │   ├── 拦截来电 (BlockCalls=true)
    │   ├── 仅允许紧急消息 (BlockExceptUrgent=true)
    │   └── 眼镜端静默提醒 (NotifyGlasses=true)
    │
    ├── 来电处理
    │   ├── HandleIncomingCall(caller, urgent)
    │   ├── 紧急来电: 允许 + 记录 AllowedMessages++
    │   └── 非紧急来电: 拦截 + 记录 BlockedCalls++
    │
    ├── 消息处理
    │   ├── 紧急消息: 允许
    │   └── 非紧急消息: 转眼镜静默提醒
    │
    └── 会议结束
        ├── HandleMeetingEnd()
        ├── 关闭 DND 模式
        └── 统计拦截信息
```

**配置参数**:

| 参数 | 默认值 | 说明 |
|------|--------|------|
| DNDModeEnabled | true | 启用勿扰模式 |
| DNDDefaultDuration | 1h | 默认勿扰时长 |
| NotifyGlasses | true | 眼镜端通知 |
| BlockCalls | true | 拦截来电 |
| BlockExceptUrgent | true | 仅允许紧急 |

**DND 模式状态**:

```go
type MeetingSession struct {
    IdentityID       string    `json:"identity_id"`
    PhoneAgentID     string    `json:"phone_agent_id"`
    GlassesAgentID   string    `json:"glasses_agent_id"`
    MeetingTitle     string    `json:"meeting_title"`
    DNDActive        bool      `json:"dnd_active"`
    BlockedCalls     int       `json:"blocked_calls"`
    AllowedMessages  int       `json:"allowed_messages"`
}
```

---

### 4.3 健康异常场景

**检测源**: 手表健康传感器

**联动逻辑**:

```
手表健康数据异常
    │
    ├── HealthAlertSceneOrchestrator.HandleHealthAlertDetection()
    │   ├── 检查冷却时间 (AlertCooldown=5min)
    │   ├── 分析健康数据 → 确定告警类型和严重度
    │   ├── 创建 HealthAlertSession
    │   └── 记录告警时间
    │
    ├── 告警广播
    │   ├── BroadcastToAll=true → 广播到所有设备
    │   ├── LogToCenter=true → 记录到 Center
    │   └── NotifyEmergencyContact=true (高危) → 通知紧急联系人
    │
    ├── 持续监测
    │   ├── HandleHealthValueUpdate()
    │   ├── 值恢复正常 → 自动解除告警
    │   └── 值持续异常 → 保持告警状态
    │
    └── 告警解除
        ├── HandleHealthAlertResolved()
        ├── 记录告警历史
        └── 移除活跃告警
```

**告警阈值**:

| 类型 | 阈值 | 严重度 |
|------|------|--------|
| 高心率 | >120 bpm | high |
| 低心率 | <50 bpm | high |
| 高心率(轻度) | >100 bpm | medium |
| 低氧 | <95% | high |
| 高血压 | 收缩压>140 | high |
| 高温 | >37.5°C | medium |

**冷却机制**:

```
AlertCooldown = 5 分钟

同一用户在冷却期内:
- 新告警数据不触发新告警
- 仅更新现有告警的值
- 冷却期结束后才能触发新告警
```

---

## 五、触发规则系统

### 5.1 规则结构

```go
type TriggerRule struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    SceneType   SceneType         `json:"scene_type"`
    Conditions  []TriggerCondition `json:"conditions"`
    Actions     []SceneAction     `json:"actions"`
    Priority    int               `json:"priority"`     // 优先级 (1-10)
    Enabled     bool              `json:"enabled"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type TriggerCondition struct {
    Type     string      `json:"type"`     // scene, duration, location, value, device
    Operator string      `json:"operator"` // equals, greater_than, less_than, contains, in_range
    Value    interface{} `json:"value"`
}

type SceneAction struct {
    Type        string                 `json:"type"`        // route, notify, block, alert, log
    TargetAgent string                 `json:"target_agent"`
    Payload     map[string]interface{} `json:"payload"`
    Delay       time.Duration          `json:"delay"`
}
```

### 5.2 条件类型

| 条件类型 | 说明 | 示例 |
|---------|------|------|
| scene | 场景类型匹配 | `scene=running` |
| duration | 持续时间阈值 | `duration>60s` |
| location | 位置区域匹配 | `location in [home, office]` |
| value | 数值阈值 | `heart_rate>120` |
| device | 设备类型匹配 | `device=watch` |

### 5.3 动作类型

| 动作类型 | 说明 | 参数 |
|---------|------|------|
| route | 路由消息 | target_agent, message_type |
| notify | 发送通知 | target_agent, message |
| block | 拦截消息 | block_type, except_urgent |
| alert | 发送告警 | severity, broadcast_to |
| log | 记录日志 | log_type, destination |

---

## 六、协调器实现

### 6.1 RunningSceneOrchestrator

**职责**: 跑步场景跨设备协调

```go
type RunningSceneOrchestrator struct {
    config        *RunningSceneConfig
    engine        *SceneEngine
    activeRuns    sync.Map // identityID -> *RunningSession
    eventHandler  RunningEventHandler
}

// 核心方法
func (o *RunningSceneOrchestrator) HandleRunningDetection(ctx, identityID, watchAgentID, data) error
func (o *RunningSceneOrchestrator) HandleRunningEnd(ctx, identityID, watchAgentID, summary) error
func (o *RunningSceneOrchestrator) RouteMessageToPhone(ctx, identityID, message) error
```

**事件接口**:

```go
type RunningEventHandler interface {
    OnRunningStart(session *RunningSession)
    OnRunningEnd(session *RunningSession)
    OnRouteToPhone(event *RouteEvent)
    OnWatchFilter(identityID, message string)
    OnHeartRateAlert(identityID string, heartRate float64)
}
```

### 6.2 MeetingSceneOrchestrator

**职责**: 会议场景跨设备协调

```go
type MeetingSceneOrchestrator struct {
    config         *MeetingSceneConfig
    engine         *SceneEngine
    activeMeetings sync.Map // identityID -> *MeetingSession
    eventHandler   MeetingEventHandler
}

// 核心方法
func (o *MeetingSceneOrchestrator) HandleMeetingDetection(ctx, identityID, phoneAgentID, data) error
func (o *MeetingSceneOrchestrator) HandleMeetingEnd(ctx, identityID) error
func (o *MeetingSceneOrchestrator) HandleIncomingCall(ctx, identityID, caller, urgent) (bool, error)
func (o *MeetingSceneOrchestrator) HandleIncomingMessage(ctx, identityID, message) (bool, error)
```

### 6.3 HealthAlertSceneOrchestrator

**职责**: 健康异常场景跨设备协调

```go
type HealthAlertSceneOrchestrator struct {
    config        *HealthAlertSceneConfig
    engine        *SceneEngine
    activeAlerts  sync.Map // identityID -> *HealthAlertSession
    alertHistory  sync.Map // identityID -> []*HealthAlertSession
    eventHandler  HealthAlertEventHandler
    lastAlertTime sync.Map // identityID -> time.Time (冷却机制)
}

// 核心方法
func (o *HealthAlertSceneOrchestrator) HandleHealthAlertDetection(ctx, identityID, watchAgentID, data) error
func (o *HealthAlertSceneOrchestrator) HandleHealthAlertResolved(ctx, identityID, resolution) error
func (o *HealthAlertSceneOrchestrator) HandleHealthValueUpdate(ctx, identityID, dataType, value) error
```

---

## 七、与 Center 集成

### 7.1 WebSocket 消息类型

| 消息类型 | 方向 | 场景用途 |
|---------|------|---------|
| BehaviorReport | Agent → Center | 上报行为观察 (含场景数据) |
| StateUpdate | Center → Agent | 推送场景状态变更 |
| SyncRequest | Agent → Center | 请求同步场景规则 |
| SyncResponse | Center → Agent | 返回同步数据 |

### 7.2 场景状态同步

```
Agent A (手表)              Center               Agent B (手机)
     │                        │                       │
     │  1. BehaviorReport     │                       │
     │  (running, heart_rate) │                       │
     │ ──────────────────────>│                       │
     │                        │  2. StateUpdate       │
     │                        │  (scene=running)      │
     │                        │ ─────────────────────>│
     │                        │                       │
     │                        │  3. 路由规则生效       │
     │                        │  手机接收跑步状态      │
     │                        │                       │
```

### 7.3 Center 存储

**场景历史记录**:

```sql
CREATE TABLE scene_history (
    id VARCHAR(36) PRIMARY KEY,
    identity_id VARCHAR(36) NOT NULL,
    scene_type VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration INTEGER,
    context JSONB,
    actions JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**告警记录**:

```sql
CREATE TABLE health_alerts (
    id VARCHAR(36) PRIMARY KEY,
    identity_id VARCHAR(36) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    value FLOAT NOT NULL,
    threshold FLOAT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    devices_notified JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 八、使用示例

### 8.1 场景引擎初始化

```go
// 创建场景引擎
engine := scene.NewSceneEngine(scene.DefaultSceneEngineConfig())

// 注册检测器
engine.RegisterDetector(scene.NewRunningDetector())
engine.RegisterDetector(scene.NewMeetingDetector())
engine.RegisterDetector(scene.NewHealthAlertDetector())

// 注册处理器
engine.RegisterHandler(scene.NewNotificationHandler())
engine.RegisterHandler(scene.NewRoutingHandler())
engine.RegisterHandler(scene.NewAlertHandler())

// 创建协调器
runningOrch := scene.NewRunningSceneOrchestrator(engine, nil)
meetingOrch := scene.NewMeetingSceneOrchestrator(engine, nil)
healthOrch := scene.NewHealthAlertSceneOrchestrator(engine, nil)
```

### 8.2 场景检测

```go
// 检测跑步场景
data := map[string]interface{}{
    "activity_type": "running",
    "heart_rate":    140.0,
    "distance":      2.5,
    "steps":         3000,
}

sceneState, err := engine.DetectScene(ctx, "watch-agent-001", data)
if sceneState.Type == scene.SceneRunning {
    // 处理跑步场景
}
```

### 8.3 添加触发规则

```go
// 创建路由规则
rule := &scene.TriggerRule{
    ID:        "rule-001",
    Name:      "Running Route Rule",
    SceneType: scene.SceneRunning,
    Conditions: []scene.TriggerCondition{
        {Type: "scene", Operator: "equals", Value: "running"},
        {Type: "duration", Operator: "greater_than", Value: 60},
    },
    Actions: []scene.SceneAction{
        {Type: "route", TargetAgent: "phone-agent-001", Payload: map[string]interface{}{
            "message_type": "running_status",
        }},
    },
    Priority: 5,
    Enabled:  true,
}

engine.AddRule(rule)
```

### 8.4 监听场景事件

```go
// 实现监听器
type MySceneListener struct{}

func (l *MySceneListener) OnSceneStart(scene *scene.SceneState) {
    log.Printf("Scene started: %s for identity %s", scene.Type, scene.IdentityID)
}

func (l *MySceneListener) OnSceneEnd(scene *scene.SceneState) {
    log.Printf("Scene ended: %s, duration: %v", scene.Type, scene.Duration)
}

func (l *MySceneListener) OnSceneAction(scene *scene.SceneState, action *scene.SceneAction) {
    log.Printf("Action executed: %s", action.Type)
}

// 注册监听器
engine.AddListener(&MySceneListener{})
```

---

## 九、性能考量

### 9.1 检测性能

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 检测延迟 | <100ms | 从数据输入到场景判定 |
| 规则匹配 | <50ms | 从场景触发到动作执行 |
| 状态更新 | <200ms | 跨设备状态同步延迟 |

### 9.2 资源使用

| 资源 | 预估 | 说明 |
|------|------|------|
| 内存 | ~10MB | 活跃场景缓存 |
| CPU | <5% | 场景检测计算 |
| 网络 | ~1KB/s | 状态同步流量 |

### 9.3 并发处理

```go
// 使用 sync.Map 处理并发场景
activeScenes sync.Map // 支持并发读写
sceneHistory sync.Map // 历史记录并发存储

// 检测器并发安全
func (d *RunningDetector) Detect(ctx context.Context, ...) (*SceneState, error) {
    // 无共享状态，线程安全
}
```

---

## 十、扩展场景

### 10.1 驾驶场景 (待实现)

**检测源**: 手机位置 + 车载蓝牙

**联动逻辑**:
- 来电转车载蓝牙
- 非紧急消息延迟
- 导航优先显示

### 10.2 睡眠场景 (待实现)

**检测源**: 手表睡眠监测

**联动逻辑**:
- 全设备静音
- 仅健康告警突破
- 智能唤醒

### 10.3 工作场景 (待实现)

**检测源**: 手机位置 + 应用使用

**联动逻辑**:
- 工作相关通知优先
- 社交消息延后
- 会议日程联动

---

## 十一、测试验证

### 11.1 单元测试

```go
// 检测器测试
func TestRunningDetector(t *testing.T) {
    detector := NewRunningDetector()
    
    data := map[string]interface{}{
        "activity_type": "running",
        "heart_rate":    120.0,
    }
    
    state, err := detector.Detect(ctx, "watch-001", data)
    assert.NoError(t, err)
    assert.Equal(t, SceneRunning, state.Type)
    assert.Greater(t, state.Confidence, 0.7)
}

// 触发条件测试
func TestTriggerCondition(t *testing.T) {
    cond := TriggerCondition{
        Type:     "heart_rate",
        Operator: "greater_than",
        Value:    100,
    }
    
    assert.True(t, evaluateCondition(cond, 120.0))
    assert.False(t, evaluateCondition(cond, 80.0))
}
```

### 11.2 性能测试

```go
func BenchmarkSceneDetection(b *testing.B) {
    engine := NewSceneEngine(DefaultSceneEngineConfig())
    data := map[string]interface{}{
        "activity_type": "running",
        "heart_rate":    120.0,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.DetectScene(ctx, "watch-001", data)
    }
}
```

---

## 十二、版本历史

| 版本 | 功能 |
|------|------|
| v7.3.0 | 场景引擎核心框架 |
| v7.3.0 | 跑步场景实现 |
| v7.3.0 | 会议场景实现 |
| v7.3.0 | 健康异常场景实现 |
| v7.3.0 | 触发规则系统 |

---

*文档版本: v7.4.0*
*更新时间: 2026-04-11*
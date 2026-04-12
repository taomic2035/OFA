# OFA 端到端联调验证计划 (v9.5.0)

## 一、验证目标

验证 Center + Android SDK + iOS SDK 的完整通信链路，确保多设备协同功能正常工作。

## 二、验证范围

### 通信链路验证
| 链路 | 测试内容 |
|------|---------|
| Center ↔ Android SDK | WebSocket 连接、消息收发、状态同步 |
| Center ↔ iOS SDK | WebSocket 连接、消息收发、身份同步 |
| Android ↔ iOS (via Center) | 跨设备消息路由、场景联动 |

### 功能验证
| 功能 | 测试场景 |
|------|---------|
| Agent 注册 | Android/iOS 设备注册到 Center |
| 身份同步 | 创建/更新/同步身份 |
| 场景检测 | 跑步/会议/健康异常场景触发 |
| 音频播放 | TTS 流播放、对话音频 |
| LLM 对话 | Claude/GPT 对话接口 |

## 三、测试环境

### 服务端
```bash
# 启动 Center 服务
cd src/center
go run cmd/center/main.go

# 或使用 Docker
docker-compose up -d
```

### 客户端
- Android: 运行 SDK 示例应用
- iOS: 运行 SwiftUI 示例应用

## 四、验证步骤

### 4.1 基础通信验证

#### Step 1: Center 服务启动验证
```bash
# 检查服务状态
curl http://localhost:8080/api/v1/status

# 检查 WebSocket 端点
curl http://localhost:8080/ws
```

#### Step 2: Android SDK 连接验证
```java
// OFAAndroidAgent 连接测试
OFAAndroidAgent agent = new OFAAndroidAgent(
    new AgentConfig.Builder()
        .centerAddress("ws://localhost:8080/ws")
        .mode(RunMode.SYNC)
        .build()
);

agent.initialize();
agent.connectCenter();

// 验证连接状态
assertEquals(AgentStatus.ONLINE, agent.getStatus());
```

#### Step 3: iOS SDK 连接验证
```swift
// OFAiOSAgent 连接测试
let config = AgentConfig(
    centerAddress: "ws://localhost:8080/ws",
    mode: .sync
)

let agent = OFAiOSAgent(config: config)
try await agent.initialize()
try await agent.connectCenter()

// 验证连接状态
XCTAssertEqual(agent.status, .online)
```

### 4.2 身份同步验证

#### Step 4: 创建身份
```java
// Android 创建身份
IdentityManager identityManager = agent.getIdentityManager();
PersonalIdentity identity = identityManager.createIdentity("测试用户");

// 验证身份创建
assertNotNull(identity.getId());
assertEquals("测试用户", identity.getName());
```

```swift
// iOS 创建身份
let identityManager = IdentityManager()
try await identityManager.initialize()
let identity = try await identityManager.createIdentity(name: "测试用户")

// 验证身份创建
XCTAssertNotNil(identity.id)
XCTAssertEqual(identity.name, "测试用户")
```

#### Step 5: 身份同步到 Center
```java
// Android 同步
identityManager.syncToCenter();

// iOS 同步
try await identityManager.syncFromCenter()
```

### 4.3 场景联动验证

#### Step 6: 跑步场景触发
```java
// Android 场景检测
SceneDetector detector = agent.getSceneDetector();
detector.detect();

// 触发跑步场景
SceneContext context = new SceneContext.Builder()
    .type(SceneType.RUNNING)
    .confidence(0.9)
    .build();

detector.triggerScene(context);
```

```swift
// iOS 场景检测
let detector = SceneDetector()
detector.initialize()
await detector.detect()

// 验证场景状态
if let scene = detector.getActiveScene() {
    XCTAssertEqual(scene.type, .running)
}
```

#### Step 7: 跨设备消息路由
```java
// 发送消息到其他设备
CrossDeviceRouter router = agent.getCrossDeviceRouter();
router.routeMessage(
    new Message.Builder()
        .type("running_status")
        .targetDevice("iphone")
        .payload("{\"distance\": 5000}")
        .build()
);
```

### 4.4 音频播放验证

#### Step 8: TTS 流请求
```java
// Android TTS 请求
AudioStreamReceiver receiver = agent.getAudioStreamReceiver();
receiver.requestTTSStream("你好，这是一个测试");

// 验证音频播放
AudioPlayer player = agent.getAudioPlayer();
assertTrue(player.isPlaying());
```

```swift
// iOS TTS 请求
let receiver = AudioStreamReceiver(audioPlayer: audioPlayer)
receiver.handleStreamStart(streamId: "tts_001", format: "pcm", sampleRate: 24000)

// 模拟音频数据接收
let audioData = Data(repeating: 0, count: 1000)
receiver.handleStreamChunk(streamId: "tts_001", audioData: audioData)
receiver.handleStreamEnd(streamId: "tts_001")

// 验证播放状态
XCTAssertEqual(audioPlayer.playbackState, .playing)
```

### 4.5 LLM 对话验证

#### Step 9: Claude 对话
```bash
# REST API 对话测试
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好", "session_id": "test_session"}'
```

#### Step 10: 流式对话
```bash
# SSE 流式对话测试
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "请介绍一下你自己", "session_id": "test_session"}'
```

## 五、验证检查清单

### 通信层
- [ ] Center 服务正常启动
- [ ] WebSocket 端点可访问
- [ ] Android SDK 可连接
- [ ] iOS SDK 可连接
- [ ] 心跳消息正常收发
- [ ] 注册消息正常处理

### 身份层
- [ ] 身份创建成功
- [ ] 身份同步到 Center
- [ ] 身份从 Center 恢复
- [ ] 行为观察上报成功
- [ ] 性格推断正常触发

### 场景层
- [ ] 场景检测器初始化
- [ ] 场景触发正常
- [ ] 场景动作执行
- [ ] 跨设备消息路由
- [ ] 场景监听器回调

### 音频层
- [ ] TTS 流请求成功
- [ ] 音频数据接收
- [ ] 音频播放正常
- [ ] 播放控制 (pause/resume/stop)

### LLM 层
- [ ] Claude API 连接成功
- [ ] 对话响应正常
- [ ] 流式响应正常
- [ ] 会话历史保存

## 六、测试脚本

### 自动化测试脚本
```bash
# 运行完整验证
./scripts/e2e-test.sh

# 运行单个模块测试
./scripts/test-center.sh
./scripts/test-android.sh
./scripts/test-ios.sh
```

### 测试报告
测试完成后生成报告:
- 测试覆盖率
- 失败案例列表
- 性能指标
- 通信延迟统计

## 七、预期结果

| 测试项 | 预期结果 |
|--------|---------|
| Center 启动 | 服务正常，端口监听 |
| Android 连接 | WebSocket 连接成功，状态 ONLINE |
| iOS 连接 | WebSocket 连接成功，状态 ONLINE |
| 身份同步 | 双端身份一致 |
| 场景联动 | 消息正确路由 |
| 音频播放 | 流播放正常 |
| LLM 对话 | 响应正确返回 |

---

*计划创建时间: 2026-04-11*
*版本: v9.5.0*

---

## 八、WebSocket集成测试 (v9.5.0)

### 测试用例列表

| 测试用例 | 说明 | 文件 |
|---------|------|------|
| TestWebSocketLifecycle | 连接生命周期（注册、心跳、断开） | integration_test.go |
| TestIdentitySynchronization | 身份同步（更新、请求、响应） | integration_test.go |
| TestSceneDetection | 场景检测（跑步场景、动作广播） | integration_test.go |
| TestMultiDeviceCoordination | 多设备协调（3设备注册、消息路由） | integration_test.go |
| TestBehaviorObservation | 行为观察（上报、性格推断） | integration_test.go |
| TestTaskAssignment | 任务分配（技能请求、结果上报） | integration_test.go |
| TestErrorHandling | 错误处理（无效消息、缺失字段） | integration_test.go |
| TestConnectionRecovery | 连接恢复（断开后重连） | integration_test.go |

### 运行集成测试
```bash
# 运行所有集成测试
cd src/center
go test -v ./tests/e2e/... -timeout 120s

# 运行单个测试
go test -v ./tests/e2e/... -run TestWebSocketLifecycle

# 运行带详细输出
go test -v ./tests/e2e/... -run TestMultiDeviceCoordination -timeout 60s
```

### WebSocket消息协议

#### 注册消息
```json
{
    "type": "Register",
    "timestamp": 1234567890,
    "payload": {
        "agent_id": "device_001",
        "device_type": "android",
        "device_name": "Test Device",
        "identity_id": "identity_001",
        "capabilities": ["voice", "display", "camera"]
    }
}
```

#### 心跳消息
```json
{
    "type": "Heartbeat",
    "timestamp": 1234567891,
    "payload": {
        "agent_id": "device_001",
        "session_id": "session_001",
        "status": "online",
        "battery": 85,
        "network_type": "wifi"
    }
}
```

#### 场景数据上报
```json
{
    "type": "StateUpdate",
    "timestamp": 1234567892,
    "payload": {
        "agent_id": "device_001",
        "session_id": "session_001",
        "update_type": "scene_data",
        "data": {
            "activity_type": "running",
            "heart_rate": 145,
            "duration": 1800
        }
    }
}
```

---

## 九、自动化集成测试脚本 (v9.5.0)

### 脚本使用
```bash
# 运行完整集成测试
./scripts/integration-test.sh

# 确保Center服务已启动
cd src/center && go run ./cmd/center

# 然后运行测试脚本
./scripts/integration-test.sh
```

### 脚本功能
| 功能 | 说明 |
|------|------|
| REST API测试 | 8个端点验证 |
| WebSocket测试 | 连接、消息协议 |
| 场景流程测试 | 检测、历史、活跃 |
| 身份同步测试 | 创建、更新、行为、推断 |
| 多设备测试 | 3设备注册、跨设备消息 |
| 测试报告生成 | JSON格式报告 |

---

## 十、场景检测验证 (v9.4.0)

### 已支持场景

| 场景 | 检测器 | 检测方式 |
|------|--------|---------|
| Running | RunningDetector | 活动类型、心率、时长、位置 |
| Meeting | MeetingDetector | 日历事件、位置、音频、时间 |
| HealthAlert | HealthAlertDetector | 心率阈值、血压、体温、血氧 |
| Driving | DrivingDetector | GPS速度、蓝牙车载、导航 |
| Exercise | ExerciseDetector | 心率、健身App、卡路里、时长 |
| Sleeping | SleepDetector | 时间、心率、光线、活动、呼吸 |
| Work | WorkDetector | 位置、时间、App、网络、会议 |
| Home | HomeDetector | 位置、时间、网络、家人、娱乐App |

### 场景测试用例
```bash
# 测试跑步场景检测
curl -X POST http://localhost:8080/api/v1/scene/detect \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "test_agent",
    "data": {
        "activity_type": "running",
        "heart_rate": 145,
        "duration": 1800,
        "steps": 3500,
        "location": "outdoor"
    }
  }'

# 测试驾驶场景检测
curl -X POST http://localhost:8080/api/v1/scene/detect \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "test_agent",
    "data": {
        "activity_type": "driving",
        "speed": 60,
        "bluetooth_devices": [{"name": "BMW Car Audio"}],
        "navigation_active": true
    }
  }'

# 测试睡眠场景检测
curl -X POST http://localhost:8080/api/v1/scene/detect \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "test_agent",
    "data": {
        "activity_type": "sleep",
        "heart_rate": 55,
        "movement": 5,
        "light_level": 0,
        "hour": 23
    }
  }'
```

---

## 十一、SDK错误处理验证 (v9.4.0)

### ErrorHandler框架测试
```java
// Android SDK错误处理测试
ErrorHandler.RetryExecutor executor = 
    new ErrorHandler.RetryExecutor(ErrorHandler.RetryConfig.defaultConfig());

// 测试重试机制
CompletableFuture<String> result = executor.execute(
    attempt -> {
        if (attempt < 3) {
            return CompletableFuture.failedFuture(new Exception("Network error"));
        }
        return CompletableFuture.completedFuture("Success");
    },
    ErrorHandler.CircuitBreaker.defaultBreaker("test")
);

// 验证结果
result.whenComplete((value, error) -> {
    if (error == null) {
        System.out.println("Retry succeeded: " + value);
    }
});
```

### CircuitBreaker测试
```java
// 测试熔断器状态转换
ErrorHandler.CircuitBreaker breaker = 
    ErrorHandler.CircuitBreaker.defaultBreaker("center");

// CLOSED状态 -> 允许执行
assertTrue(breaker.allowExecution());

// 模拟5次失败 -> OPEN状态
for (int i = 0; i < 5; i++) {
    breaker.recordFailure();
}
assertEquals(ErrorHandler.CircuitBreaker.State.OPEN, breaker.getState());

// 等待恢复时间 -> HALF_OPEN状态
Thread.sleep(30000);
assertTrue(breaker.allowExecution());
assertEquals(ErrorHandler.CircuitBreaker.State.HALF_OPEN, breaker.getState());

// 2次成功 -> CLOSED状态
breaker.recordSuccess();
breaker.recordSuccess();
assertEquals(ErrorHandler.CircuitBreaker.State.CLOSED, breaker.getState());
```

### 错误分类测试
```java
// 测试错误分类
OFAError networkError = ErrorHandler.categorizeError(
    new java.net.SocketTimeoutException("Connection timed out"));
assertEquals(ErrorHandler.ErrorCategory.NETWORK, networkError.getCategory());
assertEquals(ErrorHandler.RecoveryStrategy.BACKOFF_RETRY, networkError.getStrategy());

OFAError authError = ErrorHandler.categorizeError(
    new SecurityException("Authentication failed"));
assertEquals(ErrorHandler.ErrorCategory.AUTHENTICATION, authError.getCategory());
assertEquals(ErrorHandler.RecoveryStrategy.MANUAL_INTERVENTION, authError.getStrategy());
```
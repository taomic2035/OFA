# OFA 端到端联调验证计划 (v8.2.0)

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
*版本: v8.2.0*
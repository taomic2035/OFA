# OFA 端到端测试场景文档

## 测试概述

本文档定义了 OFA Center 与 Android SDK 之间的端到端测试场景，验证核心功能流程。

---

## 一、身份同步场景 (Identity Sync)

### 场景 1.1: 创建身份

**流程**:
1. Android SDK 通过 REST API 创建身份
2. Center 存储身份数据
3. 返回成功响应

**测试步骤**:
```bash
# 创建身份
curl -X POST http://localhost:8080/api/v1/identities \
  -H "Content-Type: application/json" \
  -d '{
    "id": "identity_001",
    "name": "测试用户",
    "personality": {
      "openness": 0.7,
      "conscientiousness": 0.6
    }
  }'

# 验证
curl http://localhost:8080/api/v1/identities/identity_001
```

**预期结果**:
- 身份创建成功 (200 OK)
- 身份数据可被检索
- 默认值已填充

### 场景 1.2: 身份同步

**流程**:
1. 设备 A 修改身份属性
2. 数据同步到 Center
3. 设备 B 从 Center 获取更新

**测试步骤**:
```bash
# 设备 A 同步更新
curl -X POST http://localhost:8080/api/v1/sync \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "device_001",
    "identity_id": "identity_001",
    "sync_type": "delta",
    "changes": [{"key": "preference.theme", "value": "dark"}]
  }'

# 设备 B 获取身份
curl http://localhost:8080/api/v1/identities/identity_001
```

**预期结果**:
- 同步成功
- 设备 B 可获取最新数据
- 版本号更新

---

## 二、设备管理场景 (Device Management)

### 场景 2.1: 设备注册

**流程**:
1. 设备启动时向 Center 注册
2. Center 分配设备 ID
3. 建立心跳机制

**测试步骤**:
```bash
# 设备注册
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "device_001",
    "identity_id": "identity_001",
    "device_type": "mobile",
    "device_name": "测试手机",
    "capabilities": ["ui_automation", "voice"]
  }'

# 发送心跳
curl -X POST http://localhost:8080/api/v1/devices/device_001/heartbeat \
  -H "Content-Type: application/json" \
  -d '{"status": "online", "battery": 85}'
```

**预期结果**:
- 设备注册成功
- 设备状态变为 online
- 心跳时间更新

### 场景 2.2: 多设备优先级

**流程**:
1. 注册多个设备
2. 设置设备优先级
3. Center 选择主设备

**预期结果**:
- 优先级设置成功
- 主设备标记正确
- 冲突时优先级设备优先

---

## 三、行为上报场景 (Behavior Report)

### 场景 3.1: 行为收集

**流程**:
1. Agent 收集用户行为
2. 上报行为观察到 Center
3. Center 存储并推断性格

**测试步骤**:
```bash
# 上报行为
curl -X POST http://localhost:8080/api/v1/behaviors \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "device_001",
    "identity_id": "identity_001",
    "type": "decision",
    "observation": {
      "decision_type": "impulse_purchase",
      "context": {"product": "电子产品"}
    }
  }'
```

**预期结果**:
- 行为存储成功
- 达到阈值触发性格推断
- 性格特质更新

### 场景 3.2: 性格推断

**流程**:
1. 收集足够行为观察
2. Center 触发性格推断
3. 更新身份性格画像

**预期结果**:
- 性格推断触发
- 性格模型更新
- 决策上下文反映新性格

---

## 四、情绪系统场景 (Emotion System)

### 场景 4.1: 情绪触发

**流程**:
1. 事件触发情绪
2. 情绪状态更新
3. 影响决策上下文

**测试步骤**:
```bash
# 触发喜悦情绪
curl -X POST http://localhost:8080/api/v1/emotions/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "identity_001",
    "trigger_type": "event",
    "trigger_desc": "获得奖励",
    "emotion_type": "joy",
    "intensity": 0.8
  }'

# 获取情绪上下文
curl http://localhost:8080/api/v1/emotions/identity_001/context
```

**预期结果**:
- 情绪状态更新
- 主导情绪为 joy
- 决策倾向反映情绪影响

### 场景 4.2: 情绪衰减

**流程**:
1. 情绪随时间衰减
2. 情绪回归基准
3. 长期稳定状态

**预期结果**:
- 情绪强度降低
- 心境稳定化
- 情绪画像形成

---

## 五、三观系统场景 (Philosophy System)

### 场景 5.1: 世界观设置

**测试步骤**:
```bash
# 更新世界观
curl -X POST http://localhost:8080/api/v1/philosophy/worldview \
  -H "Content-Type: application/json" \
  -d '{
    "identity_id": "identity_001",
    "optimism": 0.7,
    "change_belief": 0.6,
    "trust_in_people": 0.8
  }'

# 获取决策上下文
curl http://localhost:8080/api/v1/philosophy/identity_001/context
```

**预期结果**:
- 世界观设置成功
- 决策倾向计算正确
- 风险容忍度反映乐观程度

---

## 六、TTS 场景 (Text-to-Speech)

### 场景 6.1: 语音合成

**流程**:
1. 发送文本请求
2. Center 合成语音
3. 返回音频数据

**测试步骤**:
```bash
# 合成语音
curl -X POST http://localhost:8080/api/v1/tts/synthesize \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你好，欢迎使用 OFA 系统",
    "voice_id": "zh_female_qingxin",
    "speed": 1.0,
    "format": "mp3"
  }'

# 获取声音列表
curl http://localhost:8080/api/v1/tts/voices
```

**预期结果**:
- 语音合成成功
- 音频数据返回
- 身份声音映射正确

---

## 七、完整端到端流程

### 场景 7.1: 用户首次使用

**完整流程**:
1. 创建身份 → 2. 注册设备 → 3. 设置性格 → 4. 设置三观 → 5. 触发初始情绪 → 6. TTS欢迎语音

**验证点**:
- 每一步返回正确状态
- 数据持久化正确
- 各组件联动正常

### 场景 7.2: 跨设备同步

**完整流程**:
1. 设备 A 修改偏好 → 2. 同步到 Center → 3. 设备 B 获取更新 → 4. 确认一致性

**验证点**:
- 同步延迟 < 1秒
- 数据一致性保证
- 冲突正确解决

---

## 八、测试执行方式

### Go 测试

```bash
cd src/center
go test ./tests/e2e/... -v
```

### 手动测试

```bash
# 启动 Center
./center

# 执行测试脚本
./scripts/test_e2e.sh
```

### 测试报告

测试完成后生成报告:
- 测试通过率
- 响应时间统计
- 错误日志分析
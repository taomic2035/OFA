# OpenHarmony SDK

OpenHarmony 设备的 OFA Agent SDK，支持鸿蒙系统设备接入。

## 环境要求

- OpenHarmony 3.1+
- DevEco Studio 3.0+
- NDK r12+

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                   OpenHarmony Device                     │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   UI Layer  │  │  Service    │  │   Cache     │     │
│  │  (ArkUI)    │  │  Manager    │  │   Manager   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
│                           │                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │              OFA Agent Core (C++)                │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────┐ │   │
│  │  │ Connect │ │ Skills  │ │  Local  │ │  P2P  │ │   │
│  │  │ Manager │ │ Manager │ │Scheduler│ │ Mesh  │ │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └───────┘ │   │
│  └─────────────────────────────────────────────────┘   │
│                           │                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │          OpenHarmony Native API                  │   │
│  │  • 分布式软总线  • 数据存储  • 网络管理           │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## 离线能力

OpenHarmony SDK 支持以下离线能力：

| 等级 | 能力 | 说明 |
|------|------|------|
| L1 | 完全离线 | 本地技能执行，无需网络 |
| L2 | 局域网协作 | 通过分布式软总线与其他鸿蒙设备通信 |
| L3 | 弱网同步 | 偶尔连接 Center 同步数据 |
| L4 | 在线模式 | 完整功能，连接 Center |

## 快速开始

### 1. 添加依赖

```json
// oh-package.json5
{
  "dependencies": {
    "@ofa/agent": "0.9.0"
  }
}
```

### 2. 初始化 Agent

```cpp
#include <ofa/agent.h>

// 创建 Agent
OFA_Agent* agent = OFA_Agent_Create(&(OFA_AgentConfig){
    .name = "harmony-device",
    .type = OFA_AGENT_TYPE_EDGE,
    .center_addr = "center.example.com:9090",
    .offline_mode = true,  // 支持离线
    .p2p_enabled = true    // 启用 P2P
});

// 注册本地技能
OFA_Agent_RegisterSkill(agent, &(OFA_Skill){
    .id = "local.sensor",
    .name = "Local Sensor",
    .handler = SensorHandler
});

// 启动 Agent
OFA_Agent_Start(agent);
```

### 3. 离线任务执行

```cpp
// 离线模式下执行本地任务
OFA_Task* task = OFA_Task_Create(&(OFA_TaskSpec){
    .skill_id = "local.sensor",
    .input = sensor_data,
    .input_len = data_len,
    .offline_capable = true  // 标记为可离线执行
});

OFA_Result result = OFA_Agent_ExecuteLocal(agent, task);
```

### 4. P2P 设备发现

```cpp
// 发现附近的 OpenHarmony 设备
OFA_PeerDiscovery_Start(agent, ^(const OFA_Peer* peer, bool added) {
    if (added) {
        printf("发现设备: %s\n", peer->name);
    } else {
        printf("设备离线: %s\n", peer->name);
    }
});

// 向附近设备发送消息
OFA_P2P_Send(agent, peer_id, message, message_len);
```

## API 参考

### 连接管理

| API | 说明 |
|-----|------|
| `OFA_Agent_Create` | 创建 Agent 实例 |
| `OFA_Agent_Start` | 启动 Agent |
| `OFA_Agent_Stop` | 停止 Agent |
| `OFA_Agent_Destroy` | 销毁 Agent |

### 离线模式

| API | 说明 |
|-----|------|
| `OFA_Agent_SetOfflineMode` | 设置离线模式 |
| `OFA_Agent_ExecuteLocal` | 本地执行任务 |
| `OFA_Agent_CacheTask` | 缓存任务待同步 |
| `OFA_Agent_SyncWhenOnline` | 在线时同步 |

### P2P 通信

| API | 说明 |
|-----|------|
| `OFA_PeerDiscovery_Start` | 启动设备发现 |
| `OFA_PeerDiscovery_Stop` | 停止设备发现 |
| `OFA_P2P_Connect` | 连接设备 |
| `OFA_P2P_Send` | 发送消息 |
| `OFA_P2P_Broadcast` | 广播消息 |

## 目录结构

```
src/sdk/openharmony/
├── napi/                    # NAPI 接口层
│   ├── agent_napi.cpp       # Agent NAPI
│   ├── skill_napi.cpp       # Skill NAPI
│   └── p2p_napi.cpp         # P2P NAPI
├── core/                    # 核心实现
│   ├── agent.cpp            # Agent 实现
│   ├── connection.cpp       # 连接管理
│   ├── local_scheduler.cpp  # 本地调度器
│   └── offline_cache.cpp    # 离线缓存
├── p2p/                     # P2P 模块
│   ├── discovery.cpp        # 设备发现
│   ├── mesh.cpp             # 网状网络
│   └── transport.cpp        # 传输层
├── skills/                  # 内置技能
│   ├── sensor.cpp           # 传感器技能
│   └── storage.cpp          # 存储技能
├── include/                 # 头文件
│   └── ofa/
│       ├── agent.h
│       ├── skill.h
│       └── p2p.h
└── BUILD.gn                 # 构建配置
```

## 构建说明

```bash
# 使用 OpenHarmony 构建系统
./build.sh --product-name rk3568 --build-target ofa_agent

# 输出 HAP 包
out/rk3568/innerkits/ofa_agent/
```

## 安全约束

OpenHarmony SDK 内置以下安全约束：

| 约束类型 | 检查项 |
|----------|--------|
| 隐私保护 | 禁止传输用户隐私数据 |
| 财产安全 | 支付操作需要 Center 授权 |
| 本地安全 | 敏感操作需要用户确认 |
| P2P 安全 | 设备间通信加密 |

## 示例场景

### 场景 1: 智能家居控制

```cpp
// 离线控制本地智能家居设备
OFA_Agent_SetOfflineMode(agent, true);

// 通过分布式软总线控制其他鸿蒙设备
OFA_P2P_Broadcast(agent, &(OFA_Message){
    .type = OFA_MSG_DEVICE_CONTROL,
    .target = "light",
    .action = "turn_on"
});
```

### 场景 2: 数据同步

```cpp
// 缓存数据，待在线时同步
OFA_DataCache* cache = OFA_DataCache_Create();

OFA_DataCache_Push(cache, sensor_reading);
OFA_DataCache_Push(cache, event_log);

// 在线时自动同步
OFA_Agent_SyncWhenOnline(agent, cache, ^(bool success) {
    if (success) {
        printf("数据同步完成\n");
    }
});
```

## 许可证

MIT License
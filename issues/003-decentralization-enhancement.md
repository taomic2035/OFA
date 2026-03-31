# Issue: OpenHarmony 支持 + 去中心化增强

**创建时间**: 2026-03-31
**状态**: Open
**优先级**: High
**标签**: enhancement, architecture

---

## 需求描述

### 1. OpenHarmony 平台支持

当前 SDK 平台列表：
- ✅ Android
- ✅ iOS
- ✅ Desktop (Go)
- ✅ Web (TypeScript)
- ✅ Python
- ✅ Rust
- ✅ Node.js
- ✅ C++
- ✅ Lite Agent (手表/手环)
- ✅ IoT (智能家居)
- ❌ **OpenHarmony** (缺失)

OpenHarmony 是华为开源的物联网操作系统，应添加支持。

---

### 2. 去中心化交互设计

**当前问题**：
- Agent 之间的交互需要通过 Center 中转
- 平台依赖 Center 算力，无离线能力

**改进目标**：

#### 2.1 Agent 间直接通信

```
当前架构:
Agent A → Center → Agent B

目标架构:
Agent A ←→ Agent B (P2P，在允许范围内)
         ↓
      Center (仅监管/仲裁)
```

#### 2.2 交互范围约束

Agent 间交互需遵循以下规则：

| 允许 | 禁止 |
|------|------|
| 任务协作 | 隐私数据交换 |
| 状态同步 | 财产相关操作 |
| 技能调用 | 安全敏感操作 |
| 心跳检测 | 未授权的数据访问 |

#### 2.3 离线运行能力

```
┌─────────────────────────────────────────────────────┐
│                    Center (可选)                     │
│              监管 / 仲裁 / 高级服务                    │
└─────────────────────────────────────────────────────┘
                         ↑
                         │ (弱依赖)
                         ↓
┌─────────────────────────────────────────────────────┐
│              Local Agent Mesh (强依赖)               │
│  ┌───────┐     ┌───────┐     ┌───────┐             │
│  │Agent A│←───→│Agent B│←───→│Agent C│             │
│  └───────┘     └───────┘     └───────┘             │
│                                                      │
│  • 本地任务调度                                       │
│  • 本地技能执行                                       │
│  • 本地数据缓存                                       │
│  • P2P 消息传递                                      │
└─────────────────────────────────────────────────────┘
```

---

## 技术方案

### Phase 1: OpenHarmony SDK

基于 OpenHarmony NDK 开发，参考 Lite Agent SDK 设计。

**关键点**：
- 使用 OpenHarmony Native API (C++)
- 支持分布式软总线
- 适配 OpenHarmony 权限模型

### Phase 2: 去中心化通信协议

**新增组件**：

| 组件 | 功能 | 文件 |
|------|------|------|
| PeerManager | P2P 连接管理 | `pkg/p2p/peer.go` |
| ACLManager | 访问控制列表 | `pkg/acl/manager.go` |
| LocalScheduler | 本地调度器 | `pkg/local/scheduler.go` |
| OfflineCache | 离线缓存 | `pkg/cache/offline.go` |
| ConstraintEngine | 约束检查引擎 | `pkg/constraint/engine.go` |

### Phase 3: 离线运行模式

**Agent 离线能力等级**：

| 等级 | 能力 | 示例场景 |
|------|------|----------|
| L1 | 完全离线 | 本地技能执行 |
| L2 | 局域网协作 | 家庭设备互联 |
| L3 | 弱网同步 | 偶尔连接 Center |
| L4 | 在线模式 | 完整功能 |

---

## 安全约束设计

### 约束类型

```go
type ConstraintType string

const (
    ConstraintPrivacy    ConstraintType = "privacy"    // 隐私保护
    ConstraintFinancial  ConstraintType = "financial"  // 财产相关
    ConstraintSecurity   ConstraintType = "security"   // 安全敏感
    ConstraintAuth       ConstraintType = "auth"       // 需要授权
)
```

### 交互规则

```yaml
# interaction_rules.yaml
defaults:
  allow_offline: true
  require_center_for:
    - financial_operations
    - privacy_data_access
    - security_settings

agent_to_agent:
  allowed:
    - task_collaboration
    - skill_invocation
    - status_broadcast
    - heartbeat

  denied:
    - personal_data_transfer
    - payment_operations
    - credential_sharing
    - unauthorized_access
```

---

## 影响评估

| 方面 | 影响 |
|------|------|
| 架构 | 需要增强 P2P 通信模块 |
| 安全 | 需要实现约束检查引擎 |
| SDK | 新增 OpenHarmony SDK |
| 测试 | 需要离线场景测试用例 |
| 文档 | 更新架构设计文档 |

---

## 下一步

1. [ ] 创建 OpenHarmony SDK 骨架
2. [ ] 设计 Agent 间通信协议
3. [ ] 实现约束检查引擎
4. [ ] 添加离线运行模式
5. [ ] 更新架构文档
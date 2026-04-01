# OFA v0.10.0 迭代计划

**目标版本**: 0.10.0
**迭代主题**: 去中心化增强 + 质量提升
**计划周期**: 2026-03-31 ~ 2026-04-15

---

## 迭代概览

```
Phase 1: 基础设施 (测试 + 文档)
    ↓
Phase 2: 核心功能 (去中心化完善)
    ↓
Phase 3: 用户界面 (Dashboard + SDK)
    ↓
Phase 4: 集成验证
```

---

## Phase 1: 基础设施 (Day 1-4)

### 1.1 测试框架完善
- [x] 约束检查引擎单元测试
- [x] 本地调度器单元测试
- [x] P2P 通信测试
- [x] 离线模式集成测试

### 1.2 文档更新
- [ ] 架构设计文档 (更新去中心化架构)
- [ ] API 文档 (新增离线/P2P API)
- [ ] 部署指南 (离线部署场景)
- [ ] SDK 文档 (OpenHarmony)

---

## Phase 2: 核心功能 (Day 5-9)

### 2.1 去中心化完善
- [ ] 交互规则配置文件
- [ ] 离线数据同步机制
- [ ] Agent 间直接通信协议
- [ ] 约束运行时检查

### 2.2 离线能力增强
- [ ] 离线任务队列持久化
- [ ] 断点续传支持
- [ ] 冲突解决机制
- [ ] 网络状态感知

---

## Phase 3: 用户界面 (Day 10-13)

### 3.1 Dashboard 增强
- [ ] WebSocket 实时更新
- [ ] 离线模式状态显示
- [ ] 约束状态面板
- [ ] P2P 网络拓扑图

### 3.2 SDK 完善
- [ ] OpenHarmony NAPI 实现
- [ ] Go Agent 离线模式
- [ ] Python SDK 离线支持
- [ ] SDK 示例代码

---

## Phase 4: 集成验证 (Day 14-15)

### 4.1 端到端测试
- [ ] 完全离线场景测试
- [ ] 局域网协作测试
- [ ] 弱网同步测试
- [ ] 约束违反检测测试

### 4.2 发布准备
- [ ] 版本号更新 (0.9.0 → 0.10.0)
- [ ] CHANGELOG 更新
- [ ] Release Notes 编写

---

## 详细任务列表

### Phase 1.1: 测试框架

| 任务 | 文件 | 状态 |
|------|------|------|
| 约束引擎测试 | `pkg/constraint/engine_test.go` | ✅ 完成 |
| 本地调度器测试 | `pkg/local/scheduler_test.go` | ✅ 完成 |
| P2P 通信测试 | `pkg/messaging/p2p_test.go` | ✅ 完成 |
| 离线模式测试 | `pkg/local/offline_test.go` | ✅ 完成 |

### Phase 1.2: 文档更新

| 任务 | 文件 | 状态 |
|------|------|------|
| 架构文档 | `docs/03-ARCHITECTURE_DESIGN.md` | 待开始 |
| API 文档 | `docs/API.md` | 待更新 |
| 离线部署 | `docs/DEPLOYMENT.md` | 待更新 |
| OpenHarmony SDK | `src/sdk/openharmony/README.md` | 已完成基础 |

### Phase 2.1: 去中心化

| 任务 | 文件 | 状态 |
|------|------|------|
| 交互规则配置 | `configs/interaction_rules.yaml` | 待开始 |
| 同步机制 | `pkg/sync/manager.go` | 待开始 |
| 通信协议 | `pkg/p2p/protocol.go` | 待开始 |
| 运行时检查 | `pkg/constraint/runtime.go` | 待开始 |

### Phase 2.2: 离线能力

| 任务 | 文件 | 状态 |
|------|------|------|
| 任务持久化 | `pkg/local/persistence.go` | 待开始 |
| 断点续传 | `pkg/transfer/resume.go` | 待开始 |
| 冲突解决 | `pkg/sync/conflict.go` | 待开始 |
| 网络感知 | `pkg/network/detector.go` | 待开始 |

### Phase 3.1: Dashboard

| 任务 | 文件 | 状态 |
|------|------|------|
| WebSocket | `src/dashboard/src/api/websocket.ts` | 待开始 |
| 离线状态 | `src/dashboard/src/views/Status.vue` | 待开始 |
| 约束面板 | `src/dashboard/src/views/Constraints.vue` | 待开始 |
| 网络拓扑 | `src/dashboard/src/views/Topology.vue` | 待开始 |

### Phase 3.2: SDK

| 任务 | 文件 | 状态 |
|------|------|------|
| OpenHarmony NAPI | `src/sdk/openharmony/napi/` | 待开始 |
| Go Agent 离线 | `src/agent/go/pkg/offline/` | 待开始 |
| Python SDK 离线 | `src/sdk/python/ofa/offline.py` | 待开始 |
| 示例代码 | `examples/` | 待开始 |

---

## 验收标准

### 功能验收
- [ ] Agent 可完全离线运行本地技能
- [ ] Agent 间可在约束范围内直接通信
- [ ] Dashboard 实时显示系统状态
- [ ] 约束引擎正确阻止违规操作

### 质量验收
- [ ] 测试覆盖率 > 70%
- [ ] 所有测试通过
- [ ] 文档完整更新
- [ ] 代码编译无警告

---

## 当前进度

**Phase**: Phase 1 完成，开始 SDK 补齐
**已完成**: 约束引擎测试 (18 tests), 本地调度器测试 (22 tests), P2P通信测试 (18 tests)
**下一个任务**: SDK 平台补齐 (见 SDK_ROADMAP.md)

---

*创建时间: 2026-03-31*
*最后更新: 2026-04-01*
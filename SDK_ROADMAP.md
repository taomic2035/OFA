# OFA SDK 补齐计划

**版本目标**: v0.10.0
**计划周期**: 2026-04-01 ~ 2026-04-15

---

## 当前 SDK 状态概览

| SDK | 语言 | 文件数 | 离线支持 | P2P支持 | 状态 |
|-----|------|--------|----------|---------|------|
| Go Agent | Go | 4 | ✅ | ✅ | 离线完成 |
| Desktop | Go | 7 | ❌ | ❌ | 基础+系统托盘 |
| Python | Python | 11 | ✅ | ✅ | 离线完成 |
| Node.js | TypeScript | 12 | ✅ | ✅ | 离线完成 |
| Web | TypeScript | 1 | ❌ | ❌ | 最小化 |
| Rust | Rust | 10 | ✅ | ✅ | 离线完成 |
| C++ | C++ | 13 | ✅ | ✅ | 离线完成 |
| iOS | Swift | 10 | ✅ | ✅ | 离线完成 |
| Android | Java | 10 | ✅ | ✅ | 离线完成 |
| Lite (手表) | Go | 3 | ❌ | ❌ | 最小化 |
| IoT | Go | 3 | ❌ | ❌ | MQTT支持 |
| OpenHarmony | C++ | 10 | ✅ | ✅ | 核心完成 |

---

## Phase 1: 核心SDK增强 (Day 1-5)

### 1.1 OpenHarmony SDK 实现
优先级最高，因为目前只有文档没有代码。

| 任务 | 文件 | 状态 |
|------|------|------|
| 头文件定义 | `include/ofa/agent.h` | ✅ 完成 (含所有新增接口) |
| Agent核心实现 | `core/agent.cpp` | ✅ 完成 |
| 连接管理 | `core/connection.cpp` | ✅ 完成 |
| 本地调度器 | `core/local_scheduler.cpp` | ✅ 完成 |
| 离线缓存 | `core/offline_cache.cpp` | ✅ 完成 |
| NAPI接口层 | `napi/agent_napi.cpp` | ✅ 完成 (骨架) |
| 设备发现 | `p2p/discovery.cpp` | ✅ 完成 |
| 内置离线技能 | `skills/builtin.cpp` | ✅ 完成 |
| 构建配置 | `BUILD.gn` | ✅ 完成 |

### 1.2 Go Agent SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线调度器 | `offline.go` | ✅ 完成 |
| P2P客户端 | `p2p.go` | ✅ 完成 |
| 约束检查客户端 | `constraint.go` | ✅ 完成 |

### 1.3 Python SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线模块 | `ofa/offline.py` | ✅ 完成 |
| P2P模块 | `ofa/p2p.py` | ✅ 完成 |
| 约束模块 | `ofa/constraint.py` | ✅ 完成 |
| 示例代码 | `examples/offline_example.py` | ✅ 完成 |

---

## Phase 2: 移动端SDK完善 (Day 6-10)

### 2.1 Android SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线管理器 | `java/.../offline/OfflineManager.java` | ✅ 完成 |
| 本地调度器 | `java/.../offline/LocalScheduler.java` | ✅ 完成 |
| P2P客户端 | `java/.../p2p/P2PClient.java` | ✅ 完成 |
| 约束检查 | `java/.../constraint/ConstraintChecker.java` | ✅ 完成 |
| 离线缓存 | `java/.../offline/OfflineCache.java` | ✅ 完成 |
| 内置离线技能 | `java/.../skill/builtin/OfflineSkills.java` | 待开始 |

### 2.2 iOS SDK 增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 完整Agent实现 | `Sources/OFAAgent/OFAAgent.swift` | ✅ 已有基础 |
| 连接管理 | `Sources/OFAAgent/OFAAgent.swift` | ✅ 已有基础 |
| 离线支持 | `Sources/OFAAgent/Offline/` | ✅ 完成 |
| P2P支持 | `Sources/OFAAgent/P2P/` | ✅ 完成 |
| 约束检查 | `Sources/OFAAgent/Constraint/` | ✅ 完成 |
| 内置技能 | `Sources/OFAAgent/Builtins.swift` | ✅ 完成 |

---

## Phase 3: 其他SDK完善 (Day 11-13)

### 3.1 Node.js SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线模块 | `src/offline.ts` | ✅ 完成 |
| P2P模块 | `src/p2p.ts` | ✅ 完成 |
| 约束模块 | `src/constraint.ts` | ✅ 完成 |

### 3.2 Rust SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线模块 | `src/offline.rs` | ✅ 完成 |
| P2P模块 | `src/p2p.rs` | ✅ 完成 |
| 约束模块 | `src/constraint.rs` | ✅ 完成 |

### 3.3 C++ SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线调度器 | `src/offline.cpp`, `include/ofa/offline.hpp` | ✅ 完成 |
| P2P客户端 | `src/p2p.cpp`, `include/ofa/p2p.hpp` | ✅ 完成 |
| 约束检查 | `src/constraint.cpp`, `include/ofa/constraint.hpp` | ✅ 完成 |

### 3.4 Desktop SDK 离线增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线管理器 | `offline.go` | 待开始 |
| P2P集成 | `p2p.go` | 待开始 |

---

## Phase 4: Lite & IoT SDK (Day 14-15)

### 4.1 Lite Agent SDK (手表/手环)

| 任务 | 文件 | 状态 |
|------|------|------|
| 完整协议实现 | `protocol.go` | 需完善 |
| 离线技能 | `offline_skills.go` | 待开始 |
| 低功耗优化 | `power.go` | 待开始 |

### 4.2 IoT SDK 增强

| 任务 | 文件 | 状态 |
|------|------|------|
| 离线数据缓存 | `cache.go` | 待开始 |
| 边缘计算技能 | `edge_skills.go` | 待开始 |
| 设备发现增强 | `discovery.go` | 待开始 |

---

## 验收标准

### 功能验收
- [ ] 所有SDK支持离线模式 (L1-L4)
- [ ] 所有SDK支持P2P通信
- [ ] 所有SDK支持约束检查
- [ ] OpenHarmony SDK 可编译运行

### 质量验收
- [ ] 每个SDK有示例代码
- [ ] 每个SDK有单元测试
- [ ] 代码编译无警告
- [ ] 文档完整

---

## 实现优先级

1. **OpenHarmony SDK** - 新平台，从零实现
2. **Python SDK** - 用户量大，易于测试
3. **Android SDK** - 移动端主流
4. **iOS SDK** - 移动端主流
5. **Go/Node.js/Rust/C++ SDK** - 并行开发
6. **Desktop/Lite/IoT SDK** - 特殊场景

---

*创建时间: 2026-04-01*
*最后更新: 2026-04-01*
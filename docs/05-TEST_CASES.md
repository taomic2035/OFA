# OFA 测试用例文档

---

# 一、测试策略

## 1.1 测试分层

```
┌─────────────────────────────────────────────────────────┐
│                    E2E测试 (端到端)                      │
│  完整业务流程测试、跨平台集成测试                         │
├─────────────────────────────────────────────────────────┤
│                   集成测试 (集成)                        │
│  模块间交互、API接口、数据库集成                         │
├─────────────────────────────────────────────────────────┤
│                   单元测试 (单元)                        │
│  函数/方法级别、逻辑正确性验证                           │
└─────────────────────────────────────────────────────────┘
```

## 1.2 测试覆盖目标

| 测试类型 | 覆盖率目标 | 说明 |
|----------|------------|------|
| 单元测试 | > 80% | 核心业务逻辑 |
| 集成测试 | > 70% | API和存储层 |
| E2E测试 | 关键路径 | 主要业务场景 |

---

# 二、Center测试用例

## 2.1 Agent管理模块

### TC-CENTER-AGENT-001: Agent注册

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-AGENT-001 |
| **测试名称** | Agent正常注册 |
| **前置条件** | Center服务正常运行 |
| **测试步骤** | 1. 发送注册请求<br>2. 验证响应<br>3. 查询注册表 |
| **测试数据** | `{"name": "test-agent", "type": "mobile", "capabilities": ["skill-1"]}` |
| **预期结果** | 1. 返回Agent ID<br>2. 状态码200<br>3. 注册表中存在该Agent |
| **优先级** | P0 |

### TC-CENTER-AGENT-002: Agent重复注册

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-AGENT-002 |
| **测试名称** | Agent重复注册处理 |
| **前置条件** | Agent已注册 |
| **测试步骤** | 1. 使用相同信息再次注册<br>2. 验证响应 |
| **预期结果** | 返回原有Agent ID，不创建重复记录 |
| **优先级** | P1 |

### TC-CENTER-AGENT-003: Agent心跳

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-AGENT-003 |
| **测试名称** | Agent心跳处理 |
| **前置条件** | Agent已注册 |
| **测试步骤** | 1. 发送心跳请求<br>2. 等待响应<br>3. 检查LastSeen时间更新 |
| **测试数据** | `{"agent_id": "xxx", "status": "online", "resources": {...}}` |
| **预期结果** | 1. 心跳成功<br>2. LastSeen更新为当前时间 |
| **优先级** | P0 |

### TC-CENTER-AGENT-004: Agent超时下线

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-AGENT-004 |
| **测试名称** | Agent心跳超时自动下线 |
| **前置条件** | Agent已注册且在线 |
| **测试步骤** | 1. 停止发送心跳<br>2. 等待超时时间(60s)<br>3. 查询Agent状态 |
| **预期结果** | Agent状态变为offline |
| **优先级** | P0 |

### TC-CENTER-AGENT-005: Agent注销

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-AGENT-005 |
| **测试名称** | Agent主动注销 |
| **前置条件** | Agent已注册 |
| **测试步骤** | 1. 发送注销请求<br>2. 查询注册表 |
| **预期结果** | 1. 注销成功<br>2. 注册表中移除该Agent |
| **优先级** | P1 |

## 2.2 任务调度模块

### TC-CENTER-SCHED-001: 任务提交

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-001 |
| **测试名称** | 任务正常提交 |
| **前置条件** | 有在线Agent具备所需技能 |
| **测试步骤** | 1. 提交任务<br>2. 验证任务ID返回<br>3. 验证任务状态为pending |
| **测试数据** | `{"skill_id": "text.process", "input": {"text": "hello", "operation": "uppercase"}}` |
| **预期结果** | 1. 返回task_id<br>2. 任务入队成功 |
| **优先级** | P0 |

### TC-CENTER-SCHED-002: 任务调度-能力优先

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-002 |
| **测试名称** | 能力优先调度策略 |
| **前置条件** | 多个Agent在线，能力不同 |
| **测试步骤** | 1. 设置调度策略为能力优先<br>2. 提交需要特定技能的任务<br>3. 验证分配结果 |
| **预期结果** | 任务分配给具备所需技能的Agent |
| **优先级** | P0 |

### TC-CENTER-SCHED-003: 任务调度-负载均衡

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-003 |
| **测试名称** | 负载均衡调度策略 |
| **前置条件** | 多个Agent在线，负载不同 |
| **测试步骤** | 1. 设置调度策略为负载均衡<br>2. 提交多个任务<br>3. 验证分配分布 |
| **预期结果** | 任务均匀分配给各Agent |
| **优先级** | P1 |

### TC-CENTER-SCHED-004: 任务执行结果

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-004 |
| **测试名称** | 任务执行成功返回结果 |
| **前置条件** | 任务已分配并执行完成 |
| **测试步骤** | 1. 查询任务状态<br>2. 获取执行结果 |
| **预期结果** | 1. 状态为completed<br>2. 返回正确结果 |
| **优先级** | P0 |

### TC-CENTER-SCHED-005: 任务超时处理

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-005 |
| **测试名称** | 任务执行超时 |
| **前置条件** | 任务已提交 |
| **测试步骤** | 1. 提交长时间任务<br>2. 设置短超时<br>3. 等待超时 |
| **预期结果** | 任务状态变为timeout，可重试 |
| **优先级** | P1 |

### TC-CENTER-SCHED-006: 任务取消

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-SCHED-006 |
| **测试名称** | 取消运行中的任务 |
| **前置条件** | 任务正在执行 |
| **测试步骤** | 1. 发送取消请求<br>2. 验证状态变更 |
| **预期结果** | 任务状态变为cancelled |
| **优先级** | P1 |

## 2.3 消息路由模块

### TC-CENTER-MSG-001: 点对点消息

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-MSG-001 |
| **测试名称** | 发送点对点消息 |
| **前置条件** | 发送方和接收方都在线 |
| **测试步骤** | 1. AgentA发送消息给AgentB<br>2. 验证消息到达 |
| **测试数据** | `{"to": "agent-b", "action": "ping", "payload": {}}` |
| **预期结果** | AgentB收到消息，ACK返回成功 |
| **优先级** | P0 |

### TC-CENTER-MSG-002: 广播消息

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-MSG-002 |
| **测试名称** | 广播消息给所有Agent |
| **前置条件** | 多个Agent在线 |
| **测试步骤** | 1. 发送广播消息<br>2. 验证所有Agent收到 |
| **预期结果** | 所有在线Agent都收到消息 |
| **优先级** | P1 |

### TC-CENTER-MSG-003: 组播消息

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-MSG-003 |
| **测试名称** | 发送组播消息 |
| **前置条件** | 存在Agent组 |
| **测试步骤** | 1. 创建Agent组<br>2. 发送组播消息<br>3. 验证组成员收到 |
| **预期结果** | 只有组成员收到消息 |
| **优先级** | P2 |

### TC-CENTER-MSG-004: 离线消息

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-CENTER-MSG-004 |
| **测试名称** | 离线消息存储和投递 |
| **前置条件** | 目标Agent离线 |
| **测试步骤** | 1. 发送消息给离线Agent<br>2. Agent上线<br>3. 验证消息投递 |
| **预期结果** | Agent上线后收到离线消息 |
| **优先级** | P1 |

---

# 三、Agent测试用例

## 3.1 连接模块

### TC-AGENT-CONN-001: 连接Center

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-CONN-001 |
| **测试名称** | Agent正常连接Center |
| **前置条件** | Center服务正常 |
| **测试步骤** | 1. 启动Agent<br>2. 自动连接Center<br>3. 验证连接状态 |
| **预期结果** | 连接成功，状态为online |
| **优先级** | P0 |

### TC-AGENT-CONN-002: 断线重连

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-CONN-002 |
| **测试名称** | 网络断开后自动重连 |
| **前置条件** | Agent已连接 |
| **测试步骤** | 1. 断开网络<br>2. 等待重连<br>3. 恢复网络<br>4. 验证重连成功 |
| **预期结果** | Agent自动重连，状态恢复 |
| **优先级** | P0 |

### TC-AGENT-CONN-003: 心跳维持

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-CONN-003 |
| **测试名称** | 心跳正常发送 |
| **前置条件** | Agent已连接 |
| **测试步骤** | 1. 监控心跳请求<br>2. 验证发送频率 |
| **预期结果** | 每30秒发送一次心跳 |
| **优先级** | P0 |

## 3.2 任务执行模块

### TC-AGENT-EXEC-001: 接收任务

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-EXEC-001 |
| **测试名称** | 正常接收并执行任务 |
| **前置条件** | Agent已连接，具备所需技能 |
| **测试步骤** | 1. Center分配任务<br>2. Agent接收任务<br>3. 执行任务<br>4. 返回结果 |
| **预期结果** | 任务执行成功，返回正确结果 |
| **优先级** | P0 |

### TC-AGENT-EXEC-002: 并发任务执行

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-EXEC-002 |
| **测试名称** | 多任务并发执行 |
| **前置条件** | Agent已连接 |
| **测试步骤** | 1. 同时分配多个任务<br>2. 验证并发执行 |
| **预期结果** | 所有任务正确执行，不互相阻塞 |
| **优先级** | P1 |

### TC-AGENT-EXEC-003: 任务超时处理

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-EXEC-003 |
| **测试名称** | 执行超时任务终止 |
| **前置条件** | 任务设置了超时时间 |
| **测试步骤** | 1. 执行长时间任务<br>2. 超时后终止 |
| **预期结果** | 任务在超时后被终止，返回timeout状态 |
| **优先级** | P1 |

## 3.3 技能模块

### TC-AGENT-SKILL-001: 内置技能加载

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-SKILL-001 |
| **测试名称** | 加载内置技能 |
| **前置条件** | Agent启动 |
| **测试步骤** | 1. 启动Agent<br>2. 查询已加载技能列表 |
| **预期结果** | 所有内置技能正确加载 |
| **优先级** | P0 |

### TC-AGENT-SKILL-002: 技能动态安装

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-SKILL-002 |
| **测试名称** | 动态安装新技能 |
| **前置条件** | Agent已连接 |
| **测试步骤** | 1. 发送技能安装请求<br>2. 验证安装成功<br>3. 使用新技能 |
| **预期结果** | 技能安装成功，可正常使用 |
| **优先级** | P1 |

### TC-AGENT-SKILL-003: 技能输入验证

| 项目 | 内容 |
|------|------|
| **测试ID** | TC-AGENT-SKILL-003 |
| **测试名称** | 技能输入参数验证 |
| **前置条件** | 技能已加载 |
| **测试步骤** | 1. 提供非法输入<br>2. 验证错误处理 |
| **预期结果** | 返回明确的错误信息 |
| **优先级** | P0 |

---

# 四、集成测试用例

## 4.1 端到端场景测试

### TC-E2E-001: 完整任务流程

```
测试流程:
1. 启动Center
2. 启动Agent A (具备skill-1)
3. Agent A注册到Center
4. 外部服务提交任务(skill-1)
5. Center调度任务给Agent A
6. Agent A执行任务
7. Agent A返回结果
8. 外部服务获取结果

验证点:
- Agent注册成功
- 任务入队成功
- 任务正确调度
- 任务正确执行
- 结果正确返回
```

### TC-E2E-002: 跨Agent通信

```
测试流程:
1. 启动Agent A和Agent B
2. Agent A发送消息给Agent B
3. Agent B接收并处理消息
4. Agent B回复Agent A

验证点:
- 消息正确路由
- 消息不丢失
- 消息顺序正确
```

### TC-E2E-003: 高并发场景

```
测试流程:
1. 启动10个Agent
2. 同时提交100个任务
3. 验证所有任务正确执行

验证点:
- 所有任务被处理
- 无任务丢失
- 调度均衡
- 无死锁
```

---

# 五、性能测试用例

## 5.1 压力测试

### TC-PERF-001: 高并发任务提交

| 参数 | 值 |
|------|-----|
| 并发用户数 | 1000 |
| 每用户任务数 | 10 |
| 总任务数 | 10000 |
| 目标吞吐量 | > 1000 TPS |
| 平均响应时间 | < 100ms |

### TC-PERF-002: 大量Agent在线

| 参数 | 值 |
|------|-----|
| 在线Agent数 | 10000 |
| 心跳频率 | 30s |
| 消息吞吐 | > 10000条/秒 |
| 目标延迟 | < 50ms |

## 5.2 稳定性测试

### TC-STAB-001: 长时间运行

| 参数 | 值 |
|------|-----|
| 运行时长 | 72小时 |
| 任务速率 | 100 TPS |
| Agent数 | 100 |

验证点：
- 无内存泄漏
- 无连接泄漏
- CPU使用稳定
- 错误率 < 0.01%

---

# 六、安全测试用例

### TC-SEC-001: 未认证访问

**步骤**: 无Token访问API
**预期**: 返回401 Unauthorized

### TC-SEC-002: Token过期

**步骤**: 使用过期Token
**预期**: 返回401，提示Token过期

### TC-SEC-003: 权限不足

**步骤**: 普通Agent访问管理接口
**预期**: 返回403 Forbidden

### TC-SEC-004: 消息伪造

**步骤**: 伪造其他Agent发送消息
**预期**: 消息被拒绝

### TC-SEC-005: 重放攻击

**步骤**: 重放已发送的消息
**预期**: 消息被拒绝（基于timestamp和nonce）

---

# 七、测试环境

## 7.1 测试环境配置

| 环境 | 用途 | 配置 |
|------|------|------|
| 开发环境 | 日常开发测试 | 单机部署 |
| 测试环境 | 功能测试、集成测试 | 完整部署 |
| 预发环境 | 性能测试、UAT | 生产配置 |
| 生产环境 | 线上运行 | 高可用部署 |

## 7.2 测试数据准备

```sql
-- 测试用户
INSERT INTO agents (id, name, type, status)
VALUES
('test-agent-1', 'Test Agent 1', 'mobile', 'online'),
('test-agent-2', 'Test Agent 2', 'desktop', 'online'),
('test-agent-3', 'Test Agent 3', 'iot', 'offline');

-- 测试技能
INSERT INTO skills (id, name, version, category)
VALUES
('skill-test-1', 'Test Skill 1', '1.0.0', 'test');
```

---

# 八、测试自动化

## 8.1 单元测试示例

```go
// scheduler_test.go
package scheduler

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestScheduler_Submit(t *testing.T) {
    // Setup
    scheduler := NewScheduler(nil, &HybridPolicy{})
    task := &Task{
        SkillID: "skill-1",
        Input:   map[string]interface{}{"text": "hello"},
    }

    // Execute
    err := scheduler.Submit(context.Background(), task)

    // Verify
    assert.NoError(t, err)
    assert.NotEmpty(t, task.ID)
    assert.Equal(t, TaskStatusPending, task.Status)
}
```

## 8.2 集成测试示例

```go
// integration_test.go
//go:build integration

package integration

import (
    "testing"
    "github.com/ofa/sdk-go"
)

func TestTaskExecution(t *testing.T) {
    // Setup
    client := ofa.NewClient("localhost:9090")
    defer client.Close()

    // Execute
    result, err := client.SubmitTask(context.Background(), &ofa.TaskRequest{
        SkillID: "text.process",
        Input: map[string]interface{}{
            "text":      "hello",
            "operation": "uppercase",
        },
    })

    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "HELLO", result.Output["result"])
}
```

## 8.3 E2E测试示例

```kotlin
// E2ETest.kt
package com.ofa.test

import kotlinx.coroutines.runBlocking
import org.junit.Test
import kotlin.test.assertEquals

class E2ETest {

    @Test
    fun `test full task execution flow`() = runBlocking {
        // 1. Start Center (using test container)
        val center = TestCenter.start()

        // 2. Start Agent
        val agent = TestAgent.start(center.address)

        // 3. Submit task
        val client = OFAClient.create(center.address)
        val result = client.submitTask(
            skillId = "text.process",
            input = mapOf("text" to "hello", "operation" to "uppercase")
        ).await()

        // 4. Verify
        assertEquals("HELLO", result.output["result"])

        // 5. Cleanup
        agent.stop()
        center.stop()
    }
}
```

---

# 九、测试报告模板

## 测试执行报告

| 项目 | 内容 |
|------|------|
| 测试版本 | 0.9.0 Beta |
| 测试环境 | 测试环境 |
| 测试时间 | 2026-XX-XX ~ 2026-XX-XX |
| 测试人员 | XXX |

## 测试结果汇总

| 类型 | 总数 | 通过 | 失败 | 阻塞 | 通过率 |
|------|------|------|------|------|--------|
| 单元测试 | - | - | - | - | - |
| 集成测试 | - | - | - | - | - |
| E2E测试 | - | - | - | - | - |

## 缺陷统计

| 级别 | 数量 | 已修复 | 待修复 |
|------|------|--------|--------|
| 致命 | - | - | - |
| 严重 | - | - | - |
| 一般 | - | - | - |
| 轻微 | - | - | - |
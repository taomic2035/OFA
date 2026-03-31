# OFA API 测试脚本使用说明

## 脚本列表

### test_api.ps1 - API功能测试
```
D:\vibecoding\OFA\scripts\test_api.ps1
```

### test_metrics.ps1 - Prometheus指标测试
```
D:\vibecoding\OFA\scripts\test_metrics.ps1
```

## 使用方法

### 前提条件
1. Center服务已启动: `.\build\center.exe`
2. Agent客户端已启动: `.\build\agent.exe`

### 运行测试

```powershell
# API功能测试
powershell -File D:\vibecoding\OFA\scripts\test_api.ps1

# Prometheus指标测试
powershell -File D:\vibecoding\OFA\scripts\test_metrics.ps1
```

### 测试内容

#### test_api.ps1
1. **健康检查** - 检查Center服务状态
2. **系统信息** - 获取在线Agent数和任务数
3. **Agent列表** - 显示所有已注册的Agent
4. **技能列表** - 显示所有可用技能
5. **文本处理任务** - 测试text.process技能
6. **计算器任务** - 测试calculator技能

#### test_metrics.ps1
1. **健康检查** - 验证服务状态
2. **Prometheus指标** - 检查/metrics端点
3. **Agent指标** - ofa_agents_total, ofa_agents_online等
4. **Task指标** - ofa_tasks_total, ofa_task_duration_seconds等
5. **系统指标** - ofa_request_duration_seconds等
6. **Go运行时指标** - go_goroutines等

### 预期输出

```
=== OFA API 测试 ===

1. 健康检查
状态: healthy
版本: v0.9.0

2. 系统信息
在线Agent数: 1
待处理任务: 0

3. Agent列表
总数: 1
  - abc123: 状态=1

4. 技能列表
可用技能:
  - text.process: Text Process (text)
  - json.process: JSON Process (data)
  - calculator: Calculator (math)
  - echo: Echo (utility)

5. 提交任务测试
任务已提交: task-xxx
任务状态: 3
输出: {"result":"HELLO WORLD"}

6. 计算器任务测试
计算器任务已提交: task-yyy
任务状态: 3
输出: {"result":35,"operation":"add"}

=== 测试完成 ===
```
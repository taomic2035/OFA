# OFA 操作指导文档

## 目录

1. [环境准备](#1-环境准备)
2. [启动服务](#2-启动服务)
3. [REST API使用](#3-rest-api使用)
4. [gRPC API使用](#4-grpc-api使用)
5. [技能开发](#5-技能开发)
6. [故障排查](#6-故障排查)
7. [配置说明](#7-配置说明)

---

## 1. 环境准备

### 1.1 系统要求

- 操作系统: Windows 10+
- Go版本: 1.22+ (已安装到 `D:\Go\go\`)
- 内存: 最低4GB

### 1.2 环境变量配置

```powershell
# 在PowerShell中设置
$env:GOPROXY = "https://goproxy.cn,direct"
$env:PATH = "D:\Go\go\bin;" + $env:PATH

# 或在CMD中设置
set GOPROXY=https://goproxy.cn,direct
set PATH=D:\Go\go\bin;%PATH%
```

### 1.3 验证安装

```powershell
D:\Go\go\bin\go.exe version
# 输出: go version go1.22.0 windows/amd64
```

---

## 2. 启动服务

### 2.1 启动Center服务

**方式1: 直接运行编译好的程序**

```powershell
# 进入项目目录
cd D:\vibecoding\OFA

# 启动Center
.\build\center.exe
```

**方式2: 使用Go运行**

```powershell
cd D:\vibecoding\OFA\src\center
D:\Go\go\bin\go.exe run ./cmd/center
```

**启动成功输出:**

```
OFA Center started - gRPC: :9090, REST: :8080
```

### 2.2 启动Agent客户端

**新开一个终端窗口:**

```powershell
cd D:\vibecoding\OFA
.\build\agent.exe
```

**或使用Go运行:**

```powershell
cd D:\vibecoding\OFA\src\agent\go
D:\Go\go\bin\go.exe run ./cmd/agent
```

**启动成功输出:**

```
Agent started: <agent-id>
```

### 2.3 验证服务

```powershell
# 检查Center健康状态
Invoke-RestMethod -Uri "http://localhost:8080/health"

# 预期输出
# status    version
# -------   -------
# healthy   v0.9.0
```

---

## 3. REST API使用

### 3.1 API端点列表

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/v1/agents` | GET | 获取Agent列表 |
| `/api/v1/agents/{id}` | GET | 获取单个Agent |
| `/api/v1/tasks` | POST | 提交任务 |
| `/api/v1/tasks` | GET | 获取任务列表 |
| `/api/v1/tasks/{id}` | GET | 获取任务状态 |
| `/api/v1/tasks/{id}/cancel` | POST | 取消任务 |
| `/api/v1/messages` | POST | 发送消息 |
| `/api/v1/skills` | GET | 获取技能列表 |
| `/api/v1/system/info` | GET | 系统信息 |

### 3.2 PowerShell 示例

#### 健康检查

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

#### 获取Agent列表

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/agents"
```

#### 提交文本处理任务

```powershell
$body = @{
    skill_id = "text.process"
    input = [System.Text.Encoding]::UTF8.GetBytes('{"text":"hello world","operation":"uppercase"}')
    priority = 0
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body

Write-Host "Task ID: $($response.task_id)"
```

#### 查询任务状态

```powershell
$taskId = "your-task-id"
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks/$taskId"
```

#### 获取系统信息

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/system/info"
```

### 3.3 curl 示例 (Git Bash)

```bash
# 健康检查
curl http://localhost:8080/health

# 获取Agent列表
curl http://localhost:8080/api/v1/agents

# 提交任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"skill_id":"text.process","input":"eyJ0ZXh0IjoiaGVsbG8iLCJvcGVyYXRpb24iOiJ1cHBlcmNhc2UifQ=="}'

# 获取任务
curl http://localhost:8080/api/v1/tasks/{task_id}
```

---

## 4. gRPC API使用

### 4.1 使用grpcurl测试

```bash
# 列出服务
grpcurl -plaintext localhost:9090 list

# 列出Agent服务方法
grpcurl -plaintext localhost:9090 list ofa.AgentService

# 调用GetCapabilities
grpcurl -plaintext localhost:9090 ofa.AgentService/GetCapabilities

# 调用ListAgents
grpcurl -plaintext localhost:9090 ofa.ManagementService/ListAgents
```

### 4.2 Go客户端示例

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/ofa/center/proto"
)

func main() {
    // 连接Center
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 创建客户端
    client := pb.NewManagementServiceClient(conn)

    // 获取Agent列表
    resp, err := client.ListAgents(context.Background(), &pb.ListAgentsRequest{})
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Agents: %d", resp.Total)
}
```

---

## 5. 技能开发

### 5.1 创建新技能

**步骤1: 创建技能文件**

在 `src/agent/go/internal/executor/` 目录创建新文件:

```go
// my_skill.go
package executor

import (
    "context"
    "encoding/json"
)

// MySkill 自定义技能
type MySkill struct{}

func (s *MySkill) ID() string       { return "my.skill" }
func (s *MySkill) Name() string     { return "My Skill" }
func (s *MySkill) Version() string  { return "1.0.0" }
func (s *MySkill) Category() string { return "custom" }

func (s *MySkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
    // 解析输入
    var req struct {
        Message string `json:"message"`
    }
    if err := json.Unmarshal(input, &req); err != nil {
        return nil, err
    }

    // 处理逻辑
    result := map[string]interface{}{
        "result":  "processed: " + req.Message,
        "success": true,
    }

    return json.Marshal(result)
}
```

**步骤2: 注册技能**

在 `cmd/agent/main.go` 中注册:

```go
exec := executor.NewExecutor()
exec.RegisterSkill(&executor.TextProcessSkill{})
exec.RegisterSkill(&executor.JSONProcessSkill{})
exec.RegisterSkill(&executor.MySkill{})  // 添加这行
```

**步骤3: 重新编译**

```powershell
cd D:\vibecoding\OFA\src\agent\go
D:\Go\go\bin\go.exe build -o ..\..\..\build\agent.exe ./cmd/agent
```

### 5.2 技能规范

| 方法 | 必须实现 | 说明 |
|------|----------|------|
| ID() | 是 | 唯一标识符 |
| Name() | 是 | 显示名称 |
| Version() | 否 | 版本号，默认1.0.0 |
| Category() | 否 | 分类，默认general |
| Execute() | 是 | 执行逻辑 |

---

## 6. 故障排查

### 6.1 端口被占用

**问题:** 启动Center时提示端口被占用

**解决:**

```powershell
# 查看端口占用
netstat -ano | findstr :9090
netstat -ano | findstr :8080

# 结束占用进程 (PID从上面命令获取)
taskkill /PID <pid> /F
```

### 6.2 Go模块下载失败

**问题:** go mod download 超时

**解决:**

```powershell
# 使用国内代理
$env:GOPROXY = "https://goproxy.cn,direct"
```

### 6.3 Agent无法连接Center

**问题:** Agent启动后无法连接

**检查项:**

1. Center是否已启动
2. 端口9090是否可访问
3. 防火墙是否允许

```powershell
# 测试端口连通性
Test-NetConnection -ComputerName localhost -Port 9090
```

### 6.4 任务执行失败

**问题:** 任务状态为FAILED

**排查步骤:**

1. 检查技能ID是否正确
2. 检查输入格式是否匹配技能要求
3. 查看Agent日志

---

## 7. 配置说明

### 7.1 Center配置文件

位置: `configs/center.yaml`

```yaml
server:
  name: "ofa-center"
  version: "0.9.0"

grpc:
  address: ":9090"

rest:
  address: ":8080"

agent:
  heartbeat_interval_ms: 30000
  heartbeat_timeout_ms: 60000

scheduler:
  default_strategy: "hybrid"
  max_concurrent: 1000
```

### 7.2 Agent环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| CENTER_ADDR | localhost:9090 | Center地址 |
| AGENT_NAME | test-agent | Agent名称 |

**使用示例:**

```powershell
# 设置环境变量后启动
$env:CENTER_ADDR = "192.168.1.100:9090"
$env:AGENT_NAME = "my-agent"
.\build\agent.exe
```

---

## 附录

### A. 常用命令速查

```powershell
# 启动Center
.\build\center.exe

# 启动Agent
.\build\agent.exe

# 健康检查
Invoke-RestMethod -Uri "http://localhost:8080/health"

# 获取Agent列表
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/agents"

# 获取系统信息
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/system/info"

# 运行测试
cd src\center; D:\Go\go\bin\go.exe test ./... -v
```

### B. 项目文件位置

| 文件 | 路径 |
|------|------|
| Center可执行文件 | `D:\vibecoding\OFA\build\center.exe` |
| Agent可执行文件 | `D:\vibecoding\OFA\build\agent.exe` |
| Center配置 | `D:\vibecoding\OFA\configs\center.yaml` |
| Go安装位置 | `D:\Go\go\` |

### C. 技术支持

- 项目目录: `D:\vibecoding\OFA\`
- 文档目录: `D:\vibecoding\OFA\docs\`
- 问题反馈: 项目Issues
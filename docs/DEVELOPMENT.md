# OFA 开发指南

## 开发环境设置

### Go环境

```powershell
# 设置Go代理
$env:GOPROXY = "https://goproxy.cn,direct"

# 添加Go到PATH
$env:PATH = "D:\Go\go\bin;" + $env:PATH

# 验证
go version
```

### 项目结构

```
D:\vibecoding\OFA\
├── build/              # 编译产物
│   ├── center.exe
│   └── agent.exe
├── src/
│   ├── center/         # Center服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── pkg/
│   └── agent/go/       # Agent客户端
├── docs/               # 文档
├── configs/            # 配置
└── skills/             # 技能库
```

## 常用命令

### 构建

```powershell
# 构建Center
cd D:\vibecoding\OFA\src\center
go build -o ..\..\build\center.exe ./cmd/center

# 构建Agent
cd D:\vibecoding\OFA\src\agent\go
go build -o ..\..\..\build\agent.exe ./cmd/agent
```

### 运行

```powershell
# 运行Center
.\build\center.exe

# 运行Agent
.\build\agent.exe

# 使用Go run
cd src\center
go run ./cmd/center
```

### 测试

```powershell
# Center测试
cd src\center
go test ./... -v

# Agent测试
cd src\agent\go
go test ./... -v

# 运行特定测试
go test -run TestCapabilityFirstPolicy -v
```

### 依赖管理

```powershell
# 更新依赖
go mod tidy

# 下载依赖
go mod download

# 查看依赖
go list -m all
```

## 添加新技能

### 1. 创建技能文件

文件: `src/agent/go/internal/executor/my_skill.go`

```go
package executor

import (
    "context"
    "encoding/json"
)

type MySkill struct{}

func (s *MySkill) ID() string       { return "my.skill" }
func (s *MySkill) Name() string     { return "My Skill" }
func (s *MySkill) Version() string  { return "1.0.0" }
func (s *MySkill) Category() string { return "custom" }

func (s *MySkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
    var req struct {
        Data string `json:"data"`
    }
    json.Unmarshal(input, &req)

    result := map[string]interface{}{
        "result": req.Data,
    }
    return json.Marshal(result)
}
```

### 2. 注册技能

文件: `src/agent/go/cmd/agent/main.go`

```go
exec := executor.NewExecutor()
exec.RegisterSkill(&executor.MySkill{})
```

### 3. 重新编译

```powershell
cd src\agent\go
go build -o ..\..\..\build\agent.exe ./cmd/agent
```

## 添加新调度策略

文件: `src/center/internal/scheduler/scheduler.go`

```go
type MyPolicy struct{}

func (p *MyPolicy) Select(task *models.Task, agents []*AgentInfo) string {
    // 实现调度逻辑
    return agents[0].ID
}

// 在 getPolicy 函数中注册
func getPolicy(strategy string) Policy {
    switch strategy {
    case "my_policy":
        return &MyPolicy{}
    // ...
    }
}
```

## 添加REST API端点

文件: `src/center/pkg/rest/server.go`

```go
// 在 setupRoutes 中添加
s.router.HandleFunc("/api/v1/my-endpoint", s.myHandler).Methods("GET")

// 添加处理函数
func (s *Server) myHandler(w http.ResponseWriter, r *http.Request) {
    // 处理逻辑
    jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

## 调试技巧

### 日志输出

```go
import "log"

log.Printf("Debug: %+v", data)
```

### pprof性能分析

```go
import _ "net/http/pprof"

// 在main.go中添加
go func() {
    http.ListenAndServe(":6060", nil)
}()
```

然后访问:
- http://localhost:6060/debug/pprof/

### 使用delve调试

```powershell
# 安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行
dlv debug ./cmd/center
```

## 代码规范

### 命名约定

- 包名: 小写，简短
- 导出函数: 大写开头
- 私有函数: 小写开头
- 接口: 以`er`结尾

### 错误处理

```go
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### 注释

```go
// FunctionName does something.
// It returns X on success.
func FunctionName() error {
    // ...
}
```

## Git工作流

### 提交规范

```
feat: add new feature
fix: fix bug
docs: update documentation
test: add tests
refactor: code refactoring
```

### 分支策略

- `main`: 主分支
- `feature/*`: 功能分支
- `fix/*`: 修复分支

## 常见问题

### Q: go mod download 失败?

```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
```

### Q: 端口被占用?

```powershell
netstat -ano | findstr :9090
taskkill /PID <pid> /F
```

### Q: 编译缓存问题?

```powershell
go clean -cache
go build ./...
```
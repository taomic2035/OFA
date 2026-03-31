# OFA Go SDK

提供Go语言的OFA Agent SDK，用于快速开发和集成Agent。

## 安装

```bash
go get github.com/ofa/sdk-go
```

## 快速开始

### 创建一个简单的Agent

```go
package main

import (
    "context"
    "encoding/json"
    "log"

    ofa "github.com/ofa/sdk-go"
)

func main() {
    // 创建Agent
    agent, err := ofa.NewAgent(&ofa.Config{
        CenterAddr: "localhost:9090",
        Name:       "my-agent",
        Type:       ofa.AgentTypeFull,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 注册技能
    agent.RegisterSkill(&ofa.Skill{
        ID:       "my.skill",
        Name:     "My Skill",
        Version:  "1.0.0",
        Category: "custom",
        Handler: func(ctx context.Context, input []byte) ([]byte, error) {
            // 处理任务
            var req map[string]interface{}
            json.Unmarshal(input, &req)
            // ... 业务逻辑
            return json.Marshal(map[string]string{"result": "ok"})
        },
    })

    // 连接到Center
    if err := agent.Connect(context.Background()); err != nil {
        log.Fatal(err)
    }

    log.Printf("Agent started: %s", agent.ID())

    // 等待退出
    select {}
}
```

## API参考

### Config

Agent配置。

| 字段 | 类型 | 描述 |
|------|------|------|
| CenterAddr | string | Center服务地址 |
| Name | string | Agent名称 |
| Type | AgentType | Agent类型 (Full, Mobile, Lite, IoT, Edge) |
| DeviceInfo | *DeviceInfo | 设备信息 |
| Metadata | map[string]string | 元数据 |

### Agent

Agent客户端。

#### 方法

- `Connect(ctx context.Context) error` - 连接到Center
- `Disconnect()` - 断开连接
- `ID() string` - 获取Agent ID
- `RegisterSkill(skill *Skill)` - 注册技能
- `SendMessage(ctx context.Context, to, action string, payload interface{}) error` - 发送消息

### Skill

技能定义。

| 字段 | 类型 | 描述 |
|------|------|------|
| ID | string | 技能ID |
| Name | string | 技能名称 |
| Version | string | 版本号 |
| Category | string | 分类 |
| Handler | SkillHandler | 处理函数 |

### AgentType

Agent类型枚举。

```go
const (
    AgentTypeUnknown AgentType = iota
    AgentTypeFull    // 完整Agent (PC/服务器)
    AgentTypeMobile  // 移动Agent (手机/平板)
    AgentTypeLite    // 轻量Agent (手表)
    AgentTypeIoT     // IoT Agent
    AgentTypeEdge    // 边缘Agent
)
```

## 示例

### 自定义技能

```go
type CalculatorSkill struct{}

func (s *CalculatorSkill) ID() string       { return "calculator" }
func (s *CalculatorSkill) Name() string     { return "Calculator" }
func (s *CalculatorSkill) Version() string  { return "1.0.0" }
func (s *CalculatorSkill) Category() string { return "math" }

func (s *CalculatorSkill) Execute(ctx context.Context, input []byte) ([]byte, error) {
    var req struct {
        Operation string  `json:"operation"`
        A         float64 `json:"a"`
        B         float64 `json:"b"`
    }
    if err := json.Unmarshal(input, &req); err != nil {
        return nil, err
    }

    var result float64
    switch req.Operation {
    case "add":
        result = req.A + req.B
    case "sub":
        result = req.A - req.B
    case "mul":
        result = req.A * req.B
    case "div":
        result = req.A / req.B
    default:
        return nil, fmt.Errorf("unknown operation: %s", req.Operation)
    }

    return json.Marshal(map[string]float64{"result": result})
}
```

### 消息处理

```go
// 设置消息处理器
agent.SetMessageHandler(func(ctx context.Context, msg *ofa.Message) error {
    log.Printf("Received message: %s from %s", msg.Action, msg.FromAgent)

    switch msg.Action {
    case "ping":
        // 回复pong
        agent.SendMessage(ctx, msg.FromAgent, "pong", map[string]string{
            "reply_to": msg.MsgID,
        })
    }

    return nil
})
```

### 资源监控

```go
// 定期更新资源使用情况
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        agent.UpdateResources(&ofa.ResourceUsage{
            CPUUsage:    getCPUUsage(),
            MemoryUsage: getMemoryUsage(),
            BatteryLevel: getBatteryLevel(),
            NetworkType: getNetworkType(),
        })
    }
}()
```
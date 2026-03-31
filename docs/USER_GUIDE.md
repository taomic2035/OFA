# OFA v0.9 用户指南 (Beta)

## 目录

1. [快速开始](#快速开始)
2. [LLM集成使用](#llm集成使用)
3. [代码生成使用](#代码生成使用)
4. [Agent协作使用](#agent协作使用)
5. [去中心化功能使用](#去中心化功能使用)
6. [典型场景示例](#典型场景示例)

> **设备接入指南**: 移动设备(Android/iOS)、IoT智能家居设备、可穿戴设备接入请参考 [设备接入指南](DEVICE_GUIDE.md)

---

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/ofa/ofa.git
cd ofa

# 构建 Center 服务
cd src/center
go build -o ../../build/center ./cmd/center

# 构建 Agent 客户端
cd ../agent/go
go build -o ../../../build/agent ./cmd/agent
```

### 启动服务

```bash
# 启动 Center 服务
./build/center --config config.yaml

# 启动 Agent 客户端
./build/agent --center localhost:9090 --id agent-001
```

### 配置文件示例

```yaml
# config.yaml
server:
  grpc_port: 9090
  http_port: 8080

database:
  type: sqlite
  path: ./data/ofa.db

llm:
  default_provider: openai
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4
    claude:
      api_key: ${ANTHROPIC_API_KEY}
      model: claude-sonnet-4-6

decentralized:
  enabled: true
  network_type: hybrid
  replica_level: 3
```

---

## LLM集成使用

### 1. 基本对话

```go
package main

import (
    "context"
    "fmt"
    "ofa/pkg/llm"
)

func main() {
    // 创建 LLM 管理器
    manager := llm.NewManager()

    // 注册 OpenAI 适配器
    openaiAdapter := llm.NewOpenAIAdapter("your-api-key", "gpt-4")
    manager.RegisterAdapter("openai", openaiAdapter)

    // 创建对话请求
    req := &llm.ChatRequest{
        Messages: []llm.Message{
            {Role: "user", Content: "解释一下什么是分布式系统"},
        },
        Temperature: 0.7,
        MaxTokens:   1000,
    }

    // 发送请求
    resp, err := manager.Chat(context.Background(), "openai", req)
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Content)
}
```

### 2. 流式响应

```go
// 流式对话
stream, err := manager.Stream(context.Background(), "openai", req)
if err != nil {
    panic(err)
}

for chunk := range stream {
    fmt.Print(chunk.Content)
}
```

### 3. 使用 Prompt 模板

```go
// 创建 Prompt 管理器
promptMgr := llm.NewPromptManager()

// 注册自定义模板
promptMgr.RegisterTemplate("task-analysis", llm.PromptTemplate{
    Name:        "task-analysis",
    Content:     "分析以下任务并提供解决方案：\n任务：{{.task}}\n约束：{{.constraints}}",
    Variables:   []string{"task", "constraints"},
    Description: "任务分析模板",
})

// 渲染模板
prompt, err := promptMgr.Render("task-analysis", map[string]interface{}{
    "task":        "处理100万条用户数据",
    "constraints": "延迟 < 100ms, 预算 $100",
})

// 使用渲染后的 prompt
req.Messages = []llm.Message{{Role: "user", Content: prompt}}
resp, _ := manager.Chat(context.Background(), "openai", req)
```

### 4. RAG 检索增强生成

```go
// 创建向量存储
vectorStore := llm.NewMemoryVectorStore()

// 添加知识库文档
docs := []llm.Document{
    {ID: "doc1", Content: "OFA是一个分布式Agent系统...", Metadata: map[string]string{"source": "README"}},
    {ID: "doc2", Content: "Agent可以通过gRPC与Center通信...", Metadata: map[string]string{"source": "ARCHITECTURE"}},
}
vectorStore.AddDocuments(docs)

// 创建知识库
kb := llm.NewKnowledgeBase("ofa-docs", vectorStore)

// 创建 RAG Agent
ragAgent := llm.NewRAGAgent(manager, kb)

// 查询
answer, _ := ragAgent.Query(context.Background(), "OFA如何实现Agent通信？")
fmt.Println(answer)
```

### 5. LLM Agent 工具调用

```go
// 创建 LLM Agent
agent := llm.NewLLMAgent("assistant", manager)

// 注册工具
agent.RegisterTool(llm.Tool{
    Name:        "get_weather",
    Description: "获取指定城市的天气",
    Parameters: map[string]interface{}{
        "city": map[string]string{"type": "string", "description": "城市名称"},
    },
    Handler: func(params map[string]interface{}) (interface{}, error) {
        city := params["city"].(string)
        return map[string]interface{}{
            "city":    city,
            "weather": "晴天",
            "temp":    25,
        }, nil
    },
})

// 执行对话
result, _ := agent.Execute(context.Background(), "北京今天天气怎么样？")
fmt.Println(result)
```

---

## 代码生成使用

### 1. 生成 API 代码

```go
package main

import (
    "ofa/pkg/codegen"
)

func main() {
    generator := codegen.NewGenerator(nil)
    apiGen := codegen.NewAPIGenerator(generator)

    // 定义 API 规范
    spec := codegen.APISpec{
        Name:        "UserService",
        Version:     "1.0.0",
        Description: "用户管理服务",
        Models: []codegen.ModelSpec{
            {
                Name:        "User",
                TableName:   "users",
                Description: "用户模型",
                Fields: []codegen.FieldSpec{
                    {Name: "ID", Type: "int64", JSONName: "id"},
                    {Name: "Name", Type: "string", JSONName: "name"},
                    {Name: "Email", Type: "string", JSONName: "email"},
                    {Name: "CreatedAt", Type: "time.Time", JSONName: "created_at"},
                },
            },
        },
        Endpoints: []codegen.EndpointSpec{
            {
                Name:        "GetUser",
                Method:      "GET",
                Path:        "/users/{id}",
                Description: "获取用户信息",
                Input:       "GetUserRequest",
                Output:      "User",
            },
            {
                Name:        "CreateUser",
                Method:      "POST",
                Path:        "/users",
                Description: "创建用户",
                Input:       "CreateUserRequest",
                Output:      "User",
            },
        },
    }

    // 生成代码
    apiGen.GenerateModels(spec, "./models")
    apiGen.GenerateHandlers(spec, "./handlers")
    apiGen.GenerateRoutes(spec, "./routes")
    apiGen.GenerateOpenAPI(spec, "./docs/openapi.json")
}
```

### 2. 生成 SDK

```go
sdkGen := codegen.NewSDKGenerator(generator)

sdkSpec := codegen.SDKSpec{
    Name:    "OFA",
    Version: "{version}",
    Package: "ofa",
    Methods: []codegen.SDKMethod{
        {
            Name:        "ListAgents",
            Description: "列出所有Agent",
            HTTPMethod:  "GET",
            HTTPPath:    "/api/v1/agents",
            ReturnType:  "[]Agent",
        },
        {
            Name:        "SubmitTask",
            Description: "提交任务",
            Params: []codegen.ParamSpec{
                {Name: "task", Type: "Task"},
            },
            HTTPMethod: "POST",
            HTTPPath:   "/api/v1/tasks",
            ReturnType: "Task",
        },
    },
}

// 生成多语言 SDK
sdkGen.GenerateGoSDK(sdkSpec, "./sdk/go")
sdkGen.GenerateTypeScriptSDK(sdkSpec, "./sdk/typescript")
sdkGen.GeneratePythonSDK(sdkSpec, "./sdk/python")
```

### 3. 生成 Proto 文件

```go
protoGen := codegen.NewProtoGenerator(generator)

protoSpec := codegen.ProtoSpec{
    Package:   "ofa",
    GoPackage: "github.com/ofa/proto",
    Services: []codegen.ServiceSpec{
        {
            Name: "AgentService",
            Methods: []codegen.RPCMethod{
                {Name: "Register", InputType: "RegisterRequest", OutputType: "RegisterResponse"},
                {Name: "ExecuteTask", InputType: "TaskRequest", OutputType: "TaskResponse"},
            },
        },
    },
    Messages: []codegen.MessageSpec{
        {
            Name: "RegisterRequest",
            Fields: []codegen.FieldSpec{
                {Name: "AgentId", Type: "string", JSONName: "agent_id"},
                {Name: "Name", Type: "string", JSONName: "name"},
            },
        },
    },
}

protoGen.GenerateProto(protoSpec, "./proto/ofa.proto")
```

---

## Agent协作使用

### 1. 顺序执行协作

```go
package main

import (
    "context"
    "ofa/pkg/collab"
)

func main() {
    manager := collab.NewCollaborationManager()

    // 创建协作任务
    req := &collab.CreateCollabRequest{
        Name:        "数据处理流水线",
        Type:        collab.CollabTypeSequential,
        Description: "顺序处理数据",
        Goal:        "完成数据ETL流程",
        Tasks: []*collab.CollabTask{
            {
                ID:          "extract",
                Name:        "数据抽取",
                SkillID:     "data.extract",
                Operation:   "extract",
                Description: "从数据源抽取数据",
                Priority:    1,
            },
            {
                ID:          "transform",
                Name:        "数据转换",
                SkillID:     "data.transform",
                Operation:   "transform",
                Description: "转换数据格式",
                Priority:    2,
                Dependencies: []string{"extract"},
            },
            {
                ID:          "load",
                Name:        "数据加载",
                SkillID:     "data.load",
                Operation:   "load",
                Description: "加载到目标存储",
                Priority:    3,
                Dependencies: []string{"transform"},
            },
        },
    }

    // 创建协作
    collab, err := manager.CreateCollaboration(context.Background(), req)
    if err != nil {
        panic(err)
    }

    // 启动协作
    manager.StartCollaboration(context.Background(), collab.ID)

    // 订阅事件
    manager.Subscribe(func(event *collab.CollabEvent) {
        fmt.Printf("事件: %s, 协作ID: %s\n", event.Type, event.CollabID)
    })

    // 获取结果
    result, _ := manager.GetCollaboration(collab.ID)
    fmt.Printf("结果: %+v\n", result.Result)
}
```

### 2. 并行执行协作

```go
// 并行图像处理
req := &collab.CreateCollabRequest{
    Name:        "批量图像处理",
    Type:        collab.CollabTypeParallel,
    Description: "并行处理多张图片",
    Tasks: []*collab.CollabTask{
        {ID: "img1", Name: "处理图片1", SkillID: "image.process", Priority: 1},
        {ID: "img2", Name: "处理图片2", SkillID: "image.process", Priority: 1},
        {ID: "img3", Name: "处理图片3", SkillID: "image.process", Priority: 1},
    },
    Constraints: &collab.Constraints{
        MaxDuration:   5 * time.Minute,
        MinAgents:     3,
        RequiredSkills: []string{"image.process"},
    },
}
```

### 3. MapReduce 协作

```go
// 分布式数据处理
req := &collab.CreateCollabRequest{
    Name:        "日志分析",
    Type:        collab.CollabTypeMapReduce,
    Description: "分析服务器日志",
    Tasks: []*collab.CollabTask{
        // Map 任务
        {ID: "map1", Name: "Map节点1", SkillID: "log.analyze", Operation: "map"},
        {ID: "map2", Name: "Map节点2", SkillID: "log.analyze", Operation: "map"},
        {ID: "map3", Name: "Map节点3", SkillID: "log.analyze", Operation: "map"},
        // Reduce 任务
        {ID: "reduce", Name: "Reduce汇总", SkillID: "log.reduce", Operation: "reduce"},
    },
}
```

### 4. Agent 协商

```go
// 创建协商器
negotiator := collab.NewNegotiator()

// 创建提议
proposal := negotiator.CreateProposal(
    "agent-001",
    "资源分配",
    collab.ProposalResource,
    map[string]interface{}{
        "resource": "GPU",
        "quantity": 2,
    },
)

// Agent 投票
negotiator.Vote(proposal.ID, "agent-002", collab.VoteFor, "资源充足", 1.0)
negotiator.Vote(proposal.ID, "agent-003", collab.VoteAgainst, "资源紧张", 0.8)

// 解决提议
agreement, err := negotiator.ResolveProposal(proposal.ID)
if err != nil {
    fmt.Println("提议被拒绝")
} else {
    fmt.Printf("协议达成: %+v\n", agreement)
}
```

---

## 去中心化功能使用

### 1. 创建去中心化网络

```go
package main

import (
    "context"
    "ofa/pkg/decentralized"
)

func main() {
    // 创建去中心化管理器
    manager := decentralized.NewDecentralizedManager(decentralized.NetworkHybrid)

    // 加入网络
    nodeInfo := &decentralized.NodeInfo{
        ID:           "node-001",
        Address:      "192.168.1.100",
        Port:         9090,
        Type:         decentralized.NodeTypeFull,
        Capabilities: []string{"compute", "storage"},
        MaxLoad:      10,
    }

    err := manager.JoinNetwork(context.Background(), nodeInfo)
    if err != nil {
        panic(err)
    }

    // 配置启动节点
    manager.discovery.AddBootstrap("192.168.1.101", 9090)
    manager.discovery.AddBootstrap("192.168.1.102", 9090)

    // 发现其他节点
    peers, _ := manager.discovery.DiscoverPeers(context.Background(), nodeInfo.ID)
    fmt.Printf("发现 %d 个节点\n", len(peers))
}
```

### 2. 使用共识机制

```go
// 创建共识引擎
consensus := decentralized.NewConsensusEngine()
consensus.SetAlgorithm(decentralized.AlgorithmWeighted)

// 添加验证者
consensus.AddValidator("node-001", 1.0)
consensus.AddValidator("node-002", 0.8)
consensus.AddValidator("node-003", 0.6)

// 创建提案
proposal := consensus.CreateProposal(
    "policy_update",
    "调整资源分配策略",
    map[string]interface{}{
        "new_policy": "priority_based",
        "threshold":  0.7,
    },
    "node-001",
)

// 节点投票
consensus.SubmitVote(&decentralized.Vote{
    VoterID:    "node-002",
    ProposalID: proposal.ID,
    Decision:   decentralized.VoteApprove,
    Weight:     0.8,
})

consensus.SubmitVote(&decentralized.Vote{
    VoterID:    "node-003",
    ProposalID: proposal.ID,
    Decision:   decentralized.VoteApprove,
    Weight:     0.6,
})

// 做出决策
decision, _ := consensus.MakeDecision(proposal)
fmt.Printf("决策: %v, 置信度: %.2f\n", decision.Approved, decision.Confidence)
```

### 3. 数据复制

```go
// 创建数据复制器
replicator := decentralized.NewDataReplicator()
replicator.SetReplicaLevel(3) // 3副本

// 存储数据
data := map[string]interface{}{
    "task_id": "task-001",
    "result":  "success",
    "data":    []byte{1, 2, 3, 4},
}

// 复制到多个节点
err := replicator.ReplicateData(context.Background(), data, []string{
    "node-001",
    "node-002",
    "node-003",
})

// 验证复制
verified, _ := replicator.VerifyReplica("data-xxx", "node-001")
fmt.Printf("复制验证: %v\n", verified)
```

### 4. 信任管理

```go
// 创建信任管理器
trustMgr := decentralized.NewTrustManager()

// 初始化节点信任评分
trustMgr.InitializeTrust("node-001", 0.5)
trustMgr.InitializeTrust("node-002", 0.5)

// 记录正面事件
trustMgr.RecordEvent("node-001", decentralized.EventTaskCompleted, 0.1, "任务成功完成")
trustMgr.RecordEvent("node-001", decentralized.EventResponseFast, 0.05, "响应时间 < 100ms")

// 记录负面事件
trustMgr.RecordEvent("node-002", decentralized.EventTaskFailed, -0.15, "任务执行失败")

// 获取信任评分
score1, _ := trustMgr.GetTrustScore("node-001")
score2, _ := trustMgr.GetTrustScore("node-002")

fmt.Printf("Node-001 信任评分: %.2f, 等级: %s\n", score1.Score, score1.Rating)
fmt.Printf("Node-002 信任评分: %.2f, 等级: %s\n", score2.Score, score2.Rating)

// 检查是否可信
if trustMgr.IsTrusted("node-001") {
    fmt.Println("Node-001 是可信节点")
}
```

### 5. Peer 发现

```go
discovery := decentralized.NewPeerDiscovery()
discovery.SetMethod(decentralized.DiscoveryGossip)

// 配置本地节点
discovery.SetLocalNodeID("node-001")

// 添加已知节点
discovery.AddPeer("node-002", "192.168.1.102", 9090)
discovery.AddPeer("node-003", "192.168.1.103", 9090)

// 持续发现
ctx := context.Background()
discovery.StartContinuousDiscovery(ctx, 30*time.Second)

// 列出所有节点
peers := discovery.ListPeers()
for _, peer := range peers {
    fmt.Printf("Peer: %s, 地址: %s:%d\n", peer.ID, peer.Address, peer.Port)
}
```

---

## 典型场景示例

### 场景1: 智能客服系统

```go
package main

import (
    "context"
    "fmt"
    "ofa/pkg/llm"
    "ofa/pkg/collab"
)

// 智能客服系统 - 结合 LLM 和 Agent 协作
func main() {
    // 1. 初始化 LLM
    llmMgr := llm.NewManager()
    llmMgr.RegisterAdapter("claude", llm.NewClaudeAdapter("api-key", "claude-sonnet-4-6"))

    // 2. 创建知识库
    vectorStore := llm.NewMemoryVectorStore()
    kb := llm.NewKnowledgeBase("customer-service", vectorStore)

    // 导入产品文档
    docs := loadProductDocs() // 加载产品手册
    vectorStore.AddDocuments(docs)

    // 3. 创建客服 Agent
    agent := llm.NewLLMAgent("customer-service", llmMgr)

    // 注册工具
    agent.RegisterTool(llm.Tool{
        Name:        "check_order",
        Description: "查询订单状态",
        Parameters: map[string]interface{}{
            "order_id": map[string]string{"type": "string"},
        },
        Handler: checkOrderStatus,
    })

    agent.RegisterTool(llm.Tool{
        Name:        "create_ticket",
        Description: "创建工单",
        Parameters: map[string]interface{}{
            "user_id":   map[string]string{"type": "string"},
            "issue":     map[string]string{"type": "string"},
            "priority":  map[string]string{"type": "string"},
        },
        Handler: createSupportTicket,
    })

    // 4. 处理用户查询
    userQuery := "我的订单12345到哪了？"
    response, _ := agent.Execute(context.Background(), userQuery)
    fmt.Println(response)
}

func checkOrderStatus(params map[string]interface{}) (interface{}, error) {
    orderID := params["order_id"].(string)
    // 查询订单系统
    return map[string]interface{}{
        "order_id": orderID,
        "status":   "配送中",
        "location": "北京分拨中心",
    }, nil
}

func createSupportTicket(params map[string]interface{}) (interface{}, error) {
    return map[string]interface{}{
        "ticket_id": "TKT-001",
        "status":    "created",
    }, nil
}
```

### 场景2: 分布式数据处理

```go
package main

import (
    "context"
    "fmt"
    "ofa/pkg/collab"
    "ofa/pkg/decentralized"
)

// 分布式数据处理 - 使用去中心化和协作
func main() {
    // 1. 初始化去中心化网络
    network := decentralized.NewDecentralizedManager(decentralized.NetworkMesh)

    // 加入网络
    nodes := []string{"node-001", "node-002", "node-003", "node-004"}
    for _, nodeID := range nodes {
        network.JoinNetwork(context.Background(), &decentralized.NodeInfo{
            ID:           nodeID,
            Address:      fmt.Sprintf("10.0.0.%s", nodeID[6:]),
            Port:         9090,
            Type:         decentralized.NodeTypeFull,
            Capabilities: []string{"data.process"},
            MaxLoad:      100,
        })
    }

    // 2. 创建协作任务
    collabMgr := collab.NewCollaborationManager()

    req := &collab.CreateCollabRequest{
        Name:        "大数据处理",
        Type:        collab.CollabTypeMapReduce,
        Description: "分布式处理100GB数据",
        Tasks: []*collab.CollabTask{
            // Map 任务
            {ID: "map-1", Name: "Map分片1", SkillID: "data.map", Operation: "map", Priority: 1},
            {ID: "map-2", Name: "Map分片2", SkillID: "data.map", Operation: "map", Priority: 1},
            {ID: "map-3", Name: "Map分片3", SkillID: "data.map", Operation: "map", Priority: 1},
            {ID: "map-4", Name: "Map分片4", SkillID: "data.map", Operation: "map", Priority: 1},
            // Reduce 任务
            {ID: "reduce-1", Name: "Reduce汇总", SkillID: "data.reduce", Operation: "reduce", Priority: 2},
        },
        Constraints: &collab.Constraints{
            MaxDuration: 30 * time.Minute,
            MinAgents:   4,
        },
    }

    collab, _ := collabMgr.CreateCollaboration(context.Background(), req)

    // 3. 分配任务到去中心化节点
    for _, nodeID := range nodes {
        // 使用共识机制确认任务分配
        consensus := network.consensus
        proposal := consensus.CreateProposal("task_assign", "分配Map任务", nil, nodeID)
        // ... 投票和决策过程
    }

    // 4. 启动协作
    collabMgr.StartCollaboration(context.Background(), collab.ID)

    // 5. 等待结果
    for {
        time.Sleep(1 * time.Second)
        result, _ := collabMgr.GetCollaboration(collab.ID)
        if result.State == collab.CollabStateCompleted {
            fmt.Printf("处理完成: %d 成功, %d 失败\n",
                result.Result.TasksSuccess, result.Result.TasksFailed)
            break
        }
    }
}
```

### 场景3: 自动化代码生成

```go
package main

import (
    "fmt"
    "ofa/pkg/codegen"
)

// 自动化代码生成 - 从 API 规范生成完整项目
func main() {
    generator := codegen.NewGenerator(nil)

    // 1. 定义 API 规范
    apiSpec := codegen.APISpec{
        Name:        "BlogService",
        Version:     "1.0.0",
        Description: "博客管理服务",
        Models: []codegen.ModelSpec{
            {
                Name:        "Post",
                TableName:   "posts",
                Description: "博客文章",
                Fields: []codegen.FieldSpec{
                    {Name: "ID", Type: "int64", JSONName: "id", Tag: `gorm:"primaryKey"`},
                    {Name: "Title", Type: "string", JSONName: "title", Tag: `gorm:"size:200"`},
                    {Name: "Content", Type: "string", JSONName: "content", Tag: `gorm:"type:text"`},
                    {Name: "AuthorID", Type: "int64", JSONName: "author_id"},
                    {Name: "Status", Type: "string", JSONName: "status"},
                    {Name: "CreatedAt", Type: "time.Time", JSONName: "created_at"},
                    {Name: "UpdatedAt", Type: "time.Time", JSONName: "updated_at"},
                },
            },
            {
                Name:        "Comment",
                TableName:   "comments",
                Description: "评论",
                Fields: []codegen.FieldSpec{
                    {Name: "ID", Type: "int64", JSONName: "id", Tag: `gorm:"primaryKey"`},
                    {Name: "PostID", Type: "int64", JSONName: "post_id"},
                    {Name: "Content", Type: "string", JSONName: "content"},
                    {Name: "AuthorName", Type: "string", JSONName: "author_name"},
                    {Name: "CreatedAt", Type: "time.Time", JSONName: "created_at"},
                },
            },
        },
        Endpoints: []codegen.EndpointSpec{
            {Name: "ListPosts", Method: "GET", Path: "/posts", Output: "[]Post"},
            {Name: "GetPost", Method: "GET", Path: "/posts/{id}", Output: "Post"},
            {Name: "CreatePost", Method: "POST", Path: "/posts", Input: "CreatePostRequest", Output: "Post"},
            {Name: "UpdatePost", Method: "PUT", Path: "/posts/{id}", Input: "UpdatePostRequest", Output: "Post"},
            {Name: "DeletePost", Method: "DELETE", Path: "/posts/{id}"},
            {Name: "ListComments", Method: "GET", Path: "/posts/{id}/comments", Output: "[]Comment"},
            {Name: "CreateComment", Method: "POST", Path: "/posts/{id}/comments", Input: "CreateCommentRequest", Output: "Comment"},
        },
    }

    // 2. 生成模型代码
    apiGen := codegen.NewAPIGenerator(generator)
    apiGen.GenerateModels(apiSpec, "./internal/models")
    fmt.Println("✓ 模型代码生成完成")

    // 3. 生成处理器
    apiGen.GenerateHandlers(apiSpec, "./internal/handlers")
    fmt.Println("✓ 处理器代码生成完成")

    // 4. 生成路由
    apiGen.GenerateRoutes(apiSpec, "./internal/routes")
    fmt.Println("✓ 路由代码生成完成")

    // 5. 生成 OpenAPI 文档
    apiGen.GenerateOpenAPI(apiSpec, "./docs/openapi.json")
    fmt.Println("✓ OpenAPI 文档生成完成")

    // 6. 生成 SDK
    sdkGen := codegen.NewSDKGenerator(generator)
    sdkSpec := convertAPIToSDK(apiSpec)

    sdkGen.GenerateGoSDK(sdkSpec, "./sdk/go")
    fmt.Println("✓ Go SDK 生成完成")

    sdkGen.GenerateTypeScriptSDK(sdkSpec, "./sdk/typescript")
    fmt.Println("✓ TypeScript SDK 生成完成")

    sdkGen.GeneratePythonSDK(sdkSpec, "./sdk/python")
    fmt.Println("✓ Python SDK 生成完成")

    // 7. 生成文档
    docGen := codegen.NewDocGenerator(generator)
    docSpec := codegen.DocSpec{
        Title:       "Blog Service API",
        Version:     "1.0.0",
        Description: "博客管理服务 API 文档",
        APIs:        convertToEndpoints(apiSpec.Endpoints),
    }

    markdown := docGen.GenerateMarkdown(docSpec)
    // 保存 markdown...

    fmt.Println("\n🚀 代码生成完成！项目结构：")
    fmt.Println(`
./blog-service/
├── internal/
│   ├── models/      # 数据模型
│   ├── handlers/    # HTTP 处理器
│   └── routes/      # 路由定义
├── sdk/
│   ├── go/          # Go SDK
│   ├── typescript/  # TypeScript SDK
│   └── python/      # Python SDK
└── docs/
    └── openapi.json # OpenAPI 文档
    `)
}
```

### 场景4: 跨设备协同任务

```go
package main

import (
    "context"
    "fmt"
    "time"
    "ofa/pkg/collab"
    "ofa/pkg/decentralized"
    "ofa/pkg/messaging"
)

// 跨设备协同 - 手机、手表、智能家居联动
func main() {
    // 1. 创建去中心化网络
    network := decentralized.NewDecentralizedManager(decentralized.NetworkHybrid)

    // 2. 注册设备
    devices := map[string]*decentralized.NodeInfo{
        "phone": {
            ID:           "phone-001",
            Address:      "192.168.1.100",
            Type:         decentralized.NodeTypeLight,
            Capabilities: []string{"notification", "camera", "gps"},
        },
        "watch": {
            ID:           "watch-001",
            Address:      "192.168.1.101",
            Type:         decentralized.NodeTypeLight,
            Capabilities: []string{"heart_rate", "step_count", "notification"},
        },
        "home": {
            ID:           "home-001",
            Address:      "192.168.1.102",
            Type:         decentralized.NodeTypeEdge,
            Capabilities: []string{"light_control", "ac_control", "security"},
        },
    }

    for _, device := range devices {
        network.JoinNetwork(context.Background(), device)
    }

    // 3. 设置信任评分
    trustMgr := network.trustManager
    for id := range devices {
        trustMgr.InitializeTrust(id, 0.7)
    }

    // 4. 创建健康监测协作
    collabMgr := collab.NewCollaborationManager()

    healthCollab := &collab.CreateCollabRequest{
        Name:        "健康监测联动",
        Type:        collab.CollabTypePipeline,
        Description: "监测心率异常并联动设备",
        Tasks: []*collab.CollabTask{
            {
                ID:          "monitor_heart",
                Name:        "心率监测",
                SkillID:     "heart_rate",
                Description: "手表持续监测心率",
                AssignedTo:  "watch-001",
                Priority:    1,
            },
            {
                ID:          "check_anomaly",
                Name:        "异常检测",
                SkillID:     "anomaly.detect",
                Description: "检测心率异常",
                Dependencies: []string{"monitor_heart"},
                Priority:    2,
            },
            {
                ID:          "alert_user",
                Name:        "用户提醒",
                SkillID:     "notification",
                Description: "手机发送提醒",
                AssignedTo:  "phone-001",
                Dependencies: []string{"check_anomaly"},
                Priority:    3,
            },
            {
                ID:          "home_response",
                Name:        "家居响应",
                SkillID:     "home.control",
                Description: "调整家居环境",
                AssignedTo:  "home-001",
                Dependencies: []string{"check_anomaly"},
                Priority:    3,
            },
        },
    }

    collab, _ := collabMgr.CreateCollaboration(context.Background(), healthCollab)

    // 5. 订阅事件
    collabMgr.Subscribe(func(event *collab.CollabEvent) {
        switch event.Type {
        case "task_completed":
            fmt.Printf("任务完成: %s\n", event.TaskID)
        case "completed":
            result, _ := collabMgr.GetCollaboration(event.CollabID)
            fmt.Printf("协作完成: 成功=%v\n", result.Result.Success)
        }
    })

    // 6. 启动协作
    collabMgr.StartCollaboration(context.Background(), collab.ID)

    // 7. 模拟心率数据
    go func() {
        for {
            heartRate := getHeartRateFromWatch() // 从手表获取心率
            if heartRate > 100 {
                // 触发异常处理
                fmt.Printf("检测到心率异常: %d bpm\n", heartRate)
            }
            time.Sleep(5 * time.Second)
        }
    }()

    select {} // 保持运行
}

func getHeartRateFromWatch() int {
    return 75 // 模拟数据
}
```

---

## REST API 使用

### LLM API

```bash
# 对话接口
curl -X POST http://localhost:8080/api/v1/llm/chat \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'

# 流式对话
curl -X POST http://localhost:8080/api/v1/llm/stream \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "claude",
    "messages": [
      {"role": "user", "content": "Write a poem"}
    ]
  }'

# 向量搜索
curl -X POST http://localhost:8080/api/v1/rag/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何部署 OFA",
    "top_k": 5
  }'
```

### 协作 API

```bash
# 创建协作
curl -X POST http://localhost:8080/api/v1/collab \
  -H "Content-Type: application/json" \
  -d '{
    "name": "数据处理",
    "type": "parallel",
    "tasks": [
      {"id": "t1", "name": "任务1", "skill_id": "data.process"},
      {"id": "t2", "name": "任务2", "skill_id": "data.process"}
    ]
  }'

# 启动协作
curl -X POST http://localhost:8080/api/v1/collab/{id}/start

# 获取协作状态
curl http://localhost:8080/api/v1/collab/{id}

# 取消协作
curl -X DELETE http://localhost:8080/api/v1/collab/{id}
```

### 去中心化 API

```bash
# 获取网络状态
curl http://localhost:8080/api/v1/network/stats

# 列出节点
curl http://localhost:8080/api/v1/network/nodes

# 获取节点信任评分
curl http://localhost:8080/api/v1/network/nodes/{id}/trust

# 创建提案
curl -X POST http://localhost:8080/api/v1/consensus/proposal \
  -H "Content-Type: application/json" \
  -d '{
    "type": "policy_update",
    "subject": "资源分配",
    "content": {"policy": "new"}
  }'
```

---

## 命令行工具

```bash
# 构建
./scripts/ofa.bat build

# 测试
./scripts/ofa.bat test

# 运行 Center
./scripts/ofa.bat run-center

# 运行 Agent
./scripts/ofa.bat run-agent

# 生成代码
./scripts/ofa.bat generate --type api --spec api.yaml --output ./generated

# 健康检查
./scripts/ofa.bat health

# 查看版本
./scripts/ofa.bat version
```

---

## 常见问题

### Q: 如何配置多个 LLM 提供商？

```yaml
llm:
  default_provider: openai
  providers:
    openai:
      api_key: ${OPENAI_API_KEY}
      model: gpt-4
    claude:
      api_key: ${ANTHROPIC_API_KEY}
      model: claude-sonnet-4-6
    local:
      base_url: http://localhost:11434
      model: llama2
```

### Q: 如何添加自定义工具？

```go
agent.RegisterTool(llm.Tool{
    Name:        "my_tool",
    Description: "自定义工具描述",
    Parameters: map[string]interface{}{
        "param1": map[string]string{"type": "string"},
    },
    Handler: func(params map[string]interface{}) (interface{}, error) {
        return "result", nil
    },
})
```

### Q: 如何实现任务重试？

```go
task := &collab.CollabTask{
    ID:         "task-1",
    RetryCount: 0,
    MaxRetries: 3,
}

// 在任务失败时
if task.State == collab.TaskStateFailed && task.RetryCount < task.MaxRetries {
    task.RetryCount++
    task.State = collab.TaskStatePending
    // 重新执行
}
```

---

## 更多资源

- [架构文档](./docs/03-ARCHITECTURE_DESIGN.md)
- [API 文档](./docs/API.md)
- [部署指南](./docs/DEPLOYMENT.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [GitHub 仓库](https://github.com/ofa/ofa)
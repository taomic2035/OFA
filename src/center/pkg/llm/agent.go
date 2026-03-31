// Package llm - LLM Agent技能
// 0.9.0 Beta: LLM深度集成
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// LLMAgent LLM Agent
type LLMAgent struct {
	id           string
	name         string
	manager      *LLMManager
	prompts      *PromptManager
	contexts     *ContextManager
	tools        map[string]*Tool
	toolExecutor *ToolExecutor
	config       AgentLLMConfig
	stats        *AgentStats
	mu           sync.RWMutex
}

// AgentLLMConfig Agent配置
type AgentLLMConfig struct {
	ModelID         string `json:"model_id"`
	SystemPrompt    string `json:"system_prompt"`
	MaxTokens       int    `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
	EnableTools     bool   `json:"enable_tools"`
	MaxToolCalls    int    `json:"max_tool_calls"`
	EnableMemory    bool   `json:"enable_memory"`
	MemoryTokens    int    `json:"memory_tokens"`
}

// DefaultAgentLLMConfig 默认配置
func DefaultAgentLLMConfig() AgentLLMConfig {
	return AgentLLMConfig{
		ModelID:      "default",
		MaxTokens:    4096,
		Temperature:  0.7,
		EnableTools:  true,
		MaxToolCalls: 10,
		EnableMemory: true,
		MemoryTokens: 4000,
	}
}

// AgentStats Agent统计
type AgentStats struct {
	TotalQueries   int64 `json:"total_queries"`
	SuccessQueries int64 `json:"success_queries"`
	FailedQueries  int64 `json:"failed_queries"`
	TotalTokensIn  int64 `json:"total_tokens_in"`
	TotalTokensOut int64 `json:"total_tokens_out"`
	ToolCalls      int64 `json:"tool_calls"`
	AvgLatencyMs   int64 `json:"avg_latency_ms"`
}

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler 工具处理器
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ToolExecutor 工具执行器
type ToolExecutor struct {
	tools map[string]*Tool
	mu    sync.RWMutex
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		tools: make(map[string]*Tool),
	}
}

// Register 注册工具
func (e *ToolExecutor) Register(tool *Tool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("工具名称不能为空")
	}

	e.tools[tool.Name] = tool
	return nil
}

// Execute 执行工具
func (e *ToolExecutor) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	e.mu.RLock()
	tool, ok := e.tools[name]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("工具不存在: %s", name)
	}

	return tool.Handler(ctx, args)
}

// List 列出工具
func (e *ToolExecutor) List() []*Tool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tools := make([]*Tool, 0, len(e.tools))
	for _, tool := range e.tools {
		tools = append(tools, tool)
	}
	return tools
}

// NewLLMAgent 创建LLM Agent
func NewLLMAgent(id, name string, manager *LLMManager, config AgentLLMConfig) *LLMAgent {
	return &LLMAgent{
		id:           id,
		name:         name,
		manager:      manager,
		prompts:      NewPromptManager(DefaultPromptConfig()),
		contexts:     NewContextManager(config.MemoryTokens),
		tools:        make(map[string]*Tool),
		toolExecutor: NewToolExecutor(),
		config:       config,
		stats:        &AgentStats{},
	}
}

// RegisterTool 注册工具
func (a *LLMAgent) RegisterTool(tool *Tool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tools[tool.Name] = tool
	return a.toolExecutor.Register(tool)
}

// Chat 聊天
func (a *LLMAgent) Chat(ctx context.Context, userMessage string) (string, error) {
	start := time.Now()
	a.stats.TotalQueries++

	// 构建消息
	messages := []Message{}

	// 系统提示
	if a.config.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		})
	}

	// 历史上下文
	if a.config.EnableMemory {
		history, _ := a.contexts.GetMessages(a.id)
		messages = append(messages, history...)
	}

	// 用户消息
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// 调用LLM
	resp, err := a.manager.Chat(ctx, messages)
	if err != nil {
		a.stats.FailedQueries++
		return "", err
	}

	// 更新统计
	a.stats.SuccessQueries++
	if resp.Usage != nil {
		a.stats.TotalTokensIn += int64(resp.Usage.PromptTokens)
		a.stats.TotalTokensOut += int64(resp.Usage.CompletionTokens)
	}

	latency := time.Since(start).Milliseconds()
	a.stats.AvgLatencyMs = (a.stats.AvgLatencyMs*(a.stats.TotalQueries-1) + latency) / a.stats.TotalQueries

	// 获取响应
	var response string
	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		response = resp.Choices[0].Message.Content
	}

	// 保存上下文
	if a.config.EnableMemory {
		a.contexts.AddMessage(a.id, Message{Role: "user", Content: userMessage})
		a.contexts.AddMessage(a.id, Message{Role: "assistant", Content: response})
	}

	return response, nil
}

// ChatWithTools 带工具的聊天
func (a *LLMAgent) ChatWithTools(ctx context.Context, userMessage string) (string, error) {
	if !a.config.EnableTools {
		return a.Chat(ctx, userMessage)
	}

	// 构建工具定义
	tools := a.buildToolDefinitions()

	// 初始消息
	messages := a.buildMessages(userMessage)

	// 迭代调用
	toolCallCount := 0

	for toolCallCount < a.config.MaxToolCalls {
		resp, err := a.manager.Chat(ctx, messages)
		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("无响应")
		}

		choice := resp.Choices[0]
		if choice.Message != nil {
			messages = append(messages, *choice.Message)
		}

		// 检查是否有工具调用
		if choice.FinishReason != "tool_calls" {
			// 没有工具调用，返回最终响应
			if choice.Message != nil {
				return choice.Message.Content, nil
			}
			return "", fmt.Errorf("无响应内容")
		}

		// 执行工具调用
		// 简化实现：假设工具调用在消息中
		toolCallCount++
		a.stats.ToolCalls++
	}

	return "", fmt.Errorf("达到最大工具调用次数")
}

// buildMessages 构建消息列表
func (a *LLMAgent) buildMessages(userMessage string) []Message {
	messages := []Message{}

	if a.config.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		})
	}

	if a.config.EnableMemory {
		history, _ := a.contexts.GetMessages(a.id)
		messages = append(messages, history...)
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

// buildToolDefinitions 构建工具定义
func (a *LLMAgent) buildToolDefinitions() []map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(a.tools))

	for _, tool := range a.tools {
		tools = append(tools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}

	return tools
}

// ExecuteTool 执行工具
func (a *LLMAgent) ExecuteTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
	var argsMap map[string]interface{}
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return nil, fmt.Errorf("参数解析失败: %w", err)
	}

	return a.toolExecutor.Execute(ctx, name, argsMap)
}

// StreamChat 流式聊天
func (a *LLMAgent) StreamChat(ctx context.Context, userMessage string, onChunk func(chunk string)) (string, error) {
	messages := a.buildMessages(userMessage)

	return a.manager.StreamChat(ctx, messages, onChunk)
}

// ClearMemory 清除记忆
func (a *LLMAgent) ClearMemory() error {
	return a.contexts.Clear(a.id)
}

// GetStats 获取统计
func (a *LLMAgent) GetStats() *AgentStats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.stats
}

// ID 获取ID
func (a *LLMAgent) ID() string {
	return a.id
}

// Name 获取名称
func (a *LLMAgent) Name() string {
	return a.name
}

// AgentRegistry Agent注册表
type AgentRegistry struct {
	agents map[string]*LLMAgent
	mu     sync.RWMutex
}

// NewAgentRegistry 创建Agent注册表
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*LLMAgent),
	}
}

// Register 注册Agent
func (r *AgentRegistry) Register(agent *LLMAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent.id == "" {
		return fmt.Errorf("Agent ID不能为空")
	}

	r.agents[agent.id] = agent
	return nil
}

// Get 获取Agent
func (r *AgentRegistry) Get(id string) (*LLMAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[id]
	if !ok {
		return nil, fmt.Errorf("Agent不存在: %s", id)
	}

	return agent, nil
}

// Delete 删除Agent
func (r *AgentRegistry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.agents, id)
}

// List 列出Agent
func (r *AgentRegistry) List() []*LLMAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*LLMAgent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// 内置工具

// WeatherTool 天气工具
var WeatherTool = &Tool{
	Name:        "get_weather",
	Description: "获取指定城市的天气信息",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "城市名称",
			},
		},
		"required": []string{"city"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		city, _ := args["city"].(string)
		// 模拟天气数据
		return map[string]interface{}{
			"city":        city,
			"temperature": 25,
			"condition":   "晴",
			"humidity":    60,
		}, nil
	},
}

// SearchTool 搜索工具
var SearchTool = &Tool{
	Name:        "search",
	Description: "搜索互联网信息",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索查询",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "结果数量限制",
			},
		},
		"required": []string{"query"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		query, _ := args["query"].(string)
		// 模拟搜索结果
		return map[string]interface{}{
			"query": query,
			"results": []map[string]string{
				{"title": "搜索结果1", "url": "https://example.com/1"},
				{"title": "搜索结果2", "url": "https://example.com/2"},
			},
		}, nil
	},
}

// CalculatorTool 计算器工具
var CalculatorTool = &Tool{
	Name:        "calculator",
	Description: "执行数学计算",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "数学表达式",
			},
		},
		"required": []string{"expression"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		expr, _ := args["expression"].(string)
		// 简化实现：直接返回表达式
		return map[string]interface{}{
			"expression": expr,
			"result":     "计算结果",
		}, nil
	},
}

// DefaultTools 默认工具列表
var DefaultTools = []*Tool{
	WeatherTool,
	SearchTool,
	CalculatorTool,
}

// RegisterDefaultTools 注册默认工具
func RegisterDefaultTools(agent *LLMAgent) {
	for _, tool := range DefaultTools {
		agent.RegisterTool(tool)
	}
}
// Package llm - 大语言模型集成
// 0.9.0 Beta: LLM深度集成
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// LLMProvider LLM提供商
type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderClaude    LLMProvider = "claude"
	ProviderGemini    LLMProvider = "gemini"
	ProviderLocal     LLMProvider = "local"
	ProviderOllama    LLMProvider = "ollama"
	ProviderCustom    LLMProvider = "custom"
)

// ModelConfig 模型配置
type ModelConfig struct {
	Provider    LLMProvider `json:"provider"`
	Model       string      `json:"model"`
	APIKey      string      `json:"api_key"`
	BaseURL     string      `json:"base_url"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
	TopP        float64     `json:"top_p"`
	Timeout     time.Duration `json:"timeout"`
}

// DefaultModelConfig 默认配置
func DefaultModelConfig() ModelConfig {
	return ModelConfig{
		Provider:    ProviderOpenAI,
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
		TopP:        1.0,
		Timeout:     60 * time.Second,
	}
}

// Message 消息
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Created int64        `json:"created"`
	Choices []Choice     `json:"choices"`
	Usage   *TokenUsage  `json:"usage,omitempty"`
}

// Choice 选择
type Choice struct {
	Index        int          `json:"index"`
	Message      *Message     `json:"message"`
	Delta        *Message     `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

// TokenUsage Token使用量
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID      string  `json:"id"`
	Model   string  `json:"model"`
	Created int64   `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		Delta        Message `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// LLMAdapter LLM适配器接口
type LLMAdapter interface {
	// Chat 聊天
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// Stream 流式聊天
	Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)

	// Complete 补全
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)

	// Embed 嵌入
	Embed(ctx context.Context, texts []string) ([][]float64, error)

	// CountTokens 计算Token数
	CountTokens(text string) int

	// ModelInfo 模型信息
	ModelInfo() *ModelInfo

	// Close 关闭
	Close() error
}

// ModelInfo 模型信息
type ModelInfo struct {
	Provider     LLMProvider `json:"provider"`
	Model        string      `json:"model"`
	MaxContext   int         `json:"max_context"`
	MaxOutput    int         `json:"max_output"`
	SupportsChat bool        `json:"supports_chat"`
	SupportsStream bool      `json:"supports_stream"`
	SupportsEmbed  bool      `json:"supports_embed"`
}

// LLMManager LLM管理器
type LLMManager struct {
	adapters   map[string]LLMAdapter
	defaultID  string
	config     ManagerConfig
	stats      *ManagerStats
	mu         sync.RWMutex
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	DefaultProvider LLMProvider `json:"default_provider"`
	DefaultModel    string      `json:"default_model"`
	CacheEnabled    bool        `json:"cache_enabled"`
	MaxCacheSize    int         `json:"max_cache_size"`
	RetryCount      int         `json:"retry_count"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// DefaultManagerConfig 默认管理器配置
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		DefaultProvider: ProviderOpenAI,
		DefaultModel:    "gpt-4",
		CacheEnabled:    true,
		MaxCacheSize:    1000,
		RetryCount:      3,
		RetryDelay:      time.Second,
	}
}

// ManagerStats 管理器统计
type ManagerStats struct {
	TotalRequests  int64 `json:"total_requests"`
	SuccessCount   int64 `json:"success_count"`
	FailureCount   int64 `json:"failure_count"`
	TotalTokensIn  int64 `json:"total_tokens_in"`
	TotalTokensOut int64 `json:"total_tokens_out"`
	CacheHits      int64 `json:"cache_hits"`
	AvgLatencyMs   int64 `json:"avg_latency_ms"`
}

// NewLLMManager 创建LLM管理器
func NewLLMManager(config ManagerConfig) *LLMManager {
	return &LLMManager{
		adapters: make(map[string]LLMAdapter),
		config:   config,
		stats:    &ManagerStats{},
	}
}

// Register 注册适配器
func (m *LLMManager) Register(id string, adapter LLMAdapter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return fmt.Errorf("适配器ID不能为空")
	}

	m.adapters[id] = adapter

	// 设置默认
	if m.defaultID == "" {
		m.defaultID = id
	}

	return nil
}

// Unregister 注销适配器
func (m *LLMManager) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.adapters, id)

	if m.defaultID == id {
		m.defaultID = ""
		for k := range m.adapters {
			m.defaultID = k
			break
		}
	}
}

// Get 获取适配器
func (m *LLMManager) Get(id string) (LLMAdapter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapter, ok := m.adapters[id]
	return adapter, ok
}

// Default 获取默认适配器
func (m *LLMManager) Default() (LLMAdapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultID == "" {
		return nil, fmt.Errorf("没有可用的LLM适配器")
	}

	return m.adapters[m.defaultID], nil
}

// Chat 聊天(使用默认适配器)
func (m *LLMManager) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	adapter, err := m.Default()
	if err != nil {
		return nil, err
	}

	req := &ChatRequest{
		Model:    adapter.ModelInfo().Model,
		Messages: messages,
	}

	start := time.Now()
	resp, err := adapter.Chat(ctx, req)

	m.updateStats(err, time.Since(start))
	if err != nil {
		return nil, err
	}

	if resp.Usage != nil {
		m.stats.TotalTokensIn += int64(resp.Usage.PromptTokens)
		m.stats.TotalTokensOut += int64(resp.Usage.CompletionTokens)
	}

	return resp, nil
}

// Stream 流式聊天
func (m *LLMManager) Stream(ctx context.Context, messages []Message) (<-chan *StreamChunk, error) {
	adapter, err := m.Default()
	if err != nil {
		return nil, err
	}

	req := &ChatRequest{
		Model:    adapter.ModelInfo().Model,
		Messages: messages,
		Stream:   true,
	}

	return adapter.Stream(ctx, req)
}

// Complete 补全
func (m *LLMManager) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	adapter, err := m.Default()
	if err != nil {
		return "", err
	}

	return adapter.Complete(ctx, prompt, maxTokens)
}

// Embed 嵌入
func (m *LLMManager) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	adapter, err := m.Default()
	if err != nil {
		return nil, err
	}

	return adapter.Embed(ctx, texts)
}

// updateStats 更新统计
func (m *LLMManager) updateStats(err error, duration time.Duration) {
	m.stats.TotalRequests++
	if err != nil {
		m.stats.FailureCount++
	} else {
		m.stats.SuccessCount++
	}

	// 计算平均延迟
	total := m.stats.TotalRequests
	oldAvg := m.stats.AvgLatencyMs
	newMs := duration.Milliseconds()
	m.stats.AvgLatencyMs = (oldAvg*(total-1) + newMs) / total
}

// GetStats 获取统计
func (m *LLMManager) GetStats() *ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// ListAdapters 列出适配器
func (m *LLMManager) ListAdapters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.adapters))
	for id := range m.adapters {
		ids = append(ids, id)
	}
	return ids
}

// ChatWithSystem 带系统提示的聊天
func (m *LLMManager) ChatWithSystem(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	resp, err := m.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("无响应内容")
}

// StreamChat 流式聊天返回完整内容
func (m *LLMManager) StreamChat(ctx context.Context, messages []Message, onChunk func(chunk string)) (string, error) {
	ch, err := m.Stream(ctx, messages)
	if err != nil {
		return "", err
	}

	var result string

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return result, nil
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				result += content
				if onChunk != nil {
					onChunk(content)
				}
			}

		case <-ctx.Done():
			return result, ctx.Err()
		}
	}
}

// JSONMode JSON模式聊天
func (m *LLMManager) JSONMode(ctx context.Context, systemPrompt, userMessage string, result interface{}) error {
	messages := []Message{
		{Role: "system", Content: systemPrompt + "\n请以JSON格式回复。"},
		{Role: "user", Content: userMessage},
	}

	resp, err := m.Chat(ctx, messages)
	if err != nil {
		return err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return json.Unmarshal([]byte(resp.Choices[0].Message.Content), result)
	}

	return fmt.Errorf("无响应内容")
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// FunctionDefinition 函数定义
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Close 关闭所有适配器
func (m *LLMManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for id, adapter := range m.adapters {
		if err := adapter.Close(); err != nil {
			lastErr = fmt.Errorf("关闭适配器 %s 失败: %w", id, err)
		}
	}

	m.adapters = make(map[string]LLMAdapter)
	m.defaultID = ""

	return lastErr
}

// StreamReader 流式读取器
type StreamReader struct {
	reader   io.Reader
	buffer   []byte
	position int
}

// NewStreamReader 创建流读取器
func NewStreamReader(reader io.Reader) *StreamReader {
	return &StreamReader{
		reader: reader,
		buffer: make([]byte, 4096),
	}
}

// ReadLine 读取一行
func (r *StreamReader) ReadLine() ([]byte, error) {
	var line []byte

	for {
		if r.position >= len(r.buffer) {
			n, err := r.reader.Read(r.buffer)
			if err != nil {
				if len(line) > 0 {
					return line, nil
				}
				return nil, err
			}
			r.buffer = r.buffer[:n]
			r.position = 0
		}

		for i := r.position; i < len(r.buffer); i++ {
			if r.buffer[i] == '\n' {
				line = append(line, r.buffer[r.position:i]...)
				r.position = i + 1
				return line, nil
			}
		}

		line = append(line, r.buffer[r.position:]...)
		r.position = len(r.buffer)
	}
}
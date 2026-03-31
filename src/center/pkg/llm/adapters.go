// Package llm - OpenAI适配器
// 0.9.0 Beta: LLM深度集成
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIAdapter OpenAI适配器
type OpenAIAdapter struct {
	config   ModelConfig
	client   *http.Client
	modelInfo *ModelInfo
}

// NewOpenAIAdapter 创建OpenAI适配器
func NewOpenAIAdapter(config ModelConfig) *OpenAIAdapter {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	return &OpenAIAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		modelInfo: &ModelInfo{
			Provider:       ProviderOpenAI,
			Model:          config.Model,
			MaxContext:     128000, // GPT-4 Turbo
			MaxOutput:      4096,
			SupportsChat:   true,
			SupportsStream: true,
			SupportsEmbed:  true,
		},
	}
}

// Chat 聊天
func (a *OpenAIAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = a.config.Model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API错误: %s", errResp.Error.Message)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &chatResp, nil
}

// Stream 流式聊天
func (a *OpenAIAdapter) Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	if req.Model == "" {
		req.Model = a.config.Model
	}
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	ch := make(chan *StreamChunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			ch <- &chunk
		}
	}()

	return ch, nil
}

// Complete 补全
func (a *OpenAIAdapter) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	req := &ChatRequest{
		Model: a.config.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: maxTokens,
	}

	resp, err := a.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("无响应内容")
}

// Embed 嵌入
func (a *OpenAIAdapter) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	req := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{
		Model: "text-embedding-3-small",
		Input: texts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var embedResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	embeddings := make([][]float64, len(embedResp.Data))
	for i, d := range embedResp.Data {
		embeddings[i] = d.Embedding
	}

	return embeddings, nil
}

// CountTokens 计算Token数
func (a *OpenAIAdapter) CountTokens(text string) int {
	// 简化实现：平均4字符约1个token
	return len(text) / 4
}

// ModelInfo 模型信息
func (a *OpenAIAdapter) ModelInfo() *ModelInfo {
	return a.modelInfo
}

// Close 关闭
func (a *OpenAIAdapter) Close() error {
	return nil
}

// ClaudeAdapter Claude适配器
type ClaudeAdapter struct {
	config    ModelConfig
	client    *http.Client
	modelInfo *ModelInfo
}

// NewClaudeAdapter 创建Claude适配器
func NewClaudeAdapter(config ModelConfig) *ClaudeAdapter {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}

	return &ClaudeAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		modelInfo: &ModelInfo{
			Provider:       ProviderClaude,
			Model:          config.Model,
			MaxContext:     200000,
			MaxOutput:      4096,
			SupportsChat:   true,
			SupportsStream: true,
			SupportsEmbed:  false,
		},
	}
}

// Chat 聊天
func (a *ClaudeAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = a.config.Model
	}

	// 转换为Claude格式
	claudeReq := struct {
		Model       string        `json:"model"`
		MaxTokens   int           `json:"max_tokens"`
		Messages    []ClaudeMsg   `json:"messages"`
		System      string        `json:"system,omitempty"`
		Temperature float64       `json:"temperature,omitempty"`
		Stream      bool          `json:"stream"`
	}{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	// 分离系统消息和用户消息
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			claudeReq.System = msg.Content
		} else {
			claudeReq.Messages = append(claudeReq.Messages, ClaudeMsg{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	if claudeReq.MaxTokens == 0 {
		claudeReq.MaxTokens = 4096
	}

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API错误: %s", errResp.Error.Message)
	}

	var claudeResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为标准格式
	chatResp := &ChatResponse{
		ID:      claudeResp.ID,
		Model:   claudeResp.Model,
		Created: time.Now().Unix(),
		Choices: []Choice{
			{
				Index: 0,
				Message: &Message{
					Role:    "assistant",
					Content: claudeResp.Content[0].Text,
				},
			},
		},
		Usage: &TokenUsage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}

	return chatResp, nil
}

// ClaudeMsg Claude消息格式
type ClaudeMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Stream 流式聊天
func (a *ClaudeAdapter) Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	// Claude流式实现类似OpenAI
	return nil, fmt.Errorf("暂未实现")
}

// Complete 补全
func (a *ClaudeAdapter) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	req := &ChatRequest{
		Model: a.config.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: maxTokens,
	}

	resp, err := a.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("无响应内容")
}

// Embed 嵌入
func (a *ClaudeAdapter) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, fmt.Errorf("Claude不支持嵌入")
}

// CountTokens 计算Token数
func (a *ClaudeAdapter) CountTokens(text string) int {
	return len(text) / 4
}

// ModelInfo 模型信息
func (a *ClaudeAdapter) ModelInfo() *ModelInfo {
	return a.modelInfo
}

// Close 关闭
func (a *ClaudeAdapter) Close() error {
	return nil
}

// LocalAdapter 本地模型适配器
type LocalAdapter struct {
	config    ModelConfig
	client    *http.Client
	modelInfo *ModelInfo
}

// NewLocalAdapter 创建本地模型适配器
func NewLocalAdapter(config ModelConfig) *LocalAdapter {
	return &LocalAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		modelInfo: &ModelInfo{
			Provider:       ProviderLocal,
			Model:          config.Model,
			MaxContext:     8192,
			MaxOutput:      2048,
			SupportsChat:   true,
			SupportsStream: true,
			SupportsEmbed:  true,
		},
	}
}

// Chat 聊天
func (a *LocalAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &chatResp, nil
}

// Stream 流式聊天
func (a *LocalAdapter) Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	ch := make(chan *StreamChunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			ch <- &chunk
		}
	}()

	return ch, nil
}

// Complete 补全
func (a *LocalAdapter) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	req := &ChatRequest{
		Model: a.config.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: maxTokens,
	}

	resp, err := a.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("无响应内容")
}

// Embed 嵌入
func (a *LocalAdapter) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	req := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{
		Model: a.config.Model,
		Input: texts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		a.config.BaseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embedResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}

	embeddings := make([][]float64, len(embedResp.Data))
	for i, d := range embedResp.Data {
		embeddings[i] = d.Embedding
	}

	return embeddings, nil
}

// CountTokens 计算Token数
func (a *LocalAdapter) CountTokens(text string) int {
	return len(text) / 4
}

// ModelInfo 模型信息
func (a *LocalAdapter) ModelInfo() *ModelInfo {
	return a.modelInfo
}

// Close 关闭
func (a *LocalAdapter) Close() error {
	return nil
}
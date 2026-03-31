// Package llm - LLM服务集成
// 0.9.0 Beta: LLM深度集成
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LLMService LLM服务
type LLMService struct {
	manager  *LLMManager
	agents   *AgentRegistry
	registry *KnowledgeRegistry
	config   ServiceConfig
	mu       sync.RWMutex
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	DefaultModel    string `json:"default_model"`
	MaxConcurrent   int    `json:"max_concurrent"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	EnableCache     bool   `json:"enable_cache"`
	CacheTTL        time.Duration `json:"cache_ttl"`
}

// DefaultServiceConfig 默认配置
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		DefaultModel:   "gpt-4",
		MaxConcurrent:  100,
		RequestTimeout: 60 * time.Second,
		EnableCache:    true,
		CacheTTL:       time.Hour,
	}
}

// NewLLMService 创建LLM服务
func NewLLMService(config ServiceConfig) *LLMService {
	return &LLMService{
		manager:  NewLLMManager(DefaultManagerConfig()),
		agents:   NewAgentRegistry(),
		registry: NewKnowledgeRegistry(),
		config:   config,
	}
}

// RegisterAdapter 注册适配器
func (s *LLMService) RegisterAdapter(id string, adapter LLMAdapter) error {
	return s.manager.Register(id, adapter)
}

// RegisterAgent 注册Agent
func (s *LLMService) RegisterAgent(agent *LLMAgent) error {
	return s.agents.Register(agent)
}

// Chat 聊天
func (s *LLMService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return s.manager.Chat(ctx, req.Messages)
}

// AgentChat Agent聊天
func (s *LLMService) AgentChat(ctx context.Context, agentID, message string) (string, error) {
	agent, err := s.agents.Get(agentID)
	if err != nil {
		return "", err
	}

	return agent.Chat(ctx, message)
}

// CreateKnowledgeBase 创建知识库
func (s *LLMService) CreateKnowledgeBase(id, name string) (*KnowledgeBase, error) {
	adapter, err := s.manager.Default()
	if err != nil {
		return nil, err
	}

	kb := NewKnowledgeBase(id, name, NewMemoryVectorStore(1536), &llmEmbedder{adapter})
	s.registry.Register(kb)
	return kb, nil
}

// GetKnowledgeBase 获取知识库
func (s *LLMService) GetKnowledgeBase(id string) (*KnowledgeBase, error) {
	return s.registry.Get(id)
}

// HTTP处理

// HandleChat 处理聊天请求
func (s *LLMService) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.RequestTimeout)
	defer cancel()

	if req.Stream {
		s.handleStreamChat(ctx, w, req.Model, req.Messages)
		return
	}

	chatReq := &ChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
	}

	resp, err := s.manager.Chat(ctx, chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStreamChat 处理流式聊天
func (s *LLMService) handleStreamChat(ctx context.Context, w http.ResponseWriter, model string, messages []Message) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持流式响应", http.StatusInternalServerError)
		return
	}

	ch, err := s.manager.Stream(ctx, messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for chunk := range ch {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// HandleComplete 处理补全请求
func (s *LLMService) HandleComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt    string `json:"prompt"`
		MaxTokens int    `json:"max_tokens"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.RequestTimeout)
	defer cancel()

	result, err := s.manager.Complete(ctx, req.Prompt, req.MaxTokens)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": result})
}

// HandleEmbed 处理嵌入请求
func (s *LLMService) HandleEmbed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Texts []string `json:"texts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.RequestTimeout)
	defer cancel()

	embeddings, err := s.manager.Embed(ctx, req.Texts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"embeddings": embeddings})
}

// HandleAgentChat 处理Agent聊天
func (s *LLMService) HandleAgentChat(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agent_id")
	if agentID == "" {
		http.Error(w, "缺少agent_id", http.StatusBadRequest)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.config.RequestTimeout)
	defer cancel()

	response, err := s.AgentChat(ctx, agentID, req.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

// HandleAgents 列出Agent
func (s *LLMService) HandleAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.agents.List()

	list := make([]map[string]interface{}, len(agents))
	for i, agent := range agents {
		list[i] = map[string]interface{}{
			"id":    agent.ID(),
			"name":  agent.Name(),
			"stats": agent.GetStats(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetStats 获取统计
func (s *LLMService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"manager": s.manager.GetStats(),
		"adapters": s.manager.ListAdapters(),
		"agents":   len(s.agents.List()),
	}
}

// Close 关闭
func (s *LLMService) Close() error {
	return s.manager.Close()
}

// KnowledgeRegistry 知识库注册表
type KnowledgeRegistry struct {
	bases map[string]*KnowledgeBase
	mu    sync.RWMutex
}

// NewKnowledgeRegistry 创建知识库注册表
func NewKnowledgeRegistry() *KnowledgeRegistry {
	return &KnowledgeRegistry{
		bases: make(map[string]*KnowledgeBase),
	}
}

// Register 注册知识库
func (r *KnowledgeRegistry) Register(kb *KnowledgeBase) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.bases[kb.ID()] = kb
}

// Get 获取知识库
func (r *KnowledgeRegistry) Get(id string) (*KnowledgeBase, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	kb, ok := r.bases[id]
	if !ok {
		return nil, fmt.Errorf("知识库不存在: %s", id)
	}

	return kb, nil
}

// Delete 删除知识库
func (r *KnowledgeRegistry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.bases, id)
}

// List 列出知识库
func (r *KnowledgeRegistry) List() []*KnowledgeBase {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bases := make([]*KnowledgeBase, 0, len(r.bases))
	for _, kb := range r.bases {
		bases = append(bases, kb)
	}
	return bases
}

// llmEmbedder LLM嵌入器适配器
type llmEmbedder struct {
	adapter LLMAdapter
}

// Embed 嵌入
func (e *llmEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	return e.adapter.Embed(ctx, texts)
}

// LLMClient LLM客户端
type LLMClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewLLMClient 创建LLM客户端
func NewLLMClient(baseURL, apiKey string) *LLMClient {
	return &LLMClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
	}
}

// Chat 聊天
func (c *LLMClient) Chat(ctx context.Context, messages []Message) (*ChatResponse, error) {
	req := ChatRequest{
		Model:    "default",
		Messages: messages,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v1/llm/chat", nil)
	if err != nil {
		return nil, err
	}

	// 这里简化实现
	_ = body

	return &ChatResponse{}, nil
}
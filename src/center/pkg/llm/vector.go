// Package llm - 向量存储
// Sprint 30: v9.0 LLM深度集成
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// VectorStore 向量存储接口
type VectorStore interface {
	// Store 存储向量
	Store(ctx context.Context, id string, vector []float64, metadata map[string]interface{}) error

	// StoreBatch 批量存储
	StoreBatch(ctx context.Context, items []VectorItem) error

	// Search 搜索相似向量
	Search(ctx context.Context, vector []float64, k int) ([]SearchResult, error)

	// Delete 删除向量
	Delete(ctx context.Context, id string) error

	// Get 获取向量
	Get(ctx context.Context, id string) (*VectorItem, error)

	// Count 统计数量
	Count(ctx context.Context) (int, error)
}

// VectorItem 向量项
type VectorItem struct {
	ID       string                 `json:"id"`
	Vector   []float64              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MemoryVectorStore 内存向量存储
type MemoryVectorStore struct {
	items map[string]*VectorItem
	dim   int
	mu    sync.RWMutex
}

// NewMemoryVectorStore 创建内存向量存储
func NewMemoryVectorStore(dim int) *MemoryVectorStore {
	return &MemoryVectorStore{
		items: make(map[string]*VectorItem),
		dim:   dim,
	}
}

// Store 存储向量
func (s *MemoryVectorStore) Store(ctx context.Context, id string, vector []float64, metadata map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(vector) != s.dim {
		return fmt.Errorf("向量维度不匹配: 期望 %d, 实际 %d", s.dim, len(vector))
	}

	s.items[id] = &VectorItem{
		ID:       id,
		Vector:   vector,
		Metadata: metadata,
	}

	return nil
}

// StoreBatch 批量存储
func (s *MemoryVectorStore) StoreBatch(ctx context.Context, items []VectorItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		if len(item.Vector) != s.dim {
			return fmt.Errorf("向量维度不匹配: %s", item.ID)
		}
		s.items[item.ID] = &item
	}

	return nil
}

// Search 搜索相似向量
func (s *MemoryVectorStore) Search(ctx context.Context, vector []float64, k int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(vector) != s.dim {
		return nil, fmt.Errorf("向量维度不匹配")
	}

	// 计算所有相似度
	results := make([]SearchResult, 0, len(s.items))
	for id, item := range s.items {
		score := cosineSimilarity(vector, item.Vector)
		results = append(results, SearchResult{
			ID:       id,
			Score:    score,
			Metadata: item.Metadata,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回top k
	if k > 0 && k < len(results) {
		results = results[:k]
	}

	return results, nil
}

// Delete 删除向量
func (s *MemoryVectorStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, id)
	return nil
}

// Get 获取向量
func (s *MemoryVectorStore) Get(ctx context.Context, id string) (*VectorItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return nil, fmt.Errorf("向量不存在: %s", id)
	}

	return item, nil
}

// Count 统计数量
func (s *MemoryVectorStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.items), nil
}

// cosineSimilarity 余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// RAG RAG检索增强生成
type RAG struct {
	store      VectorStore
	embedder   Embedder
	llm        LLMAdapter
	chunkSize  int
	overlap    int
}

// Embedder 嵌入器接口
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// RAGConfig RAG配置
type RAGConfig struct {
	ChunkSize int `json:"chunk_size"`
	Overlap   int `json:"overlap"`
	TopK      int `json:"top_k"`
}

// DefaultRAGConfig 默认配置
func DefaultRAGConfig() RAGConfig {
	return RAGConfig{
		ChunkSize: 500,
		Overlap:   50,
		TopK:      5,
	}
}

// NewRAG 创建RAG
func NewRAG(store VectorStore, embedder Embedder, llm LLMAdapter, config RAGConfig) *RAG {
	return &RAG{
		store:     store,
		embedder:  embedder,
		llm:       llm,
		chunkSize: config.ChunkSize,
		overlap:   config.Overlap,
	}
}

// IndexDocument 索引文档
func (r *RAG) IndexDocument(ctx context.Context, id, content string, metadata map[string]interface{}) error {
	// 分块
	chunks := r.chunkText(content)

	// 嵌入
	embeddings, err := r.embedder.Embed(ctx, chunks)
	if err != nil {
		return fmt.Errorf("嵌入失败: %w", err)
	}

	// 存储
	items := make([]VectorItem, len(chunks))
	for i, chunk := range chunks {
		meta := map[string]interface{}{
			"document_id": id,
			"chunk_index": i,
			"content":     chunk,
		}
		for k, v := range metadata {
			meta[k] = v
		}

		items[i] = VectorItem{
			ID:       fmt.Sprintf("%s-%d", id, i),
			Vector:   embeddings[i],
			Metadata: meta,
		}
	}

	return r.store.StoreBatch(ctx, items)
}

// Query 查询
func (r *RAG) Query(ctx context.Context, query string) (string, error) {
	// 嵌入查询
	embeddings, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return "", fmt.Errorf("嵌入失败: %w", err)
	}

	// 搜索相似文档
	results, err := r.store.Search(ctx, embeddings[0], 5)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}

	// 构建上下文
	context := r.buildContext(results)

	// 生成回答
	messages := []Message{
		{Role: "system", Content: "根据以下上下文回答问题：\n\n" + context},
		{Role: "user", Content: query},
	}

	req := &ChatRequest{
		Model:    r.llm.ModelInfo().Model,
		Messages: messages,
	}

	resp, err := r.llm.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("生成失败: %w", err)
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("无响应")
}

// chunkText 分块文本
func (r *RAG) chunkText(text string) []string {
	// 简单分块
	chunks := make([]string, 0)

	for i := 0; i < len(text); i += r.chunkSize - r.overlap {
		end := i + r.chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
		if end >= len(text) {
			break
		}
	}

	return chunks
}

// buildContext 构建上下文
func (r *RAG) buildContext(results []SearchResult) string {
	context := ""
	for i, result := range results {
		if content, ok := result.Metadata["content"].(string); ok {
			context += fmt.Sprintf("\n[文档 %d]\n%s\n", i+1, content)
		}
	}
	return context
}

// KnowledgeBase 知识库
type KnowledgeBase struct {
	id        string
	name      string
	store     VectorStore
	embedder  Embedder
	documents map[string]*Document
	mu        sync.RWMutex
}

// Document 文档
type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewKnowledgeBase 创建知识库
func NewKnowledgeBase(id, name string, store VectorStore, embedder Embedder) *KnowledgeBase {
	return &KnowledgeBase{
		id:        id,
		name:      name,
		store:     store,
		embedder:  embedder,
		documents: make(map[string]*Document),
	}
}

// AddDocument 添加文档
func (kb *KnowledgeBase) AddDocument(ctx context.Context, id, content string, metadata map[string]interface{}) error {
	// 嵌入
	embeddings, err := kb.embedder.Embed(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("嵌入失败: %w", err)
	}

	// 存储向量
	if err := kb.store.Store(ctx, id, embeddings[0], metadata); err != nil {
		return fmt.Errorf("存储失败: %w", err)
	}

	// 存储文档
	kb.mu.Lock()
	defer kb.mu.Unlock()

	now := time.Now()
	kb.documents[id] = &Document{
		ID:        id,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return nil
}

// Search 搜索
func (kb *KnowledgeBase) Search(ctx context.Context, query string, k int) ([]SearchResult, error) {
	// 嵌入查询
	embeddings, err := kb.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入失败: %w", err)
	}

	// 搜索
	return kb.store.Search(ctx, embeddings[0], k)
}

// GetDocument 获取文档
func (kb *KnowledgeBase) GetDocument(id string) (*Document, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	doc, ok := kb.documents[id]
	if !ok {
		return nil, fmt.Errorf("文档不存在: %s", id)
	}

	return doc, nil
}

// DeleteDocument 删除文档
func (kb *KnowledgeBase) DeleteDocument(ctx context.Context, id string) error {
	// 删除向量
	if err := kb.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除向量失败: %w", err)
	}

	// 删除文档
	kb.mu.Lock()
	defer kb.mu.Unlock()

	delete(kb.documents, id)
	return nil
}

// ListDocuments 列出文档
func (kb *KnowledgeBase) ListDocuments() []*Document {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	docs := make([]*Document, 0, len(kb.documents))
	for _, doc := range kb.documents {
		docs = append(docs, doc)
	}
	return docs
}

// ToJSON 序列化为JSON
func (kb *KnowledgeBase) ToJSON() ([]byte, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	data := struct {
		ID        string                `json:"id"`
		Name      string                `json:"name"`
		Documents map[string]*Document  `json:"documents"`
	}{
		ID:        kb.id,
		Name:      kb.name,
		Documents: kb.documents,
	}

	return json.MarshalIndent(data, "", "  ")
}

// ID 获取ID
func (kb *KnowledgeBase) ID() string {
	return kb.id
}

// Name 获取名称
func (kb *KnowledgeBase) Name() string {
	return kb.name
}
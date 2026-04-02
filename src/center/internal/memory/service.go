package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// Service - 记忆服务
type Service struct {
	mu       sync.RWMutex
	store    MemoryStore
	embedder Embedder
	policy   *models.ForgettingPolicy
	registry *RecallRegistry
}

// MemoryStore - 记忆存储接口
type MemoryStore interface {
	SaveMemory(ctx context.Context, memory *models.Memory) error
	GetMemory(ctx context.Context, id string) (*models.Memory, error)
	DeleteMemory(ctx context.Context, id string) error
	ListMemories(ctx context.Context, query *models.MemoryQuery) ([]*models.Memory, int, error)
	UpdateMemory(ctx context.Context, memory *models.Memory) error

	// 批量操作
	SaveMemories(ctx context.Context, memories []*models.Memory) error
	DeleteMemories(ctx context.Context, ids []string) error

	// 统计
	GetMemoryStats(ctx context.Context, userID string) (*models.MemoryStats, error)

	// 关联
	SaveAssociation(ctx context.Context, assoc *models.MemoryAssociation) error
	GetAssociations(ctx context.Context, memoryID string) ([]*models.MemoryAssociation, error)
}

// Embedder - 向量嵌入接口
type Embedder interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
	Dimension() int
}

// RecallRegistry - 召回策略注册表
type RecallRegistry struct {
	strategies map[string]RecallStrategy
}

// RecallStrategy - 召回策略
type RecallStrategy interface {
	Name() string
	Recall(ctx context.Context, query *models.MemoryQuery) ([]*models.Memory, error)
}

// NewService 创建记忆服务
func NewService(store MemoryStore, embedder Embedder) *Service {
	return &Service{
		store:    store,
		embedder: embedder,
		policy:   models.DefaultForgettingPolicy(),
		registry: NewRecallRegistry(),
	}
}

// NewRecallRegistry 创建召回注册表
func NewRecallRegistry() *RecallRegistry {
	return &RecallRegistry{
		strategies: make(map[string]RecallStrategy),
	}
}

// RegisterStrategy 注册召回策略
func (r *RecallRegistry) RegisterStrategy(strategy RecallStrategy) {
	r.strategies[strategy.Name()] = strategy
}

// === 记忆管理 ===

// Remember 记住内容
func (s *Service) Remember(ctx context.Context, userID string, memType models.MemoryType, content string, opts ...RememberOption) (*models.Memory, error) {
	memory := models.NewMemory(userID, memType, content)

	// 应用选项
	for _, opt := range opts {
		opt(memory)
	}

	// 生成向量嵌入
	if s.embedder != nil && content != "" {
		embedding, err := s.embedder.Embed(content)
		if err == nil {
			memory.Embedding = embedding
		}
	}

	// 计算初始重要性
	s.calculateInitialImportance(memory)

	// 保存
	if err := s.store.SaveMemory(ctx, memory); err != nil {
		return nil, fmt.Errorf("failed to save memory: %w", err)
	}

	return memory, nil
}

// RememberOption 记忆选项
type RememberOption func(*models.Memory)

// WithImportance 设置重要性
func WithImportance(importance float64) RememberOption {
	return func(m *models.Memory) {
		m.Importance = importance
	}
}

// WithPriority 设置优先级
func WithPriority(priority int) RememberOption {
	return func(m *models.Memory) {
		m.Priority = priority
	}
}

// WithCategory 设置类别
func WithCategory(category string) RememberOption {
	return func(m *models.Memory) {
		m.Category = category
	}
}

// WithTags 设置标签
func WithTags(tags []string) RememberOption {
	return func(m *models.Memory) {
		m.Tags = tags
	}
}

// WithSource 设置来源
func WithSource(source string) RememberOption {
	return func(m *models.Memory) {
		m.Source = source
	}
}

// WithEmotion 设置情感
func WithEmotion(emotion string, score float64) RememberOption {
	return func(m *models.Memory) {
		m.Emotion = emotion
		m.EmotionScore = score
	}
}

// WithTimestamp 设置时间戳
func WithTimestamp(t time.Time) RememberOption {
	return func(m *models.Memory) {
		m.Timestamp = t
	}
}

// WithLayer 设置层级
func WithLayer(layer models.MemoryLayer) RememberOption {
	return func(m *models.Memory) {
		m.Layer = layer
	}
}

// Recall 召回记忆
func (s *Service) Recall(ctx context.Context, query *models.MemoryQuery) (*models.MemoryRecallResult, error) {
	startTime := time.Now()

	// 设置默认限制
	if query.Limit == 0 {
		query.Limit = 10
	}

	// 执行查询
	memories, total, err := s.store.ListMemories(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to recall memories: %w", err)
	}

	// 更新访问统计
	for _, m := range memories {
		m.Access()
		s.store.UpdateMemory(ctx, m) // 异步更新？
	}

	return &models.MemoryRecallResult{
		Memories:   memories,
		Total:      total,
		QueryTime:  time.Since(startTime).Milliseconds(),
		RecallType: string(query.Semantic != "" || query.Keywords != ""),
	}, nil
}

// RecallByID 按ID召回
func (s *Service) RecallByID(ctx context.Context, id string) (*models.Memory, error) {
	memory, err := s.store.GetMemory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	memory.Access()
	s.store.UpdateMemory(ctx, memory)

	return memory, nil
}

// RecallByType 按类型召回
func (s *Service) RecallByType(ctx context.Context, userID string, memType models.MemoryType, limit int) ([]*models.Memory, error) {
	query := &models.MemoryQuery{
		UserID: userID,
		Types:  []models.MemoryType{memType},
		Limit:  limit,
		SortBy: "timestamp",
		SortDesc: true,
	}

	result, err := s.Recall(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// RecallByTime 按时间范围召回
func (s *Service) RecallByTime(ctx context.Context, userID string, start, end time.Time, limit int) ([]*models.Memory, error) {
	query := &models.MemoryQuery{
		UserID:    userID,
		StartTime: &start,
		EndTime:   &end,
		Limit:     limit,
		SortBy:    "timestamp",
		SortDesc:  true,
	}

	result, err := s.Recall(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// RecallRecent 召回最近的记忆
func (s *Service) RecallRecent(ctx context.Context, userID string, limit int) ([]*models.Memory, error) {
	query := &models.MemoryQuery{
		UserID:   userID,
		Limit:    limit,
		SortBy:   "timestamp",
		SortDesc: true,
	}

	result, err := s.Recall(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// RecallImportant 召回重要记忆
func (s *Service) RecallImportant(ctx context.Context, userID string, minImportance float64, limit int) ([]*models.Memory, error) {
	query := &models.MemoryQuery{
		UserID:        userID,
		MinImportance: minImportance,
		Limit:         limit,
		SortBy:        "importance",
		SortDesc:      true,
	}

	result, err := s.Recall(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// RecallSemantic 语义召回
func (s *Service) RecallSemantic(ctx context.Context, userID string, query string, limit int) ([]*models.Memory, error) {
	// 生成查询向量
	var queryEmbedding []float32
	if s.embedder != nil {
		emb, err := s.embedder.Embed(query)
		if err == nil {
			queryEmbedding = emb
		}
	}

	memQuery := &models.MemoryQuery{
		UserID:   userID,
		Semantic: query,
		Limit:    limit,
	}

	// 如果有向量，使用向量搜索
	if queryEmbedding != nil {
		// 简化实现：使用关键词搜索，实际应该用向量相似度
		memQuery.Keywords = query
	}

	result, err := s.Recall(ctx, memQuery)
	if err != nil {
		return nil, err
	}

	return result.Memories, nil
}

// RecallRelated 召回关联记忆
func (s *Service) RecallRelated(ctx context.Context, memoryID string, limit int) ([]*models.Memory, error) {
	// 获取关联
	associations, err := s.store.GetAssociations(ctx, memoryID)
	if err != nil {
		return nil, err
	}

	// 收集关联记忆
	var relatedIDs []string
	for _, assoc := range associations {
		relatedIDs = append(relatedIDs, assoc.TargetID)
	}

	if len(relatedIDs) == 0 {
		return []*models.Memory{}, nil
	}

	// 获取记忆
	var memories []*models.Memory
	for _, id := range relatedIDs {
		if len(memories) >= limit {
			break
		}
		m, err := s.store.GetMemory(ctx, id)
		if err == nil {
			memories = append(memories, m)
		}
	}

	return memories, nil
}

// Forget 遗忘记忆
func (s *Service) Forget(ctx context.Context, id string) error {
	return s.store.DeleteMemory(ctx, id)
}

// UpdateMemory 更新记忆
func (s *Service) UpdateMemory(ctx context.Context, id string, updates map[string]interface{}) (*models.Memory, error) {
	memory, err := s.store.GetMemory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "content":
			if v, ok := value.(string); ok {
				memory.Content = v
			}
		case "importance":
			if v, ok := value.(float64); ok {
				memory.Importance = v
			}
		case "priority":
			if v, ok := value.(int); ok {
				memory.Priority = v
			}
		case "category":
			if v, ok := value.(string); ok {
				memory.Category = v
			}
		case "tags":
			if v, ok := value.([]string); ok {
				memory.Tags = v
			}
		case "emotion":
			if v, ok := value.(string); ok {
				memory.Emotion = v
			}
		case "emotion_score":
			if v, ok := value.(float64); ok {
				memory.EmotionScore = v
			}
		case "summary":
			if v, ok := value.(string); ok {
				memory.Summary = v
			}
		}
	}

	memory.UpdatedAt = time.Now()

	if err := s.store.SaveMemory(ctx, memory); err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}

	return memory, nil
}

// === 记忆关联 ===

// Associate 关联记忆
func (s *Service) Associate(ctx context.Context, sourceID, targetID string, relationType string, strength float64) error {
	assoc := &models.MemoryAssociation{
		SourceID:     sourceID,
		TargetID:     targetID,
		RelationType: relationType,
		Strength:     strength,
		CreatedAt:    time.Now(),
	}

	// 更新源记忆的关联列表
	source, err := s.store.GetMemory(ctx, sourceID)
	if err == nil {
		source.Relate(targetID)
		s.store.UpdateMemory(ctx, source)
	}

	return s.store.SaveAssociation(ctx, assoc)
}

// === 记忆巩固 ===

// Consolidate 执行记忆巩固
func (s *Service) Consolidate(ctx context.Context, userID string) (*models.ConsolidationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := &models.ConsolidationResult{
		UserID:         userID,
		PromotedToL2:   []string{},
		PromotedToL3:   []string{},
		Demoted:        []string{},
		Forgotten:      []string{},
		Merged:         []string{},
		ConsolidatedAt: time.Now(),
	}

	// 获取所有记忆
	query := &models.MemoryQuery{
		UserID: userID,
		Limit:  10000,
	}

	memories, _, err := s.store.ListMemories(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}

	for _, memory := range memories {
		// 应用时间衰减
		memory.Decay(s.policy.DecayRate)

		// 检查是否应该遗忘
		if memory.ShouldForget(s.policy.ForgettingThreshold) {
			s.store.DeleteMemory(ctx, memory.ID)
			result.Forgotten = append(result.Forgotten, memory.ID)
			continue
		}

		// 检查是否需要层级调整
		oldLayer := memory.Layer
		if s.shouldPromote(memory) {
			memory.Promote()
			if memory.Layer == models.MemoryLayerL2 {
				result.PromotedToL2 = append(result.PromotedToL2, memory.ID)
			} else if memory.Layer == models.MemoryLayerL3 {
				result.PromotedToL3 = append(result.PromotedToL3, memory.ID)
			}
		} else if s.shouldDemote(memory) {
			memory.Demote()
			if oldLayer != memory.Layer {
				result.Demoted = append(result.Demoted, memory.ID)
			}
		}

		memory.UpdatedAt = time.Now()
		s.store.UpdateMemory(ctx, memory)
	}

	return result, nil
}

// shouldPromote 是否应该提升
func (s *Service) shouldPromote(memory *models.Memory) bool {
	// 高重要性 + 高访问次数 = 提升
	effectiveImportance := memory.GetEffectiveImportance()
	return effectiveImportance > 0.7 && memory.AccessCount > 3
}

// shouldDemote 是否应该降级
func (s *Service) shouldDemote(memory *models.Memory) bool {
	// 低重要性 + 低访问次数 + 长时间未访问 = 降级
	effectiveImportance := memory.GetEffectiveImportance()
	timeSinceAccess := time.Since(memory.LastAccessed)

	return effectiveImportance < 0.3 &&
		memory.AccessCount < 2 &&
		timeSinceAccess > 7*24*time.Hour
}

// === 统计 ===

// GetStats 获取记忆统计
func (s *Service) GetStats(ctx context.Context, userID string) (*models.MemoryStats, error) {
	return s.store.GetMemoryStats(ctx, userID)
}

// === 辅助方法 ===

// calculateInitialImportance 计算初始重要性
func (s *Service) calculateInitialImportance(memory *models.Memory) {
	// 基础重要性
	base := 0.5

	// 根据类型调整
	switch memory.Type {
	case models.MemoryTypeEmotional:
		if memory.EmotionScore > 0.7 {
			base += 0.2
		}
	case models.MemoryTypeFact:
		base += 0.1 // 事实更重要
	case models.MemoryTypePreference:
		base += 0.1 // 偏好重要
	}

	// 根据来源调整
	if memory.Source == "manual" {
		base += 0.1 // 手动记录更重要
	}

	// 根据优先级调整
	base += float64(memory.Priority-5) * 0.05

	// 限制范围
	if base > 1 {
		base = 1
	}
	if base < 0 {
		base = 0
	}

	memory.Importance = base
}
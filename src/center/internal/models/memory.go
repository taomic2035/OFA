package models

import (
	"time"
)

// === 记忆系统数据模型 ===

// Memory - 记忆单元
type Memory struct {
	ID           string                 `json:"id" bson:"_id"`
	UserID       string                 `json:"user_id" bson:"user_id"`
	Type         MemoryType             `json:"type" bson:"type"`
	Category     string                 `json:"category" bson:"category"`

	// 内容
	Content      string                 `json:"content" bson:"content"`
	Summary      string                 `json:"summary" bson:"summary"`         // AI 生成的摘要
	Embedding    []float32              `json:"embedding,omitempty" bson:"embedding,omitempty"`

	// 元数据
	Importance   float64                `json:"importance" bson:"importance"`    // 重要性 (0-1)
	Priority     int                    `json:"priority" bson:"priority"`        // 优先级 (1-10)
	Emotion      string                 `json:"emotion" bson:"emotion"`          // 情感标签
	EmotionScore float64                `json:"emotion_score" bson:"emotion_score"` // 情感强度
	Tags         []string               `json:"tags" bson:"tags"`
	Entities     []string               `json:"entities" bson:"entities"`        // 提取的实体

	// 来源
	Source       string                 `json:"source" bson:"source"`            // agent/manual/import/derived
	SourceAgent  string                 `json:"source_agent" bson:"source_agent"`
	SourceApp    string                 `json:"source_app" bson:"source_app"`    // 来源应用

	// 时间
	Timestamp    time.Time              `json:"timestamp" bson:"timestamp"`      // 记忆发生时间
	LastAccessed time.Time              `json:"last_accessed" bson:"last_accessed"`
	AccessCount  int                    `json:"access_count" bson:"access_count"`

	// 记忆层级
	Layer        MemoryLayer            `json:"layer" bson:"layer"`              // L1/L2/L3
	DecayFactor  float64                `json:"decay_factor" bson:"decay_factor"` // 衰减因子

	// 关联
	RelatedIDs   []string               `json:"related_ids" bson:"related_ids"`
	ParentID     string                 `json:"parent_id" bson:"parent_id"`      // 父记忆ID

	// 向量检索
	VectorID     string                 `json:"vector_id" bson:"vector_id"`

	// 创建/更新时间
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" bson:"updated_at"`
}

// MemoryType - 记忆类型
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"    // 事件记忆：发生了什么
	MemoryTypeSemantic   MemoryType = "semantic"    // 语义记忆：知识、概念
	MemoryTypeProcedural MemoryType = "procedural"  // 程序记忆：技能、习惯
	MemoryTypePreference MemoryType = "preference"  // 偏好记忆：喜欢什么
	MemoryTypeEmotional  MemoryType = "emotional"   // 情感记忆：感受如何
	MemoryTypeFact       MemoryType = "fact"        // 事实记忆：个人信息
	MemoryTypeSkill      MemoryType = "skill"       // 技能记忆：会做什么
)

// MemoryLayer - 记忆层级（三层记忆架构）
type MemoryLayer string

const (
	MemoryLayerL1 MemoryLayer = "L1" // 工作记忆：当前活跃，容量小，快速访问
	MemoryLayerL2 MemoryLayer = "L2" // 情景记忆：近期重要，中等容量
	MemoryLayerL3 MemoryLayer = "L3" // 长期记忆：历史沉淀，大容量
)

// MemoryStats - 记忆统计
type MemoryStats struct {
	UserID          string         `json:"user_id"`
	TotalCount      int            `json:"total_count"`
	CountByType     map[MemoryType]int `json:"count_by_type"`
	CountByLayer    map[MemoryLayer]int `json:"count_by_layer"`
	TotalSize       int64          `json:"total_size"`       // 字节数
	OldestMemory    time.Time      `json:"oldest_memory"`
	NewestMemory    time.Time      `json:"newest_memory"`
	AvgImportance   float64        `json:"avg_importance"`
	AvgAccessCount  float64        `json:"avg_access_count"`
	TopCategories   []string       `json:"top_categories"`
	TopTags         []string       `json:"top_tags"`
}

// MemoryQuery - 记忆查询
type MemoryQuery struct {
	UserID      string     `json:"user_id"`
	Types       []MemoryType `json:"types,omitempty"`
	Categories  []string   `json:"categories,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Layer       MemoryLayer `json:"layer,omitempty"`
	Keywords    string     `json:"keywords,omitempty"`
	Semantic    string     `json:"semantic,omitempty"`     // 语义搜索文本
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	MinImportance float64  `json:"min_importance,omitempty"`
	Emotion     string     `json:"emotion,omitempty"`
	Source      string     `json:"source,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	SortBy      string     `json:"sort_by"`        // timestamp/importance/access_count
	SortDesc    bool       `json:"sort_desc"`
}

// MemoryRecallResult - 记忆召回结果
type MemoryRecallResult struct {
	Memories   []*Memory `json:"memories"`
	Total      int       `json:"total"`
	QueryTime  int64     `json:"query_time_ms"` // 查询耗时
	RecallType string    `json:"recall_type"`   // semantic/keyword/time/association
}

// MemoryAssociation - 记忆关联
type MemoryAssociation struct {
	SourceID    string  `json:"source_id"`
	TargetID    string  `json:"target_id"`
	RelationType string `json:"relation_type"` // caused_by/related_to/similar_to/follows
	Strength    float64 `json:"strength"`      // 关联强度 (0-1)
	CreatedAt   time.Time `json:"created_at"`
}

// ConsolidationResult - 记忆巩固结果
type ConsolidationResult struct {
	UserID           string   `json:"user_id"`
	PromotedToL2     []string `json:"promoted_to_l2"`     // 提升到 L2
	PromotedToL3     []string `json:"promoted_to_l3"`     // 提升到 L3
	Demoted          []string `json:"demoted"`            // 降级
	Forgotten        []string `json:"forgotten"`          // 遗忘
	Merged           []string `json:"merged"`             // 合并
	ConsolidatedAt   time.Time `json:"consolidated_at"`
}

// ForgettingPolicy - 遗忘策略
type ForgettingPolicy struct {
	// 时间衰减
	DecayRate      float64       `json:"decay_rate"`       // 衰减速率
	HalfLife       time.Duration `json:"half_life"`        // 半衰期

	// 访问加强
	AccessBoost    float64       `json:"access_boost"`     // 每次访问的重要性提升

	// 阈值
	ForgettingThreshold float64   `json:"forgetting_threshold"` // 低于此值遗忘

	// 层级保留
	L1MaxAge       time.Duration `json:"l1_max_age"`       // L1 最大保留时间
	L2MaxAge       time.Duration `json:"l2_max_age"`       // L2 最大保留时间
	L3MaxAge       time.Duration `json:"l3_max_age"`       // L3 最大保留时间

	// 容量限制
	L1Capacity     int           `json:"l1_capacity"`      // L1 容量
	L2Capacity     int           `json:"l2_capacity"`      // L2 容量
	L3Capacity     int           `json:"l3_capacity"`      // L3 容量
}

// DefaultForgettingPolicy 默认遗忘策略
func DefaultForgettingPolicy() *ForgettingPolicy {
	return &ForgettingPolicy{
		DecayRate:           0.1,
		HalfLife:            30 * 24 * time.Hour, // 30 天
		AccessBoost:         0.05,
		ForgettingThreshold: 0.1,
		L1MaxAge:            24 * time.Hour,      // 1 天
		L2MaxAge:            7 * 24 * time.Hour,  // 7 天
		L3MaxAge:            365 * 24 * time.Hour, // 1 年
		L1Capacity:          100,
		L2Capacity:          1000,
		L3Capacity:          10000,
	}
}

// === 辅助方法 ===

// NewMemory 创建新记忆
func NewMemory(userID string, memType MemoryType, content string) *Memory {
	now := time.Now()
	return &Memory{
		ID:          generateMemoryID(),
		UserID:      userID,
		Type:        memType,
		Content:     content,
		Importance:  0.5,
		Priority:    5,
		Source:      "manual",
		Timestamp:   now,
		Layer:       MemoryLayerL1,
		DecayFactor: 1.0,
		Tags:        []string{},
		Entities:    []string{},
		RelatedIDs:  []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Access 访问记忆，更新访问统计
func (m *Memory) Access() {
	m.LastAccessed = time.Now()
	m.AccessCount++
	m.DecayFactor = min(1.0, m.DecayFactor+0.1) // 访问加强
	m.UpdatedAt = time.Now()
}

// Decay 衰减记忆
func (m *Memory) Decay(rate float64) {
	m.DecayFactor = m.DecayFactor * (1 - rate)
	m.UpdatedAt = time.Now()
}

// GetEffectiveImportance 获取有效重要性
func (m *Memory) GetEffectiveImportance() float64 {
	return m.Importance * m.DecayFactor
}

// ShouldForget 是否应该遗忘
func (m *Memory) ShouldForget(threshold float64) bool {
	return m.GetEffectiveImportance() < threshold
}

// Promote 提升记忆层级
func (m *Memory) Promote() bool {
	switch m.Layer {
	case MemoryLayerL1:
		m.Layer = MemoryLayerL2
		return true
	case MemoryLayerL2:
		m.Layer = MemoryLayerL3
		return true
	default:
		return false
	}
}

// Demote 降级记忆层级
func (m *Memory) Demote() bool {
	switch m.Layer {
	case MemoryLayerL3:
		m.Layer = MemoryLayerL2
		return true
	case MemoryLayerL2:
		m.Layer = MemoryLayerL1
		return true
	default:
		return false
	}
}

// AddTag 添加标签
func (m *Memory) AddTag(tag string) {
	for _, t := range m.Tags {
		if t == tag {
			return
		}
	}
	m.Tags = append(m.Tags, tag)
	m.UpdatedAt = time.Now()
}

// AddEntity 添加实体
func (m *Memory) AddEntity(entity string) {
	for _, e := range m.Entities {
		if e == entity {
			return
		}
	}
	m.Entities = append(m.Entities, entity)
	m.UpdatedAt = time.Now()
}

// Relate 关联其他记忆
func (m *Memory) Relate(memoryID string) {
	for _, id := range m.RelatedIDs {
		if id == memoryID {
			return
		}
	}
	m.RelatedIDs = append(m.RelatedIDs, memoryID)
	m.UpdatedAt = time.Now()
}

func generateMemoryID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
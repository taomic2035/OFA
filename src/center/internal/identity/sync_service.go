package identity

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// === 同步请求/响应结构 ===

// IdentitySyncRequest - 身份同步请求
type IdentitySyncRequest struct {
	AgentID  string            `json:"agent_id"`
	Identity *models.PersonalIdentity `json:"identity"`
	Version  int64             `json:"version"`    // 版本号
	SyncType string            `json:"sync_type"`  // full/delta
}

// IdentitySyncResponse - 身份同步响应
type IdentitySyncResponse struct {
	Success  bool                    `json:"success"`
	Identity *models.PersonalIdentity `json:"identity"`  // 最新身份
	Version  int64                   `json:"version"`
	Conflict bool                    `json:"conflict"`  // 冲突标记
	Error    string                  `json:"error,omitempty"`
}

// MemorySyncRequest - 记忆同步请求
type MemorySyncRequest struct {
	AgentID  string            `json:"agent_id"`
	Memories []MemoryEntry     `json:"memories"`
	Version  int64             `json:"version"`
}

// MemorySyncResponse - 记忆同步响应
type MemorySyncResponse struct {
	Success  bool          `json:"success"`
	Memories []MemoryEntry `json:"memories"`
	Version  int64         `json:"version"`
	Error    string        `json:"error,omitempty"`
}

// MemoryEntry - 记忆条目
type MemoryEntry struct {
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// PreferenceSyncRequest - 偏好同步请求
type PreferenceSyncRequest struct {
	AgentID     string                 `json:"agent_id"`
	Preferences map[string]interface{} `json:"preferences"`
	Version     int64                  `json:"version"`
}

// PreferenceSyncResponse - 偏好同步响应
type PreferenceSyncResponse struct {
	Success     bool                   `json:"success"`
	Preferences map[string]interface{} `json:"preferences"`
	Version     int64                  `json:"version"`
	Error       string                 `json:"error,omitempty"`
}

// BehaviorReportRequest - 行为上报请求
type BehaviorReportRequest struct {
	AgentID    string                 `json:"agent_id"`
	Type       string                 `json:"type"`
	Context    map[string]interface{} `json:"context"`
	Outcome    string                 `json:"outcome"`
	Timestamp  int64                  `json:"timestamp"`
}

// === 同步服务 ===

// SyncService - 身份同步服务
type SyncService struct {
	mu            sync.RWMutex
	identityStore IdentityStore
	memoryStore   MemorySyncStore
	agentIdentity map[string]string // agentId -> identityId

	// 行为观察缓存（用于性格推断）
	behaviorCache map[string][]models.BehaviorObservation

	// 同步监听器
	syncListeners []SyncListener
}

// SyncListener - 同步事件监听器
type SyncListener interface {
	OnIdentitySynced(agentID string, identity *models.PersonalIdentity)
	OnBehaviorReported(agentID string, behavior *models.BehaviorObservation)
}

// MemorySyncStore - 记忆同步存储接口
type MemorySyncStore interface {
	SaveMemories(ctx context.Context, identityID string, memories []MemoryEntry) error
	GetMemories(ctx context.Context, identityID string, query map[string]interface{}) ([]MemoryEntry, error)
}

// NewSyncService 创建同步服务
func NewSyncService(identityStore IdentityStore) *SyncService {
	return &SyncService{
		identityStore: identityStore,
		agentIdentity: make(map[string]string),
		behaviorCache: make(map[string][]models.BehaviorObservation),
		syncListeners: make([]SyncListener, 0),
	}
}

// === 身份同步 ===

// SyncIdentity 同步身份到 Center
func (s *SyncService) SyncIdentity(ctx context.Context, req *IdentitySyncRequest) (*IdentitySyncResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	identityID := req.Identity.ID
	if identityID == "" {
		return &IdentitySyncResponse{
			Success: false,
			Error:   "identity_id is required",
		}, nil
	}

	// 绑定 agent 到 identity
	s.agentIdentity[req.AgentID] = identityID

	// 获取现有身份
	existing, err := s.identityStore.GetIdentity(ctx, identityID)
	if err != nil {
		// 不存在，直接保存
		if err := s.identityStore.SaveIdentity(ctx, req.Identity); err != nil {
			return &IdentitySyncResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		s.notifySyncListeners(req.AgentID, req.Identity)

		return &IdentitySyncResponse{
			Success:  true,
			Identity: req.Identity,
			Version:  req.Version,
			Conflict: false,
		}, nil
	}

	// 版本冲突检测
	if existing.UpdatedAt.After(req.Identity.UpdatedAt) {
		// 服务端版本更新，返回服务端版本
		return &IdentitySyncResponse{
			Success:  true,
			Identity: existing,
			Version:  int64(existing.UpdatedAt.UnixNano()),
			Conflict: true,
		}, nil
	}

	// 客户端版本更新，保存
	if err := s.identityStore.SaveIdentity(ctx, req.Identity); err != nil {
		return &IdentitySyncResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	s.notifySyncListeners(req.AgentID, req.Identity)

	log.Printf("Identity synced: %s from agent %s", identityID, req.AgentID)

	return &IdentitySyncResponse{
		Success:  true,
		Identity: req.Identity,
		Version:  req.Version,
		Conflict: false,
	}, nil
}

// GetIdentityForAgent 获取 Agent 绑定的身份
func (s *SyncService) GetIdentityForAgent(ctx context.Context, agentID string) (*models.PersonalIdentity, error) {
	s.mu.RLock()
	identityID, ok := s.agentIdentity[agentID]
	s.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	return s.identityStore.GetIdentity(ctx, identityID)
}

// === 记忆同步 ===

// SyncMemories 同步记忆
func (s *SyncService) SyncMemories(ctx context.Context, req *MemorySyncRequest) (*MemorySyncResponse, error) {
	s.mu.RLock()
	identityID, ok := s.agentIdentity[req.AgentID]
	s.mu.RUnlock()

	if !ok {
		return &MemorySyncResponse{
			Success: false,
			Error:   "agent not bound to identity",
		}, nil
	}

	if s.memoryStore == nil {
		return &MemorySyncResponse{
			Success: false,
			Error:   "memory store not configured",
		}, nil
	}

	// 保存记忆
	if err := s.memoryStore.SaveMemories(ctx, identityID, req.Memories); err != nil {
		return &MemorySyncResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取所有记忆
	memories, err := s.memoryStore.GetMemories(ctx, identityID, nil)
	if err != nil {
		return &MemorySyncResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &MemorySyncResponse{
		Success:  true,
		Memories: memories,
		Version:  time.Now().UnixNano(),
	}, nil
}

// === 偏好同步 ===

// SyncPreferences 同步偏好
func (s *SyncService) SyncPreferences(ctx context.Context, req *PreferenceSyncRequest) (*PreferenceSyncResponse, error) {
	s.mu.RLock()
	identityID, ok := s.agentIdentity[req.AgentID]
	s.mu.RUnlock()

	if !ok {
		return &PreferenceSyncResponse{
			Success: false,
			Error:   "agent not bound to identity",
		}, nil
	}

	// 获取身份
	identity, err := s.identityStore.GetIdentity(ctx, identityID)
	if err != nil {
		return &PreferenceSyncResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 更新偏好（简化版，存储在 metadata 中）
	if identity.ValueSystem == nil {
		identity.ValueSystem = &models.ValueSystem{
			CustomValues: make(map[string]float64),
		}
	}

	// 应用偏好更新
	for key, value := range req.Preferences {
		if f, ok := value.(float64); ok {
			switch key {
			case "privacy":
				identity.ValueSystem.Privacy = f
			case "efficiency":
				identity.ValueSystem.Efficiency = f
			case "health":
				identity.ValueSystem.Health = f
			case "family":
				identity.ValueSystem.Family = f
			case "career":
				identity.ValueSystem.Career = f
			}
		}
	}

	identity.UpdatedAt = time.Now()

	// 保存
	if err := s.identityStore.SaveIdentity(ctx, identity); err != nil {
		return &PreferenceSyncResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 返回更新后的偏好
	prefs := make(map[string]interface{})
	prefs["privacy"] = identity.ValueSystem.Privacy
	prefs["efficiency"] = identity.ValueSystem.Efficiency
	prefs["health"] = identity.ValueSystem.Health
	prefs["family"] = identity.ValueSystem.Family
	prefs["career"] = identity.ValueSystem.Career

	return &PreferenceSyncResponse{
		Success:     true,
		Preferences: prefs,
		Version:     identity.UpdatedAt.UnixNano(),
	}, nil
}

// === 行为上报 ===

// ReportBehavior 上报行为观察
func (s *SyncService) ReportBehavior(ctx context.Context, req *BehaviorReportRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建行为观察
	observation := models.BehaviorObservation{
		ID:        generateID(),
		UserID:    req.AgentID,
		Type:      req.Type,
		Context:   req.Context,
		Outcome:   req.Outcome,
		Timestamp: time.Unix(req.Timestamp, 0),
	}

	// 推断性格变化
	observation.Inferences = s.inferPersonalityFromBehavior(req.Type, req.Context)

	// 添加到缓存
	s.behaviorCache[req.AgentID] = append(s.behaviorCache[req.AgentID], observation)

	// 通知监听器
	for _, listener := range s.syncListeners {
		listener.OnBehaviorReported(req.AgentID, &observation)
	}

	log.Printf("Behavior reported: %s from agent %s", req.Type, req.AgentID)

	return nil
}

// inferPersonalityFromBehavior 从行为推断性格
func (s *SyncService) inferPersonalityFromBehavior(behaviorType string, context map[string]interface{}) map[string]float64 {
	inferences := make(map[string]float64)

	switch behaviorType {
	case "decision":
		decisionType, _ := context["decision_type"].(string)
		switch decisionType {
		case "impulse_purchase":
			inferences["neuroticism"] = 0.05
			inferences["conscientiousness"] = -0.03
		case "careful_planning":
			inferences["conscientiousness"] = 0.05
		case "risky_investment":
			inferences["risk_tolerance"] = 0.05
		}

	case "interaction":
		interactionType, _ := context["interaction_type"].(string)
		switch interactionType {
		case "group_chats":
			inferences["extraversion"] = 0.05
		case "private_chats":
			inferences["extraversion"] = -0.03
		case "emoji_heavy":
			inferences["openness"] = 0.03
		}

	case "preference":
		preferenceType, _ := context["preference_type"].(string)
		switch preferenceType {
		case "novel_trying":
			inferences["openness"] = 0.05
		case "routine_following":
			inferences["openness"] = -0.03
			inferences["conscientiousness"] = 0.03
		}

	case "activity":
		activityType, _ := context["activity_type"].(string)
		switch activityType {
		case "exploring_new":
			inferences["openness"] = 0.05
		case "regular_schedule":
			inferences["conscientiousness"] = 0.05
		}
	}

	return inferences
}

// === 监听器管理 ===

// AddSyncListener 添加同步监听器
func (s *SyncService) AddSyncListener(listener SyncListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncListeners = append(s.syncListeners, listener)
}

// notifySyncListeners 通知同步监听器
func (s *SyncService) notifySyncListeners(agentID string, identity *models.PersonalIdentity) {
	for _, listener := range s.syncListeners {
		listener.OnIdentitySynced(agentID, identity)
	}
}

// === 辅助方法 ===

// BindAgentToIdentity 绑定 Agent 到身份
func (s *SyncService) BindAgentToIdentity(agentID, identityID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentIdentity[agentID] = identityID
}

// UnbindAgent 解绑 Agent
func (s *SyncService) UnbindAgent(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agentIdentity, agentID)
	delete(s.behaviorCache, agentID)
}

// GetBehaviorCache 获取行为观察缓存
func (s *SyncService) GetBehaviorCache(agentID string) []models.BehaviorObservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.behaviorCache[agentID]
}

// ClearBehaviorCache 清除行为观察缓存
func (s *SyncService) ClearBehaviorCache(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.behaviorCache, agentID)
}
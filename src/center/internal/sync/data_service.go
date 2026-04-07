package sync

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ofa/center/internal/identity"
	"github.com/ofa/center/internal/models"
)

// === 数据同步服务 ===

// DataService - 数据中心服务 (v2.1.0)
//
// Center 从"控制中心"转变为"数据中心"：
// - 不再主动调度任务
// - Agent 主动拉取/推送数据
// - 提供身份、记忆、偏好的同步服务
type DataService struct {
	mu sync.RWMutex

	// 存储接口
	identityStore   IdentityDataStore
	memoryStore     MemoryDataStore
	preferenceStore PreferenceDataStore

	// 行为存储（内存缓存，用于性格推断）
	behaviorStore map[string][]models.BehaviorObservation // identityId -> observations

	// 性格推断引擎
	inferencer *identity.Inferencer

	// 同步状态
	syncStates map[string]*SyncState // identityId -> syncState

	// 监听器
	listeners []DataSyncListener
}

// SyncState - 同步状态
type SyncState struct {
	IdentityID    string
	LastSyncTime  time.Time
	Version       int64
	DeviceCount   int
	Devices       map[string]*DeviceSyncInfo // agentId -> device info
}

// DeviceSyncInfo - 设备同步信息
type DeviceSyncInfo struct {
	AgentID      string
	DeviceType   string
	LastSyncTime time.Time
	Version      int64
}

// DataSyncListener - 数据同步监听器
type DataSyncListener interface {
	OnIdentitySynced(identityID string, deviceCount int)
	OnMemorySynced(identityID string, entryCount int)
	OnPreferenceSynced(identityID string)
}

// 存储接口
type IdentityDataStore interface {
	GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error)
	SaveIdentity(ctx context.Context, identity *models.PersonalIdentity) error
}

type MemoryDataStore interface {
	GetMemories(ctx context.Context, identityID string, query *MemoryQuery) ([]*MemoryEntry, error)
	SaveMemories(ctx context.Context, identityID string, entries []*MemoryEntry) error
	DeleteMemories(ctx context.Context, identityID string, keys []string) error
}

type PreferenceDataStore interface {
	GetPreferences(ctx context.Context, identityID string) (map[string]interface{}, error)
	SavePreferences(ctx context.Context, identityID string, prefs map[string]interface{}) error
}

// MemoryQuery - 记忆查询
type MemoryQuery struct {
	KeyPattern string
	StartTime  time.Time
	EndTime    time.Time
	Limit      int
}

// MemoryEntry - 记忆条目
type MemoryEntry struct {
	Key       string
	Value     string
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// NewDataService 创建数据服务
func NewDataService() *DataService {
	return &DataService{
		syncStates:    make(map[string]*SyncState),
		listeners:     make([]DataSyncListener, 0),
		behaviorStore: make(map[string][]models.BehaviorObservation),
		inferencer:    identity.NewInferencer(),
	}
}

// SetStores 设置存储
func (s *DataService) SetStores(identity IdentityDataStore, memory MemoryDataStore, preference PreferenceDataStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identityStore = identity
	s.memoryStore = memory
	s.preferenceStore = preference
}

// === 身份数据同步 ===

// SyncIdentityRequest - 身份同步请求
type SyncIdentityRequest struct {
	AgentID    string
	IdentityID string
	Identity   *models.PersonalIdentity
	Version    int64
	SyncType   string // full/delta
}

// SyncIdentityResponse - 身份同步响应
type SyncIdentityResponse struct {
	Success     bool
	Identity    *models.PersonalIdentity
	Version     int64
	Conflict    bool
	ConflictRes string // "local" / "remote" / "merged"
}

// SyncIdentity 同步身份
func (s *DataService) SyncIdentity(ctx context.Context, req *SyncIdentityRequest) (*SyncIdentityResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.identityStore == nil {
		return &SyncIdentityResponse{Success: false}, nil
	}

	identityID := req.IdentityID
	if identityID == "" && req.Identity != nil {
		identityID = req.Identity.ID
	}

	// 获取现有身份
	existing, err := s.identityStore.GetIdentity(ctx, identityID)

	// 更新同步状态
	state := s.getOrCreateSyncState(identityID)
	if _, exists := state.Devices[req.AgentID]; !exists {
		state.DeviceCount++
	}
	state.Devices[req.AgentID] = &DeviceSyncInfo{
		AgentID:      req.AgentID,
		LastSyncTime: time.Now(),
		Version:      req.Version,
	}

	if err != nil || existing == nil {
		// 不存在，直接保存
		if req.Identity != nil {
			if err := s.identityStore.SaveIdentity(ctx, req.Identity); err != nil {
				return &SyncIdentityResponse{Success: false}, err
			}
			state.Version = req.Version
			state.LastSyncTime = time.Now()
		}

		s.notifyIdentitySynced(identityID, state.DeviceCount)

		return &SyncIdentityResponse{
			Success:  true,
			Identity: req.Identity,
			Version:  req.Version,
		}, nil
	}

	// 版本比较
	if req.Version > state.Version {
		// 客户端版本更新，保存
		if req.Identity != nil {
			if err := s.identityStore.SaveIdentity(ctx, req.Identity); err != nil {
				return &SyncIdentityResponse{Success: false}, err
			}
			state.Version = req.Version
			state.LastSyncTime = time.Now()
		}

		s.notifyIdentitySynced(identityID, state.DeviceCount)

		return &SyncIdentityResponse{
			Success:  true,
			Identity: req.Identity,
			Version:  req.Version,
		}, nil
	}

	// 服务端版本更新或相同，返回服务端版本
	return &SyncIdentityResponse{
		Success:  true,
		Identity: existing,
		Version:  state.Version,
		Conflict: req.Version < state.Version,
	}, nil
}

// GetIdentity 获取身份
func (s *DataService) GetIdentity(ctx context.Context, identityID string) (*models.PersonalIdentity, error) {
	if s.identityStore == nil {
		return nil, nil
	}
	return s.identityStore.GetIdentity(ctx, identityID)
}

// === 记忆数据同步 ===

// SyncMemoriesRequest - 记忆同步请求
type SyncMemoriesRequest struct {
	AgentID    string
	IdentityID string
	Memories   []*MemoryEntry
	Version    int64
}

// SyncMemoriesResponse - 记忆同步响应
type SyncMemoriesResponse struct {
	Success  bool
	Memories []*MemoryEntry
	Version  int64
}

// SyncMemories 同步记忆
func (s *DataService) SyncMemories(ctx context.Context, req *SyncMemoriesRequest) (*SyncMemoriesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.memoryStore == nil {
		return &SyncMemoriesResponse{Success: false}, nil
	}

	// 保存上传的记忆
	if len(req.Memories) > 0 {
		if err := s.memoryStore.SaveMemories(ctx, req.IdentityID, req.Memories); err != nil {
			return &SyncMemoriesResponse{Success: false}, err
		}
	}

	// 获取所有记忆
	memories, err := s.memoryStore.GetMemories(ctx, req.IdentityID, nil)
	if err != nil {
		return &SyncMemoriesResponse{Success: false}, err
	}

	// 更新同步状态
	state := s.getOrCreateSyncState(req.IdentityID)
	if _, exists := state.Devices[req.AgentID]; !exists {
		state.DeviceCount++
	}
	state.Devices[req.AgentID] = &DeviceSyncInfo{
		AgentID:      req.AgentID,
		LastSyncTime: time.Now(),
		Version:      req.Version,
	}

	s.notifyMemorySynced(req.IdentityID, len(memories))

	return &SyncMemoriesResponse{
		Success:  true,
		Memories: memories,
		Version:  time.Now().UnixNano(),
	}, nil
}

// GetMemories 获取记忆
func (s *DataService) GetMemories(ctx context.Context, identityID string, query *MemoryQuery) ([]*MemoryEntry, error) {
	if s.memoryStore == nil {
		return nil, nil
	}
	return s.memoryStore.GetMemories(ctx, identityID, query)
}

// === 偏好数据同步 ===

// SyncPreferencesRequest - 偏好同步请求
type SyncPreferencesRequest struct {
	AgentID     string
	IdentityID  string
	Preferences map[string]interface{}
	Version     int64
}

// SyncPreferencesResponse - 偏好同步响应
type SyncPreferencesResponse struct {
	Success     bool
	Preferences map[string]interface{}
	Version     int64
}

// SyncPreferences 同步偏好
func (s *DataService) SyncPreferences(ctx context.Context, req *SyncPreferencesRequest) (*SyncPreferencesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.preferenceStore == nil {
		return &SyncPreferencesResponse{Success: false}, nil
	}

	// 合并偏好
	existing, _ := s.preferenceStore.GetPreferences(ctx, req.IdentityID)
	if existing == nil {
		existing = make(map[string]interface{})
	}

	// 合并策略：时间戳优先
	for key, value := range req.Preferences {
		existing[key] = value
	}

	// 保存
	if err := s.preferenceStore.SavePreferences(ctx, req.IdentityID, existing); err != nil {
		return &SyncPreferencesResponse{Success: false}, err
	}

	// 更新同步状态
	state := s.getOrCreateSyncState(req.IdentityID)
	if _, exists := state.Devices[req.AgentID]; !exists {
		state.DeviceCount++
	}
	state.Devices[req.AgentID] = &DeviceSyncInfo{
		AgentID:      req.AgentID,
		LastSyncTime: time.Now(),
		Version:      req.Version,
	}

	s.notifyPreferenceSynced(req.IdentityID)

	return &SyncPreferencesResponse{
		Success:     true,
		Preferences: existing,
		Version:     time.Now().UnixNano(),
	}, nil
}

// GetPreferences 获取偏好
func (s *DataService) GetPreferences(ctx context.Context, identityID string) (map[string]interface{}, error) {
	if s.preferenceStore == nil {
		return nil, nil
	}
	return s.preferenceStore.GetPreferences(ctx, identityID)
}

// === 行为上报 ===

// BehaviorReport - 行为上报
type BehaviorReport struct {
	ID         string
	AgentID    string
	IdentityID string
	Type       string
	Context    map[string]interface{}
	Inferences map[string]float64
	Timestamp  time.Time
}

// ReportBehavior 上报行为（用于性格推断）
func (s *DataService) ReportBehavior(ctx context.Context, report *BehaviorReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Behavior reported: %s from agent %s", report.Type, report.AgentID)

	// 构建行为观察
	observation := models.BehaviorObservation{
		ID:         report.ID,
		UserID:     report.IdentityID,
		Type:       report.Type,
		Context:    report.Context,
		Inferences: report.Inferences,
		Timestamp:  report.Timestamp,
	}

	// 存储行为观察
	identityID := report.IdentityID
	if identityID == "" {
		identityID = report.AgentID // fallback
	}

	if s.behaviorStore[identityID] == nil {
		s.behaviorStore[identityID] = make([]models.BehaviorObservation, 0)
	}

	// 添加观察（保留最近100条）
	observations := s.behaviorStore[identityID]
	observations = append(observations, observation)
	if len(observations) > 100 {
		observations = observations[len(observations)-100:]
	}
	s.behaviorStore[identityID] = observations

	// 每10条触发一次性格推断
	if len(observations)%10 == 0 && s.identityStore != nil {
		s.triggerPersonalityInference(ctx, identityID, observations)
	}

	return nil
}

// triggerPersonalityInference 触发性格推断
func (s *DataService) triggerPersonalityInference(ctx context.Context, identityID string, observations []models.BehaviorObservation) {
	// 获取现有身份
	existing, err := s.identityStore.GetIdentity(ctx, identityID)
	if err != nil || existing == nil {
		log.Printf("Cannot infer personality: identity not found %s", identityID)
		return
	}

	// 使用推断引擎更新性格
	if existing.Personality == nil {
		existing.Personality = &models.Personality{
			Openness:          0.5,
			Conscientiousness: 0.5,
			Extraversion:      0.5,
			Agreeableness:     0.5,
			Neuroticism:       0.5,
			CustomTraits:      make(map[string]float64),
			SpeakingTone:      "casual",
			ResponseLength:    "moderate",
			EmojiUsage:        0.3,
		}
	}

	// 更新性格
	updatedPersonality := s.inferencer.UpdatePersonalityWithConvergence(existing.Personality, observations)
	existing.Personality = updatedPersonality

	// 更新价值观
	if existing.ValueSystem == nil {
		existing.ValueSystem = &models.ValueSystem{
			Privacy:       0.7,
			Efficiency:    0.6,
			Health:        0.7,
			Family:        0.8,
			Career:        0.6,
			Entertainment: 0.5,
			Learning:      0.6,
			Social:        0.5,
			Finance:       0.6,
			Environment:   0.5,
			RiskTolerance: 0.4,
			Impulsiveness: 0.3,
			Patience:      0.6,
			CustomValues:  make(map[string]float64),
		}
	}

	updatedValueSystem := s.inferencer.InferValueSystemFromBehavior(observations)
	if updatedValueSystem != nil {
		existing.ValueSystem = updatedValueSystem
	}

	// 保存更新
	if err := s.identityStore.SaveIdentity(ctx, existing); err != nil {
		log.Printf("Failed to save inferred personality: %v", err)
		return
	}

	log.Printf("Personality inferred for %s: MBTI=%s, stability=%.2f, observations=%d",
		identityID, updatedPersonality.MBTIType, updatedPersonality.StabilityScore, len(observations))
}

// GetBehaviorObservations 获取行为观察（用于调试）
func (s *DataService) GetBehaviorObservations(identityID string) []models.BehaviorObservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.behaviorStore[identityID]
}

// === 同步状态 ===

// GetSyncState 获取同步状态
func (s *DataService) GetSyncState(identityID string) *SyncState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.syncStates[identityID]
}

// GetDeviceCount 获取设备数量
func (s *DataService) GetDeviceCount(identityID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if state, ok := s.syncStates[identityID]; ok {
		return state.DeviceCount
	}
	return 0
}

// getOrCreateSyncState 获取或创建同步状态
func (s *DataService) getOrCreateSyncState(identityID string) *SyncState {
	if state, ok := s.syncStates[identityID]; ok {
		return state
	}

	state := &SyncState{
		IdentityID: identityID,
		Devices:    make(map[string]*DeviceSyncInfo),
	}
	s.syncStates[identityID] = state
	return state
}

// === 监听器 ===

// AddListener 添加监听器
func (s *DataService) AddListener(listener DataSyncListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

func (s *DataService) notifyIdentitySynced(identityID string, deviceCount int) {
	for _, listener := range s.listeners {
		listener.OnIdentitySynced(identityID, deviceCount)
	}
}

func (s *DataService) notifyMemorySynced(identityID string, entryCount int) {
	for _, listener := range s.listeners {
		listener.OnMemorySynced(identityID, entryCount)
	}
}

func (s *DataService) notifyPreferenceSynced(identityID string) {
	for _, listener := range s.listeners {
		listener.OnPreferenceSynced(identityID)
	}
}
package grpc

import (
	"context"

	"github.com/ofa/center/internal/decision"
	"github.com/ofa/center/internal/identity"
	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/internal/memory"
	"github.com/ofa/center/internal/preference"
	"github.com/ofa/center/internal/voice"
	"github.com/ofa/center/proto"
)

// OFAServer - OFA gRPC 服务实现
type OFAServer struct {
	// 内部服务
	memoryService     *memory.Service
	identityService   *identity.Service
	preferenceService *preference.Service
	decisionEngine    *decision.Engine
	voiceService      *voice.Service

	// 存储
	userStore  UserStore
	sessionStore SessionStore
}

// UserStore - 用户存储接口
type UserStore interface {
	CreateUser(ctx context.Context, user *proto.UserProfile) error
	GetUser(ctx context.Context, userID string) (*proto.UserProfile, error)
	UpdateUser(ctx context.Context, user *proto.UserProfile) error
	DeleteUser(ctx context.Context, userID string) error
}

// SessionStore - 会话存储接口
type SessionStore interface {
	CreateSession(ctx context.Context, session *proto.Session) error
	GetSession(ctx context.Context, sessionID string) (*proto.Session, error)
	UpdateSession(ctx context.Context, session *proto.Session) error
	EndSession(ctx context.Context, sessionID string, summary string) error
	GetActiveSessions(ctx context.Context, userID string) ([]*proto.Session, error)
}

// NewOFAServer 创建 OFA 服务
func NewOFAServer(
	memorySvc *memory.Service,
	identitySvc *identity.Service,
	preferenceSvc *preference.Service,
	decisionEng *decision.Engine,
	voiceSvc *voice.Service,
	userStore UserStore,
	sessionStore SessionStore,
) *OFAServer {
	return &OFAServer{
		memoryService:     memorySvc,
		identityService:   identitySvc,
		preferenceService: preferenceSvc,
		decisionEngine:    decisionEng,
		voiceService:      voiceSvc,
		userStore:         userStore,
		sessionStore:      sessionStore,
	}
}

// === User Service ===

// CreateUser 创建用户
func (s *OFAServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	user := &proto.UserProfile{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Locale:   req.Locale,
		Metadata: req.Metadata,
	}

	if err := s.userStore.CreateUser(ctx, user); err != nil {
		return &proto.CreateUserResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.CreateUserResponse{Success: true, User: user}, nil
}

// GetUser 获取用户
func (s *OFAServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	user, err := s.userStore.GetUser(ctx, req.UserId)
	if err != nil {
		return &proto.GetUserResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetUserResponse{Success: true, User: user}, nil
}

// UpdateUser 更新用户
func (s *OFAServer) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	user, err := s.userStore.GetUser(ctx, req.UserId)
	if err != nil {
		return &proto.UpdateUserResponse{Success: false, Error: err.Error()}, nil
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Locale != "" {
		user.Locale = req.Locale
	}
	if req.Timezone != "" {
		user.Timezone = req.Timezone
	}
	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}

	if err := s.userStore.UpdateUser(ctx, user); err != nil {
		return &proto.UpdateUserResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdateUserResponse{Success: true, User: user}, nil
}

// DeleteUser 删除用户
func (s *OFAServer) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	if err := s.userStore.DeleteUser(ctx, req.UserId); err != nil {
		return &proto.DeleteUserResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeleteUserResponse{Success: true}, nil
}

// === Identity Service ===

// GetIdentity 获取用户身份画像
func (s *OFAServer) GetIdentity(ctx context.Context, req *proto.GetIdentityRequest) (*proto.GetIdentityResponse, error) {
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.GetIdentityResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetIdentityResponse{Success: true, Identity: IdentityToProto(identity)}, nil
}

// UpdatePersonality 更新性格
func (s *OFAServer) UpdatePersonality(ctx context.Context, req *proto.UpdatePersonalityRequest) (*proto.UpdatePersonalityResponse, error) {
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.UpdatePersonalityResponse{Success: false, Error: err.Error()}, nil
	}

	// 转换为 float64 更新
	updates := make(map[string]float64)
	for key, value := range req.Updates {
		if v, ok := value.(float64); ok {
			updates[key] = v
		}
	}

	personality, err := s.identityService.UpdatePersonality(ctx, req.UserId, updates)
	if err != nil {
		return &proto.UpdatePersonalityResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdatePersonalityResponse{Success: true, Personality: PersonalityToProto(personality)}, nil
}

// InferPersonality 推断性格
func (s *OFAServer) InferPersonality(ctx context.Context, req *proto.InferPersonalityRequest) (*proto.InferPersonalityResponse, error) {
	// 转换事件
	observations := make([]models.BehaviorObservation, len(req.Events))
	for i, e := range req.Events {
		observations[i] = models.BehaviorObservation{
			Type:      e.Type,
			Timestamp: e.Timestamp,
			Data:      e.Data,
		}
	}

	personality, err := s.identityService.InferPersonalityFromBehavior(ctx, req.UserId, observations)
	if err != nil {
		return &proto.InferPersonalityResponse{Success: false, Error: err.Error()}, nil
	}

	// 计算变化和置信度
	changes := []string{}
	if personality != nil {
		changes = personality.Tags
	}
	confidence := 0.0
	if personality != nil {
		confidence = personality.StabilityScore
	}

	return &proto.InferPersonalityResponse{
		Success:     true,
		Personality: PersonalityToProto(personality),
		Changes:     changes,
		Confidence:  confidence,
	}, nil
}

// SetValueSystem 设置价值观体系
func (s *OFAServer) SetValueSystem(ctx context.Context, req *proto.SetValueSystemRequest) (*proto.SetValueSystemResponse, error) {
	// 转换为 float64 更新
	updates := make(map[string]float64)
	if req.ValueSystem != nil {
		updates["privacy"] = req.ValueSystem.Privacy
		updates["efficiency"] = req.ValueSystem.Efficiency
		updates["health"] = req.ValueSystem.Health
		updates["family"] = req.ValueSystem.Family
		updates["career"] = req.ValueSystem.Career
		updates["entertainment"] = req.ValueSystem.Entertainment
		updates["learning"] = req.ValueSystem.Learning
		updates["social"] = req.ValueSystem.Social
		updates["finance"] = req.ValueSystem.Finance
		updates["environment"] = req.ValueSystem.Environment
	}

	valueSystem, err := s.identityService.UpdateValueSystem(ctx, req.UserId, updates)
	if err != nil {
		return &proto.SetValueSystemResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SetValueSystemResponse{Success: true, ValueSystem: ValueSystemToProto(valueSystem)}, nil
}

// GetInterests 获取兴趣列表
func (s *OFAServer) GetInterests(ctx context.Context, req *proto.GetInterestsRequest) (*proto.GetInterestsResponse, error) {
	interests, err := s.identityService.GetInterests(ctx, req.UserId)
	if err != nil {
		return &proto.GetInterestsResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetInterestsResponse{Success: true, Interests: InterestsToProto(interests)}, nil
}

// AddInterest 添加兴趣
func (s *OFAServer) AddInterest(ctx context.Context, req *proto.AddInterestRequest) (*proto.AddInterestResponse, error) {
	interest := models.Interest{
		Category: req.Category,
		Name:     req.Name,
		Keywords: req.Keywords,
		Level:    req.Level,
	}

	if err := s.identityService.AddInterest(ctx, req.UserId, interest); err != nil {
		return &proto.AddInterestResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.AddInterestResponse{Success: true, Interest: &proto.Interest{
		Id:       proto.GenerateID(),
		Category: req.Category,
		Name:     req.Name,
		Keywords: req.Keywords,
		Level:    req.Level,
	}}, nil
}

// === Session Service ===

// CreateSession 创建会话
func (s *OFAServer) CreateSession(ctx context.Context, req *proto.CreateSessionRequest) (*proto.CreateSessionResponse, error) {
	session := &proto.Session{
		UserId:       req.UserId,
		AgentId:      req.AgentId,
		Status:       "active",
		Context:      req.Context,
		ActiveMemory: []*proto.Memory{},
	}

	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return &proto.CreateSessionResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.CreateSessionResponse{Success: true, Session: session}, nil
}

// GetSession 获取会话
func (s *OFAServer) GetSession(ctx context.Context, req *proto.GetSessionRequest) (*proto.GetSessionResponse, error) {
	session, err := s.sessionStore.GetSession(ctx, req.SessionId)
	if err != nil {
		return &proto.GetSessionResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetSessionResponse{Success: true, Session: session}, nil
}

// UpdateSessionContext 更新会话上下文
func (s *OFAServer) UpdateSessionContext(ctx context.Context, req *proto.UpdateSessionContextRequest) (*proto.UpdateSessionContextResponse, error) {
	session, err := s.sessionStore.GetSession(ctx, req.SessionId)
	if err != nil {
		return &proto.UpdateSessionContextResponse{Success: false, Error: err.Error()}, nil
	}

	// 合并上下文
	for k, v := range req.Context {
		session.Context[k] = v
	}

	if err := s.sessionStore.UpdateSession(ctx, session); err != nil {
		return &proto.UpdateSessionContextResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdateSessionContextResponse{Success: true, Session: session}, nil
}

// EndSession 结束会话
func (s *OFAServer) EndSession(ctx context.Context, req *proto.EndSessionRequest) (*proto.EndSessionResponse, error) {
	session, err := s.sessionStore.EndSession(ctx, req.SessionId, req.Summary)
	if err != nil {
		return &proto.EndSessionResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.EndSessionResponse{Success: true, Session: session}, nil
}

// GetActiveSessions 获取活跃会话
func (s *OFAServer) GetActiveSessions(ctx context.Context, req *proto.GetActiveSessionsRequest) (*proto.GetActiveSessionsResponse, error) {
	sessions, err := s.sessionStore.GetActiveSessions(ctx, req.UserId)
	if err != nil {
		return &proto.GetActiveSessionsResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetActiveSessionsResponse{Success: true, Sessions: sessions}, nil
}

// === Memory Service ===

// StoreMemory 存储记忆
func (s *OFAServer) StoreMemory(ctx context.Context, req *proto.StoreMemoryRequest) (*proto.StoreMemoryResponse, error) {
	memType := proto.MemoryType(req.Type)
	mem, err := s.memoryService.Store(ctx, req.UserId, memType, req.Content, req.Importance, req.Metadata)
	if err != nil {
		return &proto.StoreMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.StoreMemoryResponse{Success: true, Memory: mem}, nil
}

// RecallMemory 回忆记忆
func (s *OFAServer) RecallMemory(ctx context.Context, req *proto.RecallMemoryRequest) (*proto.RecallMemoryResponse, error) {
	memType := proto.MemoryType(req.Type)
	layer := proto.MemoryLayer(req.Layer)

	memories, err := s.memoryService.Recall(ctx, req.UserId, req.Query, memType, layer, int(req.Limit), req.Threshold)
	if err != nil {
		return &proto.RecallMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.RecallMemoryResponse{
		Success:  true,
		Memories: memories,
		Total:    int32(len(memories)),
	}, nil
}

// GetMemory 获取单个记忆
func (s *OFAServer) GetMemory(ctx context.Context, req *proto.GetMemoryRequest) (*proto.GetMemoryResponse, error) {
	mem, err := s.memoryService.Get(ctx, req.MemoryId)
	if err != nil {
		return &proto.GetMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetMemoryResponse{Success: true, Memory: mem}, nil
}

// DeleteMemory 删除记忆
func (s *OFAServer) DeleteMemory(ctx context.Context, req *proto.DeleteMemoryRequest) (*proto.DeleteMemoryResponse, error) {
	if err := s.memoryService.Delete(ctx, req.MemoryId); err != nil {
		return &proto.DeleteMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeleteMemoryResponse{Success: true}, nil
}

// ListMemories 列出记忆
func (s *OFAServer) ListMemories(ctx context.Context, req *proto.ListMemoriesRequest) (*proto.ListMemoriesResponse, error) {
	memType := proto.MemoryType(req.Type)
	layer := proto.MemoryLayer(req.Layer)

	memories, total, err := s.memoryService.List(ctx, req.UserId, memType, layer, int(req.Limit), int(req.Offset))
	if err != nil {
		return &proto.ListMemoriesResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ListMemoriesResponse{
		Success:  true,
		Memories: memories,
		Total:    int32(total),
	}, nil
}

// ConsolidateMemory 整合记忆
func (s *OFAServer) ConsolidateMemory(ctx context.Context, req *proto.ConsolidateMemoryRequest) (*proto.ConsolidateMemoryResponse, error) {
	consolidated, promoted, err := s.memoryService.Consolidate(ctx, req.UserId)
	if err != nil {
		return &proto.ConsolidateMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ConsolidateMemoryResponse{
		Success:           true,
		ConsolidatedCount: int32(consolidated),
		PromotedCount:     int32(promoted),
	}, nil
}

// === Preference Service ===

// GetPreference 获取偏好
func (s *OFAServer) GetPreference(ctx context.Context, req *proto.GetPreferenceRequest) (*proto.GetPreferenceResponse, error) {
	pref, err := s.preferenceService.Get(ctx, req.UserId, req.Key)
	if err != nil {
		return &proto.GetPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetPreferenceResponse{Success: true, Preference: pref}, nil
}

// SetPreference 设置偏好
func (s *OFAServer) SetPreference(ctx context.Context, req *proto.SetPreferenceRequest) (*proto.SetPreferenceResponse, error) {
	pref, err := s.preferenceService.Set(ctx, req.UserId, req.Key, req.Value, req.Confidence, req.Source, req.ExpiresAt)
	if err != nil {
		return &proto.SetPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SetPreferenceResponse{Success: true, Preference: pref}, nil
}

// DeletePreference 删除偏好
func (s *OFAServer) DeletePreference(ctx context.Context, req *proto.DeletePreferenceRequest) (*proto.DeletePreferenceResponse, error) {
	if err := s.preferenceService.Delete(ctx, req.UserId, req.Key); err != nil {
		return &proto.DeletePreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeletePreferenceResponse{Success: true}, nil
}

// GetAllPreferences 获取所有偏好
func (s *OFAServer) GetAllPreferences(ctx context.Context, req *proto.GetAllPreferencesRequest) (*proto.GetAllPreferencesResponse, error) {
	prefs, err := s.preferenceService.GetAll(ctx, req.UserId)
	if err != nil {
		return &proto.GetAllPreferencesResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetAllPreferencesResponse{Success: true, Preferences: prefs}, nil
}

// LearnPreference 学习偏好
func (s *OFAServer) LearnPreference(ctx context.Context, req *proto.LearnPreferenceRequest) (*proto.LearnPreferenceResponse, error) {
	learned, updated, err := s.preferenceService.Learn(ctx, req.UserId, req.EventType, req.EventData)
	if err != nil {
		return &proto.LearnPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.LearnPreferenceResponse{
		Success:      true,
		LearnedPrefs: learned,
		UpdatedPrefs: updated,
	}, nil
}

// GetPreferenceScore 获取偏好评分
func (s *OFAServer) GetPreferenceScore(ctx context.Context, req *proto.GetPreferenceScoreRequest) (*proto.GetPreferenceScoreResponse, error) {
	score, details, err := s.preferenceService.ScoreItem(ctx, req.UserId, req.Item)
	if err != nil {
		return &proto.GetPreferenceScoreResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetPreferenceScoreResponse{
		Success: true,
		Score:   score,
		Details: details,
	}, nil
}

// === Decision Service ===

// Decide 执行决策
func (s *OFAServer) Decide(ctx context.Context, req *proto.DecideRequest) (*proto.DecideResponse, error) {
	// 构建决策上下文
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.DecideResponse{Success: false, Error: err.Error()}, nil
	}

	prefs, err := s.preferenceService.GetAll(ctx, req.UserId)
	if err != nil {
		prefs = []*proto.Preference{}
	}

	decisionCtx := &proto.DecisionContext{
		UserId:            req.UserId,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		ActivePreferences: proto.PreferencesToMap(prefs),
	}

	// 转换选项
	options := make([]proto.DecisionOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = *o
	}

	result, err := s.decisionEngine.Decide(ctx, decisionCtx, req.Scenario, options, req.Context)
	if err != nil {
		return &proto.DecideResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DecideResponse{Success: true, Result: result}, nil
}

// QuickDecide 快速决策
func (s *OFAServer) QuickDecide(ctx context.Context, req *proto.QuickDecideRequest) (*proto.QuickDecideResponse, error) {
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.QuickDecideResponse{Success: false, Error: err.Error()}, nil
	}

	prefs, err := s.preferenceService.GetAll(ctx, req.UserId)
	if err != nil {
		prefs = []*proto.Preference{}
	}

	decisionCtx := &proto.DecisionContext{
		UserId:            req.UserId,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		ActivePreferences: proto.PreferencesToMap(prefs),
	}

	options := make([]proto.DecisionOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = *o
	}

	result, err := s.decisionEngine.QuickDecide(ctx, decisionCtx, req.Scenario, options)
	if err != nil {
		return &proto.QuickDecideResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.QuickDecideResponse{Success: true, Result: result}, nil
}

// ConfirmDecision 确认决策
func (s *OFAServer) ConfirmDecision(ctx context.Context, req *proto.ConfirmDecisionRequest) (*proto.ConfirmDecisionResponse, error) {
	d, err := s.decisionEngine.ConfirmDecision(ctx, req.DecisionId, int(req.OptionIndex))
	if err != nil {
		return &proto.ConfirmDecisionResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ConfirmDecisionResponse{Success: true, Decision: d}, nil
}

// RecordOutcome 记录决策结果
func (s *OFAServer) RecordOutcome(ctx context.Context, req *proto.RecordOutcomeRequest) (*proto.RecordOutcomeResponse, error) {
	d, err := s.decisionEngine.RecordOutcome(ctx, req.DecisionId, req.Outcome, req.Feedback, req.Score)
	if err != nil {
		return &proto.RecordOutcomeResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.RecordOutcomeResponse{Success: true, Decision: d}, nil
}

// GetDecisionHistory 获取决策历史
func (s *OFAServer) GetDecisionHistory(ctx context.Context, req *proto.GetDecisionHistoryRequest) (*proto.GetDecisionHistoryResponse, error) {
	query := &proto.DecisionQuery{
		UserId:     req.UserId,
		Scenario:   req.Scenario,
		Outcome:    req.Outcome,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	decisions, total, err := s.decisionEngine.GetHistory(ctx, query)
	if err != nil {
		return &proto.GetDecisionHistoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetDecisionHistoryResponse{
		Success:   true,
		Decisions: decisions,
		Total:     int32(total),
	}, nil
}

// GetDecisionStats 获取决策统计
func (s *OFAServer) GetDecisionStats(ctx context.Context, req *proto.GetDecisionStatsRequest) (*proto.GetDecisionStatsResponse, error) {
	stats, err := s.decisionEngine.GetStats(ctx, req.UserId)
	if err != nil {
		return &proto.GetDecisionStatsResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetDecisionStatsResponse{Success: true, Stats: stats}, nil
}

// === Voice Service ===

// GetVoiceProfile 获取语音配置
func (s *OFAServer) GetVoiceProfile(ctx context.Context, req *proto.GetVoiceProfileRequest) (*proto.GetVoiceProfileResponse, error) {
	profile, err := s.voiceService.GetProfile(ctx, req.UserId)
	if err != nil {
		return &proto.GetVoiceProfileResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetVoiceProfileResponse{Success: true, VoiceProfile: profile}, nil
}

// UpdateVoiceProfile 更新语音配置
func (s *OFAServer) UpdateVoiceProfile(ctx context.Context, req *proto.UpdateVoiceProfileRequest) (*proto.UpdateVoiceProfileResponse, error) {
	profile, err := s.voiceService.GetProfile(ctx, req.UserId)
	if err != nil {
		// 创建新配置
		profile = &proto.VoiceProfile{UserId: req.UserId}
	}

	if req.DefaultTtsVoice != "" {
		profile.DefaultTtsVoice = req.DefaultTtsVoice
	}
	if req.SpeechRate > 0 {
		profile.SpeechRate = req.SpeechRate
	}
	if req.SpeechPitch > 0 {
		profile.SpeechPitch = req.SpeechPitch
	}
	if req.Volume > 0 {
		profile.Volume = req.Volume
	}
	if req.PreferredLanguage != "" {
		profile.PreferredLanguage = req.PreferredLanguage
	}

	if err := s.voiceService.SaveProfile(ctx, profile); err != nil {
		return &proto.UpdateVoiceProfileResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdateVoiceProfileResponse{Success: true, VoiceProfile: profile}, nil
}

// SynthesizeSpeech 语音合成
func (s *OFAServer) SynthesizeSpeech(ctx context.Context, req *proto.SynthesizeSpeechRequest) (*proto.SynthesizeSpeechResponse, error) {
	audio, duration, err := s.voiceService.Synthesize(ctx, req.UserId, req.Text, req.VoiceId, req.Format)
	if err != nil {
		return &proto.SynthesizeSpeechResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SynthesizeSpeechResponse{
		Success:    true,
		AudioData:  audio,
		DurationMs: duration,
		Format:     req.Format,
	}, nil
}

// CloneVoice 克隆语音
func (s *OFAServer) CloneVoice(ctx context.Context, req *proto.CloneVoiceRequest) (*proto.CloneVoiceResponse, error) {
	voice, err := s.voiceService.CloneVoice(ctx, req.UserId, req.SampleData, req.SampleFormat, req.Name)
	if err != nil {
		return &proto.CloneVoiceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.CloneVoiceResponse{Success: true, Voice: voice}, nil
}

// RecognizeSpeech 语音识别
func (s *OFAServer) RecognizeSpeech(ctx context.Context, req *proto.RecognizeSpeechRequest) (*proto.RecognizeSpeechResponse, error) {
	text, confidence, lang, err := s.voiceService.Recognize(ctx, req.UserId, req.AudioData, req.AudioFormat, req.Language)
	if err != nil {
		return &proto.RecognizeSpeechResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.RecognizeSpeechResponse{
		Success:          true,
		Transcript:       text,
		Confidence:       confidence,
		DetectedLanguage: lang,
	}, nil
}

// === Context Sync ===

// SyncContext 同步上下文
func (s *OFAServer) SyncContext(ctx context.Context, req *proto.SyncContextRequest) (*proto.SyncContextResponse, error) {
	session, err := s.sessionStore.GetSession(ctx, req.SessionId)
	if err != nil {
		return &proto.SyncContextResponse{Success: false, Error: err.Error()}, nil
	}

	syncedKeys := []string{}
	for k, v := range req.Changes {
		session.Context[k] = v
		syncedKeys = append(syncedKeys, k)
	}

	if err := s.sessionStore.UpdateSession(ctx, session); err != nil {
		return &proto.SyncContextResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SyncContextResponse{
		Success:        true,
		SyncedKeys:     syncedKeys,
		CurrentContext: session.Context,
	}, nil
}

// GetFullContext 获取完整上下文
func (s *OFAServer) GetFullContext(ctx context.Context, req *proto.GetFullContextRequest) (*proto.GetFullContextResponse, error) {
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.GetFullContextResponse{Success: false, Error: err.Error()}, nil
	}

	prefs, err := s.preferenceService.GetAll(ctx, req.UserId)
	if err != nil {
		prefs = []*proto.Preference{}
	}

	// 获取最近决策
	decisions, _, err := s.decisionEngine.GetHistory(ctx, &proto.DecisionQuery{
		UserId: req.UserId,
		Limit:  10,
	})
	if err != nil {
		decisions = []*proto.Decision{}
	}

	decisionCtx := &proto.DecisionContext{
		UserId:            req.UserId,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		SpeakingTone:      identity.SpeakingTone,
		ResponseLength:    identity.ResponseLength,
		ValuePriority:     identity.ValuePriority,
		RecentDecisions:   decisions,
		ActivePreferences: proto.PreferencesToMap(prefs),
	}

	return &proto.GetFullContextResponse{
		Success: true,
		Context: decisionCtx,
	}, nil
}
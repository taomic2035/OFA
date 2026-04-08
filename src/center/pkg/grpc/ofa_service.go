package grpc

import (
	"context"
	"time"

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
	EndSession(ctx context.Context, sessionID string, summary string) (*proto.Session, error)
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
	_, err := s.identityService.GetIdentity(ctx, req.UserId)
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
			Context:   e.Data,
			Timestamp: time.Unix(e.Timestamp, 0),
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
	memType := models.MemoryType(req.Type)
	mem, err := s.memoryService.Remember(ctx, req.UserId, memType, req.Content,
		memory.WithImportance(req.Importance))
	if err != nil {
		return &proto.StoreMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.StoreMemoryResponse{Success: true, Memory: MemoryToProto(mem)}, nil
}

// RecallMemory 回忆记忆
func (s *OFAServer) RecallMemory(ctx context.Context, req *proto.RecallMemoryRequest) (*proto.RecallMemoryResponse, error) {
	query := &models.MemoryQuery{
		UserID:    req.UserId,
		Keywords:  req.Query,
		Limit:     int(req.Limit),
	}

	if req.Type != "" {
		query.Types = []models.MemoryType{models.MemoryType(req.Type)}
	}
	if req.Layer != "" {
		query.Layer = models.MemoryLayer(req.Layer)
	}
	if req.Threshold > 0 {
		query.MinImportance = req.Threshold
	}

	result, err := s.memoryService.Recall(ctx, query)
	if err != nil {
		return &proto.RecallMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.RecallMemoryResponse{
		Success:  true,
		Memories: MemoriesToProto(result.Memories),
		Total:    int32(result.Total),
	}, nil
}

// GetMemory 获取单个记忆
func (s *OFAServer) GetMemory(ctx context.Context, req *proto.GetMemoryRequest) (*proto.GetMemoryResponse, error) {
	mem, err := s.memoryService.RecallByID(ctx, req.MemoryId)
	if err != nil {
		return &proto.GetMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetMemoryResponse{Success: true, Memory: MemoryToProto(mem)}, nil
}

// DeleteMemory 删除记忆
func (s *OFAServer) DeleteMemory(ctx context.Context, req *proto.DeleteMemoryRequest) (*proto.DeleteMemoryResponse, error) {
	if err := s.memoryService.Forget(ctx, req.MemoryId); err != nil {
		return &proto.DeleteMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeleteMemoryResponse{Success: true}, nil
}

// ListMemories 列出记忆
func (s *OFAServer) ListMemories(ctx context.Context, req *proto.ListMemoriesRequest) (*proto.ListMemoriesResponse, error) {
	query := &models.MemoryQuery{
		UserID: req.UserId,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	if req.Type != "" {
		query.Types = []models.MemoryType{models.MemoryType(req.Type)}
	}
	if req.Layer != "" {
		query.Layer = models.MemoryLayer(req.Layer)
	}

	result, err := s.memoryService.Recall(ctx, query)
	if err != nil {
		return &proto.ListMemoriesResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ListMemoriesResponse{
		Success:  true,
		Memories: MemoriesToProto(result.Memories),
		Total:    int32(result.Total),
	}, nil
}

// ConsolidateMemory 整合记忆
func (s *OFAServer) ConsolidateMemory(ctx context.Context, req *proto.ConsolidateMemoryRequest) (*proto.ConsolidateMemoryResponse, error) {
	result, err := s.memoryService.Consolidate(ctx, req.UserId)
	if err != nil {
		return &proto.ConsolidateMemoryResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ConsolidateMemoryResponse{
		Success:           true,
		ConsolidatedCount: int32(len(result.PromotedToL2) + len(result.PromotedToL3) + len(result.Forgotten)),
		PromotedCount:     int32(len(result.PromotedToL2) + len(result.PromotedToL3)),
	}, nil
}

// === Preference Service ===

// GetPreference 获取偏好
func (s *OFAServer) GetPreference(ctx context.Context, req *proto.GetPreferenceRequest) (*proto.GetPreferenceResponse, error) {
	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Keys:   []string{req.Key},
		Limit:  1,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil || len(prefs) == 0 {
		return &proto.GetPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetPreferenceResponse{Success: true, Preference: PreferenceToProto(prefs[0])}, nil
}

// SetPreference 设置偏好
func (s *OFAServer) SetPreference(ctx context.Context, req *proto.SetPreferenceRequest) (*proto.SetPreferenceResponse, error) {
	pref, err := s.preferenceService.SetPreference(ctx, req.UserId, "", req.Key, req.Value)
	if err != nil {
		return &proto.SetPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SetPreferenceResponse{Success: true, Preference: PreferenceToProto(pref)}, nil
}

// DeletePreference 删除偏好
func (s *OFAServer) DeletePreference(ctx context.Context, req *proto.DeletePreferenceRequest) (*proto.DeletePreferenceResponse, error) {
	// First get the preference to get its ID
	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Keys:   []string{req.Key},
		Limit:  1,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil || len(prefs) == 0 {
		return &proto.DeletePreferenceResponse{Success: false, Error: "preference not found"}, nil
	}

	if err := s.preferenceService.DeletePreference(ctx, prefs[0].ID); err != nil {
		return &proto.DeletePreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DeletePreferenceResponse{Success: true}, nil
}

// GetAllPreferences 获取所有偏好
func (s *OFAServer) GetAllPreferences(ctx context.Context, req *proto.GetAllPreferencesRequest) (*proto.GetAllPreferencesResponse, error) {
	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Limit:  1000,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil {
		return &proto.GetAllPreferencesResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetAllPreferencesResponse{Success: true, Preferences: PreferencesToProto(prefs)}, nil
}

// LearnPreference 学习偏好
func (s *OFAServer) LearnPreference(ctx context.Context, req *proto.LearnPreferenceRequest) (*proto.LearnPreferenceResponse, error) {
	event := &models.PreferenceLearningEvent{
		UserID:    req.UserId,
		Type:      req.EventType,
		Context:   req.EventData,
		Timestamp: time.Now(),
	}
	pref, err := s.preferenceService.LearnFromEvent(ctx, event)
	if err != nil {
		return &proto.LearnPreferenceResponse{Success: false, Error: err.Error()}, nil
	}

	learned := []*proto.Preference{}
	if pref != nil {
		learned = []*proto.Preference{PreferenceToProto(pref)}
	}

	return &proto.LearnPreferenceResponse{
		Success:      true,
		LearnedPrefs: learned,
	}, nil
}

// GetPreferenceScore 获取偏好评分
func (s *OFAServer) GetPreferenceScore(ctx context.Context, req *proto.GetPreferenceScoreRequest) (*proto.GetPreferenceScoreResponse, error) {
	// 简单实现：获取用户偏好并计算分数
	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Limit:  100,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil {
		return &proto.GetPreferenceScoreResponse{Success: false, Error: err.Error()}, nil
	}

	// 简单评分逻辑
	score := 0.5
	if len(prefs) > 10 {
		score = 0.8
	} else if len(prefs) > 5 {
		score = 0.6
	}

	return &proto.GetPreferenceScoreResponse{
		Success: true,
		Score:   score,
		Details: map[string]float64{"preference_count": float64(len(prefs))},
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

	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Limit:  100,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil {
		prefs = []*models.Preference{}
	}

	decisionCtx := &models.DecisionContext{
		UserID:            req.UserId,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		ActivePreferences: make(map[string]interface{}),
	}
	for _, p := range prefs {
		decisionCtx.ActivePreferences[p.Key] = p.Value
	}

	// 转换选项
	options := make([]models.DecisionOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = models.DecisionOption{
			ID:          o.Id,
			Name:        o.Name,
			Description: o.Description,
			Attributes:  o.Attributes,
			Score:       o.Score,
			Pros:        o.Pros,
			Cons:        o.Cons,
		}
	}

	result, err := s.decisionEngine.Decide(ctx, decisionCtx, req.Scenario, options, req.Context)
	if err != nil {
		return &proto.DecideResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.DecideResponse{Success: true, Result: DecisionResultToProto(result)}, nil
}

// QuickDecide 快速决策
func (s *OFAServer) QuickDecide(ctx context.Context, req *proto.QuickDecideRequest) (*proto.QuickDecideResponse, error) {
	identity, err := s.identityService.GetIdentity(ctx, req.UserId)
	if err != nil {
		return &proto.QuickDecideResponse{Success: false, Error: err.Error()}, nil
	}

	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Limit:  100,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil {
		prefs = []*models.Preference{}
	}

	decisionCtx := &models.DecisionContext{
		UserID:            req.UserId,
		Personality:       identity.Personality,
		ValueSystem:       identity.ValueSystem,
		Interests:         identity.Interests,
		ActivePreferences: make(map[string]interface{}),
	}
	for _, p := range prefs {
		decisionCtx.ActivePreferences[p.Key] = p.Value
	}

	options := make([]models.DecisionOption, len(req.Options))
	for i, o := range req.Options {
		options[i] = models.DecisionOption{
			ID:          o.Id,
			Name:        o.Name,
			Description: o.Description,
			Attributes:  o.Attributes,
			Score:       o.Score,
			Pros:        o.Pros,
			Cons:        o.Cons,
		}
	}

	result, err := s.decisionEngine.QuickDecide(ctx, decisionCtx, req.Scenario, options)
	if err != nil {
		return &proto.QuickDecideResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.QuickDecideResponse{Success: true, Result: DecisionResultToProto(result)}, nil
}

// ConfirmDecision 确认决策
func (s *OFAServer) ConfirmDecision(ctx context.Context, req *proto.ConfirmDecisionRequest) (*proto.ConfirmDecisionResponse, error) {
	d, err := s.decisionEngine.ConfirmDecision(ctx, req.DecisionId, int(req.OptionIndex))
	if err != nil {
		return &proto.ConfirmDecisionResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.ConfirmDecisionResponse{Success: true, Decision: DecisionToProto(d)}, nil
}

// RecordOutcome 记录决策结果
func (s *OFAServer) RecordOutcome(ctx context.Context, req *proto.RecordOutcomeRequest) (*proto.RecordOutcomeResponse, error) {
	d, err := s.decisionEngine.RecordOutcome(ctx, req.DecisionId, req.Outcome, req.Feedback, req.Score)
	if err != nil {
		return &proto.RecordOutcomeResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.RecordOutcomeResponse{Success: true, Decision: DecisionToProto(d)}, nil
}

// GetDecisionHistory 获取决策历史
func (s *OFAServer) GetDecisionHistory(ctx context.Context, req *proto.GetDecisionHistoryRequest) (*proto.GetDecisionHistoryResponse, error) {
	query := &models.DecisionQuery{
		UserID:   req.UserId,
		Scenario: req.Scenario,
		Outcome:  req.Outcome,
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
	}

	if req.StartTime > 0 {
		t := time.Unix(req.StartTime, 0)
		query.StartTime = &t
	}
	if req.EndTime > 0 {
		t := time.Unix(req.EndTime, 0)
		query.EndTime = &t
	}

	decisions, total, err := s.decisionEngine.GetHistory(ctx, query)
	if err != nil {
		return &proto.GetDecisionHistoryResponse{Success: false, Error: err.Error()}, nil
	}

	protoDecisions := make([]*proto.Decision, len(decisions))
	for i, d := range decisions {
		protoDecisions[i] = DecisionToProto(d)
	}

	return &proto.GetDecisionHistoryResponse{
		Success:   true,
		Decisions: protoDecisions,
		Total:     int32(total),
	}, nil
}

// GetDecisionStats 获取决策统计
func (s *OFAServer) GetDecisionStats(ctx context.Context, req *proto.GetDecisionStatsRequest) (*proto.GetDecisionStatsResponse, error) {
	stats, err := s.decisionEngine.GetStats(ctx, req.UserId)
	if err != nil {
		return &proto.GetDecisionStatsResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetDecisionStatsResponse{Success: true, Stats: DecisionStatsToProto(stats)}, nil
}

// DecisionStatsToProto converts models.DecisionStats to proto.DecisionStats
func DecisionStatsToProto(s *models.DecisionStats) *proto.DecisionStats {
	if s == nil {
		return nil
	}
	return &proto.DecisionStats{
		UserId:           s.UserID,
		TotalDecisions:   int32(s.TotalDecisions),
		AutoDecisions:    int32(s.AutoDecisions),
		ManualDecisions:  int32(s.ManualDecisions),
		SatisfiedCount:   int32(s.SatisfiedCount),
		UnsatisfiedCount: int32(s.UnsatisfiedCount),
		AvgOutcomeScore:  s.AvgOutcomeScore,
		CountByScenario:  intMapToInt32Map(s.CountByScenario),
		TopScenarios:     s.TopScenarios,
	}
}

// intMapToInt32Map converts map[string]int to map[string]int32
func intMapToInt32Map(m map[string]int) map[string]int32 {
	result := make(map[string]int32)
	for k, v := range m {
		result[k] = int32(v)
	}
	return result
}

// === Voice Service ===

// GetVoiceProfile 获取语音配置
func (s *OFAServer) GetVoiceProfile(ctx context.Context, req *proto.GetVoiceProfileRequest) (*proto.GetVoiceProfileResponse, error) {
	profile, err := s.voiceService.GetVoiceProfile(ctx, req.UserId)
	if err != nil {
		return &proto.GetVoiceProfileResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.GetVoiceProfileResponse{Success: true, VoiceProfile: VoiceProfileToProto(profile)}, nil
}

// UpdateVoiceProfile 更新语音配置
func (s *OFAServer) UpdateVoiceProfile(ctx context.Context, req *proto.UpdateVoiceProfileRequest) (*proto.UpdateVoiceProfileResponse, error) {
	updates := make(map[string]interface{})
	if req.DefaultTtsVoice != "" {
		updates["tts_config.voice_model_id"] = req.DefaultTtsVoice
	}
	if req.SpeechRate > 0 {
		updates["voice_characteristics.speaking_rate"] = req.SpeechRate
	}
	if req.SpeechPitch > 0 {
		updates["voice_characteristics.base_pitch"] = req.SpeechPitch
	}
	if req.Volume > 0 {
		updates["voice_characteristics.base_volume"] = req.Volume
	}
	if req.PreferredLanguage != "" {
		updates["tts_config.language"] = req.PreferredLanguage
	}

	profile, err := s.voiceService.UpdateVoiceProfile(ctx, req.UserId, updates)
	if err != nil {
		return &proto.UpdateVoiceProfileResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.UpdateVoiceProfileResponse{Success: true, VoiceProfile: VoiceProfileToProto(profile)}, nil
}

// VoiceProfileToProto converts models.VoiceProfile to proto.VoiceProfile
func VoiceProfileToProto(v *models.VoiceProfile) *proto.VoiceProfile {
	if v == nil {
		return nil
	}
	return &proto.VoiceProfile{
		Id:            v.IdentityID,
		VoiceType:     v.VoiceCharacteristics.VoiceType,
		Pitch:         v.VoiceCharacteristics.BasePitch,
		Speed:         v.VoiceCharacteristics.SpeakingRate,
		Volume:        v.VoiceCharacteristics.BaseVolume,
		Tone:          v.VoiceCharacteristics.TimbreType,
		CreatedAt:     v.CreatedAt.Unix(),
		UpdatedAt:     v.UpdatedAt.Unix(),
	}
}

// SynthesizeSpeech 语音合成
func (s *OFAServer) SynthesizeSpeech(ctx context.Context, req *proto.SynthesizeSpeechRequest) (*proto.SynthesizeSpeechResponse, error) {
	// Get voice profile
	profile, err := s.voiceService.GetVoiceProfile(ctx, req.UserId)
	if err != nil {
		profile = models.NewVoiceProfile()
		profile.IdentityID = req.UserId
	}

	audio, err := s.voiceService.Synthesize(ctx, req.Text, profile)
	if err != nil {
		return &proto.SynthesizeSpeechResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SynthesizeSpeechResponse{
		Success:   true,
		AudioData: audio,
		Format:    req.Format,
	}, nil
}

// CloneVoice 克隆语音
func (s *OFAServer) CloneVoice(ctx context.Context, req *proto.CloneVoiceRequest) (*proto.CloneVoiceResponse, error) {
	voiceID, err := s.voiceService.CloneVoice(ctx, [][]byte{req.SampleData}, req.Name)
	if err != nil {
		return &proto.CloneVoiceResponse{Success: false, Error: err.Error()}, nil
	}

	// Create a basic voice profile for the response
	voiceProfile := &proto.VoiceProfile{
		Id:         voiceID,
		VoiceType:  "cloned",
	}

	return &proto.CloneVoiceResponse{Success: true, Voice: voiceProfile}, nil
}

// RecognizeSpeech 语音识别 - placeholder implementation
func (s *OFAServer) RecognizeSpeech(ctx context.Context, req *proto.RecognizeSpeechRequest) (*proto.RecognizeSpeechResponse, error) {
	// Voice service doesn't have Recognize method, return placeholder
	return &proto.RecognizeSpeechResponse{
		Success:          false,
		Error:            "speech recognition not implemented",
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

	query := &models.PreferenceQuery{
		UserID: req.UserId,
		Limit:  100,
	}
	prefs, _, err := s.preferenceService.GetPreferences(ctx, query)
	if err != nil {
		prefs = []*models.Preference{}
	}

	// 获取最近决策
	decisionQuery := &models.DecisionQuery{
		UserID: req.UserId,
		Limit:  10,
	}
	decisions, _, err := s.decisionEngine.GetHistory(ctx, decisionQuery)
	if err != nil {
		decisions = []*models.Decision{}
	}

	// Convert decisions to proto
	protoDecisions := make([]*proto.Decision, len(decisions))
	for i, d := range decisions {
		protoDecisions[i] = DecisionToProto(d)
	}

	// Convert preferences to map
	activePrefs := make(map[string]interface{})
	for _, p := range prefs {
		activePrefs[p.Key] = p.Value
	}

	decisionCtx := &proto.DecisionContext{
		UserId:            req.UserId,
		Personality:       PersonalityToProto(identity.Personality),
		ValueSystem:       ValueSystemToProto(identity.ValueSystem),
		Interests:         InterestsToProto(identity.Interests),
		RecentDecisions:   protoDecisions,
		ActivePreferences: activePrefs,
	}

	return &proto.GetFullContextResponse{
		Success: true,
		Context: decisionCtx,
	}, nil
}
package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// Service - 身份服务
type Service struct {
	mu       sync.RWMutex
	store    IdentityStore
	cache    map[string]*models.PersonalIdentity
	inferencer *Inferencer
}

// IdentityStore - 身份存储接口
type IdentityStore interface {
	SaveIdentity(ctx context.Context, identity *models.PersonalIdentity) error
	GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error)
	DeleteIdentity(ctx context.Context, id string) error
	ListIdentities(ctx context.Context, page, pageSize int) ([]*models.PersonalIdentity, int, error)
}

// NewService 创建身份服务
func NewService(store IdentityStore) *Service {
	return &Service{
		store:    store,
		cache:    make(map[string]*models.PersonalIdentity),
		inferencer: NewInferencer(),
	}
}

// === 身份管理 ===

// CreateIdentity 创建新身份
func (s *Service) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (*models.PersonalIdentity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := req.ID
	if id == "" {
		id = generateID()
	}

	identity := models.NewPersonalIdentity(id)

	// 应用请求参数
	if req.Name != "" {
		identity.Name = req.Name
	}
	if req.Nickname != "" {
		identity.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		identity.Avatar = req.Avatar
	}
	if !req.Birthday.IsZero() {
		identity.Birthday = req.Birthday
	}
	if req.Gender != "" {
		identity.Gender = req.Gender
	}
	if req.Location != "" {
		identity.Location = req.Location
	}
	if req.Occupation != "" {
		identity.Occupation = req.Occupation
	}
	if len(req.Languages) > 0 {
		identity.Languages = req.Languages
	}
	if req.Timezone != "" {
		identity.Timezone = req.Timezone
	}

	// 保存到存储
	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	// 缓存
	s.cache[id] = identity

	return identity, nil
}

// GetIdentity 获取身份
func (s *Service) GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error) {
	s.mu.RLock()
	if identity, ok := s.cache[id]; ok {
		s.mu.RUnlock()
		return identity, nil
	}
	s.mu.RUnlock()

	// 从存储获取
	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	// 缓存
	s.mu.Lock()
	s.cache[id] = identity
	s.mu.Unlock()

	return identity, nil
}

// UpdateIdentity 更新身份
func (s *Service) UpdateIdentity(ctx context.Context, id string, updates map[string]interface{}) (*models.PersonalIdentity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				identity.Name = v
			}
		case "nickname":
			if v, ok := value.(string); ok {
				identity.Nickname = v
			}
		case "avatar":
			if v, ok := value.(string); ok {
				identity.Avatar = v
			}
		case "gender":
			if v, ok := value.(string); ok {
				identity.Gender = v
			}
		case "location":
			if v, ok := value.(string); ok {
				identity.Location = v
			}
		case "occupation":
			if v, ok := value.(string); ok {
				identity.Occupation = v
			}
		case "languages":
			if v, ok := value.([]string); ok {
				identity.Languages = v
			}
		case "timezone":
			if v, ok := value.(string); ok {
				identity.Timezone = v
			}
		}
	}

	identity.UpdatedAt = time.Now()

	// 保存
	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	// 更新缓存
	s.cache[id] = identity

	return identity, nil
}

// DeleteIdentity 删除身份
func (s *Service) DeleteIdentity(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.store.DeleteIdentity(ctx, id); err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	delete(s.cache, id)
	return nil
}

// === 性格管理 ===

// GetPersonality 获取性格
func (s *Service) GetPersonality(ctx context.Context, id string) (*models.Personality, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.Personality, nil
}

// UpdatePersonality 更新性格
func (s *Service) UpdatePersonality(ctx context.Context, id string, updates map[string]float64) (*models.Personality, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	identity.UpdatePersonality(updates)

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity

	return identity.Personality, nil
}

// SetSpeakingTone 设置说话语调
func (s *Service) SetSpeakingTone(ctx context.Context, id string, tone string, responseLength string, emojiUsage float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	if identity.Personality == nil {
		identity.Personality = &models.Personality{
			CustomTraits: make(map[string]float64),
		}
	}

	identity.Personality.SpeakingTone = tone
	identity.Personality.ResponseLength = responseLength
	identity.Personality.EmojiUsage = emojiUsage
	identity.UpdatedAt = time.Now()

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// === 价值观管理 ===

// GetValueSystem 获取价值观系统
func (s *Service) GetValueSystem(ctx context.Context, id string) (*models.ValueSystem, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.ValueSystem, nil
}

// UpdateValueSystem 更新价值观
func (s *Service) UpdateValueSystem(ctx context.Context, id string, updates map[string]float64) (*models.ValueSystem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	identity.UpdateValueSystem(updates)

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity

	return identity.ValueSystem, nil
}

// GetValuePriority 获取价值观优先级
func (s *Service) GetValuePriority(ctx context.Context, id string) ([]string, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.GetValuePriority(), nil
}

// === 兴趣管理 ===

// AddInterest 添加兴趣
func (s *Service) AddInterest(ctx context.Context, id string, interest models.Interest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	identity.AddInterest(interest)

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// RemoveInterest 移除兴趣
func (s *Service) RemoveInterest(ctx context.Context, id string, interestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	if !identity.RemoveInterest(interestID) {
		return fmt.Errorf("interest not found: %s", interestID)
	}

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// GetInterests 获取所有兴趣
func (s *Service) GetInterests(ctx context.Context, id string) ([]models.Interest, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.Interests, nil
}

// GetInterestsByCategory 按类别获取兴趣
func (s *Service) GetInterestsByCategory(ctx context.Context, id string, category string) ([]models.Interest, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.GetInterestByCategory(category), nil
}

// UpdateInterestLevel 更新兴趣热衷程度
func (s *Service) UpdateInterestLevel(ctx context.Context, id string, interestID string, level float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	for i, interest := range identity.Interests {
		if interest.ID == interestID {
			identity.Interests[i].Level = level
			identity.Interests[i].LastActive = time.Now()
			identity.UpdatedAt = time.Now()
			break
		}
	}

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// === 语音配置 ===

// GetVoiceProfile 获取语音配置
func (s *Service) GetVoiceProfile(ctx context.Context, id string) (*models.VoiceProfile, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.VoiceProfile, nil
}

// UpdateVoiceProfile 更新语音配置
func (s *Service) UpdateVoiceProfile(ctx context.Context, id string, profile *models.VoiceProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	profile.UpdatedAt = time.Now()
	identity.VoiceProfile = profile
	identity.UpdatedAt = time.Now()

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// === 写作风格 ===

// GetWritingStyle 获取写作风格
func (s *Service) GetWritingStyle(ctx context.Context, id string) (*models.WritingStyle, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}
	return identity.WritingStyle, nil
}

// UpdateWritingStyle 更新写作风格
func (s *Service) UpdateWritingStyle(ctx context.Context, id string, style *models.WritingStyle) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return fmt.Errorf("identity not found: %w", err)
	}

	identity.WritingStyle = style
	identity.UpdatedAt = time.Now()

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity
	return nil
}

// === 性格推断 ===

// InferPersonalityFromBehavior 从行为推断性格
func (s *Service) InferPersonalityFromBehavior(ctx context.Context, id string, observations []models.BehaviorObservation) (*models.Personality, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	identity, err := s.store.GetIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity not found: %w", err)
	}

	// 使用推断引擎
	inferred := s.inferencer.InferFromBehavior(observations)

	// 合并到现有性格
	if identity.Personality == nil {
		identity.Personality = &models.Personality{
			CustomTraits: make(map[string]float64),
		}
	}

	// 加权平均合并
	identity.Personality.Openness = (identity.Personality.Openness + inferred.Openness) / 2
	identity.Personality.Conscientiousness = (identity.Personality.Conscientiousness + inferred.Conscientiousness) / 2
	identity.Personality.Extraversion = (identity.Personality.Extraversion + inferred.Extraversion) / 2
	identity.Personality.Agreeableness = (identity.Personality.Agreeableness + inferred.Agreeableness) / 2
	identity.Personality.Neuroticism = (identity.Personality.Neuroticism + inferred.Neuroticism) / 2

	identity.UpdatedAt = time.Now()

	if err := s.store.SaveIdentity(ctx, identity); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	s.cache[id] = identity

	return identity.Personality, nil
}

// GetDecisionContext 获取决策上下文（用于决策引擎）
func (s *Service) GetDecisionContext(ctx context.Context, id string) (*DecisionContext, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}

	return &DecisionContext{
		UserID:         id,
		Personality:    identity.Personality,
		ValueSystem:    identity.ValueSystem,
		Interests:      identity.Interests,
		SpeakingTone:   identity.Personality.SpeakingTone,
		ResponseLength: identity.Personality.ResponseLength,
		ValuePriority:  identity.GetValuePriority(),
	}, nil
}

// === 请求结构 ===

// CreateIdentityRequest - 创建身份请求
type CreateIdentityRequest struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Birthday   time.Time `json:"birthday"`
	Gender     string    `json:"gender"`
	Location   string    `json:"location"`
	Occupation string    `json:"occupation"`
	Languages  []string  `json:"languages"`
	Timezone   string    `json:"timezone"`
}

// DecisionContext - 决策上下文
type DecisionContext struct {
	UserID         string                `json:"user_id"`
	Personality    *models.Personality   `json:"personality"`
	ValueSystem    *models.ValueSystem   `json:"value_system"`
	Interests      []models.Interest     `json:"interests"`
	SpeakingTone   string                `json:"speaking_tone"`
	ResponseLength string                `json:"response_length"`
	ValuePriority  []string              `json:"value_priority"`
}

// === 辅助函数 ===

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
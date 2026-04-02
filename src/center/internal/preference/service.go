package preference

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// Service - 偏好服务
type Service struct {
	mu      sync.RWMutex
	store   PreferenceStore
	learner *Learner
}

// PreferenceStore - 偏好存储接口
type PreferenceStore interface {
	SavePreference(ctx context.Context, pref *models.Preference) error
	GetPreference(ctx context.Context, id string) (*models.Preference, error)
	DeletePreference(ctx context.Context, id string) error
	ListPreferences(ctx context.Context, query *models.PreferenceQuery) ([]*models.Preference, int, error)
	GetPreferenceByKey(ctx context.Context, userID, category, key string) (*models.Preference, error)

	// 学习事件
	SaveLearningEvent(ctx context.Context, event *models.PreferenceLearningEvent) error
	ListLearningEvents(ctx context.Context, userID string, limit int) ([]*models.PreferenceLearningEvent, error)

	// 统计
	GetPreferenceStats(ctx context.Context, userID string) (*models.PreferenceStats, error)
}

// NewService 创建偏好服务
func NewService(store PreferenceStore) *Service {
	return &Service{
		store:   store,
		learner: NewLearner(),
	}
}

// === 偏好管理 ===

// SetPreference 设置偏好
func (s *Service) SetPreference(ctx context.Context, userID, category, key string, value interface{}, opts ...PreferenceOption) (*models.Preference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已存在
	existing, _ := s.store.GetPreferenceByKey(ctx, userID, category, key)

	var pref *models.Preference
	if existing != nil {
		pref = existing
		pref.Value = value
		pref.Source = string(models.PrefSourceExplicit)
		pref.Confidence = 0.9 // 明确设置的偏好高置信度
	} else {
		pref = models.NewPreference(userID, category, key, value)
		pref.Source = string(models.PrefSourceExplicit)
		pref.Confidence = 0.9
	}

	// 应用选项
	for _, opt := range opts {
		opt(pref)
	}

	pref.UpdatedAt = time.Now()

	if err := s.store.SavePreference(ctx, pref); err != nil {
		return nil, fmt.Errorf("failed to save preference: %w", err)
	}

	return pref, nil
}

// PreferenceOption 偏好选项
type PreferenceOption func(*models.Preference)

// WithConditions 设置条件
func WithConditions(conditions []models.Condition) PreferenceOption {
	return func(p *models.Preference) {
		p.Conditions = conditions
	}
}

// WithTags 设置标签
func WithTags(tags []string) PreferenceOption {
	return func(p *models.Preference) {
		p.Tags = tags
	}
}

// WithExpiresAt 设置过期时间
func WithExpiresAt(expiresAt time.Time) PreferenceOption {
	return func(p *models.Preference) {
		p.ExpiresAt = &expiresAt
	}
}

// WithNotes 设置备注
func WithNotes(notes string) PreferenceOption {
	return func(p *models.Preference) {
		p.Notes = notes
	}
}

// GetPreference 获取偏好
func (s *Service) GetPreference(ctx context.Context, userID, category, key string) (*models.Preference, error) {
	pref, err := s.store.GetPreferenceByKey(ctx, userID, category, key)
	if err != nil {
		return nil, fmt.Errorf("preference not found: %w", err)
	}

	// 检查是否过期
	if pref.IsExpired() {
		return nil, fmt.Errorf("preference expired")
	}

	// 更新访问统计
	pref.Access()
	s.store.SavePreference(ctx, pref)

	return pref, nil
}

// GetPreferences 获取多个偏好
func (s *Service) GetPreferences(ctx context.Context, query *models.PreferenceQuery) ([]*models.Preference, int, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}

	prefs, total, err := s.store.ListPreferences(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list preferences: %w", err)
	}

	// 过滤过期的
	var valid []*models.Preference
	for _, p := range prefs {
		if !p.IsExpired() {
			valid = append(valid, p)
		}
	}

	return valid, total, nil
}

// DeletePreference 删除偏好
func (s *Service) DeletePreference(ctx context.Context, id string) error {
	return s.store.DeletePreference(ctx, id)
}

// ConfirmPreference 确认偏好（用户行为验证）
func (s *Service) ConfirmPreference(ctx context.Context, id string) (*models.Preference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pref, err := s.store.GetPreference(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("preference not found: %w", err)
	}

	pref.Confirm()
	s.store.SavePreference(ctx, pref)

	return pref, nil
}

// RejectPreference 拒绝偏好
func (s *Service) RejectPreference(ctx context.Context, id string) (*models.Preference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pref, err := s.store.GetPreference(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("preference not found: %w", err)
	}

	pref.Reject()

	// 如果置信度过低，考虑删除
	if pref.Confidence < 0.2 {
		s.store.DeletePreference(ctx, id)
		return nil, nil
	}

	s.store.SavePreference(ctx, pref)
	return pref, nil
}

// === 偏好学习 ===

// LearnFromEvent 从事件学习偏好
func (s *Service) LearnFromEvent(ctx context.Context, event *models.PreferenceLearningEvent) (*models.Preference, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存学习事件
	if err := s.store.SaveLearningEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save learning event: %w", err)
	}

	// 推断偏好
	pref := s.learner.Learn(event)
	if pref == nil {
		return nil, nil
	}

	// 检查是否与现有偏好冲突
	existing, _ := s.store.GetPreferenceByKey(ctx, event.UserID, pref.Category, pref.Key)
	if existing != nil {
		// 合并或更新
		return s.mergePreferences(ctx, existing, pref)
	}

	// 保存新偏好
	if err := s.store.SavePreference(ctx, pref); err != nil {
		return nil, fmt.Errorf("failed to save preference: %w", err)
	}

	return pref, nil
}

// mergePreferences 合并偏好
func (s *Service) mergePreferences(ctx context.Context, existing, newPref *models.Preference) (*models.Preference, error) {
	// 如果值相同，增加置信度
	if existing.Value == newPref.Value {
		existing.Confirm()
		existing.UpdatedAt = time.Now()
	} else {
		// 值不同，降低置信度
		existing.Reject()
	}

	if err := s.store.SavePreference(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// InferPreference 推断偏好
func (s *Service) InferPreference(ctx context.Context, userID, category string) ([]*models.Preference, error) {
	// 获取学习事件
	events, err := s.store.ListLearningEvents(ctx, userID, 100)
	if err != nil {
		return nil, err
	}

	// 过滤类别的学习事件
	var categoryEvents []*models.PreferenceLearningEvent
	for _, e := range events {
		if e.Category == category {
			categoryEvents = append(categoryEvents, e)
		}
	}

	// 从事件推断偏好
	prefs := s.learner.InferFromEvents(userID, category, categoryEvents)

	// 保存推断的偏好
	for _, pref := range prefs {
		pref.Source = string(models.PrefSourceInferred)
		s.store.SavePreference(ctx, pref)
	}

	return prefs, nil
}

// GetContextualPreferences 获取上下文相关的偏好
func (s *Service) GetContextualPreferences(ctx context.Context, userID string, context map[string]interface{}) ([]*models.Preference, error) {
	// 获取所有偏好
	query := &models.PreferenceQuery{
		UserID: userID,
		Limit:  100,
	}

	allPrefs, _, err := s.store.ListPreferences(ctx, query)
	if err != nil {
		return nil, err
	}

	// 过滤匹配上下文的偏好
	var matched []*models.Preference
	for _, pref := range allPrefs {
		if pref.IsExpired() {
			continue
		}

		if len(pref.Conditions) == 0 {
			matched = append(matched, pref)
			continue
		}

		if pref.MatchesConditions(context) {
			matched = append(matched, pref)
		}
	}

	return matched, nil
}

// === 统计 ===

// GetStats 获取偏好统计
func (s *Service) GetStats(ctx context.Context, userID string) (*models.PreferenceStats, error) {
	return s.store.GetPreferenceStats(ctx, userID)
}
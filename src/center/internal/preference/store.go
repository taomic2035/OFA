package preference

import (
	"context"
	"fmt"
	"sync"

	"github.com/ofa/center/internal/models"
)

// MemoryStore - 内存存储实现
type MemoryStore struct {
	mu             sync.RWMutex
	preferences    map[string]*models.Preference              // id -> preference
	userIndex      map[string]map[string]*models.Preference   // userID -> key -> preference
	keyIndex       map[string]map[string]map[string]*models.Preference // userID -> category -> key -> preference
	learningEvents map[string][]*models.PreferenceLearningEvent // userID -> events
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		preferences:    make(map[string]*models.Preference),
		userIndex:      make(map[string]map[string]*models.Preference),
		keyIndex:       make(map[string]map[string]map[string]*models.Preference),
		learningEvents: make(map[string][]*models.PreferenceLearningEvent),
	}
}

// SavePreference 保存偏好
func (s *MemoryStore) SavePreference(ctx context.Context, pref *models.Preference) error {
	if pref == nil || pref.ID == "" {
		return fmt.Errorf("invalid preference")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.preferences[pref.ID] = pref

	// 用户索引
	if s.userIndex[pref.UserID] == nil {
		s.userIndex[pref.UserID] = make(map[string]*models.Preference)
	}
	s.userIndex[pref.UserID][pref.ID] = pref

	// 键索引
	if s.keyIndex[pref.UserID] == nil {
		s.keyIndex[pref.UserID] = make(map[string]map[string]*models.Preference)
	}
	if s.keyIndex[pref.UserID][pref.Category] == nil {
		s.keyIndex[pref.UserID][pref.Category] = make(map[string]*models.Preference)
	}
	s.keyIndex[pref.UserID][pref.Category][pref.Key] = pref

	return nil
}

// GetPreference 获取偏好
func (s *MemoryStore) GetPreference(ctx context.Context, id string) (*models.Preference, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pref, ok := s.preferences[id]
	if !ok {
		return nil, fmt.Errorf("preference not found: %s", id)
	}

	return pref, nil
}

// DeletePreference 删除偏好
func (s *MemoryStore) DeletePreference(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pref, ok := s.preferences[id]
	if !ok {
		return nil
	}

	delete(s.preferences, id)
	if userPrefs, ok := s.userIndex[pref.UserID]; ok {
		delete(userPrefs, id)
	}
	if catPrefs, ok := s.keyIndex[pref.UserID]; ok {
		if keyPrefs, ok := catPrefs[pref.Category]; ok {
			delete(keyPrefs, pref.Key)
		}
	}

	return nil
}

// ListPreferences 列出偏好
func (s *MemoryStore) ListPreferences(ctx context.Context, query *models.PreferenceQuery) ([]*models.Preference, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userPrefs, ok := s.userIndex[query.UserID]
	if !ok {
		return []*models.Preference{}, 0, nil
	}

	var filtered []*models.Preference
	for _, pref := range userPrefs {
		if !s.matchesQuery(pref, query) {
			continue
		}
		filtered = append(filtered, pref)
	}

	total := len(filtered)

	// 分页
	start := query.Offset
	if start > total {
		return []*models.Preference{}, total, nil
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// matchesQuery 检查是否匹配查询
func (s *MemoryStore) matchesQuery(pref *models.Preference, query *models.PreferenceQuery) bool {
	// 类别过滤
	if len(query.Categories) > 0 {
		found := false
		for _, c := range query.Categories {
			if pref.Category == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 键过滤
	if len(query.Keys) > 0 {
		found := false
		for _, k := range query.Keys {
			if pref.Key == k {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 最小置信度
	if pref.Confidence < query.MinConfidence {
		return false
	}

	// 来源过滤
	if query.Source != "" && pref.Source != query.Source {
		return false
	}

	// 标签过滤
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			found := false
			for _, t := range pref.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// GetPreferenceByKey 按键获取偏好
func (s *MemoryStore) GetPreferenceByKey(ctx context.Context, userID, category, key string) (*models.Preference, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.keyIndex[userID] == nil {
		return nil, fmt.Errorf("preference not found")
	}
	if s.keyIndex[userID][category] == nil {
		return nil, fmt.Errorf("preference not found")
	}

	pref, ok := s.keyIndex[userID][category][key]
	if !ok {
		return nil, fmt.Errorf("preference not found")
	}

	return pref, nil
}

// SaveLearningEvent 保存学习事件
func (s *MemoryStore) SaveLearningEvent(ctx context.Context, event *models.PreferenceLearningEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.learningEvents[event.UserID] = append(s.learningEvents[event.UserID], event)
	return nil
}

// ListLearningEvents 列出学习事件
func (s *MemoryStore) ListLearningEvents(ctx context.Context, userID string, limit int) ([]*models.PreferenceLearningEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.learningEvents[userID]
	if len(events) == 0 {
		return []*models.PreferenceLearningEvent{}, nil
	}

	// 返回最近的 N 个
	if limit > 0 && len(events) > limit {
		start := len(events) - limit
		return events[start:], nil
	}

	return events, nil
}

// GetPreferenceStats 获取统计
func (s *MemoryStore) GetPreferenceStats(ctx context.Context, userID string) (*models.PreferenceStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userPrefs := s.userIndex[userID]
	if userPrefs == nil {
		return &models.PreferenceStats{UserID: userID}, nil
	}

	stats := &models.PreferenceStats{
		UserID:          userID,
		CountByCategory: make(map[string]int),
		CountBySource:   make(map[string]int),
		TopPreferences:  make(map[string][]*models.Preference),
		Conflicts:       []*models.PreferenceConflict{},
	}

	totalConfidence := 0.0
	categoryPrefs := make(map[string][]*models.Preference)

	for _, pref := range userPrefs {
		stats.TotalCount++
		stats.CountByCategory[pref.Category]++
		stats.CountBySource[pref.Source]++
		totalConfidence += pref.Confidence

		categoryPrefs[pref.Category] = append(categoryPrefs[pref.Category], pref)
	}

	if stats.TotalCount > 0 {
		stats.AvgConfidence = totalConfidence / float64(stats.TotalCount)
	}

	// 每个类别的 top 偏好
	for cat, prefs := range categoryPrefs {
		// 按置信度排序
		for i := 0; i < len(prefs)-1; i++ {
			for j := i + 1; j < len(prefs); j++ {
				if prefs[j].Confidence > prefs[i].Confidence {
					prefs[i], prefs[j] = prefs[j], prefs[i]
				}
			}
		}

		// 取前 3 个
		if len(prefs) > 3 {
			stats.TopPreferences[cat] = prefs[:3]
		} else {
			stats.TopPreferences[cat] = prefs
		}
	}

	return stats, nil
}
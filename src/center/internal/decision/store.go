package decision

import (
	"context"
	"fmt"
	"sync"

	"github.com/ofa/center/internal/models"
)

// MemoryStore - 内存存储实现
type MemoryStore struct {
	mu         sync.RWMutex
	decisions  map[string]*models.Decision           // id -> decision
	userIndex  map[string]map[string]*models.Decision // userID -> id -> decision
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		decisions: make(map[string]*models.Decision),
		userIndex: make(map[string]map[string]*models.Decision),
	}
}

// SaveDecision 保存决策
func (s *MemoryStore) SaveDecision(ctx context.Context, decision *models.Decision) error {
	if decision == nil || decision.ID == "" {
		return fmt.Errorf("invalid decision")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.decisions[decision.ID] = decision

	if s.userIndex[decision.UserID] == nil {
		s.userIndex[decision.UserID] = make(map[string]*models.Decision)
	}
	s.userIndex[decision.UserID][decision.ID] = decision

	return nil
}

// GetDecision 获取决策
func (s *MemoryStore) GetDecision(ctx context.Context, id string) (*models.Decision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	decision, ok := s.decisions[id]
	if !ok {
		return nil, fmt.Errorf("decision not found: %s", id)
	}

	return decision, nil
}

// ListDecisions 列出决策
func (s *MemoryStore) ListDecisions(ctx context.Context, query *models.DecisionQuery) ([]*models.Decision, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userDecisions, ok := s.userIndex[query.UserID]
	if !ok {
		return []*models.Decision{}, 0, nil
	}

	var filtered []*models.Decision
	for _, decision := range userDecisions {
		if !s.matchesQuery(decision, query) {
			continue
		}
		filtered = append(filtered, decision)
	}

	total := len(filtered)

	// 按时间排序（最新的在前）
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].CreatedAt.After(filtered[i].CreatedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// 分页
	start := query.Offset
	if start > total {
		return []*models.Decision{}, total, nil
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// matchesQuery 检查是否匹配查询
func (s *MemoryStore) matchesQuery(decision *models.Decision, query *models.DecisionQuery) bool {
	if query.Scenario != "" && decision.Scenario != query.Scenario {
		return false
	}

	if query.Outcome != "" && decision.Outcome != query.Outcome {
		return false
	}

	if query.AutoDecided != nil && decision.AutoDecided != *query.AutoDecided {
		return false
	}

	if query.StartTime != nil && decision.CreatedAt.Before(*query.StartTime) {
		return false
	}

	if query.EndTime != nil && decision.CreatedAt.After(*query.EndTime) {
		return false
	}

	return true
}

// GetDecisionStats 获取决策统计
func (s *MemoryStore) GetDecisionStats(ctx context.Context, userID string) (*models.DecisionStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userDecisions := s.userIndex[userID]
	if userDecisions == nil {
		return &models.DecisionStats{UserID: userID}, nil
	}

	stats := &models.DecisionStats{
		UserID:          userID,
		CountByScenario: make(map[string]int),
		ValueUsage:      make(map[string]int),
		PreferenceHits:  make(map[string]int),
	}

	totalOutcomeScore := 0.0

	for _, decision := range userDecisions {
		stats.TotalDecisions++

		if decision.AutoDecided {
			stats.AutoDecisions++
		} else {
			stats.ManualDecisions++
		}

		if decision.Outcome == "satisfied" {
			stats.SatisfiedCount++
		} else if decision.Outcome == "unsatisfied" {
			stats.UnsatisfiedCount++
		}

		if decision.OutcomeScore > 0 {
			totalOutcomeScore += decision.OutcomeScore
		}

		stats.CountByScenario[decision.Scenario]++

		for _, v := range decision.AppliedValues {
			stats.ValueUsage[v]++
		}

		for _, p := range decision.AppliedPreferences {
			stats.PreferenceHits[p]++
		}
	}

	if stats.TotalDecisions > 0 {
		stats.AvgOutcomeScore = totalOutcomeScore / float64(stats.TotalDecisions)
	}

	// Top scenarios
	for scenario := range stats.CountByScenario {
		if len(stats.TopScenarios) < 5 {
			stats.TopScenarios = append(stats.TopScenarios, scenario)
		}
	}

	return stats, nil
}
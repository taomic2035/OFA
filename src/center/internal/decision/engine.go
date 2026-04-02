package decision

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// Engine - 决策引擎
type Engine struct {
	mu          sync.RWMutex
	store       DecisionStore
	scorer      *Scorer
	explainer   *Explainer
	rules       map[string]*models.DecisionScenario
}

// DecisionStore - 决策存储接口
type DecisionStore interface {
	SaveDecision(ctx context.Context, decision *models.Decision) error
	GetDecision(ctx context.Context, id string) (*models.Decision, error)
	ListDecisions(ctx context.Context, query *models.DecisionQuery) ([]*models.Decision, int, error)
	GetDecisionStats(ctx context.Context, userID string) (*models.DecisionStats, error)
}

// NewEngine 创建决策引擎
func NewEngine(store DecisionStore) *Engine {
	return &Engine{
		store:     store,
		scorer:    NewScorer(),
		explainer: NewExplainer(),
		rules:     DefaultScenarios(),
	}
}

// === 决策执行 ===

// Decide 执行决策
func (e *Engine) Decide(ctx context.Context, decisionCtx *models.DecisionContext, scenario string, options []models.DecisionOption, context map[string]interface{}) (*models.DecisionResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 创建决策记录
	decision := models.NewDecision(decisionCtx.UserID, scenario)
	decision.Context = context
	decision.Options = options

	// 获取场景规则
	scenarioRules := e.rules[scenario]
	if scenarioRules == nil {
		scenarioRules = DefaultScenario(scenario)
	}

	// 评分所有选项
	for i := range decision.Options {
		score, breakdown := e.scorer.Score(&decision.Options[i], decisionCtx, scenarioRules)
		decision.Options[i].Score = score
		decision.Options[i].ScoreBreakdown = breakdown
	}

	// 排序选项
	e.rankOptions(decision)

	// 选择最佳选项
	if len(decision.Options) > 0 {
		decision.Select(0, "Highest score")
		decision.AppliedValues = e.getAppliedValues(decisionCtx, scenarioRules)
		decision.AppliedPreferences = e.getAppliedPreferences(decisionCtx, scenarioRules)
	}

	// 计算置信度
	decision.Confidence = decision.CalculateConfidence()

	// 生成解释
	explanation := e.explainer.Explain(decision, decisionCtx, scenarioRules)

	// 判断是否需要用户确认
	needsInput := decision.Confidence < 0.5 || len(decision.Options) == 0 ||
		(len(decision.Options) > 1 && decision.Options[0].Score-decision.Options[1].Score < 0.1)

	// 保存决策
	if err := e.store.SaveDecision(ctx, decision); err != nil {
		return nil, fmt.Errorf("failed to save decision: %w", err)
	}

	return &models.DecisionResult{
		Decision:        decision,
		Alternatives:    decision.GetTopOptions(3)[1:], // 排除第一名的备选
		Explanation:     explanation,
		Confidence:      decision.Confidence,
		NeedsUserInput:  needsInput,
		UncertainReason: e.getUncertainReason(decision),
	}, nil
}

// QuickDecide 快速决策（自动选择最高分）
func (e *Engine) QuickDecide(ctx context.Context, decisionCtx *models.DecisionContext, scenario string, options []models.DecisionOption) (*models.DecisionResult, error) {
	result, err := e.Decide(ctx, decisionCtx, scenario, options, nil)
	if err != nil {
		return nil, err
	}

	// 标记为自动决策
	result.Decision.AutoDecided = true
	e.store.SaveDecision(ctx, result.Decision)

	return result, nil
}

// ConfirmDecision 确认决策
func (e *Engine) ConfirmDecision(ctx context.Context, decisionID string, optionIndex int) (*models.Decision, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	decision, err := e.store.GetDecision(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("decision not found: %w", err)
	}

	decision.Select(optionIndex, "User confirmed")
	decision.AutoDecided = false
	decision.UpdatedAt = time.Now()

	if err := e.store.SaveDecision(ctx, decision); err != nil {
		return nil, err
	}

	return decision, nil
}

// RecordOutcome 记录决策结果
func (e *Engine) RecordOutcome(ctx context.Context, decisionID, outcome, feedback string, score float64) (*models.Decision, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	decision, err := e.store.GetDecision(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("decision not found: %w", err)
	}

	decision.SetOutcome(outcome, feedback, score)

	if err := e.store.SaveDecision(ctx, decision); err != nil {
		return nil, err
	}

	return decision, nil
}

// === 决策查询 ===

// GetDecision 获取决策
func (e *Engine) GetDecision(ctx context.Context, id string) (*models.Decision, error) {
	return e.store.GetDecision(ctx, id)
}

// GetHistory 获取历史决策
func (e *Engine) GetHistory(ctx context.Context, query *models.DecisionQuery) ([]*models.Decision, int, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}
	return e.store.ListDecisions(ctx, query)
}

// GetStats 获取决策统计
func (e *Engine) GetStats(ctx context.Context, userID string) (*models.DecisionStats, error) {
	return e.store.GetDecisionStats(ctx, userID)
}

// === 辅助方法 ===

// rankOptions 排序选项
func (e *Engine) rankOptions(decision *models.Decision) {
	// 按分数降序排序
	for i := 0; i < len(decision.Options)-1; i++ {
		for j := i + 1; j < len(decision.Options); j++ {
			if decision.Options[j].Score > decision.Options[i].Score {
				decision.Options[i], decision.Options[j] = decision.Options[j], decision.Options[i]
			}
		}
	}

	// 设置排名
	for i := range decision.Options {
		decision.Options[i].Rank = i + 1
	}

	// 记录排名
	decision.Ranking = make([]int, len(decision.Options))
	for i, opt := range decision.Options {
		decision.Ranking[i] = opt.Rank
	}
}

// getAppliedValues 获取应用的价值观
func (e *Engine) getAppliedValues(ctx *models.DecisionContext, scenario *models.DecisionScenario) []string {
	var applied []string
	if ctx.ValueSystem != nil {
		// 根据场景需要的价值观
		for _, v := range scenario.RequiredValues {
			if ctx.ValueSystem.GetPriority(v) > 0.5 {
				applied = append(applied, v)
			}
		}
	}
	return applied
}

// getAppliedPreferences 获取应用的偏好
func (e *Engine) getAppliedPreferences(ctx *models.DecisionContext, scenario *models.DecisionScenario) []string {
	var applied []string
	// 从活跃偏好中获取相关的
	for _, pref := range scenario.RequiredPrefs {
		if _, ok := ctx.ActivePreferences[pref]; ok {
			applied = append(applied, pref)
		}
	}
	return applied
}

// getUncertainReason 获取不确定原因
func (e *Engine) getUncertainReason(decision *models.Decision) string {
	if len(decision.Options) == 0 {
		return "No options available"
	}

	if len(decision.Options) == 1 {
		return "Only one option available"
	}

	gap := decision.Options[0].Score - decision.Options[1].Score
	if gap < 0.1 {
		return "Multiple options have similar scores"
	}

	return ""
}

// DefaultScenarios 默认场景
func DefaultScenarios() map[string]*models.DecisionScenario {
	return map[string]*models.DecisionScenario{
		"food_ordering": DefaultScenario("food_ordering"),
		"shopping":      DefaultScenario("shopping"),
		"travel":        DefaultScenario("travel"),
		"entertainment": DefaultScenario("entertainment"),
	}
}

// DefaultScenario 默认场景配置
func DefaultScenario(name string) *models.DecisionScenario {
	base := &models.DecisionScenario{
		ID:             name,
		Name:           name,
		RequiredValues: []string{"efficiency", "health", "finance"},
		RequiredPrefs:  []string{},
		ScoringRules:   []models.ScoringRule{},
		Constraints:    []models.Constraint{},
	}

	switch name {
	case "food_ordering":
		base.Category = "food"
		base.Description = "点餐决策"
		base.ScoringRules = []models.ScoringRule{
			{ID: "health", Name: "健康偏好", Attribute: "health_score", Weight: 0.3},
			{ID: "price", Name: "价格偏好", Attribute: "price_score", Weight: 0.2},
			{ID: "rating", Name: "评分偏好", Attribute: "rating", Weight: 0.2},
		}
	case "shopping":
		base.Category = "shop"
		base.Description = "购物决策"
		base.ScoringRules = []models.ScoringRule{
			{ID: "price", Name: "价格偏好", Attribute: "price_score", Weight: 0.3},
			{ID: "quality", Name: "质量偏好", Attribute: "quality_score", Weight: 0.25},
			{ID: "rating", Name: "评分偏好", Attribute: "rating", Weight: 0.2},
		}
	case "travel":
		base.Category = "travel"
		base.Description = "出行决策"
		base.ScoringRules = []models.ScoringRule{
			{ID: "time", Name: "时间偏好", Attribute: "time_score", Weight: 0.3},
			{ID: "cost", Name: "成本偏好", Attribute: "cost_score", Weight: 0.25},
			{ID: "comfort", Name: "舒适度偏好", Attribute: "comfort_score", Weight: 0.2},
		}
	default:
		base.Category = "general"
		base.Description = "通用决策"
	}

	return base
}
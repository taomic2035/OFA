package philosophy

import (
	"context"
	"sync"
	"time"

	"github.com/taomic2035/OFA/src/center/internal/models"
)

// PhilosophyEngine 三观管理引擎 (v4.1.0)
// 统一管理世界观、人生观、价值观
type PhilosophyEngine struct {
	mu sync.RWMutex

	// 三观存储
	worldviews    map[string]*models.Worldview    // identityID -> Worldview
	lifeViews     map[string]*models.LifeView     // identityID -> LifeView
	valueSystems  map[string]*models.EnhancedValueSystem // identityID -> ValueSystem

	// 画像存储
	worldviewProfiles map[string]*models.WorldviewProfile
	lifeViewProfiles  map[string]*models.LifeViewProfile
	valueProfiles     map[string]*models.ValueSystemProfile

	// 监听器
	listeners []PhilosophyListener
}

// PhilosophyListener 三观变化监听器
type PhilosophyListener interface {
	OnWorldviewChanged(identityID string, worldview *models.Worldview)
	OnLifeViewChanged(identityID string, lifeView *models.LifeView)
	OnValueSystemChanged(identityID string, valueSystem *models.EnhancedValueSystem)
	OnPhilosophyContextUpdated(identityID string, context *PhilosophyDecisionContext)
}

// PhilosophyDecisionContext 三观决策上下文
type PhilosophyDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 三观状态
	Worldview   *models.Worldview           `json:"worldview"`
	LifeView    *models.LifeView            `json:"life_view"`
	ValueSystem *models.EnhancedValueSystem `json:"value_system"`

	// 综合影响因子
	InfluenceFactors map[string]float64 `json:"influence_factors"`

	// 决策倾向
	DecisionTendencies DecisionTendencies `json:"decision_tendencies"`

	// 道德指南
	MoralGuidance MoralGuidance `json:"moral_guidance"`

	// 人生目标
	CurrentGoals []string `json:"current_goals"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// DecisionTendencies 决策倾向
type DecisionTendencies struct {
	// 风险相关
	RiskTolerance     float64 `json:"risk_tolerance"`     // 风险容忍度
	RiskForExperience float64 `json:"risk_for_experience"` // 为体验冒险

	// 时间相关
	DelayedGratification float64 `json:"delayed_gratification"` // 延迟满足
	LongTermPlanning     float64 `json:"long_term_planning"`    // 长期规划
	PresentEnjoyment     float64 `json:"present_enjoyment"`     // 当下享受

	// 社会相关
	Cooperation        float64 `json:"cooperation"`         // 合作倾向
	SocialResponsibility float64 `json:"social_responsibility"` // 社会责任
	Autonomy           float64 `json:"autonomy"`            // 自主性

	// 行动相关
	InnovationTendency float64 `json:"innovation_tendency"` // 创新倾向
	GoalOrientation    float64 `json:"goal_orientation"`    // 目标导向
	MeaningSeeking     float64 `json:"meaning_seeking"`     // 意义追求

	// 平衡相关
	WorkLifeBalance float64 `json:"work_life_balance"` // 工作生活平衡
}

// MoralGuidance 道德指南
type MoralGuidance struct {
	PrimaryMoralValues []string `json:"primary_moral_values"` // 主要道德价值观
	MoralFramework     string   `json:"moral_framework"`      // 道德框架
	ConflictResolution string   `json:"conflict_resolution"`  // 冲突解决策略
	EthicalSensitivity float64  `json:"ethical_sensitivity"`  // 道德敏感度
}

// NewPhilosophyEngine 创建三观管理引擎
func NewPhilosophyEngine() *PhilosophyEngine {
	return &PhilosophyEngine{
		worldviews:        make(map[string]*models.Worldview),
		lifeViews:         make(map[string]*models.LifeView),
		valueSystems:      make(map[string]*models.EnhancedValueSystem),
		worldviewProfiles: make(map[string]*models.WorldviewProfile),
		lifeViewProfiles:  make(map[string]*models.LifeViewProfile),
		valueProfiles:     make(map[string]*models.ValueSystemProfile),
		listeners:         []PhilosophyListener{},
	}
}

// === 世界观管理 ===

// GetWorldview 获取世界观
func (p *PhilosophyEngine) GetWorldview(identityID string) *models.Worldview {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.worldviews[identityID]
}

// GetOrCreateWorldview 获取或创建世界观
func (p *PhilosophyEngine) GetOrCreateWorldview(identityID string) *models.Worldview {
	p.mu.Lock()
	defer p.mu.Unlock()

	worldview, exists := p.worldviews[identityID]
	if !exists {
		worldview = models.NewWorldview()
		p.worldviews[identityID] = worldview
		p.worldviewProfiles[identityID] = models.NewWorldviewProfile(identityID)
	}
	return worldview
}

// UpdateWorldview 更新世界观
func (p *PhilosophyEngine) UpdateWorldview(identityID string, worldview *models.Worldview) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	worldview.Normalize()
	p.worldviews[identityID] = worldview

	// 通知监听器
	for _, listener := range p.listeners {
		listener.OnWorldviewChanged(identityID, worldview)
	}

	return nil
}

// UpdateWorldviewBelief 更新单个信念
func (p *PhilosophyEngine) UpdateWorldviewBelief(identityID string, beliefType string, value float64) error {
	worldview := p.GetOrCreateWorldview(identityID)
	worldview.UpdateBelief(beliefType, value)

	p.mu.Lock()
	for _, listener := range p.listeners {
		listener.OnWorldviewChanged(identityID, worldview)
	}
	p.mu.Unlock()

	return nil
}

// === 人生观管理 ===

// GetLifeView 获取人生观
func (p *PhilosophyEngine) GetLifeView(identityID string) *models.LifeView {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lifeViews[identityID]
}

// GetOrCreateLifeView 获取或创建人生观
func (p *PhilosophyEngine) GetOrCreateLifeView(identityID string) *models.LifeView {
	p.mu.Lock()
	defer p.mu.Unlock()

	lifeView, exists := p.lifeViews[identityID]
	if !exists {
		lifeView = models.NewLifeView()
		p.lifeViews[identityID] = lifeView
		p.lifeViewProfiles[identityID] = models.NewLifeViewProfile(identityID)
	}
	return lifeView
}

// UpdateLifeView 更新人生观
func (p *PhilosophyEngine) UpdateLifeView(identityID string, lifeView *models.LifeView) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	lifeView.Normalize()
	p.lifeViews[identityID] = lifeView

	for _, listener := range p.listeners {
		listener.OnLifeViewChanged(identityID, lifeView)
	}

	return nil
}

// UpdateLifeGoal 更新人生目标
func (p *PhilosophyEngine) UpdateLifeGoal(identityID string, goal string) error {
	lifeView := p.GetOrCreateLifeView(identityID)
	lifeView.LifeGoal = goal

	p.mu.Lock()
	for _, listener := range p.listeners {
		listener.OnLifeViewChanged(identityID, lifeView)
	}
	p.mu.Unlock()

	return nil
}

// AddLifeMilestone 添加人生里程碑
func (p *PhilosophyEngine) AddLifeMilestone(identityID string, milestone models.LifeMilestone) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	profile := p.lifeViewProfiles[identityID]
	if profile == nil {
		profile = models.NewLifeViewProfile(identityID)
		p.lifeViewProfiles[identityID] = profile
	}

	profile.Milestones = append(profile.Milestones, milestone)

	// 里程碑可能影响人生观
	lifeView := p.lifeViews[identityID]
	if lifeView != nil && milestone.Impact > 0.7 {
		// 重大里程碑影响人生意义来源
		if milestone.Positive {
			lifeView.HappinessLevel = min(1.0, lifeView.HappinessLevel+0.1)
		} else {
			lifeView.HappinessLevel = max(0.0, lifeView.HappinessLevel-0.1)
		}
	}

	return nil
}

// === 价值观管理 ===

// GetValueSystem 获取价值观系统
func (p *PhilosophyEngine) GetValueSystem(identityID string) *models.EnhancedValueSystem {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.valueSystems[identityID]
}

// GetOrCreateValueSystem 获取或创建价值观系统
func (p *PhilosophyEngine) GetOrCreateValueSystem(identityID string) *models.EnhancedValueSystem {
	p.mu.Lock()
	defer p.mu.Unlock()

	valueSystem, exists := p.valueSystems[identityID]
	if !exists {
		valueSystem = models.NewEnhancedValueSystem()
		p.valueSystems[identityID] = valueSystem
		p.valueProfiles[identityID] = models.NewValueSystemProfile(identityID)
	}
	return valueSystem
}

// UpdateValueSystem 更新价值观系统
func (p *PhilosophyEngine) UpdateValueSystem(identityID string, valueSystem *models.EnhancedValueSystem) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	valueSystem.Normalize()
	p.valueSystems[identityID] = valueSystem

	for _, listener := range p.listeners {
		listener.OnValueSystemChanged(identityID, valueSystem)
	}

	return nil
}

// UpdateValue 更新单个价值观
func (p *PhilosophyEngine) UpdateValue(identityID string, valueName string, value float64) error {
	valueSystem := p.GetOrCreateValueSystem(identityID)

	p.mu.Lock()
	switch valueName {
	case "privacy":
		valueSystem.Privacy = value
	case "efficiency":
		valueSystem.Efficiency = value
	case "health":
		valueSystem.Health = value
	case "family":
		valueSystem.Family = value
	case "career":
		valueSystem.Career = value
	case "honesty":
		valueSystem.Honesty = value
	case "compassion":
		valueSystem.Compassion = value
	case "freedom":
		valueSystem.Freedom = value
	case "justice":
		valueSystem.Justice = value
	// ... 其他价值观
	default:
		if valueSystem.CustomValues == nil {
			valueSystem.CustomValues = make(map[string]float64)
		}
		valueSystem.CustomValues[valueName] = value
	}
	valueSystem.Normalize()

	for _, listener := range p.listeners {
		listener.OnValueSystemChanged(identityID, valueSystem)
	}
	p.mu.Unlock()

	return nil
}

// MakeValueJudgment 进行价值判断
func (p *PhilosophyEngine) MakeValueJudgment(ctx context.Context, identityID string, situation string, options []models.ValueOption) (*models.ValueJudgment, error) {
	valueSystem := p.GetOrCreateValueSystem(identityID)

	judgment := valueSystem.MakeValueJudgment(situation, options)

	// 记录判断历史
	p.mu.Lock()
	profile := p.valueProfiles[identityID]
	if profile == nil {
		profile = models.NewValueSystemProfile(identityID)
		p.valueProfiles[identityID] = profile
	}
	profile.JudgmentHistory = append(profile.JudgmentHistory, *judgment)
	if len(profile.JudgmentHistory) > 50 {
		profile.JudgmentHistory = profile.JudgmentHistory[len(profile.JudgmentHistory)-50:]
	}
	p.mu.Unlock()

	return judgment, nil
}

// ResolveValueConflict 解决价值观冲突
func (p *PhilosophyEngine) ResolveValueConflict(identityID string, value1, value2 string, context string) (*models.ValueConflict, error) {
	valueSystem := p.GetOrCreateValueSystem(identityID)

	conflict := valueSystem.ResolveConflict(value1, value2, context)

	// 记录冲突历史
	p.mu.Lock()
	profile := p.valueProfiles[identityID]
	if profile == nil {
		profile = models.NewValueSystemProfile(identityID)
		p.valueProfiles[identityID] = profile
	}
	profile.ConflictHistory = append(profile.ConflictHistory, *conflict)
	if len(profile.ConflictHistory) > 20 {
		profile.ConflictHistory = profile.ConflictHistory[len(profile.ConflictHistory)-20:]
	}
	p.mu.Unlock()

	return conflict, nil
}

// === 综合决策上下文 ===

// GetDecisionContext 获取三观决策上下文
func (p *PhilosophyEngine) GetDecisionContext(identityID string) *PhilosophyDecisionContext {
	p.mu.RLock()
	defer p.mu.RUnlock()

	worldview := p.worldviews[identityID]
	lifeView := p.lifeViews[identityID]
	valueSystem := p.valueSystems[identityID]

	// 如果都没有，返回默认值
	if worldview == nil {
		worldview = models.NewWorldview()
	}
	if lifeView == nil {
		lifeView = models.NewLifeView()
	}
	if valueSystem == nil {
		valueSystem = models.NewEnhancedValueSystem()
	}

	// 合并所有影响因子
	influenceFactors := make(map[string]float64)

	// 世界观影响
	for k, v := range worldview.CalculateInfluence() {
		influenceFactors["worldview_"+k] = v
	}

	// 人生观影响
	for k, v := range lifeView.CalculateInfluence() {
		influenceFactors["lifeview_"+k] = v
	}

	// 价值观影响
	for k, v := range valueSystem.CalculateInfluence() {
		influenceFactors["value_"+k] = v
	}

	// 计算综合决策倾向
	tendencies := p.calculateDecisionTendencies(worldview, lifeView, valueSystem)

	// 计算道德指南
	moralGuidance := p.calculateMoralGuidance(valueSystem)

	// 获取当前目标
	goals := []string{}
	if lifeView.LifeGoal != "" {
		goals = append(goals, lifeView.LifeGoal)
	}
	profile := p.lifeViewProfiles[identityID]
	if profile != nil {
		for _, goal := range profile.Goals {
			if goal.Status == "active" {
				goals = append(goals, goal.Name)
			}
		}
	}

	return &PhilosophyDecisionContext{
		IdentityID:        identityID,
		Worldview:         worldview,
		LifeView:          lifeView,
		ValueSystem:       valueSystem,
		InfluenceFactors:  influenceFactors,
		DecisionTendencies: tendencies,
		MoralGuidance:     moralGuidance,
		CurrentGoals:      goals,
		Timestamp:         time.Now(),
	}
}

// calculateDecisionTendencies 计算决策倾向
func (p *PhilosophyEngine) calculateDecisionTendencies(worldview *models.Worldview, lifeView *models.LifeView, valueSystem *models.EnhancedValueSystem) DecisionTendencies {
	tendencies := DecisionTendencies{}

	// 风险容忍度 = 世界观乐观 + 价值观风险容忍
	tendencies.RiskTolerance = (worldview.Optimism + valueSystem.RiskTolerance) / 2

	// 为体验冒险 = 人生观享乐主义 + 世界观拥抱变化
	tendencies.RiskForExperience = (lifeView.Hedonism + worldview.ChangeBelief) / 2

	// 延迟满足 = 人生观未来关注
	tendencies.DelayedGratification = lifeView.FutureFocus

	// 长期规划 = 世界观乐观 + 社会流动性信念
	tendencies.LongTermPlanning = (worldview.Optimism + worldview.SocialMobility) / 2

	// 当下享受 = 人生观当下关注 + 享乐主义
	tendencies.PresentEnjoyment = (lifeView.PresentFocus + lifeView.Hedonism) / 2

	// 合作倾向 = 世界观信任 + 集体主义
	tendencies.Cooperation = (worldview.TrustInPeople + (1 - worldview.Individualism)) / 2

	// 社会责任 = 价值观正义 + 关系观
	tendencies.SocialResponsibility = (valueSystem.Justice + worldview.RelationshipView) / 2

	// 自主性 = 世界观自主意识 + 个人主义 + 价值观自主
	tendencies.Autonomy = (worldview.Agency + worldview.Individualism + valueSystem.Autonomy) / 3

	// 创新倾向 = 世界观科技态度 + 价值观创新 + 创造力
	tendencies.InnovationTendency = (worldview.TechnologyView + valueSystem.Innovation + valueSystem.Creativity) / 3

	// 目标导向 = 人生观未来关注 + 有明确目标
	tendencies.GoalOrientation = lifeView.FutureFocus
	if lifeView.LifeGoal != "" {
		tendencies.GoalOrientation = min(1.0, tendencies.GoalOrientation+0.2)
	}

	// 意义追求 = 人生观有目标
	if len(lifeView.MeaningSource) > 0 {
		tendencies.MeaningSeeking = 0.7
	} else {
		tendencies.MeaningSeeking = 0.3
	}

	// 工作生活平衡 = 人生观平衡主义
	tendencies.WorkLifeBalance = lifeView.Balance

	return tendencies
}

// calculateMoralGuidance 计算道德指南
func (p *PhilosophyEngine) calculateMoralGuidance(valueSystem *models.EnhancedValueSystem) MoralGuidance {
	guidance := MoralGuidance{
		PrimaryMoralValues: valueSystem.GetTopValues(5),
		ConflictResolution: valueSystem.ConflictResolution,
		EthicalSensitivity: 0.5,
	}

	if valueSystem.MoralFramework != nil {
		guidance.MoralFramework = valueSystem.MoralFramework.ReasoningStyle
		// 道德敏感度 = 各道德基础的平均值
		mf := valueSystem.MoralFramework
		guidance.EthicalSensitivity = (mf.CareHarm + mf.FairnessCheating + mf.LoyaltyBetrayal +
			mf.AuthoritySubversion + mf.SanctityDegradation + mf.LibertyOppression) / 6
	}

	return guidance
}

// === 监听器管理 ===

// AddListener 添加监听器
func (p *PhilosophyEngine) AddListener(listener PhilosophyListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners = append(p.listeners, listener)
}

// RemoveListener 移除监听器
func (p *PhilosophyEngine) RemoveListener(listener PhilosophyListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, l := range p.listeners {
		if l == listener {
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			break
		}
	}
}

// === 辅助函数 ===

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
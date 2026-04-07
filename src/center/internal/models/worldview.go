package models

import "time"

// Worldview 世界观模型 (v4.1.0)
// 世界观：对世界本质、社会运行、未来发展的认知
type Worldview struct {
	// === 世界本质认知 (0-1) ===
	Materialism    float64 `json:"materialism"`    // 唯物主义倾向 (0=唯心, 1=唯物)
	Spirituality   float64 `json:"spirituality"`   // 精神/宗教倾向
	Determinism    float64 `json:"determinism"`    // 决定论倾向 (0=自由意志, 1=宿命)
	Agency         float64 `json:"agency"`         // 自主意识信念 (人能否掌控命运)

	// === 社会认知 (0-1) ===
	TrustInPeople   float64 `json:"trust_in_people"`   // 对人的信任度
	JusticeBelief   float64 `json:"justice_belief"`    // 正义信念 (善有善报)
	SocialMobility  float64 `json:"social_mobility"`   // 社会流动性信念 (努力能改变命运)
	CompetitionView float64 `json:"competition_view"`  // 竞争观念 (0=合作优先, 1=竞争优先)

	// === 未来观 (0-1) ===
	Optimism      float64 `json:"optimism"`       // 乐观程度 (对未来的信心)
	ChangeBelief  float64 `json:"change_belief"`  // 变化可接受度 (拥抱变化)
	TechnologyView float64 `json:"technology_view"` // 科技态度 (0=怀疑, 1=拥抱)

	// === 人际关系观 (0-1) ===
	Individualism float64 `json:"individualism"` // 个人主义倾向 (0=集体主义, 1=个人主义)
	RelationshipView float64 `json:"relationship_view"` // 关系观 (人际关系的重要性)

	// === 世界观描述 ===
	Summary string `json:"summary"` // AI 生成的世界观描述

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorldviewProfile 世界观画像
type WorldviewProfile struct {
	IdentityID string `json:"identity_id"`

	// === 核心信念 ===
	CoreBeliefs []CoreBelief `json:"core_beliefs,omitempty"`

	// === 认知偏差 ===
	CognitiveBiases []CognitiveBias `json:"cognitive_biases,omitempty"`

	// === 世界观稳定性 ===
	StabilityScore float64 `json:"stability_score"` // 世界观稳定度
	OpennessToChange float64 `json:"openness_to_change"` // 对新观念的开放度

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CoreBelief 核心信念
type CoreBelief struct {
	BeliefID     string   `json:"belief_id"`
	Category     string   `json:"category"`     // world/society/future/relationship
	BeliefName   string   `json:"belief_name"`  // 信念名称
	Description  string   `json:"description"`  // 信念描述
	Strength     float64  `json:"strength"`     // 信念强度 (0-1)
	Importance   float64  `json:"importance"`   // 重要性
	Evidence     []string `json:"evidence,omitempty"` // 支持证据
	Contradictions []string `json:"contradictions,omitempty"` // 矛盾点
	FormedAt     time.Time `json:"formed_at"`
	LastReinforced time.Time `json:"last_reinforced"`
}

// CognitiveBias 认知偏差
type CognitiveBias struct {
	BiasID      string   `json:"bias_id"`
	BiasType    string   `json:"bias_type"`    // confirmation/anchoring/availability/etc
	Description string   `json:"description"`
	Impact      float64  `json:"impact"`       // 影响程度
	Triggers    []string `json:"triggers,omitempty"` // 触发场景
}

// NewWorldview 创建默认世界观
func NewWorldview() *Worldview {
	now := time.Now()
	return &Worldview{
		// 世界本质 - 中等唯物，相信自主意识
		Materialism:   0.6,
		Spirituality:  0.3,
		Determinism:   0.3,
		Agency:        0.7,

		// 社会认知 - 中等信任，相信正义
		TrustInPeople:  0.5,
		JusticeBelief:  0.6,
		SocialMobility: 0.5,
		CompetitionView: 0.4,

		// 未来观 - 谨慎乐观
		Optimism:       0.6,
		ChangeBelief:   0.5,
		TechnologyView: 0.6,

		// 人际关系 - 中等个人主义
		Individualism:     0.5,
		RelationshipView:  0.6,

		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewWorldviewProfile 创建默认世界观画像
func NewWorldviewProfile(identityID string) *WorldviewProfile {
	return &WorldviewProfile{
		IdentityID:       identityID,
		CoreBeliefs:      []CoreBelief{},
		CognitiveBiases:  []CognitiveBias{},
		StabilityScore:   0.5,
		OpennessToChange: 0.5,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// GetWorldviewType 获取世界观类型
func (w *Worldview) GetWorldviewType() string {
	// 根据主要维度判断世界观类型
	if w.Materialism > 0.7 && w.Agency > 0.7 {
		return "rationalist" // 理性主义者
	}
	if w.Spirituality > 0.7 {
		return "spiritualist" // 精神主义者
	}
	if w.Determinism > 0.7 {
		return "fatalist" // 宿命论者
	}
	if w.Individualism > 0.7 {
		return "individualist" // 个人主义者
	}
	if w.Individualism < 0.3 {
		return "collectivist" // 集体主义者
	}
	if w.Optimism > 0.7 {
		return "optimist" // 乐观主义者
	}
	if w.Optimism < 0.3 {
		return "pessimist" // 悲观主义者
	}
	return "pragmatist" // 实用主义者
}

// GetWorldviewDescription 获取世界观描述
func (w *Worldview) GetWorldviewDescription() string {
	desc := ""

	// 世界本质
	if w.Materialism > 0.7 {
		desc += "相信物质世界是根本，注重实证和逻辑。"
	} else if w.Spirituality > 0.7 {
		desc += "相信精神力量的存在，注重心灵体验。"
	}

	// 自主性
	if w.Agency > 0.7 {
		desc += "相信人可以掌控自己的命运。"
	} else if w.Determinism > 0.7 {
		desc += "认为命运有其定数。"
	}

	// 社会观
	if w.TrustInPeople > 0.7 {
		desc += "对他人持信任态度。"
	} else if w.TrustInPeople < 0.3 {
		desc += "对人持谨慎态度。"
	}

	// 未来观
	if w.Optimism > 0.7 {
		desc += "对未来充满信心。"
	} else if w.Optimism < 0.3 {
		desc += "对未来有所担忧。"
	}

	if desc == "" {
		desc = "持平衡的世界观，既看重现实也保持开放。"
	}

	return desc
}

// CalculateInfluence 计算世界观对决策的影响
func (w *Worldview) CalculateInfluence() map[string]float64 {
	influence := make(map[string]float64)

	// 冒险倾向 (乐观+自主意识 -> 更愿意冒险)
	influence["risk_taking"] = (w.Optimism + w.Agency) / 2

	// 合作倾向 (信任他人+集体主义 -> 更愿意合作)
	influence["cooperation"] = (w.TrustInPeople + (1 - w.Individualism)) / 2

	// 创新倾向 (拥抱变化+科技态度 -> 更愿意创新)
	influence["innovation"] = (w.ChangeBelief + w.TechnologyView) / 2

	// 社会责任 (正义信念+关系观 -> 更注重社会责任)
	influence["social_responsibility"] = (w.JusticeBelief + w.RelationshipView) / 2

	// 自主决策 (自主意识+个人主义 -> 更倾向于自主决策)
	influence["autonomy"] = (w.Agency + w.Individualism) / 2

	// 长期规划 (乐观+社会流动性信念 -> 更愿意长期规划)
	influence["long_term_planning"] = (w.Optimism + w.SocialMobility) / 2

	return influence
}

// Normalize 归一化
func (w *Worldview) Normalize() {
	normalizeValue := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	w.Materialism = normalizeValue(w.Materialism)
	w.Spirituality = normalizeValue(w.Spirituality)
	w.Determinism = normalizeValue(w.Determinism)
	w.Agency = normalizeValue(w.Agency)
	w.TrustInPeople = normalizeValue(w.TrustInPeople)
	w.JusticeBelief = normalizeValue(w.JusticeBelief)
	w.SocialMobility = normalizeValue(w.SocialMobility)
	w.CompetitionView = normalizeValue(w.CompetitionView)
	w.Optimism = normalizeValue(w.Optimism)
	w.ChangeBelief = normalizeValue(w.ChangeBelief)
	w.TechnologyView = normalizeValue(w.TechnologyView)
	w.Individualism = normalizeValue(w.Individualism)
	w.RelationshipView = normalizeValue(w.RelationshipView)
	w.UpdatedAt = time.Now()
}

// UpdateBelief 更新信念
func (w *Worldview) UpdateBelief(beliefType string, value float64) {
	switch beliefType {
	case "materialism":
		w.Materialism = value
	case "spirituality":
		w.Spirituality = value
	case "determinism":
		w.Determinism = value
	case "agency":
		w.Agency = value
	case "trust_in_people":
		w.TrustInPeople = value
	case "justice_belief":
		w.JusticeBelief = value
	case "social_mobility":
		w.SocialMobility = value
	case "optimism":
		w.Optimism = value
	case "change_belief":
		w.ChangeBelief = value
	case "technology_view":
		w.TechnologyView = value
	case "individualism":
		w.Individualism = value
	case "relationship_view":
		w.RelationshipView = value
	}
	w.Normalize()
}

// LifeView 人生观模型 (v4.1.0)
// 人生观：对人生意义、目标、时间、态度的认知
type LifeView struct {
	// === 人生意义 ===
	MeaningSource []string `json:"meaning_source"` // 人生意义来源 (family/career/pleasure/contribution/knowledge/creation)
	LifeGoal      string   `json:"life_goal"`      // 人生主要目标
	LifePurpose   string   `json:"life_purpose"`   // 人生使命感

	// === 人生阶段 ===
	LifeStage     string  `json:"life_stage"`     // 人生阶段 (youth/early_career/mid_career/mature/late)
	StageSatisfaction float64 `json:"stage_satisfaction"` // 当前阶段满意度

	// === 时间观 (各维度 0-1) ===
	PresentFocus float64 `json:"present_focus"` // 当下关注
	FutureFocus  float64 `json:"future_focus"`  // 未来关注
	PastFocus    float64 `json:"past_focus"`    // 过去关注 (怀旧)

	// === 生活态度 (0-1) ===
	Hedonism   float64 `json:"hedonism"`   // 享乐主义倾向
	Asceticism float64 `json:"asceticism"` // 禁欲主义倾向
	Balance    float64 `json:"balance"`    // 平衡主义倾向

	// === 成功观 ===
	SuccessDefinition string   `json:"success_definition"` // 成功定义
	SuccessMetrics    []string `json:"success_metrics"`    // 成功衡量标准

	// === 幸福观 ===
	HappinessSource []string `json:"happiness_source"` // 幸福来源
	HappinessLevel  float64  `json:"happiness_level"`  // 当前幸福水平

	// === 死亡观 ===
	DeathAcceptance float64 `json:"death_acceptance"` // 对死亡的接受程度
	LegacyWish      string  `json:"legacy_wish"`      // 希望留下的遗产

	// === 人生观描述 ===
	Summary string `json:"summary"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LifeViewProfile 人生观画像
type LifeViewProfile struct {
	IdentityID string `json:"identity_id"`

	// === 人生目标体系 ===
	Goals []LifeGoal `json:"goals,omitempty"`

	// === 人生价值观排序 ===
	ValuePriority []ValuePriority `json:"value_priority,omitempty"`

	// === 人生满意度 ===
	LifeSatisfaction float64 `json:"life_satisfaction"` // 整体人生满意度
	AreaSatisfaction map[string]float64 `json:"area_satisfaction,omitempty"` // 各领域满意度

	// === 人生轨迹 ===
	Milestones []LifeMilestone `json:"milestones,omitempty"` // 人生里程碑

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LifeGoal 人生目标
type LifeGoal struct {
	GoalID       string    `json:"goal_id"`
	Category     string    `json:"category"`     // career/family/health/wealth/growth/contribution
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Importance   float64   `json:"importance"`   // 重要性
	Progress     float64   `json:"progress"`     // 进度 (0-1)
	Status       string    `json:"status"`       // active/achieved/abandoned/paused
	TargetDate   time.Time `json:"target_date,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValuePriority 价值观排序
type ValuePriority struct {
	ValueName string  `json:"value_name"`
	Rank      int     `json:"rank"`
	Importance float64 `json:"importance"`
}

// LifeMilestone 人生里程碑
type LifeMilestone struct {
	MilestoneID   string    `json:"milestone_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`      // education/career/family/health/achievement
	Impact        float64   `json:"impact"`        // 对人生的影响程度
	Positive      bool      `json:"positive"`      // 是否正面
	LessonsLearned []string `json:"lessons_learned,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewLifeView 创建默认人生观
func NewLifeView() *LifeView {
	now := time.Now()
	return &LifeView{
		MeaningSource:     []string{"family", "career", "growth"},
		LifeGoal:          "实现自我价值",
		LifeStage:         "early_career",
		StageSatisfaction: 0.5,
		PresentFocus:      0.4,
		FutureFocus:       0.5,
		PastFocus:         0.1,
		Hedonism:          0.3,
		Asceticism:        0.2,
		Balance:           0.5,
		SuccessDefinition: "家庭幸福与事业成就的平衡",
		SuccessMetrics:    []string{"家庭和睦", "事业稳定", "健康身心"},
		HappinessSource:   []string{"家庭", "工作", "学习"},
		HappinessLevel:    0.6,
		DeathAcceptance:   0.5,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// NewLifeViewProfile 创建默认人生观画像
func NewLifeViewProfile(identityID string) *LifeViewProfile {
	return &LifeViewProfile{
		IdentityID:       identityID,
		Goals:            []LifeGoal{},
		ValuePriority:    []ValuePriority{},
		LifeSatisfaction: 0.6,
		AreaSatisfaction: map[string]float64{
			"career":    0.6,
			"family":    0.7,
			"health":    0.7,
			"wealth":    0.5,
			"growth":    0.6,
			"social":    0.5,
			"leisure":   0.5,
		},
		Milestones: []LifeMilestone{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// GetLifeViewType 获取人生观类型
func (l *LifeView) GetLifeViewType() string {
	// 根据时间观和生活态度判断
	if l.Hedonism > 0.7 && l.PresentFocus > 0.7 {
		return "hedonist" // 享乐主义者
	}
	if l.Asceticism > 0.7 {
		return "ascetic" // 禁欲主义者
	}
	if l.FutureFocus > 0.7 {
		return "ambitious" // 进取者
	}
	if l.PastFocus > 0.7 {
		return "nostalgic" // 怀旧者
	}
	if l.Balance > 0.7 {
		return "balanced" // 平衡主义者
	}
	return "pragmatic" // 务实主义者
}

// GetLifeViewDescription 获取人生观描述
func (l *LifeView) GetLifeViewDescription() string {
	desc := ""

	// 人生意义
	if len(l.MeaningSource) > 0 {
		desc += "人生意义来源于" + joinStrings(l.MeaningSource, "、") + "。"
	}

	// 时间观
	if l.FutureFocus > 0.7 {
		desc += "注重未来规划，追求长期目标。"
	} else if l.PresentFocus > 0.7 {
		desc += "注重当下体验，活在当下。"
	}

	// 生活态度
	if l.Hedonism > 0.7 {
		desc += "追求生活品质和快乐体验。"
	} else if l.Asceticism > 0.7 {
		desc += "注重精神追求，克制物质欲望。"
	} else if l.Balance > 0.7 {
		desc += "追求生活的平衡与和谐。"
	}

	// 幸福感
	if l.HappinessLevel > 0.7 {
		desc += "对生活感到幸福满足。"
	} else if l.HappinessLevel < 0.3 {
		desc += "对生活现状不太满意。"
	}

	if desc == "" {
		desc = "对人生持务实态度，在生活中寻找意义。"
	}

	return desc
}

// CalculateInfluence 计算人生观对决策的影响
func (l *LifeView) CalculateInfluence() map[string]float64 {
	influence := make(map[string]float64)

	// 时间偏好 (未来关注 -> 更愿意延迟满足)
	influence["delayed_gratification"] = l.FutureFocus

	// 风险偏好 (享乐主义 -> 更愿意冒险追求体验)
	influence["risk_for_experience"] = l.Hedonism

	// 工作生活平衡 (平衡主义 -> 更注重平衡)
	influence["work_life_balance"] = l.Balance

	// 目标导向 (未来关注 -> 更目标导向)
	influence["goal_orientation"] = l.FutureFocus

	// 当下享受 (当下关注+享乐主义 -> 更注重当下享受)
	influence["present_enjoyment"] = (l.PresentFocus + l.Hedonism) / 2

	// 意义追求 (有明确目标 -> 更追求意义)
	if l.LifeGoal != "" {
		influence["meaning_seeking"] = 0.7
	} else {
		influence["meaning_seeking"] = 0.3
	}

	return influence
}

// Normalize 归一化
func (l *LifeView) Normalize() {
	normalizeValue := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	l.StageSatisfaction = normalizeValue(l.StageSatisfaction)
	l.PresentFocus = normalizeValue(l.PresentFocus)
	l.FutureFocus = normalizeValue(l.FutureFocus)
	l.PastFocus = normalizeValue(l.PastFocus)
	l.Hedonism = normalizeValue(l.Hedonism)
	l.Asceticism = normalizeValue(l.Asceticism)
	l.Balance = normalizeValue(l.Balance)
	l.HappinessLevel = normalizeValue(l.HappinessLevel)
	l.DeathAcceptance = normalizeValue(l.DeathAcceptance)

	// 确保时间观总和合理
	total := l.PresentFocus + l.FutureFocus + l.PastFocus
	if total > 1.5 {
		factor := 1.0 / total
		l.PresentFocus *= factor
		l.FutureFocus *= factor
		l.PastFocus *= factor
	}

	l.UpdatedAt = time.Now()
}

// UpdateTimeFocus 更新时间观
func (l *LifeView) UpdateTimeFocus(present, future, past float64) {
	l.PresentFocus = present
	l.FutureFocus = future
	l.PastFocus = past
	l.Normalize()
}

// AddGoal 添加人生目标
func (l *LifeView) AddGoal(goal LifeGoal) {
	// 如果有明确目标，更新 LifeGoal
	if goal.Importance > 0.7 && goal.Status == "active" {
		l.LifeGoal = goal.Name
	}
}

// 辅助函数
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
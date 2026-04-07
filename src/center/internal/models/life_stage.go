package models

import "time"

// LifeStageSystem 人生阶段系统 (v4.4.0)
// 完整的人生阶段和事件管理
type LifeStageSystem struct {
	// === 当前阶段 ===
	CurrentStage *LifeStage `json:"current_stage"`

	// === 阶段历史 ===
	StageHistory []LifeStage `json:"stage_history,omitempty"`

	// === 人生事件 ===
	LifeEvents []LifeEvent `json:"life_events,omitempty"`

	// === 人生感悟 ===
	LifeLessons []LifeLesson `json:"life_lessons,omitempty"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LifeStage 人生阶段
type LifeStage struct {
	StageID     string    `json:"stage_id"`
	StageName   string    `json:"stage_name"`   // childhood/adolescence/youth/early_adult/mid_adult/mature/elderly
	StageLabel  string    `json:"stage_label"`  // 阶段标签（中文）
	Description string    `json:"description"`  // 阶段描述
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date,omitempty"` // 空=当前阶段

	// === 阶段特征 ===
	Challenges    []string `json:"challenges,omitempty"`    // 阶段挑战
	Opportunities []string `json:"opportunities,omitempty"` // 阶段机遇
	Goals         []string `json:"goals,omitempty"`         // 阶段目标
	Tasks         []string `json:"tasks,omitempty"`         // 发展任务

	// === 阶段属性 ===
	StageAge     int     `json:"stage_age"`     // 该阶段年龄
	StageDuration int    `json:"stage_duration"` // 阶段持续年数
	Completeness float64 `json:"completeness"`  // 完成度 (0-1)
	Satisfaction float64 `json:"satisfaction"`  // 满意度 (0-1)

	// === 发展指标 ===
	DevelopmentMetrics DevelopmentMetrics `json:"development_metrics"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DevelopmentMetrics 发展指标
type DevelopmentMetrics struct {
	// 身心发展
	PhysicalHealth   float64 `json:"physical_health"`   // 身体健康 (0-1)
	MentalHealth     float64 `json:"mental_health"`     // 心理健康 (0-1)
	CognitiveGrowth  float64 `json:"cognitive_growth"`  // 认知成长 (0-1)
	EmotionalMaturity float64 `json:"emotional_maturity"` // 情绪成熟度 (0-1)

	// 社会发展
	SocialDevelopment float64 `json:"social_development"` // 社会发展 (0-1)
	RelationshipQuality float64 `json:"relationship_quality"` // 关系质量 (0-1)
	CommunityIntegration float64 `json:"community_integration"` // 社区融入 (0-1)

	// 职业发展
	CareerProgress   float64 `json:"career_progress"`   // 职业进展 (0-1)
	FinancialStability float64 `json:"financial_stability"` // 财务稳定 (0-1)
	SkillDevelopment float64 `json:"skill_development"` // 技能发展 (0-1)

	// 自我实现
	SelfAwareness    float64 `json:"self_awareness"`    // 自我认知 (0-1)
	PurposeClarity   float64 `json:"purpose_clarity"`   // 人生目标清晰度 (0-1)
	LifeSatisfaction float64 `json:"life_satisfaction"` // 人生满意度 (0-1)
}

// LifeEvent 人生事件
type LifeEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`     // birth/education/career/family/health/achievement/loss/relocation/relationship/financial/spiritual
	EventCategory string    `json:"event_category"` // milestone/transition/challenge/achievement/loss
	Title         string    `json:"title"`
	Description   string    `json:"description"`

	// === 时间信息 ===
	EventDate   time.Time `json:"event_date"`
	EventAge    int       `json:"event_age"`     // 事件发生时的年龄
	EventType_  string    `json:"event_type_"`   // single/recurring/ongoing
	Duration    int       `json:"duration"`      // 持续时间（天），0=瞬时事件

	// === 影响评估 ===
	ImpactLevel      float64             `json:"impact_level"`       // 影响程度 (0-1)
	ImpactValence    float64             `json:"impact_valence"`     // 正面/负面 (-1到1)
	ImpactDuration   int                 `json:"impact_duration"`    // 影响持续时间（月）
	ImpactAreas      map[string]float64  `json:"impact_areas,omitempty"` // 影响领域

	// === 情绪影响 ===
	EmotionalImpact  map[string]float64 `json:"emotional_impact,omitempty"` // 对七情的影响
	PeakEmotion      string             `json:"peak_emotion"`              // 主要情绪
	EmotionalRecovery float64            `json:"emotional_recovery"`        // 情绪恢复度 (0-1)

	// === 人生感悟 ===
	LessonsLearned   []string `json:"lessons_learned,omitempty"`
	PerspectiveChange string  `json:"perspective_change"` // 视角变化描述
	ValueShift       map[string]float64 `json:"value_shift,omitempty"` // 价值观变化

	// === 关联人物 ===
	PeopleInvolved []string `json:"people_involved,omitempty"` // 相关人物

	// === 后续影响 ===
	Consequences    []string `json:"consequences,omitempty"`    // 后续影响
	RelatedEvents   []string `json:"related_events,omitempty"` // 相关事件ID

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LifeLesson 人生感悟
type LifeLesson struct {
	LessonID      string    `json:"lesson_id"`
	SourceEvent   string    `json:"source_event"`   // 来源事件ID
	SourceStage   string    `json:"source_stage"`   // 来源阶段
	Title         string    `json:"title"`          // 感悟标题
	Description   string    `json:"description"`    // 感悟描述
	Insight       string    `json:"insight"`        // 核心洞见
	Application   string    `json:"application"`    // 如何应用

	// === 感悟分类 ===
	Category      string   `json:"category"`       // relationship/career/health/finance/growth/meaning
	Tags          []string `json:"tags,omitempty"`

	// === 重要性 ===
	Importance    float64 `json:"importance"`     // 重要性 (0-1)
	Applicability float64 `json:"applicability"`  // 适用性 (0-1)
	Verified      bool    `json:"verified"`       // 是否已验证

	// === 时间属性 ===
	LearnedAt    time.Time `json:"learned_at"`
	LastApplied  time.Time `json:"last_applied,omitempty"`
	ApplyCount   int       `json:"apply_count"`    // 应用次数
}

// LifeStageProfile 人生阶段画像
type LifeStageProfile struct {
	IdentityID string `json:"identity_id"`

	// === 人生轨迹 ===
	LifeTrajectory LifeTrajectory `json:"life_trajectory"`

	// === 发展里程碑 ===
	Milestones []LifeMilestone `json:"milestones,omitempty"`

	// === 转折点 ===
	TurningPoints []TurningPoint `json:"turning_points,omitempty"`

	// === 人生总结 ===
	LifeNarrative string `json:"life_narrative"` // 人生叙事
	WisdomAccumulated float64 `json:"wisdom_accumulated"` // 智慧积累度

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LifeTrajectory 人生轨迹
type LifeTrajectory struct {
	// 轨迹方向
	Direction string `json:"direction"` // upward/stable/downward/fluctuating

	// 轨迹指标
	OverallGrowth    float64 `json:"overall_growth"`    // 整体成长
	ResilienceScore  float64 `json:"resilience_score"`  // 韧性得分
	AdaptabilityScore float64 `json:"adaptability_score"` // 适应力得分

	// 关键指标变化
	SatisfactionTrend []float64 `json:"satisfaction_trend,omitempty"` // 满意度趋势
	GrowthTrend       []float64 `json:"growth_trend,omitempty"`       // 成长趋势

	// 轨迹预测
	ExpectedDirection string  `json:"expected_direction"` // 预期方向
	Confidence        float64 `json:"confidence"`         // 置信度
}

// LifeMilestone 人生里程碑
type LifeMilestone struct {
	MilestoneID   string    `json:"milestone_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`      // personal/educational/career/family/social
	AchievedAt    time.Time `json:"achieved_at"`
	Age           int       `json:"age"`
	Significance  float64   `json:"significance"`  // 重要程度 (0-1)
	Celebration   string    `json:"celebration"`   // 如何庆祝
	Memories      []string  `json:"memories,omitempty"`
}

// TurningPoint 人生转折点
type TurningPoint struct {
	TurningPointID string    `json:"turning_point_id"`
	Event          string    `json:"event"`          // 转折事件
	BeforeState    string    `json:"before_state"`   // 之前状态
	AfterState     string    `json:"after_state"`    // 之后状态
	Trigger        string    `json:"trigger"`        // 触发因素
	Timestamp      time.Time `json:"timestamp"`
	Significance   float64   `json:"significance"`
}

// 人生阶段定义
var LifeStageDefinitions = map[string]struct {
	Label       string
	AgeRange    string
	Description string
	Challenges  []string
	Opportunities []string
	Tasks       []string
}{
	"childhood": {
		Label:       "童年期",
		AgeRange:    "0-12岁",
		Description: "依赖成长期，建立基本信任和安全感",
		Challenges:  []string{"分离焦虑", "自我意识形成", "社会化"},
		Opportunities: []string{"探索世界", "学习能力", "人格基础"},
		Tasks:       []string{"建立信任", "发展自主性", "培养主动性"},
	},
	"adolescence": {
		Label:       "青春期",
		AgeRange:    "13-18岁",
		Description: "身份认同期，自我探索和独立",
		Challenges:  []string{"身份认同", "独立性", "同伴压力", "学业压力"},
		Opportunities: []string{"自我发现", "建立价值观", "发展特长"},
		Tasks:       []string{"身份形成", "价值确立", "学业规划"},
	},
	"youth": {
		Label:       "青年期",
		AgeRange:    "19-25岁",
		Description: "探索奠基期，教育完成和职业起步",
		Challenges:  []string{"职业选择", "经济独立", "情感发展"},
		Opportunities: []string{"职业探索", "技能积累", "建立关系"},
		Tasks:       []string{"完成教育", "职业起步", "建立亲密关系"},
	},
	"early_adult": {
		Label:       "成年早期",
		AgeRange:    "26-35岁",
		Description: "建设期，事业和家庭奠基",
		Challenges:  []string{"事业压力", "家庭建立", "工作生活平衡"},
		Opportunities: []string{"职业发展", "家庭成长", "财富积累"},
		Tasks:       []string{"事业突破", "组建家庭", "建立社会地位"},
	},
	"mid_adult": {
		Label:       "成年中期",
		AgeRange:    "36-50岁",
		Description: "责任期，事业高峰和家庭责任",
		Challenges:  []string{"中年危机", "子女教育", "父母赡养", "健康关注"},
		Opportunities: []string{"事业成熟", "领导角色", "财富积累"},
		Tasks:       []string{"事业巩固", "子女培养", "健康管理"},
	},
	"mature": {
		Label:       "成熟期",
		AgeRange:    "51-65岁",
		Description: "收获期，智慧积累和传承",
		Challenges:  []string{"健康衰退", "空巢综合症", "退休准备"},
		Opportunities: []string{"经验传承", "人生智慧", "享受生活"},
		Tasks:       []string{"传承经验", "培养爱好", "退休规划"},
	},
	"elderly": {
		Label:       "老年期",
		AgeRange:    "65岁以上",
		Description: "回顾期，人生总结和精神传承",
		Challenges:  []string{"健康问题", "孤独感", "死亡意识"},
		Opportunities: []string{"人生回顾", "精神传承", "享受晚年"},
		Tasks:       []string{"接受衰老", "传承智慧", "精神满足"},
	},
}

// NewLifeStageSystem 创建默认人生阶段系统
func NewLifeStageSystem() *LifeStageSystem {
	now := time.Now()
	return &LifeStageSystem{
		CurrentStage: &LifeStage{
			StageID:        generateStageID(),
			StageName:      "early_adult",
			StageLabel:     "成年早期",
			Description:    "建设期，事业和家庭奠基",
			Challenges:     []string{"事业压力", "家庭建立", "工作生活平衡"},
			Opportunities:  []string{"职业发展", "家庭成长", "财富积累"},
			Goals:          []string{},
			Tasks:          []string{"事业突破", "组建家庭", "建立社会地位"},
			Completeness:   0.3,
			Satisfaction:   0.6,
			DevelopmentMetrics: DevelopmentMetrics{
				PhysicalHealth:      0.8,
				MentalHealth:        0.7,
				CognitiveGrowth:     0.6,
				EmotionalMaturity:   0.6,
				SocialDevelopment:   0.6,
				RelationshipQuality: 0.6,
				CareerProgress:      0.5,
				FinancialStability:  0.5,
				SkillDevelopment:    0.6,
				SelfAwareness:       0.6,
				PurposeClarity:      0.5,
				LifeSatisfaction:    0.6,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		StageHistory: []LifeStage{},
		LifeEvents:   []LifeEvent{},
		LifeLessons:  []LifeLesson{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewLifeEvent 创建人生事件
func NewLifeEvent(eventType, title string) *LifeEvent {
	now := time.Now()
	return &LifeEvent{
		EventID:          generateEventID(),
		EventType:        eventType,
		EventCategory:    "milestone",
		Title:            title,
		EventDate:        now,
		ImpactLevel:      0.5,
		ImpactValence:    0,
		ImpactAreas:      make(map[string]float64),
		EmotionalImpact:  make(map[string]float64),
		ValueShift:       make(map[string]float64),
		LessonsLearned:   []string{},
		PeopleInvolved:   []string{},
		Consequences:     []string{},
		RelatedEvents:    []string{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// GetStageDefinition 获取阶段定义
func GetStageDefinition(stageName string) (label, ageRange, description string, challenges, opportunities, tasks []string) {
	if def, ok := LifeStageDefinitions[stageName]; ok {
		return def.Label, def.AgeRange, def.Description, def.Challenges, def.Opportunities, def.Tasks
	}
	return stageName, "", "", nil, nil, nil
}

// CalculateOverallDevelopment 计算整体发展程度
func (d *DevelopmentMetrics) CalculateOverallDevelopment() float64 {
	// 加权平均
	weights := map[string]float64{
		"physical_health":   0.1,
		"mental_health":     0.1,
		"cognitive_growth":  0.08,
		"emotional_maturity": 0.08,
		"social_development": 0.1,
		"relationship_quality": 0.08,
		"career_progress":  0.12,
		"financial_stability": 0.08,
		"skill_development": 0.08,
		"self_awareness":   0.06,
		"purpose_clarity":  0.06,
		"life_satisfaction": 0.06,
	}

	total := 0.0
	weightSum := 0.0

	total += d.PhysicalHealth * weights["physical_health"]
	total += d.MentalHealth * weights["mental_health"]
	total += d.CognitiveGrowth * weights["cognitive_growth"]
	total += d.EmotionalMaturity * weights["emotional_maturity"]
	total += d.SocialDevelopment * weights["social_development"]
	total += d.RelationshipQuality * weights["relationship_quality"]
	total += d.CareerProgress * weights["career_progress"]
	total += d.FinancialStability * weights["financial_stability"]
	total += d.SkillDevelopment * weights["skill_development"]
	total += d.SelfAwareness * weights["self_awareness"]
	total += d.PurposeClarity * weights["purpose_clarity"]
	total += d.LifeSatisfaction * weights["life_satisfaction"]

	for _, w := range weights {
		weightSum += w
	}

	return total / weightSum
}

// GetDominantChallenge 获取主要挑战
func (s *LifeStage) GetDominantChallenge() string {
	if len(s.Challenges) == 0 {
		return ""
	}
	return s.Challenges[0]
}

// GetDominantGoal 获取主要目标
func (s *LifeStage) GetDominantGoal() string {
	if len(s.Goals) == 0 {
		return ""
	}
	return s.Goals[0]
}

// IsCompleted 阶段是否完成
func (s *LifeStage) IsCompleted() bool {
	return !s.EndDate.IsZero()
}

// GetDurationYears 获取阶段持续年数
func (s *LifeStage) GetDurationYears() int {
	if s.IsCompleted() {
		return int(s.EndDate.Sub(s.StartDate).Hours() / (24 * 365))
	}
	return int(time.Since(s.StartDate).Hours() / (24 * 365))
}

// GetEventTypeName 获取事件类型名称
func (e *LifeEvent) GetEventTypeName() string {
	names := map[string]string{
		"birth":        "出生",
		"education":    "教育",
		"career":       "职业",
		"family":       "家庭",
		"health":       "健康",
		"achievement":  "成就",
		"loss":         "失去",
		"relocation":   "迁居",
		"relationship": "关系",
		"financial":    "财务",
		"spiritual":    "精神",
	}
	if name, ok := names[e.EventType]; ok {
		return name
	}
	return e.EventType
}

// IsPositive 是否正面事件
func (e *LifeEvent) IsPositive() bool {
	return e.ImpactValence > 0.3
}

// IsNegative 是否负面事件
func (e *LifeEvent) IsNegative() bool {
	return e.ImpactValence < -0.3
}

// IsLifeChanging 是否改变人生的事件
func (e *LifeEvent) IsLifeChanging() bool {
	return e.ImpactLevel > 0.7
}

// Normalize 归一化
func (s *LifeStageSystem) Normalize() {
	if s.CurrentStage != nil {
		s.CurrentStage.Completeness = clamp01(s.CurrentStage.Completeness)
		s.CurrentStage.Satisfaction = clamp01(s.CurrentStage.Satisfaction)
		s.CurrentStage.DevelopmentMetrics.normalize()
	}

	for i := range s.LifeEvents {
		s.LifeEvents[i].ImpactLevel = clamp01(s.LifeEvents[i].ImpactLevel)
		s.LifeEvents[i].ImpactValence = clamp(s.LifeEvents[i].ImpactValence, -1, 1)
		s.LifeEvents[i].EmotionalRecovery = clamp01(s.LifeEvents[i].EmotionalRecovery)
	}

	for i := range s.LifeLessons {
		s.LifeLessons[i].Importance = clamp01(s.LifeLessons[i].Importance)
		s.LifeLessons[i].Applicability = clamp01(s.LifeLessons[i].Applicability)
	}

	s.UpdatedAt = time.Now()
}

func (d *DevelopmentMetrics) normalize() {
	d.PhysicalHealth = clamp01(d.PhysicalHealth)
	d.MentalHealth = clamp01(d.MentalHealth)
	d.CognitiveGrowth = clamp01(d.CognitiveGrowth)
	d.EmotionalMaturity = clamp01(d.EmotionalMaturity)
	d.SocialDevelopment = clamp01(d.SocialDevelopment)
	d.RelationshipQuality = clamp01(d.RelationshipQuality)
	d.CareerProgress = clamp01(d.CareerProgress)
	d.FinancialStability = clamp01(d.FinancialStability)
	d.SkillDevelopment = clamp01(d.SkillDevelopment)
	d.SelfAwareness = clamp01(d.SelfAwareness)
	d.PurposeClarity = clamp01(d.PurposeClarity)
	d.LifeSatisfaction = clamp01(d.LifeSatisfaction)
}

// 辅助函数
func generateStageID() string {
	return "stage_" + time.Now().Format("20060102150405")
}

func generateEventID() string {
	return "event_" + time.Now().Format("20060102150405")
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
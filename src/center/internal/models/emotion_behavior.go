package models

import "time"

// EmotionBehaviorSystem 情绪行为联动系统 (v4.5.0)
// 管理情绪对决策、表达、行为的影响
type EmotionBehaviorSystem struct {
	// === 决策影响 ===
	DecisionInfluence *EmotionDecisionInfluence `json:"decision_influence"`

	// === 表达影响 ===
	ExpressionInfluence *EmotionalExpressionInfluence `json:"expression_influence"`

	// === 行为触发 ===
	BehaviorTriggers []EmotionTriggeredBehavior `json:"behavior_triggers,omitempty"`

	// === 应对策略 ===
	CopingStrategies []CopingStrategy `json:"coping_strategies,omitempty"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmotionDecisionInfluence 情绪决策影响
type EmotionDecisionInfluence struct {
	// 风险偏好
	RiskTolerance     float64 `json:"risk_tolerance"`      // 风险承受度 (0-1)
	RiskAversion      float64 `json:"risk_aversion"`       // 风险规避度 (0-1)
	ImpulseControl    float64 `json:"impulse_control"`     // 冲动控制 (0-1)
	DelayedGratification float64 `json:"delayed_gratification"` // 延迟满足能力 (0-1)

	// 社交决策
	SocialApproach    float64 `json:"social_approach"`     // 社交趋近 (0-1)
	SocialAvoidance   float64 `json:"social_avoidance"`    // 社交回避 (0-1)
	TrustLevel        float64 `json:"trust_level"`         // 信任水平 (0-1)
	CooperationTendency float64 `json:"cooperation_tendency"` // 合作倾向 (0-1)

	// 决策速度
	DecisionSpeed     float64 `json:"decision_speed"`      // 决策速度 (0-1)
	DeliberationLevel float64 `json:"deliberation_level"`  // 深思程度 (0-1)
	Decisiveness      float64 `json:"decisiveness"`        // 果断程度 (0-1)

	// 选择偏好
	NoveltySeeking    float64 `json:"novelty_seeking"`     // 求新倾向 (0-1)
	FamiliarityPreference float64 `json:"familiarity_preference"` // 熟悉偏好 (0-1)
	QualityFocus      float64 `json:"quality_focus"`       // 质量关注 (0-1)
	PriceSensitivity  float64 `json:"price_sensitivity"`   // 价格敏感 (0-1)

	// 情绪来源
	DominantEmotion   string  `json:"dominant_emotion"`    // 主导情绪
	EmotionIntensity  float64 `json:"emotion_intensity"`   // 情绪强度
}

// EmotionalExpressionInfluence 情绪表达影响
type EmotionalExpressionInfluence struct {
	// 语言风格
	ToneStyle        string  `json:"tone_style"`         // 语调风格 (warm/formal/casual/enthusiastic/reserved)
	FormalityLevel   float64 `json:"formality_level"`    // 正式程度 (0-1)
	WarmthLevel      float64 `json:"warmth_level"`       // 温暖程度 (0-1)
	EnthusiasmLevel  float64 `json:"enthusiasm_level"`   // 热情程度 (0-1)

	// 措辞偏好
	WordChoice       string  `json:"word_choice"`        // 措辞偏好 (positive/negative/neutral/mixed)
	SentenceLength   string  `json:"sentence_length"`    // 句子长度 (short/medium/long/varied)
	ComplexityLevel  float64 `json:"complexity_level"`   // 复杂程度 (0-1)
	MetaphorUse      float64 `json:"metaphor_use"`       // 比喻使用 (0-1)

	// 表情使用
	EmojiUsage       float64 `json:"emoji_usage"`        // 表情使用频率 (0-1)
	EmojiType        string  `json:"emoji_type"`         // 表情类型 (happy/sad/neutral/varied)
	ExclamationUse   float64 `json:"exclamation_use"`    // 感叹号使用 (0-1)

	// 回应风格
	ResponseSpeed    string  `json:"response_speed"`     // 回应速度 (immediate/delayed/thoughtful)
	Proactiveness    float64 `json:"proactiveness"`      // 主动性 (0-1)
	DetailLevel      float64 `json:"detail_level"`       // 细节程度 (0-1)
	HumorLevel       float64 `json:"humor_level"`        // 幽默程度 (0-1)

	// 非语言表达
	VoiceTone        string  `json:"voice_tone"`         // 声音语调 (high/low/modulated)
	SpeechSpeed      string  `json:"speech_speed"`       // 语速 (fast/normal/slow)
	VolumeLevel      string  `json:"volume_level"`       // 音量 (loud/normal/soft)
	PauseFrequency   float64 `json:"pause_frequency"`    // 停顿频率 (0-1)

	// 情绪来源
	UnderlyingEmotion string `json:"underlying_emotion"` // 底层情绪
	ExpressionTendency string `json:"expression_tendency"` // 表达倾向 (express/suppress/mask)
}

// EmotionTriggeredBehavior 情绪触发行为
type EmotionTriggeredBehavior struct {
	BehaviorID      string    `json:"behavior_id"`
	TriggerEmotion  string    `json:"trigger_emotion"`   // 触发情绪
	IntensityThreshold float64 `json:"intensity_threshold"` // 强度阈值

	// 行为类型
	BehaviorType    string    `json:"behavior_type"`     // action/communication/withdrawal/approach/seeking
	BehaviorName    string    `json:"behavior_name"`     // 行为名称
	BehaviorDescription string `json:"behavior_description"` // 行为描述

	// 行动倾向
	ActionTendency  string    `json:"action_tendency"`   // approach/avoid/freeze/fight/flight
	UrgencyLevel    float64   `json:"urgency_level"`     // 紧迫程度 (0-1)
	Automaticity    float64   `json:"automaticity"`      // 自动化程度 (0-1)

	// 触发条件
	Triggers        []string  `json:"triggers,omitempty"` // 触发条件
	ContextFactors  []string  `json:"context_factors,omitempty"` // 情境因素

	// 调节因素
	Moderators      []string  `json:"moderators,omitempty"` // 调节因素
	Inhibitors      []string  `json:"inhibitors,omitempty"` // 抑制因素

	// 后果
	ImmediateEffect string    `json:"immediate_effect"`  // 即时效果
	LongTermEffect  string    `json:"long_term_effect"`  // 长期效果

	// 时间属性
	TriggeredAt     time.Time `json:"triggered_at,omitempty"`
	Duration        int       `json:"duration"`          // 持续时间（秒）
}

// CopingStrategy 应对策略
type CopingStrategy struct {
	StrategyID      string   `json:"strategy_id"`
	StrategyName    string   `json:"strategy_name"`
	StrategyType    string   `json:"strategy_type"`     // problem-focused/emotion-focused/avoidance/seeking-support

	// 适用情绪
	TargetEmotions  []string `json:"target_emotions"`   // 目标情绪
	Effectiveness   float64  `json:"effectiveness"`     // 有效性 (0-1)

	// 策略步骤
	Steps           []string `json:"steps,omitempty"`   // 应对步骤

	// 使用条件
	Prerequisites   []string `json:"prerequisites,omitempty"` // 前提条件
	Contraindications []string `json:"contraindications,omitempty"` // 禁忌情况

	// 效果
	ExpectedOutcome string   `json:"expected_outcome"`  // 预期效果
	SideEffects     []string `json:"side_effects,omitempty"` // 副作用

	// 使用统计
	UseCount        int      `json:"use_count"`         // 使用次数
	SuccessRate     float64  `json:"success_rate"`      // 成功率
	LastUsed        time.Time `json:"last_used,omitempty"`

	// 时间属性
	CreatedAt       time.Time `json:"created_at"`
}

// EmotionalBehaviorProfile 情绪行为画像
type EmotionalBehaviorProfile struct {
	IdentityID string `json:"identity_id"`

	// === 行为模式 ===
	BehaviorPatterns []EmotionalBehaviorPattern `json:"behavior_patterns,omitempty"`

	// === 表达习惯 ===
	ExpressionHabbits []ExpressionHabit `json:"expression_habits,omitempty"`

	// === 应对模式 ===
	CopingPatterns []CopingPattern `json:"coping_patterns,omitempty"`

	// === 触发历史 ===
	TriggerHistory []BehaviorTriggerRecord `json:"trigger_history,omitempty"`

	// === 统计指标 ===
	ImpulsiveActionRate float64 `json:"impulsive_action_rate"` // 冲动行为率
	EmotionalRegulation float64 `json:"emotional_regulation"` // 情绪调节能力
	ExpressionAuthenticity float64 `json:"expression_authenticity"` // 表达真实性

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmotionalBehaviorPattern 情绪行为模式
type EmotionalBehaviorPattern struct {
	PatternID       string   `json:"pattern_id"`
	PatternName     string   `json:"pattern_name"`

	// 触发情绪组合
	EmotionCombination []string `json:"emotion_combination"` // 情绪组合

	// 行为序列
	BehaviorSequence []string `json:"behavior_sequence"` // 行为序列

	// 模式特征
	Frequency       float64  `json:"frequency"`         // 频率
	Intensity       float64  `json:"intensity"`         // 强度
	Adaptiveness    float64  `json:"adaptiveness"`      // 适应性

	// 上下文
	CommonContexts []string `json:"common_contexts,omitempty"` // 常见情境
}

// ExpressionHabit 表达习惯
type ExpressionHabit struct {
	HabitID         string   `json:"habit_id"`
	HabitName       string   `json:"habit_name"`

	// 触发情绪
	TriggerEmotions []string `json:"trigger_emotions"`

	// 表达方式
	ExpressionMode  string   `json:"expression_mode"`   // verbal/non-verbal/written
	ExpressionStyle string   `json:"expression_style"` // direct/indirect/subtle/overt

	// 特征
	Frequency       float64  `json:"frequency"`         // 频率
	Effectiveness   float64  `json:"effectiveness"`     // 有效性

	// 社会影响
	ImpactOnOthers  string   `json:"impact_on_others"`  // 对他人的影响
}

// CopingPattern 应对模式
type CopingPattern struct {
	PatternID       string   `json:"pattern_id"`
	PatternName     string   `json:"pattern_name"`

	// 应对风格
	CopingStyle     string   `json:"coping_style"`      // active/passive/avoidant/social

	// 常用策略
	PreferredStrategies []string `json:"preferred_strategies"` // 偏好策略

	// 效果
	EffectivenessScore float64 `json:"effectiveness_score"` // 有效性得分
	ConsistencyScore   float64 `json:"consistency_score"`   // 一致性得分

	// 适用情境
	ApplicableSituations []string `json:"applicable_situations,omitempty"`
}

// BehaviorTriggerRecord 行为触发记录
type BehaviorTriggerRecord struct {
	RecordID        string    `json:"record_id"`
	Timestamp       time.Time `json:"timestamp"`

	// 触发情绪
	TriggerEmotion  string    `json:"trigger_emotion"`
	EmotionIntensity float64  `json:"emotion_intensity"`

	// 触发行为
	TriggeredBehavior string  `json:"triggered_behavior"`
	BehaviorType    string    `json:"behavior_type"`

	// 情境
	Context         string    `json:"context"`

	// 结果
	Outcome         string    `json:"outcome"`           // positive/negative/neutral
	LearnedLesson   string    `json:"learned_lesson,omitempty"` // 学到的教训

	// 是否冲动
	WasImpulsive    bool      `json:"was_impulsive"`
	RegulationUsed  string    `json:"regulation_used,omitempty"` // 使用的调节策略
}

// NewEmotionBehaviorSystem 创建默认情绪行为系统
func NewEmotionBehaviorSystem() *EmotionBehaviorSystem {
	now := time.Now()
	return &EmotionBehaviorSystem{
		DecisionInfluence: &EmotionDecisionInfluence{
			RiskTolerance:         0.5,
			RiskAversion:          0.5,
			ImpulseControl:        0.6,
			DelayedGratification:  0.5,
			SocialApproach:        0.6,
			SocialAvoidance:       0.4,
			TrustLevel:            0.5,
			CooperationTendency:   0.6,
			DecisionSpeed:         0.5,
			DeliberationLevel:     0.5,
			Decisiveness:          0.5,
			NoveltySeeking:        0.5,
			FamiliarityPreference: 0.5,
			QualityFocus:          0.5,
			PriceSensitivity:      0.5,
			DominantEmotion:       "neutral",
			EmotionIntensity:      0.3,
		},
		ExpressionInfluence: &EmotionalExpressionInfluence{
			ToneStyle:          "casual",
			FormalityLevel:     0.4,
			WarmthLevel:        0.6,
			EnthusiasmLevel:    0.5,
			WordChoice:         "neutral",
			SentenceLength:     "medium",
			ComplexityLevel:    0.5,
			MetaphorUse:        0.3,
			EmojiUsage:         0.4,
			EmojiType:          "varied",
			ExclamationUse:     0.3,
			ResponseSpeed:      "thoughtful",
			Proactiveness:      0.5,
			DetailLevel:        0.5,
			HumorLevel:         0.4,
			VoiceTone:          "modulated",
			SpeechSpeed:        "normal",
			VolumeLevel:        "normal",
			PauseFrequency:     0.3,
			UnderlyingEmotion:  "neutral",
			ExpressionTendency: "express",
		},
		BehaviorTriggers:  []EmotionTriggeredBehavior{},
		CopingStrategies:  []CopingStrategy{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// NewEmotionalBehaviorProfile 创建情绪行为画像
func NewEmotionalBehaviorProfile(identityID string) *EmotionalBehaviorProfile {
	now := time.Now()
	return &EmotionalBehaviorProfile{
		IdentityID:              identityID,
		BehaviorPatterns:        []EmotionalBehaviorPattern{},
		ExpressionHabbits:       []ExpressionHabit{},
		CopingPatterns:          []CopingPattern{},
		TriggerHistory:          []BehaviorTriggerRecord{},
		ImpulsiveActionRate:     0.3,
		EmotionalRegulation:     0.5,
		ExpressionAuthenticity:  0.6,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
}

// Normalize 归一化
func (s *EmotionBehaviorSystem) Normalize() {
	if s.DecisionInfluence != nil {
		s.DecisionInfluence.normalize()
	}
	if s.ExpressionInfluence != nil {
		s.ExpressionInfluence.normalize()
	}
	s.UpdatedAt = time.Now()
}

func (d *EmotionDecisionInfluence) normalize() {
	d.RiskTolerance = clamp01(d.RiskTolerance)
	d.RiskAversion = clamp01(d.RiskAversion)
	d.ImpulseControl = clamp01(d.ImpulseControl)
	d.DelayedGratification = clamp01(d.DelayedGratification)
	d.SocialApproach = clamp01(d.SocialApproach)
	d.SocialAvoidance = clamp01(d.SocialAvoidance)
	d.TrustLevel = clamp01(d.TrustLevel)
	d.CooperationTendency = clamp01(d.CooperationTendency)
	d.DecisionSpeed = clamp01(d.DecisionSpeed)
	d.DeliberationLevel = clamp01(d.DeliberationLevel)
	d.Decisiveness = clamp01(d.Decisiveness)
	d.NoveltySeeking = clamp01(d.NoveltySeeking)
	d.FamiliarityPreference = clamp01(d.FamiliarityPreference)
	d.QualityFocus = clamp01(d.QualityFocus)
	d.PriceSensitivity = clamp01(d.PriceSensitivity)
	d.EmotionIntensity = clamp01(d.EmotionIntensity)
}

func (e *EmotionalExpressionInfluence) normalize() {
	e.FormalityLevel = clamp01(e.FormalityLevel)
	e.WarmthLevel = clamp01(e.WarmthLevel)
	e.EnthusiasmLevel = clamp01(e.EnthusiasmLevel)
	e.ComplexityLevel = clamp01(e.ComplexityLevel)
	e.MetaphorUse = clamp01(e.MetaphorUse)
	e.EmojiUsage = clamp01(e.EmojiUsage)
	e.ExclamationUse = clamp01(e.ExclamationUse)
	e.Proactiveness = clamp01(e.Proactiveness)
	e.DetailLevel = clamp01(e.DetailLevel)
	e.HumorLevel = clamp01(e.HumorLevel)
	e.PauseFrequency = clamp01(e.PauseFrequency)
}

// IsHighRiskTolerance 是否高风险偏好
func (d *EmotionDecisionInfluence) IsHighRiskTolerance() bool {
	return d.RiskTolerance > 0.6
}

// IsImpulsive 是否冲动
func (d *EmotionDecisionInfluence) IsImpulsive() bool {
	return d.ImpulseControl < 0.4 && d.DecisionSpeed > 0.6
}

// IsSociallyApproach 是否社交趋近
func (d *EmotionDecisionInfluence) IsSociallyApproach() bool {
	return d.SocialApproach > d.SocialAvoidance
}

// IsExpressive 是否表达型
func (e *EmotionalExpressionInfluence) IsExpressive() bool {
	return e.WarmthLevel > 0.6 && e.EnthusiasmLevel > 0.5
}

// IsReserved 是否内敛型
func (e *EmotionalExpressionInfluence) IsReserved() bool {
	return e.WarmthLevel < 0.4 && e.EnthusiasmLevel < 0.4
}

// GetDecisionStyle 获取决策风格
func (d *EmotionDecisionInfluence) GetDecisionStyle() string {
	if d.DecisionSpeed > 0.7 && d.DeliberationLevel < 0.4 {
		return "intuitive" // 直觉型
	} else if d.DeliberationLevel > 0.7 && d.DecisionSpeed < 0.4 {
		return "analytical" // 分析型
	} else if d.Decisiveness > 0.6 {
		return "decisive" // 果断型
	}
	return "balanced" // 平衡型
}

// GetCommunicationStyle 获取沟通风格
func (e *EmotionalExpressionInfluence) GetCommunicationStyle() string {
	if e.WarmthLevel > 0.6 && e.HumorLevel > 0.5 {
		return "warm_friendly" // 温暖友好
	} else if e.FormalityLevel > 0.7 {
		return "professional" // 专业正式
	} else if e.DetailLevel > 0.7 {
		return "detailed" // 详细周全
	} else if e.ResponseSpeed == "immediate" {
		return "responsive" // 快速响应
	}
	return "balanced" // 平衡
}
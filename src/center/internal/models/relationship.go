package models

import "time"

// RelationshipSystem 人际关系系统 (v4.6.0)
// 管理人际关系网络和社交决策
type RelationshipSystem struct {
	// === 核心关系 ===
	Relationships []Relationship `json:"relationships,omitempty"`

	// === 社交网络 ===
	SocialNetwork *SocialNetwork `json:"social_network"`

	// === 关系模板 ===
	RelationshipTemplates []RelationshipTemplate `json:"relationship_templates,omitempty"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Relationship 人际关系
type Relationship struct {
	RelationshipID string    `json:"relationship_id"`
	PersonID       string    `json:"person_id"`        // 对方ID
	PersonName     string    `json:"person_name"`      // 对方姓名
	RelationshipType string  `json:"relationship_type"` // family/friend/colleague/romantic/acquaintance/mentor/mentee/other

	// === 关系属性 ===
	Intimacy       float64   `json:"intimacy"`         // 亲密度 (0-1)
	Trust          float64   `json:"trust"`            // 信任度 (0-1)
	Respect        float64   `json:"respect"`          // 尊重度 (0-1)
	Closeness      float64   `json:"closeness"`        // 亲近感 (0-1)
	Importance     float64   `json:"importance"`       // 重要程度 (0-1)

	// === 互动统计 ===
	InteractionFrequency float64 `json:"interaction_frequency"` // 互动频率 (次/周)
	LastInteraction     time.Time `json:"last_interaction"`
	TotalInteractions   int       `json:"total_interactions"`    // 总互动次数
	PositiveInteractions int      `json:"positive_interactions"` // 积极互动次数
	NegativeInteractions int      `json:"negative_interactions"` // 消极互动次数

	// === 关系历史 ===
	MetDate        time.Time `json:"met_date"`         // 相识日期
	RelationshipStart time.Time `json:"relationship_start"` // 关系开始
	Milestones      []RelationshipMilestone `json:"milestones,omitempty"` // 关系里程碑

	// === 关系状态 ===
	Status         string    `json:"status"`           // active/distant/strained/ended/evolving
	Stage          string    `json:"stage"`            // acquaintance/casual/close/intimate
	Trend          string    `json:"trend"`            // improving/stable/declining/fluctuating

	// === 情感连接 ===
	EmotionalBond  float64   `json:"emotional_bond"`   // 情感纽带 (0-1)
	SupportGiven   float64   `json:"support_given"`    // 给予支持 (0-1)
	SupportReceived float64  `json:"support_received"` // 获得支持 (0-1)
	ConflictLevel  float64   `json:"conflict_level"`   // 冲突程度 (0-1)

	// === 沟通模式 ===
	CommunicationStyle string `json:"communication_style"` // direct/indirect/formal/casual/mixed
	PreferredChannels  []string `json:"preferred_channels,omitempty"` // 偏好沟通渠道
	ResponseExpectation float64 `json:"response_expectation"` // 回应期望 (0-1)

	// === 共同经历 ===
	SharedExperiences []string `json:"shared_experiences,omitempty"`
	SharedInterests   []string `json:"shared_interests,omitempty"`
	SharedValues      []string `json:"shared_values,omitempty"`

	// === 期望与义务 ===
	Expectations    []string `json:"expectations,omitempty"` // 对对方的期望
	Obligations     []string `json:"obligations,omitempty"`  // 对对方的义务
	Boundaries      []string `json:"boundaries,omitempty"`   // 边界

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RelationshipMilestone 关系里程碑
type RelationshipMilestone struct {
	MilestoneID   string    `json:"milestone_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
	Significance  float64   `json:"significance"` // 重要程度 (0-1)
}

// SocialNetwork 社交网络
type SocialNetwork struct {
	// 网络规模
	TotalContacts    int `json:"total_contacts"`
	CloseContacts    int `json:"close_contacts"`     // 亲密度>0.7
	SupportContacts  int `json:"support_contacts"`   // 可提供支持的人

	// 网络密度
	Density          float64 `json:"density"`          // 网络密度 (0-1)
	Diversity        float64 `json:"diversity"`        // 网络多样性 (0-1)

	// 社交圈层
	Circles          []SocialCircle `json:"circles,omitempty"`

	// 社交资本
	SocialCapital    float64 `json:"social_capital"`   // 社交资本 (0-1)
	BridgePositions  int     `json:"bridge_positions"` // 桥接位置数

	// 弱关系
	WeakTies         int `json:"weak_ties"`         // 弱关系数量
	StrongTies       int `json:"strong_ties"`       // 强关系数量

	// 社交需求满足
	BelongingNeed    float64 `json:"belonging_need"`   // 归属需求满足度
	IntimacyNeed     float64 `json:"intimacy_need"`    // 亲密需求满足度
	SupportNeed      float64 `json:"support_need"`     // 支持需求满足度
}

// SocialCircle 社交圈层
type SocialCircle struct {
	CircleID      string   `json:"circle_id"`
	CircleName    string   `json:"circle_name"`      // 核心圈/亲密圈/朋友圈/熟人圈/工作圈
	CircleLevel   int      `json:"circle_level"`     // 圈层级别 (1-5)
	Members       []string `json:"members"`          // 成员ID列表
	Description   string   `json:"description"`
	InteractionFreq float64 `json:"interaction_freq"` // 互动频率
}

// RelationshipTemplate 关系模板
type RelationshipTemplate struct {
	TemplateID       string   `json:"template_id"`
	TemplateName     string   `json:"template_name"`
	RelationshipType string   `json:"relationship_type"`

	// 默认属性
	DefaultIntimacy  float64  `json:"default_intimacy"`
	DefaultTrust     float64  `json:"default_trust"`
	DefaultRespect   float64  `json:"default_respect"`

	// 期望行为
	ExpectedBehaviors []string `json:"expected_behaviors,omitempty"`
	ExpectedFrequencies map[string]float64 `json:"expected_frequencies,omitempty"` // 期望互动频率

	// 发展路径
	DevelopmentStages []string `json:"development_stages,omitempty"`

	// 潜在挑战
	PotentialChallenges []string `json:"potential_challenges,omitempty"`
}

// RelationshipProfile 关系画像
type RelationshipProfile struct {
	IdentityID string `json:"identity_id"`

	// === 关系倾向 ===
	RelationshipOrientation RelationshipOrientation `json:"relationship_orientation"`

	// === 社交风格 ===
	SocialStyle SocialStyleProfile `json:"social_style"`

	// === 关系模式 ===
	RelationshipPatterns []RelationshipPattern `json:"relationship_patterns,omitempty"`

	// === 冲突模式 ===
	ConflictStyle ConflictStyleProfile `json:"conflict_style"`

	// === 依恋风格 ===
	AttachmentStyle AttachmentStyleProfile `json:"attachment_style"`

	// === 关系能力 ===
	RelationshipSkills RelationshipSkillsProfile `json:"relationship_skills"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RelationshipOrientation 关系倾向
type RelationshipOrientation struct {
	// 趋向性
	SocialOrientation string  `json:"social_orientation"` // extraverted/introverted/ambiverted
	RelationshipSeeking float64 `json:"relationship_seeking"` // 关系寻求倾向 (0-1)
	IndependencePreference float64 `json:"independence_preference"` // 独立偏好 (0-1)

	// 投入程度
	CommitmentReadiness float64 `json:"commitment_readiness"` // 承诺准备度 (0-1)
	IntimacyComfort    float64  `json:"intimacy_comfort"`    // 亲密舒适度 (0-1)
	VulnerabilityWillingness float64 `json:"vulnerability_willingness"` // 脆弱展示意愿 (0-1)

	// 期望
	IdealRelationshipCount int `json:"ideal_relationship_count"` // 理想关系数量
	QualityVsQuantity      float64 `json:"quality_vs_quantity"` // 质量>数量倾向 (0-1)
}

// SocialStyleProfile 社交风格画像
type SocialStyleProfile struct {
	// 沟通风格
	Directness        float64 `json:"directness"`        // 直接程度 (0-1)
	Expressiveness    float64 `json:"expressiveness"`    // 表达程度 (0-1)
	ListeningStyle    string  `json:"listening_style"`   // active/passive/empathetic/analytical

	// 互动风格
	InitiationStyle   string  `json:"initiation_style"`  // proactive/reactive
	GroupPreference   float64 `json:"group_preference"`  // 群体偏好 (0-1) vs 一对一
	SmallTalkComfort  float64 `json:"small_talk_comfort"` // 闲聊舒适度 (0-1)
	DeepTalkPreference float64 `json:"deep_talk_preference"` // 深谈偏好 (0-1)

	// 社交能量
	SocialEnergy      float64 `json:"social_energy"`     // 社交能量 (0-1)
	RechargeStyle     string  `json:"recharge_style"`    // alone/with_close_others/social
}

// ConflictStyleProfile 冲突风格画像
type ConflictStyleProfile struct {
	// 冲突处理方式
	PrimaryStyle      string  `json:"primary_style"`     // competing/collaborating/compromising/avoiding/accommodating
	SecondaryStyle    string  `json:"secondary_style"`

	// 冲突倾向
	ConfrontationComfort float64 `json:"confrontation_comfort"` // 对抗舒适度 (0-1)
	ConflictAvoidance    float64 `json:"conflict_avoidance"`    // 冲突回避 (0-1)
	ResolutionPreference string  `json:"resolution_preference"` // win-win/compromise/avoid/dominant

	// 情绪管理
	EmotionalRegulation float64 `json:"emotional_regulation"` // 情绪调节能力 (0-1)
	ForgivenessTendency float64 `json:"forgiveness_tendency"` // 宽恕倾向 (0-1)
	GrudgeHolding       float64 `json:"grudge_holding"`       // 记仇倾向 (0-1)
}

// AttachmentStyleProfile 依恋风格画像
type AttachmentStyleProfile struct {
	PrimaryStyle      string  `json:"primary_style"`      // secure/anxious/avoidant/disorganized
	AnxietyLevel      float64 `json:"anxiety_level"`      // 焦虑水平 (0-1)
	AvoidanceLevel    float64 `json:"avoidance_level"`    // 回避水平 (0-1)

	// 依恋行为
	SeparationAnxiety float64 `json:"separation_anxiety"` // 分离焦虑 (0-1)
	ProximitySeeking  float64 `json:"proximity_seeking"`  // 亲近寻求 (0-1)
	SafeHavenUse      float64 `json:"safe_haven_use"`     // 安全港使用 (0-1)
	SecureBaseUse     float64 `json:"secure_base_use"`    // 安全基地使用 (0-1)
}

// RelationshipSkillsProfile 关系能力画像
type RelationshipSkillsProfile struct {
	// 沟通能力
	ActiveListening   float64 `json:"active_listening"`   // 积极倾听 (0-1)
	ClearExpression   float64 `json:"clear_expression"`   // 清晰表达 (0-1)
	Empathy           float64 `json:"empathy"`            // 共情能力 (0-1)
	PerspectiveTaking float64 `json:"perspective_taking"` // 观点采择 (0-1)

	// 关系维护
	TrustBuilding     float64 `json:"trust_building"`     // 信任建立 (0-1)
	BoundarySetting   float64 `json:"boundary_setting"`   // 边界设定 (0-1)
	SupportProvision  float64 `json:"support_provision"`  // 支持提供 (0-1)
	ConflictResolution float64 `json:"conflict_resolution"` // 冲突解决 (0-1)

	// 整体评分
	OverallCompetence float64 `json:"overall_competence"` // 整体关系能力 (0-1)
}

// RelationshipPattern 关系模式
type RelationshipPattern struct {
	PatternID       string   `json:"pattern_id"`
	PatternName     string   `json:"pattern_name"`      // 如 "过度付出型"、"回避亲密型"
	Description     string   `json:"description"`
	TriggerSituations []string `json:"trigger_situations,omitempty"`
	TypicalBehaviors  []string `json:"typical_behaviors,omitempty"`
	Frequency        float64  `json:"frequency"`        // 出现频率 (0-1)
	Impact          string   `json:"impact"`           // positive/negative/neutral
}

// RelationshipInteraction 关系互动记录
type RelationshipInteraction struct {
	InteractionID    string    `json:"interaction_id"`
	RelationshipID   string    `json:"relationship_id"`
	Timestamp        time.Time `json:"timestamp"`

	// 互动类型
	InteractionType  string    `json:"interaction_type"`  // conversation/activity/support/conflict/celebration
	Context          string    `json:"context"`           // 互动情境

	// 互动质量
	Satisfaction     float64   `json:"satisfaction"`      // 满意度 (0-1)
	IntimacyChange   float64   `json:"intimacy_change"`   // 亲密度变化 (-1到1)
	TrustChange      float64   `json:"trust_change"`      // 信任度变化 (-1到1)

	// 情绪体验
	MyEmotion        string    `json:"my_emotion"`        // 我的情绪
	TheirEmotion     string    `json:"their_emotion,omitempty"` // 对方情绪
	EmotionalResonance float64 `json:"emotional_resonance"` // 情感共鸣 (0-1)

	// 持续时间和深度
	Duration         int       `json:"duration"`          // 持续时间（分钟）
	Depth            string    `json:"depth"`             // shallow/moderate/deep

	// 后续影响
	FollowUpNeeded   bool      `json:"follow_up_needed"`
	Notes            string    `json:"notes,omitempty"`
}

// NewRelationshipSystem 创建默认人际关系系统
func NewRelationshipSystem() *RelationshipSystem {
	now := time.Now()
	return &RelationshipSystem{
		Relationships:         []Relationship{},
		SocialNetwork:         NewSocialNetwork(),
		RelationshipTemplates: []RelationshipTemplate{},
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// NewSocialNetwork 创建默认社交网络
func NewSocialNetwork() *SocialNetwork {
	return &SocialNetwork{
		TotalContacts:   0,
		CloseContacts:   0,
		SupportContacts: 0,
		Density:         0.3,
		Diversity:       0.5,
		Circles:         []SocialCircle{},
		SocialCapital:   0.3,
		BridgePositions: 0,
		WeakTies:        0,
		StrongTies:      0,
		BelongingNeed:   0.5,
		IntimacyNeed:    0.5,
		SupportNeed:     0.5,
	}
}

// NewRelationshipProfile 创建关系画像
func NewRelationshipProfile(identityID string) *RelationshipProfile {
	now := time.Now()
	return &RelationshipProfile{
		IdentityID: identityID,
		RelationshipOrientation: RelationshipOrientation{
			SocialOrientation:      "ambiverted",
			RelationshipSeeking:     0.5,
			IndependencePreference:  0.5,
			CommitmentReadiness:     0.5,
			IntimacyComfort:         0.5,
			VulnerabilityWillingness: 0.5,
			IdealRelationshipCount:  5,
			QualityVsQuantity:       0.6,
		},
		SocialStyle: SocialStyleProfile{
			Directness:         0.5,
			Expressiveness:     0.5,
			ListeningStyle:     "active",
			InitiationStyle:    "proactive",
			GroupPreference:    0.5,
			SmallTalkComfort:   0.5,
			DeepTalkPreference: 0.6,
			SocialEnergy:       0.5,
			RechargeStyle:      "alone",
		},
		ConflictStyle: ConflictStyleProfile{
			PrimaryStyle:         "collaborating",
			ConfrontationComfort: 0.5,
			ConflictAvoidance:    0.4,
			ResolutionPreference: "win-win",
			EmotionalRegulation:  0.6,
			ForgivenessTendency:  0.6,
			GrudgeHolding:        0.3,
		},
		AttachmentStyle: AttachmentStyleProfile{
			PrimaryStyle:      "secure",
			AnxietyLevel:      0.3,
			AvoidanceLevel:    0.3,
			SeparationAnxiety: 0.3,
			ProximitySeeking:  0.5,
			SafeHavenUse:      0.5,
			SecureBaseUse:     0.5,
		},
		RelationshipSkills: RelationshipSkillsProfile{
			ActiveListening:     0.6,
			ClearExpression:     0.6,
			Empathy:             0.6,
			PerspectiveTaking:   0.6,
			TrustBuilding:       0.5,
			BoundarySetting:     0.5,
			SupportProvision:    0.5,
			ConflictResolution:  0.5,
			OverallCompetence:   0.55,
		},
		RelationshipPatterns: []RelationshipPattern{},
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// Normalize 归一化
func (s *RelationshipSystem) Normalize() {
	for i := range s.Relationships {
		s.Relationships[i].normalize()
	}
	s.UpdatedAt = time.Now()
}

func (r *Relationship) normalize() {
	r.Intimacy = clamp01(r.Intimacy)
	r.Trust = clamp01(r.Trust)
	r.Respect = clamp01(r.Respect)
	r.Closeness = clamp01(r.Closeness)
	r.Importance = clamp01(r.Importance)
	r.InteractionFrequency = clamp01(r.InteractionFrequency)
	r.EmotionalBond = clamp01(r.EmotionalBond)
	r.SupportGiven = clamp01(r.SupportGiven)
	r.SupportReceived = clamp01(r.SupportReceived)
	r.ConflictLevel = clamp01(r.ConflictLevel)
	r.ResponseExpectation = clamp01(r.ResponseExpectation)
}

// IsClose 是否亲密关系
func (r *Relationship) IsClose() bool {
	return r.Intimacy > 0.7
}

// IsTrustworthy 是否可信任
func (r *Relationship) IsTrustworthy() bool {
	return r.Trust > 0.7
}

// IsSupportive 是否支持性关系
func (r *Relationship) IsSupportive() bool {
	return r.SupportGiven > 0.6 && r.SupportReceived > 0.6
}

// HasConflict 是否有冲突
func (r *Relationship) HasConflict() bool {
	return r.ConflictLevel > 0.5
}

// GetRelationshipHealth 获取关系健康度
func (r *Relationship) GetRelationshipHealth() float64 {
	// 综合评分
	positiveScore := (r.Intimacy + r.Trust + r.Respect + r.Closeness + r.EmotionalBond) / 5
	negativeScore := r.ConflictLevel
	return clamp01(positiveScore*(1-negativeScore*0.5))
}

// GetRelationshipTypeName 获取关系类型名称
func (r *Relationship) GetRelationshipTypeName() string {
	names := map[string]string{
		"family":       "家人",
		"friend":       "朋友",
		"colleague":    "同事",
		"romantic":     "恋人",
		"acquaintance": "熟人",
		"mentor":       "导师",
		"mentee":       "学生",
		"other":        "其他",
	}
	if name, ok := names[r.RelationshipType]; ok {
		return name
	}
	return r.RelationshipType
}

// GetStageName 获取阶段名称
func (r *Relationship) GetStageName() string {
	names := map[string]string{
		"acquaintance": "相识",
		"casual":       "普通",
		"close":        "亲密",
		"intimate":     "知己",
	}
	if name, ok := names[r.Stage]; ok {
		return name
	}
	return r.Stage
}

// IsSecurelyAttached 是否安全依恋
func (a *AttachmentStyleProfile) IsSecurelyAttached() bool {
	return a.PrimaryStyle == "secure" && a.AnxietyLevel < 0.4 && a.AvoidanceLevel < 0.4
}

// IsAnxiouslyAttached 是否焦虑依恋
func (a *AttachmentStyleProfile) IsAnxiouslyAttached() bool {
	return a.AnxietyLevel > 0.6
}

// IsAvoidantlyAttached 是否回避依恋
func (a *AttachmentStyleProfile) IsAvoidantlyAttached() bool {
	return a.AvoidanceLevel > 0.6
}

// GetOverallSocialHealth 获取整体社交健康度
func (n *SocialNetwork) GetOverallSocialHealth() float64 {
	return (n.BelongingNeed + n.IntimacyNeed + n.SupportNeed + n.SocialCapital) / 4
}
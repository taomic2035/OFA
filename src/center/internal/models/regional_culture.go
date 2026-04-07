package models

import "time"

// RegionalCulture 地域文化模型 (v4.3.0)
// 地域文化对人格的影响
type RegionalCulture struct {
	// === 基本信息 ===
	Province      string `json:"province"`       // 省份
	City          string `json:"city"`           // 城市
	CityTier      string `json:"city_tier"`      // 城市等级 (tier1/new_tier1/tier2/tier3/tier4/tier5)
	Region        string `json:"region"`         // 大区域 (north/south/east/west/central/northeast/southwest/northwest)

	// === 文化特征 ===
	Dialect         string   `json:"dialect"`          // 方言
	DialectProficiency float64 `json:"dialect_proficiency"` // 方言熟练度 (0-1)
	RegionalTraits  []string `json:"regional_traits"`  // 地域性格特征
	Customs         []string `json:"customs"`          // 习俗

	// === 文化维度 ===
	// 集体主义 vs 个人主义 (Hofstede)
	Collectivism float64 `json:"collectivism"` // 集体主义倾向 (0-1, 高=集体主义)

	// 传统 vs 现代
	TraditionOriented  float64 `json:"tradition_oriented"`  // 传统导向 (0-1)
	InnovationOriented float64 `json:"innovation_oriented"` // 创新导向 (0-1)

	// 权力距离
	PowerDistance float64 `json:"power_distance"` // 权力距离接受度 (0-1)

	// 不确定性规避
	UncertaintyAvoidance float64 `json:"uncertainty_avoidance"` // 不确定性规避 (0-1)

	// 长期导向
	LongTermOrientation float64 `json:"long_term_orientation"` // 长期导向 (0-1)

	// 阳刚 vs 阴柔气质
	Masculinity float64 `json:"masculinity"` // 阳刚气质倾向 (0-1)

	// === 地域性格特征 ===
	// 沟通风格
	CommunicationStyle string `json:"communication_style"` // direct/indirect/mixed
	ExpressionLevel    float64 `json:"expression_level"`    // 表达开放度 (0-1)

	// 社交风格
	SocialStyle    string  `json:"social_style"`    // reserved/open/warm/pragmatic
	Hospitality    float64 `json:"hospitality"`    // 好客程度 (0-1)
	FaceConscious  float64 `json:"face_conscious"` // 面子意识 (0-1)

	// 生活节奏
	LifePace     string  `json:"life_pace"`     // slow/moderate/fast
	TimeUrgency  float64 `json:"time_urgency"`  // 时间紧迫感 (0-1)

	// 饮食文化
	CuisinePreference []string `json:"cuisine_preference"` // 饮食偏好
	SpiceTolerance    float64  `json:"spice_tolerance"`    // 辣度接受度 (0-1)

	// === 迁移经历 ===
	Native           bool     `json:"native"`            // 是否本地人
	MigrationHistory []Migration `json:"migration_history,omitempty"` // 迁移历史
	CulturalAdaptation float64 `json:"cultural_adaptation"` // 文化适应能力 (0-1)

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Migration 迁移记录
type Migration struct {
	MigrationID string    `json:"migration_id"`
	FromRegion  string    `json:"from_region"`  // 原地区
	ToRegion    string    `json:"to_region"`    // 目标地区
	FromCity    string    `json:"from_city"`
	ToCity      string    `json:"to_city"`
	Year        int       `json:"year"`
	Reason      string    `json:"reason"`       // work/education/family/lifestyle
	Adaptation  float64   `json:"adaptation"`   // 适应程度 (0-1)
	Timestamp   time.Time `json:"timestamp"`
}

// RegionalCultureProfile 地域文化画像
type RegionalCultureProfile struct {
	IdentityID string `json:"identity_id"`

	// === 文化认同 ===
	RegionalIdentity    float64 `json:"regional_identity"`    // 地域认同感 (0-1)
	CulturalRootedness  float64 `json:"cultural_rootedness"`  // 文化根植性 (0-1)
	CosmopolitanOutlook float64 `json:"cosmopolitan_outlook"` // 世界观开放度 (0-1)

	// === 文化影响权重 ===
	CulturalInfluence map[string]float64 `json:"cultural_influence,omitempty"` // 各文化维度影响权重

	// === 文化适应 ===
	AdaptationHistory []CulturalAdaptation `json:"adaptation_history,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CulturalAdaptation 文化适应记录
type CulturalAdaptation struct {
	AdaptationID string    `json:"adaptation_id"`
	Aspect       string    `json:"aspect"`       // 适应方面 (language/customs/social/values)
	Challenge    string    `json:"challenge"`    // 挑战描述
	Strategy     string    `json:"strategy"`     // 应对策略
	Success      float64   `json:"success"`      // 成功程度 (0-1)
	Timestamp    time.Time `json:"timestamp"`
}

// RegionalStereotype 地域刻板印象（用于理解，非定义）
type RegionalStereotype struct {
	Region       string   `json:"region"`
	CommonTraits []string `json:"common_traits"`
	Strengths    []string `json:"strengths"`
	Stereotypes  []string `json:"stereotypes"` // 常见刻板印象
}

// 中国大区域定义
var ChineseRegions = map[string][]string{
	"north":     {"北京", "天津", "河北", "山西", "内蒙古"},
	"northeast": {"辽宁", "吉林", "黑龙江"},
	"east":      {"上海", "江苏", "浙江", "安徽", "福建", "江西", "山东"},
	"south":     {"广东", "广西", "海南"},
	"central":   {"河南", "湖北", "湖南"},
	"southwest": {"重庆", "四川", "贵州", "云南", "西藏"},
	"northwest": {"陕西", "甘肃", "青海", "宁夏", "新疆"},
}

// 城市等级定义
var CityTierDefinitions = map[string]string{
	"tier1":     "一线城市 (北上广深)",
	"new_tier1": "新一线城市",
	"tier2":     "二线城市",
	"tier3":     "三线城市",
	"tier4":     "四线城市",
	"tier5":     "五线及以下",
}

// NewRegionalCulture 创建默认地域文化
func NewRegionalCulture() *RegionalCulture {
	now := time.Now()
	return &RegionalCulture{
		Province:            "",
		City:                "",
		CityTier:            "tier2",
		Region:              "east",
		Dialect:             "",
		DialectProficiency:  0.5,
		RegionalTraits:      []string{},
		Customs:             []string{},
		Collectivism:        0.5,
		TraditionOriented:   0.4,
		InnovationOriented:  0.6,
		PowerDistance:       0.5,
		UncertaintyAvoidance: 0.5,
		LongTermOrientation: 0.6,
		Masculinity:         0.5,
		CommunicationStyle:  "mixed",
		ExpressionLevel:     0.5,
		SocialStyle:         "pragmatic",
		Hospitality:         0.6,
		FaceConscious:       0.5,
		LifePace:            "moderate",
		TimeUrgency:         0.5,
		CuisinePreference:   []string{},
		SpiceTolerance:      0.5,
		Native:              true,
		CulturalAdaptation:  0.5,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// NewRegionalCultureProfile 创建默认地域文化画像
func NewRegionalCultureProfile(identityID string) *RegionalCultureProfile {
	return &RegionalCultureProfile{
		IdentityID:          identityID,
		RegionalIdentity:    0.5,
		CulturalRootedness:  0.5,
		CosmopolitanOutlook: 0.5,
		CulturalInfluence:   make(map[string]float64),
		AdaptationHistory:   []CulturalAdaptation{},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// GetCityTierName 获取城市等级名称
func (r *RegionalCulture) GetCityTierName() string {
	if name, ok := CityTierDefinitions[r.CityTier]; ok {
		return name
	}
	return r.CityTier
}

// GetRegionName 获取大区域名称
func (r *RegionalCulture) GetRegionName() string {
	names := map[string]string{
		"north":     "华北",
		"northeast": "东北",
		"east":      "华东",
		"south":     "华南",
		"central":   "华中",
		"southwest": "西南",
		"northwest": "西北",
	}
	if name, ok := names[r.Region]; ok {
		return name
	}
	return r.Region
}

// GetCommunicationStyleName 获取沟通风格名称
func (r *RegionalCulture) GetCommunicationStyleName() string {
	names := map[string]string{
		"direct":   "直接表达",
		"indirect": "委婉含蓄",
		"mixed":    "因地制宜",
	}
	if name, ok := names[r.CommunicationStyle]; ok {
		return name
	}
	return r.CommunicationStyle
}

// GetSocialStyleName 获取社交风格名称
func (r *RegionalCulture) GetSocialStyleName() string {
	names := map[string]string{
		"reserved":  "内敛含蓄",
		"open":      "开放热情",
		"warm":      "热情好客",
		"pragmatic": "务实理性",
	}
	if name, ok := names[r.SocialStyle]; ok {
		return name
	}
	return r.SocialStyle
}

// IsMetropolitan 是否大都市
func (r *RegionalCulture) IsMetropolitan() bool {
	return r.CityTier == "tier1" || r.CityTier == "new_tier1"
}

// IsNative 是否本地人
func (r *RegionalCulture) IsNative() bool {
	return r.Native
}

// HasMigrationHistory 是否有迁移经历
func (r *RegionalCulture) HasMigrationHistory() bool {
	return len(r.MigrationHistory) > 0
}

// CalculateInfluence 计算地域文化对决策的影响
func (r *RegionalCulture) CalculateInfluence() map[string]float64 {
	influence := make(map[string]float64)

	// 集体主义影响
	influence["group_decision_preference"] = r.Collectivism
	influence["social_conformity"] = r.Collectivism * 0.8
	influence["relationship_importance"] = r.Collectivism * 0.7

	// 传统导向影响
	influence["tradition_respect"] = r.TraditionOriented
	influence["risk_aversion"] = r.TraditionOriented * 0.6
	influence["stability_preference"] = r.TraditionOriented * 0.7

	// 创新导向影响
	influence["innovation_openness"] = r.InnovationOriented
	influence["change_acceptance"] = r.InnovationOriented * 0.8
	influence["new_experience_seeking"] = r.InnovationOriented * 0.7

	// 权力距离影响
	influence["authority_respect"] = r.PowerDistance
	influence["hierarchy_acceptance"] = r.PowerDistance * 0.8

	// 不确定性规避影响
	influence["uncertainty_comfort"] = 1 - r.UncertaintyAvoidance
	influence["risk_tolerance"] = 1 - r.UncertaintyAvoidance*0.7
	influence["rule_following"] = r.UncertaintyAvoidance * 0.6

	// 长期导向影响
	influence["future_orientation"] = r.LongTermOrientation
	influence["delayed_gratification"] = r.LongTermOrientation * 0.7
	influence["investment_mindset"] = r.LongTermOrientation * 0.6

	// 沟通风格影响
	if r.CommunicationStyle == "direct" {
		influence["directness"] = 0.8
		influence["conflict_confrontation"] = 0.6
	} else if r.CommunicationStyle == "indirect" {
		influence["harmony_maintenance"] = 0.8
		influence["face_preservation"] = r.FaceConscious
	}

	// 表达开放度
	influence["self_expression"] = r.ExpressionLevel
	influence["emotion_expression"] = r.ExpressionLevel * 0.8

	// 面子意识
	influence["reputation_concern"] = r.FaceConscious
	influence["social_approval_seeking"] = r.FaceConscious * 0.7

	// 好客程度
	influence["hospitality_tendency"] = r.Hospitality
	influence["generosity"] = r.Hospitality * 0.7

	// 时间紧迫感
	influence["pace_preference"] = r.TimeUrgency
	influence["efficiency_focus"] = r.TimeUrgency * 0.8

	// 大都市影响
	if r.IsMetropolitan() {
		influence["urban_mindset"] = 0.7
		influence["diversity_exposure"] = 0.6
		influence["competition_awareness"] = 0.6
	}

	// 文化适应能力
	influence["cultural_flexibility"] = r.CulturalAdaptation
	influence["cross_cultural_competence"] = r.CulturalAdaptation * 0.8

	return influence
}

// GetCultureDescription 获取文化描述
func (r *RegionalCulture) GetCultureDescription() string {
	desc := ""

	// 地区
	if r.Province != "" {
		desc += "来自" + r.Province
		if r.City != "" {
			desc += r.City
		}
		desc += "，"
	}

	// 大区域
	desc += r.GetRegionName() + "地区。"

	// 文化倾向
	if r.Collectivism > 0.6 {
		desc += "注重集体和人际关系，"
	} else if r.Collectivism < 0.4 {
		desc += "倾向于个人独立，"
	}

	// 传统/创新
	if r.TraditionOriented > r.InnovationOriented {
		desc += "重视传统价值。"
	} else if r.InnovationOriented > r.TraditionOriented {
		desc += "开放接受新事物。"
	} else {
		desc += "在传统与创新间保持平衡。"
	}

	return desc
}

// AddMigration 添加迁移记录
func (r *RegionalCulture) AddMigration(migration Migration) {
	r.MigrationHistory = append(r.MigrationHistory, migration)
	r.Native = false
	r.UpdatedAt = time.Now()
}

// Normalize 归一化
func (r *RegionalCulture) Normalize() {
	normalizeValue := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	r.DialectProficiency = normalizeValue(r.DialectProficiency)
	r.Collectivism = normalizeValue(r.Collectivism)
	r.TraditionOriented = normalizeValue(r.TraditionOriented)
	r.InnovationOriented = normalizeValue(r.InnovationOriented)
	r.PowerDistance = normalizeValue(r.PowerDistance)
	r.UncertaintyAvoidance = normalizeValue(r.UncertaintyAvoidance)
	r.LongTermOrientation = normalizeValue(r.LongTermOrientation)
	r.Masculinity = normalizeValue(r.Masculinity)
	r.ExpressionLevel = normalizeValue(r.ExpressionLevel)
	r.Hospitality = normalizeValue(r.Hospitality)
	r.FaceConscious = normalizeValue(r.FaceConscious)
	r.TimeUrgency = normalizeValue(r.TimeUrgency)
	r.SpiceTolerance = normalizeValue(r.SpiceTolerance)
	r.CulturalAdaptation = normalizeValue(r.CulturalAdaptation)

	r.UpdatedAt = time.Now()
}

// GetPresetRegionalCulture 获取预设地域文化（根据省份）
func GetPresetRegionalCulture(province string) *RegionalCulture {
	culture := NewRegionalCulture()
	culture.Province = province

	// 根据省份设置预设值
	switch province {
	case "北京", "天津":
		culture.Region = "north"
		culture.Collectivism = 0.4
		culture.InnovationOriented = 0.7
		culture.PowerDistance = 0.5
		culture.CommunicationStyle = "direct"
		culture.SocialStyle = "pragmatic"
		culture.FaceConscious = 0.4
		culture.TimeUrgency = 0.7
		culture.RegionalTraits = []string{"直爽", "幽默", "务实"}

	case "上海", "江苏", "浙江":
		culture.Region = "east"
		culture.Collectivism = 0.5
		culture.InnovationOriented = 0.7
		culture.PowerDistance = 0.4
		culture.CommunicationStyle = "mixed"
		culture.SocialStyle = "pragmatic"
		culture.FaceConscious = 0.5
		culture.TimeUrgency = 0.8
		culture.LongTermOrientation = 0.7
		culture.RegionalTraits = []string{"精明", "务实", "开放"}

	case "广东":
		culture.Region = "south"
		culture.Collectivism = 0.6
		culture.InnovationOriented = 0.7
		culture.PowerDistance = 0.4
		culture.CommunicationStyle = "direct"
		culture.SocialStyle = "open"
		culture.FaceConscious = 0.6
		culture.TimeUrgency = 0.7
		culture.Hospitality = 0.7
		culture.RegionalTraits = []string{"务实", "开放", "包容"}

	case "四川", "重庆":
		culture.Region = "southwest"
		culture.Collectivism = 0.6
		culture.TraditionOriented = 0.5
		culture.CommunicationStyle = "direct"
		culture.SocialStyle = "open"
		culture.Hospitality = 0.8
		culture.SpiceTolerance = 0.9
		culture.TimeUrgency = 0.4
		culture.RegionalTraits = []string{"热情", "乐观", "安逸"}

	case "山东":
		culture.Region = "east"
		culture.Collectivism = 0.7
		culture.TraditionOriented = 0.6
		culture.PowerDistance = 0.6
		culture.CommunicationStyle = "direct"
		culture.SocialStyle = "warm"
		culture.Hospitality = 0.9
		culture.FaceConscious = 0.7
		culture.RegionalTraits = []string{"豪爽", "重义", "传统"}

	case "辽宁", "吉林", "黑龙江":
		culture.Region = "northeast"
		culture.Collectivism = 0.6
		culture.TraditionOriented = 0.5
		culture.CommunicationStyle = "direct"
		culture.SocialStyle = "open"
		culture.Hospitality = 0.8
		culture.FaceConscious = 0.6
		culture.RegionalTraits = []string{"豪爽", "幽默", "热情"}

	default:
		// 默认值
		culture.RegionalTraits = []string{}
	}

	return culture
}
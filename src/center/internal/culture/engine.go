package culture

import (
	"context"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// RegionalCultureEngine 地域文化管理引擎 (v4.3.0)
// 管理地域文化对人格的影响
type RegionalCultureEngine struct {
	mu sync.RWMutex

	// 地域文化存储
	regionalCultures map[string]*models.RegionalCulture // identityID -> RegionalCulture
	cultureProfiles  map[string]*models.RegionalCultureProfile

	// 监听器
	listeners []RegionalCultureListener
}

// RegionalCultureListener 地域文化变化监听器
type RegionalCultureListener interface {
	OnRegionalCultureChanged(identityID string, culture *models.RegionalCulture)
	OnMigrationAdded(identityID string, migration models.Migration)
	OnCulturalAdaptationRecorded(identityID string, adaptation models.CulturalAdaptation)
}

// CulturalDecisionContext 地域文化决策上下文
type CulturalDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 地域文化
	RegionalCulture *models.RegionalCulture `json:"regional_culture"`

	// 影响因子
	InfluenceFactors map[string]float64 `json:"influence_factors"`

	// 文化决策倾向
	CulturalTendencies CulturalTendencies `json:"cultural_tendencies"`

	// 文化背景
	CulturalBackground CulturalBackground `json:"cultural_background"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// CulturalTendencies 文化决策倾向
type CulturalTendencies struct {
	// 集体相关
	GroupDecisionPreference float64 `json:"group_decision_preference"`
	SocialConformity        float64 `json:"social_conformity"`
	RelationshipImportance  float64 `json:"relationship_importance"`

	// 传统/创新
	TraditionRespect    float64 `json:"tradition_respect"`
	InnovationOpenness  float64 `json:"innovation_openness"`
	ChangeAcceptance    float64 `json:"change_acceptance"`

	// 沟通
	Directness         float64 `json:"directness"`
	HarmonyMaintenance float64 `json:"harmony_maintenance"`
	SelfExpression     float64 `json:"self_expression"`

	// 社交
	HospitalityTendency float64 `json:"hospitality_tendency"`
	FacePreservation    float64 `json:"face_preservation"`
	ReputationConcern   float64 `json:"reputation_concern"`

	// 时间
	PacePreference float64 `json:"pace_preference"`
	EfficiencyFocus float64 `json:"efficiency_focus"`

	// 风险
	RiskTolerance float64 `json:"risk_tolerance"`
	RiskAversion  float64 `json:"risk_aversion"`

	// 文化适应
	CulturalFlexibility        float64 `json:"cultural_flexibility"`
	CrossCulturalCompetence    float64 `json:"cross_cultural_competence"`
	DiversityExposure          float64 `json:"diversity_exposure"`
}

// CulturalBackground 文化背景
type CulturalBackground struct {
	Region       string   `json:"region"`
	Province     string   `json:"province"`
	City         string   `json:"city"`
	CityTier     string   `json:"city_tier"`
	IsMetropolitan bool    `json:"is_metropolitan"`
	IsNative     bool     `json:"is_native"`
	Dialect      string   `json:"dialect"`
	RegionalTraits []string `json:"regional_traits"`
	MigrationCount int     `json:"migration_count"`
}

// NewRegionalCultureEngine 创建地域文化管理引擎
func NewRegionalCultureEngine() *RegionalCultureEngine {
	return &RegionalCultureEngine{
		regionalCultures: make(map[string]*models.RegionalCulture),
		cultureProfiles:  make(map[string]*models.RegionalCultureProfile),
		listeners:        []RegionalCultureListener{},
	}
}

// === 地域文化管理 ===

// GetRegionalCulture 获取地域文化
func (e *RegionalCultureEngine) GetRegionalCulture(identityID string) *models.RegionalCulture {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.regionalCultures[identityID]
}

// GetOrCreateRegionalCulture 获取或创建地域文化
func (e *RegionalCultureEngine) GetOrCreateRegionalCulture(identityID string) *models.RegionalCulture {
	e.mu.Lock()
	defer e.mu.Unlock()

	culture, exists := e.regionalCultures[identityID]
	if !exists {
		culture = models.NewRegionalCulture()
		e.regionalCultures[identityID] = culture
		e.cultureProfiles[identityID] = models.NewRegionalCultureProfile(identityID)
	}
	return culture
}

// UpdateRegionalCulture 更新地域文化
func (e *RegionalCultureEngine) UpdateRegionalCulture(identityID string, culture *models.RegionalCulture) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	culture.Normalize()
	e.regionalCultures[identityID] = culture

	for _, listener := range e.listeners {
		listener.OnRegionalCultureChanged(identityID, culture)
	}

	return nil
}

// SetLocation 设置地理位置
func (e *RegionalCultureEngine) SetLocation(identityID string, province string, city string, cityTier string) error {
	culture := e.GetOrCreateRegionalCulture(identityID)

	// 如果省份变化，应用预设文化
	if culture.Province != province {
		preset := models.GetPresetRegionalCulture(province)
		preset.City = city
		preset.CityTier = cityTier
		preset.Native = !culture.HasMigrationHistory()
		return e.UpdateRegionalCulture(identityID, preset)
	}

	culture.Province = province
	culture.City = city
	culture.CityTier = cityTier
	return e.UpdateRegionalCulture(identityID, culture)
}

// SetDialect 设置方言
func (e *RegionalCultureEngine) SetDialect(identityID string, dialect string, proficiency float64) error {
	culture := e.GetOrCreateRegionalCulture(identityID)
	culture.Dialect = dialect
	culture.DialectProficiency = proficiency
	return e.UpdateRegionalCulture(identityID, culture)
}

// SetCulturalDimensions 设置文化维度
func (e *RegionalCultureEngine) SetCulturalDimensions(identityID string, dimensions map[string]float64) error {
	culture := e.GetOrCreateRegionalCulture(identityID)

	for key, value := range dimensions {
		switch key {
		case "collectivism":
			culture.Collectivism = value
		case "tradition_oriented":
			culture.TraditionOriented = value
		case "innovation_oriented":
			culture.InnovationOriented = value
		case "power_distance":
			culture.PowerDistance = value
		case "uncertainty_avoidance":
			culture.UncertaintyAvoidance = value
		case "long_term_orientation":
			culture.LongTermOrientation = value
		case "masculinity":
			culture.Masculinity = value
		case "expression_level":
			culture.ExpressionLevel = value
		case "hospitality":
			culture.Hospitality = value
		case "face_conscious":
			culture.FaceConscious = value
		case "time_urgency":
			culture.TimeUrgency = value
		case "cultural_adaptation":
			culture.CulturalAdaptation = value
		}
	}

	return e.UpdateRegionalCulture(identityID, culture)
}

// SetCommunicationStyle 设置沟通风格
func (e *RegionalCultureEngine) SetCommunicationStyle(identityID string, style string) error {
	culture := e.GetOrCreateRegionalCulture(identityID)
	culture.CommunicationStyle = style
	return e.UpdateRegionalCulture(identityID, culture)
}

// SetSocialStyle 设置社交风格
func (e *RegionalCultureEngine) SetSocialStyle(identityID string, style string) error {
	culture := e.GetOrCreateRegionalCulture(identityID)
	culture.SocialStyle = style
	return e.UpdateRegionalCulture(identityID, culture)
}

// AddRegionalTrait 添加地域性格特征
func (e *RegionalCultureEngine) AddRegionalTrait(identityID string, trait string) error {
	culture := e.GetOrCreateRegionalCulture(identityID)

	// 检查是否已存在
	for _, t := range culture.RegionalTraits {
		if t == trait {
			return nil
		}
	}

	culture.RegionalTraits = append(culture.RegionalTraits, trait)
	return e.UpdateRegionalCulture(identityID, culture)
}

// AddCustom 添加习俗
func (e *RegionalCultureEngine) AddCustom(identityID string, custom string) error {
	culture := e.GetOrCreateRegionalCulture(identityID)

	// 检查是否已存在
	for _, c := range culture.Customs {
		if c == custom {
			return nil
		}
	}

	culture.Customs = append(culture.Customs, custom)
	return e.UpdateRegionalCulture(identityID, culture)
}

// === 迁移管理 ===

// AddMigration 添加迁移记录
func (e *RegionalCultureEngine) AddMigration(identityID string, migration models.Migration) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	culture := e.GetOrCreateRegionalCulture(identityID)
	culture.AddMigration(migration)

	// 更新文化适应能力
	if len(culture.MigrationHistory) > 0 {
		// 多次迁移增加适应能力
		culture.CulturalAdaptation = min(1.0, culture.CulturalAdaptation+0.1)
	}

	for _, listener := range e.listeners {
		listener.OnMigrationAdded(identityID, migration)
	}

	return nil
}

// RecordCulturalAdaptation 记录文化适应
func (e *RegionalCultureEngine) RecordCulturalAdaptation(identityID string, adaptation models.CulturalAdaptation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.cultureProfiles[identityID]
	if profile == nil {
		profile = models.NewRegionalCultureProfile(identityID)
		e.cultureProfiles[identityID] = profile
	}

	profile.AdaptationHistory = append(profile.AdaptationHistory, adaptation)
	if len(profile.AdaptationHistory) > 20 {
		profile.AdaptationHistory = profile.AdaptationHistory[len(profile.AdaptationHistory)-20:]
	}

	// 更新适应能力
	culture := e.regionalCultures[identityID]
	if culture != nil {
		culture.CulturalAdaptation = min(1.0, culture.CulturalAdaptation+adaptation.Success*0.1)
	}

	for _, listener := range e.listeners {
		listener.OnCulturalAdaptationRecorded(identityID, adaptation)
	}

	return nil
}

// === 文化画像 ===

// GetCultureProfile 获取文化画像
func (e *RegionalCultureEngine) GetCultureProfile(identityID string) *models.RegionalCultureProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cultureProfiles[identityID]
}

// UpdateCultureProfile 更新文化画像
func (e *RegionalCultureEngine) UpdateCultureProfile(identityID string, profile *models.RegionalCultureProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.UpdatedAt = time.Now()
	e.cultureProfiles[identityID] = profile
	return nil
}

// === 决策上下文 ===

// GetDecisionContext 获取地域文化决策上下文
func (e *RegionalCultureEngine) GetDecisionContext(identityID string) *CulturalDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	culture := e.regionalCultures[identityID]
	if culture == nil {
		culture = models.NewRegionalCulture()
	}

	// 计算影响因子
	influenceFactors := culture.CalculateInfluence()

	// 计算决策倾向
	tendencies := e.calculateCulturalTendencies(culture, influenceFactors)

	// 构建文化背景
	background := CulturalBackground{
		Region:         culture.Region,
		Province:       culture.Province,
		City:           culture.City,
		CityTier:       culture.CityTier,
		IsMetropolitan: culture.IsMetropolitan(),
		IsNative:       culture.Native,
		Dialect:        culture.Dialect,
		RegionalTraits: culture.RegionalTraits,
		MigrationCount: len(culture.MigrationHistory),
	}

	return &CulturalDecisionContext{
		IdentityID:        identityID,
		RegionalCulture:   culture,
		InfluenceFactors:  influenceFactors,
		CulturalTendencies: tendencies,
		CulturalBackground: background,
		Timestamp:         time.Now(),
	}
}

// calculateCulturalTendencies 计算文化决策倾向
func (e *RegionalCultureEngine) calculateCulturalTendencies(culture *models.RegionalCulture, influence map[string]float64) CulturalTendencies {
	tendencies := CulturalTendencies{}

	// 从影响因子提取
	tendencies.GroupDecisionPreference = influence["group_decision_preference"]
	tendencies.SocialConformity = influence["social_conformity"]
	tendencies.RelationshipImportance = influence["relationship_importance"]
	tendencies.TraditionRespect = influence["tradition_respect"]
	tendencies.InnovationOpenness = influence["innovation_openness"]
	tendencies.ChangeAcceptance = influence["change_acceptance"]
	tendencies.Directness = influence["directness"]
	tendencies.HarmonyMaintenance = influence["harmony_maintenance"]
	tendencies.SelfExpression = influence["self_expression"]
	tendencies.HospitalityTendency = influence["hospitality_tendency"]
	tendencies.FacePreservation = influence["face_preservation"]
	tendencies.ReputationConcern = influence["reputation_concern"]
	tendencies.PacePreference = influence["pace_preference"]
	tendencies.EfficiencyFocus = influence["efficiency_focus"]
	tendencies.RiskTolerance = influence["risk_tolerance"]
	tendencies.RiskAversion = influence["risk_aversion"]
	tendencies.CulturalFlexibility = influence["cultural_flexibility"]
	tendencies.CrossCulturalCompetence = influence["cross_cultural_competence"]
	tendencies.DiversityExposure = influence["diversity_exposure"]

	return tendencies
}

// === 文化比较 ===

// CompareCultures 比较两个身份的文化差异
func (e *RegionalCultureEngine) CompareCultures(identityID1, identityID2 string) *CulturalComparison {
	culture1 := e.GetRegionalCulture(identityID1)
	culture2 := e.GetRegionalCulture(identityID2)

	if culture1 == nil {
		culture1 = models.NewRegionalCulture()
	}
	if culture2 == nil {
		culture2 = models.NewRegionalCulture()
	}

	comparison := &CulturalComparison{
		Identity1: identityID1,
		Identity2: identityID2,
		Differences: make(map[string]float64),
		Similarities: make(map[string]float64),
	}

	// 计算各维度差异
	dimensions := []struct {
		name  string
		value1, value2 float64
	}{
		{"collectivism", culture1.Collectivism, culture2.Collectivism},
		{"tradition_oriented", culture1.TraditionOriented, culture2.TraditionOriented},
		{"innovation_oriented", culture1.InnovationOriented, culture2.InnovationOriented},
		{"power_distance", culture1.PowerDistance, culture2.PowerDistance},
		{"uncertainty_avoidance", culture1.UncertaintyAvoidance, culture2.UncertaintyAvoidance},
		{"long_term_orientation", culture1.LongTermOrientation, culture2.LongTermOrientation},
		{"expression_level", culture1.ExpressionLevel, culture2.ExpressionLevel},
		{"hospitality", culture1.Hospitality, culture2.Hospitality},
		{"face_conscious", culture1.FaceConscious, culture2.FaceConscious},
		{"time_urgency", culture1.TimeUrgency, culture2.TimeUrgency},
	}

	for _, dim := range dimensions {
		diff := abs(dim.value1 - dim.value2)
		if diff > 0.3 {
			comparison.Differences[dim.name] = diff
		} else if diff < 0.1 {
			comparison.Similarities[dim.name] = 1 - diff
		}
	}

	// 计算整体兼容性
	comparison.Compatibility = e.calculateCompatibility(culture1, culture2)

	return comparison
}

// CulturalComparison 文化比较结果
type CulturalComparison struct {
	Identity1    string             `json:"identity_1"`
	Identity2    string             `json:"identity_2"`
	Differences  map[string]float64 `json:"differences"`
	Similarities map[string]float64 `json:"similarities"`
	Compatibility float64           `json:"compatibility"`
}

// calculateCompatibility 计算文化兼容性
func (e *RegionalCultureEngine) calculateCompatibility(c1, c2 *models.RegionalCulture) float64 {
	// 同地区高兼容性
	if c1.Region == c2.Region && c1.Region != "" {
		return 0.8
	}

	// 同省份更高
	if c1.Province == c2.Province && c1.Province != "" {
		return 0.9
	}

	// 计算各维度相似度
	similarity := 0.0
	dimensions := []float64{
		1 - abs(c1.Collectivism - c2.Collectivism),
		1 - abs(c1.TraditionOriented - c2.TraditionOriented),
		1 - abs(c1.InnovationOriented - c2.InnovationOriented),
		1 - abs(c1.PowerDistance - c2.PowerDistance),
		ifThen(c1.CommunicationStyle == c2.CommunicationStyle, 1, 0),
		ifThen(c1.SocialStyle == c2.SocialStyle, 1, 0),
	}

	for _, s := range dimensions {
		similarity += s
	}

	return similarity / float64(len(dimensions))
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *RegionalCultureEngine) AddListener(listener RegionalCultureListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *RegionalCultureEngine) RemoveListener(listener RegionalCultureListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.listeners {
		if l == listener {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			break
		}
	}
}

// === 辅助函数 ===

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func ifThen(condition bool, ifValue, elseValue float64) float64 {
	if condition {
		return ifValue
	}
	return elseValue
}
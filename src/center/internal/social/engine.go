package social

import (
	"context"
	"sync"
	"time"

	"github.com/taomic2035/OFA/src/center/internal/models"
)

// SocialIdentityEngine 社会身份管理引擎 (v4.2.0)
// 管理教育背景、职业画像、社会阶层、身份认同
type SocialIdentityEngine struct {
	mu sync.RWMutex

	// 社会身份存储
	identities map[string]*models.SocialIdentity // identityID -> SocialIdentity

	// 监听器
	listeners []SocialIdentityListener
}

// SocialIdentityListener 社会身份变化监听器
type SocialIdentityListener interface {
	OnEducationChanged(identityID string, education *models.EducationBackground)
	OnCareerChanged(identityID string, career *models.CareerProfile)
	OnSocialClassChanged(identityID string, socialClass *models.SocialClassProfile)
	OnIdentityChanged(identityID string, identity *models.IdentityProfile)
	OnSocialIdentityUpdated(identityID string, socialIdentity *models.SocialIdentity)
}

// SocialDecisionContext 社会身份决策上下文
type SocialDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 社会身份
	SocialIdentity *models.SocialIdentity `json:"social_identity"`

	// 影响因子
	InfluenceFactors map[string]float64 `json:"influence_factors"`

	// 决策倾向
	DecisionTendencies SocialDecisionTendencies `json:"decision_tendencies"`

	// 社会角色
	DominantRole string   `json:"dominant_role"`
	ActiveRoles  []string `json:"active_roles"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// SocialDecisionTendencies 社会决策倾向
type SocialDecisionTendencies struct {
	// 认知相关
	CognitiveComplexity float64 `json:"cognitive_complexity"` // 认知复杂度
	Adaptability        float64 `json:"adaptability"`         // 适应性

	// 职业相关
	WorkPriority      float64 `json:"work_priority"`       // 工作优先级
	CareerRiskTaking  float64 `json:"career_risk_taking"`  // 职业风险偏好
	JobInfluencedHappiness float64 `json:"job_influenced_happiness"` // 工作影响的幸福感

	// 经济相关
	FinancialRiskTolerance float64 `json:"financial_risk_tolerance"` // 财务风险容忍
	UpwardMobilityDrive    float64 `json:"upward_mobility_drive"`    // 向上流动动力

	// 社会相关
	SocialLeverage     float64 `json:"social_leverage"`     // 社交影响力
	DecisionConsistency float64 `json:"decision_consistency"` // 决策一致性
	IdentityOpenness    float64 `json:"identity_openness"`    // 身份开放度
}

// NewSocialIdentityEngine 创建社会身份管理引擎
func NewSocialIdentityEngine() *SocialIdentityEngine {
	return &SocialIdentityEngine{
		identities: make(map[string]*models.SocialIdentity),
		listeners:  []SocialIdentityListener{},
	}
}

// === 社会身份管理 ===

// GetSocialIdentity 获取社会身份
func (e *SocialIdentityEngine) GetSocialIdentity(identityID string) *models.SocialIdentity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.identities[identityID]
}

// GetOrCreateSocialIdentity 获取或创建社会身份
func (e *SocialIdentityEngine) GetOrCreateSocialIdentity(identityID string) *models.SocialIdentity {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity, exists := e.identities[identityID]
	if !exists {
		identity = models.NewSocialIdentity()
		e.identities[identityID] = identity
	}
	return identity
}

// UpdateSocialIdentity 更新社会身份
func (e *SocialIdentityEngine) UpdateSocialIdentity(identityID string, identity *models.SocialIdentity) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity.Normalize()
	e.identities[identityID] = identity

	for _, listener := range e.listeners {
		listener.OnSocialIdentityUpdated(identityID, identity)
	}

	return nil
}

// === 教育背景管理 ===

// GetEducation 获取教育背景
func (e *SocialIdentityEngine) GetEducation(identityID string) *models.EducationBackground {
	identity := e.GetSocialIdentity(identityID)
	if identity == nil {
		return nil
	}
	return identity.Education
}

// UpdateEducation 更新教育背景
func (e *SocialIdentityEngine) UpdateEducation(identityID string, education *models.EducationBackground) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity := e.GetOrCreateSocialIdentity(identityID)
	identity.Education = education
	identity.UpdatedAt = time.Now()

	for _, listener := range e.listeners {
		listener.OnEducationChanged(identityID, education)
	}

	return nil
}

// SetEducationLevel 设置学历层次
func (e *SocialIdentityEngine) SetEducationLevel(identityID string, level string) error {
	education := e.GetOrCreateSocialIdentity(identityID).Education
	if education == nil {
		education = &models.EducationBackground{
			EducationLevel:      level,
			AcademicPerformance: 0.5,
			LearningOrientation: "balanced",
		}
	} else {
		education.EducationLevel = level
	}
	return e.UpdateEducation(identityID, education)
}

// SetMajor 设置专业
func (e *SocialIdentityEngine) SetMajor(identityID string, major string, category string) error {
	education := e.GetOrCreateSocialIdentity(identityID).Education
	if education == nil {
		education = &models.EducationBackground{
			Major:               major,
			MajorCategory:       category,
			AcademicPerformance: 0.5,
			LearningOrientation: "balanced",
		}
	} else {
		education.Major = major
		education.MajorCategory = category
	}
	return e.UpdateEducation(identityID, education)
}

// === 职业画像管理 ===

// GetCareer 获取职业画像
func (e *SocialIdentityEngine) GetCareer(identityID string) *models.CareerProfile {
	identity := e.GetSocialIdentity(identityID)
	if identity == nil {
		return nil
	}
	return identity.Career
}

// UpdateCareer 更新职业画像
func (e *SocialIdentityEngine) UpdateCareer(identityID string, career *models.CareerProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity := e.GetOrCreateSocialIdentity(identityID)
	identity.Career = career
	identity.UpdatedAt = time.Now()

	for _, listener := range e.listeners {
		listener.OnCareerChanged(identityID, career)
	}

	return nil
}

// SetOccupation 设置职业
func (e *SocialIdentityEngine) SetOccupation(identityID string, occupation string, industry string) error {
	career := e.GetOrCreateSocialIdentity(identityID).Career
	if career == nil {
		career = &models.CareerProfile{
			Occupation:      occupation,
			Industry:        industry,
			CareerStage:     "developing",
			JobSatisfaction: 0.5,
			WorkLifeBalance: 0.5,
			CareerAmbition:  0.5,
		}
	} else {
		career.Occupation = occupation
		career.Industry = industry
	}
	return e.UpdateCareer(identityID, career)
}

// SetCareerStage 设置职业阶段
func (e *SocialIdentityEngine) SetCareerStage(identityID string, stage string, years int) error {
	career := e.GetOrCreateSocialIdentity(identityID).Career
	if career == nil {
		career = &models.CareerProfile{
			CareerStage:       stage,
			YearsOfExperience: years,
			JobSatisfaction:   0.5,
			WorkLifeBalance:   0.5,
			CareerAmbition:    0.5,
		}
	} else {
		career.CareerStage = stage
		career.YearsOfExperience = years
	}
	return e.UpdateCareer(identityID, career)
}

// SetJobSatisfaction 设置工作满意度
func (e *SocialIdentityEngine) SetJobSatisfaction(identityID string, satisfaction float64, workLifeBalance float64) error {
	career := e.GetOrCreateSocialIdentity(identityID).Career
	if career == nil {
		career = &models.CareerProfile{
			JobSatisfaction:  satisfaction,
			WorkLifeBalance:  workLifeBalance,
			CareerAmbition:   0.5,
		}
	} else {
		career.JobSatisfaction = satisfaction
		career.WorkLifeBalance = workLifeBalance
	}
	return e.UpdateCareer(identityID, career)
}

// === 社会阶层管理 ===

// GetSocialClass 获取社会阶层
func (e *SocialIdentityEngine) GetSocialClass(identityID string) *models.SocialClassProfile {
	identity := e.GetSocialIdentity(identityID)
	if identity == nil {
		return nil
	}
	return identity.SocialClass
}

// UpdateSocialClass 更新社会阶层
func (e *SocialIdentityEngine) UpdateSocialClass(identityID string, socialClass *models.SocialClassProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity := e.GetOrCreateSocialIdentity(identityID)
	identity.SocialClass = socialClass
	identity.UpdatedAt = time.Now()

	for _, listener := range e.listeners {
		listener.OnSocialClassChanged(identityID, socialClass)
	}

	return nil
}

// SetIncomeLevel 设置收入层次
func (e *SocialIdentityEngine) SetIncomeLevel(identityID string, level string, percentile float64) error {
	socialClass := e.GetOrCreateSocialIdentity(identityID).SocialClass
	if socialClass == nil {
		socialClass = &models.SocialClassProfile{
			IncomeLevel:      level,
			IncomePercentile: percentile,
			FinancialHealth:  0.5,
			EconomicCapital:  0.5,
			CulturalCapital:  0.5,
			SocialCapital:    0.5,
		}
	} else {
		socialClass.IncomeLevel = level
		socialClass.IncomePercentile = percentile
	}
	return e.UpdateSocialClass(identityID, socialClass)
}

// SetCapitals 设置三种资本
func (e *SocialIdentityEngine) SetCapitals(identityID string, economic, cultural, social float64) error {
	socialClass := e.GetOrCreateSocialIdentity(identityID).SocialClass
	if socialClass == nil {
		socialClass = &models.SocialClassProfile{
			EconomicCapital: economic,
			CulturalCapital: cultural,
			SocialCapital:   social,
			FinancialHealth: (economic + cultural + social) / 3,
		}
	} else {
		socialClass.EconomicCapital = economic
		socialClass.CulturalCapital = cultural
		socialClass.SocialCapital = social
	}
	return e.UpdateSocialClass(identityID, socialClass)
}

// === 身份认同管理 ===

// GetIdentityProfile 获取身份认同
func (e *SocialIdentityEngine) GetIdentityProfile(identityID string) *models.IdentityProfile {
	identity := e.GetSocialIdentity(identityID)
	if identity == nil {
		return nil
	}
	return identity.Identity
}

// UpdateIdentityProfile 更新身份认同
func (e *SocialIdentityEngine) UpdateIdentityProfile(identityID string, profile *models.IdentityProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	identity := e.GetOrCreateSocialIdentity(identityID)
	identity.Identity = profile
	identity.UpdatedAt = time.Now()

	for _, listener := range e.listeners {
		listener.OnIdentityChanged(identityID, profile)
	}

	return nil
}

// AddSocialRole 添加社会角色
func (e *SocialIdentityEngine) AddSocialRole(identityID string, role models.SocialRole) error {
	profile := e.GetOrCreateSocialIdentity(identityID).Identity
	if profile == nil {
		profile = &models.IdentityProfile{
			SocialRoles:       []models.SocialRole{role},
			SelfConceptLabels: []string{},
			IdentityFluidity:  0.5,
		}
	} else {
		// 检查是否已存在
		for i, r := range profile.SocialRoles {
			if r.RoleID == role.RoleID {
				profile.SocialRoles[i] = role
				return e.UpdateIdentityProfile(identityID, profile)
			}
		}
		profile.SocialRoles = append(profile.SocialRoles, role)
	}
	return e.UpdateIdentityProfile(identityID, profile)
}

// RemoveSocialRole 移除社会角色
func (e *SocialIdentityEngine) RemoveSocialRole(identityID string, roleID string) error {
	profile := e.GetOrCreateSocialIdentity(identityID).Identity
	if profile == nil {
		return nil
	}

	for i, r := range profile.SocialRoles {
		if r.RoleID == roleID {
			profile.SocialRoles = append(profile.SocialRoles[:i], profile.SocialRoles[i+1:]...)
			break
		}
	}

	return e.UpdateIdentityProfile(identityID, profile)
}

// SetRolePriority 设置角色优先级
func (e *SocialIdentityEngine) SetRolePriority(identityID string, priority []string) error {
	profile := e.GetOrCreateSocialIdentity(identityID).Identity
	if profile == nil {
		profile = &models.IdentityProfile{
			RolePriority:     priority,
			SocialRoles:      []models.SocialRole{},
			IdentityFluidity: 0.5,
		}
	} else {
		profile.RolePriority = priority
	}
	return e.UpdateIdentityProfile(identityID, profile)
}

// AddIdentityConflict 添加身份冲突
func (e *SocialIdentityEngine) AddIdentityConflict(identityID string, conflict models.IdentityConflict) error {
	profile := e.GetOrCreateSocialIdentity(identityID).Identity
	if profile == nil {
		profile = &models.IdentityProfile{
			IdentityConflicts: []models.IdentityConflict{conflict},
			SocialRoles:       []models.SocialRole{},
			IdentityFluidity:  0.5,
		}
	} else {
		profile.IdentityConflicts = append(profile.IdentityConflicts, conflict)
	}
	return e.UpdateIdentityProfile(identityID, profile)
}

// === 决策上下文 ===

// GetDecisionContext 获取社会身份决策上下文
func (e *SocialIdentityEngine) GetDecisionContext(identityID string) *SocialDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	identity := e.identities[identityID]
	if identity == nil {
		identity = models.NewSocialIdentity()
	}

	// 计算影响因子
	influenceFactors := identity.CalculateInfluence()

	// 计算决策倾向
	tendencies := e.calculateDecisionTendencies(identity)

	// 获取主导角色和活跃角色
	dominantRole := ""
	activeRoles := []string{}
	if identity.Identity != nil {
		if role := identity.Identity.GetDominantRole(); role != nil {
			dominantRole = role.RoleName
		}
		for _, role := range identity.Identity.SocialRoles {
			if role.Importance > 0.5 {
				activeRoles = append(activeRoles, role.RoleName)
			}
		}
	}

	return &SocialDecisionContext{
		IdentityID:        identityID,
		SocialIdentity:    identity,
		InfluenceFactors:  influenceFactors,
		DecisionTendencies: tendencies,
		DominantRole:      dominantRole,
		ActiveRoles:       activeRoles,
		Timestamp:         time.Now(),
	}
}

// calculateDecisionTendencies 计算决策倾向
func (e *SocialIdentityEngine) calculateDecisionTendencies(identity *models.SocialIdentity) SocialDecisionTendencies {
	tendencies := SocialDecisionTendencies{}
	influence := identity.CalculateInfluence()

	// 从影响因子提取决策倾向
	tendencies.CognitiveComplexity = influence["cognitive_complexity"]
	tendencies.Adaptability = influence["adaptability"]
	tendencies.WorkPriority = influence["work_priority"]
	tendencies.CareerRiskTaking = influence["career_risk_taking"]
	tendencies.JobInfluencedHappiness = influence["job_influenced_happiness"]
	tendencies.FinancialRiskTolerance = influence["financial_risk_tolerance"]
	tendencies.UpwardMobilityDrive = influence["upward_mobility_drive"]
	tendencies.SocialLeverage = influence["social_leverage"]
	tendencies.DecisionConsistency = influence["decision_consistency"]
	tendencies.IdentityOpenness = influence["identity_openness"]

	return tendencies
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *SocialIdentityEngine) AddListener(listener SocialIdentityListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *SocialIdentityEngine) RemoveListener(listener SocialIdentityListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.listeners {
		if l == listener {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			break
		}
	}
}
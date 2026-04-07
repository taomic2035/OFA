package models

import "time"

// SocialIdentity 社会身份模型 (v4.2.0)
// 完整的社会阶层和职业身份画像
type SocialIdentity struct {
	// === 教育背景 ===
	Education *EducationBackground `json:"education"`

	// === 职业画像 ===
	Career *CareerProfile `json:"career"`

	// === 社会阶层 ===
	SocialClass *SocialClassProfile `json:"social_class"`

	// === 身份认同 ===
	Identity *IdentityProfile `json:"identity"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EducationBackground 教育背景
type EducationBackground struct {
	// 最高学历
	EducationLevel string `json:"education_level"` // high_school/associate/bachelor/master/doctorate

	// 专业领域
	Major          string `json:"major"`           // 专业名称
	MajorCategory  string `json:"major_category"`  // stem/humanities/social_science/arts/business/other

	// 学校信息
	SchoolName     string `json:"school_name"`     // 学校名称
	SchoolTier     string `json:"school_tier"`     // top/first_tier/second_tier/regular

	// 学习经历
	GraduationYear int      `json:"graduation_year"`
	Honors         []string `json:"honors,omitempty"`    // 荣誉奖项
	Activities     []string `json:"activities,omitempty"` // 社团活动

	// 继续教育
	ContinuingEducation bool     `json:"continuing_education"` // 是否持续学习
	Certifications      []string `json:"certifications,omitempty"` // 职业证书

	// 教育影响
	AcademicPerformance float64 `json:"academic_performance"` // 学业表现 (0-1)
	LearningOrientation string  `json:"learning_orientation"` // theoretical/practical/balanced
}

// CareerProfile 职业画像
type CareerProfile struct {
	// 当前职业
	Occupation       string `json:"occupation"`        // 职业名称
	Industry         string `json:"industry"`          // 行业
	IndustryCategory string `json:"industry_category"` // tech/finance/healthcare/education/manufacturing/service/government/other

	// 职业阶段
	CareerStage      string  `json:"career_stage"`       // entry/developing/established/senior/executive
	YearsOfExperience int     `json:"years_of_experience"` // 工作年限
	JobLevel         string  `json:"job_level"`          // junior/mid/senior/manager/director/executive

	// 职业状态
	EmploymentStatus string `json:"employment_status"` // employed/self_employed/freelance/student/retired

	// 职业满意度
	JobSatisfaction  float64 `json:"job_satisfaction"`  // 工作满意度 (0-1)
	WorkLifeBalance  float64 `json:"work_life_balance"` // 工作生活平衡 (0-1)
	CareerAmbition   float64 `json:"career_ambition"`   // 事业野心 (0-1)

	// 职业目标
	CareerGoal       string   `json:"career_goal"`        // 职业目标
	Skills           []string `json:"skills,omitempty"`   // 技能
	Strengths        []string `json:"strengths,omitempty"` // 优势
	DevelopmentAreas []string `json:"development_areas,omitempty"` // 待发展领域

	// 工作模式
	WorkMode string `json:"work_mode"` // remote/hybrid/onsite
	WorkHours int   `json:"work_hours"` // 每周工作时长

	// 职业历史
	CareerTransitions int `json:"career_transitions"` // 职业转型次数
}

// SocialClassProfile 社会阶层画像
type SocialClassProfile struct {
	// 收入层次
	IncomeLevel      string  `json:"income_level"`       // low/lower_middle/middle/upper_middle/upper
	IncomePercentile float64 `json:"income_percentile"`  // 收入百分位 (0-100)

	// 财富状况
	NetWorthLevel   string  `json:"net_worth_level"`   // negative/minimal/moderate/substantial/high
	FinancialHealth float64 `json:"financial_health"` // 财务健康度 (0-1)

	// 社会地位
	SocialStatus    string  `json:"social_status"`    // marginal/working/middle/professional/elite
	StatusPerception float64 `json:"status_perception"` // 自我感知的社会地位 (0-1)

	// 社会流动性
	MobilityExperience string `json:"mobility_experience"` // upward/stable/downward
	MobilityAspiration string `json:"mobility_aspiration"` // 向上流动意愿

	// 资本类型 (布迪厄)
	EconomicCapital float64 `json:"economic_capital"` // 经济资本 (0-1)
	CulturalCapital float64 `json:"cultural_capital"` // 文化资本 (0-1)
	SocialCapital   float64 `json:"social_capital"`   // 社会资本 (0-1)

	// 生活质量
	HousingType     string  `json:"housing_type"`     // renting/owned_mortgage/owned/other
	NeighborhoodTier string `json:"neighborhood_tier"` // 居住区域等级
	Lifestyle       string  `json:"lifestyle"`        // frugal/moderate/comfortable/luxurious
}

// IdentityProfile 身份认同
type IdentityProfile struct {
	// 自我概念
	SelfConceptLabels []string `json:"self_concept_labels"` // 自我概念标签

	// 社会角色
	SocialRoles []SocialRole `json:"social_roles"`

	// 身份重要性排序
	RolePriority []string `json:"role_priority"`

	// 身份冲突
	IdentityConflicts []IdentityConflict `json:"identity_conflicts,omitempty"`

	// 群体归属
	GroupMemberships []GroupMembership `json:"group_memberships,omitempty"`

	// 身份流动性
	IdentityFluidity float64 `json:"identity_fluidity"` // 身份认同的灵活度 (0-1)
}

// SocialRole 社会角色
type SocialRole struct {
	RoleID       string   `json:"role_id"`
	RoleName     string   `json:"role_name"`     // parent/employee/citizen/friend/etc
	RoleCategory string   `json:"role_category"` // family/work/community/personal
	Importance   float64  `json:"importance"`    // 重要性 (0-1)
	Satisfaction float64  `json:"satisfaction"`  // 角色满意度 (0-1)
	TimeSpent    float64  `json:"time_spent"`    // 时间投入比例 (0-1)
	Expectations []string `json:"expectations,omitempty"` // 角色期望
	Challenges   []string `json:"challenges,omitempty"`  // 角色挑战
}

// IdentityConflict 身份冲突
type IdentityConflict struct {
	ConflictID   string  `json:"conflict_id"`
	Role1        string  `json:"role_1"`
	Role2        string  `json:"role_2"`
	Description  string  `json:"description"`
	Intensity    float64 `json:"intensity"`    // 冲突强度 (0-1)
	Resolution   string  `json:"resolution"`   // compromise/prioritize/integrate/avoid
}

// GroupMembership 群体归属
type GroupMembership struct {
	GroupID     string   `json:"group_id"`
	GroupName   string   `json:"group_name"`
	GroupType   string   `json:"group_type"`   // professional/social/political/religious/hobby
	JoinDate    time.Time `json:"join_date"`
	Involvement float64  `json:"involvement"` // 参与程度 (0-1)
	Importance  float64  `json:"importance"`  // 群体对身份的重要性 (0-1)
}

// NewSocialIdentity 创建默认社会身份
func NewSocialIdentity() *SocialIdentity {
	now := time.Now()
	return &SocialIdentity{
		Education: &EducationBackground{
			EducationLevel:      "bachelor",
			MajorCategory:       "other",
			SchoolTier:          "regular",
			ContinuingEducation: true,
			AcademicPerformance: 0.6,
			LearningOrientation: "balanced",
		},
		Career: &CareerProfile{
			CareerStage:       "developing",
			YearsOfExperience: 3,
			JobLevel:          "mid",
			EmploymentStatus:  "employed",
			JobSatisfaction:   0.6,
			WorkLifeBalance:   0.5,
			CareerAmbition:    0.6,
			WorkMode:          "hybrid",
			WorkHours:         40,
		},
		SocialClass: &SocialClassProfile{
			IncomeLevel:        "middle",
			IncomePercentile:   50,
			FinancialHealth:    0.6,
			SocialStatus:       "middle",
			StatusPerception:   0.5,
			MobilityExperience: "stable",
			MobilityAspiration: "upward",
			EconomicCapital:    0.5,
			CulturalCapital:    0.5,
			SocialCapital:      0.5,
			HousingType:        "renting",
			Lifestyle:          "moderate",
		},
		Identity: &IdentityProfile{
			SelfConceptLabels: []string{"学习者", "工作者"},
			SocialRoles: []SocialRole{
				{RoleID: "role_1", RoleName: "employee", RoleCategory: "work", Importance: 0.7, Satisfaction: 0.6, TimeSpent: 0.4},
				{RoleID: "role_2", RoleName: "friend", RoleCategory: "personal", Importance: 0.6, Satisfaction: 0.7, TimeSpent: 0.15},
			},
			RolePriority:     []string{"employee", "friend"},
			IdentityFluidity: 0.5,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetEducationLevelName 获取学历名称
func (e *EducationBackground) GetEducationLevelName() string {
	names := map[string]string{
		"high_school": "高中",
		"associate":   "大专",
		"bachelor":    "本科",
		"master":      "硕士",
		"doctorate":   "博士",
	}
	if name, ok := names[e.EducationLevel]; ok {
		return name
	}
	return e.EducationLevel
}

// GetCareerStageName 获取职业阶段名称
func (c *CareerProfile) GetCareerStageName() string {
	names := map[string]string{
		"entry":      "入门期",
		"developing": "发展期",
		"established": "成熟期",
		"senior":     "资深期",
		"executive":  "高管期",
	}
	if name, ok := names[c.CareerStage]; ok {
		return name
	}
	return c.CareerStage
}

// GetSocialStatusName 获取社会地位名称
func (s *SocialClassProfile) GetSocialStatusName() string {
	names := map[string]string{
		"marginal":    "边缘阶层",
		"working":     "工薪阶层",
		"middle":      "中产阶层",
		"professional": "专业阶层",
		"elite":       "精英阶层",
	}
	if name, ok := names[s.SocialStatus]; ok {
		return name
	}
	return s.SocialStatus
}

// CalculateOverallCapital 计算综合资本
func (s *SocialClassProfile) CalculateOverallCapital() float64 {
	return (s.EconomicCapital + s.CulturalCapital + s.SocialCapital) / 3
}

// GetDominantRole 获取主导角色
func (i *IdentityProfile) GetDominantRole() *SocialRole {
	if len(i.SocialRoles) == 0 {
		return nil
	}

	dominant := &i.SocialRoles[0]
	for j := 1; j < len(i.SocialRoles); j++ {
		if i.SocialRoles[j].Importance > dominant.Importance {
			dominant = &i.SocialRoles[j]
		}
	}
	return dominant
}

// CalculateRoleBalance 计算角色平衡度
func (i *IdentityProfile) CalculateRoleBalance() float64 {
	if len(i.SocialRoles) == 0 {
		return 1.0
	}

	// 计算时间分配与重要性的匹配度
	var totalMismatch float64
	for _, role := range i.SocialRoles {
		mismatch := abs(role.TimeSpent - role.Importance)
		totalMismatch += mismatch
	}

	// 平衡度 = 1 - 平均不匹配度
	avgMismatch := totalMismatch / float64(len(i.SocialRoles))
	return max(0, 1-avgMismatch)
}

// CalculateInfluence 计算社会身份对决策的影响
func (s *SocialIdentity) CalculateInfluence() map[string]float64 {
	influence := make(map[string]float64)

	// 教育影响
	if s.Education != nil {
		// 学历影响认知复杂度
		eduLevel := map[string]float64{
			"high_school": 0.3,
			"associate":   0.4,
			"bachelor":    0.5,
			"master":      0.6,
			"doctorate":   0.7,
		}
		influence["cognitive_complexity"] = eduLevel[s.Education.EducationLevel]

		// 持续学习影响适应性
		if s.Education.ContinuingEducation {
			influence["adaptability"] = 0.7
		} else {
			influence["adaptability"] = 0.4
		}
	}

	// 职业影响
	if s.Career != nil {
		// 工作满意度影响整体幸福感
		influence["job_influenced_happiness"] = s.Career.JobSatisfaction

		// 事业野心影响风险偏好
		influence["career_risk_taking"] = s.Career.CareerAmbition

		// 工作生活平衡影响决策倾向
		influence["work_priority"] = 1 - s.Career.WorkLifeBalance
	}

	// 社会阶层影响
	if s.SocialClass != nil {
		// 收入层次影响消费倾向
		incomeRisk := map[string]float64{
			"low":          0.2,
			"lower_middle": 0.3,
			"middle":       0.4,
			"upper_middle": 0.6,
			"upper":        0.7,
		}
		influence["financial_risk_tolerance"] = incomeRisk[s.SocialClass.IncomeLevel]

		// 社会资本影响社交决策
		influence["social_leverage"] = s.SocialClass.SocialCapital

		// 流动意愿影响进取心
		if s.SocialClass.MobilityAspiration == "upward" {
			influence["upward_mobility_drive"] = 0.7
		} else {
			influence["upward_mobility_drive"] = 0.3
		}
	}

	// 身份认同影响
	if s.Identity != nil {
		// 角色平衡影响决策一致性
		influence["decision_consistency"] = s.Identity.CalculateRoleBalance()

		// 身份流动性影响开放度
		influence["identity_openness"] = s.Identity.IdentityFluidity
	}

	return influence
}

// GetSocialIdentityDescription 获取社会身份描述
func (s *SocialIdentity) GetSocialIdentityDescription() string {
	desc := ""

	// 教育背景
	if s.Education != nil {
		desc += "学历：" + s.Education.GetEducationLevelName()
		if s.Education.Major != "" {
			desc += "（" + s.Education.Major + "）"
		}
		desc += "。"
	}

	// 职业
	if s.Career != nil {
		if s.Career.Occupation != "" {
			desc += "职业：" + s.Career.Occupation + "，"
		}
		desc += "处于" + s.Career.GetCareerStageName() + "。"
	}

	// 社会阶层
	if s.SocialClass != nil {
		desc += "社会阶层：" + s.SocialClass.GetSocialStatusName() + "。"
	}

	// 主导角色
	if s.Identity != nil {
		if role := s.Identity.GetDominantRole(); role != nil {
			desc += "主要身份认同：" + role.RoleName + "。"
		}
	}

	return desc
}

// Normalize 归一化
func (s *SocialIdentity) Normalize() {
	if s.Education != nil {
		s.Education.AcademicPerformance = clamp01(s.Education.AcademicPerformance)
	}
	if s.Career != nil {
		s.Career.JobSatisfaction = clamp01(s.Career.JobSatisfaction)
		s.Career.WorkLifeBalance = clamp01(s.Career.WorkLifeBalance)
		s.Career.CareerAmbition = clamp01(s.Career.CareerAmbition)
	}
	if s.SocialClass != nil {
		s.SocialClass.IncomePercentile = clamp(s.SocialClass.IncomePercentile, 0, 100)
		s.SocialClass.FinancialHealth = clamp01(s.SocialClass.FinancialHealth)
		s.SocialClass.StatusPerception = clamp01(s.SocialClass.StatusPerception)
		s.SocialClass.EconomicCapital = clamp01(s.SocialClass.EconomicCapital)
		s.SocialClass.CulturalCapital = clamp01(s.SocialClass.CulturalCapital)
		s.SocialClass.SocialCapital = clamp01(s.SocialClass.SocialCapital)
	}
	if s.Identity != nil {
		s.Identity.IdentityFluidity = clamp01(s.Identity.IdentityFluidity)
		for i := range s.Identity.SocialRoles {
			s.Identity.SocialRoles[i].Importance = clamp01(s.Identity.SocialRoles[i].Importance)
			s.Identity.SocialRoles[i].Satisfaction = clamp01(s.Identity.SocialRoles[i].Satisfaction)
			s.Identity.SocialRoles[i].TimeSpent = clamp01(s.Identity.SocialRoles[i].TimeSpent)
		}
	}
	s.UpdatedAt = time.Now()
}

// 辅助函数 (复用其他文件中的，这里声明避免重复定义)
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func clamp(x, minVal, maxVal float64) float64 {
	if x < minVal {
		return minVal
	}
	if x > maxVal {
		return maxVal
	}
	return x
}
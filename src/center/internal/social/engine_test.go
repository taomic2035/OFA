package social

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestSocialIdentityEngineCreation 测试社会身份引擎创建
func TestSocialIdentityEngineCreation(t *testing.T) {
	engine := NewSocialIdentityEngine()

	if engine == nil {
		t.Fatal("SocialIdentityEngine should not be nil")
	}
	if engine.identities == nil {
		t.Error("identities map should be initialized")
	}
}

// TestGetOrCreateSocialIdentity 测试获取或创建社会身份
func TestGetOrCreateSocialIdentity(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 创建新社会身份
	identity := engine.GetOrCreateSocialIdentity("identity_001")
	if identity == nil {
		t.Fatal("SocialIdentity should not be nil")
	}

	// 获取已存在的身份
	identity2 := engine.GetOrCreateSocialIdentity("identity_001")
	if identity != identity2 {
		t.Error("Should return same identity instance")
	}
}

// TestUpdateSocialIdentity 测试更新社会身份
func TestUpdateSocialIdentity(t *testing.T) {
	engine := NewSocialIdentityEngine()

	identity := models.NewSocialIdentity()
	identity.Education = &models.EducationBackground{
		EducationLevel:      "本科",
		Major:               "计算机科学",
		MajorCategory:       "理工",
		AcademicPerformance: 0.8,
	}
	identity.Career = &models.CareerProfile{
		Occupation:      "软件工程师",
		Industry:        "互联网",
		CareerStage:     "developing",
		JobSatisfaction: 0.7,
	}

	err := engine.UpdateSocialIdentity("identity_001", identity)
	if err != nil {
		t.Fatalf("UpdateSocialIdentity failed: %v", err)
	}

	// 验证更新
	retrieved := engine.GetSocialIdentity("identity_001")
	if retrieved == nil {
		t.Fatal("SocialIdentity should exist after update")
	}
	if retrieved.Education.EducationLevel != "本科" {
		t.Errorf("EducationLevel should be '本科', got %s", retrieved.Education.EducationLevel)
	}
}

// TestEducationManagement 测试教育背景管理
func TestEducationManagement(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 设置学历层次
	err := engine.SetEducationLevel("identity_001", "研究生")
	if err != nil {
		t.Fatalf("SetEducationLevel failed: %v", err)
	}

	education := engine.GetEducation("identity_001")
	if education == nil {
		t.Fatal("Education should exist")
	}
	if education.EducationLevel != "研究生" {
		t.Errorf("EducationLevel should be '研究生', got %s", education.EducationLevel)
	}

	// 设置专业
	err = engine.SetMajor("identity_001", "人工智能", "理工")
	if err != nil {
		t.Fatalf("SetMajor failed: %v", err)
	}

	education = engine.GetEducation("identity_001")
	if education.Major != "人工智能" {
		t.Errorf("Major should be '人工智能', got %s", education.Major)
	}
	if education.MajorCategory != "理工" {
		t.Errorf("MajorCategory should be '理工', got %s", education.MajorCategory)
	}
}

// TestCareerManagement 测试职业画像管理
func TestCareerManagement(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 设置职业
	err := engine.SetOccupation("identity_001", "产品经理", "科技")
	if err != nil {
		t.Fatalf("SetOccupation failed: %v", err)
	}

	career := engine.GetCareer("identity_001")
	if career == nil {
		t.Fatal("Career should exist")
	}
	if career.Occupation != "产品经理" {
		t.Errorf("Occupation should be '产品经理', got %s", career.Occupation)
	}

	// 设置职业阶段
	err = engine.SetCareerStage("identity_001", "established", 5)
	if err != nil {
		t.Fatalf("SetCareerStage failed: %v", err)
	}

	career = engine.GetCareer("identity_001")
	if career.CareerStage != "established" {
		t.Errorf("CareerStage should be 'established', got %s", career.CareerStage)
	}
	if career.YearsOfExperience != 5 {
		t.Errorf("YearsOfExperience should be 5, got %d", career.YearsOfExperience)
	}

	// 设置工作满意度
	err = engine.SetJobSatisfaction("identity_001", 0.8, 0.7)
	if err != nil {
		t.Fatalf("SetJobSatisfaction failed: %v", err)
	}

	career = engine.GetCareer("identity_001")
	if career.JobSatisfaction != 0.8 {
		t.Errorf("JobSatisfaction should be 0.8, got %f", career.JobSatisfaction)
	}
	if career.WorkLifeBalance != 0.7 {
		t.Errorf("WorkLifeBalance should be 0.7, got %f", career.WorkLifeBalance)
	}
}

// TestSocialClassManagement 测试社会阶层管理
func TestSocialClassManagement(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 设置收入层次
	err := engine.SetIncomeLevel("identity_001", "中高收入", 0.75)
	if err != nil {
		t.Fatalf("SetIncomeLevel failed: %v", err)
	}

	socialClass := engine.GetSocialClass("identity_001")
	if socialClass == nil {
		t.Fatal("SocialClass should exist")
	}
	if socialClass.IncomeLevel != "中高收入" {
		t.Errorf("IncomeLevel should be '中高收入', got %s", socialClass.IncomeLevel)
	}
	if socialClass.IncomePercentile != 0.75 {
		t.Errorf("IncomePercentile should be 0.75, got %f", socialClass.IncomePercentile)
	}

	// 设置三种资本
	err = engine.SetCapitals("identity_001", 0.7, 0.6, 0.8)
	if err != nil {
		t.Fatalf("SetCapitals failed: %v", err)
	}

	socialClass = engine.GetSocialClass("identity_001")
	if socialClass.EconomicCapital != 0.7 {
		t.Errorf("EconomicCapital should be 0.7, got %f", socialClass.EconomicCapital)
	}
	if socialClass.CulturalCapital != 0.6 {
		t.Errorf("CulturalCapital should be 0.6, got %f", socialClass.CulturalCapital)
	}
	if socialClass.SocialCapital != 0.8 {
		t.Errorf("SocialCapital should be 0.8, got %f", socialClass.SocialCapital)
	}
}

// TestSocialRoleManagement 测试社会角色管理
func TestSocialIdentityEngine_Role(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 添加社会角色
	role := models.SocialRole{
		RoleID:     "role_001",
		RoleName:   "父亲",
		RoleType:   "family",
		Importance: 0.9,
		Active:     true,
	}

	err := engine.AddSocialRole("identity_001", role)
	if err != nil {
		t.Fatalf("AddSocialRole failed: %v", err)
	}

	profile := engine.GetIdentityProfile("identity_001")
	if profile == nil {
		t.Fatal("IdentityProfile should exist")
	}
	if len(profile.SocialRoles) != 1 {
		t.Errorf("Should have 1 role, got %d", len(profile.SocialRoles))
	}

	// 添加另一个角色
	role2 := models.SocialRole{
		RoleID:     "role_002",
		RoleName:   "项目经理",
		RoleType:   "work",
		Importance: 0.7,
		Active:     true,
	}
	engine.AddSocialRole("identity_001", role2)

	// 更新已有角色
	role.Importance = 0.95
	engine.AddSocialRole("identity_001", role)

	profile = engine.GetIdentityProfile("identity_001")
	if len(profile.SocialRoles) != 2 {
		t.Errorf("Should have 2 roles after update, got %d", len(profile.SocialRoles))
	}

	// 移除角色
	err = engine.RemoveSocialRole("identity_001", "role_002")
	if err != nil {
		t.Fatalf("RemoveSocialRole failed: %v", err)
	}

	profile = engine.GetIdentityProfile("identity_001")
	if len(profile.SocialRoles) != 1 {
		t.Errorf("Should have 1 role after removal, got %d", len(profile.SocialRoles))
	}
}

// TestRolePriority 测试角色优先级设置
func TestRolePriority(t *testing.T) {
	engine := NewSocialIdentityEngine()

	priority := []string{"father", "manager", "friend"}
	err := engine.SetRolePriority("identity_001", priority)
	if err != nil {
		t.Fatalf("SetRolePriority failed: %v", err)
	}

	profile := engine.GetIdentityProfile("identity_001")
	if profile == nil {
		t.Fatal("IdentityProfile should exist")
	}
	if len(profile.RolePriority) != 3 {
		t.Errorf("RolePriority should have 3 items, got %d", len(profile.RolePriority))
	}
	if profile.RolePriority[0] != "father" {
		t.Errorf("First priority should be 'father', got %s", profile.RolePriority[0])
	}
}

// TestAddIdentityConflict 测试添加身份冲突
func TestAddIdentityConflict(t *testing.T) {
	engine := NewSocialIdentityEngine()

	conflict := models.IdentityConflict{
		ConflictID:      "conflict_001",
		Role1:           "父亲",
		Role2:           "项目经理",
		ConflictType:    "time",
		ConflictLevel:   0.7,
		ResolutionStatus: "unresolved",
	}

	err := engine.AddIdentityConflict("identity_001", conflict)
	if err != nil {
		t.Fatalf("AddIdentityConflict failed: %v", err)
	}

	profile := engine.GetIdentityProfile("identity_001")
	if profile == nil {
		t.Fatal("IdentityProfile should exist")
	}
	if len(profile.IdentityConflicts) != 1 {
		t.Errorf("Should have 1 conflict, got %d", len(profile.IdentityConflicts))
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 设置完整社会身份
	identity := engine.GetOrCreateSocialIdentity("identity_001")
	identity.Education = &models.EducationBackground{
		EducationLevel:      "研究生",
		AcademicPerformance: 0.8,
	}
	identity.Career = &models.CareerProfile{
		Occupation:      "软件工程师",
		Industry:        "互联网",
		CareerStage:     "developing",
		JobSatisfaction: 0.7,
		CareerAmbition:  0.8,
	}
	identity.SocialClass = &models.SocialClassProfile{
		EconomicCapital: 0.7,
		CulturalCapital: 0.6,
		SocialCapital:   0.8,
	}

	// 添加角色
	engine.AddSocialRole("identity_001", models.SocialRole{
		RoleID:     "role_001",
		RoleName:   "技术主管",
		RoleType:   "work",
		Importance: 0.8,
		Active:     true,
	})

	// 获取上下文
	context := engine.GetDecisionContext("identity_001")

	if context == nil {
		t.Fatal("Context should not be nil")
	}

	// 验证上下文字段
	if context.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", context.IdentityID)
	}
	if context.InfluenceFactors == nil {
		t.Error("InfluenceFactors should be calculated")
	}

	// 验证决策倾向
	if context.DecisionTendencies.WorkPriority == 0 {
		t.Error("WorkPriority should be calculated")
	}
	if context.DecisionTendencies.FinancialRiskTolerance == 0 {
		t.Error("FinancialRiskTolerance should be calculated")
	}

	// 验证角色信息
	if len(context.ActiveRoles) == 0 {
		t.Error("Should have at least one active role")
	}
}

// TestListener 测试监听器
func TestSocialIdentityEngine_Listener(t *testing.T) {
	engine := NewSocialIdentityEngine()

	// 创建测试监听器
	listener := &testSocialListener{
		educationChangedCount:     0,
		careerChangedCount:        0,
		socialClassChangedCount:   0,
		identityChangedCount:      0,
		socialIdentityUpdatedCount: 0,
	}

	engine.AddListener(listener)

	// 更新教育
	education := &models.EducationBackground{
		EducationLevel:      "本科",
		AcademicPerformance: 0.7,
	}
	engine.UpdateEducation("identity_001", education)

	if listener.educationChangedCount != 1 {
		t.Errorf("educationChangedCount should be 1, got %d", listener.educationChangedCount)
	}

	// 更新职业
	career := &models.CareerProfile{
		Occupation:      "工程师",
		JobSatisfaction: 0.6,
	}
	engine.UpdateCareer("identity_001", career)

	if listener.careerChangedCount != 1 {
		t.Errorf("careerChangedCount should be 1, got %d", listener.careerChangedCount)
	}

	// 更新社会阶层
	socialClass := &models.SocialClassProfile{
		EconomicCapital: 0.6,
	}
	engine.UpdateSocialClass("identity_001", socialClass)

	if listener.socialClassChangedCount != 1 {
		t.Errorf("socialClassChangedCount should be 1, got %d", listener.socialClassChangedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次更新
	engine.UpdateEducation("identity_001", education)

	if listener.educationChangedCount != 1 {
		t.Errorf("educationChangedCount should still be 1 after removal, got %d", listener.educationChangedCount)
	}
}

// testSocialListener 测试监听器
type testSocialListener struct {
	educationChangedCount     int
	careerChangedCount        int
	socialClassChangedCount   int
	identityChangedCount      int
	socialIdentityUpdatedCount int
}

func (l *testSocialListener) OnEducationChanged(identityID string, education *models.EducationBackground) {
	l.educationChangedCount++
}

func (l *testSocialListener) OnCareerChanged(identityID string, career *models.CareerProfile) {
	l.careerChangedCount++
}

func (l *testSocialListener) OnSocialClassChanged(identityID string, socialClass *models.SocialClassProfile) {
	l.socialClassChangedCount++
}

func (l *testSocialListener) OnIdentityChanged(identityID string, identity *models.IdentityProfile) {
	l.identityChangedCount++
}

func (l *testSocialListener) OnSocialIdentityUpdated(identityID string, socialIdentity *models.SocialIdentity) {
	l.socialIdentityUpdatedCount++
}

// TestSocialIdentityNormalize 测试社会身份归一化
func TestSocialIdentityNormalize(t *testing.T) {
	identity := models.NewSocialIdentity()
	identity.Career = &models.CareerProfile{
		JobSatisfaction: 1.5, // 超出范围
		WorkLifeBalance: -0.1, // 低于范围
	}

	identity.Normalize()

	if identity.Career.JobSatisfaction != 1.0 {
		t.Errorf("JobSatisfaction should be normalized to 1.0, got %f", identity.Career.JobSatisfaction)
	}
	if identity.Career.WorkLifeBalance != 0 {
		t.Errorf("WorkLifeBalance should be normalized to 0, got %f", identity.Career.WorkLifeBalance)
	}
}
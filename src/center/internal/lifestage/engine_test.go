package lifestage

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestLifeStageEngineCreation 测试人生阶段引擎创建
func TestLifeStageEngineCreation(t *testing.T) {
	engine := NewLifeStageEngine()

	if engine == nil {
		t.Fatal("LifeStageEngine should not be nil")
	}
	if engine.lifeStageSystems == nil {
		t.Error("lifeStageSystems map should be initialized")
	}
	if engine.lifeStageProfiles == nil {
		t.Error("lifeStageProfiles map should be initialized")
	}
}

// TestGetOrCreateLifeStageSystem 测试获取或创建人生阶段系统
func TestGetOrCreateLifeStageSystem(t *testing.T) {
	engine := NewLifeStageEngine()

	// 创建新人生阶段系统
	system := engine.GetOrCreateLifeStageSystem("identity_001")
	if system == nil {
		t.Fatal("LifeStageSystem should not be nil")
	}

	// 获取已存在的系统
	system2 := engine.GetOrCreateLifeStageSystem("identity_001")
	if system != system2 {
		t.Error("Should return same system instance")
	}
}

// TestSetCurrentStage 测试设置当前阶段
func TestSetCurrentStage(t *testing.T) {
	engine := NewLifeStageEngine()

	err := engine.SetCurrentStage("identity_001", "youth", 22)
	if err != nil {
		t.Fatalf("SetCurrentStage failed: %v", err)
	}

	system := engine.GetLifeStageSystem("identity_001")
	if system == nil {
		t.Fatal("LifeStageSystem should exist")
	}
	if system.CurrentStage == nil {
		t.Fatal("CurrentStage should be set")
	}
	if system.CurrentStage.StageName != "youth" {
		t.Errorf("StageName should be 'youth', got %s", system.CurrentStage.StageName)
	}
	if system.CurrentStage.StageAge != 22 {
		t.Errorf("StageAge should be 22, got %d", system.CurrentStage.StageAge)
	}

	// 验证阶段定义
	if system.CurrentStage.StageLabel == "" {
		t.Error("StageLabel should be set from definition")
	}
	if len(system.CurrentStage.Challenges) == 0 {
		t.Error("Challenges should be set from definition")
	}
}

// TestStageTransition 测试阶段转换
func TestStageStageTransition(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置初始阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 创建监听器
	listener := &testLifeStageListener{
		transitionCount: 0,
	}
	engine.AddListener(listener)

	// 转换到新阶段
	err := engine.SetCurrentStage("identity_001", "early_adult", 28)
	if err != nil {
		t.Fatalf("SetCurrentStage (transition) failed: %v", err)
	}

	// 验证转换通知
	if listener.transitionCount != 1 {
		t.Errorf("transitionCount should be 1, got %d", listener.transitionCount)
	}

	// 验证历史记录
	system := engine.GetLifeStageSystem("identity_001")
	if len(system.StageHistory) != 1 {
		t.Errorf("StageHistory should have 1 record, got %d", len(system.StageHistory))
	}

	// 验证当前阶段
	if system.CurrentStage.StageName != "early_adult" {
		t.Errorf("CurrentStage should be 'early_adult', got %s", system.CurrentStage.StageName)
	}
}

// TestUpdateDevelopmentMetrics 测试更新发展指标
func TestUpdateDevelopmentMetrics(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 更新发展指标
	metrics := map[string]float64{
		"physical_health":    0.8,
		"mental_health":      0.7,
		"cognitive_growth":   0.6,
		"career_progress":    0.5,
		"self_awareness":     0.7,
	}

	err := engine.UpdateDevelopmentMetrics("identity_001", metrics)
	if err != nil {
		t.Fatalf("UpdateDevelopmentMetrics failed: %v", err)
	}

	system := engine.GetLifeStageSystem("identity_001")
	if system.CurrentStage.DevelopmentMetrics.PhysicalHealth != 0.8 {
		t.Errorf("PhysicalHealth should be 0.8, got %f", system.CurrentStage.DevelopmentMetrics.PhysicalHealth)
	}
	if system.CurrentStage.DevelopmentMetrics.CareerProgress != 0.5 {
		t.Errorf("CareerProgress should be 0.5, got %f", system.CurrentStage.DevelopmentMetrics.CareerProgress)
	}
}

// TestAddStageGoal 测试添加阶段目标
func TestAddStageGoal(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 添加目标
	err := engine.AddStageGoal("identity_001", "找到理想工作")
	if err != nil {
		t.Fatalf("AddStageGoal failed: %v", err)
	}

	system := engine.GetLifeStageSystem("identity_001")
	if len(system.CurrentStage.Goals) != 1 {
		t.Errorf("Should have 1 goal, got %d", len(system.CurrentStage.Goals))
	}

	// 添加重复目标
	err = engine.AddStageGoal("identity_001", "找到理想工作")
	if err != nil {
		t.Fatalf("AddStageGoal (duplicate) failed: %v", err)
	}

	system = engine.GetLifeStageSystem("identity_001")
	if len(system.CurrentStage.Goals) != 1 {
		t.Errorf("Should still have 1 goal (no duplicates), got %d", len(system.CurrentStage.Goals))
	}
}

// TestUpdateStageCompleteness 测试更新阶段完成度
func TestUpdateStageCompleteness(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 更新完成度
	err := engine.UpdateStageCompleteness("identity_001", 0.6)
	if err != nil {
		t.Fatalf("UpdateStageCompleteness failed: %v", err)
	}

	system := engine.GetLifeStageSystem("identity_001")
	if system.CurrentStage.Completeness != 0.6 {
		t.Errorf("Completeness should be 0.6, got %f", system.CurrentStage.Completeness)
	}

	// 测试边界值
	err = engine.UpdateStageCompleteness("identity_001", 1.5)
	if err != nil {
		t.Fatalf("UpdateStageCompleteness (overflow) failed: %v", err)
	}

	system = engine.GetLifeStageSystem("identity_001")
	if system.CurrentStage.Completeness != 1.0 {
		t.Errorf("Completeness should be clamped to 1.0, got %f", system.CurrentStage.Completeness)
	}
}

// TestAddLifeEvent 测试添加人生事件
func TestAddLifeEvent(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 创建监听器
	listener := &testLifeStageListener{
		eventAddedCount: 0,
	}
	engine.AddListener(listener)

	// 添加人生事件
	event := models.LifeEvent{
		EventType:        "career",
		EventTitle:       "入职第一份工作",
		EventDescription: "进入互联网公司",
		EventDate:        time.Now(),
		EventAge:         22,
		ImpactLevel:      0.7,
		ImpactValence:    0.8, // 正向
		ImpactAreas:      map[string]float64{"career": 0.8, "financial": 0.6},
	}

	err := engine.AddLifeEvent("identity_001", event)
	if err != nil {
		t.Fatalf("AddLifeEvent failed: %v", err)
	}

	// 验证事件添加
	system := engine.GetLifeStageSystem("identity_001")
	if len(system.LifeEvents) != 1 {
		t.Errorf("Should have 1 event, got %d", len(system.LifeEvents))
	}

	// 验证监听器通知
	if listener.eventAddedCount != 1 {
		t.Errorf("eventAddedCount should be 1, got %d", listener.eventAddedCount)
	}

	// 验证发展指标受影响
	if system.CurrentStage.DevelopmentMetrics.CareerProgress < 0.5 {
		t.Errorf("Career event should affect CareerProgress, got %f", system.CurrentStage.DevelopmentMetrics.CareerProgress)
	}
}

// TestGetLifeEvents 测试获取人生事件
func TestGetLifeEvents(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 添加多个事件
	events := []models.LifeEvent{
		{EventType: "career", EventTitle: "入职"},
		{EventType: "education", EventTitle: "毕业"},
		{EventType: "career", EventTitle: "升职"},
	}

	for _, event := range events {
		engine.AddLifeEvent("identity_001", event)
	}

	// 获取所有事件
	allEvents := engine.GetLifeEvents("identity_001", "")
	if len(allEvents) != 3 {
		t.Errorf("Should have 3 events, got %d", len(allEvents))
	}

	// 获取特定类型事件
	careerEvents := engine.GetLifeEvents("identity_001", "career")
	if len(careerEvents) != 2 {
		t.Errorf("Should have 2 career events, got %d", len(careerEvents))
	}
}

// TestGetSignificantEvents 测试获取重大事件
func TestGetSignificantEvents(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	// 添加事件
	events := []models.LifeEvent{
		{EventType: "career", EventTitle: "入职", ImpactLevel: 0.5},
		{EventType: "education", EventTitle: "毕业", ImpactLevel: 0.8},
		{EventType: "relationship", EventTitle: "结婚", ImpactLevel: 0.9},
	}

	for _, event := range events {
		engine.AddLifeEvent("identity_001", event)
	}

	// 获取重大事件 (threshold 0.7)
	significant := engine.GetSignificantEvents("identity_001", 0.7)
	if len(significant) != 2 {
		t.Errorf("Should have 2 significant events, got %d", len(significant))
	}
}

// TestAddLifeLesson 测试添加人生感悟
func TestAddLifeLesson(t *testing.T) {
	engine := NewLifeStageEngine()

	// 创建监听器
	listener := &testLifeStageListener{
		lessonLearnedCount: 0,
	}
	engine.AddListener(listener)

	// 添加人生感悟
	lesson := models.LifeLesson{
		Title:       "坚持的重要性",
		Description: "通过持续努力达成目标",
		Category:    "personal_growth",
		Insights:    []string{"坚持比天赋更重要"},
		Application: []string{"设定小目标", "保持耐心"},
	}

	err := engine.AddLifeLesson("identity_001", lesson)
	if err != nil {
		t.Fatalf("AddLifeLesson failed: %v", err)
	}

	// 验证感悟添加
	system := engine.GetLifeStageSystem("identity_001")
	if len(system.LifeLessons) != 1 {
		t.Errorf("Should have 1 lesson, got %d", len(system.LifeLessons))
	}

	// 验证监听器通知
	if listener.lessonLearnedCount != 1 {
		t.Errorf("lessonLearnedCount should be 1, got %d", listener.lessonLearnedCount)
	}
}

// TestApplyLifeLesson 测试应用人生感悟
func TestApplyLifeLesson(t *testing.T) {
	engine := NewLifeStageEngine()

	// 添加感悟
	lesson := models.LifeLesson{
		Title:    "坚持的重要性",
		Category: "personal_growth",
	}
	engine.AddLifeLesson("identity_001", lesson)

	// 获取感悟ID
	system := engine.GetLifeStageSystem("identity_001")
	lessonID := system.LifeLessons[0].LessonID

	// 应用感悟
	err := engine.ApplyLifeLesson("identity_001", lessonID)
	if err != nil {
		t.Fatalf("ApplyLifeLesson failed: %v", err)
	}

	// 验证应用计数
	system = engine.GetLifeStageSystem("identity_001")
	if system.LifeLessons[0].ApplyCount != 1 {
		t.Errorf("ApplyCount should be 1, got %d", system.LifeLessons[0].ApplyCount)
	}
}

// TestGetLifeLessons 测试获取人生感悟
func TestGetLifeLessons(t *testing.T) {
	engine := NewLifeStageEngine()

	// 添加多个感悟
	lessons := []models.LifeLesson{
		{Title: "坚持", Category: "personal_growth"},
		{Title: "沟通技巧", Category: "relationship"},
		{Title: "时间管理", Category: "personal_growth"},
	}

	for _, lesson := range lessons {
		engine.AddLifeLesson("identity_001", lesson)
	}

	// 获取所有感悟
	allLessons := engine.GetLifeLessons("identity_001", "")
	if len(allLessons) != 3 {
		t.Errorf("Should have 3 lessons, got %d", len(allLessons))
	}

	// 获取特定类别感悟
	growthLessons := engine.GetLifeLessons("identity_001", "personal_growth")
	if len(growthLessons) != 2 {
		t.Errorf("Should have 2 personal_growth lessons, got %d", len(growthLessons))
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewLifeStageEngine()

	// 设置阶段
	engine.SetCurrentStage("identity_001", "early_adult", 28)

	// 更新发展指标
	engine.UpdateDevelopmentMetrics("identity_001", map[string]float64{
		"career_progress":   0.7,
		"relationship_quality": 0.6,
		"life_satisfaction": 0.7,
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
	if context.CurrentStage == nil {
		t.Fatal("CurrentStage should be set")
	}
	if context.CurrentStage.StageName != "early_adult" {
		t.Errorf("CurrentStage should be 'early_adult', got %s", context.CurrentStage.StageName)
	}

	// 验证阶段影响
	if context.StageInfluence.GrowthFocus == "" {
		t.Error("GrowthFocus should be calculated")
	}
	if context.StageInfluence.TimePerspective == "" {
		t.Error("TimePerspective should be calculated")
	}
	if context.StageInfluence.RiskTaking == 0 {
		t.Error("RiskTaking should be calculated")
	}

	// 验证发展状态
	if context.DevelopmentStatus.OverallProgress == 0 {
		t.Error("OverallProgress should be calculated")
	}

	// 验证轨迹摘要
	if context.TrajectorySummary.WisdomLevel == 0 {
		t.Error("WisdomLevel should be calculated")
	}
}

// TestInferStageFromAge 测试从年龄推断阶段
func TestInferStageFromAge(t *testing.T) {
	engine := NewLifeStageEngine()

	tests := []struct {
		age      int
		expected string
	}{
		{8, "childhood"},
		{15, "adolescence"},
		{22, "youth"},
		{30, "early_adult"},
		{45, "mid_adult"},
		{55, "mature"},
		{70, "elderly"},
	}

	for _, test := range tests {
		stage := engine.InferStageFromAge(test.age)
		if stage != test.expected {
			t.Errorf("Age %d should infer stage '%s', got '%s'", test.age, test.expected, stage)
		}
	}
}

// TestAddMilestone 测试添加里程碑
func TestAddMilestone(t *testing.T) {
	engine := NewLifeStageEngine()

	milestone := models.LifeMilestone{
		Title:        "首次升职",
		Description:  "成为团队领导",
		Significance: 0.7,
		Category:     "career",
	}

	err := engine.AddMilestone("identity_001", milestone)
	if err != nil {
		t.Fatalf("AddMilestone failed: %v", err)
	}

	// 验证里程碑添加
	profile := engine.GetLifeStageProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if len(profile.Milestones) != 1 {
		t.Errorf("Should have 1 milestone, got %d", len(profile.Milestones))
	}

	// 验证智慧积累
	if profile.WisdomAccumulated < 0.07 {
		t.Errorf("WisdomAccumulated should increase, got %f", profile.WisdomAccumulated)
	}
}

// TestAddTurningPoint 测试添加转折点
func TestAddTurningPoint(t *testing.T) {
	engine := NewLifeStageEngine()

	turningPoint := models.TurningPoint{
		Title:        "职业转型",
		Description:  "从技术转向管理",
		Significance: 0.8,
		Type:         "career",
		BeforeState:  "技术工程师",
		AfterState:   "技术经理",
	}

	err := engine.AddTurningPoint("identity_001", turningPoint)
	if err != nil {
		t.Fatalf("AddTurningPoint failed: %v", err)
	}

	// 验证转折点添加
	profile := engine.GetLifeStageProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if len(profile.TurningPoints) != 1 {
		t.Errorf("Should have 1 turning point, got %d", len(profile.TurningPoints))
	}
}

// TestListener 测试监听器完整功能
func TestLifeStageEngine_Listener(t *testing.T) {
	engine := NewLifeStageEngine()

	listener := &testLifeStageListener{
		stageChangedCount:    0,
		eventAddedCount:      0,
		lessonLearnedCount:   0,
		transitionCount:      0,
	}

	engine.AddListener(listener)

	// 设置阶段
	engine.SetCurrentStage("identity_001", "youth", 22)

	if listener.stageChangedCount != 1 {
		t.Errorf("stageChangedCount should be 1, got %d", listener.stageChangedCount)
	}

	// 添加事件
	engine.AddLifeEvent("identity_001", models.LifeEvent{EventTitle: "毕业"})

	if listener.eventAddedCount != 1 {
		t.Errorf("eventAddedCount should be 1, got %d", listener.eventAddedCount)
	}

	// 添加感悟
	engine.AddLifeLesson("identity_001", models.LifeLesson{Title: "学习的重要性"})

	if listener.lessonLearnedCount != 1 {
		t.Errorf("lessonLearnedCount should be 1, got %d", listener.lessonLearnedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次操作
	engine.SetCurrentStage("identity_001", "early_adult", 28)

	if listener.stageChangedCount != 1 {
		t.Errorf("stageChangedCount should still be 1 after removal, got %d", listener.stageChangedCount)
	}
}

// testLifeStageListener 测试监听器
type testLifeStageListener struct {
	stageChangedCount    int
	eventAddedCount      int
	lessonLearnedCount   int
	transitionCount      int
}

func (l *testLifeStageListener) OnLifeStageChanged(identityID string, stage *models.LifeStage) {
	l.stageChangedCount++
}

func (l *testLifeStageListener) OnLifeEventAdded(identityID string, event models.LifeEvent) {
	l.eventAddedCount++
}

func (l *testLifeStageListener) OnLifeLessonLearned(identityID string, lesson models.LifeLesson) {
	l.lessonLearnedCount++
}

func (l *testLifeStageListener) OnStageTransition(identityID string, fromStage, toStage string) {
	l.transitionCount++
}

// TestLifeStageNormalize 测试人生阶段归一化
func TestLifeStageNormalize(t *testing.T) {
	stage := models.NewLifeStage()
	stage.Completeness = 1.5
	stage.Satisfaction = -0.2

	stage.Normalize()

	if stage.Completeness != 1.0 {
		t.Errorf("Completeness should be normalized to 1.0, got %f", stage.Completeness)
	}
	if stage.Satisfaction != 0 {
		t.Errorf("Satisfaction should be normalized to 0, got %f", stage.Satisfaction)
	}
}

// TestDevelopmentMetricsCalculateOverallDevelopment 测试发展指标计算
func TestDevelopmentMetricsCalculate(t *testing.T) {
	metrics := models.DevelopmentMetrics{
		PhysicalHealth:      0.8,
		MentalHealth:        0.7,
		CognitiveGrowth:     0.6,
		EmotionalMaturity:   0.7,
		SocialDevelopment:   0.6,
		RelationshipQuality: 0.7,
		CareerProgress:      0.6,
		FinancialStability:  0.5,
		SkillDevelopment:    0.6,
		SelfAwareness:       0.7,
		PurposeClarity:      0.5,
		LifeSatisfaction:    0.7,
	}

	overall := metrics.CalculateOverallDevelopment()

	if overall < 0.5 || overall > 0.8 {
		t.Errorf("Overall development should be between 0.5 and 0.8, got %f", overall)
	}
}
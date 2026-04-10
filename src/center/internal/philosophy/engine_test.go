package philosophy

import (
	"context"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestPhilosophyEngineCreation 测试三观引擎创建
func TestPhilosophyEngineCreation(t *testing.T) {
	engine := NewPhilosophyEngine()

	if engine == nil {
		t.Fatal("PhilosophyEngine should not be nil")
	}
	if engine.worldviews == nil {
		t.Error("worldviews map should be initialized")
	}
	if engine.lifeViews == nil {
		t.Error("lifeViews map should be initialized")
	}
	if engine.valueSystems == nil {
		t.Error("valueSystems map should be initialized")
	}
}

// TestGetOrCreateWorldview 测试获取或创建世界观
func TestGetOrCreateWorldview(t *testing.T) {
	engine := NewPhilosophyEngine()

	// 创建新世界观
	worldview := engine.GetOrCreateWorldview("identity_001")
	if worldview == nil {
		t.Fatal("Worldview should not be nil")
	}

	// 验证默认值
	if worldview.Optimism != 0.5 {
		t.Errorf("Default Optimism should be 0.5, got %f", worldview.Optimism)
	}

	// 获取已存在的世界观
	worldview2 := engine.GetOrCreateWorldview("identity_001")
	if worldview != worldview2 {
		t.Error("Should return same worldview instance")
	}
}

// TestUpdateWorldview 测试更新世界观
func TestUpdateWorldview(t *testing.T) {
	engine := NewPhilosophyEngine()

	worldview := models.NewWorldview()
	worldview.Optimism = 0.8
	worldview.ChangeBelief = 0.7
	worldview.TrustInPeople = 0.6

	err := engine.UpdateWorldview("identity_001", worldview)
	if err != nil {
		t.Fatalf("UpdateWorldview failed: %v", err)
	}

	// 验证更新
	retrieved := engine.GetWorldview("identity_001")
	if retrieved == nil {
		t.Fatal("Worldview should exist after update")
	}
	if retrieved.Optimism != 0.8 {
		t.Errorf("Optimism should be 0.8, got %f", retrieved.Optimism)
	}
}

// TestUpdateWorldviewBelief 测试更新单个信念
func TestUpdateWorldviewBelief(t *testing.T) {
	engine := NewPhilosophyEngine()

	err := engine.UpdateWorldviewBelief("identity_001", "optimism", 0.9)
	if err != nil {
		t.Fatalf("UpdateWorldviewBelief failed: %v", err)
	}

	worldview := engine.GetWorldview("identity_001")
	if worldview == nil {
		t.Fatal("Worldview should exist")
	}
	if worldview.Optimism != 0.9 {
		t.Errorf("Optimism should be 0.9, got %f", worldview.Optimism)
	}
}

// TestGetOrCreateLifeView 测试获取或创建人生观
func TestGetOrCreateLifeView(t *testing.T) {
	engine := NewPhilosophyEngine()

	lifeView := engine.GetOrCreateLifeView("identity_001")
	if lifeView == nil {
		t.Fatal("LifeView should not be nil")
	}

	// 验证默认值
	if lifeView.FutureFocus != 0.5 {
		t.Errorf("Default FutureFocus should be 0.5, got %f", lifeView.FutureFocus)
	}

	// 获取已存在的人生观
	lifeView2 := engine.GetOrCreateLifeView("identity_001")
	if lifeView != lifeView2 {
		t.Error("Should return same lifeView instance")
	}
}

// TestUpdateLifeGoal 测试更新人生目标
func TestUpdateLifeGoal(t *testing.T) {
	engine := NewPhilosophyEngine()

	err := engine.UpdateLifeGoal("identity_001", "成为更好的自己")
	if err != nil {
		t.Fatalf("UpdateLifeGoal failed: %v", err)
	}

	lifeView := engine.GetLifeView("identity_001")
	if lifeView == nil {
		t.Fatal("LifeView should exist")
	}
	if lifeView.LifeGoal != "成为更好的自己" {
		t.Errorf("LifeGoal should be '成为更好的自己', got %s", lifeView.LifeGoal)
	}
}

// TestAddLifeMilestone 测试添加人生里程碑
func TestAddLifeMilestone(t *testing.T) {
	engine := NewPhilosophyEngine()

	// 先创建人生观
	engine.GetOrCreateLifeView("identity_001")

	milestone := models.LifeMilestone{
		MilestoneID:   "milestone_001",
		Title:         "大学毕业",
		Description:   "完成学业",
		AchievedAt:    time.Now(),
		Significance:  0.8,
		Impact:        0.9,
		Positive:      true,
	}

	err := engine.AddLifeMilestone("identity_001", milestone)
	if err != nil {
		t.Fatalf("AddLifeMilestone failed: %v", err)
	}

	// 验证里程碑添加
	profile := engine.lifeViewProfiles["identity_001"]
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if len(profile.Milestones) != 1 {
		t.Errorf("Should have 1 milestone, got %d", len(profile.Milestones))
	}

	// 重大里程碑应影响人生观
	lifeView := engine.GetLifeView("identity_001")
	if lifeView.HappinessLevel < 0.5 {
		t.Errorf("Positive milestone should increase happiness, got %f", lifeView.HappinessLevel)
	}
}

// TestGetOrCreateValueSystem 测试获取或创建价值观系统
func TestGetOrCreateValueSystem(t *testing.T) {
	engine := NewPhilosophyEngine()

	valueSystem := engine.GetOrCreateValueSystem("identity_001")
	if valueSystem == nil {
		t.Fatal("ValueSystem should not be nil")
	}

	// 验证默认值
	if valueSystem.Privacy != 0.5 {
		t.Errorf("Default Privacy should be 0.5, got %f", valueSystem.Privacy)
	}

	// 获取已存在的价值观
	valueSystem2 := engine.GetOrCreateValueSystem("identity_001")
	if valueSystem != valueSystem2 {
		t.Error("Should return same valueSystem instance")
	}
}

// TestUpdateValue 测试更新单个价值观
func TestUpdateValue(t *testing.T) {
	engine := NewPhilosophyEngine()

	err := engine.UpdateValue("identity_001", "privacy", 0.9)
	if err != nil {
		t.Fatalf("UpdateValue failed: %v", err)
	}

	valueSystem := engine.GetValueSystem("identity_001")
	if valueSystem == nil {
		t.Fatal("ValueSystem should exist")
	}
	if valueSystem.Privacy != 0.9 {
		t.Errorf("Privacy should be 0.9, got %f", valueSystem.Privacy)
	}
}

// TestMakeValueJudgment 测试价值判断
func TestMakeValueJudgment(t *testing.T) {
	engine := NewPhilosophyEngine()

	options := []models.ValueOption{
		{ID: "option_1", Name: "保守方案", Values: map[string]float64{"privacy": 0.9, "efficiency": 0.5}},
		{ID: "option_2", Name: "激进方案", Values: map[string]float64{"privacy": 0.3, "efficiency": 0.9}},
	}

	judgment, err := engine.MakeValueJudgment(context.Background(), "identity_001", "隐私与效率权衡", options)
	if err != nil {
		t.Fatalf("MakeValueJudgment failed: %v", err)
	}

	if judgment == nil {
		t.Fatal("Judgment should not be nil")
	}

	// 验证判断结果
	if judgment.Situation != "隐私与效率权衡" {
		t.Errorf("Situation should be preserved, got %s", judgment.Situation)
	}

	// 判断历史应记录
	profile := engine.valueProfiles["identity_001"]
	if profile == nil || len(profile.JudgmentHistory) == 0 {
		t.Error("Judgment should be recorded in history")
	}
}

// TestResolveValueConflict 测试解决价值观冲突
func TestResolveValueConflict(t *testing.T) {
	engine := NewPhilosophyEngine()

	conflict, err := engine.ResolveValueConflict("identity_001", "privacy", "efficiency", "数据收集场景")
	if err != nil {
		t.Fatalf("ResolveValueConflict failed: %v", err)
	}

	if conflict == nil {
		t.Fatal("Conflict should not be nil")
	}

	// 验证冲突结果
	if conflict.Value1 != "privacy" {
		t.Errorf("Value1 should be privacy, got %s", conflict.Value1)
	}
	if conflict.Value2 != "efficiency" {
		t.Errorf("Value2 should be efficiency, got %s", conflict.Value2)
	}

	// 冲突历史应记录
	profile := engine.valueProfiles["identity_001"]
	if profile == nil || len(profile.ConflictHistory) == 0 {
		t.Error("Conflict should be recorded in history")
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewPhilosophyEngine()

	// 设置三观
	worldview := engine.GetOrCreateWorldview("identity_001")
	worldview.Optimism = 0.8
	worldview.ChangeBelief = 0.7

	lifeView := engine.GetOrCreateLifeView("identity_001")
	lifeView.FutureFocus = 0.8
	lifeView.LifeGoal = "追求卓越"

	valueSystem := engine.GetOrCreateValueSystem("identity_001")
	valueSystem.Justice = 0.7
	valueSystem.Freedom = 0.8

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

	// 验证决策倾向计算
	if context.DecisionTendencies.RiskTolerance < 0.5 {
		t.Errorf("High optimism should increase RiskTolerance, got %f", context.DecisionTendencies.RiskTolerance)
	}
	if context.DecisionTendencies.DelayedGratification != 0.8 {
		t.Errorf("DelayedGratification should equal FutureFocus, got %f", context.DecisionTendencies.DelayedGratification)
	}

	// 验证人生目标包含
	if len(context.CurrentGoals) == 0 {
		t.Error("CurrentGoals should contain LifeGoal")
	}

	// 验证道德指南
	if context.MoralGuidance.ConflictResolution == "" {
		t.Error("MoralGuidance should be calculated")
	}
}

// TestListener 测试监听器
func TestListener(t *testing.T) {
	engine := NewPhilosophyEngine()

	// 创建测试监听器
	listener := &testPhilosophyListener{
		worldviewChangedCount:  0,
		lifeViewChangedCount:   0,
		valueSystemChangedCount: 0,
		contextUpdatedCount:    0,
	}

	engine.AddListener(listener)

	// 更新世界观
	worldview := models.NewWorldview()
	worldview.Optimism = 0.8
	engine.UpdateWorldview("identity_001", worldview)

	if listener.worldviewChangedCount != 1 {
		t.Errorf("worldviewChangedCount should be 1, got %d", listener.worldviewChangedCount)
	}

	// 更新人生观
	lifeView := models.NewLifeView()
	lifeView.FutureFocus = 0.7
	engine.UpdateLifeView("identity_001", lifeView)

	if listener.lifeViewChangedCount != 1 {
		t.Errorf("lifeViewChangedCount should be 1, got %d", listener.lifeViewChangedCount)
	}

	// 更新价值观
	valueSystem := models.NewEnhancedValueSystem()
	valueSystem.Privacy = 0.9
	engine.UpdateValueSystem("identity_001", valueSystem)

	if listener.valueSystemChangedCount != 1 {
		t.Errorf("valueSystemChangedCount should be 1, got %d", listener.valueSystemChangedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次更新
	engine.UpdateWorldview("identity_001", worldview)

	if listener.worldviewChangedCount != 1 {
		t.Errorf("worldviewChangedCount should still be 1 after removal, got %d", listener.worldviewChangedCount)
	}
}

// testPhilosophyListener 测试监听器
type testPhilosophyListener struct {
	worldviewChangedCount  int
	lifeViewChangedCount   int
	valueSystemChangedCount int
	contextUpdatedCount    int
}

func (l *testPhilosophyListener) OnWorldviewChanged(identityID string, worldview *models.Worldview) {
	l.worldviewChangedCount++
}

func (l *testPhilosophyListener) OnLifeViewChanged(identityID string, lifeView *models.LifeView) {
	l.lifeViewChangedCount++
}

func (l *testPhilosophyListener) OnValueSystemChanged(identityID string, valueSystem *models.EnhancedValueSystem) {
	l.valueSystemChangedCount++
}

func (l *testPhilosophyListener) OnPhilosophyContextUpdated(identityID string, context *PhilosophyDecisionContext) {
	l.contextUpdatedCount++
}

// TestWorldviewNormalize 测试世界观归一化
func TestWorldviewNormalize(t *testing.T) {
	worldview := models.NewWorldview()
	worldview.Optimism = 1.5   // 超出范围
	worldview.Individualism = -0.2 // 低于范围

	worldview.Normalize()

	if worldview.Optimism != 1.0 {
		t.Errorf("Optimism should be normalized to 1.0, got %f", worldview.Optimism)
	}
	if worldview.Individualism != 0 {
		t.Errorf("Individualism should be normalized to 0, got %f", worldview.Individualism)
	}
}

// TestValueSystemNormalize 测试价值观归一化
func TestValueSystemNormalize(t *testing.T) {
	valueSystem := models.NewEnhancedValueSystem()
	valueSystem.Privacy = 1.2
	valueSystem.Family = -0.1

	valueSystem.Normalize()

	if valueSystem.Privacy != 1.0 {
		t.Errorf("Privacy should be normalized to 1.0, got %f", valueSystem.Privacy)
	}
	if valueSystem.Family != 0 {
		t.Errorf("Family should be normalized to 0, got %f", valueSystem.Family)
	}
}
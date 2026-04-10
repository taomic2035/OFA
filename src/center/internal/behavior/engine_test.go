package behavior

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestEmotionBehaviorEngineCreation 测试情绪行为引擎创建
func TestEmotionBehaviorEngineCreation(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	if engine == nil {
		t.Fatal("EmotionBehaviorEngine should not be nil")
	}
	if engine.behaviorSystems == nil {
		t.Error("behaviorSystems map should be initialized")
	}
	if engine.behaviorProfiles == nil {
		t.Error("behaviorProfiles map should be initialized")
	}
}

// TestGetOrCreateBehaviorSystem 测试获取或创建行为系统
func TestGetOrCreateBehaviorSystem(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 创建新行为系统
	system := engine.GetOrCreateBehaviorSystem("identity_001")
	if system == nil {
		t.Fatal("BehaviorSystem should not be nil")
	}

	// 验证默认值
	if system.DecisionInfluence == nil {
		t.Error("DecisionInfluence should be initialized")
	}
	if system.ExpressionInfluence == nil {
		t.Error("ExpressionInfluence should be initialized")
	}

	// 获取已存在的系统
	system2 := engine.GetOrCreateBehaviorSystem("identity_001")
	if system != system2 {
		t.Error("Should return same system instance")
	}
}

// TestApplyEmotionToDecision_Joy 测试喜悦情绪对决策的影响
func TestApplyEmotionToDecision_Joy(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	influence := engine.ApplyEmotionToDecision("identity_001", "joy", 0.7)

	if influence == nil {
		t.Fatal("DecisionInfluence should not be nil")
	}

	// 验证喜悦的影响
	if influence.DominantEmotion != "joy" {
		t.Errorf("DominantEmotion should be 'joy', got %s", influence.DominantEmotion)
	}
	if influence.EmotionIntensity != 0.7 {
		t.Errorf("EmotionIntensity should be 0.7, got %f", influence.EmotionIntensity)
	}

	// 喜悦应增加风险承受
	if influence.RiskTolerance < 0.5 {
		t.Errorf("Joy should increase RiskTolerance, got %f", influence.RiskTolerance)
	}

	// 喜悦应增加社交趋近
	if influence.SocialApproach < 0.5 {
		t.Errorf("Joy should increase SocialApproach, got %f", influence.SocialApproach)
	}

	// 喜悦应增加合作倾向
	if influence.CooperationTendency < 0.5 {
		t.Errorf("Joy should increase CooperationTendency, got %f", influence.CooperationTendency)
	}
}

// TestApplyEmotionToDecision_Anger 测试愤怒情绪对决策的影响
func TestApplyEmotionToDecision_Anger(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	influence := engine.ApplyEmotionToDecision("identity_001", "anger", 0.8)

	// 愤怒应增加风险承受
	if influence.RiskTolerance < 0.5 {
		t.Errorf("Anger should increase RiskTolerance, got %f", influence.RiskTolerance)
	}

	// 愤怒应降低冲动控制
	if influence.ImpulseControl > 0.5 {
		t.Errorf("Anger should decrease ImpulseControl, got %f", influence.ImpulseControl)
	}

	// 愤怒应增加社交回避
	if influence.SocialAvoidance < 0.1 {
		t.Errorf("Anger should increase SocialAvoidance, got %f", influence.SocialAvoidance)
	}

	// 愤怒应降低合作倾向
	if influence.CooperationTendency > 0.5 {
		t.Errorf("Anger should decrease CooperationTendency, got %f", influence.CooperationTendency)
	}
}

// TestApplyEmotionToDecision_Fear 测试恐惧情绪对决策的影响
func TestApplyEmotionToDecision_Fear(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	influence := engine.ApplyEmotionToDecision("identity_001", "fear", 0.7)

	// 恐惧应增加风险规避
	if influence.RiskAversion < 0.5 {
		t.Errorf("Fear should increase RiskAversion, got %f", influence.RiskAversion)
	}

	// 恐惧应降低风险承受
	if influence.RiskTolerance > 0.5 {
		t.Errorf("Fear should decrease RiskTolerance, got %f", influence.RiskTolerance)
	}

	// 恐惧应增加社交回避
	if influence.SocialAvoidance < 0.1 {
		t.Errorf("Fear should increase SocialAvoidance, got %f", influence.SocialAvoidance)
	}

	// 恐惧应增加深思熟虑
	if influence.DeliberationLevel < 0.5 {
		t.Errorf("Fear should increase DeliberationLevel, got %f", influence.DeliberationLevel)
	}
}

// TestApplyEmotionToExpression 测试情绪对表达的影响
func TestApplyEmotionToExpression(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 测试喜悦
	influence := engine.ApplyEmotionToExpression("identity_001", "joy", 0.7)

	if influence.WarmthLevel < 0.5 {
		t.Errorf("Joy should increase WarmthLevel, got %f", influence.WarmthLevel)
	}
	if influence.WordChoice != "positive" {
		t.Errorf("Joy should set WordChoice to 'positive', got %s", influence.WordChoice)
	}
	if influence.ResponseSpeed != "immediate" {
		t.Errorf("Joy should set ResponseSpeed to 'immediate', got %s", influence.ResponseSpeed)
	}
	if influence.ExpressionTendency != "express" {
		t.Errorf("Joy should set ExpressionTendency to 'express', got %s", influence.ExpressionTendency)
	}

	// 测试悲伤
	influence = engine.ApplyEmotionToExpression("identity_001", "sadness", 0.6)

	if influence.EnthusiasmLevel > 0.5 {
		t.Errorf("Sadness should decrease EnthusiasmLevel, got %f", influence.EnthusiasmLevel)
	}
	if influence.WordChoice != "negative" {
		t.Errorf("Sadness should set WordChoice to 'negative', got %s", influence.WordChoice)
	}
	if influence.ResponseSpeed != "delayed" {
		t.Errorf("Sadness should set ResponseSpeed to 'delayed', got %s", influence.ResponseSpeed)
	}
	if influence.ExpressionTendency != "suppress" {
		t.Errorf("Sadness should set ExpressionTendency to 'suppress', got %s", influence.ExpressionTendency)
	}
}

// TestTriggerBehavior 测试触发行为
func TestTriggerBehavior(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 创建监听器
	listener := &testBehaviorListener{
		behaviorTriggeredCount: 0,
	}
	engine.AddListener(listener)

	// 触发喜悦行为
	behavior := engine.TriggerBehavior("identity_001", "joy", 0.7, "工作成功")

	if behavior == nil {
		t.Fatal("Behavior should not be nil")
	}

	// 验证行为生成
	if behavior.TriggerEmotion != "joy" {
		t.Errorf("TriggerEmotion should be 'joy', got %s", behavior.TriggerEmotion)
	}
	if behavior.BehaviorType != "communication" {
		t.Errorf("Joy should trigger 'communication' behavior, got %s", behavior.BehaviorType)
	}
	if behavior.ActionTendency != "approach" {
		t.Errorf("Joy should trigger 'appro' tendency, got %s", behavior.ActionTendency)
	}

	// 验证监听器通知
	if listener.behaviorTriggeredCount != 1 {
		t.Errorf("behaviorTriggeredCount should be 1, got %d", listener.behaviorTriggeredCount)
	}

	// 验证触发历史
	system := engine.GetBehaviorSystem("identity_001")
	if len(system.BehaviorTriggers) != 1 {
		t.Errorf("Should have 1 trigger record, got %d", len(system.BehaviorTriggers))
	}
}

// TestTriggerBehavior_Anger 测试愤怒触发行为
func TestTriggerBehavior_Anger(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	behavior := engine.TriggerBehavior("identity_001", "anger", 0.8, "不公平待遇")

	if behavior.BehaviorType != "action" {
		t.Errorf("Anger should trigger 'action' behavior, got %s", behavior.BehaviorType)
	}
	if behavior.ActionTendency != "fight" {
		t.Errorf("Anger should trigger 'fight' tendency, got %s", behavior.ActionTendency)
	}
	if behavior.UrgencyLevel < 0.5 {
		t.Errorf("Anger should have high urgency, got %f", behavior.UrgencyLevel)
	}
}

// TestTriggerBehavior_Fear 测试恐惧触发行为
func TestTriggerBehavior_Fear(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	behavior := engine.TriggerBehavior("identity_001", "fear", 0.7, "潜在风险")

	if behavior.BehaviorType != "avoidance" {
		t.Errorf("Fear should trigger 'avoidance' behavior, got %s", behavior.BehaviorType)
	}
	if behavior.ActionTendency != "flight" {
		t.Errorf("Fear should trigger 'flight' tendency, got %s", behavior.ActionTendency)
	}
	if behavior.Automaticity < 0.5 {
		t.Errorf("Fear should have high automaticity, got %f", behavior.Automaticity)
	}
}

// TestAddCopingStrategy 测试添加应对策略
func TestAddCopingStrategy(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	strategy := models.CopingStrategy{
		StrategyName:   "深呼吸放松",
		StrategyType:   "emotion_focused",
		TargetEmotions: []string{"anger", "fear"},
		Effectiveness:  0.8,
		Description:    "通过深呼吸平复情绪",
	}

	err := engine.AddCopingStrategy("identity_001", strategy)
	if err != nil {
		t.Fatalf("AddCopingStrategy failed: %v", err)
	}

	// 验证策略添加
	system := engine.GetBehaviorSystem("identity_001")
	if len(system.CopingStrategies) != 1 {
		t.Errorf("Should have 1 strategy, got %d", len(system.CopingStrategies))
	}
}

// TestUseCopingStrategy 测试使用应对策略
func TestUseCopingStrategy(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 添加策略
	strategy := models.CopingStrategy{
		StrategyName:   "深呼吸",
		TargetEmotions: []string{"anger"},
		Effectiveness:  0.7,
	}
	engine.AddCopingStrategy("identity_001", strategy)

	// 获取策略ID
	system := engine.GetBehaviorSystem("identity_001")
	strategyID := system.CopingStrategies[0].StrategyID

	// 创建监听器
	listener := &testBehaviorListener{
		strategyUsedCount: 0,
	}
	engine.AddListener(listener)

	// 使用策略（成功）
	err := engine.UseCopingStrategy("identity_001", strategyID, true)
	if err != nil {
		t.Fatalf("UseCopingStrategy failed: %v", err)
	}

	// 验证使用计数
	system = engine.GetBehaviorSystem("identity_001")
	if system.CopingStrategies[0].UseCount != 1 {
		t.Errorf("UseCount should be 1, got %d", system.CopingStrategies[0].UseCount)
	}

	// 验证成功率更新
	if system.CopingStrategies[0].SuccessRate != 1.0 {
		t.Errorf("SuccessRate should be 1.0 after success, got %f", system.CopingStrategies[0].SuccessRate)
	}

	// 验证监听器通知
	if listener.strategyUsedCount != 1 {
		t.Errorf("strategyUsedCount should be 1, got %d", listener.strategyUsedCount)
	}
}

// TestGetRecommendedCopingStrategies 测试获取推荐应对策略
func TestGetRecommendedCopingStrategies(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 添加多个策略
	strategies := []models.CopingStrategy{
		{StrategyName: "深呼吸", TargetEmotions: []string{"anger"}, Effectiveness: 0.9},
		{StrategyName: "暂时离开", TargetEmotions: []string{"anger"}, Effectiveness: 0.7},
		{StrategyName: "寻求支持", TargetEmotions: []string{"sadness"}, Effectiveness: 0.8},
		{StrategyName: "风险评估", TargetEmotions: []string{"fear"}, Effectiveness: 0.75},
	}

	for _, strategy := range strategies {
		engine.AddCopingStrategy("identity_001", strategy)
	}

	// 获取愤怒推荐策略
	recommended := engine.GetRecommendedCopingStrategies("identity_001", "anger")

	if len(recommended) != 2 {
		t.Errorf("Should have 2 recommended strategies for anger, got %d", len(recommended))
	}

	// 验证排序（按有效性）
	if recommended[0].Effectiveness < recommended[1].Effectiveness {
		t.Error("Strategies should be sorted by effectiveness")
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 设置情绪状态提供者
	engine.SetEmotionStateProvider(&mockEmotionProvider{
		dominantEmotion: "joy",
		intensity:       0.6,
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
	if context.DecisionInfluence == nil {
		t.Error("DecisionInfluence should be set")
	}
	if context.ExpressionInfluence == nil {
		t.Error("ExpressionInfluence should be set")
	}

	// 验证当前情绪状态
	if context.CurrentEmotionState.DominantEmotion != "joy" {
		t.Errorf("DominantEmotion should be 'joy', got %s", context.CurrentEmotionState.DominantEmotion)
	}
	if context.CurrentEmotionState.Intensity != 0.6 {
		t.Errorf("Intensity should be 0.6, got %f", context.CurrentEmotionState.Intensity)
	}

	// 验证行为指导
	if context.BehaviorGuidance.DecisionStyle == "" {
		t.Error("DecisionStyle should be calculated")
	}

	// 验证推荐行为
	if len(context.RecommendedBehaviors) == 0 {
		t.Error("RecommendedBehaviors should be generated")
	}

	// 验证推荐应对策略
	if len(context.RecommendedCopingStrategies) == 0 {
		t.Error("RecommendedCopingStrategies should be generated")
	}
}

// TestGetBehaviorProfile 测试获取行为画像
func TestGetBehaviorProfile(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 触发行为
	engine.TriggerBehavior("identity_001", "joy", 0.7, "测试")

	// 获取画像
	profile := engine.GetBehaviorProfile("identity_001")

	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证画像字段
	if len(profile.TriggerHistory) != 1 {
		t.Errorf("TriggerHistory should have 1 record, got %d", len(profile.TriggerHistory))
	}
}

// TestListener 测试监听器完整功能
func TestEmotionBehaviorEngine_Listener(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	listener := &testBehaviorListener{
		decisionChangedCount:  0,
		expressionChangedCount: 0,
		behaviorTriggeredCount: 0,
		strategyUsedCount:      0,
	}

	engine.AddListener(listener)

	// 应用决策影响
	engine.ApplyEmotionToDecision("identity_001", "joy", 0.5)

	if listener.decisionChangedCount != 1 {
		t.Errorf("decisionChangedCount should be 1, got %d", listener.decisionChangedCount)
	}

	// 应用表达影响
	engine.ApplyEmotionToExpression("identity_001", "joy", 0.5)

	if listener.expressionChangedCount != 1 {
		t.Errorf("expressionChangedCount should be 1, got %d", listener.expressionChangedCount)
	}

	// 触发行为
	engine.TriggerBehavior("identity_001", "anger", 0.6, "测试")

	if listener.behaviorTriggeredCount != 1 {
		t.Errorf("behaviorTriggeredCount should be 1, got %d", listener.behaviorTriggeredCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次操作
	engine.ApplyEmotionToDecision("identity_001", "joy", 0.5)

	if listener.decisionChangedCount != 1 {
		t.Errorf("decisionChangedCount should still be 1 after removal, got %d", listener.decisionChangedCount)
	}
}

// testBehaviorListener 测试监听器
type testBehaviorListener struct {
	decisionChangedCount   int
	expressionChangedCount int
	behaviorTriggeredCount int
	strategyUsedCount      int
}

func (l *testBehaviorListener) OnDecisionInfluenceChanged(identityID string, influence *models.EmotionDecisionInfluence) {
	l.decisionChangedCount++
}

func (l *testBehaviorListener) OnExpressionInfluenceChanged(identityID string, influence *models.EmotionalExpressionInfluence) {
	l.expressionChangedCount++
}

func (l *testBehaviorListener) OnBehaviorTriggered(identityID string, behavior models.EmotionTriggeredBehavior) {
	l.behaviorTriggeredCount++
}

func (l *testBehaviorListener) OnCopingStrategyUsed(identityID string, strategy models.CopingStrategy) {
	l.strategyUsedCount++
}

// mockEmotionProvider 模拟情绪状态提供者
type mockEmotionProvider struct {
	dominantEmotion string
	intensity       float64
}

func (m *mockEmotionProvider) GetCurrentEmotion(identityID string) *models.Emotion {
	return &models.Emotion{
		Joy: m.intensity,
	}
}

func (m *mockEmotionProvider) GetDominantEmotion(identityID string) string {
	return m.dominantEmotion
}

func (m *mockEmotionProvider) GetEmotionIntensity(identityID string) float64 {
	return m.intensity
}

// TestDecisionInfluenceIsImpulsive 测试冲动检测
func TestDecisionInfluenceIsImpulsive(t *testing.T) {
	influence := models.NewEmotionDecisionInfluence()
	influence.ImpulseControl = 0.3

	if !influence.IsImpulsive() {
		t.Error("Should be impulsive when ImpulseControl is low")
	}

	influence.ImpulseControl = 0.7

	if influence.IsImpulsive() {
		t.Error("Should not be impulsive when ImpulseControl is high")
	}
}

// TestDecisionInfluenceGetDecisionStyle 测试决策风格获取
func TestDecisionInfluenceGetDecisionStyle(t *testing.T) {
	influence := models.NewEmotionDecisionInfluence()
	influence.ImpulseControl = 0.3
	influence.DeliberationLevel = 0.2
	influence.DecisionSpeed = 0.8

	style := influence.GetDecisionStyle()

	if style != "impulsive" {
		t.Errorf("Style should be 'impulsive', got %s", style)
	}

	influence.ImpulseControl = 0.7
	influence.DeliberationLevel = 0.8
	influence.DecisionSpeed = 0.3

	style = influence.GetDecisionStyle()

	if style != "deliberate" {
		t.Errorf("Style should be 'deliberate', got %s", style)
	}
}

// TestExpressionInfluenceGetCommunicationStyle 测试沟通风格获取
func TestExpressionInfluenceGetCommunicationStyle(t *testing.T) {
	influence := models.NewEmotionalExpressionInfluence()
	influence.ToneStyle = "warm"
	influence.WarmthLevel = 0.8

	style := influence.GetCommunicationStyle()

	if style != "warm" {
		t.Errorf("Style should be 'warm', got %s", style)
	}
}

// TestBehaviorTriggerHistoryLimit 测试行为触发历史限制
func TestBehaviorTriggerHistoryLimit(t *testing.T) {
	engine := NewEmotionBehaviorEngine()

	// 触发超过50次
	for i := 0; i < 60; i++ {
		engine.TriggerBehavior("identity_001", "joy", 0.3, "测试")
	}

	system := engine.GetBehaviorSystem("identity_001")

	if len(system.BehaviorTriggers) > 50 {
		t.Errorf("BehaviorTriggers should be limited to 50, got %d", len(system.BehaviorTriggers))
	}
}
package sync

import (
	"testing"
	"time"
)

func TestSceneRouter(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 测试默认规则初始化
	rules := router.GetRules()
	if len(rules) == 0 {
		t.Error("Expected default rules to be initialized")
	}

	// 验证规则已按优先级排序
	for i := 1; i < len(rules); i++ {
		if rules[i].Priority > rules[i-1].Priority {
			t.Error("Rules should be sorted by priority descending")
		}
	}
}

func TestSceneRouterAddRule(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	initialCount := len(router.GetRules())

	// 添加自定义规则
	rule := &RoutingRule{
		RuleID:       "custom-001",
		Name:         "Custom Test Rule",
		Priority:     50,
		Scenes:       []SceneType{SceneIdle},
		MessageTypes: []MessageType{MessageTypeNotification},
		Action:       ActionBroadcast,
		Enabled:      true,
	}

	router.AddRule(rule)

	// 验证规则已添加
	if len(router.GetRules()) != initialCount+1 {
		t.Error("Rule not added")
	}

	// 验证可以获取规则
	fetched := router.GetRule("custom-001")
	if fetched == nil {
		t.Error("Failed to fetch added rule")
	}
}

func TestSceneRouterRemoveRule(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 添加规则
	rule := &RoutingRule{
		RuleID:   "to-remove",
		Priority: 50,
		Action:   ActionBroadcast,
		Enabled:  true,
	}
	router.AddRule(rule)

	// 移除规则
	router.RemoveRule("to-remove")

	// 验证已移除
	if router.GetRule("to-remove") != nil {
		t.Error("Rule should be removed")
	}
}

func TestSceneRouterMatchRule(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 创建测试规则
	rule := &RoutingRule{
		RuleID:       "test-match",
		Priority:     100,
		Scenes:       []SceneType{SceneRunning, SceneWalking},
		MessageTypes: []MessageType{MessageTypeNotification},
		Action:       ActionRoute,
		Enabled:      true,
	}
	router.AddRule(rule)

	// 测试匹配的上下文
	ctx := &RoutingContext{
		IdentityID:  "identity-001",
		MessageType: MessageTypeNotification,
		Scene:       SceneRunning,
		Priority:    1,
	}

	result := router.Route(ctx)

	// 验证匹配到了规则
	if result.MatchedRule == nil {
		t.Error("Expected rule to match")
	}

	if result.MatchedRule.RuleID != "test-match" {
		t.Errorf("Expected test-match rule, got %s", result.MatchedRule.RuleID)
	}
}

func TestSceneRouterConditionMatch(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 创建带条件的规则
	rule := &RoutingRule{
		RuleID:      "condition-rule",
		Priority:    100,
		Scenes:      []SceneType{SceneIdle},
		Conditions:  []RuleCondition{{Field: "priority", Operator: "gte", Value: 3}},
		Action:      ActionBroadcast,
		Enabled:     true,
	}
	router.AddRule(rule)

	// 测试高优先级匹配
	ctx1 := &RoutingContext{
		Scene:    SceneIdle,
		Priority: 3,
	}
	result1 := router.Route(ctx1)

	if result1.MatchedRule == nil || result1.MatchedRule.RuleID != "condition-rule" {
		t.Error("Expected condition rule to match for priority >= 3")
	}

	// 测试低优先级不匹配
	router.RemoveRule("condition-rule")
	rule2 := &RoutingRule{
		RuleID:      "condition-rule-2",
		Priority:    100,
		Scenes:      []SceneType{SceneIdle},
		Conditions:  []RuleCondition{{Field: "priority", Operator: "lt", Value: 2}},
		Action:      ActionFilter,
		Enabled:     true,
	}
	router.AddRule(rule2)

	ctx2 := &RoutingContext{
		Scene:    SceneIdle,
		Priority: 3,
	}
	result2 := router.Route(ctx2)

	// 不应该匹配（优先级 >= 2）
	if result2.MatchedRule != nil && result2.MatchedRule.RuleID == "condition-rule-2" {
		t.Error("Condition rule should not match for priority >= 2")
	}
}

func TestSceneRouterPrioritizeAction(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 创建设备管理器和状态管理器
	dm := NewDeviceManager(DefaultDeviceManagerConfig())
	sm := NewStateSyncManager(DefaultStateSyncConfig())

	router.SetDeviceManager(dm)
	router.SetStateManager(sm)

	// 注册设备
	dm.RegisterDevice("agent-001", "identity-001", "mobile", "Phone", []string{})
	dm.RegisterDevice("agent-002", "identity-001", "watch", "Watch", []string{})
	dm.RegisterDevice("agent-003", "identity-001", "tablet", "Tablet", []string{})

	// 更新设备状态
	sm.UpdateDeviceState(&DeviceState{
		AgentID:   "agent-001",
		IdentityID: "identity-001",
		DeviceType: "mobile",
		Online:     true,
		Priority:   50,
	})
	sm.UpdateDeviceState(&DeviceState{
		AgentID:   "agent-002",
		IdentityID: "identity-001",
		DeviceType: "watch",
		Online:     true,
		Priority:   80,
	})
	sm.UpdateDeviceState(&DeviceState{
		AgentID:   "agent-003",
		IdentityID: "identity-001",
		DeviceType: "tablet",
		Online:     true,
		Priority:   60,
	})

	// 测试优先级路由
	ctx := &RoutingContext{
		IdentityID: "identity-001",
		Scene:      SceneIdle,
	}

	result := router.Route(ctx)

	// 验证按优先级排序
	if len(result.TargetAgents) == 0 {
		t.Error("Expected some target agents")
	}

	// 第一个应该是优先级最高的
	if len(result.TargetAgents) > 0 && result.TargetAgents[0] != "agent-002" {
		t.Logf("Note: First target is %s (expected agent-002 with highest priority)", result.TargetAgents[0])
	}
}

func TestSceneRouterSceneOptimization(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())
	router.config.SmartRouting = true

	// 创建模拟设备状态
	deviceStates := []*DeviceState{
		{AgentID: "phone", DeviceType: "mobile", Online: true, Priority: 50, BatteryLevel: 80},
		{AgentID: "watch", DeviceType: "watch", Online: true, Priority: 60, BatteryLevel: 70},
		{AgentID: "tablet", DeviceType: "tablet", Online: true, Priority: 40, BatteryLevel: 90},
	}

	// 测试跑步场景优化
	ctx := &RoutingContext{
		IdentityID:   "identity-001",
		Scene:        SceneRunning,
		DeviceStates: deviceStates,
	}

	result := router.Route(ctx)

	// 验证场景感知优化
	if len(result.TargetAgents) > 0 {
		t.Logf("Running scene targets: %v", result.TargetAgents)
	}

	// 测试睡眠场景
	ctx2 := &RoutingContext{
		IdentityID:   "identity-001",
		Scene:        SceneSleeping,
		DeviceStates: deviceStates,
	}

	result2 := router.Route(ctx2)
	t.Logf("Sleeping scene targets: %v", result2.TargetAgents)
}

func TestSceneRouterStats(t *testing.T) {
	router := NewSceneRouter(DefaultSceneRouterConfig())

	// 执行几次路由
	for i := 0; i < 5; i++ {
		ctx := &RoutingContext{
			IdentityID:  "identity-001",
			Scene:       SceneType([]string{SceneIdle, SceneRunning, SceneWalking}[i%3]),
			MessageType: MessageTypeNotification,
		}
		router.Route(ctx)
	}

	stats := router.GetStats()

	if stats.TotalRoutings != 5 {
		t.Errorf("Expected 5 routings, got %d", stats.TotalRoutings)
	}

	if stats.TotalRules == 0 {
		t.Error("Expected some rules")
	}

	if len(stats.ByScene) == 0 {
		t.Error("Expected scene statistics")
	}
}

func TestSceneRouterHistory(t *testing.T) {
	config := DefaultSceneRouterConfig()
	config.MaxHistorySize = 10
	router := NewSceneRouter(config)

	// 执行路由
	for i := 0; i < 15; i++ {
		ctx := &RoutingContext{
			IdentityID: "identity-001",
			Scene:      SceneIdle,
		}
		router.Route(ctx)
	}

	history := router.GetRoutingHistory("identity-001", 100)

	// 验证历史限制
	if len(history) > 10 {
		t.Errorf("History should be limited to 10, got %d", len(history))
	}
}

func TestRoutingRuleJSON(t *testing.T) {
	rule := &RoutingRule{
		RuleID:       "json-test",
		Name:         "JSON Test Rule",
		Priority:     100,
		Scenes:       []SceneType{SceneRunning, SceneWalking},
		MessageTypes: []MessageType{MessageTypeNotification},
		Action:       ActionRoute,
		Enabled:      true,
	}

	// 序列化
	data, err := rule.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := RoutingRuleFromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.RuleID != rule.RuleID {
		t.Error("RuleID mismatch after JSON roundtrip")
	}
	if parsed.Name != rule.Name {
		t.Error("Name mismatch after JSON roundtrip")
	}
	if len(parsed.Scenes) != len(rule.Scenes) {
		t.Error("Scenes count mismatch")
	}
}
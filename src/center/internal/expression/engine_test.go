package expression

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestExpressionGestureEngineCreation 测试引擎创建
func TestExpressionGestureEngineCreation(t *testing.T) {
	config := ExpressionGestureEngineConfig{
		DefaultExpression:      "neutral",
		DefaultGesture:         "natural",
		DefaultExpressionRange: 0.7,
		DefaultGestureRange:    0.6,
		EmotionInfluenceStrength: 0.8,
	}
	engine := NewExpressionGestureEngine(config)

	if engine == nil {
		t.Fatal("ExpressionGestureEngine should not be nil")
	}
	if engine.config.DefaultExpression != "neutral" {
		t.Errorf("DefaultExpression should be neutral, got %s", engine.config.DefaultExpression)
	}
}

// TestDefaultExpressionGestureEngineConfig 测试默认配置
func TestDefaultExpressionGestureEngineConfig(t *testing.T) {
	config := DefaultExpressionGestureEngineConfig()

	if config.DefaultExpression != "neutral" {
		t.Errorf("Default DefaultExpression should be neutral, got %s", config.DefaultExpression)
	}
	if config.DefaultAnimationFPS != 30 {
		t.Errorf("DefaultAnimationFPS should be 30, got %d", config.DefaultAnimationFPS)
	}
	if config.EmotionInfluenceStrength != 0.8 {
		t.Errorf("EmotionInfluenceStrength should be 0.8, got %f", config.EmotionInfluenceStrength)
	}
}

// TestCreateProfile 测试创建 Profile
func TestCreateProfile(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	profile := engine.CreateProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证默认值
	if profile.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", profile.IdentityID)
	}
	if profile.FacialExpressionSettings.DefaultExpression != "neutral" {
		t.Errorf("DefaultExpression should be neutral, got %s", profile.FacialExpressionSettings.DefaultExpression)
	}
	if profile.Version != 1 {
		t.Errorf("Version should be 1, got %d", profile.Version)
	}

	// 验证情绪映射
	if len(profile.EmotionExpressionMapping.EmotionMappings) == 0 {
		t.Error("EmotionMappings should have default mappings")
	}

	// 验证社交手势设置
	if profile.SocialGestureSettings.GreetingGesture == "" {
		t.Error("GreetingGesture should be set")
	}
}

// TestGetProfile 测试获取 Profile
func TestGetProfile(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 获取存在的 Profile
	profile := engine.GetProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}

	// 获取不存在的 Profile
	profile2 := engine.GetProfile("identity_999")
	if profile2 != nil {
		t.Error("Profile should not exist for unknown identity")
	}
}

// TestUpdateFacialExpressionSettings 测试更新面部表情设置
func TestUpdateFacialExpressionSettings(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 更新设置
	settings := models.FacialExpressionSettings{
		DefaultExpression:     "warm_smile",
		ExpressionRange:       0.8,
		ExpressionIntensity:   0.7,
		EyeContactTendency:    0.8,
		SmileTendency:         0.6,
		BlinkRate:             18.0,
	}

	profile := engine.UpdateFacialExpressionSettings("identity_001", settings)
	if profile == nil {
		t.Fatal("Updated profile should not be nil")
	}

	if profile.FacialExpressionSettings.DefaultExpression != "warm_smile" {
		t.Errorf("DefaultExpression should be warm_smile, got %s", profile.FacialExpressionSettings.DefaultExpression)
	}
	if profile.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", profile.Version)
	}

	// 更新不存在的 Profile
	profile2 := engine.UpdateFacialExpressionSettings("identity_999", settings)
	if profile2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestUpdateBodyGestureSettings 测试更新身体手势设置
func TestUpdateBodyGestureSettings(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 更新设置
	settings := models.BodyGestureSettings{
		DefaultPosture:       "confident",
		GestureRange:         0.7,
		GestureIntensity:     0.6,
		GestureFrequency:     0.7,
		GestureSpeed:         "fast",
		HandGestureStyle:     "expressive",
		MirroringEnabled:     true,
		MirroringDelay:       300,
	}

	profile := engine.UpdateBodyGestureSettings("identity_001", settings)
	if profile == nil {
		t.Fatal("Updated profile should not be nil")
	}

	if profile.BodyGestureSettings.DefaultPosture != "confident" {
		t.Errorf("DefaultPosture should be confident, got %s", profile.BodyGestureSettings.DefaultPosture)
	}
	if profile.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", profile.Version)
	}
}

// TestApplyEmotionExpression 测试应用情绪表情
func TestApplyEmotionExpression(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 应用喜悦情绪
	profile := engine.ApplyEmotionExpression("identity_001", "joy", 0.7)
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证版本更新
	if profile.Version != 2 {
		t.Errorf("Version should be updated, got %d", profile.Version)
	}

	// 测试不存在的 Profile
	profile2 := engine.ApplyEmotionExpression("identity_999", "joy", 0.7)
	if profile2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestApplyRelationshipInfluence 测试关系影响
func TestApplyRelationshipInfluence(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 测试安全型依恋
	profile := engine.ApplyRelationshipInfluence("identity_001", "secure", "close_friend")
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证社交手势调整
	if profile.SocialGestureSettings.TouchComfortLevel != 0.8 {
		t.Errorf("TouchComfortLevel should be 0.8 for close_friend, got %f", profile.SocialGestureSettings.TouchComfortLevel)
	}

	// 测试焦虑型依恋
	profile2 := engine.ApplyRelationshipInfluence("identity_001", "anxious", "professional")
	if profile2.SocialGestureSettings.EyeContactWhileListening != 0.8 {
		t.Errorf("EyeContactWhileListening should be high for anxious, got %f", profile2.SocialGestureSettings.EyeContactWhileListening)
	}

	// 测试回避型依恋
	profile3 := engine.ApplyRelationshipInfluence("identity_001", "avoidant", "stranger")
	if profile3.SocialGestureSettings.TouchComfortLevel != 0.1 {
		t.Errorf("TouchComfortLevel should be low for avoidant stranger, got %f", profile3.SocialGestureSettings.TouchComfortLevel)
	}
}

// TestApplyLifeStageInfluence 测试人生阶段影响
func TestApplyLifeStageInfluence(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 测试青春期
	profile := engine.ApplyLifeStageInfluence("identity_001", "adolescence")
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	if profile.BodyGestureSettings.GestureFrequency != 0.7 {
		t.Errorf("GestureFrequency should be high for adolescence, got %f", profile.BodyGestureSettings.GestureFrequency)
	}
	if profile.BodyGestureSettings.GestureSpeed != "fast" {
		t.Errorf("GestureSpeed should be fast for adolescence, got %s", profile.BodyGestureSettings.GestureSpeed)
	}

	// 测试老年阶段
	profile2 := engine.ApplyLifeStageInfluence("identity_001", "elderly")
	if profile2.BodyGestureSettings.GestureFrequency != 0.4 {
		t.Errorf("GestureFrequency should be low for elderly, got %f", profile2.BodyGestureSettings.GestureFrequency)
	}
	if profile2.BodyGestureSettings.GestureSpeed != "slow" {
		t.Errorf("GestureSpeed should be slow for elderly, got %s", profile2.BodyGestureSettings.GestureSpeed)
	}
}

// TestGenerateExpression 测试生成表情
func TestGenerateExpression(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 生成喜悦表情
	expression := engine.GenerateExpression("identity_001", "joy", 0.7, "meeting")
	if expression == nil {
		t.Fatal("Expression should not be nil")
	}

	// 验证表情属性
	if expression.ExpressionName != "happy" {
		t.Errorf("ExpressionName should be happy for joy, got %s", expression.ExpressionName)
	}
	if expression.Intensity <= 0 {
		t.Errorf("Intensity should be positive, got %f", expression.Intensity)
	}
	if expression.Duration <= 0 {
		t.Errorf("Duration should be positive, got %d", expression.Duration)
	}

	// 验证情绪特征
	if expression.EyebrowState != "raised" {
		t.Errorf("EyebrowState should be raised for joy, got %s", expression.EyebrowState)
	}
	if expression.MouthState != "smile" {
		t.Errorf("MouthState should be smile for joy, got %s", expression.MouthState)
	}

	// 测试悲伤表情
	expression2 := engine.GenerateExpression("identity_001", "sadness", 0.5, "casual")
	if expression2.ExpressionName != "sad" {
		t.Errorf("ExpressionName should be sad, got %s", expression2.ExpressionName)
	}

	// 测试不存在的身份
	expression3 := engine.GenerateExpression("identity_999", "joy", 0.7, "meeting")
	if expression3 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestGenerateGesture 测试生成手势
func TestGenerateGesture(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 生成喜悦手势
	gesture := engine.GenerateGesture("identity_001", "joy", 0.6, "close")
	if gesture == nil {
		t.Fatal("Gesture should not be nil")
	}

	// 验证手势属性
	if gesture.GestureName == "" {
		t.Error("GestureName should not be empty")
	}
	if gesture.Intensity <= 0 {
		t.Errorf("Intensity should be positive, got %f", gesture.Intensity)
	}
	if gesture.Duration <= 0 {
		t.Errorf("Duration should be positive, got %d", gesture.Duration)
	}
	if gesture.Posture == "" {
		t.Error("Posture should not be empty")
	}

	// 测试镜像行为
	if !gesture.IsMirroring {
		t.Error("IsMirroring should be true for close social context")
	}

	// 测试不存在的身份
	gesture2 := engine.GenerateGesture("identity_999", "joy", 0.6, "close")
	if gesture2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 获取决策上下文
	ctx := engine.GetDecisionContext("identity_001", "joy", "meeting", "professional")
	if ctx == nil {
		t.Fatal("DecisionContext should not be nil")
	}

	// 验证上下文
	if ctx.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", ctx.IdentityID)
	}

	// 验证表情
	if ctx.CurrentExpression.ExpressionName == "" {
		t.Error("CurrentExpression should be set")
	}
	if ctx.RecommendedExpression.ExpressionName == "" {
		t.Error("RecommendedExpression should be set")
	}

	// 验证手势
	if ctx.CurrentGesture.GestureName == "" {
		t.Error("CurrentGesture should be set")
	}
	if ctx.RecommendedGesture.GestureName == "" {
		t.Error("RecommendedGesture should be set")
	}

	// 验证场景适应
	if ctx.SceneAdaptation.Scene != "meeting" {
		t.Errorf("SceneAdaptation.Scene should be meeting, got %s", ctx.SceneAdaptation.Scene)
	}

	// 验证情绪适应
	if ctx.EmotionAdaptation.CurrentEmotion != "joy" {
		t.Errorf("EmotionAdaptation.CurrentEmotion should be joy, got %s", ctx.EmotionAdaptation.CurrentEmotion)
	}

	// 验证社交适应
	if ctx.SocialAdaptation.SocialContext != "professional" {
		t.Errorf("SocialAdaptation.SocialContext should be professional, got %s", ctx.SocialAdaptation.SocialContext)
	}

	// 测试不存在的身份
	ctx2 := engine.GetDecisionContext("identity_999", "joy", "meeting", "professional")
	if ctx2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestGetDefaultEmotionMapping 测试获取默认情绪映射
func TestGetDefaultEmotionMapping(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	testEmotions := []string{"joy", "anger", "sadness", "fear", "love", "disgust", "desire"}

	for _, emotion := range testEmotions {
		mapping := engine.getDefaultEmotionMapping(emotion)

		if mapping.EmotionName != emotion {
			t.Errorf("EmotionName should be %s, got %s", emotion, mapping.EmotionName)
		}
		if mapping.ExpressionType == "" {
			t.Errorf("ExpressionType should not be empty for %s", emotion)
		}
		if mapping.Intensity <= 0 {
			t.Errorf("Intensity should be positive for %s, got %f", emotion, mapping.Intensity)
		}
	}

	// 测试未知情绪
	mapping := engine.getDefaultEmotionMapping("unknown")
	if mapping.ExpressionType != "neutral" {
		t.Errorf("Unknown emotion should have neutral expression, got %s", mapping.ExpressionType)
	}
}

// TestCalculateDuration 测试计算表情持续时间
func TestCalculateDuration(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	testCases := []struct {
		emotion   string
		expected  int
	}{
		{"joy", 3000},
		{"anger", 3000},
		{"fear", 3000},
		{"sadness", 5000},
		{"love", 5000},
		{"unknown", 2000},
	}

	for _, tc := range testCases {
		result := engine.calculateDuration(tc.emotion)
		if result != tc.expected {
			t.Errorf("Duration for %s should be %d, got %d", tc.emotion, tc.expected, result)
		}
	}
}

// TestGenerateSceneAdaptation 测试生成场景适应
func TestGenerateSceneAdaptation(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		scene         string
		expectedRange float64
	}{
		{"meeting", 0.4},
		{"casual", 0.8},
		{"public", 0.5},
	}

	for _, tc := range testCases {
		adaptation := engine.generateSceneAdaptation(tc.scene, profile)

		if adaptation.Scene != tc.scene {
			t.Errorf("Scene should be %s, got %s", tc.scene, adaptation.Scene)
		}
		if adaptation.ExpressionRange != tc.expectedRange {
			t.Errorf("ExpressionRange for %s should be %f, got %f", tc.scene, tc.expectedRange, adaptation.ExpressionRange)
		}
	}
}

// TestGenerateSocialAdaptation 测试生成社交适应
func TestGenerateSocialAdaptation(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		context           string
		expectedEyeContact float64
	}{
		{"close", 0.8},
		{"professional", 0.7},
		{"stranger", 0.5},
	}

	for _, tc := range testCases {
		adaptation := engine.generateSocialAdaptation(tc.context, profile)

		if adaptation.SocialContext != tc.context {
			t.Errorf("SocialContext should be %s, got %s", tc.context, adaptation.SocialContext)
		}
		if adaptation.EyeContactTendency != tc.expectedEyeContact {
			t.Errorf("EyeContactTendency for %s should be %f, got %f", tc.context, tc.expectedEyeContact, adaptation.EyeContactTendency)
		}
	}
}

// TestTimestampUpdates 测试时间戳更新
func TestTimestampUpdates(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())

	// 创建 Profile
	profile := engine.CreateProfile("identity_001")
	initialTime := profile.CreatedAt

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 更新设置
	settings := models.FacialExpressionSettings{
		DefaultExpression: "warm_smile",
	}
	profile = engine.UpdateFacialExpressionSettings("identity_001", settings)

	// 验证 CreatedAt 不变
	if profile.CreatedAt != initialTime {
		t.Error("CreatedAt should not change")
	}

	// 验证 UpdatedAt 变化
	if !profile.UpdatedAt.After(initialTime) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

// TestApplyAttachmentToGestures 测试依恋风格对手势的影响
func TestApplyAttachmentToGestures(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())
	baseSettings := models.SocialGestureSettings{}

	testCases := []struct {
		style          string
		expectedEyeContact float64
		expectedTouchLevel float64
	}{
		{"secure", 0.7, 0.7},
		{"anxious", 0.8, 0.5},
		{"avoidant", 0.4, 0.3},
		{"disorganized", 0.5, 0.4},
	}

	for _, tc := range testCases {
		result := engine.applyAttachmentToGestures(tc.style, baseSettings)

		if result.EyeContactWhileListening != tc.expectedEyeContact {
			t.Errorf("EyeContact for %s should be %f, got %f", tc.style, tc.expectedEyeContact, result.EyeContactWhileListening)
		}
		if result.TouchComfortLevel != tc.expectedTouchLevel {
			t.Errorf("TouchLevel for %s should be %f, got %f", tc.style, tc.expectedTouchLevel, result.TouchComfortLevel)
		}
	}
}

// TestApplyRelationshipTypeToGestures 测试关系类型对手势的影响
func TestApplyRelationshipTypeToGestures(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())
	baseSettings := models.SocialGestureSettings{}

	testCases := []struct {
		type_              string
		expectedGreeting   float64
		expectedTouchLevel float64
	}{
		{"close_friend", 0.8, 0.8},
		{"professional", 0.5, 0.3},
		{"stranger", 0.3, 0.1},
	}

	for _, tc := range testCases {
		result := engine.applyRelationshipTypeToGestures(tc.type_, baseSettings)

		if result.GreetingIntensity != tc.expectedGreeting {
			t.Errorf("GreetingIntensity for %s should be %f, got %f", tc.type_, tc.expectedGreeting, result.GreetingIntensity)
		}
		if result.TouchComfortLevel != tc.expectedTouchLevel {
			t.Errorf("TouchComfortLevel for %s should be %f, got %f", tc.type_, tc.expectedTouchLevel, result.TouchComfortLevel)
		}
	}
}

// TestApplyLifeStageToGestures 测试人生阶段对手势的影响
func TestApplyLifeStageToGestures(t *testing.T) {
	engine := NewExpressionGestureEngine(DefaultExpressionGestureEngineConfig())
	baseSettings := models.BodyGestureSettings{}

	testCases := []struct {
		stage            string
		expectedFrequency float64
		expectedSpeed     string
	}{
		{"childhood", 0.7, "fast"},
		{"adolescence", 0.7, "fast"},
		{"youth", 0.6, "moderate"},
		{"mature", 0.5, "moderate"},
		{"elderly", 0.4, "slow"},
	}

	for _, tc := range testCases {
		result := engine.applyLifeStageToGestures(tc.stage, baseSettings)

		if result.GestureFrequency != tc.expectedFrequency {
			t.Errorf("GestureFrequency for %s should be %f, got %f", tc.stage, tc.expectedFrequency, result.GestureFrequency)
		}
		if result.GestureSpeed != tc.expectedSpeed {
			t.Errorf("GestureSpeed for %s should be %s, got %s", tc.stage, tc.expectedSpeed, result.GestureSpeed)
		}
	}
}
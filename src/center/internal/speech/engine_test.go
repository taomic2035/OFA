package speech

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestSpeechContentEngineCreation 测试引擎创建
func TestSpeechContentEngineCreation(t *testing.T) {
	config := SpeechContentEngineConfig{
		DefaultTone:         "neutral",
		DefaultFormality:    "neutral",
		DefaultLanguageLevel: "moderate",
		EnableCache:         true,
		CacheTTL:            3600,
		MaxContentLength:    1000,
		MinContentLength:    10,
	}
	engine := NewSpeechContentEngine(config)

	if engine == nil {
		t.Fatal("SpeechContentEngine should not be nil")
	}
	if engine.config.DefaultTone != "neutral" {
		t.Errorf("DefaultTone should be neutral, got %s", engine.config.DefaultTone)
	}
}

// TestDefaultSpeechContentEngineConfig 测试默认配置
func TestDefaultSpeechContentEngineConfig(t *testing.T) {
	config := DefaultSpeechContentEngineConfig()

	if config.DefaultTone != "neutral" {
		t.Errorf("Default DefaultTone should be neutral, got %s", config.DefaultTone)
	}
	if config.EnableCache != true {
		t.Error("EnableCache should be true by default")
	}
	if config.PhilosophyInfluence != 0.7 {
		t.Errorf("PhilosophyInfluence should be 0.7, got %f", config.PhilosophyInfluence)
	}
	if config.EmotionInfluence != 0.8 {
		t.Errorf("EmotionInfluence should be 0.8, got %f", config.EmotionInfluence)
	}
}

// TestCreateProfile 测试创建 Profile
func TestCreateProfile(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	profile := engine.CreateProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证默认值
	if profile.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", profile.IdentityID)
	}
	if profile.ContentStyle.ToneStyle != "neutral" {
		t.Errorf("Default ToneStyle should be neutral, got %s", profile.ContentStyle.ToneStyle)
	}
	if profile.Version != 1 {
		t.Errorf("Version should be 1, got %d", profile.Version)
	}

	// 验证表达深度
	if profile.ExpressionDepth.ThinkingDepth != "moderate" {
		t.Errorf("Default ThinkingDepth should be moderate, got %s", profile.ExpressionDepth.ThinkingDepth)
	}

	// 验证文化表达
	if profile.CulturalExpression.HonorificUsage != "light" {
		t.Errorf("Default HonorificUsage should be light, got %s", profile.CulturalExpression.HonorificUsage)
	}
}

// TestGetProfile 测试获取 Profile
func TestGetProfile(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

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

// TestUpdateContentStyle 测试更新内容风格
func TestUpdateContentStyle(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 更新风格
	style := models.ContentStyle{
		ToneStyle:         "professional",
		ToneConsistency:   0.8,
		LanguageLevel:     "sophisticated",
		Directness:        0.7,
		EuphemismUsage:    0.2,
		MetaphorUsage:     0.4,
		HumorTendency:     0.2,
		EmotionalColoring: "neutral",
		EnthusiasmLevel:   0.5,
	}

	profile := engine.UpdateContentStyle("identity_001", style)
	if profile == nil {
		t.Fatal("Updated profile should not be nil")
	}

	if profile.ContentStyle.ToneStyle != "professional" {
		t.Errorf("ToneStyle should be professional, got %s", profile.ContentStyle.ToneStyle)
	}
	if profile.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", profile.Version)
	}

	// 更新不存在的 Profile
	profile2 := engine.UpdateContentStyle("identity_999", style)
	if profile2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestApplyPhilosophyInfluence 测试三观影响
func TestApplyPhilosophyInfluence(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 创建三观数据
	worldview := &models.Worldview{
		WorldEssence:     "deterministic",
		SocialCognition:  "cooperative",
	}

	lifeView := &models.LifeView{
		LifeMeaning:       "growth_contribution",
		TimeOrientation:   "future_oriented",
	}

	valueSystem := &models.EnhancedValueSystem{
		CoreValues: []models.CoreValue{
			{Name: "honesty", Priority: 0.8},
			{Name: "professionalism", Priority: 0.7},
		},
	}

	profile := engine.ApplyPhilosophyInfluence("identity_001", worldview, lifeView, valueSystem)
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证世界观影响
	if profile.ContentStyle.PersuasionStyle != "logical" {
		t.Errorf("PersuasionStyle should be logical for deterministic worldview, got %s", profile.ContentStyle.PersuasionStyle)
	}

	// 验证人生观影响
	if profile.ExpressionDepth.ThinkingDepth != "deep" {
		t.Errorf("ThinkingDepth should be deep for growth_contribution, got %s", profile.ExpressionDepth.ThinkingDepth)
	}

	// 验证价值观影响
	if profile.ContentStyle.Directness != 0.8 {
		t.Errorf("Directness should be high for honesty value, got %f", profile.ContentStyle.Directness)
	}
}

// TestApplyCulturalInfluence 测试文化影响
func TestApplyCulturalInfluence(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 创建文化数据
	culture := &models.RegionalCulture{
		CommunicationStyle: "indirect",
		HofstedeDimensions: models.HofstedeDimensions{
			Collectivism:        0.8,
			PowerDistance:       0.7,
			LongTermOrientation: 0.8,
		},
	}

	profile := engine.ApplyCulturalInfluence("identity_001", culture)
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证文化表达
	if profile.CulturalExpression.IndirectExpression < 0.5 {
		t.Errorf("IndirectExpression should be high for collectivist culture, got %f", profile.CulturalExpression.IndirectExpression)
	}

	if profile.CulturalExpression.FaceSaving < 0.5 {
		t.Errorf("FaceSaving should be significant for high power distance, got %f", profile.CulturalExpression.FaceSaving)
	}

	// 验证风格影响
	if profile.ContentStyle.Directness > 0.5 {
		t.Errorf("Directness should be low for indirect culture, got %f", profile.ContentStyle.Directness)
	}
}

// TestApplySocialIdentityInfluence 测试社会身份影响
func TestApplySocialIdentityInfluence(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 创建社会身份数据
	socialIdentity := &models.IdentityProfile{
		IdentityConfidence: 0.8,
	}

	career := &models.CareerProfile{
		Industry: "finance",
	}

	socialClass := &models.SocialClassProfile{
		EconomicCapital: models.EconomicCapital{
			Income: "high",
		},
	}

	profile := engine.ApplySocialIdentityInfluence("identity_001", socialIdentity, career, socialClass)
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证职业影响
	if profile.ContentStyle.ToneStyle != "professional" {
		t.Errorf("ToneStyle should be professional for finance, got %s", profile.ContentStyle.ToneStyle)
	}

	if profile.SocialExpression.ProfessionalTone != "authoritative" {
		t.Errorf("ProfessionalTone should be authoritative for finance, got %s", profile.SocialExpression.ProfessionalTone)
	}

	// 验证阶层表达
	if profile.SocialExpression.ClassExpression != "understated" {
		t.Errorf("ClassExpression should be understated for high income, got %s", profile.SocialExpression.ClassExpression)
	}
}

// TestApplyEmotionInfluence 测试情绪影响
func TestApplyEmotionInfluence(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	testEmotions := []struct {
		emotion           string
		intensity         float64
		expectedColoring  string
	}{
		{"joy", 0.7, "warm"},
		{"anger", 0.6, "passionate"},
		{"sadness", 0.5, "cool"},
		{"fear", 0.6, "cool"},
		{"love", 0.8, "warm"},
		{"disgust", 0.5, "cool"},
		{"desire", 0.6, "warm"},
	}

	for _, tc := range testEmotions {
		profile := engine.ApplyEmotionInfluence("identity_001", tc.emotion, tc.intensity)
		if profile == nil {
			t.Fatalf("Profile should not be nil for emotion %s", tc.emotion)
		}

		if profile.ContentStyle.EmotionalColoring != tc.expectedColoring {
			t.Errorf("EmotionalColoring for %s should be %s, got %s", tc.emotion, tc.expectedColoring, profile.ContentStyle.EmotionalColoring)
		}
	}
}

// TestGenerateContent 测试内容生成
func TestGenerateContent(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 创建请求
	req := &models.SpeechContentRequest{
		IdentityID:   "identity_001",
		ContentType:  "greeting",
		Formality:    "formal",
		EmotionContext: models.EmotionContext{
			CurrentEmotion: "joy",
			EmotionIntensity: 0.6,
		},
	}

	result := engine.GenerateContent(req)
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// 验证结果
	if result.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", result.IdentityID)
	}
	if result.ContentType != "greeting" {
		t.Errorf("ContentType should be greeting, got %s", result.ContentType)
	}
	if result.Content == "" {
		t.Error("Content should not be empty")
	}

	// 验证质量指标
	if result.ClarityScore < 0.5 {
		t.Errorf("ClarityScore should be reasonable, got %f", result.ClarityScore)
	}
	if result.Appropriateness < 0.5 {
		t.Errorf("Appropriateness should be reasonable, got %f", result.Appropriateness)
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	engine.CreateProfile("identity_001")

	// 获取决策上下文
	ctx := engine.GetDecisionContext("identity_001", "meeting", "professional", "joy")
	if ctx == nil {
		t.Fatal("DecisionContext should not be nil")
	}

	// 验证上下文
	if ctx.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", ctx.IdentityID)
	}

	// 验证推荐
	if ctx.RecommendedTone == "" {
		t.Error("RecommendedTone should not be empty")
	}
	if ctx.RecommendedFormality == "" {
		t.Error("RecommendedFormality should not be empty")
	}
	if ctx.RecommendedDepth == "" {
		t.Error("RecommendedDepth should not be empty")
	}

	// 验证场景适应
	if ctx.SceneAdaptation.Scene != "meeting" {
		t.Errorf("SceneAdaptation.Scene should be meeting, got %s", ctx.SceneAdaptation.Scene)
	}

	// 验证情绪适应
	if ctx.EmotionAdaptation.CurrentEmotion != "joy" {
		t.Errorf("EmotionAdaptation.CurrentEmotion should be joy, got %s", ctx.EmotionAdaptation.CurrentEmotion)
	}

	// 验证建议
	if ctx.OpeningSuggestion == "" {
		t.Error("OpeningSuggestion should not be empty")
	}
	if ctx.ClosingSuggestion == "" {
		t.Error("ClosingSuggestion should not be empty")
	}

	// 测试不存在的身份
	ctx2 := engine.GetDecisionContext("identity_999", "meeting", "professional", "joy")
	if ctx2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestRecommendTone 测试推荐语调
func TestRecommendTone(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		scene    string
		emotion  string
		expected string
	}{
		{"meeting", "", "professional"},
		{"presentation", "", "professional"},
		{"casual", "", "friendly"},
		{"", "joy", "enthusiastic"},
		{"", "sadness", "gentle"},
	}

	for _, tc := range testCases {
		result := engine.recommendTone(profile, tc.scene, tc.emotion)

		if result != tc.expected {
			t.Errorf("Tone for scene=%s emotion=%s should be %s, got %s", tc.scene, tc.emotion, tc.expected, result)
		}
	}
}

// TestRecommendFormality 测试推荐正式程度
func TestRecommendFormality(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		scene    string
		context  string
		expected string
	}{
		{"meeting", "", "formal"},
		{"presentation", "", "formal"},
		{"casual", "", "casual"},
		{"home", "", "casual"},
		{"", "professional", "formal"},
		{"", "intimate", "casual"},
		{"", "", "neutral"},
	}

	for _, tc := range testCases {
		result := engine.recommendFormality(profile, tc.scene, tc.context)

		if result != tc.expected {
			t.Errorf("Formality for scene=%s context=%s should be %s, got %s", tc.scene, tc.context, tc.expected, result)
		}
	}
}

// TestRecommendDepth 测试推荐表达深度
func TestRecommendDepth(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		context  string
		expected string
	}{
		{"intimate", "deep"},
		{"professional", "moderate"},
		{"public", "surface"},
	}

	for _, tc := range testCases {
		result := engine.recommendDepth(profile, tc.context)

		if result != tc.expected {
			t.Errorf("Depth for context=%s should be %s, got %s", tc.context, tc.expected, result)
		}
	}
}

// TestRecommendLength 测试推荐长度
func TestRecommendLength(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	testCases := []struct {
		scene    string
		expected string
	}{
		{"meeting", "medium"},
		{"presentation", "long"},
		{"casual", "short"},
	}

	for _, tc := range testCases {
		result := engine.recommendLength(tc.scene)

		if result != tc.expected {
			t.Errorf("Length for scene=%s should be %s, got %s", tc.scene, tc.expected, result)
		}
	}
}

// TestCalculateIndirectness 测试计算间接程度
func TestCalculateIndirectness(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	testCases := []struct {
		collectivism  float64
		commStyle     string
		expectedMin   float64
	}{
		{0.8, "indirect", 0.7},
		{0.3, "direct", 0.5},
		{0.6, "context_dependent", 0.5},
	}

	for _, tc := range testCases {
		culture := &models.RegionalCulture{
			CommunicationStyle: tc.commStyle,
			HofstedeDimensions: models.HofstedeDimensions{
				Collectivism: tc.collectivism,
			},
		}

		result := engine.calculateIndirectness(culture)
		if result < tc.expectedMin {
			t.Errorf("Indirectness for collectivism=%f style=%s should be >= %f, got %f", tc.collectivism, tc.commStyle, tc.expectedMin, result)
		}
	}
}

// TestGetDefaultGreeting 测试默认问候语
func TestGetDefaultGreeting(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	testCases := []struct {
		scene     string
		formality string
		expected  string
	}{
		{"meeting", "formal", "您好，很高兴见到您。"},
		{"meeting", "casual", "你好，很高兴见到你。"},
		{"casual", "", "嗨，你好！"},
		{"presentation", "", "各位好，感谢大家的到来。"},
	}

	for _, tc := range testCases {
		result := engine.getDefaultGreeting(tc.scene, tc.formality)

		if result != tc.expected {
			t.Errorf("Greeting for scene=%s formality=%s should be %s, got %s", tc.scene, tc.formality, tc.expected, result)
		}
	}
}

// TestGetDefaultClosing 测试默认结束语
func TestGetDefaultClosing(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	testCases := []struct {
		scene     string
		formality string
		expected  string
	}{
		{"meeting", "formal", "感谢您的时间，期待下次交流。"},
		{"meeting", "casual", "谢谢你，下次再聊。"},
		{"presentation", "", "感谢大家的聆听。"},
	}

	for _, tc := range testCases {
		result := engine.getDefaultClosing(tc.scene, tc.formality)

		if result != tc.expected {
			t.Errorf("Closing for scene=%s formality=%s should be %s, got %s", tc.scene, tc.formality, tc.expected, result)
		}
	}
}

// TestGenerateEmotionAdaptation 测试情绪适应生成
func TestGenerateEmotionAdaptation(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		emotion          string
		expectedColoring string
	}{
		{"joy", "warm"},
		{"sadness", "cool"},
		{"anger", "passionate"},
		{"unknown", "neutral"},
	}

	for _, tc := range testCases {
		result := engine.generateEmotionAdaptation(tc.emotion, profile)

		if result.EmotionalColoring != tc.expectedColoring {
			t.Errorf("EmotionalColoring for %s should be %s, got %s", tc.emotion, tc.expectedColoring, result.EmotionalColoring)
		}
	}
}

// TestGenerateSocialAdaptation 测试社交适应生成
func TestGenerateSocialAdaptation(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())
	profile := engine.CreateProfile("identity_001")

	testCases := []struct {
		context            string
		expectedHonorific  string
	}{
		{"professional", "moderate"},
		{"intimate", "light"},
		{"public", "light"},
	}

	for _, tc := range testCases {
		result := engine.generateSocialAdaptation(tc.context, profile)

		if result.HonorificUsage != tc.expectedHonorific {
			t.Errorf("HonorificUsage for %s should be %s, got %s", tc.context, tc.expectedHonorific, result.HonorificUsage)
		}
	}
}

// TestTimestampUpdates 测试时间戳更新
func TestTimestampUpdates(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建 Profile
	profile := engine.CreateProfile("identity_001")
	initialTime := profile.CreatedAt

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 更新风格
	style := models.ContentStyle{
		ToneStyle: "professional",
	}
	profile = engine.UpdateContentStyle("identity_001", style)

	// 验证 CreatedAt 不变
	if profile.CreatedAt != initialTime {
		t.Error("CreatedAt should not change")
	}

	// 验证 UpdatedAt 变化
	if !profile.UpdatedAt.After(initialTime) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

// TestApplyEmotionToStyleClamping 测试情绪影响值约束
func TestApplyEmotionToStyleClamping(t *testing.T) {
	engine := NewSpeechContentEngine(DefaultSpeechContentEngineConfig())

	// 创建基础风格
	style := models.ContentStyle{
		EnthusiasmLevel: 0.5,
		Directness:      0.5,
		HumorTendency:   0.5,
	}

	// 测试高强度情绪不会超过边界
	testEmotions := []struct {
		emotion   string
		intensity float64
	}{
		{"joy", 1.0},
		{"anger", 1.0},
		{"sadness", 1.0},
	}

	for _, tc := range testEmotions {
		result := engine.applyEmotionToStyle(tc.emotion, tc.intensity, style)

		// 验证值在合理范围内
		if result.EnthusiasmLevel < 0.0 || result.EnthusiasmLevel > 1.0 {
			t.Errorf("EnthusiasmLevel should be in [0,1] for %s, got %f", tc.emotion, result.EnthusiasmLevel)
		}
		if result.Directness < 0.0 || result.Directness > 1.0 {
			t.Errorf("Directness should be in [0,1] for %s, got %f", tc.emotion, result.Directness)
		}
	}
}
package avatar

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestAvatarEngineCreation 测试 Avatar 引擎创建
func TestAvatarEngineCreation(t *testing.T) {
	config := AvatarEngineConfig{
		DefaultFaceShape:   "oval",
		DefaultBodyType:    "average",
		DefaultStyle:       "casual",
		DefaultRenderQuality: "medium",
	}
	engine := NewAvatarEngine(config)

	if engine == nil {
		t.Fatal("AvatarEngine should not be nil")
	}
	if engine.config.DefaultFaceShape != "oval" {
		t.Errorf("DefaultFaceShape should be oval, got %s", engine.config.DefaultFaceShape)
	}
}

// TestDefaultAvatarEngineConfig 测试默认配置
func TestDefaultAvatarEngineConfig(t *testing.T) {
	config := DefaultAvatarEngineConfig()

	if config.DefaultFaceShape != "oval" {
		t.Errorf("Default DefaultFaceShape should be oval, got %s", config.DefaultFaceShape)
	}
	if config.DefaultBodyType != "average" {
		t.Errorf("Default DefaultBodyType should be average, got %s", config.DefaultBodyType)
	}
	if !config.EnableAgeProgression {
		t.Error("EnableAgeProgression should be true by default")
	}
}

// TestCreateAvatar 测试创建 Avatar
func TestCreateAvatar(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	avatar := engine.CreateAvatar("identity_001", nil, nil)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证默认值
	if avatar.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", avatar.IdentityID)
	}
	if avatar.FacialFeatures.FaceShape != "oval" {
		t.Errorf("Default FaceShape should be oval, got %s", avatar.FacialFeatures.FaceShape)
	}
	if avatar.BodyFeatures.BodyType != "average" {
		t.Errorf("Default BodyType should be average, got %s", avatar.BodyFeatures.BodyType)
	}
	if avatar.Version != 1 {
		t.Errorf("Version should be 1, got %d", avatar.Version)
	}

	// 验证年龄默认值
	if avatar.AgeAppearance.ApparentAge != 25 {
		t.Errorf("Default ApparentAge should be 25, got %d", avatar.AgeAppearance.ApparentAge)
	}
}

// TestCreateAvatarWithFeatures 测试带特征创建 Avatar
func TestCreateAvatarWithFeatures(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	facial := &models.FacialFeatures{
		FaceShape:     "round",
		EyeShape:      "round",
		EyeColor:      "blue",
		SkinTone:      "light",
		HairStyle:     "long",
		HairColor:     "brown",
		Expressiveness: 0.7,
	}

	body := &models.BodyFeatures{
		BodyType:        "fit",
		Height:          175,
		Weight:          70,
		Posture:         "confident",
		PostureScore:    0.7,
		MovementStyle:   "energetic",
		MovementSpeed:   "fast",
		GestureFrequency: 0.6,
	}

	avatar := engine.CreateAvatar("identity_002", facial, body)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证自定义特征
	if avatar.FacialFeatures.FaceShape != "round" {
		t.Errorf("FaceShape should be round, got %s", avatar.FacialFeatures.FaceShape)
	}
	if avatar.BodyFeatures.BodyType != "fit" {
		t.Errorf("BodyType should be fit, got %s", avatar.BodyFeatures.BodyType)
	}
}

// TestGetAvatar 测试获取 Avatar
func TestGetAvatar(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 获取存在的 Avatar
	avatar := engine.GetAvatar("identity_001")
	if avatar == nil {
		t.Fatal("Avatar should exist")
	}

	// 获取不存在的 Avatar
	avatar2 := engine.GetAvatar("identity_999")
	if avatar2 != nil {
		t.Error("Avatar should not exist for unknown identity")
	}
}

// TestUpdateAvatar 测试更新 Avatar
func TestUpdateAvatar(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 更新 Avatar
	updates := &models.Avatar{
		FacialFeatures: models.FacialFeatures{
			FaceShape: "square",
		},
		BodyFeatures: models.BodyFeatures{
			BodyType: "athletic",
		},
	}

	avatar := engine.UpdateAvatar("identity_001", updates)
	if avatar == nil {
		t.Fatal("Updated avatar should not be nil")
	}

	// 验证更新
	if avatar.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", avatar.Version)
	}

	// 更新不存在的 Avatar
	avatar2 := engine.UpdateAvatar("identity_999", updates)
	if avatar2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestUpdateFacialFeatures 测试更新面部特征
func TestUpdateFacialFeatures(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 更新面部特征
	features := models.FacialFeatures{
		FaceShape:     "heart",
		EyeShape:      "almond",
		EyeColor:      "green",
		SkinTone:      "medium",
		HairStyle:     "short",
		HairColor:     "black",
		Expressiveness: 0.8,
	}

	avatar := engine.UpdateFacialFeatures("identity_001", features)
	if avatar == nil {
		t.Fatal("Updated avatar should not be nil")
	}

	if avatar.FacialFeatures.FaceShape != "heart" {
		t.Errorf("FaceShape should be heart, got %s", avatar.FacialFeatures.FaceShape)
	}
	if avatar.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", avatar.Version)
	}
}

// TestUpdateBodyFeatures 测试更新身体特征
func TestUpdateBodyFeatures(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 更新身体特征
	features := models.BodyFeatures{
		BodyType:        "slim",
		Height:          180,
		Weight:          75,
		Posture:         "tall",
		PostureScore:    0.8,
		MovementStyle:   "graceful",
		GestureFrequency: 0.7,
	}

	avatar := engine.UpdateBodyFeatures("identity_001", features)
	if avatar == nil {
		t.Fatal("Updated avatar should not be nil")
	}

	if avatar.BodyFeatures.BodyType != "slim" {
		t.Errorf("BodyType should be slim, got %s", avatar.BodyFeatures.BodyType)
	}
	if avatar.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", avatar.Version)
	}
}

// TestUpdateStylePreferences 测试更新风格偏好
func TestUpdateStylePreferences(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 更新风格偏好
	preferences := models.StylePreferences{
		ClothingStyle:   "business",
		ClothingQuality: "premium",
		ClothingColors:  []string{"black", "gray"},
		AccessoryStyle:  "moderate",
		GroomingLevel:   "polished",
		OverallVibe:     "professional",
	}

	avatar := engine.UpdateStylePreferences("identity_001", preferences)
	if avatar == nil {
		t.Fatal("Updated avatar should not be nil")
	}

	if avatar.StylePreferences.ClothingStyle != "business" {
		t.Errorf("ClothingStyle should be business, got %s", avatar.StylePreferences.ClothingStyle)
	}
	if avatar.Version != 2 {
		t.Errorf("Version should be 2 after update, got %d", avatar.Version)
	}
}

// TestApplyLifeStageInfluence 测试人生阶段影响
func TestApplyLifeStageInfluence(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 应用人生阶段影响
	avatar := engine.ApplyLifeStageInfluence("identity_001", "youth", 25)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证年龄外观更新
	if avatar.AgeAppearance.ApparentAge != 25 {
		t.Errorf("ApparentAge should be 25, got %d", avatar.AgeAppearance.ApparentAge)
	}
	if avatar.AgeAppearance.AgeRange != "young_adult" {
		t.Errorf("AgeRange should be young_adult for youth, got %s", avatar.AgeAppearance.AgeRange)
	}

	// 测试中年阶段
	avatar2 := engine.ApplyLifeStageInfluence("identity_001", "mid_adult", 40)
	if avatar2.AgeAppearance.WrinkleLevel != "moderate" {
		t.Errorf("WrinkleLevel should be moderate for mid_adult, got %s", avatar2.AgeAppearance.WrinkleLevel)
	}

	// 测试老年阶段
	avatar3 := engine.ApplyLifeStageInfluence("identity_001", "elderly", 70)
	if avatar3.AgeAppearance.SkinElasticity != "low" {
		t.Errorf("SkinElasticity should be low for elderly, got %s", avatar3.AgeAppearance.SkinElasticity)
	}
}

// TestApplySocialIdentityInfluence 测试社会身份影响
func TestApplySocialIdentityInfluence(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 创建社会身份数据
	socialIdentity := &models.IdentityProfile{
		IdentityConfidence: 0.7,
	}

	career := &models.CareerProfile{
		Industry:    "technology",
		WorkMode:    "remote",
	}

	socialClass := &models.SocialClassProfile{
		EconomicCapital: models.EconomicCapital{
			Income: "middle",
		},
	}

	avatar := engine.ApplySocialIdentityInfluence("identity_001", socialIdentity, career, socialClass)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证版本更新
	if avatar.Version < 2 {
		t.Errorf("Version should be updated, got %d", avatar.Version)
	}
}

// TestApplyCulturalInfluence 测试文化影响
func TestApplyCulturalInfluence(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 创建文化数据
	culture := &models.RegionalCulture{
		CityLevel:       "first_tier",
		CommunicationStyle: "direct",
		HofstedeDimensions: models.HofstedeDimensions{
			Collectivism:        0.7,
			LongTermOrientation: 0.8,
		},
	}

	avatar := engine.ApplyCulturalInfluence("identity_001", culture)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证风格影响
	if avatar.StylePreferences.CulturalStyle != "modern" {
		t.Errorf("CulturalStyle should be modern for first_tier, got %s", avatar.StylePreferences.CulturalStyle)
	}
}

// TestApplyEmotionExpression 测试情绪表情
func TestApplyEmotionExpression(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 测试喜悦情绪
	avatar := engine.ApplyEmotionExpression("identity_001", "joy", 0.8)
	if avatar == nil {
		t.Fatal("Avatar should not be nil")
	}

	// 验证表情强度
	if avatar.FacialFeatures.Expressiveness < 0.5 {
		t.Errorf("Expressiveness should increase for joy, got %f", avatar.FacialFeatures.Expressiveness)
	}

	// 验证姿态变化
	if avatar.BodyFeatures.Posture != "confident" {
		t.Errorf("Posture should be confident for joy, got %s", avatar.BodyFeatures.Posture)
	}

	// 测试悲伤情绪
	avatar2 := engine.ApplyEmotionExpression("identity_001", "sadness", 0.6)
	if avatar2.BodyFeatures.Posture != "slouched" {
		t.Errorf("Posture should be slouched for sadness, got %s", avatar2.BodyFeatures.Posture)
	}
}

// TestGetDecisionContext 测试决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	engine.CreateAvatar("identity_001", nil, nil)

	// 获取决策上下文
	ctx := engine.GetDecisionContext("identity_001", "meeting", "professional", "formal_culture")
	if ctx == nil {
		t.Fatal("DecisionContext should not be nil")
	}

	// 验证上下文
	if ctx.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", ctx.IdentityID)
	}

	// 验证场景适应
	if ctx.SceneAdaptation.CurrentScene != "meeting" {
		t.Errorf("SceneAdaptation.CurrentScene should be meeting, got %s", ctx.SceneAdaptation.CurrentScene)
	}

	// 验证推荐
	if ctx.RecommendedStyle == "" {
		t.Error("RecommendedStyle should not be empty")
	}
	if ctx.RecommendedPosture == "" {
		t.Error("RecommendedPosture should not be empty")
	}
	if ctx.RecommendedExpression == "" {
		t.Error("RecommendedExpression should not be empty")
	}

	// 测试不存在身份
	ctx2 := engine.GetDecisionContext("identity_999", "meeting", "professional", "formal_culture")
	if ctx2 != nil {
		t.Error("Should return nil for unknown identity")
	}
}

// TestCreateProfile 测试创建 Avatar Profile
func TestCreateProfile(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	avatar := engine.CreateAvatar("identity_001", nil, nil)

	// 创建 Profile
	profile := engine.CreateProfile("identity_001", avatar)
	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// 验证 Profile
	if profile.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", profile.IdentityID)
	}
	if profile.VisualPersonality == "" {
		t.Error("VisualPersonality should not be empty")
	}
	if profile.DistinctivenessScore < 0 || profile.DistinctivenessScore > 1 {
		t.Errorf("DistinctivenessScore should be between 0 and 1, got %f", profile.DistinctivenessScore)
	}
	if profile.CharismaScore < 0 || profile.CharismaScore > 1 {
		t.Errorf("CharismaScore should be between 0 and 1, got %f", profile.CharismaScore)
	}
}

// TestGetProfile 测试获取 Profile
func TestGetProfile(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar 和 Profile
	avatar := engine.CreateAvatar("identity_001", nil, nil)
	engine.CreateProfile("identity_001", avatar)

	// 获取 Profile
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

// TestCalculateAgeAppearance 测试年龄外观计算
func TestCalculateAgeAppearance(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 测试各人生阶段
	testCases := []struct {
		lifeStage string
		age       int
		expected  models.AgeAppearance
	}{
		{"childhood", 10, models.AgeAppearance{AgeRange: "young", WrinkleLevel: "none"}},
		{"adolescence", 15, models.AgeAppearance{AgeRange: "young", FacialMaturity: 0.3}},
		{"youth", 25, models.AgeAppearance{AgeRange: "young_adult", AgingStage: "youthful"}},
		{"early_adult", 30, models.AgeAppearance{AgeRange: "young_adult", AgingStage: "prime"}},
		{"mid_adult", 40, models.AgeAppearance{AgeRange: "adult", WrinkleLevel: "moderate"}},
		{"mature", 55, models.AgeAppearance{AgeRange: "middle_aged", AgingStage: "mature"}},
		{"elderly", 75, models.AgeAppearance{AgeRange: "senior", SkinElasticity: "low"}},
	}

	for _, tc := range testCases {
		ap := engine.calculateAgeAppearance(tc.lifeStage, tc.age)

		if ap.ApparentAge != tc.age {
			t.Errorf("ApparentAge should be %d for %s, got %d", tc.age, tc.lifeStage, ap.ApparentAge)
		}
		if ap.AgeRange != tc.expected.AgeRange {
			t.Errorf("AgeRange should be %s for %s, got %s", tc.expected.AgeRange, tc.lifeStage, ap.AgeRange)
		}
	}
}

// TestCalculateMetabolism 测试新陈代谢计算
func TestCalculateMetabolism(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	testCases := []struct {
		age      int
		expected string
	}{
		{20, "fast"},
		{25, "moderate"},
		{30, "moderate"},
		{45, "slow"},
		{60, "slow"},
	}

	for _, tc := range testCases {
		result := engine.calculateMetabolism(tc.age)
		if result != tc.expected {
			t.Errorf("Metabolism for age %d should be %s, got %s", tc.age, tc.expected, result)
		}
	}
}

// TestCalculateExpressiveness 测试表情强度计算
func TestCalculateExpressiveness(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	testCases := []struct {
		emotion   string
		intensity float64
		minValue  float64
	}{
		{"joy", 0.8, 0.7},
		{"anger", 0.7, 0.7},
		{"sadness", 0.5, 0.3},
		{"fear", 0.6, 0.3},
		{"love", 0.7, 0.6},
		{"disgust", 0.5, 0.4},
		{"unknown", 0.5, 0.5},
	}

	for _, tc := range testCases {
		result := engine.calculateExpressiveness(tc.emotion, tc.intensity)
		if result < tc.minValue {
			t.Errorf("Expressiveness for %s should be >= %f, got %f", tc.emotion, tc.minValue, result)
		}
		if result > 1.0 {
			t.Errorf("Expressiveness should not exceed 1.0, got %f", result)
		}
	}
}

// TestTimestampUpdates 测试时间戳更新
func TestTimestampUpdates(t *testing.T) {
	engine := NewAvatarEngine(DefaultAvatarEngineConfig())

	// 创建 Avatar
	avatar := engine.CreateAvatar("identity_001", nil, nil)
	initialTime := avatar.CreatedAt

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 更新 Avatar
	updates := &models.Avatar{
		FacialFeatures: models.FacialFeatures{FaceShape: "round"},
	}
	avatar = engine.UpdateAvatar("identity_001", updates)

	// 验证 CreatedAt 不变
	if avatar.CreatedAt != initialTime {
		t.Error("CreatedAt should not change")
	}

	// 验证 UpdatedAt 变化
	if !avatar.UpdatedAt.After(initialTime) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}
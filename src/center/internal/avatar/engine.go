// Package avatar provides the Avatar Management Engine for v5.0.0.
//
// AvatarEngine manages the external appearance of the digital person,
// linking internal soul characteristics (v4.x) to external presentation.
package avatar

import (
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// AvatarEngine manages avatar state and provides decision context.
type AvatarEngine struct {
	mu sync.RWMutex

	// Avatar storage (identity_id -> *Avatar)
	avatars sync.Map

	// Avatar profiles (identity_id -> *AvatarProfile)
	profiles sync.Map

	// Configuration
	config AvatarEngineConfig
}

// AvatarEngineConfig holds configuration for the avatar engine.
type AvatarEngineConfig struct {
	// Default avatar settings
	DefaultFaceShape   string
	DefaultBodyType    string
	DefaultStyle       string
	DefaultRenderQuality string

	// Age progression settings
	EnableAgeProgression bool
	AgeProgressionRate   double // how fast avatar ages

	// Style evolution settings
	EnableStyleEvolution bool
	StyleChangeThreshold double // threshold for style change
}

// NewAvatarEngine creates a new Avatar Engine.
func NewAvatarEngine(config AvatarEngineConfig) *AvatarEngine {
	return &AvatarEngine{
		config: config,
	}
}

// DefaultAvatarEngineConfig returns default configuration.
func DefaultAvatarEngineConfig() AvatarEngineConfig {
	return AvatarEngineConfig{
		DefaultFaceShape:     "oval",
		DefaultBodyType:      "average",
		DefaultStyle:         "casual",
		DefaultRenderQuality: "medium",
		EnableAgeProgression: true,
		AgeProgressionRate:   1.0,
		EnableStyleEvolution: true,
		StyleChangeThreshold: 0.3,
	}
}

// === Avatar Management ===

// CreateAvatar creates a new avatar for an identity.
func (e *AvatarEngine) CreateAvatar(identityID string, baseFeatures *models.FacialFeatures, baseBody *models.BodyFeatures) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	avatar := &models.Avatar{
		IdentityID:     identityID,
		FacialFeatures: e.applyDefaultFacialFeatures(baseFeatures),
		BodyFeatures:   e.applyDefaultBodyFeatures(baseBody),
		AgeAppearance: models.AgeAppearance{
			ApparentAge:    25, // default young adult
			AgeRange:       "young_adult",
			AgingStage:     "youthful",
			FacialMaturity: 0.3,
			WrinkleLevel:   "none",
			SkinElasticity: "high",
			BodyMaturity:   0.3,
			MetabolismType: "moderate",
			SelfCareLevel:  0.5,
			GeneticFactor:  0.5,
		},
		StylePreferences: e.getDefaultStylePreferences(),
		Model3D: models.Model3DReference{
			ModelType:    "preset",
			SourceFormat: "glb",
			CustomizationLevel: "preset",
			AnimationEnabled: true,
			AnimationSet: []string{"idle", "wave", "nod", "shake_head"},
			RenderQuality: e.config.DefaultRenderQuality,
			TextureQuality: "medium",
		},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	e.avatars.Store(identityID, avatar)
	return avatar
}

// GetAvatar retrieves an avatar by identity ID.
func (e *AvatarEngine) GetAvatar(identityID string) *models.Avatar {
	if value, ok := e.avatars.Load(identityID); ok {
		return value.(*models.Avatar)
	}
	return nil
}

// UpdateAvatar updates an existing avatar.
func (e *AvatarEngine) UpdateAvatar(identityID string, updates *models.Avatar) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	// Apply updates
	if updates.FacialFeatures.FaceShape != "" {
		avatar.FacialFeatures = updates.FacialFeatures
	}
	if updates.BodyFeatures.BodyType != "" {
		avatar.BodyFeatures = updates.BodyFeatures
	}
	if updates.AgeAppearance.AgeRange != "" {
		avatar.AgeAppearance = updates.AgeAppearance
	}
	if updates.StylePreferences.ClothingStyle != "" {
		avatar.StylePreferences = updates.StylePreferences
	}
	if updates.Model3D.ModelID != "" {
		avatar.Model3D = updates.Model3D
	}

	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// UpdateFacialFeatures updates facial features.
func (e *AvatarEngine) UpdateFacialFeatures(identityID string, features models.FacialFeatures) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	avatar.FacialFeatures = features
	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// UpdateBodyFeatures updates body features.
func (e *AvatarEngine) UpdateBodyFeatures(identityID string, features models.BodyFeatures) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	avatar.BodyFeatures = features
	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// UpdateStylePreferences updates style preferences.
func (e *AvatarEngine) UpdateStylePreferences(identityID string, preferences models.StylePreferences) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	avatar.StylePreferences = preferences
	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// === Life Stage Integration (v4.4.0) ===

// ApplyLifeStageInfluence updates avatar based on life stage changes.
func (e *AvatarEngine) ApplyLifeStageInfluence(identityID string, lifeStage string, age int) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	// Update age appearance based on life stage
	avatar.AgeAppearance = e.calculateAgeAppearance(lifeStage, age)

	// Update style preferences based on life stage
	avatar.StylePreferences = e.applyLifeStageStyle(lifeStage, avatar.StylePreferences)

	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// calculateAgeAppearance calculates age appearance based on life stage.
func (e *AvatarEngine) calculateAgeAppearance(lifeStage string, age int) models.AgeAppearance {
	ap := models.AgeAppearance{
		ApparentAge: age,
	}

	// Set age range based on life stage
	switch lifeStage {
	case "childhood":
		ap.AgeRange = "young"
		ap.AgingStage = "youthful"
		ap.FacialMaturity = 0.1
		ap.WrinkleLevel = "none"
		ap.SkinElasticity = "high"
		ap.BodyMaturity = 0.1
	case "adolescence":
		ap.AgeRange = "young"
		ap.AgingStage = "youthful"
		ap.FacialMaturity = 0.3
		ap.WrinkleLevel = "none"
		ap.SkinElasticity = "high"
		ap.BodyMaturity = 0.3
	case "youth":
		ap.AgeRange = "young_adult"
		ap.AgingStage = "youthful"
		ap.FacialMaturity = 0.5
		ap.WrinkleLevel = "none"
		ap.SkinElasticity = "high"
		ap.BodyMaturity = 0.5
	case "early_adult":
		ap.AgeRange = "young_adult"
		ap.AgingStage = "prime"
		ap.FacialMaturity = 0.6
		ap.WrinkleLevel = "minimal"
		ap.SkinElasticity = "high"
		ap.BodyMaturity = 0.7
	case "mid_adult":
		ap.AgeRange = "adult"
		ap.AgingStage = "mature"
		ap.FacialMaturity = 0.75
		ap.WrinkleLevel = "moderate"
		ap.SkinElasticity = "moderate"
		ap.BodyMaturity = 0.8
	case "mature":
		ap.AgeRange = "middle_aged"
		ap.AgingStage = "mature"
		ap.FacialMaturity = 0.85
		ap.WrinkleLevel = "moderate"
		ap.SkinElasticity = "moderate"
		ap.BodyMaturity = 0.85
	case "elderly":
		ap.AgeRange = "senior"
		ap.AgingStage = "senior"
		ap.FacialMaturity = 0.95
		ap.WrinkleLevel = "significant"
		ap.SkinElasticity = "low"
		ap.BodyMaturity = 0.9
	default:
		ap.AgeRange = "young_adult"
		ap.AgingStage = "prime"
		ap.FacialMaturity = 0.5
		ap.WrinkleLevel = "minimal"
		ap.SkinElasticity = "high"
		ap.BodyMaturity = 0.5
	}

	ap.MetabolismType = e.calculateMetabolism(age)
	ap.SelfCareLevel = 0.5
	ap.GeneticFactor = 0.5

	return ap
}

// calculateMetabolism calculates metabolism type based on age.
func (e *AvatarEngine) calculateMetabolism(age int) string {
	if age < 25 {
		return "fast"
	} else if age < 45 {
		return "moderate"
	}
	return "slow"
}

// applyLifeStageStyle adjusts style preferences based on life stage.
func (e *AvatarEngine) applyLifeStageStyle(lifeStage string, currentStyle models.StylePreferences) models.StylePreferences {
	style := currentStyle

	switch lifeStage {
	case "childhood", "adolescence":
		style.ClothingStyle = "casual"
		style.OverallVibe = "youthful"
		style.GroomingLevel = "natural"
	case "youth", "early_adult":
		style.OverallVibe = "modern"
		style.AccessoryStyle = "moderate"
	case "mid_adult", "mature":
		style.ClothingStyle = "business"
		style.OverallVibe = "professional"
		style.GroomingLevel = "polished"
	case "elderly":
		style.ClothingStyle = "classic"
		style.OverallVibe = "distinguished"
		style.GroomingLevel = "polished"
	}

	return style
}

// === Social Identity Integration (v4.2.0) ===

// ApplySocialIdentityInfluence updates avatar based on social identity.
func (e *AvatarEngine) ApplySocialIdentityInfluence(identityID string, socialIdentity *models.IdentityProfile, career *models.CareerProfile, socialClass *models.SocialClassProfile) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	// Apply career influence on style
	avatar.StylePreferences = e.applyCareerStyle(career, avatar.StylePreferences)

	// Apply social class influence
	avatar.StylePreferences = e.applyClassStyle(socialClass, avatar.StylePreferences)

	// Update body features based on lifestyle
	avatar.BodyFeatures = e.applyLifestyleBody(career, avatar.BodyFeatures)

	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// applyCareerStyle adjusts style based on career.
func (e *AvatarEngine) applyCareerStyle(career *models.CareerProfile, currentStyle models.StylePreferences) models.StylePreferences {
	style := currentStyle

	if career == nil {
		return style
	}

	// Adjust style based on occupation and industry
	switch career.Industry {
	case "technology":
		style.ClothingStyle = "casual"
		style.OverallVibe = "modern"
		style.AccessoryStyle = "minimal"
	case "finance", "law":
		style.ClothingStyle = "business"
		style.OverallVibe = "professional"
		style.GroomingLevel = "polished"
		style.ClothingQuality = "premium"
	case "creative", "media", "art":
		style.ClothingStyle = "creative"
		style.OverallVibe = "artistic"
		style.AccessoryStyle = "bold"
	case "education", "healthcare":
		style.ClothingStyle = "business_casual"
		style.OverallVibe = "professional"
		style.AccessoryStyle = "minimal"
	case "retail", "service":
		style.ClothingStyle = "casual"
		style.OverallVibe = "approachable"
	}

	return style
}

// applyClassStyle adjusts style based on social class.
func (e *AvatarEngine) applyClassStyle(socialClass *models.SocialClassProfile, currentStyle models.StylePreferences) models.StylePreferences {
	style := currentStyle

	if socialClass == nil {
		return style
	}

	// Adjust quality and brand awareness based on economic capital
	if socialClass.EconomicCapital.Income == "high" || socialClass.EconomicCapital.Income == "very_high" {
		style.ClothingQuality = "luxury"
		style.BrandAwareness = 0.8
		style.StatusDisplay = 0.6
	} else if socialClass.EconomicCapital.Income == "middle" || socialClass.EconomicCapital.Income == "middle_high" {
		style.ClothingQuality = "premium"
		style.BrandAwareness = 0.5
		style.StatusDisplay = 0.4
	} else {
		style.ClothingQuality = "mid_range"
		style.BrandAwareness = 0.3
		style.StatusDisplay = 0.2
	}

	return style
}

// applyLifestyleBody adjusts body features based on lifestyle.
func (e *AvatarEngine) applyLifestyleBody(career *models.CareerProfile, currentBody models.BodyFeatures) models.BodyFeatures {
	body := currentBody

	if career == nil {
		return body
	}

	// Adjust posture and movement based on work mode
	switch career.WorkMode {
	case "office":
		body.Posture = "modest"
		body.PostureScore = 0.5
		body.MovementStyle = "calm"
	case "remote":
		body.Posture = "casual"
		body.PostureScore = 0.4
		body.MovementStyle = "relaxed"
	case "field", "outdoor":
		body.Posture = "confident"
		body.PostureScore = 0.7
		body.MovementStyle = "energetic"
	case "hybrid":
		body.Posture = "balanced"
		body.PostureScore = 0.6
		body.MovementStyle = "versatile"
	}

	return body
}

// === Regional Culture Integration (v4.3.0) ===

// ApplyCulturalInfluence updates avatar style based on cultural background.
func (e *AvatarEngine) ApplyCulturalInfluence(identityID string, culture *models.RegionalCulture) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	// Apply cultural style influence
	avatar.StylePreferences.CulturalStyle = e.determineCulturalStyle(culture)
	avatar.StylePreferences.RegionalInfluence = e.calculateRegionalInfluence(culture)

	// Adjust grooming based on cultural norms
	avatar.StylePreferences = e.applyCulturalGrooming(culture, avatar.StylePreferences)

	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// determineCulturalStyle determines cultural style from region.
func (e *AvatarEngine) determineCulturalStyle(culture *models.RegionalCulture) string {
	if culture == nil {
		return "modern"
	}

	// Check migration history
	if len(culture.MigrationHistory) > 0 {
		return "fusion"
	}

	// Check region type
	switch culture.CityLevel {
	case "first_tier", "new_first_tier":
		return "modern"
	case "second_tier":
		return "fusion"
	default:
		return "traditional"
	}
}

// calculateRegionalInfluence calculates how much region affects style.
func (e *AvatarEngine) calculateRegionalInfluence(culture *models.RegionalCulture) double {
	if culture == nil {
		return 0.5
	}

	// Higher influence for smaller cities
	switch culture.CityLevel {
	case "first_tier", "new_first_tier":
		return 0.3 // International influence higher
	case "second_tier":
		return 0.5
	case "third_tier":
		return 0.6
	default:
		return 0.7 // Traditional influence higher
	}
}

// applyCulturalGrooming applies cultural grooming preferences.
func (e *AvatarEngine) applyCulturalGrooming(culture *models.RegionalCulture, currentStyle models.StylePreferences) models.StylePreferences {
	style := currentStyle

	if culture == nil {
		return style
	}

	// Adjust based on Hofstede dimensions
	if culture.HofstedeDimensions.Collectivism > 0.7 {
		// High collectivism - more conservative grooming
		style.GroomingLevel = "polished"
		style.MakeupStyle = "minimal"
	}

	if culture.HofstedeDimensions.LongTermOrientation > 0.7 {
		// Long-term orientation - more conservative style
		style.AestheticTheme = "classic"
	}

	return style
}

// === Emotion Integration (v4.0.0, v4.5.0) ===

// ApplyEmotionExpression updates avatar expression based on emotion state.
func (e *AvatarEngine) ApplyEmotionExpression(identityID string, dominantEmotion string, intensity double) *models.Avatar {
	e.mu.Lock()
	defer e.mu.Unlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	// Adjust facial expressiveness based on emotion
	avatar.FacialFeatures.Expressiveness = e.calculateExpressiveness(dominantEmotion, intensity)

	// Update body movement style based on emotion
	avatar.BodyFeatures = e.applyEmotionBody(dominantEmotion, avatar.BodyFeatures)

	avatar.Version++
	avatar.UpdatedAt = time.Now()

	e.avatars.Store(identityID, avatar)
	return avatar
}

// calculateExpressiveness calculates facial expressiveness from emotion.
func (e *AvatarEngine) calculateExpressiveness(emotion string, intensity double) double {
	// Base expressiveness
	base := 0.5

	// Adjust based on emotion type
	switch emotion {
	case "joy", "anger":
		base = 0.7 + intensity*0.2
	case "sadness", "fear":
		base = 0.3 + intensity*0.1
	case "love":
		base = 0.6 + intensity*0.2
	case "disgust":
		base = 0.4 + intensity*0.1
	default:
		base = 0.5 + intensity*0.1
	}

	if base > 1.0 {
		base = 1.0
	}

	return base
}

// applyEmotionBody adjusts body features based on emotion.
func (e *AvatarEngine) applyEmotionBody(emotion string, currentBody models.BodyFeatures) models.BodyFeatures {
	body := currentBody

	switch emotion {
	case "joy":
		body.Posture = "confident"
		body.MovementStyle = "energetic"
		body.GestureFrequency = 0.7
	case "anger":
		body.Posture = "tense"
		body.MovementStyle = "rigid"
		body.GestureFrequency = 0.8
	case "sadness":
		body.Posture = "slouched"
		body.MovementStyle = "slow"
		body.GestureFrequency = 0.3
	case "fear":
		body.Posture = "defensive"
		body.MovementStyle = "cautious"
		body.GestureFrequency = 0.4
	case "love":
		body.Posture = "open"
		body.MovementStyle = "graceful"
		body.GestureFrequency = 0.6
	case "disgust":
		body.Posture = "withdrawn"
		body.MovementStyle = "avoidant"
		body.GestureFrequency = 0.3
	default:
		body.Posture = "neutral"
		body.MovementStyle = "calm"
		body.GestureFrequency = 0.5
	}

	return body
}

// === Decision Context ===

// GetDecisionContext generates avatar decision context for a given situation.
func (e *AvatarEngine) GetDecisionContext(identityID string, scene string, socialContext string, culturalContext string) *models.AvatarDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	avatar := e.GetAvatar(identityID)
	if avatar == nil {
		return nil
	}

	ctx := &models.AvatarDecisionContext{
		IdentityID: identityID,
		Timestamp:  time.Now(),
	}

	// Generate scene adaptation
	ctx.SceneAdaptation = e.generateSceneAdaptation(scene, avatar)

	// Generate social adaptation
	ctx.SocialAdaptation = e.generateSocialAdaptation(socialContext, avatar)

	// Generate cultural adaptation
	ctx.CulturalAdaptation = e.generateCulturalAdaptation(culturalContext, avatar)

	// Generate display settings
	ctx.DisplaySettings = e.generateDisplaySettings(scene, avatar)

	// Generate recommendations
	ctx.RecommendedStyle = e.recommendStyle(scene, socialContext, avatar)
	ctx.RecommendedPosture = e.recommendPosture(socialContext, avatar)
	ctx.RecommendedExpression = e.recommendExpression(avatar)

	return ctx
}

// generateSceneAdaptation generates scene-specific adaptation.
func (e *AvatarEngine) generateSceneAdaptation(scene string, avatar *models.Avatar) models.SceneAdaptation {
	sa := models.SceneAdaptation{
		CurrentScene: scene,
	}

	switch scene {
	case "meeting", "work":
		sa.StyleAdjustment = "formal_up"
		sa.PostureAdjustment = "formal"
		sa.ExpressionRange = "professional"
		sa.AnimationSet = "idle_meeting"
	case "casual", "home":
		sa.StyleAdjustment = "casual_down"
		sa.PostureAdjustment = "relaxed"
		sa.ExpressionRange = "warm"
		sa.AnimationSet = "idle_casual"
	case "formal", "ceremony":
		sa.StyleAdjustment = "formal_up"
		sa.PostureAdjustment = "formal"
		sa.ExpressionRange = "neutral"
		sa.AnimationSet = "idle_formal"
	case "sport", "exercise":
		sa.StyleAdjustment = "sporty"
		sa.PostureAdjustment = "active"
		sa.ExpressionRange = "expressive"
		sa.AnimationSet = "idle_active"
	default:
		sa.StyleAdjustment = "neutral"
		sa.PostureAdjustment = "balanced"
		sa.ExpressionRange = "neutral"
		sa.AnimationSet = "idle"
	}

	return sa
}

// generateSocialAdaptation generates social context adaptation.
func (e *AvatarEngine) generateSocialAdaptation(socialContext string, avatar *models.Avatar) models.SocialAdaptation {
	sa := models.SocialAdaptation{
		SocialContext: socialContext,
	}

	switch socialContext {
	case "formal", "professional":
		sa.DistanceLevel = "far"
		sa.EyeContactLevel = 0.6
		sa.GestureLevel = 0.4
		sa.TouchPermission = "handshake"
		sa.IntimacyDisplay = 0.2
		sa.TrustDisplay = 0.5
	case "casual", "friendly":
		sa.DistanceLevel = "moderate"
		sa.EyeContactLevel = 0.7
		sa.GestureLevel = 0.6
		sa.TouchPermission = "handshake"
		sa.IntimacyDisplay = 0.4
		sa.TrustDisplay = 0.6
	case "intimate", "close":
		sa.DistanceLevel = "close"
		sa.EyeContactLevel = 0.9
		sa.GestureLevel = 0.8
		sa.TouchPermission = "hug"
		sa.IntimacyDisplay = 0.8
		sa.TrustDisplay = 0.9
	default:
		sa.DistanceLevel = "moderate"
		sa.EyeContactLevel = 0.7
		sa.GestureLevel = 0.5
		sa.TouchPermission = "handshake"
		sa.IntimacyDisplay = 0.4
		sa.TrustDisplay = 0.5
	}

	return sa
}

// generateCulturalAdaptation generates cultural context adaptation.
func (e *AvatarEngine) generateCulturalAdaptation(culturalContext string, avatar *models.Avatar) models.CulturalAdaptation {
	ca := models.CulturalAdaptation{
		CulturalContext: culturalContext,
	}

	switch culturalContext {
	case "formal_culture":
		ca.FormalityLevel = 0.8
		ca.ModestyLevel = 0.7
		ca.ExpressivenessLevel = 0.3
		ca.GreetingStyle = "bow"
		ca.CommunicationStyle = "indirect"
	case "casual_culture":
		ca.FormalityLevel = 0.3
		ca.ModestyLevel = 0.4
		ca.ExpressivenessLevel = 0.7
		ca.GreetingStyle = "wave"
		ca.CommunicationStyle = "direct"
	case "international", "multicultural":
		ca.FormalityLevel = 0.5
		ca.ModestyLevel = 0.5
		ca.ExpressivenessLevel = 0.5
		ca.GreetingStyle = "handshake"
		ca.CommunicationStyle = "context_dependent"
	default:
		ca.FormalityLevel = 0.5
		ca.ModestyLevel = 0.5
		ca.ExpressivenessLevel = 0.5
		ca.GreetingStyle = "handshake"
		ca.CommunicationStyle = "context_dependent"
	}

	return ca
}

// generateDisplaySettings generates display settings for current context.
func (e *AvatarEngine) generateDisplaySettings(scene string, avatar *models.Avatar) models.DisplaySettings {
	ds := models.DisplaySettings{
		RenderMode:     "3d",
		CameraPosition: "portrait",
		CameraDistance: "medium",
		Background:     "neutral",
		AnimationState: "idle",
		Expression:     "neutral",
	}

	// Adjust based on scene
	switch scene {
	case "meeting":
		ds.CameraPosition = "front"
		ds.Background = "scene_specific"
		ds.Expression = "professional"
	case "casual":
		ds.Expression = "warm"
	case "formal":
		ds.CameraPosition = "front"
		ds.Expression = "composed"
	}

	return ds
}

// recommendStyle recommends clothing style for context.
func (e *AvatarEngine) recommendStyle(scene string, socialContext string, avatar *models.Avatar) string {
	// Start with current style
	style := avatar.StylePreferences.ClothingStyle

	// Adjust based on scene
	switch scene {
	case "meeting", "work":
		if socialContext == "formal" {
			style = "business"
		} else {
			style = "business_casual"
		}
	case "casual", "home":
		style = "casual"
	case "formal":
		style = "formal"
	case "sport":
		style = "sporty"
	}

	return style
}

// recommendPosture recommends posture for context.
func (e *AvatarEngine) recommendPosture(socialContext string, avatar *models.Avatar) string {
	switch socialContext {
	case "formal", "professional":
		return "confident"
	case "intimate":
		return "relaxed"
	default:
		return avatar.BodyFeatures.Posture
	}
}

// recommendExpression recommends facial expression.
func (e *AvatarEngine) recommendExpression(avatar *models.Avatar) string {
	// Based on expressiveness
	if avatar.FacialFeatures.Expressiveness > 0.7 {
		return "expressive"
	} else if avatar.FacialFeatures.Expressiveness > 0.4 {
		return "neutral"
	}
	return "reserved"
}

// === Avatar Profile Management ===

// CreateProfile creates an avatar profile for an identity.
func (e *AvatarEngine) CreateProfile(identityID string, avatar *models.Avatar) *models.AvatarProfile {
	profile := &models.AvatarProfile{
		IdentityID: identityID,

		VisualPersonality:   e.determineVisualPersonality(avatar),
		FirstImpression:     e.determineFirstImpression(avatar),
		MemorableFeature:    e.determineMemorableFeature(avatar),
		DistinctivenessScore: e.calculateDistinctiveness(avatar),

		SocialPresence:       e.determineSocialPresence(avatar),
		CharismaScore:        e.calculateCharisma(avatar),
		ApproachabilityScore: e.calculateApproachability(avatar),

		StyleConsistency:  0.8,
		AuthenticityScore: 0.8,

		EvolutionTendency: "stable",
		FashionAwareness:  avatar.StylePreferences.BrandAwareness,

		Version:   1,
		UpdatedAt: time.Now(),
	}

	e.profiles.Store(identityID, profile)
	return profile
}

// GetProfile retrieves an avatar profile.
func (e *AvatarEngine) GetProfile(identityID string) *models.AvatarProfile {
	if value, ok := e.profiles.Load(identityID); ok {
		return value.(*models.AvatarProfile)
	}
	return nil
}

// Helper methods

func (e *AvatarEngine) applyDefaultFacialFeatures(base *models.FacialFeatures) models.FacialFeatures {
	if base == nil {
		return models.FacialFeatures{
			FaceShape:     e.config.DefaultFaceShape,
			EyeShape:      "almond",
			EyeColor:      "brown",
			SkinTone:      "medium",
			HairStyle:     "medium",
			HairColor:     "black",
			Expressiveness: 0.5,
		}
	}
	return *base
}

func (e *AvatarEngine) applyDefaultBodyFeatures(base *models.BodyFeatures) models.BodyFeatures {
	if base == nil {
		return models.BodyFeatures{
			BodyType:        e.config.DefaultBodyType,
			Height:          170,
			Weight:          65,
			Posture:         "balanced",
			PostureScore:    0.5,
			MovementStyle:   "calm",
			MovementSpeed:   "moderate",
			GestureFrequency: 0.5,
		}
	}
	return *base
}

func (e *AvatarEngine) getDefaultStylePreferences() models.StylePreferences {
	return models.StylePreferences{
		ClothingStyle:    e.config.DefaultStyle,
		ClothingQuality:  "mid_range",
		ClothingColors:   []string{"black", "white", "blue"},
		AccessoryStyle:   "minimal",
		GroomingLevel:    "natural",
		OverallVibe:      "balanced",
		AestheticTheme:   "modern",
		CulturalStyle:    "modern",
		RegionalInfluence: 0.5,
		StatusDisplay:    0.3,
		BrandAwareness:   0.3,
	}
}

func (e *AvatarEngine) determineVisualPersonality(avatar *models.Avatar) string {
	if avatar.BodyFeatures.Posture == "confident" && avatar.FacialFeatures.Expressiveness > 0.6 {
		return "approachable"
	} else if avatar.StylePreferences.ClothingStyle == "business" {
		return "professional"
	} else if avatar.StylePreferences.OverallVibe == "artistic" {
		return "artistic"
	}
	return "balanced"
}

func (e *AvatarEngine) determineFirstImpression(avatar *models.Avatar) string {
	if avatar.FacialFeatures.Expressiveness > 0.7 {
		return "warm"
	} else if avatar.BodyFeatures.Posture == "formal" {
		return "professional"
	} else if avatar.FacialFeatures.Expressiveness < 0.3 {
		return "reserved"
	}
	return "neutral"
}

func (e *AvatarEngine) determineMemorableFeature(avatar *models.Avatar) string {
	if avatar.FacialFeatures.EyeShape == "distinctive" {
		return "eyes"
	} else if avatar.FacialFeatures.HairStyle == "unique" {
		return "hair"
	} else if avatar.StylePreferences.AccessoryStyle == "bold" {
		return "style"
	}
	return "presence"
}

func (e *AvatarEngine) calculateDistinctiveness(avatar *models.Avatar) double {
	score := 0.3 // base

	if avatar.StylePreferences.AccessoryStyle == "bold" {
		score += 0.2
	}
	if avatar.StylePreferences.AestheticTheme == "maximalist" {
		score += 0.2
	}
	if len(avatar.FacialFeatures.Moles) > 0 || len(avatar.FacialFeatures.Scars) > 0 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (e *AvatarEngine) determineSocialPresence(avatar *models.Avatar) string {
	if avatar.BodyFeatures.GestureFrequency > 0.7 {
		return "dominant"
	} else if avatar.BodyFeatures.GestureFrequency < 0.3 {
		return "submissive"
	}
	return "balanced"
}

func (e *AvatarEngine) calculateCharisma(avatar *models.Avatar) double {
	score := 0.5

	score += avatar.FacialFeatures.Expressiveness * 0.2
	score += avatar.BodyFeatures.PostureScore * 0.2

	if avatar.StylePreferences.OverallVibe == "elegant" || avatar.StylePreferences.OverallVibe == "professional" {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (e *AvatarEngine) calculateApproachability(avatar *models.Avatar) double {
	score := 0.5

	if avatar.FacialFeatures.Expressiveness > 0.5 {
		score += 0.2
	}
	if avatar.BodyFeatures.Posture == "open" || avatar.BodyFeatures.Posture == "relaxed" {
		score += 0.15
	}
	if avatar.StylePreferences.OverallVibe == "casual" {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}
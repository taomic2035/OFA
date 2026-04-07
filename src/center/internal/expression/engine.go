// Package expression provides the Expression & Gesture Engine for v5.3.0.
//
// ExpressionGestureEngine manages facial expressions and body gestures,
// integrating with emotion and relationship systems from v4.x.
package expression

import (
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// ExpressionGestureEngine manages expression and gesture state.
type ExpressionGestureEngine struct {
	mu sync.RWMutex

	// Profiles (identity_id -> *ExpressionGestureProfile)
	profiles sync.Map

	// Configuration
	config ExpressionGestureEngineConfig
}

// ExpressionGestureEngineConfig holds configuration.
type ExpressionGestureEngineConfig struct {
	// Default settings
	DefaultExpression     string
	DefaultGesture        string
	DefaultExpressionRange double
	DefaultGestureRange    double

	// Emotion integration
	EmotionInfluenceStrength double // 0-1

	// Animation settings
	DefaultAnimationFPS int
	TransitionSpeed     string
}

// NewExpressionGestureEngine creates a new engine.
func NewExpressionGestureEngine(config ExpressionGestureEngineConfig) *ExpressionGestureEngine {
	return &ExpressionGestureEngine{
		config: config,
	}
}

// DefaultExpressionGestureEngineConfig returns default configuration.
func DefaultExpressionGestureEngineConfig() ExpressionGestureEngineConfig {
	return ExpressionGestureEngineConfig{
		DefaultExpression:         "neutral",
		DefaultGesture:            "natural",
		DefaultExpressionRange:    0.7,
		DefaultGestureRange:       0.6,
		EmotionInfluenceStrength:  0.8,
		DefaultAnimationFPS:       30,
		TransitionSpeed:           "smooth",
	}
}

// === Profile Management ===

// CreateProfile creates a new expression/gesture profile.
func (e *ExpressionGestureEngine) CreateProfile(identityID string) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	profile := &models.ExpressionGestureProfile{
		IdentityID:               identityID,
		FacialExpressionSettings: e.getDefaultFacialExpressionSettings(),
		BodyGestureSettings:      e.getDefaultBodyGestureSettings(),
		EmotionExpressionMapping: e.getDefaultEmotionExpressionMapping(),
		SocialGestureSettings:    e.getDefaultSocialGestureSettings(),
		AnimationSettings:        e.getDefaultAnimationSettings(),
		Version:                  1,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	e.profiles.Store(identityID, profile)
	return profile
}

// GetProfile retrieves a profile.
func (e *ExpressionGestureEngine) GetProfile(identityID string) *models.ExpressionGestureProfile {
	if value, ok := e.profiles.Load(identityID); ok {
		return value.(*models.ExpressionGestureProfile)
	}
	return nil
}

// UpdateFacialExpressionSettings updates facial expression settings.
func (e *ExpressionGestureEngine) UpdateFacialExpressionSettings(identityID string, settings models.FacialExpressionSettings) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	profile.FacialExpressionSettings = settings
	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// UpdateBodyGestureSettings updates body gesture settings.
func (e *ExpressionGestureEngine) UpdateBodyGestureSettings(identityID string, settings models.BodyGestureSettings) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	profile.BodyGestureSettings = settings
	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// === Emotion Integration (v4.0.0, v4.5.0) ===

// ApplyEmotionExpression applies emotion to expression/gesture.
func (e *ExpressionGestureEngine) ApplyEmotionExpression(identityID string, emotion string, intensity double) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Get emotion mapping
	mapping, exists := profile.EmotionExpressionMapping.EmotionMappings[emotion]
	if !exists {
		mapping = e.getDefaultEmotionMapping(emotion)
	}

	// Apply to facial expression settings
	profile.FacialExpressionSettings = e.applyEmotionToFacialSettings(
		profile.FacialExpressionSettings,
		mapping,
		intensity,
	)

	// Apply to body gesture settings
	profile.BodyGestureSettings = e.applyEmotionToGestureSettings(
		profile.BodyGestureSettings,
		mapping,
		intensity,
	)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyEmotionToFacialSettings applies emotion to facial settings.
func (e *ExpressionGestureEngine) applyEmotionToFacialSettings(settings models.FacialExpressionSettings, mapping models.ExpressionMapping, intensity double) models.ExpressionGestureProfile {
	// This is a placeholder - in actual implementation would modify settings
	return settings
}

// applyEmotionToGestureSettings applies emotion to gesture settings.
func (e *ExpressionGestureEngine) applyEmotionToGestureSettings(settings models.BodyGestureSettings, mapping models.ExpressionMapping, intensity double) models.BodyGestureSettings {
	// This is a placeholder - in actual implementation would modify settings
	return settings
}

// getDefaultEmotionMapping returns default emotion mapping.
func (e *ExpressionGestureEngine) getDefaultEmotionMapping(emotion string) models.ExpressionMapping {
	mappings := map[string]models.ExpressionMapping{
		"joy": {
			EmotionName:     "joy",
			EyebrowPosition: "raised",
			EyeShape:        "normal",
			MouthShape:      "smile",
			ExpressionType:  "happy",
			Intensity:       0.7,
			Posture:         "open",
			HandGesture:     "expressive",
			BodyMovement:    "energetic",
		},
		"anger": {
			EmotionName:     "anger",
			EyebrowPosition: "furrowed",
			EyeShape:        "narrowed",
			MouthShape:      "tight",
			ExpressionType:  "angry",
			Intensity:       0.8,
			Posture:         "tense",
			HandGesture:     "clenched",
			BodyMovement:    "rigid",
		},
		"sadness": {
			EmotionName:     "sadness",
			EyebrowPosition: "neutral",
			EyeShape:        "downcast",
			MouthShape:      "frown",
			ExpressionType:  "sad",
			Intensity:       0.6,
			Posture:         "slumped",
			HandGesture:     "minimal",
			BodyMovement:    "slow",
		},
		"fear": {
			EmotionName:     "fear",
			EyebrowPosition: "raised",
			EyeShape:        "wide",
			MouthShape:      "neutral",
			ExpressionType:  "fearful",
			Intensity:       0.7,
			Posture:         "defensive",
			HandGesture:     "protective",
			BodyMovement:    "cautious",
		},
		"love": {
			EmotionName:     "love",
			EyebrowPosition: "soft",
			EyeShape:        "warm",
			MouthShape:      "gentle_smile",
			ExpressionType:  "loving",
			Intensity:       0.6,
			Posture:         "open",
			HandGesture:     "gentle",
			BodyMovement:    "graceful",
		},
		"disgust": {
			EmotionName:     "disgust",
			EyebrowPosition: "furrowed",
			EyeShape:        "narrowed",
			MouthShape:      "grimace",
			ExpressionType:  "disgusted",
			Intensity:       0.5,
			Posture:         "withdrawn",
			HandGesture:     "minimal",
			BodyMovement:    "avoidant",
		},
		"desire": {
			EmotionName:     "desire",
			EyebrowPosition: "raised",
			EyeShape:        "focused",
			MouthShape:      "slight_smile",
			ExpressionType:  "desiring",
			Intensity:       0.6,
			Posture:         "leaning_forward",
			HandGesture:     "reaching",
			BodyMovement:    "attentive",
		},
	}

	if mapping, exists := mappings[emotion]; exists {
		return mapping
	}

	return models.ExpressionMapping{
		EmotionName:    emotion,
		ExpressionType: "neutral",
		Intensity:      0.5,
		Posture:        "neutral",
		HandGesture:    "natural",
		BodyMovement:   "normal",
	}
}

// === Relationship Integration (v4.6.0) ===

// ApplyRelationshipInfluence applies relationship influence.
func (e *ExpressionGestureEngine) ApplyRelationshipInfluence(identityID string, attachmentStyle string, relationshipType string) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Adjust social gesture settings based on attachment style
	profile.SocialGestureSettings = e.applyAttachmentToGestures(attachmentStyle, profile.SocialGestureSettings)

	// Adjust based on relationship type
	profile.SocialGestureSettings = e.applyRelationshipTypeToGestures(relationshipType, profile.SocialGestureSettings)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyAttachmentToGestures applies attachment style to gestures.
func (e *ExpressionGestureEngine) applyAttachmentToGestures(attachmentStyle string, settings models.SocialGestureSettings) models.SocialGestureSettings {
	result := settings

	switch attachmentStyle {
	case "secure":
		result.EyeContactWhileListening = 0.7
		result.TouchComfortLevel = 0.7
		result.MirrorCloseFriends = true
		result.PreferredDistance = "medium"
	case "anxious":
		result.EyeContactWhileListening = 0.8
		result.TouchComfortLevel = 0.5
		result.MirrorCloseFriends = true
		result.MirrorStrangers = true
		result.PreferredDistance = "close"
	case "avoidant":
		result.EyeContactWhileListening = 0.4
		result.TouchComfortLevel = 0.3
		result.MirrorCloseFriends = false
		result.PreferredDistance = "far"
	case "disorganized":
		result.EyeContactWhileListening = 0.5
		result.TouchComfortLevel = 0.4
		result.PreferredDistance = "medium"
	}

	return result
}

// applyRelationshipTypeToGestures applies relationship type to gestures.
func (e *ExpressionGestureEngine) applyRelationshipTypeToGestures(relationshipType string, settings models.SocialGestureSettings) models.SocialGestureSettings {
	result := settings

	switch relationshipType {
	case "close_friend":
		result.GreetingIntensity = 0.8
		result.TouchTypes = []string{"handshake", "pat", "hug"}
		result.TouchComfortLevel = 0.8
	case "professional":
		result.GreetingIntensity = 0.5
		result.TouchTypes = []string{"handshake"}
		result.TouchComfortLevel = 0.3
		result.EyeContactWhileListening = 0.7
	case "stranger":
		result.GreetingIntensity = 0.3
		result.TouchTypes = []string{}
		result.TouchComfortLevel = 0.1
		result.EyeContactWhileListening = 0.5
	}

	return result
}

// === Life Stage Integration (v4.4.0) ===

// ApplyLifeStageInfluence applies life stage influence.
func (e *ExpressionGestureEngine) ApplyLifeStageInfluence(identityID string, lifeStage string) *models.ExpressionGestureProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Adjust gesture settings based on life stage
	profile.BodyGestureSettings = e.applyLifeStageToGestures(lifeStage, profile.BodyGestureSettings)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyLifeStageToGestures applies life stage to gestures.
func (e *ExpressionGestureEngine) applyLifeStageToGestures(lifeStage string, settings models.BodyGestureSettings) models.BodyGestureSettings {
	result := settings

	switch lifeStage {
	case "childhood", "adolescence":
		result.GestureRange = 0.8
		result.GestureFrequency = 0.7
		result.FidgetLevel = 0.5
		result.GestureSpeed = "fast"
	case "youth", "early_adult":
		result.GestureRange = 0.7
		result.GestureFrequency = 0.6
		result.FidgetLevel = 0.3
		result.GestureSpeed = "moderate"
	case "mid_adult", "mature":
		result.GestureRange = 0.5
		result.GestureFrequency = 0.5
		result.FidgetLevel = 0.2
		result.GestureSpeed = "moderate"
	case "elderly":
		result.GestureRange = 0.4
		result.GestureFrequency = 0.4
		result.FidgetLevel = 0.1
		result.GestureSpeed = "slow"
	}

	return result
}

// === Expression/Gesture Generation ===

// GenerateExpression generates expression for context.
func (e *ExpressionGestureEngine) GenerateExpression(identityID string, emotion string, intensity double, scene string) *models.ExpressionState {
	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Get emotion mapping
	mapping, exists := profile.EmotionExpressionMapping.EmotionMappings[emotion]
	if !exists {
		mapping = e.getDefaultEmotionMapping(emotion)
	}

	// Calculate final intensity
	finalIntensity := mapping.Intensity * intensity * e.config.EmotionInfluenceStrength
	if finalIntensity > 1.0 {
		finalIntensity = 1.0
	}

	// Adjust for scene
	finalIntensity *= profile.FacialExpressionSettings.ExpressionRange

	return &models.ExpressionState{
		ExpressionName:  mapping.ExpressionType,
		Intensity:       finalIntensity,
		Duration:        e.calculateDuration(emotion),
		Transition:      profile.EmotionExpressionMapping.TransitionSpeed,
		EyebrowState:    mapping.EyebrowPosition,
		EyeState:        mapping.EyeShape,
		MouthState:      mapping.MouthShape,
		BlendWithPrevious: profile.EmotionExpressionMapping.BlendEnabled,
	}
}

// GenerateGesture generates gesture for context.
func (e *ExpressionGestureEngine) GenerateGesture(identityID string, emotion string, intensity double, socialContext string) *models.GestureState {
	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Get emotion mapping
	mapping, exists := profile.EmotionExpressionMapping.EmotionMappings[emotion]
	if !exists {
		mapping = e.getDefaultEmotionMapping(emotion)
	}

	// Calculate final intensity
	finalIntensity := 0.5 * intensity
	if finalIntensity > 1.0 {
		finalIntensity = 1.0
	}

	// Check mirroring
	mirroring := false
	if socialContext == "close" && profile.SocialGestureSettings.MirrorCloseFriends {
		mirroring = true
	} else if socialContext == "professional" && profile.SocialGestureSettings.MirrorProfessional {
		mirroring = true
	}

	return &models.GestureState{
		GestureName:   mapping.HandGesture,
		Intensity:     finalIntensity,
		Duration:      2000, // 2 seconds
		Transition:    "smooth",
		Posture:       mapping.Posture,
		HandPosition:  mapping.HandGesture,
		HeadPosition:  "neutral",
		IsMirroring:   mirroring,
		MirroringDelay: profile.BodyGestureSettings.MirroringDelay,
	}
}

// calculateDuration calculates expression duration.
func (e *ExpressionGestureEngine) calculateDuration(emotion string) int {
	switch emotion {
	case "joy", "anger", "fear":
		return 3000 // 3 seconds
	case "sadness", "love":
		return 5000 // 5 seconds
	default:
		return 2000 // 2 seconds
	}
}

// === Decision Context ===

// GetDecisionContext generates expression/gesture decision context.
func (e *ExpressionGestureEngine) GetDecisionContext(identityID string, emotion string, scene string, socialContext string) *models.ExpressionGestureContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	ctx := &models.ExpressionGestureContext{
		IdentityID: identityID,
		Timestamp:  time.Now(),
	}

	// Generate current expression
	ctx.CurrentExpression = *e.GenerateExpression(identityID, emotion, 0.5, scene)
	ctx.RecommendedExpression = *e.GenerateExpression(identityID, emotion, 0.7, scene)

	// Generate current gesture
	ctx.CurrentGesture = *e.GenerateGesture(identityID, emotion, 0.5, socialContext)
	ctx.RecommendedGesture = *e.GenerateGesture(identityID, emotion, 0.7, socialContext)

	// Generate adaptations
	ctx.SceneAdaptation = e.generateSceneAdaptation(scene, profile)
	ctx.EmotionAdaptation = e.generateEmotionAdaptation(emotion, profile)
	ctx.SocialAdaptation = e.generateSocialAdaptation(socialContext, profile)

	// Animation state
	ctx.AnimationState = models.AnimationState{
		CurrentAnimation: "idle",
		AnimationProgress: 0.0,
		AnimationFPS:      profile.AnimationSettings.ExpressionAnimationFPS,
		IsLooping:         true,
		IsBlending:        false,
	}

	return ctx
}

// generateSceneAdaptation generates scene adaptation.
func (e *ExpressionGestureEngine) generateSceneAdaptation(scene string, profile *models.ExpressionGestureProfile) models.ExpressionSceneAdaptation {
	adaptation := models.ExpressionSceneAdaptation{
		Scene: scene,
	}

	switch scene {
	case "meeting", "presentation":
		adaptation.ExpressionRange = 0.4
		adaptation.GestureRange = 0.3
		adaptation.FormalityLevel = 0.8
		adaptation.EyeContactLevel = 0.7
		adaptation.IdleAnimationStyle = "subtle"
	case "casual", "home":
		adaptation.ExpressionRange = 0.8
		adaptation.GestureRange = 0.7
		adaptation.FormalityLevel = 0.3
		adaptation.EyeContactLevel = 0.6
		adaptation.IdleAnimationStyle = "moderate"
	case "public":
		adaptation.ExpressionRange = 0.5
		adaptation.GestureRange = 0.4
		adaptation.FormalityLevel = 0.6
		adaptation.EyeContactLevel = 0.5
		adaptation.IdleAnimationStyle = "moderate"
	default:
		adaptation.ExpressionRange = profile.FacialExpressionSettings.ExpressionRange
		adaptation.GestureRange = profile.BodyGestureSettings.GestureRange
		adaptation.FormalityLevel = 0.5
		adaptation.EyeContactLevel = 0.6
		adaptation.IdleAnimationStyle = "moderate"
	}

	return adaptation
}

// generateEmotionAdaptation generates emotion adaptation.
func (e *ExpressionGestureEngine) generateEmotionAdaptation(emotion string, profile *models.ExpressionGestureProfile) models.ExpressionEmotionAdaptation {
	mapping := e.getDefaultEmotionMapping(emotion)

	return models.ExpressionEmotionAdaptation{
		CurrentEmotion:      emotion,
		TargetExpression:    mapping.ExpressionType,
		TargetGesture:       mapping.HandGesture,
		TransitionSpeed:     profile.EmotionExpressionMapping.TransitionSpeed,
		IntensityMultiplier: mapping.Intensity,
	}
}

// generateSocialAdaptation generates social adaptation.
func (e *ExpressionGestureEngine) generateSocialAdaptation(socialContext string, profile *models.ExpressionGestureProfile) models.ExpressionSocialAdaptation {
	adaptation := models.ExpressionSocialAdaptation{
		SocialContext: socialContext,
	}

	switch socialContext {
	case "close", "intimate":
		adaptation.GreetingGesture = "wave"
		adaptation.FarewellGesture = "wave"
		adaptation.EyeContactTendency = 0.8
		adaptation.MirroringEnabled = true
		adaptation.TouchPermission = "moderate"
	case "professional":
		adaptation.GreetingGesture = "nod"
		adaptation.FarewellGesture = "nod"
		adaptation.EyeContactTendency = 0.7
		adaptation.MirroringEnabled = profile.SocialGestureSettings.MirrorProfessional
		adaptation.TouchPermission = "light"
	case "stranger":
		adaptation.GreetingGesture = "nod"
		adaptation.FarewellGesture = "nod"
		adaptation.EyeContactTendency = 0.5
		adaptation.MirroringEnabled = false
		adaptation.TouchPermission = "none"
	default:
		adaptation.GreetingGesture = "nod"
		adaptation.FarewellGesture = "nod"
		adaptation.EyeContactTendency = 0.6
		adaptation.MirroringEnabled = profile.BodyGestureSettings.MirroringEnabled
		adaptation.TouchPermission = "light"
	}

	return adaptation
}

// === Helper methods ===

func (e *ExpressionGestureEngine) getDefaultFacialExpressionSettings() models.FacialExpressionSettings {
	return models.FacialExpressionSettings{
		DefaultExpression:         e.config.DefaultExpression,
		ExpressionRange:           e.config.DefaultExpressionRange,
		ExpressionIntensity:       0.5,
		ExpressionFrequency:       0.5,
		ExpressionDuration:        "medium",
		EyeExpressionEnabled:      true,
		EyeContactTendency:        0.6,
		BlinkRate:                 15.0,
		EyebrowExpressiveness:     0.5,
		MouthExpressionEnabled:    true,
		SmileTendency:             0.5,
		SmileType:                 "moderate",
		MicroExpressionEnabled:    true,
		MicroExpressionSensitivity: 0.5,
		MicroExpressionDuration:   200,
		SymmetryLevel:             0.9,
		ExpressionMasking:         0.2,
		PokerFaceAbility:          0.3,
	}
}

func (e *ExpressionGestureEngine) getDefaultBodyGestureSettings() models.BodyGestureSettings {
	return models.BodyGestureSettings{
		DefaultPosture:          "neutral",
		GestureRange:            e.config.DefaultGestureRange,
		GestureIntensity:        0.5,
		GestureFrequency:        0.5,
		GestureSpeed:            "moderate",
		HandGestureEnabled:      true,
		HandGestureStyle:        "moderate",
		HandPosition:            "natural",
		HandGestureVocabulary:   []string{"point", "wave", "nod", "shrug"},
		HeadMovementEnabled:     true,
		NodFrequency:            0.5,
		HeadTiltTendency:        0.3,
		HeadShakeFrequency:      0.3,
		BodyLeanEnabled:         true,
		BodyLeanDirection:       "neutral",
		BodyLeanTendency:        0.3,
		ShrugTendency:           0.3,
		ShoulderTension:         "relaxed",
		FidgetLevel:             0.2,
		FidgetType:              "subtle",
		MirroringEnabled:        true,
		MirroringDelay:          500,
		MirroringIntensity:      0.5,
	}
}

func (e *ExpressionGestureEngine) getDefaultEmotionExpressionMapping() models.EmotionExpressionMapping {
	return models.EmotionExpressionMapping{
		EmotionMappings: map[string]models.ExpressionMapping{
			"joy":     e.getDefaultEmotionMapping("joy"),
			"anger":   e.getDefaultEmotionMapping("anger"),
			"sadness": e.getDefaultEmotionMapping("sadness"),
			"fear":    e.getDefaultEmotionMapping("fear"),
			"love":    e.getDefaultEmotionMapping("love"),
			"disgust": e.getDefaultEmotionMapping("disgust"),
			"desire":  e.getDefaultEmotionMapping("desire"),
		},
		TransitionSpeed:   e.config.TransitionSpeed,
		TransitionDuration: 500,
		BlendEnabled:      true,
		BlendDuration:     300,
	}
}

func (e *ExpressionGestureEngine) getDefaultSocialGestureSettings() models.SocialGestureSettings {
	return models.SocialGestureSettings{
		GreetingGesture:            "nod",
		GreetingIntensity:          0.5,
		PartingGesture:             "nod",
		PartingIntensity:           0.5,
		ListeningGestureEnabled:    true,
		NodWhileListening:          0.5,
		EyeContactWhileListening:   0.6,
		SpeakingGestureEnabled:     true,
		GestureWhileSpeaking:       0.5,
		PauseGesture:              "neutral",
		AgreementGesture:          "nod",
		DisagreementGesture:       "head_shake",
		UncertaintyGesture:        "shrug",
		TouchComfortLevel:         0.5,
		TouchTypes:                []string{"handshake"},
		PreferredDistance:         "medium",
		DistanceAdjustment:        0.5,
		MirrorCloseFriends:        true,
		MirrorProfessional:        false,
		MirrorStrangers:           false,
	}
}

func (e *ExpressionGestureEngine) getDefaultAnimationSettings() models.AnimationSettings {
	return models.AnimationSettings{
		AnimationStyle:              "realistic",
		IdleAnimationEnabled:        true,
		IdleAnimationSet:            []string{"breathing", "subtle_movement", "blinking"},
		IdleVariationFrequency:      0.3,
		TransitionAnimationsEnabled: true,
		TransitionSpeed:            e.config.TransitionSpeed,
		ExpressionAnimationQuality:  "high",
		ExpressionAnimationFPS:      e.config.DefaultAnimationFPS,
		GestureAnimationQuality:     "high",
		GestureAnimationFPS:         e.config.DefaultAnimationFPS,
		LipSyncEnabled:              true,
		LipSyncQuality:              "high",
		LipSyncDelay:                0,
		BlinkAnimationEnabled:       true,
		BlinkAnimationNatural:       true,
		BreathingAnimationEnabled:   true,
		BreathingRate:               16.0,
		BreathingDepth:              0.5,
		EyeMovementEnabled:          true,
		SaccadeFrequency:            3.0,
		SaccadeRange:                0.3,
	}
}
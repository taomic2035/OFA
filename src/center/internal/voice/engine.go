// Package voice provides the Voice Synthesis Engine for v5.1.0.
//
// VoiceEngine manages voice characteristics and synthesis settings,
// integrating with emotion and culture systems from v4.x.
package voice

import (
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// VoiceEngine manages voice state and synthesis.
type VoiceEngine struct {
	mu sync.RWMutex

	// Voice profiles (identity_id -> *VoiceProfile)
	profiles sync.Map

	// Synthesis cache
	cache sync.Map // text_hash -> *VoiceSynthesisResult

	// Configuration
	config VoiceEngineConfig
}

// VoiceEngineConfig holds configuration for the voice engine.
type VoiceEngineConfig struct {
	// Default voice settings
	DefaultVoiceType   string
	DefaultSpeakingRate double
	DefaultPitch       double

	// TTS settings
	DefaultEngine      string
	DefaultQuality     string
	EnableCache        bool
	CacheSize          int

	// Emotion integration
	EmotionInfluenceStrength double // 0-1, how much emotion affects voice
}

// NewVoiceEngine creates a new Voice Engine.
func NewVoiceEngine(config VoiceEngineConfig) *VoiceEngine {
	return &VoiceEngine{
		config: config,
	}
}

// DefaultVoiceEngineConfig returns default configuration.
func DefaultVoiceEngineConfig() VoiceEngineConfig {
	return VoiceEngineConfig{
		DefaultVoiceType:          "neutral",
		DefaultSpeakingRate:       150.0, // words per minute
		DefaultPitch:              150.0, // Hz
		DefaultEngine:             "built_in",
		DefaultQuality:            "high",
		EnableCache:               true,
		CacheSize:                 100, // MB
		EmotionInfluenceStrength:  0.7,
	}
}

// === Voice Profile Management ===

// CreateProfile creates a new voice profile for an identity.
func (e *VoiceEngine) CreateProfile(identityID string, voiceType string, voiceAge string) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	profile := &models.VoiceProfile{
		IdentityID: identityID,
		VoiceCharacteristics: e.getDefaultVoiceCharacteristics(voiceType, voiceAge),
		VoiceStyle:           e.getDefaultVoiceStyle(),
		EmotionalVoice:       e.getDefaultEmotionalVoice(),
		SpeechPatterns:       e.getDefaultSpeechPatterns(),
		TTSConfig:            e.getDefaultTTSConfig(),
		Version:              1,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	e.profiles.Store(identityID, profile)
	return profile
}

// GetProfile retrieves a voice profile by identity ID.
func (e *VoiceEngine) GetProfile(identityID string) *models.VoiceProfile {
	if value, ok := e.profiles.Load(identityID); ok {
		return value.(*models.VoiceProfile)
	}
	return nil
}

// UpdateVoiceCharacteristics updates voice characteristics.
func (e *VoiceEngine) UpdateVoiceCharacteristics(identityID string, chars models.VoiceCharacteristics) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	profile.VoiceCharacteristics = chars
	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// UpdateVoiceStyle updates voice style.
func (e *VoiceEngine) UpdateVoiceStyle(identityID string, style models.VoiceStyle) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	profile.VoiceStyle = style
	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// === Emotion Integration (v4.0.0, v4.5.0) ===

// ApplyEmotionInfluence applies emotion influence to voice.
func (e *VoiceEngine) ApplyEmotionInfluence(identityID string, emotion string, intensity double) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Get emotion mapping
	mapping, exists := profile.EmotionalVoice.EmotionMappings[emotion]
	if !exists {
		mapping = e.getDefaultEmotionMapping(emotion)
	}

	// Apply influence to voice characteristics
	strength := e.config.EmotionInfluenceStrength * intensity
	profile.VoiceCharacteristics = e.applyEmotionToCharacteristics(
		profile.VoiceCharacteristics,
		mapping,
		strength,
	)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyEmotionToCharacteristics applies emotion mapping to voice characteristics.
func (e *VoiceEngine) applyEmotionToCharacteristics(chars models.VoiceCharacteristics, mapping models.VoiceEmotionMapping, strength double) models.VoiceCharacteristics {
	result := chars

	// Apply pitch shift
	pitchShift := mapping.PitchShift * strength
	result.BasePitch += pitchShift

	// Apply rate change
	rateChange := mapping.RateMultiplier - 1.0
	adjustedRateChange := rateChange * strength
	result.SpeakingRate *= (1.0 + adjustedRateChange)

	// Apply volume shift
	volumeShift := mapping.VolumeShift * strength
	result.BaseVolume += volumeShift
	if result.BaseVolume > 1.0 {
		result.BaseVolume = 1.0
	}
	if result.BaseVolume < 0.0 {
		result.BaseVolume = 0.0
	}

	// Apply timbre adjustment
	if mapping.TimbreAdjust != "" {
		switch mapping.TimbreAdjust {
		case "warmer":
			result.TimbreType = "warm"
		case "brighter":
			result.TimbreType = "bright"
		case "darker":
			result.TimbreType = "dark"
		}
	}

	// Apply tension
	result.Roughness = mapping.TensionLevel * strength

	return result
}

// getDefaultEmotionMapping returns default emotion mapping.
func (e *VoiceEngine) getDefaultEmotionMapping(emotion string) models.VoiceEmotionMapping {
	mappings := map[string]models.VoiceEmotionMapping{
		"joy": {
			EmotionName:    "joy",
			PitchShift:     2.0,  // higher pitch
			RateMultiplier: 1.1,  // slightly faster
			VolumeShift:    0.1,  // louder
			TimbreAdjust:   "brighter",
			Articulation:   "relaxed",
		},
		"anger": {
			EmotionName:    "anger",
			PitchShift:     3.0,  // higher, tense
			RateMultiplier: 1.2,  // faster
			VolumeShift:    0.2,  // louder
			TimbreAdjust:   "brighter",
			Articulation:   "tense",
			TensionLevel:   0.7,
		},
		"sadness": {
			EmotionName:    "sadness",
			PitchShift:     -2.0, // lower pitch
			RateMultiplier: 0.8,  // slower
			VolumeShift:    -0.1, // quieter
			TimbreAdjust:   "darker",
			Articulation:   "relaxed",
			BreathinessAdjust: 0.2,
		},
		"fear": {
			EmotionName:    "fear",
			PitchShift:     3.0,  // higher, shaky
			RateMultiplier: 1.15, // faster
			VolumeShift:    0.0,  // normal
			TimbreAdjust:   "brighter",
			Articulation:   "tense",
			TensionLevel:   0.6,
		},
		"love": {
			EmotionName:    "love",
			PitchShift:     1.0,  // slightly higher
			RateMultiplier: 0.95, // slightly slower
			VolumeShift:    -0.05, // slightly softer
			TimbreAdjust:   "warmer",
			Articulation:   "relaxed",
		},
		"disgust": {
			EmotionName:    "disgust",
			PitchShift:     -1.0, // slightly lower
			RateMultiplier: 0.9,  // slower
			VolumeShift:    -0.05, // quieter
			TimbreAdjust:   "darker",
			Articulation:   "tense",
		},
		"desire": {
			EmotionName:    "desire",
			PitchShift:     0.5,  // slightly higher
			RateMultiplier: 0.9,  // slower
			VolumeShift:    -0.1, // softer
			TimbreAdjust:   "warmer",
			Articulation:   "relaxed",
		},
	}

	if mapping, exists := mappings[emotion]; exists {
		return mapping
	}

	return models.VoiceEmotionMapping{
		EmotionName:    emotion,
		PitchShift:     0.0,
		RateMultiplier: 1.0,
		VolumeShift:    0.0,
		TimbreAdjust:   "neutral",
		Articulation:   "neutral",
	}
}

// === Cultural Integration (v4.3.0) ===

// ApplyCulturalInfluence applies cultural influence to voice style.
func (e *VoiceEngine) ApplyCulturalInfluence(identityID string, culture *models.RegionalCulture) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Apply accent based on region
	profile.VoiceStyle.RegionalAccent = e.determineAccent(culture)
	profile.VoiceStyle.AccentRegion = culture.CurrentLocation.Province
	profile.VoiceStyle.AccentIntensity = e.calculateAccentIntensity(culture)

	// Apply formality based on culture
	profile.VoiceStyle.FormalityLevel = e.determineFormality(culture)
	profile.VoiceStyle.CommunicationStyle = e.determineCommunicationStyle(culture)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// determineAccent determines accent type from culture.
func (e *VoiceEngine) determineAccent(culture *models.RegionalCulture) string {
	if culture == nil {
		return "standard"
	}

	// Check migration history
	if len(culture.MigrationHistory) > 0 {
		return "mixed"
	}

	// Based on city level
	switch culture.CityLevel {
	case "first_tier", "new_first_tier":
		return "standard"
	default:
		return "regional"
	}
}

// calculateAccentIntensity calculates accent intensity.
func (e *VoiceEngine) calculateAccentIntensity(culture *models.RegionalCulture) double {
	if culture == nil {
		return 0.0
	}

	// Higher intensity for smaller cities
	switch culture.CityLevel {
	case "first_tier", "new_first_tier":
		return 0.2
	case "second_tier":
		return 0.4
	case "third_tier":
		return 0.6
	default:
		return 0.8
	}
}

// determineFormality determines formality level from culture.
func (e *VoiceEngine) determineFormality(culture *models.RegionalCulture) string {
	if culture == nil {
		return "neutral"
	}

	// Based on Hofstede dimensions
	if culture.HofstedeDimensions.PowerDistance > 0.7 {
		return "formal"
	} else if culture.HofstedeDimensions.PowerDistance < 0.4 {
		return "casual"
	}
	return "neutral"
}

// determineCommunicationStyle determines communication style from culture.
func (e *VoiceEngine) determineCommunicationStyle(culture *models.RegionalCulture) string {
	if culture == nil {
		return "direct"
	}

	// Based on Hofstede collectivism
	if culture.HofstedeDimensions.Collectivism > 0.6 {
		return "indirect"
	} else if culture.HofstedeDimensions.Collectivism < 0.4 {
		return "direct"
	}
	return "context_dependent"
}

// === Life Stage Integration (v4.4.0) ===

// ApplyLifeStageInfluence applies life stage influence to voice.
func (e *VoiceEngine) ApplyLifeStageInfluence(identityID string, lifeStage string, age int) *models.VoiceProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Update voice age
	profile.VoiceCharacteristics.VoiceAge = e.determineVoiceAge(lifeStage, age)

	// Update speech patterns based on life stage
	profile.SpeechPatterns = e.applyLifeStageSpeechPatterns(lifeStage, profile.SpeechPatterns)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// determineVoiceAge determines voice age from life stage.
func (e *VoiceEngine) determineVoiceAge(lifeStage string, age int) string {
	switch lifeStage {
	case "childhood", "adolescence":
		return "child"
	case "youth", "early_adult":
		return "young"
	case "mid_adult":
		return "adult"
	case "mature":
		return "middle_aged"
	case "elderly":
		return "senior"
	default:
		if age < 18 {
			return "child"
		} else if age < 30 {
			return "young"
		} else if age < 50 {
			return "adult"
		} else if age < 65 {
			return "middle_aged"
		}
		return "senior"
	}
}

// applyLifeStageSpeechPatterns applies life stage influence to speech patterns.
func (e *VoiceEngine) applyLifeStageSpeechPatterns(lifeStage string, patterns models.SpeechPatterns) models.SpeechPatterns {
	result := patterns

	switch lifeStage {
	case "childhood", "adolescence":
		result.VocabularyLevel = "simple"
		result.SentenceLength = "short"
		result.SentenceComplexity = "simple"
		result.ThoughtOrganization = "associative"
	case "youth", "early_adult":
		result.VocabularyLevel = "moderate"
		result.SentenceLength = "medium"
		result.SentenceComplexity = "compound"
		result.ThoughtOrganization = "structured"
	case "mid_adult", "mature":
		result.VocabularyLevel = "sophisticated"
		result.SentenceLength = "long"
		result.SentenceComplexity = "complex"
		result.ThoughtOrganization = "structured"
	case "elderly":
		result.VocabularyLevel = "sophisticated"
		result.SentenceLength = "medium"
		result.SentenceComplexity = "complex"
		result.PauseBeforeSpeaking = 0.4
	}

	return result
}

// === Synthesis ===

// Synthesize synthesizes speech from text.
func (e *VoiceEngine) Synthesize(req *models.VoiceSynthesisRequest) *models.VoiceSynthesisResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.GetProfile(req.IdentityID)
	if profile == nil {
		return &models.VoiceSynthesisResult{
			IdentityID: req.IdentityID,
			Success:    false,
			Error:      "voice profile not found",
		}
	}

	// Calculate final voice parameters
	pitch := e.calculateFinalPitch(profile, req)
	rate := e.calculateFinalRate(profile, req)
	volume := e.calculateFinalVolume(profile, req)

	// Generate synthesis result
	result := &models.VoiceSynthesisResult{
		IdentityID:    req.IdentityID,
		RequestID:     generateRequestID(),
		AudioFormat:   req.OutputFormat,
		DurationMs:    e.estimateDuration(req.Text, rate),
		SampleRate:    profile.TTSConfig.SampleRate,
		EngineUsed:    profile.TTSConfig.EngineName,
		PitchApplied:  pitch,
		RateApplied:   rate,
		VolumeApplied: volume,
		QualityScore:  0.85,
		Success:       true,
	}

	// In production, this would call actual TTS engine
	// result.AudioData = e.callTTSEngine(req.Text, pitch, rate, volume)

	return result
}

// calculateFinalPitch calculates final pitch for synthesis.
func (e *VoiceEngine) calculateFinalPitch(profile *models.VoiceProfile, req *models.VoiceSynthesisRequest) double {
	pitch := profile.VoiceCharacteristics.BasePitch

	// Apply emotion influence
	if req.EmotionContext.CurrentEmotion != "" {
		mapping := e.getDefaultEmotionMapping(req.EmotionContext.CurrentEmotion)
		pitch += mapping.PitchShift * req.EmotionContext.EmotionIntensity * e.config.EmotionInfluenceStrength
	}

	// Apply scene influence
	switch req.SpeechContext.Scene {
	case "meeting", "formal":
		pitch -= 1.0 // slightly lower for authority
	case "casual":
		pitch += 0.5 // slightly higher for warmth
	}

	return pitch
}

// calculateFinalRate calculates final speaking rate.
func (e *VoiceEngine) calculateFinalRate(profile *models.VoiceProfile, req *models.VoiceSynthesisRequest) double {
	rate := profile.VoiceCharacteristics.SpeakingRate

	// Apply emotion influence
	if req.EmotionContext.CurrentEmotion != "" {
		mapping := e.getDefaultEmotionMapping(req.EmotionContext.CurrentEmotion)
		rate *= mapping.RateMultiplier
	}

	// Apply formality
	switch req.SpeechContext.Formality {
	case "formal":
		rate *= 0.9 // slower for clarity
	case "casual":
		rate *= 1.05 // slightly faster
	}

	return rate
}

// calculateFinalVolume calculates final volume for synthesis.
func (e *VoiceEngine) calculateFinalVolume(profile *models.VoiceProfile, req *models.VoiceSynthesisRequest) double {
	volume := profile.VoiceCharacteristics.BaseVolume

	// Apply emotion influence
	if req.EmotionContext.CurrentEmotion != "" {
		mapping := e.getDefaultEmotionMapping(req.EmotionContext.CurrentEmotion)
		volume += mapping.VolumeShift
	}

	// Clamp to valid range
	if volume > 1.0 {
		volume = 1.0
	}
	if volume < 0.0 {
		volume = 0.0
	}

	return volume
}

// estimateDuration estimates audio duration in ms.
func (e *VoiceEngine) estimateDuration(text string, rate double) int {
	// Simple estimation: characters / rate * 60000
	// Assuming average word is 5 characters
	wordCount := len(text) / 5
	durationMs := int(float64(wordCount) / rate * 60000)
	return durationMs
}

// === Decision Context ===

// GetDecisionContext generates voice decision context.
func (e *VoiceEngine) GetDecisionContext(identityID string, emotion string, scene string, socialContext string) *models.VoiceDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	ctx := &models.VoiceDecisionContext{
		IdentityID: identityID,
		Timestamp:  time.Now(),
	}

	// Calculate recommendations
	ctx.RecommendedPitch = e.recommendPitch(profile, emotion, scene)
	ctx.RecommendedRate = e.recommendRate(profile, scene, socialContext)
	ctx.RecommendedVolume = e.recommendVolume(profile, scene)
	ctx.RecommendedTone = e.recommendTone(profile, emotion)
	ctx.RecommendedFormality = e.recommendFormality(profile, scene, socialContext)

	// Generate adaptations
	ctx.EmotionAdaptation = e.generateEmotionAdaptation(emotion, profile)
	ctx.SceneAdaptation = e.generateSceneAdaptation(scene, profile)
	ctx.SocialAdaptation = e.generateSocialAdaptation(socialContext, profile)

	// Generate output settings
	ctx.OutputSettings = e.generateOutputSettings(profile)

	return ctx
}

// recommendPitch recommends pitch for context.
func (e *VoiceEngine) recommendPitch(profile *models.VoiceProfile, emotion string, scene string) double {
	pitch := profile.VoiceCharacteristics.BasePitch

	// Emotion influence
	if emotion != "" {
		mapping := e.getDefaultEmotionMapping(emotion)
		pitch += mapping.PitchShift * 0.5
	}

	// Scene influence
	switch scene {
	case "meeting", "formal":
		pitch -= 1.0
	case "casual", "home":
		pitch += 0.5
	}

	return pitch
}

// recommendRate recommends speaking rate for context.
func (e *VoiceEngine) recommendRate(profile *models.VoiceProfile, scene string, socialContext string) double {
	rate := profile.VoiceCharacteristics.SpeakingRate

	switch scene {
	case "meeting":
		rate *= 0.9
	case "presentation", "public":
		rate *= 0.85
	case "casual":
		rate *= 1.05
	}

	switch socialContext {
	case "intimate":
		rate *= 0.95
	case "group", "public":
		rate *= 0.9
	}

	return rate
}

// recommendVolume recommends volume for context.
func (e *VoiceEngine) recommendVolume(profile *models.VoiceProfile, scene string) double {
	volume := profile.VoiceCharacteristics.BaseVolume

	switch scene {
	case "meeting":
		volume += 0.1
	case "public":
		volume += 0.15
	case "intimate":
		volume -= 0.1
	}

	if volume > 1.0 {
		volume = 1.0
	}
	if volume < 0.0 {
		volume = 0.0
	}

	return volume
}

// recommendTone recommends tone for context.
func (e *VoiceEngine) recommendTone(profile *models.VoiceProfile, emotion string) string {
	tone := profile.VoiceCharacteristics.TimbreType

	if emotion != "" {
		mapping := e.getDefaultEmotionMapping(emotion)
		if mapping.TimbreAdjust != "" && mapping.TimbreAdjust != "neutral" {
			return mapping.TimbreAdjust
		}
	}

	return tone
}

// recommendFormality recommends formality for context.
func (e *VoiceEngine) recommendFormality(profile *models.VoiceProfile, scene string, socialContext string) string {
	formality := profile.VoiceStyle.FormalityLevel

	if scene == "formal" || socialContext == "professional" {
		return "formal"
	} else if scene == "casual" || socialContext == "intimate" {
		return "casual"
	}

	return formality
}

// generateEmotionAdaptation generates emotion adaptation.
func (e *VoiceEngine) generateEmotionAdaptation(emotion string, profile *models.VoiceProfile) models.VoiceEmotionAdaptation {
	adaptation := models.VoiceEmotionAdaptation{
		CurrentEmotion: emotion,
	}

	if emotion != "" {
		mapping := e.getDefaultEmotionMapping(emotion)
		adaptation.PitchShift = mapping.PitchShift
		adaptation.RateMultiplier = mapping.RateMultiplier
		adaptation.VolumeShift = mapping.VolumeShift
		adaptation.TimbreAdjust = mapping.TimbreAdjust
		adaptation.ExpressivenessLevel = profile.EmotionalVoice.EmotionalExpressiveness
	}

	return adaptation
}

// generateSceneAdaptation generates scene adaptation.
func (e *VoiceEngine) generateSceneAdaptation(scene string, profile *models.VoiceProfile) models.VoiceSceneAdaptation {
	adaptation := models.VoiceSceneAdaptation{
		Scene: scene,
	}

	switch scene {
	case "meeting":
		adaptation.PitchAdjust = -1.0
		adaptation.RateAdjust = 0.9
		adaptation.VolumeAdjust = 0.1
		adaptation.FormalityAdjust = "formal_up"
		adaptation.ArticulationMode = "precise"
	case "casual":
		adaptation.PitchAdjust = 0.5
		adaptation.RateAdjust = 1.05
		adaptation.VolumeAdjust = 0.0
		adaptation.FormalityAdjust = "casual_up"
		adaptation.ArticulationMode = "relaxed"
	case "presentation", "public":
		adaptation.PitchAdjust = -0.5
		adaptation.RateAdjust = 0.85
		adaptation.VolumeAdjust = 0.15
		adaptation.FormalityAdjust = "formal_up"
		adaptation.ArticulationMode = "precise"
	default:
		adaptation.ArticulationMode = "neutral"
	}

	return adaptation
}

// generateSocialAdaptation generates social adaptation.
func (e *VoiceEngine) generateSocialAdaptation(socialContext string, profile *models.VoiceProfile) models.VoiceSocialAdaptation {
	adaptation := models.VoiceSocialAdaptation{
		SocialContext: socialContext,
	}

	switch socialContext {
	case "professional", "formal":
		adaptation.AuthorityLevel = 0.7
		adaptation.WarmthLevel = 0.4
		adaptation.Directness = 0.6
		adaptation.PauseBeforeSpeech = 0.3
	case "intimate", "close":
		adaptation.AuthorityLevel = 0.3
		adaptation.WarmthLevel = 0.8
		adaptation.Directness = 0.7
		adaptation.PauseBeforeSpeech = 0.1
	case "group", "public":
		adaptation.AuthorityLevel = 0.6
		adaptation.WarmthLevel = 0.5
		adaptation.Directness = 0.5
		adaptation.PauseBeforeSpeech = 0.4
	default:
		adaptation.AuthorityLevel = 0.5
		adaptation.WarmthLevel = 0.5
		adaptation.Directness = 0.5
		adaptation.PauseBeforeSpeech = 0.2
	}

	return adaptation
}

// generateOutputSettings generates output settings.
func (e *VoiceEngine) generateOutputSettings(profile *models.VoiceProfile) models.VoiceOutputSettings {
	return models.VoiceOutputSettings{
		OutputDevice:   "speaker",
		VolumeLevel:    profile.VoiceCharacteristics.BaseVolume,
		SpatialMode:    "mono",
		EffectsEnabled: false,
		ReverbLevel:    0.0,
	}
}

// === Helper methods ===

func (e *VoiceEngine) getDefaultVoiceCharacteristics(voiceType string, voiceAge string) models.VoiceCharacteristics {
	chars := models.VoiceCharacteristics{
		VoiceType:         voiceType,
		VoiceAge:          voiceAge,
		VoiceQuality:      "clear",
		BasePitch:         150.0,
		PitchRange:        12.0,
		PitchVariation:    "moderate",
		SpeakingRate:      e.config.DefaultSpeakingRate,
		RateVariation:     "varied",
		PauseFrequency:    0.3,
		BaseVolume:        0.7,
		VolumeRange:       0.3,
		VolumeVariation:   "dynamic",
		TimbreType:        "neutral",
		Resonance:         "medium",
		Breathiness:       0.1,
		Roughness:         0.1,
		ArticulationStyle: "relaxed",
		AccentStrength:    0.3,
		Distinctiveness:   0.5,
	}

	// Adjust based on voice type
	switch voiceType {
	case "male":
		chars.BasePitch = 120.0
		chars.Resonance = "low"
	case "female":
		chars.BasePitch = 200.0
		chars.Resonance = "high"
	}

	// Adjust based on voice age
	switch voiceAge {
	case "child":
		chars.BasePitch += 50.0
		chars.SpeakingRate *= 1.1
	case "senior":
		chars.BasePitch -= 10.0
		chars.SpeakingRate *= 0.9
		chars.Roughness = 0.2
	}

	return chars
}

func (e *VoiceEngine) getDefaultVoiceStyle() models.VoiceStyle {
	return models.VoiceStyle{
		RegionalAccent:     "standard",
		AccentIntensity:    0.3,
		FormalityLevel:     "neutral",
		FormalityAdjust:    0.0,
		CommunicationStyle: "direct",
		IndirectnessLevel:  0.3,
		ProfessionalVoice:  false,
		VoiceAuthority:     0.5,
		VoiceWarmth:        0.5,
		AdaptiveStyle:      true,
		StyleConsistency:   0.8,
	}
}

func (e *VoiceEngine) getDefaultEmotionalVoice() models.EmotionalVoice {
	return models.EmotionalVoice{
		EmotionMappings: map[string]models.VoiceEmotionMapping{
			"joy":     e.getDefaultEmotionMapping("joy"),
			"anger":   e.getDefaultEmotionMapping("anger"),
			"sadness": e.getDefaultEmotionMapping("sadness"),
			"fear":    e.getDefaultEmotionMapping("fear"),
			"love":    e.getDefaultEmotionMapping("love"),
			"disgust": e.getDefaultEmotionMapping("disgust"),
			"desire":  e.getDefaultEmotionMapping("desire"),
		},
		EmotionalExpressiveness:  0.6,
		EmotionalContagion:       0.5,
		PitchModulationRange:     6.0,
		RateModulationRange:      0.3,
		VolumeModulation:         0.2,
		EmotionRegulation:        "express",
		BaselineStability:        0.7,
	}
}

func (e *VoiceEngine) getDefaultSpeechPatterns() models.SpeechPatterns {
	return models.SpeechPatterns{
		VocabularyLevel:      "moderate",
		JargonUsage:          0.3,
		IdiomUsage:           0.4,
		MetaphorUsage:        0.3,
		SentenceLength:       "medium",
		SentenceComplexity:   "compound",
		ClauseUsage:          0.5,
		FillerWords:          []string{"嗯", "那个"},
		FillerFrequency:      0.2,
		HedgingLanguage:      0.3,
		QualifyingWords:      0.3,
		ThoughtOrganization:  "structured",
		DigressionTendency:   0.3,
		TangentFrequency:     0.2,
		ListeningResponses:   []string{"嗯", "好的", "我明白了"},
		ResponseFrequency:    0.4,
		InterruptTendency:    0.2,
		PauseBeforeSpeaking:  0.2,
		ThoughtfulPauses:     0.3,
		RhetoricalPauses:     0.2,
	}
}

func (e *VoiceEngine) getDefaultTTSConfig() models.TTSConfiguration {
	return models.TTSConfiguration{
		EngineType:       e.config.DefaultEngine,
		EngineName:       "ofa_tts",
		EngineVersion:    "1.0.0",
		VoiceModelID:     "default",
		VoiceModelName:   "Default Voice",
		CustomVoice:      false,
		SampleRate:       22050,
		BitDepth:         16,
		Channels:         1,
		AudioFormat:      "wav",
		QualityLevel:     e.config.DefaultQuality,
		LatencyMode:      "balanced",
		BufferSize:       100,
		SSMLEnabled:      true,
		SSMLFeatures:     []string{"break", "emphasis", "prosody"},
		StreamingEnabled: true,
		ChunkSize:        50,
		CacheEnabled:     e.config.EnableCache,
		CacheSize:        e.config.CacheSize,
		CacheTTL:         3600,
		FallbackEngine:   "built_in",
		FailoverEnabled:  true,
	}
}

func generateRequestID() string {
	return "voice_" + time.Now().Format("20060102150405")
}
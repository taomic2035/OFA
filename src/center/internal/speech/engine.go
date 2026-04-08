// Package speech provides the Speech Content Engine for v5.2.0.
//
// SpeechContentEngine manages content generation and expression,
// integrating with philosophy, culture, and emotion systems from v4.x.
package speech

import (
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// SpeechContentEngine manages speech content state and generation.
type SpeechContentEngine struct {
	mu sync.RWMutex

	// Profiles (identity_id -> *SpeechContentProfile)
	profiles sync.Map

	// Content cache
	cache sync.Map // request_hash -> *SpeechContentResult

	// Configuration
	config SpeechContentEngineConfig
}

// SpeechContentEngineConfig holds configuration for the speech content engine.
type SpeechContentEngineConfig struct {
	// Default style
	DefaultTone        string
	DefaultFormality   string
	DefaultLanguageLevel string

	// Generation settings
	EnableCache      bool
	CacheTTL         int // seconds
	MaxContentLength int
	MinContentLength int

	// Integration strength
	PhilosophyInfluence float64 // 0-1
	CultureInfluence    float64 // 0-1
	EmotionInfluence    float64 // 0-1
}

// NewSpeechContentEngine creates a new Speech Content Engine.
func NewSpeechContentEngine(config SpeechContentEngineConfig) *SpeechContentEngine {
	return &SpeechContentEngine{
		config: config,
	}
}

// DefaultSpeechContentEngineConfig returns default configuration.
func DefaultSpeechContentEngineConfig() SpeechContentEngineConfig {
	return SpeechContentEngineConfig{
		DefaultTone:         "neutral",
		DefaultFormality:    "neutral",
		DefaultLanguageLevel: "moderate",
		EnableCache:         true,
		CacheTTL:            3600,
		MaxContentLength:    1000,
		MinContentLength:    10,
		PhilosophyInfluence: 0.7,
		CultureInfluence:    0.6,
		EmotionInfluence:    0.8,
	}
}

// === Profile Management ===

// CreateProfile creates a new speech content profile.
func (e *SpeechContentEngine) CreateProfile(identityID string) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	profile := &models.SpeechContentProfile{
		IdentityID:         identityID,
		ContentStyle:       e.getDefaultContentStyle(),
		ExpressionDepth:    e.getDefaultExpressionDepth(),
		CulturalExpression: e.getDefaultCulturalExpression(),
		SocialExpression:   e.getDefaultSocialExpression(),
		ContentTemplates:   e.getDefaultContentTemplates(),
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	e.profiles.Store(identityID, profile)
	return profile
}

// GetProfile retrieves a profile by identity ID.
func (e *SpeechContentEngine) GetProfile(identityID string) *models.SpeechContentProfile {
	if value, ok := e.profiles.Load(identityID); ok {
		return value.(*models.SpeechContentProfile)
	}
	return nil
}

// UpdateContentStyle updates content style.
func (e *SpeechContentEngine) UpdateContentStyle(identityID string, style models.ContentStyle) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	profile.ContentStyle = style
	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// === Philosophy Integration (v4.1.0) ===

// ApplyPhilosophyInfluence applies philosophy influence to content style.
func (e *SpeechContentEngine) ApplyPhilosophyInfluence(identityID string, worldview *models.Worldview, lifeView *models.LifeView, valueSystem *models.EnhancedValueSystem) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Apply worldview influence
	profile.ContentStyle = e.applyWorldviewToStyle(worldview, profile.ContentStyle)

	// Apply life view influence
	profile.ExpressionDepth = e.applyLifeViewToDepth(lifeView, profile.ExpressionDepth)

	// Apply value system influence
	profile.ContentStyle = e.applyValuesToStyle(valueSystem, profile.ContentStyle)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyWorldviewToStyle applies worldview to content style.
func (e *SpeechContentEngine) applyWorldviewToStyle(worldview *models.Worldview, style models.ContentStyle) models.ContentStyle {
	result := style

	if worldview == nil {
		return result
	}

	// World essence affects expression style
	switch worldview.WorldEssence {
	case "deterministic":
		result.PersuasionStyle = "logical"
		result.EvidenceType = "data"
	case "probabilistic":
		result.PersuasionStyle = "balanced"
		result.EvidenceType = "mixed"
	case "complex_adaptive":
		result.PersuasionStyle = "balanced"
		result.ThinkingDepth = "deep"
	}

	// Social cognition affects social expression
	switch worldview.SocialCognition {
	case "cooperative":
		result.ToneStyle = "friendly"
		result.EnthusiasmLevel = 0.6
	case "competitive":
		result.ToneStyle = "professional"
		result.Directness = 0.7
	case "cooperative_competitive":
		result.ToneStyle = "balanced"
	}

	return result
}

// applyLifeViewToDepth applies life view to expression depth.
func (e *SpeechContentEngine) applyLifeViewToDepth(lifeView *models.LifeView, depth models.ExpressionDepth) models.ExpressionDepth {
	result := depth

	if lifeView == nil {
		return result
	}

	// Life meaning affects depth
	switch lifeView.LifeMeaning {
	case "growth_contribution":
		result.ThinkingDepth = "deep"
		result.SelfDisclosureLevel = 0.6
	case "pleasure_comfort":
		result.ThinkingDepth = "surface"
		result.SelfDisclosureLevel = 0.4
	case "relationships_connection":
		result.SelfDisclosureLevel = 0.7
		result.IntimacyThreshold = 0.4
	}

	// Time orientation affects reflection
	switch lifeView.TimeOrientation {
	case "future_oriented":
		result.ReflectionTendency = 0.4
	case "present_oriented":
		result.ReflectionTendency = 0.3
	case "past_oriented":
		result.ReflectionTendency = 0.7
	}

	return result
}

// applyValuesToStyle applies value system to content style.
func (e *SpeechContentEngine) applyValuesToStyle(valueSystem *models.EnhancedValueSystem, style models.ContentStyle) models.ContentStyle {
	result := style

	if valueSystem == nil {
		return result
	}

	// Check key values
	for _, value := range valueSystem.CoreValues {
		switch value.Name {
		case "honesty", "truth":
			result.Directness = 0.8
			result.EuphemismUsage = 0.2
		case "kindness", "compassion":
			result.ToneStyle = "warm"
			result.EmotionalColoring = "warm"
		case "professionalism":
			result.ToneStyle = "professional"
			result.FormalityConsistency = 0.8
		case "creativity":
			result.MetaphorUsage = 0.7
			result.AnalogyUsage = 0.7
		case "humility":
			result.HumorTendency = 0.3
			result.EnthusiasmLevel = 0.4
		}
	}

	return result
}

// === Regional Culture Integration (v4.3.0) ===

// ApplyCulturalInfluence applies cultural influence to expression.
func (e *SpeechContentEngine) ApplyCulturalInfluence(identityID string, culture *models.RegionalCulture) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Apply cultural expression settings
	profile.CulturalExpression = e.applyCultureToExpression(culture)

	// Apply to content style
	profile.ContentStyle = e.applyCultureToStyle(culture, profile.ContentStyle)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyCultureToExpression applies culture to cultural expression.
func (e *SpeechContentEngine) applyCultureToExpression(culture *models.RegionalCulture) models.CulturalExpression {
	expr := models.CulturalExpression{
		HonorificUsage: "light",
	}

	if culture == nil {
		return expr
	}

	// Hofstede dimensions affect expression
	dims := culture.HofstedeDimensions

	// Power distance affects respect
	if dims.PowerDistance > 0.6 {
		expr.RespectLevel = 0.8
		expr.HierarchyAwareness = 0.7
		expr.HonorificUsage = "moderate"
	} else if dims.PowerDistance < 0.4 {
		expr.RespectLevel = 0.4
		expr.HierarchyAwareness = 0.3
		expr.HonorificUsage = "light"
	}

	// Collectivism affects group reference
	if dims.Collectivism > 0.6 {
		expr.CollectivistExpression = 0.7
		expr.GroupReferenceUsage = 0.7
		expr.IndirectExpression = 0.6
		expr.HighContextCommunication = true
	} else {
		expr.CollectivistExpression = 0.3
		expr.GroupReferenceUsage = 0.3
		expr.IndirectExpression = 0.3
		expr.HighContextCommunication = false
	}

	// Long-term orientation affects face saving
	if dims.LongTermOrientation > 0.6 {
		expr.FaceSaving = 0.7
	}

	// Communication style from culture
	expr.IndirectExpression = e.calculateIndirectness(culture)

	return expr
}

// applyCultureToStyle applies culture to content style.
func (e *SpeechContentEngine) applyCultureToStyle(culture *models.RegionalCulture, style models.ContentStyle) models.ContentStyle {
	result := style

	if culture == nil {
		return result
	}

	// Communication style
	switch culture.CommunicationStyle {
	case "direct":
		result.Directness = 0.8
		result.EuphemismUsage = 0.2
	case "indirect":
		result.Directness = 0.3
		result.EuphemismUsage = 0.7
	case "context_dependent":
		result.Directness = 0.5
		result.EuphemismUsage = 0.5
	}

	return result
}

// calculateIndirectness calculates indirectness level from culture.
func (e *SpeechContentEngine) calculateIndirectness(culture *models.RegionalCulture) float64 {
	if culture == nil {
		return 0.5
	}

	indirectness := 0.5

	if culture.HofstedeDimensions.Collectivism > 0.6 {
		indirectness += 0.2
	}
	if culture.HofstedeDimensions.PowerDistance > 0.6 {
		indirectness += 0.1
	}
	if culture.CommunicationStyle == "indirect" {
		indirectness += 0.2
	}

	if indirectness > 1.0 {
		indirectness = 1.0
	}

	return indirectness
}

// === Social Identity Integration (v4.2.0) ===

// ApplySocialIdentityInfluence applies social identity influence.
func (e *SpeechContentEngine) ApplySocialIdentityInfluence(identityID string, socialIdentity *models.IdentityProfile, career *models.CareerProfile, socialClass *models.SocialClassProfile) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Apply social expression settings
	profile.SocialExpression = e.applySocialToExpression(socialIdentity, career, socialClass)

	// Apply to content style
	profile.ContentStyle = e.applyCareerToStyle(career, profile.ContentStyle)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applySocialToExpression applies social identity to expression.
func (e *SpeechContentEngine) applySocialToExpression(identity *models.IdentityProfile, career *models.CareerProfile, socialClass *models.SocialClassProfile) models.SocialExpression {
	expr := models.SocialExpression{
		ProfessionalTone:   "collaborative",
		ClassExpression:    "moderate",
		NetworkingStyle:    "balanced",
		AuthorityExpression: "earned",
	}

	if identity != nil {
		expr.IdentityConfidence = identity.IdentityConfidence
	}

	if career != nil {
		// Industry affects professional tone
		switch career.Industry {
		case "finance", "law":
			expr.ProfessionalTone = "authoritative"
			expr.ExpertiseDisplay = 0.7
		case "education", "healthcare":
			expr.ProfessionalTone = "supportive"
			expr.HumilityExpression = 0.6
		case "technology":
			expr.ProfessionalTone = "collaborative"
			expr.ExpertiseDisplay = 0.6
		}
	}

	if socialClass != nil {
		// Economic capital affects class expression
		if socialClass.EconomicCapital.Income == "high" || socialClass.EconomicCapital.Income == "very_high" {
			expr.ClassExpression = "understated"
		} else {
			expr.ClassExpression = "moderate"
		}
	}

	return expr
}

// applyCareerToStyle applies career to content style.
func (e *SpeechContentEngine) applyCareerToStyle(career *models.CareerProfile, style models.ContentStyle) models.ContentStyle {
	result := style

	if career == nil {
		return result
	}

	switch career.Industry {
	case "finance", "law":
		result.ToneStyle = "professional"
		result.LanguageLevel = "sophisticated"
		result.HumorTendency = 0.2
	case "creative", "media":
		result.ToneStyle = "casual"
		result.MetaphorUsage = 0.7
		result.HumorTendency = 0.6
	case "education":
		result.ToneStyle = "friendly"
		result.LanguageLevel = "moderate"
		result.AnalogyUsage = 0.7
	case "technology":
		result.TechnicalLevel = "expert"
		result.JargonTolerance = 0.6
	}

	return result
}

// === Emotion Integration (v4.0.0) ===

// ApplyEmotionInfluence applies emotion influence to content.
func (e *SpeechContentEngine) ApplyEmotionInfluence(identityID string, emotion string, intensity float64) *models.SpeechContentProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	// Apply emotion to content style
	profile.ContentStyle = e.applyEmotionToStyle(emotion, intensity, profile.ContentStyle)

	profile.Version++
	profile.UpdatedAt = time.Now()

	e.profiles.Store(identityID, profile)
	return profile
}

// applyEmotionToStyle applies emotion to content style.
func (e *SpeechContentEngine) applyEmotionToStyle(emotion string, intensity float64, style models.ContentStyle) models.ContentStyle {
	result := style

	switch emotion {
	case "joy":
		result.EmotionalColoring = "warm"
		result.EnthusiasmLevel = 0.6 + intensity*0.3
		result.HumorTendency = style.HumorTendency + intensity*0.2
	case "anger":
		result.EmotionalColoring = "passionate"
		result.Directness = style.Directness + intensity*0.2
		result.EnthusiasmLevel = style.EnthusiasmLevel - intensity*0.2
	case "sadness":
		result.EmotionalColoring = "cool"
		result.EnthusiasmLevel = style.EnthusiasmLevel - intensity*0.3
		result.HumorTendency = style.HumorTendency - intensity*0.3
	case "fear":
		result.EmotionalColoring = "cool"
		result.Directness = style.Directness - intensity*0.2
		result.EuphemismUsage = style.EuphemismUsage + intensity*0.2
	case "love":
		result.EmotionalColoring = "warm"
		result.SelfDisclosureLevel = style.SelfDisclosureLevel + intensity*0.2
		result.EnthusiasmLevel = 0.7
	case "disgust":
		result.EmotionalColoring = "cool"
		result.Directness = style.Directness + intensity*0.1
	case "desire":
		result.EmotionalColoring = "warm"
		result.EnthusiasmLevel = 0.6 + intensity*0.2
	}

	// Clamp values
	if result.EnthusiasmLevel > 1.0 {
		result.EnthusiasmLevel = 1.0
	}
	if result.EnthusiasmLevel < 0.0 {
		result.EnthusiasmLevel = 0.0
	}
	if result.Directness > 1.0 {
		result.Directness = 1.0
	}
	if result.Directness < 0.0 {
		result.Directness = 0.0
	}

	return result
}

// === Content Generation ===

// GenerateContent generates content based on request.
func (e *SpeechContentEngine) GenerateContent(req *models.SpeechContentRequest) *models.SpeechContentResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.GetProfile(req.IdentityID)
	if profile == nil {
		return &models.SpeechContentResult{
			IdentityID: req.IdentityID,
			RequestID:  generateRequestID(),
		}
	}

	// Calculate final style settings
	style := e.calculateFinalStyle(profile, req)

	// Generate content (placeholder - in production would use LLM)
	content := e.generateContentInternal(profile, req, style)

	result := &models.SpeechContentResult{
		IdentityID:          req.IdentityID,
		RequestID:           generateRequestID(),
		Content:             content,
		ContentType:         req.ContentType,
		ContentStyle:        style.ToneStyle,
		WordCount:           countWords(content),
		CharacterCount:      len(content),
		ToneUsed:            style.ToneStyle,
		FormalityLevel:      req.Formality,
		EmotionalTone:       req.EmotionContext.AffectTone,
		LanguageComplexity:  profile.ContentStyle.LanguageLevel,
		ClarityScore:        0.85,
		Appropriateness:     0.9,
		AuthenticityScore:   0.88,
		GenerationTime:      50,
		ModelUsed:           "ofa_content_gen",
	}

	return result
}

// calculateFinalStyle calculates final style for request.
func (e *SpeechContentEngine) calculateFinalStyle(profile *models.SpeechContentProfile, req *models.SpeechContentRequest) models.ContentStyle {
	style := profile.ContentStyle

	// Apply formality from request
	switch req.Formality {
	case "formal":
		style.ToneStyle = "professional"
		style.Directness = 0.4
	case "casual":
		style.ToneStyle = "friendly"
		style.Directness = 0.7
	}

	// Apply emotion context
	if req.EmotionContext.CurrentEmotion != "" {
		style = e.applyEmotionToStyle(req.EmotionContext.CurrentEmotion, req.EmotionContext.EmotionIntensity, style)
	}

	// Apply cultural expression
	style.Directness = style.Directness * (1 - profile.CulturalExpression.IndirectExpression*0.5)

	return style
}

// generateContentInternal generates actual content.
func (e *SpeechContentEngine) generateContentInternal(profile *models.SpeechContentProfile, req *models.SpeechContentRequest, style models.ContentStyle) string {
	// Get template if available
	template := e.getTemplate(profile, req)

	// Apply style transformations
	content := e.applyStyleToContent(template, style, req)

	return content
}

// getTemplate gets appropriate template for request.
func (e *SpeechContentEngine) getTemplate(profile *models.SpeechContentProfile, req *models.SpeechContentRequest) string {
	templates := profile.ContentTemplates

	switch req.ContentType {
	case "greeting":
		if t, ok := templates.GreetingTemplates[req.Context.Scene]; ok {
			return t
		}
		return e.getDefaultGreeting(req.Context.Scene, req.Formality)
	case "closing":
		if t, ok := templates.ClosingTemplates[req.Context.Scene]; ok {
			return t
		}
		return e.getDefaultClosing(req.Context.Scene, req.Formality)
	case "apology":
		if t, ok := templates.ApologyTemplates["moderate"]; ok {
			return t
		}
		return "抱歉给您带来了不便。"
	case "gratitude":
		if t, ok := templates.GratitudeTemplates["normal"]; ok {
			return t
		}
		return "感谢您的帮助。"
	default:
		return ""
	}
}

// applyStyleToContent applies style to content.
func (e *SpeechContentEngine) applyStyleToContent(content string, style models.ContentStyle, req *models.SpeechContentRequest) string {
	// In production, this would use NLP/LLM to transform content
	// For now, return template content
	if content == "" {
		return e.generateDefaultContent(req, style)
	}
	return content
}

// generateDefaultContent generates default content.
func (e *SpeechContentEngine) generateDefaultContent(req *models.SpeechContentRequest, style models.ContentStyle) string {
	switch req.ContentType {
	case "greeting":
		return e.getDefaultGreeting(req.Context.Scene, req.Formality)
	case "response":
		return "好的，我来帮您处理。"
	case "explanation":
		return "让我为您解释一下。"
	default:
		return "好的。"
	}
}

// getDefaultGreeting returns default greeting.
func (e *SpeechContentEngine) getDefaultGreeting(scene string, formality string) string {
	switch scene {
	case "meeting":
		if formality == "formal" {
			return "您好，很高兴见到您。"
		}
		return "你好，很高兴见到你。"
	case "casual":
		return "嗨，你好！"
	case "presentation":
		return "各位好，感谢大家的到来。"
	default:
		return "你好。"
	}
}

// getDefaultClosing returns default closing.
func (e *SpeechContentEngine) getDefaultClosing(scene string, formality string) string {
	switch scene {
	case "meeting":
		if formality == "formal" {
			return "感谢您的时间，期待下次交流。"
		}
		return "谢谢你，下次再聊。"
	case "presentation":
		return "感谢大家的聆听。"
	default:
		return "再见。"
	}
}

// === Decision Context ===

// GetDecisionContext generates content decision context.
func (e *SpeechContentEngine) GetDecisionContext(identityID string, scene string, socialContext string, emotion string) *models.ContentDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.GetProfile(identityID)
	if profile == nil {
		return nil
	}

	ctx := &models.ContentDecisionContext{
		IdentityID: identityID,
		Timestamp:  time.Now(),
	}

	// Generate recommendations
	ctx.RecommendedTone = e.recommendTone(profile, scene, emotion)
	ctx.RecommendedFormality = e.recommendFormality(profile, scene, socialContext)
	ctx.RecommendedDepth = e.recommendDepth(profile, socialContext)
	ctx.RecommendedLength = e.recommendLength(scene)
	ctx.RecommendedDirectness = e.recommendDirectness(profile, socialContext)

	// Generate adaptations
	ctx.SceneAdaptation = e.generateSceneAdaptation(scene, profile)
	ctx.EmotionAdaptation = e.generateEmotionAdaptation(emotion, profile)
	ctx.SocialAdaptation = e.generateSocialAdaptation(socialContext, profile)
	ctx.CulturalAdaptation = e.generateCulturalAdaptation(profile)

	// Generate suggestions
	ctx.OpeningSuggestion = e.suggestOpening(profile, scene, socialContext)
	ctx.ClosingSuggestion = e.suggestClosing(profile, scene, socialContext)
	ctx.KeyTopicsToAvoid = e.suggestTopicsToAvoid(profile)
	ctx.KeyTopicsToInclude = e.suggestTopicsToInclude(profile, scene)

	return ctx
}

// recommendTone recommends tone for context.
func (e *SpeechContentEngine) recommendTone(profile *models.SpeechContentProfile, scene string, emotion string) string {
	baseTone := profile.ContentStyle.ToneStyle

	switch scene {
	case "meeting", "presentation":
		return "professional"
	case "casual":
		return "friendly"
	}

	if emotion == "joy" {
		return "enthusiastic"
	} else if emotion == "sadness" {
		return "gentle"
	}

	return baseTone
}

// recommendFormality recommends formality for context.
func (e *SpeechContentEngine) recommendFormality(profile *models.SpeechContentProfile, scene string, socialContext string) string {
	// Scene takes precedence
	switch scene {
	case "meeting", "presentation":
		return "formal"
	case "casual", "home":
		return "casual"
	}

	// Social context
	switch socialContext {
	case "professional", "formal":
		return "formal"
	case "intimate", "close":
		return "casual"
	}

	return "neutral"
}

// recommendDepth recommends depth for context.
func (e *SpeechContentEngine) recommendDepth(profile *models.SpeechContentProfile, socialContext string) string {
	switch socialContext {
	case "intimate":
		return "deep"
	case "professional":
		return "moderate"
	case "public":
		return "surface"
	}
	return profile.ExpressionDepth.ThinkingDepth
}

// recommendLength recommends length for context.
func (e *SpeechContentEngine) recommendLength(scene string) string {
	switch scene {
	case "meeting":
		return "medium"
	case "presentation":
		return "long"
	case "casual":
		return "short"
	}
	return "medium"
}

// recommendDirectness recommends directness for context.
func (e *SpeechContentEngine) recommendDirectness(profile *models.SpeechContentProfile, socialContext string) float64 {
	directness := profile.ContentStyle.Directness

	// Cultural influence
	directness *= (1 - profile.CulturalExpression.IndirectExpression*0.5)

	// Social context
	switch socialContext {
	case "intimate":
		directness += 0.1
	case "professional":
		directness -= 0.1
	}

	if directness > 1.0 {
		directness = 1.0
	}
	if directness < 0.0 {
		directness = 0.0
	}

	return directness
}

// generateSceneAdaptation generates scene adaptation.
func (e *SpeechContentEngine) generateSceneAdaptation(scene string, profile *models.SpeechContentProfile) models.ContentSceneAdaptation {
	return models.ContentSceneAdaptation{
		Scene:            scene,
		ToneAdjust:       e.recommendTone(profile, scene, ""),
		FormalityAdjust:  e.recommendFormality(profile, scene, ""),
		DepthAdjust:      e.recommendDepth(profile, ""),
		LengthPreference: e.recommendLength(scene),
	}
}

// generateEmotionAdaptation generates emotion adaptation.
func (e *SpeechContentEngine) generateEmotionAdaptation(emotion string, profile *models.SpeechContentProfile) models.ContentEmotionAdaptation {
	adaptation := models.ContentEmotionAdaptation{
		CurrentEmotion: emotion,
	}

	switch emotion {
	case "joy":
		adaptation.EmotionalColoring = "warm"
		adaptation.ExpressionIntensity = 0.7
		adaptation.WordChoice = "positive"
		adaptation.SentenceStyle = "flowing"
	case "sadness":
		adaptation.EmotionalColoring = "cool"
		adaptation.ExpressionIntensity = 0.4
		adaptation.WordChoice = "neutral"
		adaptation.SentenceStyle = "measured"
	case "anger":
		adaptation.EmotionalColoring = "passionate"
		adaptation.ExpressionIntensity = 0.8
		adaptation.WordChoice = "strong"
		adaptation.SentenceStyle = "choppy"
	default:
		adaptation.EmotionalColoring = "neutral"
		adaptation.ExpressionIntensity = 0.5
		adaptation.WordChoice = "neutral"
		adaptation.SentenceStyle = "measured"
	}

	return adaptation
}

// generateSocialAdaptation generates social adaptation.
func (e *SpeechContentEngine) generateSocialAdaptation(socialContext string, profile *models.SpeechContentProfile) models.ContentSocialAdaptation {
	adaptation := models.ContentSocialAdaptation{
		SocialContext: socialContext,
	}

	switch socialContext {
	case "professional":
		adaptation.RespectLevel = 0.7
		adaptation.HonorificUsage = "moderate"
		adaptation.SelfReferenceStyle = "humble"
		adaptation.OtherReferenceStyle = "respectful"
	case "intimate":
		adaptation.RespectLevel = 0.3
		adaptation.HonorificUsage = "light"
		adaptation.SelfReferenceStyle = "neutral"
		adaptation.OtherReferenceStyle = "warm"
	case "public":
		adaptation.RespectLevel = 0.6
		adaptation.HonorificUsage = "light"
		adaptation.SelfReferenceStyle = "confident"
		adaptation.OtherReferenceStyle = "neutral"
	default:
		adaptation.RespectLevel = 0.5
		adaptation.HonorificUsage = "light"
		adaptation.SelfReferenceStyle = "neutral"
		adaptation.OtherReferenceStyle = "neutral"
	}

	return adaptation
}

// generateCulturalAdaptation generates cultural adaptation.
func (e *SpeechContentEngine) generateCulturalAdaptation(profile *models.SpeechContentProfile) models.ContentCulturalAdaptation {
	return models.ContentCulturalAdaptation{
		IndirectnessLevel:    profile.CulturalExpression.IndirectExpression,
		FaceSavingLevel:      profile.CulturalExpression.FaceSaving,
		CollectivistEmphasis: profile.CulturalExpression.CollectivistExpression,
	}
}

// suggestOpening suggests opening phrase.
func (e *SpeechContentEngine) suggestOpening(profile *models.SpeechContentProfile, scene string, socialContext string) string {
	if socialContext == "professional" {
		return "您好，有什么我可以帮助您的吗？"
	}
	return "你好，很高兴见到你。"
}

// suggestClosing suggests closing phrase.
func (e *SpeechContentEngine) suggestClosing(profile *models.SpeechContentProfile, scene string, socialContext string) string {
	if socialContext == "professional" {
		return "感谢您的时间。"
	}
	return "下次再聊！"
}

// suggestTopicsToAvoid suggests topics to avoid.
func (e *SpeechContentEngine) suggestTopicsToAvoid(profile *models.SpeechContentProfile) []string {
	return profile.CulturalExpression.SensitiveTopics
}

// suggestTopicsToInclude suggests topics to include.
func (e *SpeechContentEngine) suggestTopicsToInclude(profile *models.SpeechContentProfile, scene string) []string {
	return []string{}
}

// === Helper methods ===

func (e *SpeechContentEngine) getDefaultContentStyle() models.ContentStyle {
	return models.ContentStyle{
		ToneStyle:         e.config.DefaultTone,
		ToneConsistency:   0.7,
		LanguageLevel:     e.config.DefaultLanguageLevel,
		Directness:        0.5,
		EuphemismUsage:    0.3,
		MetaphorUsage:     0.3,
		HumorTendency:     0.3,
		EmotionalColoring: "neutral",
		EnthusiasmLevel:   0.5,
		PersuasionStyle:   "balanced",
		EvidenceType:      "mixed",
	}
}

func (e *SpeechContentEngine) getDefaultExpressionDepth() models.ExpressionDepth {
	return models.ExpressionDepth{
		ThinkingDepth:         "moderate",
		AbstractionLevel:      "mixed",
		ComplexityLevel:       "moderate",
		SelfDisclosureLevel:   0.5,
		IntimacyThreshold:     0.5,
		VulnerabilityLevel:    0.3,
		ReflectionTendency:    0.5,
		SelfAwarenessLevel:    0.5,
		ProfessionalDepth:     "analytical",
		PersonalDepth:         "moderate",
		PhilosophicalDepth:    "reflective",
	}
}

func (e *SpeechContentEngine) getDefaultCulturalExpression() models.CulturalExpression {
	return models.CulturalExpression{
		HighContextCommunication: false,
		IndirectExpression:       0.3,
		FaceSaving:               0.4,
		RespectLevel:             0.5,
		HierarchyAwareness:       0.4,
		HonorificUsage:           "light",
		TabooAwareness:           0.5,
		SensitiveTopics:          []string{},
		CulturalNuances:          []string{},
		CollectivistExpression:   0.4,
		GroupReferenceUsage:      0.4,
	}
}

func (e *SpeechContentEngine) getDefaultSocialExpression() models.SocialExpression {
	return models.SocialExpression{
		ProfessionalTone:    "collaborative",
		ExpertiseDisplay:    0.5,
		HumilityExpression:  0.5,
		ClassExpression:     "moderate",
		StatusAwareness:     0.4,
		NetworkingStyle:     "balanced",
		RoleConsistency:     0.7,
		RoleAdaptability:    0.6,
		AuthorityExpression: "earned",
		IdentityConfidence:  0.6,
		AuthenticExpression: 0.7,
	}
}

func (e *SpeechContentEngine) getDefaultContentTemplates() models.ContentTemplates {
	return models.ContentTemplates{
		GreetingTemplates: map[string]string{
			"formal":   "您好，很高兴见到您。",
			"casual":   "嗨，你好！",
			"meeting":  "大家好，感谢各位的到来。",
			"default":  "你好。",
		},
		ClosingTemplates: map[string]string{
			"formal":   "感谢您的时间，期待下次交流。",
			"casual":   "下次再聊！",
			"meeting":  "感谢大家的参与。",
			"default":  "再见。",
		},
		ApologyTemplates: map[string]string{
			"light":    "不好意思。",
			"moderate": "抱歉给您带来了不便。",
			"serious":  "非常抱歉，这是我的失误，我会立即改正。",
		},
		GratitudeTemplates: map[string]string{
			"light":    "谢谢。",
			"normal":   "感谢您的帮助。",
			"deep":     "非常感谢您的支持，对我帮助很大。",
		},
		ResponseTemplates: map[string]string{
			"acknowledge": "好的，我明白了。",
			"agree":       "我同意您的看法。",
			"disagree":    "我有不同的想法。",
		},
		CustomTemplates: map[string]string{},
	}
}

func generateRequestID() string {
	return "content_" + time.Now().Format("20060102150405")
}

func countWords(content string) int {
	// Simple word count for Chinese and English
	return len([]rune(content))
}
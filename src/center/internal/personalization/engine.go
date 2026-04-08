// Package personalization provides the Avatar Personalization Engine (v5.4.0).
//
// The PersonalizationEngine manages avatar personalization including
// preferences, evolution, scene adaptation, and style management.
package personalization

import (
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// PersonalizationEngine manages avatar personalization.
type PersonalizationEngine struct {
	mu sync.RWMutex

	// Profiles
	profiles map[string]*models.AvatarPersonalizationProfile

	// Contexts
	contexts map[string]*models.PersonalizationContext

	// Dependencies (interfaces for integration with v4.x systems)
	lifeStageProvider    LifeStageProvider
	emotionProvider      EmotionProvider
	relationshipProvider RelationshipProvider
	cultureProvider      CultureProvider
	socialIdentityProvider SocialIdentityProvider
}

// LifeStageProvider provides life stage information (v4.4.0).
type LifeStageProvider interface {
	GetCurrentStage(identityID string) string
	GetStageProgress(identityID string) float64
}

// EmotionProvider provides emotion information (v4.0.0).
type EmotionProvider interface {
	GetCurrentEmotion(identityID string) string
	GetEmotionIntensity(identityID string) float64
}

// RelationshipProvider provides relationship information (v4.6.0).
type RelationshipProvider interface {
	GetSocialContext(identityID string) string
	GetIntimacyLevel(identityID string, targetID string) float64
}

// CultureProvider provides culture information (v4.3.0).
type CultureProvider interface {
	GetCulturalStyle(identityID string) string
	GetFormalityLevel(identityID string) float64
}

// SocialIdentityProvider provides social identity information (v4.2.0).
type SocialIdentityProvider interface {
	GetCareerStyle(identityID string) string
	GetSocialClass(identityID string) string
}

// NewPersonalizationEngine creates a new PersonalizationEngine.
func NewPersonalizationEngine() *PersonalizationEngine {
	return &PersonalizationEngine{
		profiles: make(map[string]*models.AvatarPersonalizationProfile),
		contexts: make(map[string]*models.PersonalizationContext),
	}
}

// SetLifeStageProvider sets the life stage provider.
func (e *PersonalizationEngine) SetLifeStageProvider(provider LifeStageProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lifeStageProvider = provider
}

// SetEmotionProvider sets the emotion provider.
func (e *PersonalizationEngine) SetEmotionProvider(provider EmotionProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.emotionProvider = provider
}

// SetRelationshipProvider sets the relationship provider.
func (e *PersonalizationEngine) SetRelationshipProvider(provider RelationshipProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.relationshipProvider = provider
}

// SetCultureProvider sets the culture provider.
func (e *PersonalizationEngine) SetCultureProvider(provider CultureProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cultureProvider = provider
}

// SetSocialIdentityProvider sets the social identity provider.
func (e *PersonalizationEngine) SetSocialIdentityProvider(provider SocialIdentityProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.socialIdentityProvider = provider
}

// === Profile Management ===

// GetProfile returns the personalization profile for an identity.
func (e *PersonalizationEngine) GetProfile(identityID string) *models.AvatarPersonalizationProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if profile, ok := e.profiles[identityID]; ok {
		return profile
	}
	return nil
}

// CreateProfile creates a new personalization profile.
func (e *PersonalizationEngine) CreateProfile(identityID string) *models.AvatarPersonalizationProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := &models.AvatarPersonalizationProfile{
		IdentityID: identityID,
		ImagePreferences: models.ImagePreferences{
			PreferredColors:    []string{"blue", "black", "white"},
			ColorIntensity:     "moderate",
			StyleMixingLevel:   0.5,
			StyleExperimentation: 0.3,
			ComfortPriority:    "medium",
			WeatherAdaptation:  true,
			AccessoryFrequency: "occasional",
			GroomingRoutine:    "moderate",
			PresentationEffort: "medium",
			AttentionToDetail:  0.5,
			OccasionAwareness:  0.7,
		},
		ImageEvolution: models.ImageEvolution{
			EvolutionMode:      "gradual",
			EvolutionSpeed:     "moderate",
			EvolutionTendency:  "balanced",
			SeasonalAdaptation: true,
			SpringStyle:        "fresh",
			SummerStyle:        "light",
			AutumnStyle:        "warm",
			WinterStyle:        "cozy",
			TrendFollowingLevel: 0.4,
			TrendAdoptionSpeed: "mainstream",
			TrendFilter:        "selective",
		},
		SceneAdaptationSettings: models.SceneAdaptationSettings{
			AdaptationMode:      "auto",
			AdaptationSpeed:     "gradual",
			AdaptationIntensity: "moderate",
			TransitionStyle:     "gradual",
			TransitionDuration:  5,
			LocationAwareness:   true,
			TimeAwareness:       true,
			CalendarAwareness:   true,
			WeatherAwareness:    true,
			PrivacyMode:         "semi_private",
		},
		StyleManagement: models.StyleManagement{
			RecommendationEnabled: true,
			RecommendationSource:  "hybrid",
			RecommendationFrequency: "weekly",
		},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	e.profiles[identityID] = profile
	return profile
}

// UpdateProfile updates the personalization profile.
func (e *PersonalizationEngine) UpdateProfile(identityID string, profile *models.AvatarPersonalizationProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.Version++
	profile.UpdatedAt = time.Now()
	e.profiles[identityID] = profile
}

// UpdateImagePreferences updates image preferences.
func (e *PersonalizationEngine) UpdateImagePreferences(identityID string, prefs models.ImagePreferences) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.ImagePreferences = prefs
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateImageEvolution updates image evolution settings.
func (e *PersonalizationEngine) UpdateImageEvolution(identityID string, evolution models.ImageEvolution) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.ImageEvolution = evolution
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateSceneAdaptationSettings updates scene adaptation settings.
func (e *PersonalizationEngine) UpdateSceneAdaptationSettings(identityID string, settings models.SceneAdaptationSettings) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.SceneAdaptationSettings = settings
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateStyleManagement updates style management settings.
func (e *PersonalizationEngine) UpdateStyleManagement(identityID string, mgmt models.StyleManagement) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.StyleManagement = mgmt
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// === Scene Adaptation ===

// AddSceneRule adds a scene adaptation rule.
func (e *PersonalizationEngine) AddSceneRule(identityID string, rule models.SceneRule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.SceneAdaptationSettings.SceneRules = append(
			profile.SceneAdaptationSettings.SceneRules, rule)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// GetSceneStyle returns the appropriate style for a scene.
func (e *PersonalizationEngine) GetSceneStyle(identityID string, scene string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if profile, ok := e.profiles[identityID]; ok {
		// Check scene rules
		for _, rule := range profile.SceneAdaptationSettings.SceneRules {
			if rule.SceneName == scene && rule.Enabled {
				return rule.RequiredStyle
			}
		}

		// Check defaults
		switch scene {
		case "work", "meeting":
			return profile.SceneAdaptationSettings.DefaultWorkStyle
		case "home":
			return profile.SceneAdaptationSettings.DefaultHomeStyle
		case "social", "party":
			return profile.SceneAdaptationSettings.DefaultSocialStyle
		case "gym", "sports":
			return profile.SceneAdaptationSettings.DefaultActiveStyle
		}
	}

	return "casual"
}

// AdaptToScene adapts the personalization to a scene.
func (e *PersonalizationEngine) AdaptToScene(identityID string, scene string) *models.SceneRule {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		for _, rule := range profile.SceneAdaptationSettings.SceneRules {
			if rule.SceneName == scene && rule.Enabled {
				return &rule
			}
		}
	}

	// Return default rule
	return &models.SceneRule{
		SceneName:     scene,
		RequiredStyle: e.GetSceneStyle(identityID, scene),
	}
}

// === Style Management ===

// AddStyleCollection adds a style collection.
func (e *PersonalizationEngine) AddStyleCollection(identityID string, collection models.StyleCollection) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		collection.CreatedAt = time.Now()
		profile.StyleManagement.StyleCollections = append(
			profile.StyleManagement.StyleCollections, collection)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// AddOutfitRecord adds an outfit record.
func (e *PersonalizationEngine) AddOutfitRecord(identityID string, outfit models.OutfitRecord) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.StyleManagement.OutfitHistory = append(
			profile.StyleManagement.OutfitHistory, outfit)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// AddWardrobeItem adds a wardrobe item.
func (e *PersonalizationEngine) AddWardrobeItem(identityID string, item models.WardrobeItem) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		item.PurchaseDate = time.Now()
		item.IsActive = true
		profile.StyleManagement.WardrobeItems = append(
			profile.StyleManagement.WardrobeItems, item)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateWardrobeStats updates wardrobe statistics.
func (e *PersonalizationEngine) UpdateWardrobeStats(identityID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		stats := models.WardrobeStats{
			ByCategory: make(map[string]int),
			ByColor:    make(map[string]int),
			BySeason:   make(map[string]int),
			ByStyle:    make(map[string]int),
		}

		totalWear := 0
		for _, item := range profile.StyleManagement.WardrobeItems {
			if item.IsActive {
				stats.TotalItems++
				stats.ByCategory[item.Category]++
				stats.ByColor[item.Color]++
				stats.BySeason[item.Season]++
				stats.ByStyle[item.Style]++
				totalWear += item.WearCount
				stats.TotalValue += item.Price

				if item.WearCount == 0 {
					stats.UnusedItems++
				}
				if item.IsFavorite {
					stats.FavoriteCount++
				}
			}
		}

		if stats.TotalItems > 0 {
			stats.AverageWearCount = float64(totalWear) / float64(stats.TotalItems)
		}

		profile.StyleManagement.WardrobeStats = stats
	}
}

// === Image Evolution ===

// RecordStyleChange records a style change in history.
func (e *PersonalizationEngine) RecordStyleChange(identityID string, styleName string, context string, satisfaction float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		record := models.StyleRecord{
			Timestamp:    time.Now(),
			StyleName:    styleName,
			Description:  context,
			Context:      context,
			Satisfaction: satisfaction,
		}
		profile.ImageEvolution.StyleHistory = append(
			profile.ImageEvolution.StyleHistory, record)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// AddStyleMilestone adds a style milestone.
func (e *PersonalizationEngine) AddStyleMilestone(identityID string, milestone models.StyleMilestone) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		milestone.Date = time.Now()
		profile.ImageEvolution.StyleMilestones = append(
			profile.ImageEvolution.StyleMilestones, milestone)
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// GetRecommendedEvolution returns recommended style evolution.
func (e *PersonalizationEngine) GetRecommendedEvolution(identityID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if profile, ok := e.profiles[identityID]; ok {
		// Consider life stage
		if e.lifeStageProvider != nil {
			stage := e.lifeStageProvider.GetCurrentStage(identityID)
			if rule, ok := profile.ImageEvolution.LifeStageStyleRules[stage]; ok {
				return rule.StyleTone
			}
		}

		// Consider trend following level
		if profile.ImageEvolution.TrendFollowingLevel > 0.6 {
			return "experimental"
		} else if profile.ImageEvolution.TrendFollowingLevel > 0.3 {
			return "moderate"
		}
		return "conservative"
	}

	return "stable"
}

// === Decision Context ===

// GetDecisionContext returns the personalization decision context.
func (e *PersonalizationEngine) GetDecisionContext(identityID string) *models.PersonalizationContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	context := &models.PersonalizationContext{
		IdentityID: identityID,
		Timestamp:  time.Now(),
	}

	if profile, ok := e.profiles[identityID]; ok {
		// Generate style recommendation
		context.RecommendedStyle = e.generateStyleRecommendation(identityID, profile)

		// Generate outfit recommendation
		context.RecommendedOutfit = e.generateOutfitRecommendation(identityID, profile)

		// Calculate scores
		context.StyleScore = e.calculateStyleScore(profile)
		context.ConsistencyScore = e.calculateConsistencyScore(profile)
		context.VersatilityScore = e.calculateVersatilityScore(profile)
		context.AuthenticityScore = e.calculateAuthenticityScore(profile)

		// Evolution status
		context.EvolutionStage = profile.ImageEvolution.EvolutionMode
		context.EvolutionReadiness = e.calculateEvolutionReadiness(profile)
		context.NextEvolutionPreview = e.getEvolutionPreview(identityID, profile)

		// Generate suggestions
		context.StyleSuggestions = e.generateStyleSuggestions(profile)
		context.ImprovementTips = e.generateImprovementTips(profile)
		context.TrendAlerts = e.generateTrendAlerts(profile)
	}

	e.contexts[identityID] = context
	return context
}

// generateStyleRecommendation generates a style recommendation.
func (e *PersonalizationEngine) generateStyleRecommendation(identityID string, profile *models.AvatarPersonalizationProfile) models.StyleRecommendation {
	rec := models.StyleRecommendation{
		Confidence: 0.7,
		Reason:     "Based on your preferences and current context",
	}

	// Consider preferred styles
	if len(profile.ImagePreferences.PreferredStyles) > 0 {
		rec.StyleName = profile.ImagePreferences.PreferredStyles[0]
	}

	// Consider season
	month := time.Now().Month()
	switch {
	case month >= 3 && month <= 5:
		rec.Season = "spring"
		rec.VibeMatch = 0.8
	case month >= 6 && month <= 8:
		rec.Season = "summer"
		rec.VibeMatch = 0.8
	case month >= 9 && month <= 11:
		rec.Season = "autumn"
		rec.VibeMatch = 0.8
	default:
		rec.Season = "winter"
		rec.VibeMatch = 0.8
	}

	// Set color palette
	rec.ColorPalette = profile.ImagePreferences.PreferredColors

	return rec
}

// generateOutfitRecommendation generates an outfit recommendation.
func (e *PersonalizationEngine) generateOutfitRecommendation(identityID string, profile *models.AvatarPersonalizationProfile) models.OutfitRecommendation {
	rec := models.OutfitRecommendation{
		OutfitID:      "rec-" + time.Now().Format("20060102150405"),
		Confidence:    0.65,
		Reason:        "Recommended based on your wardrobe and preferences",
		StyleCategory: "casual",
	}

	// Find suitable items from wardrobe
	var suitableItems []string
	for _, item := range profile.StyleManagement.WardrobeItems {
		if item.IsActive {
			suitableItems = append(suitableItems, item.ItemID)
		}
	}
	rec.Items = suitableItems

	return rec
}

// calculateStyleScore calculates the style score.
func (e *PersonalizationEngine) calculateStyleScore(profile *models.AvatarPersonalizationProfile) float64 {
	score := 0.5

	// Consider presentation effort
	switch profile.ImagePreferences.PresentationEffort {
	case "high":
		score += 0.2
	case "medium":
		score += 0.1
	}

	// Consider occasion awareness
	score += profile.ImagePreferences.OccasionAwareness * 0.2

	// Consider consistency
	if len(profile.ImageEvolution.StyleHistory) > 0 {
		var avgSatisfaction float64
		for _, record := range profile.ImageEvolution.StyleHistory {
			avgSatisfaction += record.Satisfaction
		}
		avgSatisfaction /= float64(len(profile.ImageEvolution.StyleHistory))
		score += avgSatisfaction * 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateConsistencyScore calculates the style consistency score.
func (e *PersonalizationEngine) calculateConsistencyScore(profile *models.AvatarPersonalizationProfile) float64 {
	if len(profile.ImageEvolution.StyleHistory) < 2 {
		return 0.5
	}

	// Analyze style history consistency
	consistency := 0.5

	if profile.ImageEvolution.EvolutionMode == "stable" {
		consistency += 0.3
	} else if profile.ImageEvolution.EvolutionMode == "gradual" {
		consistency += 0.1
	}

	return consistency
}

// calculateVersatilityScore calculates the style versatility score.
func (e *PersonalizationEngine) calculateVersatilityScore(profile *models.AvatarPersonalizationProfile) float64 {
	versatility := 0.0

	// Count different styles
	versatility += float64(len(profile.ImagePreferences.PreferredStyles)) * 0.1

	// Count style collections
	versatility += float64(len(profile.StyleManagement.StyleCollections)) * 0.1

	// Consider style mixing level
	versatility += profile.ImagePreferences.StyleMixingLevel * 0.3

	// Consider style experimentation
	versatility += profile.ImagePreferences.StyleExperimentation * 0.2

	if versatility > 1.0 {
		versatility = 1.0
	}

	return versatility
}

// calculateAuthenticityScore calculates the style authenticity score.
func (e *PersonalizationEngine) calculateAuthenticityScore(profile *models.AvatarPersonalizationProfile) float64 {
	authenticity := 0.5

	// Lower trend following = higher authenticity
	authenticity += (1.0 - profile.ImageEvolution.TrendFollowingLevel) * 0.2

	// Core style elements defined = higher authenticity
	authenticity += float64(len(profile.ImageEvolution.CoreStyleElements)) * 0.05

	if authenticity > 1.0 {
		authenticity = 1.0
	}

	return authenticity
}

// calculateEvolutionReadiness calculates evolution readiness.
func (e *PersonalizationEngine) calculateEvolutionReadiness(profile *models.AvatarPersonalizationProfile) float64 {
	readiness := 0.5

	// Higher experimentation = higher readiness
	readiness += profile.ImagePreferences.StyleExperimentation * 0.3

	// Higher trend following = higher readiness
	readiness += profile.ImageEvolution.TrendFollowingLevel * 0.2

	if readiness > 1.0 {
		readiness = 1.0
	}

	return readiness
}

// getEvolutionPreview returns a preview of the next evolution.
func (e *PersonalizationEngine) getEvolutionPreview(identityID string, profile *models.AvatarPersonalizationProfile) string {
	if e.lifeStageProvider != nil {
		stage := e.lifeStageProvider.GetCurrentStage(identityID)
		if rule, ok := profile.ImageEvolution.LifeStageStyleRules[stage]; ok {
			return "Transitioning to " + rule.StyleTone + " style"
		}
	}

	// Seasonal preview
	month := time.Now().Month()
	switch {
	case month >= 3 && month <= 5:
		return profile.ImageEvolution.SpringStyle + " style for spring"
	case month >= 6 && month <= 8:
		return profile.ImageEvolution.SummerStyle + " style for summer"
	case month >= 9 && month <= 11:
		return profile.ImageEvolution.AutumnStyle + " style for autumn"
	default:
		return profile.ImageEvolution.WinterStyle + " style for winter"
	}
}

// generateStyleSuggestions generates style suggestions.
func (e *PersonalizationEngine) generateStyleSuggestions(profile *models.AvatarPersonalizationProfile) []models.StyleSuggestion {
	var suggestions []models.StyleSuggestion

	// Suggest based on wardrobe gaps
	stats := profile.StyleManagement.WardrobeStats
	if stats.ByCategory["accessory"] < 3 {
		suggestions = append(suggestions, models.StyleSuggestion{
			SuggestionType: "add",
			TargetArea:     "accessory",
			Suggestion:     "Consider adding more accessories to diversify your looks",
			Priority:       3,
			Effort:         "easy",
			Impact:         0.6,
		})
	}

	// Suggest based on color variety
	if len(stats.ByColor) < 5 {
		suggestions = append(suggestions, models.StyleSuggestion{
			SuggestionType: "try",
			TargetArea:     "color",
			Suggestion:     "Experiment with new colors outside your comfort zone",
			Priority:       2,
			Effort:         "moderate",
			Impact:         0.7,
		})
	}

	return suggestions
}

// generateImprovementTips generates improvement tips.
func (e *PersonalizationEngine) generateImprovementTips(profile *models.AvatarPersonalizationProfile) []string {
	var tips []string

	// Based on style score
	if profile.ImagePreferences.AttentionToDetail < 0.5 {
		tips = append(tips, "Pay more attention to details for a polished look")
	}

	// Based on wardrobe
	if profile.StyleManagement.WardrobeStats.UnusedItems > 0 {
		tips = append(tips, "Consider wearing items that haven't been used recently")
	}

	// Based on season
	tips = append(tips, "Update your wardrobe for the current season")

	return tips
}

// generateTrendAlerts generates trend alerts.
func (e *PersonalizationEngine) generateTrendAlerts(profile *models.AvatarPersonalizationProfile) []models.TrendAlert {
	var alerts []models.TrendAlert

	// Only show alerts if trend following is enabled
	if profile.ImageEvolution.TrendFollowingLevel > 0.3 {
		alerts = append(alerts, models.TrendAlert{
			TrendName:     "Minimalist Elegance",
			Description:   "Clean lines and neutral colors are trending",
			Relevance:     profile.ImageEvolution.TrendFollowingLevel,
			AdoptionLevel: "mainstream",
			Category:      "style",
			RiskLevel:     "safe",
		})
	}

	return alerts
}
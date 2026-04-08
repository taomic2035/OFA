// Package models defines the AvatarPersonalization (形象个性化) models for v5.4.0.
//
// AvatarPersonalization represents the personalization system for avatar,
// including preferences, evolution, scene adaptation, and style management.
package models

import (
	"encoding/json"
	"time"
)

// AvatarPersonalizationProfile represents the complete personalization configuration.
type AvatarPersonalizationProfile struct {
	IdentityID string `json:"identity_id"`

	// Image preferences
	ImagePreferences ImagePreferences `json:"image_preferences"`

	// Image evolution settings
	ImageEvolution ImageEvolution `json:"image_evolution"`

	// Scene adaptation settings
	SceneAdaptationSettings SceneAdaptationSettings `json:"scene_adaptation_settings"`

	// Style management
	StyleManagement StyleManagement `json:"style_management"`

	// Personalization context
	PersonalizationContext PersonalizationContext `json:"personalization_context"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ImagePreferences defines detailed image preferences.
type ImagePreferences struct {
	// Color preferences
	PreferredColors    []string `json:"preferred_colors"`    // favorite colors
	AvoidedColors      []string `json:"avoided_colors"`      // colors to avoid
	ColorHarmonyStyle  string   `json:"color_harmony_style"` // complementary, analogous, monochromatic
	ColorIntensity     string   `json:"color_intensity"`     // muted, moderate, vibrant

	// Style preferences
	PreferredStyles    []string `json:"preferred_styles"`    // casual, formal, sporty, elegant
	AvoidedStyles      []string `json:"avoided_styles"`      // styles to avoid
	StyleMixingLevel   float64   `json:"style_mixing_level"`  // 0-1, tendency to mix styles
	StyleExperimentation float64 `json:"style_experimentation"` // 0-1, openness to new styles

	// Comfort preferences
	ComfortPriority    string `json:"comfort_priority"`    // low, medium, high
	FabricPreferences  []string `json:"fabric_preferences"` // cotton, silk, wool, synthetic
	FitPreferences     string `json:"fit_preferences"`     // tight, fitted, relaxed, loose
	WeatherAdaptation  bool    `json:"weather_adaptation"`  // auto-adjust for weather

	// Brand preferences
	FavoriteBrands     []string `json:"favorite_brands"`
	AvoidedBrands      []string `json:"avoided_brands"`
	BrandLoyalty       float64 `json:"brand_loyalty"`       // 0-1, loyalty to brands
	LocalBrandSupport  float64 `json:"local_brand_support"` // 0-1, preference for local brands

	// Accessory preferences
	AccessoryFrequency string   `json:"accessory_frequency"` // rare, occasional, frequent, always
	AccessoryTypes     []string `json:"accessory_types"`
	JewelryStyle       string   `json:"jewelry_style"`       // none, minimal, moderate, elaborate
	WatchStyle         string   `json:"watch_style"`         // none, classic, smart, sport, luxury

	// Grooming preferences
	GroomingRoutine    string `json:"grooming_routine"`    // minimal, moderate, elaborate
	HairProductUse     string `json:"hair_product_use"`    // none, minimal, moderate, extensive
	SkincareRoutine    string `json:"skincare_routine"`    // none, basic, comprehensive
	FragranceUse       string `json:"fragrance_use"`       // none, occasional, daily

	// Presentation preferences
	PresentationEffort string `json:"presentation_effort"` // low, medium, high
	AttentionToDetail  float64 `json:"attention_to_detail"` // 0-1, detail consciousness
	OccasionAwareness  float64 `json:"occasion_awareness"`  // 0-1, awareness of appropriate dress
}

// ImageEvolution defines how the image evolves over time.
type ImageEvolution struct {
	// Evolution mode
	EvolutionMode       string `json:"evolution_mode"`       // stable, gradual, dynamic, experimental
	EvolutionSpeed      string `json:"evolution_speed"`      // slow, moderate, fast
	EvolutionTendency   string `json:"evolution_tendency"`   // conservative, balanced, progressive

	// Evolution triggers
	EvolutionTriggers   []EvolutionTrigger `json:"evolution_triggers"`

	// Style history
	StyleHistory        []StyleRecord `json:"style_history"`
	StyleMilestones     []StyleMilestone `json:"style_milestones"`

	// Evolution constraints
	CoreStyleElements   []string `json:"core_style_elements"`   // elements that never change
	FlexibleElements    []string `json:"flexible_elements"`     // elements open to change
	ExperimentalZone    []string `json:"experimental_zone"`     // elements for experimentation

	// Seasonal evolution
	SeasonalAdaptation  bool     `json:"seasonal_adaptation"`
	SpringStyle         string   `json:"spring_style"`
	SummerStyle         string   `json:"summer_style"`
	AutumnStyle         string   `json:"autumn_style"`
	WinterStyle         string   `json:"winter_style"`

	// Life stage evolution (linked to v4.4.0)
	LifeStageStyleRules map[string]StageStyleRule `json:"life_stage_style_rules"`

	// Fashion trend awareness
	TrendFollowingLevel float64 `json:"trend_following_level"` // 0-1
	TrendAdoptionSpeed  string `json:"trend_adoption_speed"`  // early, mainstream, late, never
	TrendFilter         string `json:"trend_filter"`          // all, curated, selective, conservative
}

// EvolutionTrigger defines what triggers image evolution.
type EvolutionTrigger struct {
	TriggerType   string `json:"trigger_type"`   // life_event, season, trend, relationship, career
	TriggerName   string `json:"trigger_name"`
	StyleChange   string `json:"style_change"`   // expected style change
	Priority      int    `json:"priority"`       // 1-10, higher = more important
	AutoApply     bool   `json:"auto_apply"`     // auto-apply change
}

// StyleRecord represents a historical style record.
type StyleRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	StyleName     string    `json:"style_name"`
	Description   string    `json:"description"`
	Context       string    `json:"context"`       // why this style was adopted
	Satisfaction  float64    `json:"satisfaction"`  // 0-1, satisfaction score
}

// StyleMilestone represents a significant style milestone.
type StyleMilestone struct {
	Date          time.Time `json:"date"`
	MilestoneName string    `json:"milestone_name"`
	Description   string    `json:"description"`
	BeforeStyle   string    `json:"before_style"`
	AfterStyle    string    `json:"after_style"`
	Significance  float64    `json:"significance"` // 0-1
}

// StageStyleRule defines style rules for a life stage.
type StageStyleRule struct {
	StageName       string   `json:"stage_name"`
	AllowedStyles   []string `json:"allowed_styles"`
	ForbiddenStyles []string `json:"forbidden_styles"`
	StyleTone       string   `json:"style_tone"`       // professional, casual, elegant
	ComplexityLevel string   `json:"complexity_level"` // simple, moderate, complex
}

// SceneAdaptationSettings defines how avatar adapts to different scenes.
type SceneAdaptationSettings struct {
	// Scene adaptation mode
	AdaptationMode      string `json:"adaptation_mode"`      // auto, manual, hybrid
	AdaptationSpeed     string `json:"adaptation_speed"`     // instant, gradual, scheduled
	AdaptationIntensity string `json:"adaptation_intensity"` // subtle, moderate, dramatic

	// Scene rules
	SceneRules          []SceneRule `json:"scene_rules"`

	// Default adaptations
	DefaultWorkStyle    string `json:"default_work_style"`
	DefaultHomeStyle    string `json:"default_home_style"`
	DefaultSocialStyle  string `json:"default_social_style"`
	DefaultActiveStyle  string `json:"default_active_style"`

	// Transition settings
	TransitionStyle     string `json:"transition_style"`     // seamless, gradual, distinct
	TransitionDuration  int    `json:"transition_duration"`  // minutes
	TransitionAnimation string `json:"transition_animation"` // fade, morph, instant

	// Context awareness
	LocationAwareness   bool `json:"location_awareness"`
	TimeAwareness       bool `json:"time_awareness"`
	CalendarAwareness   bool `json:"calendar_awareness"`
	WeatherAwareness    bool `json:"weather_awareness"`

	// Privacy settings
	PrivacyMode         string `json:"privacy_mode"`         // public, semi_private, private
	StylePrivacyLevel   float64 `json:"style_privacy_level"`  // 0-1, how much to reveal
}

// SceneRule defines style adaptation for a specific scene.
type SceneRule struct {
	SceneName        string   `json:"scene_name"`        // meeting, date, gym, party, etc.
	SceneType        string   `json:"scene_type"`        // work, social, personal, activity
	RequiredStyle    string   `json:"required_style"`    // expected style
	AllowedVariations []string `json:"allowed_variations"`
	ForbiddenElements []string `json:"forbidden_elements"`
	AccessoryRule    string   `json:"accessory_rule"`    // none, minimal, moderate, any
	GroomingRule     string   `json:"grooming_rule"`     // natural, polished, elaborate
	ExpressionRule   string   `json:"expression_rule"`   // professional, warm, neutral, expressive
	PostureRule      string   `json:"posture_rule"`      // formal, relaxed, confident
	Priority         int      `json:"priority"`          // 1-10
	Enabled          bool     `json:"enabled"`
}

// StyleManagement defines style collection and management.
type StyleManagement struct {
	// Style collections
	StyleCollections   []StyleCollection `json:"style_collections"`
	FavoriteOutfits    []OutfitRecord    `json:"favorite_outfits"`
	OutfitHistory      []OutfitRecord    `json:"outfit_history"`

	// Style templates
	StyleTemplates     []StyleTemplate   `json:"style_templates"`
	DefaultTemplate     string            `json:"default_template"`

	// Style rules
	GlobalStyleRules   []GlobalStyleRule `json:"global_style_rules"`
	ColorRules         []ColorRule       `json:"color_rules"`
	PatternRules       []PatternRule     `json:"pattern_rules"`

	// Wardrobe management
	WardrobeItems      []WardrobeItem    `json:"wardrobe_items"`
	WardrobeStats      WardrobeStats     `json:"wardrobe_stats"`

	// Style recommendations
	RecommendationEnabled bool `json:"recommendation_enabled"`
	RecommendationSource  string `json:"recommendation_source"` // ai, manual, hybrid
	RecommendationFrequency string `json:"recommendation_frequency"` // daily, weekly, occasional
}

// StyleCollection represents a collection of styles.
type StyleCollection struct {
	CollectionID   string   `json:"collection_id"`
	CollectionName string   `json:"collection_name"`
	Description    string   `json:"description"`
	Styles         []string `json:"styles"`
	Occasions      []string `json:"occasions"`
	Seasons        []string `json:"seasons"`        // spring, summer, autumn, winter, all
	IsActive       bool     `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// OutfitRecord represents a specific outfit record.
type OutfitRecord struct {
	OutfitID      string    `json:"outfit_id"`
	OutfitName    string    `json:"outfit_name"`
	Description   string    `json:"description"`
	Items         []string  `json:"items"`         // item IDs
	StyleCategory string    `json:"style_category"`
	Occasion      string    `json:"occasion"`
	Season        string    `json:"season"`
	Rating        float64    `json:"rating"`        // 0-1, user rating
	WearCount     int       `json:"wear_count"`
	LastWorn      time.Time `json:"last_worn"`
	IsFavorite    bool      `json:"is_favorite"`
}

// StyleTemplate represents a reusable style template.
type StyleTemplate struct {
	TemplateID      string   `json:"template_id"`
	TemplateName    string   `json:"template_name"`
	Description     string   `json:"description"`
	StyleBase       string   `json:"style_base"`       // base style
	Variations      []string `json:"variations"`       // allowed variations
	RequiredItems   []string `json:"required_items"`   // must-have items
	OptionalItems   []string `json:"optional_items"`
	ColorScheme     []string `json:"color_scheme"`
	AccessorySet    []string `json:"accessory_set"`
	Occasions       []string `json:"occasions"`
	IsDefault       bool     `json:"is_default"`
}

// GlobalStyleRule defines a global style rule.
type GlobalStyleRule struct {
	RuleID       string `json:"rule_id"`
	RuleName     string `json:"rule_name"`
	RuleType     string `json:"rule_type"`     // allow, avoid, prefer, require
	TargetType   string `json:"target_type"`   // color, pattern, style, item
	TargetValue  string `json:"target_value"`
	Condition    string `json:"condition"`     // always, occasion, season
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
}

// ColorRule defines color combination rules.
type ColorRule struct {
	RuleID       string   `json:"rule_id"`
	BaseColor    string   `json:"base_color"`
	Matches      []string `json:"matches"`      // matching colors
	Contrasts    []string `json:"contrasts"`    // contrasting colors
	Avoids       []string `json:"avoids"`       // colors to avoid
	RuleStrength float64   `json:"rule_strength"` // 0-1, how strongly to follow
}

// PatternRule defines pattern combination rules.
type PatternRule struct {
	RuleID        string   `json:"rule_id"`
	PatternType   string   `json:"pattern_type"`   // solid, striped, plaid, floral, etc.
	MixesWellWith []string `json:"mixes_well_with"`
	AvoidMixing   []string `json:"avoid_mixing"`
	OccasionFit   []string `json:"occasion_fit"`   // appropriate occasions
	MaxOccurrence int      `json:"max_occurrence"` // max items with this pattern
}

// WardrobeItem represents an item in the wardrobe.
type WardrobeItem struct {
	ItemID          string    `json:"item_id"`
	ItemName        string    `json:"item_name"`
	Category        string    `json:"category"`        // top, bottom, dress, outerwear, shoes, accessory
	Subcategory     string    `json:"subcategory"`
	Color           string    `json:"color"`
	Pattern         string    `json:"pattern"`
	Style           string    `json:"style"`
	Brand           string    `json:"brand"`
	Size            string    `json:"size"`
	Material        string    `json:"material"`
	Season          string    `json:"season"`
	Occasions       []string  `json:"occasions"`
	PurchaseDate    time.Time `json:"purchase_date"`
	Price           float64    `json:"price"`
	WearCount       int       `json:"wear_count"`
	LastWorn        time.Time `json:"last_worn"`
	Condition       string    `json:"condition"`       // new, good, fair, poor
	IsFavorite      bool      `json:"is_favorite"`
	IsActive        bool      `json:"is_active"`
	ImageURL        string    `json:"image_url"`
}

// WardrobeStats represents wardrobe statistics.
type WardrobeStats struct {
	TotalItems       int            `json:"total_items"`
	ByCategory       map[string]int `json:"by_category"`
	ByColor          map[string]int `json:"by_color"`
	BySeason         map[string]int `json:"by_season"`
	ByStyle          map[string]int `json:"by_style"`
	AverageWearCount float64         `json:"average_wear_count"`
	UnusedItems      int            `json:"unused_items"`
	FavoriteCount    int            `json:"favorite_count"`
	TotalValue       float64         `json:"total_value"`
}

// PersonalizationContext provides personalization decision context.
type PersonalizationContext struct {
	IdentityID string `json:"identity_id"`

	// Current recommendations
	RecommendedStyle      StyleRecommendation `json:"recommended_style"`
	RecommendedOutfit     OutfitRecommendation `json:"recommended_outfit"`
	RecommendedAccessories []string `json:"recommended_accessories"`

	// Scene adaptation
	CurrentScene          string `json:"current_scene"`
	SceneStyleMatch       float64 `json:"scene_style_match"` // 0-1
	SceneAdaptationNeeded bool   `json:"scene_adaptation_needed"`

	// Style analysis
	StyleScore            float64 `json:"style_score"`            // 0-1, current style score
	ConsistencyScore      float64 `json:"consistency_score"`      // 0-1
	VersatilityScore      float64 `json:"versatility_score"`      // 0-1
	AuthenticityScore     float64 `json:"authenticity_score"`     // 0-1

	// Evolution status
	EvolutionStage        string `json:"evolution_stage"`
	NextEvolutionPreview  string `json:"next_evolution_preview"`
	EvolutionReadiness    float64 `json:"evolution_readiness"` // 0-1

	// Suggestions
	StyleSuggestions      []StyleSuggestion `json:"style_suggestions"`
	ImprovementTips       []string `json:"improvement_tips"`
	TrendAlerts           []TrendAlert `json:"trend_alerts"`

	// Timestamps
	Timestamp             time.Time `json:"timestamp"`
}

// StyleRecommendation represents a style recommendation.
type StyleRecommendation struct {
	StyleName        string   `json:"style_name"`
	Confidence       float64   `json:"confidence"`       // 0-1
	Reason           string   `json:"reason"`
	ColorPalette     []string `json:"color_palette"`
	KeyPieces        []string `json:"key_pieces"`
	Occasion         string   `json:"occasion"`
	Season           string   `json:"season"`
	Weather          string   `json:"weather"`
	VibeMatch        float64   `json:"vibe_match"`       // 0-1, matches user's vibe
}

// OutfitRecommendation represents an outfit recommendation.
type OutfitRecommendation struct {
	OutfitID         string   `json:"outfit_id"`
	OutfitName       string   `json:"outfit_name"`
	Items            []string `json:"items"`
	StyleCategory    string   `json:"style_category"`
	Occasion         string   `json:"occasion"`
	Weather          string   `json:"weather"`
	Confidence       float64   `json:"confidence"`
	Reason           string   `json:"reason"`
	Alternatives     []string `json:"alternatives"` // alternative outfit IDs
}

// StyleSuggestion represents a style improvement suggestion.
type StyleSuggestion struct {
	SuggestionType   string `json:"suggestion_type"`   // add, remove, replace, try
	TargetArea       string `json:"target_area"`       // color, style, accessory, grooming
	Suggestion       string `json:"suggestion"`
	Priority         int    `json:"priority"`
	Effort           string `json:"effort"`           // easy, moderate, challenging
	Impact           float64 `json:"impact"`           // 0-1, expected impact
}

// TrendAlert represents a fashion trend alert.
type TrendAlert struct {
	TrendName        string   `json:"trend_name"`
	Description      string   `json:"description"`
	Relevance        float64   `json:"relevance"`        // 0-1, relevance to user
	AdoptionLevel    string   `json:"adoption_level"`   // emerging, trending, mainstream, fading
	Category         string   `json:"category"`         // color, style, pattern, item
	SuggestedItems   []string `json:"suggested_items"`
	HowToAdopt       string   `json:"how_to_adopt"`
	RiskLevel        string   `json:"risk_level"`       // safe, moderate, bold
}

// ToJSON converts AvatarPersonalizationProfile to JSON string.
func (p *AvatarPersonalizationProfile) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// FromJSON parses AvatarPersonalizationProfile from JSON string.
func (p *AvatarPersonalizationProfile) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), p)
}

// ToJSON converts PersonalizationContext to JSON string.
func (c *PersonalizationContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses PersonalizationContext from JSON string.
func (c *PersonalizationContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}
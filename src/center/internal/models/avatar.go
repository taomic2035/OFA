// Package models defines the Avatar (外在形象) models for v5.0.0.
//
// Avatar represents the external appearance of the digital person,
// mapping internal soul characteristics (v4.x) to external presentation.
package models

import (
	"encoding/json"
	"time"
)

// Avatar represents the complete external appearance of the digital person.
type Avatar struct {
	IdentityID string `json:"identity_id"`

	// Facial features
	FacialFeatures FacialFeatures `json:"facial_features"`

	// Body features
	BodyFeatures BodyFeatures `json:"body_features"`

	// Age appearance (linked to LifeStage v4.4.0)
	AgeAppearance AgeAppearance `json:"age_appearance"`

	// Style preferences (linked to SocialIdentity v4.2.0)
	StylePreferences StylePreferences `json:"style_preferences"`

	// 3D model reference
	Model3D Model3DReference `json:"model_3d"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FacialFeatures defines the facial characteristics.
type FacialFeatures struct {
	// Face shape
	FaceShape string `json:"face_shape"` // oval, round, square, heart, oblong

	// Eyes
	EyeShape  string `json:"eye_shape"`  // almond, round, hooded, upturned, downturned
	EyeColor  string `json:"eye_color"`  // brown, black, blue, green, hazel
	EyeSize   string `json:"eye_size"`   // small, medium, large

	// Nose
	NoseShape   string `json:"nose_shape"`   // straight, curved, wide, narrow
	NoseSize    string `json:"nose_size"`    // small, medium, large
	NoseBridge  string `json:"nose_bridge"`  // low, medium, high

	// Lips
	LipShape    string `json:"lip_shape"`    // thin, medium, full
	LipColor    string `json:"lip_color"`    // natural, pink, coral, red

	// Skin
	SkinTone    string `json:"skin_tone"`    // fair, light, medium, tan, dark
	SkinTexture string `json:"skin_texture"` // smooth, normal, textured

	// Hair
	HairStyle   string `json:"hair_style"`   // short, medium, long, curly, straight, wavy
	HairColor   string `json:"hair_color"`   // black, brown, blonde, red, gray, white
	HairTexture string `json:"hair_texture"` // straight, wavy, curly, coily

	// Facial hair (male)
	FacialHair  string `json:"facial_hair"`  // none, beard, mustache, goatee, stubble

	// Special features
	Freckles    bool   `json:"freckles"`
	Moles       []MolePosition `json:"moles"`
	Scars       []ScarPosition `json:"scars"`

	// Expressiveness (linked to EmotionBehavior v4.5.0)
	Expressiveness float64 `json:"expressiveness"` // 0-1, how expressive the face is
}

// MolePosition represents a mole on the face.
type MolePosition struct {
	X      float64 `json:"x"`       // position relative to face center
	Y      float64 `json:"y"`
	Size   string `json:"size"`    // small, medium, large
	Visible bool  `json:"visible"`
}

// ScarPosition represents a scar on the face.
type ScarPosition struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Length float64 `json:"length"`
	Angle  float64 `json:"angle"`
	Type   string `json:"type"`    // linear, curved, patch
}

// BodyFeatures defines the body characteristics.
type BodyFeatures struct {
	// Basic measurements
	Height      float64 `json:"height"`      // in cm
	Weight      float64 `json:"weight"`      // in kg
	BodyType    string `json:"body_type"`   // slim, average, athletic, curvy, heavy

	// Proportions
	ShoulderWidth float64 `json:"shoulder_width"` // relative proportion 0-1
	HipWidth      float64 `json:"hip_width"`      // relative proportion 0-1
	WaistRatio    float64 `json:"waist_ratio"`    // waist-to-height ratio

	// Posture (linked to RelationshipProfile v4.6.0)
	Posture      string `json:"posture"`      // confident, modest, casual, formal, slouched
	PostureScore float64 `json:"posture_score"` // 0-1, posture quality

	// Movement style (linked to EmotionBehavior v4.5.0)
	MovementStyle string `json:"movement_style"` // graceful, energetic, calm, playful, reserved
	MovementSpeed string `json:"movement_speed"` // slow, moderate, fast

	// Body language tendency
	GestureFrequency float64 `json:"gesture_frequency"` // 0-1, how often uses gestures
	TouchTendency    float64 `json:"touch_tendency"`    // 0-1, tendency to touch others

	// Fitness level
	FitnessLevel  string `json:"fitness_level"`  // low, moderate, high, athletic
	MuscleTone    string `json:"muscle_tone"`    // minimal, moderate, defined, athletic
	Flexibility   string `json:"flexibility"`    // low, moderate, high
}

// AgeAppearance defines how age manifests in appearance.
// Linked to LifeStage v4.4.0.
type AgeAppearance struct {
	// Apparent age (may differ from actual age)
	ApparentAge    int    `json:"apparent_age"`    // perceived age in years
	AgeRange       string `json:"age_range"`       // young, young_adult, adult, middle_aged, senior

	// Aging stage
	AgingStage     string `json:"aging_stage"`     // youthful, prime, mature, senior, elderly

	// Facial maturity indicators
	FacialMaturity float64 `json:"facial_maturity"` // 0-1, facial maturity level
	WrinkleLevel   string `json:"wrinkle_level"`   // none, minimal, moderate, significant
	SkinElasticity string `json:"skin_elasticity"` // high, moderate, low

	// Body maturity indicators
	BodyMaturity   float64 `json:"body_maturity"`   // 0-1, body maturity level
	MetabolismType string `json:"metabolism_type"` // fast, moderate, slow

	// Age-defying factors
	SelfCareLevel  float64 `json:"self_care_level"`  // 0-1, anti-aging effort
	GeneticFactor  float64 `json:"genetic_factor"`   // 0-1, genetic aging tendency
}

// StylePreferences defines the styling and fashion preferences.
// Linked to SocialIdentity v4.2.0 and RegionalCulture v4.3.0.
type StylePreferences struct {
	// Clothing style
	ClothingStyle    string `json:"clothing_style"`    // casual, business, sporty, elegant, bohemian, minimal
	ClothingQuality  string `json:"clothing_quality"`  // budget, mid-range, premium, luxury
	ClothingColors   []string `json:"clothing_colors"`  // preferred colors

	// Accessory style
	AccessoryStyle   string `json:"accessory_style"`   // minimal, moderate, bold, statement
	AccessoryTypes   []string `json:"accessory_types"`  // watch, jewelry, bag, glasses, hat
	JewelryPreference string `json:"jewelry_preference"` // none, subtle, moderate, elaborate

	// Grooming style
	GroomingLevel    string `json:"grooming_level"`    // natural, polished, elaborate
	MakeupStyle      string `json:"makeup_style"`      // none, minimal, moderate, full (for female)
	MakeupFrequency  string `json:"makeup_frequency"`  // never, occasional, daily, always

	// Overall vibe/aesthetic
	OverallVibe      string `json:"overall_vibe"`      // professional, casual, artistic, sporty, elegant
	AestheticTheme   string `json:"aesthetic_theme"`   // classic, modern, vintage, minimalist, maximalist

	// Style evolution (linked to LifeStage)
	StyleEvolution   []StylePhase `json:"style_evolution"` // style history by life stage
	CurrentStylePhase string `json:"current_style_phase"` // current phase name

	// Cultural influence (linked to RegionalCulture v4.3.0)
	CulturalStyle    string `json:"cultural_style"`    // traditional, modern, fusion, western
	RegionalInfluence float64 `json:"regional_influence"` // 0-1, how much region affects style

	// Social class influence (linked to SocialClassProfile v4.2.0)
	ClassStyle       string `json:"class_style"`       // working, middle, upper-middle, upper
	StatusDisplay    float64 `json:"status_display"`    // 0-1, tendency to display status

	// Brand preferences
	BrandPreferences []string `json:"brand_preferences"` // preferred brands
	BrandAwareness   float64 `json:"brand_awareness"`   // 0-1, brand consciousness
}

// StylePhase represents a style phase in life.
type StylePhase struct {
	Stage       string `json:"stage"`       // life stage name
	StyleName   string `json:"style_name"`  // style during this phase
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// Model3DReference defines the 3D model configuration.
type Model3DReference struct {
	// Model identifiers
	ModelID      string `json:"model_id"`      // unique model identifier
	ModelType    string `json:"model_type"`    // custom, preset, generated
	ModelVersion string `json:"model_version"` // version number

	// Model sources
	SourceURL    string `json:"source_url"`    // model download URL (if external)
	SourceFormat string `json:"source_format"` // glb, gltf, fbx, obj

	// Customization level
	CustomizationLevel string `json:"customization_level"` // preset, modified, fully_custom

	// Animation support
	AnimationEnabled  bool     `json:"animation_enabled"`
	AnimationSet      []string `json:"animation_set"`      // available animations
	ExpressionMapping string   `json:"expression_mapping"` // expression to animation mapping

	// Rendering settings
	RenderQuality   string `json:"render_quality"`   // low, medium, high, ultra
	TextureQuality  string `json:"texture_quality"`  // low, medium, high
	LightingProfile string `json:"lighting_profile"` // natural, studio, dramatic

	// Optimization
	OptimizedFor    string `json:"optimized_for"`    // mobile, desktop, vr, ar
	FileSize        float64 `json:"file_size"`        // MB
	PolygonCount    int    `json:"polygon_count"`    // number of polygons
}

// AvatarDecisionContext provides avatar-related decision context.
type AvatarDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// Presentation guidance
	RecommendedStyle      string `json:"recommended_style"`      // for current context
	RecommendedPosture    string `json:"recommended_posture"`    // for social context
	RecommendedExpression string `json:"recommended_expression"` // for emotion state

	// Contextual adaptation
	SceneAdaptation    SceneAdaptation `json:"scene_adaptation"`
	SocialAdaptation   SocialAdaptation `json:"social_adaptation"`
	CulturalAdaptation CulturalAdaptation `json:"cultural_adaptation"`

	// 3D display settings
	DisplaySettings DisplaySettings `json:"display_settings"`

	// Timestamps
	Timestamp time.Time `json:"timestamp"`
}

// SceneAdaptation defines how avatar adapts to different scenes.
type SceneAdaptation struct {
	CurrentScene     string `json:"current_scene"`     // meeting, casual, formal, sport, home
	StyleAdjustment  string `json:"style_adjustment"`  // formal_up, casual_down, neutral
	PostureAdjustment string `json:"posture_adjustment"` // confident, relaxed, formal
	ExpressionRange  string `json:"expression_range"`  // professional, warm, neutral, expressive
	AnimationSet     string `json:"animation_set"`     // idle_meeting, idle_casual, etc.
}

// SocialAdaptation defines how avatar adapts to social context.
// Linked to RelationshipProfile v4.6.0.
type SocialAdaptation struct {
	SocialContext    string `json:"social_context"`    // formal, casual, intimate, professional
	DistanceLevel    string `json:"distance_level"`    // close, moderate, far (avatar distance)
	EyeContactLevel  float64 `json:"eye_contact_level"` // 0-1, how much eye contact
	GestureLevel     float64 `json:"gesture_level"`     // 0-1, how animated
	TouchPermission  string `json:"touch_permission"`  // none, handshake, hug, kiss

	// Linked to AttachmentStyle
	IntimacyDisplay  float64 `json:"intimacy_display"`  // 0-1, how much to show intimacy
	TrustDisplay     float64 `json:"trust_display"`     // 0-1, how much to show trust
}

// AvatarCulturalAdaptation defines how avatar adapts to cultural context.
// Linked to RegionalCulture v4.3.0.
type AvatarCulturalAdaptation struct {
	CulturalContext  string `json:"cultural_context"`  // local, international, multicultural
	FormalityLevel   float64 `json:"formality_level"`   // 0-1, formality degree
	ModestyLevel     float64 `json:"modesty_level"`     // 0-1, modesty degree
	ExpressivenessLevel float64 `json:"expressiveness_level"` // 0-1, cultural expressiveness
	GreetingStyle    string `json:"greeting_style"`    // handshake, bow, wave, hug, kiss

	// Communication style adaptation
	CommunicationStyle string `json:"communication_style"` // direct, indirect, context_dependent
}

// DisplaySettings defines 3D display configuration for current context.
type DisplaySettings struct {
	RenderMode     string `json:"render_mode"`     // 2d, 3d, vr, ar
	CameraPosition string `json:"camera_position"` // front, side, angle, portrait
	CameraDistance string `json:"camera_distance"` // close, medium, far
	Background     string `json:"background"`      // transparent, neutral, scene_specific
	AnimationState string `json:"animation_state"` // idle, speaking, walking, gesturing
	Expression     string `json:"expression"`      // current facial expression
}

// AvatarProfile defines the overall avatar personality in appearance.
type AvatarProfile struct {
	IdentityID string `json:"identity_id"`

	// Appearance personality
	VisualPersonality   string `json:"visual_personality"`   // approachable, professional, artistic, reserved
	FirstImpression     string `json:"first_impression"`     // warm, cool, neutral, dynamic
	MemorableFeature    string `json:"memorable_feature"`    // eyes, smile, hair, style, posture
	DistinctivenessScore float64 `json:"distinctiveness_score"` // 0-1, how distinctive

	// Social presentation
	SocialPresence     string `json:"social_presence"`     // dominant, balanced, submissive
	CharismaScore      float64 `json:"charisma_score"`      // 0-1, charisma level
	ApproachabilityScore float64 `json:"approachability_score"` // 0-1, how approachable

	// Consistency
	StyleConsistency   float64 `json:"style_consistency"`   // 0-1, style consistency over time
	AuthenticityScore  float64 `json:"authenticity_score"`  // 0-1, authenticity in appearance

	// Evolution potential
	EvolutionTendency  string `json:"evolution_tendency"`  // stable, evolving, experimental
	FashionAwareness   float64 `json:"fashion awareness"`   // 0-1, fashion trend awareness

	// Metadata
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToJSON converts Avatar to JSON string.
func (a *Avatar) ToJSON() string {
	data, _ := json.Marshal(a)
	return string(data)
}

// FromJSON parses Avatar from JSON string.
func (a *Avatar) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), a)
}

// ToJSON converts AvatarDecisionContext to JSON string.
func (c *AvatarDecisionContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses AvatarDecisionContext from JSON string.
func (c *AvatarDecisionContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}
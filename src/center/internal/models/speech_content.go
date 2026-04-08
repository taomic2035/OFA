// Package models defines the Speech Content (表达内容) models for v5.2.0.
//
// SpeechContent represents the content generation and expression system,
// integrating with philosophy, culture, and emotion systems from v4.x.
package models

import (
	"encoding/json"
	"time"
)

// SpeechContentProfile represents the complete speech content configuration.
type SpeechContentProfile struct {
	IdentityID string `json:"identity_id"`

	// Content style (linked to Philosophy v4.1.0)
	ContentStyle ContentStyle `json:"content_style"`

	// Expression depth (linked to Worldview/LifeView)
	ExpressionDepth ExpressionDepth `json:"expression_depth"`

	// Cultural expression (linked to RegionalCulture v4.3.0)
	CulturalExpression CulturalExpression `json:"cultural_expression"`

	// Social expression (linked to SocialIdentity v4.2.0)
	SocialExpression SocialExpression `json:"social_expression"`

	// Content templates
	ContentTemplates ContentTemplates `json:"content_templates"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContentStyle defines how content is styled.
// Linked to Philosophy v4.1.0 (Worldview, LifeView, ValueSystem).
type ContentStyle struct {
	// Tone
	ToneStyle       string `json:"tone_style"`       // formal, casual, professional, friendly, serious, humorous
	ToneConsistency float64 `json:"tone_consistency"` // 0-1

	// Language level
	LanguageLevel    string `json:"language_level"`    // simple, moderate, sophisticated, academic
	TechnicalLevel   string `json:"technical_level"`   // layman, intermediate, expert
	JargonTolerance  float64 `json:"jargon_tolerance"`  // 0-1, willingness to use jargon

	// Expression style
	Directness       float64 `json:"directness"`       // 0-1, direct vs indirect
	EuphemismUsage   float64 `json:"euphemism_usage"`  // 0-1, tendency to use euphemisms
	MetaphorUsage    float64 `json:"metaphor_usage"`   // 0-1
	AnalogyUsage     float64 `json:"analogy_usage"`    // 0-1
	HumorTendency    float64 `json:"humor_tendency"`   // 0-1

	// Emotional coloring
	EmotionalColoring string `json:"emotional_coloring"` // neutral, warm, cool, passionate
	EnthusiasmLevel   float64 `json:"enthusiasm_level"`   // 0-1

	// Persuasion style
	PersuasionStyle string `json:"persuasion_style"` // logical, emotional, balanced
	EvidenceType    string `json:"evidence_type"`    // data, anecdote, expert, mixed

	// Rhetorical devices
	RhetoricalDevices []string `json:"rhetorical_devices"` // repetition, contrast, question, etc.
}

// ExpressionDepth defines how deeply to express thoughts.
// Linked to Worldview and LifeView from v4.1.0.
type ExpressionDepth struct {
	// Thinking depth
	ThinkingDepth     string `json:"thinking_depth"`     // surface, moderate, deep, philosophical
	AbstractionLevel  string `json:"abstraction_level"`  // concrete, mixed, abstract
	ComplexityLevel   string `json:"complexity_level"`   // simple, moderate, complex

	// Self-disclosure
	SelfDisclosureLevel float64 `json:"self_disclosure_level"` // 0-1, how much to share about self
	IntimacyThreshold   float64 `json:"intimacy_threshold"`    // 0-1, threshold for sharing personal info
	VulnerabilityLevel  float64 `json:"vulnerability_level"`   // 0-1, willingness to show vulnerability

	// Reflection
	ReflectionTendency float64 `json:"reflection_tendency"` // 0-1
	SelfAwarenessLevel float64 `json:"self_awareness_level"` // 0-1

	// Depth by context
	ProfessionalDepth  string `json:"professional_depth"`  // task_focused, analytical, strategic
	PersonalDepth       string `json:"personal_depth"`       // surface, moderate, deep
	PhilosophicalDepth  string `json:"philosophical_depth"`  // practical, reflective, existential
}

// CulturalExpression defines how culture affects expression.
// Linked to RegionalCulture v4.3.0.
type CulturalExpression struct {
	// Communication culture
	HighContextCommunication bool `json:"high_context_communication"` // implicit vs explicit
	IndirectExpression       float64 `json:"indirect_expression"`       // 0-1
	FaceSaving               float64 `json:"face_saving"`               // 0-1, concern for face

	// Cultural references
	CulturalReferenceUsage float64 `json:"cultural_reference_usage"` // 0-1
	LocalIdiomUsage        float64 `json:"local_idiom_usage"`        // 0-1
	HistoricalReference    float64 `json:"historical_reference"`     // 0-1

	// Respect patterns
	RespectLevel       float64 `json:"respect_level"`       // 0-1
	HierarchyAwareness float64 `json:"hierarchy_awareness"` // 0-1
	HonorificUsage     string `json:"honorific_usage"`     // none, light, moderate, heavy

	// Cultural taboos
	TabooAwareness     float64   `json:"taboo_awareness"`     // 0-1
	SensitiveTopics    []string `json:"sensitive_topics"`    // topics to avoid or handle carefully
	CulturalNuances    []string `json:"cultural_nuances"`    // cultural nuances to observe

	// Collectivism vs individualism
	CollectivistExpression float64 `json:"collectivist_expression"` // 0-1
	GroupReferenceUsage    float64 `json:"group_reference_usage"`   // 0-1, "we" vs "I"
}

// SocialExpression defines how social identity affects expression.
// Linked to SocialIdentity v4.2.0.
type SocialExpression struct {
	// Professional expression
	ProfessionalTone    string `json:"professional_tone"`    // authoritative, collaborative, supportive
	ExpertiseDisplay    float64 `json:"expertise_display"`    // 0-1, how much to show expertise
	HumilityExpression  float64 `json:"humility_expression"`  // 0-1

	// Social class expression
	ClassExpression     string `json:"class_expression"`     // understated, moderate, aspirational
	StatusAwareness     float64 `json:"status_awareness"`     // 0-1
	NetworkingStyle     string `json:"networking_style"`     // reserved, balanced, proactive

	// Role expression
	RoleConsistency     float64 `json:"role_consistency"`     // 0-1
	RoleAdaptability    float64 `json:"role_adaptability"`    // 0-1
	AuthorityExpression string `json:"authority_expression"` // formal, earned, collaborative

	// Identity confidence
	IdentityConfidence  float64 `json:"identity_confidence"`  // 0-1
	AuthenticExpression float64 `json:"authentic_expression"` // 0-1
}

// ContentTemplates defines reusable content templates.
type ContentTemplates struct {
	// Greeting templates
	GreetingTemplates map[string]string `json:"greeting_templates"` // context -> template

	// Response templates
	ResponseTemplates map[string]string `json:"response_templates"` // situation -> template

	// Explanation templates
	ExplanationTemplates map[string]string `json:"explanation_templates"` // type -> template

	// Closing templates
	ClosingTemplates map[string]string `json:"closing_templates"` // context -> template

	// Apology templates
	ApologyTemplates map[string]string `json:"apology_templates"` // severity -> template

	// Gratitude templates
	GratitudeTemplates map[string]string `json:"gratitude_templates"` // intensity -> template

	// Custom templates
	CustomTemplates map[string]string `json:"custom_templates"` // name -> template
}

// SpeechContentRequest represents a content generation request.
type SpeechContentRequest struct {
	IdentityID string `json:"identity_id"`

	// Content type
	ContentType string `json:"content_type"` // greeting, response, explanation, story, opinion, etc.

	// Context
	Context SpeechContext `json:"context"`

	// Parameters
	Topic       string `json:"topic"`
	Purpose     string `json:"purpose"`     // inform, persuade, entertain, comfort, etc.
	Audience    string `json:"audience"`    // peer, superior, subordinate, public, child
	Length      string `json:"length"`      // short, medium, long
	Formality   string `json:"formality"`   // casual, neutral, formal

	// Emotion context (from v4.0.0)
	EmotionContext EmotionContext `json:"emotion_context"`

	// Constraints
	Constraints ContentConstraints `json:"constraints"`
}

// SpeechContext provides context for content generation.
type SpeechContext struct {
	Scene          string `json:"scene"`           // meeting, casual, presentation, etc.
	SocialContext  string `json:"social_context"`  // one_on_one, group, public
	CulturalContext string `json:"cultural_context"`
	TimeOfDay      string `json:"time_of_day"`     // morning, afternoon, evening, night
	Relationship   string `json:"relationship"`    // close, acquaintance, professional, stranger
	History        string `json:"history"`         // conversation history summary
}

// EmotionContext provides emotion context for content.
type EmotionContext struct {
	CurrentEmotion   string  `json:"current_emotion"`
	EmotionIntensity float64  `json:"emotion_intensity"`
	EmotionValence   string  `json:"emotion_valence"`   // positive, negative, neutral
	AffectTone       string  `json:"affect_tone"`       // warm, cool, neutral
}

// ContentConstraints defines constraints for content generation.
type ContentConstraints struct {
	MaxLength       int      `json:"max_length"`       // characters
	MinLength       int      `json:"min_length"`       // characters
	ForbiddenWords  []string `json:"forbidden_words"`
	RequiredTopics  []string `json:"required_topics"`
	AvoidTopics     []string `json:"avoid_topics"`
	ToneRequirement string   `json:"tone_requirement"` // must be friendly, etc.
}

// SpeechContentResult represents a generated content result.
type SpeechContentResult struct {
	IdentityID string `json:"identity_id"`
	RequestID  string `json:"request_id"`

	// Generated content
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	ContentStyle string `json:"content_style"`

	// Metadata
	WordCount    int    `json:"word_count"`
	CharacterCount int  `json:"character_count"`
	ReadingTime  int    `json:"reading_time"` // seconds

	// Style analysis
	ToneUsed         string `json:"tone_used"`
	FormalityLevel   string `json:"formality_level"`
	EmotionalTone    string `json:"emotional_tone"`
	LanguageComplexity string `json:"language_complexity"`

	// Cultural adaptation
	CulturalAdaptations []string `json:"cultural_adaptations"`

	// Quality metrics
	ClarityScore     float64 `json:"clarity_score"`
	Appropriateness  float64 `json:"appropriateness"`
	AuthenticityScore float64 `json:"authenticity_score"`

	// Generation info
	GenerationTime int    `json:"generation_time"` // ms
	ModelUsed      string `json:"model_used"`
	CacheHit       bool   `json:"cache_hit"`
}

// ContentDecisionContext provides content-related decision context.
type ContentDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// Recommended settings
	RecommendedTone       string `json:"recommended_tone"`
	RecommendedFormality  string `json:"recommended_formality"`
	RecommendedDepth      string `json:"recommended_depth"`
	RecommendedLength     string `json:"recommended_length"`
	RecommendedDirectness float64 `json:"recommended_directness"`

	// Context adaptations
	SceneAdaptation     ContentSceneAdaptation `json:"scene_adaptation"`
	EmotionAdaptation   ContentEmotionAdaptation `json:"emotion_adaptation"`
	SocialAdaptation    ContentSocialAdaptation `json:"social_adaptation"`
	CulturalAdaptation  ContentCulturalAdaptation `json:"cultural_adaptation"`

	// Suggested approaches
	OpeningSuggestion string `json:"opening_suggestion"`
	ClosingSuggestion string `json:"closing_suggestion"`
	KeyTopicsToAvoid  []string `json:"key_topics_to_avoid"`
	KeyTopicsToInclude []string `json:"key_topics_to_include"`

	// Timestamps
	Timestamp time.Time `json:"timestamp"`
}

// ContentSceneAdaptation defines content adaptation for scenes.
type ContentSceneAdaptation struct {
	Scene             string `json:"scene"`
	ToneAdjust        string `json:"tone_adjust"`
	FormalityAdjust   string `json:"formality_adjust"`
	DepthAdjust       string `json:"depth_adjust"`
	LengthPreference  string `json:"length_preference"`
	TopicFocus        string `json:"topic_focus"`
}

// ContentEmotionAdaptation defines content adaptation for emotions.
type ContentEmotionAdaptation struct {
	CurrentEmotion     string `json:"current_emotion"`
	EmotionalColoring  string `json:"emotional_coloring"`
	ExpressionIntensity float64 `json:"expression_intensity"`
	WordChoice         string `json:"word_choice"`      // positive, negative, neutral
	SentenceStyle      string `json:"sentence_style"`   // flowing, choppy, measured
}

// ContentSocialAdaptation defines content adaptation for social context.
type ContentSocialAdaptation struct {
	SocialContext     string `json:"social_context"`
	RespectLevel      float64 `json:"respect_level"`
	HonorificUsage    string `json:"honorific_usage"`
	SelfReferenceStyle string `json:"self_reference_style"` // humble, neutral, confident
	OtherReferenceStyle string `json:"other_reference_style"`
}

// ContentCulturalAdaptation defines content adaptation for cultural context.
type ContentCulturalAdaptation struct {
	CulturalContext      string `json:"cultural_context"`
	IndirectnessLevel    float64 `json:"indirectness_level"`
	FaceSavingLevel      float64 `json:"face_saving_level"`
	CollectivistEmphasis float64 `json:"collectivist_emphasis"`
	CulturalReferences   []string `json:"cultural_references"`
}

// ToJSON converts SpeechContentProfile to JSON string.
func (p *SpeechContentProfile) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// FromJSON parses SpeechContentProfile from JSON string.
func (p *SpeechContentProfile) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), p)
}

// ToJSON converts ContentDecisionContext to JSON string.
func (c *ContentDecisionContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses ContentDecisionContext from JSON string.
func (c *ContentDecisionContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}
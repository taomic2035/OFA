// Package models defines the Expression & Gesture (表情动作) models for v5.3.0.
//
// ExpressionGesture represents the facial expression and body gesture system,
// integrating with emotion and relationship systems from v4.x.
package models

import (
	"encoding/json"
	"time"
)

// ExpressionGestureProfile represents the complete expression and gesture configuration.
type ExpressionGestureProfile struct {
	IdentityID string `json:"identity_id"`

	// Facial expression settings
	FacialExpressionSettings FacialExpressionSettings `json:"facial_expression_settings"`

	// Body gesture settings
	BodyGestureSettings BodyGestureSettings `json:"body_gesture_settings"`

	// Emotion-to-expression mapping (linked to v4.0.0, v4.5.0)
	EmotionExpressionMapping EmotionExpressionMapping `json:"emotion_expression_mapping"`

	// Social gesture settings (linked to v4.6.0)
	SocialGestureSettings SocialGestureSettings `json:"social_gesture_settings"`

	// Animation settings
	AnimationSettings AnimationSettings `json:"animation_settings"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FacialExpressionSettings defines facial expression configuration.
type FacialExpressionSettings struct {
	// Base expression
	DefaultExpression string `json:"default_expression"` // neutral, smile, serious, relaxed

	// Expression range
	ExpressionRange     float64 `json:"expression_range"`     // 0-1, how much expressions can vary
	ExpressionIntensity float64 `json:"expression_intensity"` // 0-1, base intensity

	// Expression frequency
	ExpressionFrequency float64 `json:"expression_frequency"` // 0-1, how often expressions change
	ExpressionDuration  string `json:"expression_duration"`  // short, medium, long

	// Eye expression
	EyeExpressionEnabled bool    `json:"eye_expression_enabled"`
	EyeContactTendency   float64  `json:"eye_contact_tendency"`   // 0-1
	BlinkRate            float64  `json:"blink_rate"`             // blinks per minute
	EyebrowExpressiveness float64 `json:"eyebrow_expressiveness"` // 0-1

	// Mouth expression
	MouthExpressionEnabled bool    `json:"mouth_expression_enabled"`
	SmileTendency         float64  `json:"smile_tendency"`   // 0-1
	SmileType            string  `json:"smile_type"`      // subtle, moderate, broad
	LipMovementExpressiveness float64 `json:"lip_movement_expressiveness"` // 0-1

	// Micro-expressions
	MicroExpressionEnabled bool    `json:"micro_expression_enabled"`
	MicroExpressionSensitivity float64 `json:"micro_expression_sensitivity"` // 0-1
	MicroExpressionDuration int   `json:"micro_expression_duration"` // ms

	// Expression symmetry
	SymmetryLevel float64 `json:"symmetry_level"` // 0-1, 1 = perfectly symmetric

	// Expression masking
	ExpressionMasking float64 `json:"expression_masking"` // 0-1, how much to hide true feelings
	PokerFaceAbility  float64 `json:"poker_face_ability"`  // 0-1
}

// BodyGestureSettings defines body gesture configuration.
type BodyGestureSettings struct {
	// Base posture
	DefaultPosture string `json:"default_posture"` // neutral, confident, relaxed, formal

	// Gesture range
	GestureRange     float64 `json:"gesture_range"`     // 0-1, how much gestures can vary
	GestureIntensity float64 `json:"gesture_intensity"` // 0-1, base intensity

	// Gesture frequency
	GestureFrequency float64 `json:"gesture_frequency"` // 0-1, how often gestures occur
	GestureSpeed     string `json:"gesture_speed"`     // slow, moderate, fast

	// Hand gestures
	HandGestureEnabled bool    `json:"hand_gesture_enabled"`
	HandGestureStyle   string  `json:"hand_gesture_style"`   // minimal, moderate, expressive
	HandPosition       string  `json:"hand_position"`        // natural, crossed, pockets, animated
	HandGestureVocabulary []string `json:"hand_gesture_vocabulary"` // available gestures

	// Head movements
	HeadMovementEnabled bool    `json:"head_movement_enabled"`
	NodFrequency       float64  `json:"nod_frequency"`      // 0-1
	HeadTiltTendency   float64  `json:"head_tilt_tendency"` // 0-1
	HeadShakeFrequency float64  `json:"head_shake_frequency"` // 0-1

	// Body lean
	BodyLeanEnabled bool   `json:"body_lean_enabled"`
	BodyLeanDirection string `json:"body_lean_direction"` // forward, neutral, backward
	BodyLeanTendency float64 `json:"body_lean_tendency"` // 0-1

	// Shoulder movements
	ShrugTendency float64 `json:"shrug_tendency"` // 0-1
	ShoulderTension string `json:"shoulder_tension"` // relaxed, moderate, tense

	// Fidgeting
	FidgetLevel float64 `json:"fidget_level"` // 0-1
	FidgetType  string `json:"fidget_type"`  // none, subtle, noticeable

	// Mirroring
	MirroringEnabled bool    `json:"mirroring_enabled"`
	MirroringDelay   int     `json:"mirroring_delay"`    // ms
	MirroringIntensity float64 `json:"mirroring_intensity"` // 0-1
}

// EmotionExpressionMapping defines how emotions map to expressions/gestures.
type EmotionExpressionMapping struct {
	// Emotion-to-expression mappings
	EmotionMappings map[string]ExpressionMapping `json:"emotion_mappings"`

	// Transition settings
	TransitionSpeed string `json:"transition_speed"` // instant, smooth, gradual
	TransitionDuration int `json:"transition_duration"` // ms

	// Blending
	BlendEnabled bool `json:"blend_enabled"`
	BlendDuration int `json:"blend_duration"` // ms
}

// ExpressionMapping defines expression for a specific emotion.
type ExpressionMapping struct {
	EmotionName string `json:"emotion_name"`

	// Facial expression
	EyebrowPosition string `json:"eyebrow_position"` // raised, neutral, furrowed
	EyeShape       string `json:"eye_shape"`        // wide, normal, narrowed, closed
	MouthShape     string `json:"mouth_shape"`      // smile, frown, neutral, open
	ExpressionType string `json:"expression_type"`  // specific expression name
	Intensity      float64 `json:"intensity"`        // 0-1

	// Body gesture
	Posture        string `json:"posture"`         // confident, slumped, open, closed
	HandGesture    string `json:"hand_gesture"`    // gesture to use
	BodyMovement   string `json:"body_movement"`   // movement style

	// Micro-expression
	MicroExpression string `json:"micro_expression"` // subtle expression hint
}

// SocialGestureSettings defines social gesture behavior.
// Linked to RelationshipProfile v4.6.0.
type SocialGestureSettings struct {
	// Greeting gestures
	GreetingGesture string `json:"greeting_gesture"` // wave, nod, bow, handshake
	GreetingIntensity float64 `json:"greeting_intensity"` // 0-1

	// Parting gestures
	PartingGesture string `json:"parting_gesture"` // wave, nod, bow
	PartingIntensity float64 `json:"parting_intensity"` // 0-1

	// Listening gestures
	ListeningGestureEnabled bool    `json:"listening_gesture_enabled"`
	NodWhileListening      float64  `json:"nod_while_listening"` // 0-1
	EyeContactWhileListening float64 `json:"eye_contact_while_listening"` // 0-1

	// Speaking gestures
	SpeakingGestureEnabled bool    `json:"speaking_gesture_enabled"`
	GestureWhileSpeaking   float64  `json:"gesture_while_speaking"` // 0-1
	PauseGesture          string  `json:"pause_gesture"`          // gesture during pauses

	// Agreement/disagreement
	AgreementGesture   string `json:"agreement_gesture"`   // nod, smile
	DisagreementGesture string `json:"disagreement_gesture"` // head_shake, frown
	UncertaintyGesture string `json:"uncertainty_gesture"` // shrug, tilt

	// Touch gestures
	TouchComfortLevel float64 `json:"touch_comfort_level"` // 0-1
	TouchTypes        []string `json:"touch_types"`       // handshake, pat, hug

	// Distance management
	PreferredDistance string `json:"preferred_distance"` // close, medium, far
	DistanceAdjustment float64 `json:"distance_adjustment"` // 0-1, tendency to adjust

	// Mirroring by relationship
	MirrorCloseFriends bool `json:"mirror_close_friends"`
	MirrorProfessional bool `json:"mirror_professional"`
	MirrorStrangers    bool `json:"mirror_strangers"`
}

// AnimationSettings defines animation configuration.
type AnimationSettings struct {
	// Animation style
	AnimationStyle string `json:"animation_style"` // realistic, stylized, cartoon

	// Idle animation
	IdleAnimationEnabled bool    `json:"idle_animation_enabled"`
	IdleAnimationSet     []string `json:"idle_animation_set"`
	IdleVariationFrequency float64 `json:"idle_variation_frequency"` // 0-1

	// Transition animations
	TransitionAnimationsEnabled bool `json:"transition_animations_enabled"`
	TransitionSpeed            string `json:"transition_speed"` // slow, normal, fast

	// Expression animations
	ExpressionAnimationQuality string `json:"expression_animation_quality"` // low, medium, high
	ExpressionAnimationFPS     int    `json:"expression_animation_fps"`

	// Gesture animations
	GestureAnimationQuality string `json:"gesture_animation_quality"` // low, medium, high
	GestureAnimationFPS     int    `json:"gesture_animation_fps"`

	// Lip sync
	LipSyncEnabled bool `json:"lip_sync_enabled"`
	LipSyncQuality string `json:"lip_sync_quality"` // basic, standard, high
	LipSyncDelay  int   `json:"lip_sync_delay"`   // ms

	// Blink animation
	BlinkAnimationEnabled bool `json:"blink_animation_enabled"`
	BlinkAnimationNatural  bool `json:"blink_animation_natural"`

	// Breathing animation
	BreathingAnimationEnabled bool `json:"breathing_animation_enabled"`
	BreathingRate            float64 `json:"breathing_rate"` // breaths per minute
	BreathingDepth           float64 `json:"breathing_depth"` // 0-1

	// Eye movement
	EyeMovementEnabled bool `json:"eye_movement_enabled"`
	SaccadeFrequency   float64 `json:"saccade_frequency"` // eye movements per minute
	SaccadeRange       float64 `json:"saccade_range"`     // 0-1
}

// ExpressionGestureContext provides expression/gesture context.
type ExpressionGestureContext struct {
	IdentityID string `json:"identity_id"`

	// Current expression
	CurrentExpression ExpressionState `json:"current_expression"`

	// Current gesture
	CurrentGesture GestureState `json:"current_gesture"`

	// Recommended actions
	RecommendedExpression ExpressionState `json:"recommended_expression"`
	RecommendedGesture    GestureState `json:"recommended_gesture"`

	// Context adaptations
	SceneAdaptation      ExpressionSceneAdaptation `json:"scene_adaptation"`
	EmotionAdaptation    ExpressionEmotionAdaptation `json:"emotion_adaptation"`
	SocialAdaptation     ExpressionSocialAdaptation `json:"social_adaptation"`

	// Animation state
	AnimationState AnimationState `json:"animation_state"`

	// Timestamps
	Timestamp time.Time `json:"timestamp"`
}

// ExpressionState represents current expression state.
type ExpressionState struct {
	ExpressionName string `json:"expression_name"`
	Intensity      float64 `json:"intensity"`      // 0-1
	Duration       int    `json:"duration"`       // ms
	Transition     string `json:"transition"`     // transition type

	// Facial details
	EyebrowState string `json:"eyebrow_state"`
	EyeState     string `json:"eye_state"`
	MouthState   string `json:"mouth_state"`

	// Blending
	BlendWithPrevious bool `json:"blend_with_previous"`
	BlendProgress     float64 `json:"blend_progress"` // 0-1
}

// GestureState represents current gesture state.
type GestureState struct {
	GestureName string `json:"gesture_name"`
	Intensity   float64 `json:"intensity"`   // 0-1
	Duration    int    `json:"duration"`    // ms
	Transition  string `json:"transition"`  // transition type

	// Body details
	Posture      string `json:"posture"`
	HandPosition string `json:"hand_position"`
	HeadPosition string `json:"head_position"`

	// Mirroring
	IsMirroring    bool   `json:"is_mirroring"`
	MirroredFrom   string `json:"mirrored_from"`
	MirroringDelay int    `json:"mirroring_delay"` // ms
}

// ExpressionSceneAdaptation defines expression adaptation for scenes.
type ExpressionSceneAdaptation struct {
	Scene               string `json:"scene"`
	ExpressionRange     float64 `json:"expression_range"`     // 0-1
	GestureRange        float64 `json:"gesture_range"`        // 0-1
	FormalityLevel      float64 `json:"formality_level"`      // 0-1
	EyeContactLevel     float64 `json:"eye_contact_level"`    // 0-1
	IdleAnimationStyle  string `json:"idle_animation_style"` // subtle, moderate, expressive
}

// ExpressionEmotionAdaptation defines expression adaptation for emotions.
type ExpressionEmotionAdaptation struct {
	CurrentEmotion      string `json:"current_emotion"`
	TargetExpression    string `json:"target_expression"`
	TargetGesture       string `json:"target_gesture"`
	TransitionSpeed     string `json:"transition_speed"`
	IntensityMultiplier float64 `json:"intensity_multiplier"`
}

// ExpressionSocialAdaptation defines expression adaptation for social context.
type ExpressionSocialAdaptation struct {
	SocialContext       string `json:"social_context"`
	GreetingGesture     string `json:"greeting_gesture"`
	FarewellGesture     string `json:"farewell_gesture"`
	EyeContactTendency  float64 `json:"eye_contact_tendency"`
	MirroringEnabled    bool   `json:"mirroring_enabled"`
	TouchPermission     string `json:"touch_permission"` // none, light, moderate
}

// AnimationState represents current animation state.
type AnimationState struct {
	CurrentAnimation string `json:"current_animation"`
	AnimationProgress float64 `json:"animation_progress"` // 0-1
	AnimationFPS     int    `json:"animation_fps"`
	IsLooping        bool   `json:"is_looping"`
	IsBlending       bool   `json:"is_blending"`
	BlendProgress    float64 `json:"blend_progress"` // 0-1
}

// ToJSON converts ExpressionGestureProfile to JSON string.
func (p *ExpressionGestureProfile) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// FromJSON parses ExpressionGestureProfile from JSON string.
func (p *ExpressionGestureProfile) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), p)
}

// ToJSON converts ExpressionGestureContext to JSON string.
func (c *ExpressionGestureContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses ExpressionGestureContext from JSON string.
func (c *ExpressionGestureContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}
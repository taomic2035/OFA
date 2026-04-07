// Package models defines the Voice (语音) models for v5.1.0.
//
// Voice represents the voice characteristics and synthesis settings
// for the digital person, integrating with emotion and culture systems.
package models

import (
	"encoding/json"
	"time"
)

// VoiceProfile represents the complete voice characteristics.
type VoiceProfile struct {
	IdentityID string `json:"identity_id"`

	// Basic voice characteristics
	VoiceCharacteristics VoiceCharacteristics `json:"voice_characteristics"`

	// Voice style (linked to RegionalCulture v4.3.0)
	VoiceStyle VoiceStyle `json:"voice_style"`

	// Emotional voice settings (linked to Emotion v4.0.0)
	EmotionalVoice EmotionalVoice `json:"emotional_voice"`

	// Speech patterns (linked to Philosophy v4.1.0)
	SpeechPatterns SpeechPatterns `json:"speech_patterns"`

	// TTS configuration
	TTSConfig TTSConfiguration `json:"tts_config"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VoiceCharacteristics defines basic voice characteristics.
type VoiceCharacteristics struct {
	// Voice type
	VoiceType    string `json:"voice_type"`    // male, female, neutral
	VoiceAge     string `json:"voice_age"`     // child, young, adult, middle_aged, senior
	VoiceQuality string `json:"voice_quality"` // clear, breathy, husky, resonant

	// Pitch
	BasePitch     double `json:"base_pitch"`     // Hz, fundamental frequency
	PitchRange    double `json:"pitch_range"`    // semitones, pitch variation
	PitchVariation string `json:"pitch_variation"` // flat, moderate, dynamic

	// Rate
	SpeakingRate   double `json:"speaking_rate"`   // words per minute
	RateVariation  string `json:"rate_variation"`  // steady, varied
	PauseFrequency double `json:"pause_frequency"` // 0-1, how often pauses

	// Volume
	BaseVolume     double `json:"base_volume"`     // 0-1
	VolumeRange    double `json:"volume_range"`    // 0-1
	VolumeVariation string `json:"volume_variation"` // steady, dynamic

	// Timbre
	TimbreType     string `json:"timbre_type"`     // warm, bright, dark, neutral
	Resonance      string `json:"resonance"`       // low, medium, high
	Breathiness    double `json:"breathiness"`     // 0-1
	Roughness      double `json:"roughness"`       // 0-1

	// Articulation
	ArticulationStyle string `json:"articulation_style"` // precise, relaxed, casual
	AccentStrength    double `json:"accent_strength"`    // 0-1

	// Voice uniqueness
	VoiceFingerprint string `json:"voice_fingerprint"` // unique voice identifier
	Distinctiveness  double `json:"distinctiveness"`   // 0-1, how distinctive
}

// VoiceStyle defines voice style influenced by culture and social identity.
// Linked to RegionalCulture v4.3.0 and SocialIdentity v4.2.0.
type VoiceStyle struct {
	// Cultural influence
	RegionalAccent   string `json:"regional_accent"`   // standard, regional, dialect
	AccentRegion     string `json:"accent_region"`     // e.g., "beijing", "shanghai"
	AccentIntensity  double `json:"accent_intensity"`  // 0-1

	// Formality
	FormalityLevel   string `json:"formality_level"`   // casual, neutral, formal
	FormalityAdjust  double `json:"formality_adjust"`  // -1 to 1, adjust formality

	// Communication style
	CommunicationStyle string `json:"communication_style"` // direct, indirect, context_dependent
	IndirectnessLevel  double `json:"indirectness_level"`  // 0-1

	// Social influence
	ProfessionalVoice bool   `json:"professional_voice"`
	VoiceAuthority    double `json:"voice_authority"` // 0-1
	VoiceWarmth       double `json:"voice_warmth"`    // 0-1

	// Context adaptation
	AdaptiveStyle     bool   `json:"adaptive_style"`     // auto-adapt to context
	StyleConsistency  double `json:"style_consistency"`  // 0-1
}

// EmotionalVoice defines how emotions affect voice.
// Linked to Emotion v4.0.0 and EmotionBehavior v4.5.0.
type EmotionalVoice struct {
	// Emotion-to-voice mapping
	EmotionMappings map[string]VoiceEmotionMapping `json:"emotion_mappings"`

	// Global emotional influence
	EmotionalExpressiveness double `json:"emotional_expressiveness"` // 0-1
	EmotionalContagion      double `json:"emotional_contagion"`      // 0-1, how emotions spread to voice

	// Voice modulation by emotion
	PitchModulationRange double `json:"pitch_modulation_range"` // semitones
	RateModulationRange  double `json:"rate_modulation_range"`  // % change
	VolumeModulation     double `json:"volume_modulation"`      // 0-1

	// Emotional regulation
	EmotionRegulation string `json:"emotion_regulation"` // suppress, express, amplify
	BaselineStability double `json:"baseline_stability"` // 0-1, how stable baseline is
}

// VoiceEmotionMapping defines how a specific emotion affects voice.
type VoiceEmotionMapping struct {
	EmotionName string `json:"emotion_name"`

	// Pitch changes
	PitchShift double `json:"pitch_shift"` // semitones, positive = higher

	// Rate changes
	RateMultiplier double `json:"rate_multiplier"` // 1.0 = normal, >1 = faster

	// Volume changes
	VolumeShift double `json:"volume_shift"` // 0-1

	// Timbre changes
	TimbreAdjust string `json:"timbre_adjust"` // warmer, brighter, darker

	// Articulation changes
	Articulation string `json:"articulation"` // precise, relaxed, tense

	// Additional effects
	BreathinessAdjust double `json:"breathiness_adjust"`
	TensionLevel      double `json:"tension_level"` // 0-1
}

// SpeechPatterns defines speech patterns influenced by philosophy.
// Linked to Philosophy v4.1.0.
type SpeechPatterns struct {
	// Word choice
	VocabularyLevel   string `json:"vocabulary_level"`   // simple, moderate, sophisticated
	JargonUsage       double `json:"jargon_usage"`       // 0-1
	IdiomUsage        double `json:"idiom_usage"`        // 0-1
	MetaphorUsage     double `json:"metaphor_usage"`     // 0-1

	// Sentence structure
	SentenceLength    string `json:"sentence_length"`    // short, medium, long
	SentenceComplexity string `json:"sentence_complexity"` // simple, compound, complex
	ClauseUsage       double `json:"clause_usage"`       // 0-1

	// Speech habits
	FillerWords      []string `json:"filler_words"`      // common fillers
	FillerFrequency  double   `json:"filler_frequency"`  // 0-1
	HedgingLanguage  double   `json:"hedging_language"`  // 0-1, "maybe", "perhaps"
	QualifyingWords  double   `json:"qualifying_words"`  // 0-1, "somewhat", "quite"

	// Thought expression
	ThoughtOrganization string `json:"thought_organization"` // linear, associative, structured
	DigressionTendency  double `json:"digression_tendency"`  // 0-1
	TangentFrequency    double `json:"tangent_frequency"`    // 0-1

	// Listening style
	ListeningResponses []string `json:"listening_responses"` // "mm-hmm", "I see"
	ResponseFrequency  double   `json:"response_frequency"`  // 0-1
	InterruptTendency  double   `json:"interrupt_tendency"`  // 0-1

	// Pause patterns
	PauseBeforeSpeaking double `json:"pause_before_speaking"` // 0-1
	ThoughtfulPauses    double `json:"thoughtful_pauses"`     // 0-1
	RhetoricalPauses    double `json:"rhetorical_pauses"`     // 0-1
}

// TTSConfiguration defines TTS engine settings.
type TTSConfiguration struct {
	// Engine selection
	EngineType   string `json:"engine_type"`   // built_in, cloud, hybrid
	EngineName   string `json:"engine_name"`   // specific engine name
	EngineVersion string `json:"engine_version"`

	// Voice model
	VoiceModelID   string `json:"voice_model_id"`
	VoiceModelName string `json:"voice_model_name"`
	CustomVoice    bool   `json:"custom_voice"`

	// Synthesis settings
	SampleRate     int    `json:"sample_rate"`     // Hz
	BitDepth       int    `json:"bit_depth"`       // 16, 24
	Channels       int    `json:"channels"`        // 1 (mono), 2 (stereo)
	AudioFormat    string `json:"audio_format"`    // wav, mp3, ogg

	// Quality settings
	QualityLevel   string `json:"quality_level"`   // low, medium, high, ultra
	LatencyMode    string `json:"latency_mode"`    // real_time, balanced, quality
	BufferSize     int    `json:"buffer_size"`     // ms

	// SSML support
	SSMLEnabled    bool   `json:"ssml_enabled"`
	SSMLFeatures   []string `json:"ssml_features"` // break, emphasis, prosody, phoneme

	// Streaming
	StreamingEnabled bool   `json:"streaming_enabled"`
	ChunkSize        int    `json:"chunk_size"` // ms

	// Caching
	CacheEnabled    bool   `json:"cache_enabled"`
	CacheSize       int    `json:"cache_size"` // MB
	CacheTTL        int    `json:"cache_ttl"`  // seconds

	// Fallback
	FallbackEngine  string `json:"fallback_engine"`
	FailoverEnabled bool   `json:"failover_enabled"`
}

// VoiceSynthesisRequest represents a TTS synthesis request.
type VoiceSynthesisRequest struct {
	IdentityID string `json:"identity_id"`
	Text       string `json:"text"`

	// Emotion context (from v4.0.0)
	EmotionContext VoiceEmotionContext `json:"emotion_context"`

	// Speech context
	SpeechContext SpeechContext `json:"speech_context"`

	// Output settings
	OutputFormat string `json:"output_format"` // wav, mp3, ogg
	OutputDevice string `json:"output_device"` // speaker, headphones, bluetooth

	// Synthesis options
	Streaming    bool   `json:"streaming"`
	CacheResult  bool   `json:"cache_result"`
	Priority     int    `json:"priority"` // 0-3
}

// VoiceEmotionContext provides emotion context for synthesis.
type VoiceEmotionContext struct {
	CurrentEmotion   string  `json:"current_emotion"`
	EmotionIntensity double  `json:"emotion_intensity"` // 0-1
	EmotionTrend     string  `json:"emotion_trend"`     // rising, falling, stable
	EmotionHistory   []string `json:"emotion_history"`  // recent emotions
}

// SpeechContext provides context for speech synthesis.
type SpeechContext struct {
	Scene          string `json:"scene"`           // meeting, casual, formal
	SocialContext  string `json:"social_context"`  // one_on_one, group, public
	CulturalContext string `json:"cultural_context"` // local, international
	Formality      string `json:"formality"`       // casual, neutral, formal
	Purpose        string `json:"purpose"`         // inform, persuade, entertain
	Audience       string `json:"audience"`        // peer, superior, subordinate, public
}

// VoiceSynthesisResult represents a TTS synthesis result.
type VoiceSynthesisResult struct {
	IdentityID string `json:"identity_id"`
	RequestID  string `json:"request_id"`

	// Audio data
	AudioData     []byte `json:"audio_data"`
	AudioFormat   string `json:"audio_format"`
	DurationMs    int    `json:"duration_ms"`
	SampleRate    int    `json:"sample_rate"`

	// Synthesis info
	EngineUsed    string `json:"engine_used"`
	SynthesisTime int    `json:"synthesis_time"` // ms
	QualityScore  double `json:"quality_score"`

	// Prosody applied
	PitchApplied    double `json:"pitch_applied"`
	RateApplied     double `json:"rate_applied"`
	VolumeApplied   double `json:"volume_applied"`

	// Status
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	CacheHit   bool   `json:"cache_hit"`
}

// VoiceDecisionContext provides voice-related decision context.
type VoiceDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// Recommended settings
	RecommendedPitch    double `json:"recommended_pitch"`
	RecommendedRate     double `json:"recommended_rate"`
	RecommendedVolume   double `json:"recommended_volume"`
	RecommendedTone     string `json:"recommended_tone"`
	RecommendedFormality string `json:"recommended_formality"`

	// Context adaptations
	SceneAdaptation     VoiceSceneAdaptation `json:"scene_adaptation"`
	EmotionAdaptation   VoiceEmotionAdaptation `json:"emotion_adaptation"`
	SocialAdaptation    VoiceSocialAdaptation `json:"social_adaptation"`

	// Output settings
	OutputSettings VoiceOutputSettings `json:"output_settings"`

	// Timestamps
	Timestamp time.Time `json:"timestamp"`
}

// VoiceSceneAdaptation defines voice adaptation for scenes.
type VoiceSceneAdaptation struct {
	Scene             string `json:"scene"`
	PitchAdjust       double `json:"pitch_adjust"`       // semitones
	RateAdjust        double `json:"rate_adjust"`        // multiplier
	VolumeAdjust      double `json:"volume_adjust"`      // 0-1
	FormalityAdjust   string `json:"formality_adjust"`   // casual_up, formal_up
	ArticulationMode  string `json:"articulation_mode"`  // precise, relaxed
}

// VoiceEmotionAdaptation defines voice adaptation for emotions.
type VoiceEmotionAdaptation struct {
	CurrentEmotion    string `json:"current_emotion"`
	PitchShift        double `json:"pitch_shift"`        // semitones
	RateMultiplier    double `json:"rate_multiplier"`    // multiplier
	VolumeShift       double `json:"volume_shift"`       // 0-1
	TimbreAdjust      string `json:"timbre_adjust"`      // warmer, brighter
	ExpressivenessLevel double `json:"expressiveness_level"` // 0-1
}

// VoiceSocialAdaptation defines voice adaptation for social context.
type VoiceSocialAdaptation struct {
	SocialContext  string `json:"social_context"`
	AuthorityLevel double `json:"authority_level"` // 0-1
	WarmthLevel    double `json:"warmth_level"`    // 0-1
	Directness     double `json:"directness"`      // 0-1
	PauseBeforeSpeech double `json:"pause_before_speech"` // ms
}

// VoiceOutputSettings defines output configuration.
type VoiceOutputSettings struct {
	OutputDevice  string `json:"output_device"`
	VolumeLevel   double `json:"volume_level"`   // 0-1
	SpatialMode   string `json:"spatial_mode"`   // mono, stereo, spatial
	EffectsEnabled bool  `json:"effects_enabled"`
	ReverbLevel   double `json:"reverb_level"`   // 0-1
}

// ToJSON converts VoiceProfile to JSON string.
func (v *VoiceProfile) ToJSON() string {
	data, _ := json.Marshal(v)
	return string(data)
}

// FromJSON parses VoiceProfile from JSON string.
func (v *VoiceProfile) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), v)
}

// ToJSON converts VoiceDecisionContext to JSON string.
func (c *VoiceDecisionContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses VoiceDecisionContext from JSON string.
func (c *VoiceDecisionContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}
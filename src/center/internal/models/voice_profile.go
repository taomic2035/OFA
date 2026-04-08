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
	BasePitch     float64 `json:"base_pitch"`     // Hz, fundamental frequency
	PitchRange    float64 `json:"pitch_range"`    // semitones, pitch variation
	PitchVariation string `json:"pitch_variation"` // flat, moderate, dynamic

	// Rate
	SpeakingRate   float64 `json:"speaking_rate"`   // words per minute
	RateVariation  string `json:"rate_variation"`  // steady, varied
	PauseFrequency float64 `json:"pause_frequency"` // 0-1, how often pauses

	// Volume
	BaseVolume     float64 `json:"base_volume"`     // 0-1
	VolumeRange    float64 `json:"volume_range"`    // 0-1
	VolumeVariation string `json:"volume_variation"` // steady, dynamic

	// Timbre
	TimbreType     string `json:"timbre_type"`     // warm, bright, dark, neutral
	Resonance      string `json:"resonance"`       // low, medium, high
	Breathiness    float64 `json:"breathiness"`     // 0-1
	Roughness      float64 `json:"roughness"`       // 0-1

	// Articulation
	ArticulationStyle string `json:"articulation_style"` // precise, relaxed, casual
	AccentStrength    float64 `json:"accent_strength"`    // 0-1

	// Voice uniqueness
	VoiceFingerprint string `json:"voice_fingerprint"` // unique voice identifier
	Distinctiveness  float64 `json:"distinctiveness"`   // 0-1, how distinctive
}

// VoiceStyle defines voice style influenced by culture and social identity.
// Linked to RegionalCulture v4.3.0 and SocialIdentity v4.2.0.
type VoiceStyle struct {
	// Cultural influence
	RegionalAccent   string `json:"regional_accent"`   // standard, regional, dialect
	AccentRegion     string `json:"accent_region"`     // e.g., "beijing", "shanghai"
	AccentIntensity  float64 `json:"accent_intensity"`  // 0-1

	// Formality
	FormalityLevel   string `json:"formality_level"`   // casual, neutral, formal
	FormalityAdjust  float64 `json:"formality_adjust"`  // -1 to 1, adjust formality

	// Communication style
	CommunicationStyle string `json:"communication_style"` // direct, indirect, context_dependent
	IndirectnessLevel  float64 `json:"indirectness_level"`  // 0-1

	// Social influence
	ProfessionalVoice bool   `json:"professional_voice"`
	VoiceAuthority    float64 `json:"voice_authority"` // 0-1
	VoiceWarmth       float64 `json:"voice_warmth"`    // 0-1

	// Context adaptation
	AdaptiveStyle     bool   `json:"adaptive_style"`     // auto-adapt to context
	StyleConsistency  float64 `json:"style_consistency"`  // 0-1
}

// EmotionalVoice defines how emotions affect voice.
// Linked to Emotion v4.0.0 and EmotionBehavior v4.5.0.
type EmotionalVoice struct {
	// Emotion-to-voice mapping
	EmotionMappings map[string]VoiceEmotionMapping `json:"emotion_mappings"`

	// Global emotional influence
	EmotionalExpressiveness float64 `json:"emotional_expressiveness"` // 0-1
	EmotionalContagion      float64 `json:"emotional_contagion"`      // 0-1, how emotions spread to voice

	// Voice modulation by emotion
	PitchModulationRange float64 `json:"pitch_modulation_range"` // semitones
	RateModulationRange  float64 `json:"rate_modulation_range"`  // % change
	VolumeModulation     float64 `json:"volume_modulation"`      // 0-1

	// Emotional regulation
	EmotionRegulation string `json:"emotion_regulation"` // suppress, express, amplify
	BaselineStability float64 `json:"baseline_stability"` // 0-1, how stable baseline is
}

// VoiceEmotionMapping defines how a specific emotion affects voice.
type VoiceEmotionMapping struct {
	EmotionName string `json:"emotion_name"`

	// Pitch changes
	PitchShift float64 `json:"pitch_shift"` // semitones, positive = higher

	// Rate changes
	RateMultiplier float64 `json:"rate_multiplier"` // 1.0 = normal, >1 = faster

	// Volume changes
	VolumeShift float64 `json:"volume_shift"` // 0-1

	// Timbre changes
	TimbreAdjust string `json:"timbre_adjust"` // warmer, brighter, darker

	// Articulation changes
	Articulation string `json:"articulation"` // precise, relaxed, tense

	// Additional effects
	BreathinessAdjust float64 `json:"breathiness_adjust"`
	TensionLevel      float64 `json:"tension_level"` // 0-1
}

// SpeechPatterns defines speech patterns influenced by philosophy.
// Linked to Philosophy v4.1.0.
type SpeechPatterns struct {
	// Word choice
	VocabularyLevel   string `json:"vocabulary_level"`   // simple, moderate, sophisticated
	JargonUsage       float64 `json:"jargon_usage"`       // 0-1
	IdiomUsage        float64 `json:"idiom_usage"`        // 0-1
	MetaphorUsage     float64 `json:"metaphor_usage"`     // 0-1

	// Sentence structure
	SentenceLength    string `json:"sentence_length"`    // short, medium, long
	SentenceComplexity string `json:"sentence_complexity"` // simple, compound, complex
	ClauseUsage       float64 `json:"clause_usage"`       // 0-1

	// Speech habits
	FillerWords      []string `json:"filler_words"`      // common fillers
	FillerFrequency  float64   `json:"filler_frequency"`  // 0-1
	HedgingLanguage  float64   `json:"hedging_language"`  // 0-1, "maybe", "perhaps"
	QualifyingWords  float64   `json:"qualifying_words"`  // 0-1, "somewhat", "quite"

	// Thought expression
	ThoughtOrganization string `json:"thought_organization"` // linear, associative, structured
	DigressionTendency  float64 `json:"digression_tendency"`  // 0-1
	TangentFrequency    float64 `json:"tangent_frequency"`    // 0-1

	// Listening style
	ListeningResponses []string `json:"listening_responses"` // "mm-hmm", "I see"
	ResponseFrequency  float64   `json:"response_frequency"`  // 0-1
	InterruptTendency  float64   `json:"interrupt_tendency"`  // 0-1

	// Pause patterns
	PauseBeforeSpeaking float64 `json:"pause_before_speaking"` // 0-1
	ThoughtfulPauses    float64 `json:"thoughtful_pauses"`     // 0-1
	RhetoricalPauses    float64 `json:"rhetorical_pauses"`     // 0-1
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
	EmotionIntensity float64  `json:"emotion_intensity"` // 0-1
	EmotionTrend     string  `json:"emotion_trend"`     // rising, falling, stable
	EmotionHistory   []string `json:"emotion_history"`  // recent emotions
}

// VoiceSpeechContext provides context for speech synthesis.
type VoiceSpeechContext struct {
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
	QualityScore  float64 `json:"quality_score"`

	// Prosody applied
	PitchApplied    float64 `json:"pitch_applied"`
	RateApplied     float64 `json:"rate_applied"`
	VolumeApplied   float64 `json:"volume_applied"`

	// Status
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	CacheHit   bool   `json:"cache_hit"`
}

// VoiceDecisionContext provides voice-related decision context.
type VoiceDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// Recommended settings
	RecommendedPitch    float64 `json:"recommended_pitch"`
	RecommendedRate     float64 `json:"recommended_rate"`
	RecommendedVolume   float64 `json:"recommended_volume"`
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
	PitchAdjust       float64 `json:"pitch_adjust"`       // semitones
	RateAdjust        float64 `json:"rate_adjust"`        // multiplier
	VolumeAdjust      float64 `json:"volume_adjust"`      // 0-1
	FormalityAdjust   string `json:"formality_adjust"`   // casual_up, formal_up
	ArticulationMode  string `json:"articulation_mode"`  // precise, relaxed
}

// VoiceEmotionAdaptation defines voice adaptation for emotions.
type VoiceEmotionAdaptation struct {
	CurrentEmotion    string `json:"current_emotion"`
	PitchShift        float64 `json:"pitch_shift"`        // semitones
	RateMultiplier    float64 `json:"rate_multiplier"`    // multiplier
	VolumeShift       float64 `json:"volume_shift"`       // 0-1
	TimbreAdjust      string `json:"timbre_adjust"`      // warmer, brighter
	ExpressivenessLevel float64 `json:"expressiveness_level"` // 0-1
}

// VoiceSocialAdaptation defines voice adaptation for social context.
type VoiceSocialAdaptation struct {
	SocialContext  string `json:"social_context"`
	AuthorityLevel float64 `json:"authority_level"` // 0-1
	WarmthLevel    float64 `json:"warmth_level"`    // 0-1
	Directness     float64 `json:"directness"`      // 0-1
	PauseBeforeSpeech float64 `json:"pause_before_speech"` // ms
}

// VoiceOutputSettings defines output configuration.
type VoiceOutputSettings struct {
	OutputDevice  string `json:"output_device"`
	VolumeLevel   float64 `json:"volume_level"`   // 0-1
	SpatialMode   string `json:"spatial_mode"`   // mono, stereo, spatial
	EffectsEnabled bool  `json:"effects_enabled"`
	ReverbLevel   float64 `json:"reverb_level"`   // 0-1
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

// NewVoiceProfile creates a new VoiceProfile with default values.
func NewVoiceProfile() *VoiceProfile {
	now := time.Now()
	return &VoiceProfile{
		VoiceCharacteristics: VoiceCharacteristics{
			VoiceType:         "female",
			VoiceAge:          "adult",
			VoiceQuality:      "clear",
			BasePitch:         200.0,
			PitchRange:        12.0,
			PitchVariation:    "moderate",
			SpeakingRate:      150.0,
			RateVariation:     "varied",
			PauseFrequency:    0.3,
			BaseVolume:        0.8,
			VolumeRange:       0.3,
			VolumeVariation:   "steady",
			TimbreType:        "warm",
			Resonance:         "medium",
			Breathiness:       0.2,
			Roughness:         0.1,
			ArticulationStyle: "relaxed",
			AccentStrength:    0.5,
			Distinctiveness:   0.5,
		},
		VoiceStyle: VoiceStyle{
			RegionalAccent:     "standard",
			AccentIntensity:    0.5,
			FormalityLevel:     "neutral",
			CommunicationStyle: "direct",
			IndirectnessLevel:  0.3,
			ProfessionalVoice:  false,
			VoiceAuthority:     0.5,
			VoiceWarmth:        0.6,
			AdaptiveStyle:      true,
			StyleConsistency:   0.7,
		},
		EmotionalVoice: EmotionalVoice{
			EmotionMappings:          make(map[string]VoiceEmotionMapping),
			EmotionalExpressiveness:  0.6,
			EmotionalContagion:       0.4,
			PitchModulationRange:     6.0,
			RateModulationRange:      0.3,
			VolumeModulation:         0.3,
			EmotionRegulation:        "express",
			BaselineStability:        0.7,
		},
		SpeechPatterns: SpeechPatterns{
			VocabularyLevel:     "moderate",
			JargonUsage:         0.3,
			IdiomUsage:          0.4,
			MetaphorUsage:       0.3,
			SentenceLength:      "medium",
			SentenceComplexity:  "compound",
			ClauseUsage:         0.4,
			FillerWords:         []string{"嗯", "这个"},
			FillerFrequency:     0.2,
			HedgingLanguage:     0.3,
			QualifyingWords:     0.3,
			ThoughtOrganization: "linear",
			DigressionTendency:  0.3,
			TangentFrequency:    0.2,
			ListeningResponses:  []string{"嗯嗯", "好的"},
			ResponseFrequency:   0.5,
			InterruptTendency:   0.2,
			PauseBeforeSpeaking: 0.3,
			ThoughtfulPauses:    0.4,
			RhetoricalPauses:    0.3,
		},
		TTSConfig: TTSConfiguration{
			EngineType:      "cloud",
			EngineName:      "volcengine",
			SampleRate:      24000,
			BitDepth:        16,
			Channels:        1,
			AudioFormat:     "mp3",
			QualityLevel:    "high",
			LatencyMode:     "balanced",
			BufferSize:      200,
			SSMLEnabled:     false,
			StreamingEnabled: true,
			ChunkSize:       100,
			CacheEnabled:    true,
			CacheSize:       100,
			CacheTTL:        3600,
			FailoverEnabled: true,
		},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
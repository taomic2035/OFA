// Package tts provides Text-to-Speech synthesis capabilities (v5.6.0).
//
// The TTS engine integrates with multiple providers (Volcengine, Doubao, etc.)
// to synthesize speech from text with voice characteristics.
package tts

import (
	"context"
)

// TTSProvider defines the interface for TTS providers.
type TTSProvider interface {
	// Name returns the provider name.
	Name() string

	// Synthesize synthesizes speech from text (HTTP mode).
	Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error)

	// SynthesizeStream synthesizes speech in streaming mode (WebSocket).
	SynthesizeStream(ctx context.Context, req *SynthesisRequest) (<-chan AudioChunk, error)

	// ListVoices returns available voices.
	ListVoices(ctx context.Context) ([]VoiceInfo, error)

	// SupportsStreaming returns true if streaming is supported.
	SupportsStreaming() bool

	// SupportsCloning returns true if voice cloning is supported.
	SupportsCloning() bool

	// CloneVoice clones a voice from reference audio.
	CloneVoice(ctx context.Context, req *CloneRequest) (*CloneResult, error)
}

// SynthesisRequest represents a TTS synthesis request.
type SynthesisRequest struct {
	// Text to synthesize
	Text string `json:"text"`

	// Voice ID to use
	VoiceID string `json:"voice_id"`

	// Language code (zh-CN, en-US, etc.)
	Language string `json:"language"`

	// Voice parameters
	Pitch  float64 `json:"pitch"`  // 0.5-2.0, default 1.0
	Rate   float64 `json:"rate"`   // 0.5-2.0, default 1.0
	Volume float64 `json:"volume"` // 0-1.0, default 0.7

	// Emotion for emotional voices
	Emotion string `json:"emotion"` // happy, sad, angry, etc.

	// Style for style-based synthesis
	Style string `json:"style"`

	// Output settings
	OutputFormat string `json:"output_format"` // mp3, wav, ogg
	SampleRate   int    `json:"sample_rate"`   // 8000, 16000, 24000

	// Streaming mode
	Streaming bool `json:"streaming"`

	// Identity context (for soul integration)
	IdentityID string `json:"identity_id"`

	// Scene context
	Scene string `json:"scene"`
}

// SynthesisResult represents a TTS synthesis result.
type SynthesisResult struct {
	// Audio data
	AudioData []byte `json:"audio_data"`
	AudioURL  string `json:"audio_url"`

	// Audio metadata
	DurationMs int    `json:"duration_ms"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`

	// Synthesis info
	Provider    string  `json:"provider"`
	VoiceUsed   string  `json:"voice_used"`
	LatencyMs   int     `json:"latency_ms"`
	QualityScore float64 `json:"quality_score"`

	// Status
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AudioChunk represents a chunk of audio data in streaming mode.
type AudioChunk struct {
	// Audio data
	Data []byte

	// Chunk metadata
	Sequence   int
	IsLast     bool
	DurationMs int
}

// VoiceInfo represents information about a voice.
type VoiceInfo struct {
	VoiceID      string   `json:"voice_id"`
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	Gender       string   `json:"gender"` // male, female, neutral
	Age          string   `json:"age"`    // child, young, adult, senior
	Style        string   `json:"style"`
	Emotions     []string `json:"emotions"`
	Provider     string   `json:"provider"`
	Description  string   `json:"description"`
}

// CloneRequest represents a voice cloning request.
type CloneRequest struct {
	IdentityID string `json:"identity_id"`

	// Reference audio
	ReferenceAudios []ReferenceAudio `json:"reference_audios"`

	// Clone settings
	VoiceName string `json:"voice_name"`
	Language  string `json:"language"`
}

// ReferenceAudio represents a reference audio for voice cloning.
type ReferenceAudio struct {
	AudioURL      string `json:"audio_url"`
	DurationMs    int    `json:"duration_ms"`
	Transcription string `json:"transcription"`
}

// CloneResult represents a voice cloning result.
type CloneResult struct {
	VoiceID    string `json:"voice_id"`
	VoiceName  string `json:"voice_name"`
	Status     string `json:"status"` // pending, processing, ready, failed
	Quality    float64 `json:"quality"`
	Message    string `json:"message"`
}

// TTSEngineConfig holds configuration for the TTS engine.
type TTSEngineConfig struct {
	// Primary provider
	PrimaryProvider string `json:"primary_provider"`

	// Fallback provider
	FallbackProvider string `json:"fallback_provider"`

	// Cache settings
	EnableCache bool `json:"enable_cache"`
	CacheSizeMB int  `json:"cache_size_mb"`

	// Default settings
	DefaultVoice      string  `json:"default_voice"`
	DefaultFormat     string  `json:"default_format"`
	DefaultSampleRate int     `json:"default_sample_rate"`
	DefaultRate       float64 `json:"default_rate"`
	DefaultPitch      float64 `json:"default_pitch"`
	DefaultVolume     float64 `json:"default_volume"`

	// Volcengine settings
	VolcengineAppID string `json:"volcengine_app_id"`
	VolcengineToken string `json:"volcengine_token"`

	// Doubao settings
	DoubaoAppID string `json:"doubao_app_id"`
	DoubaoToken string `json:"doubao_token"`
}

// DefaultTTSEngineConfig returns default configuration.
func DefaultTTSEngineConfig() TTSEngineConfig {
	return TTSEngineConfig{
		PrimaryProvider:   "volcengine",
		FallbackProvider:  "doubao",
		EnableCache:       true,
		CacheSizeMB:       100,
		DefaultVoice:      "zh_female_tianmei",
		DefaultFormat:     "mp3",
		DefaultSampleRate: 24000,
		DefaultRate:       1.0,
		DefaultPitch:      1.0,
		DefaultVolume:     0.7,
	}
}
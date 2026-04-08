// Package service provides TTS service integration (v5.6.2).
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/ofa/center/internal/config"
	tts "github.com/ofa/center/internal/tts"
)

// TTSService manages TTS synthesis capabilities.
type TTSService struct {
	mu sync.RWMutex

	// TTS engine
	engine *tts.TTSEngine

	// Active synthesis sessions
	sessions map[string]*SynthesisSession

	// Voice-to-identity mappings from identity service
	identityVoiceMap map[string]string
}

// SynthesisSession represents an active synthesis session.
type SynthesisSession struct {
	ID        string
	IdentityID string
	VoiceID   string
	Text      string
	Status    string // pending, synthesizing, completed, failed
	CreatedAt int64
	UpdatedAt int64
}

// NewTTSService creates a new TTS service.
func NewTTSService(cfg *config.Config) *TTSService {
	// Build TTS engine config
	var ttsConfig tts.TTSEngineConfig
	if cfg != nil {
		ttsConfig = tts.TTSEngineConfig{
			PrimaryProvider:   cfg.TTS.PrimaryProvider,
			FallbackProvider:  cfg.TTS.FallbackProvider,
			EnableCache:       cfg.TTS.EnableCache,
			CacheSizeMB:       cfg.TTS.CacheSizeMB,
			DefaultVoice:      cfg.TTS.DefaultVoice,
			DefaultFormat:     cfg.TTS.DefaultFormat,
			DefaultSampleRate: cfg.TTS.DefaultSampleRate,
			DefaultRate:       cfg.TTS.DefaultRate,
			DefaultPitch:      cfg.TTS.DefaultPitch,
			DefaultVolume:     cfg.TTS.DefaultVolume,
			VolcengineAppID:   cfg.TTS.VolcengineAppID,
			VolcengineToken:   cfg.TTS.VolcengineToken,
			DoubaoAppID:       cfg.TTS.DoubaoAppID,
			DoubaoToken:       cfg.TTS.DoubaoToken,
		}
	}

	// Use defaults if not configured
	if ttsConfig.PrimaryProvider == "" {
		ttsConfig = tts.DefaultTTSEngineConfig()
	}

	service := &TTSService{
		engine:            tts.NewTTSEngine(ttsConfig),
		sessions:          make(map[string]*SynthesisSession),
		identityVoiceMap:  make(map[string]string),
	}

	log.Printf("TTS Service initialized with provider: %s", ttsConfig.PrimaryProvider)

	return service
}

// SynthesizeRequest represents a synthesis API request.
type SynthesizeRequest struct {
	IdentityID string  `json:"identity_id"`
	Text       string  `json:"text"`
	VoiceID    string  `json:"voice_id,omitempty"`
	Format     string  `json:"format,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Rate       float64 `json:"rate,omitempty"`
	Pitch      float64 `json:"pitch,omitempty"`
	Volume     float64 `json:"volume,omitempty"`
	Emotion    string  `json:"emotion,omitempty"`
	Style      string  `json:"style,omitempty"`
	Streaming  bool    `json:"streaming,omitempty"`
}

// SynthesizeResponse represents a synthesis API response.
type SynthesizeResponse struct {
	SessionID  string `json:"session_id"`
	AudioURL   string `json:"audio_url,omitempty"`
	AudioData  []byte `json:"audio_data,omitempty"`
	DurationMs int    `json:"duration_ms"`
	Format     string `json:"format"`
	VoiceUsed  string `json:"voice_used"`
	Provider   string `json:"provider"`
	LatencyMs  int    `json:"latency_ms"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

// Synthesize synthesizes speech from text.
func (s *TTSService) Synthesize(ctx context.Context, req *SynthesizeRequest) (*SynthesizeResponse, error) {
	s.mu.Lock()
	sessionID := generateSessionID()
	session := &SynthesisSession{
		ID:         sessionID,
		IdentityID: req.IdentityID,
		Text:       req.Text,
		Status:     "synthesizing",
		CreatedAt:  nowMs(),
		UpdatedAt:  nowMs(),
	}
	s.sessions[sessionID] = session
	s.mu.Unlock()

	// Get voice for identity
	voiceID := req.VoiceID
	if voiceID == "" && req.IdentityID != "" {
		voiceID = s.GetVoiceForIdentity(req.IdentityID)
	}

	// Build TTS request
	ttsReq := &tts.SynthesisRequest{
		Text:         req.Text,
		VoiceID:      voiceID,
		OutputFormat: req.Format,
		SampleRate:   req.SampleRate,
		Rate:         req.Rate,
		Pitch:        req.Pitch,
		Volume:       req.Volume,
		Emotion:      req.Emotion,
		Style:        req.Style,
		Streaming:    req.Streaming,
		IdentityID:   req.IdentityID,
	}

	// Synthesize
	result, err := s.engine.Synthesize(ctx, ttsReq)
	if err != nil {
		s.mu.Lock()
		session.Status = "failed"
		session.UpdatedAt = nowMs()
		s.mu.Unlock()

		return &SynthesizeResponse{
			SessionID: sessionID,
			Success:   false,
			Error:     err.Error(),
		}, nil
	}

	// Update session
	s.mu.Lock()
	session.Status = "completed"
	session.VoiceID = result.VoiceUsed
	session.UpdatedAt = nowMs()
	s.mu.Unlock()

	return &SynthesizeResponse{
		SessionID:  sessionID,
		AudioData:  result.AudioData,
		DurationMs: result.DurationMs,
		Format:     result.Format,
		VoiceUsed:  result.VoiceUsed,
		Provider:   result.Provider,
		LatencyMs:  result.LatencyMs,
		Success:    true,
	}, nil
}

// StreamChunk represents a streaming audio chunk.
type StreamChunk struct {
	SessionID string `json:"session_id"`
	Sequence  int    `json:"sequence"`
	Data      []byte `json:"data"`
	IsLast    bool   `json:"is_last"`
}

// SynthesizeStream synthesizes speech in streaming mode.
func (s *TTSService) SynthesizeStream(ctx context.Context, req *SynthesizeRequest) (<-chan StreamChunk, error) {
	// Get voice for identity
	voiceID := req.VoiceID
	if voiceID == "" && req.IdentityID != "" {
		voiceID = s.GetVoiceForIdentity(req.IdentityID)
	}

	ttsReq := &tts.SynthesisRequest{
		Text:         req.Text,
		VoiceID:      voiceID,
		OutputFormat: req.Format,
		SampleRate:   req.SampleRate,
		Rate:         req.Rate,
		Pitch:        req.Pitch,
		Volume:       req.Volume,
		Emotion:      req.Emotion,
		Style:        req.Style,
		Streaming:    true,
		IdentityID:   req.IdentityID,
	}

	// Start streaming
	chunkChan, err := s.engine.SynthesizeStream(ctx, ttsReq)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outputChan := make(chan StreamChunk, 100)

	// Session ID for this stream
	sessionID := generateSessionID()

	// Convert chunks
	go func() {
		defer close(outputChan)
		for chunk := range chunkChan {
			outputChan <- StreamChunk{
				SessionID: sessionID,
				Sequence:  chunk.Sequence,
				Data:      chunk.Data,
				IsLast:    chunk.IsLast,
			}
		}
	}()

	return outputChan, nil
}

// VoiceInfo represents voice information for API response.
type VoiceInfo struct {
	VoiceID     string   `json:"voice_id"`
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	Gender      string   `json:"gender"`
	Age         string   `json:"age,omitempty"`
	Style       string   `json:"style,omitempty"`
	Emotions    []string `json:"emotions,omitempty"`
	Provider    string   `json:"provider"`
	Description string   `json:"description,omitempty"`
}

// ListVoices returns available voices.
func (s *TTSService) ListVoices(ctx context.Context, provider string) ([]VoiceInfo, error) {
	voices, err := s.engine.ListVoices(ctx, provider)
	if err != nil {
		return nil, err
	}

	// Convert to API response format
	result := make([]VoiceInfo, len(voices))
	for i, v := range voices {
		result[i] = VoiceInfo{
			VoiceID:     v.VoiceID,
			Name:        v.Name,
			Language:    v.Language,
			Gender:      v.Gender,
			Age:         v.Age,
			Style:       v.Style,
			Emotions:    v.Emotions,
			Provider:    v.Provider,
			Description: v.Description,
		}
	}

	return result, nil
}

// CloneVoiceRequest represents a voice cloning request.
type CloneVoiceRequest struct {
	IdentityID       string           `json:"identity_id"`
	VoiceName        string           `json:"voice_name"`
	Language         string           `json:"language"`
	ReferenceAudios  []ReferenceAudio `json:"reference_audios"`
}

// ReferenceAudio represents reference audio for cloning.
type ReferenceAudio struct {
	AudioURL      string `json:"audio_url"`
	DurationMs    int    `json:"duration_ms"`
	Transcription string `json:"transcription"`
}

// CloneVoiceResponse represents a voice cloning response.
type CloneVoiceResponse struct {
	VoiceID   string  `json:"voice_id"`
	VoiceName string  `json:"voice_name"`
	Status    string  `json:"status"`
	Quality   float64 `json:"quality"`
	Message   string  `json:"message"`
}

// CloneVoice clones a voice from reference audio.
func (s *TTSService) CloneVoice(ctx context.Context, req *CloneVoiceRequest) (*CloneVoiceResponse, error) {
	// Build clone request
	cloneReq := &tts.CloneRequest{
		IdentityID:  req.IdentityID,
		VoiceName:   req.VoiceName,
		Language:    req.Language,
		ReferenceAudios: make([]tts.ReferenceAudio, len(req.ReferenceAudios)),
	}

	for i, ref := range req.ReferenceAudios {
		cloneReq.ReferenceAudios[i] = tts.ReferenceAudio{
			AudioURL:      ref.AudioURL,
			DurationMs:    ref.DurationMs,
			Transcription: ref.Transcription,
		}
	}

	// Clone voice
	result, err := s.engine.CloneVoice(ctx, cloneReq)
	if err != nil {
		return nil, err
	}

	// Update identity voice mapping if ready
	if result.Status == "ready" {
		s.SetVoiceForIdentity(req.IdentityID, result.VoiceID)
	}

	return &CloneVoiceResponse{
		VoiceID:   result.VoiceID,
		VoiceName: result.VoiceName,
		Status:    result.Status,
		Quality:   result.Quality,
		Message:   result.Message,
	}, nil
}

// SetVoiceForIdentity sets the voice for an identity.
func (s *TTSService) SetVoiceForIdentity(identityID, voiceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identityVoiceMap[identityID] = voiceID
	s.engine.SetVoice(identityID, voiceID)
}

// GetVoiceForIdentity gets the voice for an identity.
func (s *TTSService) GetVoiceForIdentity(identityID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if voiceID, ok := s.identityVoiceMap[identityID]; ok {
		return voiceID
	}
	return s.engine.GetVoice(identityID)
}

// GetSession gets a synthesis session.
func (s *TTSService) GetSession(sessionID string) *SynthesisSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// GetEngine returns the underlying TTS engine.
func (s *TTSService) GetEngine() *tts.TTSEngine {
	return s.engine
}

// Helper functions

func generateSessionID() string {
	return "tts_" + randomHex(8)
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
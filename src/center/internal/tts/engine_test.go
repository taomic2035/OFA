package tts

import (
	"context"
	"testing"
)

// MockProvider for testing
type MockProvider struct {
	name         string
	supportsStream bool
	supportsClone bool
	voices       []VoiceInfo
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error) {
	return &SynthesisResult{
		AudioData:    []byte("mock_audio"),
		DurationMs:   1000,
		Format:       req.OutputFormat,
		SampleRate:   req.SampleRate,
		Provider:     m.name,
		VoiceUsed:    req.VoiceID,
		LatencyMs:    100,
		QualityScore: 0.9,
		Success:      true,
	}, nil
}

func (m *MockProvider) SynthesizeStream(ctx context.Context, req *SynthesisRequest) (<-chan AudioChunk, error) {
	ch := make(chan AudioChunk, 1)
	go func() {
		ch <- AudioChunk{Data: []byte("mock_chunk"), Sequence: 1, IsLast: true}
		close(ch)
	}()
	return ch, nil
}

func (m *MockProvider) ListVoices(ctx context.Context) ([]VoiceInfo, error) {
	return m.voices, nil
}

func (m *MockProvider) SupportsStreaming() bool {
	return m.supportsStream
}

func (m *MockProvider) SupportsCloning() bool {
	return m.supportsClone
}

func (m *MockProvider) CloneVoice(ctx context.Context, req *CloneRequest) (*CloneResult, error) {
	return &CloneResult{
		VoiceID:   "cloned_voice_" + req.IdentityID,
		VoiceName: req.VoiceName,
		Status:    "ready",
		Quality:   0.85,
	}, nil
}

func TestNewTTSEngine(t *testing.T) {
	config := DefaultTTSEngineConfig()
	engine := NewTTSEngine(config)

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}
}

func TestTTSEngineSynthesize(t *testing.T) {
	config := TTSEngineConfig{
		PrimaryProvider:   "mock1",
		FallbackProvider:  "mock2",
		EnableCache:       false,
		DefaultVoice:      "test_voice",
		DefaultFormat:     "mp3",
		DefaultSampleRate: 24000,
		DefaultRate:       1.0,
		DefaultPitch:      1.0,
		DefaultVolume:     0.7,
	}
	engine := NewTTSEngine(config)

	// Add mock provider
	mock := &MockProvider{
		name:           "mock1",
		supportsStream: true,
		supportsClone:  true,
		voices: []VoiceInfo{
			{VoiceID: "test_voice", Name: "Test Voice", Provider: "mock1"},
		},
	}
	engine.providers["mock1"] = mock
	engine.primary = mock

	req := &SynthesisRequest{
		Text:         "Hello, world!",
		VoiceID:      "test_voice",
		OutputFormat: "mp3",
		SampleRate:   24000,
	}

	result, err := engine.Synthesize(context.Background(), req)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	if !result.Success {
		t.Error("Result should be successful")
	}
	if result.Provider != "mock1" {
		t.Errorf("Provider mismatch: got %s, expected mock1", result.Provider)
	}
	if len(result.AudioData) == 0 {
		t.Error("AudioData should not be empty")
	}
}

func TestTTSEngineSynthesizeStream(t *testing.T) {
	config := TTSEngineConfig{
		PrimaryProvider:   "mock1",
		EnableCache:       false,
		DefaultVoice:      "test_voice",
		DefaultFormat:     "mp3",
		DefaultSampleRate: 24000,
	}
	engine := NewTTSEngine(config)

	mock := &MockProvider{
		name:           "mock1",
		supportsStream: true,
	}
	engine.providers["mock1"] = mock
	engine.primary = mock

	req := &SynthesisRequest{
		Text:         "Streaming test",
		VoiceID:      "test_voice",
		OutputFormat: "mp3",
		SampleRate:   24000,
	}

	ch, err := engine.SynthesizeStream(context.Background(), req)
	if err != nil {
		t.Fatalf("SynthesizeStream failed: %v", err)
	}

	// Read chunks
	var chunks []AudioChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		t.Error("Should receive at least one chunk")
	}
}

func TestTTSEngineCache(t *testing.T) {
	config := TTSEngineConfig{
		PrimaryProvider:   "mock1",
		EnableCache:       true,
		CacheSizeMB:       10,
		DefaultVoice:      "test_voice",
		DefaultFormat:     "mp3",
		DefaultSampleRate: 24000,
		DefaultRate:       1.0,
		DefaultPitch:      1.0,
		DefaultVolume:     0.7,
	}
	engine := NewTTSEngine(config)

	mock := &MockProvider{name: "mock1"}
	engine.providers["mock1"] = mock
	engine.primary = mock

	req := &SynthesisRequest{
		Text:         "Cached text",
		VoiceID:      "test_voice",
		OutputFormat: "mp3",
		SampleRate:   24000,
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
	}

	// First call
	result1, err := engine.Synthesize(context.Background(), req)
	if err != nil {
		t.Fatalf("First Synthesize failed: %v", err)
	}

	// Second call should be cached
	result2, err := engine.Synthesize(context.Background(), req)
	if err != nil {
		t.Fatalf("Second Synthesize failed: %v", err)
	}

	// Cached result should have "(cached)" suffix
	if result2.Provider != "mock1 (cached)" {
		t.Errorf("Second call should be cached: got %s", result2.Provider)
	}
}

func TestTTSEngineListVoices(t *testing.T) {
	config := TTSEngineConfig{
		PrimaryProvider: "mock1",
	}
	engine := NewTTSEngine(config)

	mock := &MockProvider{
		name: "mock1",
		voices: []VoiceInfo{
			{VoiceID: "voice1", Name: "Voice One", Provider: "mock1"},
			{VoiceID: "voice2", Name: "Voice Two", Provider: "mock1"},
		},
	}
	engine.providers["mock1"] = mock

	voices, err := engine.ListVoices(context.Background(), "mock1")
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}

	if len(voices) != 2 {
		t.Errorf("Voice count mismatch: got %d, expected 2", len(voices))
	}
}

func TestTTSEngineVoiceMapping(t *testing.T) {
	config := TTSEngineConfig{
		DefaultVoice: "default_voice",
	}
	engine := NewTTSEngine(config)

	// Set voice mapping
	engine.SetVoice("identity1", "custom_voice")

	// Get voice for identity
	voice := engine.GetVoice("identity1")
	if voice != "custom_voice" {
		t.Errorf("Voice mapping mismatch: got %s, expected custom_voice", voice)
	}

	// Get default voice for unknown identity
	voice = engine.GetVoice("unknown")
	if voice != "default_voice" {
		t.Errorf("Default voice mismatch: got %s, expected default_voice", voice)
	}
}

func TestTTSEngineCloneVoice(t *testing.T) {
	config := TTSEngineConfig{}
	engine := NewTTSEngine(config)

	mock := &MockProvider{
		name:          "mock1",
		supportsClone: true,
	}
	engine.providers["mock1"] = mock

	req := &CloneRequest{
		IdentityID:    "identity1",
		VoiceName:     "My Voice",
		Language:      "zh-CN",
		ReferenceAudios: []ReferenceAudio{
			{AudioURL: "http://example.com/ref.mp3", DurationMs: 5000},
		},
	}

	result, err := engine.CloneVoice(context.Background(), req)
	if err != nil {
		t.Fatalf("CloneVoice failed: %v", err)
	}

	if result.Status != "ready" {
		t.Errorf("Clone status mismatch: got %s, expected ready", result.Status)
	}

	// Check voice mapping was updated
	voice := engine.GetVoice("identity1")
	if voice != result.VoiceID {
		t.Errorf("Voice mapping not updated: got %s, expected %s", voice, result.VoiceID)
	}
}

func TestApplyDefaults(t *testing.T) {
	config := TTSEngineConfig{
		DefaultVoice:      "default_voice",
		DefaultFormat:     "wav",
		DefaultSampleRate: 16000,
		DefaultRate:       1.2,
		DefaultPitch:      0.9,
		DefaultVolume:     0.5,
	}
	engine := NewTTSEngine(config)

	req := &SynthesisRequest{
		Text: "Test",
	}
	req = engine.applyDefaults(req)

	if req.VoiceID != "default_voice" {
		t.Errorf("VoiceID default mismatch: got %s", req.VoiceID)
	}
	if req.OutputFormat != "wav" {
		t.Errorf("Format default mismatch: got %s", req.OutputFormat)
	}
	if req.SampleRate != 16000 {
		t.Errorf("SampleRate default mismatch: got %d", req.SampleRate)
	}
	if req.Rate != 1.2 {
		t.Errorf("Rate default mismatch: got %.2f", req.Rate)
	}
	if req.Pitch != 0.9 {
		t.Errorf("Pitch default mismatch: got %.2f", req.Pitch)
	}
	if req.Volume != 0.5 {
		t.Errorf("Volume default mismatch: got %.2f", req.Volume)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	engine := NewTTSEngine(DefaultTTSEngineConfig())

	req1 := &SynthesisRequest{
		Text:         "Hello",
		VoiceID:      "voice1",
		OutputFormat: "mp3",
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
	}

	req2 := &SynthesisRequest{
		Text:         "Hello",
		VoiceID:      "voice1",
		OutputFormat: "mp3",
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
	}

	req3 := &SynthesisRequest{
		Text:         "Different",
		VoiceID:      "voice1",
		OutputFormat: "mp3",
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
	}

	key1 := engine.generateCacheKey(req1)
	key2 := engine.generateCacheKey(req2)
	key3 := engine.generateCacheKey(req3)

	// Same request should produce same key
	if key1 != key2 {
		t.Error("Same requests should produce same cache key")
	}

	// Different text should produce different key
	if key1 == key3 {
		t.Error("Different requests should produce different cache keys")
	}
}
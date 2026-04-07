// Package tts provides Text-to-Speech synthesis capabilities (v5.6.0).
package tts

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"ofa/center/internal/tts/providers"
)

// TTSEngine manages TTS synthesis.
type TTSEngine struct {
	mu sync.RWMutex

	// Configuration
	config TTSEngineConfig

	// Providers
	providers map[string]TTSProvider

	// Primary and fallback
	primary  TTSProvider
	fallback TTSProvider

	// Cache
	cache *synthesisCache

	// Voice mappings (identity -> voice_id)
	voiceMappings map[string]string
}

// NewTTSEngine creates a new TTS engine.
func NewTTSEngine(config TTSEngineConfig) *TTSEngine {
	engine := &TTSEngine{
		config:        config,
		providers:     make(map[string]TTSProvider),
		voiceMappings: make(map[string]string),
	}

	// Initialize cache
	if config.EnableCache {
		engine.cache = newSynthesisCache(config.CacheSizeMB)
	}

	// Initialize providers
	engine.initProviders()

	return engine
}

// initProviders initializes TTS providers.
func (e *TTSEngine) initProviders() {
	// Initialize Volcengine provider
	if e.config.VolcengineAppID != "" && e.config.VolcengineToken != "" {
		volcengine := providers.NewVolcengineProvider(e.config.VolcengineAppID, e.config.VolcengineToken)
		e.providers["volcengine"] = volcengine
	}

	// Initialize Doubao provider
	if e.config.DoubaoAppID != "" && e.config.DoubaoToken != "" {
		doubao := providers.NewDoubaoProvider(e.config.DoubaoAppID, e.config.DoubaoToken)
		e.providers["doubao"] = doubao
	}

	// Set primary provider
	if p, ok := e.providers[e.config.PrimaryProvider]; ok {
		e.primary = p
	}

	// Set fallback provider
	if p, ok := e.providers[e.config.FallbackProvider]; ok {
		e.fallback = p
	}
}

// === Synthesis ===

// Synthesize synthesizes speech from text.
func (e *TTSEngine) Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error) {
	// Apply defaults
	req = e.applyDefaults(req)

	// Check cache
	if e.cache != nil {
		cacheKey := e.generateCacheKey(req)
		if cached := e.cache.Get(cacheKey); cached != nil {
			cached.Provider = cached.Provider + " (cached)"
			return cached, nil
		}
	}

	// Try primary provider
	var result *SynthesisResult
	var err error

	if e.primary != nil {
		result, err = e.primary.Synthesize(ctx, req)
		if err == nil {
			// Cache the result
			if e.cache != nil {
				cacheKey := e.generateCacheKey(req)
				e.cache.Set(cacheKey, result)
			}
			return result, nil
		}
	}

	// Try fallback provider
	if e.fallback != nil && e.fallback != e.primary {
		result, err = e.fallback.Synthesize(ctx, req)
		if err == nil {
			if e.cache != nil {
				cacheKey := e.generateCacheKey(req)
				e.cache.Set(cacheKey, result)
			}
			return result, nil
		}
	}

	// All providers failed
	return nil, fmt.Errorf("all TTS providers failed: %v", err)
}

// SynthesizeStream synthesizes speech in streaming mode.
func (e *TTSEngine) SynthesizeStream(ctx context.Context, req *SynthesisRequest) (<-chan AudioChunk, error) {
	// Apply defaults
	req = e.applyDefaults(req)
	req.Streaming = true

	// Try primary provider
	if e.primary != nil && e.primary.SupportsStreaming() {
		return e.primary.SynthesizeStream(ctx, req)
	}

	// Try fallback provider
	if e.fallback != nil && e.fallback.SupportsStreaming() {
		return e.fallback.SynthesizeStream(ctx, req)
	}

	return nil, fmt.Errorf("no streaming-capable provider available")
}

// === Voice Management ===

// ListVoices returns available voices.
func (e *TTSEngine) ListVoices(ctx context.Context, provider string) ([]VoiceInfo, error) {
	if provider == "" {
		provider = e.config.PrimaryProvider
	}

	p, ok := e.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", provider)
	}

	return p.ListVoices(ctx)
}

// SetVoice sets the voice for an identity.
func (e *TTSEngine) SetVoice(identityID string, voiceID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.voiceMappings[identityID] = voiceID
}

// GetVoice gets the voice for an identity.
func (e *TTSEngine) GetVoice(identityID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if voiceID, ok := e.voiceMappings[identityID]; ok {
		return voiceID
	}
	return e.config.DefaultVoice
}

// === Voice Cloning ===

// CloneVoice clones a voice from reference audio.
func (e *TTSEngine) CloneVoice(ctx context.Context, req *CloneRequest) (*CloneResult, error) {
	// Try providers that support cloning
	for _, p := range e.providers {
		if p.SupportsCloning() {
			result, err := p.CloneVoice(ctx, req)
			if err == nil {
				// Map identity to cloned voice
				if result.Status == "ready" {
					e.SetVoice(req.IdentityID, result.VoiceID)
				}
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("no cloning-capable provider available")
}

// === Helper Methods ===

// applyDefaults applies default values to the request.
func (e *TTSEngine) applyDefaults(req *SynthesisRequest) *SynthesisRequest {
	if req.VoiceID == "" {
		req.VoiceID = e.config.DefaultVoice
	}
	if req.OutputFormat == "" {
		req.OutputFormat = e.config.DefaultFormat
	}
	if req.SampleRate == 0 {
		req.SampleRate = e.config.DefaultSampleRate
	}
	if req.Rate == 0 {
		req.Rate = e.config.DefaultRate
	}
	if req.Pitch == 0 {
		req.Pitch = e.config.DefaultPitch
	}
	if req.Volume == 0 {
		req.Volume = e.config.DefaultVolume
	}
	return req
}

// generateCacheKey generates a cache key for the request.
func (e *TTSEngine) generateCacheKey(req *SynthesisRequest) string {
	data := fmt.Sprintf("%s:%s:%s:%.2f:%.2f:%.2f",
		req.Text, req.VoiceID, req.OutputFormat,
		req.Rate, req.Pitch, req.Volume)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// === Cache ===

type synthesisCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	size  int
	maxMB int
}

type cacheItem struct {
	result    *SynthesisResult
	createdAt time.Time
	size      int
}

func newSynthesisCache(maxMB int) *synthesisCache {
	return &synthesisCache{
		items: make(map[string]*cacheItem),
		maxMB: maxMB,
	}
}

func (c *synthesisCache) Get(key string) *SynthesisResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil
	}

	// Check expiration (1 hour)
	if time.Since(item.createdAt) > time.Hour {
		return nil
	}

	return item.result
}

func (c *synthesisCache) Set(key string, result *SynthesisResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: remove oldest if over limit
	itemSize := len(result.AudioData)
	if c.size+itemSize > c.maxMB*1024*1024 {
		// Remove half of items
		count := 0
		for k := range c.items {
			delete(c.items, k)
			count++
			if count > len(c.items)/2 {
				break
			}
		}
	}

	c.items[key] = &cacheItem{
		result:    result,
		createdAt: time.Now(),
		size:      itemSize,
	}
	c.size += itemSize
}
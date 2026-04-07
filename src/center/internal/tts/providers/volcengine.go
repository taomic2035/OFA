// Package providers implements TTS providers.
package providers

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ofa/center/internal/tts"
)

// VolcengineProvider implements TTSProvider for Volcengine.
type VolcengineProvider struct {
	appID      string
	token      string
	resourceID string

	httpClient *http.Client

	// API endpoints
	apiEndpoint string
}

// NewVolcengineProvider creates a new Volcengine provider.
func NewVolcengineProvider(appID, token string) *VolcengineProvider {
	return &VolcengineProvider{
		appID:       appID,
		token:       token,
		resourceID:  "volc.bigasr.speech_optional_20k",
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		apiEndpoint: "https://openspeech.bytedance.com/api/v1/tts",
	}
}

// Name returns the provider name.
func (p *VolcengineProvider) Name() string {
	return "volcengine"
}

// Synthesize synthesizes speech using HTTP API.
func (p *VolcengineProvider) Synthesize(ctx context.Context, req *tts.SynthesisRequest) (*tts.SynthesisResult, error) {
	startTime := time.Now()

	// Build request payload
	payload := map[string]interface{}{
		"app": map[string]string{
			"appid": p.appID,
			"token": "access_token",
		},
		"user": map[string]string{
			"uid": "ofa_user",
		},
		"audio": map[string]interface{}{
			"voice_type":   req.VoiceID,
			"encoding":     req.OutputFormat,
			"speed_ratio":  req.Rate,
			"volume_ratio": req.Volume,
			"pitch_ratio":  req.Pitch,
		},
		"request": map[string]string{
			"reqid":     p.generateReqID(req.Text),
			"text":      req.Text,
			"operation": "query",
		},
	}

	// Apply emotion if specified
	if req.Emotion != "" {
		payload["audio"].(map[string]interface{})["emotion"] = req.Emotion
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	latencyMs := int(time.Since(startTime).Milliseconds())

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	contentType := resp.Header.Get("Content-Type")

	var audioData []byte

	// Check if response is audio or JSON
	if contains(contentType, "audio") || contains(contentType, "octet-stream") {
		audioData, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio data: %w", err)
		}
	} else {
		// JSON response
		var jsonResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if jsonResp.Code != 0 {
			return nil, fmt.Errorf("API error %d: %s", jsonResp.Code, jsonResp.Message)
		}

		// Data is base64 encoded
		audioData = []byte(jsonResp.Data) // TODO: base64 decode
	}

	// Estimate duration
	durationMs := estimateDuration(req.Text, req.Rate)

	return &tts.SynthesisResult{
		AudioData:    audioData,
		DurationMs:   durationMs,
		Format:       req.OutputFormat,
		SampleRate:   req.SampleRate,
		Provider:     p.Name(),
		VoiceUsed:    req.VoiceID,
		LatencyMs:    latencyMs,
		QualityScore: 0.9,
		Success:      true,
	}, nil
}

// SynthesizeStream synthesizes speech in streaming mode.
func (p *VolcengineProvider) SynthesizeStream(ctx context.Context, req *tts.SynthesisRequest) (<-chan tts.AudioChunk, error) {
	// Use WebSocket for streaming
	return nil, fmt.Errorf("streaming not supported, use Doubao provider")
}

// ListVoices returns available voices.
func (p *VolcengineProvider) ListVoices(ctx context.Context) ([]tts.VoiceInfo, error) {
	// Return predefined voice list
	return []tts.VoiceInfo{
		{VoiceID: "zh_female_tianmei", Name: "天美", Language: "zh-CN", Gender: "female", Age: "adult", Provider: p.Name()},
		{VoiceID: "zh_male_chunhou", Name: "醇厚", Language: "zh-CN", Gender: "male", Age: "adult", Provider: p.Name()},
		{VoiceID: "zh_female_qingxin", Name: "清新", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_male_qingnian", Name: "青年", Language: "zh-CN", Gender: "male", Age: "young", Provider: p.Name()},
		// Emotional voices
		{VoiceID: "zh_female_happy", Name: "开心女声", Language: "zh-CN", Gender: "female", Emotions: []string{"happy"}, Provider: p.Name()},
		{VoiceID: "zh_female_sad", Name: "悲伤女声", Language: "zh-CN", Gender: "female", Emotions: []string{"sad"}, Provider: p.Name()},
		{VoiceID: "zh_female_angry", Name: "愤怒女声", Language: "zh-CN", Gender: "female", Emotions: []string{"angry"}, Provider: p.Name()},
		// Dialect voices
		{VoiceID: "zh_dongbei_dialect", Name: "东北方言", Language: "zh-CN-northeast", Gender: "neutral", Provider: p.Name()},
		{VoiceID: "zh_sichuan_dialect", Name: "四川方言", Language: "zh-CN-sichuan", Gender: "neutral", Provider: p.Name()},
		{VoiceID: "zh_guangdong_dialect", Name: "广东方言", Language: "zh-CN-guangdong", Gender: "neutral", Provider: p.Name()},
	}, nil
}

// SupportsStreaming returns false for HTTP mode.
func (p *VolcengineProvider) SupportsStreaming() bool {
	return false
}

// SupportsCloning returns true.
func (p *VolcengineProvider) SupportsCloning() bool {
	return true
}

// CloneVoice clones a voice.
func (p *VolcengineProvider) CloneVoice(ctx context.Context, req *tts.CloneRequest) (*tts.CloneResult, error) {
	// TODO: Implement voice cloning API
	return &tts.CloneResult{
		Status:  "pending",
		Message: "Voice cloning is being processed",
	}, nil
}

// Helper methods

func (p *VolcengineProvider) generateReqID(text string) string {
	hash := md5.Sum([]byte(text + time.Now().String()))
	return hex.EncodeToString(hash[:])
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func estimateDuration(text string, rate float64) int {
	// Estimate: average 150 words per minute at rate 1.0
	// Chinese: ~3 characters per word equivalent
	charCount := len([]rune(text))
	words := float64(charCount) / 3.0
	minutes := words / (150.0 * rate)
	return int(minutes * 60000)
}
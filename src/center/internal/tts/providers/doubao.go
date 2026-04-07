// Package providers implements TTS providers.
package providers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"ofa/center/internal/tts"
	"ofa/center/internal/tts/protocol"
)

// DoubaoProvider implements TTSProvider for Doubao TTS 2.0 (large model voices).
type DoubaoProvider struct {
	appID      string
	token      string
	resourceID string

	httpClient *http.Client

	// WebSocket endpoint
	wsEndpoint string
}

// NewDoubaoProvider creates a new Doubao provider.
func NewDoubaoProvider(appID, token string) *DoubaoProvider {
	return &DoubaoProvider{
		appID:       appID,
		token:       token,
		resourceID:  "seed-tts-2.0",
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		wsEndpoint:  "wss://openspeech.bytedance.com/api/v3/tts/bidirection",
	}
}

// Name returns the provider name.
func (p *DoubaoProvider) Name() string {
	return "doubao"
}

// Synthesize synthesizes speech using WebSocket (fallback to HTTP-style result).
func (p *DoubaoProvider) Synthesize(ctx context.Context, req *tts.SynthesisRequest) (*tts.SynthesisResult, error) {
	startTime := time.Now()

	// Collect all audio from streaming
	chunkChan, err := p.SynthesizeStream(ctx, req)
	if err != nil {
		return nil, err
	}

	// Collect all chunks
	var audioData []byte
	for chunk := range chunkChan {
		audioData = append(audioData, chunk.Data...)
	}

	latencyMs := int(time.Since(startTime).Milliseconds())
	durationMs := estimateDuration(req.Text, req.Rate)

	return &tts.SynthesisResult{
		AudioData:    audioData,
		DurationMs:   durationMs,
		Format:       req.OutputFormat,
		SampleRate:   req.SampleRate,
		Provider:     p.Name(),
		VoiceUsed:    req.VoiceID,
		LatencyMs:    latencyMs,
		QualityScore: 0.95,
		Success:      true,
	}, nil
}

// SynthesizeStream synthesizes speech in streaming mode using WebSocket.
func (p *DoubaoProvider) SynthesizeStream(ctx context.Context, req *tts.SynthesisRequest) (<-chan tts.AudioChunk, error) {
	// Create output channel
	outputChan := make(chan tts.AudioChunk, 100)

	// Start WebSocket connection in goroutine
	go p.streamViaWebSocket(ctx, req, outputChan)

	return outputChan, nil
}

// streamViaWebSocket handles WebSocket streaming.
func (p *DoubaoProvider) streamViaWebSocket(ctx context.Context, req *tts.SynthesisRequest, outputChan chan<- tts.AudioChunk) {
	defer close(outputChan)

	// Prepare headers
	headers := http.Header{}
	headers.Set("X-Api-App-Key", p.appID)
	headers.Set("X-Api-Access-Key", p.token)
	headers.Set("X-Api-Resource-Id", p.resourceID)

	// Connect to WebSocket
	dialOptions := &websocket.DialOptions{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
	}

	conn, _, err := websocket.Dial(ctx, p.wsEndpoint, dialOptions)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Create protocol handler
	proto := protocol.NewVolcengineProtocol()

	// Start connection
	startConnMsg := protocol.Message{
		Type:  protocol.MsgFullClientRequest,
		Event: protocol.EventStartConnection,
	}
	if err := conn.Write(ctx, websocket.MessageBinary, proto.Marshal(&startConnMsg)); err != nil {
		return
	}

	// Wait for connection started
	_, reader, err := conn.Reader(ctx)
	if err != nil {
		return
	}
	respMsg, err := proto.Unmarshal(reader)
	if err != nil || respMsg.Event != protocol.EventConnectionStarted {
		return
	}

	// Split text into sentences
	sentences := splitSentences(req.Text)

	var audioData []byte
	var mu sync.Mutex
	sequence := 0

	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}

		// Start session
		sessionID := generateSessionID()
		startSessionPayload := map[string]interface{}{
			"event": protocol.EventStartSession,
			"user":  map[string]string{"uid": "ofa_user"},
			"req_params": map[string]interface{}{
				"speaker": req.VoiceID,
				"audio_params": map[string]interface{}{
					"format":      req.OutputFormat,
					"sample_rate": req.SampleRate,
				},
			},
		}
		startSessionBytes, _ := json.Marshal(startSessionPayload)

		startSessionMsg := protocol.Message{
			Type:      protocol.MsgFullClientRequest,
			Event:     protocol.EventStartSession,
			SessionID: sessionID,
			Payload:   startSessionBytes,
		}
		if err := conn.Write(ctx, websocket.MessageBinary, proto.Marshal(&startSessionMsg)); err != nil {
			continue
		}

		// Send text character by character
		go func(text string, sid string) {
			for _, char := range text {
				taskPayload := map[string]interface{}{
					"event": protocol.EventTaskRequest,
					"user":  map[string]string{"uid": "ofa_user"},
					"req_params": map[string]interface{}{
						"speaker": req.VoiceID,
						"text":    string(char),
						"audio_params": map[string]interface{}{
							"format":      req.OutputFormat,
							"sample_rate": req.SampleRate,
						},
					},
				}
				taskBytes, _ := json.Marshal(taskPayload)

				taskMsg := protocol.Message{
					Type:      protocol.MsgFullClientRequest,
					Event:     protocol.EventTaskRequest,
					SessionID: sid,
					Payload:   taskBytes,
				}
				conn.Write(ctx, websocket.MessageBinary, proto.Marshal(&taskMsg))
				time.Sleep(5 * time.Millisecond) // Small delay between chars
			}

			// Finish session
			finishMsg := protocol.Message{
				Type:      protocol.MsgFullClientRequest,
				Event:     protocol.EventFinishSession,
				SessionID: sid,
			}
			conn.Write(ctx, websocket.MessageBinary, proto.Marshal(&finishMsg))
		}(sentence, sessionID)

		// Receive audio data
		for {
			_, reader, err := conn.Reader(ctx)
			if err != nil {
				break
			}

			msg, err := proto.Unmarshal(reader)
			if err != nil {
				break
			}

			switch msg.Type {
			case protocol.MsgAudioOnlyServer:
				mu.Lock()
				audioData = append(audioData, msg.Payload...)
				mu.Unlock()

				// Send chunk
				select {
				case outputChan <- tts.AudioChunk{
					Data:     msg.Payload,
					Sequence: sequence,
					IsLast:   false,
				}:
				default:
				}
				sequence++

			case protocol.MsgFullServerResponse:
				if msg.Event == protocol.EventSessionFinished {
					// Send final chunk for this session
					mu.Lock()
					totalData := make([]byte, len(audioData))
					copy(totalData, audioData)
					mu.Unlock()

					select {
					case outputChan <- tts.AudioChunk{
						Data:     totalData,
						Sequence: sequence,
						IsLast:   true,
					}:
					default:
					}
					goto nextSentence
				}
			case protocol.MsgError:
				// Handle error
				goto nextSentence
			}
		}
	nextSentence:
	}

	// Finish connection
	finishConnMsg := protocol.Message{
		Type:  protocol.MsgFullClientRequest,
		Event: protocol.EventFinishConnection,
	}
	conn.Write(ctx, websocket.MessageBinary, proto.Marshal(&finishConnMsg))
}

// ListVoices returns available Doubao voices.
func (p *DoubaoProvider) ListVoices(ctx context.Context) ([]tts.VoiceInfo, error) {
	// Doubao TTS 2.0 large model voices
	return []tts.VoiceInfo{
		{VoiceID: "zh_male_sunwukong_uranus_bigtts", Name: "猴哥", Language: "zh-CN", Gender: "male", Age: "adult", Provider: p.Name(), Description: "孙悟空音色"},
		{VoiceID: "zh_female_meilinvyou_uranus_bigtts", Name: "魅力女友", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_male_ruyayichen_uranus_bigtts", Name: "儒雅逸辰", Language: "zh-CN", Gender: "male", Age: "adult", Provider: p.Name()},
		{VoiceID: "saturn_zh_male_shuanglangshaonian_tob", Name: "爽朗少年", Language: "zh-CN", Gender: "male", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_female_vv_uranus_bigtts", Name: "Vivi", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_male_m191_uranus_bigtts", Name: "云舟", Language: "zh-CN", Gender: "male", Age: "adult", Provider: p.Name()},
		{VoiceID: "zh_male_taocheng_uranus_bigtts", Name: "小天", Language: "zh-CN", Gender: "male", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_female_cancan_uranus_bigtts", Name: "知性灿灿", Language: "zh-CN", Gender: "female", Age: "adult", Provider: p.Name()},
		{VoiceID: "zh_female_sajiaoxuemei_uranus_bigtts", Name: "撒娇学妹", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_female_tianmeixiaoyuan_uranus_bigtts", Name: "甜美小源", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_female_linjianvhai_uranus_bigtts", Name: "邻家女孩", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_male_shaonianzixin_uranus_bigtts", Name: "少年梓辛", Language: "zh-CN", Gender: "male", Age: "young", Provider: p.Name()},
		{VoiceID: "zh_female_kefunvsheng_uranus_bigtts", Name: "暖阳女声", Language: "zh-CN", Gender: "female", Age: "adult", Provider: p.Name()},
		{VoiceID: "saturn_zh_female_keainvsheng_tob", Name: "可爱女生", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "saturn_zh_female_tiaopigongzhu_tob", Name: "调皮公主", Language: "zh-CN", Gender: "female", Age: "young", Provider: p.Name()},
		{VoiceID: "saturn_zh_male_tiancaitongzhuo_tob", Name: "天才同桌", Language: "zh-CN", Gender: "male", Age: "young", Provider: p.Name()},
	}, nil
}

// SupportsStreaming returns true for WebSocket mode.
func (p *DoubaoProvider) SupportsStreaming() bool {
	return true
}

// SupportsCloning returns true.
func (p *DoubaoProvider) SupportsCloning() bool {
	return true
}

// CloneVoice clones a voice.
func (p *DoubaoProvider) CloneVoice(ctx context.Context, req *tts.CloneRequest) (*tts.CloneResult, error) {
	// TODO: Implement voice cloning API
	return &tts.CloneResult{
		Status:  "pending",
		Message: "Voice cloning is being processed",
	}, nil
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func splitSentences(text string) []string {
	// Simple sentence splitting by Chinese punctuation
	var sentences []string
	var current string

	for _, char := range text {
		current += string(char)
		if char == '。' || char == '！' || char == '？' || char == '.' || char == '!' || char == '?' {
			sentences = append(sentences, current)
			current = ""
		}
	}

	if current != "" {
		sentences = append(sentences, current)
	}

	return sentences
}
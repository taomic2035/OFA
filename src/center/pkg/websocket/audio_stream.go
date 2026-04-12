// Package websocket provides audio streaming over WebSocket (v8.0.1).
package websocket

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/tts"
)

// AudioStreamManager manages audio streaming over WebSocket.
type AudioStreamManager struct {
	mu sync.RWMutex

	// TTS engine
	ttsEngine *tts.TTSEngine

	// Active streams (stream_id -> *AudioStream)
	activeStreams sync.Map

	// Stream counter
	streamCounter int64
}

// AudioStream represents an active audio stream.
type AudioStream struct {
	StreamID    string
	IdentityID  string
	AgentID     string
	Text        string
	VoiceID     string
	Speed       float64
	StartTime   time.Time
	EndTime     *time.Time
	TotalSize   int
	Status      string // starting, streaming, completed, error
}

// AudioStreamStartPayload represents audio stream start message.
type AudioStreamStartPayload struct {
	StreamID   string `json:"stream_id"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	VoiceID    string `json:"voice_id,omitempty"`
}

// AudioChunkPayload represents audio chunk message.
type AudioChunkPayload struct {
	StreamID string `json:"stream_id"`
	Data     string `json:"data"` // base64 encoded
	Index    int    `json:"index"`
}

// AudioEndPayload represents audio stream end message.
type AudioEndPayload struct {
	StreamID  string `json:"stream_id"`
	TotalSize int    `json:"total_size"`
	Duration  float64 `json:"duration"`
}

// NewAudioStreamManager creates a new audio stream manager.
func NewAudioStreamManager(ttsEngine *tts.TTSEngine) *AudioStreamManager {
	return &AudioStreamManager{
		ttsEngine: ttsEngine,
	}
}

// StartStream starts a new audio stream for TTS.
func (m *AudioStreamManager) StartStream(identityID, agentID, text, voiceID string, speed float64) (*AudioStream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate stream ID
	m.streamCounter++
	streamID := fmt.Sprintf("stream_%d_%d", time.Now().UnixNano(), m.streamCounter)

	// Create stream
	stream := &AudioStream{
		StreamID:   streamID,
		IdentityID: identityID,
		AgentID:    agentID,
		Text:       text,
		VoiceID:    voiceID,
		Speed:      speed,
		StartTime:  time.Now(),
		Status:     "starting",
	}

	// Store stream
	m.activeStreams.Store(streamID, stream)

	return stream, nil
}

// StreamTTS streams TTS audio to an agent.
func (m *AudioStreamManager) StreamTTS(streamID string, sendChunk func(msgType string, payload json.RawMessage) error) error {
	stream, ok := m.activeStreams.Load(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	audioStream := stream.(*AudioStream)
	audioStream.Status = "streaming"

	// Send stream start
	startPayload := AudioStreamStartPayload{
		StreamID:   streamID,
		Format:     "pcm",
		SampleRate: 24000,
		VoiceID:    audioStream.VoiceID,
	}
	startData, _ := json.Marshal(startPayload)
	sendChunk("audio_stream", startData)

	// Synthesize TTS
	ctx := context.Background()
	audioData, err := m.ttsEngine.Synthesize(ctx, audioStream.Text, audioStream.VoiceID, tts.SynthesisOptions{
		Speed:     audioStream.Speed,
		Format:    "pcm",
		SampleRate: 24000,
	})
	if err != nil {
		audioStream.Status = "error"
		return fmt.Errorf("TTS synthesis failed: %w", err)
	}

	// Send chunks (split into smaller chunks for streaming)
	chunkSize := 4096
	index := 0

	for offset := 0; offset < len(audioData); offset += chunkSize {
		end := offset + chunkSize
		if end > len(audioData) {
			end = len(audioData)
		}

		chunk := audioData[offset:end]

		// Encode chunk as base64
		chunkPayload := AudioChunkPayload{
			StreamID: streamID,
			Data:     base64.StdEncoding.EncodeToString(chunk),
			Index:    index,
		}
		chunkData, _ := json.Marshal(chunkPayload)

		if err := sendChunk("audio_chunk", chunkData); err != nil {
			audioStream.Status = "error"
			return fmt.Errorf("failed to send chunk: %w", err)
		}

		audioStream.TotalSize += len(chunk)
		index++
	}

	// Send stream end
	now := time.Now()
	audioStream.EndTime = &now
	audioStream.Status = "completed"
	duration := now.Sub(audioStream.StartTime).Seconds()

	endPayload := AudioEndPayload{
		StreamID:  streamID,
		TotalSize: audioStream.TotalSize,
		Duration:  duration,
	}
	endData, _ := json.Marshal(endPayload)
	sendChunk("audio_end", endData)

	// Remove from active streams
	m.activeStreams.Delete(streamID)

	return nil
}

// GetStream returns an active stream.
func (m *AudioStreamManager) GetStream(streamID string) *AudioStream {
	stream, ok := m.activeStreams.Load(streamID)
	if !ok {
		return nil
	}
	return stream.(*AudioStream)
}

// GetActiveStreams returns all active streams.
func (m *AudioStreamManager) GetActiveStreams() []*AudioStream {
	var streams []*AudioStream
	m.activeStreams.Range(func(key, value interface{}) bool {
		streams = append(streams, value.(*AudioStream))
		return true
	})
	return streams
}

// CancelStream cancels an active stream.
func (m *AudioStreamManager) CancelStream(streamID string) error {
	stream, ok := m.activeStreams.Load(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	audioStream := stream.(*AudioStream)
	audioStream.Status = "cancelled"
	m.activeStreams.Delete(streamID)

	return nil
}

// HandleTTSRequest handles TTS request from WebSocket.
func (m *AudioStreamManager) HandleTTSRequest(identityID, agentID string, payload map[string]interface{}, sendChunk func(msgType string, payload json.RawMessage) error) error {
	text, ok := payload["text"].(string)
	if !ok || text == "" {
		return fmt.Errorf("text is required for TTS request")
	}

	voiceID := ""
	if v, ok := payload["voice_id"].(string); ok {
		voiceID = v
	}

	speed := 1.0
	if s, ok := payload["speed"].(float64); ok {
		speed = s
	}

	// Start stream
	stream, err := m.StartStream(identityID, agentID, text, voiceID, speed)
	if err != nil {
		return err
	}

	// Stream TTS
	return m.StreamTTS(stream.StreamID, sendChunk)
}

// GetStatistics returns streaming statistics.
func (m *AudioStreamManager) GetStatistics() map[string]interface{} {
	var activeCount int
	var completedCount int
	var totalBytes int

	m.activeStreams.Range(func(key, value interface{}) bool {
		stream := value.(*AudioStream)
		if stream.Status == "streaming" {
			activeCount++
		}
		totalBytes += stream.TotalSize
		return true
	})

	return map[string]interface{}{
		"active_streams": activeCount,
		"total_bytes":    totalBytes,
		"stream_counter": m.streamCounter,
	}
}
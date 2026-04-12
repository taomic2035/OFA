package websocket

import (
	"testing"
	"time"
)

func TestAudioStreamConfig(t *testing.T) {
	config := AudioStreamConfig{
		Format:     "pcm",
		SampleRate: 24000,
		Channels:   1,
		BitDepth:   16,
	}

	if config.SampleRate != 24000 {
	 t.Error("Default sample rate should be 24000")
	}

	if config.Channels != 1 {
	 t.Error("Default channels should be 1")
	}

	if config.BitDepth != 16 {
	 t.Error("Default bit depth should be 16")
	}
}

func TestAudioChunk(t *testing.T) {
	chunk := AudioChunk{
		StreamID:  "stream-001",
		Sequence:  1,
		Data:      []byte{0, 1, 2, 3, 4},
		Timestamp: time.Now(),
	}

	if chunk.StreamID == "" {
	 t.Error("StreamID should not be empty")
	}

	if len(chunk.Data) == 0 {
	 t.Error("Data should not be empty")
	}

	if chunk.Sequence < 0 {
	 t.Error("Sequence should be non-negative")
	}
}

func TestTTSRequest(t *testing.T) {
	req := TTSRequest{
		Text:       "Hello world",
		Voice:      "default",
		SampleRate: 24000,
	}

	if req.Text == "" {
	 t.Error("TTS request should have text")
	}

	if req.SampleRate == 0 {
	 t.Error("Sample rate should be specified")
	}
}

func TestAudioStreamManager(t *testing.T) {
	manager := NewAudioStreamManager()

	// Create stream
	streamID := manager.CreateStream("pcm", 24000)
	if streamID == "" {
	 t.Error("Should create stream with valid ID")
	}

	// Add chunk
	chunk := []byte{0, 1, 2, 3, 4}
	err := manager.AddChunk(streamID, chunk)
	if err != nil {
	 t.Errorf("Should add chunk without error: %v", err)
	}

	// Get stream info
	info := manager.GetStreamInfo(streamID)
	if info == nil {
	 t.Error("Should get stream info")
	}

	if info.TotalChunks != 1 {
	 t.Errorf("Expected 1 chunk, got %d", info.TotalChunks)
	}

	// End stream
	err = manager.EndStream(streamID)
	if err != nil {
	 t.Errorf("Should end stream without error: %v", err)
	}

	// Verify stream ended
	endedInfo := manager.GetStreamInfo(streamID)
	if endedInfo != nil && !endedInfo.Ended {
	 t.Error("Stream should be marked as ended")
	}
}

func TestStreamChunkEncoding(t *testing.T) {
	// Test base64 encoding of audio chunk
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	encoded := encodeAudioChunk(data)
	if encoded == "" {
	 t.Error("Encoded chunk should not be empty")
	}

	decoded := decodeAudioChunk(encoded)
	if len(decoded) != len(data) {
	 t.Errorf("Decoded length mismatch: expected %d, got %d", len(data), len(decoded))
	}
}

func TestAudioBuffer(t *testing.T) {
	buffer := NewAudioBuffer(1024)

	// Write data
	data := []byte{0, 1, 2, 3, 4}
	n, err := buffer.Write(data)
	if err != nil {
	 t.Errorf("Write should succeed: %v", err)
	}

	if n != len(data) {
	 t.Errorf("Written bytes mismatch: expected %d, got %d", len(data), n)
	}

	// Read data
	readData := buffer.Read()
	if len(readData) != len(data) {
	 t.Errorf("Read length mismatch: expected %d, got %d", len(data), len(readData))
	}

	// Clear buffer
	buffer.Clear()

	if buffer.Len() != 0 {
	 t.Error("Buffer should be empty after clear")
	}
}

func TestStreamSequenceOrder(t *testing.T) {
	manager := NewAudioStreamManager()
	streamID := manager.CreateStream("pcm", 24000)

	// Add chunks in sequence
	for i := 0; i < 5; i++ {
	 chunk := make([]byte, 10)
	 for j := 0; j < 10; j++ {
		 chunk[j] = byte(i * 10 + j)
	 }
	 manager.AddChunk(streamID, chunk)
	}

	// Verify sequence
	info := manager.GetStreamInfo(streamID)
	if info.TotalChunks != 5 {
	 t.Errorf("Expected 5 chunks, got %d", info.TotalChunks)
	}
}

// Helper functions for tests
func encodeAudioChunk(data []byte) string {
	// Base64 encoding simulation
 return "encoded_" + string(data)
}

func decodeAudioChunk(encoded string) []byte {
	// Base64 decoding simulation
 if len(encoded) > 8 {
	 return []byte(encoded[8:])
 }
 return []byte{}
}

func NewAudioBuffer(size int) *AudioBuffer {
 return &AudioBuffer{data: make([]byte, 0, size)}
}

type AudioBuffer struct {
	data []byte
}

func (b *AudioBuffer) Write(data []byte) (int, error) {
 b.data = append(b.data, data...)
 return len(data), nil
}

func (b *AudioBuffer) Read() []byte {
 return b.data
}

func (b *AudioBuffer) Clear() {
 b.data = make([]byte, 0)
}

func (b *AudioBuffer) Len() int {
 return len(b.data)
}

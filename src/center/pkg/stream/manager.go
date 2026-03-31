package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// StreamType defines stream types
type StreamType string

const (
	StreamTypeTask   StreamType = "task"
		StreamTypeLog    StreamType = "log"
	StreamTypeMetrics StreamType = "metrics"
	StreamTypeEvent  StreamType = "event"
)

// StreamState defines stream states
type StreamState string

const (
	StreamStateActive   StreamState = "active"
	StreamStatePaused   StreamState = "paused"
	StreamStateClosed   StreamState = "closed"
	StreamStateError    StreamState = "error"
)

// StreamMessage represents a stream message
type StreamMessage struct {
	StreamID  string      `json:"stream_id"`
	Type      StreamType  `json:"type"`
	Sequence  int64       `json:"sequence"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
}

// StreamConfig holds stream configuration
type StreamConfig struct {
	BufferSize      int
	MaxStreams      int
	Timeout         time.Duration
	FlushInterval   time.Duration
	MaxSequenceAge  time.Duration // Maximum age before sequence reset
}

// DefaultStreamConfig returns default configuration
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		BufferSize:      100,
		MaxStreams:      1000,
		Timeout:         30 * time.Minute,
		FlushInterval:   100 * time.Millisecond,
		MaxSequenceAge:  24 * time.Hour,
	}
}

// Stream represents an active stream
type Stream struct {
	ID        string
	Type      StreamType
	State     StreamState
	CreatedAt time.Time
	LastSeq   int64
	LastFlush time.Time

	buffer    []StreamMessage
	subs      map[string]*Subscriber
	taskID    string
	agentID   string

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// Subscriber represents a stream subscriber
type Subscriber struct {
	ID        string
	LastSeq   int64
	Active    bool
	CreatedAt time.Time
	callback  func(StreamMessage)
}

// StreamManager manages stream processing
type StreamManager struct {
	config *StreamConfig

	streams sync.Map // map[string]*Stream

	// Task executor for stream tasks
	executor StreamExecutor

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// StreamExecutor executes stream tasks
type StreamExecutor interface {
	ExecuteStreamTask(taskID string, params map[string]interface{}) (<-chan interface{}, error)
}

// NewStreamManager creates a new stream manager
func NewStreamManager(config *StreamConfig) (*StreamManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &StreamManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	return manager, nil
}

// Start begins stream processing
func (m *StreamManager) Start() {
	go m.flushLoop()
	go m.cleanupLoop()
}

// Stop stops the stream manager
func (m *StreamManager) Stop() {
	m.cancel()
}

// SetExecutor sets the stream executor
func (m *StreamManager) SetExecutor(executor StreamExecutor) {
	m.executor = executor
}

// CreateStream creates a new stream
func (m *StreamManager) CreateStream(streamType StreamType, taskID, agentID string) (*Stream, error) {
	// Check max streams limit
	count := 0
	m.streams.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	if count >= m.config.MaxStreams {
		return nil, errors.New("maximum streams limit reached")
	}

	streamID := generateStreamID()
	ctx, cancel := context.WithCancel(m.ctx)

	stream := &Stream{
		ID:        streamID,
		Type:      streamType,
		State:     StreamStateActive,
		CreatedAt: time.Now(),
		buffer:    make([]StreamMessage, 0, m.config.BufferSize),
		subs:      make(map[string]*Subscriber),
		taskID:    taskID,
		agentID:   agentID,
		ctx:       ctx,
		cancel:    cancel,
	}

	m.streams.Store(streamID, stream)

	return stream, nil
}

// GetStream returns a stream by ID
func (m *StreamManager) GetStream(streamID string) (*Stream, bool) {
	if v, ok := m.streams.Load(streamID); ok {
		return v.(*Stream), true
	}
	return nil, false
}

// CloseStream closes a stream
func (m *StreamManager) CloseStream(streamID string) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	stream.State = StreamStateClosed
	stream.cancel()
	stream.mu.Unlock()

	m.streams.Delete(streamID)

	return nil
}

// Subscribe subscribes to a stream
func (m *StreamManager) Subscribe(streamID string, callback func(StreamMessage)) (string, error) {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return "", fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.State != StreamStateActive {
		return "", errors.New("stream not active")
	}

	subID := generateSubscriberID()
	sub := &Subscriber{
		ID:        subID,
		LastSeq:   stream.LastSeq,
		Active:    true,
		CreatedAt: time.Now(),
		callback:  callback,
	}

	stream.subs[subID] = sub

	return subID, nil
}

// Unsubscribe unsubscribes from a stream
func (m *StreamManager) Unsubscribe(streamID, subID string) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	delete(stream.subs, subID)
	stream.mu.Unlock()

	return nil
}

// PushMessage pushes a message to a stream
func (m *StreamManager) PushMessage(streamID string, data interface{}) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.State != StreamStateActive {
		return errors.New("stream not active")
	}

	stream.LastSeq++
	msg := StreamMessage{
		StreamID:  streamID,
		Type:      stream.Type,
		Sequence:  stream.LastSeq,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Add to buffer
	if len(stream.buffer) < m.config.BufferSize {
		stream.buffer = append(stream.buffer, msg)
	} else {
		// Buffer overflow - remove oldest
		stream.buffer = stream.buffer[1:]
		stream.buffer = append(stream.buffer, msg)
	}

	// Notify subscribers
	for _, sub := range stream.subs {
		if sub.Active && sub.callback != nil {
			go sub.callback(msg)
			sub.LastSeq = msg.Sequence
		}
	}

	return nil
}

// PushError pushes an error message to a stream
func (m *StreamManager) PushError(streamID string, errMsg string) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	defer stream.mu.Unlock()

	stream.State = StreamStateError

	stream.LastSeq++
	msg := StreamMessage{
		StreamID:  streamID,
		Type:      stream.Type,
		Sequence:  stream.LastSeq,
		Timestamp: time.Now(),
		Error:     errMsg,
	}

	stream.buffer = append(stream.buffer, msg)

	// Notify subscribers of error
	for _, sub := range stream.subs {
		if sub.Active && sub.callback != nil {
			go sub.callback(msg)
		}
	}

	return nil
}

// GetMessages gets messages from a stream since a sequence
func (m *StreamManager) GetMessages(streamID string, sinceSeq int64) ([]StreamMessage, error) {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.RLock()
	defer stream.mu.RUnlock()

	var messages []StreamMessage
	for _, msg := range stream.buffer {
		if msg.Sequence > sinceSeq {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// PauseStream pauses a stream
func (m *StreamManager) PauseStream(streamID string) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	stream.State = StreamStatePaused
	stream.mu.Unlock()

	return nil
}

// ResumeStream resumes a paused stream
func (m *StreamManager) ResumeStream(streamID string) error {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.Lock()
	if stream.State != StreamStatePaused {
		stream.mu.Unlock()
		return errors.New("stream not paused")
	}
	stream.State = StreamStateActive
	stream.mu.Unlock()

	return nil
}

// flushLoop periodically flushes stream buffers
func (m *StreamManager) flushLoop() {
	ticker := time.NewTicker(m.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.flushBuffers()
		}
	}
}

// flushBuffers flushes all stream buffers
func (m *StreamManager) flushBuffers() {
	m.streams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)

		stream.mu.Lock()
		if time.Since(stream.LastFlush) > m.config.FlushInterval {
			// Flush to storage or external system
			if len(stream.buffer) > 0 {
				log.Printf("Flushing %d messages for stream %s", len(stream.buffer), stream.ID)
				stream.buffer = make([]StreamMessage, 0, m.config.BufferSize)
			}
			stream.LastFlush = time.Now()
		}
		stream.mu.Unlock()

		return true
	})
}

// cleanupLoop periodically cleans up old streams
func (m *StreamManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupStreams()
		}
	}
}

// cleanupStreams removes closed/errored streams
func (m *StreamManager) cleanupStreams() {
	m.streams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)

		stream.mu.RLock()
		state := stream.State
		age := time.Since(stream.CreatedAt)
		stream.mu.RUnlock()

		// Remove closed/errored streams or old streams
		if state == StreamStateClosed || state == StreamStateError {
			m.streams.Delete(key)
			log.Printf("Removed stream %s (state: %s)", stream.ID, state)
		} else if age > m.config.Timeout {
			m.CloseStream(stream.ID)
			log.Printf("Timeout stream %s", stream.ID)
		}

		return true
	})
}

// ListStreams returns all active streams
func (m *StreamManager) ListStreams() []*Stream {
	var streams []*Stream
	m.streams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		stream.mu.RLock()
		if stream.State == StreamStateActive || stream.State == StreamStatePaused {
			streams = append(streams, stream)
		}
		stream.mu.RUnlock()
		return true
	})
	return streams
}

// GetStreamStats returns stream statistics
func (m *StreamManager) GetStreamStats(streamID string) (map[string]interface{}, error) {
	stream, ok := m.GetStream(streamID)
	if !ok {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	stream.mu.RLock()
	defer stream.mu.RUnlock()

	return map[string]interface{}{
		"id":         stream.ID,
		"type":       stream.Type,
		"state":      stream.State,
		"created_at": stream.CreatedAt,
		"sequence":   stream.LastSeq,
		"buffer_size": len(stream.buffer),
		"subscribers": len(stream.subs),
		"task_id":    stream.taskID,
		"agent_id":   stream.agentID,
	}, nil
}

// generateStreamID generates a unique stream ID
func generateStreamID() string {
	return "stream-" + time.Now().Format("20060102-150405") + "-" + randomSuffix()
}

// generateSubscriberID generates a unique subscriber ID
func generateSubscriberID() string {
	return "sub-" + time.Now().Format("20060102-150405") + "-" + randomSuffix()
}

// randomSuffix generates a random suffix
func randomSuffix() string {
	return fmt.Sprintf("%04d", time.Now().Nanosecond()%10000)
}

// StreamHandler handles WebSocket stream connections
func (m *StreamManager) StreamHandler(w http.ResponseWriter, r *http.Request) {
	// Check for WebSocket upgrade
	if r.Header.Get("Upgrade") != "websocket" {
		// HTTP streaming
		m.handleHTTPStream(w, r)
		return
	}

	// WebSocket streaming would be handled separately
	m.handleWebSocketStream(w, r)
}

// handleHTTPStream handles HTTP streaming
func (m *StreamManager) handleHTTPStream(w http.ResponseWriter, r *http.Request) {
	streamID := r.URL.Query().Get("stream_id")
	if streamID == "" {
		http.Error(w, "Missing stream_id", http.StatusBadRequest)
		return
	}

	stream, ok := m.GetStream(streamID)
	if !ok {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	// Set headers for streaming
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to stream
	var lastSeq int64
	seqStr := r.URL.Query().Get("since")
	if seqStr != "" {
		lastSeq = parseSequence(seqStr)
	}

	// Create subscriber
	subID, err := m.Subscribe(streamID, func(msg StreamMessage) {
		data, _ := json.Marshal(msg)
		w.Write(data)
		w.Write([]byte("\n"))
		flusher.Flush()
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send initial messages
	msgs, _ := m.GetMessages(streamID, lastSeq)
	for _, msg := range msgs {
		data, _ := json.Marshal(msg)
		w.Write(data)
		w.Write([]byte("\n"))
		flusher.Flush()
	}

	// Wait for stream to close or client disconnect
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		<-ctx.Done()
		m.Unsubscribe(streamID, subID)
	}()

	// Keep connection alive
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send keepalive
			w.Write([]byte("{\"type\":\"keepalive\"}\n"))
			flusher.Flush()
		}
	}
}

// handleWebSocketStream handles WebSocket streaming
func (m *StreamManager) handleWebSocketStream(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation would use gorilla/websocket or similar
	// This is a placeholder for the WebSocket upgrade logic
	http.Error(w, "WebSocket upgrade required", http.StatusBadRequest)
}

// parseSequence parses a sequence number from string
func parseSequence(s string) int64 {
	var seq int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			seq = seq*10 + int64(c-'0')
		}
	}
	return seq
}

// TaskStreamHandler handles streaming task execution
type TaskStreamHandler struct {
	manager  *StreamManager
	executor StreamExecutor
}

// NewTaskStreamHandler creates a task stream handler
func NewTaskStreamHandler(manager *StreamManager, executor StreamExecutor) *TaskStreamHandler {
	return &TaskStreamHandler{
		manager:  manager,
		executor: executor,
	}
}

// ExecuteStream executes a task and streams results
func (h *TaskStreamHandler) ExecuteStream(taskID string, params map[string]interface{}) (string, <-chan StreamMessage, error) {
	// Create stream
	stream, err := h.manager.CreateStream(StreamTypeTask, taskID, "")
	if err != nil {
		return "", nil, err
	}

	// Create output channel
	output := make(chan StreamMessage, h.manager.config.BufferSize)

	// Subscribe to stream
	_, err = h.manager.Subscribe(stream.ID, func(msg StreamMessage) {
		output <- msg
	})
	if err != nil {
		h.manager.CloseStream(stream.ID)
		return "", nil, err
	}

	// Execute task in background
	go func() {
		defer close(output)
		defer h.manager.CloseStream(stream.ID)

		// Get execution channel
		execChan, err := h.executor.ExecuteStreamTask(taskID, params)
		if err != nil {
			h.manager.PushError(stream.ID, err.Error())
			return
		}

		// Stream execution results
		for data := range execChan {
			err := h.manager.PushMessage(stream.ID, data)
			if err != nil {
				log.Printf("Failed to push message: %v", err)
				break
			}
		}

		// Send completion message
		h.manager.PushMessage(stream.ID, map[string]interface{}{
			"status": "completed",
			"task_id": taskID,
		})
	}()

	return stream.ID, output, nil
}
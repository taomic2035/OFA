// Package llm provides Claude API integration (v8.0.0).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ClaudeProvider implements LLMProvider for Anthropic Claude.
type ClaudeProvider struct {
	config LLMProviderConfig

	client *http.Client

	// Rate limiting
	requestTimes []time.Time
}

// ClaudeAPIRequest represents a Claude API request.
type ClaudeAPIRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
	Stream    bool            `json:"stream,omitempty"`
}

// ClaudeMessage represents a message for Claude API.
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeAPIResponse represents a Claude API response.
type ClaudeAPIResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []ClaudeContentBlock `json:"content"`
	Model   string `json:"model"`
	Usage   ClaudeUsage `json:"usage"`
}

// ClaudeContentBlock represents a content block in Claude response.
type ClaudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ClaudeUsage represents token usage from Claude.
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeStreamResponse represents a streaming response from Claude.
type ClaudeStreamResponse struct {
	Type         string `json:"type"`
	Index        int    `json:"index,omitempty"`
	Delta        ClaudeDelta `json:"delta,omitempty"`
	ContentBlock ClaudeContentBlock `json:"content_block,omitempty"`
	Message      ClaudeAPIResponse `json:"message,omitempty"`
	Usage        ClaudeUsage `json:"usage,omitempty"`
}

// ClaudeDelta represents a delta in streaming.
type ClaudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewClaudeProvider creates a new Claude provider.
func NewClaudeProvider(config LLMProviderConfig) (*ClaudeProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	if config.APIURL == "" {
		config.APIURL = "https://api.anthropic.com/v1/messages"
	}

	if config.Model == "" {
		config.Model = "claude-sonnet-4-20250514"
	}

	return &ClaudeProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		requestTimes: []time.Time{},
	}, nil
}

// Name returns the provider name.
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Generate generates a response.
func (p *ClaudeProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// Check rate limit
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build Claude API request
	claudeReq := ClaudeAPIRequest{
		Model:     p.config.Model,
		MaxTokens: req.MaxTokens,
		Messages:  p.convertMessages(req.Messages),
		System:    req.SystemPrompt,
		Stream:    false,
	}

	// Make request
	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API error: %s - %s", resp.Status, string(respBody))
	}

	// Parse response
	var claudeResp ClaudeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert response
	return &GenerateResponse{
		ConversationID: req.ConversationID,
		Content:        p.extractContent(claudeResp.Content),
		Role:           "assistant",
		TokenUsage: TokenUsage{
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		FinishReason: "stop",
		Model:        claudeResp.Model,
		Timestamp:    time.Now(),
	}, nil
}

// GenerateStream generates a streaming response.
func (p *ClaudeProvider) GenerateStream(ctx context.Context, req *GenerateRequest) (StreamReader, error) {
	// Check rate limit
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build Claude API request with streaming
	claudeReq := ClaudeAPIRequest{
		Model:     p.config.Model,
		MaxTokens: req.MaxTokens,
		Messages:  p.convertMessages(req.Messages),
		System:    req.SystemPrompt,
		Stream:    true,
	}

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Claude API error: %s - %s", resp.Status, string(respBody))
	}

	return NewClaudeStreamReader(resp.Body, req.ConversationID), nil
}

// CountTokens counts tokens (approximate).
func (p *ClaudeProvider) CountTokens(text string) (int, error) {
	// Approximate: 1 token ~ 4 characters for English
	// For Chinese: 1 token ~ 2 characters
	// This is a rough estimate
	chineseCount := 0
	englishCount := 0

	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		} else {
			englishCount++
		}
	}

	tokens := chineseCount/2 + englishCount/4
	return tokens + 1, nil
}

// GetMaxTokens returns max tokens.
func (p *ClaudeProvider) GetMaxTokens() int {
	return p.config.MaxTokens
}

// SupportsStreaming returns true.
func (p *ClaudeProvider) SupportsStreaming() bool {
	return true
}

// Close closes the provider.
func (p *ClaudeProvider) Close() error {
	return nil
}

// convertMessages converts messages to Claude format.
func (p *ClaudeProvider) convertMessages(messages []Message) []ClaudeMessage {
	result := []ClaudeMessage{}

	for _, msg := range messages {
		// Claude only supports user and assistant roles
		if msg.Role == "user" || msg.Role == "assistant" {
			result = append(result, ClaudeMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	return result
}

// extractContent extracts text content from Claude response.
func (p *ClaudeProvider) extractContent(blocks []ClaudeContentBlock) string {
	var texts []string
	for _, block := range blocks {
		if block.Type == "text" {
			texts = append(texts, block.Text)
		}
	}
	return strings.Join(texts, "\n")
}

// checkRateLimit checks rate limiting.
func (p *ClaudeProvider) checkRateLimit() error {
	now := time.Now()
	window := time.Minute

	// Remove old request times
	valid := []time.Time{}
	for _, t := range p.requestTimes {
		if now.Sub(t) < window {
			valid = append(valid, t)
		}
	}
	p.requestTimes = valid

	// Check limit
	if len(p.requestTimes) >= p.config.MaxRequestsPerMinute {
		return fmt.Errorf("rate limit exceeded: %d requests per minute", p.config.MaxRequestsPerMinute)
	}

	// Record this request
	p.requestTimes = append(p.requestTimes, now)
	return nil
}

// ClaudeStreamReader reads Claude streaming responses.
type ClaudeStreamReader struct {
	body       io.ReadCloser
	conversationID string
	buffer     []byte
	done       bool
	err        error
	tokenUsage *TokenUsage
}

// NewClaudeStreamReader creates a new Claude stream reader.
func NewClaudeStreamReader(body io.ReadCloser, conversationID string) *ClaudeStreamReader {
	return &ClaudeStreamReader{
		body:          body,
		conversationID: conversationID,
		buffer:        []byte{},
	}
}

// Read reads the next chunk.
func (r *ClaudeStreamReader) Read() (*StreamChunk, error) {
	if r.done || r.err != nil {
		return nil, io.EOF
	}

	// Read more data
	buf := make([]byte, 1024)
	n, err := r.body.Read(buf)
	if err != nil && err != io.EOF {
		r.err = err
		return nil, err
	}

	r.buffer = append(r.buffer, buf[:n]...)

	// Parse SSE events
	for {
		event, remaining, found := r.parseEvent()
		if !found {
			break
		}

		r.buffer = remaining

		// Parse event data
		chunk, final := r.processEvent(event)
		if chunk != nil {
			if final {
				r.done = true
			}
			return chunk, nil
		}
	}

	if err == io.EOF {
		r.done = true
		return nil, io.EOF
	}

	// Need more data
	return nil, nil
}

// Close closes the stream.
func (r *ClaudeStreamReader) Close() error {
	return r.body.Close()
}

// Err returns any error.
func (r *ClaudeStreamReader) Err() error {
	return r.err
}

// parseEvent parses a SSE event from buffer.
func (r *ClaudeStreamReader) parseEvent() (string, []byte, bool) {
	// Find event boundary (double newline)
	data := r.buffer
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\n' && data[i+1] == '\n' {
			event := string(data[:i])
			remaining := data[i+2:]
			return event, remaining, true
		}
	}
	return "", r.buffer, false
}

// processEvent processes a SSE event.
func (r *ClaudeStreamReader) processEvent(event string) (*StreamChunk, bool) {
	lines := strings.Split(event, "\n")
	var eventType string
	var eventData string

	for _, line := range lines {
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
		if strings.HasPrefix(line, "data:") {
			eventData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}

	if eventData == "" {
		return nil, false
	}

	var resp ClaudeStreamResponse
	if err := json.Unmarshal([]byte(eventData), &resp); err != nil {
		return nil, false
	}

	switch eventType {
	case "content_block_delta":
		if resp.Delta.Type == "text_delta" {
			return &StreamChunk{
				ConversationID: r.conversationID,
				Content:        resp.Delta.Text,
				IsFinal:        false,
				Timestamp:      time.Now(),
			}, false
		}

	case "message_stop":
		return &StreamChunk{
			ConversationID: r.conversationID,
			Content:        "",
			IsFinal:        true,
			FinishReason:   "stop",
			TokenUsage:     r.tokenUsage,
			Timestamp:      time.Now(),
		}, true

	case "message_delta":
		if resp.Usage.OutputTokens > 0 {
			r.tokenUsage = &TokenUsage{
				OutputTokens: resp.Usage.OutputTokens,
			}
		}
	}

	return nil, false
}
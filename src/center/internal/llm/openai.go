// Package llm provides OpenAI API integration (v8.0.0).
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

// OpenAIProvider implements LLMProvider for OpenAI GPT.
type OpenAIProvider struct {
	config LLMProviderConfig

	client *http.Client

	// Rate limiting
	requestTimes []time.Time
}

// OpenAIAPIRequest represents an OpenAI API request.
type OpenAIAPIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

// OpenAIMessage represents a message for OpenAI API.
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIAPIResponse represents an OpenAI API response.
type OpenAIAPIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage `json:"usage"`
}

// OpenAIChoice represents a choice in OpenAI response.
type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      OpenAIMessage  `json:"message,omitempty"`
	Delta        OpenAIDelta    `json:"delta,omitempty"`
	FinishReason string         `json:"finish_reason"`
}

// OpenAIDelta represents a delta in streaming.
type OpenAIDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// OpenAIUsage represents token usage from OpenAI.
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamResponse represents a streaming response from OpenAI.
type OpenAIStreamResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(config LLMProviderConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.APIURL == "" {
		config.APIURL = "https://api.openai.com/v1/chat/completions"
	}

	if config.Model == "" {
		config.Model = "gpt-4"
	}

	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		requestTimes: []time.Time{},
	}, nil
}

// Name returns the provider name.
func (p *OpenAIProvider) Name() string {
	return "gpt"
}

// Generate generates a response.
func (p *OpenAIProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// Check rate limit
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build OpenAI API request
	messages := p.buildMessages(req)

	openaiReq := OpenAIAPIRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
		Stop:        req.StopSequences,
	}

	// Make request
	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(respBody))
	}

	// Parse response
	var openaiResp OpenAIAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert response
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	finishReason := "stop"
	if choice.FinishReason != "" {
		finishReason = choice.FinishReason
	}

	return &GenerateResponse{
		ConversationID: req.ConversationID,
		Content:        choice.Message.Content,
		Role:           "assistant",
		TokenUsage: TokenUsage{
			InputTokens:  openaiResp.Usage.PromptTokens,
			OutputTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:  openaiResp.Usage.TotalTokens,
		},
		FinishReason: finishReason,
		Model:        openaiResp.Model,
		Timestamp:    time.Now(),
	}, nil
}

// GenerateStream generates a streaming response.
func (p *OpenAIProvider) GenerateStream(ctx context.Context, req *GenerateRequest) (StreamReader, error) {
	// Check rate limit
	if err := p.checkRateLimit(); err != nil {
		return nil, err
	}

	// Build OpenAI API request with streaming
	messages := p.buildMessages(req)

	openaiReq := OpenAIAPIRequest{
		Model:       p.config.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
		Stop:        req.StopSequences,
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(respBody))
	}

	return NewOpenAIStreamReader(resp.Body, req.ConversationID), nil
}

// CountTokens counts tokens (approximate).
func (p *OpenAIProvider) CountTokens(text string) (int, error) {
	// Use same approximation as Claude
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
func (p *OpenAIProvider) GetMaxTokens() int {
	return p.config.MaxTokens
}

// SupportsStreaming returns true.
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}

// Close closes the provider.
func (p *OpenAIProvider) Close() error {
	return nil
}

// buildMessages builds messages for OpenAI API.
func (p *OpenAIProvider) buildMessages(req *GenerateRequest) []OpenAIMessage {
	result := []OpenAIMessage{}

	// Add system prompt
	if req.SystemPrompt != "" {
		result = append(result, OpenAIMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range req.Messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			result = append(result, OpenAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	return result
}

// checkRateLimit checks rate limiting.
func (p *OpenAIProvider) checkRateLimit() error {
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

// OpenAIStreamReader reads OpenAI streaming responses.
type OpenAIStreamReader struct {
	body          io.ReadCloser
	conversationID string
	buffer        []byte
	done          bool
	err           error
	tokenUsage    *TokenUsage
}

// NewOpenAIStreamReader creates a new OpenAI stream reader.
func NewOpenAIStreamReader(body io.ReadCloser, conversationID string) *OpenAIStreamReader {
	return &OpenAIStreamReader{
		body:          body,
		conversationID: conversationID,
		buffer:        []byte{},
	}
}

// Read reads the next chunk.
func (r *OpenAIStreamReader) Read() (*StreamChunk, error) {
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
func (r *OpenAIStreamReader) Close() error {
	return r.body.Close()
}

// Err returns any error.
func (r *OpenAIStreamReader) Err() error {
	return r.err
}

// parseEvent parses a SSE event from buffer.
func (r *OpenAIStreamReader) parseEvent() (string, []byte, bool) {
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
func (r *OpenAIStreamReader) processEvent(event string) (*StreamChunk, bool) {
	lines := strings.Split(event, "\n")
	var eventData string

	for _, line := range lines {
		if strings.HasPrefix(line, "data:") {
			eventData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}

	if eventData == "" {
		return nil, false
	}

	if eventData == "[DONE]" {
		return &StreamChunk{
			ConversationID: r.conversationID,
			Content:        "",
			IsFinal:        true,
			FinishReason:   "stop",
			TokenUsage:     r.tokenUsage,
			Timestamp:      time.Now(),
		}, true
	}

	var resp OpenAIStreamResponse
	if err := json.Unmarshal([]byte(eventData), &resp); err != nil {
		return nil, false
	}

	if len(resp.Choices) == 0 {
		return nil, false
	}

	choice := resp.Choices[0]
	content := choice.Delta.Content

	finishReason := choice.FinishReason
	isFinal := finishReason != ""

	if content == "" && !isFinal {
		return nil, false
	}

	return &StreamChunk{
		ConversationID: r.conversationID,
		Content:        content,
		IsFinal:        isFinal,
		FinishReason:   finishReason,
		Timestamp:      time.Now(),
	}, isFinal
}
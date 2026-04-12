// Package rest provides Chat API endpoints (v8.0.0).
package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ofa/center/internal/llm"
)

// ChatAPI provides chat REST API endpoints.
type ChatAPI struct {
	llmEngine *llm.LLMEngine
}

// NewChatAPI creates a new chat API handler.
func NewChatAPI(llmEngine *llm.LLMEngine) *ChatAPI {
	return &ChatAPI{
		llmEngine: llmEngine,
	}
}

// ChatRequest represents a chat request.
type ChatRequest struct {
	IdentityID   string                 `json:"identity_id"`
	Message      string                 `json:"message"`
	Stream       bool                   `json:"stream,omitempty"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	ClearHistory bool                   `json:"clear_history,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat response.
type ChatResponse struct {
	ConversationID string                 `json:"conversation_id"`
	IdentityID     string                 `json:"identity_id"`
	Content        string                 `json:"content"`
	Role           string                 `json:"role"`
	TokenUsage     TokenUsageResponse     `json:"token_usage"`
	FinishReason   string                 `json:"finish_reason"`
	Model          string                 `json:"model"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsageResponse represents token usage in response.
type TokenUsageResponse struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// HistoryResponse represents conversation history.
type HistoryResponse struct {
	ConversationID string             `json:"conversation_id"`
	IdentityID     string             `json:"identity_id"`
	Messages       []MessageResponse  `json:"messages"`
	TokenCount     int                `json:"token_count"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// MessageResponse represents a message in history.
type MessageResponse struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Tokens    int       `json:"tokens,omitempty"`
}

// HandleChat handles chat requests.
// POST /api/v1/chat
func (a *ChatAPI) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	if req.IdentityID == "" {
		http.Error(w, "identity_id is required", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// Clear history if requested
	if req.ClearHistory {
		a.llmEngine.ClearConversation(req.IdentityID)
	}

	// Build personality context (would integrate with v4.x engines in production)
	personality := &llm.PersonalityContext{
		IdentityID: req.IdentityID,
	}

	// Generate response
	ctx := r.Context()
	response, err := a.llmEngine.Chat(ctx, req.IdentityID, req.Message, personality)
	if err != nil {
		http.Error(w, fmt.Sprintf("Chat generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	chatResp := ChatResponse{
		ConversationID: response.ConversationID,
		IdentityID:     response.IdentityID,
		Content:        response.Content,
		Role:           response.Role,
		TokenUsage: TokenUsageResponse{
			InputTokens:  response.TokenUsage.InputTokens,
			OutputTokens: response.TokenUsage.OutputTokens,
			TotalTokens:  response.TokenUsage.TotalTokens,
		},
		FinishReason: response.FinishReason,
		Model:        response.Model,
		Timestamp:    response.Timestamp,
		Metadata:     req.Metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

// HandleChatStream handles streaming chat requests.
// POST /api/v1/chat/stream
func (a *ChatAPI) HandleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	if req.IdentityID == "" {
		http.Error(w, "identity_id is required", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Build personality context
	personality := &llm.PersonalityContext{
		IdentityID: req.IdentityID,
	}

	// Generate streaming response
	ctx := r.Context()
	stream, err := a.llmEngine.ChatStream(ctx, req.IdentityID, req.Message, personality)
	if err != nil {
		http.Error(w, fmt.Sprintf("Chat stream generation failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Stream chunks
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		chunk, err := stream.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
			flusher.Flush()
			break
		}

		// Write SSE event
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data)
		flusher.Flush()

		if chunk.IsFinal {
			break
		}
	}

	// Send done event
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

// HandleChatHistory returns conversation history.
// GET /api/v1/chat/history/{identity_id}
func (a *ChatAPI) HandleChatHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract identity_id from URL
	identityID := extractPathParam(r.URL.Path, "/api/v1/chat/history/")
	if identityID == "" {
		http.Error(w, "identity_id is required", http.StatusBadRequest)
		return
	}

	// Get limit from query
	limit := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		var l int
		json.Unmarshal([]byte(limitStr), &l)
		limit = l
	}

	// Get conversation
	conv := a.llmEngine.GetConversation(identityID)
	if conv == nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Build response
	messages := a.llmEngine.GetConversationHistory(identityID, limit)

	msgResponses := []MessageResponse{}
	for _, msg := range messages {
		msgResponses = append(msgResponses, MessageResponse{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
			Tokens:    msg.Tokens,
		})
	}

	historyResp := HistoryResponse{
		ConversationID: conv.ID,
		IdentityID:     conv.IdentityID,
		Messages:       msgResponses,
		TokenCount:     conv.TokenCount,
		CreatedAt:      conv.CreatedAt,
		UpdatedAt:      conv.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(historyResp)
}

// HandleChatClear clears conversation history.
// DELETE /api/v1/chat/clear/{identity_id}
func (a *ChatAPI) HandleChatClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract identity_id from URL
	identityID := extractPathParam(r.URL.Path, "/api/v1/chat/clear/")
	if identityID == "" {
		http.Error(w, "identity_id is required", http.StatusBadRequest)
		return
	}

	// Clear conversation
	a.llmEngine.ClearConversation(identityID)

	// Build response
	resp := map[string]interface{}{
		"success":     true,
		"identity_id": identityID,
		"message":     "Conversation cleared",
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleIntentParse parses intent from message.
// POST /api/v1/chat/intent
func (a *ChatAPI) HandleIntentParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		Text string `json:"text"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	// Parse intent
	ctx := r.Context()
	intent, err := a.llmEngine.ParseIntent(ctx, req.Text)
	if err != nil {
		http.Error(w, fmt.Sprintf("Intent parsing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(intent)
}

// extractPathParam extracts a parameter from URL path.
func extractPathParam(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}

// RegisterChatRoutes registers chat routes on the server.
func (s *Server) RegisterChatRoutes(llmEngine *llm.LLMEngine) {
	chatAPI := NewChatAPI(llmEngine)

	// Chat routes
	s.router.HandleFunc("/api/v1/chat", chatAPI.HandleChat)
	s.router.HandleFunc("/api/v1/chat/stream", chatAPI.HandleChatStream)
	s.router.HandleFunc("/api/v1/chat/history/", chatAPI.HandleChatHistory)
	s.router.HandleFunc("/api/v1/chat/clear/", chatAPI.HandleChatClear)
	s.router.HandleFunc("/api/v1/chat/intent", chatAPI.HandleIntentParse)
}
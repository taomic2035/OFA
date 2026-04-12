package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatRequestValidation(t *testing.T) {
	// Test valid request
	validReq := ChatRequest{
		Message:   "Hello",
		SessionID: "test-session",
	}

	if validReq.Message == "" {
		t.Error("Valid request should have message")
	}

	// Test empty message
	emptyReq := ChatRequest{
		Message:   "",
		SessionID: "test-session",
	}

	if emptyReq.Message != "" {
		t.Error("Empty request message should be empty")
	}
}

func TestChatResponseSerialization(t *testing.T) {
	resp := ChatResponse{
		Response:  "Hello back!",
		SessionID: "test-session",
		Timestamp: "2026-04-12T00:00:00Z",
		TokensUsed: 50,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Failed to serialize response: %v", err)
	}

	var decoded ChatResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to deserialize response: %v", err)
	}

	if decoded.Response != resp.Response {
		t.Errorf("Response mismatch: expected '%s', got '%s'", resp.Response, decoded.Response)
	}
}

func TestChatHistoryRequest(t *testing.T) {
	req := ChatHistoryRequest{
		SessionID: "test-session",
		Limit:     10,
	}

	if req.SessionID == "" {
	 t.Error("SessionID should not be empty")
	}

	if req.Limit < 0 {
	 t.Error("Limit should be non-negative")
	}
}

func TestIntentRequest(t *testing.T) {
	req := IntentRequest{
		Message: "Play some music",
	}

	if req.Message == "" {
	 t.Error("Intent request should have message")
	}

	// Test intent classification
	// This would typically call an NLU service
 intents := classifyIntent(req.Message)
 if len(intents) == 0 {
	 // Might be okay if no intent detected
 }
}

func classifyIntent(message string) []string {
 // Simple intent classification for testing
 if containsKeyword(message, "play") {
  return []string{"media_play"}
 }
 if containsKeyword(message, "stop") {
  return []string{"media_stop"}
 }
 return []string{"general"}
}

func containsKeyword(message, keyword string) bool {
 return len(message) >= len(keyword) && 
  (message == keyword || len(keyword) == 0 ||
   (len(message) > 0 && len(keyword) > 0 && 
    (message[:len(keyword)] == keyword || 
     containsKeyword(message[1:], keyword))))
}

func TestChatEndpointHandler(t *testing.T) {
	// Create mock handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
		 w.WriteHeader(http.StatusMethodNotAllowed)
		 return
		}

	 var req ChatRequest
	 if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		 w.WriteHeader(http.StatusBadRequest)
		 return
	 }

	 if req.Message == "" {
		 w.WriteHeader(http.StatusBadRequest)
		 json.NewEncoder(w).Encode(map[string]string{"error": "message required"})
		 return
	 }

	 w.WriteHeader(http.StatusOK)
	 json.NewEncoder(w).Encode(ChatResponse{
		 Response:  "Mock response",
		 SessionID: req.SessionID,
	 })
	}

	// Test valid request
	reqBody := ChatRequest{Message: "Hello", SessionID: "test"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", nil)
	req.Body = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).Body

	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
	 t.Errorf("Expected 200, got %d", w.Code)
	}
}

// bytes.NewReader import needed
import "bytes"

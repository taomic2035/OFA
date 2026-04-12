// e2e_test.go
// Center End-to-End Integration Tests (v8.2.0)

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const (
	testServerURL    = "http://localhost:8080"
	testWebSocketURL = "ws://localhost:8080/ws"
	testTimeout      = 30 * time.Second
)

// TestClient represents a test client
type TestClient struct {
	httpClient    *http.Client
	websocketConn *websocket.Conn
	sessionID     string
	agentID       string
}

// NewTestClient creates a new test client
func NewTestClient() *TestClient {
	return &TestClient{
		httpClient: &http.Client{
			Timeout: testTimeout,
		},
		agentID: fmt.Sprintf("test_agent_%d", time.Now().UnixNano()),
	}
}

// TestCenterStatus tests the status endpoint
func TestCenterStatus(t *testing.T) {
	client := NewTestClient()

	resp, err := client.httpClient.Get(testServerURL + "/api/v1/status")
	if err != nil {
		t.Skipf("Center server not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status endpoint returned %d, expected 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("Status response: %s", string(body))
}

// TestIdentityCreate tests identity creation
func TestIdentityCreate(t *testing.T) {
	client := NewTestClient()

	identityReq := map[string]interface{}{
		"name": "Test User",
		"personality": map[string]interface{}{
			"openness":      0.6,
			"conscientiousness": 0.7,
			"extraversion":  0.5,
			"agreeableness": 0.8,
			"neuroticism":   0.3,
		},
	}

	body, _ := json.Marshal(identityReq)
	resp, err := client.httpClient.Post(
		testServerURL+"/api/v1/identity",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		t.Skipf("Center server not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Errorf("Identity creation returned %d, expected 200/201", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Identity response: %s", string(respBody))

	// Parse identity ID
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if id, ok := result["id"].(string); ok {
			t.Logf("Created identity ID: %s", id)
		}
	}
}

// TestWebSocketConnection tests WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	client := NewTestClient()

	// Parse WebSocket URL
	u, err := url.Parse(testWebSocketURL)
	if err != nil {
		t.Fatalf("Failed to parse WebSocket URL: %v", err)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	client.websocketConn = conn

	// Send register message
	registerMsg := map[string]interface{}{
		"type": "Register",
		"payload": map[string]interface{}{
			"agent_id":    client.agentID,
			"device_type": "test_device",
			"device_name": "Test Device",
		},
		"timestamp": time.Now().Unix(),
	}

	msgBody, _ := json.Marshal(registerMsg)
	if err := conn.WriteMessage(websocket.TextMessage, msgBody); err != nil {
		t.Errorf("Failed to send register message: %v", err)
		return
	}

	// Read response
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read WebSocket response: %v", err)
		return
	}

	t.Logf("WebSocket response: %s", string(message))

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err == nil {
		if msgType, ok := response["type"].(string); ok {
			if msgType == "RegisterAck" || strings.Contains(msgType, "Register") {
				t.Logf("Registration successful")
			}
		}
	}
}

// TestHeartbeat tests heartbeat message
func TestHeartbeat(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	client.websocketConn = conn

	// Register first
	registerMsg := map[string]interface{}{
		"type": "Register",
		"payload": map[string]interface{}{
			"agent_id": client.agentID,
		},
	}
	msgBody, _ := json.Marshal(registerMsg)
	conn.WriteMessage(websocket.TextMessage, msgBody)

	// Read register ack
	conn.ReadMessage()

	// Send heartbeat
	heartbeatMsg := map[string]interface{}{
		"type": "Heartbeat",
		"payload": map[string]interface{}{
			"agent_id": client.agentID,
			"status":   "online",
			"timestamp": time.Now().Unix(),
		},
	}
	msgBody, _ = json.Marshal(heartbeatMsg)
	if err := conn.WriteMessage(websocket.TextMessage, msgBody); err != nil {
		t.Errorf("Failed to send heartbeat: %v", err)
		return
	}

	t.Logf("Heartbeat sent successfully")
}

// TestChatEndpoint tests chat API
func TestChatEndpoint(t *testing.T) {
	client := NewTestClient()

	chatReq := map[string]interface{}{
		"message":    "Hello, this is a test message",
		"session_id": fmt.Sprintf("test_session_%d", time.Now().UnixNano()),
	}

	body, _ := json.Marshal(chatReq)
	resp, err := client.httpClient.Post(
		testServerURL+"/api/v1/chat",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		t.Skipf("Center server not available: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Chat response: %s", string(respBody))
}

// TestDeviceRegistration tests device registration API
func TestDeviceRegistration(t *testing.T) {
	client := NewTestClient()

	deviceReq := map[string]interface{}{
		"agent_id":    client.agentID,
		"device_type": "android",
		"device_name": "Test Android Device",
		"capabilities": []string{"voice", "display", "camera"},
	}

	body, _ := json.Marshal(deviceReq)
	resp, err := client.httpClient.Post(
		testServerURL+"/api/v1/devices",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		t.Skipf("Center server not available: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Device registration response: %s", string(respBody))
}

// TestSceneEndpoints tests scene API
func TestSceneEndpoints(t *testing.T) {
	client := NewTestClient()

	resp, err := client.httpClient.Get(testServerURL + "/api/v1/scene")
	if err != nil {
		t.Skipf("Center server not available: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Scene response: %s", string(respBody))
}

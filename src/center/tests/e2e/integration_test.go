// integration_test.go
// Center-SDK WebSocket Integration Tests (v9.5.0)
//
// Tests WebSocket communication between SDK and Center:
// - Connection lifecycle (register, heartbeat, disconnect)
// - Message protocol validation
// - Identity synchronization
// - Scene detection and broadcast
// - Multi-device coordination

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket Integration Test Suite

// TestWebSocketLifecycle tests full connection lifecycle
func TestWebSocketLifecycle(t *testing.T) {
	client := NewTestClient()

	u, err := url.Parse(testWebSocketURL)
	if err != nil {
		t.Fatalf("Failed to parse WebSocket URL: %v", err)
	}

	// 1. Connect
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	t.Log("Step 1: WebSocket connected")

	// 2. Register
	registerMsg := WebSocketMessage{
		Type:      "Register",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"device_type": "android",
			"device_name": "Integration Test Device",
			"identity_id": "test_identity_001",
			"capabilities": []string{
				"voice_input",
				"display",
				"camera",
				"bluetooth",
				"gps",
				"fitness",
			},
		},
	}

	if err := WriteJSONMessage(conn, registerMsg); err != nil {
		t.Errorf("Failed to send register: %v", err)
		return
	}
	t.Log("Step 2: Register message sent")

	// 3. Read RegisterAck
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read RegisterAck: %v", err)
		return
	}

	var response WebSocketMessage
	if err := json.Unmarshal(msg, &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if response.Type != "RegisterAck" {
		t.Errorf("Expected RegisterAck, got %s", response.Type)
		return
	}
	t.Log("Step 3: RegisterAck received")

	// Verify session ID assigned
	if payload, ok := response.Payload.(map[string]interface{}); ok {
		if sessionID, ok := payload["session_id"].(string); ok {
			client.sessionID = sessionID
			t.Logf("Session ID assigned: %s", sessionID)
		}
	}

	// 4. Send Heartbeats (3 rounds)
	for i := 0; i < 3; i++ {
		heartbeatMsg := WebSocketMessage{
			Type:      "Heartbeat",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"agent_id":    client.agentID,
				"session_id":  client.sessionID,
				"status":      "online",
				"battery":     85 - i*5,
				"network_type": "wifi",
				"location": map[string]interface{}{
					"latitude":  39.9042,
					"longitude": 116.4074,
				},
			},
		}

		if err := WriteJSONMessage(conn, heartbeatMsg); err != nil {
			t.Errorf("Heartbeat %d failed: %v", i+1, err)
		} else {
			t.Logf("Step 4.%d: Heartbeat sent", i+1)
		}

		time.Sleep(2 * time.Second)
	}

	// 5. Disconnect
	disconnectMsg := WebSocketMessage{
		Type:      "Disconnect",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":   client.agentID,
			"session_id": client.sessionID,
			"reason":     "test_complete",
		},
	}

	if err := WriteJSONMessage(conn, disconnectMsg); err != nil {
		t.Errorf("Failed to send disconnect: %v", err)
	}
	t.Log("Step 5: Disconnect message sent")

	// Close connection
	conn.Close()
	t.Log("Step 6: Connection closed")
}

// TestIdentitySynchronization tests identity sync over WebSocket
func TestIdentitySynchronization(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	// Register first
	registerAndAck(t, conn, client)

	// Send identity update
	identityUpdate := WebSocketMessage{
		Type:      "StateUpdate",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":   client.agentID,
			"session_id": client.sessionID,
			"update_type": "identity",
			"data": map[string]interface{}{
				"name": "Test User Updated",
				"personality": map[string]interface{}{
					"openness":          0.65,
					"conscientiousness": 0.75,
					"extraversion":      0.55,
					"agreeableness":     0.85,
					"neuroticism":       0.35,
				},
				"interests": []map[string]interface{}{
					{
						"category":     "sports",
						"name":         "running",
						"enthusiasm":   0.8,
						"keywords":     []string{"marathon", "trail"},
					},
					{
						"category":     "technology",
						"name":         "programming",
						"enthusiasm":   0.9,
						"keywords":     []string{"golang", "android"},
					},
				},
				"values": map[string]interface{}{
					"privacy":    0.9,
					"efficiency": 0.8,
					"health":     0.85,
				},
			},
		},
	}

	if err := WriteJSONMessage(conn, identityUpdate); err != nil {
		t.Errorf("Failed to send identity update: %v", err)
		return
	}
	t.Log("Identity update sent")

	// Wait for acknowledgment
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)
	t.Logf("Identity update response: %s", response.Type)

	// Request identity sync
	syncRequest := WebSocketMessage{
		Type:      "SyncRequest",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"session_id":  client.sessionID,
			"sync_type":   "identity",
			"request_full": true,
		},
	}

	if err := WriteJSONMessage(conn, syncRequest); err != nil {
		t.Errorf("Failed to send sync request: %v", err)
		return
	}
	t.Log("Sync request sent")

	// Read sync response
	_, msg, err = conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read sync response: %v", err)
		return
	}

	json.Unmarshal(msg, &response)
	t.Logf("Sync response type: %s", response.Type)

	if response.Type == "SyncResponse" {
		t.Log("Identity synchronization completed successfully")
	}
}

// TestSceneDetection tests scene detection and broadcast
func TestSceneDetection(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	registerAndAck(t, conn, client)

	// Report scene data - Running scenario
	sceneReport := WebSocketMessage{
		Type:      "StateUpdate",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"session_id":  client.sessionID,
			"update_type": "scene_data",
			"data": map[string]interface{}{
				"activity_type": "running",
				"duration":      1800, // 30 minutes
				"heart_rate":    145,
				"steps":         3500,
				"location":      "outdoor",
				"device_type":   "watch",
			},
		},
	}

	if err := WriteJSONMessage(conn, sceneReport); err != nil {
		t.Errorf("Failed to send scene report: %v", err)
		return
	}
	t.Log("Scene report sent: Running")

	// Read scene detection result
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read scene response: %v", err)
		return
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)
	t.Logf("Scene response: type=%s", response.Type)

	// Check if scene broadcast received
	if response.Type == "SceneBroadcast" {
		if payload, ok := response.Payload.(map[string]interface{}); ok {
			sceneType := payload["scene_type"]
			confidence := payload["confidence"]
			t.Logf("Scene detected: %s (confidence: %v)", sceneType, confidence)

			// Verify actions
			if actions, ok := payload["actions"].([]interface{}); ok {
				t.Logf("Scene actions count: %d", len(actions))
				for i, action := range actions {
					if act, ok := action.(map[string]interface{}); ok {
						t.Logf("  Action %d: type=%s, target=%s",
							i+1, act["type"], act["target_agent"])
					}
				}
			}
		}
	}

	// Report scene end
	sceneEnd := WebSocketMessage{
		Type:      "StateUpdate",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"session_id":  client.sessionID,
			"update_type": "scene_end",
			"data": map[string]interface{}{
				"scene_type":  "running",
				"end_reason":  "user_stopped",
				"duration":    1800,
			},
		},
	}

	WriteJSONMessage(conn, sceneEnd)
	t.Log("Scene end reported")
}

// TestMultiDeviceCoordination tests multi-device message routing
func TestMultiDeviceCoordination(t *testing.T) {
	// Create multiple device connections
	devices := make([]*TestDevice, 3)
	deviceTypes := []string{"phone", "watch", "tablet"}

	u, _ := url.Parse(testWebSocketURL)

	for i := 0; i < 3; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Skipf("Device %d connection failed: %v", i, err)
			return
		}

		devices[i] = &TestDevice{
			AgentID:    fmt.Sprintf("device_%d_%d", i, time.Now().UnixNano()),
			DeviceType: deviceTypes[i],
			Conn:       conn,
		}

		// Register each device
		registerMsg := WebSocketMessage{
			Type:      "Register",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"agent_id":    devices[i].AgentID,
				"device_type": devices[i].DeviceType,
				"device_name": fmt.Sprintf("Test %s", deviceTypes[i]),
				"identity_id": "shared_identity_001", // Same identity
				"capabilities": getCapabilities(deviceTypes[i]),
			},
		}

		WriteJSONMessage(conn, registerMsg)
		t.Logf("Device %d (%s) registered", i, deviceTypes[i])

		// Read ack
		conn.ReadMessage()
	}

	t.Log("All devices connected")

	// Simulate running scene on watch
	runningScene := WebSocketMessage{
		Type:      "StateUpdate",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    devices[1].AgentID, // Watch
			"update_type": "scene_data",
			"data": map[string]interface{}{
				"activity_type": "running",
				"heart_rate":    150,
				"duration":      600,
				"device_type":   "watch",
			},
		},
	}

	WriteJSONMessage(devices[1].Conn, runningScene)
	t.Log("Running scene reported from watch")

	// Phone should receive routing instructions
	done := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(3)

	for i, device := range devices {
		go func(idx int, d *TestDevice) {
			defer wg.Done()
			select {
			case <-done:
				return
			default:
				_, msg, err := d.Conn.ReadMessage()
				if err == nil {
					var response WebSocketMessage
					json.Unmarshal(msg, &response)
					if response.Type == "SceneBroadcast" ||
					   response.Type == "RouteCommand" {
						t.Logf("Device %d received: %s", idx, response.Type)
						done <- true
					}
				}
			}
		}(i, device)
	}

	select {
	case <-done:
		t.Log("Multi-device coordination successful")
	case <-time.After(5 * time.Second):
		t.Log("Timeout waiting for coordination")
	}

	wg.Wait()

	// Cleanup
	for _, device := range devices {
		device.Conn.Close()
	}
}

// TestBehaviorObservation tests behavior reporting and personality inference
func TestBehaviorObservation(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	registerAndAck(t, conn, client)

	// Report multiple behavior observations
	behaviors := []map[string]interface{}{
		{
			"type":    "decision",
			"subtype": "purchase",
			"data": map[string]interface{}{
				"item":        "coffee",
				"price":       35.0,
				"is_impulse":  false,
				"planned":     true,
			},
		},
		{
			"type":    "interaction",
			"subtype": "social",
			"data": map[string]interface{}{
				"platform":       "wechat",
				"participant_count": 5,
				"emoji_used":     true,
				"duration":       1200,
			},
		},
		{
			"type":    "activity",
			"subtype": "exercise",
			"data": map[string]interface{}{
				"exercise_type": "running",
				"duration":      1800,
				"intensity":     "high",
				"location":      "outdoor",
			},
		},
	}

	for i, behavior := range behaviors {
		behaviorMsg := WebSocketMessage{
			Type:      "BehaviorReport",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"agent_id":   client.agentID,
				"session_id": client.sessionID,
				"behavior":   behavior,
			},
		}

		WriteJSONMessage(conn, behaviorMsg)
		t.Logf("Behavior %d reported: %s/%s", i+1,
			behavior["type"], behavior["subtype"])

		time.Sleep(500 * time.Millisecond)
	}

	// Request personality inference
	inferenceRequest := WebSocketMessage{
		Type:      "SyncRequest",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":     client.agentID,
			"session_id":   client.sessionID,
			"sync_type":    "personality_inference",
			"request_full": true,
		},
	}

	WriteJSONMessage(conn, inferenceRequest)
	t.Log("Personality inference requested")

	// Read inference result
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read inference result: %v", err)
		return
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)
	t.Logf("Inference response: %s", response.Type)

	if response.Type == "SyncResponse" {
		if payload, ok := response.Payload.(map[string]interface{}); ok {
			if personality, ok := payload["personality"].(map[string]interface{}); ok {
				t.Log("Personality inference result:")
				for trait, value := range personality {
					t.Logf("  %s: %.2f", trait, value)
				}
			}
		}
	}
}

// TestTaskAssignment tests task assignment from Center to Agent
func TestTaskAssignment(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	registerAndAck(t, conn, client)

	// Report capabilities
	capabilityUpdate := WebSocketMessage{
		Type:      "StateUpdate",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"session_id":  client.sessionID,
			"update_type": "capabilities",
			"data": map[string]interface{}{
				"skills": []map[string]interface{}{
					{
						"id":          "search",
						"name":        "Search Skill",
						"version":     "1.0",
						"local_only":  false,
					},
					{
						"id":          "order_food",
						"name":        "Order Food",
						"version":     "1.0",
						"local_only":  true,
					},
				},
				"automation": []string{
					"click_text",
					"fill_form",
					"scroll",
				},
			},
		},
	}

	WriteJSONMessage(conn, capabilityUpdate)
	t.Log("Capabilities reported")

	// Request task from Center (simulate Center sending task)
	taskRequest := WebSocketMessage{
		Type:      "TaskRequest",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"session_id":  client.sessionID,
			"task_type":   "skill",
			"task_data": map[string]interface{}{
				"skill_id":  "search",
				"inputs": map[string]string{
					"query":    "coffee shops nearby",
					"platform": "all",
				},
			},
			"priority":   5,
			"timeout":    30000,
		},
	}

	WriteJSONMessage(conn, taskRequest)
	t.Log("Task request sent")

	// Read task assignment or ack
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read task response: %v", err)
		return
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)
	t.Logf("Task response: %s", response.Type)

	// Simulate task execution result
	if response.Type == "TaskAssign" || response.Type == "TaskAck" {
		taskResult := WebSocketMessage{
			Type:      "TaskResult",
			Timestamp: time.Now().Unix(),
			Payload: map[string]interface{}{
				"agent_id":    client.agentID,
				"session_id":  client.sessionID,
				"task_id":     "test_task_001",
				"status":      "completed",
				"output": map[string]interface{}{
					"results": []map[string]interface{}{
						{
							"name":     "Starbucks",
							"distance": "0.5km",
							"rating":   4.5,
						},
						{
							"name":     "Local Cafe",
							"distance": "0.3km",
							"rating":   4.8,
						},
					},
				},
				"execution_time": 2500,
			},
		}

		WriteJSONMessage(conn, taskResult)
		t.Log("Task result sent")
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.Close()

	registerAndAck(t, conn, client)

	// Test 1: Invalid message type
	invalidMsg := WebSocketMessage{
		Type:      "InvalidType",
		Timestamp: time.Now().Unix(),
		Payload:   map[string]interface{}{},
	}

	WriteJSONMessage(conn, invalidMsg)
	t.Log("Invalid message sent")

	_, msg, err := conn.ReadMessage()
	if err == nil {
		var response WebSocketMessage
		json.Unmarshal(msg, &response)
		t.Logf("Response to invalid: %s", response.Type)
	}

	// Test 2: Missing required fields
	missingFieldsMsg := WebSocketMessage{
		Type:      "Heartbeat",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			// Missing agent_id
			"status": "online",
		},
	}

	WriteJSONMessage(conn, missingFieldsMsg)
	t.Log("Message with missing fields sent")

	_, msg, err = conn.ReadMessage()
	if err == nil {
		var response WebSocketMessage
		json.Unmarshal(msg, &response)
		if response.Type == "Error" {
			t.Log("Error response received for missing fields")
		}
	}

	// Test 3: Malformed JSON
	conn.WriteMessage(websocket.TextMessage, []byte("{invalid json}"))
	t.Log("Malformed JSON sent")

	_, msg, err = conn.ReadMessage()
	if err == nil {
		t.Logf("Response to malformed: %s", string(msg))
	}
}

// TestConnectionRecovery tests connection recovery scenarios
func TestConnectionRecovery(t *testing.T) {
	client := NewTestClient()

	u, _ := url.Parse(testWebSocketURL)

	// First connection
	conn1, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed: %v", err)
		return
	}

	registerAndAck(t, conn1, client)
	sessionID1 := client.sessionID
	t.Logf("First session: %s", sessionID1)

	// Abrupt disconnect (no disconnect message)
	conn1.Close()
	t.Log("Connection 1 closed abruptly")

	// Wait and reconnect
	time.Sleep(2 * time.Second)

	// Second connection - should be able to reconnect
	conn2, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Errorf("Reconnection failed: %v", err)
		return
	}
	defer conn2.Close()

	// Re-register with same agent_id
	registerMsg := WebSocketMessage{
		Type:      "Register",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"device_type": "android",
			"device_name": "Reconnection Test",
			"previous_session": sessionID1, // Link to previous session
		},
	}

	WriteJSONMessage(conn2, registerMsg)
	t.Log("Re-register sent")

	_, msg, err := conn2.ReadMessage()
	if err != nil {
		t.Errorf("Failed to read re-register response: %v", err)
		return
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)
	t.Logf("Re-register response: %s", response.Type)

	// Check if previous data is restored
	if response.Type == "RegisterAck" {
		if payload, ok := response.Payload.(map[string]interface{}); ok {
			if restored, ok := payload["data_restored"].(bool); ok {
				t.Logf("Previous data restored: %v", restored)
			}
		}
	}

	t.Log("Connection recovery test completed")
}

// Helper types and functions

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type TestDevice struct {
	AgentID    string
	DeviceType string
	Conn       *websocket.Conn
	SessionID  string
}

func WriteJSONMessage(conn *websocket.Conn, msg WebSocketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func registerAndAck(t *testing.T, conn *websocket.Conn, client *TestClient) {
	registerMsg := WebSocketMessage{
		Type:      "Register",
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"agent_id":    client.agentID,
			"device_type": "test_device",
			"device_name": "Test Device",
		},
	}

	WriteJSONMessage(conn, registerMsg)

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read RegisterAck: %v", err)
	}

	var response WebSocketMessage
	json.Unmarshal(msg, &response)

	if payload, ok := response.Payload.(map[string]interface{}); ok {
		if sessionID, ok := payload["session_id"].(string); ok {
			client.sessionID = sessionID
		}
	}
}

func getCapabilities(deviceType string) []string {
	switch deviceType {
	case "phone":
		return []string{"voice_input", "display", "camera", "bluetooth", "gps", "automation"}
	case "watch":
		return []string{"voice_input", "display", "fitness", "heart_rate", "gps"}
	case "tablet":
		return []string{"display", "camera", "bluetooth", "automation"}
	default:
		return []string{"display"}
	}
}
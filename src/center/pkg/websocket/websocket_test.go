package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

// TestMessageEncodingDecoding tests message encoding and decoding
func TestMessageEncodingDecoding(t *testing.T) {
	payload := &RegisterPayload{
		AgentID:   "agent_001",
		AgentName: "Test Agent",
		AgentType: "mobile",
		DeviceInfo: DeviceInfoPayload{
			OS:        "Android",
			OSVersion: "13",
			Model:     "Pixel 7",
		},
	}

	msg := NewMessage(MsgTypeRegister, payload)

	// Encode
	data, err := EncodeMessage(msg)
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}

	// Decode
	decoded, err := DecodeMessage(data)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	// Verify
	if decoded.Type != MsgTypeRegister {
		t.Errorf("Type should be register, got %s", decoded.Type)
	}
	if decoded.MessageID == "" {
		t.Error("MessageID should not be empty")
	}
	if decoded.Timestamp <= 0 {
		t.Error("Timestamp should be positive")
	}

	// Decode payload
	var decodedPayload RegisterPayload
	if err := DecodePayload(decoded, &decodedPayload); err != nil {
		t.Fatalf("Failed to decode payload: %v", err)
	}

	if decodedPayload.AgentID != "agent_001" {
		t.Errorf("AgentID should be agent_001, got %s", decodedPayload.AgentID)
	}
	if decodedPayload.AgentName != "Test Agent" {
		t.Errorf("AgentName should be Test Agent, got %s", decodedPayload.AgentName)
	}
}

// TestAllMessageTypes tests encoding/decoding of all message types
func TestAllMessageTypes(t *testing.T) {
	testCases := []struct {
		msgType  MessageType
		payload  interface{}
	}{
		{MsgTypeRegister, &RegisterPayload{AgentID: "test"}},
		{MsgTypeRegisterAck, &RegisterAckPayload{Success: true}},
		{MsgTypeHeartbeat, &HeartbeatPayload{AgentID: "test", Status: "online"}},
		{MsgTypeStateUpdate, &StateUpdatePayload{IdentityID: "user_001", UpdateType: "emotion"}},
		{MsgTypeTaskAssign, &TaskAssignPayload{TaskID: "task_001", SkillID: "echo"}},
		{MsgTypeTaskResult, &TaskResultPayload{TaskID: "task_001", Status: "success"}},
		{MsgTypeSyncRequest, &SyncRequestPayload{IdentityID: "user_001", DataType: "memories"}},
		{MsgTypeSyncResponse, &SyncResponsePayload{Success: true, IdentityID: "user_001"}},
		{MsgTypeBehaviorReport, &BehaviorReportPayload{IdentityID: "user_001", BehaviorType: "purchase"}},
		{MsgTypeEmotionUpdate, &EmotionUpdatePayload{IdentityID: "user_001", DominantEmotion: "joy"}},
		{MsgTypeConfigUpdate, &ConfigUpdatePayload{AgentID: "agent_001"}},
		{MsgTypeError, &ErrorPayload{Code: 500, Message: "test error"}},
	}

	for _, tc := range testCases {
		msg := NewMessage(tc.msgType, tc.payload)
		data, err := EncodeMessage(msg)
		if err != nil {
			t.Errorf("Failed to encode %s: %v", tc.msgType, err)
			continue
		}

		decoded, err := DecodeMessage(data)
		if err != nil {
			t.Errorf("Failed to decode %s: %v", tc.msgType, err)
			continue
		}

		if decoded.Type != tc.msgType {
			t.Errorf("Type mismatch for %s: got %s", tc.msgType, decoded.Type)
		}
	}
}

// TestConnectionConfig tests default configuration
func TestConnectionConfig(t *testing.T) {
	config := DefaultConnectionConfig()

	if config.HeartbeatInterval != 30*time.Second {
		t.Errorf("HeartbeatInterval should be 30s, got %v", config.HeartbeatInterval)
	}
	if config.HeartbeatTimeout != 90*time.Second {
		t.Errorf("HeartbeatTimeout should be 90s, got %v", config.HeartbeatTimeout)
	}
	if config.MaxConnections != 1000 {
		t.Errorf("MaxConnections should be 1000, got %d", config.MaxConnections)
	}
	if config.MaxMessageSize != 64*1024 {
		t.Errorf("MaxMessageSize should be 64KB, got %d", config.MaxMessageSize)
	}
}

// TestConnectionState tests connection state
func TestConnectionState(t *testing.T) {
	state := &ConnectionState{
		AgentID:      "agent_001",
		SessionID:    "sess_001",
		IdentityID:   "user_001",
		Status:       "connected",
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	if state.AgentID != "agent_001" {
		t.Errorf("AgentID should be agent_001, got %s", state.AgentID)
	}
	if state.Status != "connected" {
		t.Errorf("Status should be connected, got %s", state.Status)
	}
}

// TestMockWebSocketConn tests mock WebSocket connection
func TestMockWebSocketConn(t *testing.T) {
	conn := NewMockWebSocketConn()

	// Test send
	msg := NewMessage(MsgTypeHeartbeat, &HeartbeatPayload{AgentID: "test"})
	data, _ := EncodeMessage(msg)
	if err := conn.Send(data); err != nil {
		t.Errorf("Failed to send: %v", err)
	}

	// Test receive
	received, err := conn.Receive()
	if err != nil {
		t.Errorf("Failed to receive: %v", err)
	}
	if len(received) != len(data) {
		t.Errorf("Received data length mismatch")
	}

	// Test close
	if err := conn.Close(); err != nil {
		t.Errorf("Failed to close: %v", err)
	}
	if !conn.IsClosed() {
		t.Error("Connection should be closed")
	}
}

// TestConnectionManager tests connection manager basic operations
func TestConnectionManagerBasic(t *testing.T) {
	config := DefaultConnectionConfig()
	manager := NewConnectionManager(config)

	// Test initial state
	if manager.ConnectionCount() != 0 {
		t.Error("Initial connection count should be 0")
	}

	// Test max connections config
	if config.MaxConnections <= 0 {
		t.Error("MaxConnections should be positive")
	}

	// Test heartbeat timeout config
	if config.HeartbeatTimeout <= config.HeartbeatInterval {
		t.Error("HeartbeatTimeout should be greater than HeartbeatInterval")
	}
}

// TestConnectionManagerRegister tests agent registration
func TestConnectionManagerRegister(t *testing.T) {
	config := DefaultConnectionConfig()
	manager := NewConnectionManager(config)

	payload := &RegisterPayload{
		AgentID:   "agent_001",
		AgentName: "Test Agent",
		AgentType: "mobile",
		DeviceInfo: DeviceInfoPayload{
			OS:        "Android",
			OSVersion: "13",
		},
		Capabilities: []CapabilityPayload{
			{ID: "cap_001", Name: "Echo"},
		},
		IdentityID: "user_001",
	}

	conn := NewMockWebSocketConn()
	agentConn, err := manager.RegisterAgent(conn, payload)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Verify registration
	if agentConn.AgentID != "agent_001" {
		t.Errorf("AgentID should be agent_001, got %s", agentConn.AgentID)
	}
	if agentConn.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if agentConn.IdentityID != "user_001" {
		t.Errorf("IdentityID should be user_001, got %s", agentConn.IdentityID)
	}

	// Verify connection count
	if manager.ConnectionCount() != 1 {
		t.Errorf("Connection count should be 1, got %d", manager.ConnectionCount())
	}

	// Verify state
	state := manager.GetState("agent_001")
	if state == nil {
		t.Fatal("State should not be nil")
	}
	if state.Status != "connected" {
		t.Errorf("Status should be connected, got %s", state.Status)
	}

	// Cleanup
	manager.UnregisterAgent("agent_001")
	if manager.ConnectionCount() != 0 {
		t.Error("Connection count should be 0 after unregister")
	}
}

// TestConnectionManagerHeartbeat tests heartbeat handling
func TestConnectionManagerHeartbeat(t *testing.T) {
	config := DefaultConnectionConfig()
	manager := NewConnectionManager(config)

	// Register agent first
	payload := &RegisterPayload{AgentID: "agent_001"}
	conn := NewMockWebSocketConn()
	manager.RegisterAgent(conn, payload)

	// Send heartbeat
	heartbeat := &HeartbeatPayload{
		AgentID:     "agent_001",
		Status:      "active",
		Resources:   ResourcePayload{CPUUsage: 0.5},
		PendingTasks: 5,
	}

	err := manager.HandleHeartbeat("agent_001", heartbeat)
	if err != nil {
		t.Fatalf("Failed to handle heartbeat: %v", err)
	}

	// Verify agent is online
	if !manager.IsAgentOnline("agent_001") {
		t.Error("Agent should be online after heartbeat")
	}

	// Cleanup
	manager.UnregisterAgent("agent_001")
}

// TestConnectionManagerSend tests message sending
func TestConnectionManagerSend(t *testing.T) {
	config := DefaultConnectionConfig()
	manager := NewConnectionManager(config)

	// Register agent
	payload := &RegisterPayload{AgentID: "agent_001"}
	conn := NewMockWebSocketConn()
	manager.RegisterAgent(conn, payload)

	// Send message
	msg := NewMessage(MsgTypeStateUpdate, &StateUpdatePayload{
		IdentityID: "user_001",
		UpdateType: "emotion",
	})

	err := manager.SendMessage("agent_001", msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Cleanup
	manager.UnregisterAgent("agent_001")
}

// TestConnectionManagerNotFound tests error cases
func TestConnectionManagerNotFound(t *testing.T) {
	config := DefaultConnectionConfig()
	manager := NewConnectionManager(config)

	// Test get non-existent connection
	conn, err := manager.GetConnection("nonexistent")
	if conn != nil {
		t.Error("Should return nil for non-existent agent")
	}
	if err != ErrAgentNotConnected {
		t.Errorf("Should return ErrAgentNotConnected, got %v", err)
	}

	// Test heartbeat for non-existent agent
	err = manager.HandleHeartbeat("nonexistent", &HeartbeatPayload{})
	if err != ErrAgentNotConnected {
		t.Errorf("Should return ErrAgentNotConnected, got %v", err)
	}

	// Test send to non-existent agent
	err = manager.SendMessage("nonexistent", NewMessage(MsgTypeHeartbeat, nil))
	if err != ErrAgentNotConnected {
		t.Errorf("Should return ErrAgentNotConnected, got %v", err)
	}
}

// TestBroadcasterSubscription tests broadcaster subscription
func TestBroadcasterSubscription(t *testing.T) {
	config := DefaultConnectionConfig()
	connManager := NewConnectionManager(config)
	broadcaster := NewStateBroadcaster(connManager)

	// Subscribe
	broadcaster.Subscribe("agent_001", "user_001")
	subscribers := broadcaster.GetSubscribers("user_001")
	if len(subscribers) != 1 || subscribers[0] != "agent_001" {
		t.Errorf("Should have 1 subscriber, got %d", len(subscribers))
	}

	// Subscribe another agent
	broadcaster.Subscribe("agent_002", "user_001")
	subscribers = broadcaster.GetSubscribers("user_001")
	if len(subscribers) != 2 {
		t.Errorf("Should have 2 subscribers, got %d", len(subscribers))
	}

	// Unsubscribe
	broadcaster.Unsubscribe("agent_001", "user_001")
	subscribers = broadcaster.GetSubscribers("user_001")
	if len(subscribers) != 1 || subscribers[0] != "agent_002" {
		t.Errorf("Should have 1 subscriber after unsubscribe")
	}

	// Unsubscribe all
	broadcaster.UnsubscribeAll("agent_002")
	subscribers = broadcaster.GetSubscribers("user_001")
	if len(subscribers) != 0 {
		t.Errorf("Should have 0 subscribers after unsubscribe all")
	}
}

// TestBroadcasterVersion tests version management
func TestBroadcasterVersion(t *testing.T) {
	config := DefaultConnectionConfig()
	connManager := NewConnectionManager(config)
	broadcaster := NewStateBroadcaster(connManager)

	// Initial version should be 0
	if broadcaster.GetLastVersion("user_001") != 0 {
		t.Error("Initial version should be 0")
	}

	// Update version
	broadcaster.UpdateVersion("user_001", 100)
	if broadcaster.GetLastVersion("user_001") != 100 {
		t.Errorf("Version should be 100, got %d", broadcaster.GetLastVersion("user_001"))
	}
}

// === Mock WebSocket Connection ===

type MockWebSocketConn struct {
	sendQueue    [][]byte
	receiveQueue [][]byte
	closed       bool
}

func NewMockWebSocketConn() *MockWebSocketConn {
	return &MockWebSocketConn{
		sendQueue:    make([][]byte, 0),
		receiveQueue: make([][]byte, 0),
		closed:       false,
	}
}

func (c *MockWebSocketConn) Send(data []byte) error {
	if c.closed {
		return ErrAgentNotConnected
	}
	c.sendQueue = append(c.sendQueue, data)
	c.receiveQueue = append(c.receiveQueue, data) // Echo back for testing
	return nil
}

func (c *MockWebSocketConn) Receive() ([]byte, error) {
	if c.closed {
		return nil, ErrAgentNotConnected
	}
	if len(c.receiveQueue) == 0 {
		return nil, ErrAgentNotConnected
	}
	data := c.receiveQueue[0]
	c.receiveQueue = c.receiveQueue[1:]
	return data, nil
}

func (c *MockWebSocketConn) Close() error {
	c.closed = true
	return nil
}

func (c *MockWebSocketConn) IsClosed() bool {
	return c.closed
}

// TestPayloadJSON tests JSON serialization of payloads
func TestPayloadJSON(t *testing.T) {
	// Test RegisterPayload JSON
	payload := RegisterPayload{
		AgentID:   "agent_001",
		AgentName: "Test",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded RegisterPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if decoded.AgentID != "agent_001" {
		t.Errorf("AgentID mismatch")
	}

	// Test HeartbeatPayload JSON
	heartbeat := HeartbeatPayload{
		AgentID:     "agent_001",
		Status:      "online",
		Resources:   ResourcePayload{CPUUsage: 0.75, BatteryLevel: 80},
	}
	data, err = json.Marshal(heartbeat)
	if err != nil {
		t.Fatalf("Failed to marshal heartbeat: %v", err)
	}

	var decodedHeartbeat HeartbeatPayload
	if err := json.Unmarshal(data, &decodedHeartbeat); err != nil {
		t.Fatalf("Failed to unmarshal heartbeat: %v", err)
	}
	if decodedHeartbeat.Resources.BatteryLevel != 80 {
		t.Errorf("BatteryLevel should be 80, got %d", decodedHeartbeat.Resources.BatteryLevel)
	}
}
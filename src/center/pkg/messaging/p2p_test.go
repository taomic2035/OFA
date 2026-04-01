package messaging

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func defaultTestConfig() P2PConfig {
	return P2PConfig{
		LocalAddress:   "127.0.0.1",
		Port:           0, // 随机端口
		MaxConnections: 10,
		ConnectTimeout: 5 * time.Second,
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		PingInterval:   1 * time.Second,
		PongTimeout:    3 * time.Second,
		MaxMessageSize: 1024 * 1024,
		BufferSize:     100,
	}
}

func TestNewP2PManager(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")

	if pm == nil {
		t.Fatal("NewP2PManager() returned nil")
	}
	if pm.localAgentID != "agent-1" {
		t.Errorf("localAgentID = %v, want agent-1", pm.localAgentID)
	}
	if len(pm.peers) != 0 {
		t.Errorf("peers should be empty initially")
	}
}

func TestP2PManager_StartStop(t *testing.T) {
	// 不启动监听器，仅测试生命周期
	config := defaultTestConfig()
	config.LocalAddress = "" // 不启动监听
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()

	// 启动
	err := pm.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// 检查运行状态
	if !pm.running {
		t.Error("running should be true after Start()")
	}

	// 停止
	err = pm.Stop()
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// 检查停止状态
	if pm.running {
		t.Error("running should be false after Stop()")
	}
}

func TestP2PManager_StartWithoutListener(t *testing.T) {
	// 不监听本地地址，仅作为客户端
	config := defaultTestConfig()
	config.LocalAddress = ""
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()

	err := pm.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer pm.Stop()

	// 应该可以正常启动
	if !pm.running {
		t.Error("running should be true")
	}
}

func TestP2PManager_RegisterHandler(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")

	handler := func(ctx context.Context, msg *Message) (*Message, error) {
		return &Message{ID: "response"}, nil
	}

	pm.RegisterHandler("test.action", handler)

	if len(pm.handlers) != 1 {
		t.Errorf("handlers count = %d, want 1", len(pm.handlers))
	}
}

func TestP2PManager_GroupOperations(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")

	// 创建组
	pm.CreateGroup("group-1", []string{"agent-2", "agent-3"})

	if len(pm.groups) != 1 {
		t.Errorf("groups count = %d, want 1", len(pm.groups))
	}

	// 加入组
	pm.JoinGroup("group-1", "agent-4")

	members := pm.groups["group-1"]
	if len(members) != 3 {
		t.Errorf("group members = %d, want 3", len(members))
	}

	// 离开组
	pm.LeaveGroup("group-1", "agent-3")

	members = pm.groups["group-1"]
	if len(members) != 2 {
		t.Errorf("group members after leave = %d, want 2", len(members))
	}
}

func TestP2PManager_GetPeers(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")

	// 初始应该为空
	peers := pm.GetPeers()
	if len(peers) != 0 {
		t.Errorf("GetPeers() should return empty initially")
	}

	// 获取已连接的对等节点
	connected := pm.GetConnectedPeers()
	if len(connected) != 0 {
		t.Errorf("GetConnectedPeers() should return empty initially")
	}
}

func TestP2PManager_IsConnected(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")

	if pm.IsConnected("agent-2") {
		t.Error("IsConnected() should return false for unknown peer")
	}
}

func TestP2PManager_GetStatistics(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()
	pm.Start(ctx)
	defer pm.Stop()

	stats := pm.GetStatistics()

	if stats["running"] != true {
		t.Error("running should be true")
	}
	if stats["connected_peers"] != 0 {
		t.Errorf("connected_peers should be 0")
	}
}

func TestMessageType_Constants(t *testing.T) {
	tests := []struct {
		value    MessageType
		expected string
	}{
		{TypeDirect, "direct"},
		{TypeBroadcast, "broadcast"},
		{TypeMulticast, "multicast"},
		{TypeRequest, "request"},
		{TypeResponse, "response"},
		{TypeAck, "ack"},
		{TypePing, "ping"},
		{TypePong, "pong"},
	}

	for _, tt := range tests {
		if string(tt.value) != tt.expected {
			t.Errorf("MessageType %v = %s, want %s", tt.value, string(tt.value), tt.expected)
		}
	}
}

func TestMessagePriority_Constants(t *testing.T) {
	if PriorityLow != 0 {
		t.Errorf("PriorityLow = %d, want 0", PriorityLow)
	}
	if PriorityNormal != 1 {
		t.Errorf("PriorityNormal = %d, want 1", PriorityNormal)
	}
	if PriorityHigh != 2 {
		t.Errorf("PriorityHigh = %d, want 2", PriorityHigh)
	}
	if PriorityCritical != 3 {
		t.Errorf("PriorityCritical = %d, want 3", PriorityCritical)
	}
}

func TestMessageState_Constants(t *testing.T) {
	tests := []struct {
		value    MessageState
		expected string
	}{
		{StatePending, "pending"},
		{StateSent, "sent"},
		{StateDelivered, "delivered"},
		{StateAcked, "acked"},
		{StateFailed, "failed"},
		{StateExpired, "expired"},
	}

	for _, tt := range tests {
		if string(tt.value) != tt.expected {
			t.Errorf("MessageState %v = %s, want %s", tt.value, string(tt.value), tt.expected)
		}
	}
}

func TestConnState_Constants(t *testing.T) {
	tests := []struct {
		value    ConnState
		expected string
	}{
		{ConnStateConnecting, "connecting"},
		{ConnStateConnected, "connected"},
		{ConnStateReady, "ready"},
		{ConnStateClosing, "closing"},
		{ConnStateClosed, "closed"},
		{ConnStateError, "error"},
	}

	for _, tt := range tests {
		if string(tt.value) != tt.expected {
			t.Errorf("ConnState %v = %s, want %s", tt.value, string(tt.value), tt.expected)
		}
	}
}

func TestP2PManager_SendToUnknownPeer(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()
	pm.Start(ctx)
	defer pm.Stop()

	msg := &Message{
		ID:     "msg-1",
		Type:   TypeDirect,
		Action: "test.action",
	}

	err := pm.Send("unknown-agent", msg)
	if err == nil {
		t.Error("Send() to unknown peer should fail")
	}
}

func TestP2PManager_BroadcastNoPeers(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()
	pm.Start(ctx)
	defer pm.Stop()

	msg := &Message{
		ID:     "msg-1",
		Action: "test.action",
	}

	// 广播给0个对等节点应该成功（没有错误）
	err := pm.Broadcast(msg)
	if err != nil {
		t.Errorf("Broadcast() with no peers error = %v", err)
	}
}

func TestP2PManager_MulticastNonExistentGroup(t *testing.T) {
	config := defaultTestConfig()
	pm := NewP2PManager(config, "agent-1")
	ctx := context.Background()
	pm.Start(ctx)
	defer pm.Stop()

	msg := &Message{
		ID:     "msg-1",
		Action: "test.action",
	}

	err := pm.Multicast("non-existent-group", msg)
	if err == nil {
		t.Error("Multicast() to non-existent group should fail")
	}
}

func TestPeerConnection(t *testing.T) {
	now := time.Now()
	peer := &PeerConnection{
		AgentID:      "agent-2",
		Address:      "127.0.0.1:8080",
		State:        ConnStateConnected,
		LastActive:   now,
		MessageCount: 10,
		BytesSent:    1000,
		BytesRecv:    2000,
	}

	if peer.AgentID != "agent-2" {
		t.Errorf("AgentID = %s, want agent-2", peer.AgentID)
	}
	if peer.State != ConnStateConnected {
		t.Errorf("State = %v, want connected", peer.State)
	}
}

func TestPeerInfo(t *testing.T) {
	info := &PeerInfo{
		AgentID:      "agent-2",
		Address:      "127.0.0.1:8080",
		Capabilities: []string{"skill1", "skill2"},
		Metadata:     map[string]string{"version": "1.0"},
		Online:       true,
	}

	if info.AgentID != "agent-2" {
		t.Errorf("AgentID = %s, want agent-2", info.AgentID)
	}
	if len(info.Capabilities) != 2 {
		t.Errorf("Capabilities count = %d, want 2", len(info.Capabilities))
	}
}

func TestMessage(t *testing.T) {
	now := time.Now()
	msg := &Message{
		ID:           "msg-1",
		Type:         TypeDirect,
		Priority:     PriorityHigh,
		From:         "agent-1",
		To:           "agent-2",
		Action:       "test.action",
		Payload:      map[string]interface{}{"key": "value"},
		Timestamp:    now,
		TTL:          30 * time.Second,
		RequireAck:   true,
		CorrelationID: "corr-1",
	}

	if msg.ID != "msg-1" {
		t.Errorf("ID = %s, want msg-1", msg.ID)
	}
	if msg.Type != TypeDirect {
		t.Errorf("Type = %v, want direct", msg.Type)
	}
	if msg.Priority != PriorityHigh {
		t.Errorf("Priority = %d, want 2", msg.Priority)
	}
}

func TestMessageAck(t *testing.T) {
	ack := &MessageAck{
		MessageID: "msg-1",
		From:      "agent-2",
		To:        "agent-1",
		Status:    StateDelivered,
		Timestamp: time.Now(),
		Latency:   10 * time.Millisecond,
	}

	if ack.MessageID != "msg-1" {
		t.Errorf("MessageID = %s, want msg-1", ack.MessageID)
	}
	if ack.Status != StateDelivered {
		t.Errorf("Status = %v, want delivered", ack.Status)
	}
}

// TestP2PConnection 是一个集成测试，测试两个P2P管理器之间的连接
func TestP2PConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建第一个管理器（监听）
	config1 := P2PConfig{
		LocalAddress:   "127.0.0.1",
		Port:           0, // 随机端口
		MaxConnections: 10,
		ConnectTimeout: 5 * time.Second,
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		PingInterval:   1 * time.Second,
		PongTimeout:    3 * time.Second,
		MaxMessageSize: 1024 * 1024,
		BufferSize:     100,
	}

	pm1 := NewP2PManager(config1, "agent-1")
	ctx := context.Background()
	err := pm1.Start(ctx)
	if err != nil {
		t.Fatalf("pm1 Start() error = %v", err)
	}
	defer pm1.Stop()

	// 获取实际监听地址
	addr := pm1.listener.Addr().(*net.TCPAddr)
	actualAddr := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	// 创建第二个管理器（客户端）
	config2 := P2PConfig{
		MaxConnections: 10,
		ConnectTimeout: 5 * time.Second,
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		PingInterval:   1 * time.Second,
		PongTimeout:    3 * time.Second,
		MaxMessageSize: 1024 * 1024,
		BufferSize:     100,
	}

	pm2 := NewP2PManager(config2, "agent-2")
	err = pm2.Start(ctx)
	if err != nil {
		t.Fatalf("pm2 Start() error = %v", err)
	}
	defer pm2.Stop()

	// 连接到pm1
	peerInfo := &PeerInfo{
		AgentID: "agent-1",
		Address: actualAddr,
	}

	err = pm2.Connect(ctx, peerInfo)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证连接
	if !pm2.IsConnected("agent-1") {
		t.Error("pm2 should be connected to agent-1")
	}
}
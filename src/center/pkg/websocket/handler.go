package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrRegistrationFailed = errors.New("registration failed")
)

// WebSocketHandler handles WebSocket connections from agents
type WebSocketHandler struct {
	mu sync.RWMutex

	// Connection manager
	connManager *ConnectionManager

	// State broadcaster
	broadcaster *StateBroadcaster

	// WebSocket upgrader
	upgrader websocket.Upgrader

	// Message handlers (MessageType -> handler function)
	messageHandlers map[MessageType]MessageHandlerFunc

	// Connection config
	config ConnectionConfig
}

// MessageHandlerFunc is a function that handles a specific message type
type MessageHandlerFunc func(conn *AgentConnection, msg *WebSocketMessage) error

// WebSocketHandlerConfig holds handler configuration
type WebSocketHandlerConfig struct {
	// Connection config
	ConnectionConfig ConnectionConfig

	// CORS settings
	AllowedOrigins []string

	// Authentication
	RequireAuth bool
}

// DefaultWebSocketHandlerConfig returns default configuration
func DefaultWebSocketHandlerConfig() WebSocketHandlerConfig {
	return WebSocketHandlerConfig{
		ConnectionConfig: DefaultConnectionConfig(),
		AllowedOrigins:   []string{"*"},
		RequireAuth:      false,
	}
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(config WebSocketHandlerConfig) *WebSocketHandler {
	connManager := NewConnectionManager(config.ConnectionConfig)
	broadcaster := NewStateBroadcaster(connManager)

	handler := &WebSocketHandler{
		connManager:  connManager,
		broadcaster:  broadcaster,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ConnectionConfig.ReadBufferSize,
			WriteBufferSize: config.ConnectionConfig.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				if len(config.AllowedOrigins) == 0 || config.AllowedOrigins[0] == "*" {
					return true
				}
				origin := r.Header.Get("Origin")
				for _, allowed := range config.AllowedOrigins {
					if allowed == origin {
						return true
					}
				}
				return false
			},
		},
		messageHandlers: make(map[MessageType]MessageHandlerFunc),
		config:          config.ConnectionConfig,
	}

	// Register default handlers
	handler.RegisterDefaultHandlers()

	return handler
}

// Start starts the handler background workers
func (h *WebSocketHandler) Start() {
	h.connManager.Start()
	h.broadcaster.Start()
}

// Stop stops the handler
func (h *WebSocketHandler) Stop() {
	h.broadcaster.Stop()
	h.connManager.Stop()
}

// === HTTP Handler ===

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create connection wrapper
	conn := NewGorillaWebSocketConn(wsConn)

	// Handle connection
	go h.handleConnection(conn)
}

// handleConnection handles a WebSocket connection lifecycle
func (h *WebSocketHandler) handleConnection(conn WebSocketConn) {
	// Wait for registration message
	msg, err := h.receiveMessage(conn)
	if err != nil {
		log.Printf("Failed to receive registration: %v", err)
		conn.Close()
		return
	}

	// Must be a registration message
	if msg.Type != MsgTypeRegister {
		log.Printf("First message must be register, got: %s", msg.Type)
		h.sendError(conn, ErrRegistrationFailed, "First message must be register")
		conn.Close()
		return
	}

	// Parse registration payload
	var payload RegisterPayload
	if err := DecodePayload(msg, &payload); err != nil {
		log.Printf("Failed to decode registration: %v", err)
		h.sendError(conn, ErrRegistrationFailed, err.Error())
		conn.Close()
		return
	}

	// Register agent
	agentConn, err := h.connManager.RegisterAgent(conn, &payload)
	if err != nil {
		log.Printf("Failed to register agent: %v", err)
		h.sendError(conn, err, err.Error())
		conn.Close()
		return
	}

	// Subscribe to identity updates
	if payload.IdentityID != "" {
		h.broadcaster.Subscribe(payload.AgentID, payload.IdentityID)
	}

	// Send acknowledgment
	ackPayload := &RegisterAckPayload{
		Success:           true,
		AgentID:           payload.AgentID,
		SessionID:         agentConn.SessionID,
		HeartbeatInterval: int(h.config.HeartbeatInterval.Milliseconds()),
		Config: map[string]string{
			"version": "v7.0.0",
		},
	}
	ackMsg := NewMessage(MsgTypeRegisterAck, ackPayload)
	data, _ := EncodeMessage(ackMsg)
	conn.Send(data)

	log.Printf("Agent %s connected via WebSocket (session: %s)", payload.AgentID, agentConn.SessionID)

	// Message loop
	for {
		msg, err := h.receiveMessage(conn)
		if err != nil {
			log.Printf("Connection error for %s: %v", agentConn.AgentID, err)
			break
		}

		// Handle message
		if handler, ok := h.messageHandlers[msg.Type]; ok {
			if err := handler(agentConn, msg); err != nil {
				log.Printf("Handler error for %s (%s): %v", agentConn.AgentID, msg.Type, err)
			}
		} else {
			log.Printf("Unhandled message type: %s", msg.Type)
		}
	}

	// Cleanup
	h.broadcaster.UnsubscribeAll(payload.AgentID)
	h.connManager.UnregisterAgent(payload.AgentID)
}

// receiveMessage receives a WebSocket message
func (h *WebSocketHandler) receiveMessage(conn WebSocketConn) (*WebSocketMessage, error) {
	data, err := conn.Receive()
	if err != nil {
		return nil, err
	}

	return DecodeMessage(data)
}

// sendError sends an error message
func (h *WebSocketHandler) sendError(conn WebSocketConn, err error, details string) {
	payload := &ErrorPayload{
		Code:    500,
		Message: err.Error(),
		Details: details,
	}
	msg := NewMessage(MsgTypeError, payload)
	data, _ := EncodeMessage(msg)
	conn.Send(data)
}

// === Message Handlers ===

// RegisterHandler registers a handler for a message type
func (h *WebSocketHandler) RegisterHandler(msgType MessageType, handler MessageHandlerFunc) {
	h.messageHandlers[msgType] = handler
}

// RegisterDefaultHandlers registers default message handlers
func (h *WebSocketHandler) RegisterDefaultHandlers() {
	h.RegisterHandler(MsgTypeHeartbeat, h.handleHeartbeat)
	h.RegisterHandler(MsgTypeTaskResult, h.handleTaskResult)
	h.RegisterHandler(MsgTypeSyncRequest, h.handleSyncRequest)
	h.RegisterHandler(MsgTypeBehaviorReport, h.handleBehaviorReport)
	h.RegisterHandler(MsgTypeDisconnect, h.handleDisconnect)
}

// handleHeartbeat handles heartbeat messages
func (h *WebSocketHandler) handleHeartbeat(conn *AgentConnection, msg *WebSocketMessage) error {
	var payload HeartbeatPayload
	if err := DecodePayload(msg, &payload); err != nil {
		return err
	}

	return h.connManager.HandleHeartbeat(conn.AgentID, &payload)
}

// handleTaskResult handles task result messages
func (h *WebSocketHandler) handleTaskResult(conn *AgentConnection, msg *WebSocketMessage) error {
	var payload TaskResultPayload
	if err := DecodePayload(msg, &payload); err != nil {
		return err
	}

	log.Printf("Task %s result from %s: status=%s, duration=%dms",
		payload.TaskID, conn.AgentID, payload.Status, payload.DurationMS)

	// TODO: Forward to task management system
	return nil
}

// handleSyncRequest handles sync request messages
func (h *WebSocketHandler) handleSyncRequest(conn *AgentConnection, msg *WebSocketMessage) error {
	var payload SyncRequestPayload
	if err := DecodePayload(msg, &payload); err != nil {
		return err
	}

	log.Printf("Sync request from %s: type=%s, identity=%s, version=%d",
		conn.AgentID, payload.DataType, payload.IdentityID, payload.Version)

	// TODO: Forward to sync service
	return nil
}

// handleBehaviorReport handles behavior report messages
func (h *WebSocketHandler) handleBehaviorReport(conn *AgentConnection, msg *WebSocketMessage) error {
	var payload BehaviorReportPayload
	if err := DecodePayload(msg, &payload); err != nil {
		return err
	}

	log.Printf("Behavior report from %s: type=%s, identity=%s",
		conn.AgentID, payload.BehaviorType, payload.IdentityID)

	// TODO: Forward to behavior analysis system
	return nil
}

// handleDisconnect handles disconnect messages
func (h *WebSocketHandler) handleDisconnect(conn *AgentConnection, msg *WebSocketMessage) error {
	log.Printf("Agent %s requested disconnect", conn.AgentID)

	h.broadcaster.UnsubscribeAll(conn.AgentID)
	h.connManager.UnregisterAgent(conn.AgentID)
	return nil
}

// === Accessor Methods ===

// GetConnectionManager returns the connection manager
func (h *WebSocketHandler) GetConnectionManager() *ConnectionManager {
	return h.connManager
}

// GetBroadcaster returns the state broadcaster
func (h *WebSocketHandler) GetBroadcaster() *StateBroadcaster {
	return h.broadcaster
}

// === Gorilla WebSocket Connection Wrapper ===

// GorillaWebSocketConn wraps gorilla websocket connection
type GorillaWebSocketConn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// NewGorillaWebSocketConn creates a new wrapper
func NewGorillaWebSocketConn(conn *websocket.Conn) *GorillaWebSocketConn {
	return &GorillaWebSocketConn{
		conn: conn,
	}
}

// Send sends data to the WebSocket connection
func (c *GorillaWebSocketConn) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Receive receives data from the WebSocket connection
func (c *GorillaWebSocketConn) Receive() ([]byte, error) {
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// Only handle text messages
	if msgType != websocket.TextMessage {
		return nil, ErrInvalidMessageType
	}

	return data, nil
}

// Close closes the WebSocket connection
func (c *GorillaWebSocketConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

// IsClosed checks if the connection is closed
func (c *GorillaWebSocketConn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn == nil
}

// === REST API Integration ===

// GetConnectionsAPI returns connection info for REST API
func (h *WebSocketHandler) GetConnectionsAPI() []ConnectionStateInfo {
	connections := h.connManager.ListConnections()
	result := make([]ConnectionStateInfo, len(connections))
	for i, conn := range connections {
		result[i] = ConnectionStateInfo{
			AgentID:      conn.AgentID,
			SessionID:    conn.SessionID,
			IdentityID:   conn.IdentityID,
			Status:       conn.Status,
			LastSeen:     conn.LastSeen.Format(time.RFC3339),
			RegisteredAt: conn.RegisteredAt.Format(time.RFC3339),
		}
	}
	return result
}

// ConnectionStateInfo is JSON-serializable connection state
type ConnectionStateInfo struct {
	AgentID      string `json:"agent_id"`
	SessionID    string `json:"session_id"`
	IdentityID   string `json:"identity_id"`
	Status       string `json:"status"`
	LastSeen     string `json:"last_seen"`
	RegisteredAt string `json:"registered_at"`
}

// HandleConnectionsList handles REST API connection list request
func (h *WebSocketHandler) HandleConnectionsList(w http.ResponseWriter, r *http.Request) {
	connections := h.GetConnectionsAPI()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"count":    len(connections),
		"connections": connections,
	})
}

// HandleConnectionDetail handles REST API connection detail request
func (h *WebSocketHandler) HandleConnectionDetail(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	state := h.connManager.GetState(agentID)
	if state == nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	info := ConnectionStateInfo{
		AgentID:      state.AgentID,
		SessionID:    state.SessionID,
		IdentityID:   state.IdentityID,
		Status:       state.Status,
		LastSeen:     state.LastSeen.Format(time.RFC3339),
		RegisteredAt: state.RegisteredAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"connection": info,
	})
}
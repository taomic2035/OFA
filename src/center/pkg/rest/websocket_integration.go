package rest

import (
	"github.com/ofa/center/pkg/websocket"
)

// WebSocketIntegration integrates WebSocket into REST server
type WebSocketIntegration struct {
	handler *websocket.WebSocketHandler
}

// NewWebSocketIntegration creates a new WebSocket integration
func NewWebSocketIntegration(config websocket.WebSocketHandlerConfig) *WebSocketIntegration {
	handler := websocket.NewWebSocketHandler(config)
	handler.Start()

	return &WebSocketIntegration{
		handler: handler,
	}
}

// Stop stops the WebSocket integration
func (w *WebSocketIntegration) Stop() {
	w.handler.Stop()
}

// GetHandler returns the WebSocket handler
func (w *WebSocketIntegration) GetHandler() *websocket.WebSocketHandler {
	return w.handler
}

// SetupRoutes adds WebSocket routes to the REST router
func (w *WebSocketIntegration) SetupRoutes(router interface{}) {
	// This will be called by the Server during setup
	// The actual route setup is done in server.go
}

// GetConnectionManager returns the connection manager
func (w *WebSocketIntegration) GetConnectionManager() *websocket.ConnectionManager {
	return w.handler.GetConnectionManager()
}

// GetBroadcaster returns the state broadcaster
func (w *WebSocketIntegration) GetBroadcaster() *websocket.StateBroadcaster {
	return w.handler.GetBroadcaster()
}
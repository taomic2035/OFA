package stream

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// StreamAPIHandler handles stream API requests
type StreamAPIHandler struct {
	manager *StreamManager
}

// NewStreamAPIHandler creates a new stream API handler
func NewStreamAPIHandler(manager *StreamManager) *StreamAPIHandler {
	return &StreamAPIHandler{
		manager: manager,
	}
}

// RegisterRoutes registers stream API routes
func (h *StreamAPIHandler) RegisterRoutes(r *mux.Router) {
	// Stream management
	r.HandleFunc("/api/v1/streams", h.ListStreamsHandler).Methods("GET")
	r.HandleFunc("/api/v1/streams", h.CreateStreamHandler).Methods("POST")
	r.HandleFunc("/api/v1/streams/{id}", h.GetStreamHandler).Methods("GET")
	r.HandleFunc("/api/v1/streams/{id}", h.CloseStreamHandler).Methods("DELETE")
	r.HandleFunc("/api/v1/streams/{id}/stats", h.GetStreamStatsHandler).Methods("GET")

	// Stream control
	r.HandleFunc("/api/v1/streams/{id}/pause", h.PauseStreamHandler).Methods("POST")
	r.HandleFunc("/api/v1/streams/{id}/resume", h.ResumeStreamHandler).Methods("POST")

	// Stream messages
	r.HandleFunc("/api/v1/streams/{id}/messages", h.GetMessagesHandler).Methods("GET")
	r.HandleFunc("/api/v1/streams/{id}/messages", h.PushMessageHandler).Methods("POST")

	// Streaming endpoint (HTTP streaming)
	r.HandleFunc("/api/v1/streams/{id}/stream", h.manager.StreamHandler).Methods("GET")
}

// ListStreamsHandler handles list streams request
func (h *StreamAPIHandler) ListStreamsHandler(w http.ResponseWriter, r *http.Request) {
	streams := h.manager.ListStreams()

	// Convert to summary format
	var summaries []map[string]interface{}
	for _, s := range streams {
		s.mu.RLock()
		summary := map[string]interface{}{
			"id":         s.ID,
			"type":       s.Type,
			"state":      s.State,
			"created_at": s.CreatedAt,
			"sequence":   s.LastSeq,
			"task_id":    s.taskID,
			"agent_id":   s.agentID,
		}
		s.mu.RUnlock()
		summaries = append(summaries, summary)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"streams": summaries,
		"count":   len(summaries),
	})
}

// CreateStreamHandler handles create stream request
func (h *StreamAPIHandler) CreateStreamHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    StreamType `json:"type"`
		TaskID  string     `json:"task_id,omitempty"`
		AgentID string     `json:"agent_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	stream, err := h.manager.CreateStream(req.Type, req.TaskID, req.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         stream.ID,
		"type":       stream.Type,
		"state":      stream.State,
		"created_at": stream.CreatedAt,
	})
}

// GetStreamHandler handles get stream request
func (h *StreamAPIHandler) GetStreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stream, ok := h.manager.GetStream(streamID)
	if !ok {
		writeError(w, http.StatusNotFound, "Stream not found")
		return
	}

	stream.mu.RLock()
	data := map[string]interface{}{
		"id":         stream.ID,
		"type":       stream.Type,
		"state":      stream.State,
		"created_at": stream.CreatedAt,
		"sequence":   stream.LastSeq,
		"task_id":    stream.taskID,
		"agent_id":   stream.agentID,
	}
	stream.mu.RUnlock()

	writeJSON(w, http.StatusOK, data)
}

// CloseStreamHandler handles close stream request
func (h *StreamAPIHandler) CloseStreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	if err := h.manager.CloseStream(streamID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Stream closed",
		"id":      streamID,
	})
}

// GetStreamStatsHandler handles get stream stats request
func (h *StreamAPIHandler) GetStreamStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	stats, err := h.manager.GetStreamStats(streamID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// PauseStreamHandler handles pause stream request
func (h *StreamAPIHandler) PauseStreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	if err := h.manager.PauseStream(streamID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Stream paused",
		"id":      streamID,
	})
}

// ResumeStreamHandler handles resume stream request
func (h *StreamAPIHandler) ResumeStreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	if err := h.manager.ResumeStream(streamID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Stream resumed",
		"id":      streamID,
	})
}

// GetMessagesHandler handles get messages request
func (h *StreamAPIHandler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	sinceStr := r.URL.Query().Get("since")
	var sinceSeq int64
	if sinceStr != "" {
		sinceSeq = parseSequence(sinceStr)
	}

	messages, err := h.manager.GetMessages(streamID, sinceSeq)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"stream_id": streamID,
		"messages":  messages,
		"count":     len(messages),
		"since":     sinceSeq,
	})
}

// PushMessageHandler handles push message request
func (h *StreamAPIHandler) PushMessageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	var data interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.manager.PushMessage(streamID, data); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stream, ok := h.manager.GetStream(streamID)
	if ok {
		stream.mu.RLock()
		seq := stream.LastSeq
		stream.mu.RUnlock()

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message":   "Message pushed",
			"stream_id": streamID,
			"sequence":  seq,
		})
	} else {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message":   "Message pushed",
			"stream_id": streamID,
		})
	}
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"message": message,
		"code":    status,
	})
}
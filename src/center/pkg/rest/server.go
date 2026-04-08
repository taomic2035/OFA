package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/service"
	"github.com/ofa/center/pkg/metrics"
	pb "github.com/ofa/center/proto"
)

// Server implements the REST API server
type Server struct {
	service        *service.CenterService
	config         *config.Config
	router         *mux.Router
	server         *http.Server
	metrics        *metrics.Metrics
	dashboardDir   string
}

// NewServer creates a new REST server
func NewServer(service *service.CenterService, config *config.Config) *Server {
	s := &Server{
		service:      service,
		config:       config,
		router:       mux.NewRouter(),
		metrics:      metrics.NewMetrics(),
		dashboardDir: "./dashboard",
	}

	s.setupRoutes()
	return s
}

// SetDashboardDir sets the dashboard static files directory
func (s *Server) SetDashboardDir(dir string) {
	s.dashboardDir = dir
}

// Start starts the REST server
func (s *Server) Start(address string) error {
	s.server = &http.Server{
		Addr:    address,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

// Stop stops the REST server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

// GetMetrics returns the metrics instance for external use
func (s *Server) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Prometheus metrics endpoint
	s.router.Handle("/metrics", s.metrics.Handler()).Methods("GET")

	// Health check (with metrics)
	s.router.HandleFunc("/health", s.withMetrics(s.healthCheck)).Methods("GET")

	// Agent endpoints (with metrics)
	s.router.HandleFunc("/api/v1/agents", s.withMetrics(s.listAgents)).Methods("GET")
	s.router.HandleFunc("/api/v1/agents/{id}", s.withMetrics(s.getAgent)).Methods("GET")
	s.router.HandleFunc("/api/v1/agents/{id}", s.withMetrics(s.deleteAgent)).Methods("DELETE")

	// Task endpoints (with metrics)
	s.router.HandleFunc("/api/v1/tasks", s.withMetrics(s.submitTask)).Methods("POST")
	s.router.HandleFunc("/api/v1/tasks", s.withMetrics(s.listTasks)).Methods("GET")
	s.router.HandleFunc("/api/v1/tasks/{id}", s.withMetrics(s.getTask)).Methods("GET")
	s.router.HandleFunc("/api/v1/tasks/{id}/cancel", s.withMetrics(s.cancelTask)).Methods("POST")

	// Message endpoints (with metrics)
	s.router.HandleFunc("/api/v1/messages", s.withMetrics(s.sendMessage)).Methods("POST")
	s.router.HandleFunc("/api/v1/messages/broadcast", s.withMetrics(s.broadcast)).Methods("POST")
	s.router.HandleFunc("/api/v1/messages/multicast", s.withMetrics(s.multicast)).Methods("POST")

	// System endpoints (with metrics)
	s.router.HandleFunc("/api/v1/system/info", s.withMetrics(s.getSystemInfo)).Methods("GET")
	s.router.HandleFunc("/api/v1/system/metrics", s.withMetrics(s.getMetrics)).Methods("GET")

	// Skill endpoints (with metrics)
	s.router.HandleFunc("/api/v1/skills", s.withMetrics(s.listSkills)).Methods("GET")

	// TTS endpoints (v5.6.4)
	s.router.HandleFunc("/api/v1/tts/synthesize", s.withMetrics(s.ttsSynthesize)).Methods("POST")
	s.router.HandleFunc("/api/v1/tts/voices", s.withMetrics(s.ttsListVoices)).Methods("GET")
	s.router.HandleFunc("/api/v1/tts/clone", s.withMetrics(s.ttsCloneVoice)).Methods("POST")
	s.router.HandleFunc("/api/v1/tts/identity/{id}/voice", s.withMetrics(s.ttsSetIdentityVoice)).Methods("PUT")
	s.router.HandleFunc("/api/v1/tts/identity/{id}/voice", s.withMetrics(s.ttsGetIdentityVoice)).Methods("GET")

	// Dashboard static files
	s.setupDashboardRoutes()
}

// setupDashboardRoutes configures dashboard static file serving
func (s *Server) setupDashboardRoutes() {
	// Dashboard routes
	dashboardHandler := s.serveDashboard()
	s.router.PathPrefix("/dashboard").Handler(dashboardHandler)
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}

// serveDashboard returns a handler for serving dashboard static files
func (s *Server) serveDashboard() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if dashboard directory exists
		if _, err := os.Stat(s.dashboardDir); os.IsNotExist(err) {
			http.Error(w, "Dashboard not found. Build the dashboard first: cd src/dashboard && npm install && npm run build", http.StatusNotFound)
			return
		}

		// Serve from dist directory
		distDir := filepath.Join(s.dashboardDir, "dist")
		if _, err := os.Stat(distDir); os.IsNotExist(err) {
			http.Error(w, "Dashboard not built. Run: cd src/dashboard && npm run build", http.StatusNotFound)
			return
		}

		// Get the path relative to /dashboard
		relPath := r.URL.Path
		if relPath == "/dashboard" || relPath == "/dashboard/" {
			relPath = "/dashboard/index.html"
		}

		// Remove /dashboard prefix and serve from dist
		filePath := filepath.Join(distDir, relPath[len("/dashboard/"):])

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// For SPA, serve index.html for unknown routes
			filePath = filepath.Join(distDir, "index.html")
		}

		// Serve the file
		http.ServeFile(w, r, filePath)
	})
}

// withMetrics wraps a handler with Prometheus metrics recording
func (s *Server) withMetrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(rw, r)

		// Record metrics
		duration := time.Since(start)
		path := r.URL.Path
		if mux.Vars(r) != nil {
			// Replace path parameters with placeholders for better aggregation
			for range mux.Vars(r) {
				path = r.URL.Path
				break
			}
		}
		s.metrics.RecordRequest(r.Method, path, rw.statusCode, duration)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ===== Handlers =====

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.metrics.IncrementHealthCheck()
	jsonResponse(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": s.config.Server.Version,
	})
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	agentType := parseInt(r.URL.Query().Get("type"))
	status := parseInt(r.URL.Query().Get("status"))
	page := parseInt(r.URL.Query().Get("page"))
	pageSize := parseInt(r.URL.Query().Get("page_size"))

	resp, err := s.service.ListAgents(ctx, &pb.ListAgentsRequest{
		Type:     pb.AgentType(agentType),
		Status:   pb.AgentStatus(status),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) getAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.service.GetAgent(ctx, &pb.GetAgentRequest{AgentId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	if !resp.Success {
		errorResponse(w, fmt.Errorf(resp.Error))
		return
	}

	jsonResponse(w, http.StatusOK, resp.Agent)
}

func (s *Server) deleteAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.service.DeleteAgent(ctx, &pb.DeleteAgentRequest{AgentId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) submitTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TaskSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.service.SubmitTask(ctx, &pb.SubmitTaskRequest{
		SkillId:     req.SkillID,
		Input:       req.Input,
		TargetAgent: req.TargetAgent,
		Priority:    req.Priority,
		TimeoutMs:   req.TimeoutMs,
		Metadata:    req.Metadata,
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := parseInt(r.URL.Query().Get("status"))
	agentID := r.URL.Query().Get("agent_id")
	page := parseInt(r.URL.Query().Get("page"))
	pageSize := parseInt(r.URL.Query().Get("page_size"))

	resp, err := s.service.ListTasks(ctx, &pb.ListTasksRequest{
		Status:   pb.TaskStatus(status),
		AgentId:  agentID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.service.GetTask(ctx, &pb.GetTaskRequest{TaskId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	if !resp.Success {
		errorResponse(w, fmt.Errorf(resp.Error))
		return
	}

	jsonResponse(w, http.StatusOK, resp.Task)
}

func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req CancelTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = "User requested"
	}

	resp, err := s.service.CancelTask(ctx, &pb.CancelTaskRequest{
		TaskId: id,
		Reason: req.Reason,
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MessageSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.service.SendMessage(ctx, &pb.SendMessageRequest{
		Message: &pb.Message{
			FromAgent: req.FromAgent,
			ToAgent:   req.ToAgent,
			Action:    req.Action,
			Payload:   req.Payload,
			Ttl:       req.TTL,
			Headers:   req.Headers,
		},
		RequireAck: req.RequireAck,
		TimeoutMs:  req.TimeoutMs,
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) broadcast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.service.Broadcast(ctx, &pb.BroadcastRequest{
		FromAgent: req.FromAgent,
		Action:    req.Action,
		Payload:   req.Payload,
		Ttl:       req.TTL,
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) multicast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MulticastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.service.Multicast(ctx, &pb.MulticastRequest{
		FromAgent: req.FromAgent,
		ToAgents:  req.ToAgents,
		Action:    req.Action,
		Payload:   req.Payload,
		Ttl:       req.TTL,
	})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) getSystemInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := s.service.GetSystemInfo(ctx, &pb.GetSystemInfoRequest{})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := s.service.GetMetrics(ctx, &pb.GetMetricsRequest{})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) listSkills(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	category := r.URL.Query().Get("category")

	// Get capabilities and filter
	capResp, err := s.service.GetCapabilities(ctx, &pb.GetCapabilitiesRequest{})
	if err != nil {
		errorResponse(w, err)
		return
	}

	// Filter by category if specified
	var skills []*pb.Capability
	for _, cap := range capResp.Capabilities {
		if category == "" || cap.Category == category {
			skills = append(skills, cap)
		}
	}

	jsonResponse(w, http.StatusOK, &pb.ListSkillsResponse{Skills: skills})
}

// ===== TTS Handlers (v5.6.4) =====

func (s *Server) ttsSynthesize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TTSSynthesizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	// Get TTS service
	ttsService := s.service.GetTTSService()
	if ttsService == nil {
		errorResponse(w, fmt.Errorf("TTS service not initialized"))
		return
	}

	// Synthesize
	synthesizeReq := &service.SynthesizeRequest{
		IdentityID: req.IdentityID,
		Text:       req.Text,
		VoiceID:    req.VoiceID,
		Format:     req.Format,
		SampleRate: req.SampleRate,
		Rate:       req.Rate,
		Pitch:      req.Pitch,
		Volume:     req.Volume,
		Emotion:    req.Emotion,
		Style:      req.Style,
		Streaming:  req.Streaming,
	}
	result, err := ttsService.Synthesize(ctx, synthesizeReq)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) ttsListVoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	provider := r.URL.Query().Get("provider")

	// Get TTS service
	ttsService := s.service.GetTTSService()
	if ttsService == nil {
		errorResponse(w, fmt.Errorf("TTS service not initialized"))
		return
	}

	// List voices
	voices, err := ttsService.ListVoices(ctx, provider)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"voices": voices,
	})
}

func (s *Server) ttsCloneVoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TTSCloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	// Get TTS service
	ttsService := s.service.GetTTSService()
	if ttsService == nil {
		errorResponse(w, fmt.Errorf("TTS service not initialized"))
		return
	}

	// Clone voice
	// Convert rest.TTSCloneRequest to service.CloneVoiceRequest
	refAudios := make([]service.ReferenceAudio, len(req.ReferenceAudios))
	for i, a := range req.ReferenceAudios {
		refAudios[i] = service.ReferenceAudio{
			AudioURL:      a.AudioURL,
			DurationMs:    a.DurationMs,
			Transcription: a.Transcription,
		}
	}
	cloneReq := &service.CloneVoiceRequest{
		IdentityID:      req.IdentityID,
		VoiceName:       req.VoiceName,
		Language:        req.Language,
		ReferenceAudios: refAudios,
	}
	result, err := ttsService.CloneVoice(ctx, cloneReq)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) ttsSetIdentityVoice(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["id"]

	var req struct {
		VoiceID string `json:"voice_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	// Get TTS service
	ttsService := s.service.GetTTSService()
	if ttsService == nil {
		errorResponse(w, fmt.Errorf("TTS service not initialized"))
		return
	}

	// Set voice mapping
	ttsService.SetVoiceForIdentity(identityID, req.VoiceID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identity_id": identityID,
		"voice_id":    req.VoiceID,
		"success":     true,
	})
}

func (s *Server) ttsGetIdentityVoice(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["id"]

	// Get TTS service
	ttsService := s.service.GetTTSService()
	if ttsService == nil {
		errorResponse(w, fmt.Errorf("TTS service not initialized"))
		return
	}

	// Get voice mapping
	voiceID := ttsService.GetVoiceForIdentity(identityID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identity_id": identityID,
		"voice_id":    voiceID,
	})
}

// ===== Helper Functions =====

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, err error) {
	jsonResponse(w, http.StatusInternalServerError, map[string]string{
		"error": err.Error(),
	})
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

// ===== Request Types =====

type TaskSubmitRequest struct {
	SkillID     string            `json:"skill_id"`
	Input       []byte            `json:"input"`
	TargetAgent string            `json:"target_agent,omitempty"`
	Priority    int32             `json:"priority,omitempty"`
	TimeoutMs   int64             `json:"timeout_ms,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type CancelTaskRequest struct {
	Reason string `json:"reason"`
}

type MessageSendRequest struct {
	FromAgent  string            `json:"from_agent"`
	ToAgent    string            `json:"to_agent"`
	Action     string            `json:"action"`
	Payload    []byte            `json:"payload"`
	TTL        int32             `json:"ttl"`
	Headers    map[string]string `json:"headers,omitempty"`
	RequireAck bool              `json:"require_ack"`
	TimeoutMs  int64             `json:"timeout_ms"`
}

type BroadcastRequest struct {
	FromAgent string `json:"from_agent"`
	Action    string `json:"action"`
	Payload   []byte `json:"payload"`
	TTL       int32  `json:"ttl"`
}

type MulticastRequest struct {
	FromAgent string   `json:"from_agent"`
	ToAgents  []string `json:"to_agents"`
	Action    string   `json:"action"`
	Payload   []byte   `json:"payload"`
	TTL       int32    `json:"ttl"`
}

// ===== TTS Request Types (v5.6.4) =====

type TTSSynthesizeRequest struct {
	IdentityID string  `json:"identity_id"`
	Text       string  `json:"text"`
	VoiceID    string  `json:"voice_id,omitempty"`
	Format     string  `json:"format,omitempty"`
	SampleRate int     `json:"sample_rate,omitempty"`
	Rate       float64 `json:"rate,omitempty"`
	Pitch      float64 `json:"pitch,omitempty"`
	Volume     float64 `json:"volume,omitempty"`
	Emotion    string  `json:"emotion,omitempty"`
	Style      string  `json:"style,omitempty"`
	Streaming  bool    `json:"streaming,omitempty"`
}

type TTSCloneRequest struct {
	IdentityID      string           `json:"identity_id"`
	VoiceName       string           `json:"voice_name"`
	Language        string           `json:"language"`
	ReferenceAudios []TTSRefAudio    `json:"reference_audios"`
}

type TTSRefAudio struct {
	AudioURL      string `json:"audio_url"`
	DurationMs    int    `json:"duration_ms"`
	Transcription string `json:"transcription"`
}
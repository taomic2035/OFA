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
	"github.com/ofa/center/internal/models"
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

	// === Identity API (v2.0.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/identities", s.withMetrics(s.createIdentity)).Methods("POST")
	s.router.HandleFunc("/api/v1/identities", s.withMetrics(s.listIdentities)).Methods("GET")
	s.router.HandleFunc("/api/v1/identities/{id}", s.withMetrics(s.getIdentity)).Methods("GET")
	s.router.HandleFunc("/api/v1/identities/{id}", s.withMetrics(s.updateIdentity)).Methods("PUT")
	s.router.HandleFunc("/api/v1/identities/{id}", s.withMetrics(s.deleteIdentity)).Methods("DELETE")

	// === Device API (v2.6.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/devices", s.withMetrics(s.registerDevice)).Methods("POST")
	s.router.HandleFunc("/api/v1/devices", s.withMetrics(s.listDevices)).Methods("GET")
	s.router.HandleFunc("/api/v1/devices/{id}", s.withMetrics(s.getDevice)).Methods("GET")
	s.router.HandleFunc("/api/v1/devices/{id}/heartbeat", s.withMetrics(s.deviceHeartbeat)).Methods("POST")

	// === Behavior API (v2.4.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/behaviors", s.withMetrics(s.reportBehavior)).Methods("POST")
	s.router.HandleFunc("/api/v1/behaviors/{identity_id}", s.withMetrics(s.listBehaviors)).Methods("GET")

	// === Emotion API (v4.0.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/emotions/trigger", s.withMetrics(s.triggerEmotion)).Methods("POST")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}", s.withMetrics(s.getEmotion)).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/context", s.withMetrics(s.getEmotionContext)).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/profile", s.withMetrics(s.getEmotionProfile)).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/profile", s.withMetrics(s.updateEmotionProfile)).Methods("PUT")

	// === Philosophy API (v4.1.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/philosophy/worldview", s.withMetrics(s.setWorldview)).Methods("POST")
	s.router.HandleFunc("/api/v1/philosophy/{identity_id}/worldview", s.withMetrics(s.getWorldview)).Methods("GET")
	s.router.HandleFunc("/api/v1/philosophy/{identity_id}/context", s.withMetrics(s.getPhilosophyContext)).Methods("GET")

	// === Social Identity API (v4.2.0/v6.2.0) ===
	s.router.HandleFunc("/api/v1/social/{identity_id}", s.withMetrics(s.getSocialIdentity)).Methods("GET")
	s.router.HandleFunc("/api/v1/social/{identity_id}", s.withMetrics(s.updateSocialIdentity)).Methods("PUT")
	s.router.HandleFunc("/api/v1/social/{identity_id}/education", s.withMetrics(s.getEducation)).Methods("GET")
	s.router.HandleFunc("/api/v1/social/{identity_id}/education", s.withMetrics(s.updateEducation)).Methods("PUT")
	s.router.HandleFunc("/api/v1/social/{identity_id}/career", s.withMetrics(s.getCareer)).Methods("GET")
	s.router.HandleFunc("/api/v1/social/{identity_id}/career", s.withMetrics(s.updateCareer)).Methods("PUT")
	s.router.HandleFunc("/api/v1/social/{identity_id}/context", s.withMetrics(s.getSocialContext)).Methods("GET")

	// === Culture API (v4.3.0/v6.2.0) ===
	s.router.HandleFunc("/api/v1/culture/{identity_id}", s.withMetrics(s.getCulture)).Methods("GET")
	s.router.HandleFunc("/api/v1/culture/{identity_id}", s.withMetrics(s.updateCulture)).Methods("PUT")
	s.router.HandleFunc("/api/v1/culture/{identity_id}/location", s.withMetrics(s.setLocation)).Methods("POST")
	s.router.HandleFunc("/api/v1/culture/{identity_id}/context", s.withMetrics(s.getCultureContext)).Methods("GET")

	// === LifeStage API (v4.4.0/v6.2.0) ===
	s.router.HandleFunc("/api/v1/lifestage/{identity_id}", s.withMetrics(s.getLifeStage)).Methods("GET")
	s.router.HandleFunc("/api/v1/lifestage/{identity_id}", s.withMetrics(s.updateLifeStage)).Methods("PUT")
	s.router.HandleFunc("/api/v1/lifestage/{identity_id}/stage", s.withMetrics(s.setCurrentStage)).Methods("POST")
	s.router.HandleFunc("/api/v1/lifestage/{identity_id}/event", s.withMetrics(s.addLifeEvent)).Methods("POST")
	s.router.HandleFunc("/api/v1/lifestage/{identity_id}/context", s.withMetrics(s.getLifeStageContext)).Methods("GET")

	// === Relationship API (v4.6.0/v6.2.0) ===
	s.router.HandleFunc("/api/v1/relationship/{identity_id}", s.withMetrics(s.getRelationshipSystem)).Methods("GET")
	s.router.HandleFunc("/api/v1/relationship/{identity_id}", s.withMetrics(s.updateRelationshipSystem)).Methods("PUT")
	s.router.HandleFunc("/api/v1/relationship/{identity_id}/add", s.withMetrics(s.addRelationship)).Methods("POST")
	s.router.HandleFunc("/api/v1/relationship/{identity_id}/context", s.withMetrics(s.getRelationshipContext)).Methods("GET")

	// === Sync API (v2.1.0/v6.1.0) ===
	s.router.HandleFunc("/api/v1/sync", s.withMetrics(s.syncData)).Methods("POST")
	s.router.HandleFunc("/api/v1/sync/{identity_id}/state", s.withMetrics(s.getSyncState)).Methods("GET")

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

// ===== Core API Handlers (v6.1.0) =====

// === Identity Handlers ===

func (s *Server) createIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	identityService := s.service.GetIdentityService()
	identity, err := identityService.CreateIdentity(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *Server) listIdentities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := parseInt(r.URL.Query().Get("page"))
	pageSize := parseInt(r.URL.Query().Get("page_size"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	identityService := s.service.GetIdentityService()
	identities, total, err := identityService.ListIdentities(ctx, page, pageSize)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identities": identities,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func (s *Server) getIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	identityService := s.service.GetIdentityService()
	identity, err := identityService.GetIdentity(ctx, id)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *Server) updateIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req UpdateIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	identityService := s.service.GetIdentityService()
	identity, err := identityService.GetIdentity(ctx, id)
	if err != nil {
		errorResponse(w, err)
		return
	}

	if req.Name != "" {
		identity.Name = req.Name
	}
	if req.Nickname != "" {
		identity.Nickname = req.Nickname
	}

	err = identityService.UpdateIdentity(ctx, identity)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *Server) deleteIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	identityService := s.service.GetIdentityService()
	err := identityService.DeleteIdentity(ctx, id)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// === Device Handlers ===

func (s *Server) registerDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	dataService := s.service.GetDataService()
	device, err := dataService.RegisterDevice(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, device)
}

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	identityID := r.URL.Query().Get("identity_id")

	dataService := s.service.GetDataService()
	devices := dataService.GetDevicesByIdentity(identityID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"devices": devices,
	})
}

func (s *Server) getDevice(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	dataService := s.service.GetDataService()
	device := dataService.GetDevice(id)
	if device == nil {
		errorResponse(w, fmt.Errorf("device not found"))
		return
	}

	jsonResponse(w, http.StatusOK, device)
}

func (s *Server) deviceHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	dataService := s.service.GetDataService()
	dataService.UpdateDeviceHeartbeat(id, time.Now().Unix())

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"agent_id": id,
		"status":   req.Status,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// === Behavior Handlers ===

func (s *Server) reportBehavior(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ReportBehaviorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	dataService := s.service.GetDataService()
	err := dataService.ReportBehavior(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"type":        req.Type,
	})
}

func (s *Server) listBehaviors(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	dataService := s.service.GetDataService()
	behaviors := dataService.GetBehaviorObservations(identityID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identity_id": identityID,
		"behaviors":   behaviors,
	})
}

// === Emotion Handlers ===

func (s *Server) triggerEmotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req TriggerEmotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	emotionEngine := s.service.GetEmotionEngine()
	trigger := buildEmotionTrigger(req)
	emotion, err := emotionEngine.TriggerEmotion(ctx, req.IdentityID, trigger)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"emotion":     emotion,
	})
}

func (s *Server) getEmotion(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	emotionEngine := s.service.GetEmotionEngine()
	emotion := emotionEngine.GetEmotion(identityID)
	if emotion == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no emotion state"})
		return
	}

	jsonResponse(w, http.StatusOK, emotion)
}

func (s *Server) getEmotionContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	emotionEngine := s.service.GetEmotionEngine()
	context := emotionEngine.GetEmotionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no emotion context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

func (s *Server) getEmotionProfile(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	emotionEngine := s.service.GetEmotionEngine()
	profile := emotionEngine.GetEmotionProfile(identityID)
	if profile == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no emotion profile"})
		return
	}

	jsonResponse(w, http.StatusOK, profile)
}

func (s *Server) updateEmotionProfile(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	// Build EmotionProfile from request
	profile := &models.EmotionProfile{
		BaseJoyLevel:      getFloatFromMap(req, "base_joy_level"),
		BaseAngerLevel:    getFloatFromMap(req, "base_anger_level"),
		BaseSadnessLevel:  getFloatFromMap(req, "base_sadness_level"),
		BaseFearLevel:     getFloatFromMap(req, "base_fear_level"),
		BaseLoveLevel:     getFloatFromMap(req, "base_love_level"),
		BaseDisgustLevel:  getFloatFromMap(req, "base_disgust_level"),
		BaseDesireLevel:   getFloatFromMap(req, "base_desire_level"),
		EmotionalStability: getFloatFromMap(req, "emotional_stability"),
	}

	emotionEngine := s.service.GetEmotionEngine()
	err := emotionEngine.UpdateEmotionProfile(identityID, profile)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": identityID,
		"profile":     profile,
	})
}

// === Philosophy Handlers ===

func (s *Server) setWorldview(w http.ResponseWriter, r *http.Request) {
	var req SetWorldviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	philosophyEngine := s.service.GetPhilosophyEngine()
	worldview := buildWorldview(req)
	err := philosophyEngine.UpdateWorldview(req.IdentityID, worldview)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"worldview":   worldview,
	})
}

func (s *Server) getWorldview(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	philosophyEngine := s.service.GetPhilosophyEngine()
	worldview := philosophyEngine.GetWorldview(identityID)
	if worldview == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no worldview set"})
		return
	}

	jsonResponse(w, http.StatusOK, worldview)
}

func (s *Server) getPhilosophyContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	philosophyEngine := s.service.GetPhilosophyEngine()
	context := philosophyEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no philosophy context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// === Sync Handlers ===

func (s *Server) syncData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req SyncDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	dataService := s.service.GetDataService()
	result, err := dataService.SyncIdentity(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) getSyncState(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	dataService := s.service.GetDataService()
	state := dataService.GetSyncState(identityID)
	if state == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no sync state"})
		return
	}

	jsonResponse(w, http.StatusOK, state)
}

// === Social Identity Handlers (v4.2.0/v6.2.0) ===

func (s *Server) getSocialIdentity(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	socialEngine := s.service.GetSocialEngine()
	identity := socialEngine.GetSocialIdentity(identityID)
	if identity == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no social identity"})
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *Server) updateSocialIdentity(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	socialEngine := s.service.GetSocialEngine()
	identity := socialEngine.GetOrCreateSocialIdentity(identityID)

	// Update fields from request
	if education, ok := req["education"].(map[string]interface{}); ok {
		identity.Education = parseEducation(education)
	}
	if career, ok := req["career"].(map[string]interface{}); ok {
		identity.Career = parseCareer(career)
	}

	err := socialEngine.UpdateSocialIdentity(identityID, identity)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *Server) getEducation(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	socialEngine := s.service.GetSocialEngine()
	education := socialEngine.GetEducation(identityID)
	if education == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no education data"})
		return
	}

	jsonResponse(w, http.StatusOK, education)
}

func (s *Server) updateEducation(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	socialEngine := s.service.GetSocialEngine()
	education := parseEducation(req)
	err := socialEngine.UpdateEducation(identityID, education)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, education)
}

func (s *Server) getCareer(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	socialEngine := s.service.GetSocialEngine()
	identity := socialEngine.GetSocialIdentity(identityID)
	if identity == nil || identity.Career == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no career data"})
		return
	}

	jsonResponse(w, http.StatusOK, identity.Career)
}

func (s *Server) updateCareer(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	socialEngine := s.service.GetSocialEngine()
	identity := socialEngine.GetOrCreateSocialIdentity(identityID)
	identity.Career = parseCareer(req)

	err := socialEngine.UpdateSocialIdentity(identityID, identity)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, identity.Career)
}

func (s *Server) getSocialContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	socialEngine := s.service.GetSocialEngine()
	context := socialEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no social context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// === Culture Handlers (v4.3.0/v6.2.0) ===

func (s *Server) getCulture(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	cultureEngine := s.service.GetCultureEngine()
	culture := cultureEngine.GetRegionalCulture(identityID)
	if culture == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no culture data"})
		return
	}

	jsonResponse(w, http.StatusOK, culture)
}

func (s *Server) updateCulture(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	cultureEngine := s.service.GetCultureEngine()
	culture := cultureEngine.GetOrCreateRegionalCulture(identityID)

	// Update fields
	if province, ok := req["province"].(string); ok {
		culture.Province = province
	}
	if city, ok := req["city"].(string); ok {
		culture.City = city
	}
	if cityTier, ok := req["city_tier"].(string); ok {
		culture.CityTier = cityTier
	}

	err := cultureEngine.UpdateRegionalCulture(identityID, culture)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, culture)
}

func (s *Server) setLocation(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	cultureEngine := s.service.GetCultureEngine()
	err := cultureEngine.SetLocation(identityID, req.Province, req.City, req.CityTier)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": identityID,
		"province":    req.Province,
		"city":        req.City,
		"city_tier":   req.CityTier,
	})
}

func (s *Server) getCultureContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	cultureEngine := s.service.GetCultureEngine()
	context := cultureEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no culture context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// === LifeStage Handlers (v4.4.0/v6.2.0) ===

func (s *Server) getLifeStage(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	lifestageEngine := s.service.GetLifestageEngine()
	system := lifestageEngine.GetLifeStageSystem(identityID)
	if system == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no lifestage data"})
		return
	}

	jsonResponse(w, http.StatusOK, system)
}

func (s *Server) updateLifeStage(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	lifestageEngine := s.service.GetLifestageEngine()
	system := lifestageEngine.GetOrCreateLifeStageSystem(identityID)

	// Update development metrics if provided
	if metrics, ok := req["development_metrics"].(map[string]interface{}); ok {
		floatMetrics := make(map[string]float64)
		for k, v := range metrics {
			if f, ok := v.(float64); ok {
				floatMetrics[k] = f
			}
		}
		lifestageEngine.UpdateDevelopmentMetrics(identityID, floatMetrics)
	}

	jsonResponse(w, http.StatusOK, system)
}

func (s *Server) setCurrentStage(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req StageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	lifestageEngine := s.service.GetLifestageEngine()
	err := lifestageEngine.SetCurrentStage(identityID, req.StageName, req.Age)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": identityID,
		"stage_name":  req.StageName,
		"age":         req.Age,
	})
}

func (s *Server) addLifeEvent(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	lifestageEngine := s.service.GetLifestageEngine()
	event := parseLifeEvent(req)
	err := lifestageEngine.AddLifeEvent(identityID, event)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": identityID,
		"event":       event,
	})
}

func (s *Server) getLifeStageContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	lifestageEngine := s.service.GetLifestageEngine()
	context := lifestageEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no lifestage context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// === Relationship Handlers (v4.6.0/v6.2.0) ===

func (s *Server) getRelationshipSystem(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	relationshipEngine := s.service.GetRelationshipEngine()
	system := relationshipEngine.GetRelationshipSystem(identityID)
	if system == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no relationship data"})
		return
	}

	jsonResponse(w, http.StatusOK, system)
}

func (s *Server) updateRelationshipSystem(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	relationshipEngine := s.service.GetRelationshipEngine()
	system := relationshipEngine.GetOrCreateRelationshipSystem(identityID)

	// Update profile if provided
	if profile, ok := req["profile"].(map[string]interface{}); ok {
		system.Profile = parseRelationshipProfile(profile)
	}

	err := relationshipEngine.UpdateRelationshipSystem(identityID, system)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, system)
}

func (s *Server) addRelationship(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	relationshipEngine := s.service.GetRelationshipEngine()
	rel := parseRelationship(req)
	err := relationshipEngine.AddRelationship(identityID, rel)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"identity_id":   identityID,
		"relationship":  rel,
	})
}

func (s *Server) getRelationshipContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	relationshipEngine := s.service.GetRelationshipEngine()
	context := relationshipEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no relationship context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// ===== Core API Request Types (v6.1.0) =====

type CreateIdentityRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Nickname   string `json:"nickname,omitempty"`
	Personality map[string]interface{} `json:"personality,omitempty"`
}

type UpdateIdentityRequest struct {
	Name     string `json:"name,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

type RegisterDeviceRequest struct {
	AgentID      string   `json:"agent_id"`
	IdentityID   string   `json:"identity_id"`
	DeviceType   string   `json:"device_type"`
	DeviceName   string   `json:"device_name,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type HeartbeatRequest struct {
	Status  string `json:"status"`
	Battery int    `json:"battery,omitempty"`
	Network string `json:"network,omitempty"`
}

type ReportBehaviorRequest struct {
	AgentID    string                 `json:"agent_id"`
	IdentityID string                 `json:"identity_id"`
	Type       string                 `json:"type"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

type TriggerEmotionRequest struct {
	IdentityID  string  `json:"identity_id"`
	TriggerType string  `json:"trigger_type,omitempty"`
	TriggerDesc string  `json:"trigger_desc,omitempty"`
	EmotionType string  `json:"emotion_type"`
	Intensity   float64 `json:"intensity"`
}

type SetWorldviewRequest struct {
	IdentityID      string  `json:"identity_id"`
	Optimism        float64 `json:"optimism"`
	ChangeBelief    float64 `json:"change_belief"`
	TrustInPeople   float64 `json:"trust_in_people"`
	FateControl     float64 `json:"fate_control,omitempty"`
}

type SyncDataRequest struct {
	AgentID    string `json:"agent_id"`
	IdentityID string `json:"identity_id"`
	SyncType   string `json:"sync_type"`
}

// ===== Helper Functions =====

func buildEmotionTrigger(req TriggerEmotionRequest) map[string]interface{} {
	return map[string]interface{}{
		"trigger_type": req.TriggerType,
		"trigger_desc": req.TriggerDesc,
		"emotion_type": req.EmotionType,
		"intensity":    req.Intensity,
	}
}

func buildWorldview(req SetWorldviewRequest) map[string]interface{} {
	return map[string]interface{}{
		"optimism":        req.Optimism,
		"change_belief":   req.ChangeBelief,
		"trust_in_people": req.TrustInPeople,
		"fate_control":    req.FateControl,
	}
}

// ===== V4.x Soul System Request Types (v6.2.0) =====

type LocationRequest struct {
	Province  string `json:"province"`
	City      string `json:"city"`
	CityTier  string `json:"city_tier"`
}

type StageRequest struct {
	StageName string `json:"stage_name"`
	Age       int    `json:"age"`
}

type RelationshipRequest struct {
	PersonID     string  `json:"person_id"`
	PersonName   string  `json:"person_name"`
	RelationshipType string `json:"relationship_type"`
	Intimacy     float64 `json:"intimacy"`
	Trust        float64 `json:"trust"`
	Importance   float64 `json:"importance"`
}

// ===== V4.x Parser Functions =====

func parseEducation(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if level, ok := data["level"].(string); ok {
		result["level"] = level
	}
	if school, ok := data["school"].(string); ok {
		result["school"] = school
	}
	if major, ok := data["major"].(string); ok {
		result["major"] = major
	}
	if schoolTier, ok := data["school_tier"].(string); ok {
		result["school_tier"] = schoolTier
	}
	return result
}

func parseCareer(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if occupation, ok := data["occupation"].(string); ok {
		result["occupation"] = occupation
	}
	if industry, ok := data["industry"].(string); ok {
		result["industry"] = industry
	}
	if company, ok := data["company"].(string); ok {
		result["company"] = company
	}
	if stage, ok := data["stage"].(string); ok {
		result["stage"] = stage
	}
	if satisfaction, ok := data["satisfaction"].(float64); ok {
		result["satisfaction"] = satisfaction
	}
	return result
}

func parseLifeEvent(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if eventType, ok := data["event_type"].(string); ok {
		result["event_type"] = eventType
	}
	if eventDesc, ok := data["event_desc"].(string); ok {
		result["event_desc"] = eventDesc
	}
	if eventYear, ok := data["event_year"].(int); ok {
		result["event_year"] = eventYear
	}
	if eventYearf, ok := data["event_year"].(float64); ok {
		result["event_year"] = int(eventYearf)
	}
	if impact, ok := data["impact"].(float64); ok {
		result["impact"] = impact
	}
	return result
}

func parseRelationship(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if personID, ok := data["person_id"].(string); ok {
		result["person_id"] = personID
	}
	if personName, ok := data["person_name"].(string); ok {
		result["person_name"] = personName
	}
	if relType, ok := data["relationship_type"].(string); ok {
		result["relationship_type"] = relType
	}
	if intimacy, ok := data["intimacy"].(float64); ok {
		result["intimacy"] = intimacy
	}
	if trust, ok := data["trust"].(float64); ok {
		result["trust"] = trust
	}
	if importance, ok := data["importance"].(float64); ok {
		result["importance"] = importance
	}
	return result
}

func parseRelationshipProfile(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if attachmentStyle, ok := data["attachment_style"].(string); ok {
		result["attachment_style"] = attachmentStyle
	}
	if socialStyle, ok := data["social_style"].(string); ok {
		result["social_style"] = socialStyle
	}
	if conflictStyle, ok := data["conflict_style"].(string); ok {
		result["conflict_style"] = conflictStyle
	}
	if networkSize, ok := data["network_size"].(float64); ok {
		result["network_size"] = networkSize
	}
	if socialCapital, ok := data["social_capital"].(float64); ok {
		result["social_capital"] = socialCapital
	}
	return result
}
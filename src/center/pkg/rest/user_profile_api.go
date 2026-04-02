package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ofa/center/pkg/grpc"
	"github.com/ofa/center/proto"
)

// UserProfileServer - 用户画像 REST API 服务器
type UserProfileServer struct {
	ofaServer *grpc.OFAServer
	router    *mux.Router
}

// NewUserProfileServer 创建用户画像 REST 服务器
func NewUserProfileServer(ofaServer *grpc.OFAServer) *UserProfileServer {
	s := &UserProfileServer{
		ofaServer: ofaServer,
		router:    mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

// GetRouter 返回路由器
func (s *UserProfileServer) GetRouter() *mux.Router {
	return s.router
}

// setupRoutes 配置路由
func (s *UserProfileServer) setupRoutes() {
	// User API
	s.router.HandleFunc("/api/v1/users", s.createUser).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}", s.getUser).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}", s.updateUser).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}", s.deleteUser).Methods("DELETE")

	// Identity API
	s.router.HandleFunc("/api/v1/users/{id}/identity", s.getIdentity).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/personality", s.updatePersonality).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}/personality/infer", s.inferPersonality).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/mbti", s.setMBTI).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}/values", s.setValueSystem).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}/interests", s.getInterests).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/interests", s.addInterest).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/interests/{interestId}", s.removeInterest).Methods("DELETE")

	// Session API
	s.router.HandleFunc("/api/v1/sessions", s.createSession).Methods("POST")
	s.router.HandleFunc("/api/v1/sessions/{id}", s.getSession).Methods("GET")
	s.router.HandleFunc("/api/v1/sessions/{id}/context", s.updateSessionContext).Methods("PUT")
	s.router.HandleFunc("/api/v1/sessions/{id}/end", s.endSession).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/sessions", s.getActiveSessions).Methods("GET")

	// Memory API
	s.router.HandleFunc("/api/v1/users/{id}/memories", s.storeMemory).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/memories/recall", s.recallMemory).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/memories", s.listMemories).Methods("GET")
	s.router.HandleFunc("/api/v1/memories/{id}", s.getMemory).Methods("GET")
	s.router.HandleFunc("/api/v1/memories/{id}", s.deleteMemory).Methods("DELETE")
	s.router.HandleFunc("/api/v1/users/{id}/memories/consolidate", s.consolidateMemory).Methods("POST")

	// Preference API
	s.router.HandleFunc("/api/v1/users/{id}/preferences/{key}", s.getPreference).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/preferences", s.setPreference).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/preferences/{key}", s.deletePreference).Methods("DELETE")
	s.router.HandleFunc("/api/v1/users/{id}/preferences", s.getAllPreferences).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/preferences/learn", s.learnPreference).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/preferences/score", s.getPreferenceScore).Methods("POST")

	// Decision API
	s.router.HandleFunc("/api/v1/users/{id}/decisions", s.decide).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/decisions/quick", s.quickDecide).Methods("POST")
	s.router.HandleFunc("/api/v1/decisions/{id}/confirm", s.confirmDecision).Methods("POST")
	s.router.HandleFunc("/api/v1/decisions/{id}/outcome", s.recordOutcome).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/decisions", s.getDecisionHistory).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/decisions/stats", s.getDecisionStats).Methods("GET")

	// Voice API
	s.router.HandleFunc("/api/v1/users/{id}/voice", s.getVoiceProfile).Methods("GET")
	s.router.HandleFunc("/api/v1/users/{id}/voice", s.updateVoiceProfile).Methods("PUT")
	s.router.HandleFunc("/api/v1/users/{id}/voice/synthesize", s.synthesizeSpeech).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/voice/clone", s.cloneVoice).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/voice/recognize", s.recognizeSpeech).Methods("POST")

	// Context API
	s.router.HandleFunc("/api/v1/sessions/{id}/sync", s.syncContext).Methods("POST")
	s.router.HandleFunc("/api/v1/users/{id}/context", s.getFullContext).Methods("GET")
}

// === User Handlers ===

func (s *UserProfileServer) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req proto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.ofaServer.CreateUser(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetUser(ctx, &proto.GetUserRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.UpdateUser(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.DeleteUser(ctx, &proto.DeleteUserRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Identity Handlers ===

func (s *UserProfileServer) getIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetIdentity(ctx, &proto.GetIdentityRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) updatePersonality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.UpdatePersonalityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.UpdatePersonality(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) inferPersonality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.InferPersonalityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.InferPersonality(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) setMBTI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req struct {
		MBTIType string `json:"mbti_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	// Convert to personality update
	updateReq := &proto.UpdatePersonalityRequest{
		UserId: id,
		Updates: map[string]interface{}{
			"mbti_type": req.MBTIType,
		},
	}

	resp, err := s.ofaServer.UpdatePersonality(ctx, updateReq)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) setValueSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.SetValueSystemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.SetValueSystem(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getInterests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetInterests(ctx, &proto.GetInterestsRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) addInterest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.AddInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.AddInterest(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) removeInterest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	interestId := mux.Vars(r)["interestId"]

	// Use identity service directly
	// For now, return success placeholder
	jsonResponse(w, http.StatusOK, map[string]bool{"success": true})
}

// === Session Handlers ===

func (s *UserProfileServer) createSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req proto.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}

	resp, err := s.ofaServer.CreateSession(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetSession(ctx, &proto.GetSessionRequest{SessionId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) updateSessionContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.UpdateSessionContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.SessionId = id

	resp, err := s.ofaServer.UpdateSessionContext(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) endSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.EndSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Summary = ""
	}
	req.SessionId = id

	resp, err := s.ofaServer.EndSession(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getActiveSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetActiveSessions(ctx, &proto.GetActiveSessionsRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Memory Handlers ===

func (s *UserProfileServer) storeMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.StoreMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.StoreMemory(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) recallMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.RecallMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.RecallMemory(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) listMemories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	req := &proto.ListMemoriesRequest{
		UserId: id,
		Type:   r.URL.Query().Get("type"),
		Layer:  r.URL.Query().Get("layer"),
		Limit:  20,
		Offset: 0,
	}

	resp, err := s.ofaServer.ListMemories(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetMemory(ctx, &proto.GetMemoryRequest{MemoryId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) deleteMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.DeleteMemory(ctx, &proto.DeleteMemoryRequest{MemoryId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) consolidateMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.ConsolidateMemory(ctx, &proto.ConsolidateMemoryRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Preference Handlers ===

func (s *UserProfileServer) getPreference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	key := mux.Vars(r)["key"]

	resp, err := s.ofaServer.GetPreference(ctx, &proto.GetPreferenceRequest{UserId: id, Key: key})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) setPreference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.SetPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.SetPreference(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) deletePreference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	key := mux.Vars(r)["key"]

	resp, err := s.ofaServer.DeletePreference(ctx, &proto.DeletePreferenceRequest{UserId: id, Key: key})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getAllPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetAllPreferences(ctx, &proto.GetAllPreferencesRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) learnPreference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.LearnPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.LearnPreference(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getPreferenceScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.GetPreferenceScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.GetPreferenceScore(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Decision Handlers ===

func (s *UserProfileServer) decide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.DecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.Decide(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) quickDecide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.QuickDecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.QuickDecide(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) confirmDecision(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.ConfirmDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.DecisionId = id

	resp, err := s.ofaServer.ConfirmDecision(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) recordOutcome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.RecordOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.DecisionId = id

	resp, err := s.ofaServer.RecordOutcome(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getDecisionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	req := &proto.GetDecisionHistoryRequest{
		UserId:   id,
		Scenario: r.URL.Query().Get("scenario"),
		Limit:    20,
		Offset:   0,
	}

	resp, err := s.ofaServer.GetDecisionHistory(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getDecisionStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetDecisionStats(ctx, &proto.GetDecisionStatsRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Voice Handlers ===

func (s *UserProfileServer) getVoiceProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	resp, err := s.ofaServer.GetVoiceProfile(ctx, &proto.GetVoiceProfileRequest{UserId: id})
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) updateVoiceProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.UpdateVoiceProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.UpdateVoiceProfile(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) synthesizeSpeech(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.SynthesizeSpeechRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.SynthesizeSpeech(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	// For audio response, set appropriate content type
	if resp.Success && len(resp.AudioData) > 0 {
		w.Header().Set("Content-Type", "audio/"+resp.Format)
		w.Write(resp.AudioData)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) cloneVoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.CloneVoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.CloneVoice(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) recognizeSpeech(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.RecognizeSpeechRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.UserId = id

	resp, err := s.ofaServer.RecognizeSpeech(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

// === Context Handlers ===

func (s *UserProfileServer) syncContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req proto.SyncContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, err)
		return
	}
	req.SessionId = id

	resp, err := s.ofaServer.SyncContext(ctx, &req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *UserProfileServer) getFullContext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	sessionId := r.URL.Query().Get("session_id")

	req := &proto.GetFullContextRequest{
		UserId:    id,
		SessionId: sessionId,
	}

	resp, err := s.ofaServer.GetFullContext(ctx, req)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}
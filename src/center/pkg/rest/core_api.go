package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ofa/center/internal/emotion"
	"github.com/ofa/center/internal/identity"
	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/internal/philosophy"
	"github.com/ofa/center/internal/sync"
)

// CoreAPIServer -核心功能 REST API 服务器
// 补充 v2.x-v4.x 核心模块的 REST API 端点
type CoreAPIServer struct {
	identityService  *identity.Service
	dataService      *sync.DataService
	emotionEngine    *emotion.EmotionEngine
	philosophyEngine *philosophy.Engine
	router           *mux.Router
}

// NewCoreAPIServer 创建核心 API 服务器
func NewCoreAPIServer(
	identityService *identity.Service,
	dataService *sync.DataService,
	emotionEngine *emotion.EmotionEngine,
	philosophyEngine *philosophy.Engine,
) *CoreAPIServer {
	s := &CoreAPIServer{
		identityService:  identityService,
		dataService:      dataService,
		emotionEngine:    emotionEngine,
		philosophyEngine: philosophyEngine,
		router:           mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

// GetRouter 返回路由器
func (s *CoreAPIServer) GetRouter() *mux.Router {
	return s.router
}

// setupRoutes 配置路由
func (s *CoreAPIServer) setupRoutes() {
	// === Identity API (v2.0.0) ===
	s.router.HandleFunc("/api/v1/identities", s.createIdentity).Methods("POST")
	s.router.HandleFunc("/api/v1/identities", s.listIdentities).Methods("GET")
	s.router.HandleFunc("/api/v1/identities/{id}", s.getIdentity).Methods("GET")
	s.router.HandleFunc("/api/v1/identities/{id}", s.updateIdentity).Methods("PUT")
	s.router.HandleFunc("/api/v1/identities/{id}", s.deleteIdentity).Methods("DELETE")

	// === Device API (v2.6.0) ===
	s.router.HandleFunc("/api/v1/devices", s.registerDevice).Methods("POST")
	s.router.HandleFunc("/api/v1/devices", s.listDevices).Methods("GET")
	s.router.HandleFunc("/api/v1/devices/{id}", s.getDevice).Methods("GET")
	s.router.HandleFunc("/api/v1/devices/{id}", s.updateDevice).Methods("PUT")
	s.router.HandleFunc("/api/v1/devices/{id}", s.deleteDevice).Methods("DELETE")
	s.router.HandleFunc("/api/v1/devices/{id}/heartbeat", s.deviceHeartbeat).Methods("POST")

	// === Behavior API (v2.4.0) ===
	s.router.HandleFunc("/api/v1/behaviors", s.reportBehavior).Methods("POST")
	s.router.HandleFunc("/api/v1/behaviors/{identity_id}", s.listBehaviors).Methods("GET")

	// === Emotion API (v4.0.0) ===
	s.router.HandleFunc("/api/v1/emotions/trigger", s.triggerEmotion).Methods("POST")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}", s.getEmotion).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/context", s.getEmotionContext).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/profile", s.getEmotionProfile).Methods("GET")
	s.router.HandleFunc("/api/v1/emotions/{identity_id}/profile", s.updateEmotionProfile).Methods("PUT")

	// === Philosophy API (v4.1.0) ===
	s.router.HandleFunc("/api/v1/philosophy/worldview", s.setWorldview).Methods("POST")
	s.router.HandleFunc("/api/v1/philosophy/{identity_id}/worldview", s.getWorldview).Methods("GET")
	s.router.HandleFunc("/api/v1/philosophy/{identity_id}/context", s.getPhilosophyContext).Methods("GET")

	// === Sync API (v2.1.0) ===
	s.router.HandleFunc("/api/v1/sync", s.syncData).Methods("POST")
	s.router.HandleFunc("/api/v1/sync/{identity_id}/state", s.getSyncState).Methods("GET")
}

// === Identity Handlers ===

func (s *CoreAPIServer) createIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// 构建 identity.Service 请求
	createReq := &identity.CreateIdentityRequest{
		ID:         req.ID,
		Name:       req.Name,
		Nickname:   req.Nickname,
		Avatar:     req.Avatar,
	}

	// 处理性格参数
	if req.Personality != nil {
		createReq.Personality = &models.Personality{
			Openness:          req.Personality.Openness,
			Conscientiousness: req.Personality.Conscientiousness,
			Extraversion:      req.Personality.Extraversion,
			Agreeableness:     req.Personality.Agreeableness,
			Neuroticism:       req.Personality.Neuroticism,
		}
	}

	identity, err := s.identityService.CreateIdentity(ctx, createReq)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *CoreAPIServer) listIdentities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := parseInt(r.URL.Query().Get("page"))
 pageSize := parseInt(r.URL.Query().Get("page_size"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	identities, total, err := s.identityService.ListIdentities(ctx, page, pageSize)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identities": identities,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func (s *CoreAPIServer) getIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	identity, err := s.identityService.GetIdentity(ctx, id)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *CoreAPIServer) updateIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	var req UpdateIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	identity, err := s.identityService.GetIdentity(ctx, id)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	// 应用更新
	if req.Name != "" {
		identity.Name = req.Name
	}
	if req.Nickname != "" {
		identity.Nickname = req.Nickname
	}

	err = s.identityService.UpdateIdentity(ctx, identity)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, identity)
}

func (s *CoreAPIServer) deleteIdentity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	err := s.identityService.DeleteIdentity(ctx, id)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// === Device Handlers ===

func (s *CoreAPIServer) registerDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	device, err := s.dataService.RegisterDevice(ctx, &sync.RegisterDeviceRequest{
		AgentID:      req.AgentID,
		IdentityID:   req.IdentityID,
		DeviceType:   req.DeviceType,
		DeviceName:   req.DeviceName,
		Capabilities: req.Capabilities,
	})
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, device)
}

func (s *CoreAPIServer) listDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identityID := r.URL.Query().Get("identity_id")

	var devices []*sync.DeviceInfo
	if identityID != "" {
		devices = s.dataService.GetDevicesByIdentity(identityID)
	} else {
		// 如果没有指定 identity_id，返回空列表
		devices = []*sync.DeviceInfo{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"devices": devices,
	})
}

func (s *CoreAPIServer) getDevice(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	device := s.dataService.GetDevice(id)
	if device == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	jsonResponse(w, http.StatusOK, device)
}

func (s *CoreAPIServer) updateDevice(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	device := s.dataService.GetDevice(id)
	if device == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "device not found"})
		return
	}

	// 更新状态和优先级
	if req.Status != "" {
		device.Status = req.Status
	}
	if req.Priority > 0 {
		device.Priority = req.Priority
	}

	jsonResponse(w, http.StatusOK, device)
}

func (s *CoreAPIServer) deleteDevice(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.dataService.UnregisterDevice(id)

	jsonResponse(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (s *CoreAPIServer) deviceHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// 更新心跳
	s.dataService.UpdateDeviceHeartbeat(id, time.Now().Unix())

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"agent_id": id,
		"status":   req.Status,
		"time":     time.Now().Format(time.RFC3339),
	})
}

// === Behavior Handlers ===

func (s *CoreAPIServer) reportBehavior(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ReportBehaviorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	report := &sync.BehaviorReport{
		ID:         fmt.Sprintf("behavior-%d", time.Now().UnixNano()),
		AgentID:    req.AgentID,
		IdentityID: req.IdentityID,
		Type:       req.Type,
		Context:    req.Observation,
		Timestamp:  time.Now(),
	}

	err := s.dataService.ReportBehavior(ctx, report)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"type":        req.Type,
	})
}

func (s *CoreAPIServer) listBehaviors(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	behaviors := s.dataService.GetBehaviorObservations(identityID)

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"identity_id": identityID,
		"behaviors":   behaviors,
	})
}

// === Emotion Handlers ===

func (s *CoreAPIServer) triggerEmotion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req TriggerEmotionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	trigger := models.EmotionTrigger{
		TriggerID:   generateTriggerID(),
		TriggerType: req.TriggerType,
		TriggerDesc: req.TriggerDesc,
		EmotionType: req.EmotionType,
		Intensity:   req.Intensity,
		Timestamp:   time.Now(),
	}

	emotion, err := s.emotionEngine.TriggerEmotion(ctx, req.IdentityID, trigger)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"emotion":     emotion,
	})
}

// generateTriggerID generates a unique trigger ID
func generateTriggerID() string {
	return fmt.Sprintf("trigger-%d", time.Now().UnixNano())
}

func (s *CoreAPIServer) getEmotion(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	emotion := s.emotionEngine.GetEmotion(identityID)
	if emotion == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "emotion not found"})
		return
	}

	jsonResponse(w, http.StatusOK, emotion)
}

func (s *CoreAPIServer) getEmotionContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	context := s.emotionEngine.GetEmotionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no emotion context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

func (s *CoreAPIServer) getEmotionProfile(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	profile := s.emotionEngine.GetEmotionProfile(identityID)
	if profile == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "profile not found"})
		return
	}

	jsonResponse(w, http.StatusOK, profile)
}

func (s *CoreAPIServer) updateEmotionProfile(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	var req EmotionProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	profile := &models.EmotionProfile{
		BaseJoyLevel:      req.BaseJoyLevel,
		BaseAngerLevel:    req.BaseAngerLevel,
		BaseSadnessLevel:  req.BaseSadnessLevel,
		BaseFearLevel:     req.BaseFearLevel,
		BaseLoveLevel:     req.BaseLoveLevel,
		BaseDisgustLevel:  req.BaseDisgustLevel,
		BaseDesireLevel:   req.BaseDesireLevel,
		EmotionalStability: req.EmotionalStability,
	}

	err := s.emotionEngine.UpdateEmotionProfile(identityID, profile)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, profile)
}

// === Philosophy Handlers ===

func (s *CoreAPIServer) setWorldview(w http.ResponseWriter, r *http.Request) {
	var req SetWorldviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	worldview := &models.Worldview{
		Optimism:         req.Optimism,
		ChangeBelief:     req.ChangeBelief,
		TrustInPeople:    req.TrustInPeople,
		FateControl:      req.FateControl,
		WorldEssence:     req.WorldEssence,
		SocietyView:      req.SocietyView,
		FutureView:       req.FutureView,
		RelationshipView: req.RelationshipView,
	}

	err := s.philosophyEngine.UpdateWorldview(req.IdentityID, worldview)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"identity_id": req.IdentityID,
		"worldview":   worldview,
	})
}

func (s *CoreAPIServer) getWorldview(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	worldview := s.philosophyEngine.GetWorldview(identityID)
	if worldview == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "worldview not found"})
		return
	}

	jsonResponse(w, http.StatusOK, worldview)
}

func (s *CoreAPIServer) getPhilosophyContext(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	context := s.philosophyEngine.GetDecisionContext(identityID)
	if context == nil {
		jsonResponse(w, http.StatusOK, map[string]string{"message": "no philosophy context"})
		return
	}

	jsonResponse(w, http.StatusOK, context)
}

// === Sync Handlers ===

func (s *CoreAPIServer) syncData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req SyncDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// 使用 SyncIdentity 方法
	result, err := s.dataService.SyncIdentity(ctx, &sync.SyncIdentityRequest{
		AgentID:    req.AgentID,
		IdentityID: req.IdentityID,
		SyncType:   req.SyncType,
	})
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *CoreAPIServer) getSyncState(w http.ResponseWriter, r *http.Request) {
	identityID := mux.Vars(r)["identity_id"]

	state := s.dataService.GetSyncState(identityID)
	if state == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "sync state not found"})
		return
	}

	jsonResponse(w, http.StatusOK, state)
}

// === Request Types ===

type CreateIdentityRequest struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Nickname   string              `json:"nickname,omitempty"`
	Avatar     string              `json:"avatar,omitempty"`
	Personality *PersonalityInput  `json:"personality,omitempty"`
}

type PersonalityInput struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
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

type UpdateDeviceRequest struct {
	Status    string `json:"status,omitempty"`
	Priority  int    `json:"priority,omitempty"`
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
	Observation map[string]interface{} `json:"observation,omitempty"`
}

type TriggerEmotionRequest struct {
	IdentityID   string  `json:"identity_id"`
	TriggerType  string  `json:"trigger_type,omitempty"`
	TriggerDesc  string  `json:"trigger_desc,omitempty"`
	EmotionType  string  `json:"emotion_type"`
	Intensity    float64 `json:"intensity"`
}

type EmotionProfileRequest struct {
	BaseJoyLevel      float64 `json:"base_joy_level"`
	BaseAngerLevel    float64 `json:"base_anger_level"`
	BaseSadnessLevel  float64 `json:"base_sadness_level"`
	BaseFearLevel     float64 `json:"base_fear_level"`
	BaseLoveLevel     float64 `json:"base_love_level"`
	BaseDisgustLevel  float64 `json:"base_disgust_level"`
	BaseDesireLevel   float64 `json:"base_desire_level"`
	EmotionalStability float64 `json:"emotional_stability"`
}

type SetWorldviewRequest struct {
	IdentityID       string  `json:"identity_id"`
	Optimism         float64 `json:"optimism"`
	ChangeBelief    float64 `json:"change_belief"`
	TrustInPeople    float64 `json:"trust_in_people"`
	FateControl      float64 `json:"fate_control,omitempty"`
	WorldEssence     string  `json:"world_essence,omitempty"`
	SocietyView      string  `json:"society_view,omitempty"`
	FutureView       string  `json:"future_view,omitempty"`
	RelationshipView string  `json:"relationship_view,omitempty"`
}

type SyncDataRequest struct {
	AgentID    string        `json:"agent_id"`
	IdentityID string        `json:"identity_id"`
	SyncType   string        `json:"sync_type"`
	Changes    []sync.Change `json:"changes,omitempty"`
}
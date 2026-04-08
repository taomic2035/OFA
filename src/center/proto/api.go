package proto

// === User Service Messages ===

// UserProfile message
type UserProfile struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Avatar       string            `json:"avatar"`
	Email        string            `json:"email"`
	Phone        string            `json:"phone"`
	Locale       string            `json:"locale"`
	Timezone     string            `json:"timezone"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
	Metadata     map[string]string `json:"metadata"`
}

// CreateUserRequest message
type CreateUserRequest struct {
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Locale   string            `json:"locale"`
	Metadata map[string]string `json:"metadata"`
}

// CreateUserResponse message
type CreateUserResponse struct {
	Success bool        `json:"success"`
	User    *UserProfile `json:"user"`
	Error   string      `json:"error"`
}

// GetUserRequest message
type GetUserRequest struct {
	UserId string `json:"user_id"`
}

// GetUserResponse message
type GetUserResponse struct {
	Success bool         `json:"success"`
	User    *UserProfile  `json:"user"`
	Error   string       `json:"error"`
}

// UpdateUserRequest message
type UpdateUserRequest struct {
	UserId   string            `json:"user_id"`
	Name     string            `json:"name"`
	Avatar   string            `json:"avatar"`
	Locale   string            `json:"locale"`
	Timezone string            `json:"timezone"`
	Metadata map[string]string `json:"metadata"`
}

// UpdateUserResponse message
type UpdateUserResponse struct {
	Success bool        `json:"success"`
	User    *UserProfile `json:"user"`
	Error   string      `json:"error"`
}

// DeleteUserRequest message
type DeleteUserRequest struct {
	UserId string `json:"user_id"`
}

// DeleteUserResponse message
type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// === Identity Service Messages ===

// GetIdentityRequest message
type GetIdentityRequest struct {
	UserId string `json:"user_id"`
}

// GetIdentityResponse message
type GetIdentityResponse struct {
	Success    bool            `json:"success"`
	Identity   *PersonalIdentity `json:"identity"`
	Error      string          `json:"error"`
}

// UpdatePersonalityRequest message
type UpdatePersonalityRequest struct {
	UserId string                 `json:"user_id"`
	Updates map[string]interface{} `json:"updates"`
}

// UpdatePersonalityResponse message
type UpdatePersonalityResponse struct {
	Success    bool        `json:"success"`
	Personality *Personality `json:"personality"`
	Error      string      `json:"error"`
}

// InferPersonalityRequest message
type InferPersonalityRequest struct {
	UserId string            `json:"user_id"`
	Events []*BehaviorEvent  `json:"events"`
}

// InferPersonalityResponse message
type InferPersonalityResponse struct {
	Success      bool        `json:"success"`
	Personality  *Personality `json:"personality"`
	Changes      []string    `json:"changes"`
	Confidence   float64     `json:"confidence"`
	Error        string      `json:"error"`
}

// BehaviorEvent message
type BehaviorEvent struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// SetValueSystemRequest message
type SetValueSystemRequest struct {
	UserId      string            `json:"user_id"`
	ValueSystem *ValueSystem      `json:"value_system"`
}

// SetValueSystemResponse message
type SetValueSystemResponse struct {
	Success      bool         `json:"success"`
	ValueSystem  *ValueSystem `json:"value_system"`
	Error        string       `json:"error"`
}

// GetInterestsRequest message
type GetInterestsRequest struct {
	UserId string `json:"user_id"`
}

// GetInterestsResponse message
type GetInterestsResponse struct {
	Success   bool        `json:"success"`
	Interests []*Interest `json:"interests"`
	Error     string      `json:"error"`
}

// AddInterestRequest message
type AddInterestRequest struct {
	UserId     string   `json:"user_id"`
	Category   string   `json:"category"`
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords"`
	Level      float64  `json:"level"`
}

// AddInterestResponse message
type AddInterestResponse struct {
	Success  bool      `json:"success"`
	Interest *Interest `json:"interest"`
	Error    string    `json:"error"`
}

// === Session Service Messages ===

// Session message
type Session struct {
	Id           string                 `json:"id"`
	UserId       string                 `json:"user_id"`
	AgentId      string                 `json:"agent_id"`
	Status       string                 `json:"status"`
	Context      map[string]interface{} `json:"context"`
	ActiveMemory []*Memory              `json:"active_memory"`
	StartedAt    int64                  `json:"started_at"`
	EndedAt      int64                  `json:"ended_at"`
	LastActiveAt int64                  `json:"last_active_at"`
}

// CreateSessionRequest message
type CreateSessionRequest struct {
	UserId  string                 `json:"user_id"`
	AgentId string                 `json:"agent_id"`
	Context map[string]interface{} `json:"context"`
}

// CreateSessionResponse message
type CreateSessionResponse struct {
	Success bool     `json:"success"`
	Session *Session `json:"session"`
	Error   string   `json:"error"`
}

// GetSessionRequest message
type GetSessionRequest struct {
	SessionId string `json:"session_id"`
}

// GetSessionResponse message
type GetSessionResponse struct {
	Success bool     `json:"success"`
	Session *Session `json:"session"`
	Error   string   `json:"error"`
}

// UpdateSessionContextRequest message
type UpdateSessionContextRequest struct {
	SessionId string                 `json:"session_id"`
	Context   map[string]interface{} `json:"context"`
}

// UpdateSessionContextResponse message
type UpdateSessionContextResponse struct {
	Success bool     `json:"success"`
	Session *Session `json:"session"`
	Error   string   `json:"error"`
}

// EndSessionRequest message
type EndSessionRequest struct {
	SessionId string `json:"session_id"`
	Summary   string `json:"summary"`
}

// EndSessionResponse message
type EndSessionResponse struct {
	Success bool     `json:"success"`
	Session *Session `json:"session"`
	Error   string   `json:"error"`
}

// GetActiveSessionsRequest message
type GetActiveSessionsRequest struct {
	UserId string `json:"user_id"`
}

// GetActiveSessionsResponse message
type GetActiveSessionsResponse struct {
	Success  bool       `json:"success"`
	Sessions []*Session `json:"sessions"`
	Error    string     `json:"error"`
}

// === Memory Service API Messages ===

// StoreMemoryRequest message
type StoreMemoryRequest struct {
	UserId     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Content    string                 `json:"content"`
	Importance float64                `json:"importance"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// StoreMemoryResponse message
type StoreMemoryResponse struct {
	Success bool    `json:"success"`
	Memory  *Memory `json:"memory"`
	Error   string  `json:"error"`
}

// RecallMemoryRequest message
type RecallMemoryRequest struct {
	UserId    string                 `json:"user_id"`
	Query     string                 `json:"query"`
	Type      string                 `json:"type"`
	Layer     string                 `json:"layer"`
	Limit     int32                  `json:"limit"`
	Threshold float64                `json:"threshold"`
	Context   map[string]interface{} `json:"context"`
}

// RecallMemoryResponse message
type RecallMemoryResponse struct {
	Success  bool     `json:"success"`
	Memories []*Memory `json:"memories"`
	Total    int32    `json:"total"`
	Error    string   `json:"error"`
}

// GetMemoryRequest message
type GetMemoryRequest struct {
	MemoryId string `json:"memory_id"`
}

// GetMemoryResponse message
type GetMemoryResponse struct {
	Success bool    `json:"success"`
	Memory  *Memory `json:"memory"`
	Error   string  `json:"error"`
}

// DeleteMemoryRequest message
type DeleteMemoryRequest struct {
	MemoryId string `json:"memory_id"`
}

// DeleteMemoryResponse message
type DeleteMemoryResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ListMemoriesRequest message
type ListMemoriesRequest struct {
	UserId string `json:"user_id"`
	Type   string `json:"type"`
	Layer  string `json:"layer"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

// ListMemoriesResponse message
type ListMemoriesResponse struct {
	Success  bool     `json:"success"`
	Memories []*Memory `json:"memories"`
	Total    int32    `json:"total"`
	Error    string   `json:"error"`
}

// ConsolidateMemoryRequest message
type ConsolidateMemoryRequest struct {
	UserId string `json:"user_id"`
}

// ConsolidateMemoryResponse message
type ConsolidateMemoryResponse struct {
	Success         bool    `json:"success"`
	ConsolidatedCount int32  `json:"consolidated_count"`
	PromotedCount   int32   `json:"promoted_count"`
	Error           string  `json:"error"`
}

// === Preference Service API Messages ===

// GetPreferenceRequest message
type GetPreferenceRequest struct {
	UserId string `json:"user_id"`
	Key    string `json:"key"`
}

// GetPreferenceResponse message
type GetPreferenceResponse struct {
	Success    bool        `json:"success"`
	Preference *Preference `json:"preference"`
	Error      string      `json:"error"`
}

// SetPreferenceRequest message
type SetPreferenceRequest struct {
	UserId      string                 `json:"user_id"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Confidence  float64                `json:"confidence"`
	Source      string                 `json:"source"`
	ExpiresAt   int64                  `json:"expires_at"`
}

// SetPreferenceResponse message
type SetPreferenceResponse struct {
	Success    bool        `json:"success"`
	Preference *Preference `json:"preference"`
	Error      string      `json:"error"`
}

// DeletePreferenceRequest message
type DeletePreferenceRequest struct {
	UserId string `json:"user_id"`
	Key    string `json:"key"`
}

// DeletePreferenceResponse message
type DeletePreferenceResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// GetAllPreferencesRequest message
type GetAllPreferencesRequest struct {
	UserId string `json:"user_id"`
}

// GetAllPreferencesResponse message
type GetAllPreferencesResponse struct {
	Success     bool         `json:"success"`
	Preferences []*Preference `json:"preferences"`
	Error       string       `json:"error"`
}

// LearnPreferenceRequest message
type LearnPreferenceRequest struct {
	UserId   string                 `json:"user_id"`
	EventType string                `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
}

// LearnPreferenceResponse message
type LearnPreferenceResponse struct {
	Success       bool         `json:"success"`
	LearnedPrefs  []*Preference `json:"learned_prefs"`
	UpdatedPrefs  []*Preference `json:"updated_prefs"`
	Error         string       `json:"error"`
}

// GetPreferenceScoreRequest message
type GetPreferenceScoreRequest struct {
	UserId string                 `json:"user_id"`
	Item   map[string]interface{} `json:"item"`
}

// GetPreferenceScoreResponse message
type GetPreferenceScoreResponse struct {
	Success bool               `json:"success"`
	Score   float64            `json:"score"`
	Details map[string]float64 `json:"details"`
	Error   string             `json:"error"`
}

// === Decision Service API Messages ===

// DecideRequest (already defined in decision.go, alias here)
// Using the existing DecideRequest/DecideResponse

// GetDecisionHistoryRequest message
type GetDecisionHistoryRequest struct {
	UserId    string `json:"user_id"`
	Scenario  string `json:"scenario"`
	Outcome   string `json:"outcome"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Limit     int32  `json:"limit"`
	Offset    int32  `json:"offset"`
}

// GetDecisionHistoryResponse message
type GetDecisionHistoryResponse struct {
	Success   bool        `json:"success"`
	Decisions []*Decision `json:"decisions"`
	Total     int32       `json:"total"`
	Error     string      `json:"error"`
}

// GetDecisionStatsRequest message
type GetDecisionStatsRequest struct {
	UserId string `json:"user_id"`
}

// GetDecisionStatsResponse message
type GetDecisionStatsResponse struct {
	Success bool            `json:"success"`
	Stats   *DecisionStats  `json:"stats"`
	Error   string          `json:"error"`
}

// === Voice Service API Messages ===

// GetVoiceProfileRequest message
type GetVoiceProfileRequest struct {
	UserId string `json:"user_id"`
}

// GetVoiceProfileResponse message
type GetVoiceProfileResponse struct {
	Success      bool          `json:"success"`
	VoiceProfile *VoiceProfile  `json:"voice_profile"`
	Error        string        `json:"error"`
}

// UpdateVoiceProfileRequest message
type UpdateVoiceProfileRequest struct {
	UserId             string  `json:"user_id"`
	DefaultTtsVoice    string  `json:"default_tts_voice"`
	SpeechRate         float64 `json:"speech_rate"`
	SpeechPitch        float64 `json:"speech_pitch"`
	Volume             float64 `json:"volume"`
	PreferredLanguage  string  `json:"preferred_language"`
}

// UpdateVoiceProfileResponse message
type UpdateVoiceProfileResponse struct {
	Success      bool          `json:"success"`
	VoiceProfile *VoiceProfile  `json:"voice_profile"`
	Error        string        `json:"error"`
}

// SynthesizeSpeechRequest message
type SynthesizeSpeechRequest struct {
	UserId   string `json:"user_id"`
	Text     string `json:"text"`
	VoiceId  string `json:"voice_id"`
	Format   string `json:"format"` // mp3, wav, ogg
}

// SynthesizeSpeechResponse message
type SynthesizeSpeechResponse struct {
	Success     bool    `json:"success"`
	AudioData   []byte  `json:"audio_data"`
	DurationMs  int64   `json:"duration_ms"`
	Format      string  `json:"format"`
	Error       string  `json:"error"`
}

// CloneVoiceRequest message
type CloneVoiceRequest struct {
	UserId     string `json:"user_id"`
	SampleData []byte `json:"sample_data"`
	SampleFormat string `json:"sample_format"`
	Name       string `json:"name"`
}

// CloneVoiceResponse message
type CloneVoiceResponse struct {
	Success  bool         `json:"success"`
	Voice    *VoiceProfile `json:"voice"`
	Error    string       `json:"error"`
}

// RecognizeSpeechRequest message
type RecognizeSpeechRequest struct {
	UserId     string `json:"user_id"`
	AudioData  []byte `json:"audio_data"`
	AudioFormat string `json:"audio_format"`
	Language   string `json:"language"`
}

// RecognizeSpeechResponse message
type RecognizeSpeechResponse struct {
	Success      bool    `json:"success"`
	Transcript   string  `json:"transcript"`
	Confidence   float64 `json:"confidence"`
	DetectedLanguage string `json:"detected_language"`
	Error        string  `json:"error"`
}

// === Context Sync Messages ===

// SyncContextRequest message
type SyncContextRequest struct {
	UserId    string                 `json:"user_id"`
	AgentId   string                 `json:"agent_id"`
	SessionId string                 `json:"session_id"`
	Changes   map[string]interface{} `json:"changes"`
}

// SyncContextResponse message
type SyncContextResponse struct {
	Success      bool                 `json:"success"`
	SyncedKeys   []string             `json:"synced_keys"`
	CurrentContext map[string]interface{} `json:"current_context"`
	Error        string               `json:"error"`
}

// GetFullContextRequest message
type GetFullContextRequest struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
}

// GetFullContextResponse message
type GetFullContextResponse struct {
	Success  bool                   `json:"success"`
	Context  *DecisionContext       `json:"context"`
	Error    string                  `json:"error"`
}
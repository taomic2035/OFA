package proto

// Identity-related message types for v1.2.0

// === 身份核心 ===

// Personality message placeholder
type Personality struct {
	Openness          float64            `json:"openness"`
	Conscientiousness float64            `json:"conscientiousness"`
	Extraversion      float64            `json:"extraversion"`
	Agreeableness     float64            `json:"agreeableness"`
	Neuroticism       float64            `json:"neuroticism"`
	CustomTraits      map[string]float64 `json:"custom_traits"`
	SpeakingTone      string             `json:"speaking_tone"`
	ResponseLength    string             `json:"response_length"`
	EmojiUsage        float64            `json:"emoji_usage"`
	Summary           string             `json:"summary"`
}

// ValueSystem message placeholder
type ValueSystem struct {
	Privacy        float64            `json:"privacy"`
	Efficiency     float64            `json:"efficiency"`
	Health         float64            `json:"health"`
	Family         float64            `json:"family"`
	Career         float64            `json:"career"`
	Entertainment  float64            `json:"entertainment"`
	Learning       float64            `json:"learning"`
	Social         float64            `json:"social"`
	Finance        float64            `json:"finance"`
	Environment    float64            `json:"environment"`
	RiskTolerance  float64            `json:"risk_tolerance"`
	Impulsiveness  float64            `json:"impulsiveness"`
	Patience       float64            `json:"patience"`
	CustomValues   map[string]float64 `json:"custom_values"`
	Summary        string             `json:"summary"`
}

// Interest message placeholder
type Interest struct {
	Id          string   `json:"id"`
	Category    string   `json:"category"`
	Name        string   `json:"name"`
	Level       float64  `json:"level"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
	Since       int64    `json:"since"`
	LastActive  int64    `json:"last_active"`
}

// VoiceProfile message placeholder
type VoiceProfile struct {
	Id               string  `json:"id"`
	VoiceType        string  `json:"voice_type"`
	PresetVoiceId    string  `json:"preset_voice_id"`
	CloneReferenceId string  `json:"clone_reference_id"`
	Pitch            float64 `json:"pitch"`
	Speed            float64 `json:"speed"`
	Volume           float64 `json:"volume"`
	Tone             string  `json:"tone"`
	Accent           string  `json:"accent"`
	EmotionLevel     float64 `json:"emotion_level"`
	PausePattern     string  `json:"pause_pattern"`
	EmphasisStyle    string  `json:"emphasis_style"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
}

// WritingStyle message placeholder
type WritingStyle struct {
	Formality         float64  `json:"formality"`
	Verbosity         float64  `json:"verbosity"`
	Humor             float64  `json:"humor"`
	Technicality      float64  `json:"technicality"`
	UseEmoji          bool     `json:"use_emoji"`
	UseGifs           bool     `json:"use_gifs"`
	UseMarkdown       bool     `json:"use_markdown"`
	SignaturePhrase   string   `json:"signature_phrase"`
	FrequentWords     []string `json:"frequent_words"`
	AvoidWords        []string `json:"avoid_words"`
	PreferredGreeting string   `json:"preferred_greeting"`
	PreferredClosing  string   `json:"preferred_closing"`
}

// PersonalIdentity message placeholder
type PersonalIdentity struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Nickname   string        `json:"nickname"`
	Avatar     string        `json:"avatar"`
	Birthday   int64         `json:"birthday"`
	Gender     string        `json:"gender"`
	Location   string        `json:"location"`
	Occupation string        `json:"occupation"`
	Languages  []string      `json:"languages"`
	Timezone   string        `json:"timezone"`
	Personality    *Personality    `json:"personality"`
	ValueSystem    *ValueSystem    `json:"value_system"`
	Interests      []*Interest     `json:"interests"`
	VoiceProfile   *VoiceProfile   `json:"voice_profile"`
	WritingStyle   *WritingStyle   `json:"writing_style"`
	CreatedAt      int64           `json:"created_at"`
	UpdatedAt      int64           `json:"updated_at"`
}

// BehaviorObservation message placeholder
type BehaviorObservation struct {
	Id         string                 `json:"id"`
	UserId     string                 `json:"user_id"`
	Type       string                 `json:"type"`
	Context    map[string]interface{} `json:"context"`
	Outcome    string                 `json:"outcome"`
	Inferences map[string]float64     `json:"inferences"`
	Timestamp  int64                  `json:"timestamp"`
}

// DecisionContext message placeholder
type DecisionContext struct {
	UserId         string        `json:"user_id"`
	Personality    *Personality  `json:"personality"`
	ValueSystem    *ValueSystem  `json:"value_system"`
	Interests      []*Interest   `json:"interests"`
	SpeakingTone   string        `json:"speaking_tone"`
	ResponseLength string        `json:"response_length"`
	ValuePriority  []string      `json:"value_priority"`
}

// === API 请求/响应 ===

// CreateIdentityRequest message placeholder
type CreateIdentityRequest struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Nickname   string   `json:"nickname"`
	Avatar     string   `json:"avatar"`
	Birthday   int64    `json:"birthday"`
	Gender     string   `json:"gender"`
	Location   string   `json:"location"`
	Occupation string   `json:"occupation"`
	Languages  []string `json:"languages"`
	Timezone   string   `json:"timezone"`
}

// CreateIdentityResponse message placeholder
type CreateIdentityResponse struct {
	Success  bool              `json:"success"`
	Identity *PersonalIdentity `json:"identity"`
	Error    string            `json:"error"`
}

// GetIdentityRequest message placeholder
type GetIdentityRequest struct {
	Id string `json:"id"`
}

// GetIdentityResponse message placeholder
type GetIdentityResponse struct {
	Success  bool              `json:"success"`
	Identity *PersonalIdentity `json:"identity"`
	Error    string            `json:"error"`
}

// UpdateIdentityRequest message placeholder
type UpdateIdentityRequest struct {
	Id      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

// UpdateIdentityResponse message placeholder
type UpdateIdentityResponse struct {
	Success  bool              `json:"success"`
	Identity *PersonalIdentity `json:"identity"`
	Error    string            `json:"error"`
}

// DeleteIdentityRequest message placeholder
type DeleteIdentityRequest struct {
	Id string `json:"id"`
}

// DeleteIdentityResponse message placeholder
type DeleteIdentityResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ListIdentitiesRequest message placeholder
type ListIdentitiesRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

// ListIdentitiesResponse message placeholder
type ListIdentitiesResponse struct {
	Success   bool              `json:"success"`
	Identities []*PersonalIdentity `json:"identities"`
	Total     int32             `json:"total"`
	Page      int32             `json:"page"`
	PageSize  int32             `json:"page_size"`
	Error     string            `json:"error"`
}

// UpdatePersonalityRequest message placeholder
type UpdatePersonalityRequest struct {
	Id      string                 `json:"id"`
	Updates map[string]float64     `json:"updates"`
}

// UpdatePersonalityResponse message placeholder
type UpdatePersonalityResponse struct {
	Success    bool        `json:"success"`
	Personality *Personality `json:"personality"`
	Error      string      `json:"error"`
}

// SetSpeakingToneRequest message placeholder
type SetSpeakingToneRequest struct {
	Id           string  `json:"id"`
	Tone         string  `json:"tone"`
	ResponseLength string  `json:"response_length"`
	EmojiUsage   float64 `json:"emoji_usage"`
}

// SetSpeakingToneResponse message placeholder
type SetSpeakingToneResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// UpdateValueSystemRequest message placeholder
type UpdateValueSystemRequest struct {
	Id      string             `json:"id"`
	Updates map[string]float64 `json:"updates"`
}

// UpdateValueSystemResponse message placeholder
type UpdateValueSystemResponse struct {
	Success     bool         `json:"success"`
	ValueSystem *ValueSystem `json:"value_system"`
	Error       string       `json:"error"`
}

// AddInterestRequest message placeholder
type AddInterestRequest struct {
	Id       string   `json:"id"`
	Interest *Interest `json:"interest"`
}

// AddInterestResponse message placeholder
type AddInterestResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// RemoveInterestRequest message placeholder
type RemoveInterestRequest struct {
	Id         string `json:"id"`
	InterestId string `json:"interest_id"`
}

// RemoveInterestResponse message placeholder
type RemoveInterestResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// GetInterestsRequest message placeholder
type GetInterestsRequest struct {
	Id       string `json:"id"`
	Category string `json:"category"` // optional filter
}

// GetInterestsResponse message placeholder
type GetInterestsResponse struct {
	Success   bool        `json:"success"`
	Interests []*Interest `json:"interests"`
	Error     string      `json:"error"`
}

// UpdateInterestLevelRequest message placeholder
type UpdateInterestLevelRequest struct {
	Id         string  `json:"id"`
	InterestId string  `json:"interest_id"`
	Level      float64 `json:"level"`
}

// UpdateInterestLevelResponse message placeholder
type UpdateInterestLevelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// GetVoiceProfileRequest message placeholder
type GetVoiceProfileRequest struct {
	Id string `json:"id"`
}

// GetVoiceProfileResponse message placeholder
type GetVoiceProfileResponse struct {
	Success     bool         `json:"success"`
	VoiceProfile *VoiceProfile `json:"voice_profile"`
	Error       string       `json:"error"`
}

// UpdateVoiceProfileRequest message placeholder
type UpdateVoiceProfileRequest struct {
	Id      string        `json:"id"`
	Profile *VoiceProfile `json:"profile"`
}

// UpdateVoiceProfileResponse message placeholder
type UpdateVoiceProfileResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// GetWritingStyleRequest message placeholder
type GetWritingStyleRequest struct {
	Id string `json:"id"`
}

// GetWritingStyleResponse message placeholder
type GetWritingStyleResponse struct {
	Success      bool          `json:"success"`
	WritingStyle *WritingStyle `json:"writing_style"`
	Error        string        `json:"error"`
}

// UpdateWritingStyleRequest message placeholder
type UpdateWritingStyleRequest struct {
	Id    string        `json:"id"`
	Style *WritingStyle `json:"style"`
}

// UpdateWritingStyleResponse message placeholder
type UpdateWritingStyleResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// InferPersonalityRequest message placeholder
type InferPersonalityRequest struct {
	Id           string               `json:"id"`
	Observations []*BehaviorObservation `json:"observations"`
}

// InferPersonalityResponse message placeholder
type InferPersonalityResponse struct {
	Success    bool        `json:"success"`
	Personality *Personality `json:"personality"`
	Error      string      `json:"error"`
}

// GetDecisionContextRequest message placeholder
type GetDecisionContextRequest struct {
	Id string `json:"id"`
}

// GetDecisionContextResponse message placeholder
type GetDecisionContextResponse struct {
	Success         bool             `json:"success"`
	DecisionContext *DecisionContext `json:"decision_context"`
	Error           string           `json:"error"`
}
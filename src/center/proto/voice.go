package proto

// === Voice System Messages ===

// VoiceType enum
type VoiceType string

const (
	VoiceTypeClone     VoiceType = "clone"
	VoiceTypeSynthetic VoiceType = "synthetic"
	VoiceTypePreset    VoiceType = "preset"
)

// VoiceProfile is defined in identity.go

// VoiceInfo message
type VoiceInfo struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Language    string            `json:"language"`
	Gender      string            `json:"gender"`
	Description string            `json:"description"`
	PreviewUrl  string            `json:"preview_url"`
	Attributes  map[string]string `json:"attributes"`
}

// CloneStatus message
type CloneStatus struct {
	CloneId   string `json:"clone_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	VoiceId   string `json:"voice_id"`
	CreatedAt int64  `json:"created_at"`
	ReadyAt   int64  `json:"ready_at"`
}

// ToneAdjustments message
type ToneAdjustments struct {
	PitchDelta   float64 `json:"pitch_delta"`
	SpeedDelta   float64 `json:"speed_delta"`
	VolumeDelta  float64 `json:"volume_delta"`
	EmotionDelta float64 `json:"emotion_delta"`
	Tone         string  `json:"tone"`
}

// === API Requests/Responses ===

// CreateVoiceProfileRequest message
type CreateVoiceProfileRequest struct {
	UserId      string  `json:"user_id"`
	VoiceType   string  `json:"voice_type"`
	PresetVoiceId string  `json:"preset_voice_id"`
	Pitch       float64 `json:"pitch"`
	Speed       float64 `json:"speed"`
	Volume      float64 `json:"volume"`
	Tone        string  `json:"tone"`
	EmotionLevel float64 `json:"emotion_level"`
}

// CreateVoiceProfileResponse message
type CreateVoiceProfileResponse struct {
	Success bool         `json:"success"`
	Profile *VoiceProfile `json:"profile"`
	Error   string       `json:"error"`
}

// GetVoiceProfileRequest is defined in api.go
// GetVoiceProfileResponse is defined in api.go
// UpdateVoiceProfileRequest is defined in api.go
// UpdateVoiceProfileResponse is defined in api.go

// DeleteVoiceProfileRequest message
type DeleteVoiceProfileRequest struct {
	Id string `json:"id"`
}

// DeleteVoiceProfileResponse message
type DeleteVoiceProfileResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ListVoiceProfilesRequest message
type ListVoiceProfilesRequest struct {
	UserId string `json:"user_id"`
}

// ListVoiceProfilesResponse message
type ListVoiceProfilesResponse struct {
	Success  bool           `json:"success"`
	Profiles []*VoiceProfile `json:"profiles"`
	Error    string         `json:"error"`
}

// CloneVoiceRequest is defined in api.go
// CloneVoiceResponse is defined in api.go

// GetCloneStatusRequest message
type GetCloneStatusRequest struct {
	CloneId string `json:"clone_id"`
}

// GetCloneStatusResponse message
type GetCloneStatusResponse struct {
	Success bool         `json:"success"`
	Status  *CloneStatus `json:"status"`
	Error   string       `json:"error"`
}

// DeleteCloneRequest message
type DeleteCloneRequest struct {
	CloneId string `json:"clone_id"`
}

// DeleteCloneResponse message
type DeleteCloneResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SynthesizeRequest message
type SynthesizeRequest struct {
	Text        string        `json:"text"`
	Profile     *VoiceProfile `json:"profile"`
	ProfileId   string        `json:"profile_id"`
}

// SynthesizeResponse message
type SynthesizeResponse struct {
	Success bool   `json:"success"`
	Audio   []byte `json:"audio"`
	Error   string `json:"error"`
}

// GetAvailableVoicesRequest message
type GetAvailableVoicesRequest struct{}

// GetAvailableVoicesResponse message
type GetAvailableVoicesResponse struct {
	Success bool        `json:"success"`
	Voices  []*VoiceInfo `json:"voices"`
	Error   string      `json:"error"`
}

// AdjustToneRequest message
type AdjustToneRequest struct {
	ProfileId    string           `json:"profile_id"`
	Adjustments  *ToneAdjustments `json:"adjustments"`
	Preset       string           `json:"preset"` // energetic/calm/professional/friendly/serious
}

// AdjustToneResponse message
type AdjustToneResponse struct {
	Success bool         `json:"success"`
	Profile *VoiceProfile `json:"profile"`
	Error   string       `json:"error"`
}
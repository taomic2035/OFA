package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// Service - 语音服务
type Service struct {
	mu       sync.RWMutex
	store    VoiceStore
	tts      TTSProvider
	clone    VoiceCloner
}

// VoiceStore - 语音存储接口
type VoiceStore interface {
	SaveVoiceProfile(ctx context.Context, profile *models.VoiceProfile) error
	GetVoiceProfile(ctx context.Context, id string) (*models.VoiceProfile, error)
	DeleteVoiceProfile(ctx context.Context, id string) error
	ListVoiceProfiles(ctx context.Context, userID string) ([]*models.VoiceProfile, error)
}

// TTSProvider - TTS 提供者接口
type TTSProvider interface {
	Synthesize(text string, profile *models.VoiceProfile) ([]byte, error)
	SynthesizeStream(text string, profile *models.VoiceProfile) (<-chan []byte, error)
	GetVoices() ([]VoiceInfo, error)
}

// VoiceCloner - 音色克隆接口
type VoiceCloner interface {
	CloneVoice(samples [][]byte, name string) (string, error) // 返回 voice ID
	GetCloneStatus(cloneID string) (*CloneStatus, error)
	DeleteClone(cloneID string) error
}

// VoiceInfo - 音色信息
type VoiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Language    string            `json:"language"`
	Gender      string            `json:"gender"`
	Description string            `json:"description"`
	PreviewURL  string            `json:"preview_url"`
	Attributes  map[string]string `json:"attributes"`
}

// CloneStatus - 克隆状态
type CloneStatus struct {
	CloneID    string    `json:"clone_id"`
	Status     string    `json:"status"` // processing/ready/failed
	Message    string    `json:"message"`
	VoiceID    string    `json:"voice_id"`
	CreatedAt  time.Time `json:"created_at"`
	ReadyAt    time.Time `json:"ready_at"`
}

// NewService 创建语音服务
func NewService(store VoiceStore, tts TTSProvider, cloner VoiceCloner) *Service {
	return &Service{
		store: store,
		tts:   tts,
		clone: cloner,
	}
}

// === 音色管理 ===

// CreateVoiceProfile 创建语音配置
func (s *Service) CreateVoiceProfile(ctx context.Context, userID string, opts ...VoiceProfileOption) (*models.VoiceProfile, error) {
	profile := models.NewDefaultVoiceProfile()
	profile.ID = generateVoiceID()
	profile.VoiceType = string(models.VoiceTypePreset)

	// 应用选项
	for _, opt := range opts {
		opt(profile)
	}

	// 保存
	if err := s.store.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save voice profile: %w", err)
	}

	return profile, nil
}

// VoiceProfileOption 音色配置选项
type VoiceProfileOption func(*models.VoiceProfile)

// WithVoiceType 设置音色类型
func WithVoiceType(voiceType string) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.VoiceType = voiceType
	}
}

// WithPresetVoice 设置预设音色
func WithPresetVoice(voiceID string) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.VoiceType = string(models.VoiceTypePreset)
		p.PresetVoiceID = voiceID
	}
}

// WithPitch 设置音高
func WithPitch(pitch float64) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.Pitch = clamp(pitch, 0, 2)
	}
}

// WithSpeed 设置语速
func WithSpeed(speed float64) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.Speed = clamp(speed, 0, 2)
	}
}

// WithVolume 设置音量
func WithVolume(volume float64) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.Volume = clamp(volume, 0, 1)
	}
}

// WithTone 设置语调
func WithTone(tone string) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.Tone = tone
	}
}

// WithEmotionLevel 设置情感级别
func WithEmotionLevel(level float64) VoiceProfileOption {
	return func(p *models.VoiceProfile) {
		p.EmotionLevel = clamp(level, 0, 1)
	}
}

// GetVoiceProfile 获取语音配置
func (s *Service) GetVoiceProfile(ctx context.Context, id string) (*models.VoiceProfile, error) {
	return s.store.GetVoiceProfile(ctx, id)
}

// UpdateVoiceProfile 更新语音配置
func (s *Service) UpdateVoiceProfile(ctx context.Context, id string, updates map[string]interface{}) (*models.VoiceProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, err := s.store.GetVoiceProfile(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("voice profile not found: %w", err)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "voice_type":
			if v, ok := value.(string); ok {
				profile.VoiceType = v
			}
		case "preset_voice_id":
			if v, ok := value.(string); ok {
				profile.PresetVoiceID = v
			}
		case "clone_reference_id":
			if v, ok := value.(string); ok {
				profile.CloneReferenceID = v
			}
		case "pitch":
			if v, ok := value.(float64); ok {
				profile.Pitch = clamp(v, 0, 2)
			}
		case "speed":
			if v, ok := value.(float64); ok {
				profile.Speed = clamp(v, 0, 2)
			}
		case "volume":
			if v, ok := value.(float64); ok {
				profile.Volume = clamp(v, 0, 1)
			}
		case "tone":
			if v, ok := value.(string); ok {
				profile.Tone = v
			}
		case "accent":
			if v, ok := value.(string); ok {
				profile.Accent = v
			}
		case "emotion_level":
			if v, ok := value.(float64); ok {
				profile.EmotionLevel = clamp(v, 0, 1)
			}
		case "pause_pattern":
			if v, ok := value.(string); ok {
				profile.PausePattern = v
			}
		case "emphasis_style":
			if v, ok := value.(string); ok {
				profile.EmphasisStyle = v
			}
		}
	}

	profile.UpdatedAt = time.Now()

	if err := s.store.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save voice profile: %w", err)
	}

	return profile, nil
}

// DeleteVoiceProfile 删除语音配置
func (s *Service) DeleteVoiceProfile(ctx context.Context, id string) error {
	return s.store.DeleteVoiceProfile(ctx, id)
}

// ListVoiceProfiles 列出语音配置
func (s *Service) ListVoiceProfiles(ctx context.Context, userID string) ([]*models.VoiceProfile, error) {
	return s.store.ListVoiceProfiles(ctx, userID)
}

// === 音色克隆 ===

// CloneVoice 克隆音色
func (s *Service) CloneVoice(ctx context.Context, samples [][]byte, name string) (string, error) {
	if s.clone == nil {
		return "", fmt.Errorf("voice cloner not available")
	}

	cloneID, err := s.clone.CloneVoice(samples, name)
	if err != nil {
		return "", fmt.Errorf("failed to clone voice: %w", err)
	}

	return cloneID, nil
}

// GetCloneStatus 获取克隆状态
func (s *Service) GetCloneStatus(ctx context.Context, cloneID string) (*CloneStatus, error) {
	if s.clone == nil {
		return nil, fmt.Errorf("voice cloner not available")
	}

	return s.clone.GetCloneStatus(cloneID)
}

// DeleteClone 删除克隆
func (s *Service) DeleteClone(ctx context.Context, cloneID string) error {
	if s.clone == nil {
		return fmt.Errorf("voice cloner not available")
	}

	return s.clone.DeleteClone(cloneID)
}

// CreateClonedProfile 创建克隆音色配置
func (s *Service) CreateClonedProfile(ctx context.Context, cloneID string, opts ...VoiceProfileOption) (*models.VoiceProfile, error) {
	// 检查克隆状态
	status, err := s.GetCloneStatus(ctx, cloneID)
	if err != nil {
		return nil, err
	}

	if status.Status != "ready" {
		return nil, fmt.Errorf("clone not ready: %s", status.Status)
	}

	profile := models.NewDefaultVoiceProfile()
	profile.ID = generateVoiceID()
	profile.VoiceType = string(models.VoiceTypeClone)
	profile.CloneReferenceID = cloneID

	for _, opt := range opts {
		opt(profile)
	}

	if err := s.store.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save voice profile: %w", err)
	}

	return profile, nil
}

// === 语音合成 ===

// Synthesize 合成语音
func (s *Service) Synthesize(ctx context.Context, text string, profile *models.VoiceProfile) ([]byte, error) {
	if s.tts == nil {
		return nil, fmt.Errorf("TTS provider not available")
	}

	return s.tts.Synthesize(text, profile)
}

// SynthesizeStream 流式合成语音
func (s *Service) SynthesizeStream(ctx context.Context, text string, profile *models.VoiceProfile) (<-chan []byte, error) {
	if s.tts == nil {
		return nil, fmt.Errorf("TTS provider not available")
	}

	return s.tts.SynthesizeStream(text, profile)
}

// GetAvailableVoices 获取可用音色列表
func (s *Service) GetAvailableVoices(ctx context.Context) ([]VoiceInfo, error) {
	if s.tts == nil {
		return nil, fmt.Errorf("TTS provider not available")
	}

	return s.tts.GetVoices()
}

// === 语调调整 ===

// AdjustTone 调整语调参数
func (s *Service) AdjustTone(ctx context.Context, profileID string, adjustments ToneAdjustments) (*models.VoiceProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, err := s.store.GetVoiceProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("voice profile not found: %w", err)
	}

	// 应用调整
	if adjustments.PitchDelta != 0 {
		profile.Pitch = clamp(profile.Pitch+adjustments.PitchDelta, 0, 2)
	}
	if adjustments.SpeedDelta != 0 {
		profile.Speed = clamp(profile.Speed+adjustments.SpeedDelta, 0, 2)
	}
	if adjustments.VolumeDelta != 0 {
		profile.Volume = clamp(profile.Volume+adjustments.VolumeDelta, 0, 1)
	}
	if adjustments.EmotionDelta != 0 {
		profile.EmotionLevel = clamp(profile.EmotionLevel+adjustments.EmotionDelta, 0, 1)
	}
	if adjustments.Tone != "" {
		profile.Tone = adjustments.Tone
	}

	profile.UpdatedAt = time.Now()

	if err := s.store.SaveVoiceProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save voice profile: %w", err)
	}

	return profile, nil
}

// ToneAdjustments 语调调整参数
type ToneAdjustments struct {
	PitchDelta   float64 `json:"pitch_delta"`
	SpeedDelta   float64 `json:"speed_delta"`
	VolumeDelta  float64 `json:"volume_delta"`
	EmotionDelta float64 `json:"emotion_delta"`
	Tone         string  `json:"tone"`
}

// SetPresetTone 设置预设语调
func (s *Service) SetPresetTone(ctx context.Context, profileID string, preset string) (*models.VoiceProfile, error) {
	adjustments := getPresetAdjustments(preset)
	return s.AdjustTone(ctx, profileID, adjustments)
}

// getPresetAdjustments 获取预设语调调整
func getPresetAdjustments(preset string) ToneAdjustments {
	switch preset {
	case "energetic":
		return ToneAdjustments{
			PitchDelta:   0.1,
			SpeedDelta:   0.1,
			EmotionDelta: 0.2,
			Tone:         "energetic",
		}
	case "calm":
		return ToneAdjustments{
			PitchDelta:   -0.1,
			SpeedDelta:   -0.1,
			EmotionDelta: -0.1,
			Tone:         "calm",
		}
	case "professional":
		return ToneAdjustments{
			SpeedDelta: -0.1,
			Tone:       "professional",
		}
	case "friendly":
		return ToneAdjustments{
			PitchDelta:   0.05,
			EmotionDelta: 0.1,
			Tone:         "friendly",
		}
	case "serious":
		return ToneAdjustments{
			SpeedDelta:   -0.05,
			EmotionDelta: -0.2,
			Tone:         "serious",
		}
	default:
		return ToneAdjustments{}
	}
}

// === 辅助函数 ===

func generateVoiceID() string {
	return fmt.Sprintf("voice_%d", time.Now().UnixNano())
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
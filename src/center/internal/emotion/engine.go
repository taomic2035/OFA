package emotion

import (
	"context"
	"sync"
	"time"

	"github.com/taomic2035/OFA/src/center/internal/models"
)

// EmotionEngine 情绪引擎 (v4.0.0)
// 核心功能: 情绪触发、衰减、传播、记忆
type EmotionEngine struct {
	mu sync.RWMutex

	// 情绪存储
	emotions     map[string]*models.Emotion            // identityID -> Emotion
	emotionProfiles map[string]*models.EmotionProfile  // identityID -> Profile
	emotionStates   map[string][]models.EmotionState   // identityID -> States历史

	// 欲望存储
	desires      map[string]*models.Desire
	desireProfiles map[string]*models.DesireProfile

	// 配置
	config EmotionConfig

	// 监听器
	listeners []EmotionListener

	// 时间追踪
	lastUpdate map[string]time.Time
}

// EmotionConfig 情绪引擎配置
type EmotionConfig struct {
	// 衰减配置
	DefaultDecayRate      float64 `json:"default_decay_rate"`       // 默认衰减率 (per minute)
	PositiveDecayRate     float64 `json:"positive_decay_rate"`     // 正面情绪衰减率
	NegativeDecayRate     float64 `json:"negative_decay_rate"`     // 负面情绪衰减率

	// 触发配置
	MaxTriggerIntensity   float64 `json:"max_trigger_intensity"`   // 最大触发强度
	TriggerCooldown       int     `json:"trigger_cooldown"`        // 触发冷却时间(秒)

	// 历史配置
	MaxHistoryStates      int     `json:"max_history_states"`      // 最大历史状态数
	MaxTriggerHistory     int     `json:"max_trigger_history"`     // 最大触发历史数

	// 传播配置
	EmotionInfluenceRadius float64 `json:"emotion_influence_radius"` // 情绪影响范围
}

// EmotionListener 情绪事件监听器
type EmotionListener interface {
	OnEmotionTriggered(identityID string, emotion *models.Emotion, trigger models.EmotionTrigger)
	OnEmotionDecayed(identityID string, emotion *models.Emotion)
	OnMoodChanged(identityID string, oldMood, newMood string)
	OnDesireChanged(identityID string, desire *models.Desire)
}

// NewEmotionEngine 创建情绪引擎
func NewEmotionEngine(config EmotionConfig) *EmotionEngine {
	// 默认配置
	if config.DefaultDecayRate == 0 {
		config.DefaultDecayRate = 0.02
	}
	if config.PositiveDecayRate == 0 {
		config.PositiveDecayRate = 0.03 // 正面情绪衰减快
	}
	if config.NegativeDecayRate == 0 {
		config.NegativeDecayRate = 0.01 // 负面情绪衰减慢
	}
	if config.MaxTriggerIntensity == 0 {
		config.MaxTriggerIntensity = 0.8
	}
	if config.TriggerCooldown == 0 {
		config.TriggerCooldown = 30
	}
	if config.MaxHistoryStates == 0 {
		config.MaxHistoryStates = 100
	}
	if config.MaxTriggerHistory == 0 {
		config.MaxTriggerHistory = 50
	}
	if config.EmotionInfluenceRadius == 0 {
		config.EmotionInfluenceRadius = 0.5
	}

	return &EmotionEngine{
		emotions:        make(map[string]*models.Emotion),
		emotionProfiles: make(map[string]*models.EmotionProfile),
		emotionStates:   make(map[string][]models.EmotionState),
		desires:         make(map[string]*models.Desire),
		desireProfiles:  make(map[string]*models.DesireProfile),
		config:          config,
		listeners:       []EmotionListener{},
		lastUpdate:      make(map[string]time.Time),
	}
}

// === 情绪管理 ===

// GetEmotion 获取情绪状态
func (e *EmotionEngine) GetEmotion(identityID string) *models.Emotion {
	e.mu.RLock()
	defer e.mu.RUnlock()

	emotion, exists := e.emotions[identityID]
	if !exists {
		return nil
	}
	return emotion
}

// GetOrCreateEmotion 获取或创建情绪
func (e *EmotionEngine) GetOrCreateEmotion(identityID string) *models.Emotion {
	e.mu.Lock()
	defer e.mu.Unlock()

	emotion, exists := e.emotions[identityID]
	if !exists {
		emotion = models.NewEmotion()
		e.emotions[identityID] = emotion
		e.emotionProfiles[identityID] = models.NewEmotionProfile(identityID)
		e.lastUpdate[identityID] = time.Now()
	}
	return emotion
}

// GetEmotionProfile 获取情绪画像
func (e *EmotionEngine) GetEmotionProfile(identityID string) *models.EmotionProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.emotionProfiles[identityID]
}

// UpdateEmotionProfile 更新情绪画像
func (e *EmotionEngine) UpdateEmotionProfile(identityID string, profile *models.EmotionProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.UpdatedAt = time.Now()
	e.emotionProfiles[identityID] = profile
	return nil
}

// === 情绪触发 ===

// TriggerEmotion 触发情绪
func (e *EmotionEngine) TriggerEmotion(ctx context.Context, identityID string, trigger models.EmotionTrigger) (*models.Emotion, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 获取或创建情绪
	emotion, exists := e.emotions[identityID]
	if !exists {
		emotion = models.NewEmotion()
		e.emotions[identityID] = emotion
	}

	// 获取情绪画像（用于调整触发敏感度）
	profile := e.emotionProfiles[identityID]

	// 调整触发强度
	triggerIntensity := trigger.Intensity
	if profile != nil {
		sensitivity, ok := profile.TriggerSensitivity[trigger.TriggerType]
		if ok {
			triggerIntensity *= sensitivity
		}
	}

	// 限制最大强度
	if triggerIntensity > e.config.MaxTriggerIntensity {
		triggerIntensity = e.config.MaxTriggerIntensity
	}

	trigger.Intensity = triggerIntensity

	// 记录旧心境
	oldMood := emotion.CurrentMood

	// 触发情绪
	emotion.TriggerEmotion(trigger)

	// 更新心境
	emotion.UpdateMood()

	// 记录状态
	e.recordState(identityID, emotion, trigger)

	// 更新时间
	e.lastUpdate[identityID] = time.Now()

	// 通知监听器
	newMood := emotion.CurrentMood
	for _, listener := range e.listeners {
		listener.OnEmotionTriggered(identityID, emotion, trigger)
		if oldMood != newMood {
			listener.OnMoodChanged(identityID, oldMood, newMood)
		}
	}

	return emotion, nil
}

// TriggerEmotionByEvent 根据事件触发情绪
func (e *EmotionEngine) TriggerEmotionByEvent(ctx context.Context, identityID string, eventType string, eventDesc string, contextData map[string]interface{}) (*models.Emotion, error) {
	// 根据事件类型确定情绪映射
	trigger := e.mapEventToTrigger(eventType, eventDesc, contextData)
	return e.TriggerEmotion(ctx, identityID, trigger)
}

// mapEventToTrigger 事件到情绪触发映射
func (e *EmotionEngine) mapEventToTrigger(eventType string, eventDesc string, contextData map[string]interface{}) models.EmotionTrigger {
	trigger := models.EmotionTrigger{
		TriggerID:   generateTriggerID(),
		TriggerType: "event",
		TriggerDesc: eventDesc,
		Context:     contextData,
		Timestamp:   time.Now(),
	}

	// 事件类型到情绪映射
	switch eventType {
	// 正面事件
	case "achievement", "success", "reward":
		trigger.EmotionType = "joy"
		trigger.Intensity = 0.5
		trigger.Duration = 60
	case "compliment", "praise", "recognition":
		trigger.EmotionType = "joy"
		trigger.Intensity = 0.3
		trigger.Duration = 30
	case "connection", "reunion", "friendship":
		trigger.EmotionType = "love"
		trigger.Intensity = 0.4
		trigger.Duration = 45
	case "gift", "surprise_positive":
		trigger.EmotionType = "joy"
		trigger.Intensity = 0.4
		trigger.Duration = 40

	// 负面事件
	case "failure", "mistake", "error":
		trigger.EmotionType = "sadness"
		trigger.Intensity = 0.4
		trigger.Duration = 45
	case "criticism", "rejection":
		trigger.EmotionType = "sadness"
		trigger.Intensity = 0.3
		trigger.Duration = 30
	case "loss", "death", "separation":
		trigger.EmotionType = "sadness"
		trigger.Intensity = 0.7
		trigger.Duration = 120
	case "threat", "danger":
		trigger.EmotionType = "fear"
		trigger.Intensity = 0.6
		trigger.Duration = 60
	case "injustice", "betrayal":
		trigger.EmotionType = "anger"
		trigger.Intensity = 0.5
		trigger.Duration = 50
	case "conflict", "argument":
		trigger.EmotionType = "anger"
		trigger.Intensity = 0.4
		trigger.Duration = 40
	case "disappointment":
		trigger.EmotionType = "sadness"
		trigger.Intensity = 0.3
		trigger.Duration = 30

	// 中性/复杂事件
	case "challenge":
		trigger.EmotionType = "desire"
		trigger.Intensity = 0.4
		trigger.Duration = 30
	case "novelty", "discovery":
		trigger.EmotionType = "joy"
		trigger.Intensity = 0.3
		trigger.Duration = 20
	case "uncertainty":
		trigger.EmotionType = "fear"
		trigger.Intensity = 0.2
		trigger.Duration = 25

	default:
		trigger.EmotionType = "joy"
		trigger.Intensity = 0.2
		trigger.Duration = 15
	}

	return trigger
}

// === 情绪衰减 ===

// DecayEmotion 衰减情绪
func (e *EmotionEngine) DecayEmotion(identityID string, minutes int) (*models.Emotion, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	emotion, exists := e.emotions[identityID]
	if !exists {
		return nil, nil
	}

	// 根据情绪效价选择衰减率
	rate := e.config.DefaultDecayRate
	if emotion.Valence > 0 {
		rate = e.config.PositiveDecayRate
	} else if emotion.Valence < 0 {
		rate = e.config.NegativeDecayRate
	}

	// 应用衰减
	emotion.Decay(rate, minutes)

	// 更新心境
	oldMood := emotion.CurrentMood
	emotion.UpdateMood()
	newMood := emotion.CurrentMood

	// 更新时间
	e.lastUpdate[identityID] = time.Now()

	// 通知监听器
	for _, listener := range e.listeners {
		listener.OnEmotionDecayed(identityID, emotion)
		if oldMood != newMood {
			listener.OnMoodChanged(identityID, oldMood, newMood)
		}
	}

	return emotion, nil
}

// DecayAllEmotions 衰减所有情绪
func (e *EmotionEngine) DecayAllEmotions(minutes int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	for identityID, emotion := range e.emotions {
		lastUpdate := e.lastUpdate[identityID]
		actualMinutes := int(now.Sub(lastUpdate).Minutes())
		if actualMinutes > 0 {
			// 根据情绪效价选择衰减率
			rate := e.config.DefaultDecayRate
			if emotion.Valence > 0 {
				rate = e.config.PositiveDecayRate
			} else if emotion.Valence < 0 {
				rate = e.config.NegativeDecayRate
			}

			emotion.Decay(rate, actualMinutes)
			emotion.UpdateMood()
			e.lastUpdate[identityID] = now
		}
	}
}

// === 情绪传播 ===

// PropagateEmotion 情绪传播（影响决策）
func (e *EmotionEngine) PropagateEmotion(identityID string) map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	emotion, exists := e.emotions[identityID]
	if !exists {
		return nil
	}

	// 情绪对决策的影响因子
	influence := make(map[string]float64)

	// 风险偏好 (高兴时风险偏好增加，恐惧时降低)
	influence["risk_tolerance"] = 0.5 + emotion.Joy * 0.3 - emotion.Fear * 0.3

	// 时间折扣 (悲伤时时间折扣增加，急于寻求慰藉)
	influence["time_discount"] = 0.1 + emotion.Sadness * 0.2 + emotion.Desire * 0.1

	// 社交倾向 (爱和喜悦增加，厌恶和恐惧降低)
	influence["social_tendency"] = emotion.Love * 0.3 + emotion.Joy * 0.2 - emotion.Disgust * 0.2 - emotion.Fear * 0.1

	// 决策速度 (愤怒时加快，恐惧时减慢)
	influence["decision_speed"] = 0.5 + emotion.Anger * 0.3 - emotion.Fear * 0.2

	// 信任度 (爱增加，恐惧和厌恶降低)
	influence["trust_level"] = 0.5 + emotion.Love * 0.2 - emotion.Fear * 0.2 - emotion.Disgust * 0.2

	// 创造性 (喜悦和自我实现欲望增加)
	influence["creativity"] = 0.3 + emotion.Joy * 0.3 + emotion.Desire * 0.2

	// 限制范围 [0, 1]
	for key, value := range influence {
		if value < 0 {
			influence[key] = 0
		}
		if value > 1 {
			influence[key] = 1
		}
	}

	return influence
}

// GetEmotionContext 获取情绪决策上下文
func (e *EmotionEngine) GetEmotionContext(identityID string) *EmotionDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	emotion, exists := e.emotions[identityID]
	if !exists {
		return nil
	}

	desire, desireExists := e.desires[identityID]

	context := &EmotionDecisionContext{
		IdentityID:      identityID,
		CurrentEmotion:  emotion,
		DominantEmotion: emotion.GetDominantEmotion(),
		Mood:            emotion.CurrentMood,
		MoodTrend:       emotion.MoodTrend,
		Valence:         emotion.Valence,
		Arousal:         emotion.Arousal,
		Intensity:       emotion.Intensity,
		InfluenceFactors: e.PropagateEmotion(identityID),
	}

	if desireExists {
		context.CurrentDesire = desire
		context.PrimaryDesire = desire.PrimaryDesire
		context.DesireStrength = desire.DesireStrength
		context.DesireTarget = desire.DesireTarget
		context.SatisfactionLevel = desire.CalculateOverallSatisfaction()
		context.FrustrationLevel = desire.CalculateFrustrationLevel()
	}

	return context
}

// EmotionDecisionContext 情绪决策上下文
type EmotionDecisionContext struct {
	IdentityID      string                 `json:"identity_id"`
	CurrentEmotion  *models.Emotion        `json:"current_emotion"`
	DominantEmotion string                 `json:"dominant_emotion"`
	Mood            string                 `json:"mood"`
	MoodTrend       string                 `json:"mood_trend"`
	Valence         float64                `json:"valence"`
	Arousal         float64                `json:"arousal"`
	Intensity       float64                `json:"intensity"`
	InfluenceFactors map[string]float64    `json:"influence_factors"`

	// 欲望相关
	CurrentDesire     *models.Desire `json:"current_desire,omitempty"`
	PrimaryDesire     string         `json:"primary_desire"`
	DesireStrength    float64        `json:"desire_strength"`
	DesireTarget      string         `json:"desire_target"`
	SatisfactionLevel float64        `json:"satisfaction_level"`
	FrustrationLevel  float64        `json:"frustration_level"`
}

// === 欲望管理 ===

// GetDesire 获取欲望状态
func (e *EmotionEngine) GetDesire(identityID string) *models.Desire {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.desires[identityID]
}

// GetOrCreateDesire 获取或创建欲望
func (e *EmotionEngine) GetOrCreateDesire(identityID string) *models.Desire {
	e.mu.Lock()
	defer e.mu.Unlock()

	desire, exists := e.desires[identityID]
	if !exists {
		desire = models.NewDesire()
		e.desires[identityID] = desire
		e.desireProfiles[identityID] = models.NewDesireProfile(identityID)
	}
	return desire
}

// TriggerDesire 触发欲望
func (e *EmotionEngine) TriggerDesire(identityID string, desireType string, strength float64, target string) (*models.Desire, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	desire := e.GetOrCreateDesire(identityID)
	desire.TriggerDesire(desireType, strength, target)

	e.lastUpdate[identityID] = time.Now()

	// 通知监听器
	for _, listener := range e.listeners {
		listener.OnDesireChanged(identityID, desire)
	}

	return desire, nil
}

// SatisfyDesire 满足欲望
func (e *EmotionEngine) SatisfyDesire(identityID string, desireType string, amount float64) (*models.Desire, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	desire := e.GetOrCreateDesire(identityID)
	desire.SatisfyDesire(desireType, amount)

	// 满足欲望可能触发正面情绪
	emotion := e.emotions[identityID]
	if emotion != nil {
		trigger := models.EmotionTrigger{
			TriggerID:   generateTriggerID(),
			TriggerType: "satisfaction",
			TriggerDesc: "欲望满足: " + desireType,
			EmotionType: "joy",
			Intensity:   amount * 0.3,
			Duration:    20,
			Timestamp:   time.Now(),
		}
		emotion.TriggerEmotion(trigger)
		emotion.UpdateMood()
	}

	e.lastUpdate[identityID] = time.Now()

	// 通知监听器
	for _, listener := range e.listeners {
		listener.OnDesireChanged(identityID, desire)
	}

	return desire, nil
}

// DecayDesire 衰减欲望满足度
func (e *EmotionEngine) DecayDesire(identityID string, minutes int) (*models.Desire, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	desire, exists := e.desires[identityID]
	if !exists {
		return nil, nil
	}

	// 需求满足度随时间降低
	desire.Decay(0.02, minutes)

	e.lastUpdate[identityID] = time.Now()

	// 通知监听器
	for _, listener := range e.listeners {
		listener.OnDesireChanged(identityID, desire)
	}

	return desire, nil
}

// === 情绪记忆 ===

// GetEmotionHistory 获取情绪历史
func (e *EmotionEngine) GetEmotionHistory(identityID string, limit int) []models.EmotionState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	states := e.emotionStates[identityID]
	if limit > 0 && len(states) > limit {
		return states[len(states)-limit:]
	}
	return states
}

// GetEmotionPatterns 获取情绪模式
func (e *EmotionEngine) GetEmotionPatterns(identityID string) []models.EmotionPattern {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile := e.emotionProfiles[identityID]
	if profile == nil {
		return nil
	}
	return profile.EmotionPatterns
}

// recordState 记录情绪状态
func (e *EmotionEngine) recordState(identityID string, emotion *models.Emotion, trigger models.EmotionTrigger) {
	state := models.EmotionState{
		StateID:    generateStateID(),
		IdentityID: identityID,
		Emotion:    *emotion,
		Triggers:   []string{trigger.TriggerID},
		Stability:  calculateStability(emotion),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 添加到历史
	states := e.emotionStates[identityID]
	states = append(states, state)

	// 限制历史数量
	if len(states) > e.config.MaxHistoryStates {
		states = states[len(states)-e.config.MaxHistoryStates:]
	}

	e.emotionStates[identityID] = states
}

// calculateStability 计算情绪稳定性
func calculateStability(emotion *models.Emotion) float64 {
	// 稳定性 = 1 - 波动程度
	// 波动程度基于情绪强度和唤醒度
	variance := (emotion.Intensity + emotion.Arousal) / 2
	stability := 1 - variance

	if stability < 0 {
		stability = 0
	}
	return stability
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *EmotionEngine) AddListener(listener EmotionListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *EmotionEngine) RemoveListener(listener EmotionListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.listeners {
		if l == listener {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			break
		}
	}
}

// === 辅助函数 ===

func generateTriggerID() string {
	return "trigger_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

func generateStateID() string {
	return "state_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
package models

import "time"

// Emotion 情绪模型 (v4.0.0)
// 七情: 喜、怒、哀、惧、爱、恶、欲
type Emotion struct {
	// === 七情 (0-1 强度) ===
	Joy     float64 `json:"joy"`     // 喜 - 快乐、愉悦
	Anger   float64 `json:"anger"`   // 怒 - 愤怒、不满
	Sadness float64 `json:"sadness"` // 哀 - 悲伤、忧郁
	Fear    float64 `json:"fear"`    // 惧 - 恐惧、担忧
	Love    float64 `json:"love"`    // 爱 - 爱意、喜欢
	Disgust float64 `json:"disgust"` // 恶 - 厌恶、反感
	Desire  float64 `json:"desire"`  // 欲 - 欲望、渴求

	// === 情绪元数据 ===
	Intensity float64 `json:"intensity"` // 情绪强度 (0-1)
	Valence   float64 `json:"valence"`   // 情感效价 (-1负到1正)
	Arousal   float64 `json:"arousal"`   // 唤醒度 (0-1)
	Dominance float64 `json:"dominance"` // 支配感 (0-1)

	// === 动态属性 ===
	CurrentMood string    `json:"current_mood"` // 当前心境 (happy/sad/angry/neutral/anxious/excited/content)
	MoodTrend   string    `json:"mood_trend"`   // 情绪趋势 (improving/stable/declining)
	LastTrigger string    `json:"last_trigger"` // 最后触发原因
	Duration    int       `json:"duration"`     // 持续时间 (分钟)
	Timestamp   time.Time `json:"timestamp"`    // 时间戳

	// === 情绪历史 ===
	RecentTriggers []EmotionTrigger `json:"recent_triggers,omitempty"` // 近期触发事件
}

// EmotionTrigger 情绪触发事件
type EmotionTrigger struct {
	TriggerID   string                 `json:"trigger_id"`
	TriggerType string                 `json:"trigger_type"` // event/interaction/thought/environment/body
	TriggerDesc string                 `json:"trigger_desc"` // 触发描述
	EmotionType string                 `json:"emotion_type"` // joy/anger/sadness/fear/love/disgust/desire
	Intensity   float64                `json:"intensity"`    // 触发强度
	Duration    int                    `json:"duration"`     // 持续时间(分钟)
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// EmotionState 情绪状态快照
type EmotionState struct {
	StateID      string    `json:"state_id"`
	IdentityID   string    `json:"identity_id"`
	Emotion      Emotion   `json:"emotion"`
	Triggers     []string  `json:"triggers"` // 触发ID列表
	Stability    float64   `json:"stability"` // 情绪稳定性 (0-1)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EmotionProfile 情绪特征画像
type EmotionProfile struct {
	IdentityID string `json:"identity_id"`

	// === 基线情绪 (常态) ===
	BaselineJoy     float64 `json:"baseline_joy"`
	BaselineAnger   float64 `json:"baseline_anger"`
	BaselineSadness float64 `json:"baseline_sadness"`
	BaselineFear    float64 `json:"baseline_fear"`
	BaselineLove    float64 `json:"baseline_love"`
	BaselineDisgust float64 `json:"baseline_disgust"`
	BaselineDesire  float64 `json:"baseline_desire"`

	// === 情绪特质 ===
	EmotionalRange    float64 `json:"emotional_range"`    // 情绪波动范围 (0-1)
	RecoveryRate      float64 `json:"recovery_rate"`      // 情绪恢复速率 (0-1)
	TriggerSensitivity map[string]float64 `json:"trigger_sensitivity,omitempty"` // 触发敏感度

	// === 情绪偏好 ===
	PreferredEmotions []string `json:"preferred_emotions,omitempty"` // 喜欢的情绪状态
	AvoidedEmotions   []string `json:"avoided_emotions,omitempty"`   // 避免的情绪状态

	// === 情绪习惯 ===
	EmotionPatterns []EmotionPattern `json:"emotion_patterns,omitempty"` // 情绪模式

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmotionPattern 情绪模式
type EmotionPattern struct {
	PatternID   string   `json:"pattern_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TriggerConditions []string `json:"trigger_conditions,omitempty"`
	ExpectedEmotions  map[string]float64 `json:"expected_emotions,omitempty"`
	Frequency         float64 `json:"frequency"` // 发生频率
	LastOccurrence    time.Time `json:"last_occurrence"`
}

// NewEmotion 创建默认情绪
func NewEmotion() *Emotion {
	return &Emotion{
		Joy:           0.5,
		Anger:         0.1,
		Sadness:       0.1,
		Fear:          0.1,
		Love:          0.3,
		Disgust:       0.1,
		Desire:        0.3,
		Intensity:     0.3,
		Valence:       0.0,
		Arousal:       0.3,
		Dominance:     0.5,
		CurrentMood:   "neutral",
		MoodTrend:     "stable",
		LastTrigger:   "",
		Duration:      0,
		Timestamp:     time.Now(),
		RecentTriggers: []EmotionTrigger{},
	}
}

// NewEmotionProfile 创建默认情绪画像
func NewEmotionProfile(identityID string) *EmotionProfile {
	return &EmotionProfile{
		IdentityID:        identityID,
		BaselineJoy:       0.5,
		BaselineAnger:     0.1,
		BaselineSadness:   0.1,
		BaselineFear:      0.1,
		BaselineLove:      0.3,
		BaselineDisgust:   0.1,
		BaselineDesire:    0.3,
		EmotionalRange:    0.3,
		RecoveryRate:      0.5,
		TriggerSensitivity: map[string]float64{
			"positive_event": 0.5,
			"negative_event": 0.7,
			"social":         0.5,
			"achievement":    0.6,
		},
		PreferredEmotions: []string{"joy", "love", "content"},
		AvoidedEmotions:   []string{"anger", "fear", "sadness"},
		EmotionPatterns:   []EmotionPattern{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// GetDominantEmotion 获取主导情绪
func (e *Emotion) GetDominantEmotion() string {
	emotions := map[string]float64{
		"joy":     e.Joy,
		"anger":   e.Anger,
		"sadness": e.Sadness,
		"fear":    e.Fear,
		"love":    e.Love,
		"disgust": e.Disgust,
		"desire":  e.Desire,
	}

	maxEmotion := "joy"
	maxValue := e.Joy
	for name, value := range emotions {
		if value > maxValue {
			maxValue = value
			maxEmotion = name
		}
	}
	return maxEmotion
}

// CalculateValence 计算情感效价
func (e *Emotion) CalculateValence() float64 {
	// 正面情绪
	positive := e.Joy + e.Love + e.Desire*0.5
	// 负面情绪
	negative := e.Anger + e.Sadness + e.Fear + e.Disgust

	// 计算效价 (-1 到 1)
	valence := (positive - negative) / 2
	if valence > 1 {
		valence = 1
	}
	if valence < -1 {
		valence = -1
	}
	return valence
}

// CalculateArousal 计算唤醒度
func (e *Emotion) CalculateArousal() float64 {
	// 高唤醒情绪
	highArousal := e.Anger + e.Fear + e.Joy + e.Desire
	// 低唤醒情绪
	lowArousal := e.Sadness + e.Love + e.Disgust

	arousal := (highArousal - lowArousal*0.3) / 1.7
	if arousal > 1 {
		arousal = 1
	}
	if arousal < 0 {
		arousal = 0
	}
	return arousal
}

// Normalize 归一化情绪值
func (e *Emotion) Normalize() {
	normalizeValue := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	e.Joy = normalizeValue(e.Joy)
	e.Anger = normalizeValue(e.Anger)
	e.Sadness = normalizeValue(e.Sadness)
	e.Fear = normalizeValue(e.Fear)
	e.Love = normalizeValue(e.Love)
	e.Disgust = normalizeValue(e.Disgust)
	e.Desire = normalizeValue(e.Desire)
	e.Intensity = normalizeValue(e.Intensity)
	e.Valence = normalizeValue(e.Valence)
	e.Arousal = normalizeValue(e.Arousal)
	e.Dominance = normalizeValue(e.Dominance)
}

// Decay 情绪衰减
func (e *Emotion) Decay(rate float64, minutes int) {
	// 衰减因子
	decayFactor := 1 - rate * float64(minutes) / 60.0
	if decayFactor < 0.1 {
		decayFactor = 0.1
	}

	// 应用衰减
	e.Joy *= decayFactor
	e.Anger *= decayFactor
	e.Sadness *= decayFactor
	e.Fear *= decayFactor
	e.Love *= decayFactor
	e.Disgust *= decayFactor
	e.Desire *= decayFactor
	e.Intensity *= decayFactor

	// 归一化
	e.Normalize()

	// 更新时间
	e.Duration += minutes
	e.Timestamp = time.Now()
}

// TriggerEmotion 触发情绪
func (e *Emotion) TriggerEmotion(trigger EmotionTrigger) {
	// 增加对应情绪
	switch trigger.EmotionType {
	case "joy":
		e.Joy += trigger.Intensity
	case "anger":
		e.Anger += trigger.Intensity
	case "sadness":
		e.Sadness += trigger.Intensity
	case "fear":
		e.Fear += trigger.Intensity
	case "love":
		e.Love += trigger.Intensity
	case "disgust":
		e.Disgust += trigger.Intensity
	case "desire":
		e.Desire += trigger.Intensity
	}

	// 更新元数据
	e.Intensity = max(e.Intensity, trigger.Intensity)
	e.LastTrigger = trigger.TriggerDesc
	e.Duration = trigger.Duration
	e.Timestamp = time.Now()

	// 记录触发
	e.RecentTriggers = append(e.RecentTriggers, trigger)
	if len(e.RecentTriggers) > 10 {
		e.RecentTriggers = e.RecentTriggers[len(e.RecentTriggers)-10:]
	}

	// 归一化
	e.Normalize()

	// 更新效价和唤醒度
	e.Valence = e.CalculateValence()
	e.Arousal = e.CalculateArousal()
}

// UpdateMood 更新心境状态
func (e *Emotion) UpdateMood() {
	dominant := e.GetDominantEmotion()

	// 根据主导情绪确定心境
	switch dominant {
	case "joy":
		if e.Intensity > 0.7 {
			e.CurrentMood = "excited"
		} else {
			e.CurrentMood = "happy"
		}
	case "sadness":
		if e.Intensity > 0.7 {
			e.CurrentMood = "depressed"
		} else {
			e.CurrentMood = "sad"
		}
	case "anger":
		if e.Intensity > 0.7 {
			e.CurrentMood = "furious"
		} else {
			e.CurrentMood = "angry"
		}
	case "fear":
		if e.Intensity > 0.7 {
			e.CurrentMood = "panicked"
		} else {
			e.CurrentMood = "anxious"
		}
	case "love":
		e.CurrentMood = "content"
	case "disgust":
		e.CurrentMood = "annoyed"
	case "desire":
		e.CurrentMood = "motivated"
	default:
		e.CurrentMood = "neutral"
	}

	// 更新趋势 (简化逻辑)
	if e.Valence > 0.3 {
		e.MoodTrend = "improving"
	} else if e.Valence < -0.3 {
		e.MoodTrend = "declining"
	} else {
		e.MoodTrend = "stable"
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
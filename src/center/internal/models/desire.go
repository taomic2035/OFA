package models

import "time"

// Desire 欲望模型 (v4.0.0)
// 基于马斯洛需求层次理论
type Desire struct {
	// === 马斯洛需求层次 (0-1 满足度) ===
	Physiological     float64 `json:"physiological"`      // 生理需求 (食、睡、性、安全感)
	Safety            float64 `json:"safety"`             // 安全需求 (稳定、保障、秩序)
	LoveBelonging     float64 `json:"love_belonging"`     // 爱与归属 (友谊、爱情、家庭)
	Esteem            float64 `json:"esteem"`             // 尊重需求 (自尊、认可、成就)
	SelfActualization float64 `json:"self_actualization"` // 自我实现 (潜能、创造力、意义)

	// === 当前驱动力 ===
	PrimaryDesire  string    `json:"primary_desire"`  // 当前主要欲望
	DesireStrength float64   `json:"desire_strength"` // 欲望强度 (0-1)
	DesireTarget   string    `json:"desire_target"`   // 欲望目标
	DesireAction   string    `json:"desire_action"`   // 欲望驱动的行为

	// === 欲望历史 ===
	SatisfiedDesires  []DesireRecord `json:"satisfied_desires,omitempty"`  // 已满足欲望
	UnsatisfiedDesires []DesireRecord `json:"unsatisfied_desires,omitempty"` // 未满足欲望
	PendingDesires    []DesireRecord `json:"pending_desires,omitempty"`    // 待处理欲望

	// === 时间属性 ===
	Timestamp time.Time `json:"timestamp"`
}

// DesireRecord 欲望记录
type DesireRecord struct {
	RecordID     string                 `json:"record_id"`
	DesireType   string                 `json:"desire_type"`   // physiological/safety/love_belonging/esteem/self_actualization
	DesireName   string                 `json:"desire_name"`   // 欲望名称
	Strength     float64                `json:"strength"`      // 欲望强度
	Target       string                 `json:"target"`        // 欲望目标
	Status       string                 `json:"status"`        // pending/satisfied/unsatisfied/partial
	Satisfaction float64                `json:"satisfaction"`  // 满足程度
	Duration     int                    `json:"duration"`      // 持续时间(分钟)
	Attempts     int                    `json:"attempts"`      // 尝试次数
	Context      map[string]interface{} `json:"context,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ResolvedAt   time.Time              `json:"resolved_at,omitempty"`
}

// DesireProfile 欲望特征画像
type DesireProfile struct {
	IdentityID string `json:"identity_id"`

	// === 需求层次偏好 ===
	HierarchyPreference map[string]float64 `json:"hierarchy_preference"` // 各层次重要性权重

	// === 需求阈值 ===
	SatisfactionThreshold map[string]float64 `json:"satisfaction_threshold"` // 满足阈值
	FrustrationThreshold  map[string]float64 `json:"frustration_threshold"`  // 挫折阈值

	// === 欲望特质 ===
	DesireIntensity    float64 `json:"desire_intensity"`    // 整体欲望强度倾向
	DesirePersistence  float64 `json:"desire_persistence"`  // 欲望持久性
	DesireVariety      float64 `json:"desire_variety"`      // 欲望多样性
	DesirePatience     float64 `json:"desire_patience"`     // 欲望耐心程度

	// === 常见欲望 ===
	CommonDesires []string `json:"common_desires,omitempty"` // 常见欲望类型

	// === 欲望模式 ===
	DesirePatterns []DesirePattern `json:"desire_patterns,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DesirePattern 欲望模式
type DesirePattern struct {
	PatternID      string   `json:"pattern_id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TriggerConditions []string `json:"trigger_conditions,omitempty"`
	DesireSequence []string `json:"desire_sequence,omitempty"` // 欲望触发序列
	Frequency      float64  `json:"frequency"`
	LastOccurrence time.Time `json:"last_occurrence"`
}

// NeedLevel 需求层次级别
type NeedLevel struct {
	Level         int     `json:"level"`          // 层级 (1-5)
	Name          string  `json:"name"`           // 名称
	Description   string  `json:"description"`    // 描述
	Importance    float64 `json:"importance"`     // 重要性
	Satisfaction  float64 `json:"satisfaction"`   // 当前满足度
	Urgency       float64 `json:"urgency"`        // 紧迫性
	Dependencies  []int   `json:"dependencies"`   // 依赖的底层需求
}

// NewDesire 创建默认欲望
func NewDesire() *Desire {
	return &Desire{
		Physiological:     0.7,  // 生理需求基本满足
		Safety:            0.6,  // 安全需求中等
		LoveBelonging:     0.5,  // 爱与归属中等
		Esteem:            0.4,  // 尊重需求略低
		SelfActualization: 0.3,  // 自我实现待发展
		PrimaryDesire:     "esteem",
		DesireStrength:    0.5,
		DesireTarget:      "个人成就",
		DesireAction:      "",
		SatisfiedDesires:  []DesireRecord{},
		UnsatisfiedDesires: []DesireRecord{},
		PendingDesires:    []DesireRecord{},
		Timestamp:         time.Now(),
	}
}

// NewDesireProfile 创建默认欲望画像
func NewDesireProfile(identityID string) *DesireProfile {
	return &DesireProfile{
		IdentityID: identityID,
		HierarchyPreference: map[string]float64{
			"physiological":      0.15,
			"safety":             0.15,
			"love_belonging":     0.20,
			"esteem":             0.25,
			"self_actualization": 0.25,
		},
		SatisfactionThreshold: map[string]float64{
			"physiological":      0.7,
			"safety":             0.6,
			"love_belonging":     0.5,
			"esteem":             0.5,
			"self_actualization": 0.4,
		},
		FrustrationThreshold: map[string]float64{
			"physiological":      0.3,
			"safety":             0.4,
			"love_belonging":     0.3,
			"esteem":             0.2,
			"self_actualization": 0.1,
		},
		DesireIntensity:   0.5,
		DesirePersistence: 0.5,
		DesireVariety:     0.5,
		DesirePatience:    0.5,
		CommonDesires:     []string{"achievement", "connection", "security", "growth"},
		DesirePatterns:    []DesirePattern{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// GetNeedLevels 获取需求层次
func (d *Desire) GetNeedLevels() []NeedLevel {
	return []NeedLevel{
		{Level: 1, Name: "生理需求", Description: "食物、睡眠、安全、性", Importance: 1.0, Satisfaction: d.Physiological, Urgency: 1.0 - d.Physiological, Dependencies: []int{}},
		{Level: 2, Name: "安全需求", Description: "稳定、保障、秩序", Importance: 0.9, Satisfaction: d.Safety, Urgency: 1.0 - d.Safety, Dependencies: []int{1}},
		{Level: 3, Name: "爱与归属", Description: "友谊、爱情、家庭", Importance: 0.8, Satisfaction: d.LoveBelonging, Urgency: 1.0 - d.LoveBelonging, Dependencies: []int{1, 2}},
		{Level: 4, Name: "尊重需求", Description: "自尊、认可、成就", Importance: 0.7, Satisfaction: d.Esteem, Urgency: 1.0 - d.Esteem, Dependencies: []int{1, 2, 3}},
		{Level: 5, Name: "自我实现", Description: "潜能、创造力、意义", Importance: 0.6, Satisfaction: d.SelfActualization, Urgency: 1.0 - d.SelfActualization, Dependencies: []int{1, 2, 3, 4}},
	}
}

// GetDominantNeed 获取主导需求（最紧迫的不满足需求）
func (d *Desire) GetDominantNeed() NeedLevel {
	levels := d.GetNeedLevels()

	// 找到满足度最低的需求
	dominant := levels[0]
	minSatisfaction := d.Physiological

	for _, level := range levels {
		if level.Satisfaction < minSatisfaction {
			minSatisfaction = level.Satisfaction
			dominant = level
		}
	}

	return dominant
}

// CalculateOverallSatisfaction 计算整体满足度
func (d *Desire) CalculateOverallSatisfaction() float64 {
	// 加权平均
	weights := map[string]float64{
		"physiological":      1.0,
		"safety":             0.8,
		"love_belonging":     0.6,
		"esteem":             0.4,
		"self_actualization": 0.2,
	}

	total := 0.0
	weightSum := 0.0

	for need, weight := range weights {
		satisfaction := 0.0
		switch need {
		case "physiological":
			satisfaction = d.Physiological
		case "safety":
			satisfaction = d.Safety
		case "love_belonging":
			satisfaction = d.LoveBelonging
		case "esteem":
			satisfaction = d.Esteem
		case "self_actualization":
			satisfaction = d.SelfActualization
		}
		total += satisfaction * weight
		weightSum += weight
	}

	return total / weightSum
}

// CalculateFrustrationLevel 计算挫折程度
func (d *Desire) CalculateFrustrationLevel() float64 {
	// 找到最不满足的需求
	dominant := d.GetDominantNeed()

	// 挫折 = 1 - 满足度
	frustration := 1.0 - dominant.Satisfaction

	// 根据层级调整（底层需求未满足更痛苦）
	levelFactor := 1.0 + float64(5 - dominant.Level) * 0.1

	return frustration * levelFactor
}

// UpdateDesire 更新欲望状态
func (d *Desire) UpdateDesire(record DesireRecord) {
	// 更新对应需求层次
	switch record.DesireType {
	case "physiological":
		d.Physiological = (d.Physiological + record.Satisfaction) / 2
	case "safety":
		d.Safety = (d.Safety + record.Satisfaction) / 2
	case "love_belonging":
		d.LoveBelonging = (d.LoveBelonging + record.Satisfaction) / 2
	case "esteem":
		d.Esteem = (d.Esteem + record.Satisfaction) / 2
	case "self_actualization":
		d.SelfActualization = (d.SelfActualization + record.Satisfaction) / 2
	}

	// 根据状态归类
	switch record.Status {
	case "satisfied":
		d.SatisfiedDesires = append(d.SatisfiedDesires, record)
		if len(d.SatisfiedDesires) > 20 {
			d.SatisfiedDesires = d.SatisfiedDesires[len(d.SatisfiedDesires)-20:]
		}
	case "unsatisfied":
		d.UnsatisfiedDesires = append(d.UnsatisfiedDesires, record)
		if len(d.UnsatisfiedDesires) > 20 {
			d.UnsatisfiedDesires = d.UnsatisfiedDesires[len(d.UnsatisfiedDesires)-20:]
		}
	case "pending":
		d.PendingDesires = append(d.PendingDesires, record)
		if len(d.PendingDesires) > 10 {
			d.PendingDesires = d.PendingDesires[len(d.PendingDesires)-10:]
		}
	}

	// 更新主导欲望
	dominant := d.GetDominantNeed()
	d.PrimaryDesire = dominant.Name
	d.DesireStrength = dominant.Urgency
	d.Timestamp = time.Now()
}

// Normalize 归一化
func (d *Desire) Normalize() {
	normalizeValue := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	d.Physiological = normalizeValue(d.Physiological)
	d.Safety = normalizeValue(d.Safety)
	d.LoveBelonging = normalizeValue(d.LoveBelonging)
	d.Esteem = normalizeValue(d.Esteem)
	d.SelfActualization = normalizeValue(d.SelfActualization)
	d.DesireStrength = normalizeValue(d.DesireStrength)
}

// Decay 需求满足度衰减（随时间降低）
func (d *Desire) Decay(rate float64, minutes int) {
	// 衰减因子
	decayFactor := 1 - rate * float64(minutes) / 120.0 // 2小时完全衰减
	if decayFactor < 0.3 {
		decayFactor = 0.3
	}

	// 应用衰减（底层需求衰减更快）
	d.Physiological *= (decayFactor - 0.1) // 生理需求衰减最快
	d.Safety *= (decayFactor - 0.05)
	d.LoveBelonging *= decayFactor
	d.Esteem *= (decayFactor + 0.05) // 高层需求衰减慢
	d.SelfActualization *= (decayFactor + 0.1)

	d.Normalize()
	d.Timestamp = time.Now()
}

// TriggerDesire 触发欲望
func (d *Desire) TriggerDesire(desireType string, strength float64, target string) {
	// 增加对应需求的紧迫性（降低满足度）
	switch desireType {
	case "physiological":
		d.Physiological -= strength * 0.3
	case "safety":
		d.Safety -= strength * 0.25
	case "love_belonging":
		d.LoveBelonging -= strength * 0.2
	case "esteem":
		d.Esteem -= strength * 0.15
	case "self_actualization":
		d.SelfActualization -= strength * 0.1
	}

	d.PrimaryDesire = desireType
	d.DesireStrength = strength
	d.DesireTarget = target
	d.Normalize()
	d.Timestamp = time.Now()
}

// SatisfyDesire 满足欲望
func (d *Desire) SatisfyDesire(desireType string, amount float64) {
	// 增加满足度
	switch desireType {
	case "physiological":
		d.Physiological += amount * 0.3
	case "safety":
		d.Safety += amount * 0.25
	case "love_belonging":
		d.LoveBelonging += amount * 0.2
	case "esteem":
		d.Esteem += amount * 0.15
	case "self_actualization":
		d.SelfActualization += amount * 0.1
	}

	d.Normalize()
	d.Timestamp = time.Now()
}
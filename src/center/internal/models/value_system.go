package models

import "time"

// EnhancedValueSystem 增强版价值观系统 (v4.1.0)
// 扩展现有 ValueSystem，增加价值判断逻辑和层级结构
type EnhancedValueSystem struct {
	// === 核心价值观 (继承并扩展) ===
	// 基础价值观 (0-1)
	Privacy       float64 `json:"privacy"`
	Efficiency    float64 `json:"efficiency"`
	Health        float64 `json:"health"`
	Family        float64 `json:"family"`
	Career        float64 `json:"career"`
	Entertainment float64 `json:"entertainment"`
	Learning      float64 `json:"learning"`
	Social        float64 `json:"social"`
	Finance       float64 `json:"finance"`
	Environment   float64 `json:"environment"`

	// === 扩展价值观 ===
	Freedom       float64 `json:"freedom"`       // 自由
	Justice       float64 `json:"justice"`       // 公正
	Honesty       float64 `json:"honesty"`       // 诚实
	Compassion    float64 `json:"compassion"`    // 同情心
	Creativity    float64 `json:"creativity"`    // 创造力
	Tradition     float64 `json:"tradition"`     // 传统
	Innovation    float64 `json:"innovation"`    // 创新
	Achievement   float64 `json:"achievement"`   // 成就
	Security      float64 `json:"security"`      // 安全
	Autonomy      float64 `json:"autonomy"`      // 自主

	// === 决策倾向 ===
	RiskTolerance float64 `json:"risk_tolerance"`
	Impulsiveness float64 `json:"impulsiveness"`
	Patience      float64 `json:"patience"`

	// === 道德判断框架 ===
	MoralFramework *MoralFramework `json:"moral_framework,omitempty"`

	// === 价值冲突解决策略 ===
	ConflictResolution string `json:"conflict_resolution"` // utilitarian/deontological/virtue/care

	// === 自定义价值观 ===
	CustomValues map[string]float64 `json:"custom_values,omitempty"`

	// === 价值观描述 ===
	Summary string `json:"summary"`

	// === 时间属性 ===
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MoralFramework 道德判断框架
type MoralFramework struct {
	// 道德基础理论 (Jonathan Haidt)
	CareHarm      float64 `json:"care_harm"`      // 关爱/伤害
	FairnessCheating float64 `json:"fairness_cheating"` // 公平/欺骗
	LoyaltyBetrayal float64 `json:"loyalty_betrayal"` // 忠诚/背叛
	AuthoritySubversion float64 `json:"authority_subversion"` // 权威/颠覆
	SanctityDegradation float64 `json:"sanctity_degradation"` // 圣洁/堕落
	LibertyOppression float64 `json:"liberty_oppression"` // 自由/压迫

	// 道德推理风格
	ReasoningStyle string `json:"reasoning_style"` // intuitive/deliberative/mixed
}

// ValueJudgment 价值判断结果
type ValueJudgment struct {
	JudgmentID   string             `json:"judgment_id"`
	Situation    string             `json:"situation"`     // 情境描述
	Options      []ValueOption      `json:"options"`       // 可选项
	ChosenOption string             `json:"chosen_option"` // 选择
	Reasoning    string             `json:"reasoning"`     // 推理过程
	ValuesUsed   map[string]float64 `json:"values_used"`   // 使用的价值观
	Confidence   float64            `json:"confidence"`    // 置信度
	Timestamp    time.Time          `json:"timestamp"`
}

// ValueOption 价值选项
type ValueOption struct {
	OptionID    string             `json:"option_id"`
	Description string             `json:"description"`
	ValueScores map[string]float64 `json:"value_scores"` // 各价值观评分
	OverallScore float64           `json:"overall_score"`
}

// ValueConflict 价值冲突
type ValueConflict struct {
	ConflictID   string    `json:"conflict_id"`
	Value1       string    `json:"value_1"`
	Value2       string    `json:"value_2"`
	Description  string    `json:"description"`
	Resolution   string    `json:"resolution"`   // compromise/prioritize/transcend
	ResolutionNote string  `json:"resolution_note"`
	Timestamp    time.Time `json:"timestamp"`
}

// ValueSystemProfile 价值观画像
type ValueSystemProfile struct {
	IdentityID string `json:"identity_id"`

	// === 核心价值 ===
	CoreValues []CoreValue `json:"core_values,omitempty"`

	// === 价值层级 ===
	ValueHierarchy []ValueLevel `json:"value_hierarchy,omitempty"`

	// === 价值稳定性 ===
	StabilityScore float64 `json:"stability_score"`
	ConsistencyScore float64 `json:"consistency_score"` // 价值观与行为一致性

	// === 历史判断记录 ===
	JudgmentHistory []ValueJudgment `json:"judgment_history,omitempty"`

	// === 价值冲突历史 ===
	ConflictHistory []ValueConflict `json:"conflict_history,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CoreValue 核心价值
type CoreValue struct {
	ValueID     string   `json:"value_id"`
	Name        string   `json:"name"`
	Importance  float64  `json:"importance"`
	Description string   `json:"description"`
	Source      string   `json:"source"`      // family/experience/education/reflection
	Evidence    []string `json:"evidence,omitempty"` // 体现此价值的行为
}

// ValueLevel 价值层级
type ValueLevel struct {
	Level    int      `json:"level"`     // 1=最高优先级
	Values   []string `json:"values"`    // 该层级包含的价值观
	Category string   `json:"category"`  // terminal/instrumental
}

// NewEnhancedValueSystem 创建默认增强版价值观
func NewEnhancedValueSystem() *EnhancedValueSystem {
	now := time.Now()
	return &EnhancedValueSystem{
		// 基础价值观
		Privacy:        0.7,
		Efficiency:     0.6,
		Health:         0.8,
		Family:         0.8,
		Career:         0.6,
		Entertainment:  0.5,
		Learning:       0.7,
		Social:         0.5,
		Finance:        0.6,
		Environment:    0.5,

		// 扩展价值观
		Freedom:     0.7,
		Justice:     0.6,
		Honesty:     0.8,
		Compassion:  0.7,
		Creativity:  0.6,
		Tradition:   0.4,
		Innovation:  0.6,
		Achievement: 0.6,
		Security:    0.6,
		Autonomy:    0.7,

		// 决策倾向
		RiskTolerance: 0.4,
		Impulsiveness: 0.3,
		Patience:      0.6,

		// 道德框架
		MoralFramework: &MoralFramework{
			CareHarm:           0.8,
			FairnessCheating:   0.7,
			LoyaltyBetrayal:    0.5,
			AuthoritySubversion: 0.4,
			SanctityDegradation: 0.3,
			LibertyOppression:  0.7,
			ReasoningStyle:     "mixed",
		},

		ConflictResolution: "utilitarian",
		CustomValues:       make(map[string]float64),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// NewValueSystemProfile 创建默认价值观画像
func NewValueSystemProfile(identityID string) *ValueSystemProfile {
	return &ValueSystemProfile{
		IdentityID:       identityID,
		CoreValues:       []CoreValue{},
		ValueHierarchy:   []ValueLevel{},
		StabilityScore:   0.5,
		ConsistencyScore: 0.5,
		JudgmentHistory:  []ValueJudgment{},
		ConflictHistory:  []ValueConflict{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// GetTopValues 获取最重要的价值观
func (v *EnhancedValueSystem) GetTopValues(n int) []string {
	values := map[string]float64{
		"privacy":       v.Privacy,
		"efficiency":    v.Efficiency,
		"health":        v.Health,
		"family":        v.Family,
		"career":        v.Career,
		"entertainment": v.Entertainment,
		"learning":      v.Learning,
		"social":        v.Social,
		"finance":       v.Finance,
		"environment":   v.Environment,
		"freedom":       v.Freedom,
		"justice":       v.Justice,
		"honesty":       v.Honesty,
		"compassion":    v.Compassion,
		"creativity":    v.Creativity,
		"tradition":     v.Tradition,
		"innovation":    v.Innovation,
		"achievement":   v.Achievement,
		"security":      v.Security,
		"autonomy":      v.Autonomy,
	}

	// 添加自定义价值观
	for key, value := range v.CustomValues {
		values[key] = value
	}

	// 简单排序
	type kv struct {
		Key   string
		Value float64
	}

	var pairs []kv
	for k, val := range values {
		pairs = append(pairs, kv{k, val})
	}

	// 冒泡排序
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Value > pairs[i].Value {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	var result []string
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].Key)
	}

	return result
}

// GetValueCategory 获取价值观类别
func (v *EnhancedValueSystem) GetValueCategory(valueName string) string {
	terminalValues := map[string]bool{
		"health": true, "family": true, "freedom": true,
		"happiness": true, "achievement": true, "security": true,
	}

	instrumentalValues := map[string]bool{
		"honesty": true, "efficiency": true, "creativity": true,
		"compassion": true, "innovation": true, "autonomy": true,
	}

	if terminalValues[valueName] {
		return "terminal"
	}
	if instrumentalValues[valueName] {
		return "instrumental"
	}
	return "other"
}

// MakeValueJudgment 进行价值判断
func (v *EnhancedValueSystem) MakeValueJudgment(situation string, options []ValueOption) *ValueJudgment {
	judgment := &ValueJudgment{
		JudgmentID: generateJudgmentID(),
		Situation:  situation,
		Options:    options,
		ValuesUsed: make(map[string]float64),
		Timestamp:  time.Now(),
	}

	// 评估每个选项
	topValues := v.GetTopValues(5)
	for i := range options {
		score := 0.0
		for _, valueName := range topValues {
			valueScore, exists := options[i].ValueScores[valueName]
			if !exists {
				// 根据系统价值观评分
				valueScore = v.getValueScore(valueName)
			}
			// 加权评分
			weight := v.getValueScore(valueName)
			score += valueScore * weight
		}
		options[i].OverallScore = score / float64(len(topValues))
	}

	// 选择最高分选项
	maxScore := -1.0
	for _, opt := range options {
		if opt.OverallScore > maxScore {
			maxScore = opt.OverallScore
			judgment.ChosenOption = opt.OptionID
		}
	}

	// 记录使用的价值观
	for _, valueName := range topValues {
		judgment.ValuesUsed[valueName] = v.getValueScore(valueName)
	}

	// 计算置信度
	judgment.Confidence = 0.7 // 基础置信度

	return judgment
}

// getValueScore 获取价值观分数
func (v *EnhancedValueSystem) getValueScore(valueName string) float64 {
	switch valueName {
	case "privacy":
		return v.Privacy
	case "efficiency":
		return v.Efficiency
	case "health":
		return v.Health
	case "family":
		return v.Family
	case "career":
		return v.Career
	case "entertainment":
		return v.Entertainment
	case "learning":
		return v.Learning
	case "social":
		return v.Social
	case "finance":
		return v.Finance
	case "environment":
		return v.Environment
	case "freedom":
		return v.Freedom
	case "justice":
		return v.Justice
	case "honesty":
		return v.Honesty
	case "compassion":
		return v.Compassion
	case "creativity":
		return v.Creativity
	case "tradition":
		return v.Tradition
	case "innovation":
		return v.Innovation
	case "achievement":
		return v.Achievement
	case "security":
		return v.Security
	case "autonomy":
		return v.Autonomy
	default:
		if v.CustomValues != nil {
			return v.CustomValues[valueName]
		}
		return 0.5
	}
}

// CalculateInfluence 计算价值观对决策的影响
func (v *EnhancedValueSystem) CalculateInfluence() map[string]float64 {
	influence := make(map[string]float64)

	// 隐私相关决策
	influence["privacy_awareness"] = v.Privacy

	// 效率优先
	influence["efficiency_priority"] = v.Efficiency

	// 健康考量
	influence["health_consciousness"] = v.Health

	// 家庭优先
	influence["family_priority"] = v.Family

	// 事业抱负
	influence["career_ambition"] = v.Career

	// 社交倾向
	influence["social_orientation"] = v.Social

	// 创新倾向
	influence["innovation_tendency"] = (v.Innovation + v.Creativity) / 2

	// 道德敏感度
	influence["moral_sensitivity"] = (v.Honesty + v.Justice + v.Compassion) / 3

	// 自主性
	influence["autonomy_level"] = v.Autonomy

	// 安全意识
	influence["security_awareness"] = v.Security

	return influence
}

// ResolveConflict 解决价值观冲突
func (v *EnhancedValueSystem) ResolveConflict(value1, value2 string, context string) *ValueConflict {
	conflict := &ValueConflict{
		ConflictID:  generateConflictID(),
		Value1:      value1,
		Value2:      value2,
		Description: context,
		Timestamp:   time.Now(),
	}

	score1 := v.getValueScore(value1)
	score2 := v.getValueScore(value2)

	switch v.ConflictResolution {
	case "utilitarian":
		// 功利主义：选择产生最大善的
		if score1 > score2 {
			conflict.Resolution = "prioritize"
			conflict.ResolutionNote = "优先 " + value1 + " (功利主义考量)"
		} else {
			conflict.Resolution = "prioritize"
			conflict.ResolutionNote = "优先 " + value2 + " (功利主义考量)"
		}
	case "deontological":
		// 义务论：遵循道德规则
		conflict.Resolution = "compromise"
		conflict.ResolutionNote = "寻求折中方案 (义务论考量)"
	case "virtue":
		// 德性伦理：考虑行为者的品格
		conflict.Resolution = "transcend"
		conflict.ResolutionNote = "超越冲突，寻找更高层面的解决方案"
	case "care":
		// 关怀伦理：考虑关系和情境
		conflict.Resolution = "compromise"
		conflict.ResolutionNote = "在具体情境中寻求平衡"
	default:
		conflict.Resolution = "prioritize"
		if score1 > score2 {
			conflict.ResolutionNote = "优先 " + value1
		} else {
			conflict.ResolutionNote = "优先 " + value2
		}
	}

	return conflict
}

// Normalize 归一化
func (v *EnhancedValueSystem) Normalize() {
	normalizeValue := func(val float64) float64 {
		if val < 0 {
			return 0
		}
		if val > 1 {
			return 1
		}
		return val
	}

	v.Privacy = normalizeValue(v.Privacy)
	v.Efficiency = normalizeValue(v.Efficiency)
	v.Health = normalizeValue(v.Health)
	v.Family = normalizeValue(v.Family)
	v.Career = normalizeValue(v.Career)
	v.Entertainment = normalizeValue(v.Entertainment)
	v.Learning = normalizeValue(v.Learning)
	v.Social = normalizeValue(v.Social)
	v.Finance = normalizeValue(v.Finance)
	v.Environment = normalizeValue(v.Environment)
	v.Freedom = normalizeValue(v.Freedom)
	v.Justice = normalizeValue(v.Justice)
	v.Honesty = normalizeValue(v.Honesty)
	v.Compassion = normalizeValue(v.Compassion)
	v.Creativity = normalizeValue(v.Creativity)
	v.Tradition = normalizeValue(v.Tradition)
	v.Innovation = normalizeValue(v.Innovation)
	v.Achievement = normalizeValue(v.Achievement)
	v.Security = normalizeValue(v.Security)
	v.Autonomy = normalizeValue(v.Autonomy)
	v.RiskTolerance = normalizeValue(v.RiskTolerance)
	v.Impulsiveness = normalizeValue(v.Impulsiveness)
	v.Patience = normalizeValue(v.Patience)

	v.UpdatedAt = time.Now()
}

// ToValueSystem 转换为基础 ValueSystem
func (v *EnhancedValueSystem) ToValueSystem() *ValueSystem {
	return &ValueSystem{
		Privacy:        v.Privacy,
		Efficiency:     v.Efficiency,
		Health:         v.Health,
		Family:         v.Family,
		Career:         v.Career,
		Entertainment:  v.Entertainment,
		Learning:       v.Learning,
		Social:         v.Social,
		Finance:        v.Finance,
		Environment:    v.Environment,
		RiskTolerance:  v.RiskTolerance,
		Impulsiveness:  v.Impulsiveness,
		Patience:       v.Patience,
		CustomValues:   v.CustomValues,
		Summary:        v.Summary,
	}
}

// 辅助函数
func generateJudgmentID() string {
	return "judgment_" + time.Now().Format("20060102150405")
}

func generateConflictID() string {
	return "conflict_" + time.Now().Format("20060102150405")
}
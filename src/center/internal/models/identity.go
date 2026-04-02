package models

import (
	"time"
)

// === 核心身份数据模型 ===

// PersonalIdentity - 个人身份核心
type PersonalIdentity struct {
	ID         string    `json:"id" bson:"_id"`
	Name       string    `json:"name" bson:"name"`
	Nickname   string    `json:"nickname" bson:"nickname"`
	Avatar     string    `json:"avatar" bson:"avatar"`
	Birthday   time.Time `json:"birthday" bson:"birthday"`
	Gender     string    `json:"gender" bson:"gender"`
	Location   string    `json:"location" bson:"location"`
	Occupation string    `json:"occupation" bson:"occupation"`
	Languages  []string  `json:"languages" bson:"languages"`
	Timezone   string    `json:"timezone" bson:"timezone"`

	// 核心特质
	Personality  *Personality  `json:"personality" bson:"personality"`
	ValueSystem  *ValueSystem  `json:"value_system" bson:"value_system"`
	Interests    []Interest    `json:"interests" bson:"interests"`

	// 数字资产
	VoiceProfile  *VoiceProfile  `json:"voice_profile" bson:"voice_profile"`
	WritingStyle  *WritingStyle  `json:"writing_style" bson:"writing_style"`

	// 元数据
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// Personality - 性格特质（基于 Big Five 模型）
type Personality struct {
	// Big Five 核心特质 (0-1)
	Openness          float64 `json:"openness" bson:"openness"`                     // 开放性
	Conscientiousness float64 `json:"conscientiousness" bson:"conscientiousness"`   // 尽责性
	Extraversion      float64 `json:"extraversion" bson:"extraversion"`             // 外向性
	Agreeableness     float64 `json:"agreeableness" bson:"agreeableness"`           // 宜人性
	Neuroticism       float64 `json:"neuroticism" bson:"neuroticism"`               // 神经质

	// 自定义特质
	CustomTraits map[string]float64 `json:"custom_traits" bson:"custom_traits"`

	// 说话风格
	SpeakingTone   string  `json:"speaking_tone" bson:"speaking_tone"`       // formal/casual/humorous
	ResponseLength string  `json:"response_length" bson:"response_length"`   // brief/moderate/detailed
	EmojiUsage     float64 `json:"emoji_usage" bson:"emoji_usage"`           // 表情使用频率 (0-1)

	// 性格描述（AI 生成）
	Summary string `json:"summary" bson:"summary"`
}

// PersonalityTone - 说话语调类型
type PersonalityTone string

const (
	ToneFormal   PersonalityTone = "formal"
	ToneCasual   PersonalityTone = "casual"
	ToneHumorous PersonalityTone = "humorous"
	ToneWarm     PersonalityTone = "warm"
	ToneNeutral  PersonalityTone = "neutral"
)

// ResponseLength - 回复长度类型
type ResponseLength string

const (
	LengthBrief    ResponseLength = "brief"
	LengthModerate ResponseLength = "moderate"
	LengthDetailed ResponseLength = "detailed"
)

// ValueSystem - 价值观系统
type ValueSystem struct {
	// 核心价值观权重 (0-1)
	Privacy       float64 `json:"privacy" bson:"privacy"`                 // 隐私重视程度
	Efficiency    float64 `json:"efficiency" bson:"efficiency"`           // 效率优先程度
	Health        float64 `json:"health" bson:"health"`                   // 健康重视程度
	Family        float64 `json:"family" bson:"family"`                   // 家庭重视程度
	Career        float64 `json:"career" bson:"career"`                   // 事业重视程度
	Entertainment float64 `json:"entertainment" bson:"entertainment"`     // 娱乐重视程度
	Learning      float64 `json:"learning" bson:"learning"`               // 学习重视程度
	Social        float64 `json:"social" bson:"social"`                   // 社交重视程度
	Finance       float64 `json:"finance" bson:"finance"`                 // 财务重视程度
	Environment   float64 `json:"environment" bson:"environment"`         // 环保重视程度

	// 决策倾向
	RiskTolerance float64 `json:"risk_tolerance" bson:"risk_tolerance"` // 风险承受度 (0-1)
	Impulsiveness float64 `json:"impulsiveness" bson:"impulsiveness"`   // 冲动程度 (0-1)
	Patience      float64 `json:"patience" bson:"patience"`             // 耐心程度 (0-1)

	// 自定义价值观
	CustomValues map[string]float64 `json:"custom_values" bson:"custom_values"`

	// 价值观描述（AI 生成）
	Summary string `json:"summary" bson:"summary"`
}

// Interest - 兴趣爱好
type Interest struct {
	ID          string    `json:"id" bson:"id"`
	Category    string    `json:"category" bson:"category"`       // sports/tech/art/music/food/travel...
	Name        string    `json:"name" bson:"name"`               // 具体名称
	Level       float64   `json:"level" bson:"level"`             // 热衷程度 (0-1)
	Keywords    []string  `json:"keywords" bson:"keywords"`       // 关键词
	Description string    `json:"description" bson:"description"` // 描述
	Since       time.Time `json:"since" bson:"since"`             // 开始时间
	LastActive  time.Time `json:"last_active" bson:"last_active"` // 最近活跃
}

// InterestCategory - 兴趣类别
type InterestCategory string

const (
	CategorySports   InterestCategory = "sports"
	CategoryTech     InterestCategory = "tech"
	CategoryArt      InterestCategory = "art"
	CategoryMusic    InterestCategory = "music"
	CategoryFood     InterestCategory = "food"
	CategoryTravel   InterestCategory = "travel"
	CategoryReading  InterestCategory = "reading"
	CategoryGaming   InterestCategory = "gaming"
	CategoryFitness  InterestCategory = "fitness"
	CategoryMovies   InterestCategory = "movies"
	CategoryFashion  InterestCategory = "fashion"
	CategoryFinance  InterestCategory = "finance"
	CategorySocial   InterestCategory = "social"
	CategoryOther    InterestCategory = "other"
)

// VoiceProfile - 语音音色配置
type VoiceProfile struct {
	ID               string    `json:"id" bson:"id"`
	VoiceType        string    `json:"voice_type" bson:"voice_type"`                     // clone/synthetic/preset
	PresetVoiceID    string    `json:"preset_voice_id" bson:"preset_voice_id"`           // 预设音色ID
	CloneReferenceID string    `json:"clone_reference_id" bson:"clone_reference_id"`     // 克隆参考ID

	// 音色参数
	Pitch  float64 `json:"pitch" bson:"pitch"`     // 音高 (0-2, 1.0 为正常)
	Speed  float64 `json:"speed" bson:"speed"`     // 语速 (0-2, 1.0 为正常)
	Volume float64 `json:"volume" bson:"volume"`   // 音量 (0-1)

	// 风格参数
	Tone         string  `json:"tone" bson:"tone"`                 // warm/neutral/energetic
	Accent       string  `json:"accent" bson:"accent"`             // 口音
	EmotionLevel float64 `json:"emotion_level" bson:"emotion_level"` // 情感表达程度 (0-1)

	// 语调模式
	PausePattern  string `json:"pause_pattern" bson:"pause_pattern"`     // 停顿模式
	EmphasisStyle string `json:"emphasis_style" bson:"emphasis_style"`   // 重音风格

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// VoiceType - 音色类型
type VoiceType string

const (
	VoiceTypeClone     VoiceType = "clone"
	VoiceTypeSynthetic VoiceType = "synthetic"
	VoiceTypePreset    VoiceType = "preset"
)

// WritingStyle - 写作风格
type WritingStyle struct {
	// 风格参数
	Formality    float64 `json:"formality" bson:"formality"`       // 正式程度 (0-1)
	Verbosity    float64 `json:"verbosity" bson:"verbosity"`       // 冗长程度 (0-1)
	Humor        float64 `json:"humor" bson:"humor"`               // 幽默程度 (0-1)
	Technicality float64 `json:"technicality" bson:"technicality"` // 专业程度 (0-1)

	// 文风特征
	UseEmoji    bool `json:"use_emoji" bson:"use_emoji"`
	UseGIFs     bool `json:"use_gifs" bson:"use_gifs"`
	UseMarkdown bool `json:"use_markdown" bson:"use_markdown"`

	// 标志性用语
	SignaturePhrase string `json:"signature_phrase" bson:"signature_phrase"`

	// 常用词汇
	FrequentWords []string `json:"frequent_words" bson:"frequent_words"`
	AvoidWords    []string `json:"avoid_words" bson:"avoid_words"`

	// 语言习惯
	PreferredGreeting string `json:"preferred_greeting" bson:"preferred_greeting"` // 偏好的问候语
	PreferredClosing  string `json:"preferred_closing" bson:"preferred_closing"`   // 偏好的结束语
}

// === 辅助结构 ===

// PersonalityAssessment - 性格评估结果
type PersonalityAssessment struct {
	ID          string                 `json:"id" bson:"_id"`
	UserID      string                 `json:"user_id" bson:"user_id"`
	Method      string                 `json:"method" bson:"method"`           // questionnaire/behavior_inference/combined
	Questions   []AssessmentQuestion   `json:"questions" bson:"questions"`
	Responses   []AssessmentResponse   `json:"responses" bson:"responses"`
	Result      *Personality           `json:"result" bson:"result"`
	Confidence  float64                `json:"confidence" bson:"confidence"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
}

// AssessmentQuestion - 评估问题
type AssessmentQuestion struct {
	ID           string   `json:"id" bson:"id"`
	Text         string   `json:"text" bson:"text"`
	Category     string   `json:"category" bson:"category"`
	Options      []string `json:"options" bson:"options"`
	Weights      []float64 `json:"weights" bson:"weights"` // 每个选项对各特质的影响权重
}

// AssessmentResponse - 评估响应
type AssessmentResponse struct {
	QuestionID string `json:"question_id" bson:"question_id"`
	Response   string `json:"response" bson:"response"`
	Score      float64 `json:"score" bson:"score"`
}

// BehaviorObservation - 行为观察记录（用于推断性格）
type BehaviorObservation struct {
	ID          string                 `json:"id" bson:"_id"`
	UserID      string                 `json:"user_id" bson:"user_id"`
	Type        string                 `json:"type" bson:"type"`         // decision/interaction/preference/activity
	Context     map[string]interface{} `json:"context" bson:"context"`
	Outcome     string                 `json:"outcome" bson:"outcome"`
	Inferences  map[string]float64     `json:"inferences" bson:"inferences"` // 推断的性格特质
	Timestamp   time.Time              `json:"timestamp" bson:"timestamp"`
}

// === 工厂函数 ===

// NewPersonalIdentity 创建新的个人身份
func NewPersonalIdentity(id string) *PersonalIdentity {
	now := time.Now()
	return &PersonalIdentity{
		ID:        id,
		Languages: []string{"zh-CN"},
		Timezone:  "Asia/Shanghai",
		Personality: &Personality{
			Openness:          0.5,
			Conscientiousness: 0.5,
			Extraversion:      0.5,
			Agreeableness:     0.5,
			Neuroticism:       0.5,
			CustomTraits:      make(map[string]float64),
			SpeakingTone:      string(ToneCasual),
			ResponseLength:    string(LengthModerate),
			EmojiUsage:        0.3,
		},
		ValueSystem: &ValueSystem{
			Privacy:        0.7,
			Efficiency:     0.6,
			Health:         0.7,
			Family:         0.8,
			Career:         0.6,
			Entertainment:  0.5,
			Learning:       0.6,
			Social:         0.5,
			Finance:        0.6,
			Environment:    0.5,
			RiskTolerance:  0.4,
			Impulsiveness:  0.3,
			Patience:       0.6,
			CustomValues:   make(map[string]float64),
		},
		Interests:    []Interest{},
		VoiceProfile: NewDefaultVoiceProfile(),
		WritingStyle: NewDefaultWritingStyle(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewDefaultVoiceProfile 创建默认语音配置
func NewDefaultVoiceProfile() *VoiceProfile {
	now := time.Now()
	return &VoiceProfile{
		VoiceType:     string(VoiceTypePreset),
		Pitch:         1.0,
		Speed:         1.0,
		Volume:        0.8,
		Tone:          "warm",
		EmotionLevel:  0.6,
		PausePattern:  "natural",
		EmphasisStyle: "moderate",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewDefaultWritingStyle 创建默认写作风格
func NewDefaultWritingStyle() *WritingStyle {
	return &WritingStyle{
		Formality:         0.4,
		Verbosity:         0.5,
		Humor:             0.3,
		Technicality:      0.5,
		UseEmoji:          true,
		UseGIFs:           false,
		UseMarkdown:       true,
		FrequentWords:     []string{},
		AvoidWords:        []string{},
		PreferredGreeting: "你好",
		PreferredClosing:  "祝好",
	}
}

// === 方法 ===

// UpdatePersonality 更新性格特质
func (p *PersonalIdentity) UpdatePersonality(updates map[string]float64) {
	if p.Personality == nil {
		p.Personality = &Personality{
			CustomTraits: make(map[string]float64),
		}
	}

	for key, value := range updates {
		switch key {
		case "openness":
			p.Personality.Openness = clamp01(value)
		case "conscientiousness":
			p.Personality.Conscientiousness = clamp01(value)
		case "extraversion":
			p.Personality.Extraversion = clamp01(value)
		case "agreeableness":
			p.Personality.Agreeableness = clamp01(value)
		case "neuroticism":
			p.Personality.Neuroticism = clamp01(value)
		default:
			if p.Personality.CustomTraits == nil {
				p.Personality.CustomTraits = make(map[string]float64)
			}
			p.Personality.CustomTraits[key] = clamp01(value)
		}
	}

	p.UpdatedAt = time.Now()
}

// UpdateValueSystem 更新价值观
func (p *PersonalIdentity) UpdateValueSystem(updates map[string]float64) {
	if p.ValueSystem == nil {
		p.ValueSystem = &ValueSystem{
			CustomValues: make(map[string]float64),
		}
	}

	for key, value := range updates {
		switch key {
		case "privacy":
			p.ValueSystem.Privacy = clamp01(value)
		case "efficiency":
			p.ValueSystem.Efficiency = clamp01(value)
		case "health":
			p.ValueSystem.Health = clamp01(value)
		case "family":
			p.ValueSystem.Family = clamp01(value)
		case "career":
			p.ValueSystem.Career = clamp01(value)
		case "entertainment":
			p.ValueSystem.Entertainment = clamp01(value)
		case "learning":
			p.ValueSystem.Learning = clamp01(value)
		case "social":
			p.ValueSystem.Social = clamp01(value)
		case "finance":
			p.ValueSystem.Finance = clamp01(value)
		case "environment":
			p.ValueSystem.Environment = clamp01(value)
		case "risk_tolerance":
			p.ValueSystem.RiskTolerance = clamp01(value)
		case "impulsiveness":
			p.ValueSystem.Impulsiveness = clamp01(value)
		case "patience":
			p.ValueSystem.Patience = clamp01(value)
		default:
			if p.ValueSystem.CustomValues == nil {
				p.ValueSystem.CustomValues = make(map[string]float64)
			}
			p.ValueSystem.CustomValues[key] = clamp01(value)
		}
	}

	p.UpdatedAt = time.Now()
}

// AddInterest 添加兴趣
func (p *PersonalIdentity) AddInterest(interest Interest) {
	// 检查是否已存在
	for i, existing := range p.Interests {
		if existing.ID == interest.ID || (existing.Category == interest.Category && existing.Name == interest.Name) {
			// 更新现有兴趣
			p.Interests[i] = interest
			p.UpdatedAt = time.Now()
			return
		}
	}

	// 添加新兴趣
	if interest.ID == "" {
		interest.ID = generateID()
	}
	if interest.Since.IsZero() {
		interest.Since = time.Now()
	}
	interest.LastActive = time.Now()

	p.Interests = append(p.Interests, interest)
	p.UpdatedAt = time.Now()
}

// RemoveInterest 移除兴趣
func (p *PersonalIdentity) RemoveInterest(interestID string) bool {
	for i, interest := range p.Interests {
		if interest.ID == interestID {
			p.Interests = append(p.Interests[:i], p.Interests[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetInterestByCategory 按类别获取兴趣
func (p *PersonalIdentity) GetInterestByCategory(category string) []Interest {
	var result []Interest
	for _, interest := range p.Interests {
		if interest.Category == category {
			result = append(result, interest)
		}
	}
	return result
}

// GetTopInterests 获取最热衷的兴趣
func (p *PersonalIdentity) GetTopInterests(limit int) []Interest {
	if len(p.Interests) <= limit {
		return p.Interests
	}

	// 按热衷程度排序
	sorted := make([]Interest, len(p.Interests))
	copy(sorted, p.Interests)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Level > sorted[i].Level {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:limit]
}

// GetPersonalityDescription 获取性格描述
func (p *PersonalIdentity) GetPersonalityDescription() string {
	if p.Personality == nil {
		return "性格待完善"
	}

	desc := ""

	// 开放性
	if p.Personality.Openness > 0.7 {
		desc += "富有创造力和好奇心，喜欢尝试新事物。"
	} else if p.Personality.Openness < 0.3 {
		desc += "务实稳重，偏好熟悉的事物。"
	}

	// 尽责性
	if p.Personality.Conscientiousness > 0.7 {
		desc += "做事认真负责，善于规划和执行。"
	} else if p.Personality.Conscientiousness < 0.3 {
		desc += "随性灵活，不拘泥于细节。"
	}

	// 外向性
	if p.Personality.Extraversion > 0.7 {
		desc += "外向开朗，善于社交。"
	} else if p.Personality.Extraversion < 0.3 {
		desc += "内向沉静，享受独处时光。"
	}

	// 宜人性
	if p.Personality.Agreeableness > 0.7 {
		desc += "友善亲和，乐于助人。"
	} else if p.Personality.Agreeableness < 0.3 {
		desc += "独立直接，注重效率。"
	}

	return desc
}

// GetValuePriority 获取价值观优先级排序
func (p *PersonalIdentity) GetValuePriority() []string {
	if p.ValueSystem == nil {
		return []string{}
	}

	values := map[string]float64{
		"privacy":       p.ValueSystem.Privacy,
		"efficiency":    p.ValueSystem.Efficiency,
		"health":        p.ValueSystem.Health,
		"family":        p.ValueSystem.Family,
		"career":        p.ValueSystem.Career,
		"entertainment": p.ValueSystem.Entertainment,
		"learning":      p.ValueSystem.Learning,
		"social":        p.ValueSystem.Social,
		"finance":       p.ValueSystem.Finance,
		"environment":   p.ValueSystem.Environment,
	}

	// 简单排序
	type kv struct {
		Key   string
		Value float64
	}

	var pairs []kv
	for k, v := range values {
		pairs = append(pairs, kv{k, v})
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
	for _, pair := range pairs {
		result = append(result, pair.Key)
	}

	return result
}

// === 辅助函数 ===

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
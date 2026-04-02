package identity

import (
	"time"

	"github.com/ofa/center/internal/models"
)

// Inferencer - 性格推断引擎
// 从用户行为观察中推断性格特质，支持收敛趋势
type Inferencer struct {
	// 推断规则配置
	rules *InferenceRules
	// 收敛参数
	convergenceConfig *ConvergenceConfig
}

// ConvergenceConfig - 收敛配置
type ConvergenceConfig struct {
	// 收敛速率：观察次数越多，变化越小
	BaseLearningRate   float64 `json:"base_learning_rate"`    // 基础学习率 (0-1)
	MinLearningRate    float64 `json:"min_learning_rate"`     // 最小学习率
	ConvergenceSpeed   float64 `json:"convergence_speed"`     // 收敛速度 (越大越快收敛)
	StabilityThreshold int     `json:"stability_threshold"`   // 稳定阈值（观察次数）
	MaxObservations    int     `json:"max_observations"`      // 最大观察次数
}

// DefaultConvergenceConfig 默认收敛配置
func DefaultConvergenceConfig() *ConvergenceConfig {
	return &ConvergenceConfig{
		BaseLearningRate:   0.3,
		MinLearningRate:    0.05,
		ConvergenceSpeed:   0.1,
		StabilityThreshold: 20,
		MaxObservations:    100,
	}
}

// InferenceRules - 推断规则
type InferenceRules struct {
	// 各类行为对性格特质的影响权重
	DecisionWeights   map[string]TraitImpact `json:"decision_weights"`
	InteractionWeights map[string]TraitImpact `json:"interaction_weights"`
	PreferenceWeights map[string]TraitImpact `json:"preference_weights"`
	ActivityWeights   map[string]TraitImpact `json:"activity_weights"`
}

// TraitImpact - 行为对特质的影响
type TraitImpact struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
}

// NewInferencer 创建推断引擎
func NewInferencer() *Inferencer {
	return &Inferencer{
		rules:             DefaultInferenceRules(),
		convergenceConfig: DefaultConvergenceConfig(),
	}
}

// DefaultInferenceRules 默认推断规则
func DefaultInferenceRules() *InferenceRules {
	return &InferenceRules{
		DecisionWeights: map[string]TraitImpact{
			// 决策类型 -> 性格影响
			"impulsive_purchase": {
				Neuroticism: 0.3,
				Conscientiousness: -0.2,
			},
			"careful_comparison": {
				Conscientiousness: 0.3,
				Neuroticism: -0.1,
			},
			"social_activity": {
				Extraversion: 0.4,
				Agreeableness: 0.2,
			},
			"solo_activity": {
				Extraversion: -0.3,
				Openness: 0.1,
			},
			"routine_following": {
				Conscientiousness: 0.3,
				Openness: -0.2,
			},
			"novel_trying": {
				Openness: 0.4,
				Extraversion: 0.1,
			},
			"risk_taking": {
				Openness: 0.2,
				Neuroticism: -0.1,
			},
			"cautious_approach": {
				Neuroticism: 0.2,
				Conscientiousness: 0.2,
			},
			"helping_others": {
				Agreeableness: 0.4,
				Extraversion: 0.1,
			},
			"self_focus": {
				Agreeableness: -0.2,
			},
		},
		InteractionWeights: map[string]TraitImpact{
			// 交互类型 -> 性格影响
			"frequentMessaging": {
				Extraversion: 0.3,
				Agreeableness: 0.2,
			},
			"briefResponses": {
				Extraversion: -0.2,
				Conscientiousness: 0.1,
			},
			"longConversations": {
				Extraversion: 0.3,
				Agreeableness: 0.3,
			},
			"groupChats": {
				Extraversion: 0.4,
			},
			"privateChats": {
				Extraversion: -0.2,
				Agreeableness: 0.1,
			},
			"emojiHeavy": {
				Openness: 0.2,
				Extraversion: 0.2,
			},
			"formalLanguage": {
				Conscientiousness: 0.3,
				Openness: -0.1,
			},
			"casualLanguage": {
				Openness: 0.1,
				Extraversion: 0.1,
			},
			"quickReplies": {
				Extraversion: 0.2,
				Impulsiveness: 0.2,
			},
			"delayedReplies": {
				Conscientiousness: 0.2,
				Neuroticism: 0.1,
			},
		},
		PreferenceWeights: map[string]TraitImpact{
			// 偏好类型 -> 性格影响
			"reading": {
				Openness: 0.3,
				Extraversion: -0.1,
			},
			"sports": {
				Extraversion: 0.3,
				Conscientiousness: 0.2,
			},
			"music": {
				Openness: 0.3,
			},
			"travel": {
				Openness: 0.4,
				Extraversion: 0.2,
			},
			"gaming": {
				Openness: 0.2,
				Extraversion: -0.1,
			},
			"socialEvents": {
				Extraversion: 0.4,
				Agreeableness: 0.2,
			},
			"soloHobbies": {
				Extraversion: -0.3,
				Openness: 0.2,
			},
			"structuredActivities": {
				Conscientiousness: 0.3,
				Openness: -0.1,
			},
			"spontaneousActivities": {
				Openness: 0.3,
				Conscientiousness: -0.2,
			},
			"helping": {
				Agreeableness: 0.4,
			},
			"learning": {
				Openness: 0.3,
				Conscientiousness: 0.2,
			},
			"relaxing": {
				Neuroticism: -0.2,
			},
		},
		ActivityWeights: map[string]TraitImpact{
			// 活动类型 -> 性格影响
			"early riser": {
				Conscientiousness: 0.3,
			},
			"late sleeper": {
				Conscientiousness: -0.2,
				Openness: 0.1,
			},
			"regular_schedule": {
				Conscientiousness: 0.4,
				Neuroticism: -0.1,
			},
			"irregular_schedule": {
				Conscientiousness: -0.3,
				Openness: 0.2,
			},
			"exercise_regular": {
				Conscientiousness: 0.3,
				Neuroticism: -0.2,
			},
			"exercise_rare": {
				Conscientiousness: -0.1,
			},
			"multitasking": {
				Openness: 0.2,
				Conscientiousness: -0.1,
			},
			"focused_work": {
				Conscientiousness: 0.3,
			},
			"exploring_new": {
				Openness: 0.4,
			},
			"staying_familiar": {
				Openness: -0.2,
			},
		},
	}
}

// InferFromBehavior 从行为观察推断性格
func (i *Inferencer) InferFromBehavior(observations []models.BehaviorObservation) *models.Personality {
	result := &models.Personality{
		Openness:          0.5,
		Conscientiousness: 0.5,
		Extraversion:      0.5,
		Agreeableness:     0.5,
		Neuroticism:       0.5,
		CustomTraits:      make(map[string]float64),
		SpeakingTone:      "casual",
		ResponseLength:    "moderate",
		EmojiUsage:        0.3,
	}

	if len(observations) == 0 {
		return result
	}

	// 统计各维度的影响值
	impactSum := TraitImpact{}
	count := 0

	for _, obs := range observations {
		impact := i.getImpactForObservation(obs)
		if impact != nil {
			impactSum.Openness += impact.Openness
			impactSum.Conscientiousness += impact.Conscientiousness
			impactSum.Extraversion += impact.Extraversion
			impactSum.Agreeableness += impact.Agreeableness
			impactSum.Neuroticism += impact.Neuroticism
			count++

			// 处理观察中已有的推断值
			for trait, value := range obs.Inferences {
				switch trait {
				case "openness":
					impactSum.Openness += value
				case "conscientiousness":
					impactSum.Conscientiousness += value
				case "extraversion":
					impactSum.Extraversion += value
				case "agreeableness":
					impactSum.Agreeableness += value
				case "neuroticism":
					impactSum.Neuroticism += value
				default:
					// 自定义特质
					if result.CustomTraits != nil {
						result.CustomTraits[trait] += value
					}
				}
			}
		}
	}

	// 应用推断结果（基础值 + 平均影响）
	if count > 0 {
		weight := 0.3 // 推断影响权重
		result.Openness = clamp01(0.5 + weight*impactSum.Openness/float64(count))
		result.Conscientiousness = clamp01(0.5 + weight*impactSum.Conscientiousness/float64(count))
		result.Extraversion = clamp01(0.5 + weight*impactSum.Extraversion/float64(count))
		result.Agreeableness = clamp01(0.5 + weight*impactSum.Agreeableness/float64(count))
		result.Neuroticism = clamp01(0.5 + weight*impactSum.Neuroticism/float64(count))

		// 自定义特质也需要平均
		for trait, value := range result.CustomTraits {
			result.CustomTraits[trait] = clamp01(0.5 + weight*value/float64(count))
		}
	}

	// 推断说话风格
	i.inferSpeakingStyle(result, observations)

	return result
}

// getImpactForObservation 获取观察对应的特质影响
func (i *Inferencer) getImpactForObservation(obs models.BehaviorObservation) *TraitImpact {
	var weights map[string]TraitImpact

	switch obs.Type {
	case "decision":
		weights = i.rules.DecisionWeights
	case "interaction":
		weights = i.rules.InteractionWeights
	case "preference":
		weights = i.rules.PreferenceWeights
	case "activity":
		weights = i.rules.ActivityWeights
	default:
		return nil
	}

	// 从上下文中获取具体行为类型
	if obs.Context == nil {
		return nil
	}

	// 检查 outcome 是否在权重表中
	if impact, ok := weights[obs.Outcome]; ok {
		return &impact
	}

	// 检查 context 中的 subtype
	if subtype, ok := obs.Context["subtype"].(string); ok {
		if impact, ok := weights[subtype]; ok {
			return &impact
		}
	}

	// 检查 context 中的 action
	if action, ok := obs.Context["action"].(string); ok {
		if impact, ok := weights[action]; ok {
			return &impact
		}
	}

	return nil
}

// inferSpeakingStyle 推断说话风格
func (i *Inferencer) inferSpeakingStyle(personality *models.Personality, observations []models.BehaviorObservation) {
	// 基于性格特质推断
	if personality.Openness > 0.6 && personality.Extraversion > 0.6 {
		personality.SpeakingTone = "humorous"
		personality.EmojiUsage = 0.5
	} else if personality.Conscientiousness > 0.6 {
		personality.SpeakingTone = "formal"
		personality.EmojiUsage = 0.1
	} else if personality.Agreeableness > 0.7 {
		personality.SpeakingTone = "warm"
		personality.EmojiUsage = 0.4
	}

	// 从交互行为调整
	for _, obs := range observations {
		if obs.Type == "interaction" {
			if obs.Outcome == "emojiHeavy" {
				personality.EmojiUsage = clamp01(personality.EmojiUsage + 0.1)
			}
			if obs.Outcome == "formalLanguage" {
				personality.SpeakingTone = "formal"
			}
			if obs.Outcome == "casualLanguage" {
				personality.SpeakingTone = "casual"
			}
			if obs.Outcome == "briefResponses" {
				personality.ResponseLength = "brief"
			}
			if obs.Outcome == "longConversations" {
				personality.ResponseLength = "detailed"
			}
		}
	}
}

// InferValueSystemFromBehavior 从行为推断价值观
func (i *Inferencer) InferValueSystemFromBehavior(observations []models.BehaviorObservation) *models.ValueSystem {
	result := &models.ValueSystem{
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
	}

	if len(observations) == 0 {
		return result
	}

	// 从决策行为推断价值观
	for _, obs := range observations {
		if obs.Type == "decision" {
			switch obs.Outcome {
			case "privacy_concern":
				result.Privacy = clamp01(result.Privacy + 0.1)
			case "efficiency_focus":
				result.Efficiency = clamp01(result.Efficiency + 0.1)
			case "health_priority":
				result.Health = clamp01(result.Health + 0.1)
			case "family_priority":
				result.Family = clamp01(result.Family + 0.1)
			case "career_priority":
				result.Career = clamp01(result.Career + 0.1)
			case "entertainment_priority":
				result.Entertainment = clamp01(result.Entertainment + 0.1)
			case "learning_priority":
				result.Learning = clamp01(result.Learning + 0.1)
			case "social_priority":
				result.Social = clamp01(result.Social + 0.1)
			case "finance_priority":
				result.Finance = clamp01(result.Finance + 0.1)
			case "environment_priority":
				result.Environment = clamp01(result.Environment + 0.1)
			case "risk_taking":
				result.RiskTolerance = clamp01(result.RiskTolerance + 0.1)
			case "impulsive_decision":
				result.Impulsiveness = clamp01(result.Impulsiveness + 0.1)
			case "patient_decision":
				result.Patience = clamp01(result.Patience + 0.1)
			}
		}
	}

	return result
}

// InferInterestsFromActivity 从活动推断兴趣
func (i *Inferencer) InferInterestsFromActivity(activities []models.BehaviorObservation) []models.Interest {
	var interests []models.Interest

	for _, activity := range activities {
		if activity.Type != "activity" && activity.Type != "preference" {
			continue
		}

		// 从上下文提取兴趣信息
		interest := extractInterestFromActivity(activity)
		if interest != nil {
			interests = append(interests, *interest)
		}
	}

	return interests
}

// extractInterestFromActivity 从活动提取兴趣
func extractInterestFromActivity(activity models.BehaviorObservation) *models.Interest {
	if activity.Context == nil {
		return nil
	}

	interest := &models.Interest{}

	// 获取类别
	if category, ok := activity.Context["category"].(string); ok {
		interest.Category = category
	} else {
		interest.Category = "other"
	}

	// 获取名称
	if name, ok := activity.Context["name"].(string); ok {
		interest.Name = name
	} else if activity.Outcome != "" {
		interest.Name = activity.Outcome
	} else {
		return nil // 无名称则无效
	}

	// 获取热衷程度
	if level, ok := activity.Context["level"].(float64); ok {
		interest.Level = level
	} else {
		interest.Level = 0.5 // 默认中等
	}

	// 获取关键词
	if keywords, ok := activity.Context["keywords"].([]string); ok {
		interest.Keywords = keywords
	}

	// 获取描述
	if desc, ok := activity.Context["description"].(string); ok {
		interest.Description = desc
	}

	// 设置时间
	interest.Since = activity.Timestamp
	interest.LastActive = activity.Timestamp

	return interest
}

// 辅助函数
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// === MBTI 推断方法 ===

// TraitImpactWithMBTI - 包含 MBTI 维度的影响
type TraitImpactWithMBTI struct {
	// Big Five
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
	// MBTI 维度 (-1 到 1)
	MBTI_EI float64 `json:"mbti_ei"` // E(+), I(-)
	MBTI_SN float64 `json:"mbti_sn"` // N(+), S(-)
	MBTI_TF float64 `json:"mbti_tf"` // F(+), T(-)
	MBTI_JP float64 `json:"mbti_jp"` // P(+), J(-)
}

// CalculateMBTIType 从维度值计算 MBTI 类型
func CalculateMBTIType(ei, sn, tf, jp float64) string {
	result := ""
	// E-I
	if ei >= 0 {
		result += "E"
	} else {
		result += "I"
	}
	// S-N
	if sn >= 0 {
		result += "N"
	} else {
		result += "S"
	}
	// T-F
	if tf >= 0 {
		result += "F"
	} else {
		result += "T"
	}
	// J-P
	if jp >= 0 {
		result += "P"
	} else {
		result += "J"
	}
	return result
}

// CalculateMBTIConfidence 计算 MBTI 置信度
func CalculateMBTIConfidence(ei, sn, tf, jp float64) float64 {
	// 各维度的绝对值（偏离中性程度）
	absEI := abs(ei)
	absSN := abs(sn)
	absTF := abs(tf)
	absJP := abs(jp)

	// 平均偏离程度作为置信度
	avgDeviation := (absEI + absSN + absTF + absJP) / 4.0

	// 归一化到 0-1
	return clamp01(avgDeviation)
}

// BigFiveToMBTI 从 Big Five 推断 MBTI 维度
func BigFiveToMBTI(openness, conscientiousness, extraversion, agreeableness, neuroticism float64) (ei, sn, tf, jp float64) {
	// E-I: 主要由 Extraversion 决定
	ei = (extraversion - 0.5) * 2 // 映射到 -1 到 1

	// S-N: 主要由 Openness 决定
	sn = (openness - 0.5) * 2

	// T-F: 主要由 Agreeableness 决定
	// 高宜人性 -> F, 低宜人性 -> T
	tf = (agreeableness - 0.5) * 2

	// J-P: 主要由 Conscientiousness 决定
	// 高尽责性 -> J, 低尽责性 -> P
	jp = -(conscientiousness - 0.5) * 2

	return ei, sn, tf, jp
}

// UpdatePersonalityWithConvergence 带收敛机制的性格更新
func (i *Inferencer) UpdatePersonalityWithConvergence(current *models.Personality, observations []models.BehaviorObservation) *models.Personality {
	if current == nil {
		current = &models.Personality{
			Openness:          0.5,
			Conscientiousness: 0.5,
			Extraversion:      0.5,
			Agreeableness:     0.5,
			Neuroticism:       0.5,
			CustomTraits:      make(map[string]float64),
			SpeakingTone:      "casual",
			ResponseLength:    "moderate",
			EmojiUsage:        0.3,
			Tags:              []string{},
		}
	}

	// 初始化 MBTI 维度
	if current.MBTI_EI == 0 && current.MBTI_SN == 0 && current.MBTI_TF == 0 && current.MBTI_JP == 0 {
		// 从 Big Five 初始化 MBTI
		current.MBTI_EI, current.MBTI_SN, current.MBTI_TF, current.MBTI_JP =
			BigFiveToMBTI(current.Openness, current.Conscientiousness, current.Extraversion, current.Agreeableness, current.Neuroticism)
	}

	// 计算学习率（收敛趋势）
	learningRate := i.calculateLearningRate(current.ObservedCount)

	// 统计各维度的变化
	impactSum := TraitImpactWithMBTI{}
	count := 0

	for _, obs := range observations {
		impact := i.getImpactWithMBTI(obs)
		if impact != nil {
			impactSum.Openness += impact.Openness
			impactSum.Conscientiousness += impact.Conscientiousness
			impactSum.Extraversion += impact.Extraversion
			impactSum.Agreeableness += impact.Agreeableness
			impactSum.Neuroticism += impact.Neuroticism
			impactSum.MBTI_EI += impact.MBTI_EI
			impactSum.MBTI_SN += impact.MBTI_SN
			impactSum.MBTI_TF += impact.MBTI_TF
			impactSum.MBTI_JP += impact.MBTI_JP
			count++
		}
	}

	if count > 0 {
		// 应用变化（带收敛）
		avgOpenness := impactSum.Openness / float64(count)
		avgConscientiousness := impactSum.Conscientiousness / float64(count)
		avgExtraversion := impactSum.Extraversion / float64(count)
		avgAgreeableness := impactSum.Agreeableness / float64(count)
		avgNeuroticism := impactSum.Neuroticism / float64(count)

		// 更新 Big Five
		current.Openness = clamp01(current.Openness + learningRate*avgOpenness)
		current.Conscientiousness = clamp01(current.Conscientiousness + learningRate*avgConscientiousness)
		current.Extraversion = clamp01(current.Extraversion + learningRate*avgExtraversion)
		current.Agreeableness = clamp01(current.Agreeableness + learningRate*avgAgreeableness)
		current.Neuroticism = clamp01(current.Neuroticism + learningRate*avgNeuroticism)

		// 更新 MBTI 维度（带收敛）
		mbtiLR := learningRate * 0.5 // MBTI 变化更慢
		avgEI := impactSum.MBTI_EI / float64(count)
		avgSN := impactSum.MBTI_SN / float64(count)
		avgTF := impactSum.MBTI_TF / float64(count)
		avgJP := impactSum.MBTI_JP / float64(count)

		current.MBTI_EI = clampRange(current.MBTI_EI+mbtiLR*avgEI, -1, 1)
		current.MBTI_SN = clampRange(current.MBTI_SN+mbtiLR*avgSN, -1, 1)
		current.MBTI_TF = clampRange(current.MBTI_TF+mbtiLR*avgTF, -1, 1)
		current.MBTI_JP = clampRange(current.MBTI_JP+mbtiLR*avgJP, -1, 1)
	}

	// 计算 MBTI 类型和置信度
	current.MBTIType = CalculateMBTIType(current.MBTI_EI, current.MBTI_SN, current.MBTI_TF, current.MBTI_JP)
	current.MBTIConfidence = CalculateMBTIConfidence(current.MBTI_EI, current.MBTI_SN, current.MBTI_TF, current.MBTI_JP)

	// 更新统计
	current.ObservedCount++
	current.LastInferredAt = time.Now()
	current.StabilityScore = i.calculateStability(current.ObservedCount)

	// 生成性格标签
	current.Tags = i.generatePersonalityTags(current)

	// 推断说话风格
	i.inferSpeakingStyle(current, observations)

	return current
}

// calculateLearningRate 计算学习率（收敛趋势）
func (i *Inferencer) calculateLearningRate(observedCount int) float64 {
	cfg := i.convergenceConfig

	// 观察次数越多，学习率越低（趋于稳定）
	// 使用指数衰减
	decay := 1.0 / (1.0 + float64(observedCount)*cfg.ConvergenceSpeed)
	rate := cfg.BaseLearningRate * decay

	// 确保不低于最小学习率
	if rate < cfg.MinLearningRate {
		rate = cfg.MinLearningRate
	}

	return rate
}

// calculateStability 计算稳定度
func (i *Inferencer) calculateStability(observedCount int) float64 {
	cfg := i.convergenceConfig

	// 观察次数越多越稳定
	stability := float64(observedCount) / float64(cfg.StabilityThreshold)
	if stability > 1.0 {
		stability = 1.0
	}

	return stability
}

// getImpactWithMBTI 获取包含 MBTI 的影响
func (i *Inferencer) getImpactWithMBTI(obs models.BehaviorObservation) *TraitImpactWithMBTI {
	impact := i.getImpactForObservation(obs)
	if impact == nil {
		return nil
	}

	result := &TraitImpactWithMBTI{
		Openness:          impact.Openness,
		Conscientiousness: impact.Conscientiousness,
		Extraversion:      impact.Extraversion,
		Agreeableness:     impact.Agreeableness,
		Neuroticism:       impact.Neuroticism,
	}

	// 从 Big Five 影响推断 MBTI 维度变化
	// E-I: Extraversion 影响
	if impact.Extraversion != 0 {
		result.MBTI_EI = impact.Extraversion * 2
	}
	// S-N: Openness 影响
	if impact.Openness != 0 {
		result.MBTI_SN = impact.Openness * 2
	}
	// T-F: Agreeableness 影响
	if impact.Agreeableness != 0 {
		result.MBTI_TF = impact.Agreeableness * 2
	}
	// J-P: Conscientiousness 影响（反向）
	if impact.Conscientiousness != 0 {
		result.MBTI_JP = -impact.Conscientiousness * 2
	}

	return result
}

// generatePersonalityTags 生成性格标签
func (i *Inferencer) generatePersonalityTags(p *models.Personality) []string {
	tags := []string{}

	// MBTI 类型标签
	if p.MBTIType != "" {
		tags = append(tags, p.MBTIType)
		if name, ok := models.MBTITypeNames[models.MBTIType(p.MBTIType)]; ok {
			tags = append(tags, name)
		}
	}

	// MBTI 分组标签
	for group, types := range models.MBTIGroups {
		for _, t := range types {
			if string(t) == p.MBTIType {
				tags = append(tags, group)
				break
			}
		}
	}

	// Big Five 特质标签
	if p.Openness > 0.7 {
		tags = append(tags, "开放创新")
	} else if p.Openness < 0.3 {
		tags = append(tags, "务实稳重")
	}

	if p.Conscientiousness > 0.7 {
		tags = append(tags, "严谨负责")
	} else if p.Conscientiousness < 0.3 {
		tags = append(tags, "随性灵活")
	}

	if p.Extraversion > 0.7 {
		tags = append(tags, "外向开朗")
	} else if p.Extraversion < 0.3 {
		tags = append(tags, "内向沉稳")
	}

	if p.Agreeableness > 0.7 {
		tags = append(tags, "友善亲和")
	} else if p.Agreeableness < 0.3 {
		tags = append(tags, "独立直接")
	}

	if p.Neuroticism > 0.7 {
		tags = append(tags, "敏感细腻")
	} else if p.Neuroticism < 0.3 {
		tags = append(tags, "情绪稳定")
	}

	// 稳定度标签
	if p.StabilityScore > 0.8 {
		tags = append(tags, "性格稳定")
	}

	return tags
}

// GetMBTIDescription 获取 MBTI 描述
func GetMBTIDescription(mbtiType string) string {
	descriptions := map[string]string{
		"INTJ": "富有想象力和战略性的思想家，一切皆在计划之中。",
		"INTP": "具有创造力的发明家，对知识有着永不满足的渴望。",
		"ENTJ": "大胆、富有想象力的领导者，总能找到解决方法。",
		"ENTP": "聪明好奇的思想家，无法抗拒智力挑战。",
		"INFJ": "安静而神秘，但能深刻启发和感染他人。",
		"INFP": "诗意、善良的利他主义者，总是渴望帮助善行。",
		"ENFJ": "富有魅力的领袖，能够激励听众。",
		"ENFP": "热情、有创造力的社交达人，总能找到微笑的理由。",
		"ISTJ": "可靠且勤奋，非常注重义务。",
		"ISFJ": "非常敬业且善良，守护者传统价值观。",
		"ESTJ": "出色的管理者，在管理事务方面无与伦比。",
		"ESFJ": "极有同情心，爱交际，总是乐于助人。",
		"ISTP": "大胆而实际的实验家，善于使用各种工具。",
		"ISFP": "灵活而有魅力的艺术家，时刻准备探索新事物。",
		"ESTP": "聪明、精力充沛，享受生活边缘的冒险家。",
		"ESFP": "自发、精力充沛的娱乐达人，生活从不无聊。",
	}
	if desc, ok := descriptions[mbtiType]; ok {
		return desc
	}
	return "性格独特，难以简单归类。"
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func clampRange(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
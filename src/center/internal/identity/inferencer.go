package identity

import (
	"github.com/ofa/center/internal/models"
)

// Inferencer - 性格推断引擎
// 从用户行为观察中推断性格特质
type Inferencer struct {
	// 推断规则配置
	rules *InferenceRules
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
		rules: DefaultInferenceRules(),
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
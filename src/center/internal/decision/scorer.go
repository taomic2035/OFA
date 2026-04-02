package decision

import (
	"github.com/ofa/center/internal/models"
)

// Scorer - 评分器
type Scorer struct {
	valueWeights map[string]float64
}

// NewScorer 创建评分器
func NewScorer() *Scorer {
	return &Scorer{
		valueWeights: DefaultValueWeights(),
	}
}

// DefaultValueWeights 默认价值观权重
func DefaultValueWeights() map[string]float64 {
	return map[string]float64{
		"efficiency": 0.15,
		"health":     0.15,
		"finance":    0.15,
		"privacy":    0.10,
		"family":     0.10,
		"career":     0.10,
		"entertainment": 0.05,
		"learning":   0.10,
		"social":     0.05,
		"environment": 0.05,
	}
}

// Score 计算选项得分
func (s *Scorer) Score(option *models.DecisionOption, ctx *models.DecisionContext, scenario *models.DecisionScenario) (float64, map[string]float64) {
	breakdown := make(map[string]float64)
	totalScore := 0.0
	totalWeight := 0.0

	// 1. 基于场景规则评分
	for _, rule := range scenario.ScoringRules {
		score := s.scoreByRule(option, rule)
		weight := rule.Weight

		// 根据价值观调整权重
		if ctx.ValueSystem != nil {
			adjustedWeight := s.adjustWeightByValues(weight, rule.ID, ctx.ValueSystem)
			weight = adjustedWeight
		}

		breakdown[rule.ID] = score * weight
		totalScore += score * weight
		totalWeight += weight
	}

	// 2. 基于偏好评分
	if ctx.ActivePreferences != nil {
		prefScore := s.scoreByPreferences(option, ctx.ActivePreferences)
		breakdown["preference_match"] = prefScore * 0.2
		totalScore += prefScore * 0.2
		totalWeight += 0.2
	}

	// 3. 基于兴趣评分
	if len(ctx.Interests) > 0 {
		interestScore := s.scoreByInterests(option, ctx.Interests)
		breakdown["interest_match"] = interestScore * 0.1
		totalScore += interestScore * 0.1
		totalWeight += 0.1
	}

	// 4. 根据性格特质微调
	if ctx.Personality != nil {
		personalityBonus := s.scoreByPersonality(option, ctx.Personality)
		breakdown["personality_fit"] = personalityBonus * 0.05
		totalScore += personalityBonus * 0.05
		totalWeight += 0.05
	}

	// 归一化
	if totalWeight > 0 {
		totalScore = totalScore / totalWeight
	}

	return totalScore, breakdown
}

// scoreByRule 按规则评分
func (s *Scorer) scoreByRule(option *models.DecisionOption, rule models.ScoringRule) float64 {
	if option.Attributes == nil {
		return 0.5
	}

	attrValue, ok := option.Attributes[rule.Attribute]
	if !ok {
		return 0.5
	}

	// 检查值匹配
	switch v := attrValue.(type) {
	case string:
		if v == rule.ValueMatch {
			return rule.ScoreIfMatch
		}
		return rule.ScoreIfNoMatch
	case float64:
		// 数值类型：假设值越高分数越高
		return v // 假设已经是归一化的
	case int:
		return float64(v) / 100.0 // 假设需要归一化
	case bool:
		if v {
			return rule.ScoreIfMatch
		}
		return rule.ScoreIfNoMatch
	default:
		return 0.5
	}
}

// adjustWeightByValues 根据价值观调整权重
func (s *Scorer) adjustWeightByValues(baseWeight float64, ruleID string, vs *models.ValueSystem) float64 {
	// 根据规则 ID 和价值观映射
	valueMapping := map[string]string{
		"health":  "health",
		"price":   "finance",
		"cost":    "finance",
		"quality": "efficiency",
		"time":    "efficiency",
		"rating":  "efficiency",
	}

	if valueKey, ok := valueMapping[ruleID]; ok {
		valueScore := getValueScore(vs, valueKey)
		// 价值观分数越高，权重越大
		return baseWeight * (0.5 + valueScore)
	}

	return baseWeight
}

// getValueScore 获取价值观分数
func getValueScore(vs *models.ValueSystem, key string) float64 {
	switch key {
	case "privacy":
		return vs.Privacy
	case "efficiency":
		return vs.Efficiency
	case "health":
		return vs.Health
	case "family":
		return vs.Family
	case "career":
		return vs.Career
	case "entertainment":
		return vs.Entertainment
	case "learning":
		return vs.Learning
	case "social":
		return vs.Social
	case "finance":
		return vs.Finance
	case "environment":
		return vs.Environment
	default:
		return 0.5
	}
}

// scoreByPreferences 基于偏好评分
func (s *Scorer) scoreByPreferences(option *models.DecisionOption, prefs map[string]interface{}) float64 {
	if option.Attributes == nil {
		return 0.5
	}

	matchCount := 0
	totalPrefs := len(prefs)

	for key, prefValue := range prefs {
		if attrValue, ok := option.Attributes[key]; ok {
			if attrValue == prefValue {
				matchCount++
			}
		}
	}

	if totalPrefs == 0 {
		return 0.5
	}

	return float64(matchCount) / float64(totalPrefs)
}

// scoreByInterests 基于兴趣评分
func (s *Scorer) scoreByInterests(option *models.DecisionOption, interests []models.Interest) float64 {
	if option.Attributes == nil {
		return 0.5
	}

	// 检查选项标签或名称是否与兴趣匹配
	optionName := option.Name
	optionTags := []string{}
	if tags, ok := option.Attributes["tags"].([]string); ok {
		optionTags = tags
	}

	matchScore := 0.0
	totalLevel := 0.0

	for _, interest := range interests {
		totalLevel += interest.Level

		// 检查名称匹配
		if containsKeyword(optionName, interest.Keywords) {
			matchScore += interest.Level
		}

		// 检查标签匹配
		for _, tag := range optionTags {
			if containsKeyword(tag, interest.Keywords) {
				matchScore += interest.Level * 0.5
			}
		}
	}

	if totalLevel == 0 {
		return 0.5
	}

	return matchScore / totalLevel
}

// scoreByPersonality 基于性格评分
func (s *Scorer) scoreByPersonality(option *models.DecisionOption, personality *models.Personality) float64 {
	if option.Attributes == nil {
		return 0.5
	}

	score := 0.5

	// 开放性高 -> 偏好新选项
	if personality.Openness > 0.6 {
		if isNew, ok := option.Attributes["is_new"].(bool); ok && isNew {
			score += 0.2
		}
	}

	// 尽责性高 -> 偏好可靠选项
	if personality.Conscientiousness > 0.6 {
		if rating, ok := option.Attributes["rating"].(float64); ok && rating > 0.8 {
			score += 0.1
		}
	}

	// 外向性 -> 偏好社交相关选项
	if personality.Extraversion > 0.6 {
		if isSocial, ok := option.Attributes["is_social"].(bool); ok && isSocial {
			score += 0.1
		}
	}

	return score
}

// containsKeyword 检查是否包含关键词
func containsKeyword(text string, keywords []string) bool {
	for _, kw := range keywords {
		if len(text) >= len(kw) {
			for i := 0; i <= len(text)-len(kw); i++ {
				if text[i:i+len(kw)] == kw {
					return true
				}
			}
		}
	}
	return false
}
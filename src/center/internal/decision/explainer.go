package decision

import (
	"fmt"
	"strings"

	"github.com/ofa/center/internal/models"
)

// Explainer - 决策解释生成器
type Explainer struct{}

// NewExplainer 创建解释器
func NewExplainer() *Explainer {
	return &Explainer{}
}

// Explain 生成决策解释
func (e *Explainer) Explain(decision *models.Decision, ctx *models.DecisionContext, scenario *models.DecisionScenario) string {
	if decision.SelectedOption == nil {
		return "No option was selected."
	}

	var parts []string

	// 1. 说明选择结果
	parts = append(parts, e.explainSelection(decision.SelectedOption))

	// 2. 说明主要评分因素
	if len(decision.SelectedOption.ScoreBreakdown) > 0 {
		topFactors := e.getTopFactors(decision.SelectedOption.ScoreBreakdown)
		parts = append(parts, e.explainFactors(topFactors))
	}

	// 3. 说明价值观影响
	if len(decision.AppliedValues) > 0 {
		parts = append(parts, e.explainValues(decision.AppliedValues, ctx))
	}

	// 4. 说明偏好影响
	if len(decision.AppliedPreferences) > 0 {
		parts = append(parts, e.explainPreferences(decision.AppliedPreferences))
	}

	// 5. 与备选比较
	if len(decision.Options) > 1 {
		parts = append(parts, e.explainComparison(decision))
	}

	return strings.Join(parts, " ")
}

// explainSelection 解释选择
func (e *Explainer) explainSelection(option *models.DecisionOption) string {
	return fmt.Sprintf("选择了「%s」，", option.Name)
}

// explainFactors 解释评分因素
func (e *Explainer) explainFactors(factors map[string]float64) string {
	if len(factors) == 0 {
		return ""
	}

	factorNames := map[string]string{
		"health":           "健康因素",
		"price":            "价格因素",
		"cost":             "成本因素",
		"quality":          "质量因素",
		"time":             "时间因素",
		"rating":           "评分因素",
		"preference_match": "偏好匹配",
		"interest_match":   "兴趣匹配",
		"personality_fit":  "性格契合",
	}

	var descriptions []string
	for factor, score := range factors {
		name := factorNames[factor]
		if name == "" {
			name = factor
		}
		if score > 0.7 {
			descriptions = append(descriptions, fmt.Sprintf("%s表现优秀", name))
		} else if score > 0.5 {
			descriptions = append(descriptions, fmt.Sprintf("%s表现良好", name))
		}
	}

	if len(descriptions) == 0 {
		return ""
	}

	return fmt.Sprintf("主要因为%s。", strings.Join(descriptions, "、"))
}

// explainValues 解释价值观影响
func (e *Explainer) explainValues(values []string, ctx *models.DecisionContext) string {
	if ctx.ValueSystem == nil {
		return ""
	}

	valueNames := map[string]string{
		"efficiency":    "效率",
		"health":        "健康",
		"finance":       "财务",
		"privacy":       "隐私",
		"family":        "家庭",
		"career":        "事业",
		"entertainment": "娱乐",
		"learning":      "学习",
		"social":        "社交",
		"environment":   "环保",
	}

	var important []string
	for _, v := range values {
		score := getValueScore(ctx.ValueSystem, v)
		if score > 0.6 {
			name := valueNames[v]
			if name != "" {
				important = append(important, name)
			}
		}
	}

	if len(important) == 0 {
		return ""
	}

	return fmt.Sprintf("考虑到您重视%s，", strings.Join(important, "和"))
}

// explainPreferences 解释偏好影响
func (e *Explainer) explainPreferences(prefs []string) string {
	if len(prefs) == 0 {
		return ""
	}

	return fmt.Sprintf("结合您的历史偏好（%s），", strings.Join(prefs, "、"))
}

// explainComparison 解释比较
func (e *Explainer) explainComparison(decision *models.Decision) string {
	if len(decision.Options) < 2 {
		return ""
	}

	selected := decision.SelectedOption
	second := decision.Options[1]

	gap := selected.Score - second.Score
	if gap < 0.05 {
		return fmt.Sprintf("与「%s」得分接近，选择较为艰难。", second.Name)
	} else if gap < 0.15 {
		return fmt.Sprintf("相比「%s」略胜一筹。", second.Name)
	} else {
		return fmt.Sprintf("明显优于备选「%s」。", second.Name)
	}
}

// getTopFactors 获取主要评分因素
func (e *Explainer) getTopFactors(breakdown map[string]float64) map[string]float64 {
	top := make(map[string]float64)

	for factor, score := range breakdown {
		if score > 0.5 {
			top[factor] = score
		}
	}

	// 只保留前3个
	if len(top) > 3 {
		// 简单排序取前3
		type kv struct {
			Key   string
			Value float64
		}
		var pairs []kv
		for k, v := range top {
			pairs = append(pairs, kv{k, v})
		}
		for i := 0; i < len(pairs)-1; i++ {
			for j := i + 1; j < len(pairs); j++ {
				if pairs[j].Value > pairs[i].Value {
					pairs[i], pairs[j] = pairs[j], pairs[i]
				}
			}
		}
		top = make(map[string]float64)
		for i := 0; i < 3 && i < len(pairs); i++ {
			top[pairs[i].Key] = pairs[i].Value
		}
	}

	return top
}

// GenerateSummary 生成决策摘要
func (e *Explainer) GenerateSummary(decision *models.Decision) string {
	if decision == nil || decision.SelectedOption == nil {
		return ""
	}

	summary := fmt.Sprintf("在%s场景中，", decision.Scenario)

	if decision.AutoDecided {
		summary += "系统自动"
	} else {
		summary += "您"
	}

	summary += fmt.Sprintf("选择了「%s」。", decision.SelectedOption.Name)

	if decision.Outcome != "" {
		outcomeText := map[string]string{
			"satisfied":   "结果令人满意",
			"neutral":     "结果一般",
			"unsatisfied": "结果不太理想",
		}
		if text, ok := outcomeText[decision.Outcome]; ok {
			summary += text + "。"
		}
	}

	return summary
}
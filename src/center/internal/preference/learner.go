package preference

import (
	"time"

	"github.com/ofa/center/internal/models"
)

// Learner - 偏好学习引擎
type Learner struct {
	rules *LearningRules
}

// LearningRules - 学习规则
type LearningRules struct {
	// 各类行为对偏好置信度的影响
	ChoiceWeight    float64 `json:"choice_weight"`     // 选择行为权重
	FeedbackWeight  float64 `json:"feedback_weight"`   // 反馈权重
	BehaviorWeight  float64 `json:"behavior_weight"`   // 行为权重
	SkipWeight      float64 `json:"skip_weight"`       // 跳过权重

	// 置信度阈值
	MinConfidence   float64 `json:"min_confidence"`    // 最低置信度
	HighConfidence  float64 `json:"high_confidence"`   // 高置信度阈值

	// 学习次数阈值
	MinObservations int     `json:"min_observations"`  // 最少观察次数
}

// NewLearner 创建学习引擎
func NewLearner() *Learner {
	return &Learner{
		rules: DefaultLearningRules(),
	}
}

// DefaultLearningRules 默认学习规则
func DefaultLearningRules() *LearningRules {
	return &LearningRules{
		ChoiceWeight:    0.3,
		FeedbackWeight:  0.4,
		BehaviorWeight:  0.2,
		SkipWeight:      -0.1,
		MinConfidence:   0.3,
		HighConfidence:  0.8,
		MinObservations: 3,
	}
}

// Learn 从学习事件学习偏好
func (l *Learner) Learn(event *models.PreferenceLearningEvent) *models.Preference {
	if event == nil {
		return nil
	}

	var pref *models.Preference

	switch event.Type {
	case "choice":
		pref = l.learnFromChoice(event)
	case "feedback":
		pref = l.learnFromFeedback(event)
	case "behavior":
		pref = l.learnFromBehavior(event)
	case "skip":
		pref = l.learnFromSkip(event)
	}

	return pref
}

// learnFromChoice 从选择学习
func (l *Learner) learnFromChoice(event *models.PreferenceLearningEvent) *models.Preference {
	if event.Selected == nil {
		return nil
	}

	// 创建偏好
	pref := models.NewPreference(event.UserID, event.Category, "selected", event.Selected)
	pref.Source = string(models.PrefSourceImplicit)
	pref.Confidence = l.rules.ChoiceWeight
	pref.SourceEvent = event.ID

	// 如果有多个选项，说明有比较
	if len(event.Options) > 1 {
		pref.Notes = "Learned from choice among multiple options"
	}

	return pref
}

// learnFromFeedback 从反馈学习
func (l *Learner) learnFromFeedback(event *models.PreferenceLearningEvent) *models.Preference {
	if event.Feedback == "" {
		return nil
	}

	// 根据反馈类型推断偏好
	pref := models.NewPreference(event.UserID, event.Category, "feedback", event.Feedback)
	pref.Source = string(models.PrefSourceExplicit)
	pref.Confidence = l.rules.FeedbackWeight
	pref.SourceEvent = event.ID

	// 根据情感倾向调整置信度
	if event.Sentiment > 0 {
		pref.Confidence += event.Sentiment * 0.2
	} else if event.Sentiment < 0 {
		pref.Confidence += event.Sentiment * 0.2 // 负面反馈降低置信度
	}

	if pref.Confidence < 0 {
		pref.Confidence = 0
	}
	if pref.Confidence > 1 {
		pref.Confidence = 1
	}

	return pref
}

// learnFromBehavior 从行为学习
func (l *Learner) learnFromBehavior(event *models.PreferenceLearningEvent) *models.Preference {
	if event.Selected == nil {
		return nil
	}

	pref := models.NewPreference(event.UserID, event.Category, "preferred", event.Selected)
	pref.Source = string(models.PrefSourceImplicit)
	pref.Confidence = l.rules.BehaviorWeight
	pref.SourceEvent = event.ID

	return pref
}

// learnFromSkip 从跳过学习
func (l *Learner) learnFromSkip(event *models.PreferenceLearningEvent) *models.Preference {
	if event.Selected == nil {
		return nil
	}

	pref := models.NewPreference(event.UserID, event.Category, "avoided", event.Selected)
	pref.Source = string(models.PrefSourceImplicit)
	pref.Confidence = -l.rules.SkipWeight // 跳过表示不喜欢
	pref.SourceEvent = event.ID

	return pref
}

// InferFromEvents 从多个事件推断偏好
func (l *Learner) InferFromEvents(userID, category string, events []*models.PreferenceLearningEvent) []*models.Preference {
	if len(events) < l.rules.MinObservations {
		return nil
	}

	// 统计选择频率
	choiceCount := make(map[interface{}]int)
	choiceSentiment := make(map[interface{}]float64)

	for _, event := range events {
		if event.Selected == nil {
			continue
		}

		key := event.Selected
		choiceCount[key]++
		choiceSentiment[key] += event.Sentiment
	}

	// 找出高频选择
	var prefs []*models.Preference
	totalEvents := float64(len(events))

	for selected, count := range choiceCount {
		frequency := float64(count) / totalEvents
		avgSentiment := choiceSentiment[selected] / float64(count)

		// 高频 + 正面情感 = 强偏好
		confidence := frequency * 0.5
		if avgSentiment > 0 {
			confidence += avgSentiment * 0.3
		}

		if confidence >= l.rules.MinConfidence {
			pref := models.NewPreference(userID, category, "preferred", selected)
			pref.Source = string(models.PrefSourceLearned)
			pref.Confidence = confidence
			pref.SourceEvent = "inferred_from_" + string(rune(len(events))) + "_events"

			// 添加上下文条件
			pref.Notes = "Inferred from " + string(rune(count)) + " observations"

			prefs = append(prefs, pref)
		}
	}

	return prefs
}

// ResolveConflicts 解决偏好冲突
func (l *Learner) ResolveConflicts(conflicts []*models.Preference) *models.Preference {
	if len(conflicts) == 0 {
		return nil
	}

	if len(conflicts) == 1 {
		return conflicts[0]
	}

	// 选择置信度最高的
	best := conflicts[0]
	for _, p := range conflicts[1:] {
		if p.Confidence > best.Confidence {
			best = p
		}
		// 置信度相同时，选择更新的
		if p.Confidence == best.Confidence && p.UpdatedAt.After(best.UpdatedAt) {
			best = p
		}
	}

	return best
}

// GenerateRecommendations 生成推荐
func (l *Learner) GenerateRecommendations(userID, category string, preferences []*models.Preference, context map[string]interface{}) []*models.Recommendation {
	var recommendations []*models.Recommendation

	for _, pref := range preferences {
		if pref.Category != category {
			continue
		}

		if pref.IsExpired() {
			continue
		}

		// 检查条件匹配
		if len(pref.Conditions) > 0 && !pref.MatchesConditions(context) {
			continue
		}

		// 创建推荐
		rec := &models.Recommendation{
			ID:        generateRecID(),
			UserID:    userID,
			Category:  category,
			Item:      pref.Value,
			Score:     pref.Confidence,
			Reason:    "Based on your preferences",
			BasedOn:   []string{pref.ID},
			Context:   context,
			CreatedAt: time.Now(),
		}

		// 设置过期时间（1小时）
		expiresAt := time.Now().Add(time.Hour)
		rec.ExpiresAt = expiresAt

		recommendations = append(recommendations, rec)
	}

	// 按分数排序
	sortRecommendations(recommendations)

	return recommendations
}

func generateRecID() string {
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

func sortRecommendations(recs []*models.Recommendation) {
	// 简单冒泡排序
	for i := 0; i < len(recs)-1; i++ {
		for j := i + 1; j < len(recs); j++ {
			if recs[j].Score > recs[i].Score {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
	}
}
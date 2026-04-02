package models

import (
	"time"
)

// === 决策系统数据模型 ===

// Decision - 决策记录
type Decision struct {
	ID              string                 `json:"id" bson:"_id"`
	UserID          string                 `json:"user_id" bson:"user_id"`

	// 决策场景
	Scenario        string                 `json:"scenario" bson:"scenario"`           // 点餐/购物/出行/娱乐
	ScenarioType    string                 `json:"scenario_type" bson:"scenario_type"` // single/recurring/batch
	Context         map[string]interface{} `json:"context" bson:"context"`

	// 决策选项
	Options         []DecisionOption       `json:"options" bson:"options"`
	SelectedIndex   int                    `json:"selected_index" bson:"selected_index"`
	SelectedOption  *DecisionOption        `json:"selected_option" bson:"selected_option"`
	SelectedReason  string                 `json:"selected_reason" bson:"selected_reason"`

	// 决策依据
	AppliedValues   []string               `json:"applied_values" bson:"applied_values"`
	AppliedRules    []string               `json:"applied_rules" bson:"applied_rules"`
	AppliedPreferences []string            `json:"applied_preferences" bson:"applied_preferences"`

	// 决策过程
	ScoreDetails    map[string]float64     `json:"score_details" bson:"score_details"` // 各选项得分详情
	Ranking         []int                  `json:"ranking" bson:"ranking"`            // 选项排名

	// 结果反馈
	Outcome         string                 `json:"outcome" bson:"outcome"`            // satisfied/neutral/unsatisfied
	UserFeedback    string                 `json:"user_feedback" bson:"user_feedback"`
	OutcomeScore    float64                `json:"outcome_score" bson:"outcome_score"` // 结果评分

	// 执行信息
	ExecutedAt      *time.Time             `json:"executed_at" bson:"executed_at"`
	CompletedAt     *time.Time             `json:"completed_at" bson:"completed_at"`

	// 元数据
	AutoDecided     bool                   `json:"auto_decided" bson:"auto_decided"`   // 是否自动决策
	Confidence      float64                `json:"confidence" bson:"confidence"`       // 决策置信度
	Tags            []string               `json:"tags" bson:"tags"`

	CreatedAt       time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" bson:"updated_at"`
}

// DecisionOption - 决策选项
type DecisionOption struct {
	ID          string                 `json:"id" bson:"id"`
	Name        string                 `json:"name" bson:"name"`
	Description string                 `json:"description" bson:"description"`
	Attributes  map[string]interface{} `json:"attributes" bson:"attributes"`
	Score       float64                `json:"score" bson:"score"`
	ScoreBreakdown map[string]float64  `json:"score_breakdown" bson:"score_breakdown"`
	Pros        []string               `json:"pros" bson:"pros"`
	Cons        []string               `json:"cons" bson:"cons"`
	Rank        int                    `json:"rank" bson:"rank"`
}

// DecisionScenario - 决策场景
type DecisionScenario struct {
	ID              string                 `json:"id" bson:"_id"`
	Name            string                 `json:"name" bson:"name"`
	Category        string                 `json:"category" bson:"category"`
	Description     string                 `json:"description" bson:"description"`
	RequiredValues  []string               `json:"required_values" bson:"required_values"`
	RequiredPrefs   []string               `json:"required_prefs" bson:"required_prefs"`
	ScoringRules    []ScoringRule          `json:"scoring_rules" bson:"scoring_rules"`
	Constraints     []Constraint           `json:"constraints" bson:"constraints"`
	Templates       []DecisionTemplate     `json:"templates" bson:"templates"`
}

// ScoringRule - 评分规则
type ScoringRule struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Attribute     string  `json:"attribute"`     // 要评估的属性
	Weight        float64 `json:"weight"`        // 权重
	ValueMatch    string  `json:"value_match"`   // 匹配的值
	ScoreIfMatch  float64 `json:"score_if_match"`
	ScoreIfNoMatch float64 `json:"score_if_no_match"`
}

// Constraint - 约束条件
type Constraint struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`        // hard/soft
	Attribute   string      `json:"attribute"`
	Operator    string      `json:"operator"`    // eq/neq/gt/lt/in/range
	Value       interface{} `json:"value"`
	Penalty     float64     `json:"penalty"`     // 违反软约束的惩罚
	Description string      `json:"description"`
}

// DecisionTemplate - 决策模板
type DecisionTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Scenario    string                 `json:"scenario"`
	Options     []DecisionOption       `json:"options"`
	Context     map[string]interface{} `json:"context"`
	Priority    int                    `json:"priority"`
}

// DecisionContext - 决策上下文（从 Identity 获取）
type DecisionContext struct {
	UserID         string                `json:"user_id"`
	Personality    *Personality          `json:"personality"`
	ValueSystem    *ValueSystem          `json:"value_system"`
	Interests      []Interest            `json:"interests"`
	SpeakingTone   string                `json:"speaking_tone"`
	ResponseLength string                `json:"response_length"`
	ValuePriority  []string              `json:"value_priority"`
	RecentDecisions []*Decision          `json:"recent_decisions"`
	ActivePreferences map[string]interface{} `json:"active_preferences"`
}

// DecisionResult - 决策结果
type DecisionResult struct {
	Decision        *Decision        `json:"decision"`
	Alternatives    []*DecisionOption `json:"alternatives"` // 备选方案
	Explanation     string           `json:"explanation"`
	Confidence      float64          `json:"confidence"`
	NeedsUserInput  bool             `json:"needs_user_input"` // 是否需要用户确认
	UncertainReason string           `json:"uncertain_reason"` // 不确定的原因
}

// DecisionQuery - 决策查询
type DecisionQuery struct {
	UserID      string   `json:"user_id"`
	Scenario    string   `json:"scenario"`
	Outcome     string   `json:"outcome"`
	AutoDecided *bool    `json:"auto_decided"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
}

// DecisionStats - 决策统计
type DecisionStats struct {
	UserID           string                   `json:"user_id"`
	TotalDecisions   int                      `json:"total_decisions"`
	AutoDecisions    int                      `json:"auto_decisions"`
	ManualDecisions  int                      `json:"manual_decisions"`
	SatisfiedCount   int                      `json:"satisfied_count"`
	UnsatisfiedCount int                      `json:"unsatisfied_count"`
	AvgOutcomeScore  float64                  `json:"avg_outcome_score"`
	CountByScenario  map[string]int           `json:"count_by_scenario"`
	TopScenarios     []string                 `json:"top_scenarios"`
	ValueUsage       map[string]int           `json:"value_usage"`      // 价值观使用频率
	PreferenceHits   map[string]int           `json:"preference_hits"`  // 偏好命中次数
	RecentDecisions  []*Decision              `json:"recent_decisions"`
}

// === 辅助方法 ===

// NewDecision 创建新决策
func NewDecision(userID, scenario string) *Decision {
	now := time.Now()
	return &Decision{
		ID:           generateDecisionID(),
		UserID:       userID,
		Scenario:     scenario,
		Context:      make(map[string]interface{}),
		Options:      []DecisionOption{},
		AppliedValues: []string{},
		AppliedRules: []string{},
		AppliedPreferences: []string{},
		ScoreDetails: make(map[string]float64),
		Tags:         []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AddOption 添加选项
func (d *Decision) AddOption(option DecisionOption) {
	d.Options = append(d.Options, option)
	d.UpdatedAt = time.Now()
}

// Select 选择选项
func (d *Decision) Select(index int, reason string) {
	if index >= 0 && index < len(d.Options) {
		d.SelectedIndex = index
		d.SelectedOption = &d.Options[index]
		d.SelectedReason = reason
		d.UpdatedAt = time.Now()
	}
}

// SetOutcome 设置结果
func (d *Decision) SetOutcome(outcome string, feedback string, score float64) {
	d.Outcome = outcome
	d.UserFeedback = feedback
	d.OutcomeScore = score
	now := time.Now()
	d.CompletedAt = &now
	d.UpdatedAt = now
}

// IsCompleted 是否完成
func (d *Decision) IsCompleted() bool {
	return d.CompletedAt != nil
}

// GetTopOptions 获取前 N 个选项
func (d *Decision) GetTopOptions(n int) []DecisionOption {
	if len(d.Options) <= n {
		return d.Options
	}

	// 按分数排序
	sorted := make([]DecisionOption, len(d.Options))
	copy(sorted, d.Options)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:n]
}

// CalculateConfidence 计算置信度
func (d *Decision) CalculateConfidence() float64 {
	if len(d.Options) == 0 {
		return 0
	}

	if d.SelectedOption == nil {
		return 0
	}

	// 最高分与次高分的差距
	topScore := d.SelectedOption.Score
	secondScore := 0.0

	for _, opt := range d.Options {
		if opt.ID != d.SelectedOption.ID && opt.Score > secondScore {
			secondScore = opt.Score
		}
	}

	// 差距越大，置信度越高
	gap := topScore - secondScore
	if gap > 0.5 {
		return 0.9
	} else if gap > 0.3 {
		return 0.7
	} else if gap > 0.1 {
		return 0.5
	}

	return 0.3
}

func generateDecisionID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}
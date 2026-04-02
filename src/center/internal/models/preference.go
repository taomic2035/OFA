package models

import (
	"time"
)

// === 偏好系统数据模型 ===

// Preference - 偏好记录
type Preference struct {
	ID          string      `json:"id" bson:"_id"`
	UserID      string      `json:"user_id" bson:"user_id"`
	Category    string      `json:"category" bson:"category"`     // food/shop/travel/music/app...
	Key         string      `json:"key" bson:"key"`               // preferred_tea_shop
	Value       interface{} `json:"value" bson:"value"`           // 喜茶
	ValueType   string      `json:"value_type" bson:"value_type"` // string/number/bool/array/object
	Confidence  float64     `json:"confidence" bson:"confidence"` // 置信度 (0-1)
	Source      string      `json:"source" bson:"source"`         // explicit/implicit/learned/inferred
	Context     string      `json:"context" bson:"context"`       // 偏好上下文
	Conditions  []Condition `json:"conditions" bson:"conditions"` // 条件偏好

	// 统计
	AccessCount int       `json:"access_count" bson:"access_count"`
	ConfirmCount int      `json:"confirm_count" bson:"confirm_count"` // 确认次数
	RejectCount int       `json:"reject_count" bson:"reject_count"`   // 拒绝次数

	// 时间
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	LastUsed    time.Time `json:"last_used" bson:"last_used"`
	ExpiresAt   *time.Time `json:"expires_at" bson:"expires_at"`

	// 元数据
	Tags        []string  `json:"tags" bson:"tags"`
	SourceApp   string    `json:"source_app" bson:"source_app"`
	SourceEvent string    `json:"source_event" bson:"source_event"`
	Notes       string    `json:"notes" bson:"notes"`
}

// Condition - 偏好条件
type Condition struct {
	Type     string      `json:"type"`     // time/location/weather/context
	Key      string      `json:"key"`      // 具体条件键
	Value    interface{} `json:"value"`    // 条件值
	Operator string      `json:"operator"` // eq/neq/gt/lt/in/contains
}

// PreferenceCategory - 偏好类别
type PreferenceCategory string

const (
	PrefCategoryFood       PreferenceCategory = "food"        // 饮食偏好
	PrefCategoryShop       PreferenceCategory = "shop"        // 购物偏好
	PrefCategoryTravel     PreferenceCategory = "travel"      // 出行偏好
	PrefCategoryMusic      PreferenceCategory = "music"       // 音乐偏好
	PrefCategoryApp        PreferenceCategory = "app"         // 应用偏好
	PrefCategoryContent    PreferenceCategory = "content"     // 内容偏好
	PrefCategorySocial     PreferenceCategory = "social"      // 社交偏好
	PrefCategoryWork       PreferenceCategory = "work"        // 工作偏好
	PrefCategoryHealth     PreferenceCategory = "health"      // 健康偏好
	PrefCategoryFinance    PreferenceCategory = "finance"     // 财务偏好
	PrefCategoryEntertainment PreferenceCategory = "entertainment" // 娱乐偏好
	PrefCategoryOther      PreferenceCategory = "other"       // 其他
)

// PreferenceSource - 偏好来源
type PreferenceSource string

const (
	PrefSourceExplicit PreferenceSource = "explicit" // 用户明确指定
	PrefSourceImplicit PreferenceSource = "implicit" // 从行为推断
	PrefSourceLearned  PreferenceSource = "learned"  // 从历史学习
	PrefSourceInferred PreferenceSource = "inferred" // 从其他偏好推断
)

// PreferenceConflict - 偏好冲突
type PreferenceConflict struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	Preference1  *Preference  `json:"preference_1"`
	Preference2  *Preference  `json:"preference_2"`
	ConflictType string       `json:"conflict_type"` // value/mutual_exclusive/priority
	Resolution   string       `json:"resolution"`    // keep_first/keep_second/merge/ask_user
	ResolvedAt   time.Time    `json:"resolved_at"`
}

// PreferenceQuery - 偏好查询
type PreferenceQuery struct {
	UserID     string   `json:"user_id"`
	Categories []string `json:"categories"`
	Keys       []string `json:"keys"`
	MinConfidence float64 `json:"min_confidence"`
	Source      string   `json:"source"`
	Context     string   `json:"context"`
	Tags        []string `json:"tags"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
}

// PreferenceStats - 偏好统计
type PreferenceStats struct {
	UserID          string                        `json:"user_id"`
	TotalCount      int                           `json:"total_count"`
	CountByCategory map[string]int                `json:"count_by_category"`
	CountBySource   map[string]int                `json:"count_by_source"`
	AvgConfidence   float64                       `json:"avg_confidence"`
	TopPreferences  map[string][]*Preference      `json:"top_preferences"` // category -> top prefs
	RecentChanges   []*Preference                 `json:"recent_changes"`
	Conflicts       []*PreferenceConflict         `json:"conflicts"`
}

// PreferenceLearningEvent - 偏好学习事件
type PreferenceLearningEvent struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        string                 `json:"type"`        // choice/feedback/behavior/skip
	Category    string                 `json:"category"`
	Context     map[string]interface{} `json:"context"`
	Options     []interface{}          `json:"options"`     // 可选项
	Selected    interface{}            `json:"selected"`    // 选择项
	Feedback    string                 `json:"feedback"`    // 反馈
	Sentiment   float64                `json:"sentiment"`   // 情感倾向
	Timestamp   time.Time              `json:"timestamp"`
}

// Recommendation - 推荐结果
type Recommendation struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Category    string                 `json:"category"`
	Item        interface{}            `json:"item"`
	Score       float64                `json:"score"`
	Reason      string                 `json:"reason"`
	BasedOn     []string               `json:"based_on"`     // 基于的偏好ID
	Context     map[string]interface{} `json:"context"`
	ExpiresAt   time.Time              `json:"expires_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// === 辅助方法 ===

// NewPreference 创建新偏好
func NewPreference(userID, category, key string, value interface{}) *Preference {
	now := time.Now()
	return &Preference{
		ID:         generatePrefID(),
		UserID:     userID,
		Category:   category,
		Key:        key,
		Value:      value,
		ValueType:  detectValueType(value),
		Confidence: 0.5,
		Source:     string(PrefSourceExplicit),
		Tags:       []string{},
		Conditions: []Condition{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Access 访问偏好
func (p *Preference) Access() {
	p.AccessCount++
	p.LastUsed = time.Now()
	p.UpdatedAt = time.Now()
}

// Confirm 确认偏好
func (p *Preference) Confirm() {
	p.ConfirmCount++
	p.updateConfidence()
}

// Reject 拒绝偏好
func (p *Preference) Reject() {
	p.RejectCount++
	p.updateConfidence()
}

// updateConfidence 更新置信度
func (p *Preference) updateConfidence() {
	total := p.ConfirmCount + p.RejectCount
	if total > 0 {
		p.Confidence = float64(p.ConfirmCount) / float64(total)
	}
	p.UpdatedAt = time.Now()
}

// IsExpired 是否过期
func (p *Preference) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

// MatchesConditions 是否匹配条件
func (p *Preference) MatchesConditions(context map[string]interface{}) bool {
	if len(p.Conditions) == 0 {
		return true
	}

	for _, cond := range p.Conditions {
		ctxValue, ok := context[cond.Key]
		if !ok {
			return false
		}

		if !matchCondition(ctxValue, cond.Value, cond.Operator) {
			return false
		}
	}

	return true
}

// AddCondition 添加条件
func (p *Preference) AddCondition(cond Condition) {
	p.Conditions = append(p.Conditions, cond)
	p.UpdatedAt = time.Now()
}

// AddTag 添加标签
func (p *Preference) AddTag(tag string) {
	for _, t := range p.Tags {
		if t == tag {
			return
		}
	}
	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now()
}

func generatePrefID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

func detectValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64, float32, float64:
		return "number"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func matchCondition(ctxValue, condValue interface{}, operator string) bool {
	switch operator {
	case "eq":
		return ctxValue == condValue
	case "neq":
		return ctxValue != condValue
	case "gt":
		return compareNumbers(ctxValue, condValue) > 0
	case "lt":
		return compareNumbers(ctxValue, condValue) < 0
	case "in":
		arr, ok := condValue.([]interface{})
		if !ok {
			return false
		}
		for _, v := range arr {
			if v == ctxValue {
				return true
			}
		}
		return false
	case "contains":
		str, ok := ctxValue.(string)
		if !ok {
			return false
		}
		substr, ok := condValue.(string)
		if !ok {
			return false
		}
		return len(str) > 0 && len(substr) > 0 && containsSubstring(str, substr)
	default:
		return false
	}
}

func compareNumbers(a, b interface{}) int {
	aFloat := toFloat(a)
	bFloat := toFloat(b)
	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	default:
		return 0
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
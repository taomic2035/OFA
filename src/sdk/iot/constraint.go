// Package iot - 约束检查
package iot

import (
	"encoding/json"
	"sync"
	"time"
)

// ConstraintType 约束类型
type ConstraintType string

const (
	ConstraintPower    ConstraintType = "power"    // 电源约束
	ConstraintTime     ConstraintType = "time"     // 时间约束
	ConstraintSafety   ConstraintType = "safety"   // 安全约束
	ConstraintPrivacy  ConstraintType = "privacy"  // 隐私约束
	ConstraintSchedule ConstraintType = "schedule" // 调度约束
)

// ConstraintLevel 约束级别
type ConstraintLevel string

const (
	LevelWarn  ConstraintLevel = "warn"  // 警告
	LevelBlock ConstraintLevel = "block" // 阻止
	LevelLog   ConstraintLevel = "log"   // 记录
)

// ConstraintRule 约束规则
type ConstraintRule struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        ConstraintType   `json:"type"`
	Level       ConstraintLevel  `json:"level"`
	Description string           `json:"description"`
	Condition   string           `json:"condition"` // 条件表达式
	Enabled     bool             `json:"enabled"`
	CreatedAt   time.Time        `json:"created_at"`
}

// ConstraintResult 检查结果
type ConstraintResult struct {
	RuleID      string           `json:"rule_id"`
	RuleName    string           `json:"rule_name"`
	Type        ConstraintType   `json:"type"`
	Level       ConstraintLevel  `json:"level"`
	Passed      bool             `json:"passed"`
	Message     string           `json:"message"`
	Suggestion  string           `json:"suggestion,omitempty"`
}

// ConstraintReport 检查报告
type ConstraintReport struct {
	DeviceID  string              `json:"device_id"`
	CommandID string              `json:"command_id"`
	AllPassed bool                `json:"all_passed"`
	Results   []*ConstraintResult `json:"results"`
	Timestamp time.Time           `json:"timestamp"`
}

// ConstraintChecker 约束检查器
type ConstraintChecker struct {
	rules       map[string]*ConstraintRule
	customCheck map[string]func(interface{}) bool

	// 设备状态
	powerStatus  bool
	currentTime  time.Time

	mu sync.RWMutex
}

// NewConstraintChecker 创建检查器
func NewConstraintChecker() *ConstraintChecker {
	cc := &ConstraintChecker{
		rules:       make(map[string]*ConstraintRule),
		customCheck: make(map[string]func(interface{}) bool),
	}

	cc.loadDefaultRules()
	return cc
}

// AddRule 添加规则
func (cc *ConstraintChecker) AddRule(rule *ConstraintRule) {
	cc.mu.Lock()
	rule.CreatedAt = time.Now()
	cc.rules[rule.ID] = rule
	cc.mu.Unlock()
}

// RemoveRule 移除规则
func (cc *ConstraintChecker) RemoveRule(id string) {
	cc.mu.Lock()
	delete(cc.rules, id)
	delete(cc.customCheck, id)
	cc.mu.Unlock()
}

// SetPowerStatus 设置电源状态
func (cc *ConstraintChecker) SetPowerStatus(on bool) {
	cc.mu.Lock()
	cc.powerStatus = on
	cc.mu.Unlock()
}

// Check 检查命令
func (cc *ConstraintChecker) Check(deviceID, commandID, commandType string, params map[string]interface{}) *ConstraintReport {
	report := &ConstraintReport{
		DeviceID:  deviceID,
		CommandID: commandID,
		AllPassed: true,
		Timestamp: time.Now(),
	}

	cc.mu.RLock()
	rules := cc.getEnabledRules()
	cc.mu.RUnlock()

	// 检查每条规则
	for _, rule := range rules {
		result := cc.checkRule(rule, commandType, params)
		if result != nil {
			report.Results = append(report.Results, result)
			if !result.Passed && result.Level == LevelBlock {
				report.AllPassed = false
			}
		}
	}

	return report
}

// CheckProperty 检查属性变更
func (cc *ConstraintChecker) CheckProperty(deviceID, key string, value interface{}) *ConstraintReport {
	report := &ConstraintReport{
		DeviceID:  deviceID,
		CommandID: "property_" + key,
		AllPassed: true,
		Timestamp: time.Now(),
	}

	// 检查隐私字段
	privateFields := []string{"password", "token", "secret", "key"}
	for _, field := range privateFields {
		if key == field {
			report.Results = append(report.Results, &ConstraintResult{
				RuleID:   "privacy_field",
				Type:     ConstraintPrivacy,
				Level:    LevelBlock,
				Passed:   false,
				Message:  "禁止设置敏感字段: " + key,
			})
			report.AllPassed = false
			break
		}
	}

	return report
}

// SetCustomCheck 设置自定义检查
func (cc *ConstraintChecker) SetCustomCheck(id string, check func(interface{}) bool) {
	cc.mu.Lock()
	cc.customCheck[id] = check
	cc.mu.Unlock()
}

// GetRules 获取规则
func (cc *ConstraintChecker) GetRules() []*ConstraintRule {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	rules := make([]*ConstraintRule, 0, len(cc.rules))
	for _, r := range cc.rules {
		rules = append(rules, r)
	}
	return rules
}

// Export 导出规则
func (cc *ConstraintChecker) Export() json.RawMessage {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	rules := make([]*ConstraintRule, 0, len(cc.rules))
	for _, r := range cc.rules {
		rules = append(rules, r)
	}

	data, _ := json.Marshal(rules)
	return data
}

func (cc *ConstraintChecker) checkRule(rule *ConstraintRule, commandType string, params map[string]interface{}) *ConstraintResult {
	result := &ConstraintResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Type:     rule.Type,
		Level:    rule.Level,
		Passed:   true,
	}

	switch rule.Type {
	case ConstraintPower:
		// 电源约束：设备关电时的操作限制
		if commandType == "set_property" || commandType == "action" {
			// 某些操作在断电时不允许
		}

	case ConstraintTime:
		// 时间约束：特定时间段限制
		now := time.Now()
		hour := now.Hour()
		if hour >= 22 || hour < 6 {
			if commandType == "action" {
				// 深夜操作警告
				result.Passed = true
				result.Message = "深夜操作，请确认"
				result.Suggestion = "建议白天执行此操作"
				result.Level = LevelWarn
			}
		}

	case ConstraintSafety:
		// 安全约束：危险操作检查
		if commandType == "ota" {
			if params != nil {
				if url, ok := params["url"].(string); ok {
					if url == "" {
						result.Passed = false
						result.Message = "OTA地址无效"
						return result
					}
				}
			}
		}

	case ConstraintPrivacy:
		// 隐私约束：敏感数据处理
		if params != nil {
			if _, ok := params["password"]; ok {
				result.Passed = false
				result.Message = "参数包含敏感信息"
				return result
			}
		}

	case ConstraintSchedule:
		// 调度约束：操作频率限制
		// 可扩展实现
	}

	return result
}

func (cc *ConstraintChecker) getEnabledRules() []*ConstraintRule {
	rules := make([]*ConstraintRule, 0)
	for _, r := range cc.rules {
		if r.Enabled {
			rules = append(rules, r)
		}
	}
	return rules
}

func (cc *ConstraintChecker) loadDefaultRules() {
	// 电源规则
	cc.AddRule(&ConstraintRule{
		ID:          "power_off_limit",
		Name:        "断电操作限制",
		Type:        ConstraintPower,
		Level:       LevelWarn,
		Description: "设备断电时部分功能受限",
		Enabled:     true,
	})

	// 时间规则
	cc.AddRule(&ConstraintRule{
		ID:          "night_operation",
		Name:        "深夜操作提醒",
		Type:        ConstraintTime,
		Level:       LevelWarn,
		Description: "22:00-06:00 操作提醒",
		Enabled:     true,
	})

	// 安全规则
	cc.AddRule(&ConstraintRule{
		ID:          "ota_safety",
		Name:        "OTA安全检查",
		Type:        ConstraintSafety,
		Level:       LevelBlock,
		Description: "OTA升级安全验证",
		Enabled:     true,
	})

	// 隐私规则
	cc.AddRule(&ConstraintRule{
		ID:          "privacy_data",
		Name:        "隐私数据保护",
		Type:        ConstraintPrivacy,
		Level:       LevelBlock,
		Description: "敏感数据传输限制",
		Enabled:     true,
	})

	// 调度规则
	cc.AddRule(&ConstraintRule{
		ID:          "rate_limit",
		Name:        "操作频率限制",
		Type:        ConstraintSchedule,
		Level:       LevelWarn,
		Description: "防止频繁操作",
		Enabled:     true,
	})
}
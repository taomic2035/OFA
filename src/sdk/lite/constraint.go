// Package lite - 约束检查(轻量级，适合手表/手环)
package lite

import (
	"encoding/json"
	"sync"
	"time"
)

// ConstraintType 约束类型
type ConstraintType string

const (
	ConstraintPrivacy   ConstraintType = "privacy"
	ConstraintSecurity  ConstraintType = "security"
	ConstraintBattery   ConstraintType = "battery"
	ConstraintBandwidth ConstraintType = "bandwidth"
)

// ConstraintSeverity 严重程度
type ConstraintSeverity string

const (
	SeverityWarning ConstraintSeverity = "warning"
	SeverityBlock   ConstraintSeverity = "block"
)

// ConstraintRule 约束规则(简化版)
type ConstraintRule struct {
	ID          string            `json:"id"`
	Type        ConstraintType    `json:"type"`
	Severity    ConstraintSeverity `json:"severity"`
	Description string            `json:"description"`
	Field       string            `json:"field"` // 限制的字段
	Enabled     bool              `json:"enabled"`
}

// ConstraintResult 检查结果
type ConstraintResult struct {
	RuleID   string            `json:"rule_id"`
	Type     ConstraintType    `json:"type"`
	Severity ConstraintSeverity `json:"severity"`
	Passed   bool              `json:"passed"`
	Message  string            `json:"message"`
}

// ConstraintReport 检查报告
type ConstraintReport struct {
	TaskID    string             `json:"task_id"`
	AllPassed bool               `json:"all_passed"`
	Results   []*ConstraintResult `json:"results"`
	Timestamp time.Time          `json:"timestamp"`
}

// ConstraintChecker 约束检查器(轻量级)
type ConstraintChecker struct {
	rules       map[string]*ConstraintRule
	customRules map[string]func(interface{}) bool

	// 省电模式
	powerSaveMode bool
	batteryLevel  int

	mu sync.RWMutex
}

// NewConstraintChecker 创建检查器
func NewConstraintChecker() *ConstraintChecker {
	c := &ConstraintChecker{
		rules:         make(map[string]*ConstraintRule),
		customRules:   make(map[string]func(interface{}) bool),
		powerSaveMode: true,
	}
	c.loadDefaultRules()
	return c
}

// AddRule 添加规则
func (c *ConstraintChecker) AddRule(rule *ConstraintRule) {
	c.mu.Lock()
	c.rules[rule.ID] = rule
	c.mu.Unlock()
}

// RemoveRule 移除规则
func (c *ConstraintChecker) RemoveRule(id string) {
	c.mu.Lock()
	delete(c.rules, id)
	delete(c.customRules, id)
	c.mu.Unlock()
}

// SetPowerSaveMode 设置省电模式
func (c *ConstraintChecker) SetPowerSaveMode(enabled bool) {
	c.mu.Lock()
	c.powerSaveMode = enabled
	c.mu.Unlock()
}

// SetBatteryLevel 设置电量
func (c *ConstraintChecker) SetBatteryLevel(level int) {
	c.mu.Lock()
	c.batteryLevel = level
	c.mu.Unlock()
}

// Check 检查数据
func (c *ConstraintChecker) Check(taskID, skill string, data interface{}) *ConstraintReport {
	report := &ConstraintReport{
		TaskID:    taskID,
		AllPassed: true,
		Timestamp: time.Now(),
	}

	c.mu.RLock()
	rules := c.getEnabledRules()
	powerSave := c.powerSaveMode
	battery := c.batteryLevel
	c.mu.RUnlock()

	// 检查电量约束
	if powerSave && battery < 20 {
		report.Results = append(report.Results, &ConstraintResult{
			RuleID:   "battery_low",
			Type:     ConstraintBattery,
			Severity: SeverityWarning,
			Passed:   false,
			Message:  "低电量模式，部分功能受限",
		})
	}

	// 检查字段约束
	if m, ok := data.(map[string]interface{}); ok {
		for _, rule := range rules {
			if rule.Field == "" {
				continue
			}

			if val, exists := m[rule.Field]; exists {
				result := c.checkField(rule, val)
				report.Results = append(report.Results, result)
				if !result.Passed && result.Severity == SeverityBlock {
					report.AllPassed = false
				}
			}
		}
	}

	// 检查自定义规则
	for id, checker := range c.customRules {
		if !checker(data) {
			report.Results = append(report.Results, &ConstraintResult{
				RuleID:   id,
				Severity: SeverityWarning,
				Passed:   false,
				Message:  "自定义约束检查失败",
			})
		}
	}

	return report
}

// CheckField 检查单个字段
func (c *ConstraintChecker) checkField(rule *ConstraintRule, val interface{}) *ConstraintResult {
	result := &ConstraintResult{
		RuleID:   rule.ID,
		Type:     rule.Type,
		Severity: rule.Severity,
		Passed:   true,
	}

	// 隐私字段检查
	if rule.Type == ConstraintPrivacy {
		if rule.Field == "password" || rule.Field == "token" || rule.Field == "secret" {
			result.Passed = false
			result.Message = "敏感字段不允许传输: " + rule.Field
		}
	}

	return result
}

// SetCustomRule 设置自定义规则
func (c *ConstraintChecker) SetCustomRule(id string, checker func(interface{}) bool) {
	c.mu.Lock()
	c.customRules[id] = checker
	c.mu.Unlock()
}

// GetRules 获取所有规则
func (c *ConstraintChecker) GetRules() []*ConstraintRule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rules := make([]*ConstraintRule, 0, len(c.rules))
	for _, r := range c.rules {
		rules = append(rules, r)
	}
	return rules
}

// Export 导出规则
func (c *ConstraintChecker) Export() json.RawMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rules := make([]*ConstraintRule, 0, len(c.rules))
	for _, r := range c.rules {
		rules = append(rules, r)
	}

	data, _ := json.Marshal(rules)
	return data
}

// Import 导入规则
func (c *ConstraintChecker) Import(data json.RawMessage) {
	var rules []*ConstraintRule
	if json.Unmarshal(data, &rules) == nil {
		c.mu.Lock()
		for _, r := range rules {
			c.rules[r.ID] = r
		}
		c.mu.Unlock()
	}
}

func (c *ConstraintChecker) getEnabledRules() []*ConstraintRule {
	rules := make([]*ConstraintRule, 0)
	for _, r := range c.rules {
		if r.Enabled {
			rules = append(rules, r)
		}
	}
	return rules
}

func (c *ConstraintChecker) loadDefaultRules() {
	// 隐私规则
	c.AddRule(&ConstraintRule{
		ID:          "privacy_password",
		Type:        ConstraintPrivacy,
		Severity:    SeverityBlock,
		Description: "禁止传输密码",
		Field:       "password",
		Enabled:     true,
	})

	c.AddRule(&ConstraintRule{
		ID:          "privacy_token",
		Type:        ConstraintPrivacy,
		Severity:    SeverityBlock,
		Description: "禁止传输令牌",
		Field:       "token",
		Enabled:     true,
	})

	c.AddRule(&ConstraintRule{
		ID:          "privacy_secret",
		Type:        ConstraintPrivacy,
		Severity:    SeverityBlock,
		Description: "禁止传输密钥",
		Field:       "secret",
		Enabled:     true,
	})

	// 安全规则
	c.AddRule(&ConstraintRule{
		ID:          "security_location",
		Type:        ConstraintSecurity,
		Severity:    SeverityWarning,
		Description: "位置数据传输需谨慎",
		Field:       "location",
		Enabled:     true,
	})

	// 带宽规则(省电模式)
	c.AddRule(&ConstraintRule{
		ID:          "bandwidth_limit",
		Type:        ConstraintBandwidth,
		Severity:    SeverityWarning,
		Description: "省电模式限制大数据传输",
		Enabled:     true,
	})
}
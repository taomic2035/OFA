// Package desktop provides constraint checking
package desktop

import (
	"encoding/json"
	"regexp"
	"sync"
	"time"
)

// ConstraintType defines constraint type
type ConstraintType string

const (
	TypePrivacy    ConstraintType = "privacy"
	TypeFinancial  ConstraintType = "financial"
	TypeSecurity   ConstraintType = "security"
	TypeAuth       ConstraintType = "auth"
	TypeLocation   ConstraintType = "location"
	TypeTime       ConstraintType = "time"
)

// ConstraintSeverity defines severity
type ConstraintSeverity string

const (
	SeverityWarning ConstraintSeverity = "warning"
	SeverityBlock   ConstraintSeverity = "block"
	SeverityAudit   ConstraintSeverity = "audit"
)

// ConstraintRule defines a rule
type ConstraintRule struct {
	ID          string
	Name        string
	Type        ConstraintType
	Severity    ConstraintSeverity
	Description string
	Pattern     string
	Fields      []string
	Enabled     bool
	CreatedAt   time.Time
}

// ConstraintResult represents check result
type ConstraintResult struct {
	RuleID       string
	RuleName     string
	Type         ConstraintType
	Severity     ConstraintSeverity
	Passed       bool
	Message      string
	Field        string
	MatchedValue string
}

// ConstraintReport represents full report
type ConstraintReport struct {
	TaskID    string
	SkillID   string
	Operation string
	AllPassed bool
	Results   []*ConstraintResult
	Timestamp time.Time
}

// HasBlockers checks if there are blocking results
func (r *ConstraintReport) HasBlockers() bool {
	for _, result := range r.Results {
		if !result.Passed && result.Severity == SeverityBlock {
			return true
		}
	}
	return false
}

// GetBlockerMessages returns blocker messages
func (r *ConstraintReport) GetBlockerMessages() []string {
	var messages []string
	for _, result := range r.Results {
		if !result.Passed && result.Severity == SeverityBlock {
			messages = append(messages, result.Message)
		}
	}
	return messages
}

// ConstraintChecker manages constraint checking
type ConstraintChecker struct {
	rules    map[string]*ConstraintRule
	checkers map[string]func(json.RawMessage) bool
	mu       sync.RWMutex
}

// NewConstraintChecker creates a new checker
func NewConstraintChecker() *ConstraintChecker {
	c := &ConstraintChecker{
		rules:    make(map[string]*ConstraintRule),
		checkers: make(map[string]func(json.RawMessage) bool),
	}
	c.loadDefaultRules()
	return c
}

// AddRule adds a rule
func (c *ConstraintChecker) AddRule(rule *ConstraintRule) {
	c.mu.Lock()
	rule.CreatedAt = time.Now()
	c.rules[rule.ID] = rule
	c.mu.Unlock()
}

// RemoveRule removes a rule
func (c *ConstraintChecker) RemoveRule(ruleID string) {
	c.mu.Lock()
	delete(c.rules, ruleID)
	delete(c.checkers, ruleID)
	c.mu.Unlock()
}

// SetRuleEnabled enables/disables a rule
func (c *ConstraintChecker) SetRuleEnabled(ruleID string, enabled bool) {
	c.mu.Lock()
	if rule, ok := c.rules[ruleID]; ok {
		rule.Enabled = enabled
	}
	c.mu.Unlock()
}

// GetRules returns all rules
func (c *ConstraintChecker) GetRules() []*ConstraintRule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rules := make([]*ConstraintRule, 0, len(c.rules))
	for _, rule := range c.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Check checks data against rules
func (c *ConstraintChecker) Check(taskID, skillID, operation string, data json.RawMessage) *ConstraintReport {
	report := &ConstraintReport{
		TaskID:    taskID,
		SkillID:   skillID,
		Operation: operation,
		Timestamp: time.Now(),
		AllPassed: true,
	}

	c.mu.RLock()
	rules := c.getApplicableRules()
	c.mu.RUnlock()

	c.checkRecursive(data, "", rules, report)

	for _, result := range report.Results {
		if !result.Passed {
			report.AllPassed = false
		}
	}

	return report
}

// SetCustomChecker sets a custom checker
func (c *ConstraintChecker) SetCustomChecker(ruleID string, checker func(json.RawMessage) bool) {
	c.mu.Lock()
	c.checkers[ruleID] = checker
	c.mu.Unlock()
}

func (c *ConstraintChecker) getApplicableRules() []*ConstraintRule {
	rules := make([]*ConstraintRule, 0)
	for _, rule := range c.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}

func (c *ConstraintChecker) checkRecursive(data json.RawMessage, prefix string, rules []*ConstraintRule, report *ConstraintReport) {
	var v interface{}
	if json.Unmarshal(data, &v) != nil {
		return
	}

	switch val := v.(type) {
	case map[string]interface{}:
		for key, value := range val {
			fieldPath := prefix + "." + key
			if prefix == "" {
				fieldPath = key
			}
			subData, _ := json.Marshal(value)
			c.checkRecursive(subData, fieldPath, rules, report)
		}
	case []interface{}:
		for i, value := range val {
			fieldPath := prefix + "[" + itoa(i) + "]"
			subData, _ := json.Marshal(value)
			c.checkRecursive(subData, fieldPath, rules, report)
		}
	default:
		result := c.checkField(prefix, data, rules)
		if !result.Passed || result.Severity == SeverityAudit {
			report.Results = append(report.Results, result)
		}
	}
}

func (c *ConstraintChecker) checkField(field string, value json.RawMessage, rules []*ConstraintRule) *ConstraintResult {
	result := &ConstraintResult{
		Field:  field,
		Passed: true,
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, rule := range rules {
		// Check field match
		if len(rule.Fields) > 0 {
			match := false
			for _, f := range rule.Fields {
				if f == field {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Check pattern
		if rule.Pattern != "" {
			var strValue string
			var v interface{}
			if json.Unmarshal(value, &v) == nil {
				if s, ok := v.(string); ok {
					strValue = s
				}
			}

			if strValue != "" {
				re, err := regexp.Compile(rule.Pattern)
				if err == nil && re.MatchString(strValue) {
					result.RuleID = rule.ID
					result.RuleName = rule.Name
					result.Type = rule.Type
					result.Severity = rule.Severity
					result.Passed = false
					result.Message = rule.Description
					result.MatchedValue = strValue
					return result
				}
			}
		}

		// Check custom checker
		if checker, ok := c.checkers[rule.ID]; ok {
			if !checker(value) {
				result.RuleID = rule.ID
				result.RuleName = rule.Name
				result.Type = rule.Type
				result.Severity = rule.Severity
				result.Passed = false
				result.Message = rule.Description
				return result
			}
		}
	}

	return result
}

func (c *ConstraintChecker) loadDefaultRules() {
	// Privacy rule
	c.AddRule(&ConstraintRule{
		ID:          "privacy-sensitive",
		Name:        "Sensitive Field Detection",
		Type:        TypePrivacy,
		Severity:    SeverityBlock,
		Description: "Sensitive data detected",
		Fields: []string{
			"password", "passwd", "pwd", "secret", "token",
			"credit_card", "ssn", "phone", "email",
		},
		Enabled: true,
	})

	// Financial rule
	c.AddRule(&ConstraintRule{
		ID:          "financial-card",
		Name:        "Credit Card Pattern",
		Type:        TypeFinancial,
		Severity:    SeverityAudit,
		Description: "Credit card number detected",
		Pattern:     "\\d{4}-\\d{4}-\\d{4}-\\d{4}",
		Fields:      []string{"card_number", "credit_card"},
		Enabled:     true,
	})

	// Security rule
	c.AddRule(&ConstraintRule{
		ID:          "security-cmd",
		Name:        "Command Injection Check",
		Type:        TypeSecurity,
		Severity:    SeverityBlock,
		Description: "Potential command injection",
		Pattern:     "(exec|eval|system|shell).*",
		Fields:      []string{"command", "cmd", "script"},
		Enabled:     true,
	})

	// Auth rule
	c.AddRule(&ConstraintRule{
		ID:          "auth-required",
		Name:        "Authentication Required",
		Type:        TypeAuth,
		Severity:    SeverityBlock,
		Description: "Operation requires auth",
		Enabled:     true,
	})
}
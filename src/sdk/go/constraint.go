package ofa

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// ConstraintType defines constraint type
type ConstraintType int

const (
	ConstraintPrivacy ConstraintType = iota
	ConstraintFinancial
	ConstraintSecurity
	ConstraintAuth
	ConstraintLocation
	ConstraintTime
	ConstraintCustom
)

// ConstraintSeverity defines constraint severity
type ConstraintSeverity int

const (
	SeverityWarning ConstraintSeverity = iota
	SeverityBlock
	SeverityAudit
)

// ConstraintRule defines a constraint rule
type ConstraintRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        ConstraintType    `json:"type"`
	Severity    ConstraintSeverity `json:"severity"`
	Description string            `json:"description"`
	Pattern     string            `json:"pattern"`
	Fields      []string          `json:"fields"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   int64             `json:"createdAt"`
}

// ConstraintResult represents a constraint check result
type ConstraintResult struct {
	RuleID       string            `json:"ruleId"`
	RuleName     string            `json:"ruleName"`
	Type         ConstraintType    `json:"type"`
	Severity     ConstraintSeverity `json:"severity"`
	Passed       bool              `json:"passed"`
	Message      string            `json:"message"`
	Field        string            `json:"field"`
	MatchedValue string            `json:"matchedValue"`
}

// ConstraintReport represents a full constraint report
type ConstraintReport struct {
	TaskID    string             `json:"taskId"`
	SkillID   string             `json:"skillId"`
	Operation string             `json:"operation"`
	AllPassed bool               `json:"allPassed"`
	Results   []*ConstraintResult `json:"results"`
	Timestamp int64              `json:"timestamp"`
}

// HasBlockers checks if report has blocking results
func (r *ConstraintReport) HasBlockers() bool {
	for _, result := range r.Results {
		if !result.Passed && result.Severity == SeverityBlock {
			return true
		}
	}
	return false
}

// GetBlockerMessages returns blocking messages
func (r *ConstraintReport) GetBlockerMessages() []string {
	var messages []string
	for _, result := range r.Results {
		if !result.Passed && result.Severity == SeverityBlock {
			messages = append(messages, result.Message)
		}
	}
	return messages
}

// CustomChecker is a custom constraint checker function
type CustomChecker func(json.RawMessage) bool

// ConstraintChecker manages constraint checking
type ConstraintChecker struct {
	rules         map[string]*ConstraintRule
	customCheckers map[string]CustomChecker
	mu            sync.Mutex
}

// NewConstraintChecker creates a new constraint checker
func NewConstraintChecker() *ConstraintChecker {
	c := &ConstraintChecker{
		rules:         make(map[string]*ConstraintRule),
		customCheckers: make(map[string]CustomChecker),
	}
	c.loadDefaultRules()
	return c
}

// AddRule adds a constraint rule
func (c *ConstraintChecker) AddRule(rule *ConstraintRule) {
	c.mu.Lock()
	rule.CreatedAt = time.Now().UnixMilli()
	c.rules[rule.ID] = rule
	c.mu.Unlock()
}

// RemoveRule removes a rule
func (c *ConstraintChecker) RemoveRule(ruleID string) {
	c.mu.Lock()
	delete(c.rules, ruleID)
	delete(c.customCheckers, ruleID)
	c.mu.Unlock()
}

// SetRuleEnabled enables or disables a rule
func (c *ConstraintChecker) SetRuleEnabled(ruleID string, enabled bool) {
	c.mu.Lock()
	if rule, ok := c.rules[ruleID]; ok {
		rule.Enabled = enabled
	}
	c.mu.Unlock()
}

// GetRules returns all rules
func (c *ConstraintChecker) GetRules() []*ConstraintRule {
	c.mu.Lock()
	defer c.mu.Unlock()

	var rules []*ConstraintRule
	for _, rule := range c.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetRulesByType returns rules by type
func (c *ConstraintChecker) GetRulesByType(typ ConstraintType) []*ConstraintRule {
	c.mu.Lock()
	defer c.mu.Unlock()

	var rules []*ConstraintRule
	for _, rule := range c.rules {
		if rule.Type == typ {
			rules = append(rules, rule)
		}
	}
	return rules
}

// Check checks data against all rules
func (c *ConstraintChecker) Check(taskID, skillID, operation string, data json.RawMessage) *ConstraintReport {
	report := &ConstraintReport{
		TaskID:    taskID,
		SkillID:   skillID,
		Operation: operation,
		Timestamp: time.Now().UnixMilli(),
		AllPassed: true,
	}

	c.mu.Lock()
	applicableRules := c.getApplicableRules()
	c.mu.Unlock()

	c.checkJSONRecursive(data, "", applicableRules, report)

	for _, result := range report.Results {
		if !result.Passed {
			report.AllPassed = false
		}
	}

	return report
}

// SetCustomChecker sets a custom checker for a rule
func (c *ConstraintChecker) SetCustomChecker(ruleID string, checker CustomChecker) {
	c.mu.Lock()
	c.customCheckers[ruleID] = checker
	c.mu.Unlock()
}

// LoadRules loads rules from JSON
func (c *ConstraintChecker) LoadRules(data json.RawMessage) error {
	var rules []*ConstraintRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	c.mu.Lock()
	for _, rule := range rules {
		c.rules[rule.ID] = rule
	}
	c.mu.Unlock()

	return nil
}

// ExportRules exports rules to JSON
func (c *ConstraintChecker) ExportRules() json.RawMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	rules := make([]*ConstraintRule, 0, len(c.rules))
	for _, rule := range c.rules {
		rules = append(rules, rule)
	}

	data, _ := json.Marshal(rules)
	return data
}

// ClearRules clears all rules
func (c *ConstraintChecker) ClearRules() {
	c.mu.Lock()
	c.rules = make(map[string]*ConstraintRule)
	c.customCheckers = make(map[string]CustomChecker)
	c.mu.Unlock()
}

func (c *ConstraintChecker) getApplicableRules() []*ConstraintRule {
	var rules []*ConstraintRule
	for _, rule := range c.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}

func (c *ConstraintChecker) checkJSONRecursive(data json.RawMessage, prefix string, rules []*ConstraintRule, report *ConstraintReport) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
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
			c.checkJSONRecursive(subData, fieldPath, rules, report)
		}
	case []interface{}:
		for i, value := range val {
			fieldPath := fmt.Sprintf("%s[%d]", prefix, i)
			subData, _ := json.Marshal(value)
			c.checkJSONRecursive(subData, fieldPath, rules, report)
		}
	default:
		// Leaf value - check it
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

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, rule := range rules {
		// Check if rule applies to this field
		if len(rule.Fields) > 0 {
			fieldMatch := false
			for _, f := range rule.Fields {
				if f == field || regexpMatch(f, field) {
					fieldMatch = true
					break
				}
			}
			if !fieldMatch {
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

			if strValue != "" && regexpMatch(rule.Pattern, strValue) {
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

		// Check custom checker
		if checker, ok := c.customCheckers[rule.ID]; ok {
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

func regexpMatch(pattern, value string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

func (c *ConstraintChecker) loadDefaultRules() {
	// Privacy rule
	c.AddRule(&ConstraintRule{
		ID:          "privacy-sensitive-fields",
		Name:        "Sensitive Field Detection",
		Type:        ConstraintPrivacy,
		Severity:    SeverityBlock,
		Description: "Sensitive data field detected",
		Fields: []string{
			"password", "passwd", "pwd", "secret", "token", "api_key",
			"credit_card", "ssn", "social_security", "phone", "email",
			"address", "birth_date", "name",
		},
		Enabled: true,
	})

	// Financial rule
	c.AddRule(&ConstraintRule{
		ID:          "financial-validation",
		Name:        "Financial Data Validation",
		Type:        ConstraintFinancial,
		Severity:    SeverityAudit,
		Description: "Financial transaction detected",
		Pattern:     "\\d{4}-\\d{4}-\\d{4}-\\d{4}",
		Fields:      []string{"credit_card", "card_number", "account"},
		Enabled:     true,
	})

	// Security rule
	c.AddRule(&ConstraintRule{
		ID:          "security-command-check",
		Name:        "Security Command Check",
		Type:        ConstraintSecurity,
		Severity:    SeverityBlock,
		Description: "Potential command injection detected",
		Pattern:     "(exec|eval|system|shell|cmd|command).*",
		Fields:      []string{"command", "cmd", "script", "code"},
		Enabled:     true,
	})

	// Auth rule
	c.AddRule(&ConstraintRule{
		ID:          "auth-required",
		Name:        "Authentication Required",
		Type:        ConstraintAuth,
		Severity:    SeverityBlock,
		Description: "Operation requires authentication",
		Enabled:     true,
	})

	// Location rule
	c.AddRule(&ConstraintRule{
		ID:          "location-restriction",
		Name:        "Location Restriction",
		Type:        ConstraintLocation,
		Severity:    SeverityWarning,
		Description: "Location-based operation restricted",
		Enabled:     true,
	})

	// Time rule
	c.AddRule(&ConstraintRule{
		ID:          "time-restriction",
		Name:        "Time Restriction",
		Type:        ConstraintTime,
		Severity:    SeverityWarning,
		Description: "Operation restricted by time",
		Enabled:     true,
	})
}
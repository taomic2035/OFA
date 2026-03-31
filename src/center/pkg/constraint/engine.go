// Package constraint - 交互约束检查引擎
// 用于检查 Agent 间交互是否符合安全约束规则
package constraint

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// ConstraintType 约束类型
type ConstraintType int

const (
	ConstraintNone ConstraintType = 0
	// 隐私保护 - 涉及个人隐私数据
	ConstraintPrivacy ConstraintType = 1 << iota
	// 财产相关 - 涉及支付、转账等
	ConstraintFinancial
	// 安全敏感 - 涉及系统安全设置
	ConstraintSecurity
	// 需要授权 - 需要用户明确授权
	ConstraintAuthRequired
	// 本地限制 - 仅允许本地执行
	ConstraintLocalOnly
	// 需要在线 - 需要 Center 参与
	ConstraintRequireOnline
)

// ActionCategory 操作类别
type ActionCategory string

const (
	ActionTaskCollaboration ActionCategory = "task_collaboration" // 任务协作
	ActionSkillInvocation    ActionCategory = "skill_invocation"  // 技能调用
	ActionStatusBroadcast    ActionCategory = "status_broadcast"  // 状态广播
	ActionHeartbeat          ActionCategory = "heartbeat"         // 心跳检测
	ActionDataTransfer       ActionCategory = "data_transfer"     // 数据传输
	ActionPayment            ActionCategory = "payment"           // 支付操作
	ActionCredentialAccess   ActionCategory = "credential_access" // 凭证访问
	ActionSecuritySetting    ActionCategory = "security_setting"  // 安全设置
	ActionPrivacyDataAccess  ActionCategory = "privacy_data"      // 隐私数据访问
)

// CheckResult 约束检查结果
type CheckResult struct {
	Allowed   bool           `json:"allowed"`
	Violated  ConstraintType `json:"violated,omitempty"`
	Reason    string         `json:"reason,omitempty"`
	Require   ConstraintType `json:"require,omitempty"`
	Action    string         `json:"action"`
	Source    string         `json:"source,omitempty"`
	Target    string         `json:"target,omitempty"`
}

// Rule 约束规则
type Rule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Category    ActionCategory `json:"category"`
	Pattern     string         `json:"pattern"`      // 匹配模式 (正则)
	Allowed     bool           `json:"allowed"`      // 是否允许
	Constraints ConstraintType `json:"constraints"`  // 约束条件
	RequireAuth bool           `json:"require_auth"` // 是否需要授权
	OfflineOK   bool           `json:"offline_ok"`  // 是否允许离线
	P2POK       bool           `json:"p2p_ok"`      // 是否允许 P2P
	Priority    int            `json:"priority"`    // 优先级
}

// Engine 约束检查引擎
type Engine struct {
	rules       []*Rule
	patterns    map[string]*regexp.Regexp
	sensitivePatterns []*SensitivePattern
	mu          sync.RWMutex
}

// SensitivePattern 敏感数据模式
type SensitivePattern struct {
	Name    string
	Pattern *regexp.Regexp
	Type    ConstraintType
}

// NewEngine 创建约束检查引擎
func NewEngine() *Engine {
	e := &Engine{
		rules:    make([]*Rule, 0),
		patterns: make(map[string]*regexp.Regexp),
		sensitivePatterns: []*SensitivePattern{
			// 身份证号
			{Name: "IDCard", Pattern: regexp.MustCompile(`\d{17}[\dXx]`), Type: ConstraintPrivacy},
			// 手机号
			{Name: "Phone", Pattern: regexp.MustCompile(`1[3-9]\d{9}`), Type: ConstraintPrivacy},
			// 银行卡号
			{Name: "BankCard", Pattern: regexp.MustCompile(`\d{16,19}`), Type: ConstraintFinancial},
			// 密码字段
			{Name: "Password", Pattern: regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]+"`), Type: ConstraintSecurity},
			// Token
			{Name: "Token", Pattern: regexp.MustCompile(`(?i)"token"\s*:\s*"[^"]+"`), Type: ConstraintSecurity},
		},
	}

	// 加载默认规则
	e.loadDefaultRules()

	return e
}

// loadDefaultRules 加载默认约束规则
func (e *Engine) loadDefaultRules() {
	defaultRules := []*Rule{
		// === 允许的操作 ===
		{
			ID:        "allow_task_collab",
			Name:      "允许任务协作",
			Category:  ActionTaskCollaboration,
			Pattern:   `task\.(submit|execute|cancel|query)`,
			Allowed:   true,
			OfflineOK: true,
			P2POK:     true,
			Priority:  100,
		},
		{
			ID:        "allow_skill_invoke",
			Name:      "允许技能调用",
			Category:  ActionSkillInvocation,
			Pattern:   `skill\.(execute|query|list)`,
			Allowed:   true,
			OfflineOK: true,
			P2POK:     true,
			Priority:  100,
		},
		{
			ID:        "allow_status",
			Name:      "允许状态广播",
			Category:  ActionStatusBroadcast,
			Pattern:   `status\.(broadcast|update|query)`,
			Allowed:   true,
			OfflineOK: true,
			P2POK:     true,
			Priority:  100,
		},
		{
			ID:        "allow_heartbeat",
			Name:      "允许心跳检测",
			Category:  ActionHeartbeat,
			Pattern:   `heartbeat\.(ping|pong)`,
			Allowed:   true,
			OfflineOK: true,
			P2POK:     true,
			Priority:  100,
		},

		// === 禁止的操作 ===
		{
			ID:          "deny_privacy_transfer",
			Name:        "禁止隐私数据传输",
			Category:    ActionPrivacyDataAccess,
			Pattern:     `data\.(transfer|sync).*personal`,
			Allowed:     false,
			Constraints: ConstraintPrivacy,
			RequireAuth: true,
			OfflineOK:   false,
			P2POK:       false,
			Priority:    200,
		},
		{
			ID:          "deny_payment",
			Name:        "禁止离线支付",
			Category:    ActionPayment,
			Pattern:     `payment\.(create|confirm|cancel)`,
			Allowed:     false,
			Constraints: ConstraintFinancial | ConstraintRequireOnline,
			RequireAuth: true,
			OfflineOK:   false,
			P2POK:       false,
			Priority:    300,
		},
		{
			ID:          "deny_credential",
			Name:        "禁止凭证共享",
			Category:    ActionCredentialAccess,
			Pattern:     `credential\.(share|transfer)`,
			Allowed:     false,
			Constraints: ConstraintSecurity,
			RequireAuth: true,
			OfflineOK:   false,
			P2POK:       false,
			Priority:    300,
		},
		{
			ID:          "deny_security_change",
			Name:        "禁止安全设置变更",
			Category:    ActionSecuritySetting,
			Pattern:     `security\.(change|reset|disable)`,
			Allowed:     false,
			Constraints: ConstraintSecurity | ConstraintRequireOnline,
			RequireAuth: true,
			OfflineOK:   false,
			P2POK:       false,
			Priority:    300,
		},
	}

	for _, rule := range defaultRules {
		e.AddRule(rule)
	}
}

// AddRule 添加规则
func (e *Engine) AddRule(rule *Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 编译正则
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	e.patterns[rule.ID] = re
	e.rules = append(e.rules, rule)

	// 按优先级排序
	sortRules(e.rules)

	return nil
}

// Check 检查操作是否符合约束
func (e *Engine) Check(ctx context.Context, action string, data []byte, opts ...CheckOption) *CheckResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	options := &checkOptions{
		source:  "unknown",
		target:  "unknown",
		offline: false,
		p2p:     false,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 1. 匹配规则
	var matchedRule *Rule
	for _, rule := range e.rules {
		if re, ok := e.patterns[rule.ID]; ok {
			if re.MatchString(action) {
				matchedRule = rule
				break
			}
		}
	}

	// 2. 无匹配规则时，默认允许
	if matchedRule == nil {
		return &CheckResult{
			Allowed: true,
			Action:  action,
			Source:  options.source,
			Target:  options.target,
		}
	}

	// 3. 检查约束条件
	result := &CheckResult{
		Action: action,
		Source: options.source,
		Target: options.target,
	}

	// 检查是否禁止
	if !matchedRule.Allowed {
		result.Allowed = false
		result.Violated = matchedRule.Constraints
		result.Reason = fmt.Sprintf("操作 '%s' 被 '%s' 规则禁止", action, matchedRule.Name)
		return result
	}

	// 检查离线限制
	if options.offline && !matchedRule.OfflineOK {
		result.Allowed = false
		result.Violated = ConstraintRequireOnline
		result.Reason = "此操作需要在线执行"
		return result
	}

	// 检查 P2P 限制
	if options.p2p && !matchedRule.P2POK {
		result.Allowed = false
		result.Violated = ConstraintSecurity
		result.Reason = "此操作不允许 P2P 直接通信"
		return result
	}

	// 4. 检查数据中的敏感信息
	if len(data) > 0 {
		violations := e.checkSensitiveData(data)
		if len(violations) > 0 {
			result.Allowed = false
			result.Violated = violations[0]
			result.Reason = "数据包含敏感信息，需要特殊处理"
			return result
		}
	}

	// 5. 检查是否需要授权
	if matchedRule.RequireAuth {
		result.Require = ConstraintAuthRequired
	}

	result.Allowed = true
	return result
}

// checkSensitiveData 检查敏感数据
func (e *Engine) checkSensitiveData(data []byte) []ConstraintType {
	var violations []ConstraintType

	dataStr := string(data)
	for _, sp := range e.sensitivePatterns {
		if sp.Pattern.MatchString(dataStr) {
			violations = append(violations, sp.Type)
		}
	}

	return violations
}

// CheckAgentInteraction 检查 Agent 间交互
func (e *Engine) CheckAgentInteraction(ctx context.Context, source, target, action string, data []byte) *CheckResult {
	return e.Check(ctx, action, data,
		WithSource(source),
		WithTarget(target),
		WithP2P(true),
	)
}

// CheckOfflineAction 检查离线操作
func (e *Engine) CheckOfflineAction(ctx context.Context, action string, data []byte) *CheckResult {
	return e.Check(ctx, action, data, WithOffline(true))
}

// === 选项 ===

type checkOptions struct {
	source  string
	target  string
	offline bool
	p2p     bool
}

// CheckOption 检查选项
type CheckOption func(*checkOptions)

// WithSource 设置来源
func WithSource(source string) CheckOption {
	return func(o *checkOptions) {
		o.source = source
	}
}

// WithTarget 设置目标
func WithTarget(target string) CheckOption {
	return func(o *checkOptions) {
		o.target = target
	}
}

// WithOffline 设置离线模式
func WithOffline(offline bool) CheckOption {
	return func(o *checkOptions) {
		o.offline = offline
	}
}

// WithP2P 设置 P2P 模式
func WithP2P(p2p bool) CheckOption {
	return func(o *checkOptions) {
		o.p2p = p2p
	}
}

// === 辅助函数 ===

func sortRules(rules []*Rule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

// GetConstraintName 获取约束名称
func GetConstraintName(c ConstraintType) string {
	var names []string
	if c&ConstraintPrivacy != 0 {
		names = append(names, "隐私保护")
	}
	if c&ConstraintFinancial != 0 {
		names = append(names, "财产安全")
	}
	if c&ConstraintSecurity != 0 {
		names = append(names, "安全敏感")
	}
	if c&ConstraintAuthRequired != 0 {
		names = append(names, "需要授权")
	}
	if c&ConstraintRequireOnline != 0 {
		names = append(names, "需要在线")
	}
	if len(names) == 0 {
		return "无"
	}
	return strings.Join(names, ", ")
}

// MarshalJSON 实现 JSON 序列化
func (c ConstraintType) MarshalJSON() ([]byte, error) {
	return json.Marshal(GetConstraintName(c))
}
package constraint

import (
	"context"
	"testing"
)

func TestEngine_Check(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	tests := []struct {
		name      string
		action    string
		data      []byte
		opts      []CheckOption
		wantAllow bool
		wantConst ConstraintType
	}{
		{
			name:      "允许任务协作",
			action:    "task.submit",
			data:      nil,
			wantAllow: true,
		},
		{
			name:      "允许技能调用",
			action:    "skill.execute",
			data:      nil,
			wantAllow: true,
		},
		{
			name:      "允许心跳",
			action:    "heartbeat.ping",
			data:      nil,
			wantAllow: true,
		},
		{
			name:      "禁止离线支付",
			action:    "payment.create",
			data:      nil,
			opts:      []CheckOption{WithOffline(true)},
			wantAllow: false,
			wantConst: ConstraintFinancial | ConstraintRequireOnline,
		},
		{
			name:      "禁止凭证共享",
			action:    "credential.share",
			data:      nil,
			wantAllow: false,
			wantConst: ConstraintSecurity,
		},
		{
			name:      "禁止P2P隐私数据传输",
			action:    "data.transfer.personal",
			data:      nil,
			opts:      []CheckOption{WithP2P(true)},
			wantAllow: false,
			wantConst: ConstraintPrivacy,
		},
		{
			name:      "检测身份证号",
			action:    "data.transfer",
			data:      []byte(`{"idcard": "330102199001011234"}`),
			wantAllow: false,
			wantConst: ConstraintPrivacy,
		},
		{
			name:      "检测手机号",
			action:    "data.transfer",
			data:      []byte(`{"phone": "13812345678"}`),
			wantAllow: false,
			wantConst: ConstraintPrivacy,
		},
		{
			name:      "检测银行卡号",
			action:    "data.transfer",
			data:      []byte(`{"card": "6222021234567890123"}`),
			wantAllow: false,
			wantConst: ConstraintFinancial,
		},
		{
			name:      "正常数据传输",
			action:    "data.transfer",
			data:      []byte(`{"message": "hello"}`),
			wantAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Check(ctx, tt.action, tt.data, tt.opts...)

			if result.Allowed != tt.wantAllow {
				t.Errorf("Check() allowed = %v, want %v, reason: %s", result.Allowed, tt.wantAllow, result.Reason)
			}

			if !tt.wantAllow && result.Violated != tt.wantConst {
				t.Errorf("Check() violated = %v, want %v", result.Violated, tt.wantConst)
			}
		})
	}
}

func TestEngine_CheckAgentInteraction(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	tests := []struct {
		name      string
		source    string
		target    string
		action    string
		data      []byte
		wantAllow bool
	}{
		{
			name:      "Agent间任务协作",
			source:    "agent-1",
			target:    "agent-2",
			action:    "task.collaboration",
			data:      []byte(`{"task_id": "task-123"}`),
			wantAllow: true,
		},
		{
			name:      "Agent间禁止凭证共享",
			source:    "agent-1",
			target:    "agent-2",
			action:    "credential.share",
			data:      nil,
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.CheckAgentInteraction(ctx, tt.source, tt.target, tt.action, tt.data)

			if result.Allowed != tt.wantAllow {
				t.Errorf("CheckAgentInteraction() allowed = %v, want %v, reason: %s", result.Allowed, tt.wantAllow, result.Reason)
			}
		})
	}
}

func TestEngine_CheckOfflineAction(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	tests := []struct {
		name      string
		action    string
		data      []byte
		wantAllow bool
	}{
		{
			name:      "离线执行本地技能",
			action:    "skill.execute.local",
			data:      nil,
			wantAllow: true,
		},
		{
			name:      "离线任务查询",
			action:    "task.query",
			data:      nil,
			wantAllow: true,
		},
		{
			name:      "离线支付被禁止",
			action:    "payment.create",
			data:      nil,
			wantAllow: false,
		},
		{
			name:      "离线安全设置被禁止",
			action:    "security.change",
			data:      nil,
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.CheckOfflineAction(ctx, tt.action, tt.data)

			if result.Allowed != tt.wantAllow {
				t.Errorf("CheckOfflineAction() allowed = %v, want %v, reason: %s", result.Allowed, tt.wantAllow, result.Reason)
			}
		})
	}
}

func TestEngine_AddRule(t *testing.T) {
	engine := NewEngine()

	// 测试添加自定义规则
	rule := &Rule{
		ID:        "custom_test",
		Name:      "自定义测试规则",
		Category:  ActionTaskCollaboration,
		Pattern:   `custom\.action`,
		Allowed:   true,
		OfflineOK: true,
		P2POK:     true,
		Priority:  150,
	}

	err := engine.AddRule(rule)
	if err != nil {
		t.Fatalf("AddRule() error = %v", err)
	}

	// 验证规则生效
	ctx := context.Background()
	result := engine.Check(ctx, "custom.action", nil)

	if !result.Allowed {
		t.Errorf("Custom rule not applied, allowed = %v", result.Allowed)
	}
}

func TestEngine_InvalidPattern(t *testing.T) {
	engine := NewEngine()

	rule := &Rule{
		ID:       "invalid",
		Name:     "无效规则",
		Category: ActionTaskCollaboration,
		Pattern:  `[invalid(`, // 无效正则
	}

	err := engine.AddRule(rule)
	if err == nil {
		t.Error("AddRule() should fail with invalid pattern")
	}
}

func TestGetConstraintName(t *testing.T) {
	tests := []struct {
		constraint ConstraintType
		want       string
	}{
		{ConstraintNone, "无"},
		{ConstraintPrivacy, "隐私保护"},
		{ConstraintFinancial, "财产安全"},
		{ConstraintSecurity, "安全敏感"},
		{ConstraintPrivacy | ConstraintFinancial, "隐私保护, 财产安全"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetConstraintName(tt.constraint)
			if got != tt.want {
				t.Errorf("GetConstraintName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSensitiveDataDetection(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	tests := []struct {
		name      string
		data      string
		wantAllow bool
	}{
		{
			name:      "包含身份证",
			data:      `{"name": "张三", "idcard": "110101199001011234"}`,
			wantAllow: false,
		},
		{
			name:      "包含手机号",
			data:      `{"contact": "13912345678"}`,
			wantAllow: false,
		},
		{
			name:      "包含密码",
			data:      `{"password": "secret123"}`,
			wantAllow: false,
		},
		{
			name:      "包含Token",
			data:      `{"token": "abc123xyz"}`,
			wantAllow: false,
		},
		{
			name:      "正常数据",
			data:      `{"name": "张三", "age": 25}`,
			wantAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Check(ctx, "data.transfer", []byte(tt.data))

			if result.Allowed != tt.wantAllow {
				t.Errorf("Check() allowed = %v, want %v, data: %s", result.Allowed, tt.wantAllow, tt.data)
			}
		})
	}
}
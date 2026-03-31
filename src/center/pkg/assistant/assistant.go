// Package assistant - AI助手集成模块
package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AssistantType 助手类型
type AssistantType string

const (
	AssistantTask       AssistantType = "task"       // 任务助手
	AssistantDiagnosis  AssistantType = "diagnosis"  // 诊断助手
	AssistantRecommend  AssistantType = "recommend"  // 推荐助手
	AssistantTemplate   AssistantType = "template"   // 模板助手
)

// IntentType 意图类型
type IntentType string

const (
	IntentCreateTask    IntentType = "create_task"
	IntentQueryStatus   IntentType = "query_status"
	IntentDiagnoseError IntentType = "diagnose_error"
	IntentGetRecommend  IntentType = "get_recommend"
	IntentCreateTemplate IntentType = "create_template"
	IntentModifyTask    IntentType = "modify_task"
	IntentCancelTask    IntentType = "cancel_task"
)

// AssistantRequest 助手请求
type AssistantRequest struct {
	ID          string                 `json:"id"`
	Type        AssistantType          `json:"type"`
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context"`
	UserID      string                 `json:"user_id"`
	TenantID    string                 `json:"tenant_id"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AssistantResponse 助手响应
type AssistantResponse struct {
	ID          string                 `json:"id"`
	RequestID   string                 `json:"request_id"`
	Intent      IntentType             `json:"intent"`
	Result      map[string]interface{} `json:"result"`
	Suggestions []string               `json:"suggestions"`
	Confidence  float64                `json:"confidence"` // 0-1
	Action      string                 `json:"action"`     // 建议的下一步操作
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Parameters  map[string]interface{} `json:"parameters"`
	Skills      []string               `json:"skills"`
	Priority    int                    `json:"priority"`
	Example     string                 `json:"example"`
	CreatedAt   time.Time              `json:"created_at"`
}

// DiagnosisResult 诊断结果
type DiagnosisResult struct {
	ID           string   `json:"id"`
	ErrorType    string   `json:"error_type"`
	Description  string   `json:"description"`
	RootCause    string   `json:"root_cause"`
	Solutions    []string `json:"solutions"`
	Severity     string   `json:"severity"` // low, medium, high, critical
	AutoFixable  bool     `json:"auto_fixable"`
	FixCommand   string   `json:"fix_command,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// Recommendation 推荐
type Recommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // agent, skill, parameter, strategy
	Target      string    `json:"target"`
	Reason      string    `json:"reason"`
	Score       float64   `json:"score"` // 0-10
	Alternative []string  `json:"alternative,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// AssistantConfig 助手配置
type AssistantConfig struct {
	EnableAutoDiagnosis bool    `json:"enable_auto_diagnosis"`
	EnableAutoRecommend bool    `json:"enable_auto_recommend"`
	ConfidenceThreshold float64 `json:"confidence_threshold"` // 低于此值会请求澄清
	MaxSuggestions      int     `json:"max_suggestions"`
	HistorySize         int     `json:"history_size"` // 历史记录大小
}

// AssistantManager 助手管理器
type AssistantManager struct {
	config      AssistantConfig
	templates   map[string]*TaskTemplate
	history     []*AssistantResponse
	dialogCtx   map[string]*DialogContext // 用户对话上下文
	mu          sync.RWMutex
}

// DialogContext 对话上下文
type DialogContext struct {
	UserID      string
	SessionID   string
	IntentStack []IntentType
	Entities    map[string]interface{}
	LastQuery   string
	StartTime   time.Time
	TurnCount   int
}

// NewAssistantManager 创建助手管理器
func NewAssistantManager(config AssistantConfig) *AssistantManager {
	return &AssistantManager{
		config:    config,
		templates: make(map[string]*TaskTemplate),
		history:   make([]*AssistantResponse, 0),
		dialogCtx: make(map[string]*DialogContext),
	}
}

// Initialize 初始化助手系统
func (am *AssistantManager) Initialize() error {
	// 加载预置模板
	am.loadPresetTemplates()
	return nil
}

// loadPresetTemplates 加载预置模板
func (am *AssistantManager) loadPresetTemplates() {
	templates := []TaskTemplate{
		{
			ID:          "text-analysis",
			Name:        "文本分析",
			Description: "分析文本内容，提取关键信息",
			Category:    "text",
			Parameters:  map[string]interface{}{"input": "", "operations": []string{"extract", "classify"}},
			Skills:      []string{"text.process"},
			Priority:    5,
			Example:     "帮我分析这段文本的关键词",
		},
		{
			ID:          "data-calc",
			Name:        "数据计算",
			Description: "执行数学计算和统计分析",
			Category:    "calculation",
			Parameters:  map[string]interface{}{"expression": "", "precision": 2},
			Skills:      []string{"calculator"},
			Priority:    5,
			Example:     "帮我计算 sqrt(144) + 25 * 3",
		},
		{
			ID:          "json-transform",
			Name:        "JSON转换",
			Description: "处理JSON数据，提取或转换",
			Category:    "data",
			Parameters:  map[string]interface{}{"input": "", "operation": "pretty"},
			Skills:      []string{"json.process"},
			Priority:    5,
			Example:     "帮我格式化这段JSON数据",
		},
		{
			ID:          "batch-process",
			Name:        "批量处理",
			Description: "批量执行多个任务",
			Category:    "workflow",
			Parameters:  map[string]interface{}{"tasks": []interface{}{}, "parallel": true},
			Skills:      []string{"text.process", "json.process"},
			Priority:    6,
			Example:     "批量处理这些文件",
		},
		{
			ID:          "scheduled-task",
			Name:        "定时任务",
			Description: "创建定时执行的任务",
			Category:    "schedule",
			Parameters:  map[string]interface{}{"task": "", "schedule": "daily", "time": "09:00"},
			Skills:      []string{},
			Priority:    7,
			Example:     "每天早上9点执行数据备份",
		},
	}

	am.mu.Lock()
	for _, t := range templates {
		t.CreatedAt = time.Now()
		am.templates[t.ID] = &t
	}
	am.mu.Unlock()
}

// ProcessQuery 处理用户查询
func (am *AssistantManager) ProcessQuery(ctx context.Context, req *AssistantRequest) (*AssistantResponse, error) {
	// 解析意图
	intent, entities, confidence := am.parseIntent(req.Query)

	// 低置信度时请求澄清
	if confidence < am.config.ConfidenceThreshold {
		return &AssistantResponse{
			ID:          fmt.Sprintf("resp-%d", time.Now().UnixNano()),
			RequestID:   req.ID,
			Intent:      intent,
			Confidence:  confidence,
			Suggestions: am.generateClarifications(intent, req.Query),
			Action:      "clarify",
			Timestamp:   time.Now(),
		}, nil
	}

	// 处理意图
	result, suggestions, action, err := am.handleIntent(ctx, intent, entities, req)

	resp := &AssistantResponse{
		ID:          fmt.Sprintf("resp-%d", time.Now().UnixNano()),
		RequestID:   req.ID,
		Intent:      intent,
		Result:      result,
		Suggestions: suggestions,
		Confidence:  confidence,
		Action:      action,
		Timestamp:   time.Now(),
	}

	if err != nil {
		resp.Error = err.Error()
	}

	// 保存历史
	am.addToHistory(resp)

	return resp, nil
}

// parseIntent 解析意图
func (am *AssistantManager) parseIntent(query string) (IntentType, map[string]interface{}, float64) {
	queryLower := strings.ToLower(query)
	var intent IntentType
	entities := make(map[string]interface{})
	confidence := 0.8

	// 意图识别规则
	if strings.Contains(queryLower, "创建") || strings.Contains(queryLower, "帮我") ||
		strings.Contains(queryLower, "执行") || strings.Contains(queryLower, "运行") {
		intent = IntentCreateTask
		confidence = 0.9

		// 提取任务类型
		if strings.Contains(queryLower, "计算") {
			entities["task_type"] = "calculator"
		} else if strings.Contains(queryLower, "分析") || strings.Contains(queryLower, "提取") {
			entities["task_type"] = "text.process"
		} else if strings.Contains(queryLower, "json") || strings.Contains(queryLower, "格式化") {
			entities["task_type"] = "json.process"
		}

		// 提取输入内容
		if idx := strings.Index(query, " "); idx > 0 {
			entities["input"] = query[idx+1:]
		}
	} else if strings.Contains(queryLower, "状态") || strings.Contains(queryLower, "进度") ||
		strings.Contains(queryLower, "结果") {
		intent = IntentQueryStatus
		confidence = 0.85

		// 提取任务ID
		if parts := strings.Fields(query); len(parts) > 1 {
			entities["task_id"] = parts[len(parts)-1]
		}
	} else if strings.Contains(queryLower, "错误") || strings.Contains(queryLower, "问题") ||
		strings.Contains(queryLower, "失败") || strings.Contains(queryLower, "异常") {
		intent = IntentDiagnoseError
		confidence = 0.9

		// 提取错误信息
		entities["error_query"] = query
	} else if strings.Contains(queryLower, "推荐") || strings.Contains(queryLower, "选择") ||
		strings.Contains(queryLower, "哪个") {
		intent = IntentGetRecommend
		confidence = 0.85
	} else if strings.Contains(queryLower, "模板") || strings.Contains(queryLower, "示例") {
		intent = IntentCreateTemplate
		confidence = 0.9
	} else if strings.Contains(queryLower, "修改") || strings.Contains(queryLower, "更新") {
		intent = IntentModifyTask
		confidence = 0.8
	} else if strings.Contains(queryLower, "取消") || strings.Contains(queryLower, "停止") {
		intent = IntentCancelTask
		confidence = 0.85
	} else {
		intent = IntentCreateTask
		confidence = 0.5 // 默认意图，低置信度
	}

	return intent, entities, confidence
}

// handleIntent 处理意图
func (am *AssistantManager) handleIntent(ctx context.Context, intent IntentType, entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	switch intent {
	case IntentCreateTask:
		return am.handleCreateTask(entities, req)
	case IntentQueryStatus:
		return am.handleQueryStatus(entities, req)
	case IntentDiagnoseError:
		return am.handleDiagnosis(entities, req)
	case IntentGetRecommend:
		return am.handleRecommendation(entities, req)
	case IntentCreateTemplate:
		return am.handleTemplate(entities, req)
	case IntentModifyTask:
		return am.handleModifyTask(entities, req)
	case IntentCancelTask:
		return am.handleCancelTask(entities, req)
	default:
		return nil, []string{"请提供更多信息"}, "clarify", fmt.Errorf("未知意图: %s", intent)
	}
}

// handleCreateTask 处理创建任务意图
func (am *AssistantManager) handleCreateTask(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})

	// 匹配模板
	templateID := entities["task_type"]
	if templateID != nil {
		am.mu.RLock()
		template := am.findTemplateBySkill(templateID.(string))
		am.mu.RUnlock()

		if template != nil {
			result["matched_template"] = template.ID
			result["template_name"] = template.Name
			result["suggested_params"] = template.Parameters
			result["required_skills"] = template.Skills

			// 提取输入
			if input, ok := entities["input"]; ok {
				result["input"] = input
			}

			return result, []string{
				"确认执行此任务？",
				"修改参数设置？",
			}, "confirm_execute", nil
		}
	}

	// 无匹配模板，生成通用任务建议
	result["task_type"] = templateID
	return result, []string{
		"请指定具体的任务类型",
		"查看可用模板列表",
	}, "clarify", nil
}

// handleQueryStatus 处理状态查询意图
func (am *AssistantManager) handleQueryStatus(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})

	taskID, ok := entities["task_id"]
	if ok {
		result["query_target"] = taskID
		result["status_hint"] = "请使用 GET /api/v1/tasks/" + taskID.(string)
	} else {
		result["status_hint"] = "请提供任务ID"
	}

	return result, []string{
		"查询所有任务状态",
		"查询最近执行的任务",
	}, "query_api", nil
}

// handleDiagnosis 处理诊断意图
func (am *AssistantManager) handleDiagnosis(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})

	// 模拟诊断逻辑
	diagnosis := &DiagnosisResult{
		ID:          fmt.Sprintf("diag-%d", time.Now().UnixNano()),
		ErrorType:   "task_timeout",
		Description: "任务执行超时",
		RootCause:   "Agent响应时间过长，可能由于资源不足或网络延迟",
		Solutions: []string{
			"检查Agent资源使用情况",
			"增加超时时间设置",
			"切换到负载较低的Agent",
			"检查网络连接状态",
		},
		Severity:    "medium",
		AutoFixable: true,
		FixCommand:  "ofa task retry --timeout=60",
		Timestamp:   time.Now(),
	}

	result["diagnosis"] = diagnosis

	return result, []string{
		"自动修复此问题",
		"查看更多解决方案",
		"查看问题历史",
	}, "diagnose", nil
}

// handleRecommendation 处理推荐意图
func (am *AssistantManager) handleRecommendation(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})

	recommendations := []Recommendation{
		{
			ID:     "rec-1",
			Type:   "agent",
			Target: "agent-high-cpu",
			Reason: "该Agent当前负载最低(15%)，响应速度最快",
			Score:  9.2,
		},
		{
			ID:     "rec-2",
			Type:   "skill",
			Target: "text.process.v2",
			Reason: "新版本技能支持更多操作，性能提升30%",
			Score:  8.5,
			Alternative: []string{"text.process.v1"},
		},
		{
			ID:     "rec-3",
			Type:   "strategy",
			Target: "hybrid",
			Reason: "混合策略综合考虑负载、延迟和资源利用率",
			Score:  8.0,
		},
	}

	result["recommendations"] = recommendations

	return result, []string{
		"使用推荐的Agent",
		"使用推荐的策略",
		"查看更多选项",
	}, "recommend", nil
}

// handleTemplate 处理模板意图
func (am *AssistantManager) handleTemplate(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})

	am.mu.RLock()
	templateList := make([]map[string]interface{}, 0)
	for _, t := range am.templates {
		templateList = append(templateList, map[string]interface{}{
			"id":          t.ID,
			"name":        t.Name,
			"description": t.Description,
			"example":     t.Example,
		})
	}
	am.mu.RUnlock()

	result["templates"] = templateList
	result["count"] = len(templateList)

	return result, []string{
		"选择一个模板使用",
		"创建自定义模板",
	}, "show_templates", nil
}

// handleModifyTask 处理修改任务意图
func (am *AssistantManager) handleModifyTask(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})
	result["hint"] = "请使用 PATCH /api/v1/tasks/{id} 修改任务参数"
	return result, []string{"查看当前任务参数"}, "modify_api", nil
}

// handleCancelTask 处理取消任务意图
func (am *AssistantManager) handleCancelTask(entities map[string]interface{}, req *AssistantRequest) (map[string]interface{}, []string, string, error) {
	result := make(map[string]interface{})
	taskID, ok := entities["task_id"]
	if ok {
		result["cancel_target"] = taskID
		result["hint"] = "请使用 DELETE /api/v1/tasks/" + taskID.(string)
	} else {
		result["hint"] = "请提供要取消的任务ID"
	}
	return result, []string{"确认取消"}, "cancel_api", nil
}

// generateClarifications 生成澄清建议
func (am *AssistantManager) generateClarifications(intent IntentType, query string) []string {
	switch intent {
	case IntentCreateTask:
		return []string{
			"请指定任务类型（如：文本分析、数据计算、JSON处理）",
			"请提供具体的输入内容",
		}
	case IntentQueryStatus:
		return []string{
			"请提供要查询的任务ID",
			"查询所有任务状态？",
		}
	default:
		return []string{
			"请提供更多详细信息",
			"查看帮助文档",
		}
	}
}

// findTemplateBySkill 根据技能查找模板
func (am *AssistantManager) findTemplateBySkill(skill string) *TaskTemplate {
	for _, t := range am.templates {
		for _, s := range t.Skills {
			if s == skill {
				return t
			}
		}
	}
	return nil
}

// addToHistory 添加到历史记录
func (am *AssistantManager) addToHistory(resp *AssistantResponse) {
	am.mu.Lock()
	am.history = append(am.history, resp)
	// 限制历史大小
	if len(am.history) > am.config.HistorySize {
		am.history = am.history[1:]
	}
	am.mu.Unlock()
}

// AddCustomTemplate 添加自定义模板
func (am *AssistantManager) AddCustomTemplate(template *TaskTemplate) error {
	if template.ID == "" {
		template.ID = fmt.Sprintf("custom-%d", time.Now().UnixNano())
	}
	template.CreatedAt = time.Now()

	am.mu.Lock()
	am.templates[template.ID] = template
	am.mu.Unlock()

	return nil
}

// GetTemplate 获取模板
func (am *AssistantManager) GetTemplate(id string) (*TaskTemplate, error) {
	am.mu.RLock()
	template, ok := am.templates[id]
	am.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("模板不存在: %s", id)
	}
	return template, nil
}

// ListTemplates 列出模板
func (am *AssistantManager) ListTemplates() []*TaskTemplate {
	am.mu.RLock()
	list := make([]*TaskTemplate, 0, len(am.templates))
	for _, t := range am.templates {
		list = append(list, t)
	}
	am.mu.RUnlock()
	return list
}

// GetHistory 获取历史记录
func (am *AssistantManager) GetHistory(limit int) []*AssistantResponse {
	am.mu.RLock()
	size := len(am.history)
	if limit > size || limit <= 0 {
		limit = size
	}
	history := am.history[size-limit:]
	am.mu.RUnlock()
	return history
}

// CreateDialogContext 创建对话上下文
func (am *AssistantManager) CreateDialogContext(userID string) *DialogContext {
	ctx := &DialogContext{
		UserID:      userID,
		SessionID:   fmt.Sprintf("sess-%d", time.Now().UnixNano()),
		IntentStack: make([]IntentType, 0),
		Entities:    make(map[string]interface{}),
		StartTime:   time.Now(),
		TurnCount:   0,
	}

	am.mu.Lock()
	am.dialogCtx[userID] = ctx
	am.mu.Unlock()

	return ctx
}

// GetDialogContext 获取对话上下文
func (am *AssistantManager) GetDialogContext(userID string) *DialogContext {
	am.mu.RLock()
	ctx := am.dialogCtx[userID]
	am.mu.RUnlock()
	return ctx
}

// ExportTemplates 导出模板为JSON
func (am *AssistantManager) ExportTemplates() ([]byte, error) {
	am.mu.RLock()
	data, err := json.MarshalIndent(am.templates, "", "  ")
	am.mu.RUnlock()
	return data, err
}
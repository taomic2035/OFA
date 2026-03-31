// Package nlp - 对话管理模块
package nlp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DialogState 对话状态
type DialogState string

const (
	StateInitial    DialogState = "initial"    // 初始状态
	StateIntentDetected DialogState = "intent_detected" // 意图已检测
	StateNeedClarify  DialogState = "need_clarify"  // 需要澄清
	StateProcessing   DialogState = "processing"   // 正在处理
	StateConfirming   DialogState = "confirming"   // 等待确认
	StateCompleted    DialogState = "completed"    // 已完成
	StateError        DialogState = "error"        // 错误状态
	StateFollowUp     DialogState = "follow_up"    // 后续追问
)

// DialogTurn 对话轮次
type DialogTurn struct {
	ID           string                 `json:"id"`
	UserInput    string                 `json:"user_input"`
	Intent       *NLPIntent             `json:"intent"`
	Entities     []Entity               `json:"entities"`
	Response     string                 `json:"response"`
	Action       string                 `json:"action"`       // suggested action
	State        DialogState            `json:"state"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

// DialogSession 对话会话
type DialogSession struct {
	ID           string        `json:"id"`
	UserID       string        `json:"user_id"`
	TenantID     string        `json:"tenant_id"`
	State        DialogState   `json:"state"`
	Turns        []*DialogTurn `json:"turns"`
	Context      DialogContext `json:"context"`
	IntentHistory []string     `json:"intent_history"` // 意图历史
	StartTime    time.Time     `json:"start_time"`
	LastActivity time.Time     `json:"last_activity"`
	ExpiresAt    time.Time     `json:"expires_at"`
	Completed    bool          `json:"completed"`
}

// DialogContext 对话上下文
type DialogContext struct {
	CurrentIntent     string                 `json:"current_intent"`
	PendingAction     string                 `json:"pending_action"`
	CollectedParams   map[string]interface{} `json:"collected_params"`
	MissingParams     []string               `json:"missing_params"`
	PreviousResults   map[string]interface{} `json:"previous_results"`
	UserPreferences   map[string]interface{} `json:"user_preferences"`
	EntityMemory      map[string][]Entity    `json:"entity_memory"` // 实体记忆
	TopicStack        []string               `json:"topic_stack"`   // 话题栈
	ClarificationCount int                   `json:"clarification_count"`
}

// Clarification 澄清请求
type Clarification struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`     // entity_missing, ambiguous, confirm, choice
	Question    string   `json:"question"` // 澄清问题
	Options     []string `json:"options"`  // 选项列表
	ExpectedEntity EntityType `json:"expected_entity"` // 期望的实体类型
	Context     string   `json:"context"`  // 上下文描述
	Priority    int      `json:"priority"` // 优先级
}

// Feedback 用户反馈
type Feedback struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	TurnID      string    `json:"turn_id"`
	Type        string    `json:"type"`     // positive, negative, correction, suggestion
	Content     string    `json:"content"`  // 反馈内容
	ActionTaken bool      `json:"action_taken"` // 是否采取行动
	Timestamp   time.Time `json:"timestamp"`
}

// DialogResponse 对话响应
type DialogResponse struct {
	SessionID     string        `json:"session_id"`
	TurnID        string        `json:"turn_id"`
	Response      string        `json:"response"`
	State         DialogState   `json:"state"`
	Clarification *Clarification `json:"clarification,omitempty"`
	Suggestions   []string      `json:"suggestions"`
	Action        string        `json:"action"`
	RequiresInput bool          `json:"requires_input"` // 需要用户输入
	TaskResult    map[string]interface{} `json:"task_result,omitempty"`
	Confidence    float64       `json:"confidence"`
	Timestamp     time.Time     `json:"timestamp"`
}

// DialogManagerConfig 对话管理器配置
type DialogManagerConfig struct {
	MaxTurnsPerSession  int           `json:"max_turns_per_session"`  // 最大轮次
	SessionTimeout      time.Duration `json:"session_timeout"`        // 会话超时
	MaxClarifications   int           `json:"max_clarifications"`     // 最大澄清次数
	AutoConfirm         bool          `json:"auto_confirm"`           // 自动确认
	ContextRetention    time.Duration `json:"context_retention"`      // 上下文保留时间
	EnableMultiIntent   bool          `json:"enable_multi_intent"`    // 多意图支持
	FeedbackEnabled     bool          `json:"feedback_enabled"`       // 反馈收集
}

// DialogManager 对话管理器
type DialogManager struct {
	config       DialogManagerConfig
	processor    *NLPProcessor
	sessions     map[string]*DialogSession
	feedbacks    []*Feedback
	responseTemplates map[string][]string // 响应模板
	mu           sync.RWMutex
}

// NewDialogManager 创建对话管理器
func NewDialogManager(config DialogManagerConfig, processor *NLPProcessor) *DialogManager {
	return &DialogManager{
		config:        config,
		processor:     processor,
		sessions:      make(map[string]*DialogSession),
		feedbacks:     make([]*Feedback, 0),
		responseTemplates: make(map[string][]string),
	}
}

// Initialize 初始化
func (dm *DialogManager) Initialize() error {
	// 加载响应模板
	dm.loadResponseTemplates()
	return nil
}

// loadResponseTemplates 加载响应模板
func (dm *DialogManager) loadResponseTemplates() {
	dm.mu.Lock()
	dm.responseTemplates = map[string][]string{
		"create-task-success": []string{
			"好的，我正在为您创建任务...",
			"任务已创建，正在执行中。",
			"已开始处理您的请求。",
		},
		"create-task-confirm": []string{
			"请确认您想要执行的任务：%s",
			"我理解您想要%s，请确认？",
		},
		"need-task-id": []string{
			"请提供任务ID，格式如：task-123",
			"请告诉我您要操作的任务编号。",
		},
		"need-skill": []string{
			"请选择一种处理方式：文本处理、JSON处理、数学计算等。",
			"您希望执行什么类型的任务？",
		},
		"need-input": []string{
			"请提供需要处理的内容。",
			"请告诉我具体要处理什么。",
		},
		"task-completed": []string{
			"任务已完成，结果如下：%s",
			"处理完成！%s",
		},
		"task-failed": []string{
			"任务执行失败：%s",
			"抱歉，执行过程中出现错误：%s",
		},
		"clarify-ambiguous": []string{
			"我理解您可能有多个意图，请选择：",
			"您是想%s还是%s？请帮我确认。",
		},
		"help-response": []string{
			"我可以帮您：创建任务、查询状态、获取推荐、诊断问题等。",
			"支持的操作包括：任务管理、状态查询、系统配置等。",
		},
		"unknown-intent": []string{
			"抱歉，我没有理解您的意思，请再说详细一些。",
			"请您描述得更具体一些，我才能更好地帮助您。",
		},
		"session-expired": []string{
			"会话已过期，请重新开始。",
			"长时间没有操作，请重新发起请求。",
		},
		"feedback-thanks": []string{
			"感谢您的反馈！",
			"您的建议已记录，谢谢！",
		},
	}
	dm.mu.Unlock()
}

// StartSession 开始会话
func (dm *DialogManager) StartSession(userID string, tenantID string) (*DialogSession, error) {
	session := &DialogSession{
		ID:           fmt.Sprintf("sess-%d", time.Now().UnixNano()),
		UserID:       userID,
		TenantID:     tenantID,
		State:        StateInitial,
		Turns:        make([]*DialogTurn, 0),
		IntentHistory: make([]string, 0),
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(dm.config.SessionTimeout),
		Completed:    false,
		Context: DialogContext{
			CollectedParams:  make(map[string]interface{}),
			MissingParams:    make([]string, 0),
			PreviousResults:  make(map[string]interface{}),
			UserPreferences:  make(map[string]interface{}),
			EntityMemory:     make(map[string][]Entity),
			TopicStack:       make([]string, 0),
		},
	}

	dm.mu.Lock()
	dm.sessions[session.ID] = session
	dm.mu.Unlock()

	return session, nil
}

// ProcessInput 处理用户输入
func (dm *DialogManager) ProcessInput(ctx context.Context, sessionID string, userInput string) (*DialogResponse, error) {
	dm.mu.RLock()
	session, ok := dm.sessions[sessionID]
	dm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		return dm.endSession(session), nil
	}

	// 检查轮次限制
	if len(session.Turns) >= dm.config.MaxTurnsPerSession {
		session.State = StateCompleted
		session.Completed = true
		return &DialogResponse{
			SessionID:   session.ID,
			Response:    "对话轮次已达到上限，请开始新会话。",
			State:       StateCompleted,
			RequiresInput: false,
			Timestamp:   time.Now(),
		}, nil
	}

	// 使用NLP处理器分析输入
	nlpResult, err := dm.processor.Process(ctx, userInput, "")
	if err != nil {
		return dm.handleError(session, err), nil
	}

	// 创建对话轮次
	turn := &DialogTurn{
		ID:         fmt.Sprintf("turn-%d-%d", len(session.Turns), time.Now().UnixNano()),
		UserInput:  userInput,
		Intent:     nlpResult.Intent,
		Entities:   nlpResult.Entities,
		State:      session.State,
		Confidence: nlpResult.Confidence,
		Metadata:   nlpResult.TaskParams,
		Timestamp:  time.Now(),
	}

	// 更新会话状态
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(dm.config.SessionTimeout)

	// 处理对话流程
	response := dm.processTurn(ctx, session, turn, nlpResult)

	// 保存轮次
	session.Turns = append(session.Turns, turn)

	return response, nil
}

// processTurn 处理对话轮次
func (dm *DialogManager) processTurn(ctx context.Context, session *DialogSession, turn *DialogTurn, nlpResult *NLPResult) *DialogResponse {
	// 低置信度处理
	if nlpResult.Confidence < 0.5 {
		session.Context.ClarificationCount++
		if session.Context.ClarificationCount > dm.config.MaxClarifications {
			return &DialogResponse{
				SessionID:   session.ID,
				TurnID:      turn.ID,
				Response:    dm.getResponse("unknown-intent"),
				State:       StateError,
				RequiresInput: true,
				Confidence:  nlpResult.Confidence,
				Timestamp:   time.Now(),
			}
		}

		return &DialogResponse{
			SessionID:   session.ID,
			TurnID:      turn.ID,
			Response:    dm.getResponse("unknown-intent"),
			State:       StateNeedClarify,
			RequiresInput: true,
			Confidence:  nlpResult.Confidence,
			Timestamp:   time.Now(),
		}
	}

	// 根据意图处理
	intentID := nlpResult.Intent.ID
	session.IntentHistory = append(session.IntentHistory, intentID)
	session.Context.CurrentIntent = intentID

	// 检查必需参数
	missing := dm.checkMissingParams(nlpResult.Intent, nlpResult.TaskParams)
	if len(missing) > 0 {
		session.Context.MissingParams = missing
		clarification := dm.generateClarification(nlpResult.Intent, missing)
		session.State = StateNeedClarify

		return &DialogResponse{
			SessionID:     session.ID,
			TurnID:        turn.ID,
			Response:      clarification.Question,
			State:         StateNeedClarify,
			Clarification: clarification,
			RequiresInput: true,
			Confidence:    nlpResult.Confidence,
			Timestamp:     time.Now(),
		}
	}

	// 合并上下文中的参数
	for k, v := range session.Context.CollectedParams {
		if nlpResult.TaskParams[k] == nil {
			nlpResult.TaskParams[k] = v
		}
	}

	// 更新上下文
	dm.updateContext(session, nlpResult)

	// 生成响应
	return dm.generateResponse(session, turn, nlpResult)
}

// checkMissingParams 检查缺失参数
func (dm *DialogManager) checkMissingParams(intent *NLPIntent, params map[string]interface{}) []string {
	missing := make([]string, 0)

	for _, entity := range intent.RequiredEntities {
		found := false
		entityMap, ok := params["entities"].(map[string][]string)
		if ok && len(entityMap[string(entity)]) > 0 {
			found = true
		}
		if !found && params[string(entity)] == nil {
			missing = append(missing, string(entity))
		}
	}

	return missing
}

// generateClarification 生成澄清请求
func (dm *DialogManager) generateClarification(intent *NLPIntent, missing []string) *Clarification {
	questions := map[string]string{
		EntityTaskID:    "请提供任务ID",
		EntitySkillName: "请选择一种操作类型",
		EntityParameter: "请提供必要的参数",
	}

	options := map[string][]string{
		EntitySkillName: []string{"文本处理", "JSON处理", "数学计算", "文件操作"},
	}

	question := ""
	expected := EntityTaskID
	optionList := []string{}

	for _, m := range missing {
		if questions[m] != "" {
			question = questions[m]
			expected = EntityType(m)
			if options[m] != nil {
				optionList = options[m]
			}
			break
		}
	}

	return &Clarification{
		ID:            fmt.Sprintf("clarify-%d", time.Now().UnixNano()),
		Type:          "entity_missing",
		Question:      question,
		Options:       optionList,
		ExpectedEntity: expected,
		Priority:      1,
	}
}

// updateContext 更新上下文
func (dm *DialogManager) updateContext(session *DialogSession, nlpResult *NLPResult) {
	// 保存实体到记忆
	for _, e := range nlpResult.Entities {
		key := string(e.Type)
		if session.Context.EntityMemory[key] == nil {
			session.Context.EntityMemory[key] = make([]Entity, 0)
		}
		session.Context.EntityMemory[key] = append(session.Context.EntityMemory[key], e)
	}

	// 保存收集的参数
	for k, v := range nlpResult.TaskParams {
		session.Context.CollectedParams[k] = v
	}

	// 更新话题栈
	if session.Context.CurrentIntent != "" {
		session.Context.TopicStack = append(session.Context.TopicStack, session.Context.CurrentIntent)
	}
}

// generateResponse 生成响应
func (dm *DialogManager) generateResponse(session *DialogSession, turn *DialogTurn, nlpResult *NLPResult) *DialogResponse {
	response := ""
	action := ""
	state := StateCompleted
	suggestions := []string{}
	taskResult := nlpResult.TaskParams

	switch nlpResult.Intent.ID {
	case "create-task":
		skill := nlpResult.TaskParams["skill"]
		if skill != nil {
			response = fmt.Sprintf("好的，我将为您执行%s任务。", skill)
			action = "execute_task"
			state = StateProcessing
			suggestions = []string{"查看进度", "修改参数"}
		} else {
			response = dm.getResponse("create-task-confirm")
			state = StateConfirming
			suggestions = []string{"文本处理", "JSON处理", "计算"}
		}

	case "query-status":
		taskID := nlpResult.TaskParams["task_id"]
		if taskID != nil {
			response = fmt.Sprintf("正在查询任务%s的状态...", taskID)
			action = "query_status"
			state = StateProcessing
		} else {
			response = dm.getResponse("need-task-id")
			state = StateNeedClarify
		}

	case "cancel-task":
		taskID := nlpResult.TaskParams["task_id"]
		if taskID != nil {
			response = fmt.Sprintf("正在取消任务%s...", taskID)
			action = "cancel_task"
			state = StateProcessing
		} else {
			response = dm.getResponse("need-task-id")
			state = StateNeedClarify
		}

	case "list-tasks":
		response = "正在获取任务列表..."
		action = "list_tasks"
		state = StateProcessing
		suggestions = []string{"查看详情", "筛选状态"}

	case "get-recommendation":
		response = "正在分析并生成推荐..."
		action = "get_recommendation"
		state = StateProcessing

	case "diagnose-error":
		response = "正在诊断问题..."
		action = "diagnose"
		state = StateProcessing

	case "system-info":
		response = "正在获取系统信息..."
		action = "system_info"
		state = StateProcessing

	case "get-help":
		response = dm.getResponse("help-response")
		state = StateCompleted
		suggestions = []string{"创建任务", "查询状态", "获取推荐"}

	case "config-setting":
		response = "正在更新配置..."
		action = "update_config"
		state = StateProcessing

	default:
		response = dm.getResponse("unknown-intent")
		state = StateNeedClarify
	}

	// 更新会话状态
	session.State = state
	turn.State = state
	turn.Response = response
	turn.Action = action

	return &DialogResponse{
		SessionID:   session.ID,
		TurnID:      turn.ID,
		Response:    response,
		State:       state,
		Suggestions: suggestions,
		Action:      action,
		RequiresInput: state == StateNeedClarify || state == StateConfirming,
		TaskResult:  taskResult,
		Confidence:  nlpResult.Confidence,
		Timestamp:   time.Now(),
	}
}

// getResponse 获取响应模板
func (dm *DialogManager) getResponse(templateKey string) string {
	dm.mu.RLock()
	templates, ok := dm.responseTemplates[templateKey]
	dm.mu.RUnlock()

	if !ok || len(templates) == 0 {
		return "请提供更多信息。"
	}

	// 随机选择一个模板
	idx := time.Now().Nanosecond() % len(templates)
	return templates[idx]
}

// handleError 处理错误
func (dm *DialogManager) handleError(session *DialogSession, err error) *DialogResponse {
	session.State = StateError

	return &DialogResponse{
		SessionID:   session.ID,
		Response:    fmt.Sprintf("处理出错：%s", err.Error()),
		State:       StateError,
		RequiresInput: true,
		Timestamp:   time.Now(),
	}
}

// endSession 结束会话
func (dm *DialogManager) endSession(session *DialogSession) *DialogResponse {
	session.State = StateCompleted
	session.Completed = true

	dm.mu.Lock()
	delete(dm.sessions, session.ID)
	dm.mu.Unlock()

	return &DialogResponse{
		SessionID:   session.ID,
		Response:    dm.getResponse("session-expired"),
		State:       StateCompleted,
		RequiresInput: false,
		Timestamp:   time.Now(),
	}
}

// ProvideFeedback 提供反馈
func (dm *DialogManager) ProvideFeedback(sessionID string, turnID string, feedbackType string, content string) error {
	if !dm.config.FeedbackEnabled {
		return fmt.Errorf("反馈功能未启用")
	}

	feedback := &Feedback{
		ID:        fmt.Sprintf("fb-%d", time.Now().UnixNano()),
		SessionID: sessionID,
		TurnID:    turnID,
		Type:      feedbackType,
		Content:   content,
		Timestamp: time.Now(),
	}

	dm.mu.Lock()
	dm.feedbacks = append(dm.feedbacks, feedback)
	dm.mu.Unlock()

	return nil
}

// GetSession 获取会话
func (dm *DialogManager) GetSession(sessionID string) (*DialogSession, error) {
	dm.mu.RLock()
	session, ok := dm.sessions[sessionID]
	dm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	return session, nil
}

// GetActiveSessions 获取活跃会话列表
func (dm *DialogManager) GetActiveSessions(userID string) []*DialogSession {
	dm.mu.RLock()
	sessions := make([]*DialogSession, 0)
	for _, s := range dm.sessions {
		if userID == "" || s.UserID == userID {
			if !s.Completed && !time.Now().After(s.ExpiresAt) {
				sessions = append(sessions, s)
			}
		}
	}
	dm.mu.RUnlock()
	return sessions
}

// CleanupExpiredSessions 清理过期会话
func (dm *DialogManager) CleanupExpiredSessions() int {
	dm.mu.Lock()
	count := 0
	now := time.Now()
	for id, s := range dm.sessions {
		if now.After(s.ExpiresAt) || s.Completed {
			delete(dm.sessions, id)
			count++
		}
	}
	dm.mu.Unlock()
	return count
}

// GetFeedbacks 获取反馈列表
func (dm *DialogManager) GetFeedbacks(limit int) []*Feedback {
	dm.mu.RLock()
	size := len(dm.feedbacks)
	if limit > size || limit <= 0 {
		limit = size
	}
	list := dm.feedbacks[size-limit:]
	dm.mu.RUnlock()
	return list
}

// AddResponseTemplate 添加响应模板
func (dm *DialogManager) AddResponseTemplate(key string, templates []string) error {
	dm.mu.Lock()
	dm.responseTemplates[key] = templates
	dm.mu.Unlock()
	return nil
}

// ResumeContext 恢复上下文（从历史会话）
func (dm *DialogManager) ResumeContext(session *DialogSession, previousSession *DialogSession) {
	// 复制实体记忆
	for k, v := range previousSession.Context.EntityMemory {
		session.Context.EntityMemory[k] = v
	}

	// 复制用户偏好
	for k, v := range previousSession.Context.UserPreferences {
		session.Context.UserPreferences[k] = v
	}

	// 复制历史结果
	for k, v := range previousSession.Context.PreviousResults {
		session.Context.PreviousResults[k] = v
	}
}

// ExportSession 导出会话数据
func (dm *DialogManager) ExportSession(sessionID string) ([]byte, error) {
	dm.mu.RLock()
	session, ok := dm.sessions[sessionID]
	dm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	return json.MarshalIndent(session, "", "  ")
}

// GetStatistics 获取统计信息
func (dm *DialogManager) GetStatistics() map[string]interface{} {
	dm.mu.RLock()
	activeCount := 0
	totalTurns := 0
	intentCounts := make(map[string]int)

	for _, s := range dm.sessions {
		if !s.Completed && !time.Now().After(s.ExpiresAt) {
			activeCount++
			totalTurns += len(s.Turns)
			for _, intent := range s.IntentHistory {
				intentCounts[intent]++
			}
		}
	}

	stats := map[string]interface{}{
		"active_sessions":  activeCount,
		"total_sessions":   len(dm.sessions),
		"total_turns":      totalTurns,
		"feedback_count":   len(dm.feedbacks),
		"intent_distribution": intentCounts,
		"template_count":   len(dm.responseTemplates),
	}
	dm.mu.RUnlock()
	return stats
}

// MultiIntentHandler 多意图处理
func (dm *DialogManager) MultiIntentHandler(ctx context.Context, sessionID string, intents []string) (*DialogResponse, error) {
	if !dm.config.EnableMultiIntent {
		return nil, fmt.Errorf("多意图处理未启用")
	}

	dm.mu.RLock()
	session, ok := dm.sessions[sessionID]
	dm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 生成意图选择澄清
	options := make([]string, 0)
	for _, intentID := range intents {
		intent, err := dm.processor.GetIntent(intentID)
		if err == nil {
			options = append(options, intent.Name)
		}
	}

	clarification := &Clarification{
		ID:       fmt.Sprintf("multi-%d", time.Now().UnixNano()),
		Type:     "choice",
		Question: "您有多种可能的意图，请选择：",
		Options:  options,
		Priority: 1,
	}

	return &DialogResponse{
		SessionID:     session.ID,
		Response:      "请选择您想要执行的操作：",
		State:         StateNeedClarify,
		Clarification: clarification,
		RequiresInput: true,
		Timestamp:     time.Now(),
	}, nil
}

// AnalyzeSentiment 分析情感（简化实现）
func (dm *DialogManager) AnalyzeSentiment(text string) string {
	positiveWords := []string{"好", "谢谢", "感谢", "满意", "great", "thanks", "good"}
	negativeWords := []string{"不好", "错误", "失败", "问题", "差", "bad", "error", "problem"}

	textLower := strings.ToLower(text)

	positiveScore := 0
	for _, w := range positiveWords {
		if strings.Contains(textLower, w) {
			positiveScore++
		}
	}

	negativeScore := 0
	for _, w := range negativeWords {
		if strings.Contains(textLower, w) {
			negativeScore++
		}
	}

	if positiveScore > negativeScore {
		return "positive"
	} else if negativeScore > positiveScore {
		return "negative"
	}
	return "neutral"
}
// Package llm - Prompt 模板管理
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

// PromptTemplate Prompt模板
type PromptTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Template    string            `json:"template"`
	Variables   []TemplateVariable `json:"variables"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Model       string            `json:"model,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, number, boolean, array, object
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Validation  string      `json:"validation,omitempty"` // regex pattern
}

// PromptManager Prompt管理器
type PromptManager struct {
	templates map[string]*PromptTemplate
	categories map[string][]string
	mu        sync.RWMutex
	config    PromptConfig
}

// PromptConfig Prompt配置
type PromptConfig struct {
	TemplateDir string `json:"template_dir"`
	AutoSave    bool   `json:"auto_save"`
}

// DefaultPromptConfig 默认配置
func DefaultPromptConfig() PromptConfig {
	return PromptConfig{
		TemplateDir: "./prompts",
		AutoSave:    true,
	}
}

// NewPromptManager 创建Prompt管理器
func NewPromptManager(config PromptConfig) *PromptManager {
	return &PromptManager{
		templates:  make(map[string]*PromptTemplate),
		categories: make(map[string][]string),
		config:     config,
	}
}

// Register 注册模板
func (m *PromptManager) Register(tmpl *PromptTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tmpl.ID == "" {
		return fmt.Errorf("模板ID不能为空")
	}

	if tmpl.CreatedAt.IsZero() {
		tmpl.CreatedAt = time.Now()
	}
	tmpl.UpdatedAt = time.Now()

	m.templates[tmpl.ID] = tmpl

	// 分类索引
	for _, tag := range tmpl.Tags {
		m.categories[tag] = append(m.categories[tag], tmpl.ID)
	}

	return nil
}

// Get 获取模板
func (m *PromptManager) Get(id string) (*PromptTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tmpl, ok := m.templates[id]
	if !ok {
		return nil, fmt.Errorf("模板不存在: %s", id)
	}

	return tmpl, nil
}

// Delete 删除模板
func (m *PromptManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tmpl, ok := m.templates[id]
	if !ok {
		return fmt.Errorf("模板不存在: %s", id)
	}

	// 从分类中移除
	for _, tag := range tmpl.Tags {
		m.categories[tag] = removeString(m.categories[tag], id)
	}

	delete(m.templates, id)
	return nil
}

// Render 渲染模板
func (m *PromptManager) Render(id string, variables map[string]interface{}) (string, error) {
	tmpl, err := m.Get(id)
	if err != nil {
		return "", err
	}

	// 合并默认值
	merged := m.mergeDefaults(tmpl, variables)

	// 验证必需变量
	if err := m.validateVariables(tmpl, merged); err != nil {
		return "", err
	}

	// 解析模板
	t, err := template.New(id).Parse(tmpl.Template)
	if err != nil {
		return "", fmt.Errorf("模板解析失败: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, merged); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}

	return buf.String(), nil
}

// RenderMessages 渲染为消息列表
func (m *PromptManager) RenderMessages(id string, variables map[string]interface{}) ([]Message, error) {
	tmpl, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	content, err := m.Render(id, variables)
	if err != nil {
		return nil, err
	}

	messages := []Message{}

	// 添加系统提示
	if tmpl.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: tmpl.SystemPrompt,
		})
	}

	// 添加用户消息
	messages = append(messages, Message{
		Role:    "user",
		Content: content,
	})

	return messages, nil
}

// mergeDefaults 合并默认值
func (m *PromptManager) mergeDefaults(tmpl *PromptTemplate, variables map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// 设置默认值
	for _, v := range tmpl.Variables {
		if v.Default != nil {
			merged[v.Name] = v.Default
		}
	}

	// 覆盖用户值
	for k, v := range variables {
		merged[k] = v
	}

	return merged
}

// validateVariables 验证变量
func (m *PromptManager) validateVariables(tmpl *PromptTemplate, variables map[string]interface{}) error {
	for _, v := range tmpl.Variables {
		if v.Required {
			if _, ok := variables[v.Name]; !ok {
				return fmt.Errorf("缺少必需变量: %s", v.Name)
			}
		}
	}
	return nil
}

// List 列出模板
func (m *PromptManager) List() []*PromptTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	templates := make([]*PromptTemplate, 0, len(m.templates))
	for _, tmpl := range m.templates {
		templates = append(templates, tmpl)
	}
	return templates
}

// Search 搜索模板
func (m *PromptManager) Search(query string) []*PromptTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]*PromptTemplate, 0)

	for _, tmpl := range m.templates {
		if strings.Contains(strings.ToLower(tmpl.Name), query) ||
			strings.Contains(strings.ToLower(tmpl.Description), query) {
			results = append(results, tmpl)
		}
	}

	return results
}

// ByTag 按标签获取
func (m *PromptManager) ByTag(tag string) []*PromptTemplate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := m.categories[tag]
	templates := make([]*PromptTemplate, 0, len(ids))

	for _, id := range ids {
		if tmpl, ok := m.templates[id]; ok {
			templates = append(templates, tmpl)
		}
	}

	return templates
}

// Save 保存模板到文件
func (m *PromptManager) Save(id string) error {
	m.mu.RLock()
	tmpl, ok := m.templates[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("模板不存在: %s", id)
	}

	os.MkdirAll(m.config.TemplateDir, 0755)

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	path := filepath.Join(m.config.TemplateDir, id+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	return nil
}

// Load 从文件加载模板
func (m *PromptManager) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取失败: %w", err)
	}

	var tmpl PromptTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	return m.Register(&tmpl)
}

// LoadAll 加载所有模板
func (m *PromptManager) LoadAll() error {
	files, err := filepath.Glob(filepath.Join(m.config.TemplateDir, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := m.Load(file); err != nil {
			return err
		}
	}

	return nil
}

// 预置模板
var DefaultTemplates = []*PromptTemplate{
	{
		ID:          "agent-task",
		Name:        "Agent任务处理",
		Description: "通用Agent任务处理模板",
		Template:    "请执行以下任务：\n\n任务类型：{{.taskType}}\n任务描述：{{.description}}\n\n请提供详细的结果。",
		SystemPrompt: "你是一个智能Agent助手，负责执行各种任务。请根据任务类型和描述，提供准确、详细的执行结果。",
		Variables: []TemplateVariable{
			{Name: "taskType", Type: "string", Required: true, Description: "任务类型"},
			{Name: "description", Type: "string", Required: true, Description: "任务描述"},
		},
		Tags: []string{"agent", "task"},
	},
	{
		ID:          "code-review",
		Name:        "代码审查",
		Description: "代码审查模板",
		Template:    "请审查以下代码：\n\n```{{.language}}\n{{.code}}\n```\n\n请从以下方面进行审查：\n1. 代码质量\n2. 潜在问题\n3. 改进建议",
		SystemPrompt: "你是一位资深的代码审查专家。请从代码质量、安全性、可维护性等方面进行审查，并提供具体的改进建议。",
		Variables: []TemplateVariable{
			{Name: "code", Type: "string", Required: true, Description: "要审查的代码"},
			{Name: "language", Type: "string", Required: false, Default: "go", Description: "编程语言"},
		},
		Tags: []string{"code", "review"},
	},
	{
		ID:          "summarize",
		Name:        "内容摘要",
		Description: "文本内容摘要模板",
		Template:    "请对以下内容进行摘要：\n\n{{.content}}\n\n摘要要求：\n- 控制在{{.maxLength}}字以内\n- 突出重点信息",
		SystemPrompt: "你是一位专业的内容编辑，擅长提取和总结关键信息。",
		Variables: []TemplateVariable{
			{Name: "content", Type: "string", Required: true, Description: "要摘要的内容"},
			{Name: "maxLength", Type: "number", Required: false, Default: 200, Description: "最大字数"},
		},
		Tags: []string{"text", "summary"},
	},
	{
		ID:          "translate",
		Name:        "翻译",
		Description: "多语言翻译模板",
		Template:    "请将以下{{.sourceLang}}文本翻译成{{.targetLang}}：\n\n{{.text}}",
		SystemPrompt: "你是一位专业的翻译专家，精通多种语言。请提供准确、自然的翻译。",
		Variables: []TemplateVariable{
			{Name: "text", Type: "string", Required: true, Description: "要翻译的文本"},
			{Name: "sourceLang", Type: "string", Required: true, Description: "源语言"},
			{Name: "targetLang", Type: "string", Required: true, Description: "目标语言"},
		},
		Tags: []string{"translation", "multilingual"},
	},
	{
		ID:          "analyze",
		Name:        "数据分析",
		Description: "数据分析模板",
		Template:    "请分析以下数据：\n\n{{.data}}\n\n分析要求：{{.requirements}}",
		SystemPrompt: "你是一位数据分析专家，请对提供的数据进行深入分析，并提供见解和建议。",
		Variables: []TemplateVariable{
			{Name: "data", Type: "string", Required: true, Description: "数据内容"},
			{Name: "requirements", Type: "string", Required: false, Default: "请进行全面分析", Description: "分析要求"},
		},
		Tags: []string{"data", "analysis"},
	},
}

// InitDefaultTemplates 初始化默认模板
func (m *PromptManager) InitDefaultTemplates() {
	for _, tmpl := range DefaultTemplates {
		m.Register(tmpl)
	}
}

// removeString 从切片中移除字符串
func removeString(slice []string, s string) []string {
	result := make([]string, 0)
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// ContextManager 上下文管理器
type ContextManager struct {
	contexts  map[string]*ConversationContext
	maxTokens int
	mu        sync.RWMutex
}

// ConversationContext 对话上下文
type ConversationContext struct {
	ID           string         `json:"id"`
	Messages     []Message      `json:"messages"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	LastActiveAt time.Time      `json:"last_active_at"`
	TokenCount   int            `json:"token_count"`
}

// NewContextManager 创建上下文管理器
func NewContextManager(maxTokens int) *ContextManager {
	return &ContextManager{
		contexts:  make(map[string]*ConversationContext),
		maxTokens: maxTokens,
	}
}

// Create 创建新上下文
func (m *ContextManager) Create(id string) *ConversationContext {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := &ConversationContext{
		ID:           id,
		Messages:     make([]Message, 0),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	m.contexts[id] = ctx
	return ctx
}

// Get 获取上下文
func (m *ContextManager) Get(id string) (*ConversationContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, ok := m.contexts[id]
	if !ok {
		return nil, fmt.Errorf("上下文不存在: %s", id)
	}

	return ctx, nil
}

// AddMessage 添加消息
func (m *ContextManager) AddMessage(id string, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.contexts[id]
	if !ok {
		ctx = &ConversationContext{
			ID:           id,
			Messages:     make([]Message, 0),
			Metadata:     make(map[string]interface{}),
			CreatedAt:    time.Now(),
		}
		m.contexts[id] = ctx
	}

	ctx.Messages = append(ctx.Messages, msg)
	ctx.LastActiveAt = time.Now()

	// 检查是否需要截断
	m.truncateIfNeeded(ctx)

	return nil
}

// GetMessages 获取消息
func (m *ContextManager) GetMessages(id string) ([]Message, error) {
	ctx, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	return ctx.Messages, nil
}

// Clear 清除上下文
func (m *ContextManager) Clear(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.contexts, id)
	return nil
}

// truncateIfNeeded 必要时截断
func (m *ContextManager) truncateIfNeeded(ctx *ConversationContext) {
	// 简单实现：保留最近的N条消息
	maxMessages := m.maxTokens / 100 // 假设每条消息平均100 token

	if len(ctx.Messages) > maxMessages {
		// 保留系统消息和最近的消息
		systemMessages := []Message{}
		userMessages := []Message{}

		for _, msg := range ctx.Messages {
			if msg.Role == "system" {
				systemMessages = append(systemMessages, msg)
			} else {
				userMessages = append(userMessages, msg)
			}
		}

		// 保留最近的用户消息
		if len(userMessages) > maxMessages-1 {
			userMessages = userMessages[len(userMessages)-maxMessages+1:]
		}

		ctx.Messages = append(systemMessages, userMessages...)
	}
}

// List 列出所有上下文
func (m *ContextManager) List() []*ConversationContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contexts := make([]*ConversationContext, 0, len(m.contexts))
	for _, ctx := range m.contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}
// Package nlp - 自然语言处理模块
package nlp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Language 支持的语言
type Language string

const (
	LangChinese  Language = "zh"  // 中文
	LangEnglish  Language = "en"  // 英文
	LangJapanese Language = "ja"  // 日文
	LangKorean   Language = "ko"  // 韩文
	LangGerman   Language = "de"  // 德文
	LangFrench   Language = "fr"  // 法文
)

// EntityType 实体类型
type EntityType string

const (
	EntityTaskID    EntityType = "task_id"
	EntityAgentID   EntityType = "agent_id"
	EntitySkillName EntityType = "skill_name"
	EntityParameter EntityType = "parameter"
	EntityTime      EntityType = "time"
	EntityNumber    EntityType = "number"
	EntityURL       EntityType = "url"
	EntityFilePath  EntityType = "file_path"
	EntityCommand   EntityType = "command"
	EntityStatus    EntityType = "status"
	EntityPriority  EntityType = "priority"
)

// IntentCategory 意图类别
type IntentCategory string

const (
	CategoryTask      IntentCategory = "task"      // 任务相关
	CategoryQuery     IntentCategory = "query"     // 查询相关
	CategoryConfig    IntentCategory = "config"    // 配置相关
	CategorySystem    IntentCategory = "system"    // 系统相关
	CategoryHelp      IntentCategory = "help"      // 帮助相关
	CategoryFeedback  IntentCategory = "feedback"  // 反馈相关
)

// NLPIntent NLP意图
type NLPIntent struct {
	ID           string            `json:"id"`
	Category     IntentCategory    `json:"category"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Keywords     []string          `json:"keywords"`     // 关键词列表
	Patterns     []string          `json:"patterns"`     // 正则模式
	RequiredEntities []EntityType  `json:"required_entities"` // 必需实体
	OptionalEntities []EntityType  `json:"optional_entities"` // 可选实体
	Examples     []string          `json:"examples"`     // 示例句子
	Confidence   float64           `json:"confidence"`
}

// Entity 实体
type Entity struct {
	ID       string      `json:"id"`
	Type     EntityType  `json:"type"`
	Name     string      `json:"name"`     // 实体名称
	Value    string      `json:"value"`    // 提取的值
	Original string      `json:"original"` // 原始文本
	Position int         `json:"position"` // 位置索引
	Length   int         `json:"length"`   // 长度
	Confidence float64   `json:"confidence"`
}

// NLPResult NLP解析结果
type NLPResult struct {
	ID          string      `json:"id"`
	Input       string      `json:"input"`
	Language    Language    `json:"language"`
	Intent      *NLPIntent  `json:"intent"`
	Entities    []Entity    `json:"entities"`
	TaskParams  map[string]interface{} `json:"task_params"` // 提取的任务参数
	Normalized  string      `json:"normalized"` // 规范化文本
	Confidence  float64     `json:"confidence"` // 总体置信度
	Tokens      []string    `json:"tokens"`     // 分词结果
	Timestamp   time.Time   `json:"timestamp"`
}

// TaskSpec 任务规格
type TaskSpec struct {
	TaskType    string                 `json:"task_type"`
	Skills      []string               `json:"skills"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Description string                 `json:"description"`
	Timeout     int                    `json:"timeout"` // 毫秒
}

// NLPProcessorConfig NLP处理器配置
type NLPProcessorConfig struct {
	DefaultLanguage    Language    `json:"default_language"`
	MinConfidence      float64     `json:"min_confidence"`      // 最小置信度阈值
	EnableMultiLang    bool        `json:"enable_multi_lang"`   // 多语言支持
	EnableEntityMerge  bool        `json:"enable_entity_merge"` // 实体合并
	ContextWindowSize  int         `json:"context_window_size"` // 上下文窗口
	DebugMode          bool        `json:"debug_mode"`
}

// NLPProcessor NLP处理器
type NLPProcessor struct {
	config      NLPProcessorConfig
	intents     map[string]*NLPIntent
	entityPatterns map[EntityType][]*regexp.Regexp
	skillKeywords map[string][]string
	taskTemplates map[string]*TaskSpec
	mu          sync.RWMutex
}

// NewNLPProcessor 创建NLP处理器
func NewNLPProcessor(config NLPProcessorConfig) *NLPProcessor {
	return &NLPProcessor{
		config:        config,
		intents:       make(map[string]*NLPIntent),
		entityPatterns: make(map[EntityType][]*regexp.Regexp),
		skillKeywords: make(map[string][]string),
		taskTemplates: make(map[string]*TaskSpec),
	}
}

// Initialize 初始化
func (np *NLPProcessor) Initialize() error {
	// 加载意图定义
	np.loadIntents()
	// 加载实体模式
	np.loadEntityPatterns()
	// 加载技能关键词
	np.loadSkillKeywords()
	// 加载任务模板
	np.loadTaskTemplates()
	return nil
}

// loadIntents 加载意图定义
func (np *NLPProcessor) loadIntents() {
	intents := []NLPIntent{
		{
			ID:          "create-task",
			Category:    CategoryTask,
			Name:        "创建任务",
			Description: "创建并执行一个新任务",
			Keywords:    []string{"创建", "执行", "运行", "帮我", "处理", "计算", "分析", "create", "run", "execute", "help"},
			Patterns:    []string{"帮我.*", "创建.*任务", "执行.*", "计算.*", "分析.*"},
			RequiredEntities: []EntityType{EntitySkillName},
			OptionalEntities: []EntityType{EntityParameter, EntityPriority},
			Examples: []string{"帮我计算123+456", "创建一个文本分析任务", "执行JSON格式化"},
		},
		{
			ID:          "query-status",
			Category:    CategoryQuery,
			Name:        "查询状态",
			Description: "查询任务或系统状态",
			Keywords:    []string{"状态", "进度", "结果", "查询", "status", "progress", "query"},
			Patterns:    []string{"查询.*状态", ".*进度", ".*结果", "status.*"},
			RequiredEntities: []EntityType{EntityTaskID},
			OptionalEntities: []EntityType{},
			Examples: []string{"查询任务123的状态", "task-abc的进度"},
		},
		{
			ID:          "cancel-task",
			Category:    CategoryTask,
			Name:        "取消任务",
			Description: "取消正在执行的任务",
			Keywords:    []string{"取消", "停止", "终止", "cancel", "stop", "abort"},
			Patterns:    []string{"取消.*", "停止.*", "终止.*"},
			RequiredEntities: []EntityType{EntityTaskID},
			OptionalEntities: []EntityType{},
			Examples: []string{"取消任务123", "停止task-abc"},
		},
		{
			ID:          "modify-task",
			Category:    CategoryTask,
			Name:        "修改任务",
			Description: "修改任务参数或配置",
			Keywords:    []string{"修改", "更新", "调整", "modify", "update", "adjust"},
			Patterns:    []string{"修改.*", "更新.*", "调整.*"},
			RequiredEntities: []EntityType{EntityTaskID, EntityParameter},
			OptionalEntities: []EntityType{},
			Examples: []string{"修改任务123的参数", "更新超时时间"},
		},
		{
			ID:          "list-tasks",
			Category:    CategoryQuery,
			Name:        "列出任务",
			Description: "列出所有或特定类型的任务",
			Keywords:    []string{"列出", "显示", "所有", "list", "show", "all"},
			Patterns:    []string{"列出.*", "显示.*", "所有任务"},
			RequiredEntities: []EntityType{},
			OptionalEntities: []EntityType{EntityStatus, EntitySkillName},
			Examples: []string{"列出所有任务", "显示正在运行的任务"},
		},
		{
			ID:          "get-recommendation",
			Category:    CategoryQuery,
			Name:        "获取推荐",
			Description: "获取Agent或策略推荐",
			Keywords:    []string{"推荐", "选择", "哪个", "recommend", "suggest", "which"},
			Patterns:    []string{"推荐.*", "选择.*", "哪个.*"},
			RequiredEntities: []EntityType{},
			OptionalEntities: []EntityType{EntitySkillName},
			Examples: []string{"推荐一个Agent", "哪个策略最好"},
		},
		{
			ID:          "diagnose-error",
			Category:    CategorySystem,
			Name:        "诊断错误",
			Description: "诊断系统或任务错误",
			Keywords:    []string{"错误", "问题", "失败", "异常", "诊断", "error", "problem", "diagnose"},
			Patterns:    []string{".*错误", ".*问题", "诊断.*"},
			RequiredEntities: []EntityType{},
			OptionalEntities: []EntityType{EntityTaskID, EntityCommand},
			Examples: []string{"诊断这个错误", "任务失败的原因"},
		},
		{
			ID:          "system-info",
			Category:    CategorySystem,
			Name:        "系统信息",
			Description: "获取系统状态和信息",
			Keywords:    []string{"系统", "信息", "状态", "system", "info", "status"},
			Patterns:    []string{"系统.*", ".*信息"},
			RequiredEntities: []EntityType{},
			OptionalEntities: []EntityType{},
			Examples: []string{"系统信息", "当前状态"},
		},
		{
			ID:          "get-help",
			Category:    CategoryHelp,
			Name:        "获取帮助",
			Description: "获取使用帮助和文档",
			Keywords:    []string{"帮助", "怎么", "如何", "help", "how", "what"},
			Patterns:    []string{"怎么.*", "如何.*", "帮助"},
			RequiredEntities: []EntityType{},
			OptionalEntities: []EntityType{EntitySkillName},
			Examples: []string{"如何创建任务", "帮助"},
		},
		{
			ID:          "config-setting",
			Category:    CategoryConfig,
			Name:        "配置设置",
			Description: "修改系统配置",
			Keywords:    []string{"配置", "设置", "config", "setting", "set"},
			Patterns:    []string{"配置.*", "设置.*"},
			RequiredEntities: []EntityType{EntityParameter},
			OptionalEntities: []EntityType{},
			Examples: []string{"配置超时时间", "设置默认策略"},
		},
	}

	np.mu.Lock()
	for _, intent := range intents {
		np.intents[intent.ID] = &intent
	}
	np.mu.Unlock()
}

// loadEntityPatterns 加载实体模式
func (np *NLPProcessor) loadEntityPatterns() {
	patterns := map[EntityType][]string{
		EntityTaskID: {
			`task[-_]?\d+`,
			`task[-_]?[a-zA-Z0-9]+`,
			`\d{3,}`,
		},
		EntityAgentID: {
			`agent[-_]?\d+`,
			`agent[-_]?[a-zA-Z0-9]+`,
		},
		EntitySkillName: {
			`text\.process`,
			`json\.process`,
			`calculator`,
			`echo`,
			`文件操作`,
			`文本处理`,
			`JSON`,
			`计算`,
		},
		EntityNumber: {
			`\d+\.?\d*`,
		},
		EntityTime: {
			`\d{1,2}:\d{2}`,
			`\d{4}-\d{2}-\d{2}`,
			`今天|明天|每天|每小时`,
		},
		EntityFilePath: {
			`[a-zA-Z]:\\[^:]*`,
			`/[^:]*`,
			`.*\.(txt|json|csv|log)`,
		},
		EntityURL: {
			`https?://[^\s]+`,
		},
		EntityPriority: {
			`高优先|低优先|紧急|普通`,
			`priority[-_]?high`,
			`priority[-_]?low`,
		},
		EntityStatus: {
			`运行|完成|失败|等待|pending|running|completed|failed`,
		},
		EntityCommand: {
			`run|start|stop|restart|reload`,
		},
	}

	np.mu.Lock()
	for entityType, patternList := range patterns {
		np.entityPatterns[entityType] = make([]*regexp.Regexp, 0)
		for _, p := range patternList {
			re, err := regexp.Compile(p)
			if err == nil {
				np.entityPatterns[entityType] = append(np.entityPatterns[entityType], re)
			}
		}
	}
	np.mu.Unlock()
}

// loadSkillKeywords 加载技能关键词
func (np *NLPProcessor) loadSkillKeywords() {
	np.mu.Lock()
	np.skillKeywords = map[string][]string{
		"text.process":   []string{"文本", "字符串", "大写", "小写", "反转", "长度", "text", "string", "uppercase", "lowercase"},
		"json.process":   []string{"JSON", "格式化", "解析", "键", "值", "json", "format", "parse"},
		"calculator":     []string{"计算", "加减", "乘除", "数学", "calculate", "math", "add", "sub", "mul", "div"},
		"echo":           []string{"回显", "测试", "返回", "echo", "test"},
		"file.operation": []string{"文件", "读取", "写入", "删除", "file", "read", "write", "delete"},
		"command.execute": []string{"命令", "执行", "脚本", "command", "execute", "script"},
	}
	np.mu.Unlock()
}

// loadTaskTemplates 加载任务模板
func (np *NLPProcessor) loadTaskTemplates() {
	np.mu.Lock()
	np.taskTemplates = map[string]*TaskSpec{
		"calculate": {
			TaskType:    "calculator",
			Skills:      []string{"calculator"},
			Parameters:  map[string]interface{}{"operation": "auto"},
			Priority:    5,
			Description: "数学计算任务",
			Timeout:     5000,
		},
		"text_process": {
			TaskType:    "text.process",
			Skills:      []string{"text.process"},
			Parameters:  map[string]interface{}{"operation": "auto"},
			Priority:    5,
			Description: "文本处理任务",
			Timeout:     3000,
		},
		"json_format": {
			TaskType:    "json.process",
			Skills:      []string{"json.process"},
			Parameters:  map[string]interface{}{"operation": "pretty"},
			Priority:    5,
			Description: "JSON格式化任务",
			Timeout:     3000,
		},
		"batch": {
			TaskType:    "batch",
			Skills:      []string{"text.process", "json.process"},
			Parameters:  map[string]interface{}{"parallel": true},
			Priority:    6,
			Description: "批量处理任务",
			Timeout:     30000,
		},
	}
	np.mu.Unlock()
}

// Process 处理自然语言输入
func (np *NLPProcessor) Process(ctx context.Context, input string, lang Language) (*NLPResult, error) {
	// 检测语言
	if lang == "" {
		lang = np.detectLanguage(input)
	}

	// 分词
	tokens := np.tokenize(input, lang)

	// 规范化
	normalized := np.normalize(input, lang)

	// 意图识别
	intent, confidence := np.recognizeIntent(normalized, tokens)

	// 实体抽取
	entities := np.extractEntities(input, normalized)

	// 构建任务参数
	taskParams := np.buildTaskParams(intent, entities, normalized)

	result := &NLPResult{
		ID:         fmt.Sprintf("nlp-%d", time.Now().UnixNano()),
		Input:      input,
		Language:   lang,
		Intent:     intent,
		Entities:   entities,
		TaskParams: taskParams,
		Normalized: normalized,
		Confidence: confidence,
		Tokens:     tokens,
		Timestamp:  time.Now(),
	}

	return result, nil
}

// detectLanguage 检测语言
func (np *NLPProcessor) detectLanguage(input string) Language {
	// 简化的语言检测
	chinesePattern := regexp.MustCompile(`[\p{Han}]`)
	if chinesePattern.MatchString(input) {
		return LangChinese
	}

	// 检查常见英文词
	englishWords := []string{"the", "is", "are", "create", "run", "execute", "help", "status"}
	inputLower := strings.ToLower(input)
	for _, word := range englishWords {
		if strings.Contains(inputLower, word) {
			return LangEnglish
		}
	}

	return np.config.DefaultLanguage
}

// tokenize 分词
func (np *NLPProcessor) tokenize(input string, lang Language) []string {
	var tokens []string

	switch lang {
	case LangChinese:
		// 中文分词（简化实现）
		// 按空格和标点分割
		separators := regexp.MustCompile(`[\s,，。！？；：""''（）【】]+`)
		tokens = separators.Split(input, -1)
		// 过滤空token
		tokens = filterEmpty(tokens)
	case LangEnglish:
		// 英文分词
		separators := regexp.MustCompile(`[\s,.!?;:()]+`)
		tokens = separators.Split(strings.ToLower(input), -1)
		tokens = filterEmpty(tokens)
	default:
		separators := regexp.MustCompile(`[\s]+`)
		tokens = separators.Split(input, -1)
		tokens = filterEmpty(tokens)
	}

	return tokens
}

// normalize 规范化文本
func (np *NLPProcessor) normalize(input string, lang Language) string {
	// 去除多余空格
	input = strings.TrimSpace(input)
	// 统一标点
	if lang == LangChinese {
		input = strings.ReplaceAll(input, "，", ",")
		input = strings.ReplaceAll(input, "。", ".")
	}
	// 小写化（英文）
	if lang == LangEnglish {
		input = strings.ToLower(input)
	}
	return input
}

// recognizeIntent 意图识别
func (np *NLPProcessor) recognizeIntent(normalized string, tokens []string) (*NLPIntent, float64) {
	np.mu.RLock()
	defer np.mu.RUnlock()

	bestIntent := ""
	bestScore := 0.0

	for id, intent := range np.intents {
		score := 0.0

		// 关键词匹配
		keywordScore := np.matchKeywords(normalized, intent.Keywords)
		score += keywordScore * 0.4

		// 模式匹配
		patternScore := np.matchPatterns(normalized, intent.Patterns)
		score += patternScore * 0.4

		// Token匹配
		tokenScore := np.matchTokens(tokens, intent.Keywords)
		score += tokenScore * 0.2

		if score > bestScore {
			bestScore = score
			bestIntent = id
		}
	}

	if bestIntent == "" || bestScore < np.config.MinConfidence {
		// 返回默认意图
		return np.intents["create-task"], bestScore
	}

	intent := np.intents[bestIntent]
	intent.Confidence = bestScore
	return intent, bestScore
}

// matchKeywords 关键词匹配
func (np *NLPProcessor) matchKeywords(text string, keywords []string) float64 {
	matchCount := 0
	for _, kw := range keywords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(kw)) {
			matchCount++
		}
	}
	if len(keywords) == 0 {
		return 0
	}
	return float64(matchCount) / float64(len(keywords))
}

// matchPatterns 模式匹配
func (np *NLPProcessor) matchPatterns(text string, patterns []string) float64 {
	matchCount := 0
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err == nil && re.MatchString(text) {
			matchCount++
		}
	}
	if len(patterns) == 0 {
		return 0
	}
	return float64(matchCount) / float64(len(patterns))
}

// matchTokens Token匹配
func (np *NLPProcessor) matchTokens(tokens []string, keywords []string) float64 {
	matchCount := 0
	for _, token := range tokens {
		for _, kw := range keywords {
			if strings.ToLower(token) == strings.ToLower(kw) {
				matchCount++
			}
		}
	}
	if len(keywords) == 0 || len(tokens) == 0 {
		return 0
	}
	return float64(matchCount) / float64(len(tokens))
}

// extractEntities 实体抽取
func (np *NLPProcessor) extractEntities(original string, normalized string) []Entity {
	np.mu.RLock()
	defer np.mu.RUnlock()

	entities := make([]Entity, 0)

	for entityType, patterns := range np.entityPatterns {
		for _, re := range patterns {
			matches := re.FindAllStringIndex(normalized, -1)
			for _, match := range matches {
				value := normalized[match[0]:match[1]]
				// 在原文中找对应位置
				origPos := strings.Index(original, value)
				if origPos == -1 {
					origPos = match[0]
				}

				entity := Entity{
					ID:         fmt.Sprintf("ent-%d-%d", entityType, time.Now().UnixNano()),
					Type:       entityType,
					Name:       entityTypeToName(entityType),
					Value:      value,
					Original:   value,
					Position:   origPos,
					Length:     len(value),
					Confidence: 0.8,
				}
				entities = append(entities, entity)
			}
		}
	}

	// 合并相同位置的实体
	if np.config.EnableEntityMerge {
		entities = np.mergeEntities(entities)
	}

	return entities
}

// mergeEntities 合并实体
func (np *NLPProcessor) mergeEntities(entities []Entity) []Entity {
	merged := make([]Entity, 0)
	used := make(map[int]bool)

	for _, e1 := range entities {
		if used[e1.Position] {
			continue
		}

		bestEntity := e1
		for _, e2 := range entities {
			if e2.Position == e1.Position && e2.Confidence > bestEntity.Confidence {
				bestEntity = e2
			}
		}

		merged = append(merged, bestEntity)
		used[bestEntity.Position] = true
	}

	return merged
}

// buildTaskParams 构建任务参数
func (np *NLPProcessor) buildTaskParams(intent *NLPIntent, entities []Entity, normalized string) map[string]interface{} {
	params := make(map[string]interface{})

	if intent == nil {
		return params
	}

	// 根据意图类型提取参数
	switch intent.ID {
	case "create-task":
		// 查找技能
		skill := np.findSkill(normalized, entities)
		if skill != "" {
			params["skill"] = skill
		}

		// 查找输入内容
		input := np.extractInput(normalized)
		if input != "" {
			params["input"] = input
		}

		// 查找优先级
		priority := np.extractPriority(entities)
		if priority > 0 {
			params["priority"] = priority
		}

	case "query-status", "cancel-task", "modify-task":
		// 查找任务ID
		taskID := np.extractTaskID(entities)
		if taskID != "" {
			params["task_id"] = taskID
		}

	case "config-setting":
		// 提取配置参数
		for _, e := range entities {
			if e.Type == EntityParameter {
				params["parameter"] = e.Value
			}
		}
	}

	// 添加检测到的所有实体
	entityMap := make(map[string][]string)
	for _, e := range entities {
		key := string(e.Type)
		if entityMap[key] == nil {
			entityMap[key] = make([]string, 0)
		}
		entityMap[key] = append(entityMap[key], e.Value)
	}
	params["entities"] = entityMap

	return params
}

// findSkill 查找技能
func (np *NLPProcessor) findSkill(text string, entities []Entity) string {
	np.mu.RLock()
	defer np.mu.RUnlock()

	// 检查实体
	for _, e := range entities {
		if e.Type == EntitySkillName {
			return e.Value
		}
	}

	// 关键词匹配
	for skill, keywords := range np.skillKeywords {
		for _, kw := range keywords {
			if strings.Contains(strings.ToLower(text), strings.ToLower(kw)) {
				return skill
			}
		}
	}

	return ""
}

// extractInput 提取输入内容
func (np *NLPProcessor) extractInput(text string) string {
	// 尝试提取"帮我"后面的内容
	helpPattern := regexp.MustCompile(`帮我\s*(.+)`)
	matches := helpPattern.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	// 尝试提取"计算"后面的内容
	calcPattern := regexp.MustCompile(`计算\s*(.+)`)
	matches = calcPattern.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractPriority 提取优先级
func (np *NLPProcessor) extractPriority(entities []Entity) int {
	for _, e := range entities {
		if e.Type == EntityPriority {
			switch e.Value {
			case "高优先", "紧急", "priority-high":
				return 10
			case "低优先", "priority-low":
				return 3
			default:
				return 5
			}
		}
	}
	return 0
}

// extractTaskID 提取任务ID
func (np *NLPProcessor) extractTaskID(entities []Entity) string {
	for _, e := range entities {
		if e.Type == EntityTaskID {
			return e.Value
		}
	}
	return ""
}

// entityTypeToName 实体类型转名称
func entityTypeToName(t EntityType) string {
	names := map[EntityType]string{
		EntityTaskID:    "任务ID",
		EntityAgentID:   "AgentID",
		EntitySkillName: "技能名称",
		EntityParameter: "参数",
		EntityTime:      "时间",
		EntityNumber:    "数字",
		EntityURL:       "URL",
		EntityFilePath:  "文件路径",
		EntityCommand:   "命令",
		EntityStatus:    "状态",
		EntityPriority:  "优先级",
	}
	return names[t]
}

// filterEmpty 过滤空字符串
func filterEmpty(tokens []string) []string {
	result := make([]string, 0)
	for _, t := range tokens {
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

// GetIntent 获取意图定义
func (np *NLPProcessor) GetIntent(id string) (*NLPIntent, error) {
	np.mu.RLock()
	intent, ok := np.intents[id]
	np.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("意图不存在: %s", id)
	}
	return intent, nil
}

// ListIntents 列出所有意图
func (np *NLPProcessor) ListIntents() []*NLPIntent {
	np.mu.RLock()
	list := make([]*NLPIntent, 0, len(np.intents))
	for _, intent := range np.intents {
		list = append(list, intent)
	}
	np.mu.RUnlock()
	return list
}

// AddIntent 添加自定义意图
func (np *NLPProcessor) AddIntent(intent *NLPIntent) error {
	if intent.ID == "" {
		intent.ID = fmt.Sprintf("intent-%d", time.Now().UnixNano())
	}

	np.mu.Lock()
	np.intents[intent.ID] = intent
	np.mu.Unlock()

	return nil
}

// GetTaskTemplate 获取任务模板
func (np *NLPProcessor) GetTaskTemplate(name string) (*TaskSpec, error) {
	np.mu.RLock()
	template, ok := np.taskTemplates[name]
	np.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("模板不存在: %s", name)
	}
	return template, nil
}

// AddTaskTemplate 添加任务模板
func (np *NLPProcessor) AddTaskTemplate(name string, template *TaskSpec) error {
	np.mu.Lock()
	np.taskTemplates[name] = template
	np.mu.Unlock()
	return nil
}

// Translate 翻译（简化实现）
func (np *NLPProcessor) Translate(ctx context.Context, text string, from Language, to Language) (string, error) {
	// 简化的翻译映射
	translations := map[string]map[string]string{
		"zh-en": map[string]string{
			"创建任务": "create task",
			"查询状态": "query status",
			"取消任务": "cancel task",
			"帮助":    "help",
			"计算":    "calculate",
			"分析":    "analyze",
		},
		"en-zh": map[string]string{
			"create task": "创建任务",
			"query status": "查询状态",
			"cancel task": "取消任务",
			"help":        "帮助",
			"calculate":   "计算",
			"analyze":     "分析",
		},
	}

	key := fmt.Sprintf("%s-%s", from, to)
	if transMap, ok := translations[key]; ok {
		for src, dst := range transMap {
			if strings.Contains(text, src) {
				return strings.ReplaceAll(text, src, dst), nil
			}
		}
	}

	return text, nil
}

// ExportConfig 导出配置
func (np *NLPProcessor) ExportConfig() ([]byte, error) {
	np.mu.RLock()
	data := map[string]interface{}{
		"config":      np.config,
		"intents":     np.intents,
		"templates":   np.taskTemplates,
	}
	np.mu.RUnlock()

	return json.MarshalIndent(data, "", "  ")
}

// GetStatistics 获取统计信息
func (np *NLPProcessor) GetStatistics() map[string]interface{} {
	np.mu.RLock()
	stats := map[string]interface{}{
		"intent_count":    len(np.intents),
		"template_count":  len(np.taskTemplates),
		"entity_types":    len(np.entityPatterns),
		"skill_count":     len(np.skillKeywords),
		"languages":       []Language{LangChinese, LangEnglish, LangJapanese, LangKorean, LangGerman, LangFrench},
	}
	np.mu.RUnlock()
	return stats
}
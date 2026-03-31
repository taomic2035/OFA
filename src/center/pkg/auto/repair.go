// Package auto - 自动修复模块
package auto

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// FaultType 故障类型
type FaultType string

const (
	FaultAgentOffline     FaultType = "agent_offline"     // Agent离线
	FaultTaskTimeout      FaultType = "task_timeout"      // 任务超时
	FaultTaskFailed       FaultType = "task_failed"       // 任务失败
	FaultResourceExhausted FaultType = "resource_exhausted" // 资源耗尽
	FaultNetworkError     FaultType = "network_error"     // 网络错误
	FaultMemoryLeak       FaultType = "memory_leak"       // 内存泄漏
	FaultHighCPU          FaultType = "high_cpu"          // CPU过高
	FaultDiskFull         FaultType = "disk_full"         // 磁盘满
	FaultServiceCrash     FaultType = "service_crash"     // 服务崩溃
	FaultConfigError      FaultType = "config_error"      // 配置错误
)

// FaultLevel 故障级别
type FaultLevel string

const (
	LevelCritical FaultLevel = "critical" // 严重
	LevelHigh     FaultLevel = "high"     // 高
	LevelMedium   FaultLevel = "medium"   // 中
	LevelLow      FaultLevel = "low"      // 低
)

// RepairStatus 修复状态
type RepairStatus string

const (
	RepairPending    RepairStatus = "pending"    // 待修复
	RepairInProgress RepairStatus = "in_progress" // 修复中
	RepairSuccess    RepairStatus = "success"    // 成功
	RepairFailed     RepairStatus = "failed"     // 失败
	RepairSkipped    RepairStatus = "skipped"    // 跳过
	RepairManual     RepairStatus = "manual"     // 需人工干预
)

// Fault 故障定义
type Fault struct {
	ID          string            `json:"id"`
	Type        FaultType         `json:"type"`
	Level       FaultLevel        `json:"level"`
	Source      string            `json:"source"`      // 故障源 (AgentID/TaskID)
	Description string            `json:"description"`
	Timestamp   time.Time         `json:"timestamp"`
	Metrics     map[string]float64 `json:"metrics"`    // 相关指标
	Context     map[string]interface{} `json:"context"` // 上下文信息
	AffectedResources []string    `json:"affected_resources"` // 受影响资源
	Count       int               `json:"count"`       // 发生次数
}

// RepairAction 修复动作
type RepairAction struct {
	ID          string            `json:"id"`
	FaultID     string            `json:"fault_id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // restart, reload, scale, migrate, cleanup
	Description string            `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration     `json:"timeout"`
	RetryCount  int               `json:"retry_count"`
	MaxRetries  int               `json:"max_retries"`
	RequireConfirmation bool      `json:"require_confirmation"`
	Dependencies []string         `json:"dependencies"` // 依赖的其他动作
}

// RepairRecord 修复记录
type RepairRecord struct {
	ID           string         `json:"id"`
	FaultID      string         `json:"fault_id"`
	Action       *RepairAction  `json:"action"`
	Status       RepairStatus   `json:"status"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Duration     time.Duration  `json:"duration"`
	Result       string         `json:"result"`
	Error        string         `json:"error,omitempty"`
	RollbackAction string       `json:"rollback_action,omitempty"`
	Metrics      map[string]float64 `json:"metrics"` // 修复前后指标
	Operator     string         `json:"operator"` // auto / manual
}

// RepairRule 修复规则
type RepairRule struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	FaultType       FaultType    `json:"fault_type"`
	FaultLevel      FaultLevel   `json:"fault_level"`
	Conditions      []Condition  `json:"conditions"`
	Actions         []RepairAction `json:"actions"`
	Enabled         bool         `json:"enabled"`
	Priority        int          `json:"priority"`
	Cooldown        time.Duration `json:"cooldown"`    // 冷却时间
	MaxAttempts     int          `json:"max_attempts"` // 最大尝试次数
	NotifyOnRepair  bool         `json:"notify_on_repair"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// Condition 触发条件
type Condition struct {
	Type     string      `json:"type"`     // metric, count, duration
	Operator string      `json:"operator"` // >, <, ==, >=, <=
	Value    interface{} `json:"value"`
	Duration time.Duration `json:"duration,omitempty"`
}

// AutoRepairConfig 自动修复配置
type AutoRepairConfig struct {
	Enabled           bool          `json:"enabled"`
	DetectionInterval time.Duration `json:"detection_interval"`
	RepairTimeout     time.Duration `json:"repair_timeout"`
	MaxConcurrent     int           `json:"max_concurrent"`    // 最大并发修复数
	HistoryRetention  time.Duration `json:"history_retention"` // 历史保留时间
	AutoConfirm       bool          `json:"auto_confirm"`      // 自动确认高风险操作
	NotifyChannel     string        `json:"notify_channel"`
}

// AutoRepairManager 自动修复管理器
type AutoRepairManager struct {
	config       AutoRepairConfig
	rules        map[string]*RepairRule
	faults       map[string]*Fault
	records      []*RepairRecord
	activeRepairs map[string]bool
	cooldowns    map[string]time.Time
	notifyFunc   func(level FaultLevel, message string)
	mu           sync.RWMutex
}

// NewAutoRepairManager 创建自动修复管理器
func NewAutoRepairManager(config AutoRepairConfig) *AutoRepairManager {
	return &AutoRepairManager{
		config:       config,
		rules:        make(map[string]*RepairRule),
		faults:       make(map[string]*Fault),
		records:      make([]*RepairRecord, 0),
		activeRepairs: make(map[string]bool),
		cooldowns:    make(map[string]time.Time),
	}
}

// Initialize 初始化
func (am *AutoRepairManager) Initialize() error {
	am.loadDefaultRules()
	return nil
}

// loadDefaultRules 加载默认修复规则
func (am *AutoRepairManager) loadDefaultRules() {
	rules := []RepairRule{
		{
			ID:         "rule-agent-offline",
			Name:       "Agent离线自动重连",
			FaultType:  FaultAgentOffline,
			FaultLevel: LevelHigh,
			Conditions: []Condition{
				{Type: "duration", Operator: ">", Value: 30 * time.Second},
			},
			Actions: []RepairAction{
				{
					ID:          "action-reconnect",
					Name:        "重新连接Agent",
					Type:        "reconnect",
					Description: "尝试重新建立与Agent的连接",
					Timeout:     30 * time.Second,
					MaxRetries:  3,
				},
			},
			Enabled:     true,
			Priority:    10,
			Cooldown:    5 * time.Minute,
			MaxAttempts: 3,
		},
		{
			ID:         "rule-task-timeout",
			Name:       "任务超时自动重试",
			FaultType:  FaultTaskTimeout,
			FaultLevel: LevelMedium,
			Conditions: []Condition{
				{Type: "count", Operator: "<", Value: 3},
			},
			Actions: []RepairAction{
				{
					ID:          "action-retry",
					Name:        "重新执行任务",
					Type:        "retry",
					Description: "在其他Agent上重新执行任务",
					Timeout:     60 * time.Second,
					MaxRetries:  2,
				},
			},
			Enabled:     true,
			Priority:   8,
			Cooldown:   2 * time.Minute,
			MaxAttempts: 2,
		},
		{
			ID:         "rule-high-cpu",
			Name:       "高CPU自动处理",
			FaultType:  FaultHighCPU,
			FaultLevel: LevelMedium,
			Conditions: []Condition{
				{Type: "metric", Operator: ">", Value: 90.0, Duration: 2 * time.Minute},
			},
			Actions: []RepairAction{
				{
					ID:          "action-scale",
					Name:        "扩容实例",
					Type:        "scale",
					Description: "增加实例数量分散负载",
					Parameters:  map[string]interface{}{"scale_factor": 1.5},
					Timeout:     5 * time.Minute,
					RequireConfirmation: true,
				},
				{
					ID:          "action-migrate",
					Name:        "迁移任务",
					Type:        "migrate",
					Description: "将部分任务迁移到其他实例",
					Timeout:     2 * time.Minute,
				},
			},
			Enabled:     true,
			Priority:   7,
			Cooldown:   10 * time.Minute,
			MaxAttempts: 1,
		},
		{
			ID:         "rule-memory-leak",
			Name:       "内存泄漏自动处理",
			FaultType:  FaultMemoryLeak,
			FaultLevel: LevelHigh,
			Conditions: []Condition{
				{Type: "metric", Operator: ">", Value: 85.0, Duration: 5 * time.Minute},
			},
			Actions: []RepairAction{
				{
					ID:          "action-restart",
					Name:        "重启服务",
					Type:        "restart",
					Description: "重启存在内存泄漏的服务",
					Timeout:     1 * time.Minute,
					MaxRetries:  1,
					RollbackAction: "rollback-restart",
				},
			},
			Enabled:     true,
			Priority:   9,
			Cooldown:   15 * time.Minute,
			MaxAttempts: 1,
		},
		{
			ID:         "rule-disk-full",
			Name:       "磁盘满自动清理",
			FaultType:  FaultDiskFull,
			FaultLevel: LevelCritical,
			Conditions: []Condition{
				{Type: "metric", Operator: ">", Value: 90.0},
			},
			Actions: []RepairAction{
				{
					ID:          "action-cleanup",
					Name:        "清理临时文件",
					Type:        "cleanup",
					Description: "清理临时文件和日志",
					Timeout:     5 * time.Minute,
				},
				{
					ID:          "action-alert",
					Name:        "发送告警",
					Type:        "alert",
					Description: "通知管理员进行人工处理",
					RequireConfirmation: false,
				},
			},
			Enabled:     true,
			Priority:   10,
			Cooldown:   30 * time.Minute,
			MaxAttempts: 2,
		},
		{
			ID:         "rule-service-crash",
			Name:       "服务崩溃自动恢复",
			FaultType:  FaultServiceCrash,
			FaultLevel: LevelCritical,
			Conditions: []Condition{},
			Actions: []RepairAction{
				{
					ID:          "action-restart",
					Name:        "重启服务",
					Type:        "restart",
					Description: "自动重启崩溃的服务",
					Timeout:     2 * time.Minute,
					MaxRetries:  3,
				},
			},
			Enabled:     true,
			Priority:   10,
			Cooldown:   5 * time.Minute,
			MaxAttempts: 3,
		},
	}

	am.mu.Lock()
	for _, rule := range rules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		am.rules[rule.ID] = &rule
	}
	am.mu.Unlock()
}

// DetectFault 检测故障
func (am *AutoRepairManager) DetectFault(ctx context.Context, faultType FaultType, source string, metrics map[string]float64, context map[string]interface{}) (*Fault, error) {
	// 确定故障级别
	level := am.determineFaultLevel(faultType, metrics)

	// 检查是否已存在相同故障
	faultKey := fmt.Sprintf("%s-%s", faultType, source)
	am.mu.RLock()
	existingFault, exists := am.faults[faultKey]
	am.mu.RUnlock()

	if exists {
		// 更新现有故障
		am.mu.Lock()
		existingFault.Count++
		existingFault.Timestamp = time.Now()
		for k, v := range metrics {
			existingFault.Metrics[k] = v
		}
		am.mu.Unlock()
		return existingFault, nil
	}

	// 创建新故障
	fault := &Fault{
		ID:          fmt.Sprintf("fault-%d", time.Now().UnixNano()),
		Type:        faultType,
		Level:       level,
		Source:      source,
		Description: am.generateFaultDescription(faultType, source, metrics),
		Timestamp:   time.Now(),
		Metrics:     metrics,
		Context:     context,
		Count:       1,
	}

	am.mu.Lock()
	am.faults[faultKey] = fault
	am.mu.Unlock()

	return fault, nil
}

// determineFaultLevel 确定故障级别
func (am *AutoRepairManager) determineFaultLevel(faultType FaultType, metrics map[string]float64) FaultLevel {
	switch faultType {
	case FaultServiceCrash, FaultDiskFull:
		return LevelCritical
	case FaultAgentOffline, FaultMemoryLeak:
		return LevelHigh
	case FaultTaskTimeout, FaultNetworkError:
		return LevelMedium
	default:
		// 根据指标判断
		if cpu, ok := metrics["cpu"]; ok && cpu > 95 {
			return LevelCritical
		}
		if mem, ok := metrics["memory"]; ok && mem > 90 {
			return LevelHigh
		}
		return LevelMedium
	}
}

// generateFaultDescription 生成故障描述
func (am *AutoRepairManager) generateFaultDescription(faultType FaultType, source string, metrics map[string]float64) string {
	descriptions := map[FaultType]string{
		FaultAgentOffline:     "Agent离线，无法通信",
		FaultTaskTimeout:      "任务执行超时",
		FaultTaskFailed:       "任务执行失败",
		FaultResourceExhausted: "资源耗尽",
		FaultNetworkError:     "网络连接错误",
		FaultMemoryLeak:       "检测到内存持续增长",
		FaultHighCPU:          "CPU使用率过高",
		FaultDiskFull:         "磁盘空间不足",
		FaultServiceCrash:     "服务异常崩溃",
		FaultConfigError:      "配置错误",
	}

	desc := descriptions[faultType]
	if desc == "" {
		desc = "未知故障类型"
	}

	// 添加指标信息
	if len(metrics) > 0 {
		desc += fmt.Sprintf(" [指标: ")
		for k, v := range metrics {
			desc += fmt.Sprintf("%s=%.1f ", k, v)
		}
		desc += "]"
	}

	return desc
}

// AutoRepair 自动修复
func (am *AutoRepairManager) AutoRepair(ctx context.Context, fault *Fault) (*RepairRecord, error) {
	if !am.config.Enabled {
		return nil, fmt.Errorf("自动修复未启用")
	}

	// 查找匹配的规则
	rule := am.findMatchingRule(fault)
	if rule == nil {
		return nil, fmt.Errorf("未找到匹配的修复规则")
	}

	// 检查冷却期
	faultKey := fmt.Sprintf("%s-%s", fault.Type, fault.Source)
	am.mu.RLock()
	cooldownTime, inCooldown := am.cooldowns[faultKey]
	am.mu.RUnlock()

	if inCooldown && time.Now().Before(cooldownTime) {
		return nil, fmt.Errorf("修复处于冷却期，剩余时间: %v", time.Until(cooldownTime))
	}

	// 检查最大并发数
	am.mu.Lock()
	if len(am.activeRepairs) >= am.config.MaxConcurrent {
		am.mu.Unlock()
		return nil, fmt.Errorf("已达到最大并发修复数")
	}

	// 开始修复
	record := &RepairRecord{
		ID:        fmt.Sprintf("repair-%d", time.Now().UnixNano()),
		FaultID:   fault.ID,
		Status:    RepairInProgress,
		StartTime: time.Now(),
		Operator:  "auto",
		Metrics:   make(map[string]float64),
	}

	am.activeRepairs[record.ID] = true
	am.mu.Unlock()

	// 执行修复动作
	success := false
	var lastError error

	for _, action := range rule.Actions {
		// 检查是否需要确认
		if action.RequireConfirmation && !am.config.AutoConfirm {
			record.Status = RepairManual
			record.Result = "需要人工确认"
			break
		}

		// 执行动作
		actionCopy := action
		actionCopy.RetryCount = 0

		for actionCopy.RetryCount < actionCopy.MaxRetries+1 {
			result, err := am.executeAction(ctx, &actionCopy, fault)

			if err == nil {
				record.Action = &actionCopy
				record.Result = result
				success = true
				break
			}

			lastError = err
			actionCopy.RetryCount++

			if actionCopy.RetryCount < actionCopy.MaxRetries+1 {
				time.Sleep(time.Second * time.Duration(actionCopy.RetryCount))
			}
		}

		if !success {
			record.Status = RepairFailed
			record.Error = lastError.Error()
			break
		}
	}

	if success {
		record.Status = RepairSuccess
		// 设置冷却期
		am.mu.Lock()
		am.cooldowns[faultKey] = time.Now().Add(rule.Cooldown)
		am.mu.Unlock()
	}

	record.EndTime = time.Now()
	record.Duration = record.EndTime.Sub(record.StartTime)

	// 保存记录
	am.mu.Lock()
	delete(am.activeRepairs, record.ID)
	am.records = append(am.records, record)
	am.mu.Unlock()

	// 发送通知
	if rule.NotifyOnRepair && am.notifyFunc != nil {
		message := fmt.Sprintf("故障 %s 已%s修复", fault.Type, map[RepairStatus]string{
			RepairSuccess: "成功",
			RepairFailed:  "失败",
			RepairManual:  "等待人工",
		}[record.Status])
		am.notifyFunc(fault.Level, message)
	}

	return record, nil
}

// findMatchingRule 查找匹配的修复规则
func (am *AutoRepairManager) findMatchingRule(fault *Fault) *RepairRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var bestMatch *RepairRule
	bestPriority := -1

	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}
		if rule.FaultType != fault.Type {
			continue
		}
		if rule.FaultLevel != fault.Level && rule.FaultLevel != "" {
			continue
		}

		// 检查条件
		conditionsMet := true
		for _, cond := range rule.Conditions {
			if !am.checkCondition(cond, fault) {
				conditionsMet = false
				break
			}
		}

		if conditionsMet && rule.Priority > bestPriority {
			bestMatch = rule
			bestPriority = rule.Priority
		}
	}

	return bestMatch
}

// checkCondition 检查条件
func (am *AutoRepairManager) checkCondition(cond Condition, fault *Fault) bool {
	switch cond.Type {
	case "metric":
		if val, ok := fault.Metrics[cond.Type]; ok {
			switch cond.Operator {
			case ">":
				return val > cond.Value.(float64)
			case "<":
				return val < cond.Value.(float64)
			case ">=":
				return val >= cond.Value.(float64)
			case "<=":
				return val <= cond.Value.(float64)
			}
		}
	case "count":
		switch cond.Operator {
		case "<":
			return fault.Count < cond.Value.(int)
		case ">":
			return fault.Count > cond.Value.(int)
		case "==":
			return fault.Count == cond.Value.(int)
		}
	case "duration":
		elapsed := time.Since(fault.Timestamp)
		switch cond.Operator {
		case ">":
			return elapsed > cond.Value.(time.Duration)
		case "<":
			return elapsed < cond.Value.(time.Duration)
		}
	}
	return false
}

// executeAction 执行修复动作
func (am *AutoRepairManager) executeAction(ctx context.Context, action *RepairAction, fault *Fault) (string, error) {
	// 模拟执行修复动作
	// 实际实现中会调用具体的服务接口

	switch action.Type {
	case "restart":
		return "服务已重启", nil
	case "reconnect":
		return "已重新建立连接", nil
	case "retry":
		return "任务已在其他Agent上重新执行", nil
	case "scale":
		factor := action.Parameters["scale_factor"]
		return fmt.Sprintf("已扩容 %.1f 倍", factor), nil
	case "migrate":
		return "任务已迁移", nil
	case "cleanup":
		return "已清理临时文件", nil
	case "alert":
		return "已发送告警通知", nil
	default:
		return "", fmt.Errorf("未知动作类型: %s", action.Type)
	}
}

// ManualRepair 手动修复
func (am *AutoRepairManager) ManualRepair(ctx context.Context, faultID string, actionType string, params map[string]interface{}) (*RepairRecord, error) {
	am.mu.RLock()
	var fault *Fault
	for _, f := range am.faults {
		if f.ID == faultID {
			fault = f
			break
		}
	}
	am.mu.RUnlock()

	if fault == nil {
		return nil, fmt.Errorf("故障不存在: %s", faultID)
	}

	record := &RepairRecord{
		ID:        fmt.Sprintf("repair-%d", time.Now().UnixNano()),
		FaultID:   fault.ID,
		Status:    RepairInProgress,
		StartTime: time.Now(),
		Operator:  "manual",
		Metrics:   make(map[string]float64),
	}

	action := &RepairAction{
		ID:         fmt.Sprintf("manual-%s", actionType),
		Type:       actionType,
		Parameters: params,
		Timeout:    am.config.RepairTimeout,
	}

	result, err := am.executeAction(ctx, action, fault)

	record.EndTime = time.Now()
	record.Duration = record.EndTime.Sub(record.StartTime)
	record.Action = action

	if err != nil {
		record.Status = RepairFailed
		record.Error = err.Error()
	} else {
		record.Status = RepairSuccess
		record.Result = result
	}

	am.mu.Lock()
	am.records = append(am.records, record)
	am.mu.Unlock()

	return record, nil
}

// AddRule 添加修复规则
func (am *AutoRepairManager) AddRule(rule *RepairRule) error {
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	am.mu.Lock()
	am.rules[rule.ID] = rule
	am.mu.Unlock()

	return nil
}

// GetFaults 获取故障列表
func (am *AutoRepairManager) GetFaults(level FaultLevel) []*Fault {
	am.mu.RLock()
	list := make([]*Fault, 0)
	for _, f := range am.faults {
		if level == "" || f.Level == level {
			list = append(list, f)
		}
	}
	am.mu.RUnlock()
	return list
}

// GetRecords 获取修复记录
func (am *AutoRepairManager) GetRecords(limit int) []*RepairRecord {
	am.mu.RLock()
	size := len(am.records)
	if limit > size || limit <= 0 {
		limit = size
	}
	records := am.records[size-limit:]
	am.mu.RUnlock()
	return records
}

// ClearFault 清除故障
func (am *AutoRepairManager) ClearFault(faultType FaultType, source string) error {
	faultKey := fmt.Sprintf("%s-%s", faultType, source)

	am.mu.Lock()
	delete(am.faults, faultKey)
	am.mu.Unlock()

	return nil
}

// SetNotifyFunc 设置通知函数
func (am *AutoRepairManager) SetNotifyFunc(fn func(level FaultLevel, message string)) {
	am.mu.Lock()
	am.notifyFunc = fn
	am.mu.Unlock()
}

// GetStatistics 获取统计信息
func (am *AutoRepairManager) GetStatistics() map[string]interface{} {
	am.mu.RLock()

	totalFaults := len(am.faults)
	totalRecords := len(am.records)
	successCount := 0
	failedCount := 0

	for _, r := range am.records {
		if r.Status == RepairSuccess {
			successCount++
		} else if r.Status == RepairFailed {
			failedCount++
		}
	}

	faultByType := make(map[FaultType]int)
	for _, f := range am.faults {
		faultByType[f.Type]++
	}

	am.mu.RUnlock()

	return map[string]interface{}{
		"total_faults":     totalFaults,
		"total_repairs":    totalRecords,
		"success_count":    successCount,
		"failed_count":     failedCount,
		"active_repairs":   len(am.activeRepairs),
		"rules_count":      len(am.rules),
		"fault_by_type":    faultByType,
		"success_rate":     float64(successCount) / float64(totalRecords+1) * 100,
	}
}

// ExportRules 导出规则
func (am *AutoRepairManager) ExportRules() ([]byte, error) {
	am.mu.RLock()
	data, err := json.MarshalIndent(am.rules, "", "  ")
	am.mu.RUnlock()
	return data, err
}

// ImportRules 导入规则
func (am *AutoRepairManager) ImportRules(data []byte) error {
	rules := make(map[string]*RepairRule)
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	am.mu.Lock()
	for id, rule := range rules {
		rule.UpdatedAt = time.Now()
		am.rules[id] = rule
	}
	am.mu.Unlock()

	return nil
}
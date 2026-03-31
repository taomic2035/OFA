// Package auto - 智能巡检模块
package auto

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CheckType 检查类型
type CheckType string

const (
	CheckHealth      CheckType = "health"       // 健康检查
	CheckPerformance CheckType = "performance"  // 性能检查
	CheckSecurity    CheckType = "security"     // 安全检查
	CheckResource    CheckType = "resource"     // 资源检查
	CheckNetwork     CheckType = "network"      // 网络检查
	CheckStorage     CheckType = "storage"      // 存储检查
	CheckConfig      CheckType = "config"       // 配置检查
	CheckDependency  CheckType = "dependency"   // 依赖检查
)

// CheckStatus 检查状态
type CheckStatus string

const (
	StatusPass      CheckStatus = "pass"      // 通过
	StatusWarn      CheckStatus = "warn"      // 警告
	StatusFail      CheckStatus = "fail"      // 失败
	StatusSkipped   CheckStatus = "skipped"   // 跳过
	StatusUnknown   CheckStatus = "unknown"   // 未知
)

// CheckResult 检查结果
type CheckResult struct {
	ID          string            `json:"id"`
	CheckID     string            `json:"check_id"`
	CheckType   CheckType         `json:"check_type"`
	Name        string            `json:"name"`
	Status      CheckStatus       `json:"status"`
	Score       float64           `json:"score"`       // 0-100
	Message     string            `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Metrics     map[string]float64 `json:"metrics"`
	Suggestions []string          `json:"suggestions"`
	Timestamp   time.Time         `json:"timestamp"`
	Duration    time.Duration     `json:"duration"`
}

// PatrolCheck 巡检项定义
type PatrolCheck struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         CheckType     `json:"type"`
	Description  string        `json:"description"`
	Interval     time.Duration `json:"interval"`     // 检查间隔
	Timeout      time.Duration `json:"timeout"`      // 超时时间
	Enabled      bool          `json:"enabled"`
	Priority     int           `json:"priority"`     // 优先级
	Thresholds   Thresholds    `json:"thresholds"`   // 阈值配置
	DependsOn    []string      `json:"depends_on"`   // 依赖的检查项
	Actions      []CheckAction `json:"actions"`      // 失败时的动作
	LastRun      time.Time     `json:"last_run"`
	NextRun      time.Time     `json:"next_run"`
}

// Thresholds 阈值配置
type Thresholds struct {
	Warn  ThresholdConfig `json:"warn"`
	Fail  ThresholdConfig `json:"fail"`
}

// ThresholdConfig 阈值配置项
type ThresholdConfig struct {
	CPU         float64 `json:"cpu"`          // CPU %
	Memory      float64 `json:"memory"`       // Memory %
	Disk        float64 `json:"disk"`         // Disk %
	Latency     float64 `json:"latency"`      // Latency ms
	ErrorRate   float64 `json:"error_rate"`   // Error rate %
	Connections int     `json:"connections"`  // Connection count
	ResponseTime float64 `json:"response_time"` // Response time ms
}

// CheckAction 检查动作
type CheckAction struct {
	Type        string                 `json:"type"`        // alert, repair, restart, scale
	Condition   string                 `json:"condition"`   // 执行条件
	Parameters  map[string]interface{} `json:"parameters"`
}

// PatrolReport 巡检报告
type PatrolReport struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Duration     time.Duration  `json:"duration"`
	TotalChecks  int            `json:"total_checks"`
	PassCount    int            `json:"pass_count"`
	WarnCount    int            `json:"warn_count"`
	FailCount    int            `json:"fail_count"`
	SkipCount    int            `json:"skip_count"`
	OverallScore float64        `json:"overall_score"` // 0-100
	HealthLevel  string         `json:"health_level"`  // excellent/good/fair/poor/critical
	Results      []*CheckResult `json:"results"`
	Summary      string         `json:"summary"`
	GeneratedAt  time.Time      `json:"generated_at"`
}

// AnomalyEvent 异常事件
type AnomalyEvent struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`        // spike, drop, trend, outlier
	Source      string            `json:"source"`      // 来源
	Metric      string            `json:"metric"`      // 指标名
	Value       float64           `json:"value"`       // 当前值
	Baseline    float64           `json:"baseline"`    // 基线值
	Deviation   float64           `json:"deviation"`   // 偏差度
	Severity    string            `json:"severity"`    // low/medium/high/critical
	Description string            `json:"description"`
	Timestamp   time.Time         `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	ID           string   `json:"id"`
	Category     string   `json:"category"`     // performance, cost, security, reliability
	Priority     string   `json:"priority"`     // high/medium/low
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Impact       string   `json:"impact"`       // 预期影响
	Effort       string   `json:"effort"`       // 实施难度
	ActionItems  []string `json:"action_items"` // 具体步骤
	RelatedCheck string   `json:"related_check"`
	CreatedAt    time.Time `json:"created_at"`
}

// PatrolConfig 巡检配置
type PatrolConfig struct {
	DefaultInterval   time.Duration `json:"default_interval"`
	DefaultTimeout    time.Duration `json:"default_timeout"`
	EnableAutoRepair  bool          `json:"enable_auto_repair"`
	ReportRetention   time.Duration `json:"report_retention"`
	AnomalyDetection  bool          `json:"anomaly_detection"`
	BaselineWindow    time.Duration `json:"baseline_window"`    // 基线计算窗口
	MaxConcurrent     int           `json:"max_concurrent"`     // 最大并发检查数
	NotifyOnFail      bool          `json:"notify_on_fail"`
	NotifyOnWarn      bool          `json:"notify_on_warn"`
}

// PatrolManager 巡检管理器
type PatrolManager struct {
	config       PatrolConfig
	checks       map[string]*PatrolCheck
	results      []*CheckResult
	reports      []*PatrolReport
	anomalies    []*AnomalyEvent
	suggestions  []*OptimizationSuggestion
	baselines    map[string]*MetricBaseline
	scheduler    *PatrolScheduler
	notifyFunc   func(level string, message string)
	mu           sync.RWMutex
}

// MetricBaseline 指标基线
type MetricBaseline struct {
	Metric       string    `json:"metric"`
	Mean         float64   `json:"mean"`
	StdDev       float64   `json:"std_dev"`
	Min          float64   `json:"min"`
	Max          float64   `json:"max"`
	P95          float64   `json:"p95"`
	P99          float64   `json:"p99"`
	SampleCount  int       `json:"sample_count"`
	LastUpdated  time.Time `json:"last_updated"`
}

// PatrolScheduler 巡检调度器
type PatrolScheduler struct {
	running bool
	stopCh  chan struct{}
}

// NewPatrolManager 创建巡检管理器
func NewPatrolManager(config PatrolConfig) *PatrolManager {
	return &PatrolManager{
		config:      config,
		checks:      make(map[string]*PatrolCheck),
		results:     make([]*CheckResult, 0),
		reports:     make([]*PatrolReport, 0),
		anomalies:   make([]*AnomalyEvent, 0),
		suggestions: make([]*OptimizationSuggestion, 0),
		baselines:   make(map[string]*MetricBaseline),
		scheduler:   &PatrolScheduler{stopCh: make(chan struct{})},
	}
}

// Initialize 初始化
func (pm *PatrolManager) Initialize() error {
	pm.loadDefaultChecks()
	pm.loadBaselines()
	return nil
}

// loadDefaultChecks 加载默认检查项
func (pm *PatrolManager) loadDefaultChecks() {
	checks := []PatrolCheck{
		{
			ID:          "check-agent-health",
			Name:        "Agent健康检查",
			Type:        CheckHealth,
			Description: "检查所有Agent的运行状态和心跳",
			Interval:    30 * time.Second,
			Timeout:     10 * time.Second,
			Enabled:     true,
			Priority:    10,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{Latency: 100},
				Fail: ThresholdConfig{Latency: 500},
			},
		},
		{
			ID:          "check-cpu-usage",
			Name:        "CPU使用率检查",
			Type:        CheckResource,
			Description: "检查系统CPU使用率",
			Interval:    1 * time.Minute,
			Timeout:     5 * time.Second,
			Enabled:     true,
			Priority:    9,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{CPU: 70},
				Fail: ThresholdConfig{CPU: 90},
			},
		},
		{
			ID:          "check-memory-usage",
			Name:        "内存使用率检查",
			Type:        CheckResource,
			Description: "检查系统内存使用率",
			Interval:    1 * time.Minute,
			Timeout:     5 * time.Second,
			Enabled:     true,
			Priority:    9,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{Memory: 75},
				Fail: ThresholdConfig{Memory: 90},
			},
		},
		{
			ID:          "check-disk-usage",
			Name:        "磁盘使用率检查",
			Type:        CheckStorage,
			Description: "检查磁盘空间使用情况",
			Interval:    5 * time.Minute,
			Timeout:     10 * time.Second,
			Enabled:     true,
			Priority:    8,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{Disk: 80},
				Fail: ThresholdConfig{Disk: 95},
			},
		},
		{
			ID:          "check-network-latency",
			Name:        "网络延迟检查",
			Type:        CheckNetwork,
			Description: "检查网络连接延迟",
			Interval:    30 * time.Second,
			Timeout:     5 * time.Second,
			Enabled:     true,
			Priority:    8,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{Latency: 50},
				Fail: ThresholdConfig{Latency: 200},
			},
		},
		{
			ID:          "check-error-rate",
			Name:        "错误率检查",
			Type:        CheckPerformance,
			Description: "检查任务执行错误率",
			Interval:    1 * time.Minute,
			Timeout:     10 * time.Second,
			Enabled:     true,
			Priority:    9,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{ErrorRate: 5},
				Fail: ThresholdConfig{ErrorRate: 15},
			},
		},
		{
			ID:          "check-response-time",
			Name:        "响应时间检查",
			Type:        CheckPerformance,
			Description: "检查API响应时间",
			Interval:    30 * time.Second,
			Timeout:     5 * time.Second,
			Enabled:     true,
			Priority:    8,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{ResponseTime: 200},
				Fail: ThresholdConfig{ResponseTime: 500},
			},
		},
		{
			ID:          "check-connection-pool",
			Name:        "连接池状态检查",
			Type:        CheckResource,
			Description: "检查连接池使用情况",
			Interval:    1 * time.Minute,
			Timeout:     5 * time.Second,
			Enabled:     true,
			Priority:    7,
			Thresholds: Thresholds{
				Warn: ThresholdConfig{Connections: 80},
				Fail: ThresholdConfig{Connections: 95},
			},
		},
		{
			ID:          "check-config-validity",
			Name:        "配置有效性检查",
			Type:        CheckConfig,
			Description: "检查系统配置是否有效",
			Interval:    10 * time.Minute,
			Timeout:     30 * time.Second,
			Enabled:     true,
			Priority:    6,
		},
		{
			ID:          "check-dependency-health",
			Name:        "依赖服务健康检查",
			Type:        CheckDependency,
			Description: "检查外部依赖服务状态",
			Interval:    1 * time.Minute,
			Timeout:     10 * time.Second,
			Enabled:     true,
			Priority:    8,
		},
	}

	pm.mu.Lock()
	for _, check := range checks {
		check.NextRun = time.Now().Add(check.Interval)
		pm.checks[check.ID] = &check
	}
	pm.mu.Unlock()
}

// loadBaselines 加载基线数据
func (pm *PatrolManager) loadBaselines() {
	// 初始化默认基线
	baselines := map[string]*MetricBaseline{
		"cpu_usage":     {Metric: "cpu_usage", Mean: 30, StdDev: 15, Min: 5, Max: 80, P95: 60, P99: 70},
		"memory_usage":  {Metric: "memory_usage", Mean: 50, StdDev: 20, Min: 20, Max: 85, P95: 75, P99: 82},
		"disk_usage":    {Metric: "disk_usage", Mean: 60, StdDev: 15, Min: 30, Max: 90, P95: 85, P99: 88},
		"latency":       {Metric: "latency", Mean: 20, StdDev: 10, Min: 5, Max: 100, P95: 50, P99: 80},
		"error_rate":    {Metric: "error_rate", Mean: 1, StdDev: 2, Min: 0, Max: 10, P95: 5, P99: 8},
		"response_time": {Metric: "response_time", Mean: 50, StdDev: 30, Min: 10, Max: 300, P95: 150, P99: 250},
	}

	pm.mu.Lock()
	for k, v := range baselines {
		v.LastUpdated = time.Now()
		pm.baselines[k] = v
	}
	pm.mu.Unlock()
}

// RunCheck 执行单个检查
func (pm *PatrolManager) RunCheck(ctx context.Context, checkID string) (*CheckResult, error) {
	pm.mu.RLock()
	check, ok := pm.checks[checkID]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("检查项不存在: %s", checkID)
	}

	startTime := time.Now()
	result := &CheckResult{
		ID:        fmt.Sprintf("result-%d", time.Now().UnixNano()),
		CheckID:   check.ID,
		CheckType: check.Type,
		Name:      check.Name,
		Details:   make(map[string]interface{}),
		Metrics:   make(map[string]float64),
		Timestamp: startTime,
	}

	// 执行检查
	status, metrics, message := pm.executeCheck(ctx, check)

	result.Status = status
	result.Message = message
	result.Metrics = metrics
	result.Duration = time.Since(startTime)

	// 计算分数
	result.Score = pm.calculateScore(check, metrics)

	// 生成建议
	result.Suggestions = pm.generateSuggestions(check, metrics, status)

	// 检测异常
	if pm.config.AnomalyDetection {
		pm.detectAnomalies(check, metrics)
	}

	// 更新检查项
	pm.mu.Lock()
	check.LastRun = startTime
	check.NextRun = startTime.Add(check.Interval)
	pm.results = append(pm.results, result)
	pm.mu.Unlock()

	// 触发通知
	if pm.notifyFunc != nil {
		if status == StatusFail && pm.config.NotifyOnFail {
			pm.notifyFunc("critical", fmt.Sprintf("检查失败: %s - %s", check.Name, message))
		} else if status == StatusWarn && pm.config.NotifyOnWarn {
			pm.notifyFunc("warn", fmt.Sprintf("检查警告: %s - %s", check.Name, message))
		}
	}

	return result, nil
}

// executeCheck 执行检查逻辑
func (pm *PatrolManager) executeCheck(ctx context.Context, check *PatrolCheck) (CheckStatus, map[string]float64, string) {
	metrics := make(map[string]float64)

	switch check.ID {
	case "check-agent-health":
		return pm.checkAgentHealth(metrics)
	case "check-cpu-usage":
		return pm.checkCPUUsage(metrics)
	case "check-memory-usage":
		return pm.checkMemoryUsage(metrics)
	case "check-disk-usage":
		return pm.checkDiskUsage(metrics)
	case "check-network-latency":
		return pm.checkNetworkLatency(metrics)
	case "check-error-rate":
		return pm.checkErrorRate(metrics)
	case "check-response-time":
		return pm.checkResponseTime(metrics)
	case "check-connection-pool":
		return pm.checkConnectionPool(metrics)
	case "check-config-validity":
		return pm.checkConfigValidity(metrics)
	case "check-dependency-health":
		return pm.checkDependencyHealth(metrics)
	default:
		return StatusUnknown, metrics, "未知检查类型"
	}
}

// 具体检查实现
func (pm *PatrolManager) checkAgentHealth(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	// 模拟检查
	metrics["total_agents"] = 10
	metrics["online_agents"] = 9
	metrics["avg_latency"] = 25

	if metrics["online_agents"].(float64) < metrics["total_agents"].(float64)*0.8 {
		return StatusFail, metrics, "在线Agent数量过低"
	} else if metrics["avg_latency"].(float64) > 100 {
		return StatusWarn, metrics, "Agent平均延迟较高"
	}
	return StatusPass, metrics, "所有Agent运行正常"
}

func (pm *PatrolManager) checkCPUUsage(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["cpu_usage"] = 45.5
	metrics["cpu_cores"] = 8
	metrics["load_avg"] = 3.2

	cpu := metrics["cpu_usage"].(float64)
	if cpu > 90 {
		return StatusFail, metrics, fmt.Sprintf("CPU使用率过高: %.1f%%", cpu)
	} else if cpu > 70 {
		return StatusWarn, metrics, fmt.Sprintf("CPU使用率较高: %.1f%%", cpu)
	}
	return StatusPass, metrics, fmt.Sprintf("CPU使用率正常: %.1f%%", cpu)
}

func (pm *PatrolManager) checkMemoryUsage(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["memory_usage"] = 62.3
	metrics["memory_total_gb"] = 16
	metrics["memory_used_gb"] = 10

	mem := metrics["memory_usage"].(float64)
	if mem > 90 {
		return StatusFail, metrics, fmt.Sprintf("内存使用率过高: %.1f%%", mem)
	} else if mem > 75 {
		return StatusWarn, metrics, fmt.Sprintf("内存使用率较高: %.1f%%", mem)
	}
	return StatusPass, metrics, fmt.Sprintf("内存使用率正常: %.1f%%", mem)
}

func (pm *PatrolManager) checkDiskUsage(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["disk_usage"] = 55.0
	metrics["disk_total_gb"] = 500
	metrics["disk_free_gb"] = 225

	disk := metrics["disk_usage"].(float64)
	if disk > 95 {
		return StatusFail, metrics, fmt.Sprintf("磁盘空间不足: %.1f%%", disk)
	} else if disk > 80 {
		return StatusWarn, metrics, fmt.Sprintf("磁盘空间紧张: %.1f%%", disk)
	}
	return StatusPass, metrics, fmt.Sprintf("磁盘空间充足: %.1f%%", disk)
}

func (pm *PatrolManager) checkNetworkLatency(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["latency_ms"] = 18.5
	metrics["packet_loss"] = 0.1
	metrics["bandwidth_mbps"] = 1000

	latency := metrics["latency_ms"].(float64)
	if latency > 200 {
		return StatusFail, metrics, fmt.Sprintf("网络延迟过高: %.1fms", latency)
	} else if latency > 50 {
		return StatusWarn, metrics, fmt.Sprintf("网络延迟较高: %.1fms", latency)
	}
	return StatusPass, metrics, fmt.Sprintf("网络延迟正常: %.1fms", latency)
}

func (pm *PatrolManager) checkErrorRate(metrics map[string]float64) (CheckStatus, map[string]float64, message string) {
	metrics["error_rate"] = 2.5
	metrics["total_requests"] = 10000
	metrics["error_count"] = 250

	errRate := metrics["error_rate"].(float64)
	if errRate > 15 {
		return StatusFail, metrics, fmt.Sprintf("错误率过高: %.1f%%", errRate)
	} else if errRate > 5 {
		return StatusWarn, metrics, fmt.Sprintf("错误率偏高: %.1f%%", errRate)
	}
	return StatusPass, metrics, fmt.Sprintf("错误率正常: %.1f%%", errRate)
}

func (pm *PatrolManager) checkResponseTime(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["response_time_ms"] = 85.0
	metrics["p50"] = 50
	metrics["p95"] = 150
	metrics["p99"] = 280

	rt := metrics["response_time_ms"].(float64)
	if rt > 500 {
		return StatusFail, metrics, fmt.Sprintf("响应时间过长: %.1fms", rt)
	} else if rt > 200 {
		return StatusWarn, metrics, fmt.Sprintf("响应时间较长: %.1fms", rt)
	}
	return StatusPass, metrics, fmt.Sprintf("响应时间正常: %.1fms", rt)
}

func (pm *PatrolManager) checkConnectionPool(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["active_connections"] = 65
	metrics["max_connections"] = 100
	metrics["idle_connections"] = 35

	active := metrics["active_connections"].(float64)
	max := metrics["max_connections"].(float64)
	usage := active / max * 100

	if usage > 95 {
		return StatusFail, metrics, fmt.Sprintf("连接池接近满载: %.1f%%", usage)
	} else if usage > 80 {
		return StatusWarn, metrics, fmt.Sprintf("连接池使用较高: %.1f%%", usage)
	}
	return StatusPass, metrics, fmt.Sprintf("连接池使用正常: %.1f%%", usage)
}

func (pm *PatrolManager) checkConfigValidity(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["config_valid"] = 1
	metrics["config_items"] = 25
	metrics["warnings"] = 0

	if metrics["warnings"].(float64) > 0 {
		return StatusWarn, metrics, "配置存在警告项"
	}
	return StatusPass, metrics, "配置有效"
}

func (pm *PatrolManager) checkDependencyHealth(metrics map[string]float64) (CheckStatus, map[string]float64, string) {
	metrics["total_dependencies"] = 5
	metrics["healthy_dependencies"] = 5
	metrics["unhealthy_dependencies"] = 0

	unhealthy := metrics["unhealthy_dependencies"].(float64)
	if unhealthy > 2 {
		return StatusFail, metrics, fmt.Sprintf("多个依赖服务不健康: %d", int(unhealthy))
	} else if unhealthy > 0 {
		return StatusWarn, metrics, fmt.Sprintf("部分依赖服务不健康: %d", int(unhealthy))
	}
	return StatusPass, metrics, "所有依赖服务正常"
}

// calculateScore 计算健康分数
func (pm *PatrolManager) calculateScore(check *PatrolCheck, metrics map[string]float64) float64 {
	baseScore := 100.0

	// 根据指标偏差扣分
	for metric, value := range metrics {
		baseline, ok := pm.baselines[metric]
		if !ok {
			continue
		}

		// 计算与基线的偏差
		deviation := 0.0
		if baseline.StdDev > 0 {
			deviation = (value - baseline.Mean) / baseline.StdDev
		}

		// 偏差大于2个标准差扣分
		if deviation > 2 {
			baseScore -= deviation * 5
		} else if deviation > 1 {
			baseScore -= deviation * 2
		}
	}

	if baseScore < 0 {
		baseScore = 0
	}
	return baseScore
}

// generateSuggestions 生成优化建议
func (pm *PatrolManager) generateSuggestions(check *PatrolCheck, metrics map[string]float64, status CheckStatus) []string {
	suggestions := make([]string, 0)

	if status == StatusPass {
		return suggestions
	}

	switch check.ID {
	case "check-cpu-usage":
		if metrics["cpu_usage"].(float64) > 70 {
			suggestions = append(suggestions, "考虑增加计算资源或优化高CPU占用的进程")
			suggestions = append(suggestions, "检查是否有异常进程消耗过多CPU")
		}
	case "check-memory-usage":
		if metrics["memory_usage"].(float64) > 75 {
			suggestions = append(suggestions, "检查内存泄漏问题")
			suggestions = append(suggestions, "考虑增加内存或调整应用内存配置")
		}
	case "check-disk-usage":
		if metrics["disk_usage"].(float64) > 80 {
			suggestions = append(suggestions, "清理不必要的日志和临时文件")
			suggestions = append(suggestions, "考虑扩展存储容量")
		}
	case "check-network-latency":
		if metrics["latency_ms"].(float64) > 50 {
			suggestions = append(suggestions, "检查网络拥塞情况")
			suggestions = append(suggestions, "考虑启用连接复用")
		}
	case "check-error-rate":
		if metrics["error_rate"].(float64) > 5 {
			suggestions = append(suggestions, "分析错误日志找出根本原因")
			suggestions = append(suggestions, "检查最近部署的变更")
		}
	case "check-response-time":
		if metrics["response_time_ms"].(float64) > 200 {
			suggestions = append(suggestions, "优化数据库查询")
			suggestions = append(suggestions, "考虑增加缓存层")
		}
	}

	return suggestions
}

// detectAnomalies 检测异常
func (pm *PatrolManager) detectAnomalies(check *PatrolCheck, metrics map[string]float64) {
	for metric, value := range metrics {
		baseline, ok := pm.baselines[metric]
		if !ok || baseline.StdDev == 0 {
			continue
		}

		deviation := (value - baseline.Mean) / baseline.StdDev

		// 偏差超过3个标准差视为异常
		if deviation > 3 || deviation < -3 {
			severity := "medium"
			if deviation > 5 || deviation < -5 {
				severity = "high"
			}
			if deviation > 8 || deviation < -8 {
				severity = "critical"
			}

			anomaly := &AnomalyEvent{
				ID:          fmt.Sprintf("anomaly-%d", time.Now().UnixNano()),
				Type:        "outlier",
				Source:      check.Name,
				Metric:      metric,
				Value:       value,
				Baseline:    baseline.Mean,
				Deviation:   deviation,
				Severity:    severity,
				Description: fmt.Sprintf("%s 指标 %.1f 超出正常范围(基线 %.1f)", metric, value, baseline.Mean),
				Timestamp:   time.Now(),
			}

			pm.mu.Lock()
			pm.anomalies = append(pm.anomalies, anomaly)
			pm.mu.Unlock()
		}
	}
}

// RunFullPatrol 执行完整巡检
func (pm *PatrolManager) RunFullPatrol(ctx context.Context) (*PatrolReport, error) {
	report := &PatrolReport{
		ID:          fmt.Sprintf("report-%d", time.Now().UnixNano()),
		Name:        fmt.Sprintf("系统巡检报告 %s", time.Now().Format("2006-01-02 15:04")),
		StartTime:   time.Now(),
		Results:     make([]*CheckResult, 0),
		GeneratedAt: time.Now(),
	}

	// 执行所有启用的检查
	pm.mu.RLock()
	checks := make([]*PatrolCheck, 0)
	for _, check := range pm.checks {
		if check.Enabled {
			checks = append(checks, check)
		}
	}
	pm.mu.RUnlock()

	for _, check := range checks {
		result, err := pm.RunCheck(ctx, check.ID)
		if err != nil {
			result = &CheckResult{
				CheckID: check.ID,
				Name:    check.Name,
				Status:  StatusSkipped,
				Message: err.Error(),
			}
		}
		report.Results = append(report.Results, result)

		report.TotalChecks++
		switch result.Status {
		case StatusPass:
			report.PassCount++
		case StatusWarn:
			report.WarnCount++
		case StatusFail:
			report.FailCount++
		case StatusSkipped:
			report.SkipCount++
		}
	}

	// 计算总体分数
	report.OverallScore = pm.calculateOverallScore(report)
	report.HealthLevel = pm.getHealthLevel(report.OverallScore)
	report.Summary = pm.generateSummary(report)

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	// 保存报告
	pm.mu.Lock()
	pm.reports = append(pm.reports, report)
	pm.mu.Unlock()

	return report, nil
}

// calculateOverallScore 计算总体分数
func (pm *PatrolManager) calculateOverallScore(report *PatrolReport) float64 {
	if report.TotalChecks == 0 {
		return 100
	}

	// 加权计算
	weights := map[CheckStatus]float64{
		StatusPass:    100,
		StatusWarn:    70,
		StatusFail:    0,
		StatusSkipped: 50,
	}

	totalWeight := 0.0
	for _, result := range report.Results {
		totalWeight += weights[result.Status]
	}

	return totalWeight / float64(report.TotalChecks)
}

// getHealthLevel 获取健康等级
func (pm *PatrolManager) getHealthLevel(score float64) string {
	switch {
	case score >= 95:
		return "excellent"
	case score >= 85:
		return "good"
	case score >= 70:
		return "fair"
	case score >= 50:
		return "poor"
	default:
		return "critical"
	}
}

// generateSummary 生成摘要
func (pm *PatrolManager) generateSummary(report *PatrolReport) string {
	summary := fmt.Sprintf("共检查 %d 项，", report.TotalChecks)
	summary += fmt.Sprintf("通过 %d 项，", report.PassCount)
	summary += fmt.Sprintf("警告 %d 项，", report.WarnCount)
	summary += fmt.Sprintf("失败 %d 项。", report.FailCount)
	summary += fmt.Sprintf("健康评分: %.1f，等级: %s。", report.OverallScore, report.HealthLevel)

	if report.FailCount > 0 {
		summary += "建议立即处理失败的检查项。"
	} else if report.WarnCount > 0 {
		summary += "建议关注警告项并进行优化。"
	}

	return summary
}

// StartScheduler 启动调度器
func (pm *PatrolManager) StartScheduler(ctx context.Context) {
	pm.mu.Lock()
	pm.scheduler.running = true
	pm.mu.Unlock()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-pm.scheduler.stopCh:
				return
			case <-ticker.C:
				pm.runScheduledChecks(ctx)
			}
		}
	}()
}

// StopScheduler 停止调度器
func (pm *PatrolManager) StopScheduler() {
	pm.mu.Lock()
	pm.scheduler.running = false
	pm.mu.Unlock()

	close(pm.scheduler.stopCh)
}

// runScheduledChecks 运行定时检查
func (pm *PatrolManager) runScheduledChecks(ctx context.Context) {
	pm.mu.RLock()
	checks := make([]*PatrolCheck, 0)
	now := time.Now()
	for _, check := range pm.checks {
		if check.Enabled && now.After(check.NextRun) {
			checks = append(checks, check)
		}
	}
	pm.mu.RUnlock()

	for _, check := range checks {
		pm.RunCheck(ctx, check.ID)
	}
}

// GetLatestReport 获取最新报告
func (pm *PatrolManager) GetLatestReport() *PatrolReport {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.reports) == 0 {
		return nil
	}
	return pm.reports[len(pm.reports)-1]
}

// GetReports 获取报告列表
func (pm *PatrolManager) GetReports(limit int) []*PatrolReport {
	pm.mu.RLock()
	size := len(pm.reports)
	if limit > size || limit <= 0 {
		limit = size
	}
	reports := pm.reports[size-limit:]
	pm.mu.RUnlock()
	return reports
}

// GetAnomalies 获取异常列表
func (pm *PatrolManager) GetAnomalies(limit int) []*AnomalyEvent {
	pm.mu.RLock()
	size := len(pm.anomalies)
	if limit > size || limit <= 0 {
		limit = size
	}
	anomalies := pm.anomalies[size-limit:]
	pm.mu.RUnlock()
	return anomalies
}

// GenerateOptimizationReport 生成优化报告
func (pm *PatrolManager) GenerateOptimizationReport() []*OptimizationSuggestion {
	suggestions := make([]*OptimizationSuggestion, 0)

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 基于最近的检查结果生成建议
	for _, result := range pm.results {
		if result.Status == StatusPass {
			continue
		}

		for i, sug := range result.Suggestions {
			priority := "medium"
			if result.Status == StatusFail {
				priority = "high"
			}

			suggestion := &OptimizationSuggestion{
				ID:           fmt.Sprintf("sug-%d-%d", time.Now().UnixNano(), i),
				Category:     string(result.CheckType),
				Priority:     priority,
				Title:        fmt.Sprintf("%s 优化建议", result.Name),
				Description:  sug,
				Impact:       "提升系统稳定性和性能",
				Effort:       "medium",
				RelatedCheck: result.CheckID,
				CreatedAt:    time.Now(),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// AddCheck 添加检查项
func (pm *PatrolManager) AddCheck(check *PatrolCheck) error {
	if check.ID == "" {
		check.ID = fmt.Sprintf("check-%d", time.Now().UnixNano())
	}
	check.NextRun = time.Now().Add(check.Interval)

	pm.mu.Lock()
	pm.checks[check.ID] = check
	pm.mu.Unlock()

	return nil
}

// SetNotifyFunc 设置通知函数
func (pm *PatrolManager) SetNotifyFunc(fn func(level string, message string)) {
	pm.mu.Lock()
	pm.notifyFunc = fn
	pm.mu.Unlock()
}

// GetStatistics 获取统计信息
func (pm *PatrolManager) GetStatistics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"total_checks":    len(pm.checks),
		"enabled_checks":  countEnabled(pm.checks),
		"total_reports":   len(pm.reports),
		"total_anomalies": len(pm.anomalies),
		"scheduler_running": pm.scheduler.running,
		"baseline_count":  len(pm.baselines),
	}
}

func countEnabled(checks map[string]*PatrolCheck) int {
	count := 0
	for _, c := range checks {
		if c.Enabled {
			count++
		}
	}
	return count
}

// ExportReport 导出报告
func (pm *PatrolManager) ExportReport(reportID string) ([]byte, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, r := range pm.reports {
		if r.ID == reportID {
			return json.MarshalIndent(r, "", "  ")
		}
	}
	return nil, fmt.Errorf("报告不存在: %s", reportID)
}
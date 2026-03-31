// Package smart - 智能调度器模块
package smart

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

// PredictionType 预测类型
type PredictionType string

const (
	PredictLoad      PredictionType = "load"       // 负载预测
	PredictResource  PredictionType = "resource"   // 资源预测
	PredictLatency   PredictionType = "latency"    // 延迟预测
	PredictCost      PredictionType = "cost"       // 成本预测
	PredictTaskCount PredictionType = "task_count" // 任务数量预测
)

// OptimizationType 优化类型
type OptimizationType string

const (
	OptMinCost      OptimizationType = "min_cost"      // 最小成本
	OptMinLatency   OptimizationType = "min_latency"   // 最小延迟
	OptMaxResource  OptimizationType = "max_resource"  // 最大资源利用率
	OptMaxThroughput OptimizationType = "max_throughput" // 最大吞吐量
	OptBalance      OptimizationType = "balance"      // 平衡优化
)

// PredictionModel 预测模型
type PredictionModel struct {
	ID           string            `json:"id"`
	Type         PredictionType    `json:"type"`
	Algorithm    string            `json:"algorithm"` // linear, exponential, arima, ml
	Parameters   map[string]float64 `json:"parameters"`
	Accuracy     float64           `json:"accuracy"` // 0-1
	LastTrained  time.Time         `json:"last_trained"`
	DataPoints   int               `json:"data_points"`
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Timestamps []time.Time `json:"timestamps"`
	Values     []float64   `json:"values"`
	Labels     []string    `json:"labels,omitempty"`
}

// PredictionResult 预测结果
type PredictionResult struct {
	ID          string        `json:"id"`
	Type        PredictionType `json:"type"`
	Timestamp   time.Time     `json:"timestamp"`
	Predicted   []float64     `json:"predicted"`
	Actual      []float64     `json:"actual,omitempty"`
	Confidence  float64       `json:"confidence"`
	Horizon     int           `json:"horizon"` // 预测步数
	StepSize    time.Duration `json:"step_size"`
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	ID            string    `json:"id"`
	Type          OptimizationType `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	CurrentValue  float64   `json:"current_value"`
	OptimizedValue float64  `json:"optimized_value"`
	Improvement   float64   `json:"improvement"` // 改进百分比
	Actions       []string  `json:"actions"`
	Constraints   map[string]interface{} `json:"constraints"`
}

// AgentScore Agent评分
type AgentScore struct {
	AgentID      string    `json:"agent_id"`
	Score        float64   `json:"score"` // 0-100
	LoadFactor   float64   `json:"load_factor"`
	LatencyFactor float64  `json:"latency_factor"`
	ResourceFactor float64 `json:"resource_factor"`
	HistoryScore float64   `json:"history_score"` // 历史成功率
	PredictedLoad float64  `json:"predicted_load"` // 预测负载
	LastUpdated  time.Time `json:"last_updated"`
}

// SmartScheduleResult 智能调度结果
type SmartScheduleResult struct {
	ID           string       `json:"id"`
	TaskID       string       `json:"task_id"`
	SelectedAgent string      `json:"selected_agent"`
	Score        float64      `json:"score"`
	Alternative  []AgentScore `json:"alternative"`
	PredictedWait float64     `json:"predicted_wait"` // 预计等待时间
	PredictedExec float64     `json:"predicted_exec"` // 预计执行时间
	Confidence   float64      `json:"confidence"`
	Reason       string       `json:"reason"`
	Timestamp    time.Time    `json:"timestamp"`
}

// HistoricalRecord 历史记录
type HistoricalRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	TaskID      string    `json:"task_id"`
	AgentID     string    `json:"agent_id"`
	Load        float64   `json:"load"`
	Latency     float64   `json:"latency"` // 毫秒
	ResourceUsed float64  `json:"resource_used"`
	Success     bool      `json:"success"`
	Duration    float64   `json:"duration"` // 毫秒
}

// SmartSchedulerConfig 智能调度器配置
type SmartSchedulerConfig struct {
	HistoryWindowSize    int           `json:"history_window_size"`    // 历史窗口大小
	PredictionHorizon    int           `json:"prediction_horizon"`     // 预测步数
	UpdateInterval       time.Duration `json:"update_interval"`        // 更新间隔
	MinDataPoints        int           `json:"min_data_points"`        // 最小数据点
	LearningRate         float64       `json:"learning_rate"`          // 学习率
	OptimizationType     OptimizationType `json:"optimization_type"`
	EnableAutoAdjust     bool          `json:"enable_auto_adjust"`     // 自动调整参数
	EnablePredictive     bool          `json:"enable_predictive"`      // 启用预测调度
}

// SmartScheduler 智能调度器
type SmartScheduler struct {
	config       SmartSchedulerConfig
	models       map[PredictionType]*PredictionModel
	history      []*HistoricalRecord
	agentScores  map[string]*AgentScore
	optimizations []*OptimizationResult
	mu           sync.RWMutex
}

// NewSmartScheduler 创建智能调度器
func NewSmartScheduler(config SmartSchedulerConfig) *SmartScheduler {
	return &SmartScheduler{
		config:        config,
		models:        make(map[PredictionType]*PredictionModel),
		history:       make([]*HistoricalRecord, 0),
		agentScores:   make(map[string]*AgentScore),
		optimizations: make([]*OptimizationResult, 0),
	}
}

// Initialize 初始化
func (ss *SmartScheduler) Initialize() error {
	// 初始化预测模型
	ss.initModels()
	return nil
}

// initModels 初始化预测模型
func (ss *SmartScheduler) initModels() {
	models := []PredictionModel{
		{
			ID:          "load-predict",
			Type:        PredictLoad,
			Algorithm:   "exponential",
			Parameters:  map[string]float64{"alpha": 0.3, "beta": 0.1},
			Accuracy:    0.75,
			LastTrained: time.Now(),
		},
		{
			ID:          "latency-predict",
			Type:        PredictLatency,
			Algorithm:   "linear",
			Parameters:  map[string]float64{"slope": 0.5, "intercept": 10},
			Accuracy:    0.70,
			LastTrained: time.Now(),
		},
		{
			ID:          "resource-predict",
			Type:        PredictResource,
			Algorithm:   "arima",
			Parameters:  map[string]float64{"p": 1, "d": 1, "q": 1},
			Accuracy:    0.65,
			LastTrained: time.Now(),
		},
		{
			ID:          "cost-predict",
			Type:        PredictCost,
			Algorithm:   "ml",
			Parameters:  map[string]float64{"layers": 3, "units": 64},
			Accuracy:    0.80,
			LastTrained: time.Now(),
		},
	}

	ss.mu.Lock()
	for _, m := range models {
		ss.models[m.Type] = &m
	}
	ss.mu.Unlock()
}

// AddHistoricalRecord 添加历史记录
func (ss *SmartScheduler) AddHistoricalRecord(record *HistoricalRecord) {
	ss.mu.Lock()
	ss.history = append(ss.history, record)

	// 保持窗口大小
	if len(ss.history) > ss.config.HistoryWindowSize {
		ss.history = ss.history[1:]
	}

	// 更新Agent评分
	ss.updateAgentScore(record)

	ss.mu.Unlock()
}

// updateAgentScore 更新Agent评分
func (ss *SmartScheduler) updateAgentScore(record *HistoricalRecord) {
	score, ok := ss.agentScores[record.AgentID]
	if !ok {
		score = &AgentScore{
			AgentID:     record.AgentID,
			LastUpdated: time.Now(),
		}
		ss.agentScores[record.AgentID] = score
	}

	// 计算评分因子
	// 负载因子: 低负载 = 高分
	score.LoadFactor = 100 - record.Load

	// 延迟因子: 低延迟 = 高分 (假设100ms为基准)
	latencyNorm := math.Min(record.Latency/100, 1)
	score.LatencyFactor = 100 - latencyNorm*100

	// 资源因子: 资源利用率适中 = 高分 (50%最佳)
	resourceDiff := math.Abs(record.ResourceUsed - 50)
	score.ResourceFactor = 100 - resourceDiff

	// 历史成功率
	if record.Success {
		score.HistoryScore = math.Min(score.HistoryScore + 5, 100)
	} else {
		score.HistoryScore = math.Max(score.HistoryScore - 10, 0)
	}

	// 综合评分 (加权平均)
	score.Score = score.LoadFactor*0.3 +
		score.LatencyFactor*0.25 +
		score.ResourceFactor*0.2 +
		score.HistoryScore*0.25

	score.LastUpdated = time.Now()
}

// Predict 预测
func (ss *SmartScheduler) Predict(ctx context.Context, predictType PredictionType, horizon int) (*PredictionResult, error) {
	ss.mu.RLock()
	model, ok := ss.models[predictType]
	history := ss.getHistoryValues(predictType)
	ss.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("预测模型不存在: %s", predictType)
	}

	if len(history) < ss.config.MinDataPoints {
		return nil, fmt.Errorf("数据点不足: 需要%d，当前%d", ss.config.MinDataPoints, len(history))
	}

	// 执行预测
	predicted := ss.runPrediction(model, history, horizon)

	result := &PredictionResult{
		ID:         fmt.Sprintf("pred-%d", time.Now().UnixNano()),
		Type:       predictType,
		Timestamp:  time.Now(),
		Predicted:  predicted,
		Confidence: model.Accuracy,
		Horizon:    horizon,
	}

	return result, nil
}

// runPrediction 执行预测算法
func (ss *SmartScheduler) runPrediction(model *PredictionModel, history []float64, horizon int) []float64 {
	predicted := make([]float64, horizon)

	switch model.Algorithm {
	case "exponential":
		// 指数平滑
		alpha := model.Parameters["alpha"]
		lastValue := history[len(history)-1]
		for i := 0; i < horizon; i++ {
			predicted[i] = alpha * lastValue + (1 - alpha) * predicted[i-1]
			if i == 0 {
				predicted[i] = alpha * lastValue + (1 - alpha) * lastValue
			}
		}

	case "linear":
		// 线性回归
		slope := model.Parameters["slope"]
		intercept := model.Parameters["intercept"]
		lastIdx := len(history)
		for i := 0; i < horizon; i++ {
			predicted[i] = intercept + slope * float64(lastIdx + i + 1)
		}

	case "arima":
		// 简化ARIMA (使用移动平均)
		window := 5
		if len(history) < window {
			window = len(history)
		}
		avg := 0.0
		for i := len(history) - window; i < len(history); i++ {
			avg += history[i]
		}
		avg /= float64(window)
		for i := 0; i < horizon; i++ {
			predicted[i] = avg
		}

	case "ml":
		// 简化ML预测 (使用趋势+周期)
		trend := 0.0
		if len(history) >= 2 {
			trend = history[len(history)-1] - history[len(history)-2]
		}
		base := history[len(history)-1]
		for i := 0; i < horizon; i++ {
			predicted[i] = base + trend * float64(i + 1)
		}

	default:
		// 默认: 最后一个值
		lastValue := history[len(history)-1]
		for i := 0; i < horizon; i++ {
			predicted[i] = lastValue
		}
	}

	return predicted
}

// getHistoryValues 获取历史值序列
func (ss *SmartScheduler) getHistoryValues(predictType PredictionType) []float64 {
	values := make([]float64, 0)
	for _, r := range ss.history {
		switch predictType {
		case PredictLoad:
			values = append(values, r.Load)
		case PredictLatency:
			values = append(values, r.Latency)
		case PredictResource:
			values = append(values, r.ResourceUsed)
		case PredictCost:
			// 模拟成本计算
			cost := r.Duration * 0.001 // 每毫秒0.001成本单位
			values = append(values, cost)
		}
	}
	return values
}

// Optimize 优化调度
func (ss *SmartScheduler) Optimize(ctx context.Context, optType OptimizationType) (*OptimizationResult, error) {
	ss.mu.RLock()
	agentScores := ss.agentScores
	ss.mu.RUnlock()

	if len(agentScores) == 0 {
		return nil, fmt.Errorf("没有Agent数据")
	}

	// 计算当前值和优化值
	currentValue, optimizedValue, actions := ss.runOptimization(optType, agentScores)

	improvement := 0.0
	if currentValue != 0 {
		improvement = ((optimizedValue - currentValue) / currentValue) * 100
		if optType == OptMinCost || optType == OptMinLatency {
			improvement = -improvement // 成本/延迟优化，负数表示减少
		}
	}

	result := &OptimizationResult{
		ID:             fmt.Sprintf("opt-%d", time.Now().UnixNano()),
		Type:           optType,
		Timestamp:      time.Now(),
		CurrentValue:   currentValue,
		OptimizedValue: optimizedValue,
		Improvement:    improvement,
		Actions:        actions,
	}

	ss.mu.Lock()
	ss.optimizations = append(ss.optimizations, result)
	ss.mu.Unlock()

	return result, nil
}

// runOptimization 执行优化算法
func (ss *SmartScheduler) runOptimization(optType OptimizationType, agentScores map[string]*AgentScore) (float64, float64, []string) {
	actions := make([]string, 0)
	currentValue := 0.0
	optimizedValue := 0.0

	// 计算当前指标
	totalLoad := 0.0
	totalLatency := 0.0
	totalResource := 0.0
	bestAgent := ""
	bestScore := 0.0

	for _, score := range agentScores {
		totalLoad += score.LoadFactor
		totalLatency += score.LatencyFactor
		totalResource += score.ResourceFactor

		if score.Score > bestScore {
			bestScore = score.Score
			bestAgent = score.AgentID
		}
	}
	count := float64(len(agentScores))

	switch optType {
	case OptMinCost:
		// 最小成本: 选择性价比最高的Agent
		currentValue = totalLatency / count * 0.001 // 当前平均成本
		optimizedValue = 0.8 * currentValue // 优化后成本降低20%
		actions = []string{
			"优先使用低延迟Agent",
			"合并小任务减少调度次数",
			"使用批量执行降低成本",
		}

	case OptMinLatency:
		// 最小延迟: 选择响应最快的Agent
		currentValue = 100 - totalLatency / count // 当前平均延迟指标
		optimizedValue = bestAgent != "" ? agentScores[bestAgent].LatencyFactor : currentValue * 1.2
		actions = []string{
			fmt.Sprintf("优先调度到Agent: %s", bestAgent),
			"预加载常用技能减少初始化时间",
			"使用就近数据中心",
		}

	case OptMaxResource:
		// 最大资源利用率: 平衡分配
		currentValue = totalResource / count
		optimizedValue = 75.0 // 目标利用率75%
		actions = []string{
			"动态调整任务分配",
			"启用自动伸缩",
			"优化任务粒度",
		}

	case OptMaxThroughput:
		// 最大吞吐量
		currentValue = count * 10 // 当前吞吐量估算
		optimizedValue = currentValue * 1.5
		actions = []string{
			"增加并行任务数",
			"优化任务队列深度",
			"使用异步处理",
		}

	case OptBalance:
		// 平衡优化
		currentValue = (totalLoad + totalLatency + totalResource) / (count * 3)
		optimizedValue = 80.0 // 目标平衡值80
		actions = []string{
			"使用混合调度策略",
			"动态调整权重",
			"定期重新评估Agent",
		}
	}

	return currentValue, optimizedValue, actions
}

// SmartSchedule 智能调度选择
func (ss *SmartScheduler) SmartSchedule(ctx context.Context, taskID string, requiredSkills []string) (*SmartScheduleResult, error) {
	ss.mu.RLock()
	agentScores := ss.agentScores
	ss.mu.RUnlock()

	if len(agentScores) == 0 {
		return nil, fmt.Errorf("没有可用的Agent")
	}

	// 如果启用预测调度
	if ss.config.EnablePredictive {
		// 获取负载预测
		predLoad, err := ss.Predict(ctx, PredictLoad, 1)
		if err == nil && len(predLoad.Predicted) > 0 {
			// 更新预测负载
			ss.mu.Lock()
			for id, score := range ss.agentScores {
				// 模拟预测值分配
				score.PredictedLoad = predLoad.Predicted[0] * (1 + 0.1*float64(len(agentScores)-len(agentScores)/2))
				ss.agentScores[id] = score
			}
			ss.mu.Unlock()
		}
	}

	// 计算最终评分并选择
	bestAgent := ""
	bestFinalScore := -1.0
	alternatives := make([]AgentScore, 0)

	ss.mu.RLock()
	for _, score := range ss.agentScores {
		// 考虑预测负载调整评分
		finalScore := score.Score
		if ss.config.EnablePredictive && score.PredictedLoad > 0 {
			// 预测负载高则降低评分
			loadPenalty := math.Min(score.PredictedLoad/100, 0.5)
			finalScore = finalScore - loadPenalty * 20
		}

		// 检查技能匹配 (简化: 所有Agent都有技能)
		// 实际应用中需要检查Agent的能力列表

		altScore := *score
		altScore.Score = finalScore
		alternatives = append(alternatives, altScore)

		if finalScore > bestFinalScore {
			bestFinalScore = finalScore
			bestAgent = score.AgentID
		}
	}
	ss.mu.RUnlock()

	if bestAgent == "" {
		return nil, fmt.Errorf("无法找到合适的Agent")
	}

	// 估算执行时间
	predictedExec := 50.0 // 默认50ms
	predictedWait := 0.0

	ss.mu.RLock()
	if best, ok := ss.agentScores[bestAgent]; ok {
		predictedWait = best.Load * 10 // 负载越高等待越长
		predictedExec = 100 - best.LatencyFactor + 10 // 基于延迟因子估算
	}
	ss.mu.RUnlock()

	result := &SmartScheduleResult{
		ID:            fmt.Sprintf("sched-%d", time.Now().UnixNano()),
		TaskID:        taskID,
		SelectedAgent: bestAgent,
		Score:         bestFinalScore,
		Alternative:   alternatives,
		PredictedWait: predictedWait,
		PredictedExec: predictedExec,
		Confidence:    0.85,
		Reason:        fmt.Sprintf("Agent %s 评分最高 (%.2f)，预测负载 %.1f%%", bestAgent, bestFinalScore, predictedWait/10),
		Timestamp:     time.Now(),
	}

	return result, nil
}

// GetAgentScores 获取Agent评分
func (ss *SmartScheduler) GetAgentScores() map[string]*AgentScore {
	ss.mu.RLock()
	scores := make(map[string]*AgentScore)
	for k, v := range ss.agentScores {
		scores[k] = v
	}
	ss.mu.RUnlock()
	return scores
}

// GetPredictionModel 获取预测模型
func (ss *SmartScheduler) GetPredictionModel(predictType PredictionType) (*PredictionModel, error) {
	ss.mu.RLock()
	model, ok := ss.models[predictType]
	ss.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("模型不存在: %s", predictType)
	}
	return model, nil
}

// UpdateModel 更新模型参数
func (ss *SmartScheduler) UpdateModel(predictType PredictionType, params map[string]float64) error {
	ss.mu.Lock()
	model, ok := ss.models[predictType]
	if !ok {
		ss.mu.Unlock()
		return fmt.Errorf("模型不存在: %s", predictType)
	}

	// 更新参数
	for k, v := range params {
		model.Parameters[k] = v
	}
	model.LastTrained = time.Now()

	// 计算新的准确率 (简化)
	model.Accuracy = math.Min(model.Accuracy + 0.05, 0.95)

	ss.mu.Unlock()
	return nil
}

// GetOptimizationHistory 获取优化历史
func (ss *SmartScheduler) GetOptimizationHistory(limit int) []*OptimizationResult {
	ss.mu.RLock()
	size := len(ss.optimizations)
	if limit > size || limit <= 0 {
		limit = size
	}
	history := ss.optimizations[size-limit:]
	ss.mu.RUnlock()
	return history
}

// GetHistoricalRecords 获取历史记录
func (ss *SmartScheduler) GetHistoricalRecords(limit int) []*HistoricalRecord {
	ss.mu.RLock()
	size := len(ss.history)
	if limit > size || limit <= 0 {
		limit = size
	}
	records := ss.history[size-limit:]
	ss.mu.RUnlock()
	return records
}

// ExportModel 导出模型
func (ss *SmartScheduler) ExportModel(predictType PredictionType) ([]byte, error) {
	ss.mu.RLock()
	model, ok := ss.models[predictType]
	ss.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("模型不存在: %s", predictType)
	}

	return json.MarshalIndent(model, "", "  ")
}

// ImportModel 导入模型
func (ss *SmartScheduler) ImportModel(data []byte) error {
	model := &PredictionModel{}
	if err := json.Unmarshal(data, model); err != nil {
		return err
	}

	ss.mu.Lock()
	ss.models[model.Type] = model
	ss.mu.Unlock()

	return nil
}

// GetStatistics 获取统计信息
func (ss *SmartScheduler) GetStatistics() map[string]interface{} {
	ss.mu.RLock()
	stats := map[string]interface{}{
		"history_count":     len(ss.history),
		"agent_count":       len(ss.agentScores),
		"optimization_count": len(ss.optimizations),
		"model_count":       len(ss.models),
		"config":            ss.config,
	}

	// 计算平均指标
	if len(ss.agentScores) > 0 {
		avgScore := 0.0
		avgLoad := 0.0
		for _, s := range ss.agentScores {
			avgScore += s.Score
			avgLoad += s.LoadFactor
		}
		count := float64(len(ss.agentScores))
		stats["avg_agent_score"] = avgScore / count
		stats["avg_agent_load"] = 100 - avgLoad / count
	}

	ss.mu.RUnlock()
	return stats
}
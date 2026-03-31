// Package collab - 结果聚合器
package collab

import (
	"sync"
	"time"
)

// ResultAggregator 结果聚合器
type ResultAggregator struct {
	strategy AggregationStrategy
	mu       sync.RWMutex
}

// AggregationStrategy 聚合策略
type AggregationStrategy string

const (
	StrategyMerge     AggregationStrategy = "merge"     // 合并所有结果
	StrategyBest      AggregationStrategy = "best"      // 选择最佳结果
	StrategyConsensus AggregationStrategy = "consensus" // 共识结果
	StrategyVote      AggregationStrategy = "vote"      // 投票结果
	StrategyAverage   AggregationStrategy = "average"   // 平均值
)

// NewResultAggregator 创建结果聚合器
func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{
		strategy: StrategyMerge,
	}
}

// SetStrategy 设置聚合策略
func (a *ResultAggregator) SetStrategy(strategy AggregationStrategy) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.strategy = strategy
}

// Aggregate 聚合结果
func (a *ResultAggregator) Aggregate(collab *Collaboration) *CollabResult {
	a.mu.RLock()
	strategy := a.strategy
	a.mu.RUnlock()

	result := &CollabResult{
		Output:       make(map[string]interface{}),
		TasksTotal:   len(collab.Tasks),
		TasksSuccess: 0,
		TasksFailed:  0,
		Errors:       make([]string, 0),
	}

	// 统计任务结果
	taskOutputs := make([]map[string]interface{}, 0)
	for _, task := range collab.Tasks {
		switch task.State {
		case TaskStateCompleted:
			result.TasksSuccess++
			if task.Output != nil {
				taskOutputs = append(taskOutputs, task.Output)
			}
		case TaskStateFailed:
			result.TasksFailed++
			result.Errors = append(result.Errors, task.ID+" failed")
		}
	}

	// 根据策略聚合输出
	result.Output = a.aggregateOutputs(taskOutputs, strategy)

	// 计算成功率
	if result.TasksTotal > 0 {
		result.Success = result.TasksSuccess == result.TasksTotal
	}

	// 计算成本
	result.Cost = a.calculateCost(collab)

	return result
}

// aggregateOutputs 聚合输出
func (a *ResultAggregator) aggregateOutputs(outputs []map[string]interface{}, strategy AggregationStrategy) map[string]interface{} {
	if len(outputs) == 0 {
		return make(map[string]interface{})
	}

	switch strategy {
	case StrategyMerge:
		return a.mergeOutputs(outputs)

	case StrategyBest:
		return a.selectBest(outputs)

	case StrategyConsensus:
		return a.findConsensus(outputs)

	case StrategyVote:
		return a.voteResult(outputs)

	case StrategyAverage:
		return a.averageOutputs(outputs)

	default:
		return a.mergeOutputs(outputs)
	}
}

// mergeOutputs 合并输出
func (a *ResultAggregator) mergeOutputs(outputs []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, output := range outputs {
		for k, v := range output {
			// 如果键已存在，转为数组
			if existing, ok := result[k]; ok {
				if arr, ok := existing.([]interface{}); ok {
					result[k] = append(arr, v)
				} else {
					result[k] = []interface{}{existing, v}
				}
			} else {
				result[k] = v
			}
		}
	}

	return result
}

// selectBest 选择最佳结果
func (a *ResultAggregator) selectBest(outputs []map[string]interface{}) map[string]interface{} {
	best := outputs[0]
	bestScore := a.scoreOutput(best)

	for _, output := range outputs[1:] {
		score := a.scoreOutput(output)
		if score > bestScore {
			best = output
			bestScore = score
		}
	}

	return best
}

// scoreOutput 评分输出
func (a *ResultAggregator) scoreOutput(output map[string]interface{}) float64 {
	score := 0.0

	// 有status字段加分
	if status, ok := output["status"]; ok {
		if status == "completed" || status == "success" {
			score += 10.0
		}
	}

	// 有结果数据加分
	if result, ok := output["result"]; ok {
		if result != nil {
			score += 5.0
		}
	}

	// 字段数量加分
	score += float64(len(output)) * 0.5

	return score
}

// findConsensus 查找共识
func (a *ResultAggregator) findConsensus(outputs []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 统计每个键值的出现次数
	valueCounts := make(map[string]map[interface{}]int)

	for _, output := range outputs {
		for k, v := range output {
			if valueCounts[k] == nil {
				valueCounts[k] = make(map[interface{}]int)
			}
			valueCounts[k][v]++
		}
	}

	// 选择出现次数最多的值
	for k, counts := range valueCounts {
		maxCount := 0
		var maxValue interface{}

		for v, count := range counts {
			if count > maxCount {
				maxCount = count
				maxValue = v
			}
		}

		// 需要超过半数
		if maxCount > len(outputs)/2 {
			result[k] = maxValue
		}
	}

	return result
}

// voteResult 投票结果
func (a *ResultAggregator) voteResult(outputs []map[string]interface{}) map[string]interface{} {
	// 简化实现 - 选择出现次数最多的输出模式
	return a.findConsensus(outputs)
}

// averageOutputs 平均输出
func (a *ResultAggregator) averageOutputs(outputs []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 收集数值字段
	numericFields := make(map[string][]float64)

	for _, output := range outputs {
		for k, v := range output {
			if num, ok := toFloat64(v); ok {
				if numericFields[k] == nil {
					numericFields[k] = make([]float64, 0)
				}
				numericFields[k] = append(numericFields[k], num)
			}
		}
	}

	// 计算平均值
	for k, values := range numericFields {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			result[k] = sum / float64(len(values))
		}
	}

	// 合并非数值字段
	for _, output := range outputs {
		for k, v := range output {
			if _, ok := toFloat64(v); !ok {
				if _, exists := result[k]; !exists {
					result[k] = v
				}
			}
		}
	}

	return result
}

// calculateCost 计算成本
func (a *ResultAggregator) calculateCost(collab *Collaboration) float64 {
	cost := 0.0

	for _, task := range collab.Tasks {
		if task.State == TaskStateCompleted {
			// 基本成本计算
			cost += 1.0

			// 根据任务优先级加权
			cost += float64(task.Priority) * 0.1
		}
	}

	// 时间成本
	if collab.Result != nil {
		durationCost := float64(collab.Result.Duration.Milliseconds()) / 1000.0 * 0.01
		cost += durationCost
	}

	return cost
}

// toFloat64 转换为float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// AggregatedStats 聚合统计
type AggregatedStats struct {
	TotalCollaborations int64         `json:"total_collaborations"`
	SuccessRate         float64       `json:"success_rate"`
	AvgTasksPerCollab   float64       `json:"avg_tasks_per_collab"`
	AvgDuration         time.Duration `json:"avg_duration"`
	AvgCost             float64       `json:"avg_cost"`
	ByType              map[string]int64 `json:"by_type"`
}

// GetStats 获取统计
func (a *ResultAggregator) GetStats(collabs []*Collaboration) *AggregatedStats {
	stats := &AggregatedStats{
		TotalCollaborations: int64(len(collabs)),
		ByType:              make(map[string]int64),
	}

	if len(collabs) == 0 {
		return stats
	}

	totalTasks := 0
	totalSuccess := 0
	totalDuration := time.Duration(0)
	totalCost := 0.0

	for _, collab := range collabs {
		totalTasks += len(collab.Tasks)

		if collab.Result != nil {
			if collab.Result.Success {
				totalSuccess++
			}
			totalDuration += collab.Result.Duration
			totalCost += collab.Result.Cost
		}

		stats.ByType[string(collab.Type)]++
	}

	stats.SuccessRate = float64(totalSuccess) / float64(len(collabs))
	stats.AvgTasksPerCollab = float64(totalTasks) / float64(len(collabs))
	stats.AvgDuration = totalDuration / time.Duration(len(collabs))
	stats.AvgCost = totalCost / float64(len(collabs))

	return stats
}
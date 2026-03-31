// Package decentralized - 信任管理
package decentralized

import (
	"fmt"
	"sync"
	"time"
)

// TrustManager 信任管理器
type TrustManager struct {
	scores      map[string]*TrustScore
	events      map[string][]*TrustEvent
	thresholds  *TrustThresholds
	algorithm   TrustAlgorithm
	mu          sync.RWMutex
}

// TrustScore 信任评分
type TrustScore struct {
	NodeID          string    `json:"node_id"`
	Score           float64   `json:"score"`           // 0-1
	Confidence      float64   `json:"confidence"`      // 置信度
	TotalEvents     int       `json:"total_events"`    // 总事件数
	PositiveEvents  int       `json:"positive_events"` // 正面事件数
	NegativeEvents  int       `json:"negative_events"` // 负面事件数
	LastUpdate      time.Time `json:"last_update"`
	Rating          TrustRating `json:"rating"`
	DecayFactor     float64   `json:"decay_factor"`    // 衰减因子
}

// TrustRating 信任等级
type TrustRating string

const (
	RatingUntrusted  TrustRating = "untrusted"  // 不信任 (< 0.3)
	RatingLow        TrustRating = "low"        // 低信任 (0.3-0.5)
	RatingMedium     TrustRating = "medium"     // 中等信任 (0.5-0.7)
	RatingHigh       TrustRating = "high"       // 高信任 (0.7-0.9)
	RatingFull       TrustRating = "full"       // 完全信任 (> 0.9)
)

// TrustEvent 信任事件
type TrustEvent struct {
	ID          string        `json:"id"`
	NodeID      string        `json:"node_id"`
	Type        TrustEventType `json:"type"`
	Impact      float64       `json:"impact"`      // 影响 -1.0 到 1.0
	Description string        `json:"description"`
	Timestamp   time.Time     `json:"timestamp"`
	Source      string        `json:"source"`      // 事件来源
	Verified    bool          `json:"verified"`
}

// TrustEventType 信任事件类型
type TrustEventType string

const (
	EventTaskCompleted   TrustEventType = "task_completed"   // 任务完成
	EventTaskFailed      TrustEventType = "task_failed"      // 任务失败
	EventResponseFast    TrustEventType = "response_fast"    // 响应快速
	EventResponseSlow    TrustEventType = "response_slow"    // 响应缓慢
	EventDataCorrect     TrustEventType = "data_correct"     // 数据正确
	EventDataCorrupted   TrustEventType = "data_corrupted"   // 数据损坏
	EventMaliciousAct    TrustEventType = "malicious_act"    // 恶意行为
	EventHelpfulAct      TrustEventType = "helpful_act"      // 帮助行为
	EventUptimeHigh      TrustEventType = "uptime_high"      // 高可用性
	EventDowntime        TrustEventType = "downtime"         // 宕机
	EventConsensusFollow TrustEventType = "consensus_follow" // 遵守共识
	EventConsensusBreak  TrustEventType = "consensus_break"  // 违反共识
)

// TrustThresholds 信任阈值
type TrustThresholds struct {
	UntrustedThreshold float64 `json:"untrusted_threshold"` // 不信任阈值
	LowThreshold       float64 `json:"low_threshold"`       // 低信任阈值
	MediumThreshold    float64 `json:"medium_threshold"`    // 中等信任阈值
	HighThreshold      float64 `json:"high_threshold"`      // 高信任阈值
	MinConfidence      float64 `json:"min_confidence"`      // 最小置信度
}

// TrustAlgorithm 信任算法
type TrustAlgorithm string

const (
	AlgorithmSimple     TrustAlgorithm = "simple"     // 简单平均
	AlgorithmWeighted   TrustAlgorithm = "weighted"   // 加权平均
	AlgorithmBayesian   TrustAlgorithm = "bayesian"   // 贝叶斯
	AlgorithmSliding    TrustAlgorithm = "sliding"    // 滑动窗口
	AlgorithmReputation TrustAlgorithm = "reputation" // 声誉系统
)

// NewTrustManager 创建信任管理器
func NewTrustManager() *TrustManager {
	return &TrustManager{
		scores: make(map[string]*TrustScore),
		events: make(map[string][]*TrustEvent),
		thresholds: &TrustThresholds{
			UntrustedThreshold: 0.3,
			LowThreshold:       0.5,
			MediumThreshold:    0.7,
			HighThreshold:      0.9,
			MinConfidence:      0.1,
		},
		algorithm: AlgorithmWeighted,
	}
}

// SetAlgorithm 设置算法
func (tm *TrustManager) SetAlgorithm(algorithm TrustAlgorithm) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.algorithm = algorithm
}

// InitializeTrust 初始化信任评分
func (tm *TrustManager) InitializeTrust(nodeID string, initialScore float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if initialScore < 0 {
		initialScore = 0
	}
	if initialScore > 1 {
		initialScore = 1
	}

	tm.scores[nodeID] = &TrustScore{
		NodeID:     nodeID,
		Score:      initialScore,
		Confidence: 0.1, // 初始低置信度
		LastUpdate: time.Now(),
		Rating:     tm.calculateRating(initialScore),
		DecayFactor: 0.99,
	}

	tm.events[nodeID] = make([]*TrustEvent, 0)
}

// UpdateTrustScore 更新信任评分
func (tm *TrustManager) UpdateTrustScore(nodeID string, success bool) {
	var impact float64
	var eventType TrustEventType

	if success {
		impact = 0.05
		eventType = EventTaskCompleted
	} else {
		impact = -0.1
		eventType = EventTaskFailed
	}

	tm.RecordEvent(nodeID, eventType, impact, "task execution")
}

// RecordEvent 记录事件
func (tm *TrustManager) RecordEvent(nodeID string, eventType TrustEventType, impact float64, description string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 创建事件
	event := &TrustEvent{
		ID:          generateEventID(),
		NodeID:      nodeID,
		Type:        eventType,
		Impact:      impact,
		Description: description,
		Timestamp:   time.Now(),
		Source:      "system",
		Verified:    true,
	}

	// 记录事件
	if tm.events[nodeID] == nil {
		tm.events[nodeID] = make([]*TrustEvent, 0)
	}
	tm.events[nodeID] = append(tm.events[nodeID], event)

	// 更新评分
	tm.updateScoreFromEvent(nodeID, event)
}

// updateScoreFromEvent 从事件更新评分
func (tm *TrustManager) updateScoreFromEvent(nodeID string, event *TrustEvent) {
	score, ok := tm.scores[nodeID]
	if !ok {
		score = &TrustScore{
			NodeID:     nodeID,
			Score:      0.5, // 初始中等信任
			Confidence: 0.1,
			LastUpdate: time.Now(),
			Rating:     RatingMedium,
			DecayFactor: 0.99,
		}
		tm.scores[nodeID] = score
	}

	// 更新统计
	score.TotalEvents++
	if event.Impact > 0 {
		score.PositiveEvents++
	} else if event.Impact < 0 {
		score.NegativeEvents++
	}

	// 应用影响
	switch tm.algorithm {
	case AlgorithmSimple:
		// 简单平均
		score.Score = (score.Score*float64(score.TotalEvents-1) + (score.Score + event.Impact)) / float64(score.TotalEvents)

	case AlgorithmWeighted:
		// 加权平均 - 新事件权重更高
		weight := 1.0 / float64(score.TotalEvents+1)
		score.Score = score.Score*(1-weight) + (score.Score+event.Impact)*weight

	case AlgorithmSliding:
		// 滑动窗口
		windowSize := 100
		if len(tm.events[nodeID]) > windowSize {
			// 重新计算
			score.Score = tm.calculateFromWindow(nodeID, windowSize)
		} else {
			score.Score = score.Score + event.Impact/float64(windowSize)
		}

	default:
		score.Score = score.Score + event.Impact*0.1
	}

	// 限制范围
	if score.Score < 0 {
		score.Score = 0
	}
	if score.Score > 1 {
		score.Score = 1
	}

	// 更新置信度
	score.Confidence = float64(score.TotalEvents) / 100.0
	if score.Confidence > 1 {
		score.Confidence = 1
	}

	// 更新等级
	score.Rating = tm.calculateRating(score.Score)
	score.LastUpdate = time.Now()
}

// calculateFromWindow 从窗口计算
func (tm *TrustManager) calculateFromWindow(nodeID string, windowSize int) float64 {
	events := tm.events[nodeID]
	if len(events) < windowSize {
		windowSize = len(events)
	}

	sum := 0.0
	for i := len(events) - windowSize; i < len(events); i++ {
		sum += events[i].Impact
	}

	return 0.5 + sum/float64(windowSize)
}

// calculateRating 计算等级
func (tm *TrustManager) calculateRating(score float64) TrustRating {
	if score < tm.thresholds.UntrustedThreshold {
		return RatingUntrusted
	}
	if score < tm.thresholds.LowThreshold {
		return RatingLow
	}
	if score < tm.thresholds.MediumThreshold {
		return RatingMedium
	}
	if score < tm.thresholds.HighThreshold {
		return RatingHigh
	}
	return RatingFull
}

// GetTrustScore 获取信任评分
func (tm *TrustManager) GetTrustScore(nodeID string) (*TrustScore, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	score, ok := tm.scores[nodeID]
	if !ok {
		return nil, fmt.Errorf("节点信任评分不存在: %s", nodeID)
	}
	return score, nil
}

// GetTrustEvents 获取信任事件
func (tm *TrustManager) GetTrustEvents(nodeID string, limit int) ([]*TrustEvent, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	events, ok := tm.events[nodeID]
	if !ok {
		return nil, fmt.Errorf("节点事件不存在: %s", nodeID)
	}

	if limit > 0 && len(events) > limit {
		return events[len(events)-limit:], nil
	}
	return events, nil
}

// ListNodesByTrust 按信任列节点
func (tm *TrustManager) ListNodesByTrust(rating TrustRating) []*TrustScore {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	list := make([]*TrustScore, 0)
	for _, score := range tm.scores {
		if rating == "" || score.Rating == rating {
			list = append(list, score)
		}
	}

	// 排序
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Score > list[i].Score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	return list
}

// IsTrusted 检查是否可信
func (tm *TrustManager) IsTrusted(nodeID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	score, ok := tm.scores[nodeID]
	if !ok {
		return false
	}

	return score.Score >= tm.thresholds.LowThreshold && score.Confidence >= tm.thresholds.MinConfidence
}

// DecayScores 衰减评分
func (tm *TrustManager) DecayScores() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, score := range tm.scores {
		// 时间衰减
		hoursSinceUpdate := time.Since(score.LastUpdate).Hours()
		decay := score.DecayFactor * (1.0 - hoursSinceUpdate/720.0) // 30天半衰期

		if decay < 0.5 {
			decay = 0.5
		}

		score.Score = score.Score * decay
		if score.Score < 0.3 {
			score.Score = 0.3 // 最低不低于0.3
		}

		score.Rating = tm.calculateRating(score.Score)
	}
}

// StartDecayRoutine 启动衰减例行程序
func (tm *TrustManager) StartDecayRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			tm.DecayScores()
		}
	}()
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("evt-%d", time.Now().UnixNano())
}

// TrustStats 信任统计
type TrustStats struct {
	TotalNodes       int       `json:"total_nodes"`
	AvgScore         float64   `json:"avg_score"`
	AvgConfidence    float64   `json:"avg_confidence"`
	UntrustedNodes   int       `json:"untrusted_nodes"`
	LowTrustNodes    int       `json:"low_trust_nodes"`
	MediumTrustNodes int       `json:"medium_trust_nodes"`
	HighTrustNodes   int       `json:"high_trust_nodes"`
	FullTrustNodes   int       `json:"full_trust_nodes"`
	TotalEvents      int       `json:"total_events"`
}

// GetStats 获取统计
func (tm *TrustManager) GetStats() *TrustStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := &TrustStats{
		TotalNodes: len(tm.scores),
	}

	if len(tm.scores) == 0 {
		return stats
	}

	totalScore := 0.0
	totalConfidence := 0.0

	for _, score := range tm.scores {
		totalScore += score.Score
		totalConfidence += score.Confidence
		stats.TotalEvents += score.TotalEvents

		switch score.Rating {
		case RatingUntrusted:
			stats.UntrustedNodes++
		case RatingLow:
			stats.LowTrustNodes++
		case RatingMedium:
			stats.MediumTrustNodes++
		case RatingHigh:
			stats.HighTrustNodes++
		case RatingFull:
			stats.FullTrustNodes++
		}
	}

	stats.AvgScore = totalScore / float64(len(tm.scores))
	stats.AvgConfidence = totalConfidence / float64(len(tm.scores))

	return stats
}
// Package identity - 性格进化引擎 (v2.9.0)
//
// Center 是永远在线的灵魂载体，性格进化引擎确保：
// - 性格随时间稳定收敛
// - MBTI 类型趋于稳定
// - 长期行为模式识别
// - 重大事件影响追踪
package identity

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// === 性格进化引擎 ===

// EvolutionEngine 性格进化引擎
type EvolutionEngine struct {
	mu sync.RWMutex

	// 推断器
	inferencer *Inferencer

	// 配置
	config EvolutionConfig

	// 性格历史记录：identityId -> snapshots
	history map[string]*PersonalityHistory

	// 重大事件记录
	events map[string][]MajorEvent
}

// EvolutionConfig 进化配置
type EvolutionConfig struct {
	// 稳定性阈值
	StabilityThreshold    float64       `json:"stability_threshold"`     // 稳定度阈值 (0.8)
	ConvergenceWindow     int           `json:"convergence_window"`      // 收敛窗口大小 (20)
	MBTIStabilityWindow   int           `json:"mbti_stability_window"`   // MBTI 稳定窗口 (50)

	// 时间衰减
	DecayHalfLife         time.Duration `json:"decay_half_life"`         // 衰减半衰期 (30天)
	MinObservationWeight  float64       `json:"min_observation_weight"`  // 最小观察权重 (0.1)

	// 重大事件
	MajorEventThreshold   float64       `json:"major_event_threshold"`   // 重大事件阈值 (0.3)
	EventImpactDecay      time.Duration `json:"event_impact_decay"`      // 事件影响衰减 (7天)

	// 快照
	SnapshotInterval      time.Duration `json:"snapshot_interval"`       // 快照间隔 (24小时)
	MaxSnapshots          int           `json:"max_snapshots"`           // 最大快照数 (90)
}

// DefaultEvolutionConfig 默认配置
func DefaultEvolutionConfig() EvolutionConfig {
	return EvolutionConfig{
		StabilityThreshold:    0.8,
		ConvergenceWindow:     20,
		MBTIStabilityWindow:   50,
		DecayHalfLife:         30 * 24 * time.Hour,
		MinObservationWeight:  0.1,
		MajorEventThreshold:   0.3,
		EventImpactDecay:      7 * 24 * time.Hour,
		SnapshotInterval:      24 * time.Hour,
		MaxSnapshots:          90,
	}
}

// PersonalityHistory 性格历史
type PersonalityHistory struct {
	IdentityID    string              `json:"identity_id"`
	Snapshots     []PersonalitySnapshot `json:"snapshots"`
	CurrentState  PersonalityState    `json:"current_state"`
	TrendAnalysis *TrendAnalysis      `json:"trend_analysis,omitempty"`
}

// PersonalitySnapshot 性格快照
type PersonalitySnapshot struct {
	Timestamp      time.Time  `json:"timestamp"`
	Personality    *models.Personality `json:"personality"`
	MBTIType       string     `json:"mbti_type"`
	MBTIConfidence float64    `json:"mbti_confidence"`
	StabilityScore float64    `json:"stability_score"`
	ObservedCount  int        `json:"observed_count"`
	Trigger        string     `json:"trigger"` // scheduled, event, manual
}

// PersonalityState 性格状态
type PersonalityState struct {
	// 稳定性状态
	IsStable           bool      `json:"is_stable"`
	StabilityScore     float64   `json:"stability_score"`
	StabilizedAt       *time.Time `json:"stabilized_at,omitempty"`

	// MBTI 稳定性
	MBTIStable         bool      `json:"mbti_stable"`
	StableMBTIType     string    `json:"stable_mbti_type"`
	MBTILocked         bool      `json:"mbti_locked"` // 是否已锁定（不可更改）

	// 收敛状态
	ConvergenceRate    float64   `json:"convergence_rate"` // 收敛速率
	Variance           float64   `json:"variance"`         // 方差（波动程度）

	// 最后更新
	LastUpdated        time.Time `json:"last_updated"`
	ObservationCount   int       `json:"observation_count"`
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	// 各维度趋势 (-1 到 1，负为下降，正为上升)
	OpennessTrend          float64 `json:"openness_trend"`
	ConscientiousnessTrend float64 `json:"conscientiousness_trend"`
	ExtraversionTrend      float64 `json:"extraversion_trend"`
	AgreeablenessTrend     float64 `json:"agreeableness_trend"`
	NeuroticismTrend       float64 `json:"neuroticism_trend"`

	// MBTI 变化历史
	MBTIHistory            []MBTITransition `json:"mbti_history"`

	// 预测
	PredictedMBTI          string  `json:"predicted_mbti"`
	PredictionConfidence   float64 `json:"prediction_confidence"`

	// 分析时间
	AnalyzedAt             time.Time `json:"analyzed_at"`
}

// MBTITransition MBTI 转换记录
type MBTITransition struct {
	FromType    string    `json:"from_type"`
	ToType      string    `json:"to_type"`
	Timestamp   time.Time `json:"timestamp"`
	Trigger     string    `json:"trigger"` // observation, event, convergence
	Confidence  float64   `json:"confidence"`
}

// MajorEvent 重大事件
type MajorEvent struct {
	ID          string                 `json:"id"`
	IdentityID  string                 `json:"identity_id"`
	Type        string                 `json:"type"` // life_change, trauma, achievement, relationship
	Description string                 `json:"description"`
	Impact      TraitImpact            `json:"impact"`
	ObservedAt  time.Time              `json:"observed_at"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Decayed     bool                   `json:"decayed"`
}

// NewEvolutionEngine 创建进化引擎
func NewEvolutionEngine(config EvolutionConfig) *EvolutionEngine {
	if config.StabilityThreshold == 0 {
		config = DefaultEvolutionConfig()
	}

	return &EvolutionEngine{
		inferencer: NewInferencer(),
		config:     config,
		history:    make(map[string]*PersonalityHistory),
		events:     make(map[string][]MajorEvent),
	}
}

// === 核心方法 ===

// ProcessObservations 处理行为观察并更新性格
func (e *EvolutionEngine) ProcessObservations(identityID string, observations []models.BehaviorObservation, current *models.Personality) (*models.Personality, *PersonalityState) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 获取或创建历史记录
	history := e.getOrCreateHistory(identityID)

	// 应用时间衰减
	observations = e.applyTimeDecay(observations)

	// 检测重大事件
	majorEvents := e.detectMajorEvents(identityID, observations)
	if len(majorEvents) > 0 {
		e.events[identityID] = append(e.events[identityID], majorEvents...)
	}

	// 应用重大事件影响
	current = e.applyEventImpact(current, majorEvents)

	// 使用推断器更新性格
	updated := e.inferencer.UpdatePersonalityWithConvergence(current, observations)

	// 更新状态
	state := e.updateState(identityID, updated)

	// 检查是否需要创建快照
	e.maybeCreateSnapshot(identityID, updated, "observation")

	// 更新趋势分析
	e.updateTrendAnalysis(history)

	return updated, state
}

// GetPersonalityState 获取性格状态
func (e *EvolutionEngine) GetPersonalityState(identityID string) *PersonalityState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if history, ok := e.history[identityID]; ok {
		state := history.CurrentState
		return &state
	}
	return nil
}

// GetPersonalityHistory 获取性格历史
func (e *EvolutionEngine) GetPersonalityHistory(identityID string, limit int) []PersonalitySnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if history, ok := e.history[identityID]; ok {
		if limit <= 0 || limit > len(history.Snapshots) {
			limit = len(history.Snapshots)
		}
		start := len(history.Snapshots) - limit
		if start < 0 {
			start = 0
		}
		return history.Snapshots[start:]
	}
	return nil
}

// GetTrendAnalysis 获取趋势分析
func (e *EvolutionEngine) GetTrendAnalysis(identityID string) *TrendAnalysis {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if history, ok := e.history[identityID]; ok {
		return history.TrendAnalysis
	}
	return nil
}

// === 稳定性检测 ===

// CheckStability 检查性格稳定性
func (e *EvolutionEngine) CheckStability(identityID string, personality *models.Personality) StabilityReport {
	e.mu.RLock()
	defer e.mu.RUnlock()

	report := StabilityReport{
		IdentityID: identityID,
		CheckedAt:  time.Now(),
	}

	history, ok := e.history[identityID]
	if !ok || len(history.Snapshots) < e.config.ConvergenceWindow {
		report.IsStable = false
		report.Reason = "insufficient_data"
		report.RequiredObservations = e.config.ConvergenceWindow
		return report
	}

	// 计算最近的方差
	snapshots := history.Snapshots
	if len(snapshots) > e.config.ConvergenceWindow {
		snapshots = snapshots[len(snapshots)-e.config.ConvergenceWindow:]
	}

	variance := e.calculateVariance(snapshots)
	report.Variance = variance

	// 检查是否稳定
	if variance < 0.05 && personality.StabilityScore >= e.config.StabilityThreshold {
		report.IsStable = true
		report.Reason = "converged"
		report.StabilityScore = personality.StabilityScore
	} else if personality.StabilityScore >= e.config.StabilityThreshold {
		report.IsStable = true
		report.Reason = "high_stability"
		report.StabilityScore = personality.StabilityScore
	} else {
		report.IsStable = false
		report.Reason = "still_evolving"
		report.StabilityScore = personality.StabilityScore
	}

	// MBTI 稳定性
	report.MBTIStable = e.checkMBTIStability(history)
	report.StableMBTIType = history.CurrentState.StableMBTIType

	return report
}

// StabilityReport 稳定性报告
type StabilityReport struct {
	IdentityID          string    `json:"identity_id"`
	IsStable            bool      `json:"is_stable"`
	Reason              string    `json:"reason"` // converged, high_stability, still_evolving, insufficient_data
	StabilityScore      float64   `json:"stability_score"`
	Variance            float64   `json:"variance"`
	MBTIStable          bool      `json:"mbti_stable"`
	StableMBTIType      string    `json:"stable_mbti_type"`
	RequiredObservations int      `json:"required_observations,omitempty"`
	CheckedAt           time.Time `json:"checked_at"`
}

// checkMBTIStability 检查 MBTI 稳定性
func (e *EvolutionEngine) checkMBTIStability(history *PersonalityHistory) bool {
	if len(history.Snapshots) < e.config.MBTIStabilityWindow {
		return false
	}

	// 获取最近的 MBTI 类型
	recentSnapshots := history.Snapshots[len(history.Snapshots)-e.config.MBTIStabilityWindow:]
	currentType := recentSnapshots[len(recentSnapshots)-1].MBTIType

	// 检查是否一致
	for _, s := range recentSnapshots {
		if s.MBTIType != currentType {
			return false
		}
	}

	return true
}

// === 时间衰减 ===

// applyTimeDecay 应用时间衰减
func (e *EvolutionEngine) applyTimeDecay(observations []models.BehaviorObservation) []models.BehaviorObservation {
	now := time.Now()
	halfLife := float64(e.config.DecayHalfLife)

	result := make([]models.BehaviorObservation, 0, len(observations))
	for _, obs := range observations {
		// 计算衰减权重
		age := now.Sub(obs.Timestamp).Hours()
		decayFactor := math.Pow(0.5, age/(halfLife/float64(time.Hour)))

		// 如果权重太低，跳过
		if decayFactor < e.config.MinObservationWeight {
			continue
		}

		// 应用衰减到推断值
		if obs.Inferences != nil {
			decayedInferences := make(map[string]float64)
			for k, v := range obs.Inferences {
				decayedInferences[k] = v * decayFactor
			}
			obs.Inferences = decayedInferences
		}

		result = append(result, obs)
	}

	return result
}

// === 重大事件检测 ===

// detectMajorEvents 检测重大事件
func (e *EvolutionEngine) detectMajorEvents(identityID string, observations []models.BehaviorObservation) []MajorEvent {
	var events []MajorEvent

	for _, obs := range observations {
		// 检查是否有大影响
		if obs.Inferences == nil {
			continue
		}

		maxImpact := 0.0
		for _, v := range obs.Inferences {
			if math.Abs(v) > maxImpact {
				maxImpact = math.Abs(v)
			}
		}

		// 超过阈值则认为是重大事件
		if maxImpact >= e.config.MajorEventThreshold {
			event := MajorEvent{
				ID:         fmt.Sprintf("evt_%d", time.Now().UnixNano()),
				IdentityID: identityID,
				Type:       e.classifyEventType(obs),
				ObservedAt: obs.Timestamp,
				Context:    obs.Context,
			}

			// 提取影响
			event.Impact = TraitImpact{
				Openness:          obs.Inferences["openness"],
				Conscientiousness: obs.Inferences["conscientiousness"],
				Extraversion:      obs.Inferences["extraversion"],
				Agreeableness:     obs.Inferences["agreeableness"],
				Neuroticism:       obs.Inferences["neuroticism"],
			}

			events = append(events, event)
		}
	}

	return events
}

// classifyEventType 分类事件类型
func (e *EvolutionEngine) classifyEventType(obs models.BehaviorObservation) string {
	// 根据观察类型和上下文分类
	if obs.Context == nil {
		return "unknown"
	}

	if t, ok := obs.Context["event_type"].(string); ok {
		return t
	}

	// 默认分类
	switch obs.Type {
	case "decision":
		return "life_change"
	case "interaction":
		return "relationship"
	case "activity":
		return "achievement"
	default:
		return "unknown"
	}
}

// applyEventImpact 应用事件影响
func (e *EvolutionEngine) applyEventImpact(personality *models.Personality, events []MajorEvent) *models.Personality {
	if personality == nil || len(events) == 0 {
		return personality
	}

	result := *personality

	for _, event := range events {
		// 事件影响有更大的权重
		impactWeight := 1.5 // 事件影响权重更大

		result.Openness = clamp01(result.Openness + event.Impact.Openness*impactWeight)
		result.Conscientiousness = clamp01(result.Conscientiousness + event.Impact.Conscientiousness*impactWeight)
		result.Extraversion = clamp01(result.Extraversion + event.Impact.Extraversion*impactWeight)
		result.Agreeableness = clamp01(result.Agreeableness + event.Impact.Agreeableness*impactWeight)
		result.Neuroticism = clamp01(result.Neuroticism + event.Impact.Neuroticism*impactWeight)
	}

	return &result
}

// === 内部方法 ===

func (e *EvolutionEngine) getOrCreateHistory(identityID string) *PersonalityHistory {
	if h, ok := e.history[identityID]; ok {
		return h
	}

	h := &PersonalityHistory{
		IdentityID: identityID,
		Snapshots:  make([]PersonalitySnapshot, 0),
		CurrentState: PersonalityState{
			LastUpdated: time.Now(),
		},
	}
	e.history[identityID] = h
	return h
}

func (e *EvolutionEngine) updateState(identityID string, personality *models.Personality) *PersonalityState {
	history := e.getOrCreateHistory(identityID)

	state := &history.CurrentState
	state.LastUpdated = time.Now()
	state.ObservationCount++
	state.StabilityScore = personality.StabilityScore

	// 检查稳定性
	if personality.StabilityScore >= e.config.StabilityThreshold {
		state.IsStable = true
		if state.StabilizedAt == nil {
			now := time.Now()
			state.StabilizedAt = &now
		}
	}

	// MBTI 稳定性
	if e.checkMBTIStability(history) {
		state.MBTIStable = true
		state.StableMBTIType = personality.MBTIType
	}

	return state
}

func (e *EvolutionEngine) maybeCreateSnapshot(identityID string, personality *models.Personality, trigger string) {
	history := e.getOrCreateHistory(identityID)

	// 检查是否需要创建快照
	now := time.Now()
	if len(history.Snapshots) > 0 {
		lastSnapshot := history.Snapshots[len(history.Snapshots)-1]
		if now.Sub(lastSnapshot.Timestamp) < e.config.SnapshotInterval {
			return
		}
	}

	snapshot := PersonalitySnapshot{
		Timestamp:      now,
		Personality:    personality,
		MBTIType:       personality.MBTIType,
		MBTIConfidence: personality.MBTIConfidence,
		StabilityScore: personality.StabilityScore,
		ObservedCount:  personality.ObservedCount,
		Trigger:        trigger,
	}

	history.Snapshots = append(history.Snapshots, snapshot)

	// 限制快照数量
	if len(history.Snapshots) > e.config.MaxSnapshots {
		history.Snapshots = history.Snapshots[len(history.Snapshots)-e.config.MaxSnapshots:]
	}
}

func (e *EvolutionEngine) calculateVariance(snapshots []PersonalitySnapshot) float64 {
	if len(snapshots) < 2 {
		return 1.0
	}

	// 计算各维度的方差
	calculateVarianceForTrait := func(getValue func(*models.Personality) float64) float64 {
		var sum, sumSq float64
		for _, s := range snapshots {
			v := getValue(s.Personality)
			sum += v
			sumSq += v * v
		}
		n := float64(len(snapshots))
		mean := sum / n
		variance := sumSq/n - mean*mean
		return variance
	}

	opennessVar := calculateVarianceForTrait(func(p *models.Personality) float64 { return p.Openness })
	conscientiousnessVar := calculateVarianceForTrait(func(p *models.Personality) float64 { return p.Conscientiousness })
	extraversionVar := calculateVarianceForTrait(func(p *models.Personality) float64 { return p.Extraversion })
	agreeablenessVar := calculateVarianceForTrait(func(p *models.Personality) float64 { return p.Agreeableness })
	neuroticismVar := calculateVarianceForTrait(func(p *models.Personality) float64 { return p.Neuroticism })

	// 平均方差
	return (opennessVar + conscientiousnessVar + extraversionVar + agreeablenessVar + neuroticismVar) / 5
}

func (e *EvolutionEngine) updateTrendAnalysis(history *PersonalityHistory) {
	if len(history.Snapshots) < 5 {
		return
	}

	analysis := &TrendAnalysis{
		AnalyzedAt:  time.Now(),
		MBTIHistory: make([]MBTITransition, 0),
	}

	// 计算趋势（使用线性回归简化版）
	snapshots := history.Snapshots
	n := float64(len(snapshots))

	// 计算各维度趋势
	analysis.OpennessTrend = e.calculateTrend(snapshots, func(s PersonalitySnapshot) float64 { return s.Personality.Openness })
	analysis.ConscientiousnessTrend = e.calculateTrend(snapshots, func(s PersonalitySnapshot) float64 { return s.Personality.Conscientiousness })
	analysis.ExtraversionTrend = e.calculateTrend(snapshots, func(s PersonalitySnapshot) float64 { return s.Personality.Extraversion })
	analysis.AgreeablenessTrend = e.calculateTrend(snapshots, func(s PersonalitySnapshot) float64 { return s.Personality.Agreeableness })
	analysis.NeuroticismTrend = e.calculateTrend(snapshots, func(s PersonalitySnapshot) float64 { return s.Personality.Neuroticism })

	// MBTI 历史变化
	prevType := ""
	for _, s := range snapshots {
		if prevType != "" && s.MBTIType != prevType {
			analysis.MBTIHistory = append(analysis.MBTIHistory, MBTITransition{
				FromType:   prevType,
				ToType:     s.MBTIType,
				Timestamp:  s.Timestamp,
				Confidence: s.MBTIConfidence,
			})
		}
		prevType = s.MBTIType
	}

	// 预测 MBTI（使用最近的稳定类型）
	if len(snapshots) > 0 {
		latest := snapshots[len(snapshots)-1]
		analysis.PredictedMBTI = latest.MBTIType
		analysis.PredictionConfidence = latest.MBTIConfidence
	}

	history.TrendAnalysis = analysis
}

// calculateTrend 计算趋势（简化线性回归）
func (e *EvolutionEngine) calculateTrend(snapshots []PersonalitySnapshot, getValue func(PersonalitySnapshot) float64) float64 {
	n := float64(len(snapshots))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, s := range snapshots {
		x := float64(i)
		y := getValue(s)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 线性回归斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	return slope
}

// === 序列化 ===

// ToJSON 序列化历史记录
func (h *PersonalityHistory) ToJSON() ([]byte, error) {
	return json.Marshal(h)
}

// FromJSON 反序列化历史记录
func (h *PersonalityHistory) FromJSON(data []byte) error {
	return json.Unmarshal(data, h)
}

// LockMBTI 锁定 MBTI 类型（用于固化稳定性格）
func (e *EvolutionEngine) LockMBTI(identityID string, mbtiType string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	history := e.getOrCreateHistory(identityID)
	history.CurrentState.MBTILocked = true
	history.CurrentState.StableMBTIType = mbtiType

	log.Printf("MBTI locked for %s: %s", identityID, mbtiType)
	return nil
}

// UnlockMBTI 解锁 MBTI 类型
func (e *EvolutionEngine) UnlockMBTI(identityID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if history, ok := e.history[identityID]; ok {
		history.CurrentState.MBTILocked = false
	}
}

// GetMajorEvents 获取重大事件
func (e *EvolutionEngine) GetMajorEvents(identityID string) []MajorEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.events[identityID]
}
package lifestage

import (
	"context"
	"sync"
	"time"

	"github.com/taomic2035/OFA/src/center/internal/models"
)

// LifeStageEngine 人生阶段管理引擎 (v4.4.0)
// 管理人生阶段、事件、感悟
type LifeStageEngine struct {
	mu sync.RWMutex

	// 人生阶段存储
	lifeStageSystems map[string]*models.LifeStageSystem // identityID -> LifeStageSystem
	lifeStageProfiles map[string]*models.LifeStageProfile

	// 监听器
	listeners []LifeStageListener
}

// LifeStageListener 人生阶段变化监听器
type LifeStageListener interface {
	OnLifeStageChanged(identityID string, stage *models.LifeStage)
	OnLifeEventAdded(identityID string, event models.LifeEvent)
	OnLifeLessonLearned(identityID string, lesson models.LifeLesson)
	OnStageTransition(identityID string, fromStage, toStage string)
}

// LifeStageDecisionContext 人生阶段决策上下文
type LifeStageDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 当前阶段
	CurrentStage *models.LifeStage `json:"current_stage"`

	// 阶段影响
	StageInfluence StageInfluence `json:"stage_influence"`

	// 发展状态
	DevelopmentStatus DevelopmentStatus `json:"development_status"`

	// 人生轨迹
	TrajectorySummary TrajectorySummary `json:"trajectory_summary"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// StageInfluence 阶段影响
type StageInfluence struct {
	// 阶段特征影响
	ChallengeLevel    float64 `json:"challenge_level"`    // 面临挑战程度
	OpportunityLevel  float64 `json:"opportunity_level"`  // 机遇程度
	GrowthFocus       string  `json:"growth_focus"`       // 成长焦点
	TaskPriority      string  `json:"task_priority"`      // 任务优先级

	// 发展倾向
	RiskTaking       float64 `json:"risk_taking"`        // 冒险倾向
	SocialFocus      float64 `json:"social_focus"`       // 社交焦点
	CareerFocus      float64 `json:"career_focus"`       // 事业焦点
	FamilyFocus      float64 `json:"family_focus"`       // 家庭焦点
	HealthFocus      float64 `json:"health_focus"`       // 健康焦点

	// 时间感知
	TimePerspective  string  `json:"time_perspective"`   // 时间视角 (future/present/past)
	UrgencyLevel     float64 `json:"urgency_level"`      // 紧迫感
	PatienceLevel    float64 `json:"patience_level"`     // 耐心程度
}

// DevelopmentStatus 发展状态
type DevelopmentStatus struct {
	// 整体发展
	OverallProgress   float64 `json:"overall_progress"`   // 整体进度
	SatisfactionLevel float64 `json:"satisfaction_level"` // 满意度水平
	GrowthRate        float64 `json:"growth_rate"`        // 成长速率

	// 各维度
	PhysicalDevelopment float64 `json:"physical_development"` // 身体发展
	MentalDevelopment   float64 `json:"mental_development"`   // 心理发展
	SocialDevelopment   float64 `json:"social_development"`   // 社会发展
	CareerDevelopment   float64 `json:"career_development"`   // 职业发展

	// 发展瓶颈
	Bottlenecks []string `json:"bottlenecks"` // 发展瓶颈
}

// TrajectorySummary 人生轨迹摘要
type TrajectorySummary struct {
	// 轨迹特征
	Direction      string  `json:"direction"`       // 方向 (upward/stable/downward/fluctuating)
	Resilience     float64 `json:"resilience"`      // 韧性
	Adaptability   float64 `json:"adaptability"`    // 适应力
	WisdomLevel    float64 `json:"wisdom_level"`    // 智慧积累

	// 关键里程碑数
	MilestoneCount int `json:"milestone_count"`

	// 转折点数
	TurningPointCount int `json:"turning_point_count"`

	// 人生感悟数
	LessonCount int `json:"lesson_count"`
}

// NewLifeStageEngine 创建人生阶段管理引擎
func NewLifeStageEngine() *LifeStageEngine {
	return &LifeStageEngine{
		lifeStageSystems:  make(map[string]*models.LifeStageSystem),
		lifeStageProfiles: make(map[string]*models.LifeStageProfile),
		listeners:         []LifeStageListener{},
	}
}

// === 人生阶段管理 ===

// GetLifeStageSystem 获取人生阶段系统
func (e *LifeStageEngine) GetLifeStageSystem(identityID string) *models.LifeStageSystem {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lifeStageSystems[identityID]
}

// GetOrCreateLifeStageSystem 获取或创建人生阶段系统
func (e *LifeStageEngine) GetOrCreateLifeStageSystem(identityID string) *models.LifeStageSystem {
	e.mu.Lock()
	defer e.mu.Unlock()

	system, exists := e.lifeStageSystems[identityID]
	if !exists {
		system = models.NewLifeStageSystem()
		e.lifeStageSystems[identityID] = system
		e.lifeStageProfiles[identityID] = models.NewLifeStageProfile(identityID)
	}
	return system
}

// UpdateLifeStageSystem 更新人生阶段系统
func (e *LifeStageEngine) UpdateLifeStageSystem(identityID string, system *models.LifeStageSystem) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system.Normalize()
	e.lifeStageSystems[identityID] = system

	if system.CurrentStage != nil {
		for _, listener := range e.listeners {
			listener.OnLifeStageChanged(identityID, system.CurrentStage)
		}
	}

	return nil
}

// SetCurrentStage 设置当前阶段
func (e *LifeStageEngine) SetCurrentStage(identityID string, stageName string, age int) error {
	system := e.GetOrCreateLifeStageSystem(identityID)

	// 记录旧阶段
	if system.CurrentStage != nil && system.CurrentStage.StageName != stageName {
		oldStageName := system.CurrentStage.StageName
		system.CurrentStage.EndDate = time.Now()
		system.StageHistory = append(system.StageHistory, *system.CurrentStage)

		// 通知阶段转换
		for _, listener := range e.listeners {
			listener.OnStageTransition(identityID, oldStageName, stageName)
		}
	}

	// 创建新阶段
	label, ageRange, desc, challenges, opportunities, tasks := models.GetStageDefinition(stageName)
	now := time.Now()

	newStage := &models.LifeStage{
		StageID:        generateStageID(),
		StageName:      stageName,
		StageLabel:     label,
		Description:    desc,
		StartDate:      now,
		Challenges:     challenges,
		Opportunities:  opportunities,
		Goals:          []string{},
		Tasks:          tasks,
		StageAge:       age,
		Completeness:   0.0,
		Satisfaction:   0.5,
		DevelopmentMetrics: models.DevelopmentMetrics{
			PhysicalHealth:      0.8,
			MentalHealth:        0.7,
			CognitiveGrowth:     0.6,
			EmotionalMaturity:   0.6,
			SocialDevelopment:   0.6,
			RelationshipQuality: 0.6,
			CareerProgress:      0.5,
			FinancialStability:  0.5,
			SkillDevelopment:    0.6,
			SelfAwareness:       0.6,
			PurposeClarity:      0.5,
			LifeSatisfaction:    0.6,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	system.CurrentStage = newStage
	return e.UpdateLifeStageSystem(identityID, system)
}

// UpdateDevelopmentMetrics 更新发展指标
func (e *LifeStageEngine) UpdateDevelopmentMetrics(identityID string, metrics map[string]float64) error {
	system := e.GetOrCreateLifeStageSystem(identityID)
	if system.CurrentStage == nil {
		return nil
	}

	for key, value := range metrics {
		switch key {
		case "physical_health":
			system.CurrentStage.DevelopmentMetrics.PhysicalHealth = value
		case "mental_health":
			system.CurrentStage.DevelopmentMetrics.MentalHealth = value
		case "cognitive_growth":
			system.CurrentStage.DevelopmentMetrics.CognitiveGrowth = value
		case "emotional_maturity":
			system.CurrentStage.DevelopmentMetrics.EmotionalMaturity = value
		case "social_development":
			system.CurrentStage.DevelopmentMetrics.SocialDevelopment = value
		case "relationship_quality":
			system.CurrentStage.DevelopmentMetrics.RelationshipQuality = value
		case "career_progress":
			system.CurrentStage.DevelopmentMetrics.CareerProgress = value
		case "financial_stability":
			system.CurrentStage.DevelopmentMetrics.FinancialStability = value
		case "skill_development":
			system.CurrentStage.DevelopmentMetrics.SkillDevelopment = value
		case "self_awareness":
			system.CurrentStage.DevelopmentMetrics.SelfAwareness = value
		case "purpose_clarity":
			system.CurrentStage.DevelopmentMetrics.PurposeClarity = value
		case "life_satisfaction":
			system.CurrentStage.DevelopmentMetrics.LifeSatisfaction = value
		}
	}

	return e.UpdateLifeStageSystem(identityID, system)
}

// AddStageGoal 添加阶段目标
func (e *LifeStageEngine) AddStageGoal(identityID string, goal string) error {
	system := e.GetOrCreateLifeStageSystem(identityID)
	if system.CurrentStage == nil {
		return nil
	}

	// 检查是否已存在
	for _, g := range system.CurrentStage.Goals {
		if g == goal {
			return nil
		}
	}

	system.CurrentStage.Goals = append(system.CurrentStage.Goals, goal)
	return e.UpdateLifeStageSystem(identityID, system)
}

// UpdateStageCompleteness 更新阶段完成度
func (e *LifeStageEngine) UpdateStageCompleteness(identityID string, completeness float64) error {
	system := e.GetOrCreateLifeStageSystem(identityID)
	if system.CurrentStage == nil {
		return nil
	}

	system.CurrentStage.Completeness = clamp01(completeness)
	return e.UpdateLifeStageSystem(identityID, system)
}

// UpdateStageSatisfaction 更新阶段满意度
func (e *LifeStageEngine) UpdateStageSatisfaction(identityID string, satisfaction float64) error {
	system := e.GetOrCreateLifeStageSystem(identityID)
	if system.CurrentStage == nil {
		return nil
	}

	system.CurrentStage.Satisfaction = clamp01(satisfaction)
	return e.UpdateLifeStageSystem(identityID, system)
}

// === 人生事件管理 ===

// AddLifeEvent 添加人生事件
func (e *LifeStageEngine) AddLifeEvent(identityID string, event models.LifeEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.GetOrCreateLifeStageSystem(identityID)
	event.EventID = generateEventID()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	system.LifeEvents = append(system.LifeEvents, event)

	// 通知事件添加
	for _, listener := range e.listeners {
		listener.OnLifeEventAdded(identityID, event)
	}

	// 根据事件类型更新发展指标
	e.applyEventImpact(system, event)

	return nil
}

// applyEventImpact 应用事件影响
func (e *LifeStageEngine) applyEventImpact(system *models.LifeStageSystem, event models.LifeEvent) {
	if system.CurrentStage == nil {
		return
	}

	// 根据事件影响领域更新发展指标
	for area, impact := range event.ImpactAreas {
		impactValue := impact * event.ImpactLevel * (1 + event.ImpactValence)

		switch area {
		case "physical_health":
			system.CurrentStage.DevelopmentMetrics.PhysicalHealth = clamp01(
				system.CurrentStage.DevelopmentMetrics.PhysicalHealth + impactValue*0.1)
		case "mental_health":
			system.CurrentStage.DevelopmentMetrics.MentalHealth = clamp01(
				system.CurrentStage.DevelopmentMetrics.MentalHealth + impactValue*0.1)
		case "career":
			system.CurrentStage.DevelopmentMetrics.CareerProgress = clamp01(
				system.CurrentStage.DevelopmentMetrics.CareerProgress + impactValue*0.1)
		case "relationship":
			system.CurrentStage.DevelopmentMetrics.RelationshipQuality = clamp01(
				system.CurrentStage.DevelopmentMetrics.RelationshipQuality + impactValue*0.1)
		case "financial":
			system.CurrentStage.DevelopmentMetrics.FinancialStability = clamp01(
				system.CurrentStage.DevelopmentMetrics.FinancialStability + impactValue*0.1)
		case "self_awareness":
			system.CurrentStage.DevelopmentMetrics.SelfAwareness = clamp01(
				system.CurrentStage.DevelopmentMetrics.SelfAwareness + impactValue*0.15)
		case "purpose":
			system.CurrentStage.DevelopmentMetrics.PurposeClarity = clamp01(
				system.CurrentStage.DevelopmentMetrics.PurposeClarity + impactValue*0.15)
		}
	}
}

// GetLifeEvents 获取人生事件
func (e *LifeStageEngine) GetLifeEvents(identityID string, eventType string) []models.LifeEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.lifeStageSystems[identityID]
	if system == nil {
		return []models.LifeEvent{}
	}

	if eventType == "" {
		return system.LifeEvents
	}

	var filtered []models.LifeEvent
	for _, event := range system.LifeEvents {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetSignificantEvents 获取重大事件
func (e *LifeStageEngine) GetSignificantEvents(identityID string, threshold float64) []models.LifeEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.lifeStageSystems[identityID]
	if system == nil {
		return []models.LifeEvent{}
	}

	var significant []models.LifeEvent
	for _, event := range system.LifeEvents {
		if event.ImpactLevel >= threshold {
			significant = append(significant, event)
		}
	}
	return significant
}

// === 人生感悟管理 ===

// AddLifeLesson 添加人生感悟
func (e *LifeStageEngine) AddLifeLesson(identityID string, lesson models.LifeLesson) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.GetOrCreateLifeStageSystem(identityID)
	lesson.LessonID = generateLessonID()
	lesson.LearnedAt = time.Now()

	system.LifeLessons = append(system.LifeLessons, lesson)

	// 通知感悟学习
	for _, listener := range e.listeners {
		listener.OnLifeLessonLearned(identityID, lesson)
	}

	return nil
}

// ApplyLifeLesson 应用人生感悟
func (e *LifeStageEngine) ApplyLifeLesson(identityID string, lessonID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.lifeStageSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.LifeLessons {
		if system.LifeLessons[i].LessonID == lessonID {
			system.LifeLessons[i].ApplyCount++
			system.LifeLessons[i].LastApplied = time.Now()
			break
		}
	}

	return nil
}

// GetLifeLessons 获取人生感悟
func (e *LifeStageEngine) GetLifeLessons(identityID string, category string) []models.LifeLesson {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.lifeStageSystems[identityID]
	if system == nil {
		return []models.LifeLesson{}
	}

	if category == "" {
		return system.LifeLessons
	}

	var filtered []models.LifeLesson
	for _, lesson := range system.LifeLessons {
		if lesson.Category == category {
			filtered = append(filtered, lesson)
		}
	}
	return filtered
}

// === 决策上下文 ===

// GetDecisionContext 获取人生阶段决策上下文
func (e *LifeStageEngine) GetDecisionContext(identityID string) *LifeStageDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.lifeStageSystems[identityID]
	if system == nil {
		system = models.NewLifeStageSystem()
	}

	// 计算阶段影响
	stageInfluence := e.calculateStageInfluence(system)

	// 计算发展状态
	devStatus := e.calculateDevelopmentStatus(system)

	// 计算轨迹摘要
	trajectory := e.calculateTrajectorySummary(identityID)

	return &LifeStageDecisionContext{
		IdentityID:        identityID,
		CurrentStage:      system.CurrentStage,
		StageInfluence:    stageInfluence,
		DevelopmentStatus: devStatus,
		TrajectorySummary: trajectory,
		Timestamp:         time.Now(),
	}
}

// calculateStageInfluence 计算阶段影响
func (e *LifeStageEngine) calculateStageInfluence(system *models.LifeStageSystem) StageInfluence {
	influence := StageInfluence{
		RiskTaking:    0.5,
		SocialFocus:   0.5,
		CareerFocus:   0.5,
		FamilyFocus:   0.5,
		HealthFocus:   0.5,
		UrgencyLevel:  0.5,
		PatienceLevel: 0.5,
	}

	if system.CurrentStage == nil {
		return influence
	}

	stage := system.CurrentStage

	// 根据阶段特征计算影响
	switch stage.StageName {
	case "childhood":
		influence.GrowthFocus = "基础建立"
		influence.TimePerspective = "present"
		influence.RiskTaking = 0.3
		influence.SocialFocus = 0.4
		influence.CareerFocus = 0.1
		influence.FamilyFocus = 0.9
	case "adolescence":
		influence.GrowthFocus = "身份探索"
		influence.TimePerspective = "future"
		influence.RiskTaking = 0.7
		influence.SocialFocus = 0.8
		influence.CareerFocus = 0.3
		influence.FamilyFocus = 0.4
	case "youth":
		influence.GrowthFocus = "职业奠基"
		influence.TimePerspective = "future"
		influence.RiskTaking = 0.6
		influence.SocialFocus = 0.6
		influence.CareerFocus = 0.7
		influence.FamilyFocus = 0.4
	case "early_adult":
		influence.GrowthFocus = "事业家庭"
		influence.TimePerspective = "future"
		influence.RiskTaking = 0.5
		influence.SocialFocus = 0.5
		influence.CareerFocus = 0.7
		influence.FamilyFocus = 0.6
	case "mid_adult":
		influence.GrowthFocus = "责任承担"
		influence.TimePerspective = "present"
		influence.RiskTaking = 0.4
		influence.SocialFocus = 0.6
		influence.CareerFocus = 0.6
		influence.FamilyFocus = 0.7
		influence.HealthFocus = 0.6
	case "mature":
		influence.GrowthFocus = "智慧传承"
		influence.TimePerspective = "past"
		influence.RiskTaking = 0.3
		influence.SocialFocus = 0.5
		influence.CareerFocus = 0.4
		influence.FamilyFocus = 0.6
		influence.HealthFocus = 0.8
	case "elderly":
		influence.GrowthFocus = "人生回顾"
		influence.TimePerspective = "past"
		influence.RiskTaking = 0.2
		influence.SocialFocus = 0.4
		influence.CareerFocus = 0.1
		influence.FamilyFocus = 0.5
		influence.HealthFocus = 0.9
	}

	// 根据挑战和机遇调整
	if len(stage.Challenges) > 0 {
		influence.ChallengeLevel = 0.5 + float64(len(stage.Challenges))*0.1
	}
	if len(stage.Opportunities) > 0 {
		influence.OpportunityLevel = 0.5 + float64(len(stage.Opportunities))*0.1
	}

	// 根据任务确定优先级
	if len(stage.Tasks) > 0 {
		influence.TaskPriority = stage.Tasks[0]
	}

	// 根据完成度调整紧迫感
	if stage.Completeness < 0.5 {
		influence.UrgencyLevel = 0.5 + (0.5 - stage.Completeness)
	}

	return influence
}

// calculateDevelopmentStatus 计算发展状态
func (e *LifeStageEngine) calculateDevelopmentStatus(system *models.LifeStageSystem) DevelopmentStatus {
	status := DevelopmentStatus{
		Bottlenecks: []string{},
	}

	if system.CurrentStage == nil {
		return status
	}

	metrics := system.CurrentStage.DevelopmentMetrics

	// 计算整体发展
	status.OverallProgress = metrics.CalculateOverallDevelopment()
	status.SatisfactionLevel = system.CurrentStage.Satisfaction
	status.PhysicalDevelopment = metrics.PhysicalHealth
	status.MentalDevelopment = (metrics.MentalHealth + metrics.CognitiveGrowth + metrics.EmotionalMaturity) / 3
	status.SocialDevelopment = (metrics.SocialDevelopment + metrics.RelationshipQuality) / 2
	status.CareerDevelopment = (metrics.CareerProgress + metrics.FinancialStability + metrics.SkillDevelopment) / 3

	// 识别瓶颈
	if metrics.PhysicalHealth < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "身体健康")
	}
	if metrics.MentalHealth < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "心理健康")
	}
	if metrics.CareerProgress < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "职业发展")
	}
	if metrics.RelationshipQuality < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "人际关系")
	}
	if metrics.FinancialStability < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "财务稳定")
	}
	if metrics.PurposeClarity < 0.4 {
		status.Bottlenecks = append(status.Bottlenecks, "人生目标")
	}

	// 计算成长速率（基于时间）
	duration := system.CurrentStage.GetDurationYears()
	if duration > 0 {
		status.GrowthRate = status.OverallProgress / float64(duration)
	} else {
		status.GrowthRate = 0.1 // 新阶段默认成长率
	}

	return status
}

// calculateTrajectorySummary 计算轨迹摘要
func (e *LifeStageEngine) calculateTrajectorySummary(identityID string) TrajectorySummary {
	summary := TrajectorySummary{}

	profile := e.lifeStageProfiles[identityID]
	if profile != nil {
		summary.Direction = profile.LifeTrajectory.Direction
		summary.Resilience = profile.LifeTrajectory.ResilienceScore
		summary.Adaptability = profile.LifeTrajectory.AdaptabilityScore
		summary.WisdomLevel = profile.WisdomAccumulated
		summary.MilestoneCount = len(profile.Milestones)
		summary.TurningPointCount = len(profile.TurningPoints)
	}

	system := e.lifeStageSystems[identityID]
	if system != nil {
		summary.LessonCount = len(system.LifeLessons)
	}

	return summary
}

// === 人生画像 ===

// GetLifeStageProfile 获取人生画像
func (e *LifeStageEngine) GetLifeStageProfile(identityID string) *models.LifeStageProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lifeStageProfiles[identityID]
}

// UpdateLifeStageProfile 更新人生画像
func (e *LifeStageEngine) UpdateLifeStageProfile(identityID string, profile *models.LifeStageProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.UpdatedAt = time.Now()
	e.lifeStageProfiles[identityID] = profile
	return nil
}

// AddMilestone 添加里程碑
func (e *LifeStageEngine) AddMilestone(identityID string, milestone models.LifeMilestone) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.lifeStageProfiles[identityID]
	if profile == nil {
		profile = models.NewLifeStageProfile(identityID)
		e.lifeStageProfiles[identityID] = profile
	}

	milestone.MilestoneID = generateMilestoneID()
	milestone.AchievedAt = time.Now()
	profile.Milestones = append(profile.Milestones, milestone)

	// 更新智慧积累
	profile.WisdomAccumulated = clamp01(profile.WisdomAccumulated + milestone.Significance*0.1)

	return nil
}

// AddTurningPoint 添加转折点
func (e *LifeStageEngine) AddTurningPoint(identityID string, turningPoint models.TurningPoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.lifeStageProfiles[identityID]
	if profile == nil {
		profile = models.NewLifeStageProfile(identityID)
		e.lifeStageProfiles[identityID] = profile
	}

	turningPoint.TurningPointID = generateTurningPointID()
	turningPoint.Timestamp = time.Now()
	profile.TurningPoints = append(profile.TurningPoints, turningPoint)

	// 更新轨迹方向
	e.updateTrajectoryDirection(profile)

	return nil
}

// updateTrajectoryDirection 更新轨迹方向
func (e *LifeStageEngine) updateTrajectoryDirection(profile *models.LifeStageProfile) {
	if len(profile.TurningPoints) < 3 {
		return
	}

	// 根据最近转折点判断方向
	recent := profile.TurningPoints[len(profile.TurningPoints)-3:]
	upCount := 0
	downCount := 0

	for _, tp := range recent {
		if tp.Significance > 0.7 {
			// 根据状态变化判断
			if tp.AfterState != "" && tp.BeforeState != "" {
				if tp.AfterState > tp.BeforeState { // 简单字符串比较
					upCount++
				} else {
					downCount++
				}
			}
		}
	}

	if upCount > downCount {
		profile.LifeTrajectory.Direction = "upward"
	} else if downCount > upCount {
		profile.LifeTrajectory.Direction = "downward"
	} else {
		profile.LifeTrajectory.Direction = "fluctuating"
	}
}

// === 阶段推断 ===

// InferStageFromAge 根据年龄推断阶段
func (e *LifeStageEngine) InferStageFromAge(age int) string {
	if age <= 12 {
		return "childhood"
	} else if age <= 18 {
		return "adolescence"
	} else if age <= 25 {
		return "youth"
	} else if age <= 35 {
		return "early_adult"
	} else if age <= 50 {
		return "mid_adult"
	} else if age <= 65 {
		return "mature"
	} else {
		return "elderly"
	}
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *LifeStageEngine) AddListener(listener LifeStageListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *LifeStageEngine) RemoveListener(listener LifeStageListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.listeners {
		if l == listener {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			break
		}
	}
}

// === 辅助函数 ===

func generateStageID() string {
	return "stage_" + time.Now().Format("20060102150405")
}

func generateEventID() string {
	return "event_" + time.Now().Format("20060102150405")
}

func generateLessonID() string {
	return "lesson_" + time.Now().Format("20060102150405")
}

func generateMilestoneID() string {
	return "milestone_" + time.Now().Format("20060102150405")
}

func generateTurningPointID() string {
	return "turning_" + time.Now().Format("20060102150405")
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
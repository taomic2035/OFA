package behavior

import (
	"context"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// EmotionBehaviorEngine 情绪行为联动引擎 (v4.5.0)
// 管理情绪对决策、表达、行为的影响
type EmotionBehaviorEngine struct {
	mu sync.RWMutex

	// 情绪行为系统存储
	behaviorSystems map[string]*models.EmotionBehaviorSystem // identityID -> EmotionBehaviorSystem
	behaviorProfiles map[string]*models.EmotionalBehaviorProfile

	// 情绪引擎引用（用于获取情绪状态）
	emotionStateProvider EmotionStateProvider

	// 监听器
	listeners []EmotionBehaviorListener
}

// EmotionStateProvider 情绪状态提供者接口
type EmotionStateProvider interface {
	GetCurrentEmotion(identityID string) *models.Emotion
	GetDominantEmotion(identityID string) string
	GetEmotionIntensity(identityID string) float64
}

// EmotionBehaviorListener 情绪行为监听器
type EmotionBehaviorListener interface {
	OnDecisionInfluenceChanged(identityID string, influence *models.EmotionDecisionInfluence)
	OnExpressionInfluenceChanged(identityID string, influence *models.EmotionalExpressionInfluence)
	OnBehaviorTriggered(identityID string, behavior models.EmotionTriggeredBehavior)
	OnCopingStrategyUsed(identityID string, strategy models.CopingStrategy)
}

// EmotionBehaviorDecisionContext 情绪行为决策上下文
type EmotionBehaviorDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 决策影响
	DecisionInfluence *models.EmotionDecisionInfluence `json:"decision_influence"`

	// 表达影响
	ExpressionInfluence *models.EmotionalExpressionInfluence `json:"expression_influence"`

	// 推荐行为
	RecommendedBehaviors []BehaviorRecommendation `json:"recommended_behaviors"`

	// 推荐应对策略
	RecommendedCopingStrategies []CopingRecommendation `json:"recommended_coping_strategies"`

	// 当前情绪状态
	CurrentEmotionState EmotionStateSummary `json:"current_emotion_state"`

	// 行为建议
	BehaviorGuidance BehaviorGuidance `json:"behavior_guidance"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// BehaviorRecommendation 行为推荐
type BehaviorRecommendation struct {
	BehaviorType    string  `json:"behavior_type"`
	BehaviorName    string  `json:"behavior_name"`
	Reason          string  `json:"reason"`
	Urgency         float64 `json:"urgency"`
	Appropriateness float64 `json:"appropriateness"`
}

// CopingRecommendation 应对策略推荐
type CopingRecommendation struct {
	StrategyID   string  `json:"strategy_id"`
	StrategyName string  `json:"strategy_name"`
	Reason       string  `json:"reason"`
	Effectiveness float64 `json:"effectiveness"`
	Priority     int     `json:"priority"`
}

// EmotionStateSummary 情绪状态摘要
type EmotionStateSummary struct {
	DominantEmotion string  `json:"dominant_emotion"`
	Intensity       float64 `json:"intensity"`
	Valence         float64 `json:"valence"` // 正负效价
	Arousal         float64 `json:"arousal"` // 唤醒度
}

// BehaviorGuidance 行为指导
type BehaviorGuidance struct {
	// 决策建议
	DecisionStyle     string `json:"decision_style"`      // 建议决策风格
	ShouldDelay       bool   `json:"should_delay"`        // 是否应延迟决策
	DelayReason       string `json:"delay_reason"`        // 延迟原因

	// 表达建议
	ExpressionStyle   string `json:"expression_style"`    // 表达风格
	ShouldExpress     bool   `json:"should_express"`      // 是否应表达
	ExpressionChannel string `json:"expression_channel"`  // 表达渠道

	// 社交建议
	SocialApproach    string `json:"social_approach"`     // 社交方式
	ShouldInteract    bool   `json:"should_interact"`     // 是否应互动

	// 风险提示
	RiskWarning       string `json:"risk_warning"`        // 风险提示
	ImpulseRisk       float64 `json:"impulse_risk"`       // 冲动风险

	// 自我调节建议
	RegulationSuggestion string `json:"regulation_suggestion"` // 调节建议
}

// NewEmotionBehaviorEngine 创建情绪行为联动引擎
func NewEmotionBehaviorEngine() *EmotionBehaviorEngine {
	return &EmotionBehaviorEngine{
		behaviorSystems:  make(map[string]*models.EmotionBehaviorSystem),
		behaviorProfiles: make(map[string]*models.EmotionalBehaviorProfile),
		listeners:        []EmotionBehaviorListener{},
	}
}

// SetEmotionStateProvider 设置情绪状态提供者
func (e *EmotionBehaviorEngine) SetEmotionStateProvider(provider EmotionStateProvider) {
	e.emotionStateProvider = provider
}

// === 情绪行为系统管理 ===

// GetBehaviorSystem 获取情绪行为系统
func (e *EmotionBehaviorEngine) GetBehaviorSystem(identityID string) *models.EmotionBehaviorSystem {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.behaviorSystems[identityID]
}

// GetOrCreateBehaviorSystem 获取或创建情绪行为系统
func (e *EmotionBehaviorEngine) GetOrCreateBehaviorSystem(identityID string) *models.EmotionBehaviorSystem {
	e.mu.Lock()
	defer e.mu.Unlock()

	system, exists := e.behaviorSystems[identityID]
	if !exists {
		system = models.NewEmotionBehaviorSystem()
		e.behaviorSystems[identityID] = system
		e.behaviorProfiles[identityID] = models.NewEmotionalBehaviorProfile(identityID)
	}
	return system
}

// UpdateBehaviorSystem 更新情绪行为系统
func (e *EmotionBehaviorEngine) UpdateBehaviorSystem(identityID string, system *models.EmotionBehaviorSystem) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system.Normalize()
	e.behaviorSystems[identityID] = system

	return nil
}

// === 情绪影响决策 ===

// ApplyEmotionToDecision 应用情绪到决策
func (e *EmotionBehaviorEngine) ApplyEmotionToDecision(identityID string, emotion string, intensity float64) *models.EmotionDecisionInfluence {
	system := e.GetOrCreateBehaviorSystem(identityID)
	influence := system.DecisionInfluence

	// 更新主导情绪
	influence.DominantEmotion = emotion
	influence.EmotionIntensity = intensity

	// 根据情绪类型调整决策影响
	switch emotion {
	case "joy":
		// 喜悦：增加风险承受、社交趋近、合作倾向
		influence.RiskTolerance = clamp01(influence.RiskTolerance + intensity*0.2)
		influence.SocialApproach = clamp01(influence.SocialApproach + intensity*0.2)
		influence.CooperationTendency = clamp01(influence.CooperationTendency + intensity*0.2)
		influence.TrustLevel = clamp01(influence.TrustLevel + intensity*0.15)
		influence.NoveltySeeking = clamp01(influence.NoveltySeeking + intensity*0.1)
		influence.DecisionSpeed = clamp01(influence.DecisionSpeed + intensity*0.1)

	case "anger":
		// 愤怒：增加风险承受、冲动、攻击性
		influence.RiskTolerance = clamp01(influence.RiskTolerance + intensity*0.25)
		influence.ImpulseControl = clamp01(influence.ImpulseControl - intensity*0.2)
		influence.DecisionSpeed = clamp01(influence.DecisionSpeed + intensity*0.2)
		influence.SocialAvoidance = clamp01(influence.SocialAvoidance + intensity*0.15)
		influence.TrustLevel = clamp01(influence.TrustLevel - intensity*0.15)
		influence.CooperationTendency = clamp01(influence.CooperationTendency - intensity*0.2)

	case "sadness":
		// 悲伤：降低风险承受、增加回避、降低决策速度
		influence.RiskTolerance = clamp01(influence.RiskTolerance - intensity*0.2)
		influence.RiskAversion = clamp01(influence.RiskAversion + intensity*0.2)
		influence.SocialAvoidance = clamp01(influence.SocialAvoidance + intensity*0.25)
		influence.SocialApproach = clamp01(influence.SocialApproach - intensity*0.15)
		influence.DecisionSpeed = clamp01(influence.DecisionSpeed - intensity*0.1)
		influence.DeliberationLevel = clamp01(influence.DeliberationLevel + intensity*0.1)

	case "fear":
		// 恐惧：增加风险规避、回避、决策谨慎
		influence.RiskAversion = clamp01(influence.RiskAversion + intensity*0.3)
		influence.RiskTolerance = clamp01(influence.RiskTolerance - intensity*0.25)
		influence.SocialAvoidance = clamp01(influence.SocialAvoidance + intensity*0.2)
		influence.TrustLevel = clamp01(influence.TrustLevel - intensity*0.1)
		influence.DeliberationLevel = clamp01(influence.DeliberationLevel + intensity*0.2)
		influence.DecisionSpeed = clamp01(influence.DecisionSpeed - intensity*0.15)

	case "love":
		// 爱：增加社交趋近、合作、信任
		influence.SocialApproach = clamp01(influence.SocialApproach + intensity*0.3)
		influence.CooperationTendency = clamp01(influence.CooperationTendency + intensity*0.25)
		influence.TrustLevel = clamp01(influence.TrustLevel + intensity*0.25)
		influence.WarmthLevel = influence.TrustLevel // 复用
		influence.FamiliarityPreference = clamp01(influence.FamiliarityPreference + intensity*0.1)

	case "disgust":
		// 厌恶：增加回避、降低合作
		influence.SocialAvoidance = clamp01(influence.SocialAvoidance + intensity*0.25)
		influence.CooperationTendency = clamp01(influence.CooperationTendency - intensity*0.2)
		influence.TrustLevel = clamp01(influence.TrustLevel - intensity*0.15)
		influence.NoveltySeeking = clamp01(influence.NoveltySeeking - intensity*0.1)

	case "desire":
		// 欲望：增加冲动、求新、风险承受
		influence.ImpulseControl = clamp01(influence.ImpulseControl - intensity*0.15)
		influence.NoveltySeeking = clamp01(influence.NoveltySeeking + intensity*0.2)
		influence.RiskTolerance = clamp01(influence.RiskTolerance + intensity*0.15)
		influence.DecisionSpeed = clamp01(influence.DecisionSpeed + intensity*0.15)
		influence.DelayedGratification = clamp01(influence.DelayedGratification - intensity*0.1)
	}

	// 通知变更
	for _, listener := range e.listeners {
		listener.OnDecisionInfluenceChanged(identityID, influence)
	}

	return influence
}

// GetDecisionInfluence 获取决策影响
func (e *EmotionBehaviorEngine) GetDecisionInfluence(identityID string) *models.EmotionDecisionInfluence {
	system := e.GetOrCreateBehaviorSystem(identityID)
	return system.DecisionInfluence
}

// === 情绪影响表达 ===

// ApplyEmotionToExpression 应用情绪到表达
func (e *EmotionBehaviorEngine) ApplyEmotionToExpression(identityID string, emotion string, intensity float64) *models.EmotionalExpressionInfluence {
	system := e.GetOrCreateBehaviorSystem(identityID)
	influence := system.ExpressionInfluence

	// 更新底层情绪
	influence.UnderlyingEmotion = emotion

	// 根据情绪类型调整表达影响
	switch emotion {
	case "joy":
		// 喜悦：温暖、热情、积极措辞
		influence.WarmthLevel = clamp01(influence.WarmthLevel + intensity*0.25)
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel + intensity*0.25)
		influence.WordChoice = "positive"
		influence.EmojiUsage = clamp01(influence.EmojiUsage + intensity*0.2)
		influence.EmojiType = "happy"
		influence.ExclamationUse = clamp01(influence.ExclamationUse + intensity*0.2)
		influence.ResponseSpeed = "immediate"
		influence.HumorLevel = clamp01(influence.HumorLevel + intensity*0.15)
		influence.VoiceTone = "high"
		influence.SpeechSpeed = "fast"
		influence.ExpressionTendency = "express"

	case "anger":
		// 愤怒：尖锐、直接、快速
		influence.ToneStyle = "sharp"
		influence.WarmthLevel = clamp01(influence.WarmthLevel - intensity*0.2)
		influence.FormalityLevel = clamp01(influence.FormalityLevel - intensity*0.15)
		influence.WordChoice = "negative"
		influence.ResponseSpeed = "immediate"
		influence.VoiceTone = "high"
		influence.SpeechSpeed = "fast"
		influence.VolumeLevel = "loud"
		influence.ExpressionTendency = "express"

	case "sadness":
		// 悲伤：低沉、缓慢、回避
		influence.WarmthLevel = clamp01(influence.WarmthLevel - intensity*0.15)
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel - intensity*0.25)
		influence.WordChoice = "negative"
		influence.EmojiUsage = clamp01(influence.EmojiUsage - intensity*0.2)
		influence.ResponseSpeed = "delayed"
		influence.HumorLevel = clamp01(influence.HumorLevel - intensity*0.2)
		influence.VoiceTone = "low"
		influence.SpeechSpeed = "slow"
		influence.VolumeLevel = "soft"
		influence.ExpressionTendency = "suppress"

	case "fear":
		// 恐惧：犹豫、谨慎、回避
		influence.WarmthLevel = clamp01(influence.WarmthLevel - intensity*0.1)
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel - intensity*0.15)
		influence.FormalityLevel = clamp01(influence.FormalityLevel + intensity*0.1)
		influence.WordChoice = "neutral"
		influence.ResponseSpeed = "delayed"
		influence.VoiceTone = "low"
		influence.SpeechSpeed = "slow"
		influence.PauseFrequency = clamp01(influence.PauseFrequency + intensity*0.2)
		influence.ExpressionTendency = "suppress"

	case "love":
		// 爱：温暖、亲密、表达
		influence.ToneStyle = "warm"
		influence.WarmthLevel = clamp01(influence.WarmthLevel + intensity*0.3)
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel + intensity*0.2)
		influence.WordChoice = "positive"
		influence.EmojiUsage = clamp01(influence.EmojiUsage + intensity*0.25)
		influence.ResponseSpeed = "immediate"
		influence.DetailLevel = clamp01(influence.DetailLevel + intensity*0.1)
		influence.ExpressionTendency = "express"

	case "disgust":
		// 厌恶：冷淡、回避、负面
		influence.WarmthLevel = clamp01(influence.WarmthLevel - intensity*0.25)
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel - intensity*0.2)
		influence.WordChoice = "negative"
		influence.ResponseSpeed = "delayed"
		influence.ExpressionTendency = "mask"

	case "desire":
		// 欲望：热情、积极、主动
		influence.EnthusiasmLevel = clamp01(influence.EnthusiasmLevel + intensity*0.2)
		influence.WordChoice = "positive"
		influence.ResponseSpeed = "immediate"
		influence.Proactiveness = clamp01(influence.Proactiveness + intensity*0.2)
		influence.ExpressionTendency = "express"
	}

	// 通知变更
	for _, listener := range e.listeners {
		listener.OnExpressionInfluenceChanged(identityID, influence)
	}

	return influence
}

// GetExpressionInfluence 获取表达影响
func (e *EmotionBehaviorEngine) GetExpressionInfluence(identityID string) *models.EmotionalExpressionInfluence {
	system := e.GetOrCreateBehaviorSystem(identityID)
	return system.ExpressionInfluence
}

// === 行为触发 ===

// TriggerBehavior 触发行为
func (e *EmotionBehaviorEngine) TriggerBehavior(identityID string, emotion string, intensity float64, context string) *models.EmotionTriggeredBehavior {
	// 根据情绪和强度生成行为
	behavior := e.generateBehaviorFromEmotion(emotion, intensity, context)

	e.mu.Lock()
	system := e.GetOrCreateBehaviorSystem(identityID)
	system.BehaviorTriggers = append(system.BehaviorTriggers, *behavior)
	// 保持最近50条触发记录
	if len(system.BehaviorTriggers) > 50 {
		system.BehaviorTriggers = system.BehaviorTriggers[len(system.BehaviorTriggers)-50:]
	}
	e.mu.Unlock()

	// 通知行为触发
	for _, listener := range e.listeners {
		listener.OnBehaviorTriggered(identityID, *behavior)
	}

	// 更新画像
	e.updateBehaviorProfile(identityID, emotion, behavior)

	return behavior
}

// generateBehaviorFromEmotion 从情绪生成行为
func (e *EmotionBehaviorEngine) generateBehaviorFromEmotion(emotion string, intensity float64, context string) *models.EmotionTriggeredBehavior {
	now := time.Now()
	behavior := &models.EmotionTriggeredBehavior{
		BehaviorID:         generateBehaviorID(),
		TriggerEmotion:     emotion,
		IntensityThreshold: intensity,
		Triggers:           []string{},
		ContextFactors:     []string{context},
		TriggeredAt:        now,
		Duration:           60, // 默认60秒
	}

	// 根据情绪设置行为
	switch emotion {
	case "joy":
		behavior.BehaviorType = "communication"
		behavior.BehaviorName = "分享喜悦"
		behavior.BehaviorDescription = "主动分享好消息，与他人互动"
		behavior.ActionTendency = "approach"
		behavior.UrgencyLevel = 0.4
		behavior.Automaticity = 0.6
		behavior.ImmediateEffect = "提升他人情绪"
		behavior.LongTermEffect = "增强社会联结"

	case "anger":
		behavior.BehaviorType = "action"
		behavior.BehaviorName = "表达不满"
		behavior.BehaviorDescription = "直接表达愤怒或采取行动"
		behavior.ActionTendency = "fight"
		behavior.UrgencyLevel = 0.8
		behavior.Automaticity = 0.7
		behavior.ImmediateEffect = "释放情绪"
		behavior.LongTermEffect = "可能损害关系"
		behavior.Triggers = append(behavior.Triggers, "不公正待遇", "期望落空")

	case "sadness":
		behavior.BehaviorType = "withdrawal"
		behavior.BehaviorName = "寻求安慰或独处"
		behavior.BehaviorDescription = "退缩、寻求支持或独自消化"
		behavior.ActionTendency = "withdraw"
		behavior.UrgencyLevel = 0.3
		behavior.Automaticity = 0.5
		behavior.ImmediateEffect = "获得情感支持"
		behavior.LongTermEffect = "促进情感处理"
		behavior.Triggers = append(behavior.Triggers, "失去", "失望")

	case "fear":
		behavior.BehaviorType = "avoidance"
		behavior.BehaviorName = "规避风险"
		behavior.BehaviorDescription = "避免潜在威胁，寻求安全"
		behavior.ActionTendency = "flight"
		behavior.UrgencyLevel = 0.7
		behavior.Automaticity = 0.8
		behavior.ImmediateEffect = "获得安全感"
		behavior.LongTermEffect = "可能限制成长机会"
		behavior.Triggers = append(behavior.Triggers, "威胁", "不确定性")

	case "love":
		behavior.BehaviorType = "approach"
		behavior.BehaviorName = "表达关爱"
		behavior.BehaviorDescription = "主动表达爱意，增加亲密互动"
		behavior.ActionTendency = "approach"
		behavior.UrgencyLevel = 0.5
		behavior.Automaticity = 0.5
		behavior.ImmediateEffect = "增强亲密感"
		behavior.LongTermEffect = "加深关系"
		behavior.Triggers = append(behavior.Triggers, "亲密关系", "陪伴")

	case "disgust":
		behavior.BehaviorType = "avoidance"
		behavior.BehaviorName = "回避厌恶对象"
		behavior.BehaviorDescription = "远离厌恶的事物或人"
		behavior.ActionTendency = "avoid"
		behavior.UrgencyLevel = 0.6
		behavior.Automaticity = 0.7
		behavior.ImmediateEffect = "减少不适"
		behavior.LongTermEffect = "可能限制社交圈"
		behavior.Triggers = append(behavior.Triggers, "不适环境", "价值观冲突")

	case "desire":
		behavior.BehaviorType = "seeking"
		behavior.BehaviorName = "追求目标"
		behavior.BehaviorDescription = "积极追求想要的事物"
		behavior.ActionTendency = "approach"
		behavior.UrgencyLevel = 0.7
		behavior.Automaticity = 0.6
		behavior.ImmediateEffect = "获得满足感"
		behavior.LongTermEffect = "可能过度追求"
		behavior.Triggers = append(behavior.Triggers, "诱惑", "机会")
	}

	// 根据强度调整紧迫程度
	behavior.UrgencyLevel = clamp01(behavior.UrgencyLevel * intensity)

	return behavior
}

// === 应对策略 ===

// AddCopingStrategy 添加应对策略
func (e *EmotionBehaviorEngine) AddCopingStrategy(identityID string, strategy models.CopingStrategy) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.GetOrCreateBehaviorSystem(identityID)
	strategy.StrategyID = generateStrategyID()
	strategy.CreatedAt = time.Now()

	system.CopingStrategies = append(system.CopingStrategies, strategy)

	return nil
}

// UseCopingStrategy 使用应对策略
func (e *EmotionBehaviorEngine) UseCopingStrategy(identityID string, strategyID string, success bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.behaviorSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.CopingStrategies {
		if system.CopingStrategies[i].StrategyID == strategyID {
			system.CopingStrategies[i].UseCount++
			system.CopingStrategies[i].LastUsed = time.Now()

			// 更新成功率
			if success {
				currentRate := system.CopingStrategies[i].SuccessRate
				count := float64(system.CopingStrategies[i].UseCount)
				system.CopingStrategies[i].SuccessRate = (currentRate*(count-1) + 1) / count
			}

			// 通知策略使用
			for _, listener := range e.listeners {
				listener.OnCopingStrategyUsed(identityID, system.CopingStrategies[i])
			}
			break
		}
	}

	return nil
}

// GetRecommendedCopingStrategies 获取推荐应对策略
func (e *EmotionBehaviorEngine) GetRecommendedCopingStrategies(identityID string, emotion string) []models.CopingStrategy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.behaviorSystems[identityID]
	if system == nil {
		return []models.CopingStrategy{}
	}

	var recommended []models.CopingStrategy
	for _, strategy := range system.CopingStrategies {
		for _, targetEmotion := range strategy.TargetEmotions {
			if targetEmotion == emotion && strategy.Effectiveness > 0.5 {
				recommended = append(recommended, strategy)
			}
		}
	}

	// 按有效性排序
	sortStrategiesByEffectiveness(recommended)

	// 返回前5个
	if len(recommended) > 5 {
		return recommended[:5]
	}
	return recommended
}

// === 决策上下文 ===

// GetDecisionContext 获取情绪行为决策上下文
func (e *EmotionBehaviorEngine) GetDecisionContext(identityID string) *EmotionBehaviorDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.behaviorSystems[identityID]
	if system == nil {
		system = models.NewEmotionBehaviorSystem()
	}

	// 获取当前情绪状态
	emotionState := e.getEmotionStateSummary(identityID)

	// 更新情绪影响
	if emotionState.DominantEmotion != "" {
		e.ApplyEmotionToDecision(identityID, emotionState.DominantEmotion, emotionState.Intensity)
		e.ApplyEmotionToExpression(identityID, emotionState.DominantEmotion, emotionState.Intensity)
	}

	// 生成行为建议
	behaviorGuidance := e.generateBehaviorGuidance(system, emotionState)

	// 生成推荐行为
	recommendedBehaviors := e.generateRecommendedBehaviors(emotionState)

	// 生成推荐应对策略
	recommendedCoping := e.generateRecommendedCoping(identityID, emotionState)

	return &EmotionBehaviorDecisionContext{
		IdentityID:                  identityID,
		DecisionInfluence:           system.DecisionInfluence,
		ExpressionInfluence:         system.ExpressionInfluence,
		RecommendedBehaviors:        recommendedBehaviors,
		RecommendedCopingStrategies: recommendedCoping,
		CurrentEmotionState:         emotionState,
		BehaviorGuidance:            behaviorGuidance,
		Timestamp:                   time.Now(),
	}
}

// getEmotionStateSummary 获取情绪状态摘要
func (e *EmotionBehaviorEngine) getEmotionStateSummary(identityID string) EmotionStateSummary {
	summary := EmotionStateSummary{
		DominantEmotion: "neutral",
		Intensity:       0.3,
		Valence:         0,
		Arousal:         0.3,
	}

	if e.emotionStateProvider != nil {
		summary.DominantEmotion = e.emotionStateProvider.GetDominantEmotion(identityID)
		summary.Intensity = e.emotionStateProvider.GetEmotionIntensity(identityID)
	}

	// 根据情绪类型设置效价和唤醒度
	switch summary.DominantEmotion {
	case "joy", "love":
		summary.Valence = 0.7
		summary.Arousal = 0.6
	case "anger":
		summary.Valence = -0.6
		summary.Arousal = 0.8
	case "sadness":
		summary.Valence = -0.5
		summary.Arousal = 0.3
	case "fear":
		summary.Valence = -0.7
		summary.Arousal = 0.7
	case "disgust":
		summary.Valence = -0.4
		summary.Arousal = 0.4
	case "desire":
		summary.Valence = 0.4
		summary.Arousal = 0.6
	}

	return summary
}

// generateBehaviorGuidance 生成行为指导
func (e *EmotionBehaviorEngine) generateBehaviorGuidance(system *models.EmotionBehaviorSystem, emotionState EmotionStateSummary) BehaviorGuidance {
	guidance := BehaviorGuidance{}

	// 决策风格建议
	guidance.DecisionStyle = system.DecisionInfluence.GetDecisionStyle()

	// 是否应延迟决策
	if emotionState.Intensity > 0.7 && (emotionState.DominantEmotion == "anger" || emotionState.DominantEmotion == "fear") {
		guidance.ShouldDelay = true
		guidance.DelayReason = "情绪强度较高，建议冷静后再做决策"
		guidance.ImpulseRisk = emotionState.Intensity
	}

	// 表达建议
	guidance.ExpressionStyle = system.ExpressionInfluence.GetCommunicationStyle()
	guidance.ShouldExpress = system.ExpressionInfluence.ExpressionTendency == "express"

	// 社交建议
	guidance.SocialApproach = "moderate"
	guidance.ShouldInteract = system.DecisionInfluence.SocialApproach > system.DecisionInfluence.SocialAvoidance

	// 风险提示
	if system.DecisionInfluence.IsImpulsive() {
		guidance.RiskWarning = "当前冲动风险较高，建议深思熟虑后再行动"
		guidance.ImpulseRisk = 1 - system.DecisionInfluence.ImpulseControl
	}

	// 调节建议
	guidance.RegulationSuggestion = e.getRegulationSuggestion(emotionState.DominantEmotion)

	return guidance
}

// getRegulationSuggestion 获取调节建议
func (e *EmotionBehaviorEngine) getRegulationSuggestion(emotion string) string {
	switch emotion {
	case "anger":
		return "深呼吸、暂时离开情境、换位思考"
	case "fear":
		return "分析风险来源、寻求支持、逐步面对"
	case "sadness":
		return "允许自己感受、寻求支持、适度运动"
	case "joy":
		return "享受当下、分享快乐、保持感恩"
	case "disgust":
		return "理解厌恶来源、保持距离、寻求理解"
	case "desire":
		return "评估后果、延迟满足、寻找替代"
	default:
		return "保持觉察、接纳情绪"
	}
}

// generateRecommendedBehaviors 生成推荐行为
func (e *EmotionBehaviorEngine) generateRecommendedBehaviors(emotionState EmotionStateSummary) []BehaviorRecommendation {
	var recommendations []BehaviorRecommendation

	switch emotionState.DominantEmotion {
	case "joy":
		recommendations = append(recommendations, BehaviorRecommendation{
			BehaviorType:    "social",
			BehaviorName:    "分享好消息",
			Reason:          "喜悦时分享能增强幸福感",
			Urgency:         0.4,
			Appropriateness: 0.9,
		})
	case "anger":
		recommendations = append(recommendations, BehaviorRecommendation{
			BehaviorType:    "regulation",
			BehaviorName:    "冷静下来",
			Reason:          "愤怒时先冷静再做决定",
			Urgency:         0.8,
			Appropriateness: 0.95,
		})
	case "sadness":
		recommendations = append(recommendations, BehaviorRecommendation{
			BehaviorType:    "support",
			BehaviorName:    "寻求支持",
			Reason:          "悲伤时寻求他人支持有助于恢复",
			Urgency:         0.5,
			Appropriateness: 0.85,
		})
	case "fear":
		recommendations = append(recommendations, BehaviorRecommendation{
			BehaviorType:    "safety",
			BehaviorName:    "评估安全",
			Reason:          "恐惧提示潜在风险，需要评估",
			Urgency:         0.7,
			Appropriateness: 0.9,
		})
	}

	return recommendations
}

// generateRecommendedCoping 生成推荐应对策略
func (e *EmotionBehaviorEngine) generateRecommendedCoping(identityID string, emotionState EmotionStateSummary) []CopingRecommendation {
	var recommendations []CopingRecommendation

	// 默认应对策略
	defaultStrategies := map[string][]CopingRecommendation{
		"anger": {
			{StrategyID: "deep_breathing", StrategyName: "深呼吸放松", Reason: "快速平复情绪", Effectiveness: 0.8, Priority: 1},
			{StrategyID: "timeout", StrategyName: "暂时离开", Reason: "避免冲动行为", Effectiveness: 0.85, Priority: 2},
		},
		"fear": {
			{StrategyID: "risk_assessment", StrategyName: "风险评估", Reason: "理性分析威胁", Effectiveness: 0.75, Priority: 1},
			{StrategyID: "seek_support", StrategyName: "寻求支持", Reason: "获得安全感", Effectiveness: 0.8, Priority: 2},
		},
		"sadness": {
			{StrategyID: "self_compassion", StrategyName: "自我关怀", Reason: "接纳悲伤情绪", Effectiveness: 0.85, Priority: 1},
			{StrategyID: "social_connection", StrategyName: "社会联结", Reason: "获得情感支持", Effectiveness: 0.8, Priority: 2},
		},
		"joy": {
			{StrategyID: "savor", StrategyName: "品味当下", Reason: "延长积极体验", Effectiveness: 0.9, Priority: 1},
			{StrategyID: "gratitude", StrategyName: "感恩练习", Reason: "增强幸福感", Effectiveness: 0.85, Priority: 2},
		},
	}

	if strategies, ok := defaultStrategies[emotionState.DominantEmotion]; ok {
		recommendations = strategies
	}

	return recommendations
}

// updateBehaviorProfile 更新行为画像
func (e *EmotionBehaviorEngine) updateBehaviorProfile(identityID string, emotion string, behavior *models.EmotionTriggeredBehavior) {
	profile := e.behaviorProfiles[identityID]
	if profile == nil {
		profile = models.NewEmotionalBehaviorProfile(identityID)
		e.behaviorProfiles[identityID] = profile
	}

	// 记录触发历史
	record := models.BehaviorTriggerRecord{
		RecordID:         generateRecordID(),
		Timestamp:        time.Now(),
		TriggerEmotion:   emotion,
		EmotionIntensity: behavior.IntensityThreshold,
		TriggeredBehavior: behavior.BehaviorName,
		BehaviorType:     behavior.BehaviorType,
	}

	profile.TriggerHistory = append(profile.TriggerHistory, record)
	if len(profile.TriggerHistory) > 100 {
		profile.TriggerHistory = profile.TriggerHistory[len(profile.TriggerHistory)-100:]
	}

	// 更新冲动行为率
	if behavior.ActionTendency == "fight" || behavior.ActionTendency == "flight" {
		if behavior.Automaticity > 0.6 {
			profile.ImpulsiveActionRate = clamp01(profile.ImpulsiveActionRate + 0.05)
		}
	}
}

// GetBehaviorProfile 获取行为画像
func (e *EmotionBehaviorEngine) GetBehaviorProfile(identityID string) *models.EmotionalBehaviorProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.behaviorProfiles[identityID]
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *EmotionBehaviorEngine) AddListener(listener EmotionBehaviorListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *EmotionBehaviorEngine) RemoveListener(listener EmotionBehaviorListener) {
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

func generateBehaviorID() string {
	return "behavior_" + time.Now().Format("20060102150405")
}

func generateStrategyID() string {
	return "strategy_" + time.Now().Format("20060102150405")
}

func generateRecordID() string {
	return "record_" + time.Now().Format("20060102150405")
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

func sortStrategiesByEffectiveness(strategies []models.CopingStrategy) {
	// 简单冒泡排序
	for i := 0; i < len(strategies)-1; i++ {
		for j := i + 1; j < len(strategies); j++ {
			if strategies[j].Effectiveness > strategies[i].Effectiveness {
				strategies[i], strategies[j] = strategies[j], strategies[i]
			}
		}
	}
}
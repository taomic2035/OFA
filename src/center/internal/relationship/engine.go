package relationship

import (
	"context"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
)

// RelationshipEngine 人际关系管理引擎 (v4.6.0)
// 管理人际关系网络和社交决策
type RelationshipEngine struct {
	mu sync.RWMutex

	// 关系系统存储
	relationshipSystems map[string]*models.RelationshipSystem // identityID -> RelationshipSystem
	relationshipProfiles map[string]*models.RelationshipProfile

	// 监听器
	listeners []RelationshipListener
}

// RelationshipListener 关系变化监听器
type RelationshipListener interface {
	OnRelationshipAdded(identityID string, relationship models.Relationship)
	OnRelationshipUpdated(identityID string, relationship models.Relationship)
	OnRelationshipEnded(identityID string, relationshipID string)
	OnInteractionRecorded(identityID string, interaction models.RelationshipInteraction)
	OnConflictOccurred(identityID string, relationshipID string, conflictLevel float64)
}

// RelationshipDecisionContext 人际关系决策上下文
type RelationshipDecisionContext struct {
	IdentityID string `json:"identity_id"`

	// 关系网络摘要
	NetworkSummary NetworkSummary `json:"network_summary"`

	// 关系倾向
	RelationshipOrientation models.RelationshipOrientation `json:"relationship_orientation"`

	// 社交风格
	SocialStyle models.SocialStyleProfile `json:"social_style"`

	// 依恋风格
	AttachmentStyle models.AttachmentStyleProfile `json:"attachment_style"`

	// 关系决策影响
	DecisionInfluence RelationshipDecisionInfluence `json:"decision_influence"`

	// 社交建议
	SocialGuidance SocialGuidance `json:"social_guidance"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// NetworkSummary 关系网络摘要
type NetworkSummary struct {
	TotalContacts   int     `json:"total_contacts"`
	CloseContacts   int     `json:"close_contacts"`
	SupportContacts int     `json:"support_contacts"`
	SocialCapital   float64 `json:"social_capital"`
	StrongTies      int     `json:"strong_ties"`
	WeakTies        int     `json:"weak_ties"`
	NetworkHealth   float64 `json:"network_health"`
}

// RelationshipDecisionInfluence 关系决策影响
type RelationshipDecisionInfluence struct {
	// 社交决策
	SocialApproach       float64 `json:"social_approach"`        // 社交趋近 (0-1)
	TrustInOthers        float64 `json:"trust_in_others"`        // 对他人信任 (0-1)
	CooperationReadiness float64 `json:"cooperation_readiness"` // 合作准备度 (0-1)
	VulnerabilityComfort float64 `json:"vulnerability_comfort"` // 脆弱展示舒适度 (0-1)

	// 关系投入
	CommitmentWillingness float64 `json:"commitment_willingness"` // 承诺意愿 (0-1)
	SupportWillingness    float64 `json:"support_willingness"`    // 支持意愿 (0-1)
	IntimacySeeking       float64 `json:"intimacy_seeking"`       // 亲密寻求 (0-1)

	// 冲突处理
	ConflictStyle         string  `json:"conflict_style"`         // 冲突风格
	ConfrontationReadiness float64 `json:"confrontation_readiness"` // 对抗准备度 (0-1)
	ForgivenessReadiness  float64 `json:"forgiveness_readiness"`  // 宽恕准备度 (0-1)

	// 依恋影响
	SecurityLevel        float64 `json:"security_level"`         // 安全感水平 (0-1)
	SeparationAnxiety    float64 `json:"separation_anxiety"`     // 分离焦虑 (0-1)
	ReassuranceNeed      float64 `json:"reassurance_need"`       // 再确认需求 (0-1)
}

// SocialGuidance 社交建议
type SocialGuidance struct {
	// 社交活动建议
	ShouldSocialize     bool    `json:"should_socialize"`      // 是否应社交
	SocialEnergyLevel   float64 `json:"social_energy_level"`   // 社交能量水平
	PreferredGroupSize  string  `json:"preferred_group_size"`  // 偏好群体规模 (one_on_one/small/medium/large)
	PreferredDepth      string  `json:"preferred_depth"`       // 偏好深度 (casual/deep/mixed)

	// 关系维护建议
	MaintenanceNeeded   []string `json:"maintenance_needed,omitempty"` // 需维护的关系
	OutreachSuggestions []string `json:"outreach_suggestions,omitempty"` // 主动联系建议

	// 冲突风险提示
	ConflictRisk       float64 `json:"conflict_risk"`        // 冲突风险 (0-1)
	TensionRelationships []string `json:"tension_relationships,omitempty"` // 紧张关系

	// 边界建议
	BoundaryReminders  []string `json:"boundary_reminders,omitempty"` // 边界提醒
	SelfCareNeeded     bool     `json:"self_care_needed"`              // 需要自我关怀

	// 发展建议
	GrowthOpportunities []string `json:"growth_opportunities,omitempty"` // 关系成长机会
	NewConnectionSuggestions []string `json:"new_connection_suggestions,omitempty"` // 新连接建议
}

// NewRelationshipEngine 创建人际关系管理引擎
func NewRelationshipEngine() *RelationshipEngine {
	return &RelationshipEngine{
		relationshipSystems:  make(map[string]*models.RelationshipSystem),
		relationshipProfiles: make(map[string]*models.RelationshipProfile),
		listeners:            []RelationshipListener{},
	}
}

// === 关系系统管理 ===

// GetRelationshipSystem 获取关系系统
func (e *RelationshipEngine) GetRelationshipSystem(identityID string) *models.RelationshipSystem {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.relationshipSystems[identityID]
}

// GetOrCreateRelationshipSystem 获取或创建关系系统
func (e *RelationshipEngine) GetOrCreateRelationshipSystem(identityID string) *models.RelationshipSystem {
	e.mu.Lock()
	defer e.mu.Unlock()

	system, exists := e.relationshipSystems[identityID]
	if !exists {
		system = models.NewRelationshipSystem()
		e.relationshipSystems[identityID] = system
		e.relationshipProfiles[identityID] = models.NewRelationshipProfile(identityID)
	}
	return system
}

// UpdateRelationshipSystem 更新关系系统
func (e *RelationshipEngine) UpdateRelationshipSystem(identityID string, system *models.RelationshipSystem) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system.Normalize()
	e.relationshipSystems[identityID] = system

	return nil
}

// === 关系管理 ===

// AddRelationship 添加关系
func (e *RelationshipEngine) AddRelationship(identityID string, relationship models.Relationship) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.GetOrCreateRelationshipSystem(identityID)
	relationship.RelationshipID = generateRelationshipID()
	relationship.CreatedAt = time.Now()
	relationship.UpdatedAt = time.Now()

	// 设置初始阶段
	if relationship.Stage == "" {
		relationship.Stage = "acquaintance"
	}
	if relationship.Status == "" {
		relationship.Status = "active"
	}

	system.Relationships = append(system.Relationships, relationship)

	// 更新社交网络
	e.updateSocialNetwork(system)

	// 通知关系添加
	for _, listener := range e.listeners {
		listener.OnRelationshipAdded(identityID, relationship)
	}

	return nil
}

// UpdateRelationship 更新关系
func (e *RelationshipEngine) UpdateRelationship(identityID string, relationship models.Relationship) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.Relationships {
		if system.Relationships[i].RelationshipID == relationship.RelationshipID {
			relationship.UpdatedAt = time.Now()
			relationship.normalize()
			system.Relationships[i] = relationship

			// 更新社交网络
			e.updateSocialNetwork(system)

			// 通知关系更新
			for _, listener := range e.listeners {
				listener.OnRelationshipUpdated(identityID, relationship)
			}
			break
		}
	}

	return nil
}

// RemoveRelationship 移除关系
func (e *RelationshipEngine) RemoveRelationship(identityID string, relationshipID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i, r := range system.Relationships {
		if r.RelationshipID == relationshipID {
			system.Relationships = append(system.Relationships[:i], system.Relationships[i+1:]...)

			// 更新社交网络
			e.updateSocialNetwork(system)

			// 通知关系结束
			for _, listener := range e.listeners {
				listener.OnRelationshipEnded(identityID, relationshipID)
			}
			break
		}
	}

	return nil
}

// GetRelationship 获取特定关系
func (e *RelationshipEngine) GetRelationship(identityID string, relationshipID string) *models.Relationship {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.Relationships {
		if system.Relationships[i].RelationshipID == relationshipID {
			return &system.Relationships[i]
		}
	}
	return nil
}

// GetRelationshipByPersonID 通过人员ID获取关系
func (e *RelationshipEngine) GetRelationshipByPersonID(identityID string, personID string) *models.Relationship {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.Relationships {
		if system.Relationships[i].PersonID == personID {
			return &system.Relationships[i]
		}
	}
	return nil
}

// GetRelationshipsByType 按类型获取关系
func (e *RelationshipEngine) GetRelationshipsByType(identityID string, relType string) []models.Relationship {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return []models.Relationship{}
	}

	var filtered []models.Relationship
	for _, r := range system.Relationships {
		if r.RelationshipType == relType {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// GetCloseRelationships 获取亲密关系
func (e *RelationshipEngine) GetCloseRelationships(identityID string) []models.Relationship {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return []models.Relationship{}
	}

	var close []models.Relationship
	for _, r := range system.Relationships {
		if r.IsClose() {
			close = append(close, r)
		}
	}
	return close
}

// === 互动记录 ===

// RecordInteraction 记录互动
func (e *RelationshipEngine) RecordInteraction(identityID string, interaction models.RelationshipInteraction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	interaction.InteractionID = generateInteractionID()
	interaction.Timestamp = time.Now()

	// 更新关系状态
	for i := range system.Relationships {
		if system.Relationships[i].RelationshipID == interaction.RelationshipID {
			r := &system.Relationships[i]

			// 更新互动统计
			r.TotalInteractions++
			r.LastInteraction = interaction.Timestamp

			if interaction.Satisfaction > 0.6 {
				r.PositiveInteractions++
			} else if interaction.Satisfaction < 0.4 {
				r.NegativeInteractions++
			}

			// 更新关系属性
			r.Intimacy = clamp01(r.Intimacy + interaction.IntimacyChange*0.1)
			r.Trust = clamp01(r.Trust + interaction.TrustChange*0.1)

			// 更新阶段
			e.updateRelationshipStage(r)

			// 检查冲突
			if r.ConflictLevel > 0.5 {
				for _, listener := range e.listeners {
					listener.OnConflictOccurred(identityID, r.RelationshipID, r.ConflictLevel)
				}
			}

			break
		}
	}

	// 更新社交网络
	e.updateSocialNetwork(system)

	// 通知互动记录
	for _, listener := range e.listeners {
		listener.OnInteractionRecorded(identityID, interaction)
	}

	return nil
}

// updateRelationshipStage 更新关系阶段
func (e *RelationshipEngine) updateRelationshipStage(r *models.Relationship) {
	health := r.GetRelationshipHealth()

	// 根据健康度更新阶段
	if health > 0.8 {
		r.Stage = "intimate"
	} else if health > 0.6 {
		r.Stage = "close"
	} else if health > 0.4 {
		r.Stage = "casual"
	} else {
		r.Stage = "acquaintance"
	}

	// 更新趋势
	if r.PositiveInteractions > r.NegativeInteractions*2 {
		r.Trend = "improving"
	} else if r.NegativeInteractions > r.PositiveInteractions*2 {
		r.Trend = "declining"
	} else if r.ConflictLevel > 0.5 {
		r.Trend = "fluctuating"
	} else {
		r.Trend = "stable"
	}
}

// updateSocialNetwork 更新社交网络
func (e *RelationshipEngine) updateSocialNetwork(system *models.RelationshipSystem) {
	network := system.SocialNetwork

	totalContacts := len(system.Relationships)
	closeContacts := 0
	supportContacts := 0
	strongTies := 0
	weakTies := 0

	for _, r := range system.Relationships {
		if r.IsClose() {
			closeContacts++
			strongTies++
		} else {
			weakTies++
		}

		if r.IsSupportive() {
			supportContacts++
		}
	}

	network.TotalContacts = totalContacts
	network.CloseContacts = closeContacts
	network.SupportContacts = supportContacts
	network.StrongTies = strongTies
	network.WeakTies = weakTies

	// 计算社交资本
	if totalContacts > 0 {
		network.SocialCapital = (float64(closeContacts)*0.5 + float64(supportContacts)*0.3 + float64(weakTies)*0.1) / float64(totalContacts)
	}

	// 计算需求满足度
	var totalIntimacy, totalTrust, totalSupport float64
	for _, r := range system.Relationships {
		totalIntimacy += r.Intimacy
		totalTrust += r.Trust
		totalSupport += (r.SupportGiven + r.SupportReceived) / 2
	}
	if totalContacts > 0 {
		network.BelongingNeed = totalTrust / float64(totalContacts)
		network.IntimacyNeed = totalIntimacy / float64(totalContacts)
		network.SupportNeed = totalSupport / float64(totalContacts)
	}
}

// === 社交圈管理 ===

// AddToCircle 添加到社交圈
func (e *RelationshipEngine) AddToCircle(identityID string, circleID string, personID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.SocialNetwork.Circles {
		if system.SocialNetwork.Circles[i].CircleID == circleID {
			// 检查是否已存在
			for _, m := range system.SocialNetwork.Circles[i].Members {
				if m == personID {
					return nil
				}
			}
			system.SocialNetwork.Circles[i].Members = append(system.SocialNetwork.Circles[i].Members, personID)
			break
		}
	}

	return nil
}

// CreateCircle 创建社交圈
func (e *RelationshipEngine) CreateCircle(identityID string, circle models.SocialCircle) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.GetOrCreateRelationshipSystem(identityID)
	circle.CircleID = generateCircleID()
	system.SocialNetwork.Circles = append(system.SocialNetwork.Circles, circle)

	return nil
}

// === 关系画像 ===

// GetRelationshipProfile 获取关系画像
func (e *RelationshipEngine) GetRelationshipProfile(identityID string) *models.RelationshipProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.relationshipProfiles[identityID]
}

// UpdateRelationshipProfile 更新关系画像
func (e *RelationshipEngine) UpdateRelationshipProfile(identityID string, profile *models.RelationshipProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.UpdatedAt = time.Now()
	e.relationshipProfiles[identityID] = profile
	return nil
}

// UpdateAttachmentStyle 更新依恋风格
func (e *RelationshipEngine) UpdateAttachmentStyle(identityID string, style models.AttachmentStyleProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.relationshipProfiles[identityID]
	if profile == nil {
		profile = models.NewRelationshipProfile(identityID)
		e.relationshipProfiles[identityID] = profile
	}

	profile.AttachmentStyle = style
	profile.UpdatedAt = time.Now()

	return nil
}

// UpdateSocialStyle 更新社交风格
func (e *RelationshipEngine) UpdateSocialStyle(identityID string, style models.SocialStyleProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.relationshipProfiles[identityID]
	if profile == nil {
		profile = models.NewRelationshipProfile(identityID)
		e.relationshipProfiles[identityID] = profile
	}

	profile.SocialStyle = style
	profile.UpdatedAt = time.Now()

	return nil
}

// UpdateConflictStyle 更新冲突风格
func (e *RelationshipEngine) UpdateConflictStyle(identityID string, style models.ConflictStyleProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := e.relationshipProfiles[identityID]
	if profile == nil {
		profile = models.NewRelationshipProfile(identityID)
		e.relationshipProfiles[identityID] = profile
	}

	profile.ConflictStyle = style
	profile.UpdatedAt = time.Now()

	return nil
}

// === 决策上下文 ===

// GetDecisionContext 获取人际关系决策上下文
func (e *RelationshipEngine) GetDecisionContext(identityID string) *RelationshipDecisionContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		system = models.NewRelationshipSystem()
	}

	profile := e.relationshipProfiles[identityID]
	if profile == nil {
		profile = models.NewRelationshipProfile(identityID)
	}

	// 构建网络摘要
	networkSummary := e.buildNetworkSummary(system)

	// 计算决策影响
	decisionInfluence := e.calculateDecisionInfluence(profile, system)

	// 生成社交建议
	socialGuidance := e.generateSocialGuidance(profile, system)

	return &RelationshipDecisionContext{
		IdentityID:             identityID,
		NetworkSummary:         networkSummary,
		RelationshipOrientation: profile.RelationshipOrientation,
		SocialStyle:            profile.SocialStyle,
		AttachmentStyle:        profile.AttachmentStyle,
		DecisionInfluence:      decisionInfluence,
		SocialGuidance:         socialGuidance,
		Timestamp:              time.Now(),
	}
}

// buildNetworkSummary 构建网络摘要
func (e *RelationshipEngine) buildNetworkSummary(system *models.RelationshipSystem) NetworkSummary {
	network := system.SocialNetwork
	return NetworkSummary{
		TotalContacts:   network.TotalContacts,
		CloseContacts:   network.CloseContacts,
		SupportContacts: network.SupportContacts,
		SocialCapital:   network.SocialCapital,
		StrongTies:      network.StrongTies,
		WeakTies:        network.WeakTies,
		NetworkHealth:   network.GetOverallSocialHealth(),
	}
}

// calculateDecisionInfluence 计算决策影响
func (e *RelationshipEngine) calculateDecisionInfluence(profile *models.RelationshipProfile, system *models.RelationshipSystem) RelationshipDecisionInfluence {
	influence := RelationshipDecisionInfluence{}

	// 从关系倾向提取
	influence.SocialApproach = profile.RelationshipOrientation.RelationshipSeeking
	influence.TrustInOthers = 0.5 // 默认中等信任
	influence.CooperationReadiness = profile.RelationshipSkills.OverallCompetence
	influence.VulnerabilityComfort = profile.RelationshipOrientation.VulnerabilityWillingness
	influence.CommitmentWillingness = profile.RelationshipOrientation.CommitmentReadiness
	influence.SupportWillingness = profile.RelationshipSkills.SupportProvision
	influence.IntimacySeeking = profile.RelationshipOrientation.IntimacyComfort

	// 从冲突风格提取
	influence.ConflictStyle = profile.ConflictStyle.PrimaryStyle
	influence.ConfrontationReadiness = profile.ConflictStyle.ConfrontationComfort
	influence.ForgivenessReadiness = profile.ConflictStyle.ForgivenessTendency

	// 从依恋风格提取
	influence.SecurityLevel = 1 - profile.AttachmentStyle.AnxietyLevel - profile.AttachmentStyle.AvoidanceLevel + 0.5
	influence.SecurityLevel = clamp01(influence.SecurityLevel)
	influence.SeparationAnxiety = profile.AttachmentStyle.SeparationAnxiety
	influence.ReassuranceNeed = profile.AttachmentStyle.AnxietyLevel

	// 根据网络状态调整
	if system.SocialNetwork.TotalContacts > 0 {
		// 有更多关系时信任度可能更高
		avgTrust := 0.0
		for _, r := range system.Relationships {
			avgTrust += r.Trust
		}
		avgTrust /= float64(len(system.Relationships))
		influence.TrustInOthers = avgTrust
	}

	return influence
}

// generateSocialGuidance 生成社交建议
func (e *RelationshipEngine) generateSocialGuidance(profile *models.RelationshipProfile, system *models.RelationshipSystem) SocialGuidance {
	guidance := SocialGuidance{}

	// 社交能量评估
	guidance.SocialEnergyLevel = profile.SocialStyle.SocialEnergy
	guidance.ShouldSocialize = profile.SocialStyle.SocialEnergy > 0.4

	// 群体偏好
	if profile.SocialStyle.GroupPreference < 0.3 {
		guidance.PreferredGroupSize = "one_on_one"
	} else if profile.SocialStyle.GroupPreference < 0.7 {
		guidance.PreferredGroupSize = "small"
	} else {
		guidance.PreferredGroupSize = "medium"
	}

	// 深度偏好
	if profile.SocialStyle.DeepTalkPreference > 0.6 {
		guidance.PreferredDepth = "deep"
	} else if profile.SocialStyle.SmallTalkComfort > 0.6 {
		guidance.PreferredDepth = "casual"
	} else {
		guidance.PreferredDepth = "mixed"
	}

	// 关系维护建议
	for _, r := range system.Relationships {
		// 需要维护的关系
		if r.LastInteraction.IsZero() || time.Since(r.LastInteraction) > 7*24*time.Hour {
			if r.Importance > 0.6 {
				guidance.MaintenanceNeeded = append(guidance.MaintenanceNeeded, r.PersonName)
			}
		}

		// 紧张关系
		if r.ConflictLevel > 0.5 {
			guidance.TensionRelationships = append(guidance.TensionRelationships, r.PersonName)
			guidance.ConflictRisk = max(guidance.ConflictRisk, r.ConflictLevel)
		}
	}

	// 自我关怀提醒
	if profile.SocialStyle.SocialEnergy < 0.3 || profile.AttachmentStyle.AnxietyLevel > 0.7 {
		guidance.SelfCareNeeded = true
	}

	// 边界提醒
	if profile.RelationshipSkills.BoundarySetting < 0.4 {
		guidance.BoundaryReminders = append(guidance.BoundaryReminders, "注意设定健康边界")
	}

	// 成长机会
	if system.SocialNetwork.WeakTies < 3 {
		guidance.GrowthOpportunities = append(guidance.GrowthOpportunities, "拓展社交网络")
	}
	if profile.RelationshipSkills.Empathy < 0.5 {
		guidance.GrowthOpportunities = append(guidance.GrowthOpportunities, "提升共情能力")
	}

	return guidance
}

// === 关系冲突处理 ===

// RecordConflict 记录冲突
func (e *RelationshipEngine) RecordConflict(identityID string, relationshipID string, description string, intensity float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.Relationships {
		if system.Relationships[i].RelationshipID == relationshipID {
			r := &system.Relationships[i]
			r.ConflictLevel = clamp01(intensity)
			r.NegativeInteractions++
			r.Trend = "fluctuating"

			// 通知冲突发生
			for _, listener := range e.listeners {
				listener.OnConflictOccurred(identityID, relationshipID, intensity)
			}
			break
		}
	}

	return nil
}

// ResolveConflict 解决冲突
func (e *RelationshipEngine) ResolveConflict(identityID string, relationshipID string, resolution string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	system := e.relationshipSystems[identityID]
	if system == nil {
		return nil
	}

	for i := range system.Relationships {
		if system.Relationships[i].RelationshipID == relationshipID {
			r := &system.Relationships[i]
			r.ConflictLevel = clamp01(r.ConflictLevel - 0.3)
			r.PositiveInteractions++

			// 根据解决方式调整信任
			if resolution == "collaborative" || resolution == "compromise" {
				r.Trust = clamp01(r.Trust + 0.1)
			}

			// 更新阶段
			e.updateRelationshipStage(r)
			break
		}
	}

	return nil
}

// === 监听器管理 ===

// AddListener 添加监听器
func (e *RelationshipEngine) AddListener(listener RelationshipListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemoveListener 移除监听器
func (e *RelationshipEngine) RemoveListener(listener RelationshipListener) {
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

func generateRelationshipID() string {
	return "rel_" + time.Now().Format("20060102150405")
}

func generateInteractionID() string {
	return "inter_" + time.Now().Format("20060102150405")
}

func generateCircleID() string {
	return "circle_" + time.Now().Format("20060102150405")
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

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
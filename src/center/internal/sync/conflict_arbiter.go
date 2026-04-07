package sync

import (
	"log"
	"time"

	"github.com/ofa/center/internal/models"
)

// === 冲突仲裁器 ===

// ConflictArbiter - 冲突仲裁器 (v2.6.0)
//
// Center 作为永远在线的灵魂载体，负责统一决策和纠偏。
// 当设备间数据冲突时，由 Center 仲裁最终状态。
type ConflictArbiter struct {
	// 默认策略
	defaultStrategy ConflictStrategy
	// 特定字段的策略
	fieldStrategies map[string]ConflictStrategy
}

// ConflictStrategy - 冲突解决策略接口
type ConflictStrategy interface {
	Resolve(local, remote *models.PersonalIdentity, deviceID string) *models.PersonalIdentity
	Name() string
}

// NewConflictArbiter 创建冲突仲裁器
func NewConflictArbiter() *ConflictArbiter {
	arbiter := &ConflictArbiter{
		defaultStrategy: &AuthorityStrategy{},
		fieldStrategies: make(map[string]ConflictStrategy),
	}

	// 配置字段级策略
	arbiter.fieldStrategies["personality"] = &MergeStrategy{}
	arbiter.fieldStrategies["value_system"] = &MergeStrategy{}
	arbiter.fieldStrategies["interests"] = &MergeStrategy{}

	return arbiter
}

// Arbitrate 仲裁冲突
//
// local: Center 当前存储的身份（基准）
// remote: 设备上报的身份
// deviceID: 设备 ID
// 返回: 仲裁后的最终身份
func (a *ConflictArbiter) Arbitrate(local, remote *models.PersonalIdentity, deviceID string) *models.PersonalIdentity {
	if local == nil {
		return remote
	}
	if remote == nil {
		return local
	}

	// 使用默认策略仲裁
	result := a.defaultStrategy.Resolve(local, remote, deviceID)

	log.Printf("Conflict arbitrated by %s for device %s: local_v%d vs remote_v%d -> v%d",
		a.defaultStrategy.Name(), deviceID,
		local.Version, remote.Version, result.Version)

	return result
}

// ArbitrateField 仲裁特定字段
func (a *ConflictArbiter) ArbitrateField(fieldName string, local, remote *models.PersonalIdentity, deviceID string) interface{} {
	if strategy, ok := a.fieldStrategies[fieldName]; ok {
		// 使用字段级策略
		resolved := strategy.Resolve(local, remote, deviceID)

		switch fieldName {
		case "personality":
			return resolved.Personality
		case "value_system":
			return resolved.ValueSystem
		case "interests":
			return resolved.Interests
		}
	}

	return nil
}

// === 策略实现 ===

// AuthorityStrategy - Center 权威策略
//
// Center 是永远在线的灵魂载体，保持最终基准。
// 冲突时以 Center 的数据为准，设备数据仅作参考。
type AuthorityStrategy struct{}

func (s *AuthorityStrategy) Name() string {
	return "Authority"
}

func (s *AuthorityStrategy) Resolve(local, remote *models.PersonalIdentity, deviceID string) *models.PersonalIdentity {
	// Center 权威策略：以 Center 的数据为基准
	// 仅当设备版本更新时，才考虑合并

	if remote.Version > local.Version {
		// 设备版本更新，但 Center 保持基准
		// 选择性合并新数据
		return s.mergeWithAuthority(local, remote)
	}

	// Center 版本更新或相同，保持 Center 状态
	return local
}

func (s *AuthorityStrategy) mergeWithAuthority(base, incoming *models.PersonalIdentity) *models.PersonalIdentity {
	result := *base // 复制 Center 基准

	// 仅合并不冲突的字段
	if incoming.Name != "" && incoming.Name != base.Name {
		// 名称冲突：保持 Center 的名称
		log.Printf("Name conflict: Center=%s, Device=%s, keeping Center", base.Name, incoming.Name)
	}

	// 合并新兴趣（不覆盖现有）
	if incoming.Interests != nil {
		existingCategories := make(map[string]bool)
		for _, interest := range base.Interests {
			existingCategories[interest.Category] = true
		}

		for _, interest := range incoming.Interests {
			if !existingCategories[interest.Category] {
				result.Interests = append(result.Interests, interest)
			}
		}
	}

	// 版本号取较大值
	result.Version = max(base.Version, incoming.Version)
	result.UpdatedAt = time.Now()

	return &result
}

// MergeStrategy - 智能合并策略
//
// 用于性格、价值观等复杂对象的合并。
// 采用加权平均、取最新值等方法。
type MergeStrategy struct{}

func (s *MergeStrategy) Name() string {
	return "Merge"
}

func (s *MergeStrategy) Resolve(local, remote *models.PersonalIdentity, deviceID string) *models.PersonalIdentity {
	// 智能合并：取较新的非空值
	result := *local

	if remote.Personality != nil {
		if local.Personality == nil {
			result.Personality = remote.Personality
		} else {
			// 性格特质合并（加权平均，Center 权重更高）
			result.Personality = s.mergePersonality(local.Personality, remote.Personality)
		}
	}

	if remote.ValueSystem != nil {
		if local.ValueSystem == nil {
			result.ValueSystem = remote.ValueSystem
		} else {
			result.ValueSystem = s.mergeValueSystem(local.ValueSystem, remote.ValueSystem)
		}
	}

	result.Version = max(local.Version, remote.Version)
	result.UpdatedAt = time.Now()

	return &result
}

func (s *MergeStrategy) mergePersonality(local, remote *models.Personality) *models.Personality {
	result := *local

	// Center 权重 0.7，设备权重 0.3
	localWeight := 0.7
	remoteWeight := 0.3

	// Big Five 加权合并
	result.Openness = local.Openness*localWeight + remote.Openness*remoteWeight
	result.Conscientiousness = local.Conscientiousness*localWeight + remote.Conscientiousness*remoteWeight
	result.Extraversion = local.Extraversion*localWeight + remote.Extraversion*remoteWeight
	result.Agreeableness = local.Agreeableness*localWeight + remote.Agreeableness*remoteWeight
	result.Neuroticism = local.Neuroticism*localWeight + remote.Neuroticism*remoteWeight

	// MBTI 类型保持 Center 的
	// 设备的 MBTI 推断仅作参考

	return &result
}

func (s *MergeStrategy) mergeValueSystem(local, remote *models.ValueSystem) *models.ValueSystem {
	result := *local

	localWeight := 0.7
	remoteWeight := 0.3

	result.Privacy = local.Privacy*localWeight + remote.Privacy*remoteWeight
	result.Efficiency = local.Efficiency*localWeight + remote.Efficiency*remoteWeight
	result.Health = local.Health*localWeight + remote.Health*remoteWeight
	result.Family = local.Family*localWeight + remote.Family*remoteWeight
	result.Career = local.Career*localWeight + remote.Career*remoteWeight
	result.Entertainment = local.Entertainment*localWeight + remote.Entertainment*remoteWeight
	result.Learning = local.Learning*localWeight + remote.Learning*remoteWeight
	result.Social = local.Social*localWeight + remote.Social*remoteWeight
	result.Finance = local.Finance*localWeight + remote.Finance*remoteWeight
	result.Environment = local.Environment*localWeight + remote.Environment*remoteWeight

	return &result
}

// TimestampStrategy - 时间戳优先策略
//
// 简单策略：取 updatedAt 更新的数据。
// 适用于不需要复杂合并的场景。
type TimestampStrategy struct{}

func (s *TimestampStrategy) Name() string {
	return "Timestamp"
}

func (s *TimestampStrategy) Resolve(local, remote *models.PersonalIdentity, deviceID string) *models.PersonalIdentity {
	if remote.UpdatedAt.After(local.UpdatedAt) {
		return remote
	}
	return local
}

// === 辅助函数 ===

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
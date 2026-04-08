package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// GroupType 群组类型
type GroupType string

const (
	GroupTypeHome     GroupType = "home"     // 家庭群组
	GroupTypeWork     GroupType = "work"     // 工作群组
	GroupTypePersonal GroupType = "personal" // 个人群组
	GroupTypeTravel   GroupType = "travel"   // 出行群组
	GroupTypeFitness  GroupType = "fitness"  // 健身群组
	GroupTypeCustom   GroupType = "custom"   // 自定义群组
)

// GroupMemberRole 群组成员角色
type GroupMemberRole string

const (
	GroupRoleOwner   GroupMemberRole = "owner"   // 群主
	GroupRoleAdmin   GroupMemberRole = "admin"   // 管理员
	GroupRoleMember  GroupMemberRole = "member"  // 普通成员
	GroupRoleGuest   GroupMemberRole = "guest"   // 访客
)

// DeviceGroup 设备群组
type DeviceGroup struct {
	GroupID       string            `json:"group_id"`
	IdentityID    string            `json:"identity_id"`
	Name          string            `json:"name"`
	Type          GroupType         `json:"type"`
	Description   string            `json:"description,omitempty"`
	Members       []*GroupMember    `json:"members"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	IsActive      bool              `json:"is_active"`
	Priority      int               `json:"priority"`          // 群组优先级 (0-100)
	Settings      GroupSettings     `json:"settings"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// GroupMember 群组成员
type GroupMember struct {
	AgentID       string          `json:"agent_id"`
	DeviceName    string          `json:"device_name"`
	DeviceType    string          `json:"device_type"`
	Role          GroupMemberRole `json:"role"`
	JoinedAt      time.Time       `json:"joined_at"`
	LastActiveAt  time.Time       `json:"last_active_at"`
	IsOnline      bool            `json:"is_online"`
	Priority      int             `json:"priority"`          // 成员在群组内的优先级
	Capabilities  []string        `json:"capabilities,omitempty"`
}

// GroupSettings 群组设置
type GroupSettings struct {
	MaxMembers         int       `json:"max_members"`
	AllowGuests        bool      `json:"allow_guests"`
	BroadcastEnabled   bool      `json:"broadcast_enabled"`
	SyncEnabled        bool      `json:"sync_enabled"`
	QuietHoursStart    int       `json:"quiet_hours_start"`
	QuietHoursEnd      int       `json:"quiet_hours_end"`
	DefaultPriority    int       `json:"default_priority"`
	AutoActivateScene  string    `json:"auto_activate_scene,omitempty"`
}

// DefaultGroupSettings 默认群组设置
func DefaultGroupSettings() GroupSettings {
	return GroupSettings{
		MaxMembers:       10,
		AllowGuests:      true,
		BroadcastEnabled: true,
		SyncEnabled:      true,
		QuietHoursStart:  22,
		QuietHoursEnd:    7,
		DefaultPriority:  50,
	}
}

// DeviceGroupManager 设备群组管理器
type DeviceGroupManager struct {
	mu            sync.RWMutex
	groups        map[string]*DeviceGroup          // groupID -> group
	identityGroups map[string][]string             // identityID -> []groupID
	agentGroups   map[string][]string              // agentID -> []groupID
	config        GroupManagerConfig
	listeners     []GroupListener
	stateManager  *StateSyncManager
	messageBus    *MessageBus
}

// GroupManagerConfig 群组管理器配置
type GroupManagerConfig struct {
	MaxGroupsPerIdentity int           `json:"max_groups_per_identity"`
	DefaultMaxMembers    int           `json:"default_max_members"`
	AutoCleanupInterval  time.Duration `json:"auto_cleanup_interval"`
	EnableAutoSync       bool          `json:"enable_auto_sync"`
}

// DefaultGroupManagerConfig 默认配置
func DefaultGroupManagerConfig() GroupManagerConfig {
	return GroupManagerConfig{
		MaxGroupsPerIdentity: 20,
		DefaultMaxMembers:    10,
		AutoCleanupInterval:  5 * time.Minute,
		EnableAutoSync:       true,
	}
}

// GroupListener 群组事件监听器
type GroupListener interface {
	OnGroupCreated(group *DeviceGroup)
	OnGroupUpdated(group *DeviceGroup)
	OnGroupDeleted(groupID string)
	OnMemberAdded(group *DeviceGroup, member *GroupMember)
	OnMemberRemoved(group *DeviceGroup, agentID string)
	OnMemberRoleChanged(group *DeviceGroup, agentID string, oldRole, newRole GroupMemberRole)
	OnGroupBroadcast(group *DeviceGroup, message *Message)
}

// NewDeviceGroupManager 创建群组管理器
func NewDeviceGroupManager(config GroupManagerConfig) *DeviceGroupManager {
	return &DeviceGroupManager{
		groups:         make(map[string]*DeviceGroup),
		identityGroups: make(map[string][]string),
		agentGroups:    make(map[string][]string),
		config:         config,
		listeners:      make([]GroupListener, 0),
	}
}

// SetStateManager 设置状态管理器
func (m *DeviceGroupManager) SetStateManager(manager *StateSyncManager) {
	m.mu.Lock()
	m.stateManager = manager
	m.mu.Unlock()
}

// SetMessageBus 设置消息总线
func (m *DeviceGroupManager) SetMessageBus(bus *MessageBus) {
	m.mu.Lock()
	m.messageBus = bus
	m.mu.Unlock()
}

// AddListener 添加监听器
func (m *DeviceGroupManager) AddListener(listener GroupListener) {
	m.mu.Lock()
	m.listeners = append(m.listeners, listener)
	m.mu.Unlock()
}

// === 群组操作 ===

// CreateGroup 创建群组
func (m *DeviceGroupManager) CreateGroup(identityID, name string, groupType GroupType) (*DeviceGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查群组数量限制
	if len(m.identityGroups[identityID]) >= m.config.MaxGroupsPerIdentity {
		return nil, fmt.Errorf("maximum groups reached for identity %s", identityID)
	}

	groupID := generateGroupID(identityID)
	now := time.Now()

	group := &DeviceGroup{
		GroupID:    groupID,
		IdentityID: identityID,
		Name:       name,
		Type:       groupType,
		Members:    make([]*GroupMember, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
		IsActive:   true,
		Priority:   50,
		Settings:   DefaultGroupSettings(),
		Metadata:   make(map[string]string),
	}

	m.groups[groupID] = group
	m.identityGroups[identityID] = append(m.identityGroups[identityID], groupID)

	// 通知监听器
	m.notifyGroupCreated(group)

	return group, nil
}

// GetGroup 获取群组
func (m *DeviceGroupManager) GetGroup(groupID string) *DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.groups[groupID]
}

// GetIdentityGroups 获取身份的所有群组
func (m *DeviceGroupManager) GetIdentityGroups(identityID string) []*DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupIDs := m.identityGroups[identityID]
	groups := make([]*DeviceGroup, 0, len(groupIDs))
	for _, id := range groupIDs {
		if g := m.groups[id]; g != nil {
			groups = append(groups, g)
		}
	}
	return groups
}

// GetAgentGroups 获取设备所属的群组
func (m *DeviceGroupManager) GetAgentGroups(agentID string) []*DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupIDs := m.agentGroups[agentID]
	groups := make([]*DeviceGroup, 0, len(groupIDs))
	for _, id := range groupIDs {
		if g := m.groups[id]; g != nil && g.IsActive {
			groups = append(groups, g)
		}
	}
	return groups
}

// UpdateGroup 更新群组
func (m *DeviceGroupManager) UpdateGroup(groupID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				group.Name = v
			}
		case "description":
			if v, ok := value.(string); ok {
				group.Description = v
			}
		case "priority":
			if v, ok := value.(int); ok {
				group.Priority = v
			}
		case "is_active":
			if v, ok := value.(bool); ok {
				group.IsActive = v
			}
		case "settings":
			if v, ok := value.(GroupSettings); ok {
				group.Settings = v
			}
		}
	}

	group.UpdatedAt = time.Now()

	// 通知监听器
	m.notifyGroupUpdated(group)

	return nil
}

// DeleteGroup 删除群组
func (m *DeviceGroupManager) DeleteGroup(groupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// 从身份群组列表移除
	identityGroups := m.identityGroups[group.IdentityID]
	for i, id := range identityGroups {
		if id == groupID {
			m.identityGroups[group.IdentityID] = append(identityGroups[:i], identityGroups[i+1:]...)
			break
		}
	}

	// 从设备群组列表移除
	for _, member := range group.Members {
		agentGroups := m.agentGroups[member.AgentID]
		for i, id := range agentGroups {
			if id == groupID {
				m.agentGroups[member.AgentID] = append(agentGroups[:i], agentGroups[i+1:]...)
				break
			}
		}
	}

	// 删除群组
	delete(m.groups, groupID)

	// 通知监听器
	m.notifyGroupDeleted(groupID)

	return nil
}

// === 成员操作 ===

// AddMember 添加成员
func (m *DeviceGroupManager) AddMember(groupID, agentID, deviceName, deviceType string, role GroupMemberRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// 检查成员数量限制
	if len(group.Members) >= group.Settings.MaxMembers {
		return fmt.Errorf("group full: %s", groupID)
	}

	// 检查是否已存在
	for _, member := range group.Members {
		if member.AgentID == agentID {
			return fmt.Errorf("member already exists: %s", agentID)
		}
	}

	now := time.Now()
	member := &GroupMember{
		AgentID:      agentID,
		DeviceName:   deviceName,
		DeviceType:   deviceType,
		Role:         role,
		JoinedAt:     now,
		LastActiveAt: now,
		IsOnline:     true,
		Priority:     group.Settings.DefaultPriority,
		Capabilities: make([]string, 0),
	}

	group.Members = append(group.Members, member)
	group.UpdatedAt = now

	// 更新设备群组映射
	m.agentGroups[agentID] = append(m.agentGroups[agentID], groupID)

	// 通知监听器
	m.notifyMemberAdded(group, member)

	return nil
}

// RemoveMember 移除成员
func (m *DeviceGroupManager) RemoveMember(groupID, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	// 查找并移除成员
	for i, member := range group.Members {
		if member.AgentID == agentID {
			group.Members = append(group.Members[:i], group.Members[i+1:]...)
			group.UpdatedAt = time.Now()

			// 更新设备群组映射
			agentGroups := m.agentGroups[agentID]
			for j, id := range agentGroups {
				if id == groupID {
					m.agentGroups[agentID] = append(agentGroups[:j], agentGroups[j+1:]...)
					break
				}
			}

			// 通知监听器
			m.notifyMemberRemoved(group, agentID)
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", agentID)
}

// UpdateMemberRole 更新成员角色
func (m *DeviceGroupManager) UpdateMemberRole(groupID, agentID string, newRole GroupMemberRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	for _, member := range group.Members {
		if member.AgentID == agentID {
			oldRole := member.Role
			member.Role = newRole
			group.UpdatedAt = time.Now()

			// 通知监听器
			m.notifyMemberRoleChanged(group, agentID, oldRole, newRole)
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", agentID)
}

// UpdateMemberPriority 更新成员优先级
func (m *DeviceGroupManager) UpdateMemberPriority(groupID, agentID string, priority int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	for _, member := range group.Members {
		if member.AgentID == agentID {
			member.Priority = priority
			group.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", agentID)
}

// UpdateMemberOnline 更新成员在线状态
func (m *DeviceGroupManager) UpdateMemberOnline(groupID, agentID string, isOnline bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return fmt.Errorf("group not found: %s", groupID)
	}

	for _, member := range group.Members {
		if member.AgentID == agentID {
			member.IsOnline = isOnline
			member.LastActiveAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", agentID)
}

// GetMember 获取群组成员
func (m *DeviceGroupManager) GetMember(groupID, agentID string) *GroupMember {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group := m.groups[groupID]
	if group == nil {
		return nil
	}

	for _, member := range group.Members {
		if member.AgentID == agentID {
			return member
		}
	}
	return nil
}

// GetOnlineMembers 获取在线成员
func (m *DeviceGroupManager) GetOnlineMembers(groupID string) []*GroupMember {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group := m.groups[groupID]
	if group == nil {
		return nil
	}

	members := make([]*GroupMember, 0)
	for _, member := range group.Members {
		if member.IsOnline {
			members = append(members, member)
		}
	}
	return members
}

// === 群组广播 ===

// BroadcastToGroup 群组内广播消息
func (m *DeviceGroupManager) BroadcastToGroup(groupID string, message *Message) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	group := m.groups[groupID]
	if group == nil {
		return nil, fmt.Errorf("group not found: %s", groupID)
	}

	if !group.Settings.BroadcastEnabled {
		return nil, fmt.Errorf("broadcast disabled for group: %s", groupID)
	}

	// 检查勿扰时段
	if m.isQuietHours(group) {
		// 只发送给高优先级成员
		message.Priority = PriorityLow
	}

	// 获取目标成员
	targets := make([]string, 0)
	for _, member := range group.Members {
		if member.IsOnline && member.Role != GroupRoleGuest {
			targets = append(targets, member.AgentID)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no online members in group: %s", groupID)
	}

	// 通过消息总线发送
	delivered := make([]string, 0)
	if m.messageBus != nil {
		for _, agentID := range targets {
			msg := &Message{
				ID:         generateMessageID(),
				FromAgent:  "center",
				ToAgent:    agentID,
				IdentityID: group.IdentityID,
				Type:       message.Type,
				Priority:   message.Priority,
				Payload:    message.Payload,
				CreatedAt:  time.Now(),
				Metadata: map[string]string{
					"group_id":    groupID,
					"group_name":  group.Name,
					"broadcast":   "true",
				},
			}

			if err := m.messageBus.SendMessage(context.Background(), msg); err == nil {
				delivered = append(delivered, agentID)
			}
		}
	}

	// 通知监听器
	m.notifyGroupBroadcast(group, message)

	return delivered, nil
}

// BroadcastToAllGroups 广播到身份的所有群组
func (m *DeviceGroupManager) BroadcastToAllGroups(identityID string, message *Message) (map[string][]string, error) {
	m.mu.RLock()
	groupIDs := m.identityGroups[identityID]
	m.mu.RUnlock()

	results := make(map[string][]string)
	for _, groupID := range groupIDs {
		delivered, err := m.BroadcastToGroup(groupID, message)
		if err == nil && len(delivered) > 0 {
			results[groupID] = delivered
		}
	}

	return results, nil
}

// === 群组优先级 ===

// GetHighestPriorityGroup 获取最高优先级群组
func (m *DeviceGroupManager) GetHighestPriorityGroup(identityID string) *DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupIDs := m.identityGroups[identityID]
	if len(groupIDs) == 0 {
		return nil
	}

	var highest *DeviceGroup
	maxPriority := -1

	for _, id := range groupIDs {
		g := m.groups[id]
		if g != nil && g.IsActive && g.Priority > maxPriority {
			maxPriority = g.Priority
			highest = g
		}
	}

	return highest
}

// GetActiveGroupByScene 根据场景获取活动群组
func (m *DeviceGroupManager) GetActiveGroupByScene(identityID, scene string) *DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupIDs := m.identityGroups[identityID]
	for _, id := range groupIDs {
		g := m.groups[id]
		if g != nil && g.IsActive && g.Settings.AutoActivateScene == scene {
			return g
		}
	}

	// 尔回默认群组类型
	switch scene {
	case "home", "relaxing":
		return m.findGroupByType(identityID, GroupTypeHome)
	case "work", "meeting":
		return m.findGroupByType(identityID, GroupTypeWork)
	case "running", "fitness":
		return m.findGroupByType(identityID, GroupTypeFitness)
	case "driving", "traveling":
		return m.findGroupByType(identityID, GroupTypeTravel)
	}

	return nil
}

func (m *DeviceGroupManager) findGroupByType(identityID string, groupType GroupType) *DeviceGroup {
	groupIDs := m.identityGroups[identityID]
	for _, id := range groupIDs {
		g := m.groups[id]
		if g != nil && g.IsActive && g.Type == groupType {
			return g
		}
	}
	return nil
}

// === 统计与查询 ===

// GetGroupStats 获取群组统计
func (m *DeviceGroupManager) GetGroupStats(groupID string) *GroupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group := m.groups[groupID]
	if group == nil {
		return nil
	}

	stats := &GroupStats{
		GroupID:      groupID,
		TotalMembers: len(group.Members),
		OnlineMembers: 0,
		ByRole:       make(map[GroupMemberRole]int),
		ByDeviceType: make(map[string]int),
	}

	for _, member := range group.Members {
		if member.IsOnline {
			stats.OnlineMembers++
		}
		stats.ByRole[member.Role]++
		stats.ByDeviceType[member.DeviceType]++
	}

	return stats
}

// GroupStats 群组统计
type GroupStats struct {
	GroupID       string                    `json:"group_id"`
	TotalMembers  int                       `json:"total_members"`
	OnlineMembers int                       `json:"online_members"`
	ByRole        map[GroupMemberRole]int   `json:"by_role"`
	ByDeviceType  map[string]int            `json:"by_device_type"`
}

// GetIdentityStats 获取身份群组统计
func (m *DeviceGroupManager) GetIdentityStats(identityID string) *IdentityGroupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &IdentityGroupStats{
		IdentityID:   identityID,
		TotalGroups:  len(m.identityGroups[identityID]),
		ActiveGroups: 0,
		ByType:       make(map[GroupType]int),
		TotalDevices: 0,
	}

	for _, groupID := range m.identityGroups[identityID] {
		g := m.groups[groupID]
		if g != nil {
			if g.IsActive {
				stats.ActiveGroups++
			}
			stats.ByType[g.Type]++
			stats.TotalDevices += len(g.Members)
		}
	}

	return stats
}

// IdentityGroupStats 身份群组统计
type IdentityGroupStats struct {
	IdentityID   string            `json:"identity_id"`
	TotalGroups  int               `json:"total_groups"`
	ActiveGroups int               `json:"active_groups"`
	ByType       map[GroupType]int `json:"by_type"`
	TotalDevices int               `json:"total_devices"`
}

// === JSON 序列化 ===

// ToJSON 序列化群组
func (g *DeviceGroup) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

// DeviceGroupFromJSON 从 JSON 解析群组
func DeviceGroupFromJSON(data []byte) (*DeviceGroup, error) {
	group := &DeviceGroup{}
	if err := json.Unmarshal(data, group); err != nil {
		return nil, err
	}
	return group, nil
}

// ToMap 转换为 map
func (g *DeviceGroup) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"group_id":    g.GroupID,
		"identity_id": g.IdentityID,
		"name":        g.Name,
		"type":        string(g.Type),
		"is_active":   g.IsActive,
		"priority":    g.Priority,
		"created_at":  g.CreatedAt,
		"updated_at":  g.UpdatedAt,
		"member_count": len(g.Members),
	}

	if g.Description != "" {
		m["description"] = g.Description
	}

	return m
}

// === 辅助方法 ===

func generateGroupID(identityID string) string {
	return fmt.Sprintf("group-%s-%d", identityID, time.Now().UnixNano())
}

func (m *DeviceGroupManager) isQuietHours(group *DeviceGroup) bool {
	now := time.Now()
	hour := now.Hour()

	start := group.Settings.QuietHoursStart
	end := group.Settings.QuietHoursEnd

	if start > end {
		// 跨午夜: 22:00 - 07:00
		return hour >= start || hour < end
	}

	return hour >= start && hour < end
}

// === 监听器通知 ===

func (m *DeviceGroupManager) notifyGroupCreated(group *DeviceGroup) {
	for _, l := range m.listeners {
		l.OnGroupCreated(group)
	}
}

func (m *DeviceGroupManager) notifyGroupUpdated(group *DeviceGroup) {
	for _, l := range m.listeners {
		l.OnGroupUpdated(group)
	}
}

func (m *DeviceGroupManager) notifyGroupDeleted(groupID string) {
	for _, l := range m.listeners {
		l.OnGroupDeleted(groupID)
	}
}

func (m *DeviceGroupManager) notifyMemberAdded(group *DeviceGroup, member *GroupMember) {
	for _, l := range m.listeners {
		l.OnMemberAdded(group, member)
	}
}

func (m *DeviceGroupManager) notifyMemberRemoved(group *DeviceGroup, agentID string) {
	for _, l := range m.listeners {
		l.OnMemberRemoved(group, agentID)
	}
}

func (m *DeviceGroupManager) notifyMemberRoleChanged(group *DeviceGroup, agentID string, oldRole, newRole GroupMemberRole) {
	for _, l := range m.listeners {
		l.OnMemberRoleChanged(group, agentID, oldRole, newRole)
	}
}

func (m *DeviceGroupManager) notifyGroupBroadcast(group *DeviceGroup, message *Message) {
	for _, l := range m.listeners {
		l.OnGroupBroadcast(group, message)
	}
}
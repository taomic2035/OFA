package sync

import (
	"testing"
	"time"
)

func TestDeviceGroupManager(t *testing.T) {
	config := DefaultGroupManagerConfig()
	manager := NewDeviceGroupManager(config)

	// 测试初始化
	if manager.groups == nil {
		t.Error("groups map should be initialized")
	}
	if manager.identityGroups == nil {
		t.Error("identityGroups map should be initialized")
	}
	if manager.agentGroups == nil {
		t.Error("agentGroups map should be initialized")
	}
}

func TestDeviceGroupManagerCreateGroup(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, err := manager.CreateGroup("identity-001", "家庭群组", GroupTypeHome)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	// 验证群组属性
	if group.GroupID == "" {
		t.Error("GroupID should be generated")
	}
	if group.Name != "家庭群组" {
		t.Errorf("Expected name '家庭群组', got '%s'", group.Name)
	}
	if group.Type != GroupTypeHome {
		t.Errorf("Expected type 'home', got '%s'", group.Type)
	}
	if !group.IsActive {
		t.Error("Group should be active by default")
	}

	// 验证存储
	stored := manager.GetGroup(group.GroupID)
	if stored == nil {
		t.Fatal("Group not stored")
	}

	// 验证身份群组映射
	identityGroups := manager.GetIdentityGroups("identity-001")
	if len(identityGroups) != 1 {
		t.Errorf("Expected 1 identity group, got %d", len(identityGroups))
	}
}

func TestDeviceGroupManagerAddMember(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "工作群组", GroupTypeWork)

	// 添加成员
	err := manager.AddMember(group.GroupID, "agent-phone", "我的手机", "phone", GroupRoleOwner)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	// 验证成员
	member := manager.GetMember(group.GroupID, "agent-phone")
	if member == nil {
		t.Fatal("Member not found")
	}

	if member.DeviceName != "我的手机" {
		t.Errorf("Expected device name '我的手机', got '%s'", member.DeviceName)
	}
	if member.Role != GroupRoleOwner {
		t.Errorf("Expected role 'owner', got '%s'", member.Role)
	}
	if !member.IsOnline {
		t.Error("Member should be online by default")
	}

	// 验证设备群组映射
	agentGroups := manager.GetAgentGroups("agent-phone")
	if len(agentGroups) != 1 {
		t.Errorf("Expected 1 agent group, got %d", len(agentGroups))
	}
}

func TestDeviceGroupManagerRemoveMember(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)

	// 移除成员
	err := manager.RemoveMember(group.GroupID, "agent-001")
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	// 验证已移除
	member := manager.GetMember(group.GroupID, "agent-001")
	if member != nil {
		t.Error("Member should be removed")
	}

	// 验证设备群组映射已更新
	agentGroups := manager.GetAgentGroups("agent-001")
	if len(agentGroups) != 0 {
		t.Errorf("Expected 0 agent groups, got %d", len(agentGroups))
	}
}

func TestDeviceGroupManagerUpdateMemberRole(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)

	// 更新角色
	err := manager.UpdateMemberRole(group.GroupID, "agent-001", GroupRoleAdmin)
	if err != nil {
		t.Fatalf("UpdateMemberRole failed: %v", err)
	}

	member := manager.GetMember(group.GroupID, "agent-001")
	if member.Role != GroupRoleAdmin {
		t.Errorf("Expected role 'admin', got '%s'", member.Role)
	}
}

func TestDeviceGroupManagerUpdateMemberPriority(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)

	// 更新优先级
	err := manager.UpdateMemberPriority(group.GroupID, "agent-001", 80)
	if err != nil {
		t.Fatalf("UpdateMemberPriority failed: %v", err)
	}

	member := manager.GetMember(group.GroupID, "agent-001")
	if member.Priority != 80 {
		t.Errorf("Expected priority 80, got %d", member.Priority)
	}
}

func TestDeviceGroupManagerUpdateMemberOnline(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)

	// 设置离线
	err := manager.UpdateMemberOnline(group.GroupID, "agent-001", false)
	if err != nil {
		t.Fatalf("UpdateMemberOnline failed: %v", err)
	}

	member := manager.GetMember(group.GroupID, "agent-001")
	if member.IsOnline {
		t.Error("Member should be offline")
	}

	// 获取在线成员
	online := manager.GetOnlineMembers(group.GroupID)
	if len(online) != 0 {
		t.Errorf("Expected 0 online members, got %d", len(online))
	}
}

func TestDeviceGroupManagerUpdateGroup(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)

	// 更新群组
	err := manager.UpdateGroup(group.GroupID, map[string]interface{}{
		"name":        "更新后的群组",
		"description": "这是描述",
		"priority":    90,
	})
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	updated := manager.GetGroup(group.GroupID)
	if updated.Name != "更新后的群组" {
		t.Errorf("Expected name '更新后的群组', got '%s'", updated.Name)
	}
	if updated.Description != "这是描述" {
		t.Errorf("Expected description '这是描述', got '%s'", updated.Description)
	}
	if updated.Priority != 90 {
		t.Errorf("Expected priority 90, got %d", updated.Priority)
	}
}

func TestDeviceGroupManagerDeleteGroup(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)

	// 删除群组
	err := manager.DeleteGroup(group.GroupID)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	// 验证已删除
	if manager.GetGroup(group.GroupID) != nil {
		t.Error("Group should be deleted")
	}

	// 验证身份群组映射已更新
	identityGroups := manager.GetIdentityGroups("identity-001")
	if len(identityGroups) != 0 {
		t.Errorf("Expected 0 identity groups, got %d", len(identityGroups))
	}

	// 验证设备群组映射已更新
	agentGroups := manager.GetAgentGroups("agent-001")
	if len(agentGroups) != 0 {
		t.Errorf("Expected 0 agent groups, got %d", len(agentGroups))
	}
}

func TestDeviceGroupManagerMaxMembers(t *testing.T) {
	config := DefaultGroupManagerConfig()
	config.DefaultMaxMembers = 3
	manager := NewDeviceGroupManager(config)

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	group.Settings.MaxMembers = 2

	// 添加两个成员
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)
	manager.AddMember(group.GroupID, "agent-002", "设备2", "phone", GroupRoleMember)

	// 第三个应该失败
	err := manager.AddMember(group.GroupID, "agent-003", "设备3", "phone", GroupRoleMember)
	if err == nil {
		t.Error("Should fail when group is full")
	}
}

func TestDeviceGroupManagerMaxGroups(t *testing.T) {
	config := DefaultGroupManagerConfig()
	config.MaxGroupsPerIdentity = 2
	manager := NewDeviceGroupManager(config)

	// 创建两个群组
	manager.CreateGroup("identity-001", "群组1", GroupTypeHome)
	manager.CreateGroup("identity-001", "群组2", GroupTypeWork)

	// 第三个应该失败
	_, err := manager.CreateGroup("identity-001", "群组3", GroupTypePersonal)
	if err == nil {
		t.Error("Should fail when max groups reached")
	}
}

func TestDeviceGroupManagerGetGroupStats(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "手机", "phone", GroupRoleOwner)
	manager.AddMember(group.GroupID, "agent-002", "手表", "watch", GroupRoleMember)
	manager.UpdateMemberOnline(group.GroupID, "agent-002", false)

	stats := manager.GetGroupStats(group.GroupID)

	if stats.TotalMembers != 2 {
		t.Errorf("Expected 2 total members, got %d", stats.TotalMembers)
	}
	if stats.OnlineMembers != 1 {
		t.Errorf("Expected 1 online member, got %d", stats.OnlineMembers)
	}
	if stats.ByRole[GroupRoleOwner] != 1 {
		t.Errorf("Expected 1 owner, got %d", stats.ByRole[GroupRoleOwner])
	}
	if stats.ByDeviceType["phone"] != 1 {
		t.Errorf("Expected 1 phone, got %d", stats.ByDeviceType["phone"])
	}
}

func TestDeviceGroupManagerGetIdentityStats(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	manager.CreateGroup("identity-001", "家庭", GroupTypeHome)
	manager.CreateGroup("identity-001", "工作", GroupTypeWork)

	stats := manager.GetIdentityStats("identity-001")

	if stats.TotalGroups != 2 {
		t.Errorf("Expected 2 total groups, got %d", stats.TotalGroups)
	}
	if stats.ActiveGroups != 2 {
		t.Errorf("Expected 2 active groups, got %d", stats.ActiveGroups)
	}
	if stats.ByType[GroupTypeHome] != 1 {
		t.Errorf("Expected 1 home group, got %d", stats.ByType[GroupTypeHome])
	}
}

func TestDeviceGroupManagerHighestPriority(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	group1, _ := manager.CreateGroup("identity-001", "低优先级", GroupTypePersonal)
	group1.Priority = 30

	group2, _ := manager.CreateGroup("identity-001", "高优先级", GroupTypeWork)
	group2.Priority = 90

	highest := manager.GetHighestPriorityGroup("identity-001")
	if highest == nil {
		t.Fatal("Expected highest priority group")
	}
	if highest.Priority != 90 {
		t.Errorf("Expected priority 90, got %d", highest.Priority)
	}
}

func TestDeviceGroupManagerSceneGroup(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	// 创建群组并设置自动激活场景
	group1, _ := manager.CreateGroup("identity-001", "家庭群组", GroupTypeHome)
	group1.Settings.AutoActivateScene = "home"

	group2, _ := manager.CreateGroup("identity-001", "工作群组", GroupTypeWork)
	group2.Settings.AutoActivateScene = "work"

	// 测试场景匹配
	homeGroup := manager.GetActiveGroupByScene("identity-001", "home")
	if homeGroup == nil || homeGroup.Type != GroupTypeHome {
		t.Error("Expected home group for home scene")
	}

	workGroup := manager.GetActiveGroupByScene("identity-001", "meeting")
	if workGroup == nil || workGroup.Type != GroupTypeWork {
		t.Error("Expected work group for meeting scene")
	}
}

func TestDeviceGroupJSON(t *testing.T) {
	group := &DeviceGroup{
		GroupID:     "group-001",
		IdentityID:  "identity-001",
		Name:        "测试群组",
		Type:        GroupTypeHome,
		Description: "这是描述",
		Members: []*GroupMember{
			{
				AgentID:    "agent-001",
				DeviceName: "手机",
				DeviceType: "phone",
				Role:       GroupRoleOwner,
				IsOnline:   true,
				Priority:   80,
			},
		},
		CreatedAt: time.Now(),
		IsActive:  true,
		Priority:  70,
		Settings:  DefaultGroupSettings(),
	}

	// 序列化
	data, err := group.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := DeviceGroupFromJSON(data)
	if err != nil {
		t.Fatalf("DeviceGroupFromJSON failed: %v", err)
	}

	if parsed.GroupID != group.GroupID {
		t.Error("GroupID mismatch after JSON roundtrip")
	}
	if parsed.Name != group.Name {
		t.Error("Name mismatch after JSON roundtrip")
	}
	if len(parsed.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(parsed.Members))
	}
}

func TestDeviceGroupToMap(t *testing.T) {
	group := &DeviceGroup{
		GroupID:     "group-001",
		IdentityID:  "identity-001",
		Name:        "测试群组",
		Type:        GroupTypeHome,
		Description: "描述",
		IsActive:    true,
		Priority:    70,
		Members:     []*GroupMember{{AgentID: "agent-001"}},
	}

	m := group.ToMap()

	if m["group_id"] != group.GroupID {
		t.Error("group_id mismatch")
	}
	if m["name"] != group.Name {
		t.Error("name mismatch")
	}
	if m["type"] != string(GroupTypeHome) {
		t.Error("type mismatch")
	}
	if m["member_count"] != 1 {
		t.Errorf("Expected member_count 1, got %d", m["member_count"])
	}
}

func TestDeviceGroupManagerListener(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())

	created := false
	memberAdded := false
	updated := false

	listener := &TestGroupListener{
		onCreated: func(g *DeviceGroup) {
			created = true
		},
		onMemberAdded: func(g *DeviceGroup, m *GroupMember) {
			memberAdded = true
		},
		onUpdated: func(g *DeviceGroup) {
			updated = true
		},
	}

	manager.AddListener(listener)

	// 创建群组
	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)

	time.Sleep(50 * time.Millisecond)

	if !created {
		t.Error("Listener should have been notified on create")
	}

	// 添加成员
	manager.AddMember(group.GroupID, "agent-001", "设备", "phone", GroupRoleMember)

	time.Sleep(50 * time.Millisecond)

	if !memberAdded {
		t.Error("Listener should have been notified on member add")
	}

	// 更新群组
	manager.UpdateGroup(group.GroupID, map[string]interface{}{"name": "新名称"})

	time.Sleep(50 * time.Millisecond)

	if !updated {
		t.Error("Listener should have been notified on update")
	}
}

func TestDeviceGroupManagerBroadcast(t *testing.T) {
	manager := NewDeviceGroupManager(DefaultGroupManagerConfig())
	messageBus := NewMessageBus(DefaultMessageBusConfig())
	manager.SetMessageBus(messageBus)

	// 创建群组并添加成员
	group, _ := manager.CreateGroup("identity-001", "测试群组", GroupTypePersonal)
	manager.AddMember(group.GroupID, "agent-001", "设备1", "phone", GroupRoleMember)
	manager.AddMember(group.GroupID, "agent-002", "设备2", "watch", GroupRoleMember)

	// 广播消息
	msg := &Message{
		Type:     MessageTypeData,
		Priority: PriorityNormal,
		Payload:  map[string]interface{}{"content": "测试广播"},
	}

	delivered, err := manager.BroadcastToGroup(group.GroupID, msg)
	if err != nil {
		t.Fatalf("BroadcastToGroup failed: %v", err)
	}

	if len(delivered) != 2 {
		t.Errorf("Expected 2 delivered, got %d", len(delivered))
	}
}

// 测试监听器
type TestGroupListener struct {
	onCreated          func(g *DeviceGroup)
	onUpdated          func(g *DeviceGroup)
	onDeleted          func(groupID string)
	onMemberAdded      func(g *DeviceGroup, m *GroupMember)
	onMemberRemoved    func(g *DeviceGroup, agentID string)
	onMemberRoleChanged func(g *DeviceGroup, agentID string, oldRole, newRole GroupMemberRole)
	onBroadcast        func(g *DeviceGroup, msg *Message)
}

func (l *TestGroupListener) OnGroupCreated(g *DeviceGroup) {
	if l.onCreated != nil {
		l.onCreated(g)
	}
}

func (l *TestGroupListener) OnGroupUpdated(g *DeviceGroup) {
	if l.onUpdated != nil {
		l.onUpdated(g)
	}
}

func (l *TestGroupListener) OnGroupDeleted(groupID string) {
	if l.onDeleted != nil {
		l.onDeleted(groupID)
	}
}

func (l *TestGroupListener) OnMemberAdded(g *DeviceGroup, m *GroupMember) {
	if l.onMemberAdded != nil {
		l.onMemberAdded(g, m)
	}
}

func (l *TestGroupListener) OnMemberRemoved(g *DeviceGroup, agentID string) {
	if l.onMemberRemoved != nil {
		l.onMemberRemoved(g, agentID)
	}
}

func (l *TestGroupListener) OnMemberRoleChanged(g *DeviceGroup, agentID string, oldRole, newRole GroupMemberRole) {
	if l.onMemberRoleChanged != nil {
		l.onMemberRoleChanged(g, agentID, oldRole, newRole)
	}
}

func (l *TestGroupListener) OnGroupBroadcast(g *DeviceGroup, msg *Message) {
	if l.onBroadcast != nil {
		l.onBroadcast(g, msg)
	}
}
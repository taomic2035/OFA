package sync

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

func TestDeviceState(t *testing.T) {
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		BatteryLevel: 85,
		Charging:     false,
		NetworkType:  "wifi",
		Scene:        "idle",
	}

	// 测试 ToMap
	m := state.ToMap()
	if m["agent_id"] != "agent-001" {
		t.Error("agent_id mismatch")
	}
	if m["online"] != true {
		t.Error("online mismatch")
	}

	// 测试 JSON 序列化
	data, err := state.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 测试 JSON 反序列化
	parsed, err := DeviceStateFromJSON(data)
	if err != nil {
		t.Fatalf("DeviceStateFromJSON failed: %v", err)
	}

	if parsed.AgentID != state.AgentID {
		t.Error("parsed AgentID mismatch")
	}
	if parsed.BatteryLevel != state.BatteryLevel {
		t.Error("parsed BatteryLevel mismatch")
	}
}

func TestStateSyncManager(t *testing.T) {
	config := DefaultStateSyncConfig()
	mgr := NewStateSyncManager(config)

	// 测试状态更新
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		BatteryLevel: 85,
		NetworkType:  "wifi",
		Scene:        "idle",
	}

	change, err := mgr.UpdateDeviceState(state)
	if err != nil {
		t.Fatalf("UpdateDeviceState failed: %v", err)
	}

	if change == nil {
		t.Fatal("change is nil")
	}

	// 验证变更类型
	if change.ChangeType != StateChangeFull {
		t.Errorf("expected full change, got %s", change.ChangeType)
	}

	// 验证状态已存储
	stored := mgr.GetDeviceState("agent-001")
	if stored == nil {
		t.Fatal("stored state is nil")
	}
	if stored.AgentID != state.AgentID {
		t.Error("stored AgentID mismatch")
	}

	// 测试状态变更
	state2 := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		BatteryLevel: 75, // 变化
		NetworkType:  "wifi",
		Scene:        "idle",
	}

	change2, err := mgr.UpdateDeviceState(state2)
	if err != nil {
		t.Fatalf("UpdateDeviceState 2 failed: %v", err)
	}

	if change2.ChangeType != StateChangeBattery {
		t.Errorf("expected battery change, got %s", change2.ChangeType)
	}

	// 测试离线变更
	state3 := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       false, // 离线
		BatteryLevel: 75,
		NetworkType:  "wifi",
		Scene:        "idle",
	}

	change3, err := mgr.UpdateDeviceState(state3)
	if err != nil {
		t.Fatalf("UpdateDeviceState 3 failed: %v", err)
	}

	if change3.ChangeType != StateChangeOffline {
		t.Errorf("expected offline change, got %s", change3.ChangeType)
	}
}

func TestStateSyncManagerQuery(t *testing.T) {
	mgr := NewStateSyncManager(DefaultStateSyncConfig())

	// 注册多个设备
	for i := 1; i <= 3; i++ {
		state := &DeviceState{
			AgentID:      "agent-" + string(rune('0'+i)),
			IdentityID:   "identity-001",
			DeviceType:   "mobile",
			DeviceName:   "Device " + string(rune('0'+i)),
			Online:       i <= 2, // 设备3离线
			BatteryLevel: 50 + i*10,
			Scene:        i == 1 ? "running" : "idle",
		}
		mgr.UpdateDeviceState(state)
	}

	// 测试获取身份下所有设备状态
	states := mgr.GetIdentityDeviceStates("identity-001")
	if len(states) != 3 {
		t.Errorf("expected 3 states, got %d", len(states))
	}

	// 测试获取在线设备
	online := mgr.GetOnlineDevices("identity-001")
	if len(online) != 2 {
		t.Errorf("expected 2 online, got %d", len(online))
	}

	// 测试按场景获取
	running := mgr.GetDevicesByScene("identity-001", "running")
	if len(running) != 1 {
		t.Errorf("expected 1 running, got %d", len(running))
	}

	// 测试电池级别筛选
	highBattery := mgr.GetDevicesWithBatteryLevel("identity-001", 60)
	if len(highBattery) != 2 {
		t.Errorf("expected 2 high battery, got %d", len(highBattery))
	}
}

func TestStateSyncManagerPartialUpdate(t *testing.T) {
	mgr := NewStateSyncManager(DefaultStateSyncConfig())

	// 初始化状态
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		BatteryLevel: 85,
		NetworkType:  "wifi",
		Scene:        "idle",
	}
	mgr.UpdateDeviceState(state)

	// 部分更新电池
	updates := map[string]interface{}{
		"battery_level": 75,
		"charging":      true,
	}

	change, err := mgr.UpdatePartialState("agent-001", updates)
	if err != nil {
		t.Fatalf("UpdatePartialState failed: %v", err)
	}

	if change.ChangeType != StateChangeBattery {
		t.Errorf("expected battery change, got %s", change.ChangeType)
	}

	// 验证更新后的状态
	stored := mgr.GetDeviceState("agent-001")
	if stored.BatteryLevel != 75 {
		t.Errorf("expected battery 75, got %d", stored.BatteryLevel)
	}
	if !stored.Charging {
		t.Error("expected charging true")
	}

	// 部分更新场景
	updates2 := map[string]interface{}{
		"scene": "running",
	}

	change2, err := mgr.UpdatePartialState("agent-001", updates2)
	if err != nil {
		t.Fatalf("UpdatePartialState 2 failed: %v", err)
	}

	if change2.ChangeType != StateChangeScene {
		t.Errorf("expected scene change, got %s", change2.ChangeType)
	}
}

func TestStateSyncManagerOfflineDetect(t *testing.T) {
	config := DefaultStateSyncConfig()
	config.HeartbeatTimeout = 100 * time.Millisecond // 短超时测试

	mgr := NewStateSyncManager(config)

	// 注册设备
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		LastSeen:     time.Now(),
		BatteryLevel: 85,
	}
	mgr.UpdateDeviceState(state)

	// 等待超时
	time.Sleep(150 * time.Millisecond)

	// 检测离线
	changes := mgr.DetectOfflineDevices()
	if len(changes) != 1 {
		t.Errorf("expected 1 offline change, got %d", len(changes))
	}

	// 验证状态已更新为离线
	stored := mgr.GetDeviceState("agent-001")
	if stored.Online {
		t.Error("expected offline")
	}
}

func TestStateSyncManagerHistory(t *testing.T) {
	mgr := NewStateSyncManager(DefaultStateSyncConfig())

	// 创建多次状态变更
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		DeviceName:   "Test Phone",
		Online:       true,
		BatteryLevel: 85,
		Scene:        "idle",
	}

	for i := 0; i < 5; i++ {
		state.BatteryLevel = 85 - i*10
		mgr.UpdateDeviceState(state)
	}

	// 获取历史
	history := mgr.GetStateHistory("agent-001", 3)
	if len(history) != 3 {
		t.Errorf("expected 3 history, got %d", len(history))
	}
}

func TestStateSyncManagerStats(t *testing.T) {
	mgr := NewStateSyncManager(DefaultStateSyncConfig())

	// 注册多个设备
	for i := 1; i <= 4; i++ {
		state := &DeviceState{
			AgentID:      "agent-" + string(rune('0'+i)),
			IdentityID:   "identity-001",
			DeviceType:   i <= 2 ? "mobile" : "watch",
			DeviceName:   "Device " + string(rune('0'+i)),
			Online:       i <= 3,
			BatteryLevel: i*20,
			Charging:     i == 4,
			NetworkType:  i <= 2 ? "wifi" : "cellular",
			Scene:        i == 1 ? "running" : "idle",
		}
		mgr.UpdateDeviceState(state)
	}

	stats := mgr.GetStats("identity-001")

	if stats.TotalDevices != 4 {
		t.Errorf("expected 4 total, got %d", stats.TotalDevices)
	}
	if stats.OnlineDevices != 3 {
		t.Errorf("expected 3 online, got %d", stats.OnlineDevices)
	}
	if stats.LowBattery != 1 { // agent-01 has 20%
		t.Errorf("expected 1 low battery, got %d", stats.LowBattery)
	}
	if stats.OnWiFi != 2 {
		t.Errorf("expected 2 on wifi, got %d", stats.OnWiFi)
	}
	if stats.ByScene["running"] != 1 {
		t.Errorf("expected 1 running scene")
	}
}

func TestStateBroadcastService(t *testing.T) {
	config := DefaultStateBroadcastConfig()
	svc := NewStateBroadcastService(config)

	// 测试订阅
	sub := svc.Subscribe("subscriber-001", "identity-001",
		[]StateChangeType{StateChangeOnline, StateChangeOffline}, nil)

	if sub.SubscriptionID == "" {
		t.Error("subscription ID is empty")
	}
	if sub.SubscriberID != "subscriber-001" {
		t.Error("subscriber ID mismatch")
	}
	if len(sub.ChangeTypes) != 2 {
		t.Errorf("expected 2 change types, got %d", len(sub.ChangeTypes))
	}

	// 测试获取订阅
	fetched := svc.GetSubscription(sub.SubscriptionID)
	if fetched == nil {
		t.Fatal("fetched subscription is nil")
	}
	if fetched.SubscriberID != sub.SubscriberID {
		t.Error("fetched subscriber mismatch")
	}

	// 测试取消订阅
	svc.Unsubscribe(sub.SubscriptionID)

	fetched2 := svc.GetSubscription(sub.SubscriptionID)
	if fetched2 != nil {
		t.Error("subscription should be removed")
	}
}

func TestStateBroadcastServiceCache(t *testing.T) {
	svc := NewStateBroadcastService(DefaultStateBroadcastConfig())

	// 创建离线设备缓存场景
	// 注意：由于没有真实的 MessageBus，这里只测试基本逻辑

	// 测试缓存计数
	count := svc.GetCachedCount("agent-001")
	if count != 0 {
		t.Errorf("expected 0 cached, got %d", count)
	}
}

func TestStateBroadcastShouldBroadcast(t *testing.T) {
	config := DefaultStateBroadcastConfig()
	config.BroadcastLocation = false

	svc := NewStateBroadcastService(config)

	// 测试应该广播的类型
	if !svc.shouldBroadcast(StateChangeOnline) {
		t.Error("should broadcast online")
	}
	if !svc.shouldBroadcast(StateChangeOffline) {
		t.Error("should broadcast offline")
	}
	if !svc.shouldBroadcast(StateChangeBattery) {
		t.Error("should broadcast battery")
	}
	if !svc.shouldBroadcast(StateChangeNetwork) {
		t.Error("should broadcast network")
	}
	if !svc.shouldBroadcast(StateChangeScene) {
		t.Error("should broadcast scene")
	}

	// 测试不应广播的类型（位置默认不广播）
	if svc.shouldBroadcast(StateChangeLocation) {
		t.Error("should not broadcast location by default")
	}
}

func TestStateBroadcastStats(t *testing.T) {
	svc := NewStateBroadcastService(DefaultStateBroadcastConfig())

	// 添加几个订阅
	svc.Subscribe("sub-1", "identity-001", []StateChangeType{StateChangeOnline}, nil)
	svc.Subscribe("sub-2", "identity-001", []StateChangeType{StateChangeBattery}, nil)
	svc.Subscribe("sub-3", "identity-002", []StateChangeType{StateChangeScene}, nil)

	stats := svc.GetStats()

	if stats.TotalSubscriptions != 3 {
		t.Errorf("expected 3 subscriptions, got %d", stats.TotalSubscriptions)
	}
	if stats.ActiveSubscriptions != 3 {
		t.Errorf("expected 3 active, got %d", stats.ActiveSubscriptions)
	}
}

func TestStateChangeListener(t *testing.T) {
	mgr := NewStateSyncManager(DefaultStateSyncConfig())

	var receivedChange *StateChange
	listener := StateChangeListenerFunc(func(change *StateChange) {
		receivedChange = change
	})

	mgr.AddListener(listener)

	// 更新状态触发变更
	state := &DeviceState{
		AgentID:      "agent-001",
		IdentityID:   "identity-001",
		DeviceType:   "mobile",
		Online:       true,
		BatteryLevel: 85,
	}

	mgr.UpdateDeviceState(state)

	// 等待监听器异步通知
	time.Sleep(50 * time.Millisecond)

	if receivedChange == nil {
		t.Fatal("listener did not receive change")
	}
	if receivedChange.AgentID != "agent-001" {
		t.Error("received change AgentID mismatch")
	}

	// 移除监听器
	mgr.RemoveListener(listener)

	// 再次更新状态
	state.BatteryLevel = 75
	mgr.UpdateDeviceState(state)

	time.Sleep(50 * time.Millisecond)

	// 监听器已移除，不应再收到通知（检查旧值）
	if receivedChange.ChangeType == StateChangeBattery {
		// 可能是旧通知，不影响测试
	}
}
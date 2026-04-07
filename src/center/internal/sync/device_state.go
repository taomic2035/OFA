package sync

import (
	"encoding/json"
	"sync"
	"time"
)

// === 设备状态同步 (v3.1.0) ===
//
// Center 是永远在线的灵魂载体，设备状态同步确保：
// - 所有设备实时了解其他设备的状态
// - Center 维护全局状态视图
// - 状态变更自动广播到相关设备

// DeviceState - 设备状态模型
type DeviceState struct {
	AgentID      string                 `json:"agent_id"`
	IdentityID   string                 `json:"identity_id"`
	DeviceType   string                 `json:"device_type"`
	DeviceName   string                 `json:"device_name"`

	// 连接状态
	Online       bool                   `json:"online"`
	LastSeen     time.Time              `json:"last_seen"`

	// 电源状态
	BatteryLevel int                    `json:"battery_level"`  // 0-100
	Charging     bool                   `json:"charging"`
	PowerSaver   bool                   `json:"power_saver"`

	// 网络状态
	NetworkType  string                 `json:"network_type"`   // wifi, cellular, offline
	NetworkStrength int                 `json:"network_strength"` // 0-100
	Roaming      bool                   `json:"roaming"`

	// 场景状态 (v3.2.0 场景感知路由基础)
	Scene        string                 `json:"scene"`          // running, walking, driving, meeting, idle
	SceneContext map[string]interface{} `json:"scene_context"`

	// 能力状态
	Capabilities []string               `json:"capabilities"`
	ActiveApps   []string               `json:"active_apps"`

	// 位置信息 (可选)
	Location     *DeviceLocation        `json:"location,omitempty"`

	// 优先级与信任
	TrustLevel   TrustLevel             `json:"trust_level"`
	Priority     int                    `json:"priority"`

	// 状态时间戳
	StateVersion int64                  `json:"state_version"`
	UpdatedAt    time.Time              `json:"updated_at"`

	// 扩展状态
	Metadata     map[string]interface{} `json:"metadata"`
}

// DeviceLocation - 设备位置
type DeviceLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Accuracy    float64 `json:"accuracy"`
	LocationType string `json:"location_type"` // home, work, outdoor, unknown
}

// StateChangeType - 状态变更类型
type StateChangeType string

const (
	StateChangeOnline      StateChangeType = "online"
	StateChangeOffline     StateChangeType = "offline"
	StateChangeBattery     StateChangeType = "battery"
	StateChangeNetwork     StateChangeType = "network"
	StateChangeScene       StateChangeType = "scene"
	StateChangeLocation    StateChangeType = "location"
	StateChangeCapabilities StateChangeType = "capabilities"
	StateChangeFull        StateChangeType = "full"
)

// StateChange - 状态变更事件
type StateChange struct {
	AgentID      string         `json:"agent_id"`
	IdentityID   string         `json:"identity_id"`
	ChangeType   StateChangeType `json:"change_type"`
	OldState     *DeviceState   `json:"old_state,omitempty"`
	NewState     *DeviceState   `json:"new_state"`
	Timestamp    time.Time      `json:"timestamp"`
	Version      int64          `json:"version"`
}

// StateSyncConfig - 状态同步配置
type StateSyncConfig struct {
	// 状态广播间隔
	BroadcastInterval time.Duration
	// 心跳超时
	HeartbeatTimeout time.Duration
	// 状态过期时间
	StateExpiration time.Duration
	// 最大状态历史记录
	MaxStateHistory int
	// 是否广播位置变更
	BroadcastLocation bool
}

// DefaultStateSyncConfig 默认配置
func DefaultStateSyncConfig() StateSyncConfig {
	return StateSyncConfig{
		BroadcastInterval:  10 * time.Second,
		HeartbeatTimeout:   5 * time.Minute,
		StateExpiration:    30 * time.Minute,
		MaxStateHistory:    100,
		BroadcastLocation:  false, // 位置敏感，默认不广播
	}
}

// StateSyncManager - 状态同步管理器
type StateSyncManager struct {
	mu sync.RWMutex

	// 设备状态: agentId -> DeviceState
	deviceStates map[string]*DeviceState

	// 身份设备列表: identityId -> []agentId
	identityDevices map[string][]string

	// 状态历史: agentId -> []StateChange
	stateHistory map[string][]StateChange

	// 状态版本
	stateVersion int64

	// 配置
	config StateSyncConfig

	// 状态监听器
	listeners []StateChangeListener

	// 设备管理器引用
	deviceManager *DeviceManager

	// 消息总线引用
	messageBus *MessageBus
}

// StateChangeListener - 状态变更监听器
type StateChangeListener interface {
	OnStateChange(change *StateChange)
}

// StateChangeListenerFunc - 函数式监听器
type StateChangeListenerFunc func(change *StateChange)

func (f StateChangeListenerFunc) OnStateChange(change *StateChange) {
	f(change)
}

// NewStateSyncManager 创建状态同步管理器
func NewStateSyncManager(config StateSyncConfig) *StateSyncManager {
	if config.BroadcastInterval == 0 {
		config = DefaultStateSyncConfig()
	}

	return &StateSyncManager{
		deviceStates:    make(map[string]*DeviceState),
		identityDevices: make(map[string][]string),
		stateHistory:    make(map[string][]StateChange),
		config:          config,
		listeners:       make([]StateChangeListener, 0),
		stateVersion:    1,
	}
}

// SetDeviceManager 设置设备管理器
func (m *StateSyncManager) SetDeviceManager(dm *DeviceManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deviceManager = dm
}

// SetMessageBus 设置消息总线
func (m *StateSyncManager) SetMessageBus(mb *MessageBus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messageBus = mb
}

// === 状态上报 ===

// UpdateDeviceState 更新设备状态
func (m *StateSyncManager) UpdateDeviceState(state *DeviceState) (*StateChange, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取旧状态
	oldState := m.deviceStates[state.AgentID]

	// 设置时间戳
	state.UpdatedAt = time.Now()
	state.StateVersion = m.stateVersion
	m.stateVersion++

	// 存储新状态
	m.deviceStates[state.AgentID] = state

	// 更新身份设备列表
	if state.IdentityID != "" {
		found := false
		for _, aid := range m.identityDevices[state.IdentityID] {
			if aid == state.AgentID {
				found = true
				break
			}
		}
		if !found {
			m.identityDevices[state.IdentityID] = append(m.identityDevices[state.IdentityID], state.AgentID)
		}
	}

	// 检测变更类型
	changeType := m.detectChangeType(oldState, state)

	// 创建变更事件
	change := &StateChange{
		AgentID:    state.AgentID,
		IdentityID: state.IdentityID,
		ChangeType: changeType,
		OldState:   oldState,
		NewState:   state,
		Timestamp:  time.Now(),
		Version:    m.stateVersion,
	}

	// 记录历史
	m.recordHistory(state.AgentID, change)

	// 通知监听器
	m.notifyListeners(change)

	// 广播到消息总线
	if m.messageBus != nil && changeType != StateChangeFull {
		m.broadcastStateChange(change)
	}

	return change, nil
}

// UpdatePartialState 更新部分状态
func (m *StateSyncManager) UpdatePartialState(agentID string, updates map[string]interface{}) (*StateChange, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取现有状态
	state := m.deviceStates[agentID]
	if state == nil {
		return nil, nil
	}

	oldState := &DeviceState{}
	*oldState = *state // 复制旧状态

	// 应用更新
	for key, value := range updates {
		switch key {
		case "battery_level":
			if v, ok := value.(int); ok {
				state.BatteryLevel = v
			}
		case "charging":
			if v, ok := value.(bool); ok {
				state.Charging = v
			}
		case "power_saver":
			if v, ok := value.(bool); ok {
				state.PowerSaver = v
			}
		case "network_type":
			if v, ok := value.(string); ok {
				state.NetworkType = v
			}
		case "network_strength":
			if v, ok := value.(int); ok {
				state.NetworkStrength = v
			}
		case "scene":
			if v, ok := value.(string); ok {
				state.Scene = v
			}
		case "scene_context":
			if v, ok := value.(map[string]interface{}); ok {
				state.SceneContext = v
			}
		case "online":
			if v, ok := value.(bool); ok {
				state.Online = v
			}
		case "location":
			if v, ok := value.(*DeviceLocation); ok {
				state.Location = v
			}
		case "active_apps":
			if v, ok := value.([]string); ok {
				state.ActiveApps = v
			}
		default:
			// 扩展字段存入 Metadata
			state.Metadata[key] = value
		}
	}

	// 更新时间戳
	state.UpdatedAt = time.Now()
	state.StateVersion = m.stateVersion
	m.stateVersion++

	// 检测变更
	changeType := m.detectChangeType(oldState, state)

	change := &StateChange{
		AgentID:    agentID,
		IdentityID: state.IdentityID,
		ChangeType: changeType,
		OldState:   oldState,
		NewState:   state,
		Timestamp:  time.Now(),
		Version:    m.stateVersion,
	}

	m.recordHistory(agentID, change)
	m.notifyListeners(change)

	if m.messageBus != nil {
		m.broadcastStateChange(change)
	}

	return change, nil
}

// === 状态查询 ===

// GetDeviceState 获取设备状态
func (m *StateSyncManager) GetDeviceState(agentID string) *DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deviceStates[agentID]
}

// GetIdentityDeviceStates 获取身份下所有设备状态
func (m *StateSyncManager) GetIdentityDeviceStates(identityID string) []*DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var states []*DeviceState
	for _, aid := range m.identityDevices[identityID] {
		if state := m.deviceStates[aid]; state != nil {
			states = append(states, state)
		}
	}
	return states
}

// GetOnlineDevices 获取在线设备
func (m *StateSyncManager) GetOnlineDevices(identityID string) []*DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var online []*DeviceState
	for _, aid := range m.identityDevices[identityID] {
		if state := m.deviceStates[aid]; state != nil && state.Online {
			online = append(online, state)
		}
	}
	return online
}

// GetDevicesByScene 按场景获取设备
func (m *StateSyncManager) GetDevicesByScene(identityID string, scene string) []*DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var devices []*DeviceState
	for _, aid := range m.identityDevices[identityID] {
		if state := m.deviceStates[aid]; state != nil && state.Scene == scene {
			devices = append(devices, state)
		}
	}
	return devices
}

// GetDevicesWithBatteryLevel 获取指定电池级别的设备
func (m *StateSyncManager) GetDevicesWithBatteryLevel(identityID string, minLevel int) []*DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var devices []*DeviceState
	for _, aid := range m.identityDevices[identityID] {
		if state := m.deviceStates[aid]; state != nil && state.BatteryLevel >= minLevel {
			devices = append(devices, state)
		}
	}
	return devices
}

// GetDevicesWithNetwork 获取指定网络类型的设备
func (m *StateSyncManager) GetDevicesWithNetwork(identityID string, networkType string) []*DeviceState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var devices []*DeviceState
	for _, aid := range m.identityDevices[identityID] {
		if state := m.deviceStates[aid]; state != nil && state.NetworkType == networkType {
			devices = append(devices, state)
		}
	}
	return devices
}

// === 状态历史 ===

// GetStateHistory 获取状态历史
func (m *StateSyncManager) GetStateHistory(agentID string, limit int) []StateChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := m.stateHistory[agentID]
	if limit > 0 && len(history) > limit {
		return history[:limit]
	}
	return history
}

// GetRecentChanges 获取最近的变更
func (m *StateSyncManager) GetRecentChanges(identityID string, since time.Time) []StateChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var changes []StateChange
	for _, aid := range m.identityDevices[identityID] {
		history := m.stateHistory[aid]
		for _, change := range history {
			if change.Timestamp.After(since) {
				changes = append(changes, change)
			}
		}
	}
	return changes
}

// === 心跳与离线检测 ===

// UpdateHeartbeat 更新心跳
func (m *StateSyncManager) UpdateHeartbeat(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.deviceStates[agentID]
	if state != nil {
		state.Online = true
		state.LastSeen = time.Now()
		state.UpdatedAt = time.Now()
	}
}

// DetectOfflineDevices 检测离线设备
func (m *StateSyncManager) DetectOfflineDevices() []*StateChange {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var changes []*StateChange

	for agentID, state := range m.deviceStates {
		if state.Online && now.Sub(state.LastSeen) > m.config.HeartbeatTimeout {
			oldState := &DeviceState{}
			*oldState = *state

			state.Online = false
			state.UpdatedAt = now
			state.StateVersion = m.stateVersion
			m.stateVersion++

			change := &StateChange{
				AgentID:    agentID,
				IdentityID: state.IdentityID,
				ChangeType: StateChangeOffline,
				OldState:   oldState,
				NewState:   state,
				Timestamp:  now,
				Version:    m.stateVersion,
			}

			m.recordHistory(agentID, change)
			changes = append(changes, change)
		}
	}

	for _, change := range changes {
		m.notifyListeners(change)
		if m.messageBus != nil {
			m.broadcastStateChange(change)
		}
	}

	return changes
}

// === 状态清理 ===

// CleanExpiredStates 清理过期状态
func (m *StateSyncManager) CleanExpiredStates() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for agentID, state := range m.deviceStates {
		if !state.Online && now.Sub(state.UpdatedAt) > m.config.StateExpiration {
			expired = append(expired, agentID)
		}
	}

	for _, agentID := range expired {
		delete(m.deviceStates, agentID)
		delete(m.stateHistory, agentID)

		// 从身份列表移除
		for identityID, devices := range m.identityDevices {
			var newDevices []string
			for _, aid := range devices {
				if aid != agentID {
					newDevices = append(newDevices, aid)
				}
			}
			m.identityDevices[identityID] = newDevices
		}
	}

	return len(expired)
}

// TrimHistory 清理历史记录
func (m *StateSyncManager) TrimHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for agentID, history := range m.stateHistory {
		if len(history) > m.config.MaxStateHistory {
			m.stateHistory[agentID] = history[len(history)-m.config.MaxStateHistory:]
		}
	}
}

// === 监听器管理 ===

// AddListener 添加状态监听器
func (m *StateSyncManager) AddListener(listener StateChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// RemoveListener 移除状态监听器
func (m *StateSyncManager) RemoveListener(listener StateChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, l := range m.listeners {
		if l == listener {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			break
		}
	}
}

// === 私有方法 ===

func (m *StateSyncManager) detectChangeType(oldState, newState *DeviceState) StateChangeType {
	if oldState == nil {
		return StateChangeFull
	}

	// 检测上线/离线
	if oldState.Online != newState.Online {
		if newState.Online {
			return StateChangeOnline
		}
		return StateChangeOffline
	}

	// 检测电池变更
	if oldState.BatteryLevel != newState.BatteryLevel ||
		oldState.Charging != newState.Charging ||
		oldState.PowerSaver != newState.PowerSaver {
		return StateChangeBattery
	}

	// 检测网络变更
	if oldState.NetworkType != newState.NetworkType ||
		oldState.NetworkStrength != newState.NetworkStrength {
		return StateChangeNetwork
	}

	// 检测场景变更
	if oldState.Scene != newState.Scene {
		return StateChangeScene
	}

	// 检测位置变更（如果配置允许）
	if m.config.BroadcastLocation && oldState.Location != newState.Location {
		return StateChangeLocation
	}

	// 检测能力变更
	if !equalStringArrays(oldState.Capabilities, newState.Capabilities) {
		return StateChangeCapabilities
	}

	return StateChangeFull
}

func (m *StateSyncManager) recordHistory(agentID string, change *StateChange) {
	history := m.stateHistory[agentID]
	history = append(history, *change)

	// 限制历史记录大小
	if len(history) > m.config.MaxStateHistory {
		history = history[len(history)-m.config.MaxStateHistory:]
	}

	m.stateHistory[agentID] = history
}

func (m *StateSyncManager) notifyListeners(change *StateChange) {
	for _, listener := range m.listeners {
		go func(l StateChangeListener, c *StateChange) {
			l.OnStateChange(c)
		}(listener, change)
	}
}

func (m *StateSyncManager) broadcastStateChange(change *StateChange) {
	if m.messageBus == nil {
		return
	}

	// 构造状态变更消息
	msg := &Message{
		ID:          generateMessageID(),
		FromAgent:   "center",
		ToAgent:     "*", // 广播到身份所有设备
		IdentityID:  change.IdentityID,
		Type:        MessageTypeNotification,
		Priority:    MessagePriorityNormal,
		Payload:     change.ToMap(),
		CreatedAt:   time.Now(),
	}

	m.messageBus.Broadcast(msg)
}

func equalStringArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ToMap 将 DeviceState 转换为 map
func (s *DeviceState) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"agent_id":        s.AgentID,
		"identity_id":     s.IdentityID,
		"device_type":     s.DeviceType,
		"device_name":     s.DeviceName,
		"online":          s.Online,
		"battery_level":   s.BatteryLevel,
		"charging":        s.Charging,
		"power_saver":     s.PowerSaver,
		"network_type":    s.NetworkType,
		"network_strength": s.NetworkStrength,
		"scene":           s.Scene,
		"trust_level":     s.TrustLevel,
		"priority":        s.Priority,
		"state_version":   s.StateVersion,
		"updated_at":      s.UpdatedAt,
	}

	if s.SceneContext != nil {
		m["scene_context"] = s.SceneContext
	}
	if s.Capabilities != nil {
		m["capabilities"] = s.Capabilities
	}
	if s.ActiveApps != nil {
		m["active_apps"] = s.ActiveApps
	}
	if s.Location != nil {
		m["location"] = map[string]interface{}{
			"latitude":     s.Location.Latitude,
			"longitude":    s.Location.Longitude,
			"accuracy":     s.Location.Accuracy,
			"location_type": s.Location.LocationType,
		}
	}
	if s.Metadata != nil {
		m["metadata"] = s.Metadata
	}

	return m
}

// ToJSON 转换为 JSON
func (s *DeviceState) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON 从 JSON 解析
func DeviceStateFromJSON(data []byte) (*DeviceState, error) {
	var state DeviceState
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// ToMap 将 StateChange 转换为 map
func (c *StateChange) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"agent_id":    c.AgentID,
		"identity_id": c.IdentityID,
		"change_type": c.ChangeType,
		"timestamp":   c.Timestamp,
		"version":     c.Version,
	}

	if c.NewState != nil {
		m["new_state"] = c.NewState.ToMap()
	}
	if c.OldState != nil {
		m["old_state"] = c.OldState.ToMap()
	}

	return m
}

// === 状态统计 ===

// StateStats 状态统计
type StateStats struct {
	TotalDevices    int            `json:"total_devices"`
	OnlineDevices   int            `json:"online_devices"`
	OfflineDevices  int            `json:"offline_devices"`
	LowBattery      int            `json:"low_battery"`     // < 20%
	Charging        int            `json:"charging"`
	OnWiFi          int            `json:"on_wifi"`
	OnCellular      int            `json:"on_cellular"`
	ByScene         map[string]int `json:"by_scene"`
	ByDeviceType    map[string]int `json:"by_device_type"`
}

// GetStats 获取状态统计
func (m *StateSyncManager) GetStats(identityID string) *StateStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &StateStats{
		ByScene:      make(map[string]int),
		ByDeviceType: make(map[string]int),
	}

	for _, aid := range m.identityDevices[identityID] {
		state := m.deviceStates[aid]
		if state == nil {
			continue
		}

		stats.TotalDevices++
		if state.Online {
			stats.OnlineDevices++
		} else {
			stats.OfflineDevices++
		}

		if state.BatteryLevel < 20 {
			stats.LowBattery++
		}
		if state.Charging {
			stats.Charging++
		}

		if state.NetworkType == "wifi" {
			stats.OnWiFi++
		} else if state.NetworkType == "cellular" {
			stats.OnCellular++
		}

		if state.Scene != "" {
			stats.ByScene[state.Scene]++
		}
		stats.ByDeviceType[state.DeviceType]++
	}

	return stats
}
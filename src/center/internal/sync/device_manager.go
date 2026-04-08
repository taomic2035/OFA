package sync

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// === 设备管理器 ===

// DeviceManager - 设备管理器 (v2.8.0)
//
// 管理所有设备的状态、心跳、离线检测。
// Center 是永远在线的灵魂载体，设备可能离线、更换。
// v2.8.0: 新增信任链集成，支持设备优先级和安全绑定。
type DeviceManager struct {
	mu sync.RWMutex

	// 设备列表：agentId -> DeviceInfo
	devices map[string]*DeviceInfo

	// 身份绑定的设备：identityId -> []agentId
	identityDevices map[string][]string

	// v2.8.0: 信任管理器
	trustManager *TrustManager

	// 配置
	config DeviceManagerConfig
}

// DeviceInfo - 设备信息 (v2.8.0)
type DeviceInfo struct {
	AgentID       string       `json:"agent_id"`
	IdentityID    string       `json:"identity_id"`
	DeviceType    string       `json:"device_type"`    // mobile, tablet, watch, iot...
	DeviceName    string       `json:"device_name"`
	Status        DeviceStatus `json:"status"`         // online, offline, suspended
	TrustLevel    TrustLevel   `json:"trust_level"`    // v2.8.0: 信任级别
	Priority      int          `json:"priority"`       // v2.8.0: 设备优先级 (0-100)
	LastSeen      time.Time    `json:"last_seen"`
	RegisteredAt  time.Time    `json:"registered_at"`
	SyncVersion   int64        `json:"sync_version"`   // 最后同步版本
	Capabilities  []string     `json:"capabilities"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DeviceStatus - 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline    DeviceStatus = "online"
	DeviceStatusOffline   DeviceStatus = "offline"
	DeviceStatusSuspended DeviceStatus = "suspended"
)

// DeviceManagerConfig - 设备管理器配置
type DeviceManagerConfig struct {
	// 离线检测阈值
	OfflineThreshold time.Duration
	// 心跳间隔
	HeartbeatInterval time.Duration
	// 最大设备数
	MaxDevicesPerIdentity int
}

// DefaultDeviceManagerConfig 默认配置
func DefaultDeviceManagerConfig() DeviceManagerConfig {
	return DeviceManagerConfig{
		OfflineThreshold:      5 * time.Minute,
		HeartbeatInterval:     30 * time.Second,
		MaxDevicesPerIdentity: 10,
	}
}

// NewDeviceManager 创建设备管理器
func NewDeviceManager(config DeviceManagerConfig) *DeviceManager {
	if config.OfflineThreshold == 0 {
		config = DefaultDeviceManagerConfig()
	}

	return &DeviceManager{
		devices:         make(map[string]*DeviceInfo),
		identityDevices: make(map[string][]string),
		trustManager:    NewTrustManager(DefaultTrustManagerConfig()),
		config:          config,
	}
}

// SetTrustManager 设置信任管理器 (v2.8.0)
func (m *DeviceManager) SetTrustManager(tm *TrustManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trustManager = tm
}

// GetTrustManager 获取信任管理器 (v2.8.0)
func (m *DeviceManager) GetTrustManager() *TrustManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trustManager
}

// === 设备注册 ===

// RegisterDevice 注册设备
func (m *DeviceManager) RegisterDevice(agentID, identityID, deviceType, deviceName string, capabilities []string) *DeviceInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查设备数限制
	devices := m.identityDevices[identityID]
	if len(devices) >= m.config.MaxDevicesPerIdentity {
		// 移除最旧的离线设备
		m.removeOldestOfflineDevice(identityID)
	}

	// 确定信任级别和优先级
	trustLevel := TrustLevelLow
	priority := 50 // 默认优先级

	// 第一个设备为主设备
	if len(devices) == 0 {
		trustLevel = TrustLevelPrimary
		priority = 100
	}

	// 如果有信任管理器，注册信任设备
	if m.trustManager != nil {
		req := &RegisterDeviceRequest{
			AgentID:      agentID,
			IdentityID:   identityID,
			DeviceType:   deviceType,
			DeviceName:   deviceName,
			Capabilities: capabilities,
		}
		resp, err := m.trustManager.RegisterTrustedDevice(req)
		if err == nil && resp.Success {
			trustLevel = resp.TrustLevel
		}
	}

	now := time.Now()
	device := &DeviceInfo{
		AgentID:      agentID,
		IdentityID:   identityID,
		DeviceType:   deviceType,
		DeviceName:   deviceName,
		Status:       DeviceStatusOnline,
		TrustLevel:   trustLevel,
		Priority:     priority,
		LastSeen:     now,
		RegisteredAt: now,
		Capabilities: capabilities,
		Metadata:     make(map[string]interface{}),
	}

	m.devices[agentID] = device
	m.identityDevices[identityID] = append(m.identityDevices[identityID], agentID)

	log.Printf("Device registered: %s (%s) for identity %s", agentID, deviceType, identityID)

	return device
}

// UnregisterDevice 注销设备
func (m *DeviceManager) UnregisterDevice(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[agentID]
	if !exists {
		return
	}

	// 从身份设备列表中移除
	devices := m.identityDevices[device.IdentityID]
	for i, id := range devices {
		if id == agentID {
			m.identityDevices[device.IdentityID] = append(devices[:i], devices[i+1:]...)
			break
		}
	}

	// 删除设备
	delete(m.devices, agentID)

	log.Printf("Device unregistered: %s", agentID)
}

// === 心跳与状态 ===

// UpdateHeartbeat 更新心跳
func (m *DeviceManager) UpdateHeartbeat(agentID string, syncVersion int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[agentID]
	if !exists {
		return false
	}

	device.LastSeen = time.Now()
	device.Status = DeviceStatusOnline
	device.SyncVersion = syncVersion

	return true
}

// GetDevice 获取设备信息
func (m *DeviceManager) GetDevice(agentID string) *DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.devices[agentID]
}

// GetDevicesByIdentity 获取身份关联的所有设备
func (m *DeviceManager) GetDevicesByIdentity(identityID string) []*DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agentIDs := m.identityDevices[identityID]
	devices := make([]*DeviceInfo, 0, len(agentIDs))

	for _, agentID := range agentIDs {
		if device, exists := m.devices[agentID]; exists {
			devices = append(devices, device)
		}
	}

	return devices
}

// GetActiveDevices 获取活跃设备
func (m *DeviceManager) GetActiveDevices(identityID string) []*DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*DeviceInfo
	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists && device.Status == DeviceStatusOnline {
			active = append(active, device)
		}
	}

	return active
}

// === 离线检测 ===

// DetectOfflineDevices 检测离线设备
func (m *DeviceManager) DetectOfflineDevices() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var offline []string

	for agentID, device := range m.devices {
		if device.Status == DeviceStatusOnline {
			if now.Sub(device.LastSeen) > m.config.OfflineThreshold {
				device.Status = DeviceStatusOffline
				offline = append(offline, agentID)
				log.Printf("Device went offline: %s (last seen: %v)", agentID, device.LastSeen)
			}
		}
	}

	return offline
}

// IsDeviceOnline 检查设备是否在线
func (m *DeviceManager) IsDeviceOnline(agentID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, exists := m.devices[agentID]
	if !exists {
		return false
	}

	return device.Status == DeviceStatusOnline
}

// GetDeviceCount 获取设备数量
func (m *DeviceManager) GetDeviceCount(identityID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.identityDevices[identityID])
}

// === 设备版本管理 ===

// GetOutdatedDevices 获取版本落后的设备
func (m *DeviceManager) GetOutdatedDevices(identityID string, currentVersion int64) []*DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var outdated []*DeviceInfo
	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			if device.SyncVersion < currentVersion {
				outdated = append(outdated, device)
			}
		}
	}

	return outdated
}

// === 辅助方法 ===

func (m *DeviceManager) removeOldestOfflineDevice(identityID string) {
	devices := m.identityDevices[identityID]

	var oldestAgentID string
	var oldestSeen time.Time

	for _, agentID := range devices {
		if device, exists := m.devices[agentID]; exists {
			if device.Status == DeviceStatusOffline {
				if oldestAgentID == "" || device.LastSeen.Before(oldestSeen) {
					oldestAgentID = agentID
					oldestSeen = device.LastSeen
				}
			}
		}
	}

	if oldestAgentID != "" {
		delete(m.devices, oldestAgentID)
		log.Printf("Removed oldest offline device: %s", oldestAgentID)

		// 更新列表
		var newDevices []string
		for _, agentID := range devices {
			if agentID != oldestAgentID {
				newDevices = append(newDevices, agentID)
			}
		}
		m.identityDevices[identityID] = newDevices
	}
}

// === 设备统计 ===

// DeviceStats 设备统计
type DeviceStats struct {
	TotalDevices   int            `json:"total_devices"`
	OnlineDevices  int            `json:"online_devices"`
	OfflineDevices int            `json:"offline_devices"`
	ByType         map[string]int `json:"by_type"`
}

// GetStats 获取设备统计
func (m *DeviceManager) GetStats(identityID string) *DeviceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &DeviceStats{
		ByType: make(map[string]int),
	}

	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			stats.TotalDevices++
			if device.Status == DeviceStatusOnline {
				stats.OnlineDevices++
			} else {
				stats.OfflineDevices++
			}
			stats.ByType[device.DeviceType]++
		}
	}

	return stats
}

// === 纠偏检查 ===

// NeedsReconciliation 检查是否需要纠偏
func (m *DeviceManager) NeedsReconciliation(agentID string, centerVersion int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, exists := m.devices[agentID]
	if !exists {
		return false
	}

	// 设备版本落后于 Center
	return device.SyncVersion < centerVersion
}

// UpdateSyncVersion 更新同步版本
func (m *DeviceManager) UpdateSyncVersion(agentID string, version int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if device, exists := m.devices[agentID]; exists {
		device.SyncVersion = version
		device.LastSeen = time.Now()
	}
}

// === 设备优先级管理 (v2.8.0) ===

// SetDevicePriority 设置设备优先级
func (m *DeviceManager) SetDevicePriority(agentID string, priority int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[agentID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	if priority < 0 {
		priority = 0
	}
	if priority > 100 {
		priority = 100
	}

	device.Priority = priority
	log.Printf("Device priority updated: %s -> %d", agentID, priority)
	return nil
}

// GetHighestPriorityDevice 获取最高优先级设备
func (m *DeviceManager) GetHighestPriorityDevice(identityID string) *DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var highest *DeviceInfo
	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			if device.Status == DeviceStatusOnline {
				if highest == nil || device.Priority > highest.Priority {
					highest = device
				}
			}
		}
	}
	return highest
}

// GetDevicesByPriority 按优先级排序获取设备
func (m *DeviceManager) GetDevicesByPriority(identityID string) []*DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*DeviceInfo, 0)
	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			devices = append(devices, device)
		}
	}

	// 按优先级降序排序
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Priority > devices[j].Priority
	})

	return devices
}

// SetPrimaryDevice 设置主设备
func (m *DeviceManager) SetPrimaryDevice(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[agentID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	// 降级其他主设备
	for _, aid := range m.identityDevices[device.IdentityID] {
		if d, ok := m.devices[aid]; ok && d.TrustLevel == TrustLevelPrimary {
			d.TrustLevel = TrustLevelHigh
			d.Priority = 80
		}
	}

	// 设置新主设备
	device.TrustLevel = TrustLevelPrimary
	device.Priority = 100

	// 更新信任管理器
	if m.trustManager != nil {
		m.trustManager.SetTrustLevel(agentID, TrustLevelPrimary, "center")
	}

	log.Printf("Primary device set: %s for identity %s", agentID, device.IdentityID)
	return nil
}

// === 设备更换支持 (v2.8.0) ===

// PrepareDeviceTransfer 准备设备迁移
func (m *DeviceManager) PrepareDeviceTransfer(fromAgentID, toAgentID, identityID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证原设备
	fromDevice, exists := m.devices[fromAgentID]
	if !exists {
		return "", fmt.Errorf("source device not found")
	}

	if fromDevice.IdentityID != identityID {
		return "", fmt.Errorf("identity mismatch")
	}

	// 验证目标设备
	toDevice, exists := m.devices[toAgentID]
	if !exists {
		return "", fmt.Errorf("target device not found")
	}

	if toDevice.IdentityID != identityID {
		return "", fmt.Errorf("target device belongs to different identity")
	}

	// 生成迁移令牌
	transferToken, err := generateToken()
	if err != nil {
		return "", err
	}

	// 存储迁移请求（简化处理，实际应使用持久化存储）
	// 这里仅返回令牌，实际迁移由 TrustManager 处理

	return transferToken, nil
}

// CompleteDeviceTransfer 完成设备迁移
func (m *DeviceManager) CompleteDeviceTransfer(fromAgentID, toAgentID, identityID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证设备和身份
	fromDevice, fromExists := m.devices[fromAgentID]
	toDevice, toExists := m.devices[toAgentID]

	if !fromExists || !toExists {
		return fmt.Errorf("device not found")
	}

	if fromDevice.IdentityID != identityID || toDevice.IdentityID != identityID {
		return fmt.Errorf("identity mismatch")
	}

	switch reason {
	case "lost":
		// 设备丢失：注销原设备，提升新设备
		m.unregisterDeviceLocked(fromAgentID)
		toDevice.TrustLevel = TrustLevelPrimary
		toDevice.Priority = 100

	case "upgrade":
		// 设备升级：降级原设备，提升新设备
		fromDevice.TrustLevel = TrustLevelHigh
		fromDevice.Priority = 80
		toDevice.TrustLevel = TrustLevelPrimary
		toDevice.Priority = 100

	case "migration":
		// 主动迁移：保持原设备信任级别
		toDevice.TrustLevel = fromDevice.TrustLevel
		toDevice.Priority = fromDevice.Priority
	}

	// 更新信任管理器
	if m.trustManager != nil {
		req := &TransferSoulRequest{
			FromAgentID: fromAgentID,
			ToAgentID:   toAgentID,
			IdentityID:  identityID,
			Reason:      reason,
		}
		m.trustManager.TransferSoulToNewDevice(req)
	}

	log.Printf("Device transfer completed: %s -> %s (reason: %s)", fromAgentID, toAgentID, reason)
	return nil
}

// unregisterDeviceLocked 内部注销设备（已持有锁）
func (m *DeviceManager) unregisterDeviceLocked(agentID string) {
	device, exists := m.devices[agentID]
	if !exists {
		return
	}

	// 从身份设备列表中移除
	devices := m.identityDevices[device.IdentityID]
	var newDevices []string
	for _, aid := range devices {
		if aid != agentID {
			newDevices = append(newDevices, aid)
		}
	}
	m.identityDevices[device.IdentityID] = newDevices

	// 删除设备
	delete(m.devices, agentID)

	// 从信任管理器注销
	if m.trustManager != nil {
		m.trustManager.RevokeDevice(agentID)
	}

	log.Printf("Device unregistered: %s", agentID)
}

// === 设备信任查询 (v2.8.0) ===

// GetDevicesByTrustLevel 按信任级别获取设备
func (m *DeviceManager) GetDevicesByTrustLevel(identityID string, level TrustLevel) []*DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DeviceInfo
	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			if device.TrustLevel == level {
				result = append(result, device)
			}
		}
	}
	return result
}

// GetPrimaryDevice 获取主设备
func (m *DeviceManager) GetPrimaryDevice(identityID string) *DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agentID := range m.identityDevices[identityID] {
		if device, exists := m.devices[agentID]; exists {
			if device.TrustLevel == TrustLevelPrimary {
				return device
			}
		}
	}
	return nil
}
// Package sync - 设备信任链管理 (v2.8.0)
//
// Center 是永远在线的灵魂载体，设备信任链确保：
// - 只有授权设备能访问身份
// - 设备更换时安全迁移灵魂
// - 多设备优先级和权限管理
package sync

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// === 信任级别 ===

// TrustLevel 设备信任级别
type TrustLevel int

const (
	TrustLevelNone      TrustLevel = 0 // 未信任
	TrustLevelLow       TrustLevel = 1 // 低信任（新设备）
	TrustLevelMedium    TrustLevel = 2 // 中信任（验证过）
	TrustLevelHigh      TrustLevel = 3 // 高信任（主设备）
	TrustLevelPrimary   TrustLevel = 4 // 主设备（唯一）
)

// String 信任级别字符串
func (t TrustLevel) String() string {
	switch t {
	case TrustLevelNone:
		return "none"
	case TrustLevelLow:
		return "low"
	case TrustLevelMedium:
		return "medium"
	case TrustLevelHigh:
		return "high"
	case TrustLevelPrimary:
		return "primary"
	default:
		return "unknown"
	}
}

// ParseTrustLevel 解析信任级别
func ParseTrustLevel(s string) TrustLevel {
	switch s {
	case "low":
		return TrustLevelLow
	case "medium":
		return TrustLevelMedium
	case "high":
		return TrustLevelHigh
	case "primary":
		return TrustLevelPrimary
	default:
		return TrustLevelNone
	}
}

// === 设备授权 ===

// DeviceAuthorization 设备授权
type DeviceAuthorization struct {
	AgentID        string      `json:"agent_id"`
	IdentityID     string      `json:"identity_id"`
	TrustLevel     TrustLevel  `json:"trust_level"`
	Permissions    []string    `json:"permissions"` // read, write, sync, admin
	GrantedAt      time.Time   `json:"granted_at"`
	GrantedBy      string      `json:"granted_by"` // 授权来源设备
	ExpiresAt      *time.Time  `json:"expires_at,omitempty"`
	LastVerified   time.Time   `json:"last_verified"`
	VerificationCount int       `json:"verification_count"`
}

// IsExpired 检查授权是否过期
func (a *DeviceAuthorization) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// HasPermission 检查是否有权限
func (a *DeviceAuthorization) HasPermission(perm string) bool {
	for _, p := range a.Permissions {
		if p == perm || p == "admin" {
			return true
		}
	}
	return false
}

// === 设备认证 ===

// DeviceCredential 设备凭证
type DeviceCredential struct {
	AgentID       string    `json:"agent_id"`
	IdentityID    string    `json:"identity_id"`
	DeviceToken   string    `json:"device_token"`   // 设备令牌
	RefreshToken  string    `json:"refresh_token"`  // 刷新令牌
	SecretHash    string    `json:"secret_hash"`    // 密钥哈希
	CreatedAt     time.Time `json:"created_at"`
	LastUsed      time.Time `json:"last_used"`
	TokenExpiry   time.Time `json:"token_expiry"`
}

// === 信任管理器 ===

// TrustManager 设备信任管理器
type TrustManager struct {
	mu sync.RWMutex

	// 设备授权：agentId -> authorization
	authorizations map[string]*DeviceAuthorization

	// 设备凭证：agentId -> credential
	credentials map[string]*DeviceCredential

	// 身份-设备绑定：identityId -> []agentId
	identityBindings map[string][]string

	// 配置
	config TrustManagerConfig
}

// TrustManagerConfig 信任管理器配置
type TrustManagerConfig struct {
	// 默认信任级别
	DefaultTrustLevel TrustLevel
	// 新设备需要验证
	RequireVerification bool
	// 令牌有效期
	TokenExpiry time.Duration
	// 最大设备数
	MaxDevicesPerIdentity int
	// 主设备数量限制
	MaxPrimaryDevices int
}

// DefaultTrustManagerConfig 默认配置
func DefaultTrustManagerConfig() TrustManagerConfig {
	return TrustManagerConfig{
		DefaultTrustLevel:     TrustLevelLow,
		RequireVerification:   true,
		TokenExpiry:           30 * 24 * time.Hour, // 30天
		MaxDevicesPerIdentity: 10,
		MaxPrimaryDevices:     1,
	}
}

// NewTrustManager 创建信任管理器
func NewTrustManager(config TrustManagerConfig) *TrustManager {
	if config.DefaultTrustLevel == 0 {
		config = DefaultTrustManagerConfig()
	}

	return &TrustManager{
		authorizations:    make(map[string]*DeviceAuthorization),
		credentials:       make(map[string]*DeviceCredential),
		identityBindings:  make(map[string][]string),
		config:            config,
	}
}

// === 设备注册与认证 ===

// RegisterDeviceRequest 设备注册请求
type RegisterDeviceRequest struct {
	AgentID      string   `json:"agent_id"`
	IdentityID   string   `json:"identity_id"`
	DeviceType   string   `json:"device_type"`
	DeviceName   string   `json:"device_name"`
	Capabilities []string `json:"capabilities"`
	PublicKey    string   `json:"public_key"` // 设备公钥（可选）
}

// RegisterDeviceResponse 设备注册响应
type RegisterDeviceResponse struct {
	Success      bool              `json:"success"`
	DeviceToken  string            `json:"device_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresAt    time.Time         `json:"expires_at"`
	TrustLevel   TrustLevel        `json:"trust_level"`
	Permissions  []string          `json:"permissions"`
	Error        string            `json:"error,omitempty"`
}

// RegisterTrustedDevice 注册信任设备
func (m *TrustManager) RegisterTrustedDevice(req *RegisterDeviceRequest) (*RegisterDeviceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查设备数限制
	devices := m.identityBindings[req.IdentityID]
	if len(devices) >= m.config.MaxDevicesPerIdentity {
		return &RegisterDeviceResponse{
			Success: false,
			Error:   "max devices limit reached",
		}, nil
	}

	// 检查是否已注册
	if _, exists := m.authorizations[req.AgentID]; exists {
		return &RegisterDeviceResponse{
			Success: false,
			Error:   "device already registered",
		}, nil
	}

	// 生成令牌
	deviceToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 计算密钥哈希
	secretHash := hashSecret(deviceToken + req.AgentID)

	// 确定信任级别
	trustLevel := m.config.DefaultTrustLevel
	if len(devices) == 0 {
		// 第一个设备默认为主设备
		trustLevel = TrustLevelPrimary
	}

	// 创建授权
	now := time.Now()
	expiry := now.Add(m.config.TokenExpiry)

	auth := &DeviceAuthorization{
		AgentID:     req.AgentID,
		IdentityID:  req.IdentityID,
		TrustLevel:  trustLevel,
		Permissions: getPermissionsForTrustLevel(trustLevel),
		GrantedAt:   now,
		GrantedBy:   "center",
		LastVerified: now,
	}

	cred := &DeviceCredential{
		AgentID:      req.AgentID,
		IdentityID:   req.IdentityID,
		DeviceToken:  deviceToken,
		RefreshToken: refreshToken,
		SecretHash:   secretHash,
		CreatedAt:    now,
		LastUsed:     now,
		TokenExpiry:  expiry,
	}

	m.authorizations[req.AgentID] = auth
	m.credentials[req.AgentID] = cred
	m.identityBindings[req.IdentityID] = append(m.identityBindings[req.IdentityID], req.AgentID)

	log.Printf("Device registered: %s -> identity %s, trust=%s", req.AgentID, req.IdentityID, trustLevel)

	return &RegisterDeviceResponse{
		Success:      true,
		DeviceToken:  deviceToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiry,
		TrustLevel:   trustLevel,
		Permissions:  auth.Permissions,
	}, nil
}

// AuthenticateDevice 认证设备
func (m *TrustManager) AuthenticateDevice(agentID, deviceToken string) (*DeviceAuthorization, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cred, exists := m.credentials[agentID]
	if !exists {
		return nil, false
	}

	// 验证令牌
	if cred.DeviceToken != deviceToken {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(cred.TokenExpiry) {
		return nil, false
	}

	// 获取授权
	auth, exists := m.authorizations[agentID]
	if !exists {
		return nil, false
	}

	// 检查授权是否过期
	if auth.IsExpired() {
		return nil, false
	}

	// 更新最后使用时间
	cred.LastUsed = time.Now()
	auth.LastVerified = time.Now()
	auth.VerificationCount++

	return auth, true
}

// RefreshDeviceToken 刷新设备令牌
func (m *TrustManager) RefreshDeviceToken(agentID, refreshToken string) (*RegisterDeviceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cred, exists := m.credentials[agentID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}

	if cred.RefreshToken != refreshToken {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// 生成新令牌
	deviceToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 更新凭证
	now := time.Now()
	expiry := now.Add(m.config.TokenExpiry)

	cred.DeviceToken = deviceToken
	cred.RefreshToken = newRefreshToken
	cred.TokenExpiry = expiry
	cred.LastUsed = now

	auth := m.authorizations[agentID]

	return &RegisterDeviceResponse{
		Success:      true,
		DeviceToken:  deviceToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiry,
		TrustLevel:   auth.TrustLevel,
		Permissions:  auth.Permissions,
	}, nil
}

// === 信任级别管理 ===

// SetTrustLevel 设置设备信任级别
func (m *TrustManager) SetTrustLevel(agentID string, level TrustLevel, grantedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auth, exists := m.authorizations[agentID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	// 检查主设备数量限制
	if level == TrustLevelPrimary {
		// 先降级现有主设备
		for _, aid := range m.identityBindings[auth.IdentityID] {
			if a, ok := m.authorizations[aid]; ok && a.TrustLevel == TrustLevelPrimary && aid != agentID {
				a.TrustLevel = TrustLevelHigh
				a.Permissions = getPermissionsForTrustLevel(TrustLevelHigh)
			}
		}
	}

	auth.TrustLevel = level
	auth.Permissions = getPermissionsForTrustLevel(level)
	auth.GrantedBy = grantedBy
	auth.LastVerified = time.Now()

	log.Printf("Trust level updated: %s -> %s (by %s)", agentID, level, grantedBy)
	return nil
}

// GetTrustLevel 获取设备信任级别
func (m *TrustManager) GetTrustLevel(agentID string) TrustLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if auth, exists := m.authorizations[agentID]; exists {
		return auth.TrustLevel
	}
	return TrustLevelNone
}

// GetAuthorization 获取设备授权
func (m *TrustManager) GetAuthorization(agentID string) *DeviceAuthorization {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.authorizations[agentID]
}

// === 设备更换 ===

// TransferSoulRequest 灵魂迁移请求
type TransferSoulRequest struct {
	FromAgentID   string `json:"from_agent_id"`
	ToAgentID     string `json:"to_agent_id"`
	IdentityID    string `json:"identity_id"`
	TransferToken string `json:"transfer_token"` // 迁移令牌（由原设备生成）
	Reason        string `json:"reason"`         // lost, upgrade, migration
}

// TransferSoulResponse 灵魂迁移响应
type TransferSoulResponse struct {
	Success   bool   `json:"success"`
	NewToken  string `json:"new_token,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TransferSoulToNewDevice 将灵魂迁移到新设备
//
// 场景1: 设备丢失 - 需要验证身份后迁移
// 场景2: 设备升级 - 需要原设备确认
func (m *TrustManager) TransferSoulToNewDevice(req *TransferSoulRequest) (*TransferSoulResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证原设备
	fromAuth, fromExists := m.authorizations[req.FromAgentID]
	if !fromExists {
		return &TransferSoulResponse{
			Success: false,
			Error:   "source device not found",
		}, nil
	}

	// 验证身份匹配
	if fromAuth.IdentityID != req.IdentityID {
		return &TransferSoulResponse{
			Success: false,
			Error:   "identity mismatch",
		}, nil
	}

	// 验证目标设备已注册
	toAuth, toExists := m.authorizations[req.ToAgentID]
	if !toExists {
		return &TransferSoulResponse{
			Success: false,
			Error:   "target device not registered",
		}, nil
	}

	// 验证目标设备属于同一身份
	if toAuth.IdentityID != req.IdentityID {
		return &TransferSoulResponse{
			Success: false,
			Error:   "target device belongs to different identity",
		}, nil
	}

	// 根据迁移原因处理
	switch req.Reason {
	case "lost":
		// 设备丢失：直接迁移，提升新设备信任级别
		m.revokeDevice(req.FromAgentID)
		m.SetTrustLevel(req.ToAgentID, TrustLevelPrimary, "transfer:lost")

	case "upgrade":
		// 设备升级：验证迁移令牌
		// 简化处理，实际需要更复杂的验证
		m.SetTrustLevel(req.ToAgentID, TrustLevelPrimary, "transfer:upgrade")
		m.SetTrustLevel(req.FromAgentID, TrustLevelHigh, "transfer:upgrade")

	case "migration":
		// 主动迁移：保持原设备信任级别
		m.SetTrustLevel(req.ToAgentID, fromAuth.TrustLevel, "transfer:migration")
	}

	// 生成新令牌
	newToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	// 更新目标设备凭证
	cred := m.credentials[req.ToAgentID]
	cred.DeviceToken = newToken
	cred.LastUsed = time.Now()

	log.Printf("Soul transferred: %s -> %s (reason: %s)", req.FromAgentID, req.ToAgentID, req.Reason)

	return &TransferSoulResponse{
		Success:  true,
		NewToken: newToken,
	}, nil
}

// === 设备注销 ===

// RevokeDevice 注销设备
func (m *TrustManager) RevokeDevice(agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.revokeDevice(agentID)
}

func (m *TrustManager) revokeDevice(agentID string) error {
	auth, exists := m.authorizations[agentID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	// 从身份绑定中移除
	devices := m.identityBindings[auth.IdentityID]
	var newDevices []string
	for _, aid := range devices {
		if aid != agentID {
			newDevices = append(newDevices, aid)
		}
	}
	m.identityBindings[auth.IdentityID] = newDevices

	// 删除授权和凭证
	delete(m.authorizations, agentID)
	delete(m.credentials, agentID)

	log.Printf("Device revoked: %s", agentID)
	return nil
}

// === 查询 ===

// GetDevicesByIdentity 获取身份的所有设备
func (m *TrustManager) GetDevicesByIdentity(identityID string) []*DeviceAuthorization {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agentIDs := m.identityBindings[identityID]
	devices := make([]*DeviceAuthorization, 0, len(agentIDs))

	for _, agentID := range agentIDs {
		if auth, exists := m.authorizations[agentID]; exists {
			devices = append(devices, auth)
		}
	}

	return devices
}

// GetPrimaryDevice 获取主设备
func (m *TrustManager) GetPrimaryDevice(identityID string) *DeviceAuthorization {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agentID := range m.identityBindings[identityID] {
		if auth, exists := m.authorizations[agentID]; exists {
			if auth.TrustLevel == TrustLevelPrimary {
				return auth
			}
		}
	}
	return nil
}

// === 辅助函数 ===

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

func getPermissionsForTrustLevel(level TrustLevel) []string {
	switch level {
	case TrustLevelPrimary:
		return []string{"admin", "read", "write", "sync", "transfer"}
	case TrustLevelHigh:
		return []string{"read", "write", "sync"}
	case TrustLevelMedium:
		return []string{"read", "sync"}
	case TrustLevelLow:
		return []string{"read"}
	default:
		return []string{}
	}
}

// === 序列化 ===

// ToJSON 序列化授权
func (a *DeviceAuthorization) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON 反序列化授权
func (a *DeviceAuthorization) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}
package sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// KeyType 密钥类型
type KeyType string

const (
	KeyTypeMaster   KeyType = "master"   // 主密钥
	KeyTypeSession  KeyType = "session"  // 会话密钥
	KeyTypeDevice   KeyType = "device"   // 设备密钥
	KeyTypeIdentity KeyType = "identity" // 身份密钥
)

// EncryptionAlgorithm 加密算法
type EncryptionAlgorithm string

const (
	EncryptionAES256GCM EncryptionAlgorithm = "AES-256-GCM"
	EncryptionAES256CBC EncryptionAlgorithm = "AES-256-CBC"
	EncryptionChaCha20  EncryptionAlgorithm = "ChaCha20-Poly1305"
)

// SecurityLevel 安全级别
type SecurityLevel string

const (
	SecurityLevelNone    SecurityLevel = "none"
	SecurityLevelBasic   SecurityLevel = "basic"
	SecurityLevelE2E     SecurityLevel = "e2e"     // 端到端加密
	SecurityLevelE2EAuth SecurityLevel = "e2e_auth" // 端到端加密+认证
)

// SecurityKey 安全密钥
type SecurityKey struct {
	KeyID       string            `json:"key_id"`
	KeyType     KeyType           `json:"key_type"`
	Algorithm   EncryptionAlgorithm `json:"algorithm"`
	KeyData     []byte            `json:"-"` // 不序列化
	KeyDataB64  string            `json:"key_data,omitempty"`
	IdentityID  string            `json:"identity_id,omitempty"`
	AgentID     string            `json:"agent_id,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	IsActive    bool              `json:"is_active"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// EncryptedData 加密数据
type EncryptedData struct {
	DataID      string             `json:"data_id"`
	Algorithm   EncryptionAlgorithm `json:"algorithm"`
	KeyID       string             `json:"key_id"`
	IV          string             `json:"iv"`           // 初始化向量
	Ciphertext  string             `json:"ciphertext"`   // 密文
	Tag         string             `json:"tag,omitempty"` // 认证标签 (GCM)
	Timestamp   time.Time          `json:"timestamp"`
	DataType    string             `json:"data_type"`
	Checksum    string             `json:"checksum"`
}

// SecuritySession 安全会话
type SecuritySession struct {
	SessionID     string        `json:"session_id"`
	IdentityID    string        `json:"identity_id"`
	AgentID       string        `json:"agent_id"`
	SessionKey    *SecurityKey  `json:"-"`
	SessionKeyID  string        `json:"session_key_id"`
	SecurityLevel SecurityLevel `json:"security_level"`
	CreatedAt     time.Time     `json:"created_at"`
	ExpiresAt     time.Time     `json:"expires_at"`
	LastActiveAt  time.Time     `json:"last_active_at"`
	IsActive      bool          `json:"is_active"`
	DeviceInfo    map[string]string `json:"device_info,omitempty"`
}

// SecureChannel 安全通道
type SecureChannel struct {
	ChannelID    string        `json:"channel_id"`
	IdentityID   string        `json:"identity_id"`
	SourceAgent  string        `json:"source_agent"`
	TargetAgent  string        `json:"target_agent"`
	ChannelKey   *SecurityKey  `json:"-"`
	ChannelKeyID string        `json:"channel_key_id"`
	SecurityLevel SecurityLevel `json:"security_level"`
	CreatedAt    time.Time     `json:"created_at"`
	ExpiresAt    *time.Time    `json:"expires_at,omitempty"`
	IsActive     bool          `json:"is_active"`
	MessageCount int           `json:"message_count"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	DefaultAlgorithm    EncryptionAlgorithm `json:"default_algorithm"`
	SessionKeyTTL       time.Duration       `json:"session_key_ttl"`
	ChannelKeyTTL       time.Duration       `json:"channel_key_ttl"`
	MaxKeyAge           time.Duration       `json:"max_key_age"`
	KeyRotationInterval time.Duration       `json:"key_rotation_interval"`
	DefaultSecurityLevel SecurityLevel      `json:"default_security_level"`
	EnableKeyRotation   bool                `json:"enable_key_rotation"`
	MaxSessionsPerAgent int                 `json:"max_sessions_per_agent"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		DefaultAlgorithm:     EncryptionAES256GCM,
		SessionKeyTTL:        24 * time.Hour,
		ChannelKeyTTL:        1 * time.Hour,
		MaxKeyAge:            30 * 24 * time.Hour,
		KeyRotationInterval:  7 * 24 * time.Hour,
		DefaultSecurityLevel: SecurityLevelE2E,
		EnableKeyRotation:    true,
		MaxSessionsPerAgent:  5,
	}
}

// SecurityService 安全服务
type SecurityService struct {
	mu             sync.RWMutex
	keys           map[string]*SecurityKey      // keyID -> key
	sessions       map[string]*SecuritySession  // sessionID -> session
	channels       map[string]*SecureChannel    // channelID -> channel
	agentSessions  map[string][]string          // agentID -> []sessionID
	identityKeys   map[string][]string          // identityID -> []keyID
	config         SecurityConfig
	masterKey      *SecurityKey
	listeners      []SecurityListener
}

// SecurityListener 安全事件监听器
type SecurityListener interface {
	OnKeyCreated(key *SecurityKey)
	OnKeyRotated(oldKeyID, newKeyID string)
	OnSessionCreated(session *SecuritySession)
	OnSessionExpired(sessionID string)
	OnChannelEstablished(channel *SecureChannel)
	OnChannelClosed(channelID string)
	OnEncryptionSuccess(dataID, keyID string)
	OnDecryptionSuccess(dataID, keyID string)
	OnSecurityViolation(violationType, details string)
}

// NewSecurityService 创建安全服务
func NewSecurityService(config SecurityConfig) *SecurityService {
	service := &SecurityService{
		keys:          make(map[string]*SecurityKey),
		sessions:      make(map[string]*SecuritySession),
		channels:      make(map[string]*SecureChannel),
		agentSessions: make(map[string][]string),
		identityKeys:  make(map[string][]string),
		config:        config,
		listeners:     make([]SecurityListener, 0),
	}

	// 生成主密钥
	service.masterKey = service.generateKey(KeyTypeMaster, "")

	return service
}

// AddListener 添加监听器
func (s *SecurityService) AddListener(listener SecurityListener) {
	s.mu.Lock()
	s.listeners = append(s.listeners, listener)
	s.mu.Unlock()
}

// === 密钥管理 ===

// generateKey 生成密钥
func (s *SecurityService) generateKey(keyType KeyType, identityID string) *SecurityKey {
	keyID := generateKeyID()
	keyData := make([]byte, 32) // AES-256

	if _, err := rand.Read(keyData); err != nil {
		// 回退到时间戳
		h := sha256.New()
		h.Write([]byte(keyID + time.Now().String()))
		keyData = h.Sum(nil)[:32]
	}

	key := &SecurityKey{
		KeyID:      keyID,
		KeyType:    keyType,
		Algorithm:  s.config.DefaultAlgorithm,
		KeyData:    keyData,
		KeyDataB64: base64.StdEncoding.EncodeToString(keyData),
		IdentityID: identityID,
		CreatedAt:  time.Now(),
		IsActive:   true,
		Metadata:   make(map[string]string),
	}

	s.mu.Lock()
	s.keys[keyID] = key
	if identityID != "" {
		s.identityKeys[identityID] = append(s.identityKeys[identityID], keyID)
	}
	s.mu.Unlock()

	s.notifyKeyCreated(key)

	return key
}

// CreateSessionKey 创建会话密钥
func (s *SecurityService) CreateSessionKey(identityID, agentID string) *SecurityKey {
	key := s.generateKey(KeyTypeSession, identityID)
	key.AgentID = agentID

	expiresAt := time.Now().Add(s.config.SessionKeyTTL)
	key.ExpiresAt = &expiresAt

	return key
}

// CreateDeviceKey 创建设备密钥
func (s *SecurityService) CreateDeviceKey(identityID, agentID string) *SecurityKey {
	key := s.generateKey(KeyTypeDevice, identityID)
	key.AgentID = agentID

	return key
}

// CreateIdentityKey 创建身份密钥
func (s *SecurityService) CreateIdentityKey(identityID string) *SecurityKey {
	return s.generateKey(KeyTypeIdentity, identityID)
}

// GetKey 获取密钥
func (s *SecurityService) GetKey(keyID string) *SecurityKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keys[keyID]
}

// GetActiveKey 获取活动密钥
func (s *SecurityService) GetActiveKey(identityID string, keyType KeyType) *SecurityKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keyIDs := s.identityKeys[identityID]
	for i := len(keyIDs) - 1; i >= 0; i-- {
		key := s.keys[keyIDs[i]]
		if key != nil && key.IsActive && key.KeyType == keyType {
			// 检查是否过期
			if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
				continue
			}
			return key
		}
	}
	return nil
}

// RotateKey 轮换密钥
func (s *SecurityService) RotateKey(oldKeyID string) (*SecurityKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldKey := s.keys[oldKeyID]
	if oldKey == nil {
		return nil, fmt.Errorf("key not found: %s", oldKeyID)
	}

	// 生成新密钥
	newKey := s.generateKey(oldKey.KeyType, oldKey.IdentityID)
	newKey.AgentID = oldKey.AgentID

	// 标记旧密钥为非活动
	oldKey.IsActive = false

	s.notifyKeyRotated(oldKeyID, newKey.KeyID)

	return newKey, nil
}

// === 加密/解密 ===

// Encrypt 加密数据
func (s *SecurityService) Encrypt(plaintext []byte, keyID string) (*EncryptedData, error) {
	s.mu.RLock()
	key := s.keys[keyID]
	s.mu.RUnlock()

	if key == nil || !key.IsActive {
		return nil, fmt.Errorf("key not found or inactive: %s", keyID)
	}

	switch key.Algorithm {
	case EncryptionAES256GCM:
		return s.encryptAESGCM(plaintext, key)
	case EncryptionAES256CBC:
		return s.encryptAESCBC(plaintext, key)
	default:
		return s.encryptAESGCM(plaintext, key)
	}
}

// encryptAESGCM AES-GCM 加密
func (s *SecurityService) encryptAESGCM(plaintext []byte, key *SecurityKey) (*EncryptedData, error) {
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 提取认证标签
	tagSize := 16
	tag := ""
	if len(ciphertext) > tagSize {
		tag = base64.StdEncoding.EncodeToString(ciphertext[len(ciphertext)-tagSize:])
		ciphertext = ciphertext[:len(ciphertext)-tagSize]
	}

	data := &EncryptedData{
		DataID:     generateDataID(),
		Algorithm:  EncryptionAES256GCM,
		KeyID:      key.KeyID,
		IV:         base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Tag:        tag,
		Timestamp:  time.Now(),
		Checksum:   calculateDataChecksum(plaintext),
	}

	s.notifyEncryptionSuccess(data.DataID, key.KeyID)

	return data, nil
}

// encryptAESCBC AES-CBC 加密
func (s *SecurityService) encryptAESCBC(plaintext []byte, key *SecurityKey) (*EncryptedData, error) {
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// PKCS7 填充
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padded := make([]byte, len(plaintext)+padding)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	// 生成 IV
	iv := make([]byte, blockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// 加密
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	data := &EncryptedData{
		DataID:     generateDataID(),
		Algorithm:  EncryptionAES256CBC,
		KeyID:      key.KeyID,
		IV:         base64.StdEncoding.EncodeToString(iv),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Timestamp:  time.Now(),
		Checksum:   calculateDataChecksum(plaintext),
	}

	s.notifyEncryptionSuccess(data.DataID, key.KeyID)

	return data, nil
}

// Decrypt 解密数据
func (s *SecurityService) Decrypt(data *EncryptedData) ([]byte, error) {
	s.mu.RLock()
	key := s.keys[data.KeyID]
	s.mu.RUnlock()

	if key == nil || !key.IsActive {
		return nil, fmt.Errorf("key not found or inactive: %s", data.KeyID)
	}

	switch data.Algorithm {
	case EncryptionAES256GCM:
		return s.decryptAESGCM(data, key)
	case EncryptionAES256CBC:
		return s.decryptAESCBC(data, key)
	default:
		return s.decryptAESGCM(data, key)
	}
}

// decryptAESGCM AES-GCM 解密
func (s *SecurityService) decryptAESGCM(data *EncryptedData, key *SecurityKey) ([]byte, error) {
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 解码
	nonce, err := base64.StdEncoding.DecodeString(data.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(data.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// 重新组合密文和标签
	if data.Tag != "" {
		tag, err := base64.StdEncoding.DecodeString(data.Tag)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tag: %w", err)
		}
		ciphertext = append(ciphertext, tag...)
	}

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	s.notifyDecryptionSuccess(data.DataID, key.KeyID)

	return plaintext, nil
}

// decryptAESCBC AES-CBC 解密
func (s *SecurityService) decryptAESCBC(data *EncryptedData, key *SecurityKey) ([]byte, error) {
	block, err := aes.NewCipher(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 解码
	iv, err := base64.StdEncoding.DecodeString(data.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(data.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// 解密
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// 移除 PKCS7 填充
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) || padding > block.BlockSize() {
		return nil, fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padding]

	s.notifyDecryptionSuccess(data.DataID, key.KeyID)

	return plaintext, nil
}

// === 会话管理 ===

// CreateSession 创建安全会话
func (s *SecurityService) CreateSession(identityID, agentID string, level SecurityLevel) (*SecuritySession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查会话数量限制
	sessions := s.agentSessions[agentID]
	if len(sessions) >= s.config.MaxSessionsPerAgent {
		// 清理过期会话
		s.cleanupExpiredSessionsLocked(agentID)
	}

	// 创建会话密钥
	sessionKey := s.generateKey(KeyTypeSession, identityID)
	sessionKey.AgentID = agentID

	session := &SecuritySession{
		SessionID:     generateSessionID(),
		IdentityID:    identityID,
		AgentID:       agentID,
		SessionKey:    sessionKey,
		SessionKeyID:  sessionKey.KeyID,
		SecurityLevel: level,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(s.config.SessionKeyTTL),
		LastActiveAt:  time.Now(),
		IsActive:      true,
		DeviceInfo:    make(map[string]string),
	}

	s.sessions[session.SessionID] = session
	s.agentSessions[agentID] = append(s.agentSessions[agentID], session.SessionID)

	s.notifySessionCreated(session)

	return session, nil
}

// GetSession 获取会话
func (s *SecurityService) GetSession(sessionID string) *SecuritySession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// ValidateSession 验证会话
func (s *SecurityService) ValidateSession(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[sessionID]
	if session == nil || !session.IsActive {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		return false
	}

	return true
}

// RefreshSession 刷新会话
func (s *SecurityService) RefreshSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[sessionID]
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.LastActiveAt = time.Now()
	session.ExpiresAt = time.Now().Add(s.config.SessionKeyTTL)

	return nil
}

// CloseSession 关闭会话
func (s *SecurityService) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[sessionID]
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.IsActive = false

	// 从代理会话列表移除
	sessions := s.agentSessions[session.AgentID]
	for i, id := range sessions {
		if id == sessionID {
			s.agentSessions[session.AgentID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}

	return nil
}

func (s *SecurityService) cleanupExpiredSessionsLocked(agentID string) {
	sessions := s.agentSessions[agentID]
	active := make([]string, 0)

	for _, sessionID := range sessions {
		session := s.sessions[sessionID]
		if session != nil && session.IsActive && time.Now().Before(session.ExpiresAt) {
			active = append(active, sessionID)
		} else if session != nil {
			session.IsActive = false
		}
	}

	s.agentSessions[agentID] = active
}

// === 安全通道 ===

// EstablishChannel 建立安全通道
func (s *SecurityService) EstablishChannel(identityID, sourceAgent, targetAgent string) (*SecureChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成通道密钥
	channelKey := s.generateKey(KeyTypeSession, identityID)

	channel := &SecureChannel{
		ChannelID:     generateChannelID(),
		IdentityID:    identityID,
		SourceAgent:   sourceAgent,
		TargetAgent:   targetAgent,
		ChannelKey:    channelKey,
		ChannelKeyID:  channelKey.KeyID,
		SecurityLevel: s.config.DefaultSecurityLevel,
		CreatedAt:     time.Now(),
		IsActive:      true,
		MessageCount:  0,
	}

	expiresAt := time.Now().Add(s.config.ChannelKeyTTL)
	channel.ExpiresAt = &expiresAt

	s.channels[channel.ChannelID] = channel

	s.notifyChannelEstablished(channel)

	return channel, nil
}

// GetChannel 获取通道
func (s *SecurityService) GetChannel(channelID string) *SecureChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.channels[channelID]
}

// CloseChannel 关闭通道
func (s *SecurityService) CloseChannel(channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel := s.channels[channelID]
	if channel == nil {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	channel.IsActive = false

	s.notifyChannelClosed(channelID)

	return nil
}

// EncryptForChannel 通过通道加密
func (s *SecurityService) EncryptForChannel(channelID string, plaintext []byte) (*EncryptedData, error) {
	s.mu.RLock()
	channel := s.channels[channelID]
	s.mu.RUnlock()

	if channel == nil || !channel.IsActive {
		return nil, fmt.Errorf("channel not found or inactive: %s", channelID)
	}

	if channel.ExpiresAt != nil && time.Now().After(*channel.ExpiresAt) {
		return nil, fmt.Errorf("channel expired: %s", channelID)
	}

	data, err := s.Encrypt(plaintext, channel.ChannelKeyID)
	if err != nil {
		return nil, err
	}

	// 更新消息计数
	s.mu.Lock()
	if ch := s.channels[channelID]; ch != nil {
		ch.MessageCount++
	}
	s.mu.Unlock()

	return data, nil
}

// === 安全验证 ===

// ValidateEncryption 验证加密
func (s *SecurityService) ValidateEncryption(data *EncryptedData) error {
	if data == nil {
		return fmt.Errorf("encrypted data is nil")
	}

	if data.KeyID == "" {
		return fmt.Errorf("key ID is empty")
	}

	if data.Ciphertext == "" {
		return fmt.Errorf("ciphertext is empty")
	}

	if data.IV == "" {
		return fmt.Errorf("IV is empty")
	}

	return nil
}

// VerifyChecksum 验证校验和
func (s *SecurityService) VerifyChecksum(plaintext []byte, expectedChecksum string) bool {
	return calculateDataChecksum(plaintext) == expectedChecksum
}

// === 统计信息 ===

// GetSecurityStats 获取安全统计
func (s *SecurityService) GetSecurityStats(identityID string) *SecurityStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SecurityStats{
		IdentityID:     identityID,
		TotalKeys:      len(s.identityKeys[identityID]),
		ActiveKeys:     0,
		ActiveSessions: 0,
		ActiveChannels: 0,
	}

	for _, keyID := range s.identityKeys[identityID] {
		if key := s.keys[keyID]; key != nil && key.IsActive {
			stats.ActiveKeys++
		}
	}

	for _, session := range s.sessions {
		if session.IdentityID == identityID && session.IsActive {
			stats.ActiveSessions++
		}
	}

	for _, channel := range s.channels {
		if channel.IdentityID == identityID && channel.IsActive {
			stats.ActiveChannels++
		}
	}

	return stats
}

// SecurityStats 安全统计
type SecurityStats struct {
	IdentityID     string `json:"identity_id"`
	TotalKeys      int    `json:"total_keys"`
	ActiveKeys     int    `json:"active_keys"`
	ActiveSessions int    `json:"active_sessions"`
	ActiveChannels int    `json:"active_channels"`
}

// === JSON 序列化 ===

// ToJSON 序列化密钥 (不含敏感数据)
func (k *SecurityKey) ToJSON() ([]byte, error) {
	// 创建副本，不包含 KeyData
	copy := &SecurityKey{
		KeyID:      k.KeyID,
		KeyType:    k.KeyType,
		Algorithm:  k.Algorithm,
		IdentityID: k.IdentityID,
		AgentID:    k.AgentID,
		CreatedAt:  k.CreatedAt,
		ExpiresAt:  k.ExpiresAt,
		IsActive:   k.IsActive,
		Metadata:   k.Metadata,
	}
	return json.Marshal(copy)
}

// ToJSON 序列化加密数据
func (d *EncryptedData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// EncryptedDataFromJSON 从 JSON 解析
func EncryptedDataFromJSON(data []byte) (*EncryptedData, error) {
	ed := &EncryptedData{}
	if err := json.Unmarshal(data, ed); err != nil {
		return nil, err
	}
	return ed, nil
}

// === 辅助方法 ===

func generateKeyID() string {
	return fmt.Sprintf("key-%d-%x", time.Now().UnixNano(), randBytes(4))
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d-%x", time.Now().UnixNano(), randBytes(4))
}

func generateChannelID() string {
	return fmt.Sprintf("channel-%d-%x", time.Now().UnixNano(), randBytes(4))
}

func generateDataID() string {
	return fmt.Sprintf("data-%d-%x", time.Now().UnixNano(), randBytes(4))
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func calculateDataChecksum(data []byte) string {
	h := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(h[:8])
}

// === 监听器通知 ===

func (s *SecurityService) notifyKeyCreated(key *SecurityKey) {
	for _, l := range s.listeners {
		l.OnKeyCreated(key)
	}
}

func (s *SecurityService) notifyKeyRotated(oldKeyID, newKeyID string) {
	for _, l := range s.listeners {
		l.OnKeyRotated(oldKeyID, newKeyID)
	}
}

func (s *SecurityService) notifySessionCreated(session *SecuritySession) {
	for _, l := range s.listeners {
		l.OnSessionCreated(session)
	}
}

func (s *SecurityService) notifySessionExpired(sessionID string) {
	for _, l := range s.listeners {
		l.OnSessionExpired(sessionID)
	}
}

func (s *SecurityService) notifyChannelEstablished(channel *SecureChannel) {
	for _, l := range s.listeners {
		l.OnChannelEstablished(channel)
	}
}

func (s *SecurityService) notifyChannelClosed(channelID string) {
	for _, l := range s.listeners {
		l.OnChannelClosed(channelID)
	}
}

func (s *SecurityService) notifyEncryptionSuccess(dataID, keyID string) {
	for _, l := range s.listeners {
		l.OnEncryptionSuccess(dataID, keyID)
	}
}

func (s *SecurityService) notifyDecryptionSuccess(dataID, keyID string) {
	for _, l := range s.listeners {
		l.OnDecryptionSuccess(dataID, keyID)
	}
}

func (s *SecurityService) notifySecurityViolation(violationType, details string) {
	for _, l := range s.listeners {
		l.OnSecurityViolation(violationType, details)
	}
}
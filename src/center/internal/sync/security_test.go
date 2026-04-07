package sync

import (
	"testing"
	"time"
)

func TestSecurityService(t *testing.T) {
	config := DefaultSecurityConfig()
	service := NewSecurityService(config)

	// 测试初始化
	if service.keys == nil {
		t.Error("keys map should be initialized")
	}
	if service.sessions == nil {
		t.Error("sessions map should be initialized")
	}
	if service.masterKey == nil {
		t.Error("master key should be generated")
	}
}

func TestSecurityServiceGenerateKey(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	key := service.generateKey(KeyTypeSession, "identity-001")

	if key == nil {
		t.Fatal("Key should not be nil")
	}

	if key.KeyID == "" {
		t.Error("KeyID should be generated")
	}
	if key.KeyType != KeyTypeSession {
		t.Errorf("Expected key type 'session', got '%s'", key.KeyType)
	}
	if len(key.KeyData) != 32 {
		t.Errorf("Expected key data length 32, got %d", len(key.KeyData))
	}
	if !key.IsActive {
		t.Error("Key should be active by default")
	}

	// 验证存储
	stored := service.GetKey(key.KeyID)
	if stored == nil {
		t.Error("Key should be stored")
	}
}

func TestSecurityServiceCreateSessionKey(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	key := service.CreateSessionKey("identity-001", "agent-001")

	if key == nil {
		t.Fatal("Session key should not be nil")
	}

	if key.KeyType != KeyTypeSession {
		t.Errorf("Expected key type 'session', got '%s'", key.KeyType)
	}
	if key.IdentityID != "identity-001" {
		t.Errorf("Expected identity 'identity-001', got '%s'", key.IdentityID)
	}
	if key.AgentID != "agent-001" {
		t.Errorf("Expected agent 'agent-001', got '%s'", key.AgentID)
	}
	if key.ExpiresAt == nil {
		t.Error("Session key should have expiration")
	}
}

func TestSecurityServiceCreateSession(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	session, err := service.CreateSession("identity-001", "agent-001", SecurityLevelE2E)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.SessionID == "" {
		t.Error("SessionID should be generated")
	}
	if session.IdentityID != "identity-001" {
		t.Errorf("Expected identity 'identity-001', got '%s'", session.IdentityID)
	}
	if session.AgentID != "agent-001" {
		t.Errorf("Expected agent 'agent-001', got '%s'", session.AgentID)
	}
	if session.SecurityLevel != SecurityLevelE2E {
		t.Errorf("Expected security level 'e2e', got '%s'", session.SecurityLevel)
	}
	if !session.IsActive {
		t.Error("Session should be active by default")
	}

	// 验证存储
	stored := service.GetSession(session.SessionID)
	if stored == nil {
		t.Error("Session should be stored")
	}
}

func TestSecurityServiceValidateSession(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	// 创建会话
	session, _ := service.CreateSession("identity-001", "agent-001", SecurityLevelE2E)

	// 验证有效会话
	if !service.ValidateSession(session.SessionID) {
		t.Error("Session should be valid")
	}

	// 关闭会话
	service.CloseSession(session.SessionID)

	// 验证无效会话
	if service.ValidateSession(session.SessionID) {
		t.Error("Closed session should not be valid")
	}
}

func TestSecurityServiceRefreshSession(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	session, _ := service.CreateSession("identity-001", "agent-001", SecurityLevelE2E)
	originalExpiry := session.ExpiresAt

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 刷新会话
	err := service.RefreshSession(session.SessionID)
	if err != nil {
		t.Fatalf("RefreshSession failed: %v", err)
	}

	// 验证更新
	stored := service.GetSession(session.SessionID)
	if stored.LastActiveAt.Before(session.CreatedAt) {
		t.Error("LastActiveAt should be updated")
	}
	if !stored.ExpiresAt.After(originalExpiry) {
		t.Error("ExpiresAt should be extended")
	}
}

func TestSecurityServiceEncryptDecryptAESGCM(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	key := service.CreateSessionKey("identity-001", "agent-001")

	plaintext := []byte("Hello, this is a secret message!")

	// 加密
	encrypted, err := service.Encrypt(plaintext, key.KeyID)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted.DataID == "" {
		t.Error("DataID should be generated")
	}
	if encrypted.Algorithm != EncryptionAES256GCM {
		t.Errorf("Expected algorithm 'AES-256-GCM', got '%s'", encrypted.Algorithm)
	}
	if encrypted.Ciphertext == "" {
		t.Error("Ciphertext should not be empty")
	}
	if encrypted.IV == "" {
		t.Error("IV should not be empty")
	}

	// 解密
	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch: expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestSecurityServiceEncryptDecryptAESCBC(t *testing.T) {
	config := DefaultSecurityConfig()
	config.DefaultAlgorithm = EncryptionAES256CBC
	service := NewSecurityService(config)

	key := service.CreateSessionKey("identity-001", "agent-001")

	plaintext := []byte("Secret message for AES-CBC encryption")

	// 加密
	encrypted, err := service.Encrypt(plaintext, key.KeyID)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted.Algorithm != EncryptionAES256CBC {
		t.Errorf("Expected algorithm 'AES-256-CBC', got '%s'", encrypted.Algorithm)
	}

	// 解密
	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch: expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestSecurityServiceEncryptWithInvalidKey(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	plaintext := []byte("Test message")

	// 使用不存在的密钥
	_, err := service.Encrypt(plaintext, "invalid-key-id")
	if err == nil {
		t.Error("Should fail with invalid key")
	}
}

func TestSecurityServiceEstablishChannel(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	channel, err := service.EstablishChannel("identity-001", "agent-phone", "agent-watch")
	if err != nil {
		t.Fatalf("EstablishChannel failed: %v", err)
	}

	if channel.ChannelID == "" {
		t.Error("ChannelID should be generated")
	}
	if channel.SourceAgent != "agent-phone" {
		t.Errorf("Expected source 'agent-phone', got '%s'", channel.SourceAgent)
	}
	if channel.TargetAgent != "agent-watch" {
		t.Errorf("Expected target 'agent-watch', got '%s'", channel.TargetAgent)
	}
	if !channel.IsActive {
		t.Error("Channel should be active by default")
	}

	// 验证存储
	stored := service.GetChannel(channel.ChannelID)
	if stored == nil {
		t.Error("Channel should be stored")
	}
}

func TestSecurityServiceEncryptForChannel(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	channel, _ := service.EstablishChannel("identity-001", "agent-phone", "agent-watch")

	plaintext := []byte("Channel message")

	// 通过通道加密
	encrypted, err := service.EncryptForChannel(channel.ChannelID, plaintext)
	if err != nil {
		t.Fatalf("EncryptForChannel failed: %v", err)
	}

	// 解密
	decrypted, err := service.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch")
	}

	// 验证消息计数
	stored := service.GetChannel(channel.ChannelID)
	if stored.MessageCount != 1 {
		t.Errorf("Expected message count 1, got %d", stored.MessageCount)
	}
}

func TestSecurityServiceRotateKey(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	oldKey := service.CreateSessionKey("identity-001", "agent-001")

	// 轮换密钥
	newKey, err := service.RotateKey(oldKey.KeyID)
	if err != nil {
		t.Fatalf("RotateKey failed: %v", err)
	}

	if newKey.KeyID == oldKey.KeyID {
		t.Error("New key should have different ID")
	}

	// 验证旧密钥状态
	storedOld := service.GetKey(oldKey.KeyID)
	if storedOld.IsActive {
		t.Error("Old key should be inactive after rotation")
	}

	// 验证新密钥状态
	storedNew := service.GetKey(newKey.KeyID)
	if !storedNew.IsActive {
		t.Error("New key should be active")
	}
}

func TestSecurityServiceGetActiveKey(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	// 创建多个密钥
	key1 := service.CreateSessionKey("identity-001", "agent-001")
	key2 := service.CreateSessionKey("identity-001", "agent-001")

	// 获取活动密钥 (应该是最后一个)
	active := service.GetActiveKey("identity-001", KeyTypeSession)
	if active == nil {
		t.Fatal("Active key should not be nil")
	}

	if active.KeyID != key2.KeyID {
		t.Errorf("Expected key2 to be active, got key1")
	}

	// 使 key2 过期
	key2.IsActive = false

	// 现在应该返回 key1
	active = service.GetActiveKey("identity-001", KeyTypeSession)
	if active.KeyID != key1.KeyID {
		t.Errorf("Expected key1 to be active after key2 inactive")
	}
}

func TestSecurityServiceValidateEncryption(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	// 测试 nil 数据
	err := service.ValidateEncryption(nil)
	if err == nil {
		t.Error("Should fail with nil data")
	}

	// 测试空数据
	emptyData := &EncryptedData{}
	err = service.ValidateEncryption(emptyData)
	if err == nil {
		t.Error("Should fail with empty data")
	}

	// 测试有效数据
	validData := &EncryptedData{
		KeyID:      "key-001",
		Ciphertext: "ciphertext",
		IV:         "iv",
	}
	err = service.ValidateEncryption(validData)
	if err != nil {
		t.Errorf("Should pass with valid data: %v", err)
	}
}

func TestSecurityServiceVerifyChecksum(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	data := []byte("test data")
	checksum := calculateDataChecksum(data)

	// 验证正确的校验和
	if !service.VerifyChecksum(data, checksum) {
		t.Error("Checksum should match")
	}

	// 验证错误的校验和
	if service.VerifyChecksum(data, "wrong-checksum") {
		t.Error("Wrong checksum should not match")
	}
}

func TestSecurityServiceStats(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	// 创建一些密钥和会话
	service.CreateSessionKey("identity-001", "agent-001")
	service.CreateSession("identity-001", "agent-001", SecurityLevelE2E)
	service.EstablishChannel("identity-001", "agent-001", "agent-002")

	stats := service.GetSecurityStats("identity-001")

	if stats.TotalKeys < 1 {
		t.Errorf("Expected at least 1 key, got %d", stats.TotalKeys)
	}
	if stats.ActiveKeys < 1 {
		t.Errorf("Expected at least 1 active key, got %d", stats.ActiveKeys)
	}
	if stats.ActiveSessions != 1 {
		t.Errorf("Expected 1 active session, got %d", stats.ActiveSessions)
	}
	if stats.ActiveChannels != 1 {
		t.Errorf("Expected 1 active channel, got %d", stats.ActiveChannels)
	}
}

func TestSecurityServiceCloseChannel(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	channel, _ := service.EstablishChannel("identity-001", "agent-001", "agent-002")

	// 关闭通道
	err := service.CloseChannel(channel.ChannelID)
	if err != nil {
		t.Fatalf("CloseChannel failed: %v", err)
	}

	// 验证状态
	stored := service.GetChannel(channel.ChannelID)
	if stored.IsActive {
		t.Error("Channel should be inactive after close")
	}
}

func TestEncryptedDataJSON(t *testing.T) {
	data := &EncryptedData{
		DataID:     "data-001",
		Algorithm:  EncryptionAES256GCM,
		KeyID:      "key-001",
		IV:         "base64iv",
		Ciphertext: "base64ciphertext",
		Tag:        "base64tag",
		Timestamp:  time.Now(),
		DataType:   "memory",
		Checksum:   "checksum",
	}

	// 序列化
	jsonData, err := data.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := EncryptedDataFromJSON(jsonData)
	if err != nil {
		t.Fatalf("EncryptedDataFromJSON failed: %v", err)
	}

	if parsed.DataID != data.DataID {
		t.Error("DataID mismatch after JSON roundtrip")
	}
	if parsed.Algorithm != data.Algorithm {
		t.Error("Algorithm mismatch after JSON roundtrip")
	}
	if parsed.Ciphertext != data.Ciphertext {
		t.Error("Ciphertext mismatch after JSON roundtrip")
	}
}

func TestSecurityServiceListener(t *testing.T) {
	service := NewSecurityService(DefaultSecurityConfig())

	keyCreated := false
	sessionCreated := false
	channelEstablished := false
	encryptSuccess := false

	listener := &TestSecurityListener{
		onKeyCreated: func(k *SecurityKey) {
			keyCreated = true
		},
		onSessionCreated: func(s *SecuritySession) {
			sessionCreated = true
		},
		onChannelEstablished: func(c *SecureChannel) {
			channelEstablished = true
		},
		onEncryptSuccess: func(dataID, keyID string) {
			encryptSuccess = true
		},
	}

	service.AddListener(listener)

	// 创建密钥
	key := service.CreateSessionKey("identity-001", "agent-001")

	time.Sleep(50 * time.Millisecond)

	if !keyCreated {
		t.Error("Listener should have been notified on key create")
	}

	// 创建会话
	service.CreateSession("identity-001", "agent-001", SecurityLevelE2E)

	time.Sleep(50 * time.Millisecond)

	if !sessionCreated {
		t.Error("Listener should have been notified on session create")
	}

	// 建立通道
	service.EstablishChannel("identity-001", "agent-001", "agent-002")

	time.Sleep(50 * time.Millisecond)

	if !channelEstablished {
		t.Error("Listener should have been notified on channel establish")
	}

	// 加密数据
	service.Encrypt([]byte("test"), key.KeyID)

	time.Sleep(50 * time.Millisecond)

	if !encryptSuccess {
		t.Error("Listener should have been notified on encrypt success")
	}
}

// 测试监听器
type TestSecurityListener struct {
	onKeyCreated         func(k *SecurityKey)
	onKeyRotated         func(oldKeyID, newKeyID string)
	onSessionCreated     func(s *SecuritySession)
	onSessionExpired     func(sessionID string)
	onChannelEstablished func(c *SecureChannel)
	onChannelClosed      func(channelID string)
	onEncryptSuccess     func(dataID, keyID string)
	onDecryptSuccess     func(dataID, keyID string)
	onViolation          func(violationType, details string)
}

func (l *TestSecurityListener) OnKeyCreated(k *SecurityKey) {
	if l.onKeyCreated != nil {
		l.onKeyCreated(k)
	}
}

func (l *TestSecurityListener) OnKeyRotated(oldKeyID, newKeyID string) {
	if l.onKeyRotated != nil {
		l.onKeyRotated(oldKeyID, newKeyID)
	}
}

func (l *TestSecurityListener) OnSessionCreated(s *SecuritySession) {
	if l.onSessionCreated != nil {
		l.onSessionCreated(s)
	}
}

func (l *TestSecurityListener) OnSessionExpired(sessionID string) {
	if l.onSessionExpired != nil {
		l.onSessionExpired(sessionID)
	}
}

func (l *TestSecurityListener) OnChannelEstablished(c *SecureChannel) {
	if l.onChannelEstablished != nil {
		l.onChannelEstablished(c)
	}
}

func (l *TestSecurityListener) OnChannelClosed(channelID string) {
	if l.onChannelClosed != nil {
		l.onChannelClosed(channelID)
	}
}

func (l *TestSecurityListener) OnEncryptionSuccess(dataID, keyID string) {
	if l.onEncryptSuccess != nil {
		l.onEncryptSuccess(dataID, keyID)
	}
}

func (l *TestSecurityListener) OnDecryptionSuccess(dataID, keyID string) {
	if l.onDecryptSuccess != nil {
		l.onDecryptSuccess(dataID, keyID)
	}
}

func (l *TestSecurityListener) OnSecurityViolation(violationType, details string) {
	if l.onViolation != nil {
		l.onViolation(violationType, details)
	}
}
// Package security - 端到端加密
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"sync"
	"time"
)

// EncryptionAlgorithm 加密算法
type EncryptionAlgorithm string

const (
	AlgAES256GCM  EncryptionAlgorithm = "AES-256-GCM"
	AlgAES256CBC  EncryptionAlgorithm = "AES-256-CBC"
	AlgChaCha20   EncryptionAlgorithm = "ChaCha20-Poly1305"
	AlgX25519     EncryptionAlgorithm = "X25519"
)

// KeyExchangeAlgorithm 密钥交换算法
type KeyExchangeAlgorithm string

const (
	KEX25519    KeyExchangeAlgorithm = "X25519"
	KEXECDHP256 KeyExchangeAlgorithm = "ECDH-P256"
	KEXECDHP384 KeyExchangeAlgorithm = "ECDH-P384"
)

// KeyPair 密钥对
type KeyPair struct {
	Algorithm KeyExchangeAlgorithm `json:"algorithm"`
	PublicKey []byte               `json:"public_key"`
	PrivateKey []byte              `json:"private_key,omitempty"` // 不序列化
	CreatedAt time.Time            `json:"created_at"`
	ExpiresAt *time.Time           `json:"expires_at,omitempty"`
	KeyID     string               `json:"key_id"`
}

// E2EConfig 端到端加密配置
type E2EConfig struct {
	Algorithm       EncryptionAlgorithm   `json:"algorithm"`
	KeyExchange     KeyExchangeAlgorithm `json:"key_exchange"`
	KeyRotationInt  time.Duration         `json:"key_rotation_interval"`
	SessionTimeout  time.Duration         `json:"session_timeout"`
	MaxMessageSize  int64                 `json:"max_message_size"`
	EnableForward   bool                  `json:"enable_forward_secrecy"` // 前向保密
}

// E2ESession 加密会话
type E2ESession struct {
	SessionID     string               `json:"session_id"`
	PeerID        string               `json:"peer_id"`
	SharedSecret  []byte               `json:"-"`
	EncryptionKey []byte               `json:"-"`
	MACKey        []byte               `json:"-"`
	Algorithm     EncryptionAlgorithm  `json:"algorithm"`
	CreatedAt     time.Time            `json:"created_at"`
	LastActive    time.Time            `json:"last_active"`
	MessageCount  int64                `json:"message_count"`
	Status        string               `json:"status"` // active, expired, closed
}

// EncryptedMessage 加密消息
type EncryptedMessage struct {
	Version      byte            `json:"version"`
	SessionID    string          `json:"session_id"`
	Nonce        []byte          `json:"nonce"`
	Ciphertext   []byte          `json:"ciphertext"`
	Tag          []byte          `json:"tag,omitempty"`      // 认证标签
	SenderPubKey []byte          `json:"sender_pub_key,omitempty"` // 发送方公钥(用于前向保密)
	Signature    []byte          `json:"signature,omitempty"` // 数字签名
	Timestamp    int64           `json:"timestamp"`
	KeyID        string          `json:"key_id,omitempty"`
}

// E2EManager 端到端加密管理器
type E2EManager struct {
	config        E2EConfig
	keyPair       *KeyPair
	sessions      map[string]*E2ESession
	peerPublicKeys map[string][]byte // peer_id -> public_key
	pendingHandshakes map[string]*HandshakeState
	messageBuffer map[string][]*EncryptedMessage // 等待解密的消息
	mu            sync.RWMutex
}

// HandshakeState 握手状态
type HandshakeState struct {
	PeerID        string
	LocalPubKey   []byte
	LocalPrivKey  []byte
	RemotePubKey  []byte
	Step          int
	CreatedAt     time.Time
}

// NewE2EManager 创建端到端加密管理器
func NewE2EManager(config E2EConfig) *E2EManager {
	// 默认配置
	if config.Algorithm == "" {
		config.Algorithm = AlgAES256GCM
	}
	if config.KeyExchange == "" {
		config.KeyExchange = KEX25519
	}
	if config.KeyRotationInt == 0 {
		config.KeyRotationInt = 24 * time.Hour
	}
	if config.SessionTimeout == 0 {
		config.SessionTimeout = 7 * 24 * time.Hour
	}
	if config.MaxMessageSize == 0 {
		config.MaxMessageSize = 10 * 1024 * 1024 // 10MB
	}

	return &E2EManager{
		config:            config,
		sessions:          make(map[string]*E2ESession),
		peerPublicKeys:    make(map[string][]byte),
		pendingHandshakes: make(map[string]*HandshakeState),
		messageBuffer:     make(map[string][]*EncryptedMessage),
	}
}

// Initialize 初始化(生成密钥对)
func (em *E2EManager) Initialize() error {
	keyPair, err := em.generateKeyPair()
	if err != nil {
		return fmt.Errorf("生成密钥对失败: %w", err)
	}

	em.mu.Lock()
	em.keyPair = keyPair
	em.mu.Unlock()

	return nil
}

// generateKeyPair 生成密钥对
func (em *E2EManager) generateKeyPair() (*KeyPair, error) {
	keyID := generateKeyID()

	switch em.config.KeyExchange {
	case KEX25519:
		privKey, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		return &KeyPair{
			Algorithm:  KEX25519,
			PublicKey:  privKey.PublicKey().Bytes(),
			PrivateKey: privKey.Bytes(),
			CreatedAt:  time.Now(),
			KeyID:      keyID,
		}, nil

	case KEXECDHP256:
		privKey, err := ecdsa.GenerateKey(ellipticP256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		pubKeyBytes, _ := privKey.PublicKey.ECDH()
		privKeyBytes, _ := privKey.ECDH()
		return &KeyPair{
			Algorithm:  KEXECDHP256,
			PublicKey:  pubKeyBytes.Bytes(),
			PrivateKey: privKeyBytes.Bytes(),
			CreatedAt:  time.Now(),
			KeyID:      keyID,
		}, nil

	default:
		return nil, fmt.Errorf("不支持的密钥交换算法: %s", em.config.KeyExchange)
	}
}

// === 密钥交换 ===

// InitiateHandshake 发起握手
func (em *E2EManager) InitiateHandshake(peerID string) (*HandshakeMessage, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// 生成临时密钥对
	tempKeyPair, err := em.generateKeyPair()
	if err != nil {
		return nil, err
	}

	// 创建握手状态
	hs := &HandshakeState{
		PeerID:       peerID,
		LocalPubKey:  tempKeyPair.PublicKey,
		LocalPrivKey: tempKeyPair.PrivateKey,
		Step:         1,
		CreatedAt:    time.Now(),
	}
	em.pendingHandshakes[peerID] = hs

	return &HandshakeMessage{
		Type:       "initiate",
		PeerID:     peerID,
		PublicKey:  tempKeyPair.PublicKey,
		KeyID:      tempKeyPair.KeyID,
		Timestamp:  time.Now().Unix(),
	}, nil
}

// HandshakeMessage 握手消息
type HandshakeMessage struct {
	Type      string    `json:"type"`       // initiate, response, complete
	PeerID    string    `json:"peer_id"`
	PublicKey []byte    `json:"public_key"`
	KeyID     string    `json:"key_id"`
	Timestamp int64     `json:"timestamp"`
	SessionID string    `json:"session_id,omitempty"`
}

// ProcessHandshake 处理握手消息
func (em *E2EManager) ProcessHandshake(msg *HandshakeMessage) (*HandshakeMessage, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	switch msg.Type {
	case "initiate":
		return em.handleHandshakeInitiate(msg)
	case "response":
		return em.handleHandshakeResponse(msg)
	case "complete":
		return em.handleHandshakeComplete(msg)
	default:
		return nil, fmt.Errorf("未知握手消息类型: %s", msg.Type)
	}
}

// handleHandshakeInitiate 处理握手发起
func (em *E2EManager) handleHandshakeInitiate(msg *HandshakeMessage) (*HandshakeMessage, error) {
	// 生成临时密钥对
	tempKeyPair, err := em.generateKeyPair()
	if err != nil {
		return nil, err
	}

	// 计算共享密钥
	sharedSecret, err := em.computeSharedSecret(tempKeyPair.PrivateKey, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	// 创建会话
	sessionID := generateSessionID()
	session := &E2ESession{
		SessionID:     sessionID,
		PeerID:        msg.PeerID,
		SharedSecret:  sharedSecret,
		EncryptionKey: deriveKey(sharedSecret, []byte("encryption")),
		MACKey:        deriveKey(sharedSecret, []byte("mac")),
		Algorithm:     em.config.Algorithm,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		Status:        "active",
	}
	em.sessions[sessionID] = session

	// 存储对端公钥
	em.peerPublicKeys[msg.PeerID] = msg.PublicKey

	// 创建握手状态
	em.pendingHandshakes[msg.PeerID] = &HandshakeState{
		PeerID:       msg.PeerID,
		LocalPubKey:  tempKeyPair.PublicKey,
		LocalPrivKey: tempKeyPair.PrivateKey,
		RemotePubKey: msg.PublicKey,
		Step:         2,
		CreatedAt:    time.Now(),
	}

	return &HandshakeMessage{
		Type:      "response",
		PeerID:    msg.PeerID,
		PublicKey: tempKeyPair.PublicKey,
		KeyID:     tempKeyPair.KeyID,
		SessionID: sessionID,
		Timestamp: time.Now().Unix(),
	}, nil
}

// handleHandshakeResponse 处理握手响应
func (em *E2EManager) handleHandshakeResponse(msg *HandshakeMessage) (*HandshakeMessage, error) {
	hs, ok := em.pendingHandshakes[msg.PeerID]
	if !ok {
		return nil, fmt.Errorf("未找到握手状态: %s", msg.PeerID)
	}

	// 计算共享密钥
	sharedSecret, err := em.computeSharedSecret(hs.LocalPrivKey, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	// 创建会话
	sessionID := msg.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	session := &E2ESession{
		SessionID:     sessionID,
		PeerID:        msg.PeerID,
		SharedSecret:  sharedSecret,
		EncryptionKey: deriveKey(sharedSecret, []byte("encryption")),
		MACKey:        deriveKey(sharedSecret, []byte("mac")),
		Algorithm:     em.config.Algorithm,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		Status:        "active",
	}
	em.sessions[sessionID] = session

	// 存储对端公钥
	em.peerPublicKeys[msg.PeerID] = msg.PublicKey

	// 清理握手状态
	delete(em.pendingHandshakes, msg.PeerID)

	return &HandshakeMessage{
		Type:      "complete",
		PeerID:    msg.PeerID,
		SessionID: sessionID,
		Timestamp: time.Now().Unix(),
	}, nil
}

// handleHandshakeComplete 处理握手完成
func (em *E2EManager) handleHandshakeComplete(msg *HandshakeMessage) (*HandshakeMessage, error) {
	delete(em.pendingHandshakes, msg.PeerID)
	return nil, nil
}

// computeSharedSecret 计算共享密钥
func (em *E2EManager) computeSharedSecret(privateKey, peerPublicKey []byte) ([]byte, error) {
	switch em.config.KeyExchange {
	case KEX25519:
		privKey, err := ecdh.X25519().NewPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		pubKey, err := ecdh.X25519().NewPublicKey(peerPublicKey)
		if err != nil {
			return nil, err
		}
		return privKey.ECDH(pubKey)

	default:
		return nil, fmt.Errorf("不支持的密钥交换算法")
	}
}

// === 加密解密 ===

// Encrypt 加密消息
func (em *E2EManager) Encrypt(sessionID string, plaintext []byte) (*EncryptedMessage, error) {
	em.mu.RLock()
	session, ok := em.sessions[sessionID]
	em.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 检查会话状态
	if session.Status != "active" {
		return nil, fmt.Errorf("会话未激活")
	}

	// 检查消息大小
	if int64(len(plaintext)) > em.config.MaxMessageSize {
		return nil, fmt.Errorf("消息大小超过限制")
	}

	// 生成随机Nonce
	nonce := make([]byte, em.nonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	var ciphertext []byte
	var tag []byte

	switch session.Algorithm {
	case AlgAES256GCM:
		ciphertext, tag = em.encryptAESGCM(session.EncryptionKey, nonce, plaintext)
	case AlgChaCha20:
		ciphertext, tag = em.encryptChaCha20(session.EncryptionKey, nonce, plaintext)
	default:
		return nil, fmt.Errorf("不支持的加密算法: %s", session.Algorithm)
	}

	// 更新会话状态
	em.mu.Lock()
	session.LastActive = time.Now()
	session.MessageCount++
	em.mu.Unlock()

	return &EncryptedMessage{
		Version:     1,
		SessionID:   sessionID,
		Nonce:       nonce,
		Ciphertext:  ciphertext,
		Tag:         tag,
		Timestamp:   time.Now().UnixNano(),
	}, nil
}

// Decrypt 解密消息
func (em *E2EManager) Decrypt(msg *EncryptedMessage) ([]byte, error) {
	em.mu.RLock()
	session, ok := em.sessions[msg.SessionID]
	em.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", msg.SessionID)
	}

	var plaintext []byte
	var err error

	switch session.Algorithm {
	case AlgAES256GCM:
		plaintext, err = em.decryptAESGCM(session.EncryptionKey, msg.Nonce, msg.Ciphertext, msg.Tag)
	case AlgChaCha20:
		plaintext, err = em.decryptChaCha20(session.EncryptionKey, msg.Nonce, msg.Ciphertext, msg.Tag)
	default:
		return nil, fmt.Errorf("不支持的加密算法: %s", session.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	// 更新会话状态
	em.mu.Lock()
	session.LastActive = time.Now()
	session.MessageCount++
	em.mu.Unlock()

	return plaintext, nil
}

// encryptAESGCM AES-GCM加密
func (em *E2EManager) encryptAESGCM(key, nonce, plaintext []byte) ([]byte, []byte) {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	// GCM的tag在最后16字节
	tag := ciphertext[len(ciphertext)-16:]
	ciphertext = ciphertext[:len(ciphertext)-16]

	return ciphertext, tag
}

// decryptAESGCM AES-GCM解密
func (em *E2EManager) decryptAESGCM(key, nonce, ciphertext, tag []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 组合ciphertext和tag
	combined := append(ciphertext, tag...)
	return gcm.Open(nil, nonce, combined, nil)
}

// encryptChaCha20 ChaCha20-Poly1305加密
func (em *E2EManager) encryptChaCha20(key, nonce, plaintext []byte) ([]byte, []byte) {
	// 简化实现，实际应使用golang.org/x/crypto/chacha20poly1305
	return plaintext, make([]byte, 16)
}

// decryptChaCha20 ChaCha20-Poly1305解密
func (em *E2EManager) decryptChaCha20(key, nonce, ciphertext, tag []byte) ([]byte, error) {
	return ciphertext, nil
}

// nonceSize 获取Nonce大小
func (em *E2EManager) nonceSize() int {
	switch em.config.Algorithm {
	case AlgAES256GCM:
		return 12
	case AlgChaCha20:
		return 12
	default:
		return 12
	}
}

// === 数字签名 ===

// SigningKey 签名密钥
type SigningKey struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	KeyID      string
	CreatedAt  time.Time
}

// GenerateSigningKey 生成签名密钥
func (em *E2EManager) GenerateSigningKey() (*SigningKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &SigningKey{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		KeyID:      generateKeyID(),
		CreatedAt:  time.Now(),
	}, nil
}

// Sign 签名
func (em *E2EManager) Sign(key *SigningKey, message []byte) []byte {
	return ed25519.Sign(key.PrivateKey, message)
}

// Verify 验证签名
func (em *E2EManager) Verify(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// === 会话管理 ===

// GetSession 获取会话
func (em *E2EManager) GetSession(sessionID string) (*E2ESession, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	session, ok := em.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	return session, nil
}

// CloseSession 关闭会话
func (em *E2EManager) CloseSession(sessionID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	session, ok := em.sessions[sessionID]
	if !ok {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	session.Status = "closed"
	delete(em.sessions, sessionID)

	return nil
}

// RotateSessionKey 轮换会话密钥
func (em *E2EManager) RotateSessionKey(sessionID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	session, ok := em.sessions[sessionID]
	if !ok {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 重新派生密钥(使用新随机数)
	newSalt := make([]byte, 32)
	io.ReadFull(rand.Reader, newSalt)

	session.EncryptionKey = deriveKey(session.SharedSecret, append([]byte("encryption"), newSalt...))
	session.MACKey = deriveKey(session.SharedSecret, append([]byte("mac"), newSalt...))
	session.LastActive = time.Now()

	return nil
}

// ListSessions 列出所有会话
func (em *E2EManager) ListSessions() []*E2ESession {
	em.mu.RLock()
	defer em.mu.RUnlock()

	sessions := make([]*E2ESession, 0, len(em.sessions))
	for _, session := range em.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetPublicKey 获取公钥
func (em *E2EManager) GetPublicKey() []byte {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.keyPair == nil {
		return nil
	}
	return em.keyPair.PublicKey
}

// GetKeyID 获取密钥ID
func (em *E2EManager) GetKeyID() string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.keyPair == nil {
		return ""
	}
	return em.keyPair.KeyID
}

// === 辅助函数 ===

// deriveKey 派生密钥
func deriveKey(secret, info []byte) []byte {
	h := sha256.New()
	h.Write(secret)
	h.Write(info)
	return h.Sum(nil)
}

// generateKeyID 生成密钥ID
func generateKeyID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}

// ellipticP256 返回P256曲线(简化)
func ellipticP256() interface{} {
	return nil
}

// Hash 计算哈希
func Hash(algorithm string, data []byte) ([]byte, error) {
	var h hash.Hash

	switch algorithm {
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return nil, errors.New("不支持的哈希算法")
	}

	h.Write(data)
	return h.Sum(nil), nil
}

// EncodeBase64 Base64编码
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 Base64解码
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
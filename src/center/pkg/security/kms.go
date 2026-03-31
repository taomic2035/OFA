// Package security provides Key Management System integration
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KeyProvider defines key provider interface
type KeyProvider interface {
	GenerateKey(keyID string, keySize int) ([]byte, error)
	GetKey(keyID string) ([]byte, error)
	DeleteKey(keyID string) error
	RotateKey(keyID string) ([]byte, error)
}

// KMSConfig holds KMS configuration
type KMSConfig struct {
	Provider     string `json:"provider"`      // local, aws, azure, gcp, vault
	KeyPrefix    string `json:"key_prefix"`
	KeyRotationDays int  `json:"key_rotation_days"`
	StoragePath  string `json:"storage_path"`
}

// KMSManager manages encryption keys
type KMSManager struct {
	config KMSConfig

	// Key storage
	keys       sync.Map // map[string]*KeyInfo
	provider   KeyProvider

	// Key cache
	keyCache   sync.Map // map[string][]byte
	cacheTTL   time.Duration

	// Encryption contexts
	contexts   sync.Map // map[string]*EncryptionContext

	mu sync.RWMutex
}

// KeyInfo holds key metadata
type KeyInfo struct {
	KeyID       string    `json:"key_id"`
	KeyType     string    `json:"key_type"`     // symmetric, asymmetric
	KeySize     int       `json:"key_size"`     // bits
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	RotatedAt   time.Time `json:"rotated_at,omitempty"`
	Version     int       `json:"version"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	TenantID    string    `json:"tenant_id,omitempty"`
}

// EncryptionContext holds encryption context
type EncryptionContext struct {
	KeyID       string            `json:"key_id"`
	Algorithm   string            `json:"algorithm"`
	Context     map[string]string `json:"context"`
	CreatedAt   time.Time         `json:"created_at"`
}

// EncryptedData represents encrypted data
type EncryptedData struct {
	KeyID       string    `json:"key_id"`
	IV          []byte    `json:"iv"`
	Ciphertext  []byte    `json:"ciphertext"`
	Tag         []byte    `json:"tag,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
	EncryptedAt time.Time `json:"encrypted_at"`
}

// NewKMSManager creates a new KMS manager
func NewKMSManager(config KMSConfig) (*KMSManager, error) {
	manager := &KMSManager{
		config:   config,
		cacheTTL: 5 * time.Minute,
	}

	// Initialize provider
	switch config.Provider {
	case "local":
		manager.provider = &LocalKeyProvider{storagePath: config.StoragePath}
	case "aws":
		// manager.provider = &AWSKMSProvider{}
	case "azure":
		// manager.provider = &AzureKeyVaultProvider{}
	case "gcp":
		// manager.provider = &GCPKMSProvider{}
	case "vault":
		// manager.provider = &VaultProvider{}
	default:
		manager.provider = &LocalKeyProvider{storagePath: config.StoragePath}
	}

	// Load existing keys
	if err := manager.load(); err != nil {
		return nil, err
	}

	// Start key rotation checker
	go manager.rotationChecker()

	return manager, nil
}

// CreateKey creates a new encryption key
func (m *KMSManager) CreateKey(keyID string, keySize int, tenantID string) (*KeyInfo, error) {
	// Generate key
	key, err := m.provider.GenerateKey(keyID, keySize)
	if err != nil {
		return nil, err
	}

	// Store key info
	info := &KeyInfo{
		KeyID:     keyID,
		KeyType:   "symmetric",
		KeySize:   keySize,
		CreatedAt: time.Now(),
		Version:   1,
		Enabled:   true,
		TenantID:  tenantID,
	}

	// Set rotation date
	if m.config.KeyRotationDays > 0 {
		info.ExpiresAt = time.Now().AddDate(0, 0, m.config.KeyRotationDays)
	}

	m.keys.Store(keyID, info)
	m.keyCache.Store(keyID, key)

	// Save metadata
	m.save()

	return info, nil
}

// GetKey retrieves a key
func (m *KMSManager) GetKey(keyID string) ([]byte, error) {
	// Check cache first
	if v, ok := m.keyCache.Load(keyID); ok {
		return v.([]byte), nil
	}

	// Get from provider
	key, err := m.provider.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	// Cache the key
	m.keyCache.Store(keyID, key)

	return key, nil
}

// DeleteKey deletes a key
func (m *KMSManager) DeleteKey(keyID string) error {
	if err := m.provider.DeleteKey(keyID); err != nil {
		return err
	}

	m.keys.Delete(keyID)
	m.keyCache.Delete(keyID)

	m.save()

	return nil
}

// RotateKey rotates a key
func (m *KMSManager) RotateKey(keyID string) (*KeyInfo, error) {
	// Get current key info
	info, err := m.GetKeyInfo(keyID)
	if err != nil {
		return nil, err
	}

	// Generate new key
	key, err := m.provider.RotateKey(keyID)
	if err != nil {
		return nil, err
	}

	// Update key info
	info.Version++
	info.RotatedAt = time.Now()

	if m.config.KeyRotationDays > 0 {
		info.ExpiresAt = time.Now().AddDate(0, 0, m.config.KeyRotationDays)
	}

	m.keys.Store(keyID, info)
	m.keyCache.Store(keyID, key)

	m.save()

	return info, nil
}

// GetKeyInfo retrieves key metadata
func (m *KMSManager) GetKeyInfo(keyID string) (*KeyInfo, error) {
	if v, ok := m.keys.Load(keyID); ok {
		return v.(*KeyInfo), nil
	}
	return nil, fmt.Errorf("key not found: %s", keyID)
}

// ListKeys lists all keys
func (m *KMSManager) ListKeys(tenantID string) []*KeyInfo {
	var keys []*KeyInfo

	m.keys.Range(func(key, value interface{}) bool {
		info := value.(*KeyInfo)
		if tenantID == "" || info.TenantID == tenantID {
			keys = append(keys, info)
		}
		return true
	})

	return keys
}

// Encrypt encrypts data with a key
func (m *KMSManager) Encrypt(keyID string, plaintext []byte, context map[string]string) (*EncryptedData, error) {
	// Get key
	key, err := m.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	// Split ciphertext and tag
	tagSize := 16
	tag := ciphertext[len(ciphertext)-tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-tagSize]

	return &EncryptedData{
		KeyID:      keyID,
		IV:         iv,
		Ciphertext: ciphertext,
		Tag:        tag,
		Context:    context,
		EncryptedAt: time.Now(),
	}, nil
}

// Decrypt decrypts data
func (m *KMSManager) Decrypt(encrypted *EncryptedData) ([]byte, error) {
	// Get key
	key, err := m.GetKey(encrypted.KeyID)
	if err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Combine ciphertext and tag
	combined := append(encrypted.Ciphertext, encrypted.Tag...)

	// Decrypt
	plaintext, err := gcm.Open(nil, encrypted.IV, combined, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts a string
func (m *KMSManager) EncryptString(keyID string, plaintext string, context map[string]string) (string, error) {
	encrypted, err := m.Encrypt(keyID, []byte(plaintext), context)
	if err != nil {
		return "", err
	}

	// Serialize
	data, err := json.Marshal(encrypted)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// DecryptString decrypts a string
func (m *KMSManager) DecryptString(encryptedStr string) (string, error) {
	// Decode
	data, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", err
	}

	var encrypted EncryptedData
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return "", err
	}

	plaintext, err := m.Decrypt(&encrypted)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// rotationChecker periodically checks for key rotation
func (m *KMSManager) rotationChecker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.checkRotation()
	}
}

// checkRotation checks if any keys need rotation
func (m *KMSManager) checkRotation() {
	m.keys.Range(func(key, value interface{}) bool {
		info := value.(*KeyInfo)

		if !info.Enabled {
			return true
		}

		if !info.ExpiresAt.IsZero() && time.Now().After(info.ExpiresAt) {
			// Rotate key
			m.RotateKey(info.KeyID)
		}

		return true
	})
}

// load loads keys from disk
func (m *KMSManager) load() error {
	if m.config.StoragePath == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(m.config.StoragePath, "keys.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var keys map[string]*KeyInfo
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}

	for k, v := range keys {
		m.keys.Store(k, v)
	}

	return nil
}

// save saves keys to disk
func (m *KMSManager) save() error {
	if m.config.StoragePath == "" {
		return nil
	}

	os.MkdirAll(m.config.StoragePath, 0755)

	keys := make(map[string]*KeyInfo)
	m.keys.Range(func(key, value interface{}) bool {
		keys[key.(string)] = value.(*KeyInfo)
		return true
	})

	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.config.StoragePath, "keys.json"), data, 0644)
}

// LocalKeyProvider provides local key storage
type LocalKeyProvider struct {
	storagePath string
	keys        sync.Map
	mu          sync.RWMutex
}

// GenerateKey generates a new key
func (p *LocalKeyProvider) GenerateKey(keyID string, keySize int) ([]byte, error) {
	key := make([]byte, keySize/8)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	p.keys.Store(keyID, key)

	// Save to disk
	if p.storagePath != "" {
		os.MkdirAll(p.storagePath, 0755)
		os.WriteFile(filepath.Join(p.storagePath, keyID+".key"), key, 0600)
	}

	return key, nil
}

// GetKey retrieves a key
func (p *LocalKeyProvider) GetKey(keyID string) ([]byte, error) {
	// Check memory
	if v, ok := p.keys.Load(keyID); ok {
		return v.([]byte), nil
	}

	// Try to load from disk
	if p.storagePath != "" {
		key, err := os.ReadFile(filepath.Join(p.storagePath, keyID+".key"))
		if err != nil {
			return nil, fmt.Errorf("key not found: %s", keyID)
		}
		p.keys.Store(keyID, key)
		return key, nil
	}

	return nil, fmt.Errorf("key not found: %s", keyID)
}

// DeleteKey deletes a key
func (p *LocalKeyProvider) DeleteKey(keyID string) error {
	p.keys.Delete(keyID)

	if p.storagePath != "" {
		os.Remove(filepath.Join(p.storagePath, keyID+".key"))
	}

	return nil
}

// RotateKey rotates a key
func (p *LocalKeyProvider) RotateKey(keyID string) ([]byte, error) {
	// Get old key info
	oldKey, err := p.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	// Generate new key
	newKey := make([]byte, len(oldKey))
	if _, err := rand.Read(newKey); err != nil {
		return nil, err
	}

	// Store new key
	p.keys.Store(keyID, newKey)

	if p.storagePath != "" {
		os.WriteFile(filepath.Join(p.storagePath, keyID+".key"), newKey, 0600)
	}

	return newKey, nil
}

// DeriveKey derives a key from a master key
func DeriveKey(masterKey []byte, context string, keySize int) ([]byte) {
	h := sha256.New()
	h.Write(masterKey)
	h.Write([]byte(context))
	return h.Sum(nil)[:keySize/8]
}
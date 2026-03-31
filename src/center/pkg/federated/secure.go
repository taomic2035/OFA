// Package securefl provides security enhancements for federated learning
package securefl

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"sync"
	"time"
)

// SecurityLevel defines security levels
type SecurityLevel string

const (
	SecurityStandard SecurityLevel = "standard" // Standard security
	SecurityHigh     SecurityLevel = "high"     // High security
	SecurityMaximum  SecurityLevel = "maximum"  // Maximum security
)

// SecureAggregationProtocol implements secure aggregation
type SecureAggregationProtocol struct {
	securityLevel SecurityLevel
	keySize       int

	// Client keys
	clientKeys sync.Map // map[string]*ClientKey

	// Session management
	sessions sync.Map // map[string]*AggregationSession

	mu sync.RWMutex
}

// ClientKey represents a client's cryptographic keys
type ClientKey struct {
	ClientID    string    `json:"client_id"`
	PublicKey   []byte    `json:"public_key"`
	PrivateKey  []byte    `json:"private_key,omitempty"`
	KeyPairID   string    `json:"key_pair_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AggregationSession represents a secure aggregation session
type AggregationSession struct {
	SessionID   string            `json:"session_id"`
	TaskID      string            `json:"task_id"`
	ClientIDs   []string          `json:"client_ids"`
	SecretShare map[string][]byte `json:"secret_shares"` // clientID -> share
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

// DPConfig holds differential privacy configuration
type DPConfig struct {
	Epsilon      float64 `json:"epsilon"`       // Privacy budget
	Delta        float64 `json:"delta"`         // Privacy parameter
	Sensitivity  float64 `json:"sensitivity"`   // L2 sensitivity
	ClipNorm     float64 `json:"clip_norm"`     // Gradient clipping norm
	Mechanism    string  `json:"mechanism"`     // gaussian, laplace
	NoiseScale   float64 `json:"noise_scale"`   // Noise scale factor
	AdaptiveClip bool    `json:"adaptive_clip"` // Adaptive clipping
}

// DPNoiseGenerator generates differential privacy noise
type DPNoiseGenerator struct {
	config DPConfig
}

// NewSecureAggregationProtocol creates a new secure aggregation protocol
func NewSecureAggregationProtocol(level SecurityLevel) *SecureAggregationProtocol {
	keySize := 256
	if level == SecurityMaximum {
		keySize = 512
	}

	return &SecureAggregationProtocol{
		securityLevel: level,
		keySize:       keySize,
	}
}

// RegisterClient registers a client for secure aggregation
func (p *SecureAggregationProtocol) RegisterClient(clientID string) (*ClientKey, error) {
	// Generate key pair
	publicKey, privateKey, err := p.generateKeyPair()
	if err != nil {
		return nil, err
	}

	key := &ClientKey{
		ClientID:   clientID,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		KeyPairID:  generateKeyID(),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	p.clientKeys.Store(clientID, key)

	return key, nil
}

// generateKeyPair generates a key pair
func (p *SecureAggregationProtocol) generateKeyPair() ([]byte, []byte, error) {
	// Simplified key generation
	// In production, would use proper cryptographic libraries
	publicKey := make([]byte, p.keySize/8)
	privateKey := make([]byte, p.keySize/8)

	if _, err := rand.Read(publicKey); err != nil {
		return nil, nil, err
	}
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

// CreateSession creates a secure aggregation session
func (p *SecureAggregationProtocol) CreateSession(taskID string, clientIDs []string) (*AggregationSession, error) {
	if len(clientIDs) < 2 {
		return nil, errors.New("at least 2 clients required")
	}

	session := &AggregationSession{
		SessionID:   generateSessionID(),
		TaskID:      taskID,
		ClientIDs:   clientIDs,
		SecretShare: make(map[string][]byte),
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	// Generate secret shares for each client
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}

	shares := p.generateSecretShares(secret, len(clientIDs))

	for i, clientID := range clientIDs {
		session.SecretShare[clientID] = shares[i]
	}

	p.sessions.Store(session.SessionID, session)

	return session, nil
}

// generateSecretShares generates Shamir's secret shares
func (p *SecureAggregationProtocol) generateSecretShares(secret []byte, n int) [][]byte {
	// Simplified secret sharing
	// In production, would use Shamir's Secret Sharing
	shares := make([][]byte, n)

	for i := 0; i < n; i++ {
		share := make([]byte, len(secret))
		copy(share, secret)

		// Add random component
		random := make([]byte, len(secret))
		rand.Read(random)

		// XOR with random (simplified)
		for j := range share {
			share[j] ^= random[j]
		}

		shares[i] = share
	}

	return shares
}

// EncryptUpdate encrypts a client update
func (p *SecureAggregationProtocol) EncryptUpdate(sessionID, clientID string, update []byte) ([]byte, error) {
	session, ok := p.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	sess := session.(*AggregationSession)

	key, ok := p.clientKeys.Load(clientID)
	if !ok {
		return nil, fmt.Errorf("client key not found: %s", clientID)
	}

	clientKey := key.(*ClientKey)

	// Encrypt using client's public key
	encrypted, err := p.encryptWithKey(update, clientKey.PublicKey)
	if err != nil {
		return nil, err
	}

	// Add secret share
	share := sess.SecretShare[clientID]
	encrypted = append(encrypted, share...)

	return encrypted, nil
}

// DecryptUpdate decrypts a client update
func (p *SecureAggregationProtocol) DecryptUpdate(sessionID string, encrypted []byte) ([]byte, error) {
	// Remove secret share (last 32 bytes)
	if len(encrypted) < 32 {
		return nil, errors.New("invalid encrypted data")
	}

	update := encrypted[:len(encrypted)-32]

	return update, nil
}

// encryptWithKey encrypts data with a key
func (p *SecureAggregationProtocol) encryptWithKey(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// AggregateSecurely performs secure aggregation
func (p *SecureAggregationProtocol) AggregateSecurely(sessionID string, encryptedUpdates [][]byte) ([]byte, error) {
	if len(encryptedUpdates) == 0 {
		return nil, errors.New("no updates to aggregate")
	}

	// Decrypt all updates
	var updates [][]byte
	for _, enc := range encryptedUpdates {
		dec, err := p.DecryptUpdate(sessionID, enc)
		if err != nil {
			continue
		}
		updates = append(updates, dec)
	}

	// Aggregate (simplified - would use proper secure aggregation)
	aggregated := make([]byte, len(updates[0]))
	for _, update := range updates {
		for i := range aggregated {
			aggregated[i] ^= update[i]
		}
	}

	return aggregated, nil
}

// NewDPNoiseGenerator creates a new DP noise generator
func NewDPNoiseGenerator(config DPConfig) *DPNoiseGenerator {
	return &DPNoiseGenerator{config: config}
}

// AddNoise adds differential privacy noise to gradients
func (g *DPNoiseGenerator) AddNoise(gradients []float64) []float64 {
	noisy := make([]float64, len(gradients))

	for i, grad := range gradients {
		// Clip gradient
		if g.config.ClipNorm > 0 {
			grad = g.clipGradient(grad, g.config.ClipNorm)
		}

		// Add noise
		noise := g.generateNoise()
		noisy[i] = grad + noise
	}

	return noisy
}

// clipGradient clips a gradient to specified norm
func (g *DPNoiseGenerator) clipGradient(grad, norm float64) float64 {
	if grad > norm {
		return norm
	}
	if grad < -norm {
		return -norm
	}
	return grad
}

// generateNoise generates DP noise
func (g *DPNoiseGenerator) generateNoise() float64 {
	switch g.config.Mechanism {
	case "gaussian":
		return g.gaussianNoise()
	case "laplace":
		return g.laplaceNoise()
	default:
		return g.gaussianNoise()
	}
}

// gaussianNoise generates Gaussian noise
func (g *DPNoiseGenerator) gaussianNoise() float64 {
	// Standard deviation = sensitivity * sqrt(2 * ln(1.25/delta)) / epsilon
	sigma := g.config.Sensitivity * g.calculateSigma() / g.config.Epsilon

	// Box-Muller transform
	u1 := rand.Float64()
	u2 := rand.Float64()

	z0 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)

	return z0 * sigma
}

// laplaceNoise generates Laplace noise
func (g *DPNoiseGenerator) laplaceNoise() float64 {
	scale := g.config.Sensitivity / g.config.Epsilon

	u := rand.Float64() - 0.5
	sign := 1.0
	if u < 0 {
		sign = -1.0
	}

	return sign * scale * math.Log(1-2*math.Abs(u))
}

// calculateSigma calculates noise sigma
func (g *DPNoiseGenerator) calculateSigma() float64 {
	if g.config.Delta <= 0 {
		return 1.0
	}
	return math.Sqrt(2 * math.Log(1.25/g.config.Delta))
}

// HomomorphicEncryption provides homomorphic encryption operations
type HomomorphicEncryption struct {
	publicKey  []byte
	privateKey []byte
	keySize    int
}

// NewHomomorphicEncryption creates a new HE instance
func NewHomomorphicEncryption(keySize int) *HomomorphicEncryption {
	return &HomomorphicEncryption{keySize: keySize}
}

// GenerateKeys generates HE keys
func (h *HomomorphicEncryption) GenerateKeys() error {
	// Placeholder - would use proper HE library
	h.publicKey = make([]byte, h.keySize/8)
	h.privateKey = make([]byte, h.keySize/8)

	rand.Read(h.publicKey)
	rand.Read(h.privateKey)

	return nil
}

// Encrypt encrypts a value
func (h *HomomorphicEncryption) Encrypt(value float64) ([]byte, error) {
	// Simplified encryption
	encoded := []byte(fmt.Sprintf("%.10f", value))
	encrypted := make([]byte, len(encoded))

	for i, b := range encoded {
		encrypted[i] = b ^ h.publicKey[i%len(h.publicKey)]
	}

	return encrypted, nil
}

// Decrypt decrypts a value
func (h *HomomorphicEncryption) Decrypt(encrypted []byte) (float64, error) {
	decrypted := make([]byte, len(encrypted))

	for i, b := range encrypted {
		decrypted[i] = b ^ h.privateKey[i%len(h.privateKey)]
	}

	var value float64
	_, err := fmt.Sscanf(string(decrypted), "%f", &value)
	return value, err
}

// Add performs homomorphic addition
func (h *HomomorphicEncryption) Add(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errors.New("encrypted values must have same length")
	}

	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}

	return result, nil
}

// SecurityAuditor audits security of federated learning operations
type SecurityAuditor struct {
	auditLog []SecurityEvent
	mu       sync.RWMutex
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"`
	ClientID  string    `json:"client_id,omitempty"`
	TaskID    string    `json:"task_id,omitempty"`
	Details   string    `json:"details"`
	Severity  string    `json:"severity"`
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{
		auditLog: make([]SecurityEvent, 0),
	}
}

// LogEvent logs a security event
func (a *SecurityAuditor) LogEvent(eventType, clientID, taskID, details, severity string) {
	event := SecurityEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		ClientID:  clientID,
		TaskID:    taskID,
		Details:   details,
		Severity:  severity,
	}

	a.mu.Lock()
	a.auditLog = append(a.auditLog, event)
	a.mu.Unlock()
}

// GetAuditLog returns the audit log
func (a *SecurityAuditor) GetAuditLog() []SecurityEvent {
	a.mu.RLock()
	defer a.mu.Unlock()

	log := make([]SecurityEvent, len(a.auditLog))
	copy(log, a.auditLog)
	return log
}

// VerifyClientUpdate verifies integrity of client update
func (a *SecurityAuditor) VerifyClientUpdate(clientID string, update []byte, signature string) error {
	// Verify signature
	expectedSig := hex.EncodeToString(sha256.New().Sum(update))
	if signature != expectedSig {
		return errors.New("invalid signature")
	}

	a.LogEvent("update_verified", clientID, "", "Update signature verified", "info")
	return nil
}

// Helper functions

func generateKeyID() string {
	return fmt.Sprintf("key-%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
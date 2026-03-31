package market

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SkillMetadata represents skill metadata
type SkillMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Category    string            `json:"category"`    // text, json, calculator, etc.
	Tags        []string          `json:"tags"`
	Depends     []string          `json:"depends"`     // Skill dependencies
	Provides    []string          `json:"provides"`    // Provided operations
	Inputs      []SkillInput      `json:"inputs"`
	Outputs     []SkillOutput     `json:"outputs"`
	Signature   string            `json:"signature"`   // Ed25519 signature
	Hash        string            `json:"hash"`        // SHA256 hash
	PublishedAt time.Time         `json:"published_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata"`
}

// SkillInput defines skill input parameters
type SkillInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // string, number, boolean, object, array
	Required bool   `json:"required"`
	Default  string `json:"default"`
	Description string `json:"description"`
}

// SkillOutput defines skill output format
type SkillOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Description string `json:"description"`
}

// SkillPackage contains the skill code and metadata
type SkillPackage struct {
	Metadata SkillMetadata `json:"metadata"`
	Code     []byte        `json:"code"`      // Skill implementation
	Config   []byte        `json:"config"`    // Configuration schema
}

// MarketConfig holds capability market configuration
type MarketConfig struct {
	StoragePath    string        // Local storage path
	RemoteURL      string        // Remote market URL
	CacheEnabled   bool
	CacheDuration  time.Duration
	VerifySignatures bool
}

// DefaultMarketConfig returns default configuration
func DefaultMarketConfig() *MarketConfig {
	return &MarketConfig{
		StoragePath:     "skills",
		CacheEnabled:    true,
		CacheDuration:   1 * time.Hour,
		VerifySignatures: true,
	}
}

// CapabilityMarket manages the skill repository
type CapabilityMarket struct {
	config  *MarketConfig
	client  *http.Client

	// Local skill registry
	skills sync.Map // map[string]*SkillPackage

	// Version tracking
	versions sync.Map // map[string][]string (skillID -> versions)

	// Cache
	cache sync.Map // map[string]*cachedSkill

	// Verification keys
	pubKeys sync.Map // map[string]ed25519.PublicKey (author -> key)

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// cachedSkill represents a cached skill entry
type cachedSkill struct {
	skill     *SkillPackage
	cachedAt  time.Time
}

// NewCapabilityMarket creates a new capability market
func NewCapabilityMarket(config *MarketConfig) (*CapabilityMarket, error) {
	ctx, cancel := context.WithCancel(context.Background())

	market := &CapabilityMarket{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		ctx:     ctx,
		cancel:  cancel,
	}

	// Load local skills
	if err := market.loadLocalSkills(); err != nil {
		log.Printf("Failed to load local skills: %v", err)
	}

	return market, nil
}

// Start begins market sync
func (m *CapabilityMarket) Start() {
	if m.config.RemoteURL != "" {
		go m.syncLoop()
	}
}

// Stop stops the market
func (m *CapabilityMarket) Stop() {
	m.cancel()
}

// loadLocalSkills loads skills from local storage
func (m *CapabilityMarket) loadLocalSkills() error {
	if m.config.StoragePath == "" {
		return nil
	}

	if _, err := os.Stat(m.config.StoragePath); os.IsNotExist(err) {
		return os.MkdirAll(m.config.StoragePath, 0755)
	}

	return filepath.Walk(m.config.StoragePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".skill" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var pkg SkillPackage
			if err := json.Unmarshal(data, &pkg); err != nil {
				return err
			}

			m.skills.Store(pkg.Metadata.ID, &pkg)
		}

		return nil
	})
}

// syncLoop syncs with remote market
func (m *CapabilityMarket) syncLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Initial sync
	m.syncRemote()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.syncRemote()
		}
	}
}

// syncRemote syncs with remote market
func (m *CapabilityMarket) syncRemote() {
	if m.config.RemoteURL == "" {
		return
	}

	url := fmt.Sprintf("%s/skills", m.config.RemoteURL)
	resp, err := m.client.Get(url)
	if err != nil {
		log.Printf("Failed to sync with remote market: %v", err)
		return
	}
	defer resp.Body.Close()

	var skills []*SkillMetadata
	if err := json.NewDecoder(resp.Body).Decode(&skills); err != nil {
		log.Printf("Failed to decode remote skills: %v", err)
		return
	}

	for _, meta := range skills {
		// Check if we have this skill
		if v, ok := m.skills.Load(meta.ID); ok {
			local := v.(*SkillPackage)
			if local.Metadata.Version >= meta.Version {
				continue // Already have latest
			}
		}

		// Download new/updated skill
		m.downloadSkill(meta.ID)
	}
}

// downloadSkill downloads a skill from remote market
func (m *CapabilityMarket) downloadSkill(skillID string) error {
	url := fmt.Sprintf("%s/skills/%s", m.config.RemoteURL, skillID)
	resp, err := m.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var pkg SkillPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	// Verify signature if enabled
	if m.config.VerifySignatures {
		if err := m.verifySkill(&pkg); err != nil {
			return fmt.Errorf("signature verification failed: %v", err)
		}
	}

	// Store skill
	m.skills.Store(skillID, &pkg)

	// Update version tracking
	m.updateVersions(skillID, pkg.Metadata.Version)

	// Save to local storage
	m.saveSkill(skillID, &pkg)

	return nil
}

// PublishSkill publishes a new skill to the market
func (m *CapabilityMarket) PublishSkill(pkg *SkillPackage, privateKey ed25519.PrivateKey) error {
	// Calculate hash
	hash := sha256.Sum256(pkg.Code)
	pkg.Metadata.Hash = hex.EncodeToString(hash[:])

	// Sign skill
	signature := ed25519.Sign(privateKey, pkg.Code)
	pkg.Metadata.Signature = hex.EncodeToString(signature)

	// Set timestamps
	now := time.Now()
	pkg.Metadata.PublishedAt = now
	pkg.Metadata.UpdatedAt = now

	// Verify signature
	if m.config.VerifySignatures {
		pubKey, ok := m.pubKeys.Load(pkg.Metadata.Author)
		if ok {
			if !ed25519.Verify(pubKey.(ed25519.PublicKey), pkg.Code, signature) {
				return errors.New("signature verification failed")
			}
		}
	}

	// Store locally
	m.skills.Store(pkg.Metadata.ID, pkg)
	m.updateVersions(pkg.Metadata.ID, pkg.Metadata.Version)

	// Save to storage
	m.saveSkill(pkg.Metadata.ID, pkg)

	// Publish to remote if configured
	if m.config.RemoteURL != "" {
		return m.publishRemote(pkg)
	}

	return nil
}

// publishRemote publishes skill to remote market
func (m *CapabilityMarket) publishRemote(pkg *SkillPackage) error {
	data, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/skills", m.config.RemoteURL)
	req, err := http.NewRequestWithContext(m.ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("publish failed: %s", resp.Status)
	}

	return nil
}

// verifySkill verifies skill signature
func (m *CapabilityMarket) verifySkill(pkg *SkillPackage) error {
	if pkg.Metadata.Signature == "" {
		return errors.New("missing signature")
	}

	pubKey, ok := m.pubKeys.Load(pkg.Metadata.Author)
	if !ok {
		return errors.New("unknown author")
	}

	sig, err := hex.DecodeString(pkg.Metadata.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %v", err)
	}

	if !ed25519.Verify(pubKey.(ed25519.PublicKey), pkg.Code, sig) {
		return errors.New("signature mismatch")
	}

	// Verify hash
	hash := sha256.Sum256(pkg.Code)
	if hex.EncodeToString(hash[:]) != pkg.Metadata.Hash {
		return errors.New("hash mismatch")
	}

	return nil
}

// saveSkill saves skill to local storage
func (m *CapabilityMarket) saveSkill(skillID string, pkg *SkillPackage) error {
	if m.config.StoragePath == "" {
		return nil
	}

	data, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	filename := filepath.Join(m.config.StoragePath, fmt.Sprintf("%s_%s.skill", pkg.Metadata.ID, pkg.Metadata.Version))
	return os.WriteFile(filename, data, 0644)
}

// updateVersions updates version tracking
func (m *CapabilityMarket) updateVersions(skillID, version string) {
	if v, ok := m.versions.Load(skillID); ok {
		versions := v.([]string)
		// Check if version already exists
		for _, v := range versions {
			if v == version {
				return
			}
		}
		versions = append(versions, version)
		m.versions.Store(skillID, versions)
	} else {
		m.versions.Store(skillID, []string{version})
	}
}

// GetSkill retrieves a skill by ID
func (m *CapabilityMarket) GetSkill(skillID string) (*SkillPackage, error) {
	// Check cache first
	if m.config.CacheEnabled {
		if v, ok := m.cache.Load(skillID); ok {
			cached := v.(*cachedSkill)
			if time.Since(cached.cachedAt) < m.config.CacheDuration {
				return cached.skill, nil
			}
		}
	}

	// Get from registry
	if v, ok := m.skills.Load(skillID); ok {
		pkg := v.(*SkillPackage)

		// Update cache
		if m.config.CacheEnabled {
			m.cache.Store(skillID, &cachedSkill{
				skill:    pkg,
				cachedAt: time.Now(),
			})
		}

		return pkg, nil
	}

	// Try to download from remote
	if m.config.RemoteURL != "" {
		if err := m.downloadSkill(skillID); err != nil {
			return nil, fmt.Errorf("skill not found: %s", skillID)
		}

		if v, ok := m.skills.Load(skillID); ok {
			return v.(*SkillPackage), nil
		}
	}

	return nil, fmt.Errorf("skill not found: %s", skillID)
}

// GetSkillVersion retrieves a specific version
func (m *CapabilityMarket) GetSkillVersion(skillID, version string) (*SkillPackage, error) {
	// Check versions
	if v, ok := m.versions.Load(skillID); ok {
		versions := v.([]string)
		found := false
		for _, v := range versions {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found for skill %s", version, skillID)
		}
	}

	// Check local storage
	if m.config.StoragePath != "" {
		filename := filepath.Join(m.config.StoragePath, fmt.Sprintf("%s_%s.skill", skillID, version))
		data, err := os.ReadFile(filename)
		if err == nil {
			var pkg SkillPackage
			if err := json.Unmarshal(data, &pkg); err == nil {
				return &pkg, nil
			}
		}
	}

	// Try remote
	if m.config.RemoteURL != "" {
		url := fmt.Sprintf("%s/skills/%s/%s", m.config.RemoteURL, skillID, version)
		resp, err := m.client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err == nil {
				var pkg SkillPackage
				if err := json.Unmarshal(data, &pkg); err == nil {
					return &pkg, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("skill version not found: %s@%s", skillID, version)
}

// SearchSkills searches skills by criteria
func (m *CapabilityMarket) SearchSkills(query string) []*SkillMetadata {
	var results []*SkillMetadata

	m.skills.Range(func(key, value interface{}) bool {
		pkg := value.(*SkillPackage)
		meta := pkg.Metadata

		// Simple search: check name, description, tags
		if contains(meta.Name, query) ||
			contains(meta.Description, query) ||
			containsSlice(meta.Tags, query) ||
			contains(meta.Category, query) {
			results = append(results, &meta)
		}

		return true
	})

	return results
}

// ListSkills returns all available skills
func (m *CapabilityMarket) ListSkills() []*SkillMetadata {
	var skills []*SkillMetadata

	m.skills.Range(func(key, value interface{}) bool {
		pkg := value.(*SkillPackage)
		skills = append(skills, &pkg.Metadata)
		return true
	})

	return skills
}

// GetVersions returns all versions of a skill
func (m *CapabilityMarket) GetVersions(skillID string) ([]string, error) {
	if v, ok := m.versions.Load(skillID); ok {
		return v.([]string), nil
	}
	return nil, fmt.Errorf("skill not found: %s", skillID)
}

// RegisterPublicKey registers an author's public key
func (m *CapabilityMarket) RegisterPublicKey(author string, pubKey ed25519.PublicKey) {
	m.pubKeys.Store(author, pubKey)
}

// ResolveDependencies resolves skill dependencies
func (m *CapabilityMarket) ResolveDependencies(skillID string) ([]*SkillPackage, error) {
	pkg, err := m.GetSkill(skillID)
	if err != nil {
		return nil, err
	}

	if len(pkg.Metadata.Depends) == 0 {
		return nil, nil
	}

	var deps []*SkillPackage
	for _, depID := range pkg.Metadata.Depends {
		dep, err := m.GetSkill(depID)
		if err != nil {
			return nil, fmt.Errorf("dependency %s not found", depID)
		}
		deps = append(deps, dep)

		// Recursively resolve
		subDeps, err := m.ResolveDependencies(depID)
		if err != nil {
			return nil, err
		}
		deps = append(deps, subDeps...)
	}

	return deps, nil
}

// contains checks if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && s[:len(substr)] == substr)
}

// containsSlice checks if slice contains element
func containsSlice(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
// Package wasm - WASM技能加载器
package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SkillLoader 技能加载器
type SkillLoader struct {
	config      LoaderConfig
	registry    *SkillRegistry
	cache       *SkillCache
	httpClient  *http.Client
	mu          sync.RWMutex
}

// LoaderConfig 加载器配置
type LoaderConfig struct {
	CacheDir       string        `json:"cache_dir"`
	DownloadTimeout time.Duration `json:"download_timeout"`
	MaxSkillSize   int64         `json:"max_skill_size"` // 最大技能大小(字节)
	AllowedSources []string      `json:"allowed_sources"`
	VerifySignature bool          `json:"verify_signature"`
}

// DefaultLoaderConfig 默认加载器配置
func DefaultLoaderConfig() LoaderConfig {
	return LoaderConfig{
		CacheDir:       "./skills/cache",
		DownloadTimeout: 60 * time.Second,
		MaxSkillSize:   50 * 1024 * 1024, // 50MB
		AllowedSources: []string{"local", "http", "https"},
		VerifySignature: true,
	}
}

// SkillSource 技能来源
type SkillSource struct {
	Type     string `json:"type"`     // local, http, registry
	Location string `json:"location"` // 路径或URL
	Checksum string `json:"checksum"` // SHA256校验和
	Version  string `json:"version"`
}

// NewSkillLoader 创建技能加载器
func NewSkillLoader(config LoaderConfig) *SkillLoader {
	if config.CacheDir == "" {
		config.CacheDir = "./skills/cache"
	}
	os.MkdirAll(config.CacheDir, 0755)

	return &SkillLoader{
		config:     config,
		registry:   NewSkillRegistry(),
		cache:      NewSkillCache(config.CacheDir),
		httpClient: &http.Client{Timeout: config.DownloadTimeout},
	}
}

// LoadFromBytes 从字节数据加载
func (l *SkillLoader) LoadFromBytes(skillID string, data []byte) (*WASMSkill, error) {
	// 计算校验和
	checksum := sha256.Sum256(data)
	checksumStr := hex.EncodeToString(checksum[:])

	// 解析技能元数据
	metadata, err := l.parseMetadata(data)
	if err != nil {
		// 没有元数据，使用默认值
		metadata = &SkillMetadata{
			ID:         skillID,
			Name:       skillID,
			Version:    "1.0.0",
			Operations: []string{"execute"},
		}
	}

	skill := &WASMSkill{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Version:     metadata.Version,
		Description: metadata.Description,
		Author:      metadata.Author,
		Operations:  metadata.Operations,
		Binary:      data,
		Exports:     metadata.Exports,
		Config: SkillConfig{
			MemoryPages: metadata.MemoryPages,
			Timeout:     metadata.Timeout,
			Permissions: metadata.Permissions,
		},
	}

	// 验证技能
	if checksumStr != "" && l.config.VerifySignature {
		skill.Config.Permissions = append(skill.Config.Permissions, "verified")
	}

	return skill, nil
}

// LoadFromFile 从文件加载
func (l *SkillLoader) LoadFromFile(path string) (*WASMSkill, error) {
	// 检查文件大小
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	if info.Size() > l.config.MaxSkillSize {
		return nil, fmt.Errorf("文件过大: %d > %d", info.Size(), l.config.MaxSkillSize)
	}

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	skillID := strings.TrimSuffix(filepath.Base(path), ".wasm")
	return l.LoadFromBytes(skillID, data)
}

// LoadFromURL 从URL加载
func (l *SkillLoader) LoadFromURL(ctx context.Context, url, checksum string) (*WASMSkill, error) {
	// 检查协议
	if !l.isSourceAllowed(url) {
		return nil, fmt.Errorf("不允许的来源: %s", url)
	}

	// 检查缓存
	if cached, ok := l.cache.Get(url); ok {
		if checksum == "" || cached.Checksum == checksum {
			return l.LoadFromBytes(cached.ID, cached.Binary)
		}
	}

	// 下载
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 检查大小
	if resp.ContentLength > l.config.MaxSkillSize {
		return nil, fmt.Errorf("文件过大: %d > %d", resp.ContentLength, l.config.MaxSkillSize)
	}

	// 读取内容
	limited := io.LimitReader(resp.Body, l.config.MaxSkillSize)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("读取失败: %w", err)
	}

	// 验证校验和
	actualChecksum := sha256.Sum256(data)
	actualChecksumStr := hex.EncodeToString(actualChecksum[:])
	if checksum != "" && actualChecksumStr != checksum {
		return nil, fmt.Errorf("校验和不匹配")
	}

	// 缓存
	skill, err := l.LoadFromBytes(filepath.Base(url), data)
	if err != nil {
		return nil, err
	}

	l.cache.Set(url, &CachedSkill{
		ID:       skill.ID,
		Binary:   data,
		Checksum: actualChecksumStr,
		LoadedAt: time.Now(),
	})

	return skill, nil
}

// LoadFromRegistry 从注册表加载
func (l *SkillLoader) LoadFromRegistry(ctx context.Context, skillID, version string) (*WASMSkill, error) {
	// 查询注册表
	info, err := l.registry.Get(skillID, version)
	if err != nil {
		return nil, fmt.Errorf("技能不在注册表中: %w", err)
	}

	// 从URL加载
	return l.LoadFromURL(ctx, info.DownloadURL, info.Checksum)
}

// SaveSkill 保存技能到本地
func (l *SkillLoader) SaveSkill(skill *WASMSkill, dir string) error {
	if dir == "" {
		dir = l.config.CacheDir
	}
	os.MkdirAll(dir, 0755)

	// 保存二进制
	binaryPath := filepath.Join(dir, skill.ID+".wasm")
	if err := os.WriteFile(binaryPath, skill.Binary, 0644); err != nil {
		return fmt.Errorf("保存二进制失败: %w", err)
	}

	// 保存元数据
	metadata := SkillMetadata{
		ID:          skill.ID,
		Name:        skill.Name,
		Version:     skill.Version,
		Description: skill.Description,
		Author:      skill.Author,
		Operations:  skill.Operations,
		Exports:     skill.Exports,
		MemoryPages: skill.Config.MemoryPages,
		Timeout:     skill.Config.Timeout,
		Permissions: skill.Config.Permissions,
	}

	metadataPath := filepath.Join(dir, skill.ID+".json")
	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("保存元数据失败: %w", err)
	}

	return nil
}

// ListLocalSkills 列出本地技能
func (l *SkillLoader) ListLocalSkills(dir string) ([]string, error) {
	if dir == "" {
		dir = l.config.CacheDir
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.wasm"))
	if err != nil {
		return nil, err
	}

	skills := make([]string, 0, len(files))
	for _, f := range files {
		skills = append(skills, strings.TrimSuffix(filepath.Base(f), ".wasm"))
	}
	return skills, nil
}

// parseMetadata 解析技能元数据
func (l *SkillLoader) parseMetadata(data []byte) (*SkillMetadata, error) {
	// 在WASM二进制中查找自定义段
	// 这里简化处理，实际需要解析WASM格式
	return nil, fmt.Errorf("no metadata found")
}

// isSourceAllowed 检查来源是否允许
func (l *SkillLoader) isSourceAllowed(url string) bool {
	if strings.HasPrefix(url, "http://") {
		for _, src := range l.config.AllowedSources {
			if src == "http" {
				return true
			}
		}
		return false
	}
	if strings.HasPrefix(url, "https://") {
		for _, src := range l.config.AllowedSources {
			if src == "https" {
				return true
			}
		}
		return false
	}
	return true
}

// SkillMetadata 技能元数据
type SkillMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Operations  []string          `json:"operations"`
	Exports     map[string]string `json:"exports"`
	MemoryPages int               `json:"memory_pages"`
	Timeout     int               `json:"timeout"`
	Permissions []string          `json:"permissions"`
}

// SkillRegistry 技能注册表
type SkillRegistry struct {
	skills map[string]map[string]*RegistryEntry // skillID -> version -> entry
	mu     sync.RWMutex
}

// RegistryEntry 注册表条目
type RegistryEntry struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DownloadURL string    `json:"download_url"`
	Checksum    string    `json:"checksum"`
	PublishedAt time.Time `json:"published_at"`
}

// NewSkillRegistry 创建技能注册表
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]map[string]*RegistryEntry),
	}
}

// Register 注册技能
func (r *SkillRegistry) Register(entry *RegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.skills[entry.ID]; !ok {
		r.skills[entry.ID] = make(map[string]*RegistryEntry)
	}
	r.skills[entry.ID][entry.Version] = entry
	return nil
}

// Get 获取技能信息
func (r *SkillRegistry) Get(skillID, version string) (*RegistryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.skills[skillID]
	if !ok {
		return nil, fmt.Errorf("技能不存在: %s", skillID)
	}

	if version == "" || version == "latest" {
		// 返回最新版本
		var latest *RegistryEntry
		for _, v := range versions {
			if latest == nil || v.PublishedAt.After(latest.PublishedAt) {
				latest = v
			}
		}
		return latest, nil
	}

	entry, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("版本不存在: %s", version)
	}
	return entry, nil
}

// List 列出技能
func (r *SkillRegistry) List() []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.skills {
		for _, entry := range versions {
			entries = append(entries, entry)
		}
	}
	return entries
}

// SkillCache 技能缓存
type SkillCache struct {
	dir    string
	cache  map[string]*CachedSkill
	mu     sync.RWMutex
}

// CachedSkill 缓存的技能
type CachedSkill struct {
	ID       string    `json:"id"`
	Binary   []byte    `json:"binary"`
	Checksum string    `json:"checksum"`
	LoadedAt time.Time `json:"loaded_at"`
}

// NewSkillCache 创建技能缓存
func NewSkillCache(dir string) *SkillCache {
	return &SkillCache{
		dir:   dir,
		cache: make(map[string]*CachedSkill),
	}
}

// Get 获取缓存
func (c *SkillCache) Get(key string) (*CachedSkill, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	skill, ok := c.cache[key]
	return skill, ok
}

// Set 设置缓存
func (c *SkillCache) Set(key string, skill *CachedSkill) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = skill
}

// Delete 删除缓存
func (c *SkillCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

// Clear 清空缓存
func (c *SkillCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*CachedSkill)
}
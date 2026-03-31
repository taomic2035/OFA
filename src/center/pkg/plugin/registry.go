// Package plugin - 插件注册表
// Sprint 28: v8.1插件系统增强
package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// PluginRegistry 插件注册表
type PluginRegistry struct {
	plugins   map[string]map[string]*RegistryEntry // pluginID -> version -> entry
	local     map[string]Plugin                    // 本地已加载的插件
	categories map[string][]string                 // 分类索引
	mu        sync.RWMutex
	config    RegistryConfig
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	DataFile      string `json:"data_file"`       // 注册表数据文件
	EnableRemote  bool   `json:"enable_remote"`   // 启用远程注册表
	RemoteURL     string `json:"remote_url"`      // 远程注册表URL
	CacheDuration int    `json:"cache_duration"`  // 缓存时长（秒）
}

// DefaultRegistryConfig 默认注册表配置
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		DataFile:      "./plugins/registry.json",
		EnableRemote:  true,
		RemoteURL:     "https://plugins.ofa.io/registry",
		CacheDuration: 3600,
	}
}

// RegistryEntry 注册表条目
type RegistryEntry struct {
	ID           string      `json:"id"`
	Version      string      `json:"version"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Author       string      `json:"author"`
	Type         PluginType  `json:"type"`
	Tags         []string    `json:"tags"`
	Category     string      `json:"category"`
	DownloadURL  string      `json:"download_url"`
	Checksum     string      `json:"checksum"`
	Signature    string      `json:"signature"`
	Homepage     string      `json:"homepage"`
	Repository   string      `json:"repository"`
	License      string      `json:"license"`
	Rating       float64     `json:"rating"`
	Downloads    int64        `json:"downloads"`
	Dependencies []Dependency `json:"dependencies"`
	Permissions  []string    `json:"permissions"`
	PublishedAt  time.Time   `json:"published_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Verified     bool        `json:"verified"`
	Featured     bool        `json:"featured"`
}

// NewPluginRegistry 创建插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:    make(map[string]map[string]*RegistryEntry),
		local:      make(map[string]Plugin),
		categories: make(map[string][]string),
		config:     DefaultRegistryConfig(),
	}
}

// Register 注册插件到注册表
func (r *PluginRegistry) Register(entry *RegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry.ID == "" {
		return fmt.Errorf("插件ID不能为空")
	}

	if entry.Version == "" {
		entry.Version = "1.0.0"
	}

	if _, ok := r.plugins[entry.ID]; !ok {
		r.plugins[entry.ID] = make(map[string]*RegistryEntry)
	}

	r.plugins[entry.ID][entry.Version] = entry

	// 更新分类索引
	if entry.Category != "" {
		r.addToCategory(entry.Category, entry.ID)
	}

	return nil
}

// Unregister 从注册表注销插件
func (r *PluginRegistry) Unregister(pluginID, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.plugins[pluginID]
	if !ok {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	if version == "" || version == "all" {
		// 删除所有版本
		delete(r.plugins, pluginID)
	} else {
		delete(versions, version)
		if len(versions) == 0 {
			delete(r.plugins, pluginID)
		}
	}

	return nil
}

// Get 获取插件信息
func (r *PluginRegistry) Get(pluginID, version string) (*RegistryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.plugins[pluginID]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", pluginID)
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

// GetLocal 获取本地已加载的插件
func (r *PluginRegistry) GetLocal(pluginID string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, ok := r.local[pluginID]
	return plugin, ok
}

// RegisterLocal 注册本地插件
func (r *PluginRegistry) RegisterLocal(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := plugin.Info()
	r.local[info.ID] = plugin

	return nil
}

// UnregisterLocal 注销本地插件
func (r *PluginRegistry) UnregisterLocal(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.local, pluginID)
}

// List 列出所有插件
func (r *PluginRegistry) List() []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.plugins {
		for _, entry := range versions {
			entries = append(entries, entry)
		}
	}

	// 按评分排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rating > entries[j].Rating
	})

	return entries
}

// ListByCategory 按分类列出插件
func (r *PluginRegistry) ListByCategory(category string) []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	pluginIDs, ok := r.categories[category]
	if !ok {
		return entries
	}

	for _, id := range pluginIDs {
		if versions, ok := r.plugins[id]; ok {
			for _, entry := range versions {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// ListByType 按类型列出插件
func (r *PluginRegistry) ListByType(pluginType PluginType) []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.plugins {
		for _, entry := range versions {
			if entry.Type == pluginType {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// Search 搜索插件
func (r *PluginRegistry) Search(query string) []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	entries := make([]*RegistryEntry, 0)

	for _, versions := range r.plugins {
		for _, entry := range versions {
			// 匹配名称、描述、标签
			if strings.Contains(strings.ToLower(entry.Name), query) ||
				strings.Contains(strings.ToLower(entry.Description), query) ||
				r.matchTags(entry.Tags, query) {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// matchTags 匹配标签
func (r *PluginRegistry) matchTags(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetCategories 获取分类列表
func (r *PluginRegistry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]string, 0, len(r.categories))
	for cat := range r.categories {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// addToCategory 添加到分类
func (r *PluginRegistry) addToCategory(category, pluginID string) {
	if _, ok := r.categories[category]; !ok {
		r.categories[category] = make([]string, 0)
	}

	// 检查是否已存在
	for _, id := range r.categories[category] {
		if id == pluginID {
			return
		}
	}

	r.categories[category] = append(r.categories[category], pluginID)
}

// GetFeatured 获取精选插件
func (r *PluginRegistry) GetFeatured(limit int) []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.plugins {
		for _, entry := range versions {
			if entry.Featured {
				entries = append(entries, entry)
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rating > entries[j].Rating
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

// GetPopular 获取热门插件
func (r *PluginRegistry) GetPopular(limit int) []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.plugins {
		for _, entry := range versions {
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Downloads > entries[j].Downloads
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}

// GetVerified 获取已验证插件
func (r *PluginRegistry) GetVerified() []*RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]*RegistryEntry, 0)
	for _, versions := range r.plugins {
		for _, entry := range versions {
			if entry.Verified {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// LoadFromFile 从文件加载注册表
func (r *PluginRegistry) LoadFromFile(path string) error {
	if path == "" {
		path = r.config.DataFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，使用空注册表
		}
		return fmt.Errorf("读取注册表文件失败: %w", err)
	}

	var entries []*RegistryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("解析注册表数据失败: %w", err)
	}

	for _, entry := range entries {
		r.Register(entry)
	}

	return nil
}

// SaveToFile 保存注册表到文件
func (r *PluginRegistry) SaveToFile(path string) error {
	if path == "" {
		path = r.config.DataFile
	}

	entries := r.List()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化注册表失败: %w", err)
	}

	os.MkdirAll(filepath.Dir(path), 0755)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("保存注册表文件失败: %w", err)
	}

	return nil
}

// GetVersions 获取插件版本列表
func (r *PluginRegistry) GetVersions(pluginID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.plugins[pluginID]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", pluginID)
	}

	versionList := make([]string, 0, len(versions))
	for v := range versions {
		versionList = append(versionList, v)
	}

	// 按发布时间排序
	sort.Slice(versionList, func(i, j int) bool {
		return versions[versionList[i]].PublishedAt.After(versions[versionList[j]].PublishedAt)
	})

	return versionList, nil
}

// CheckDependencies 检查依赖是否满足
func (r *PluginRegistry) CheckDependencies(pluginID, version string) ([]Dependency, error) {
	entry, err := r.Get(pluginID, version)
	if err != nil {
		return nil, err
	}

	missing := make([]Dependency, 0)
	for _, dep := range entry.Dependencies {
		if _, err := r.Get(dep.ID, dep.Version); err != nil {
			if dep.Required {
				missing = append(missing, dep)
			}
		}
	}

	return missing, nil
}

// UpdateStats 更新统计信息
func (r *PluginRegistry) UpdateStats(pluginID, version string, downloads int64, rating float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, ok := r.plugins[pluginID]
	if !ok {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	if version == "" || version == "latest" {
		// 更新最新版本
		var latest *RegistryEntry
		for _, v := range versions {
			if latest == nil || v.PublishedAt.After(latest.PublishedAt) {
				latest = v
			}
		}
		if latest != nil {
			latest.Downloads = downloads
			latest.Rating = rating
		}
	} else {
		entry, ok := versions[version]
		if !ok {
			return fmt.Errorf("版本不存在: %s", version)
		}
		entry.Downloads = downloads
		entry.Rating = rating
	}

	return nil
}

// Stats 注册表统计
type RegistryStats struct {
	TotalPlugins   int     `json:"total_plugins"`
	TotalVersions  int     `json:"total_versions"`
	TotalDownloads int64   `json:"total_downloads"`
	AvgRating      float64 `json:"avg_rating"`
	Categories     int     `json:"categories"`
	Verified       int     `json:"verified"`
	Featured       int     `json:"featured"`
}

// GetStats 获取统计信息
func (r *PluginRegistry) GetStats() *RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &RegistryStats{}
	totalRating := 0.0

	for _, versions := range r.plugins {
		stats.TotalPlugins++
		for _, entry := range versions {
			stats.TotalVersions++
			stats.TotalDownloads += entry.Downloads
			totalRating += entry.Rating
			if entry.Verified {
				stats.Verified++
			}
			if entry.Featured {
				stats.Featured++
			}
		}
	}

	stats.Categories = len(r.categories)
	if stats.TotalVersions > 0 {
		stats.AvgRating = totalRating / float64(stats.TotalVersions)
	}

	return stats
}

// 预置分类
var DefaultCategories = []string{
	"AI & Machine Learning",
	"Data Processing",
	"Communication",
	"Automation",
	"Security",
	"Integration",
	"Utility",
	"Developer Tools",
}

// InitDefaultCategories 初始化默认分类
func (r *PluginRegistry) InitDefaultCategories() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cat := range DefaultCategories {
		if _, ok := r.categories[cat]; !ok {
			r.categories[cat] = make([]string, 0)
		}
	}
}
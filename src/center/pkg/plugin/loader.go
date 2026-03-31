// Package plugin - 插件加载器
package plugin

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

// PluginLoader 插件加载器
type PluginLoader struct {
	config      LoaderConfig
	registry    *PluginRegistry
	cache       *PluginCache
	httpClient  *http.Client
	mu          sync.RWMutex
}

// LoaderConfig 加载器配置
type LoaderConfig struct {
	PluginDir       string        `json:"plugin_dir"`
	CacheDir        string        `json:"cache_dir"`
	DownloadTimeout time.Duration `json:"download_timeout"`
	MaxPluginSize   int64         `json:"max_plugin_size"`
	AllowedSources  []string      `json:"allowed_sources"`
	VerifySignature bool          `json:"verify_signature"`
	EnableHotLoad   bool          `json:"enable_hot_load"`
}

// DefaultLoaderConfig 默认加载器配置
func DefaultLoaderConfig(pluginDir string) LoaderConfig {
	return LoaderConfig{
		PluginDir:       pluginDir,
		CacheDir:        filepath.Join(pluginDir, "cache"),
		DownloadTimeout: 60 * time.Second,
		MaxPluginSize:   100 * 1024 * 1024, // 100MB
		AllowedSources:  []string{"local", "http", "https", "registry"},
		VerifySignature: true,
		EnableHotLoad:   true,
	}
}

// PluginSource 插件来源
type PluginSource struct {
	Type     string `json:"type"`     // local, http, registry, wasm
	Location string `json:"location"` // 路径或URL
	Checksum string `json:"checksum"` // SHA256校验和
	Version  string `json:"version"`
}

// NewPluginLoader 创建插件加载器
func NewPluginLoader(config LoaderConfig) *PluginLoader {
	if config.PluginDir == "" {
		config.PluginDir = "./plugins"
	}
	if config.CacheDir == "" {
		config.CacheDir = filepath.Join(config.PluginDir, "cache")
	}

	os.MkdirAll(config.PluginDir, 0755)
	os.MkdirAll(config.CacheDir, 0755)

	return &PluginLoader{
		config:     config,
		registry:   NewPluginRegistry(),
		cache:      NewPluginCache(config.CacheDir),
		httpClient: &http.Client{Timeout: config.DownloadTimeout},
	}
}

// LoadPlugin 加载插件
func (l *PluginLoader) LoadPlugin(ctx context.Context, source PluginSource) (Plugin, error) {
	var data []byte
	var checksum string

	switch source.Type {
	case "local":
		data, checksum, err := l.loadFromLocal(source.Location)
		if err != nil {
			return nil, err
		}
		return l.parsePlugin(data, checksum)

	case "http", "https":
		data, checksum, err := l.loadFromURL(ctx, source.Location)
		if err != nil {
			return nil, err
		}
		return l.parsePlugin(data, checksum)

	case "registry":
		return l.loadFromRegistry(ctx, source.Location, source.Version)

	case "wasm":
		return l.loadWASMPlugin(ctx, source)

	default:
		return nil, fmt.Errorf("不支持的插件来源类型: %s", source.Type)
	}
}

// loadFromLocal 从本地加载
func (l *PluginLoader) loadFromLocal(path string) ([]byte, string, error) {
	// 检查路径是否在允许目录内
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("路径解析失败: %w", err)
	}

	pluginDir, err := filepath.Abs(l.config.PluginDir)
	if err != nil {
		return nil, "", fmt.Errorf("插件目录解析失败: %w", err)
	}

	// 安全检查：确保路径在插件目录内
	if !strings.HasPrefix(absPath, pluginDir) {
		return nil, "", fmt.Errorf("插件路径不在允许目录内")
	}

	// 检查文件大小
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("文件不存在: %w", err)
	}

	if info.Size() > l.config.MaxPluginSize {
		return nil, "", fmt.Errorf("文件过大: %d > %d", info.Size(), l.config.MaxPluginSize)
	}

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 计算校验和
	checksum := sha256.Sum256(data)
	return data, hex.EncodeToString(checksum[:]), nil
}

// loadFromURL 从URL加载
func (l *PluginLoader) loadFromURL(ctx context.Context, url string) ([]byte, string, error) {
	// 检查协议
	if !l.isSourceAllowed(url) {
		return nil, "", fmt.Errorf("不允许的来源: %s", url)
	}

	// 检查缓存
	if cached, ok := l.cache.Get(url); ok {
		return cached.Data, cached.Checksum, nil
	}

	// 下载
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 检查大小
	if resp.ContentLength > l.config.MaxPluginSize {
		return nil, "", fmt.Errorf("文件过大: %d > %d", resp.ContentLength, l.config.MaxPluginSize)
	}

	// 读取内容
	limited := io.LimitReader(resp.Body, l.config.MaxPluginSize)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", fmt.Errorf("读取失败: %w", err)
	}

	// 计算校验和
	checksum := sha256.Sum256(data)
	checksumStr := hex.EncodeToString(checksum[:])

	// 缓存
	l.cache.Set(url, &CachedPlugin{
		URL:      url,
		Data:     data,
		Checksum: checksumStr,
		LoadedAt: time.Now(),
	})

	return data, checksumStr, nil
}

// loadFromRegistry 从注册表加载
func (l *PluginLoader) loadFromRegistry(ctx context.Context, pluginID, version string) (Plugin, error) {
	// 查询注册表
	info, err := l.registry.Get(pluginID, version)
	if err != nil {
		return nil, fmt.Errorf("插件不在注册表中: %w", err)
	}

	// 从URL加载
	data, checksum, err := l.loadFromURL(ctx, info.DownloadURL)
	if err != nil {
		return nil, err
	}

	// 验证校验和
	if info.Checksum != "" && checksum != info.Checksum {
		return nil, fmt.Errorf("校验和不匹配")
	}

	return l.parsePlugin(data, checksum)
}

// loadWASMPlugin 加载WASM插件
func (l *PluginLoader) loadWASMPlugin(ctx context.Context, source PluginSource) (Plugin, error) {
	// WASM插件需要特殊处理
	// 这里返回一个WASMPlugin包装器
	data, checksum, err := l.loadFromLocal(source.Location)
	if err != nil {
		return nil, err
	}

	return &WASMPluginWrapper{
		id:       filepath.Base(source.Location),
		binary:   data,
		checksum: checksum,
	}, nil
}

// parsePlugin 解析插件
func (l *PluginLoader) parsePlugin(data []byte, checksum string) (Plugin, error) {
	// 尝试解析JSON格式的插件描述
	var info PluginInfo
	if err := json.Unmarshal(data, &info); err == nil {
		// JSON格式的插件描述
		return &JSONPlugin{info: &info}, nil
	}

	// 尝试解析Go插件（需要实际实现）
	// 这里返回一个基本插件
	return &BasicPlugin{
		id:       "unknown",
		data:     data,
		checksum: checksum,
	}, nil
}

// isSourceAllowed 检查来源是否允许
func (l *PluginLoader) isSourceAllowed(url string) bool {
	for _, src := range l.config.AllowedSources {
		if src == "http" && strings.HasPrefix(url, "http://") {
			return true
		}
		if src == "https" && strings.HasPrefix(url, "https://") {
			return true
		}
		if src == "registry" && strings.HasPrefix(url, "registry://") {
			return true
		}
	}
	return false
}

// ListLocalPlugins 列出本地插件
func (l *PluginLoader) ListLocalPlugins() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(l.config.PluginDir, "*.json"))
	if err != nil {
		return nil, err
	}

	plugins := make([]string, 0, len(files))
	for _, f := range files {
		plugins = append(plugins, strings.TrimSuffix(filepath.Base(f), ".json"))
	}

	// 也查找.so文件（Go插件）
	soFiles, err := filepath.Glob(filepath.Join(l.config.PluginDir, "*.so"))
	if err == nil {
		for _, f := range soFiles {
			plugins = append(plugins, strings.TrimSuffix(filepath.Base(f), ".so"))
		}
	}

	// 查找.wasm文件
	wasmFiles, err := filepath.Glob(filepath.Join(l.config.PluginDir, "*.wasm"))
	if err == nil {
		for _, f := range wasmFiles {
			plugins = append(plugins, strings.TrimSuffix(filepath.Base(f), ".wasm"))
		}
	}

	return plugins, nil
}

// WatchPluginDir 监听插件目录变化（热加载）
func (l *PluginLoader) WatchPluginDir(ctx context.Context, onChange func(path string, event string)) error {
	if !l.config.EnableHotLoad {
		return nil
	}

	// 简化的热加载实现
	// 实际可以使用fsnotify库
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		lastFiles := make(map[string]time.Time)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				files, err := l.ListLocalPlugins()
				if err != nil {
					continue
				}

				for _, f := range files {
					path := filepath.Join(l.config.PluginDir, f)
					info, err := os.Stat(path)
					if err != nil {
						continue
					}

					if lastMod, exists := lastFiles[f]; exists {
						if info.ModTime().After(lastMod) {
							onChange(path, "modified")
						}
					} else {
						onChange(path, "added")
					}
					lastFiles[f] = info.ModTime()
				}
			}
		}
	}()

	return nil
}

// SavePlugin 保存插件到本地
func (l *PluginLoader) SavePlugin(plugin Plugin, dir string) error {
	if dir == "" {
		dir = l.config.PluginDir
	}
	os.MkdirAll(dir, 0755)

	info := plugin.Info()

	// 保存插件描述
	metadataPath := filepath.Join(dir, info.ID+".json")
	metadataJSON, _ := json.MarshalIndent(info, "", "  ")
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("保存元数据失败: %w", err)
	}

	return nil
}

// ResolveDependencies 解析依赖
func (l *PluginLoader) ResolveDependencies(plugin Plugin) ([]Plugin, error) {
	info := plugin.Info()
	if len(info.Dependencies) == 0 {
		return nil, nil
	}

	dependencies := make([]Plugin, 0)
	ctx := context.Background()

	for _, dep := range info.Dependencies {
		// 检查是否已加载
		if loaded, ok := l.registry.GetLocal(dep.ID); ok {
			dependencies = append(dependencies, loaded)
			continue
		}

		// 尝试从注册表加载
		source := PluginSource{
			Type:     "registry",
			Location: dep.ID,
			Version:  dep.Version,
		}

		depPlugin, err := l.LoadPlugin(ctx, source)
		if err != nil {
			if dep.Required {
				return nil, fmt.Errorf("依赖加载失败: %s - %w", dep.ID, err)
			}
			// 可选依赖，跳过
			continue
		}

		dependencies = append(dependencies, depPlugin)
	}

	return dependencies, nil
}

// ValidatePlugin 验证插件
func (l *PluginLoader) ValidatePlugin(plugin Plugin) error {
	info := plugin.Info()

	// 验证ID
	if info.ID == "" {
		return fmt.Errorf("插件ID不能为空")
	}

	// 验证名称
	if info.Name == "" {
		return fmt.Errorf("插件名称不能为空")
	}

	// 验证版本
	if info.Version == "" {
		return fmt.Errorf("插件版本不能为空")
	}

	// 验证权限
	for _, perm := range info.Permissions {
		if !isValidPermission(perm) {
			return fmt.Errorf("无效的权限: %s", perm)
		}
	}

	return nil
}

// isValidPermission 检查权限是否有效
func isValidPermission(perm string) bool {
	validPerms := []string{
		"file.read", "file.write", "file.execute",
		"network.connect", "network.listen",
		"env.read", "env.write",
		"system.info", "system.execute",
		"agent.list", "agent.manage",
		"task.create", "task.execute",
		"message.send", "message.receive",
	}

	for _, v := range validPerms {
		if perm == v || strings.HasPrefix(perm, v+".") {
			return true
		}
	}
	return false
}

// PluginCache 插件缓存
type PluginCache struct {
	dir   string
	cache map[string]*CachedPlugin
	mu    sync.RWMutex
}

// CachedPlugin 缓存的插件
type CachedPlugin struct {
	URL      string    `json:"url"`
	Data     []byte    `json:"data"`
	Checksum string    `json:"checksum"`
	LoadedAt time.Time `json:"loaded_at"`
}

// NewPluginCache 创建插件缓存
func NewPluginCache(dir string) *PluginCache {
	return &PluginCache{
		dir:   dir,
		cache: make(map[string]*CachedPlugin),
	}
}

// Get 获取缓存
func (c *PluginCache) Get(key string) (*CachedPlugin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	plugin, ok := c.cache[key]
	return plugin, ok
}

// Set 设置缓存
func (c *PluginCache) Set(key string, plugin *CachedPlugin) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = plugin
}

// Delete 删除缓存
func (c *PluginCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

// Clear 清空缓存
func (c *PluginCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*CachedPlugin)
}

// 基本插件实现
type BasicPlugin struct {
	id       string
	data     []byte
	checksum string
}

func (p *BasicPlugin) Info() *PluginInfo {
	return &PluginInfo{
		ID:      p.id,
		Name:    p.id,
		Version: "1.0.0",
	}
}

func (p *BasicPlugin) Init(ctx context.Context, config *PluginConfig) error {
	return nil
}

func (p *BasicPlugin) Start(ctx context.Context) error {
	return nil
}

func (p *BasicPlugin) Stop(ctx context.Context) error {
	return nil
}

func (p *BasicPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return input, nil
}

func (p *BasicPlugin) Health(ctx context.Context) error {
	return nil
}

func (p *BasicPlugin) Cleanup(ctx context.Context) error {
	return nil
}

// JSON插件实现
type JSONPlugin struct {
	info *PluginInfo
}

func (p *JSONPlugin) Info() *PluginInfo {
	return p.info
}

func (p *JSONPlugin) Init(ctx context.Context, config *PluginConfig) error {
	return nil
}

func (p *JSONPlugin) Start(ctx context.Context) error {
	return nil
}

func (p *JSONPlugin) Stop(ctx context.Context) error {
	return nil
}

func (p *JSONPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return input, nil
}

func (p *JSONPlugin) Health(ctx context.Context) error {
	return nil
}

func (p *JSONPlugin) Cleanup(ctx context.Context) error {
	return nil
}

// WASM插件包装器
type WASMPluginWrapper struct {
	id       string
	binary   []byte
	checksum string
}

func (p *WASMPluginWrapper) Info() *PluginInfo {
	return &PluginInfo{
		ID:      p.id,
		Name:    p.id + " (WASM)",
		Version: "1.0.0",
		Type:    PluginTypeSkill,
	}
}

func (p *WASMPluginWrapper) Init(ctx context.Context, config *PluginConfig) error {
	return nil
}

func (p *WASMPluginWrapper) Start(ctx context.Context) error {
	return nil
}

func (p *WASMPluginWrapper) Stop(ctx context.Context) error {
	return nil
}

func (p *WASMPluginWrapper) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	// WASM执行需要通过WASMRuntime
	return nil, fmt.Errorf("WASM插件需要通过WASMRuntime执行")
}

func (p *WASMPluginWrapper) Health(ctx context.Context) error {
	return nil
}

func (p *WASMPluginWrapper) Cleanup(ctx context.Context) error {
	return nil
}
package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ofa/center/internal/models"
)

// FileStore - 文件存储实现
type FileStore struct {
	mu      sync.RWMutex
	baseDir string
	cache   map[string]*models.PersonalIdentity
}

// NewFileStore 创建文件存储
func NewFileStore(baseDir string) (*FileStore, error) {
	if baseDir == "" {
		baseDir = "./data/identity"
	}

	// 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	store := &FileStore{
		baseDir: baseDir,
		cache:   make(map[string]*models.PersonalIdentity),
	}

	// 加载已有数据
	if err := store.loadAll(); err != nil {
		// 非致命错误，继续运行
		fmt.Printf("Warning: failed to load existing identities: %v\n", err)
	}

	return store, nil
}

// loadAll 加载所有身份数据
func (s *FileStore) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，无数据
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		id := file.Name()[:len(file.Name())-5] // 移除 .json
		identity, err := s.loadFile(filepath.Join(s.baseDir, file.Name()))
		if err != nil {
			fmt.Printf("Warning: failed to load %s: %v\n", file.Name(), err)
			continue
		}

		s.cache[id] = identity
	}

	return nil
}

// loadFile 加载单个文件
func (s *FileStore) loadFile(path string) (*models.PersonalIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var identity models.PersonalIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

// SaveIdentity 保存身份
func (s *FileStore) SaveIdentity(ctx context.Context, identity *models.PersonalIdentity) error {
	if identity == nil || identity.ID == "" {
		return fmt.Errorf("invalid identity")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 序列化
	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}

	// 写文件
	path := filepath.Join(s.baseDir, identity.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 更新缓存
	s.cache[identity.ID] = identity

	return nil
}

// GetIdentity 获取身份
func (s *FileStore) GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error) {
	if id == "" {
		return nil, fmt.Errorf("empty id")
	}

	s.mu.RLock()
	if identity, ok := s.cache[id]; ok {
		s.mu.RUnlock()
		return identity, nil
	}
	s.mu.RUnlock()

	// 从文件加载
	path := filepath.Join(s.baseDir, id+".json")
	identity, err := s.loadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("identity not found: %s", id)
		}
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	// 缓存
	s.mu.Lock()
	s.cache[id] = identity
	s.mu.Unlock()

	return identity, nil
}

// DeleteIdentity 删除身份
func (s *FileStore) DeleteIdentity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除文件
	path := filepath.Join(s.baseDir, id+".json")
	if err := os.Remove(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// 从缓存移除
	delete(s.cache, id)

	return nil
}

// ListIdentities 列出所有身份
func (s *FileStore) ListIdentities(ctx context.Context, page, pageSize int) ([]*models.PersonalIdentity, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.cache)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		return []*models.PersonalIdentity{}, total, nil
	}
	if end > total {
		end = total
	}

	// 收集所有身份
	all := make([]*models.PersonalIdentity, 0, total)
	for _, identity := range s.cache {
		all = append(all, identity)
	}

	// 返回分页结果
	return all[start:end], total, nil
}

// MemoryStore - 内存存储实现（用于测试）
type MemoryStore struct {
	mu        sync.RWMutex
	identities map[string]*models.PersonalIdentity
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		identities: make(map[string]*models.PersonalIdentity),
	}
}

// SaveIdentity 保存身份
func (s *MemoryStore) SaveIdentity(ctx context.Context, identity *models.PersonalIdentity) error {
	if identity == nil || identity.ID == "" {
		return fmt.Errorf("invalid identity")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.identities[identity.ID] = identity
	return nil
}

// GetIdentity 获取身份
func (s *MemoryStore) GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error) {
	if id == "" {
		return nil, fmt.Errorf("empty id")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	identity, ok := s.identities[id]
	if !ok {
		return nil, fmt.Errorf("identity not found: %s", id)
	}

	return identity, nil
}

// DeleteIdentity 删除身份
func (s *MemoryStore) DeleteIdentity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.identities, id)
	return nil
}

// ListIdentities 列出所有身份
func (s *MemoryStore) ListIdentities(ctx context.Context, page, pageSize int) ([]*models.PersonalIdentity, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.identities)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 收集所有身份
	all := make([]*models.PersonalIdentity, 0, total)
	for _, identity := range s.identities {
		all = append(all, identity)
	}

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*models.PersonalIdentity{}, total, nil
	}
	if end > total {
		end = total
	}

	return all[start:end], total, nil
}
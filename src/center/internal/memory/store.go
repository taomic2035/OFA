package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ofa/center/internal/models"
)

// FileStore - 文件存储实现
type FileStore struct {
	mu          sync.RWMutex
	baseDir     string
	memories    map[string]*models.Memory              // id -> memory
	userIndex   map[string]map[string]*models.Memory   // userID -> id -> memory
	typeIndex   map[string]map[models.MemoryType][]string // userID -> type -> []id
	associations map[string][]*models.MemoryAssociation // memoryID -> associations
}

// NewFileStore 创建文件存储
func NewFileStore(baseDir string) (*FileStore, error) {
	if baseDir == "" {
		baseDir = "./data/memory"
	}

	// 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	store := &FileStore{
		baseDir:     baseDir,
		memories:    make(map[string]*models.Memory),
		userIndex:   make(map[string]map[string]*models.Memory),
		typeIndex:   make(map[string]map[models.MemoryType][]string),
		associations: make(map[string][]*models.MemoryAssociation),
	}

	// 加载已有数据
	if err := store.loadAll(); err != nil {
		fmt.Printf("Warning: failed to load existing memories: %v\n", err)
	}

	return store, nil
}

// loadAll 加载所有记忆
func (s *FileStore) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		memory, err := s.loadFile(filepath.Join(s.baseDir, file.Name()))
		if err != nil {
			continue
		}

		s.indexMemory(memory)
	}

	return nil
}

// loadFile 加载单个文件
func (s *FileStore) loadFile(path string) (*models.Memory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var memory models.Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// indexMemory 索引记忆
func (s *FileStore) indexMemory(memory *models.Memory) {
	// 主索引
	s.memories[memory.ID] = memory

	// 用户索引
	if s.userIndex[memory.UserID] == nil {
		s.userIndex[memory.UserID] = make(map[string]*models.Memory)
	}
	s.userIndex[memory.UserID][memory.ID] = memory

	// 类型索引
	if s.typeIndex[memory.UserID] == nil {
		s.typeIndex[memory.UserID] = make(map[models.MemoryType][]string)
	}
	s.typeIndex[memory.UserID][memory.Type] = append(
		s.typeIndex[memory.UserID][memory.Type],
		memory.ID,
	)
}

// SaveMemory 保存记忆
func (s *FileStore) SaveMemory(ctx context.Context, memory *models.Memory) error {
	if memory == nil || memory.ID == "" {
		return fmt.Errorf("invalid memory")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 序列化
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	// 写文件
	path := filepath.Join(s.baseDir, memory.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 更新索引
	s.indexMemory(memory)

	return nil
}

// GetMemory 获取记忆
func (s *FileStore) GetMemory(ctx context.Context, id string) (*models.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memory, ok := s.memories[id]
	if !ok {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	return memory, nil
}

// DeleteMemory 删除记忆
func (s *FileStore) DeleteMemory(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memory, ok := s.memories[id]
	if !ok {
		return nil
	}

	// 删除文件
	path := filepath.Join(s.baseDir, id+".json")
	os.Remove(path)

	// 从索引移除
	delete(s.memories, id)
	if userMemories, ok := s.userIndex[memory.UserID]; ok {
		delete(userMemories, id)
	}
	if typeList, ok := s.typeIndex[memory.UserID]; ok {
		newList := []string{}
		for _, memID := range typeList[memory.Type] {
			if memID != id {
				newList = append(newList, memID)
			}
		}
		typeList[memory.Type] = newList
	}
	delete(s.associations, id)

	return nil
}

// ListMemories 列出记忆
func (s *FileStore) ListMemories(ctx context.Context, query *models.MemoryQuery) ([]*models.Memory, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 获取用户的记忆
	userMemories, ok := s.userIndex[query.UserID]
	if !ok {
		return []*models.Memory{}, 0, nil
	}

	// 过滤
	var filtered []*models.Memory
	for _, memory := range userMemories {
		if !s.matchesQuery(memory, query) {
			continue
		}
		filtered = append(filtered, memory)
	}

	total := len(filtered)

	// 排序
	s.sortMemories(filtered, query.SortBy, query.SortDesc)

	// 分页
	start := query.Offset
	if start > total {
		return []*models.Memory{}, total, nil
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// matchesQuery 检查是否匹配查询
func (s *FileStore) matchesQuery(memory *models.Memory, query *models.MemoryQuery) bool {
	// 类型过滤
	if len(query.Types) > 0 {
		found := false
		for _, t := range query.Types {
			if memory.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 类别过滤
	if len(query.Categories) > 0 {
		found := false
		for _, c := range query.Categories {
			if memory.Category == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 标签过滤
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			found := false
			for _, t := range memory.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 层级过滤
	if query.Layer != "" && memory.Layer != query.Layer {
		return false
	}

	// 重要性过滤
	if memory.GetEffectiveImportance() < query.MinImportance {
		return false
	}

	// 时间范围过滤
	if query.StartTime != nil && memory.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && memory.Timestamp.After(*query.EndTime) {
		return false
	}

	// 情感过滤
	if query.Emotion != "" && memory.Emotion != query.Emotion {
		return false
	}

	// 来源过滤
	if query.Source != "" && memory.Source != query.Source {
		return false
	}

	// 关键词搜索
	if query.Keywords != "" {
		// 简单实现：检查内容是否包含关键词
		// 实际应该用更复杂的搜索算法
		found := strings.Contains(memory.Content, query.Keywords) ||
			strings.Contains(memory.Summary, query.Keywords)
		if !found {
			return false
		}
	}

	return true
}

// sortMemories 排序记忆
func (s *FileStore) sortMemories(memories []*models.Memory, sortBy string, desc bool) {
	if sortBy == "" {
		sortBy = "timestamp"
	}

	// 简单冒泡排序
	for i := 0; i < len(memories)-1; i++ {
		for j := i + 1; j < len(memories); j++ {
			var shouldSwap bool
			switch sortBy {
			case "timestamp":
				shouldSwap = memories[j].Timestamp.After(memories[i].Timestamp)
			case "importance":
				shouldSwap = memories[j].GetEffectiveImportance() > memories[i].GetEffectiveImportance()
			case "access_count":
				shouldSwap = memories[j].AccessCount > memories[i].AccessCount
			default:
				shouldSwap = memories[j].Timestamp.After(memories[i].Timestamp)
			}

			if shouldSwap != desc {
				memories[i], memories[j] = memories[j], memories[i]
			}
		}
	}
}

// UpdateMemory 更新记忆
func (s *FileStore) UpdateMemory(ctx context.Context, memory *models.Memory) error {
	return s.SaveMemory(ctx, memory)
}

// SaveMemories 批量保存
func (s *FileStore) SaveMemories(ctx context.Context, memories []*models.Memory) error {
	for _, m := range memories {
		if err := s.SaveMemory(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMemories 批量删除
func (s *FileStore) DeleteMemories(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.DeleteMemory(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// GetMemoryStats 获取统计
func (s *FileStore) GetMemoryStats(ctx context.Context, userID string) (*models.MemoryStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userMemories, ok := s.userIndex[userID]
	if !ok {
		return &models.MemoryStats{UserID: userID}, nil
	}

	stats := &models.MemoryStats{
		UserID:        userID,
		CountByType:   make(map[models.MemoryType]int),
		CountByLayer:  make(map[models.MemoryLayer]int),
		TopCategories: []string{},
		TopTags:       []string{},
	}

	totalImportance := 0.0
	totalAccess := 0
	categoryCount := make(map[string]int)
	tagCount := make(map[string]int)

	for _, memory := range userMemories {
		stats.TotalCount++
		stats.CountByType[memory.Type]++
		stats.CountByLayer[memory.Layer]++
		totalImportance += memory.GetEffectiveImportance()
		totalAccess += memory.AccessCount

		if stats.OldestMemory.IsZero() || memory.Timestamp.Before(stats.OldestMemory) {
			stats.OldestMemory = memory.Timestamp
		}
		if memory.Timestamp.After(stats.NewestMemory) {
			stats.NewestMemory = memory.Timestamp
		}

		categoryCount[memory.Category]++
		for _, tag := range memory.Tags {
			tagCount[tag]++
		}
	}

	if stats.TotalCount > 0 {
		stats.AvgImportance = totalImportance / float64(stats.TotalCount)
		stats.AvgAccessCount = float64(totalAccess) / float64(stats.TotalCount)
	}

	// Top categories
	for i := 0; i < 5 && len(categoryCount) > 0; i++ {
		maxCat := ""
		maxCount := 0
		for cat, count := range categoryCount {
			if count > maxCount {
				maxCat = cat
				maxCount = count
			}
		}
		if maxCat != "" {
			stats.TopCategories = append(stats.TopCategories, maxCat)
			delete(categoryCount, maxCat)
		}
	}

	// Top tags
	for i := 0; i < 10 && len(tagCount) > 0; i++ {
		maxTag := ""
		maxCount := 0
		for tag, count := range tagCount {
			if count > maxCount {
				maxTag = tag
				maxCount = count
			}
		}
		if maxTag != "" {
			stats.TopTags = append(stats.TopTags, maxTag)
			delete(tagCount, maxTag)
		}
	}

	return stats, nil
}

// SaveAssociation 保存关联
func (s *FileStore) SaveAssociation(ctx context.Context, assoc *models.MemoryAssociation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.associations[assoc.SourceID] = append(s.associations[assoc.SourceID], assoc)

	return nil
}

// GetAssociations 获取关联
func (s *FileStore) GetAssociations(ctx context.Context, memoryID string) ([]*models.MemoryAssociation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.associations[memoryID], nil
}

// InMemoryStore - 内存存储实现（用于测试）
type InMemoryStore struct {
	mu          sync.RWMutex
	memories    map[string]*models.Memory
	userIndex   map[string]map[string]*models.Memory
	associations map[string][]*models.MemoryAssociation
}

// NewInMemoryStore 创建内存存储
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories:    make(map[string]*models.Memory),
		userIndex:   make(map[string]map[string]*models.Memory),
		associations: make(map[string][]*models.MemoryAssociation),
	}
}

// SaveMemory 保存记忆
func (s *InMemoryStore) SaveMemory(ctx context.Context, memory *models.Memory) error {
	if memory == nil || memory.ID == "" {
		return fmt.Errorf("invalid memory")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.memories[memory.ID] = memory

	if s.userIndex[memory.UserID] == nil {
		s.userIndex[memory.UserID] = make(map[string]*models.Memory)
	}
	s.userIndex[memory.UserID][memory.ID] = memory

	return nil
}

// GetMemory 获取记忆
func (s *InMemoryStore) GetMemory(ctx context.Context, id string) (*models.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memory, ok := s.memories[id]
	if !ok {
		return nil, fmt.Errorf("memory not found: %s", id)
	}

	return memory, nil
}

// DeleteMemory 删除记忆
func (s *InMemoryStore) DeleteMemory(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory, ok := s.memories[id]; ok {
		delete(s.memories, id)
		if userMemories, ok := s.userIndex[memory.UserID]; ok {
			delete(userMemories, id)
		}
	}
	delete(s.associations, id)

	return nil
}

// ListMemories 列出记忆
func (s *InMemoryStore) ListMemories(ctx context.Context, query *models.MemoryQuery) ([]*models.Memory, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userMemories, ok := s.userIndex[query.UserID]
	if !ok {
		return []*models.Memory{}, 0, nil
	}

	var filtered []*models.Memory
	for _, memory := range userMemories {
		// 简化过滤逻辑
		if len(query.Types) > 0 {
			found := false
			for _, t := range query.Types {
				if memory.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if query.MinImportance > 0 && memory.GetEffectiveImportance() < query.MinImportance {
			continue
		}

		filtered = append(filtered, memory)
	}

	total := len(filtered)

	// 简单排序
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Timestamp.After(filtered[i].Timestamp) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// 分页
	start := query.Offset
	if start > total {
		return []*models.Memory{}, total, nil
	}

	end := start + query.Limit
	if end > total || query.Limit == 0 {
		end = total
	}

	return filtered[start:end], total, nil
}

// UpdateMemory 更新记忆
func (s *InMemoryStore) UpdateMemory(ctx context.Context, memory *models.Memory) error {
	return s.SaveMemory(ctx, memory)
}

// SaveMemories 批量保存
func (s *InMemoryStore) SaveMemories(ctx context.Context, memories []*models.Memory) error {
	for _, m := range memories {
		if err := s.SaveMemory(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMemories 批量删除
func (s *InMemoryStore) DeleteMemories(ctx context.Context, ids []string) error {
	for _, id := range ids {
		s.DeleteMemory(ctx, id)
	}
	return nil
}

// GetMemoryStats 获取统计
func (s *InMemoryStore) GetMemoryStats(ctx context.Context, userID string) (*models.MemoryStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userMemories, ok := s.userIndex[userID]
	if !ok {
		return &models.MemoryStats{UserID: userID}, nil
	}

	stats := &models.MemoryStats{
		UserID:       userID,
		TotalCount:   len(userMemories),
		CountByType:  make(map[models.MemoryType]int),
		CountByLayer: make(map[models.MemoryLayer]int),
	}

	for _, m := range userMemories {
		stats.CountByType[m.Type]++
		stats.CountByLayer[m.Layer]++
	}

	return stats, nil
}

// SaveAssociation 保存关联
func (s *InMemoryStore) SaveAssociation(ctx context.Context, assoc *models.MemoryAssociation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.associations[assoc.SourceID] = append(s.associations[assoc.SourceID], assoc)
	return nil
}

// GetAssociations 获取关联
func (s *InMemoryStore) GetAssociations(ctx context.Context, memoryID string) ([]*models.MemoryAssociation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.associations[memoryID], nil
}
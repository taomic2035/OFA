package voice

import (
	"context"
	"fmt"
	"sync"

	"github.com/ofa/center/internal/models"
)

// MemoryStore - 内存存储实现
type MemoryStore struct {
	mu       sync.RWMutex
	profiles map[string]*models.VoiceProfile
	userIndex map[string][]string // userID -> []profileID
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		profiles:  make(map[string]*models.VoiceProfile),
		userIndex: make(map[string][]string),
	}
}

// SaveVoiceProfile 保存语音配置
func (s *MemoryStore) SaveVoiceProfile(ctx context.Context, profile *models.VoiceProfile) error {
	if profile == nil || profile.ID == "" {
		return fmt.Errorf("invalid profile")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.profiles[profile.ID] = profile

	return nil
}

// GetVoiceProfile 获取语音配置
func (s *MemoryStore) GetVoiceProfile(ctx context.Context, id string) (*models.VoiceProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.profiles[id]
	if !ok {
		return nil, fmt.Errorf("voice profile not found: %s", id)
	}

	return profile, nil
}

// DeleteVoiceProfile 删除语音配置
func (s *MemoryStore) DeleteVoiceProfile(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.profiles, id)

	// 从用户索引移除
	for userID, ids := range s.userIndex {
		for i, profileID := range ids {
			if profileID == id {
				s.userIndex[userID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}

	return nil
}

// ListVoiceProfiles 列出语音配置
func (s *MemoryStore) ListVoiceProfiles(ctx context.Context, userID string) ([]*models.VoiceProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var profiles []*models.VoiceProfile

	if userID == "" {
		for _, p := range s.profiles {
			profiles = append(profiles, p)
		}
	} else {
		ids := s.userIndex[userID]
		for _, id := range ids {
			if p, ok := s.profiles[id]; ok {
				profiles = append(profiles, p)
			}
		}
	}

	return profiles, nil
}
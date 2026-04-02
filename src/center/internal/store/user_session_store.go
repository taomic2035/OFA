package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/ofa/center/proto"
)

// MemoryUserStore - 内存用户存储
type MemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*proto.UserProfile
}

// NewMemoryUserStore 创建内存用户存储
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users: make(map[string]*proto.UserProfile),
	}
}

// CreateUser 创建用户
func (s *MemoryUserStore) CreateUser(ctx context.Context, user *proto.UserProfile) error {
	if user.Id == "" {
		user.Id = proto.GenerateID()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.Id] = user
	return nil
}

// GetUser 获取用户
func (s *MemoryUserStore) GetUser(ctx context.Context, userID string) (*proto.UserProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return user, nil
}

// UpdateUser 更新用户
func (s *MemoryUserStore) UpdateUser(ctx context.Context, user *proto.UserProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[user.Id]; !ok {
		return fmt.Errorf("user not found: %s", user.Id)
	}

	s.users[user.Id] = user
	return nil
}

// DeleteUser 删除用户
func (s *MemoryUserStore) DeleteUser(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, userID)
	return nil
}

// MemorySessionStore - 内存会话存储
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*proto.Session
	userSessions map[string]map[string]*proto.Session // userID -> sessionID -> session
}

// NewMemorySessionStore 创建内存会话存储
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions:     make(map[string]*proto.Session),
		userSessions: make(map[string]map[string]*proto.Session),
	}
}

// CreateSession 创建会话
func (s *MemorySessionStore) CreateSession(ctx context.Context, session *proto.Session) error {
	if session.Id == "" {
		session.Id = proto.GenerateID()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.Id] = session

	if s.userSessions[session.UserId] == nil {
		s.userSessions[session.UserId] = make(map[string]*proto.Session)
	}
	s.userSessions[session.UserId][session.Id] = session

	return nil
}

// GetSession 获取会话
func (s *MemorySessionStore) GetSession(ctx context.Context, sessionID string) (*proto.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// UpdateSession 更新会话
func (s *MemorySessionStore) UpdateSession(ctx context.Context, session *proto.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[session.Id]; !ok {
		return fmt.Errorf("session not found: %s", session.Id)
	}

	s.sessions[session.Id] = session
	s.userSessions[session.UserId][session.Id] = session

	return nil
}

// EndSession 结束会话
func (s *MemorySessionStore) EndSession(ctx context.Context, sessionID string, summary string) (*proto.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session.Status = "ended"
	session.EndedAt = proto.Now()
	// Store summary in context
	if session.Context == nil {
		session.Context = make(map[string]interface{})
	}
	session.Context["summary"] = summary

	s.sessions[sessionID] = session
	s.userSessions[session.UserId][session.Id] = session

	return session, nil
}

// GetActiveSessions 获取活跃会话
func (s *MemorySessionStore) GetActiveSessions(ctx context.Context, userID string) ([]*proto.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userSessions := s.userSessions[userID]
	if userSessions == nil {
		return []*proto.Session{}, nil
	}

	var active []*proto.Session
	for _, session := range userSessions {
		if session.Status == "active" {
			active = append(active, session)
		}
	}

	return active, nil
}
// Package wasm - WASM技能沙箱安全
package wasm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	MaxMemoryMB     int           `json:"max_memory_mb"`
	MaxCPUSeconds   int           `json:"max_cpu_seconds"`
	MaxFileSize     int64         `json:"max_file_size"`
	AllowedPaths    []string      `json:"allowed_paths"`
	AllowedHosts    []string      `json:"allowed_hosts"`
	AllowedPorts    []int         `json:"allowed_ports"`
	EnableTimeout   bool          `json:"enable_timeout"`
	EnableMemoryLimit bool        `json:"enable_memory_limit"`
	EnableCPULimit  bool          `json:"enable_cpu_limit"`
	EnableNetworkFilter bool      `json:"enable_network_filter"`
	EnableFileFilter bool         `json:"enable_file_filter"`
}

// DefaultSandboxConfig 默认沙箱配置
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		MaxMemoryMB:       64,
		MaxCPUSeconds:     5,
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		AllowedPaths:      []string{},
		AllowedHosts:      []string{},
		AllowedPorts:      []int{},
		EnableTimeout:     true,
		EnableMemoryLimit: true,
		EnableCPULimit:    true,
		EnableNetworkFilter: true,
		EnableFileFilter:  true,
	}
}

// Sandbox 沙箱
type Sandbox struct {
	config    SandboxConfig
	context   *SandboxContext
	resources *ResourceTracker
	violations *ViolationLog
	mu        sync.RWMutex
}

// SandboxContext 沙箱上下文
type SandboxContext struct {
	SkillID     string
	StartTime   time.Time
	MemoryUsed  int64
	CPUUsed     time.Duration
	Operations  int
	FilesAccess []string
	NetAccess   []string
}

// ResourceTracker 资源跟踪器
type ResourceTracker struct {
	memoryUsed  int64
	cpuStart    time.Time
	operations  int64
	memoryLimit int64
	cpuLimit    time.Duration
	mu          sync.RWMutex
}

// ViolationLog 违规日志
type ViolationLog struct {
	entries []Violation
	mu      sync.RWMutex
}

// Violation 违规记录
type Violation struct {
	Time      time.Time `json:"time"`
	SkillID   string    `json:"skill_id"`
	Type      string    `json:"type"` // memory, cpu, file, network
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // warning, critical, fatal
	Blocked   bool      `json:"blocked"`
}

// NewSandbox 创建沙箱
func NewSandbox(config SandboxConfig) *Sandbox {
	return &Sandbox{
		config:    config,
		context:   &SandboxContext{},
		resources: &ResourceTracker{},
		violations: &ViolationLog{},
	}
}

// Enter 进入沙箱
func (s *Sandbox) Enter(skillID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.context = &SandboxContext{
		SkillID:    skillID,
		StartTime:  time.Now(),
		MemoryUsed: 0,
		CPUUsed:    0,
		Operations: 0,
	}

	s.resources = &ResourceTracker{
		memoryLimit: int64(s.config.MaxMemoryMB) * 1024 * 1024,
		cpuLimit:    time.Duration(s.config.MaxCPUSeconds) * time.Second,
	}

	return nil
}

// Exit 退出沙箱
func (s *Sandbox) Exit() *SandboxContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.context
}

// CheckMemory 检查内存限制
func (s *Sandbox) CheckMemory(delta int64) error {
	if !s.config.EnableMemoryLimit {
		return nil
	}

	s.resources.mu.Lock()
	defer s.resources.mu.Unlock()

	newMemory := s.resources.memoryUsed + delta
	if newMemory > s.resources.memoryLimit {
		s.logViolation("memory", fmt.Sprintf("内存超限: %d > %d", newMemory, s.resources.memoryLimit), "critical", true)
		return fmt.Errorf("内存超限")
	}

	s.resources.memoryUsed = newMemory
	s.context.MemoryUsed = newMemory
	return nil
}

// CheckCPU 检查CPU限制
func (s *Sandbox) CheckCPU(delta time.Duration) error {
	if !s.config.EnableCPULimit {
		return nil
	}

	s.resources.mu.Lock()
	defer s.resources.mu.Unlock()

	newCPU := time.Since(s.context.StartTime)
	if newCPU > s.resources.cpuLimit {
		s.logViolation("cpu", fmt.Sprintf("CPU超限: %v > %v", newCPU, s.resources.cpuLimit), "critical", true)
		return fmt.Errorf("CPU超限")
	}

	s.context.CPUUsed = newCPU
	return nil
}

// CheckFileAccess 检查文件访问权限
func (s *Sandbox) CheckFileAccess(path string) error {
	if !s.config.EnableFileFilter {
		return nil
	}

	if len(s.config.AllowedPaths) == 0 {
		s.logViolation("file", fmt.Sprintf("文件访问被拒绝: %s (无允许路径)", path), "warning", true)
		return fmt.Errorf("文件访问被拒绝")
	}

	allowed := false
	for _, allowedPath := range s.config.AllowedPaths {
		if path == allowedPath || len(path) > len(allowedPath) && path[:len(allowedPath)] == allowedPath {
			allowed = true
			break
		}
	}

	if !allowed {
		s.logViolation("file", fmt.Sprintf("文件访问被拒绝: %s", path), "warning", true)
		return fmt.Errorf("文件访问被拒绝: %s", path)
	}

	s.context.FilesAccess = append(s.context.FilesAccess, path)
	return nil
}

// CheckNetworkAccess 检查网络访问权限
func (s *Sandbox) CheckNetworkAccess(host string, port int) error {
	if !s.config.EnableNetworkFilter {
		return nil
	}

	// 检查主机
	hostAllowed := len(s.config.AllowedHosts) == 0
	for _, allowedHost := range s.config.AllowedHosts {
		if host == allowedHost {
			hostAllowed = true
			break
		}
	}

	if !hostAllowed {
		s.logViolation("network", fmt.Sprintf("主机访问被拒绝: %s", host), "warning", true)
		return fmt.Errorf("主机访问被拒绝: %s", host)
	}

	// 检查端口
	portAllowed := len(s.config.AllowedPorts) == 0
	for _, allowedPort := range s.config.AllowedPorts {
		if port == allowedPort {
			portAllowed = true
			break
		}
	}

	if !portAllowed {
		s.logViolation("network", fmt.Sprintf("端口访问被拒绝: %d", port), "warning", true)
		return fmt.Errorf("端口访问被拒绝: %d", port)
	}

	s.context.NetAccess = append(s.context.NetAccess, fmt.Sprintf("%s:%d", host, port))
	return nil
}

// logViolation 记录违规
func (s *Sandbox) logViolation(violationType, message, severity string, blocked bool) {
	s.violations.mu.Lock()
	defer s.violations.mu.Unlock()

	s.violations.entries = append(s.violations.entries, Violation{
		Time:     time.Now(),
		SkillID:  s.context.SkillID,
		Type:     violationType,
		Message:  message,
		Severity: severity,
		Blocked:  blocked,
	})
}

// GetViolations 获取违规记录
func (s *Sandbox) GetViolations() []Violation {
	s.violations.mu.RLock()
	defer s.violations.mu.RUnlock()

	result := make([]Violation, len(s.violations.entries))
	copy(result, s.violations.entries)
	return result
}

// GetContext 获取上下文
func (s *Sandbox) GetContext() *SandboxContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.context
}

// Permission 权限
type Permission struct {
	Type    string `json:"type"`    // file, network, env, system
	Action  string `json:"action"`  // read, write, execute, connect
	Target  string `json:"target"`  // 目标路径/主机等
	Granted bool   `json:"granted"`
}

// PermissionManager 权限管理器
type PermissionManager struct {
	permissions map[string][]Permission
	mu          sync.RWMutex
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		permissions: make(map[string][]Permission),
	}
}

// Grant 授予权限
func (pm *PermissionManager) Grant(skillID string, permission Permission) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.permissions[skillID] = append(pm.permissions[skillID], permission)
}

// Revoke 撤销权限
func (pm *PermissionManager) Revoke(skillID string, permType, action string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	perms := pm.permissions[skillID]
	newPerms := make([]Permission, 0)
	for _, p := range perms {
		if !(p.Type == permType && p.Action == action) {
			newPerms = append(newPerms, p)
		}
	}
	pm.permissions[skillID] = newPerms
}

// Check 检查权限
func (pm *PermissionManager) Check(skillID, permType, action, target string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.permissions[skillID] {
		if p.Type == permType && (p.Action == action || p.Action == "*") {
			if p.Target == target || p.Target == "*" {
				return true
			}
		}
	}
	return false
}

// GetPermissions 获取权限列表
func (pm *PermissionManager) GetPermissions(skillID string) []Permission {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	perms := pm.permissions[skillID]
	result := make([]Permission, len(perms))
	copy(result, perms)
	return result
}

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []SecurityRule    `json:"rules"`
	Default     string            `json:"default"` // allow, deny
}

// SecurityRule 安全规则
type SecurityRule struct {
	Type     string `json:"type"`     // allow, deny
	Resource string `json:"resource"` // *, file:*, network:*, etc.
	Action   string `json:"action"`   // *, read, write, connect
}

// DefaultSecurityPolicy 默认安全策略
func DefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		Name:        "default",
		Description: "默认安全策略",
		Default:     "deny",
		Rules: []SecurityRule{
			{Type: "allow", Resource: "env:PATH", Action: "read"},
			{Type: "allow", Resource: "env:HOME", Action: "read"},
			{Type: "allow", Resource: "file:/tmp/*", Action: "read,write"},
		},
	}
}

// Evaluate 评估权限
func (p *SecurityPolicy) Evaluate(resource, action string) bool {
	for _, rule := range p.Rules {
		if p.matchResource(rule.Resource, resource) && p.matchAction(rule.Action, action) {
			return rule.Type == "allow"
		}
	}
	return p.Default == "allow"
}

// matchResource 匹配资源
func (p *SecurityPolicy) matchResource(pattern, resource string) bool {
	if pattern == "*" {
		return true
	}
	// 简化的通配符匹配
	return pattern == resource || pattern+"*" == resource[:len(pattern)]+"*"
}

// matchAction 匹配动作
func (p *SecurityPolicy) matchAction(pattern, action string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == action
}

// Isolate 隔离执行
func (s *Sandbox) Isolate(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		s.logViolation("timeout", "执行超时", "critical", true)
		return fmt.Errorf("执行超时")
	}
}
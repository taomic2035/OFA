package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SkillManager manages skill loading and lifecycle
type SkillManager struct {
	dataDir string
	skills  sync.Map // map[string]Skill

	mu sync.RWMutex
}

// NewSkillManager creates a new skill manager
func NewSkillManager(dataDir string) *SkillManager {
	return &SkillManager{
		dataDir: dataDir,
	}
}

// LoadSkills loads all skills from the skills directory
func (sm *SkillManager) LoadSkills() ([]Skill, error) {
	skillsDir := filepath.Join(sm.dataDir, "skills")

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, nil // No skills directory yet
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []Skill

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		skill, err := sm.loadSkill(skillPath)
		if err != nil {
			fmt.Printf("Failed to load skill %s: %v\n", entry.Name(), err)
			continue
		}

		skills = append(skills, skill)
	}

	return skills, nil
}

// loadSkill loads a single skill from directory
func (sm *SkillManager) loadSkill(path string) (Skill, error) {
	// Load skill manifest
	manifestPath := filepath.Join(path, "skill.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill.json: %v", err)
	}

	var manifest SkillManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse skill.json: %v", err)
	}

	// Create skill based on type
	switch manifest.Type {
	case "script":
		return sm.loadScriptSkill(path, &manifest)
	case "binary":
		return sm.loadBinarySkill(path, &manifest)
	case "wasm":
		return sm.loadWASMSkill(path, &manifest)
	default:
		return nil, fmt.Errorf("unknown skill type: %s", manifest.Type)
	}
}

// SkillManifest defines skill metadata
type SkillManifest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type"`     // script, binary, wasm
	EntryPoint  string                 `json:"entrypoint"`
	Operations  []string               `json:"operations"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
}

// loadScriptSkill loads a script-based skill
func (sm *SkillManager) loadScriptSkill(path string, manifest *SkillManifest) (Skill, error) {
	return &ScriptSkill{
		path:      path,
		manifest:  manifest,
		entryPoint: manifest.EntryPoint,
	}, nil
}

// loadBinarySkill loads a binary executable skill
func (sm *SkillManager) loadBinarySkill(path string, manifest *SkillManifest) (Skill, error) {
	return &BinarySkill{
		path:      path,
		manifest:  manifest,
		entryPoint: manifest.EntryPoint,
	}, nil
}

// loadWASMSkill loads a WebAssembly skill
func (sm *SkillManager) loadWASMSkill(path string, manifest *SkillManifest) (Skill, error) {
	return &WASMSkill{
		path:      path,
		manifest:  manifest,
	}, nil
}

// ScriptSkill represents a script-based skill
type ScriptSkill struct {
	path       string
	manifest   *SkillManifest
	entryPoint string
}

func (s *ScriptSkill) ID() string          { return s.manifest.ID }
func (s *ScriptSkill) Name() string        { return s.manifest.Name }
func (s *ScriptSkill) Operations() []string { return s.manifest.Operations }

func (s *ScriptSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	// Execute script with operation and params
	scriptPath := filepath.Join(s.path, s.entryPoint)
	return ExecuteScript(scriptPath, operation, params)
}

// BinarySkill represents a binary executable skill
type BinarySkill struct {
	path       string
	manifest   *SkillManifest
	entryPoint string
}

func (s *BinarySkill) ID() string          { return s.manifest.ID }
func (s *BinarySkill) Name() string        { return s.manifest.Name }
func (s *BinarySkill) Operations() []string { return s.manifest.Operations }

func (s *BinarySkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	binaryPath := filepath.Join(s.path, s.entryPoint)
	return ExecuteBinary(binaryPath, operation, params)
}

// WASMSkill represents a WebAssembly skill
type WASMSkill struct {
	path     string
	manifest *SkillManifest
}

func (s *WASMSkill) ID() string          { return s.manifest.ID }
func (s *WASMSkill) Name() string        { return s.manifest.Name }
func (s *WASMSkill) Operations() []string { return s.manifest.Operations }

func (s *WASMSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	// WASM execution would be implemented here
	return nil, fmt.Errorf("WASM skills not yet supported")
}

// GRPCConnector implements Connector using gRPC
type GRPCConnector struct {
	port       int
	connected  bool
	taskChan   chan Task
	ctx        context.Context
	cancel     context.CancelFunc

	mu sync.RWMutex
}

// NewGRPCConnector creates a new gRPC connector
func NewGRPCConnector(port int) *GRPCConnector {
	ctx, cancel := context.WithCancel(context.Background())
	return &GRPCConnector{
		port:     port,
		taskChan: make(chan Task, 100),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (c *GRPCConnector) Connect(ctx context.Context, url string) error {
	// gRPC connection implementation
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// Start receiving tasks
	go c.receiveLoop()

	return nil
}

func (c *GRPCConnector) Disconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	c.cancel()
	return nil
}

func (c *GRPCConnector) Register(capabilities []Capability) error {
	// Register capabilities with Center
	return nil
}

func (c *GRPCConnector) SendHeartbeat() error {
	// Send heartbeat to Center
	return nil
}

func (c *GRPCConnector) ReceiveTasks() <-chan Task {
	return c.taskChan
}

func (c *GRPCConnector) SendResult(result TaskResult) error {
	// Send result to Center
	return nil
}

func (c *GRPCConnector) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *GRPCConnector) receiveLoop() {
	// Receive tasks from Center via gRPC stream
}